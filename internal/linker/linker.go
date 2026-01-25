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
	"github.com/tgaines/agent-smith/pkg/paths"
)

// ComponentLinker handles linking components to the opencode directory
type ComponentLinker struct {
	agentsDir   string
	opencodeDir string
	detector    *detector.RepositoryDetector
}

// NewComponentLinker creates a new ComponentLinker with dependency injection
func NewComponentLinker(agentsDir, opencodeDir string, det *detector.RepositoryDetector) (*ComponentLinker, error) {
	// Validate inputs
	if agentsDir == "" {
		return nil, fmt.Errorf("agentsDir cannot be empty")
	}
	if opencodeDir == "" {
		return nil, fmt.Errorf("opencodeDir cannot be empty")
	}
	if det == nil {
		return nil, fmt.Errorf("detector cannot be nil")
	}

	// Create opencode directory if it doesn't exist
	if err := fileutil.CreateDirectoryWithPermissions(opencodeDir); err != nil {
		return nil, fmt.Errorf("failed to create opencode directory: %w", err)
	}

	return &ComponentLinker{
		agentsDir:   agentsDir,
		opencodeDir: opencodeDir,
		detector:    det,
	}, nil
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

// LinkComponent links a single component to opencode
func (cl *ComponentLinker) LinkComponent(componentType, componentName string) error {
	srcDir := filepath.Join(cl.agentsDir, componentType, componentName)
	dstDir := filepath.Join(cl.opencodeDir, componentType, componentName)

	// Check if source component exists
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("component %s/%s does not exist in %s", componentType, componentName, cl.agentsDir)
	}

	// All components are now stored type-based, no special plugin handling needed
	metadata := cl.loadComponentMetadata(componentType, componentName)
	_ = metadata // Keep metadata loading for potential future use

	// Create destination directory
	if err := fileutil.CreateDirectoryWithPermissions(filepath.Dir(dstDir)); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Create symlink or copy
	if err := cl.createSymlink(srcDir, dstDir); err != nil {
		return fmt.Errorf("failed to link component: %w", err)
	}

	fmt.Printf("Successfully linked %s '%s' to opencode\n", componentType, componentName)
	fmt.Printf("Source: %s\n", srcDir)
	fmt.Printf("Target: %s\n", dstDir)

	return nil
}

// loadComponentMetadata loads metadata for a component from lock files or metadata files
func (cl *ComponentLinker) loadComponentMetadata(componentType, componentName string) *models.ComponentMetadata {
	return metadataPkg.LoadComponentMetadata(cl.agentsDir, componentType, componentName)
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
				dstDir := filepath.Join(cl.opencodeDir, componentType, linkName)

				// Create destination directory
				if err := fileutil.CreateDirectoryWithPermissions(filepath.Dir(dstDir)); err != nil {
					fmt.Printf("Warning: failed to create destination directory for %s: %v\n", linkName, err)
					continue
				}

				// Create symlink
				if err := cl.createSymlink(srcDir, dstDir); err != nil {
					fmt.Printf("Warning: failed to link monorepo component %s: %v\n", linkName, err)
					continue
				}

				fmt.Printf("Successfully linked monorepo component %s from %s\n", linkName, repoName)
				linkedCount++
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
		componentTypeStr := string(component.Type)
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

// ListLinkedComponents lists all components linked to opencode
func (cl *ComponentLinker) ListLinkedComponents() error {
	componentTypes := paths.GetComponentTypes()

	allLinks := make(map[string][]LinkStatus)
	totalCount := 0
	validCount := 0
	brokenCount := 0

	for _, componentType := range componentTypes {
		typeDir := filepath.Join(cl.opencodeDir, componentType)

		// Check if directory exists
		if _, err := os.Stat(typeDir); os.IsNotExist(err) {
			allLinks[componentType] = []LinkStatus{}
			continue
		}

		// Read directory entries
		entries, err := os.ReadDir(typeDir)
		if err != nil {
			return fmt.Errorf("failed to read %s directory: %w", componentType, err)
		}

		links := []LinkStatus{}
		for _, entry := range entries {
			// Skip hidden files
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			fullPath := filepath.Join(typeDir, entry.Name())
			linkType, target, valid := cl.analyzeLinkStatus(fullPath)

			status := LinkStatus{
				Name:       entry.Name(),
				Type:       componentType,
				LinkType:   linkType,
				Target:     target,
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

	// Display results
	if totalCount == 0 {
		fmt.Println("No components are currently linked to opencode.")
		fmt.Printf("Link location: %s\n", cl.opencodeDir)
		return nil
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

	return nil
}

// UnlinkComponent removes a linked component from opencode
func (cl *ComponentLinker) UnlinkComponent(componentType, componentName string) error {
	// Validate component type
	if componentType != "skills" && componentType != "agents" && componentType != "commands" {
		return fmt.Errorf("invalid component type: %s (must be skills, agents, or commands)", componentType)
	}

	linkPath := filepath.Join(cl.opencodeDir, componentType, componentName)

	// Check if link exists
	if _, err := os.Lstat(linkPath); os.IsNotExist(err) {
		return fmt.Errorf("component %s/%s is not linked to opencode", componentType, componentName)
	}

	// Analyze what we're removing
	linkType, target, _ := cl.analyzeLinkStatus(linkPath)

	// For copied directories, ask for confirmation
	if linkType == "copied" {
		fmt.Printf("Warning: '%s' is a copied directory, not a symlink.\n", componentName)
		fmt.Printf("This will permanently delete: %s\n", linkPath)
		fmt.Print("Continue? [y/N]: ")

		var response string
		fmt.Scanln(&response)

		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("Unlink cancelled.")
			return nil
		}
	}

	// Remove the link or directory
	if linkType == "copied" {
		if err := os.RemoveAll(linkPath); err != nil {
			return fmt.Errorf("failed to remove copied directory: %w", err)
		}
	} else {
		// For symlinks and broken links
		if err := os.Remove(linkPath); err != nil {
			return fmt.Errorf("failed to remove link: %w", err)
		}
	}

	fmt.Printf("Successfully unlinked %s '%s' from opencode\n", componentType, componentName)

	if linkType == "symlink" && target != "" {
		fmt.Printf("Source still available at: %s\n", target)
	}

	return nil
}

// UnlinkComponentsByType removes all linked components of a specific type from opencode
func (cl *ComponentLinker) UnlinkComponentsByType(componentType string, force bool) error {
	typeDir := filepath.Join(cl.opencodeDir, componentType)

	if _, err := os.Stat(typeDir); os.IsNotExist(err) {
		fmt.Printf("No linked %s found.\n", componentType)
		return nil
	}

	// First, collect all symlinks (skip copied directories)
	totalLinks := 0
	copiedDirs := 0

	entries, err := os.ReadDir(typeDir)
	if err != nil {
		return fmt.Errorf("failed to read %s directory: %w", componentType, err)
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		fullPath := filepath.Join(typeDir, entry.Name())
		linkType, _, _ := cl.analyzeLinkStatus(fullPath)

		if linkType == "copied" {
			copiedDirs++
			continue // Skip copied directories
		}
		totalLinks++
	}

	if totalLinks == 0 && copiedDirs == 0 {
		fmt.Printf("No linked %s found.\n", componentType)
		return nil
	}

	// Require force flag or confirmation
	if !force {
		if totalLinks > 0 {
			fmt.Printf("This will unlink %d %s from opencode", totalLinks, componentType)
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

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		fullPath := filepath.Join(typeDir, entry.Name())
		linkType, _, _ := cl.analyzeLinkStatus(fullPath)

		// Skip copied directories - don't delete them
		if linkType == "copied" {
			skippedCount++
			continue
		}

		// Remove symlinks and broken links
		err := os.Remove(fullPath)

		if err != nil {
			fmt.Printf("Warning: failed to unlink %s/%s: %v\n", componentType, entry.Name(), err)
			errorCount++
		} else {
			removedCount++
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

// UnlinkAllComponents removes all linked components from opencode
func (cl *ComponentLinker) UnlinkAllComponents(force bool) error {
	componentTypes := paths.GetComponentTypes()

	// First, collect all symlinks (skip copied directories)
	totalLinks := 0
	copiedDirs := 0

	for _, componentType := range componentTypes {
		typeDir := filepath.Join(cl.opencodeDir, componentType)

		if _, err := os.Stat(typeDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(typeDir)
		if err != nil {
			return fmt.Errorf("failed to read %s directory: %w", componentType, err)
		}

		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			fullPath := filepath.Join(typeDir, entry.Name())
			linkType, _, _ := cl.analyzeLinkStatus(fullPath)

			if linkType == "copied" {
				copiedDirs++
				continue // Skip copied directories
			}
			totalLinks++
		}
	}

	if totalLinks == 0 && copiedDirs == 0 {
		fmt.Println("No linked components found.")
		return nil
	}

	// Require force flag or confirmation
	if !force {
		if totalLinks > 0 {
			fmt.Printf("This will unlink %d symlinked components from opencode", totalLinks)
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

	for _, componentType := range componentTypes {
		typeDir := filepath.Join(cl.opencodeDir, componentType)

		if _, err := os.Stat(typeDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(typeDir)
		if err != nil {
			fmt.Printf("Warning: failed to read %s directory: %v\n", componentType, err)
			continue
		}

		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			fullPath := filepath.Join(typeDir, entry.Name())
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
				fmt.Printf("Warning: failed to unlink %s/%s: %v\n", componentType, entry.Name(), err)
				errorCount++
			} else {
				removedCount++
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
