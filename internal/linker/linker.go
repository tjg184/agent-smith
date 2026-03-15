package linker

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tjg184/agent-smith/internal/detector"
	"github.com/tjg184/agent-smith/internal/fileutil"
	"github.com/tjg184/agent-smith/internal/formatter"
	metadataPkg "github.com/tjg184/agent-smith/internal/metadata"
	"github.com/tjg184/agent-smith/internal/models"
	"github.com/tjg184/agent-smith/pkg/colors"
	"github.com/tjg184/agent-smith/pkg/config"
	"github.com/tjg184/agent-smith/pkg/paths"
	"github.com/tjg184/agent-smith/pkg/styles"
)

type ComponentLinker struct {
	agentsDir      string
	targets        []config.Target
	detector       *detector.RepositoryDetector
	profileManager ProfileManager // Optional - can be nil
	formatter      *formatter.Formatter
}

// ProfileManager prevents circular dependencies between linker and profiles packages.
type ProfileManager interface {
	ScanProfiles() ([]*Profile, error)
	GetActiveProfile() (string, error)
}

func displayName(name string) string {
	if name == "" {
		return ""
	}
	displayNames := map[string]string{
		"opencode":   "OpenCode",
		"claudecode": "ClaudeCode",
		"copilot":    "Copilot",
		"universal":  "Universal",
	}
	if d, ok := displayNames[name]; ok {
		return d
	}
	replaced := strings.ReplaceAll(name, "-", " ")
	replaced = strings.ReplaceAll(replaced, "_", " ")
	words := strings.Fields(replaced)
	for i, word := range words {
		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}
	return strings.Join(words, "")
}

// Profile must match the Profile struct from pkg/profiles/profiles.go
type Profile struct {
	Name        string
	BasePath    string
	HasAgents   bool
	HasSkills   bool
	HasCommands bool
}

func NewComponentLinker(agentsDir string, targets []config.Target, det *detector.RepositoryDetector, pm ProfileManager) (*ComponentLinker, error) {
	if agentsDir == "" {
		return nil, fmt.Errorf("agentsDir cannot be empty")
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("at least one target must be provided")
	}
	if det == nil {
		return nil, fmt.Errorf("detector cannot be nil")
	}

	for _, target := range targets {
		targetDir, err := target.GetGlobalBaseDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get target base directory: %w", err)
		}

		if err := fileutil.CreateDirectoryWithPermissions(targetDir); err != nil {
			return nil, fmt.Errorf("failed to create target directory: %w", err)
		}
	}

	return &ComponentLinker{
		agentsDir:      agentsDir,
		targets:        targets,
		detector:       det,
		profileManager: pm,
		formatter:      formatter.New(),
	}, nil
}

func (cl *ComponentLinker) SetFormatter(f *formatter.Formatter) {
	cl.formatter = f
}

// filterTargets returns all targets when targetFilter is empty or "all", otherwise returns only the matching target.
func (cl *ComponentLinker) filterTargets(targetFilter string) []config.Target {
	if targetFilter == "" || targetFilter == "all" {
		return cl.targets
	}

	filtered := make([]config.Target, 0)
	for _, target := range cl.targets {
		if target.GetName() == targetFilter {
			filtered = append(filtered, target)
		}
	}

	return filtered
}

func (cl *ComponentLinker) createSymlink(src, dst string) error {
	if _, err := os.Lstat(dst); err == nil {
		os.Remove(dst)
	}

	dstDir := filepath.Dir(dst)
	if realDir, err := filepath.EvalSymlinks(dstDir); err == nil {
		dstDir = realDir
	}

	relPath, err := filepath.Rel(dstDir, src)
	if err != nil {
		return fmt.Errorf("failed to create relative path: %w", err)
	}

	if err := os.Symlink(relPath, dst); err != nil {
		if runtime.GOOS == "windows" {
			return cl.createJunction(src, dst)
		}
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

// createJunction creates a Windows junction, falling back to directory copy.
func (cl *ComponentLinker) createJunction(src, dst string) error {
	// Windows junctions require Windows API; fall back to copying as a cross-platform solution.
	return cl.copyDirectory(src, dst)
}

func (cl *ComponentLinker) copyDirectory(src, dst string) error {
	return fileutil.CopyDirectoryContents(src, dst)
}

func (cl *ComponentLinker) copyFile(src, dst string) error {
	return fileutil.CopyFile(src, dst)
}

func (cl *ComponentLinker) LinkComponent(componentType, componentName string) error {
	return cl.linkComponentInternal(componentType, componentName, true)
}

func (cl *ComponentLinker) linkComponentInternal(componentType, componentName string, verbose bool) error {
	srcDir := filepath.Join(cl.agentsDir, componentType, componentName)
	selectedProfileName := ""

	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		matches, searchErr := cl.searchComponentInProfiles(componentType, componentName)
		if searchErr != nil {
			return fmt.Errorf("failed to search profiles: %w", searchErr)
		}

		if len(matches) == 0 {
			return fmt.Errorf("component %s/%s does not exist in any profile", componentType, componentName)
		}

		if len(matches) > 1 {
			profilePath, profileName, err := cl.promptProfileSelection(componentType, componentName, matches)
			if err != nil {
				return err
			}
			srcDir = filepath.Join(profilePath, componentType, componentName)
			selectedProfileName = profileName
		} else {
			srcDir = filepath.Join(matches[0].ProfilePath, componentType, componentName)
			selectedProfileName = matches[0].ProfileName
			fmt.Printf("  %s Component found in profile: %s\n", colors.Muted("→"), selectedProfileName)
		}
	} else {
		selectedProfileName = getProfileFromPath(srcDir)
	}

	// All components are now stored type-based, no special plugin handling needed
	metadata := cl.loadComponentMetadata(componentType, componentName)
	_ = metadata
	type linkResult struct {
		name    string
		path    string
		success bool
		errMsg  string
	}
	var linkResults []linkResult

	for _, target := range cl.targets {
		targetName := target.GetName()

		componentDir, err := target.GetGlobalComponentDir(componentType)
		if err != nil {
			linkResults = append(linkResults, linkResult{
				name:    targetName,
				success: false,
				errMsg:  fmt.Sprintf("failed to get target component directory: %v", err),
			})
			continue
		}
		if componentType == "commands" || componentType == "agents" {
			if err := fileutil.CreateDirectoryWithPermissions(componentDir); err != nil {
				linkResults = append(linkResults, linkResult{
					name:    targetName,
					success: false,
					errMsg:  fmt.Sprintf("failed to create destination directory: %v", err),
				})
				continue
			}

			linked, err := linkFlatMdFiles(srcDir, componentDir)
			if err != nil || len(linked) == 0 {
				msg := fmt.Sprintf("failed to link: %v", err)
				if err == nil {
					msg = "no .md files found to link"
				}
				linkResults = append(linkResults, linkResult{
					name:    targetName,
					success: false,
					errMsg:  msg,
				})
				continue
			}

			linkResults = append(linkResults, linkResult{
				name:    targetName,
				path:    componentDir,
				success: true,
			})
			continue
		}

		dstDir := filepath.Join(componentDir, componentName)

		if err := fileutil.CreateDirectoryWithPermissions(filepath.Dir(dstDir)); err != nil {
			linkResults = append(linkResults, linkResult{
				name:    targetName,
				path:    dstDir,
				success: false,
				errMsg:  fmt.Sprintf("failed to create destination directory: %v", err),
			})
			continue
		}

		if err := cl.createSymlink(srcDir, dstDir); err != nil {
			linkResults = append(linkResults, linkResult{
				name:    targetName,
				path:    dstDir,
				success: false,
				errMsg:  fmt.Sprintf("failed to link: %v", err),
			})
			continue
		}

		linkResults = append(linkResults, linkResult{
			name:    targetName,
			path:    dstDir,
			success: true,
		})
	}

	if len(linkResults) > 0 && verbose {
		hasSuccess := false
		for _, result := range linkResults {
			if result.success {
				hasSuccess = true
				break
			}
		}

		if hasSuccess {
			profileNote := styles.ProfileNoteFormat(selectedProfileName)
			fmt.Printf("%s%s\n", styles.InlineSuccessFormat("Linking", componentType, componentName), profileNote)

			for _, result := range linkResults {
				if result.success {
					fmt.Printf("%s\n", styles.IndentedDetailFormat(result.name, result.path))
				}
			}
		} else {
			fmt.Printf("%s\n", styles.InlineFailedFormat("Linking", componentType, componentName))

			for _, result := range linkResults {
				if !result.success {
					fmt.Printf("%s\n", styles.IndentedDetailFormat(result.name, result.errMsg))
				}
			}
			return fmt.Errorf("failed to link to target")
		}
	}

	if len(linkResults) > 0 {
		hasSuccess := false
		for _, result := range linkResults {
			if result.success {
				hasSuccess = true
				break
			}
		}
		if !hasSuccess {
			return fmt.Errorf("failed to link to any target")
		}
	}

	return nil
}

func (cl *ComponentLinker) loadComponentMetadata(componentType, componentName string) *models.ComponentEntry {
	entry, err := metadataPkg.LoadLockFileEntry(cl.agentsDir, componentType, componentName)
	if err != nil {
		return nil
	}
	return entry
}

func (cl *ComponentLinker) LinkComponentsByType(componentType string) error {
	typeDir := filepath.Join(cl.agentsDir, componentType)

	if _, err := os.Stat(typeDir); os.IsNotExist(err) {
		fmt.Printf("No %s found in %s\n", componentType, cl.agentsDir)
		return nil
	}

	entries, err := os.ReadDir(typeDir)
	if err != nil {
		return fmt.Errorf("failed to read %s directory: %w", componentType, err)
	}

	var successCount, failedCount, skippedCount int
	var failedComponents []string

	targetNames := make([]string, len(cl.targets))
	for i, target := range cl.targets {
		targetNames[i] = target.GetName()
	}
	targetList := strings.Join(targetNames, ", ")

	cl.formatter.EmptyLine()
	fmt.Printf("%s\n", colors.InfoBold(fmt.Sprintf("Linking %s to: %s", componentType, targetList)))
	cl.formatter.EmptyLine()

	for _, entry := range entries {
		if entry.IsDir() {
			componentName := entry.Name()

			if cl.isMonorepoContainer(componentType, componentName) {
				skippedCount++
				continue
			}

			fmt.Printf("Linking %s: %s... ", componentType, componentName)

			if err := cl.linkComponentInternal(componentType, componentName, false); err != nil {
				fmt.Printf("%s\n", colors.Error(formatter.SymbolError+" Failed"))
				fmt.Printf("  %s %v\n", colors.Muted("→"), err)
				failedCount++
				failedComponents = append(failedComponents, fmt.Sprintf("%s/%s", componentType, componentName))
			} else {
				fmt.Printf("%s\n", colors.Success(formatter.SymbolSuccess+" Done"))
				successCount++
			}
		}
	}

	cl.formatter.EmptyLine()

	table := formatter.NewBoxTable(cl.formatter.Writer(), []string{"Status", "Count"})
	table.AddRow([]string{colors.Success(formatter.SymbolSuccess + " Success"), fmt.Sprintf("%d", successCount)})
	if skippedCount > 0 {
		table.AddRow([]string{colors.Warning(formatter.SymbolWarning + " Skipped"), fmt.Sprintf("%d (monorepos)", skippedCount)})
	}
	if failedCount > 0 {
		table.AddRow([]string{colors.Error(formatter.SymbolError + " Failed"), fmt.Sprintf("%d", failedCount)})
	}
	table.Render()

	if failedCount > 0 {
		cl.formatter.EmptyLine()
		fmt.Println("Failed components:")
		for _, comp := range failedComponents {
			fmt.Printf("  • %s\n", comp)
		}
	}

	if skippedCount > 0 {
		cl.formatter.EmptyLine()
		fmt.Printf("%s Monorepo containers were skipped - their individual components should be linked separately\n",
			colors.Muted("Note:"))
	}

	return nil
}

func (cl *ComponentLinker) LinkAllComponents() error {
	componentTypes := paths.GetComponentTypes()

	var successCount, failedCount, skippedCount int
	var failedComponents []string

	targetNames := make([]string, len(cl.targets))
	for i, target := range cl.targets {
		targetNames[i] = target.GetName()
	}
	targetList := strings.Join(targetNames, ", ")

	cl.formatter.EmptyLine()
	fmt.Printf("%s\n", colors.InfoBold("Linking components to: "+targetList))
	cl.formatter.EmptyLine()

	for _, componentType := range componentTypes {
		typeDir := filepath.Join(cl.agentsDir, componentType)
		if _, err := os.Stat(typeDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(typeDir)
		if err != nil {
			cl.formatter.WarningMsg("Failed to read %s directory: %v", componentType, err)
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				componentName := entry.Name()

				// Skip monorepo containers - they shouldn't be linked as individual components
				if cl.isMonorepoContainer(componentType, componentName) {
					skippedCount++
					continue
				}

				fmt.Printf("Linking %s: %s... ", componentType, componentName)

				if err := cl.linkComponentInternal(componentType, componentName, false); err != nil {
					fmt.Printf("%s\n", colors.Error(formatter.SymbolError+" Failed"))
					fmt.Printf("  %s %v\n", colors.Muted("→"), err)
					failedCount++
					failedComponents = append(failedComponents, fmt.Sprintf("%s/%s", componentType, componentName))
				} else {
					fmt.Printf("%s\n", colors.Success(formatter.SymbolSuccess+" Done"))
					successCount++
				}
			}
		}
	}

	cl.formatter.EmptyLine()

	table := formatter.NewBoxTable(cl.formatter.Writer(), []string{"Status", "Count"})
	table.AddRow([]string{colors.Success(formatter.SymbolSuccess + " Success"), fmt.Sprintf("%d", successCount)})
	if skippedCount > 0 {
		table.AddRow([]string{colors.Warning(formatter.SymbolWarning + " Skipped"), fmt.Sprintf("%d (monorepos)", skippedCount)})
	}
	if failedCount > 0 {
		table.AddRow([]string{colors.Error(formatter.SymbolError + " Failed"), fmt.Sprintf("%d", failedCount)})
	}
	table.Render()

	if failedCount > 0 {
		cl.formatter.EmptyLine()
		fmt.Println("Failed components:")
		for _, comp := range failedComponents {
			fmt.Printf("  • %s\n", comp)
		}
	}

	if skippedCount > 0 {
		cl.formatter.EmptyLine()
		fmt.Printf("%s Monorepo containers were skipped - their individual components should be linked separately\n",
			colors.Muted("Note:"))
	}

	return nil
}

// isMonorepoContainer checks if a component directory contains other component directories
// and should not be linked as a single component.
func (cl *ComponentLinker) isMonorepoContainer(componentType, componentName string) bool {
	componentDir := filepath.Join(cl.agentsDir, componentType, componentName)

	entries, err := os.ReadDir(componentDir)
	if err != nil {
		return false
	}

	var markerFiles []string
	switch componentType {
	case "skills":
		markerFiles = []string{"SKILL.md"}
	case "agents":
		markerFiles = []string{componentName + ".md"}
	case "commands":
		markerFiles = []string{componentName + ".md"}
	default:
		return false
	}

	subComponentCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			subDir := filepath.Join(componentDir, entry.Name())
			for _, markerFile := range markerFiles {
				if _, err := os.Stat(filepath.Join(subDir, markerFile)); err == nil {
					subComponentCount++
					break
				}
			}
		}
	}

	return subComponentCount > 1
}

func (cl *ComponentLinker) LinkMonorepoComponents(componentType, repoName string) error {
	repoDir := filepath.Join(cl.agentsDir, componentType, repoName)

	entries, err := os.ReadDir(repoDir)
	if err != nil {
		return fmt.Errorf("failed to read monorepo directory: %w", err)
	}

	var markerFiles []string
	switch componentType {
	case "skills":
		markerFiles = []string{"SKILL.md"}
	case "agents":
		markerFiles = []string{}
	case "commands":
		markerFiles = []string{}
	default:
		return fmt.Errorf("unknown component type: %s", componentType)
	}

	linkedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			subComponentName := entry.Name()
			subComponentDir := filepath.Join(repoDir, subComponentName)

			hasMarker := false
			for _, markerFile := range markerFiles {
				if _, err := os.Stat(filepath.Join(subComponentDir, markerFile)); err == nil {
					hasMarker = true
					break
				}
			}

			if !hasMarker {
				if _, err := os.Stat(filepath.Join(subComponentDir, subComponentName+".md")); err == nil {
					hasMarker = true
				}
			}

			if hasMarker {
				linkName := fmt.Sprintf("%s-%s", repoName, subComponentName)

				srcDir := subComponentDir

				var successfulTargets []struct {
					name string
					path string
				}
				for _, target := range cl.targets {
					componentDir, err := target.GetGlobalComponentDir(componentType)
					if err != nil {
						fmt.Printf("Warning: failed to get target component directory for %s: %v\n", target.GetName(), err)
						continue
					}
					dstDir := filepath.Join(componentDir, linkName)

					if err := fileutil.CreateDirectoryWithPermissions(filepath.Dir(dstDir)); err != nil {
						fmt.Printf("Warning: failed to create destination directory for %s in %s: %v\n", linkName, target.GetName(), err)
						continue
					}

					if err := cl.createSymlink(srcDir, dstDir); err != nil {
						fmt.Printf("Warning: failed to link monorepo component %s to %s: %v\n", linkName, target.GetName(), err)
						continue
					}

					successfulTargets = append(successfulTargets, struct {
						name string
						path string
					}{
						name: target.GetName(),
						path: dstDir,
					})
				}

				if len(successfulTargets) > 0 {
					fmt.Printf("Successfully linked monorepo component '%s':\n", linkName)
					for _, t := range successfulTargets {
						fmt.Printf("  → %s: %s\n", t.name, t.path)
					}
					fmt.Printf("  Source: %s\n", srcDir)
					linkedCount++
				}
			}
		}
	}

	if linkedCount > 0 {
		fmt.Printf("Linked %d components from monorepo %s\n", linkedCount, repoName)
	}

	return nil
}

func (cl *ComponentLinker) DetectAndLinkLocalRepositories() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	if !cl.detector.IsLocalPath(cwd) {
		return fmt.Errorf("current directory is not a git repository")
	}

	components, err := cl.detector.DetectComponentsInRepo(cwd)
	if err != nil {
		return fmt.Errorf("failed to detect components in repository: %w", err)
	}

	if len(components) == 0 {
		fmt.Println("No components detected in current repository")
		return nil
	}

	fmt.Printf("Detected %d components in current repository:\n", len(components))
	for _, component := range components {
		fmt.Printf("  - %s: %s (%s)\n", component.Type, component.Name, component.Path)
	}

	// Link each detected component
	for _, component := range components {
		componentTypeStr := string(component.Type) + "s"
		componentPath := filepath.Join(cwd, component.Path)
		if info, err := os.Stat(componentPath); err == nil && !info.IsDir() {
			componentPath = filepath.Dir(componentPath)
		}

		tempLinkName := fmt.Sprintf("auto-detected-%s", component.Name)
		tempLinkPath := filepath.Join(cl.agentsDir, componentTypeStr, tempLinkName)

		if err := fileutil.CreateDirectoryWithPermissions(filepath.Dir(tempLinkPath)); err != nil {
			fmt.Printf("Warning: failed to create directory for %s: %v\n", component.Name, err)
			continue
		}

		if err := cl.createSymlink(componentPath, tempLinkPath); err != nil {
			fmt.Printf("Warning: failed to link component %s: %v\n", component.Name, err)
			continue
		}

		if err := cl.LinkComponent(componentTypeStr, tempLinkName); err != nil {
			fmt.Printf("Warning: failed to link %s to opencode: %v\n", component.Name, err)
			continue
		}

		fmt.Printf("✓ Automatically linked %s '%s' from current repository\n", component.Type, component.Name)
	}

	return nil
}

// ListLinkedComponents lists all components linked to the configured targets
func (cl *ComponentLinker) ListLinkedComponents() error {
	componentTypes := paths.GetComponentTypes()

	for _, target := range cl.targets {
		allLinks := make(map[string][]LinkStatus)
		totalCount := 0
		validCount := 0
		brokenCount := 0

		for _, componentType := range componentTypes {
			componentDir, err := target.GetGlobalComponentDir(componentType)
			if err != nil {
				return fmt.Errorf("failed to get target component directory: %w", err)
			}

			if _, err := os.Stat(componentDir); os.IsNotExist(err) {
				allLinks[componentType] = []LinkStatus{}
				continue
			}

			entries, err := os.ReadDir(componentDir)
			if err != nil {
				return fmt.Errorf("failed to read %s directory: %w", componentType, err)
			}

			links := []LinkStatus{}
			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), ".") {
					continue
				}

				fullPath := filepath.Join(componentDir, entry.Name())
				linkType, targetPath, valid := cl.analyzeLinkStatus(fullPath)

				profile := paths.BaseProfileName
				if targetPath != "" {
					profile = getProfileFromPath(targetPath)
				}

				status := LinkStatus{
					Name:       entry.Name(),
					Type:       componentType,
					LinkType:   linkType,
					Target:     targetPath,
					Valid:      valid,
					TargetPath: fullPath,
					Profile:    profile,
				}

				links = append(links, status)
				totalCount++

				if valid {
					validCount++
				} else if linkType == "broken" {
					brokenCount++
				}
			}

			allLinks[componentType] = links
		}

		targetName := target.GetName()
		targetDir, _ := target.GetGlobalBaseDir()

		fmt.Printf("\n=== %s ===\n", displayName(targetName))
		fmt.Printf("%s\n", cl.getSourceDescription())

		if totalCount == 0 {
			fmt.Printf("No components are currently linked to %s.\n", targetName)
			fmt.Printf("Link location: %s\n", targetDir)
			continue
		}

		for _, componentType := range componentTypes {
			links := allLinks[componentType]
			if len(links) == 0 {
				continue
			}

			displayType := strings.Title(componentType)
			fmt.Printf("\n%s (%d):\n", displayType, len(links))

			for _, link := range links {
				var symbol, statusMsg string

				switch link.LinkType {
				case "symlink":
					if link.Valid {
						symbol = "✓"
						statusMsg = fmt.Sprintf("→ %s", link.Target)
					} else {
						symbol = "✗"
						statusMsg = "[broken link]"
					}
				case "copied":
					symbol = "◆"
					statusMsg = "[copied directory]"
				case "broken":
					symbol = "✗"
					statusMsg = "[broken link]"
				case "missing":
					symbol = "?"
					statusMsg = "[unknown state]"
				default:
					symbol = "?"
					statusMsg = "[unknown]"
				}

				profileDisplay := fmt.Sprintf("(%s)", link.Profile)
				fmt.Printf("  %s %s %s %s\n", symbol, link.Name, profileDisplay, statusMsg)
			}
		}

		fmt.Printf("\nTotal: %d components", totalCount)
		if brokenCount > 0 {
			fmt.Printf(" (%d valid, %d broken)", validCount, brokenCount)
		}
		fmt.Println()
	}

	return nil
}

// ShowLinkStatus displays a matrix view of components and their status across all targets
func (cl *ComponentLinker) ShowLinkStatus(linkedOnly bool) error {
	componentTypes := paths.GetComponentTypes()

	// Collect all unique components from source directory
	type ComponentInfo struct {
		Name     string
		Type     string
		Profile  string
		BasePath string // Full path to the profile or base directory
	}
	allComponents := make([]ComponentInfo, 0)

	for _, componentType := range componentTypes {
		sourceDir := filepath.Join(cl.agentsDir, componentType)
		if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(sourceDir)
		if err != nil {
			return fmt.Errorf("failed to read %s directory: %w", componentType, err)
		}

		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			componentPath := filepath.Join(sourceDir, entry.Name())

			var profile string
			info, err := os.Lstat(componentPath)
			if err == nil && info.Mode()&os.ModeSymlink != 0 {
				profile = GetProfileNameFromSymlink(componentPath)
				if profile == "" {
					profile = getProfileFromPath(componentPath)
				}
			} else {
				profile = getProfileFromPath(componentPath)
			}

			allComponents = append(allComponents, ComponentInfo{
				Name:     entry.Name(),
				Type:     componentType,
				Profile:  profile,
				BasePath: cl.agentsDir,
			})
		}
	}

	if len(allComponents) == 0 {
		fmt.Fprintf(cl.formatter.Writer(), "No components found in ~/.agent-smith/\n")
		return nil
	}

	type ComponentStatus struct {
		Component ComponentInfo
		Targets   map[string]string // target name -> status symbol
	}

	statuses := make([]ComponentStatus, 0)

	for _, comp := range allComponents {
		status := ComponentStatus{
			Component: comp,
			Targets:   make(map[string]string),
		}

		for _, target := range cl.targets {
			componentDir, err := target.GetGlobalComponentDir(comp.Type)
			if err != nil {
				status.Targets[target.GetName()] = colors.Warning("?")
				continue
			}

			if comp.Type == "commands" || comp.Type == "agents" {
				componentTypeDir := filepath.Join(comp.BasePath, comp.Type)
				if isFlatMdLinked(comp.Name, componentTypeDir, componentDir) {
					status.Targets[target.GetName()] = colors.Success("✓")
				} else {
					status.Targets[target.GetName()] = colors.Muted("-")
				}
				continue
			}

			linkPath := filepath.Join(componentDir, comp.Name)

			if _, err := os.Lstat(linkPath); os.IsNotExist(err) {
				status.Targets[target.GetName()] = colors.Muted("-")
				continue
			}

			linkType, targetPath, valid := cl.analyzeLinkStatus(linkPath)

			expectedSource := filepath.Join(comp.BasePath, comp.Type, comp.Name)

			if linkType == "symlink" && valid {
				expectedSource, _ = filepath.EvalSymlinks(expectedSource)
				targetPath, _ = filepath.EvalSymlinks(targetPath)

				if expectedSource != targetPath {
					status.Targets[target.GetName()] = colors.Muted("-")
					continue
				}
			}

			var symbol string
			switch linkType {
			case "symlink":
				if valid {
					symbol = colors.Success("✓")
				} else {
					symbol = colors.Error("✗")
				}
			case "copied":
				symbol = colors.Success("◆")
			case "broken":
				symbol = colors.Error("✗")
			default:
				symbol = colors.Warning("?")
			}

			status.Targets[target.GetName()] = symbol
		}

		statuses = append(statuses, status)
	}

	cl.formatter.EmptyLine()
	cl.formatter.SectionHeader("Link Status Across All Targets")
	cl.formatter.InfoMsg("%s", cl.getSourceDescription())
	cl.formatter.EmptyLine()

	targetNames := make([]string, 0, len(cl.targets))
	for _, target := range cl.targets {
		targetNames = append(targetNames, target.GetName())
	}

	headers := []string{"Component", "Profile"}
	for _, targetName := range targetNames {
		headers = append(headers, displayName(targetName))
	}

	table := formatter.NewBoxTable(cl.formatter.Writer(), headers)

	byType := make(map[string][]ComponentStatus)
	for _, status := range statuses {
		byType[status.Component.Type] = append(byType[status.Component.Type], status)
	}

	for _, componentType := range componentTypes {
		components := byType[componentType]
		if len(components) == 0 {
			continue
		}

		sectionRow := []string{strings.Title(componentType) + ":", ""}
		for range targetNames {
			sectionRow = append(sectionRow, "")
		}
		table.AddRow(sectionRow)

		for _, status := range components {
			if linkedOnly {
				hasAnyLink := false
				for _, symbol := range status.Targets {
					if symbol != colors.Muted("-") {
						hasAnyLink = true
						break
					}
				}
				if !hasAnyLink {
					continue
				}
			}

			componentName := fmt.Sprintf("  %s", status.Component.Name)
			row := []string{componentName, status.Component.Profile}

			for _, targetName := range targetNames {
				symbol := status.Targets[targetName]
				row = append(row, symbol)
			}
			table.AddRow(row)
		}
	}

	table.Render()

	cl.formatter.EmptyLine()
	cl.formatter.SubsectionHeader("Legend")
	legendItems := []formatter.LegendItem{
		{Symbol: colors.Success("✓"), Description: colors.Success("Valid symlink")},
		{Symbol: colors.Success("◆"), Description: colors.Success("Copied directory")},
		{Symbol: colors.Error("✗"), Description: colors.Error("Broken link")},
		{Symbol: colors.Muted("-"), Description: colors.Muted("Not linked")},
		{Symbol: colors.Warning("?"), Description: colors.Warning("Unknown status")},
	}
	cl.formatter.DisplayLegendTable(legendItems)

	cl.formatter.EmptyLine()
	cl.formatter.SubsectionHeader("Summary")
	for _, targetName := range targetNames {
		linkedCount := 0
		for _, status := range statuses {
			symbol := status.Targets[targetName]
			if symbol == colors.Success("✓") || symbol == colors.Success("◆") {
				linkedCount++
			}
		}
		cl.formatter.ListItem("%s: %d/%d components linked", displayName(targetName), linkedCount, len(statuses))
	}

	return nil
}

// ShowAllProfilesLinkStatus displays link status for components across all profiles
// profileFilter can filter to specific profiles, or empty to show all
func (cl *ComponentLinker) ShowAllProfilesLinkStatus(profileFilter []string, linkedOnly bool) error {
	if cl.profileManager == nil {
		return fmt.Errorf("profile manager not available - this operation requires a profile manager")
	}

	componentTypes := paths.GetComponentTypes()

	// Collect all unique components from all profiles
	type ComponentInfo struct {
		Name     string
		Type     string
		Profile  string
		BasePath string // Full path to the profile or base directory
	}
	allComponents := make([]ComponentInfo, 0)

	baseDir, err := paths.GetAgentsDir()
	if err != nil {
		return fmt.Errorf("failed to get base directory: %w", err)
	}

	baseComponents := make([]ComponentInfo, 0)
	for _, componentType := range componentTypes {
		sourceDir := filepath.Join(baseDir, componentType)
		if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(sourceDir)
		if err != nil {
			continue // Skip directories we can't read
		}

		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			if !entry.IsDir() {
				continue
			}

			componentPath := filepath.Join(sourceDir, entry.Name())

			var profile string
			info, err := os.Lstat(componentPath)
			if err == nil && info.Mode()&os.ModeSymlink != 0 {
				profile = GetProfileNameFromSymlink(componentPath)
				if profile == "" {
					profile = paths.BaseProfileName
				}
			} else {
				profile = paths.BaseProfileName
			}

			baseComponents = append(baseComponents, ComponentInfo{
				Name:     entry.Name(),
				Type:     componentType,
				Profile:  profile,
				BasePath: baseDir,
			})
		}
	}

	profiles, err := cl.profileManager.ScanProfiles()
	if err != nil {
		return fmt.Errorf("failed to scan profiles: %w", err)
	}

	var filteredProfiles []*Profile
	if len(profileFilter) > 0 {
		filterMap := make(map[string]bool)
		for _, name := range profileFilter {
			filterMap[name] = true
		}

		profileMap := make(map[string]bool)
		for _, p := range profiles {
			profileMap[p.Name] = true
		}

		for _, filterName := range profileFilter {
			if !profileMap[filterName] {
				return fmt.Errorf("profile '%s' does not exist", filterName)
			}
		}

		for _, p := range profiles {
			if filterMap[p.Name] {
				filteredProfiles = append(filteredProfiles, p)
			}
		}
	} else {
		filteredProfiles = profiles
	}

	profileComponents := make([]ComponentInfo, 0)
	for _, profile := range filteredProfiles {
		for _, componentType := range componentTypes {
			var sourceDir string
			switch componentType {
			case "agents":
				if !profile.HasAgents {
					continue
				}
				sourceDir = filepath.Join(profile.BasePath, "agents")
			case "skills":
				if !profile.HasSkills {
					continue
				}
				sourceDir = filepath.Join(profile.BasePath, "skills")
			case "commands":
				if !profile.HasCommands {
					continue
				}
				sourceDir = filepath.Join(profile.BasePath, "commands")
			default:
				continue
			}

			if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
				continue
			}

			entries, err := os.ReadDir(sourceDir)
			if err != nil {
				continue // Skip directories we can't read
			}
			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), ".") {
					continue
				}
				if !entry.IsDir() {
					continue
				}
				profileComponents = append(profileComponents, ComponentInfo{
					Name:     entry.Name(),
					Type:     componentType,
					Profile:  profile.Name,
					BasePath: profile.BasePath,
				})
			}
		}
	}

	if len(profileFilter) == 0 {
		allComponents = append(allComponents, baseComponents...)
	}
	allComponents = append(allComponents, profileComponents...)

	if len(allComponents) == 0 {
		if len(profileFilter) > 0 {
			fmt.Fprintf(cl.formatter.Writer(), "No components found in the specified profiles\n")
		} else {
			fmt.Fprintf(cl.formatter.Writer(), "No components found\n")
		}
		return nil
	}

	type ComponentStatus struct {
		Component ComponentInfo
		Targets   map[string]string // target name -> status symbol
	}

	statuses := make([]ComponentStatus, 0)

	for _, comp := range allComponents {
		status := ComponentStatus{
			Component: comp,
			Targets:   make(map[string]string),
		}

		for _, target := range cl.targets {
			componentDir, err := target.GetGlobalComponentDir(comp.Type)
			if err != nil {
				status.Targets[target.GetName()] = colors.Warning("?")
				continue
			}

			if comp.Type == "commands" || comp.Type == "agents" {
				componentTypeDir := filepath.Join(comp.BasePath, comp.Type)
				if isFlatMdLinked(comp.Name, componentTypeDir, componentDir) {
					status.Targets[target.GetName()] = colors.Success("✓")
				} else {
					status.Targets[target.GetName()] = colors.Muted("-")
				}
				continue
			}

			linkPath := filepath.Join(componentDir, comp.Name)

			if _, err := os.Lstat(linkPath); os.IsNotExist(err) {
				status.Targets[target.GetName()] = colors.Muted("-")
				continue
			}

			linkType, linkTarget, valid := cl.analyzeLinkStatus(linkPath)

			if linkType == "symlink" && valid {
				expectedSource := filepath.Join(comp.BasePath, comp.Type, comp.Name)

				absTarget, err1 := filepath.Abs(linkTarget)
				absExpected, err2 := filepath.Abs(expectedSource)

				if err1 == nil && err2 == nil {
					absTarget = filepath.Clean(absTarget)
					absExpected = filepath.Clean(absExpected)

					if absTarget != absExpected {
						status.Targets[target.GetName()] = colors.Muted("-")
						continue
					}
				}
			}

			var symbol string
			switch linkType {
			case "symlink":
				if valid {
					symbol = colors.Success("✓")
				} else {
					symbol = colors.Error("✗")
				}
			case "copied":
				symbol = colors.Success("◆")
			case "broken":
				symbol = colors.Error("✗")
			default:
				symbol = colors.Warning("?")
			}

			status.Targets[target.GetName()] = symbol
		}

		statuses = append(statuses, status)
	}

	cl.formatter.EmptyLine()
	cl.formatter.SectionHeader("Link Status Across All Profiles")
	cl.formatter.EmptyLine()

	targetNames := make([]string, 0, len(cl.targets))
	for _, target := range cl.targets {
		targetNames = append(targetNames, target.GetName())
	}

	headers := []string{"Component", "Type", "Profile"}
	for _, targetName := range targetNames {
		headers = append(headers, displayName(targetName))
	}

	table := formatter.NewBoxTable(cl.formatter.Writer(), headers)

	byType := make(map[string][]ComponentStatus)
	for _, status := range statuses {
		byType[status.Component.Type] = append(byType[status.Component.Type], status)
	}

	for _, componentType := range componentTypes {
		components := byType[componentType]
		if len(components) == 0 {
			continue
		}

		sectionRow := []string{strings.Title(componentType) + ":", "", ""}
		for range targetNames {
			sectionRow = append(sectionRow, "")
		}
		table.AddRow(sectionRow)

		for _, status := range components {
			if linkedOnly {
				hasAnyLink := false
				for _, targetName := range targetNames {
					symbol := status.Targets[targetName]
					if symbol != colors.Muted("-") {
						hasAnyLink = true
						break
					}
				}
				if !hasAnyLink {
					continue
				}
			}

			componentName := fmt.Sprintf("  %s", status.Component.Name)
			row := []string{componentName, status.Component.Type, status.Component.Profile}

			for _, targetName := range targetNames {
				symbol := status.Targets[targetName]
				row = append(row, symbol)
			}
			table.AddRow(row)
		}
	}

	// Render the table
	table.Render()

	cl.formatter.EmptyLine()
	cl.formatter.SubsectionHeader("Legend")
	legendItems := []formatter.LegendItem{
		{Symbol: colors.Success("✓"), Description: colors.Success("Valid symlink")},
		{Symbol: colors.Success("◆"), Description: colors.Success("Copied directory")},
		{Symbol: colors.Error("✗"), Description: colors.Error("Broken link")},
		{Symbol: colors.Muted("-"), Description: colors.Muted("Not linked")},
		{Symbol: colors.Warning("?"), Description: colors.Warning("Unknown status")},
	}
	cl.formatter.DisplayLegendTable(legendItems)

	cl.formatter.EmptyLine()
	cl.formatter.SubsectionHeader("Summary")

	profileCount := len(filteredProfiles)
	if len(profileFilter) == 0 {
		profileCount++ // Include base
	}
	profileCountStr := fmt.Sprintf("%d", profileCount)
	if len(profileFilter) == 0 {
		if len(filteredProfiles) == 0 {
			profileCountStr = fmt.Sprintf("1 (%s only)", paths.BaseProfileName)
		} else {
			profileCountStr = fmt.Sprintf("%d (%s + %d custom)", profileCount, paths.BaseProfileName, len(filteredProfiles))
		}
	}
	cl.formatter.ListItem("Profiles scanned: %s", profileCountStr)
	cl.formatter.ListItem("Total components: %d", len(statuses))

	for _, targetName := range targetNames {
		linkedCount := 0
		for _, status := range statuses {
			symbol := status.Targets[targetName]
			if symbol == colors.Success("✓") || symbol == colors.Success("◆") {
				linkedCount++
			}
		}
		percentage := 0
		if len(statuses) > 0 {
			percentage = (linkedCount * 100) / len(statuses)
		}
		cl.formatter.ListItem("%s: %d/%d linked (%d%%)", displayName(targetName), linkedCount, len(statuses), percentage)
	}

	activeProfile, err := cl.profileManager.GetActiveProfile()
	if err == nil && activeProfile != "" {
		cl.formatter.EmptyLine()
		cl.formatter.InfoMsg("Active Profile: %s", activeProfile)
	}

	return nil
}

// UnlinkComponent removes a linked component from configured targets
// targetFilter can be:
//   - "" (empty): unlink from all targets
//   - "all": unlink from all targets
//   - specific target name (e.g., "opencode", "claudecode"): unlink from only that target
func (cl *ComponentLinker) UnlinkComponent(componentType, componentName, targetFilter string) error {
	if componentType != "skills" && componentType != "agents" && componentType != "commands" {
		return fmt.Errorf("invalid component type: %s (must be skills, agents, or commands)", componentType)
	}

	targetsToUnlink := cl.filterTargets(targetFilter)
	if len(targetsToUnlink) == 0 {
		if targetFilter != "" && targetFilter != "all" {
			availableTargets := make([]string, 0, len(cl.targets))
			for _, target := range cl.targets {
				availableTargets = append(availableTargets, target.GetName())
			}

			if len(availableTargets) == 0 {
				return fmt.Errorf("target '%s' does not exist and no targets are configured", targetFilter)
			}

			return fmt.Errorf("target '%s' does not exist\n\nAvailable targets:\n  - %s\n\nExample:\n  agent-smith unlink %s %s --target %s",
				targetFilter,
				strings.Join(availableTargets, "\n  - "),
				componentType,
				componentName,
				availableTargets[0])
		}
		return fmt.Errorf("no targets available")
	}

	successCount := 0
	failedCount := 0
	var errors []string
	var unlinkedTargets []string

	if targetFilter != "" && targetFilter != "all" {
		cl.formatter.SectionHeader(fmt.Sprintf("Unlinking %s '%s' from: %s", componentType, componentName, targetFilter))
	} else {
		targetNames := make([]string, 0, len(targetsToUnlink))
		for _, target := range targetsToUnlink {
			targetNames = append(targetNames, target.GetName())
		}
		cl.formatter.SectionHeader(fmt.Sprintf("Unlinking %s '%s' from: %s", componentType, componentName, strings.Join(targetNames, ", ")))
	}

	for _, target := range targetsToUnlink {
		componentDir, err := target.GetGlobalComponentDir(componentType)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to get target component directory for %s: %v", target.GetName(), err))
			failedCount++
			continue
		}

		targetName := target.GetName()

		if componentType == "commands" || componentType == "agents" {
			srcComponentTypeDir := filepath.Join(cl.agentsDir, componentType)
			if !isFlatMdLinked(componentName, srcComponentTypeDir, componentDir) {
				continue
			}

			cl.formatter.ProgressMsg(fmt.Sprintf("Unlinking from %s", targetName), componentName)

			if err := unlinkFlatMdFiles(componentName, srcComponentTypeDir, componentDir); err != nil {
				cl.formatter.ProgressFailed()
				errors = append(errors, fmt.Sprintf("failed to unlink from %s: %v", targetName, err))
				failedCount++
				continue
			}

			cl.formatter.ProgressComplete()
			unlinkedTargets = append(unlinkedTargets, targetName)
			successCount++
			continue
		}

		linkPath := filepath.Join(componentDir, componentName)

		if _, err := os.Lstat(linkPath); os.IsNotExist(err) {
			continue
		}

		linkType, targetPath, _ := cl.analyzeLinkStatus(linkPath)

		if linkType == "copied" {
			cl.formatter.WarningMsg("'%s' is a copied directory in %s, not a symlink", componentName, targetName)
			fmt.Printf("This will permanently delete: %s\n", linkPath)
			fmt.Print("Continue? [y/N]: ")

			var response string
			fmt.Scanln(&response)

			if strings.ToLower(strings.TrimSpace(response)) != "y" {
				cl.formatter.InfoMsg("Unlink cancelled for %s", targetName)
				continue
			}
		}

		cl.formatter.ProgressMsg(fmt.Sprintf("Unlinking from %s", targetName), componentName)

		if linkType == "copied" {
			if err := os.RemoveAll(linkPath); err != nil {
				cl.formatter.ProgressFailed()
				errors = append(errors, fmt.Sprintf("failed to remove copied directory from %s: %v", targetName, err))
				failedCount++
				continue
			}
		} else {
			if err := os.Remove(linkPath); err != nil {
				cl.formatter.ProgressFailed()
				errors = append(errors, fmt.Sprintf("failed to remove link from %s: %v", targetName, err))
				failedCount++
				continue
			}
		}

		cl.formatter.ProgressComplete()
		unlinkedTargets = append(unlinkedTargets, targetName)

		if linkType == "symlink" && targetPath != "" {
			cl.formatter.DetailItem("Source", targetPath)
		}
		successCount++
	}

	if len(errors) > 0 {
		for _, errMsg := range errors {
			cl.formatter.WarningMsg("%s", errMsg)
		}
		if successCount == 0 {
			return fmt.Errorf("failed to unlink from any target")
		}
	}

	if successCount == 0 {
		return fmt.Errorf("component %s/%s is not linked to any target", componentType, componentName)
	}

	cl.formatter.EmptyLine()
	cl.formatter.CounterSummary(successCount+failedCount, successCount, failedCount, 0)

	return nil
}

// UnlinkComponentsByType removes all linked components of a specific type from configured targets
// targetFilter can be:
//   - "" (empty): unlink from all targets
//   - "all": unlink from all targets
//   - specific target name (e.g., "opencode", "claudecode"): unlink from only that target
func (cl *ComponentLinker) UnlinkComponentsByType(componentType, targetFilter string, force bool) error {
	targetsToUnlink := cl.filterTargets(targetFilter)
	if len(targetsToUnlink) == 0 {
		if targetFilter != "" && targetFilter != "all" {
			availableTargets := make([]string, 0, len(cl.targets))
			for _, target := range cl.targets {
				availableTargets = append(availableTargets, target.GetName())
			}

			if len(availableTargets) == 0 {
				return fmt.Errorf("target '%s' does not exist and no targets are configured", targetFilter)
			}

			return fmt.Errorf("target '%s' does not exist\n\nAvailable targets:\n  - %s\n\nExample:\n  agent-smith unlink %s --target %s",
				targetFilter,
				strings.Join(availableTargets, "\n  - "),
				componentType,
				availableTargets[0])
		}
		return fmt.Errorf("no targets available")
	}

	totalLinks := 0
	copiedDirs := 0

	for _, target := range targetsToUnlink {
		componentDir, err := target.GetGlobalComponentDir(componentType)
		if err != nil {
			return fmt.Errorf("failed to get target component directory: %w", err)
		}

		if _, err := os.Stat(componentDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(componentDir)
		if err != nil {
			return fmt.Errorf("failed to read %s directory: %w", componentType, err)
		}

		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			fullPath := filepath.Join(componentDir, entry.Name())
			linkType, _, _ := cl.analyzeLinkStatus(fullPath)

			if linkType == "copied" {
				copiedDirs++
				continue
			}
			totalLinks++
		}
	}

	if totalLinks == 0 && copiedDirs == 0 {
		cl.formatter.InfoMsg("No linked %s found", componentType)
		return nil
	}

	if !force {
		if totalLinks > 0 {
			targetNames := make([]string, 0, len(targetsToUnlink))
			for _, target := range targetsToUnlink {
				targetNames = append(targetNames, target.GetName())
			}
			targetStr := strings.Join(targetNames, ", ")

			if targetFilter != "" && targetFilter != "all" {
				fmt.Printf("This will unlink %d %s from: %s", totalLinks, componentType, targetStr)
			} else {
				fmt.Printf("This will unlink %d %s from all targets", totalLinks, componentType)
			}
			fmt.Println()
		}
		if copiedDirs > 0 {
			cl.formatter.InfoMsg("Note: %d copied directories will be skipped (not deleted)", copiedDirs)
		}
		if totalLinks == 0 {
			cl.formatter.InfoMsg("No symlinked %s to unlink (only copied directories found)", componentType)
			return nil
		}
		fmt.Print("Continue? [y/N]: ")

		var response string
		fmt.Scanln(&response)

		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("Unlink cancelled.")
			return nil
		}
	}

	if targetFilter != "" && targetFilter != "all" {
		cl.formatter.SectionHeader(fmt.Sprintf("Unlinking all %s from: %s", componentType, targetFilter))
	} else {
		targetNames := make([]string, 0, len(targetsToUnlink))
		for _, target := range targetsToUnlink {
			targetNames = append(targetNames, target.GetName())
		}
		cl.formatter.SectionHeader(fmt.Sprintf("Unlinking all %s from: %s", componentType, strings.Join(targetNames, ", ")))
	}

	removedCount := 0
	skippedCount := 0
	errorCount := 0

	for _, target := range targetsToUnlink {
		componentDir, err := target.GetGlobalComponentDir(componentType)
		if err != nil {
			return fmt.Errorf("failed to get target component directory: %w", err)
		}

		if _, err := os.Stat(componentDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(componentDir)
		if err != nil {
			cl.formatter.WarningMsg("Failed to read %s directory for %s: %v", componentType, target.GetName(), err)
			continue
		}

		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			fullPath := filepath.Join(componentDir, entry.Name())
			linkType, _, _ := cl.analyzeLinkStatus(fullPath)

			if linkType == "copied" {
				skippedCount++
				continue
			}

			cl.formatter.ProgressMsg(fmt.Sprintf("Unlinking %s from %s", componentType, target.GetName()), entry.Name())

			err := os.Remove(fullPath)

			if err != nil {
				cl.formatter.ProgressFailed()
				cl.formatter.WarningMsg("Failed to unlink %s/%s from %s: %v", componentType, entry.Name(), target.GetName(), err)
				errorCount++
			} else {
				cl.formatter.ProgressComplete()
				removedCount++
			}
		}
	}

	cl.formatter.EmptyLine()
	cl.formatter.CounterSummary(removedCount+errorCount, removedCount, errorCount, skippedCount)

	return nil
}

// isSymlinkFromCurrentProfile checks if a symlink belongs to the current profile
// by comparing the symlink target path with the ComponentLinker's agentsDir
func (cl *ComponentLinker) isSymlinkFromCurrentProfile(symlinkPath string) (bool, error) {
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		return false, err
	}

	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(symlinkPath), target)
	}

	target = filepath.Clean(target)

	// This correctly distinguishes base installation from profile installations
	// because profile directories are physically inside ~/.agent-smith/ but
	// getProfileFromPath() extracts the actual profile name from the path
	currentProfile := getProfileFromPath(cl.agentsDir)
	targetProfile := getProfileFromPath(target)
	return currentProfile == targetProfile, nil
}

// isSymlinkFromAgentSmith checks if a symlink points to any agent-smith directory
// (base installation, any profile, etc). Returns false for manually-created external symlinks.
func (cl *ComponentLinker) isSymlinkFromAgentSmith(symlinkPath string) (bool, error) {
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		return false, err
	}

	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(symlinkPath), target)
	}

	target = filepath.Clean(target)

	// This is important for tests and for the current profile/base installation
	agentsDir := filepath.Clean(cl.agentsDir)
	if strings.HasPrefix(target, agentsDir) {
		return true, nil
	}

	baseAgentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return false, nil
	}

	if strings.HasPrefix(target, baseAgentsDir) {
		return true, nil
	}

	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		return false, nil
	}

	if strings.HasPrefix(target, profilesDir) {
		return true, nil
	}

	return false, nil
}

// anyProfilesExist checks if any profiles directory exists and contains profiles
func (cl *ComponentLinker) anyProfilesExist() bool {
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		return false
	}

	info, err := os.Stat(profilesDir)
	if err != nil || !info.IsDir() {
		return false
	}

	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			return true
		}
	}

	return false
}

// UnlinkAllComponents removes all linked components from configured targets
// targetFilter can be:
//   - "" (empty): unlink from all targets
//   - "all": unlink from all targets
//   - specific target name (e.g., "opencode", "claudecode"): unlink from only that target
//
// allProfiles: if true, unlinks components from all profiles; if false, only from current profile
func (cl *ComponentLinker) UnlinkAllComponents(targetFilter string, force bool, allProfiles bool) error {
	targetsToUnlink := cl.filterTargets(targetFilter)
	if len(targetsToUnlink) == 0 {
		if targetFilter != "" && targetFilter != "all" {
			availableTargets := make([]string, 0, len(cl.targets))
			for _, target := range cl.targets {
				availableTargets = append(availableTargets, target.GetName())
			}

			if len(availableTargets) == 0 {
				return fmt.Errorf("target '%s' does not exist and no targets are configured", targetFilter)
			}

			return fmt.Errorf("target '%s' does not exist\n\nAvailable targets:\n  - %s\n\nExample:\n  agent-smith unlink all --target %s",
				targetFilter,
				strings.Join(availableTargets, "\n  - "),
				availableTargets[0])
		}
		return fmt.Errorf("no targets available")
	}

	componentTypes := paths.GetComponentTypes()

	totalLinks := 0
	copiedDirs := 0
	skippedProfilesCount := 0
	skippedProfilesMap := make(map[string]int)

	for _, target := range targetsToUnlink {
		for _, componentType := range componentTypes {
			componentDir, err := target.GetGlobalComponentDir(componentType)
			if err != nil {
				return fmt.Errorf("failed to get target component directory: %w", err)
			}

			if _, err := os.Stat(componentDir); os.IsNotExist(err) {
				continue
			}

			entries, err := os.ReadDir(componentDir)
			if err != nil {
				return fmt.Errorf("failed to read %s directory: %w", componentType, err)
			}

			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), ".") {
					continue
				}

				fullPath := filepath.Join(componentDir, entry.Name())
				linkType, _, _ := cl.analyzeLinkStatus(fullPath)

				if linkType == "copied" {
					copiedDirs++
					continue
				}

				if linkType == "symlink" || linkType == "broken" {
					fromAgentSmith, err := cl.isSymlinkFromAgentSmith(fullPath)
					if err == nil && !fromAgentSmith {
						skippedProfilesCount++
						continue
					}
				}

				if !allProfiles && (linkType == "symlink" || linkType == "broken") {
					belongsToProfile, err := cl.isSymlinkFromCurrentProfile(fullPath)
					if err == nil && !belongsToProfile {
						profileName := GetProfileNameFromSymlink(fullPath)
						if profileName != "" {
							skippedProfilesMap[profileName]++
						}
						skippedProfilesCount++
						continue
					}
				}

				totalLinks++
			}
		}
	}

	if totalLinks == 0 && copiedDirs == 0 && skippedProfilesCount == 0 {
		cl.formatter.InfoMsg("No linked components found")
		return nil
	}

	currentProfileName := getProfileFromPath(cl.agentsDir)
	profilesExist := cl.anyProfilesExist()

	if !force {
		if totalLinks > 0 {
			targetStr := "all targets"
			if targetFilter != "" && targetFilter != "all" {
				targetStr = targetFilter
			} else if len(targetsToUnlink) > 0 {
				targetNames := make([]string, 0, len(targetsToUnlink))
				for _, target := range targetsToUnlink {
					targetNames = append(targetNames, target.GetName())
				}
				targetStr = strings.Join(targetNames, ", ")
			}

			profileMsg := ""
			if allProfiles {
				profileMsg = " from all profiles"
			} else if profilesExist {
				if currentProfileName == paths.BaseProfileName {
					profileMsg = " from base installation"
				} else {
					profileMsg = fmt.Sprintf(" from profile '%s'", currentProfileName)
				}
			}
			fmt.Printf("This will unlink %d symlinked components%s from: %s", totalLinks, profileMsg, targetStr)
			fmt.Println()
		}
		if copiedDirs > 0 {
			cl.formatter.InfoMsg("Note: %d copied directories will be skipped (not deleted)", copiedDirs)
		}
		if skippedProfilesCount > 0 && !allProfiles {
			profileNames := make([]string, 0, len(skippedProfilesMap))
			for profileName := range skippedProfilesMap {
				profileNames = append(profileNames, profileName)
			}
			if len(profileNames) == 1 {
				cl.formatter.InfoMsg("Note: %d components from profile '%s' will be skipped", skippedProfilesCount, profileNames[0])
			} else if len(profileNames) > 1 {
				cl.formatter.InfoMsg("Note: %d components from other profiles will be skipped:", skippedProfilesCount)
				for _, profileName := range profileNames {
					count := skippedProfilesMap[profileName]
					fmt.Printf("  - %s: %d components\n", profileName, count)
				}
			}
		}
		if totalLinks == 0 {
			if skippedProfilesCount > 0 {
				if currentProfileName == paths.BaseProfileName {
					cl.formatter.InfoMsg("No symlinked components from base installation to unlink (found %d from profiles)", skippedProfilesCount)
				} else {
					cl.formatter.InfoMsg("No symlinked components from profile '%s' to unlink (found %d from other profiles)", currentProfileName, skippedProfilesCount)
				}
			} else {
				cl.formatter.InfoMsg("No symlinked components to unlink (only copied directories found)")
			}
			return nil
		}
		fmt.Print("Continue? [y/N]: ")

		var response string
		fmt.Scanln(&response)

		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("Unlink cancelled.")
			return nil
		}
	}

	headerMsg := ""
	if targetFilter != "" && targetFilter != "all" {
		if allProfiles {
			headerMsg = fmt.Sprintf("Unlinking all components (all profiles) from: %s", targetFilter)
		} else if profilesExist {
			if currentProfileName == paths.BaseProfileName {
				headerMsg = fmt.Sprintf("Unlinking components (base installation) from: %s", targetFilter)
			} else {
				headerMsg = fmt.Sprintf("Unlinking components (profile '%s') from: %s", currentProfileName, targetFilter)
			}
		} else {
			headerMsg = fmt.Sprintf("Unlinking components from: %s", targetFilter)
		}
	} else {
		targetNames := make([]string, 0, len(targetsToUnlink))
		for _, target := range targetsToUnlink {
			targetNames = append(targetNames, target.GetName())
		}
		targetList := strings.Join(targetNames, ", ")
		if allProfiles {
			headerMsg = fmt.Sprintf("Unlinking all components (all profiles) from: %s", targetList)
		} else if profilesExist {
			if currentProfileName == paths.BaseProfileName {
				headerMsg = fmt.Sprintf("Unlinking components (base installation) from: %s", targetList)
			} else {
				headerMsg = fmt.Sprintf("Unlinking components (profile '%s') from: %s", currentProfileName, targetList)
			}
		} else {
			headerMsg = fmt.Sprintf("Unlinking components from: %s", targetList)
		}
	}
	cl.formatter.SectionHeader(headerMsg)

	removedCount := 0
	skippedCount := 0
	skippedByProfile := make(map[string][]string)
	errorCount := 0

	for _, target := range targetsToUnlink {
		for _, componentType := range componentTypes {
			componentDir, err := target.GetGlobalComponentDir(componentType)
			if err != nil {
				return fmt.Errorf("failed to get target component directory: %w", err)
			}

			if _, err := os.Stat(componentDir); os.IsNotExist(err) {
				continue
			}

			entries, err := os.ReadDir(componentDir)
			if err != nil {
				cl.formatter.WarningMsg("Failed to read %s directory for %s: %v", componentType, target.GetName(), err)
				continue
			}

			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), ".") {
					continue
				}

				fullPath := filepath.Join(componentDir, entry.Name())
				linkType, _, _ := cl.analyzeLinkStatus(fullPath)

				if linkType == "copied" {
					skippedCount++
					continue
				}

				if linkType == "symlink" || linkType == "broken" {
					fromAgentSmith, err := cl.isSymlinkFromAgentSmith(fullPath)
					if err == nil && !fromAgentSmith {
						skippedCount++
						continue
					}
				}

				if !allProfiles && (linkType == "symlink" || linkType == "broken") {
					belongsToProfile, err := cl.isSymlinkFromCurrentProfile(fullPath)
					if err == nil && !belongsToProfile {
						profileName := GetProfileNameFromSymlink(fullPath)
						if profileName != "" {
							skippedByProfile[profileName] = append(skippedByProfile[profileName], fmt.Sprintf("%s/%s", componentType, entry.Name()))
						}
						skippedCount++
						continue
					}
				}

				profileNote := ""
				if profilesExist && (linkType == "symlink" || linkType == "broken") {
					profileName := GetProfileNameFromSymlink(fullPath)
					if profileName != "" && profileName != paths.BaseProfileName {
						profileNote = fmt.Sprintf(" [%s]", profileName)
					}
				}

				cl.formatter.ProgressMsg(fmt.Sprintf("Unlinking %s from %s%s", componentType, target.GetName(), profileNote), entry.Name())

				var err error
				err = os.Remove(fullPath)

				if err != nil {
					cl.formatter.ProgressFailed()
					cl.formatter.WarningMsg("Failed to unlink %s/%s from %s: %v", componentType, entry.Name(), target.GetName(), err)
					errorCount++
				} else {
					cl.formatter.ProgressComplete()
					removedCount++
				}
			}
		}
	}

	cl.formatter.EmptyLine()
	cl.formatter.CounterSummary(removedCount+errorCount, removedCount, errorCount, skippedCount)

	if len(skippedByProfile) > 0 && !allProfiles {
		cl.formatter.EmptyLine()
		fmt.Println("Skipped components from other profiles:")
		for profileName, components := range skippedByProfile {
			fmt.Printf("  Profile '%s' (%d components):\n", profileName, len(components))
			for _, comp := range components {
				fmt.Printf("    - %s\n", comp)
			}
		}
		cl.formatter.EmptyLine()
		if currentProfileName == paths.BaseProfileName {
			cl.formatter.InfoMsg("Use --all-profiles flag to unlink components from all profiles")
		} else {
			cl.formatter.InfoMsg("Use --all-profiles flag to unlink components from all profiles, or switch to the respective profile")
		}
	}

	return nil
}

// ProfileMatch represents a profile that contains a specific component
type ProfileMatch struct {
	ProfileName string
	ProfilePath string
	IsActive    bool
	SourceUrl   string
}

// searchComponentInProfiles searches for a component across all profiles
// Returns a list of profiles that contain the specified component
func (cl *ComponentLinker) searchComponentInProfiles(componentType, componentName string) ([]ProfileMatch, error) {
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get profiles directory: %w", err)
	}

	if _, err := os.Stat(profilesDir); os.IsNotExist(err) {
		return []ProfileMatch{}, nil
	}

	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}
	activeProfilePath := filepath.Join(agentsDir, ".active-profile")
	activeProfileData, _ := os.ReadFile(activeProfilePath)
	activeProfile := strings.TrimSpace(string(activeProfileData))

	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read profiles directory: %w", err)
	}

	var matches []ProfileMatch
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		profileName := entry.Name()
		profilePath := filepath.Join(profilesDir, profileName)
		componentPath := filepath.Join(profilePath, componentType, componentName)

		if _, err := os.Stat(componentPath); err == nil {
			sourceUrl := ""
			lockEntry, err := metadataPkg.LoadLockFileEntry(profilePath, componentType, componentName)
			if err == nil && lockEntry != nil {
				sourceUrl = lockEntry.SourceUrl
			}

			matches = append(matches, ProfileMatch{
				ProfileName: profileName,
				ProfilePath: profilePath,
				IsActive:    profileName == activeProfile,
				SourceUrl:   sourceUrl,
			})
		}
	}

	return matches, nil
}

// promptProfileSelection displays an interactive prompt for the user to select a profile
// Returns the selected profile path and name, or error if cancelled
func (cl *ComponentLinker) promptProfileSelection(componentType, componentName string, matches []ProfileMatch) (string, string, error) {
	if len(matches) == 0 {
		return "", "", fmt.Errorf("no profiles contain component %s", componentName)
	}

	fmt.Printf("\n⚠️  Component \"%s\" found in multiple profiles:\n\n", componentName)

	for i, match := range matches {
		activeIndicator := ""
		if match.IsActive {
			activeIndicator = " (active)"
		}
		fmt.Printf("  %d. %s%s\n", i+1, match.ProfileName, activeIndicator)

		if match.SourceUrl != "" {
			fmt.Printf("     Source: %s\n", match.SourceUrl)
		}

		if i < len(matches)-1 {
			fmt.Println()
		}
	}

	fmt.Printf("\nSelect profile to link from [1-%d] (or 'c' to cancel): ", len(matches))

	var response string
	fmt.Scanln(&response)
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "c" || response == "" {
		return "", "", fmt.Errorf("profile selection cancelled")
	}

	var selection int
	_, err := fmt.Sscanf(response, "%d", &selection)
	if err != nil || selection < 1 || selection > len(matches) {
		return "", "", fmt.Errorf("invalid selection: %s", response)
	}

	selectedMatch := matches[selection-1]
	return selectedMatch.ProfilePath, selectedMatch.ProfileName, nil
}
