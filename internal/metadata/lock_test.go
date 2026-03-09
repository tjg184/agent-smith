package metadata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tjg184/agent-smith/internal/models"
)

func TestUnifiedComponentEntry(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Test SaveComponentEntry for install operation
	t.Run("SaveForInstall", func(t *testing.T) {
		err := SaveComponentEntry(
			tempDir,
			"skills",
			"test-skill",
			"github",
			"git",
			"https://github.com/test/repo",
			"abc123",
			"skills/test-skill",
			ComponentEntryOptions{
				UpdatedAt:      time.Now().Format(time.RFC3339),
				Components:     5,
				Detection:      "auto",
				SourceHash:     "source-hash-123",
				CurrentHash:    "current-hash-123",
				FilesystemName: "test-skill",
			},
		)

		if err != nil {
			t.Fatalf("Failed to save component entry: %v", err)
		}

		// Verify file was created
		lockPath := filepath.Join(tempDir, ".component-lock.json")
		if _, err := os.Stat(lockPath); os.IsNotExist(err) {
			t.Fatalf("Lock file was not created")
		}

		// Load and verify
		entry, err := LoadLockFileEntry(tempDir, "skills", "test-skill")
		if err != nil {
			t.Fatalf("Failed to load component entry: %v", err)
		}

		if entry.SourceUrl != "https://github.com/test/repo" {
			t.Errorf("Expected SourceUrl 'https://github.com/test/repo', got '%s'", entry.SourceUrl)
		}

		if entry.Components != 5 {
			t.Errorf("Expected Components 5, got %d", entry.Components)
		}

		if entry.Detection != "auto" {
			t.Errorf("Expected Detection 'auto', got '%s'", entry.Detection)
		}

		if entry.SourceHash != "source-hash-123" {
			t.Errorf("Expected SourceHash 'source-hash-123', got '%s'", entry.SourceHash)
		}

		if entry.CurrentHash != "current-hash-123" {
			t.Errorf("Expected CurrentHash 'current-hash-123', got '%s'", entry.CurrentHash)
		}

		if entry.FilesystemName != "test-skill" {
			t.Errorf("Expected FilesystemName 'test-skill', got '%s'", entry.FilesystemName)
		}

		if entry.InstalledAt == "" {
			t.Error("Expected InstalledAt to be set")
		}

		if entry.UpdatedAt == "" {
			t.Error("Expected UpdatedAt to be set")
		}

		if entry.MaterializedAt != "" {
			t.Error("Expected MaterializedAt to be empty for install operation")
		}
	})

	// Test SaveComponentEntry for materialize operation
	t.Run("SaveForMaterialize", func(t *testing.T) {
		err := SaveComponentEntry(
			tempDir,
			"agents",
			"test-agent",
			"github",
			"git",
			"https://github.com/test/repo",
			"def456",
			"agents/test-agent",
			ComponentEntryOptions{
				MaterializedAt: time.Now().Format(time.RFC3339),
				SourceProfile:  "default",
				SourceHash:     "mat-source-hash",
				CurrentHash:    "mat-current-hash",
				FilesystemName: "test-agent",
			},
		)

		if err != nil {
			t.Fatalf("Failed to save materialized component entry: %v", err)
		}

		// Load and verify
		entry, err := LoadLockFileEntry(tempDir, "agents", "test-agent")
		if err != nil {
			t.Fatalf("Failed to load materialized component entry: %v", err)
		}

		if entry.MaterializedAt == "" {
			t.Error("Expected MaterializedAt to be set for materialize operation")
		}

		if entry.SourceProfile != "default" {
			t.Errorf("Expected SourceProfile 'default', got '%s'", entry.SourceProfile)
		}

		if entry.InstalledAt != "" {
			t.Error("Expected InstalledAt to be empty for materialize operation")
		}

		if entry.UpdatedAt != "" {
			t.Error("Expected UpdatedAt to be empty for materialize operation")
		}
	})

	// Test multiple sources with same component name
	t.Run("MultipleSources", func(t *testing.T) {
		// Add same component from different source
		err := SaveComponentEntry(
			tempDir,
			"skills",
			"test-skill",
			"github",
			"git",
			"https://github.com/other/repo",
			"xyz789",
			"skills/test-skill",
			ComponentEntryOptions{
				UpdatedAt:      time.Now().Format(time.RFC3339),
				Components:     3,
				Detection:      "manual",
				FilesystemName: "test-skill-2",
			},
		)

		if err != nil {
			t.Fatalf("Failed to save second source: %v", err)
		}

		// Find all sources
		sources, err := FindComponentSources(tempDir, "skills", "test-skill")
		if err != nil {
			t.Fatalf("Failed to find component sources: %v", err)
		}

		if len(sources) != 2 {
			t.Errorf("Expected 2 sources, got %d", len(sources))
		}

		// Load by specific source
		entry1, err := LoadLockFileEntryBySource(tempDir, "skills", "test-skill", "https://github.com/test/repo")
		if err != nil {
			t.Fatalf("Failed to load from first source: %v", err)
		}

		if entry1.Components != 5 {
			t.Errorf("Expected Components 5 for first source, got %d", entry1.Components)
		}

		entry2, err := LoadLockFileEntryBySource(tempDir, "skills", "test-skill", "https://github.com/other/repo")
		if err != nil {
			t.Fatalf("Failed to load from second source: %v", err)
		}

		if entry2.Components != 3 {
			t.Errorf("Expected Components 3 for second source, got %d", entry2.Components)
		}
	})

	// Test RemoveComponentEntryBySource
	t.Run("RemoveBySource", func(t *testing.T) {
		err := RemoveComponentEntryBySource(tempDir, "skills", "test-skill", "https://github.com/other/repo")
		if err != nil {
			t.Fatalf("Failed to remove by source: %v", err)
		}

		// Verify only one source remains
		sources, err := FindComponentSources(tempDir, "skills", "test-skill")
		if err != nil {
			t.Fatalf("Failed to find component sources after removal: %v", err)
		}

		if len(sources) != 1 {
			t.Errorf("Expected 1 source after removal, got %d", len(sources))
		}

		if sources[0] != "https://github.com/test/repo" {
			t.Errorf("Expected remaining source to be 'https://github.com/test/repo', got '%s'", sources[0])
		}
	})

	// Test version field
	t.Run("Version", func(t *testing.T) {
		lockPath := filepath.Join(tempDir, ".component-lock.json")
		data, err := os.ReadFile(lockPath)
		if err != nil {
			t.Fatalf("Failed to read lock file: %v", err)
		}

		var lockFile models.ComponentLockFile
		if err := json.Unmarshal(data, &lockFile); err != nil {
			t.Fatalf("Failed to unmarshal lock file: %v", err)
		}

		if lockFile.Version != 5 {
			t.Errorf("Expected version 5, got %d", lockFile.Version)
		}
	})
}

func TestFilesystemNameConflicts(t *testing.T) {
	tempDir := t.TempDir()
	skillsDir := filepath.Join(tempDir, "skills")
	os.MkdirAll(skillsDir, 0755)

	componentName := "conflict-skill"
	sourceURL1 := "https://github.com/test/repo1"
	sourceURL2 := "https://github.com/test/repo2"

	// First install uses base name
	os.MkdirAll(filepath.Join(skillsDir, componentName), 0755)
	name1, err := ResolveInstallFilesystemName(tempDir, "skills", componentName, sourceURL1)
	if err != nil {
		t.Fatalf("Failed to resolve first name: %v", err)
	}
	if name1 != componentName {
		t.Errorf("First install should use base name %s, got %s", componentName, name1)
	}

	SaveComponentEntry(tempDir, "skills", componentName, "github", "git", sourceURL1, "abc123", "",
		ComponentEntryOptions{FilesystemName: name1, UpdatedAt: time.Now().Format(time.RFC3339)})

	// Second install gets suffixed name
	name2, err := ResolveInstallFilesystemName(tempDir, "skills", componentName, sourceURL2)
	if err != nil {
		t.Fatalf("Failed to resolve second name: %v", err)
	}
	expected := componentName + "-2"
	if name2 != expected {
		t.Errorf("Second install should get name %s, got %s", expected, name2)
	}
}

func TestBackwardCompatibility(t *testing.T) {
	// Test that ComponentLockEntry type alias still works
	var entry models.ComponentLockEntry
	entry.Source = "test"
	entry.SourceUrl = "https://example.com"
	entry.Version = 5

	if entry.Source != "test" {
		t.Errorf("ComponentLockEntry alias not working correctly")
	}
}
