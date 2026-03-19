package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjg184/agent-smith/pkg/config"
)

// ProjectMarkers are the directory names that indicate a project root.
// Derived from the built-in target registry so adding a new target automatically
// extends detection without touching this file.
var ProjectMarkers = config.GetAllTargetProjectDirNames()

// ProjectBoundaryMarkers are files/directories that indicate a project boundary
// checked as fallbacks if no ProjectMarkers are found.
var ProjectBoundaryMarkers = []string{
	".git",
	"go.mod",
	"package.json",
	"pyproject.toml",
	"Cargo.toml",
	"composer.json",
	"pom.xml",
	"build.gradle",
	"Gemfile",
	"mix.exs",
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

	for {
		for _, marker := range ProjectMarkers {
			markerPath := filepath.Join(currentDir, marker)
			if info, err := os.Stat(markerPath); err == nil && info.IsDir() {
				return currentDir, nil
			}
		}

		for _, marker := range ProjectBoundaryMarkers {
			markerPath := filepath.Join(currentDir, marker)
			if info, err := os.Stat(markerPath); err == nil {
				if marker == ".git" && !info.IsDir() {
					continue
				}
				return currentDir, nil
			}
		}

		parentDir := filepath.Dir(currentDir)
		if currentDir == parentDir || currentDir == homeDir || currentDir == filepath.VolumeName(currentDir)+string(filepath.Separator) {
			break
		}

		currentDir = parentDir
	}

	var markerLines strings.Builder
	for _, m := range ProjectMarkers {
		markerLines.WriteString(fmt.Sprintf("  • %s/\n", m))
	}
	for _, m := range ProjectBoundaryMarkers {
		if m == ".git" {
			markerLines.WriteString(fmt.Sprintf("  • %s/\n", m))
		} else {
			markerLines.WriteString(fmt.Sprintf("  • %s\n", m))
		}
	}

	msg := "no project boundary detected\n\n" +
		"agent-smith looks for project markers to determine where to materialize components.\n\n" +
		"Supported project markers:\n" +
		markerLines.String() + "\n" +
		"To fix this:\n" +
		"  1. Create a project marker: mkdir -p .opencode/\n" +
		"  2. Or use --project-dir flag: agent-smith materialize --project-dir /path/to/project\n" +
		"  3. Or initialize version control: git init"
	return "", errors.New(msg)
}

func EnsureTargetStructure(targetDir string) (bool, error) {
	created := false

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		created = true
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return false, fmt.Errorf("failed to create target directory: %w", err)
	}

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

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		created = true
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return false, fmt.Errorf("failed to create target directory: %w", err)
	}

	subdirPath := filepath.Join(targetDir, componentType)
	if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
		created = true
	}
	if err := os.MkdirAll(subdirPath, 0755); err != nil {
		return false, fmt.Errorf("failed to create subdirectory %s: %w", componentType, err)
	}

	return created, nil
}
