package linker

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tgaines/agent-smith/internal/detector"
	"github.com/tgaines/agent-smith/internal/fileutil"
	metadataPkg "github.com/tgaines/agent-smith/internal/metadata"
	"github.com/tgaines/agent-smith/internal/models"
	"github.com/tgaines/agent-smith/pkg/config"
	"github.com/tgaines/agent-smith/pkg/paths"
)

// ComponentLinker handles linking components to configured targets
type ComponentLinker struct {
	agentsDir string
	targets   []config.Target
	detector  *detector.RepositoryDetector
}

// NewComponentLinker creates a new ComponentLinker with dependency injection
func NewComponentLinker(agentsDir string, targets []config.Target, det *detector.RepositoryDetector) (*ComponentLinker, error) {
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
		agentsDir: agentsDir,
		targets:   targets,
		detector:  det,
	}, nil
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
	srcDir := filepath.Join(cl.agentsDir, componentType, componentName)

	// Check if source component exists
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("component %s/%s does not exist in %s", componentType, componentName, cl.agentsDir)
	}

	// All components are now stored type-based, no special plugin handling needed
	metadata := cl.loadComponentMetadata(componentType, componentName)
	_ = metadata // Keep metadata loading for potential future use

	// Link to all configured targets
	var errors []string
	successCount := 0

	for _, target := range cl.targets {
		// Get destination directory from target
		componentDir, err := target.GetComponentDir(componentType)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to get target component directory for %s: %v", target.GetName(), err))
			continue
		}
		dstDir := filepath.Join(componentDir, componentName)

		// Create destination directory
		if err := fileutil.CreateDirectoryWithPermissions(filepath.Dir(dstDir)); err != nil {
			errors = append(errors, fmt.Sprintf("failed to create destination directory for %s: %v", target.GetName(), err))
			continue
		}

		// Create symlink or copy
		if err := cl.createSymlink(srcDir, dstDir); err != nil {
			errors = append(errors, fmt.Sprintf("failed to link component to %s: %v", target.GetName(), err))
			continue
		}

		targetName := target.GetName()
		fmt.Printf("Successfully linked %s '%s' to %s\n", componentType, componentName, targetName)
		fmt.Printf("  Target: %s\n", dstDir)
		successCount++
	}

	if successCount > 0 {
		fmt.Printf("  Source: %s\n", srcDir)
	}

	if len(errors) > 0 {
		for _, errMsg := range errors {
			fmt.Printf("Warning: %s\n", errMsg)
		}
		if successCount == 0 {
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

	linkedCount := 0
	errorCount := 0

	for _, entry := range entries {
		if entry.IsDir() {
			componentName := entry.Name()

			// Skip monorepo containers - they shouldn't be linked as individual components
			if cl.isMonorepoContainer(componentType, componentName) {
				continue
			}

			// Link as a regular single component
			if err := cl.LinkComponent(componentType, componentName); err != nil {
				fmt.Printf("Warning: failed to link %s/%s: %v\n", componentType, componentName, err)
				errorCount++
			} else {
				linkedCount++
			}
		}
	}

	fmt.Printf("\nSuccessfully linked %d %s", linkedCount, componentType)
	if errorCount > 0 {
		fmt.Printf(" (%d errors)", errorCount)
	}
	fmt.Println()

	return nil
}

// LinkAllComponents links all components to opencode
func (cl *ComponentLinker) LinkAllComponents() error {
	componentTypes := paths.GetComponentTypes()

	for _, componentType := range componentTypes {
		typeDir := filepath.Join(cl.agentsDir, componentType)
		if _, err := os.Stat(typeDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(typeDir)
		if err != nil {
			fmt.Printf("Warning: failed to read %s directory: %v\n", componentType, err)
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				componentName := entry.Name()

				// Skip monorepo containers - they shouldn't be linked as individual components
				if cl.isMonorepoContainer(componentType, componentName) {
					continue
				}

				// Link as a regular single component
				if err := cl.LinkComponent(componentType, componentName); err != nil {
					fmt.Printf("Warning: failed to link %s/%s: %v\n", componentType, componentName, err)
				}
			}
		}
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
				linkedToAnyTarget := false
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

					linkedToAnyTarget = true
				}

				if linkedToAnyTarget {
					fmt.Printf("Successfully linked monorepo component %s from %s\n", linkName, repoName)
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

				status := LinkStatus{
					Name:       entry.Name(),
					Type:       componentType,
					LinkType:   linkType,
					Target:     targetPath,
					Valid:      valid,
					TargetPath: fullPath,
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

				fmt.Printf("  %s %s %s\n", symbol, link.Name, statusMsg)
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
		Name string
		Type string
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
			allComponents = append(allComponents, ComponentInfo{
				Name: entry.Name(),
				Type: componentType,
			})
		}
	}

	if len(allComponents) == 0 {
		fmt.Println("No components found in ~/.agent-smith/")
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
				status.Targets[target.GetName()] = "?"
				continue
			}

			linkPath := filepath.Join(componentDir, comp.Name)

			// Check if link exists
			if _, err := os.Lstat(linkPath); os.IsNotExist(err) {
				status.Targets[target.GetName()] = "-"
				continue
			}

			// Get link status
			linkType, _, valid := cl.analyzeLinkStatus(linkPath)

			var symbol string
			switch linkType {
			case "symlink":
				if valid {
					symbol = "✓"
				} else {
					symbol = "✗"
				}
			case "copied":
				symbol = "◆"
			case "broken":
				symbol = "✗"
			default:
				symbol = "?"
			}

			status.Targets[target.GetName()] = symbol
		}

		statuses = append(statuses, status)
	}

	// Display header
	fmt.Println("\n=== Link Status Across All Targets ===")
	fmt.Printf("%s\n\n", cl.getSourceDescription())

	// Get target names for header
	targetNames := make([]string, 0, len(cl.targets))
	for _, target := range cl.targets {
		targetNames = append(targetNames, target.GetName())
	}

	// Calculate column widths
	maxNameLen := 20
	for _, status := range statuses {
		nameLen := len(status.Component.Name) + 2 // +2 for indent
		if nameLen > maxNameLen {
			maxNameLen = nameLen
		}
	}

	// Print header
	fmt.Printf("%-*s", maxNameLen+2, "Component")
	for _, targetName := range targetNames {
		fmt.Printf("  %-12s", strings.ToUpper(targetName))
	}
	fmt.Println()

	// Print separator
	fmt.Print(strings.Repeat("-", maxNameLen+2))
	for range targetNames {
		fmt.Print("  " + strings.Repeat("-", 12))
	}
	fmt.Println()

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

		fmt.Printf("\n%s:\n", strings.Title(componentType))

		for _, status := range components {
			componentName := fmt.Sprintf("  %s", status.Component.Name)
			fmt.Printf("%-*s", maxNameLen+2, componentName)

			for _, targetName := range targetNames {
				symbol := status.Targets[targetName]
				fmt.Printf("  %-12s", symbol)
			}
			fmt.Println()
		}
	}

	// Print legend
	fmt.Println("\nLegend:")
	fmt.Println("  ✓  Valid symlink")
	fmt.Println("  ◆  Copied directory")
	fmt.Println("  ✗  Broken link")
	fmt.Println("  -  Not linked")
	fmt.Println("  ?  Unknown status")

	// Print summary
	fmt.Println("\nSummary:")
	for _, targetName := range targetNames {
		linkedCount := 0
		for _, status := range statuses {
			symbol := status.Targets[targetName]
			if symbol == "✓" || symbol == "◆" {
				linkedCount++
			}
		}
		fmt.Printf("  %s: %d/%d components linked\n", strings.ToUpper(targetName), linkedCount, len(statuses))
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
	var errors []string
	var unlinkedTargets []string

	for _, target := range targetsToUnlink {
		componentDir, err := target.GetComponentDir(componentType)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to get target component directory for %s: %v", target.GetName(), err))
			continue
		}
		linkPath := filepath.Join(componentDir, componentName)

		targetName := target.GetName()

		// Check if link exists
		if _, err := os.Lstat(linkPath); os.IsNotExist(err) {
			// Not an error, just skip this target
			continue
		}

		// Analyze what we're removing
		linkType, targetPath, _ := cl.analyzeLinkStatus(linkPath)

		// For copied directories, ask for confirmation
		if linkType == "copied" {
			fmt.Printf("Warning: '%s' is a copied directory in %s, not a symlink.\n", componentName, targetName)
			fmt.Printf("This will permanently delete: %s\n", linkPath)
			fmt.Print("Continue? [y/N]: ")

			var response string
			fmt.Scanln(&response)

			if strings.ToLower(strings.TrimSpace(response)) != "y" {
				fmt.Printf("Unlink cancelled for %s.\n", targetName)
				continue
			}
		}

		// Remove the link or directory
		if linkType == "copied" {
			if err := os.RemoveAll(linkPath); err != nil {
				errors = append(errors, fmt.Sprintf("failed to remove copied directory from %s: %v", targetName, err))
				continue
			}
		} else {
			// For symlinks and broken links
			if err := os.Remove(linkPath); err != nil {
				errors = append(errors, fmt.Sprintf("failed to remove link from %s: %v", targetName, err))
				continue
			}
		}

		unlinkedTargets = append(unlinkedTargets, targetName)

		if linkType == "symlink" && targetPath != "" {
			fmt.Printf("  Source still available at: %s\n", targetPath)
		}
		successCount++
	}

	if len(errors) > 0 {
		for _, errMsg := range errors {
			fmt.Printf("Warning: %s\n", errMsg)
		}
		if successCount == 0 {
			return fmt.Errorf("failed to unlink from any target")
		}
	}

	if successCount == 0 {
		return fmt.Errorf("component %s/%s is not linked to any target", componentType, componentName)
	}

	// Display summary of affected targets
	fmt.Printf("Successfully unlinked %s '%s' from %d target(s): %s\n",
		componentType, componentName, successCount, strings.Join(unlinkedTargets, ", "))

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
		fmt.Printf("No linked %s found.\n", componentType)
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
			fmt.Printf("Note: %d copied directories will be skipped (not deleted)\n", copiedDirs)
		}
		if totalLinks == 0 {
			fmt.Printf("No symlinked %s to unlink (only copied directories found).\n", componentType)
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
			fmt.Printf("Warning: failed to read %s directory for %s: %v\n", componentType, target.GetName(), err)
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

			// Remove symlinks and broken links
			err := os.Remove(fullPath)

			if err != nil {
				fmt.Printf("Warning: failed to unlink %s/%s from %s: %v\n", componentType, entry.Name(), target.GetName(), err)
				errorCount++
			} else {
				removedCount++
			}
		}
	}

	fmt.Printf("\nSuccessfully unlinked %d %s", removedCount, componentType)
	if skippedCount > 0 {
		fmt.Printf(" (%d copied directories skipped)", skippedCount)
	}
	if errorCount > 0 {
		fmt.Printf(" (%d errors)", errorCount)
	}
	fmt.Println()

	return nil
}

// UnlinkAllComponents removes all linked components from configured targets
// targetFilter can be:
//   - "" (empty): unlink from all targets
//   - "all": unlink from all targets
//   - specific target name (e.g., "opencode", "claudecode"): unlink from only that target
func (cl *ComponentLinker) UnlinkAllComponents(targetFilter string, force bool) error {
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
				totalLinks++
			}
		}
	}

	if totalLinks == 0 && copiedDirs == 0 {
		fmt.Println("No linked components found.")
		return nil
	}

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
			fmt.Printf("This will unlink %d symlinked components from: %s", totalLinks, targetStr)
			fmt.Println()
		}
		if copiedDirs > 0 {
			fmt.Printf("Note: %d copied directories will be skipped (not deleted)\n", copiedDirs)
		}
		if totalLinks == 0 {
			fmt.Println("No symlinked components to unlink (only copied directories found).")
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

	// Remove all symlinks (skip copied directories)
	removedCount := 0
	skippedCount := 0
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
				fmt.Printf("Warning: failed to read %s directory for %s: %v\n", componentType, target.GetName(), err)
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

				// Remove symlinks and broken links
				var err error
				err = os.Remove(fullPath)

				if err != nil {
					fmt.Printf("Warning: failed to unlink %s/%s from %s: %v\n", componentType, entry.Name(), target.GetName(), err)
					errorCount++
				} else {
					removedCount++
				}
			}
		}
	}

	fmt.Printf("\nSuccessfully unlinked %d components", removedCount)
	if skippedCount > 0 {
		fmt.Printf(" (%d copied directories skipped)", skippedCount)
	}
	if errorCount > 0 {
		fmt.Printf(" (%d errors)", errorCount)
	}
	fmt.Println()

	return nil
}
