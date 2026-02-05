package downloader

import (
	"path/filepath"
	"strings"

	"github.com/tgaines/agent-smith/internal/metadata"
	"github.com/tgaines/agent-smith/internal/models"
)

// Re-export types for backward compatibility
// These now reference models.ComponentLockFile and models.ComponentEntry
type ComponentLockFile = models.ComponentLockFile
type ComponentLockEntry = models.ComponentEntry

// Re-export functions for backward compatibility
var ComputeLocalFolderHash = metadata.ComputeLocalFolderHash

// DetermineDestinationFolderName determines the destination folder name using hierarchy heuristic
// Walks up from component file directory, skipping component-type names (agents/commands/skills)
// Returns first non-component-type directory name for preserving optional hierarchy
func DetermineDestinationFolderName(componentFilePath string) string {
	componentTypeNames := []string{"skills", "agents", "commands"}

	// Get directory containing the component file
	currentDir := filepath.Dir(componentFilePath)

	// Walk up the directory tree
	for {
		dirName := filepath.Base(currentDir)

		// Check if current directory name is a component type
		isComponentType := false
		for _, typeName := range componentTypeNames {
			if dirName == typeName {
				isComponentType = true
				break
			}
		}

		// If not a component type name, use it
		if !isComponentType && dirName != "." && dirName != "" {
			return dirName
		}

		// Go up one directory
		parentDir := filepath.Dir(currentDir)

		// Check if we've reached the root
		if parentDir == currentDir || parentDir == "." || parentDir == "/" || dirName == "" {
			// Reached root, fall back to "root"
			return "root"
		}

		currentDir = parentDir
	}
}

// extractGitHubOwnerRepo extracts "owner/repo" from a GitHub URL
// Returns empty string if not a GitHub URL or parsing fails
func extractGitHubOwnerRepo(url string) string {
	// Handle various GitHub URL formats:
	// https://github.com/owner/repo
	// https://github.com/owner/repo.git
	// git@github.com:owner/repo.git
	// ssh://git@github.com/owner/repo.git

	url = strings.TrimSpace(url)

	// Handle SSH format: git@github.com:owner/repo.git
	if strings.HasPrefix(url, "git@github.com:") {
		path := strings.TrimPrefix(url, "git@github.com:")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			return parts[0] + "/" + parts[1]
		}
		return ""
	}

	// Handle HTTPS format: https://github.com/owner/repo or ssh://git@github.com/owner/repo
	if strings.Contains(url, "github.com/") {
		idx := strings.Index(url, "github.com/")
		path := url[idx+len("github.com/"):]
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			return parts[0] + "/" + parts[1]
		}
	}

	return ""
}
