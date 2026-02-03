package profiles

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFindProfileBySourceURL_URLEquivalence tests that different URL formats for the same repository
// are recognized as equivalent when finding profiles.
// This is the integration test for Story-002.
func TestFindProfileBySourceURL_URLEquivalence(t *testing.T) {
	// Create a temporary profiles directory
	tempDir := t.TempDir()

	pm := &ProfileManager{
		profilesDir: tempDir,
		linker:      nil,
	}

	// Create a test profile
	profileName := "test-profile"
	profileDir := filepath.Join(tempDir, profileName)
	if err := os.MkdirAll(filepath.Join(profileDir, "skills"), 0755); err != nil {
		t.Fatalf("failed to create profile directory: %v", err)
	}

	// Save metadata with HTTPS URL
	httpsURL := "https://github.com/owner/repo"
	if err := pm.SaveProfileMetadata(profileName, httpsURL); err != nil {
		t.Fatalf("failed to save profile metadata: %v", err)
	}

	// Test that various HTTPS URL formats all find the same profile
	// Note: SSH URLs are normalized separately and won't match HTTPS profiles
	urlVariations := []string{
		"https://github.com/owner/repo",
		"https://github.com/owner/repo/",
		"https://github.com/owner/repo.git",
		"http://github.com/owner/repo",
		"owner/repo",
	}

	for _, url := range urlVariations {
		t.Run(url, func(t *testing.T) {
			foundProfile, err := pm.FindProfileBySourceURL(url)
			if err != nil {
				t.Errorf("FindProfileBySourceURL(%s) returned error: %v", url, err)
			}
			if foundProfile != profileName {
				t.Errorf("FindProfileBySourceURL(%s) = %s, want %s", url, foundProfile, profileName)
			}
		})
	}
}

// TestFindProfileBySourceURL_DifferentRepos tests that different repositories don't match
func TestFindProfileBySourceURL_DifferentRepos(t *testing.T) {
	// Create a temporary profiles directory
	tempDir := t.TempDir()

	pm := &ProfileManager{
		profilesDir: tempDir,
		linker:      nil,
	}

	// Create a test profile for owner1/repo1
	profileName := "test-profile"
	profileDir := filepath.Join(tempDir, profileName)
	if err := os.MkdirAll(filepath.Join(profileDir, "skills"), 0755); err != nil {
		t.Fatalf("failed to create profile directory: %v", err)
	}

	// Save metadata with HTTPS URL for owner1/repo1
	if err := pm.SaveProfileMetadata(profileName, "https://github.com/owner1/repo1"); err != nil {
		t.Fatalf("failed to save profile metadata: %v", err)
	}

	// Try to find profile with a different repository URL
	differentRepoURLs := []string{
		"https://github.com/owner2/repo2",
		"owner2/repo2",
		"git@github.com:owner2/repo2.git",
	}

	for _, url := range differentRepoURLs {
		t.Run(url, func(t *testing.T) {
			foundProfile, err := pm.FindProfileBySourceURL(url)
			if err != nil {
				t.Errorf("FindProfileBySourceURL(%s) returned error: %v", url, err)
			}
			if foundProfile != "" {
				t.Errorf("FindProfileBySourceURL(%s) = %s, want empty string (no match)", url, foundProfile)
			}
		})
	}
}

// TestSaveProfileMetadata_Normalization tests that profile metadata is saved with normalized URL
func TestSaveProfileMetadata_Normalization(t *testing.T) {
	// Create a temporary profiles directory
	tempDir := t.TempDir()

	pm := &ProfileManager{
		profilesDir: tempDir,
		linker:      nil,
	}

	// Create a test profile
	profileName := "test-profile"
	profileDir := filepath.Join(tempDir, profileName)
	if err := os.MkdirAll(filepath.Join(profileDir, "skills"), 0755); err != nil {
		t.Fatalf("failed to create profile directory: %v", err)
	}

	// Save metadata with SSH URL
	sshURL := "git@github.com:owner/repo.git"
	if err := pm.SaveProfileMetadata(profileName, sshURL); err != nil {
		t.Fatalf("failed to save profile metadata: %v", err)
	}

	// Load metadata and verify it's normalized (SSH preserved, .git removed)
	metadata, err := pm.LoadProfileMetadata(profileName)
	if err != nil {
		t.Fatalf("failed to load profile metadata: %v", err)
	}

	expectedNormalizedURL := "git@github.com:owner/repo"
	if metadata.SourceURL != expectedNormalizedURL {
		t.Errorf("SaveProfileMetadata normalized URL = %s, want %s", metadata.SourceURL, expectedNormalizedURL)
	}
}

// TestFindProfileBySourceURL_BackwardCompatibility tests that profiles without metadata files
// are skipped during duplicate detection (backward compatibility).
// This is the integration test for Story-003 backward compatibility requirement.
func TestFindProfileBySourceURL_BackwardCompatibility(t *testing.T) {
	// Create a temporary profiles directory
	tempDir := t.TempDir()

	pm := &ProfileManager{
		profilesDir: tempDir,
		linker:      nil,
	}

	// Create two test profiles
	legacyProfileName := "legacy-profile"
	legacyProfileDir := filepath.Join(tempDir, legacyProfileName)
	if err := os.MkdirAll(filepath.Join(legacyProfileDir, "skills"), 0755); err != nil {
		t.Fatalf("failed to create legacy profile directory: %v", err)
	}
	// Note: NO metadata file created for legacy profile

	newProfileName := "new-profile"
	newProfileDir := filepath.Join(tempDir, newProfileName)
	if err := os.MkdirAll(filepath.Join(newProfileDir, "skills"), 0755); err != nil {
		t.Fatalf("failed to create new profile directory: %v", err)
	}

	// Save metadata for the new profile
	repoURL := "https://github.com/owner/repo"
	if err := pm.SaveProfileMetadata(newProfileName, repoURL); err != nil {
		t.Fatalf("failed to save profile metadata: %v", err)
	}

	// Search for profile by URL - should find the new profile, not the legacy one
	foundProfile, err := pm.FindProfileBySourceURL(repoURL)
	if err != nil {
		t.Fatalf("FindProfileBySourceURL returned error: %v", err)
	}

	if foundProfile != newProfileName {
		t.Errorf("FindProfileBySourceURL = %s, want %s (legacy profile without metadata should be skipped)", foundProfile, newProfileName)
	}

	// Verify legacy profile metadata returns nil without error
	legacyMetadata, err := pm.LoadProfileMetadata(legacyProfileName)
	if err != nil {
		t.Errorf("LoadProfileMetadata for legacy profile returned error: %v, want nil", err)
	}
	if legacyMetadata != nil {
		t.Errorf("LoadProfileMetadata for legacy profile = %v, want nil", legacyMetadata)
	}
}

// TestProfileMetadata_FileFormat tests that the .profile-metadata file is created
// with the correct format, encoding, and structure as specified in Story-003.
func TestProfileMetadata_FileFormat(t *testing.T) {
	// Create a temporary profiles directory
	tempDir := t.TempDir()

	pm := &ProfileManager{
		profilesDir: tempDir,
		linker:      nil,
	}

	// Create a test profile
	profileName := "test-profile"
	profileDir := filepath.Join(tempDir, profileName)
	if err := os.MkdirAll(filepath.Join(profileDir, "skills"), 0755); err != nil {
		t.Fatalf("failed to create profile directory: %v", err)
	}

	// Save metadata
	repoURL := "https://github.com/owner/repo"
	if err := pm.SaveProfileMetadata(profileName, repoURL); err != nil {
		t.Fatalf("failed to save profile metadata: %v", err)
	}

	// Verify the .profile-metadata file exists
	metadataPath := filepath.Join(profileDir, ".profile-metadata")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Fatalf(".profile-metadata file not created")
	}

	// Read the file and verify it's human-readable JSON
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatalf("failed to read metadata file: %v", err)
	}

	// Verify content is human-readable and contains expected format
	content := string(data)
	if !strings.Contains(content, "source_url") {
		t.Errorf("metadata file content missing 'source_url' field, got: %s", content)
	}
	if !strings.Contains(content, repoURL) {
		t.Errorf("metadata file content missing repository URL, got: %s", content)
	}

	// Verify it's properly formatted JSON with indentation (human-readable)
	if !strings.Contains(content, "{\n") {
		t.Errorf("metadata file not formatted as indented JSON (not human-readable), got: %s", content)
	}
}
