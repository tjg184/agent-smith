package profiles

import (
	"os"
	"path/filepath"
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

	// Test that various URL formats all find the same profile
	urlVariations := []string{
		"https://github.com/owner/repo",
		"https://github.com/owner/repo/",
		"https://github.com/owner/repo.git",
		"http://github.com/owner/repo",
		"git@github.com:owner/repo",
		"git@github.com:owner/repo.git",
		"ssh://git@github.com/owner/repo",
		"ssh://git@github.com/owner/repo.git",
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

	// Load metadata and verify it's normalized to HTTPS
	metadata, err := pm.LoadProfileMetadata(profileName)
	if err != nil {
		t.Fatalf("failed to load profile metadata: %v", err)
	}

	expectedNormalizedURL := "https://github.com/owner/repo"
	if metadata.SourceURL != expectedNormalizedURL {
		t.Errorf("SaveProfileMetadata normalized URL = %s, want %s", metadata.SourceURL, expectedNormalizedURL)
	}
}
