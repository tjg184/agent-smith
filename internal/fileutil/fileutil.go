// Package fileutil provides cross-platform file system utilities for agent-smith.
// This package includes functions for:
//   - Cross-platform file and directory permissions
//   - File and directory copying operations
//   - YAML frontmatter parsing from markdown files
package fileutil

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tgaines/agent-smith/internal/models"
	"gopkg.in/yaml.v3"
)

// GetCrossPlatformPermissions returns the appropriate directory permissions
// for the current operating system.
// Returns 0666 for Windows, 0755 for Unix-like systems.
func GetCrossPlatformPermissions() os.FileMode {
	if runtime.GOOS == "windows" {
		return 0666 // Windows has less granular permissions
	}
	return 0755 // Unix-like systems
}

// GetCrossPlatformFilePermissions returns the appropriate file permissions
// for the current operating system.
// Returns 0644 for all systems.
func GetCrossPlatformFilePermissions() os.FileMode {
	if runtime.GOOS == "windows" {
		return 0644 // Windows has less granular permissions
	}
	return 0644 // Unix-like systems
}

// CreateDirectoryWithPermissions creates a directory with cross-platform
// appropriate permissions. Creates all parent directories as needed.
func CreateDirectoryWithPermissions(path string) error {
	perm := GetCrossPlatformPermissions()
	return os.MkdirAll(path, perm)
}

// CreateFileWithPermissions writes data to a file with cross-platform
// appropriate permissions. Creates parent directories if needed.
func CreateFileWithPermissions(path string, data []byte) error {
	perm := GetCrossPlatformFilePermissions()
	return os.WriteFile(path, data, perm)
}

// CopyFile copies a single file from src to dst with appropriate permissions.
func CopyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return CreateFileWithPermissions(dst, data)
}

// CopyDirectoryContents recursively copies all contents from src directory to dst directory.
// Maintains the relative directory structure and copies all files with appropriate permissions.
func CopyDirectoryContents(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return CreateDirectoryWithPermissions(dstPath)
		}

		return CopyFile(path, dstPath)
	})
}

// CopyComponentFiles copies files for a detected component from a repository.
// For single-file components (non-SKILL.md, non-AGENT.md, non-COMMAND.md .md files),
// copies only the component file.
// For directory-based components, recursively copies all files and subdirectories
// in the component directory, preserving the directory structure.
func CopyComponentFiles(repoPath string, component models.DetectedComponent, dst string) error {
	componentPath := filepath.Join(repoPath, component.FilePath)
	componentDir := filepath.Dir(componentPath)

	// Check if this is a single file component
	// Directory-based components use SKILL.md, AGENT.md, or COMMAND.md
	baseName := filepath.Base(component.FilePath)
	if filepath.Ext(component.FilePath) == ".md" &&
		baseName != "SKILL.md" &&
		baseName != "AGENT.md" &&
		baseName != "COMMAND.md" {
		// Single file component - copy just this file
		fileName := filepath.Base(component.FilePath)
		return CopyFile(componentPath, filepath.Join(dst, fileName))
	}

	// Directory-based component - recursively copy all contents
	return CopyDirectoryContents(componentDir, dst)
}

// ParseFrontmatter extracts YAML frontmatter from a markdown file.
// Frontmatter must be delimited by "---" at the start of the file.
// Returns nil if no frontmatter is found (not an error).
// Returns error only if the file cannot be read.
func ParseFrontmatter(filePath string) (*models.ComponentFrontmatter, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	contentStr := string(content)

	// Check if file starts with frontmatter delimiter
	if !strings.HasPrefix(contentStr, "---\n") && !strings.HasPrefix(contentStr, "---\r\n") {
		// No frontmatter found, return nil (not an error)
		return nil, nil
	}

	// Find the closing delimiter
	lines := strings.Split(contentStr, "\n")
	var frontmatterLines []string
	foundClosing := false

	for i := 1; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r")
		if line == "---" {
			foundClosing = true
			break
		}
		frontmatterLines = append(frontmatterLines, lines[i])
	}

	if !foundClosing {
		log.Printf("Warning: Malformed frontmatter in %s (missing closing delimiter)", filePath)
		return nil, nil
	}

	// Parse YAML
	frontmatterStr := strings.Join(frontmatterLines, "\n")
	var frontmatter models.ComponentFrontmatter

	if err := yaml.Unmarshal([]byte(frontmatterStr), &frontmatter); err != nil {
		log.Printf("Warning: Failed to parse YAML frontmatter in %s: %v", filePath, err)
		return nil, nil
	}

	return &frontmatter, nil
}

// DetermineComponentName determines the component name using frontmatter or filename.
// Priority: frontmatter.name > filename (without extension)
// Special files (README.md, index.md, main.md) return empty string.
func DetermineComponentName(frontmatter *models.ComponentFrontmatter, fileName string) string {
	// Skip special files
	lowerFileName := strings.ToLower(fileName)
	if lowerFileName == "readme.md" || lowerFileName == "index.md" || lowerFileName == "main.md" {
		return ""
	}

	// Use frontmatter name if available
	if frontmatter != nil && strings.TrimSpace(frontmatter.Name) != "" {
		return strings.TrimSpace(frontmatter.Name)
	}

	// Fall back to filename without extension
	name := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	// Handle edge case: no extension or empty name
	if name == "" || name == "." {
		return ""
	}

	return name
}
