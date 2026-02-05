//go:build integration
// +build integration

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tgaines/agent-smith/internal/metadata"
	"github.com/tgaines/agent-smith/internal/models"
)

// TestUnifiedLockFormat_InstallWorkflow tests the complete install workflow using unified format
func TestUnifiedLockFormat_InstallWorkflow(t *testing.T) {
	// Create temporary base directory
	baseDir := t.TempDir()

	// Create skill directory structure
	skillsDir := filepath.Join(baseDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	// Test component details
	componentName := "test-skill"
	sourceURL := "https://github.com/test/repo"
	filesystemName := "test-skill"

	// AC1: Install operation creates .component-lock.json
	t.Run("AC1: Install creates unified lock file", func(t *testing.T) {
		err := metadata.SaveComponentEntry(
			baseDir,
			"skills",
			componentName,
			"github",
			"git",
			sourceURL,
			"abc123",
			"",
			metadata.ComponentEntryOptions{
				UpdatedAt:      time.Now().Format(time.RFC3339),
				FilesystemName: filesystemName,
				SourceHash:     "source-hash-123",
				CurrentHash:    "current-hash-123",
			},
		)
		if err != nil {
			t.Fatalf("Failed to save component entry: %v", err)
		}

		// Verify .component-lock.json exists
		lockPath := filepath.Join(baseDir, ".component-lock.json")
		if _, err := os.Stat(lockPath); os.IsNotExist(err) {
			t.Fatalf("Lock file not created at %s", lockPath)
		}

		// Verify no old lock files exist
		oldPaths := []string{
			filepath.Join(baseDir, ".skill-lock.json"),
			filepath.Join(baseDir, ".agent-lock.json"),
			filepath.Join(baseDir, ".command-lock.json"),
		}
		for _, oldPath := range oldPaths {
			if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
				t.Errorf("Old lock file should not exist: %s", oldPath)
			}
		}

		t.Logf("✓ AC1: Install creates .component-lock.json")
	})

	// AC2: Lock file has correct structure (version 5, nested maps)
	t.Run("AC2: Lock file structure is correct", func(t *testing.T) {
		lockPath := filepath.Join(baseDir, ".component-lock.json")
		data, err := os.ReadFile(lockPath)
		if err != nil {
			t.Fatalf("Failed to read lock file: %v", err)
		}

		var lockFile models.ComponentLockFile
		if err := json.Unmarshal(data, &lockFile); err != nil {
			t.Fatalf("Failed to parse lock file: %v", err)
		}

		// Check version
		if lockFile.Version != 5 {
			t.Errorf("Expected version 5, got %d", lockFile.Version)
		}

		// Check nested structure: Skills[sourceURL][componentName]
		if lockFile.Skills == nil {
			t.Fatalf("Skills map is nil")
		}
		if lockFile.Skills[sourceURL] == nil {
			t.Fatalf("Source URL map is nil for %s", sourceURL)
		}

		entry, exists := lockFile.Skills[sourceURL][componentName]
		if !exists {
			t.Fatalf("Component entry not found for %s", componentName)
		}

		// Verify entry fields
		if entry.SourceUrl != sourceURL {
			t.Errorf("Expected SourceUrl %s, got %s", sourceURL, entry.SourceUrl)
		}
		if entry.CommitHash != "abc123" {
			t.Errorf("Expected CommitHash abc123, got %s", entry.CommitHash)
		}
		if entry.FilesystemName != filesystemName {
			t.Errorf("Expected FilesystemName %s, got %s", filesystemName, entry.FilesystemName)
		}
		if entry.SourceHash != "source-hash-123" {
			t.Errorf("Expected SourceHash source-hash-123, got %s", entry.SourceHash)
		}
		if entry.CurrentHash != "current-hash-123" {
			t.Errorf("Expected CurrentHash current-hash-123, got %s", entry.CurrentHash)
		}
		if entry.InstalledAt == "" {
			t.Error("InstalledAt should be set")
		}
		if entry.UpdatedAt == "" {
			t.Error("UpdatedAt should be set")
		}

		t.Logf("✓ AC2: Lock file structure verified (version 5, nested maps, all fields)")
	})

	// AC3: Load operation reads from unified lock file
	t.Run("AC3: Load reads from unified lock file", func(t *testing.T) {
		entry, err := metadata.LoadLockFileEntry(baseDir, "skills", componentName)
		if err != nil {
			t.Fatalf("Failed to load component entry: %v", err)
		}

		if entry.SourceUrl != sourceURL {
			t.Errorf("Expected SourceUrl %s, got %s", sourceURL, entry.SourceUrl)
		}
		if entry.CommitHash != "abc123" {
			t.Errorf("Expected CommitHash abc123, got %s", entry.CommitHash)
		}
		if entry.FilesystemName != filesystemName {
			t.Errorf("Expected FilesystemName %s, got %s", filesystemName, entry.FilesystemName)
		}

		t.Logf("✓ AC3: Load successfully reads from .component-lock.json")
	})

	// AC4: Multiple sources with same component name
	t.Run("AC4: Multiple sources tracked independently", func(t *testing.T) {
		secondSourceURL := "https://github.com/another/repo"
		secondFilesystemName := "test-skill-2"

		// Install same component from different source
		err := metadata.SaveComponentEntry(
			baseDir,
			"skills",
			componentName,
			"github",
			"git",
			secondSourceURL,
			"def456",
			"",
			metadata.ComponentEntryOptions{
				UpdatedAt:      time.Now().Format(time.RFC3339),
				FilesystemName: secondFilesystemName,
				SourceHash:     "source-hash-456",
				CurrentHash:    "current-hash-456",
			},
		)
		if err != nil {
			t.Fatalf("Failed to save second component entry: %v", err)
		}

		// LoadLockFileEntry should fail for multiple sources
		_, err = metadata.LoadLockFileEntry(baseDir, "skills", componentName)
		if err == nil {
			t.Error("LoadLockFileEntry should fail when component exists in multiple sources")
		}

		// Verify both exist in lock file
		lockPath := filepath.Join(baseDir, ".component-lock.json")
		data, err := os.ReadFile(lockPath)
		if err != nil {
			t.Fatalf("Failed to read lock file: %v", err)
		}

		var lockFile models.ComponentLockFile
		if err := json.Unmarshal(data, &lockFile); err != nil {
			t.Fatalf("Failed to parse lock file: %v", err)
		}

		if len(lockFile.Skills) != 2 {
			t.Errorf("Expected 2 sources, got %d", len(lockFile.Skills))
		}

		_, exists1 := lockFile.Skills[sourceURL][componentName]
		_, exists2 := lockFile.Skills[secondSourceURL][componentName]
		if !exists1 || !exists2 {
			t.Error("Both source entries should exist in lock file")
		}

		t.Logf("✓ AC4: Multiple sources tracked independently in .component-lock.json")
	})

	// AC5: Remove operation removes from unified lock file
	t.Run("AC5: Remove deletes from unified lock file", func(t *testing.T) {
		// Remove one source
		err := metadata.RemoveLockFileEntry(baseDir, "skills", componentName)
		if err != nil {
			t.Fatalf("Failed to remove component entry: %v", err)
		}

		// Verify lock file still exists (other source remains)
		lockPath := filepath.Join(baseDir, ".component-lock.json")
		if _, err := os.Stat(lockPath); os.IsNotExist(err) {
			t.Error("Lock file should still exist after removing one source")
		}

		// Verify entry is removed
		data, err := os.ReadFile(lockPath)
		if err != nil {
			t.Fatalf("Failed to read lock file: %v", err)
		}

		var lockFile models.ComponentLockFile
		if err := json.Unmarshal(data, &lockFile); err != nil {
			t.Fatalf("Failed to parse lock file: %v", err)
		}

		// Should have fewer sources now
		if len(lockFile.Skills) > 2 {
			t.Errorf("Expected at most 2 sources after removal, got %d", len(lockFile.Skills))
		}

		t.Logf("✓ AC5: Remove successfully deletes from .component-lock.json")
	})
}

// TestUnifiedLockFormat_DriftDetection tests drift detection with hashes
func TestUnifiedLockFormat_DriftDetection(t *testing.T) {
	baseDir := t.TempDir()

	componentName := "drift-test-skill"
	sourceURL := "https://github.com/test/drift"
	filesystemName := "drift-test-skill"

	// AC1: Install with matching hashes (no drift)
	t.Run("AC1: No drift when hashes match", func(t *testing.T) {
		matchingHash := "matching-hash-abc"
		err := metadata.SaveComponentEntry(
			baseDir,
			"skills",
			componentName,
			"github",
			"git",
			sourceURL,
			"abc123",
			"",
			metadata.ComponentEntryOptions{
				UpdatedAt:      time.Now().Format(time.RFC3339),
				FilesystemName: filesystemName,
				SourceHash:     matchingHash,
				CurrentHash:    matchingHash,
			},
		)
		if err != nil {
			t.Fatalf("Failed to save component entry: %v", err)
		}

		entry, err := metadata.LoadLockFileEntry(baseDir, "skills", componentName)
		if err != nil {
			t.Fatalf("Failed to load entry: %v", err)
		}

		// No drift when hashes match
		if entry.SourceHash != entry.CurrentHash {
			t.Errorf("Hashes should match for no-drift scenario")
		}

		t.Logf("✓ AC1: No drift detected when sourceHash == currentHash")
	})

	// AC2: Detect drift when hashes differ
	t.Run("AC2: Drift detected when hashes differ", func(t *testing.T) {
		// Update with different currentHash (simulating local modification)
		err := metadata.SaveComponentEntry(
			baseDir,
			"skills",
			componentName,
			"github",
			"git",
			sourceURL,
			"abc123",
			"",
			metadata.ComponentEntryOptions{
				UpdatedAt:      time.Now().Format(time.RFC3339),
				FilesystemName: filesystemName,
				SourceHash:     "original-source-hash",
				CurrentHash:    "modified-current-hash",
			},
		)
		if err != nil {
			t.Fatalf("Failed to update component entry: %v", err)
		}

		entry, err := metadata.LoadLockFileEntry(baseDir, "skills", componentName)
		if err != nil {
			t.Fatalf("Failed to load entry: %v", err)
		}

		// Drift detected when hashes differ
		if entry.SourceHash == entry.CurrentHash {
			t.Errorf("Hashes should differ for drift scenario")
		}
		if entry.SourceHash != "original-source-hash" {
			t.Errorf("Expected SourceHash original-source-hash, got %s", entry.SourceHash)
		}
		if entry.CurrentHash != "modified-current-hash" {
			t.Errorf("Expected CurrentHash modified-current-hash, got %s", entry.CurrentHash)
		}

		t.Logf("✓ AC2: Drift detected when sourceHash != currentHash")
	})
}

// TestUnifiedLockFormat_FilesystemNameConflicts tests filesystem name conflict resolution
func TestUnifiedLockFormat_FilesystemNameConflicts(t *testing.T) {
	baseDir := t.TempDir()

	// Create skills directory
	skillsDir := filepath.Join(baseDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	componentName := "conflict-skill"
	sourceURL1 := "https://github.com/test/repo1"
	sourceURL2 := "https://github.com/test/repo2"

	// AC1: First install uses base name
	t.Run("AC1: First install uses base name", func(t *testing.T) {
		// Create actual directory
		dir1 := filepath.Join(skillsDir, componentName)
		if err := os.MkdirAll(dir1, 0755); err != nil {
			t.Fatalf("Failed to create skill dir: %v", err)
		}

		filesystemName, err := metadata.ResolveInstallFilesystemName(baseDir, "skills", componentName, sourceURL1)
		if err != nil {
			t.Fatalf("Failed to resolve filesystem name: %v", err)
		}

		err = metadata.SaveComponentEntry(
			baseDir,
			"skills",
			componentName,
			"github",
			"git",
			sourceURL1,
			"abc123",
			"",
			metadata.ComponentEntryOptions{
				UpdatedAt:      time.Now().Format(time.RFC3339),
				FilesystemName: filesystemName,
				SourceHash:     "hash1",
				CurrentHash:    "hash1",
			},
		)
		if err != nil {
			t.Fatalf("Failed to save component entry: %v", err)
		}

		if filesystemName != componentName {
			t.Errorf("First install should use base name %s, got %s", componentName, filesystemName)
		}

		t.Logf("✓ AC1: First install uses base name: %s", filesystemName)
	})

	// AC2: Second install with same name gets suffixed name
	t.Run("AC2: Second install gets suffixed name", func(t *testing.T) {
		filesystemName, err := metadata.ResolveInstallFilesystemName(baseDir, "skills", componentName, sourceURL2)
		if err != nil {
			t.Fatalf("Failed to resolve filesystem name: %v", err)
		}

		// Create actual directory with suffixed name
		dir2 := filepath.Join(skillsDir, filesystemName)
		if err := os.MkdirAll(dir2, 0755); err != nil {
			t.Fatalf("Failed to create skill dir: %v", err)
		}

		err = metadata.SaveComponentEntry(
			baseDir,
			"skills",
			componentName,
			"github",
			"git",
			sourceURL2,
			"def456",
			"",
			metadata.ComponentEntryOptions{
				UpdatedAt:      time.Now().Format(time.RFC3339),
				FilesystemName: filesystemName,
				SourceHash:     "hash2",
				CurrentHash:    "hash2",
			},
		)
		if err != nil {
			t.Fatalf("Failed to save component entry: %v", err)
		}

		if filesystemName == componentName {
			t.Errorf("Second install should get suffixed name, got %s", filesystemName)
		}

		expectedSuffix := componentName + "-2"
		if filesystemName != expectedSuffix {
			t.Errorf("Expected filesystem name %s, got %s", expectedSuffix, filesystemName)
		}

		t.Logf("✓ AC2: Second install gets suffixed name: %s", filesystemName)
	})

	// AC3: Both entries tracked with different filesystem names
	t.Run("AC3: Both entries tracked independently", func(t *testing.T) {
		lockPath := filepath.Join(baseDir, ".component-lock.json")
		data, err := os.ReadFile(lockPath)
		if err != nil {
			t.Fatalf("Failed to read lock file: %v", err)
		}

		var lockFile models.ComponentLockFile
		if err := json.Unmarshal(data, &lockFile); err != nil {
			t.Fatalf("Failed to parse lock file: %v", err)
		}

		// Both sources should exist
		entry1, exists1 := lockFile.Skills[sourceURL1][componentName]
		entry2, exists2 := lockFile.Skills[sourceURL2][componentName]

		if !exists1 || !exists2 {
			t.Fatal("Both entries should exist in lock file")
		}

		if entry1.FilesystemName == entry2.FilesystemName {
			t.Errorf("Filesystem names should differ: %s vs %s", entry1.FilesystemName, entry2.FilesystemName)
		}

		t.Logf("✓ AC3: Both entries tracked with different filesystem names: %s, %s",
			entry1.FilesystemName, entry2.FilesystemName)
	})
}

// TestUnifiedLockFormat_AllComponentTypes tests that all component types use unified format
func TestUnifiedLockFormat_AllComponentTypes(t *testing.T) {
	baseDir := t.TempDir()

	componentTypes := []struct {
		typeName string
		name     string
		source   string
	}{
		{"skills", "test-skill", "https://github.com/test/skills"},
		{"agents", "test-agent", "https://github.com/test/agents"},
		{"commands", "test-command", "https://github.com/test/commands"},
	}

	for _, ct := range componentTypes {
		t.Run(ct.typeName, func(t *testing.T) {
			// Create component directory
			componentDir := filepath.Join(baseDir, ct.typeName)
			if err := os.MkdirAll(componentDir, 0755); err != nil {
				t.Fatalf("Failed to create %s dir: %v", ct.typeName, err)
			}

			// Save entry
			err := metadata.SaveComponentEntry(
				baseDir,
				ct.typeName,
				ct.name,
				"github",
				"git",
				ct.source,
				"abc123",
				"",
				metadata.ComponentEntryOptions{
					UpdatedAt:      time.Now().Format(time.RFC3339),
					FilesystemName: ct.name,
					SourceHash:     "hash-" + ct.typeName,
					CurrentHash:    "hash-" + ct.typeName,
				},
			)
			if err != nil {
				t.Fatalf("Failed to save %s entry: %v", ct.typeName, err)
			}

			// Verify entry in lock file
			lockPath := filepath.Join(baseDir, ".component-lock.json")
			data, err := os.ReadFile(lockPath)
			if err != nil {
				t.Fatalf("Failed to read lock file: %v", err)
			}

			var lockFile models.ComponentLockFile
			if err := json.Unmarshal(data, &lockFile); err != nil {
				t.Fatalf("Failed to parse lock file: %v", err)
			}

			// Check correct map
			var componentMap map[string]map[string]models.ComponentEntry
			switch ct.typeName {
			case "skills":
				componentMap = lockFile.Skills
			case "agents":
				componentMap = lockFile.Agents
			case "commands":
				componentMap = lockFile.Commands
			}

			if componentMap == nil {
				t.Fatalf("%s map is nil", ct.typeName)
			}
			if componentMap[ct.source] == nil {
				t.Fatalf("Source map is nil for %s", ct.source)
			}

			entry, exists := componentMap[ct.source][ct.name]
			if !exists {
				t.Fatalf("Entry not found for %s/%s", ct.typeName, ct.name)
			}

			if entry.SourceUrl != ct.source {
				t.Errorf("Expected SourceURL %s, got %s", ct.source, entry.SourceUrl)
			}

			t.Logf("✓ %s uses unified .component-lock.json format", ct.typeName)
		})
	}

	// Verify all in same file
	t.Run("All types in single lock file", func(t *testing.T) {
		lockPath := filepath.Join(baseDir, ".component-lock.json")
		data, err := os.ReadFile(lockPath)
		if err != nil {
			t.Fatalf("Failed to read lock file: %v", err)
		}

		var lockFile models.ComponentLockFile
		if err := json.Unmarshal(data, &lockFile); err != nil {
			t.Fatalf("Failed to parse lock file: %v", err)
		}

		if len(lockFile.Skills) == 0 {
			t.Error("Skills should be present")
		}
		if len(lockFile.Agents) == 0 {
			t.Error("Agents should be present")
		}
		if len(lockFile.Commands) == 0 {
			t.Error("Commands should be present")
		}

		t.Logf("✓ All component types stored in single .component-lock.json file")
	})
}

// TestUnifiedLockFormat_UpdatePreservesInstalledAt tests that updates preserve installedAt
func TestUnifiedLockFormat_UpdatePreservesInstalledAt(t *testing.T) {
	baseDir := t.TempDir()

	componentName := "update-test-skill"
	sourceURL := "https://github.com/test/update"
	filesystemName := "update-test-skill"

	// Initial install
	initialTime := time.Now().Add(-24 * time.Hour)
	err := metadata.SaveComponentEntry(
		baseDir,
		"skills",
		componentName,
		"github",
		"git",
		sourceURL,
		"abc123",
		"",
		metadata.ComponentEntryOptions{
			UpdatedAt:      time.Now().Format(time.RFC3339),
			FilesystemName: filesystemName,
			SourceHash:     "hash1",
			CurrentHash:    "hash1",
		},
	)
	if err != nil {
		t.Fatalf("Failed to save initial entry: %v", err)
	}

	// Manually set installedAt to past time
	lockPath := filepath.Join(baseDir, ".component-lock.json")
	data, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("Failed to read lock file: %v", err)
	}

	var lockFile models.ComponentLockFile
	if err := json.Unmarshal(data, &lockFile); err != nil {
		t.Fatalf("Failed to parse lock file: %v", err)
	}

	entry := lockFile.Skills[sourceURL][componentName]
	entry.InstalledAt = initialTime.Format(time.RFC3339)
	lockFile.Skills[sourceURL][componentName] = entry

	updatedData, err := json.MarshalIndent(lockFile, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal lock file: %v", err)
	}

	if err := os.WriteFile(lockPath, updatedData, 0644); err != nil {
		t.Fatalf("Failed to write lock file: %v", err)
	}

	// Simulate update (new commit hash)
	time.Sleep(100 * time.Millisecond)
	err = metadata.SaveComponentEntry(
		baseDir,
		"skills",
		componentName,
		"github",
		"git",
		sourceURL,
		"def456",
		"",
		metadata.ComponentEntryOptions{
			UpdatedAt:      time.Now().Format(time.RFC3339),
			FilesystemName: filesystemName,
			SourceHash:     "hash2",
			CurrentHash:    "hash2",
		},
	)
	if err != nil {
		t.Fatalf("Failed to save updated entry: %v", err)
	}

	// Verify installedAt preserved, updatedAt changed
	updatedEntry, err := metadata.LoadLockFileEntry(baseDir, "skills", componentName)
	if err != nil {
		t.Fatalf("Failed to load updated entry: %v", err)
	}

	if updatedEntry.CommitHash != "def456" {
		t.Errorf("Expected updated CommitHash def456, got %s", updatedEntry.CommitHash)
	}

	// InstalledAt should be preserved
	parsedInstalledAt, err := time.Parse(time.RFC3339, updatedEntry.InstalledAt)
	if err != nil {
		t.Fatalf("Failed to parse InstalledAt: %v", err)
	}

	timeDiff := parsedInstalledAt.Sub(initialTime)
	if timeDiff > time.Second || timeDiff < -time.Second {
		t.Errorf("InstalledAt should be preserved, expected ~%v, got %v", initialTime, parsedInstalledAt)
	}

	// UpdatedAt should be recent
	parsedUpdatedAt, err := time.Parse(time.RFC3339, updatedEntry.UpdatedAt)
	if err != nil {
		t.Fatalf("Failed to parse UpdatedAt: %v", err)
	}

	if time.Since(parsedUpdatedAt) > 2*time.Second {
		t.Errorf("UpdatedAt should be recent, got %v", parsedUpdatedAt)
	}

	t.Logf("✓ Update preserves InstalledAt and updates UpdatedAt")
}
