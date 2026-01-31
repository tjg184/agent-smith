package profiles

import (
	"os"
	"path/filepath"
	"testing"
)

// Story-006 Integration Test
//
// This test verifies that Story-006 acceptance criteria are met:
// "As a developer, I want the ProfileManager to expose methods for finding
// profiles by source URL so that the installation logic can detect duplicates."
//
// Acceptance Criteria:
// 1. ProfileManager has a FindProfileBySourceURL(repoURL string) (string, error) method
// 2. ProfileManager has a SaveProfileMetadata(profileName, sourceURL string) error method
// 3. ProfileManager has a LoadProfileMetadata(profileName string) (*ProfileMetadata, error) method
// 4. Methods handle errors gracefully (missing files, corrupt data, etc.)
// 5. Methods are well-documented with godoc comments

// TestStory006_ProfileManagerMethods verifies all required methods exist and work correctly
func TestStory006_ProfileManagerMethods(t *testing.T) {
	// Create a temporary profiles directory
	tempDir := t.TempDir()

	pm := &ProfileManager{
		profilesDir: tempDir,
		linker:      nil,
	}

	// Create a test profile
	profileName := "test-story-006"
	profileDir := filepath.Join(tempDir, profileName)
	if err := os.MkdirAll(filepath.Join(profileDir, "skills"), 0755); err != nil {
		t.Fatalf("failed to create profile directory: %v", err)
	}

	repoURL := "https://github.com/test/repo"

	// Test 1: SaveProfileMetadata method exists and works
	t.Run("SaveProfileMetadata", func(t *testing.T) {
		err := pm.SaveProfileMetadata(profileName, repoURL)
		if err != nil {
			t.Errorf("SaveProfileMetadata failed: %v", err)
		}

		// Verify metadata file was created
		metadataPath := filepath.Join(profileDir, ".profile-metadata")
		if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
			t.Error("SaveProfileMetadata did not create .profile-metadata file")
		}
	})

	// Test 2: LoadProfileMetadata method exists and returns ProfileMetadata
	t.Run("LoadProfileMetadata", func(t *testing.T) {
		metadata, err := pm.LoadProfileMetadata(profileName)
		if err != nil {
			t.Errorf("LoadProfileMetadata failed: %v", err)
		}
		if metadata == nil {
			t.Fatal("LoadProfileMetadata returned nil metadata")
		}
		if metadata.SourceURL != repoURL {
			t.Errorf("LoadProfileMetadata returned wrong URL: got %s, want %s", metadata.SourceURL, repoURL)
		}
	})

	// Test 3: FindProfileBySourceURL method exists and finds profiles
	t.Run("FindProfileBySourceURL", func(t *testing.T) {
		foundProfile, err := pm.FindProfileBySourceURL(repoURL)
		if err != nil {
			t.Errorf("FindProfileBySourceURL failed: %v", err)
		}
		if foundProfile != profileName {
			t.Errorf("FindProfileBySourceURL returned wrong profile: got %s, want %s", foundProfile, profileName)
		}
	})

	// Test 4: FindProfileBySourceURL with URL variations
	t.Run("FindProfileBySourceURL_URLVariations", func(t *testing.T) {
		urlVariations := []string{
			"https://github.com/test/repo/",
			"https://github.com/test/repo.git",
			"git@github.com:test/repo",
			"test/repo",
		}

		for _, url := range urlVariations {
			foundProfile, err := pm.FindProfileBySourceURL(url)
			if err != nil {
				t.Errorf("FindProfileBySourceURL(%s) failed: %v", url, err)
			}
			if foundProfile != profileName {
				t.Errorf("FindProfileBySourceURL(%s) = %s, want %s", url, foundProfile, profileName)
			}
		}
	})
}

// TestStory006_ErrorHandling verifies graceful error handling
func TestStory006_ErrorHandling(t *testing.T) {
	tempDir := t.TempDir()

	pm := &ProfileManager{
		profilesDir: tempDir,
		linker:      nil,
	}

	// Test 1: LoadProfileMetadata with non-existent profile returns nil without error
	t.Run("LoadProfileMetadata_MissingFile", func(t *testing.T) {
		metadata, err := pm.LoadProfileMetadata("non-existent-profile")
		if err != nil {
			t.Errorf("LoadProfileMetadata should not error on missing file, got: %v", err)
		}
		if metadata != nil {
			t.Error("LoadProfileMetadata should return nil for missing metadata file")
		}
	})

	// Test 2: LoadProfileMetadata with corrupt data returns error
	t.Run("LoadProfileMetadata_CorruptData", func(t *testing.T) {
		// Create profile with corrupt metadata
		corruptProfile := "corrupt-profile"
		corruptProfileDir := filepath.Join(tempDir, corruptProfile)
		if err := os.MkdirAll(filepath.Join(corruptProfileDir, "skills"), 0755); err != nil {
			t.Fatalf("failed to create profile directory: %v", err)
		}

		// Write invalid JSON to metadata file
		metadataPath := filepath.Join(corruptProfileDir, ".profile-metadata")
		if err := os.WriteFile(metadataPath, []byte("not valid json{{{"), 0644); err != nil {
			t.Fatalf("failed to write corrupt metadata: %v", err)
		}

		metadata, err := pm.LoadProfileMetadata(corruptProfile)
		if err == nil {
			t.Error("LoadProfileMetadata should return error for corrupt data")
		}
		if metadata != nil {
			t.Error("LoadProfileMetadata should return nil metadata on error")
		}
	})

	// Test 3: FindProfileBySourceURL returns empty string when not found
	t.Run("FindProfileBySourceURL_NotFound", func(t *testing.T) {
		foundProfile, err := pm.FindProfileBySourceURL("https://github.com/non/existent")
		if err != nil {
			t.Errorf("FindProfileBySourceURL should not error when not found, got: %v", err)
		}
		if foundProfile != "" {
			t.Errorf("FindProfileBySourceURL should return empty string when not found, got: %s", foundProfile)
		}
	})

	// Test 4: SaveProfileMetadata handles normalization errors gracefully
	t.Run("SaveProfileMetadata_InvalidURL", func(t *testing.T) {
		// Create a test profile
		testProfile := "test-invalid-url"
		testProfileDir := filepath.Join(tempDir, testProfile)
		if err := os.MkdirAll(filepath.Join(testProfileDir, "skills"), 0755); err != nil {
			t.Fatalf("failed to create profile directory: %v", err)
		}

		// Even with potentially invalid URL, should not error (will save original)
		err := pm.SaveProfileMetadata(testProfile, "not-a-valid-url")
		if err != nil {
			t.Errorf("SaveProfileMetadata should handle invalid URLs gracefully, got error: %v", err)
		}
	})
}

// TestStory006_BackwardCompatibility verifies backward compatibility with legacy profiles
func TestStory006_BackwardCompatibility(t *testing.T) {
	tempDir := t.TempDir()

	pm := &ProfileManager{
		profilesDir: tempDir,
		linker:      nil,
	}

	// Create a legacy profile without metadata
	legacyProfile := "legacy-profile"
	legacyProfileDir := filepath.Join(tempDir, legacyProfile)
	if err := os.MkdirAll(filepath.Join(legacyProfileDir, "skills"), 0755); err != nil {
		t.Fatalf("failed to create legacy profile directory: %v", err)
	}

	// Test 1: LoadProfileMetadata returns nil for legacy profiles
	t.Run("LoadProfileMetadata_Legacy", func(t *testing.T) {
		metadata, err := pm.LoadProfileMetadata(legacyProfile)
		if err != nil {
			t.Errorf("LoadProfileMetadata should not error for legacy profiles, got: %v", err)
		}
		if metadata != nil {
			t.Error("LoadProfileMetadata should return nil for legacy profiles without metadata")
		}
	})

	// Test 2: FindProfileBySourceURL skips legacy profiles
	t.Run("FindProfileBySourceURL_SkipsLegacy", func(t *testing.T) {
		// Should not find anything since the only profile has no metadata
		foundProfile, err := pm.FindProfileBySourceURL("https://github.com/test/repo")
		if err != nil {
			t.Errorf("FindProfileBySourceURL failed: %v", err)
		}
		if foundProfile != "" {
			t.Error("FindProfileBySourceURL should skip legacy profiles without metadata")
		}
	})
}
