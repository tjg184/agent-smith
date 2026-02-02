package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureTargetStructure_CreatesNewStructure(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "agent-smith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	targetDir := filepath.Join(tempDir, ".opencode")

	// Call EnsureTargetStructure - should create everything
	created, err := EnsureTargetStructure(targetDir)
	if err != nil {
		t.Fatalf("EnsureTargetStructure failed: %v", err)
	}

	// Should return true since directories were created
	if !created {
		t.Errorf("Expected created=true for new structure, got false")
	}

	// Verify target directory exists
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		t.Errorf("Target directory was not created")
	}

	// Verify subdirectories exist
	subdirs := []string{"skills", "agents", "commands"}
	for _, subdir := range subdirs {
		subdirPath := filepath.Join(targetDir, subdir)
		if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
			t.Errorf("Subdirectory %s was not created", subdir)
		}
	}
}

func TestEnsureTargetStructure_ExistingStructure(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "agent-smith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	targetDir := filepath.Join(tempDir, ".claude")

	// Pre-create the structure
	subdirs := []string{"skills", "agents", "commands"}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("Failed to pre-create target directory: %v", err)
	}
	for _, subdir := range subdirs {
		subdirPath := filepath.Join(targetDir, subdir)
		if err := os.MkdirAll(subdirPath, 0755); err != nil {
			t.Fatalf("Failed to pre-create subdirectory %s: %v", subdir, err)
		}
	}

	// Call EnsureTargetStructure - structure already exists
	created, err := EnsureTargetStructure(targetDir)
	if err != nil {
		t.Fatalf("EnsureTargetStructure failed: %v", err)
	}

	// Should return false since directories already existed
	if created {
		t.Errorf("Expected created=false for existing structure, got true")
	}
}

func TestEnsureTargetStructure_PartialStructure(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "agent-smith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	targetDir := filepath.Join(tempDir, ".opencode")

	// Pre-create only the target directory, but not subdirectories
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("Failed to pre-create target directory: %v", err)
	}

	// Pre-create only some subdirectories
	if err := os.MkdirAll(filepath.Join(targetDir, "skills"), 0755); err != nil {
		t.Fatalf("Failed to pre-create skills directory: %v", err)
	}

	// Call EnsureTargetStructure - missing subdirectories should be created
	created, err := EnsureTargetStructure(targetDir)
	if err != nil {
		t.Fatalf("EnsureTargetStructure failed: %v", err)
	}

	// Should return true since some directories needed to be created
	if !created {
		t.Errorf("Expected created=true for partial structure, got false")
	}

	// Verify all subdirectories now exist
	subdirs := []string{"skills", "agents", "commands"}
	for _, subdir := range subdirs {
		subdirPath := filepath.Join(targetDir, subdir)
		if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
			t.Errorf("Subdirectory %s was not created", subdir)
		}
	}
}
