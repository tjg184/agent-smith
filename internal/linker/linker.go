package linker

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tgaines/agent-smith/internal/detector"
	"github.com/tgaines/agent-smith/internal/fileutil"
	"github.com/tgaines/agent-smith/internal/formatter"
	metadataPkg "github.com/tgaines/agent-smith/internal/metadata"
	"github.com/tgaines/agent-smith/internal/models"
	"github.com/tgaines/agent-smith/pkg/colors"
	"github.com/tgaines/agent-smith/pkg/config"
	"github.com/tgaines/agent-smith/pkg/paths"
	"github.com/tgaines/agent-smith/pkg/styles"
)

// ComponentLinker handles linking components to configured targets
type ComponentLinker struct {
	agentsDir      string
	targets        []config.Target
	detector       *detector.RepositoryDetector
	profileManager ProfileManager // Optional - can be nil
	formatter      *formatter.Formatter
}

// ProfileManager is an interface for profile scanning operations
// This interface prevents circular dependencies between linker and profiles packages
type ProfileManager interface {
	ScanProfiles() ([]*Profile, error)
	GetActiveProfile() (string, error)
}

// Profile represents a user profile (minimal interface to avoid circular dependency)
// This must match the Profile struct from pkg/profiles/profiles.go
type Profile struct {
	Name        string
	BasePath    string
	HasAgents   bool
	HasSkills   bool
	HasCommands bool
}

// NewComponentLinker creates a new ComponentLinker with dependency injection
// The pm parameter is optional and can be nil for backward compatibility
func NewComponentLinker(agentsDir string, targets []config.Target, det *detector.RepositoryDetector, pm ProfileManager) (*ComponentLinker, error) {
	// Validate inputs
	if agentsDir == "" {
		return nil, fmt.Errorf("agentsDir cannot be empty")
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("at least one target must be provided")
	}
	if det == nil {
		return nil, fmt.Errorf("detector cannot be nil")
	}

	// Create target base directories if they don't exist
	for _, target := range targets {
		targetDir, err := target.GetBaseDir()
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

// SetFormatter sets a custom formatter for this linker (useful for testing)
func (cl *ComponentLinker) SetFormatter(f *formatter.Formatter) {
	cl.formatter = f
}

// filterTargets filters the targets based on the targetFilter parameter.
// Returns all targets if targetFilter is empty or "all", otherwise returns only the matching target.
func (cl *ComponentLinker) filterTargets(targetFilter string) []config.Target {
	// If no filter or "all", return all targets
	if targetFilter == "" || targetFilter == "all" {
		return cl.targets
	}

	// Filter for specific target
	filtered := make([]config.Target, 0)
	for _, target := range cl.targets {
		if target.GetName() == targetFilter {
			filtered = append(filtered, target)
		}
	}

	return filtered
}

// createSymlink creates a symbolic link from src to dst
func (cl *ComponentLinker) createSymlink(src, dst string) error {
	// Remove existing destination if it exists
	if _, err := os.Lstat(dst); err == nil {
		os.Remove(dst)
	}

	// Create relative path for the symlink
	relPath, err := filepath.Rel(filepath.Dir(dst), src)
	if err != nil {
		return fmt.Errorf("failed to create relative path: %w", err)
	}

	// Create the symbolic link
	if err := os.Symlink(relPath, dst); err != nil {
		// Try fallback to junction on Windows
		if runtime.GOOS == "windows" {
			return cl.createJunction(src, dst)
		}
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

// createJunction creates a Windows junction or falls back to copying
func (cl *ComponentLinker) createJunction(src, dst string) error {
	// For Windows, we would need to use Windows API calls for proper junctions
	// For now, fall back to copying the directory as cross-platform solution
	return cl.copyDirectory(src, dst)
}

// copyDirectory copies a directory from src to dst
func (cl *ComponentLinker) copyDirectory(src, dst string) error {
	return fileutil.CopyDirectoryContents(src, dst)
}

// copyFile copies a file from src to dst
func (cl *ComponentLinker) copyFile(src, dst string) error {
	return fileutil.CopyFile(src, dst)
}

// LinkComponent links a single component to all configured targets
func (cl *ComponentLinker) LinkComponent(componentType, componentName string) error {
	return cl.linkComponentInternal(componentType, componentName, true)
}

// linkComponentInternal links a single component with optional quiet mode for bulk operations
func (cl *ComponentLinker) linkComponentInternal(componentType, componentName string, verbose bool) error {
	srcDir := filepath.Join(cl.agentsDir, componentType, componentName)
	selectedProfileName := "" // Track which profile was used

	// Check if source component exists in current directory (active profile or base)
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		// Component not found in active profile/base directory
		// Search across all profiles
		matches, searchErr := cl.searchComponentInProfiles(componentType, componentName)
		if searchErr != nil {
			return fmt.Errorf("failed to search profiles: %w", searchErr)
		}

		if len(matches) == 0 {
			return fmt.Errorf("component %s/%s does not exist in any profile", componentType, componentName)
		}

		// If found in multiple profiles, prompt user to select
		if len(matches) > 1 {
			profilePath, profileName, err := cl.promptProfileSelection(componentType, componentName, matches)
			if err != nil {
				return err
			}
			srcDir = filepath.Join(profilePath, componentType, componentName)
			selectedProfileName = profileName
		} else {
			// Only one match, use it automatically
			srcDir = filepath.Join(matches[0].ProfilePath, componentType, componentName)
			selectedProfileName = matches[0].ProfileName
			fmt.Printf("  %s Component found in profile: %s\n", colors.Muted("→"), selectedProfileName)
		}
	} else {
		// Component found in current directory - determine profile name
		selectedProfileName = getProfileFromPath(srcDir)
	}

	// All components are now stored type-based, no special plugin handling needed
	metadata := cl.loadComponentMetadata(componentType, componentName)
	_ = metadata // Keep metadata loading for potential future use

	// Link to all configured targets
	// Track results for inline display
	type linkResult struct {
		name    string
		path    string
		success bool
		errMsg  string
	}
	var linkResults []linkResult

	for _, target := range cl.targets {
		targetName := target.GetName()

		// Get destination directory from target
		componentDir, err := target.GetComponentDir(componentType)
		if err != nil {
			linkResults = append(linkResults, linkResult{
				name:    targetName,
				success: false,
				errMsg:  fmt.Sprintf("failed to get target component directory: %v", err),
			})
			continue
		}
		dstDir := filepath.Join(componentDir, componentName)

		// Create destination directory
		if err := fileutil.CreateDirectoryWithPermissions(filepath.Dir(dstDir)); err != nil {
			linkResults = append(linkResults, linkResult{
				name:    targetName,
				path:    dstDir,
				success: false,
				errMsg:  fmt.Sprintf("failed to create destination directory: %v", err),
			})
			continue
		}

		// Create symlink or copy
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

	// Display results with modern inline progress format
	if len(linkResults) > 0 && verbose {
		hasSuccess := false
		for _, result := range linkResults {
			if result.success {
				hasSuccess = true
				break
			}
		}

		// Display inline progress: "Linking {type}: {name}... ✓ Done" or "✗ Failed"
		if hasSuccess {
			profileNote := styles.ProfileNoteFormat(selectedProfileName)
			fmt.Printf("%s%s\n", styles.InlineSuccessFormat("Linking", componentType, componentName), profileNote)

			// Show target details
			for _, result := range linkResults {
				if result.success {
					fmt.Printf("%s\n", styles.IndentedDetailFormat(result.name, result.path))
				}
			}
		} else {
			fmt.Printf("%s\n", styles.InlineFailedFormat("Linking", componentType, componentName))

			// Show errors indented
			for _, result := range linkResults {
				if !result.success {
					fmt.Printf("%s\n", styles.IndentedDetailFormat(result.name, result.errMsg))
				}
			}
			return fmt.Errorf("failed to link to target")
		}
	}

	// Return error for failed links even in quiet mode
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

// loadComponentMetadata loads metadata for a component from lock files only
func (cl *ComponentLinker) loadComponentMetadata(componentType, componentName string) *models.ComponentMetadata {
	return cl.loadFromLockFile(componentType, componentName)
}

// loadFromLockFile loads metadata from lock file
func (cl *ComponentLinker) loadFromLockFile(componentType, componentName string) *models.ComponentMetadata {
	metadata, err := metadataPkg.LoadFromLockFile(cl.agentsDir, componentType, componentName)
	if err != nil {
		return nil
	}
	return metadata
}

// LinkComponentsByType links all components of a specific type to opencode
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

	// Track results for summary
	var successCount, failedCount, skippedCount int
	var failedComponents []string

	// Display header with target info
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

			// Skip monorepo containers - they shouldn't be linked as individual components
			if cl.isMonorepoContainer(componentType, componentName) {
				skippedCount++
				continue
			}

			// Show inline progress: "Linking {type}: {name}... ✓ Done"
			fmt.Printf("Linking %s: %s... ", componentType, componentName)

			// Link as a regular single component (quiet mode for bulk operations)
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

	// Display summary table with box-drawing
	cl.formatter.EmptyLine()

	// Create summary table
	table := formatter.NewBoxTable(cl.formatter.Writer(), []string{"Status", "Count"})
	table.AddRow([]string{colors.Success(formatter.SymbolSuccess + " Success"), fmt.Sprintf("%d", successCount)})
	if skippedCount > 0 {
		table.AddRow([]string{colors.Warning(formatter.SymbolWarning + " Skipped"), fmt.Sprintf("%d (monorepos)", skippedCount)})
	}
	if failedCount > 0 {
		table.AddRow([]string{colors.Error(formatter.SymbolError + " Failed"), fmt.Sprintf("%d", failedCount)})
	}
	table.Render()

	// List failed components below table if any
	if failedCount > 0 {
		cl.formatter.EmptyLine()
		fmt.Println("Failed components:")
		for _, comp := range failedComponents {
			fmt.Printf("  • %s\n", comp)
		}
	}

	// Skipped components explanation
	if skippedCount > 0 {
		cl.formatter.EmptyLine()
		fmt.Printf("%s Monorepo containers were skipped - their individual components should be linked separately\n",
			colors.Muted("Note:"))
	}

	return nil
}

// LinkAllComponents links all components to opencode
func (cl *ComponentLinker) LinkAllComponents() error {
	componentTypes := paths.GetComponentTypes()

	// Track results for summary
	var successCount, failedCount, skippedCount int
	var failedComponents []string

	// Display header with target info
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

				// Show inline progress: "Linking {type}: {name}... ✓ Done"
				fmt.Printf("Linking %s: %s... ", componentType, componentName)

				// Link as a regular single component (quiet mode for bulk operations)
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

	// Display summary table with box-drawing
	cl.formatter.EmptyLine()

	// Create summary table
	table := formatter.NewBoxTable(cl.formatter.Writer(), []string{"Status", "Count"})
	table.AddRow([]string{colors.Success(formatter.SymbolSuccess + " Success"), fmt.Sprintf("%d", successCount)})
	if skippedCount > 0 {
		table.AddRow([]string{colors.Warning(formatter.SymbolWarning + " Skipped"), fmt.Sprintf("%d (monorepos)", skippedCount)})
	}
	if failedCount > 0 {
		table.AddRow([]string{colors.Error(formatter.SymbolError + " Failed"), fmt.Sprintf("%d", failedCount)})
	}
	table.Render()

	// List failed components below table if any
	if failedCount > 0 {
		cl.formatter.EmptyLine()
		fmt.Println("Failed components:")
		for _, comp := range failedComponents {
			fmt.Printf("  • %s\n", comp)
		}
	}

	// Skipped components explanation
	if skippedCount > 0 {
		cl.formatter.EmptyLine()
		fmt.Printf("%s Monorepo containers were skipped - their individual components should be linked separately\n",
			colors.Muted("Note:"))
	}

	return nil
}

// isMonorepoContainer checks if a component directory contains other component directories
// and should not be linked as a single component
func (cl *ComponentLinker) isMonorepoContainer(componentType, componentName string) bool {
	componentDir := filepath.Join(cl.agentsDir, componentType, componentName)

	// Check if this directory contains other component directories
	entries, err := os.ReadDir(componentDir)
	if err != nil {
		return false
	}

	// Determine possible marker files for this component type
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

	// Count how many subdirectories contain a marker file
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

	// If there are multiple sub-components, this is a monorepo container
	return subComponentCount > 1
}

// LinkMonorepoComponents links individual components from a monorepo container
func (cl *ComponentLinker) LinkMonorepoComponents(componentType, repoName string) error {
	repoDir := filepath.Join(cl.agentsDir, componentType, repoName)

	entries, err := os.ReadDir(repoDir)
	if err != nil {
		return fmt.Errorf("failed to read monorepo directory: %w", err)
	}

	// Determine possible marker files for this component type
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

			// Check if this subdirectory contains any marker file or a {name}.md file
			hasMarker := false
			for _, markerFile := range markerFiles {
				if _, err := os.Stat(filepath.Join(subComponentDir, markerFile)); err == nil {
					hasMarker = true
					break
				}
			}

			// Also check for {name}.md pattern
			if !hasMarker {
				if _, err := os.Stat(filepath.Join(subComponentDir, subComponentName+".md")); err == nil {
					hasMarker = true
				}
			}

			if hasMarker {
				// Link this sub-component using a unique name that includes the repo name
				linkName := fmt.Sprintf("%s-%s", repoName, subComponentName)

				// Create the link from the sub-component directory
				srcDir := subComponentDir

				// Link to all configured targets
				var successfulTargets []struct {
					name string
					path string
				}
				for _, target := range cl.targets {
					// Get destination directory from target
					componentDir, err := target.GetComponentDir(componentType)
					if err != nil {
						fmt.Printf("Warning: failed to get target component directory for %s: %v\n", target.GetName(), err)
						continue
					}
					dstDir := filepath.Join(componentDir, linkName)

					// Create destination directory
					if err := fileutil.CreateDirectoryWithPermissions(filepath.Dir(dstDir)); err != nil {
						fmt.Printf("Warning: failed to create destination directory for %s in %s: %v\n", linkName, target.GetName(), err)
						continue
					}

					// Create symlink
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

// DetectAndLinkLocalRepositories detects and links components from the current directory
func (cl *ComponentLinker) DetectAndLinkLocalRepositories() error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Check if current directory is a git repository
	if !cl.detector.IsLocalPath(cwd) {
		return fmt.Errorf("current directory is not a git repository")
	}

	// Detect components in the current repository
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
		// Convert component type to plural form for directory structure
		componentTypeStr := string(component.Type) + "s"
		componentPath := filepath.Join(cwd, component.Path)

		// Create a temporary link to the detected component
		tempLinkName := fmt.Sprintf("auto-detected-%s", component.Name)
		tempLinkPath := filepath.Join(cl.agentsDir, componentTypeStr, tempLinkName)

		// Create destination directory
		if err := fileutil.CreateDirectoryWithPermissions(filepath.Dir(tempLinkPath)); err != nil {
			fmt.Printf("Warning: failed to create directory for %s: %v\n", component.Name, err)
			continue
		}

		// Create symlink to the detected component
		if err := cl.createSymlink(componentPath, tempLinkPath); err != nil {
			fmt.Printf("Warning: failed to link component %s: %v\n", component.Name, err)
			continue
		}

		// Now link it to opencode
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

	// Loop through each target and display links
	for _, target := range cl.targets {
		allLinks := make(map[string][]LinkStatus)
		totalCount := 0
		validCount := 0
		brokenCount := 0

		for _, componentType := range componentTypes {
			componentDir, err := target.GetComponentDir(componentType)
			if err != nil {
				return fmt.Errorf("failed to get target component directory: %w", err)
			}

			// Check if directory exists
			if _, err := os.Stat(componentDir); os.IsNotExist(err) {
				allLinks[componentType] = []LinkStatus{}
				continue
			}

			// Read directory entries
			entries, err := os.ReadDir(componentDir)
			if err != nil {
				return fmt.Errorf("failed to read %s directory: %w", componentType, err)
			}

			links := []LinkStatus{}
			for _, entry := range entries {
				// Skip hidden files
				if strings.HasPrefix(entry.Name(), ".") {
					continue
				}

				fullPath := filepath.Join(componentDir, entry.Name())
				linkType, targetPath, valid := cl.analyzeLinkStatus(fullPath)

				// Determine profile from target path
				profile := "base"
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

		// Get target info for display
		targetName := target.GetName()
		targetDir, _ := target.GetBaseDir()

		// Display results for this target
		fmt.Printf("\n=== %s ===\n", strings.ToUpper(targetName))
		fmt.Printf("%s\n", cl.getSourceDescription())

		if totalCount == 0 {
			fmt.Printf("No components are currently linked to %s.\n", targetName)
			fmt.Printf("Link location: %s\n", targetDir)
			continue
		}

		// Display by type
		for _, componentType := range componentTypes {
			links := allLinks[componentType]
			if len(links) == 0 {
				continue
			}

			// Capitalize first letter for display
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

		// Summary
		fmt.Printf("\nTotal: %d components", totalCount)
		if brokenCount > 0 {
			fmt.Printf(" (%d valid, %d broken)", validCount, brokenCount)
		}
		fmt.Println()
	}

	return nil
}

// ShowLinkStatus displays a matrix view of components and their status across all targets
func (cl *ComponentLinker) ShowLinkStatus() error {
	componentTypes := paths.GetComponentTypes()

	// Collect all unique components from source directory
	type ComponentInfo struct {
		Name    string
		Type    string
		Profile string
	}
	allComponents := make([]ComponentInfo, 0)

	// Scan source directories for all available components
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
			// Determine profile from source directory
			componentPath := filepath.Join(sourceDir, entry.Name())
			profile := getProfileFromPath(componentPath)
			allComponents = append(allComponents, ComponentInfo{
				Name:    entry.Name(),
				Type:    componentType,
				Profile: profile,
			})
		}
	}

	if len(allComponents) == 0 {
		fmt.Fprintf(cl.formatter.Writer(), "No components found in ~/.agent-smith/\n")
		return nil
	}

	// Get link status for each component across all targets
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
			componentDir, err := target.GetComponentDir(comp.Type)
			if err != nil {
				status.Targets[target.GetName()] = colors.Warning("?")
				continue
			}

			linkPath := filepath.Join(componentDir, comp.Name)

			// Check if link exists
			if _, err := os.Lstat(linkPath); os.IsNotExist(err) {
				status.Targets[target.GetName()] = colors.Muted("-")
				continue
			}

			// Get link status
			linkType, _, valid := cl.analyzeLinkStatus(linkPath)

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

	// Display header
	cl.formatter.EmptyLine()
	cl.formatter.SectionHeader("Link Status Across All Targets")
	cl.formatter.InfoMsg(cl.getSourceDescription())
	cl.formatter.EmptyLine()

	// Get target names for header
	targetNames := make([]string, 0, len(cl.targets))
	for _, target := range cl.targets {
		targetNames = append(targetNames, target.GetName())
	}

	// Build headers for table
	headers := []string{"Component", "Profile"}
	for _, targetName := range targetNames {
		headers = append(headers, strings.ToUpper(targetName))
	}

	// Create table using formatter's writer
	table := formatter.NewBoxTable(cl.formatter.Writer(), headers)

	// Group by type and sort by name within each type
	byType := make(map[string][]ComponentStatus)
	for _, status := range statuses {
		byType[status.Component.Type] = append(byType[status.Component.Type], status)
	}

	// Display each component type
	for _, componentType := range componentTypes {
		components := byType[componentType]
		if len(components) == 0 {
			continue
		}

		// Add section header row
		sectionRow := []string{strings.Title(componentType) + ":", ""}
		for range targetNames {
			sectionRow = append(sectionRow, "")
		}
		table.AddRow(sectionRow)

		for _, status := range components {
			componentName := fmt.Sprintf("  %s", status.Component.Name)
			row := []string{componentName, status.Component.Profile}

			for _, targetName := range targetNames {
				symbol := status.Targets[targetName]
				row = append(row, symbol)
			}
			table.AddRow(row)
		}
	}

	// Render the table
	table.Render()

	// Print legend
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

	// Print summary
	cl.formatter.EmptyLine()
	cl.formatter.SubsectionHeader("Summary")
	for _, targetName := range targetNames {
		linkedCount := 0
		for _, status := range statuses {
			symbol := status.Targets[targetName]
			// Compare against color-wrapped symbols
			if symbol == colors.Success("✓") || symbol == colors.Success("◆") {
				linkedCount++
			}
		}
		cl.formatter.ListItem("%s: %d/%d components linked", strings.ToUpper(targetName), linkedCount, len(statuses))
	}

	return nil
}

// ShowAllProfilesLinkStatus displays link status for components across all profiles
// profileFilter can filter to specific profiles, or empty to show all
func (cl *ComponentLinker) ShowAllProfilesLinkStatus(profileFilter []string) error {
	// Validate that profileManager is available
	if cl.profileManager == nil {
		return fmt.Errorf("profile manager not available - this operation requires a profile manager")
	}

	componentTypes := paths.GetComponentTypes()

	// Collect all unique components from all profiles
	type ComponentInfo struct {
		Name    string
		Type    string
		Profile string
	}
	allComponents := make([]ComponentInfo, 0)

	// Get base installation directory
	baseDir, err := paths.GetAgentsDir()
	if err != nil {
		return fmt.Errorf("failed to get base directory: %w", err)
	}

	// Scan base installation
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
			baseComponents = append(baseComponents, ComponentInfo{
				Name:    entry.Name(),
				Type:    componentType,
				Profile: "base",
			})
		}
	}

	// Scan all profiles
	profiles, err := cl.profileManager.ScanProfiles()
	if err != nil {
		return fmt.Errorf("failed to scan profiles: %w", err)
	}

	// Apply profile filter if specified
	var filteredProfiles []*Profile
	if len(profileFilter) > 0 {
		filterMap := make(map[string]bool)
		for _, name := range profileFilter {
			filterMap[name] = true
		}

		// Validate that all filter names exist
		profileMap := make(map[string]bool)
		for _, p := range profiles {
			profileMap[p.Name] = true
		}

		for _, filterName := range profileFilter {
			if !profileMap[filterName] {
				return fmt.Errorf("profile '%s' does not exist", filterName)
			}
		}

		// Apply filter
		for _, p := range profiles {
			if filterMap[p.Name] {
				filteredProfiles = append(filteredProfiles, p)
			}
		}
	} else {
		filteredProfiles = profiles
	}

	// Scan each profile for components
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
					Name:    entry.Name(),
					Type:    componentType,
					Profile: profile.Name,
				})
			}
		}
	}

	// Combine base and profile components
	// Only include base components if no filter is applied or if we're showing all
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

	// Get link status for each component across all targets
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

		// For each target, check the link status
		for _, target := range cl.targets {
			componentDir, err := target.GetComponentDir(comp.Type)
			if err != nil {
				status.Targets[target.GetName()] = colors.Warning("?")
				continue
			}

			linkPath := filepath.Join(componentDir, comp.Name)

			// Check if link exists
			if _, err := os.Lstat(linkPath); os.IsNotExist(err) {
				status.Targets[target.GetName()] = colors.Muted("-")
				continue
			}

			// Get link status
			linkType, _, valid := cl.analyzeLinkStatus(linkPath)

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

	// Display header
	cl.formatter.EmptyLine()
	cl.formatter.SectionHeader("Link Status Across All Profiles")
	cl.formatter.EmptyLine()

	// Get target names for header
	targetNames := make([]string, 0, len(cl.targets))
	for _, target := range cl.targets {
		targetNames = append(targetNames, target.GetName())
	}

	// Build headers for table
	headers := []string{"Component", "Type", "Profile"}
	for _, targetName := range targetNames {
		headers = append(headers, strings.ToUpper(targetName))
	}

	// Create table
	table := formatter.NewBoxTable(cl.formatter.Writer(), headers)

	// Group by type and sort by name within each type
	byType := make(map[string][]ComponentStatus)
	for _, status := range statuses {
		byType[status.Component.Type] = append(byType[status.Component.Type], status)
	}

	// Display each component type
	for _, componentType := range componentTypes {
		components := byType[componentType]
		if len(components) == 0 {
			continue
		}

		// Add section header row
		sectionRow := []string{strings.Title(componentType) + ":", "", ""}
		for range targetNames {
			sectionRow = append(sectionRow, "")
		}
		table.AddRow(sectionRow)

		for _, status := range components {
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

	// Print legend
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

	// Print summary
	cl.formatter.EmptyLine()
	cl.formatter.SubsectionHeader("Summary")

	// Calculate profile count
	profileCount := len(filteredProfiles)
	if len(profileFilter) == 0 {
		profileCount++ // Include base
	}
	profileCountStr := fmt.Sprintf("%d", profileCount)
	if len(profileFilter) == 0 {
		if len(filteredProfiles) == 0 {
			profileCountStr = "1 (base only)"
		} else {
			profileCountStr = fmt.Sprintf("%d (base + %d custom)", profileCount, len(filteredProfiles))
		}
	}
	cl.formatter.ListItem("Profiles scanned: %s", profileCountStr)
	cl.formatter.ListItem("Total components: %d", len(statuses))

	for _, targetName := range targetNames {
		linkedCount := 0
		for _, status := range statuses {
			symbol := status.Targets[targetName]
			// Compare against color-wrapped symbols
			if symbol == colors.Success("✓") || symbol == colors.Success("◆") {
				linkedCount++
			}
		}
		percentage := 0
		if len(statuses) > 0 {
			percentage = (linkedCount * 100) / len(statuses)
		}
		cl.formatter.ListItem("%s: %d/%d linked (%d%%)", strings.ToUpper(targetName), linkedCount, len(statuses), percentage)
	}

	// Show active profile
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
	// Validate component type
	if componentType != "skills" && componentType != "agents" && componentType != "commands" {
		return fmt.Errorf("invalid component type: %s (must be skills, agents, or commands)", componentType)
	}

	// Filter targets based on targetFilter parameter
	targetsToUnlink := cl.filterTargets(targetFilter)
	if len(targetsToUnlink) == 0 {
		if targetFilter != "" && targetFilter != "all" {
			// Build list of available target names for the error message
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

	// Print header showing what we're unlinking from
	if targetFilter != "" && targetFilter != "all" {
		cl.formatter.SectionHeader(fmt.Sprintf("Unlinking %s '%s' from: %s", componentType, componentName, targetFilter))
	} else {
		// Build list of target names
		targetNames := make([]string, 0, len(targetsToUnlink))
		for _, target := range targetsToUnlink {
			targetNames = append(targetNames, target.GetName())
		}
		cl.formatter.SectionHeader(fmt.Sprintf("Unlinking %s '%s' from: %s", componentType, componentName, strings.Join(targetNames, ", ")))
	}

	for _, target := range targetsToUnlink {
		componentDir, err := target.GetComponentDir(componentType)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to get target component directory for %s: %v", target.GetName(), err))
			failedCount++
			continue
		}
		linkPath := filepath.Join(componentDir, componentName)

		targetName := target.GetName()

		// Check if link exists
		if _, err := os.Lstat(linkPath); os.IsNotExist(err) {
			// Not an error, just skip this target - component is already unlinked
			continue
		}

		// Analyze what we're removing
		linkType, targetPath, _ := cl.analyzeLinkStatus(linkPath)

		// For copied directories, ask for confirmation
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

		// Show progress
		cl.formatter.ProgressMsg(fmt.Sprintf("Unlinking from %s", targetName), componentName)

		// Remove the link or directory
		if linkType == "copied" {
			if err := os.RemoveAll(linkPath); err != nil {
				cl.formatter.ProgressFailed()
				errors = append(errors, fmt.Sprintf("failed to remove copied directory from %s: %v", targetName, err))
				failedCount++
				continue
			}
		} else {
			// For symlinks and broken links
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

	// Display warnings for any errors
	if len(errors) > 0 {
		for _, errMsg := range errors {
			cl.formatter.WarningMsg(errMsg)
		}
		if successCount == 0 {
			return fmt.Errorf("failed to unlink from any target")
		}
	}

	if successCount == 0 {
		return fmt.Errorf("component %s/%s is not linked to any target", componentType, componentName)
	}

	// Display summary
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
	// Filter targets based on targetFilter parameter
	targetsToUnlink := cl.filterTargets(targetFilter)
	if len(targetsToUnlink) == 0 {
		if targetFilter != "" && targetFilter != "all" {
			// Build list of available target names for the error message
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

	// First, collect all symlinks across all targets
	for _, target := range targetsToUnlink {
		componentDir, err := target.GetComponentDir(componentType)
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
				continue // Skip copied directories
			}
			totalLinks++
		}
	}

	if totalLinks == 0 && copiedDirs == 0 {
		cl.formatter.InfoMsg("No linked %s found", componentType)
		return nil
	}

	// Require force flag or confirmation
	if !force {
		if totalLinks > 0 {
			// Build target names string
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

	// Print header showing what we're unlinking
	if targetFilter != "" && targetFilter != "all" {
		cl.formatter.SectionHeader(fmt.Sprintf("Unlinking all %s from: %s", componentType, targetFilter))
	} else {
		// Build list of target names
		targetNames := make([]string, 0, len(targetsToUnlink))
		for _, target := range targetsToUnlink {
			targetNames = append(targetNames, target.GetName())
		}
		cl.formatter.SectionHeader(fmt.Sprintf("Unlinking all %s from: %s", componentType, strings.Join(targetNames, ", ")))
	}

	// Remove all symlinks (skip copied directories)
	removedCount := 0
	skippedCount := 0
	errorCount := 0

	for _, target := range targetsToUnlink {
		componentDir, err := target.GetComponentDir(componentType)
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

			// Skip copied directories - don't delete them
			if linkType == "copied" {
				skippedCount++
				continue
			}

			// Show progress for each item
			cl.formatter.ProgressMsg(fmt.Sprintf("Unlinking %s from %s", componentType, target.GetName()), entry.Name())

			// Remove symlinks and broken links
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

	// Display summary
	cl.formatter.EmptyLine()
	cl.formatter.CounterSummary(removedCount+errorCount, removedCount, errorCount, skippedCount)

	return nil
}

// isSymlinkFromCurrentProfile checks if a symlink belongs to the current profile
// by comparing the symlink target path with the ComponentLinker's agentsDir
func (cl *ComponentLinker) isSymlinkFromCurrentProfile(symlinkPath string) (bool, error) {
	// Read the symlink target
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		return false, err
	}

	// Resolve relative paths to absolute
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(symlinkPath), target)
	}

	// Clean the target path
	target = filepath.Clean(target)

	// Get profile names and compare
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
	// Read the symlink target
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		return false, err
	}

	// Resolve relative paths to absolute
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(symlinkPath), target)
	}

	// Clean path for comparison
	target = filepath.Clean(target)

	// First check if target starts with the ComponentLinker's agentsDir
	// This is important for tests and for the current profile/base installation
	agentsDir := filepath.Clean(cl.agentsDir)
	if strings.HasPrefix(target, agentsDir) {
		return true, nil
	}

	// Get base agent-smith directory (parent of profiles)
	baseAgentsDir, err := paths.GetAgentsDir()
	if err != nil {
		// If we can't get base agents dir, just rely on the ComponentLinker's agentsDir check above
		return false, nil
	}

	// Check if target starts with base agents directory
	if strings.HasPrefix(target, baseAgentsDir) {
		return true, nil
	}

	// Check if target is within any profile
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		// If we can't get profiles dir, just rely on checks above
		return false, nil
	}

	if strings.HasPrefix(target, profilesDir) {
		return true, nil
	}

	// Not from agent-smith
	return false, nil
}

// anyProfilesExist checks if any profiles directory exists and contains profiles
// Returns false for fresh installations with no profiles (backward compatibility)
func (cl *ComponentLinker) anyProfilesExist() bool {
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		return false
	}

	// Check if profiles directory exists
	info, err := os.Stat(profilesDir)
	if err != nil || !info.IsDir() {
		return false
	}

	// Check if there are any profile directories inside
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return false
	}

	// Check for at least one valid profile directory
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
	// Filter targets based on targetFilter parameter
	targetsToUnlink := cl.filterTargets(targetFilter)
	if len(targetsToUnlink) == 0 {
		if targetFilter != "" && targetFilter != "all" {
			// Build list of available target names for the error message
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

	// First, collect all symlinks across all targets
	totalLinks := 0
	copiedDirs := 0
	skippedProfilesCount := 0
	skippedProfilesMap := make(map[string]int) // Track count per profile

	for _, target := range targetsToUnlink {
		for _, componentType := range componentTypes {
			componentDir, err := target.GetComponentDir(componentType)
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
					continue // Skip copied directories
				}

				// Skip manual/external symlinks (not from agent-smith)
				if linkType == "symlink" || linkType == "broken" {
					fromAgentSmith, err := cl.isSymlinkFromAgentSmith(fullPath)
					if err == nil && !fromAgentSmith {
						// This is a manually created symlink outside agent-smith - always preserve
						skippedProfilesCount++
						continue
					}
				}

				// If not unlinking all profiles, check if this symlink belongs to current profile
				// This applies to both valid symlinks and broken symlinks
				if !allProfiles && (linkType == "symlink" || linkType == "broken") {
					belongsToProfile, err := cl.isSymlinkFromCurrentProfile(fullPath)
					if err == nil && !belongsToProfile {
						// Track which profile this component belongs to
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

	// Determine current profile name for messaging
	currentProfileName := getProfileFromPath(cl.agentsDir)

	// Check if any profiles exist at all (for backward compatibility messaging)
	profilesExist := cl.anyProfilesExist()

	// Require force flag or confirmation
	if !force {
		if totalLinks > 0 {
			// Build target description for confirmation message
			targetStr := "all targets"
			if targetFilter != "" && targetFilter != "all" {
				targetStr = targetFilter
			} else if len(targetsToUnlink) > 0 {
				// List specific targets
				targetNames := make([]string, 0, len(targetsToUnlink))
				for _, target := range targetsToUnlink {
					targetNames = append(targetNames, target.GetName())
				}
				targetStr = strings.Join(targetNames, ", ")
			}

			profileMsg := ""
			// Only show profile-related messages if profiles exist or --all-profiles was used
			if allProfiles {
				profileMsg = " from all profiles"
			} else if profilesExist {
				// Show profile context only when profiles actually exist
				if currentProfileName == "base" {
					profileMsg = " from base installation"
				} else {
					profileMsg = fmt.Sprintf(" from profile '%s'", currentProfileName)
				}
			}
			// If no profiles exist and not using --all-profiles, don't add any profile message
			fmt.Printf("This will unlink %d symlinked components%s from: %s", totalLinks, profileMsg, targetStr)
			fmt.Println()
		}
		if copiedDirs > 0 {
			cl.formatter.InfoMsg("Note: %d copied directories will be skipped (not deleted)", copiedDirs)
		}
		if skippedProfilesCount > 0 && !allProfiles {
			// Show detailed breakdown of skipped profiles
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
				if currentProfileName == "base" {
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

	// Print header showing what we're unlinking with profile context
	headerMsg := ""
	if targetFilter != "" && targetFilter != "all" {
		if allProfiles {
			headerMsg = fmt.Sprintf("Unlinking all components (all profiles) from: %s", targetFilter)
		} else if profilesExist {
			// Only show profile context in header when profiles actually exist
			if currentProfileName == "base" {
				headerMsg = fmt.Sprintf("Unlinking components (base installation) from: %s", targetFilter)
			} else {
				headerMsg = fmt.Sprintf("Unlinking components (profile '%s') from: %s", currentProfileName, targetFilter)
			}
		} else {
			// No profiles exist - use simple messaging for backward compatibility
			headerMsg = fmt.Sprintf("Unlinking components from: %s", targetFilter)
		}
	} else {
		// Build list of target names
		targetNames := make([]string, 0, len(targetsToUnlink))
		for _, target := range targetsToUnlink {
			targetNames = append(targetNames, target.GetName())
		}
		targetList := strings.Join(targetNames, ", ")
		if allProfiles {
			headerMsg = fmt.Sprintf("Unlinking all components (all profiles) from: %s", targetList)
		} else if profilesExist {
			// Only show profile context in header when profiles actually exist
			if currentProfileName == "base" {
				headerMsg = fmt.Sprintf("Unlinking components (base installation) from: %s", targetList)
			} else {
				headerMsg = fmt.Sprintf("Unlinking components (profile '%s') from: %s", currentProfileName, targetList)
			}
		} else {
			// No profiles exist - use simple messaging for backward compatibility
			headerMsg = fmt.Sprintf("Unlinking components from: %s", targetList)
		}
	}
	cl.formatter.SectionHeader(headerMsg)

	// Remove all symlinks (skip copied directories and other profiles' components)
	removedCount := 0
	skippedCount := 0
	skippedByProfile := make(map[string][]string) // Track skipped items by profile
	errorCount := 0

	for _, target := range targetsToUnlink {
		for _, componentType := range componentTypes {
			componentDir, err := target.GetComponentDir(componentType)
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

				// Skip copied directories - don't delete them
				if linkType == "copied" {
					skippedCount++
					continue
				}

				// Skip manual/external symlinks (not from agent-smith)
				if linkType == "symlink" || linkType == "broken" {
					fromAgentSmith, err := cl.isSymlinkFromAgentSmith(fullPath)
					if err == nil && !fromAgentSmith {
						// This is a manually created symlink outside agent-smith - always preserve
						skippedCount++
						continue
					}
				}

				// If not unlinking all profiles, check if this symlink belongs to current profile
				// This applies to both valid symlinks and broken symlinks
				if !allProfiles && (linkType == "symlink" || linkType == "broken") {
					belongsToProfile, err := cl.isSymlinkFromCurrentProfile(fullPath)
					if err == nil && !belongsToProfile {
						// Track which profile this skipped component belongs to
						profileName := GetProfileNameFromSymlink(fullPath)
						if profileName != "" {
							skippedByProfile[profileName] = append(skippedByProfile[profileName], fmt.Sprintf("%s/%s", componentType, entry.Name()))
						}
						skippedCount++
						continue
					}
				}

				// Determine profile for progress message
				profileNote := ""
				// Only show profile tags when profiles actually exist
				// This applies to both valid symlinks and broken symlinks
				if profilesExist && (linkType == "symlink" || linkType == "broken") {
					profileName := GetProfileNameFromSymlink(fullPath)
					if profileName != "" && profileName != "base" {
						profileNote = fmt.Sprintf(" [%s]", profileName)
					}
				}

				// Show progress for each item with profile context
				cl.formatter.ProgressMsg(fmt.Sprintf("Unlinking %s from %s%s", componentType, target.GetName(), profileNote), entry.Name())

				// Remove symlinks and broken links
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

	// Display summary with profile breakdown
	cl.formatter.EmptyLine()
	cl.formatter.CounterSummary(removedCount+errorCount, removedCount, errorCount, skippedCount)

	// Show detailed breakdown of skipped items by profile if any
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
		if currentProfileName == "base" {
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
	// Get profiles directory
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get profiles directory: %w", err)
	}

	// Check if profiles directory exists
	if _, err := os.Stat(profilesDir); os.IsNotExist(err) {
		return []ProfileMatch{}, nil // No profiles yet
	}

	// Get active profile
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}
	activeProfilePath := filepath.Join(agentsDir, ".active-profile")
	activeProfileData, _ := os.ReadFile(activeProfilePath)
	activeProfile := strings.TrimSpace(string(activeProfileData))

	// Scan all profile directories
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read profiles directory: %w", err)
	}

	var matches []ProfileMatch
	for _, entry := range entries {
		// Skip files and hidden directories
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		profileName := entry.Name()
		profilePath := filepath.Join(profilesDir, profileName)
		componentPath := filepath.Join(profilePath, componentType, componentName)

		// Check if component exists in this profile
		if _, err := os.Stat(componentPath); err == nil {
			// Try to load source URL from lock file
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

		// Display source URL if available
		if match.SourceUrl != "" {
			fmt.Printf("     Source: %s\n", match.SourceUrl)
		}

		// Add blank line between options for readability
		if i < len(matches)-1 {
			fmt.Println()
		}
	}

	fmt.Printf("\nSelect profile to link from [1-%d] (or 'c' to cancel): ", len(matches))

	var response string
	fmt.Scanln(&response)
	response = strings.TrimSpace(strings.ToLower(response))

	// Check for cancellation
	if response == "c" || response == "" {
		return "", "", fmt.Errorf("profile selection cancelled")
	}

	// Parse selection
	var selection int
	_, err := fmt.Sscanf(response, "%d", &selection)
	if err != nil || selection < 1 || selection > len(matches) {
		return "", "", fmt.Errorf("invalid selection: %s", response)
	}

	selectedMatch := matches[selection-1]
	return selectedMatch.ProfilePath, selectedMatch.ProfileName, nil
}
