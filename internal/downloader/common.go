package downloader

import (
	"path/filepath"

	"github.com/tgaines/agent-smith/internal/metadata"
)

// Re-export types for backward compatibility
type ComponentLockFile = metadata.ComponentLockFile
type ComponentLockEntry = metadata.ComponentLockEntry

// Re-export functions for backward compatibility
var ComputeGitHubTreeSHA = metadata.ComputeGitHubTreeSHA
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
