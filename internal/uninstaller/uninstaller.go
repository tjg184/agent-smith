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

// NewUninstaller creates a new Uninstaller instance
func NewUninstaller(baseDir string, componentLinker *linker.ComponentLinker) *Uninstaller {
	return &Uninstaller{
		baseDir:   baseDir,
		linker:    componentLinker,
		formatter: formatter.New(),
	}
}

// UninstallComponent removes a single component
func (u *Uninstaller) UninstallComponent(componentType, name string) error {
	// Validate component type
	if componentType != "skills" && componentType != "agents" && componentType != "commands" {
		return fmt.Errorf("invalid component type: %s (must be skills, agents, or commands)", componentType)
	}

	// Check if component exists in lock file
	entry, err := metadata.LoadLockFileEntry(u.baseDir, componentType, name)
	if err != nil {
		return fmt.Errorf("component '%s' not installed", name)
	}

	// Get component directory path - use FilesystemName if available (for conflict resolution)
	dirName := name
	if entry.FilesystemName != "" {
		dirName = entry.FilesystemName
	}
	componentDir := filepath.Join(u.baseDir, componentType, dirName)

	// Display section header
	u.formatter.SectionHeader(fmt.Sprintf("Uninstalling %s: %s", componentType, name))

	// Show what will be removed
	u.formatter.InfoMsg("The following will be removed:")
	u.formatter.EmptyLine()

	// Show source information if available
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

	// Find and display linked targets
	linkedTargets := u.findLinkedTargets(componentType, name)
	if len(linkedTargets) > 0 {
		u.formatter.DetailItem("Linked to", strings.Join(linkedTargets, ", "))
	} else {
		u.formatter.DetailItem("Linked to", "(none)")
	}

	u.formatter.EmptyLine()

	// Auto-unlink component from all targets (silent if not linked)
	if u.linker != nil && len(linkedTargets) > 0 {
		u.formatter.ProgressMsg("Unlinking from targets", name)
		// Try to unlink, but don't fail if it's not linked
		// Pass empty targetFilter to unlink from all targets
		_ = u.linker.UnlinkComponent(componentType, name, "")
		u.formatter.ProgressComplete()
	}

	// Remove component directory from filesystem
	u.formatter.ProgressMsg("Removing directory", componentDir)
	if err := os.RemoveAll(componentDir); err != nil {
		u.formatter.ProgressFailed()
		return fmt.Errorf("failed to remove component directory: %w", err)
	}
	u.formatter.ProgressComplete()

	// Remove entry from lock file
	u.formatter.ProgressMsg("Updating lock file", "")
	if err := metadata.RemoveLockFileEntry(u.baseDir, componentType, name); err != nil {
		u.formatter.ProgressFailed()
		// Log warning but continue - directory is already removed
		u.formatter.WarningMsg("Could not update lock file: %v", err)
	} else {
		u.formatter.ProgressComplete()
	}

	// Display success message
	u.formatter.EmptyLine()
	u.formatter.SuccessMsg("Removed %s: %s", componentType, name)

	return nil
}

// UninstallAllFromSource removes all components from a specified repository URL
func (u *Uninstaller) UninstallAllFromSource(repoURL string, force bool) error {
	// Normalize repository URL
	detector := detector.NewRepositoryDetector()
	normalizedURL, err := detector.NormalizeURL(repoURL)
	if err != nil {
		return fmt.Errorf("invalid repository URL: %w", err)
	}

	// Find all components from this source
	componentsByType, err := u.findComponentsBySource(normalizedURL)
	if err != nil {
		return fmt.Errorf("failed to find components: %w", err)
	}

	// Check if any components found
	totalComponents := 0
	for _, names := range componentsByType {
		totalComponents += len(names)
	}

	if totalComponents == 0 {
		u.formatter.InfoMsg("No components found from %s", repoURL)
		return nil
	}

	// Display section header
	u.formatter.SectionHeader("Uninstall Preview")

	// Display repository and total count
	u.formatter.DetailItem("Repository", repoURL)
	u.formatter.DetailItem("Total components", fmt.Sprintf("%d", totalComponents))
	u.formatter.EmptyLine()

	// Show breakdown by type with more details
	u.formatter.InfoMsg("Components to be removed:")
	u.formatter.EmptyLine()

	for componentType, names := range componentsByType {
		if len(names) > 0 {
			// Component type header
			u.formatter.Info("%s (%d):", strings.Title(componentType), len(names))
			for _, name := range names {
				// Find linked targets for this component
				linkedTargets := u.findLinkedTargets(componentType, name)

				if len(linkedTargets) > 0 {
					u.formatter.ListItem("%s (linked to: %s)", name, strings.Join(linkedTargets, ", "))
				} else {
					u.formatter.ListItem("%s", name)
				}
			}
			u.formatter.EmptyLine()
		}
	}

	// Show what directories will be deleted
	u.formatter.InfoMsg("Directories to be deleted:")
	u.formatter.EmptyLine()
	for componentType, names := range componentsByType {
		if len(names) > 0 {
			for _, name := range names {
				// Load lock entry to get filesystem name
				entry, err := metadata.LoadLockFileEntry(u.baseDir, componentType, name)
				dirName := name
				if err == nil && entry.FilesystemName != "" {
					dirName = entry.FilesystemName
				}
				componentDir := filepath.Join(u.baseDir, componentType, dirName)
				u.formatter.ListItem("%s", componentDir)
			}
		}
	}
	u.formatter.EmptyLine()

	// Prompt for confirmation unless --force is set
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

	// Display removal header
	u.formatter.SectionHeader("Removing Components")

	// Remove all components
	removed := 0
	failed := 0
	var failedComponents []string

	// Process in order: skills, agents, commands
	for _, componentType := range []string{"skills", "agents", "commands"} {
		names, exists := componentsByType[componentType]
		if !exists || len(names) == 0 {
			continue
		}

		for _, name := range names {
			// Auto-unlink component from all targets (silent if not linked)
			if u.linker != nil {
				// Pass empty targetFilter to unlink from all targets
				_ = u.linker.UnlinkComponent(componentType, name, "")
			}

			// Show progress for removal
			u.formatter.ProgressMsg(fmt.Sprintf("Removing %s", componentType), name)

			// Load lock entry to get filesystem name
			entry, err := metadata.LoadLockFileEntry(u.baseDir, componentType, name)
			dirName := name
			if err == nil && entry.FilesystemName != "" {
				dirName = entry.FilesystemName
			}

			// Remove component directory from filesystem
			componentDir := filepath.Join(u.baseDir, componentType, dirName)
			if err := os.RemoveAll(componentDir); err != nil {
				u.formatter.ProgressFailed()
				u.formatter.ErrorMsg("Failed to remove %s: %s (%v)", componentType, name, err)
				failed++
				failedComponents = append(failedComponents, fmt.Sprintf("%s/%s", componentType, name))
				continue
			}

			// Remove entry from lock file
			if err := metadata.RemoveLockFileEntry(u.baseDir, componentType, name); err != nil {
				u.formatter.WarningMsg("Could not update lock file for %s: %s", name, err)
			}

			u.formatter.ProgressComplete()
			removed++
		}
	}

	// Display summary
	u.formatter.EmptyLine()
	u.formatter.CounterSummary(totalComponents, removed, failed, 0)

	if failed > 0 {
		return fmt.Errorf("failed to remove %d component(s)", failed)
	}

	return nil
}

// findComponentsBySource scans lock files and finds all components from a source URL
func (u *Uninstaller) findComponentsBySource(normalizedURL string) (map[string][]string, error) {
	componentsByType := make(map[string][]string)

	// Component types to check
	componentTypes := []string{"skills", "agents", "commands"}

	for _, componentType := range componentTypes {
		lockFilePath := paths.GetComponentLockPath(u.baseDir, componentType)

		// Check if lock file exists
		if _, err := os.Stat(lockFilePath); os.IsNotExist(err) {
			continue
		}

		// Read lock file
		lockData, err := os.ReadFile(lockFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read lock file for %s: %w", componentType, err)
		}

		// Parse lock file
		var lockFile models.ComponentLockFile
		if err := json.Unmarshal(lockData, &lockFile); err != nil {
			return nil, fmt.Errorf("failed to parse lock file for %s: %w", componentType, err)
		}

		// Get the appropriate nested map for this component type
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

		// Find components matching the source URL (iterate through nested structure)
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

// matchesSourceURL checks if two URLs match (with normalization)
func (u *Uninstaller) matchesSourceURL(url1, url2 string) bool {
	if url1 == "" || url2 == "" {
		return false
	}

	// Normalize both URLs for comparison
	normalized1 := normalizeURLForComparison(url1)
	normalized2 := normalizeURLForComparison(url2)

	return normalized1 == normalized2
}

// normalizeURLForComparison normalizes a URL for comparison purposes
func normalizeURLForComparison(url string) string {
	// Trim whitespace
	url = strings.TrimSpace(url)

	// Remove trailing slash
	url = strings.TrimSuffix(url, "/")

	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")

	// Convert to lowercase for case-insensitive comparison
	url = strings.ToLower(url)

	return url
}

// findLinkedTargets checks which targets a component is currently linked to
func (u *Uninstaller) findLinkedTargets(componentType, componentName string) []string {
	var linkedTargets []string

	if u.linker == nil {
		return linkedTargets
	}

	// Get the list of targets from the linker
	// We need to check each target to see if the component is linked there
	// This requires accessing the targets, which we can do through the detector

	// For now, we'll check common target directories
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return linkedTargets
	}

	// Check common target locations (both macOS and Linux paths)
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
