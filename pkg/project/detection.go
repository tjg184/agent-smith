package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// ProjectMarkers are the directory names that indicate a project root
var ProjectMarkers = []string{".opencode", ".claude"}

// FindProjectRoot walks up the directory tree from the current working directory
// looking for project markers (.opencode or .claude directories).
// Returns the project root path or an error if no project is found.
func FindProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	return FindProjectRootFromDir(cwd)
}

// FindProjectRootFromDir walks up the directory tree from the specified directory
// looking for project markers (.opencode or .claude directories) or .git directory.
// Returns the project root path or an error if no project boundary is found.
// Stops at .git/ directory (project root), home directory, or filesystem root.
func FindProjectRootFromDir(startDir string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	currentDir := startDir

	// Walk up the directory tree
	for {
		// Check for project markers (.opencode or .claude)
		for _, marker := range ProjectMarkers {
			markerPath := filepath.Join(currentDir, marker)
			if info, err := os.Stat(markerPath); err == nil && info.IsDir() {
				return currentDir, nil
			}
		}

		// Check if we've reached a .git/ directory (project root boundary)
		// If we find .git, this is the project root even without .opencode/.claude
		gitPath := filepath.Join(currentDir, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			// We're at a git project root, return this as the project root
			return currentDir, nil
		}

		// Check if we've reached the home directory or filesystem root
		parentDir := filepath.Dir(currentDir)
		if currentDir == parentDir || currentDir == homeDir || currentDir == filepath.VolumeName(currentDir)+string(filepath.Separator) {
			break
		}

		currentDir = parentDir
	}

	return "", fmt.Errorf("no project found (.opencode/, .claude/, or .git/ directory not found)\n\nTo materialize components to a project:\n  1. Create a project directory: mkdir -p .opencode/\n  2. Run the materialize command from within the project")
}

// GetTargetDirectory returns the target directory path for a given target name
// (opencode or claudecode) within the project root.
func GetTargetDirectory(projectRoot, targetName string) string {
	switch targetName {
	case "opencode":
		return filepath.Join(projectRoot, ".opencode")
	case "claudecode":
		return filepath.Join(projectRoot, ".claude")
	default:
		return ""
	}
}

// EnsureTargetStructure creates the target directory structure if it doesn't exist.
// Creates the target directory and subdirectories: skills/, agents/, commands/
// Returns true if any directories were created (structure was initialized), false if all existed.
func EnsureTargetStructure(targetDir string) (bool, error) {
	created := false

	// Check if target directory exists
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		created = true
	}

	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return false, fmt.Errorf("failed to create target directory: %w", err)
	}

	// Create subdirectories
	subdirs := []string{"skills", "agents", "commands"}
	for _, subdir := range subdirs {
		subdirPath := filepath.Join(targetDir, subdir)
		if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
			created = true
		}
		if err := os.MkdirAll(subdirPath, 0755); err != nil {
			return false, fmt.Errorf("failed to create subdirectory %s: %w", subdir, err)
		}
	}

	return created, nil
}
