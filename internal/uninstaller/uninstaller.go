package uninstaller

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tgaines/agent-smith/internal/detector"
	"github.com/tgaines/agent-smith/internal/linker"
	"github.com/tgaines/agent-smith/internal/metadata"
	"github.com/tgaines/agent-smith/pkg/paths"
)

// Uninstaller handles component removal
type Uninstaller struct {
	baseDir string
	linker  *linker.ComponentLinker
}

// NewUninstaller creates a new Uninstaller instance
func NewUninstaller(baseDir string, componentLinker *linker.ComponentLinker) *Uninstaller {
	return &Uninstaller{
		baseDir: baseDir,
		linker:  componentLinker,
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

	// Get component directory path
	componentDir := filepath.Join(u.baseDir, componentType, name)

	// Display what will be removed
	fmt.Printf("\nRemoving %s: %s\n", componentType, name)

	// Show source information if available
	if entry != nil {
		if entry.SourceUrl != "" {
			fmt.Printf("  Source: %s\n", entry.SourceUrl)
		} else if entry.Source != "" {
			fmt.Printf("  Source: %s\n", entry.Source)
		}
		if entry.CommitHash != "" {
			fmt.Printf("  Commit: %s\n", entry.CommitHash)
		}
	}

	fmt.Printf("  Directory: %s\n", componentDir)

	// Find and display linked targets
	linkedTargets := u.findLinkedTargets(componentType, name)
	if len(linkedTargets) > 0 {
		fmt.Printf("  Linked to: %s\n", strings.Join(linkedTargets, ", "))
	} else {
		fmt.Printf("  Linked to: (none)\n")
	}

	fmt.Println()

	// Auto-unlink component from all targets (silent if not linked)
	if u.linker != nil && len(linkedTargets) > 0 {
		// Try to unlink, but don't fail if it's not linked
		// Pass empty targetFilter to unlink from all targets
		_ = u.linker.UnlinkComponent(componentType, name, "")
	}

	// Remove component directory from filesystem
	if err := os.RemoveAll(componentDir); err != nil {
		return fmt.Errorf("failed to remove component directory: %w", err)
	}

	// Remove entry from lock file
	if err := metadata.RemoveLockFileEntry(u.baseDir, componentType, name); err != nil {
		// Log warning but continue - directory is already removed
		fmt.Printf("Warning: Could not update lock file: %v\n", err)
	}

	// Display success message
	fmt.Printf("✓ Removed %s: %s\n", componentType, name)

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
		fmt.Printf("No components found from %s\n", repoURL)
		return nil
	}

	// Display detailed information about what will be removed
	fmt.Printf("\nThe following will be removed:\n\n")
	fmt.Printf("Repository: %s\n", repoURL)
	fmt.Printf("Total components: %d\n\n", totalComponents)

	// Show breakdown by type with more details
	for componentType, names := range componentsByType {
		if len(names) > 0 {
			fmt.Printf("%s (%d):\n", strings.Title(componentType), len(names))
			for _, name := range names {
				// Find linked targets for this component
				linkedTargets := u.findLinkedTargets(componentType, name)

				if len(linkedTargets) > 0 {
					fmt.Printf("  • %s (linked to: %s)\n", name, strings.Join(linkedTargets, ", "))
				} else {
					fmt.Printf("  • %s\n", name)
				}
			}
			fmt.Println()
		}
	}

	// Show what directories will be deleted
	fmt.Printf("Directories to be deleted:\n")
	for componentType, names := range componentsByType {
		if len(names) > 0 {
			for _, name := range names {
				componentDir := filepath.Join(u.baseDir, componentType, name)
				fmt.Printf("  • %s\n", componentDir)
			}
		}
	}
	fmt.Println()

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
			fmt.Println("Uninstall cancelled")
			return nil
		}
	}

	// Unlink message (displayed once for all components)
	fmt.Println("Unlinking from targets...")

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

			// Remove component directory from filesystem
			componentDir := filepath.Join(u.baseDir, componentType, name)
			if err := os.RemoveAll(componentDir); err != nil {
				fmt.Printf("✗ Failed to remove %s: %s (%v)\n", componentType, name, err)
				failed++
				failedComponents = append(failedComponents, fmt.Sprintf("%s/%s", componentType, name))
				continue
			}

			// Remove entry from lock file
			if err := metadata.RemoveLockFileEntry(u.baseDir, componentType, name); err != nil {
				fmt.Printf("Warning: Could not update lock file for %s: %s\n", name, err)
			}

			// Display success message
			fmt.Printf("✓ Removed %s: %s\n", componentType, name)
			removed++
		}
	}

	// Display summary
	fmt.Println()
	if failed > 0 {
		fmt.Printf("Removed %d of %d components (%d failed)\n", removed, totalComponents, failed)
		return fmt.Errorf("failed to remove %d component(s)", failed)
	} else {
		fmt.Printf("Successfully removed %d component(s) from repository\n", removed)
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
		var lockFile metadata.ComponentLockFile
		if err := json.Unmarshal(lockData, &lockFile); err != nil {
			return nil, fmt.Errorf("failed to parse lock file for %s: %w", componentType, err)
		}

		// Get the appropriate map for this component type
		var entries map[string]metadata.ComponentLockEntry
		switch componentType {
		case "skills":
			entries = lockFile.Skills
		case "agents":
			entries = lockFile.Agents
		case "commands":
			entries = lockFile.Commands
		default:
			continue
		}

		// Find components matching the source URL
		var matchingNames []string
		for name, entry := range entries {
			if u.matchesSourceURL(entry.SourceUrl, normalizedURL) || u.matchesSourceURL(entry.Source, normalizedURL) {
				matchingNames = append(matchingNames, name)
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

	// Check common target locations
	targetPaths := map[string]string{
		"OpenCode":   filepath.Join(homeDir, "Library", "Application Support", "OpenCode", componentType),
		"ClaudeCode": filepath.Join(homeDir, "Library", "Application Support", "Claude", componentType),
	}

	for targetName, targetDir := range targetPaths {
		componentPath := filepath.Join(targetDir, componentName)
		if _, err := os.Lstat(componentPath); err == nil {
			linkedTargets = append(linkedTargets, targetName)
		}
	}

	return linkedTargets
}
