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
	_, err := metadata.LoadLockFileEntry(u.baseDir, componentType, name)
	if err != nil {
		return fmt.Errorf("component '%s' not installed", name)
	}

	// Auto-unlink component from all targets (silent if not linked)
	if u.linker != nil {
		// Try to unlink, but don't fail if it's not linked
		_ = u.linker.UnlinkComponent(componentType, name)
	}

	// Remove component directory from filesystem
	componentDir := filepath.Join(u.baseDir, componentType, name)
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

	// Display what will be removed
	fmt.Printf("Found %d component(s) from %s:\n", totalComponents, repoURL)
	for componentType, names := range componentsByType {
		if len(names) > 0 {
			fmt.Printf("  %s (%d): %s\n", componentType, len(names), strings.Join(names, ", "))
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
	if u.linker != nil {
		fmt.Println("Unlinking from targets...")
	}

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
				_ = u.linker.UnlinkComponent(componentType, name)
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
		fmt.Printf("Removed %d component(s) from repository\n", removed)
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
