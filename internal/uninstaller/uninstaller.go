package uninstaller

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjg184/agent-smith/internal/detector"
	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/internal/linker"
	"github.com/tjg184/agent-smith/internal/metadata"
	"github.com/tjg184/agent-smith/internal/models"
	"github.com/tjg184/agent-smith/pkg/paths"
)

// Uninstaller handles component removal
type Uninstaller struct {
	baseDir   string
	linker    *linker.ComponentLinker
	formatter *formatter.Formatter
}

func NewUninstaller(baseDir string, componentLinker *linker.ComponentLinker) *Uninstaller {
	return &Uninstaller{
		baseDir:   baseDir,
		linker:    componentLinker,
		formatter: formatter.New(),
	}
}

func (u *Uninstaller) UninstallComponent(componentType, name, source string) error {
	if componentType != "skills" && componentType != "agents" && componentType != "commands" {
		return fmt.Errorf("invalid component type: %s (must be skills, agents, or commands)", componentType)
	}

	var entry *models.ComponentEntry
	var err error
	if source != "" {
		entry, err = metadata.LoadLockFileEntryBySource(u.baseDir, componentType, name, source)
	} else {
		entry, err = metadata.LoadLockFileEntry(u.baseDir, componentType, name)
	}
	if err != nil {
		return fmt.Errorf("component '%s' not installed", name)
	}

	dirName := name
	if entry.FilesystemName != "" {
		dirName = entry.FilesystemName
	}
	componentDir := filepath.Join(u.baseDir, componentType, dirName)

	u.formatter.SectionHeader(fmt.Sprintf("Uninstalling %s: %s", componentType, name))

	u.formatter.InfoMsg("The following will be removed:")
	u.formatter.EmptyLine()

	if entry != nil {
		if entry.SourceUrl != "" {
			u.formatter.DetailItem("Source", entry.SourceUrl)
		} else if entry.Source != "" {
			u.formatter.DetailItem("Source", entry.Source)
		}
		if entry.CommitHash != "" {
			u.formatter.DetailItem("Commit", entry.CommitHash)
		}
	}

	u.formatter.DetailItem("Directory", componentDir)

	linkedTargets := u.findLinkedTargets(componentType, name)
	if len(linkedTargets) > 0 {
		u.formatter.DetailItem("Linked to", strings.Join(linkedTargets, ", "))
	} else {
		u.formatter.DetailItem("Linked to", "(none)")
	}

	u.formatter.EmptyLine()

	if u.linker != nil && len(linkedTargets) > 0 {
		u.formatter.ProgressMsg("Unlinking from targets", name)
		_ = u.linker.UnlinkComponent(componentType, name, "")
		u.formatter.ProgressComplete()
	}

	u.formatter.ProgressMsg("Removing directory", componentDir)
	resolvedSource := source
	if resolvedSource == "" && entry != nil {
		resolvedSource = entry.SourceUrl
	}
	sharedDir, err := u.isDirectorySharedByOtherSource(componentType, name, dirName, resolvedSource)
	if err != nil {
		u.formatter.ProgressFailed()
		return fmt.Errorf("failed to check directory sharing: %w", err)
	}
	if sharedDir {
		u.formatter.ProgressComplete()
		u.formatter.DetailItem("Note", "Directory kept (referenced by another source)")
	} else {
		if err := os.RemoveAll(componentDir); err != nil {
			u.formatter.ProgressFailed()
			return fmt.Errorf("failed to remove component directory: %w", err)
		}
		u.formatter.ProgressComplete()
	}

	u.formatter.ProgressMsg("Updating lock file", "")
	removeSource := source
	if removeSource == "" && entry != nil {
		removeSource = entry.SourceUrl
	}
	var lockErr error
	if removeSource != "" {
		lockErr = metadata.RemoveComponentEntryBySource(u.baseDir, componentType, name, removeSource)
	} else {
		lockErr = metadata.RemoveComponentEntry(u.baseDir, componentType, name)
	}
	if lockErr != nil {
		u.formatter.ProgressFailed()
		u.formatter.WarningMsg("Could not update lock file: %v", lockErr)
	} else {
		u.formatter.ProgressComplete()
	}

	u.formatter.EmptyLine()
	u.formatter.SuccessMsg("Removed %s: %s", componentType, name)

	return nil
}

type componentsByTypeInDir struct {
	baseDir          string
	componentsByType map[string][]string
}

// UninstallAllFromSource removes all components from a specified repository URL.
// It searches both the base directory and all profile directories.
func (u *Uninstaller) UninstallAllFromSource(repoURL string, force bool) error {
	return u.UninstallAllFromSourceAcrossDirs(repoURL, nil, force)
}

func (u *Uninstaller) UninstallAllFromSourceAcrossDirs(repoURL string, extraDirs []string, force bool) error {
	det := detector.NewRepositoryDetector()
	normalizedURL, err := det.NormalizeURL(repoURL)
	if err != nil {
		return fmt.Errorf("invalid repository URL: %w", err)
	}

	searchDirs := append([]string{u.baseDir}, extraDirs...)

	var allMatches []componentsByTypeInDir
	for _, dir := range searchDirs {
		byType, err := u.findComponentsBySourceInDir(dir, normalizedURL)
		if err != nil {
			return fmt.Errorf("failed to find components in %s: %w", dir, err)
		}
		if len(byType) > 0 {
			allMatches = append(allMatches, componentsByTypeInDir{baseDir: dir, componentsByType: byType})
		}
	}

	totalComponents := 0
	for _, m := range allMatches {
		for _, names := range m.componentsByType {
			totalComponents += len(names)
		}
	}

	if totalComponents == 0 {
		u.formatter.InfoMsg("No components found from repository: %s", repoURL)
		return nil
	}

	u.formatter.SectionHeader("Uninstall Preview")

	u.formatter.DetailItem("Repository", repoURL)
	u.formatter.DetailItem("Total components", fmt.Sprintf("%d", totalComponents))
	u.formatter.EmptyLine()

	u.formatter.InfoMsg("Components to be removed:")
	u.formatter.EmptyLine()

	for _, m := range allMatches {
		if len(m.componentsByType) > 0 {
			u.formatter.Info("From %s:", m.baseDir)
			for componentType, names := range m.componentsByType {
				if len(names) > 0 {
					u.formatter.Info("  %s (%d):", strings.Title(componentType), len(names))
					for _, name := range names {
						linkedTargets := u.findLinkedTargets(componentType, name)
						if len(linkedTargets) > 0 {
							u.formatter.ListItem("  %s (linked to: %s)", name, strings.Join(linkedTargets, ", "))
						} else {
							u.formatter.ListItem("  %s", name)
						}
					}
				}
			}
			u.formatter.EmptyLine()
		}
	}

	u.formatter.InfoMsg("Directories to be deleted:")
	u.formatter.EmptyLine()
	for _, m := range allMatches {
		for componentType, names := range m.componentsByType {
			for _, name := range names {
				entry, err := metadata.LoadLockFileEntry(m.baseDir, componentType, name)
				dirName := name
				if err == nil && entry.FilesystemName != "" {
					dirName = entry.FilesystemName
				}
				componentDir := filepath.Join(m.baseDir, componentType, dirName)
				u.formatter.ListItem("%s", componentDir)
			}
		}
	}
	u.formatter.EmptyLine()

	if !force {
		fmt.Print("Remove these components? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read user input: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			u.formatter.InfoMsg("Uninstall cancelled")
			return nil
		}
		u.formatter.EmptyLine()
	}

	u.formatter.SectionHeader("Removing Components")

	removed := 0
	failed := 0

	for _, m := range allMatches {
		for _, componentType := range []string{"skills", "agents", "commands"} {
			names, exists := m.componentsByType[componentType]
			if !exists || len(names) == 0 {
				continue
			}

			for _, name := range names {
				if u.linker != nil {
					_ = u.linker.UnlinkComponent(componentType, name, "")
				}

				u.formatter.ProgressMsg(fmt.Sprintf("Removing %s", componentType), name)

				entry, err := metadata.LoadLockFileEntry(m.baseDir, componentType, name)
				dirName := name
				if err == nil && entry.FilesystemName != "" {
					dirName = entry.FilesystemName
				}

				componentDir := filepath.Join(m.baseDir, componentType, dirName)
				if err := os.RemoveAll(componentDir); err != nil {
					u.formatter.ProgressFailed()
					u.formatter.ErrorMsg("Failed to remove %s: %s (%v)", componentType, name, err)
					failed++
					continue
				}

				if err := metadata.RemoveComponentEntry(m.baseDir, componentType, name); err != nil {
					u.formatter.WarningMsg("Could not update lock file for %s: %s", name, err)
				}

				u.formatter.ProgressComplete()
				removed++
			}
		}
	}

	u.formatter.EmptyLine()
	u.formatter.CounterSummary(totalComponents, removed, failed, 0)

	if failed > 0 {
		return fmt.Errorf("failed to remove %d component(s)", failed)
	}

	return nil
}

func (u *Uninstaller) findComponentsBySource(normalizedURL string) (map[string][]string, error) {
	return u.findComponentsBySourceInDir(u.baseDir, normalizedURL)
}

func (u *Uninstaller) findComponentsBySourceInDir(dir, normalizedURL string) (map[string][]string, error) {
	componentsByType := make(map[string][]string)

	componentTypes := []string{"skills", "agents", "commands"}

	for _, componentType := range componentTypes {
		lockFilePath := paths.GetComponentLockPath(dir, componentType)

		if _, err := os.Stat(lockFilePath); os.IsNotExist(err) {
			continue
		}

		lockData, err := os.ReadFile(lockFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read lock file for %s: %w", componentType, err)
		}

		var lockFile models.ComponentLockFile
		if err := json.Unmarshal(lockData, &lockFile); err != nil {
			return nil, fmt.Errorf("failed to parse lock file for %s: %w", componentType, err)
		}

		var nestedEntries map[string]map[string]models.ComponentEntry
		switch componentType {
		case "skills":
			nestedEntries = lockFile.Skills
		case "agents":
			nestedEntries = lockFile.Agents
		case "commands":
			nestedEntries = lockFile.Commands
		default:
			continue
		}

		var matchingNames []string
		for sourceURL, entries := range nestedEntries {
			for name, entry := range entries {
				if u.matchesSourceURL(sourceURL, normalizedURL) ||
					u.matchesSourceURL(entry.SourceUrl, normalizedURL) ||
					u.matchesSourceURL(entry.Source, normalizedURL) {
					matchingNames = append(matchingNames, name)
				}
			}
		}

		if len(matchingNames) > 0 {
			componentsByType[componentType] = matchingNames
		}
	}

	return componentsByType, nil
}

func (u *Uninstaller) matchesSourceURL(url1, url2 string) bool {
	if url1 == "" || url2 == "" {
		return false
	}

	normalized1 := normalizeURLForComparison(url1)
	normalized2 := normalizeURLForComparison(url2)

	return normalized1 == normalized2
}

func normalizeURLForComparison(url string) string {
	url = strings.TrimSpace(url)
	url = strings.TrimSuffix(url, "/")
	url = strings.TrimSuffix(url, ".git")
	url = strings.ToLower(url)
	return url
}

func (u *Uninstaller) findLinkedTargets(componentType, componentName string) []string {
	var linkedTargets []string

	if u.linker == nil {
		return linkedTargets
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return linkedTargets
	}

	targetPaths := map[string]string{
		"opencode":   filepath.Join(homeDir, ".config", "opencode", componentType),
		"claudecode": filepath.Join(homeDir, ".claude", componentType),
	}

	for targetName, targetDir := range targetPaths {
		componentPath := filepath.Join(targetDir, componentName)
		if _, err := os.Lstat(componentPath); err == nil {
			linkedTargets = append(linkedTargets, targetName)
		}
	}

	return linkedTargets
}

// isDirectorySharedByOtherSource returns true when another source entry in the lock
// file maps to the same filesystemName, meaning the on-disk directory must not be
// deleted — only the lock entry for the given source should be removed.
func (u *Uninstaller) isDirectorySharedByOtherSource(componentType, name, filesystemName, removingSource string) (bool, error) {
	instances, err := metadata.FindAllComponentInstances(u.baseDir, componentType, name)
	if err != nil {
		return false, err
	}

	for _, inst := range instances {
		if inst.SourceUrl == removingSource {
			continue
		}
		instDirName := inst.Entry.FilesystemName
		if instDirName == "" {
			instDirName = name
		}
		if instDirName == filesystemName {
			return true, nil
		}
	}

	return false, nil
}
