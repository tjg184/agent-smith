package fileutil_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tgaines/agent-smith/internal/fileutil"
	"github.com/tgaines/agent-smith/internal/models"
)

func TestCopyComponentFilesRecursive(t *testing.T) {
	// Create temporary directories
	srcDir, err := os.MkdirTemp("", "src-*")
	if err != nil {
		t.Fatalf("Failed to create temp src directory: %v", err)
	}
	defer os.RemoveAll(srcDir)

	dstDir, err := os.MkdirTemp("", "dst-*")
	if err != nil {
		t.Fatalf("Failed to create temp dst directory: %v", err)
	}
	defer os.RemoveAll(dstDir)

	// Create source structure with files and subdirectories
	// Component directory structure:
	// src/
	//   SKILL.md (should be copied)
	//   README.md (should be copied)
	//   resources/
	//     image.png (should be copied - in subdirectory)
	//   subdirectory/
	//     file.txt (should be copied - in subdirectory)

	skillFile := filepath.Join(srcDir, "SKILL.md")
	readmeFile := filepath.Join(srcDir, "README.md")
	resourcesDir := filepath.Join(srcDir, "resources")
	subdirDir := filepath.Join(srcDir, "subdirectory")
	imageFile := filepath.Join(resourcesDir, "image.png")
	nestedFile := filepath.Join(subdirDir, "file.txt")

	// Create files
	if err := os.WriteFile(skillFile, []byte("# Skill"), 0644); err != nil {
		t.Fatalf("Failed to create SKILL.md: %v", err)
	}
	if err := os.WriteFile(readmeFile, []byte("# README"), 0644); err != nil {
		t.Fatalf("Failed to create README.md: %v", err)
	}

	// Create subdirectories with files
	if err := os.MkdirAll(resourcesDir, 0755); err != nil {
		t.Fatalf("Failed to create resources directory: %v", err)
	}
	if err := os.WriteFile(imageFile, []byte("image data"), 0644); err != nil {
		t.Fatalf("Failed to create image.png: %v", err)
	}

	if err := os.MkdirAll(subdirDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	if err := os.WriteFile(nestedFile, []byte("nested content"), 0644); err != nil {
		t.Fatalf("Failed to create file.txt: %v", err)
	}

	// Create a models.DetectedComponent with FilePath pointing to SKILL.md
	component := models.DetectedComponent{
		Type:       models.ComponentSkill,
		Name:       "test-skill",
		Path:       ".",
		SourceFile: "SKILL.md",
		FilePath:   "SKILL.md",
	}

	// Copy component files using fileutil directly
	err = fileutil.CopyComponentFiles(srcDir, component, dstDir)
	if err != nil {
		t.Fatalf("copyComponentFiles failed: %v", err)
	}

	// Verify that ALL files including subdirectories were copied
	// SKILL.md should exist
	if _, err := os.Stat(filepath.Join(dstDir, "SKILL.md")); os.IsNotExist(err) {
		t.Errorf("SKILL.md was not copied")
	}

	// README.md should exist
	if _, err := os.Stat(filepath.Join(dstDir, "README.md")); os.IsNotExist(err) {
		t.Errorf("README.md was not copied")
	}

	// resources/ directory should exist
	if _, err := os.Stat(filepath.Join(dstDir, "resources")); os.IsNotExist(err) {
		t.Errorf("resources/ directory was not copied")
	}

	// resources/image.png should exist
	if _, err := os.Stat(filepath.Join(dstDir, "resources", "image.png")); os.IsNotExist(err) {
		t.Errorf("resources/image.png was not copied")
	}

	// subdirectory/ should exist
	if _, err := os.Stat(filepath.Join(dstDir, "subdirectory")); os.IsNotExist(err) {
		t.Errorf("subdirectory/ was not copied")
	}

	// subdirectory/file.txt should exist
	if _, err := os.Stat(filepath.Join(dstDir, "subdirectory", "file.txt")); os.IsNotExist(err) {
		t.Errorf("subdirectory/file.txt was not copied")
	}

	// Verify content of nested files
	imageContent, err := os.ReadFile(filepath.Join(dstDir, "resources", "image.png"))
	if err != nil {
		t.Errorf("Failed to read resources/image.png: %v", err)
	} else if string(imageContent) != "image data" {
		t.Errorf("Expected 'image data' in resources/image.png, got '%s'", string(imageContent))
	}

	nestedContent, err := os.ReadFile(filepath.Join(dstDir, "subdirectory", "file.txt"))
	if err != nil {
		t.Errorf("Failed to read subdirectory/file.txt: %v", err)
	} else if string(nestedContent) != "nested content" {
		t.Errorf("Expected 'nested content' in subdirectory/file.txt, got '%s'", string(nestedContent))
	}

	t.Logf("SUCCESS: CopyComponentFiles correctly copied all files and subdirectories recursively")
}
