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

	"github.com/tjg184/agent-smith/internal/models"
	"gopkg.in/yaml.v3"
)

// GetCrossPlatformPermissions returns the appropriate directory permissions
// for the current operating system.
func GetCrossPlatformPermissions() os.FileMode {
	if runtime.GOOS == "windows" {
		return 0666 // Windows has less granular permissions
	}
	return 0755 // Unix-like systems
}

// GetCrossPlatformFilePermissions returns the appropriate file permissions
// for the current operating system.
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
		if os.IsNotExist(err) {
			return fmt.Errorf("cannot copy file: source file does not exist: %s", src)
		}
		if os.IsPermission(err) {
			return fmt.Errorf("cannot copy file: permission denied reading source file: %s", src)
		}
		return fmt.Errorf("cannot copy file: failed to read source file %s: %w", src, err)
	}

	if err := CreateFileWithPermissions(dst, data); err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("cannot copy file: permission denied writing to destination: %s", dst)
		}
		// Check if it's a directory write error (parent directory doesn't exist)
		if os.IsNotExist(err) {
			return fmt.Errorf("cannot copy file: destination directory does not exist: %s", filepath.Dir(dst))
		}
		return fmt.Errorf("cannot copy file: failed to write to destination %s: %w", dst, err)
	}

	return nil
}

// CopyDirectoryContents recursively copies all contents from src directory to dst directory.
// Maintains the relative directory structure and copies all files with appropriate permissions.
func CopyDirectoryContents(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("cannot copy directory: source directory does not exist: %s", src)
		}
		if os.IsPermission(err) {
			return fmt.Errorf("cannot copy directory: permission denied accessing source directory: %s", src)
		}
		return fmt.Errorf("cannot copy directory: failed to access source directory %s: %w", src, err)
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("cannot copy directory: source is not a directory: %s", src)
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("cannot copy directory: error walking path %s: %w", path, err)
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("cannot copy directory: failed to determine relative path for %s: %w", path, err)
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			if err := CreateDirectoryWithPermissions(dstPath); err != nil {
				return fmt.Errorf("cannot copy directory: failed to create destination directory %s: %w", dstPath, err)
			}
			return nil
		}

		if err := CopyFile(path, dstPath); err != nil {
			// CopyFile already provides detailed error messages
			return err
		}

		return nil
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

	if _, err := os.Stat(componentPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("cannot copy component '%s': component file does not exist at %s", component.Name, componentPath)
		}
		return fmt.Errorf("cannot copy component '%s': failed to access component file %s: %w", component.Name, componentPath, err)
	}

	baseName := filepath.Base(component.FilePath)
	// Directory-based components use SKILL.md, AGENT.md, or COMMAND.md
	if filepath.Ext(component.FilePath) == ".md" &&
		baseName != "SKILL.md" &&
		baseName != "AGENT.md" &&
		baseName != "COMMAND.md" {
		fileName := filepath.Base(component.FilePath)
		dstFilePath := filepath.Join(dst, fileName)
		if err := CopyFile(componentPath, dstFilePath); err != nil {
			return fmt.Errorf("cannot copy component '%s': %w", component.Name, err)
		}
		return nil
	}

	if err := CopyDirectoryContents(componentDir, dst); err != nil {
		return fmt.Errorf("cannot copy component '%s': %w", component.Name, err)
	}
	return nil
}

func ParseFrontmatter(filePath string) (*models.ComponentFrontmatter, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	contentStr := string(content)

	if !strings.HasPrefix(contentStr, "---\n") && !strings.HasPrefix(contentStr, "---\r\n") {
		return nil, nil
	}

	lines := strings.Split(contentStr, "\n")
	var frontmatterLines []string
	foundClosing := false

	for i := 1; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r")
		if line == "---" {
			foundClosing = true
			break
		}
		frontmatterLines = append(frontmatterLines, line)
	}

	if !foundClosing {
		log.Printf("Warning: Malformed frontmatter in %s (missing closing delimiter)", filePath)
		return nil, nil
	}

	frontmatterStr := strings.Join(frontmatterLines, "\n")
	var frontmatter models.ComponentFrontmatter

	if err := yaml.Unmarshal([]byte(frontmatterStr), &frontmatter); err != nil {
		// Retry with unquoted values auto-quoted — SKILL.md authors often write
		// prose descriptions containing ": " without YAML quoting.
		quoted := quoteUnquotedValues(frontmatterStr)
		if quoted != frontmatterStr {
			if retryErr := yaml.Unmarshal([]byte(quoted), &frontmatter); retryErr == nil {
				return &frontmatter, nil
			}
		}
		log.Printf("Warning: Failed to parse YAML frontmatter in %s: %v", filePath, err)
		return nil, nil
	}

	return &frontmatter, nil
}

// quoteUnquotedValues wraps scalar values containing ": " in double quotes so
// prose descriptions in SKILL.md frontmatter parse without YAML errors.
func quoteUnquotedValues(src string) string {
	lines := strings.Split(src, "\n")
	for i, line := range lines {
		colonIdx := strings.Index(line, ": ")
		if colonIdx < 0 {
			continue
		}
		key := line[:colonIdx]
		value := line[colonIdx+2:]
		if strings.HasPrefix(value, `"`) || strings.HasPrefix(value, `'`) {
			continue
		}
		if !strings.Contains(value, ": ") {
			continue
		}
		escaped := strings.ReplaceAll(value, `"`, `\"`)
		lines[i] = key + `: "` + escaped + `"`
	}
	return strings.Join(lines, "\n")
}

// DetermineComponentName determines the component name using frontmatter or filename.// Priority: frontmatter.name > filename (without extension)
// Special files (README.md, index.md, main.md) return empty string.
func DetermineComponentName(frontmatter *models.ComponentFrontmatter, fileName string) string {
	lowerFileName := strings.ToLower(fileName)
	if lowerFileName == "readme.md" || lowerFileName == "index.md" || lowerFileName == "main.md" {
		return ""
	}

	if frontmatter != nil && strings.TrimSpace(frontmatter.Name) != "" {
		return strings.TrimSpace(frontmatter.Name)
	}

	name := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	if name == "" || name == "." {
		return ""
	}

	return name
}
