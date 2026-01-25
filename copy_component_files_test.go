package main

import (
	"github.com/tgaines/agent-smith/internal/fileutil"
	"github.com/tgaines/agent-smith/internal/models"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyComponentFilesNonRecursive(t *testing.T) {
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
	//     image.png (should NOT be copied - in subdirectory)
	//   subdirectory/
	//     file.txt (should NOT be copied - in subdirectory)

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

	// Verify that ONLY files in the component directory were copied
	// SKILL.md should exist
	if _, err := os.Stat(filepath.Join(dstDir, "SKILL.md")); os.IsNotExist(err) {
		t.Errorf("SKILL.md was not copied")
	}

	// README.md should exist
	if _, err := os.Stat(filepath.Join(dstDir, "README.md")); os.IsNotExist(err) {
		t.Errorf("README.md was not copied")
	}

	// resources/ directory should NOT exist
	if _, err := os.Stat(filepath.Join(dstDir, "resources")); !os.IsNotExist(err) {
		t.Errorf("resources/ directory should not have been copied")
	}

	// subdirectory/ should NOT exist
	if _, err := os.Stat(filepath.Join(dstDir, "subdirectory")); !os.IsNotExist(err) {
		t.Errorf("subdirectory/ should not have been copied")
	}

	// Verify only 2 files in destination
	entries, err := os.ReadDir(dstDir)
	if err != nil {
		t.Fatalf("Failed to read destination directory: %v", err)
	}

	fileCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			fileCount++
		}
	}

	if fileCount != 2 {
		t.Errorf("Expected exactly 2 files in destination, got %d", fileCount)
		for _, entry := range entries {
			t.Logf("  - %s (isDir: %v)", entry.Name(), entry.IsDir())
		}
	}

	t.Logf("SUCCESS: copyComponentFiles correctly copied only files in component directory (non-recursive)")
}
