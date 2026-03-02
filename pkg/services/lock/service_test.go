package lock

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tjg184/agent-smith/internal/models"
	"github.com/tjg184/agent-smith/pkg/logger"
)

// TestLoadEntry tests loading a component entry
func TestLoadEntry(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "lock-service-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create service
	log := logger.New(logger.LevelError) // Use error level to suppress debug output in tests
	service := NewService(log)

	// Test loading non-existent entry
	_, err = service.LoadEntry(tmpDir, "skills", "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent entry")
	}

	// Test validation
	_, err = service.LoadEntry("", "skills", "test")
	if err == nil {
		t.Error("Expected error for empty baseDir")
	}

	_, err = service.LoadEntry(tmpDir, "", "test")
	if err == nil {
		t.Error("Expected error for empty componentType")
	}

	_, err = service.LoadEntry(tmpDir, "skills", "")
	if err == nil {
		t.Error("Expected error for empty componentName")
	}
}

// TestSaveAndLoadEntry tests saving and loading a component entry
func TestSaveAndLoadEntry(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "lock-service-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create service
	log := logger.New(logger.LevelError) // Use error level to suppress debug output in tests
	service := NewService(log)

	// Create test entry
	entry := &models.ComponentEntry{
		Source:       "github",
		SourceType:   "git",
		SourceUrl:    "https://github.com/test/repo",
		CommitHash:   "abc123",
		OriginalPath: "skills/test",
		Version:      5,
	}

	// Save entry
	err = service.SaveEntry(tmpDir, "skills", "test-skill", entry)
	if err != nil {
		t.Fatalf("Failed to save entry: %v", err)
	}

	// Verify lock file exists
	lockFile := filepath.Join(tmpDir, ".component-lock.json")
	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		t.Error("Lock file was not created")
	}

	// Load entry
	loaded, err := service.LoadEntry(tmpDir, "skills", "test-skill")
	if err != nil {
		t.Fatalf("Failed to load entry: %v", err)
	}

	// Verify entry fields
	if loaded.SourceUrl != entry.SourceUrl {
		t.Errorf("SourceUrl mismatch: got %s, want %s", loaded.SourceUrl, entry.SourceUrl)
	}
	if loaded.CommitHash != entry.CommitHash {
		t.Errorf("CommitHash mismatch: got %s, want %s", loaded.CommitHash, entry.CommitHash)
	}
}

// TestSaveEntryValidation tests validation in SaveEntry
func TestSaveEntryValidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-service-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	log := logger.New(logger.LevelError) // Use error level to suppress debug output in tests
	service := NewService(log)

	// Test nil entry
	err = service.SaveEntry(tmpDir, "skills", "test", nil)
	if err == nil {
		t.Error("Expected error for nil entry")
	}

	// Test missing SourceUrl
	entry := &models.ComponentEntry{
		CommitHash: "abc123",
	}
	err = service.SaveEntry(tmpDir, "skills", "test", entry)
	if err == nil {
		t.Error("Expected error for missing SourceUrl")
	}

	// Test missing CommitHash
	entry = &models.ComponentEntry{
		SourceUrl: "https://github.com/test/repo",
	}
	err = service.SaveEntry(tmpDir, "skills", "test", entry)
	if err == nil {
		t.Error("Expected error for missing CommitHash")
	}
}

// TestRemoveEntry tests removing a component entry
func TestRemoveEntry(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-service-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	log := logger.New(logger.LevelError) // Use error level to suppress debug output in tests
	service := NewService(log)

	// Create and save entry
	entry := &models.ComponentEntry{
		Source:     "github",
		SourceType: "git",
		SourceUrl:  "https://github.com/test/repo",
		CommitHash: "abc123",
		Version:    5,
	}

	err = service.SaveEntry(tmpDir, "skills", "test-skill", entry)
	if err != nil {
		t.Fatalf("Failed to save entry: %v", err)
	}

	// Verify entry exists
	_, err = service.LoadEntry(tmpDir, "skills", "test-skill")
	if err != nil {
		t.Fatalf("Entry should exist after save: %v", err)
	}

	// Remove entry
	err = service.RemoveEntry(tmpDir, "skills", "test-skill")
	if err != nil {
		t.Fatalf("Failed to remove entry: %v", err)
	}

	// Verify entry no longer exists
	_, err = service.LoadEntry(tmpDir, "skills", "test-skill")
	if err == nil {
		t.Error("Entry should not exist after removal")
	}
}

// TestFindComponentSources tests finding sources for a component
func TestFindComponentSources(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-service-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	log := logger.New(logger.LevelError) // Use error level to suppress debug output in tests
	service := NewService(log)

	// Create entries from different sources
	entry1 := &models.ComponentEntry{
		Source:     "github",
		SourceType: "git",
		SourceUrl:  "https://github.com/test/repo1",
		CommitHash: "abc123",
		Version:    5,
	}

	entry2 := &models.ComponentEntry{
		Source:     "github",
		SourceType: "git",
		SourceUrl:  "https://github.com/test/repo2",
		CommitHash: "def456",
		Version:    5,
	}

	// Save same component from two different sources
	err = service.SaveEntry(tmpDir, "skills", "duplicate-skill", entry1)
	if err != nil {
		t.Fatalf("Failed to save entry1: %v", err)
	}

	err = service.SaveEntry(tmpDir, "skills", "duplicate-skill", entry2)
	if err != nil {
		t.Fatalf("Failed to save entry2: %v", err)
	}

	// Find sources
	sources, err := service.FindComponentSources(tmpDir, "skills", "duplicate-skill")
	if err != nil {
		t.Fatalf("Failed to find sources: %v", err)
	}

	if len(sources) != 2 {
		t.Errorf("Expected 2 sources, got %d", len(sources))
	}

	// Check for conflict
	hasConflict, err := service.HasConflict(tmpDir, "skills", "duplicate-skill")
	if err != nil {
		t.Fatalf("Failed to check conflict: %v", err)
	}

	if !hasConflict {
		t.Error("Expected conflict to be detected")
	}
}

// TestGetAllComponentNames tests getting all component names
func TestGetAllComponentNames(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-service-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	log := logger.New(logger.LevelError) // Use error level to suppress debug output in tests
	service := NewService(log)

	// Create multiple entries
	for i, name := range []string{"skill-a", "skill-b", "skill-c"} {
		entry := &models.ComponentEntry{
			Source:     "github",
			SourceType: "git",
			SourceUrl:  "https://github.com/test/repo",
			CommitHash: string(rune('a' + i)),
			Version:    5,
		}

		err = service.SaveEntry(tmpDir, "skills", name, entry)
		if err != nil {
			t.Fatalf("Failed to save entry %s: %v", name, err)
		}
	}

	// Get all names
	names, err := service.GetAllComponentNames(tmpDir, "skills")
	if err != nil {
		t.Fatalf("Failed to get component names: %v", err)
	}

	if len(names) != 3 {
		t.Errorf("Expected 3 components, got %d", len(names))
	}
}

// TestResolveFilesystemName tests filesystem name resolution
func TestResolveFilesystemName(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-service-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	log := logger.New(logger.LevelError) // Use error level to suppress debug output in tests
	service := NewService(log)

	// First resolution should return the original name
	name, err := service.ResolveFilesystemName(tmpDir, "skills", "test-skill", "https://github.com/test/repo")
	if err != nil {
		t.Fatalf("Failed to resolve name: %v", err)
	}

	if name != "test-skill" {
		t.Errorf("Expected 'test-skill', got %s", name)
	}
}
