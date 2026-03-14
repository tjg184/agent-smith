package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// ProjectMarkers are the directory names that indicate a project root
var ProjectMarkers = []string{".opencode", ".claude", ".github"}

// ProjectBoundaryMarkers are files/directories that indicate a project boundary
// These are checked as fallbacks if no ProjectMarkers are found
var ProjectBoundaryMarkers = []string{
	".git",           // Git repository
	"go.mod",         // Go project
	"package.json",   // Node.js project
	"pyproject.toml", // Python project
	"Cargo.toml",     // Rust project
	"composer.json",  // PHP project
	"pom.xml",        // Java Maven project
	"build.gradle",   // Java Gradle project
	"Gemfile",        // Ruby project
	"mix.exs",        // Elixir project
}

func FindProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	return FindProjectRootFromDir(cwd)
}

func FindProjectRootFromDir(startDir string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	currentDir := startDir

	// Walk up the directory tree
	for {
		// Check for project markers (.opencode, .claude, or .github) - these are preferred
		for _, marker := range ProjectMarkers {
			markerPath := filepath.Join(currentDir, marker)
			if info, err := os.Stat(markerPath); err == nil && info.IsDir() {
				return currentDir, nil
			}
		}

		// Check for project boundary markers (files or directories)
		for _, marker := range ProjectBoundaryMarkers {
			markerPath := filepath.Join(currentDir, marker)
			if info, err := os.Stat(markerPath); err == nil {
				// Found a project boundary marker
				// If it's .git, verify it's a directory
				if marker == ".git" {
					if !info.IsDir() {
						continue
					}
				}
				// Return immediately - this is our project root
				return currentDir, nil
			}
		}

		// Check if we've reached the home directory or filesystem root
		parentDir := filepath.Dir(currentDir)
		if currentDir == parentDir || currentDir == homeDir || currentDir == filepath.VolumeName(currentDir)+string(filepath.Separator) {
			break
		}

		currentDir = parentDir
	}

	return "", fmt.Errorf("no project boundary detected\n\n" +
		"agent-smith looks for project markers to determine where to materialize components.\n\n" +
		"Supported project markers:\n" +
		"  • .opencode/    (preferred - agent-smith project)\n" +
		"  • .claude/      (preferred - Claude project)\n" +
		"  • .github/      (preferred - GitHub Copilot project)\n" +
		"  • .agents/      (universal - target-agnostic storage)\n" +
		"  • .git/         (version control)\n" +
		"  • go.mod        (Go projects)\n" +
		"  • package.json  (Node.js projects)\n" +
		"  • pyproject.toml (Python projects)\n" +
		"  • Cargo.toml    (Rust projects)\n" +
		"  • composer.json (PHP projects)\n" +
		"  • pom.xml       (Java Maven projects)\n" +
		"  • build.gradle  (Java Gradle projects)\n" +
		"  • Gemfile       (Ruby projects)\n" +
		"  • mix.exs       (Elixir projects)\n\n" +
		"To fix this:\n" +
		"  1. Create a project marker: mkdir -p .opencode/\n" +
		"  2. Or use --project-dir flag: agent-smith materialize --project-dir /path/to/project\n" +
		"  3. Or initialize version control: git init")
}

func EnsureTargetStructure(targetDir string) (bool, error) {
	created := false

	// Check if target directory exists
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		created = true
	}

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

func EnsureComponentDirectory(targetDir, componentType string) (bool, error) {
	created := false

	// Check if target directory exists
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		created = true
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return false, fmt.Errorf("failed to create target directory: %w", err)
	}

	// Create only the specific component subdirectory
	subdirPath := filepath.Join(targetDir, componentType)
	if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
		created = true
	}
	if err := os.MkdirAll(subdirPath, 0755); err != nil {
		return false, fmt.Errorf("failed to create subdirectory %s: %w", componentType, err)
	}

	return created, nil
}
