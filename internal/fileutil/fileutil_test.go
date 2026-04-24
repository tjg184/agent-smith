package fileutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/tjg184/agent-smith/internal/models"
)

// TestGetCrossPlatformPermissions tests cross-platform directory permissions
func TestGetCrossPlatformPermissions(t *testing.T) {
	perm := GetCrossPlatformPermissions()

	if runtime.GOOS == "windows" {
		if perm != 0666 {
			t.Errorf("expected 0666 on Windows, got %o", perm)
		}
	} else {
		if perm != 0755 {
			t.Errorf("expected 0755 on Unix-like systems, got %o", perm)
		}
	}
}

// TestGetCrossPlatformFilePermissions tests cross-platform file permissions
func TestGetCrossPlatformFilePermissions(t *testing.T) {
	perm := GetCrossPlatformFilePermissions()

	// Both Windows and Unix-like systems should return 0644
	if perm != 0644 {
		t.Errorf("expected 0644, got %o", perm)
	}
}

// TestCreateDirectoryWithPermissions tests directory creation with proper permissions
func TestCreateDirectoryWithPermissions(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "test", "nested", "dirs")

	err := CreateDirectoryWithPermissions(testDir)
	if err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	// Verify directory exists
	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("directory does not exist: %v", err)
	}

	if !info.IsDir() {
		t.Error("expected path to be a directory")
	}

	// On Unix-like systems, verify permissions
	if runtime.GOOS != "windows" {
		expectedPerm := os.FileMode(0755)
		if info.Mode().Perm() != expectedPerm {
			t.Errorf("expected permissions %o, got %o", expectedPerm, info.Mode().Perm())
		}
	}
}

// TestCreateFileWithPermissions tests file creation with proper permissions
func TestCreateFileWithPermissions(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "file.txt")
	testData := []byte("test content")

	err := CreateFileWithPermissions(testFile, testData)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Verify file exists and has correct content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(content) != string(testData) {
		t.Errorf("expected content %q, got %q", testData, content)
	}

	// Verify permissions
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	if runtime.GOOS != "windows" {
		expectedPerm := os.FileMode(0644)
		if info.Mode().Perm() != expectedPerm {
			t.Errorf("expected permissions %o, got %o", expectedPerm, info.Mode().Perm())
		}
	}
}

// TestCopyFile tests single file copying
func TestCopyFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(tempDir, "source.txt")
	srcData := []byte("file content to copy")
	if err := os.WriteFile(srcFile, srcData, 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Copy file
	dstFile := filepath.Join(tempDir, "destination.txt")
	err := CopyFile(srcFile, dstFile)
	if err != nil {
		t.Fatalf("failed to copy file: %v", err)
	}

	// Verify destination file exists and has same content
	dstData, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}

	if string(dstData) != string(srcData) {
		t.Errorf("expected content %q, got %q", srcData, dstData)
	}
}

// TestCopyFileNonExistent tests copying non-existent file
func TestCopyFileNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "nonexistent.txt")
	dstFile := filepath.Join(tempDir, "destination.txt")

	err := CopyFile(srcFile, dstFile)
	if err == nil {
		t.Error("expected error when copying non-existent file")
	}
}

// TestCopyDirectoryContents tests recursive directory copying
func TestCopyDirectoryContents(t *testing.T) {
	tempDir := t.TempDir()

	// Create source directory structure
	srcDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755); err != nil {
		t.Fatalf("failed to create source directory: %v", err)
	}

	// Create test files
	files := map[string]string{
		"file1.txt":        "content 1",
		"file2.txt":        "content 2",
		"subdir/file3.txt": "content 3",
	}

	for path, content := range files {
		filePath := filepath.Join(srcDir, path)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", path, err)
		}
	}

	// Copy directory contents
	dstDir := filepath.Join(tempDir, "destination")
	err := CopyDirectoryContents(srcDir, dstDir)
	if err != nil {
		t.Fatalf("failed to copy directory contents: %v", err)
	}

	// Verify all files were copied
	for path, expectedContent := range files {
		dstPath := filepath.Join(dstDir, path)
		content, err := os.ReadFile(dstPath)
		if err != nil {
			t.Errorf("failed to read copied file %s: %v", path, err)
			continue
		}

		if string(content) != expectedContent {
			t.Errorf("file %s: expected content %q, got %q", path, expectedContent, content)
		}
	}

	// Verify directory structure
	subdirInfo, err := os.Stat(filepath.Join(dstDir, "subdir"))
	if err != nil {
		t.Error("subdirectory was not copied")
	} else if !subdirInfo.IsDir() {
		t.Error("subdir should be a directory")
	}
}

// TestCopyComponentFilesSingleFile tests copying single-file components
func TestCopyComponentFilesSingleFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create repo structure
	repoPath := filepath.Join(tempDir, "repo")
	componentPath := filepath.Join(repoPath, "components")
	if err := os.MkdirAll(componentPath, 0755); err != nil {
		t.Fatalf("failed to create repo directory: %v", err)
	}

	// Create a single-file component (non-SKILL/AGENT/COMMAND.md)
	componentFile := filepath.Join(componentPath, "example.md")
	componentContent := []byte("# Example Component")
	if err := os.WriteFile(componentFile, componentContent, 0644); err != nil {
		t.Fatalf("failed to create component file: %v", err)
	}

	// Create component
	component := models.DetectedComponent{
		Type:     models.ComponentSkill,
		Name:     "example",
		FilePath: "components/example.md",
	}

	// Copy component
	dstDir := filepath.Join(tempDir, "dest")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("failed to create destination directory: %v", err)
	}

	err := CopyComponentFiles(repoPath, component, dstDir)
	if err != nil {
		t.Fatalf("failed to copy component files: %v", err)
	}

	// Verify only the single file was copied
	copiedFile := filepath.Join(dstDir, "example.md")
	content, err := os.ReadFile(copiedFile)
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}

	if string(content) != string(componentContent) {
		t.Errorf("expected content %q, got %q", componentContent, content)
	}
}

// TestCopyComponentFilesDirectory tests copying directory-based components
func TestCopyComponentFilesDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Create repo structure with directory-based component
	repoPath := filepath.Join(tempDir, "repo")
	componentDir := filepath.Join(repoPath, "agents", "example")
	if err := os.MkdirAll(componentDir, 0755); err != nil {
		t.Fatalf("failed to create component directory: %v", err)
	}

	// Create AGENT.md and supporting files
	files := map[string]string{
		"AGENT.md":    "# Agent",
		"config.json": "{}",
		"utils.js":    "// utils",
	}

	for filename, content := range files {
		filePath := filepath.Join(componentDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", filename, err)
		}
	}

	// Create component
	component := models.DetectedComponent{
		Type:     models.ComponentAgent,
		Name:     "example",
		FilePath: "agents/example/AGENT.md",
	}

	// Copy component
	dstDir := filepath.Join(tempDir, "dest")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("failed to create destination directory: %v", err)
	}

	err := CopyComponentFiles(repoPath, component, dstDir)
	if err != nil {
		t.Fatalf("failed to copy component files: %v", err)
	}

	// Verify all files were copied
	for filename, expectedContent := range files {
		copiedFile := filepath.Join(dstDir, filename)
		content, err := os.ReadFile(copiedFile)
		if err != nil {
			t.Errorf("failed to read copied file %s: %v", filename, err)
			continue
		}

		if string(content) != expectedContent {
			t.Errorf("file %s: expected content %q, got %q", filename, expectedContent, content)
		}
	}
}

// TestParseFrontmatter tests YAML frontmatter parsing
func TestParseFrontmatter(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		content        string
		expectedName   string
		expectedDesc   string
		shouldBeNil    bool
		shouldHaveData bool
	}{
		{
			name: "Valid frontmatter",
			content: `---
name: test-component
description: Test Description
model: gpt-4
---
# Content`,
			expectedName:   "test-component",
			expectedDesc:   "Test Description",
			shouldBeNil:    false,
			shouldHaveData: true,
		},
		{
			name: "No frontmatter",
			content: `# Just a markdown file
No frontmatter here`,
			shouldBeNil:    true,
			shouldHaveData: false,
		},
		{
			name: "Empty frontmatter",
			content: `---
---
# Content`,
			shouldBeNil:    false,
			shouldHaveData: false,
		},
		{
			name: "Malformed frontmatter (no closing)",
			content: `---
name: test
# Content without closing delimiter`,
			shouldBeNil:    true,
			shouldHaveData: false,
		},
		{
			name: "Description with unquoted colons (SKILL.md style)",
			content: `---
name: quick-execute
description: Lightweight pipeline. Triggers on: quick feature, simple change, skip the pipeline, quick-execute, fast-track.
---
# Content`,
			expectedName:   "quick-execute",
			expectedDesc:   "Lightweight pipeline. Triggers on: quick feature, simple change, skip the pipeline, quick-execute, fast-track.",
			shouldBeNil:    false,
			shouldHaveData: true,
		},
		{
			name: "CRLF line endings",
			content: "---\r\nname: test-component\r\ndescription: Test description\r\n---\r\n# Content",
			expectedName:   "test-component",
			expectedDesc:   "Test description",
			shouldBeNil:    false,
			shouldHaveData: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tempDir, "test.md")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			// Parse frontmatter
			frontmatter, err := ParseFrontmatter(testFile)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.shouldBeNil && frontmatter != nil {
				t.Error("expected nil frontmatter")
			}

			if !tt.shouldBeNil && frontmatter == nil && tt.shouldHaveData {
				t.Error("expected non-nil frontmatter")
			}

			if tt.shouldHaveData && frontmatter != nil {
				if frontmatter.Name != tt.expectedName {
					t.Errorf("expected name %q, got %q", tt.expectedName, frontmatter.Name)
				}
				if frontmatter.Description != tt.expectedDesc {
					t.Errorf("expected description %q, got %q", tt.expectedDesc, frontmatter.Description)
				}
			}

			// Clean up
			os.Remove(testFile)
		})
	}
}

// TestParseFrontmatterNonExistent tests parsing non-existent file
func TestParseFrontmatterNonExistent(t *testing.T) {
	_, err := ParseFrontmatter("/nonexistent/file.md")
	if err == nil {
		t.Error("expected error when parsing non-existent file")
	}
}

// TestDetermineComponentName tests component name determination
func TestDetermineComponentName(t *testing.T) {
	tests := []struct {
		name         string
		frontmatter  *models.ComponentFrontmatter
		fileName     string
		expectedName string
	}{
		{
			name: "Frontmatter name takes precedence",
			frontmatter: &models.ComponentFrontmatter{
				Name: "custom-name",
			},
			fileName:     "filename.md",
			expectedName: "custom-name",
		},
		{
			name:         "Use filename when no frontmatter",
			frontmatter:  nil,
			fileName:     "my-component.md",
			expectedName: "my-component",
		},
		{
			name: "Use filename when frontmatter name is empty",
			frontmatter: &models.ComponentFrontmatter{
				Name: "",
			},
			fileName:     "fallback.md",
			expectedName: "fallback",
		},
		{
			name: "Use filename when frontmatter name is whitespace",
			frontmatter: &models.ComponentFrontmatter{
				Name: "   ",
			},
			fileName:     "component.md",
			expectedName: "component",
		},
		{
			name:         "Skip README.md",
			frontmatter:  nil,
			fileName:     "README.md",
			expectedName: "",
		},
		{
			name:         "Skip index.md",
			frontmatter:  nil,
			fileName:     "index.md",
			expectedName: "",
		},
		{
			name:         "Skip main.md",
			frontmatter:  nil,
			fileName:     "main.md",
			expectedName: "",
		},
		{
			name:         "Case insensitive special files",
			frontmatter:  nil,
			fileName:     "ReAdMe.Md",
			expectedName: "",
		},
		{
			name:         "Trim frontmatter name whitespace",
			frontmatter:  &models.ComponentFrontmatter{Name: "  trimmed  "},
			fileName:     "test.md",
			expectedName: "trimmed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineComponentName(tt.frontmatter, tt.fileName)
			if result != tt.expectedName {
				t.Errorf("expected %q, got %q", tt.expectedName, result)
			}
		})
	}
}

// TestDetermineComponentNameEdgeCases tests edge cases for component name determination
func TestDetermineComponentNameEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		fileName     string
		expectedName string
	}{
		{
			name:         "File without extension",
			fileName:     "component",
			expectedName: "component",
		},
		{
			name:         "File with multiple dots",
			fileName:     "my.component.md",
			expectedName: "my.component",
		},
		{
			name:         "Hidden file",
			fileName:     ".hidden.md",
			expectedName: ".hidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineComponentName(nil, tt.fileName)
			if result != tt.expectedName {
				t.Errorf("expected %q, got %q", tt.expectedName, result)
			}
		})
	}
}
