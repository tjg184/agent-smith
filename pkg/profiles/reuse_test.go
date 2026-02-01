package profiles

import (
	"os"
	"path/filepath"
	"testing"
)

// TestProfileReuse_FindExistingProfileByNormalizedURL tests that profiles can be found
// by their source URL regardless of URL format (HTTPS, SSH, shorthand)
func TestProfileReuse_FindExistingProfileByNormalizedURL(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	pm := &ProfileManager{
		profilesDir: tempDir,
		linker:      nil,
	}

	// Create a test profile
	profileName := "existing-profile"
	profileDir := filepath.Join(tempDir, profileName)
	if err := os.MkdirAll(filepath.Join(profileDir, "skills"), 0755); err != nil {
		t.Fatalf("Failed to create profile directory: %v", err)
	}

	// Save metadata with HTTPS URL
	originalURL := "https://github.com/owner/repo"
	if err := pm.SaveProfileMetadata(profileName, originalURL); err != nil {
		t.Fatalf("Failed to save profile metadata: %v", err)
	}

	// Test that various URL formats all find the same profile
	urlVariations := []struct {
		name string
		url  string
	}{
		{"HTTPS standard", "https://github.com/owner/repo"},
		{"HTTPS with trailing slash", "https://github.com/owner/repo/"},
		{"HTTPS with .git", "https://github.com/owner/repo.git"},
		{"HTTP (upgraded to HTTPS)", "http://github.com/owner/repo"},
		{"SSH git@", "git@github.com:owner/repo"},
		{"SSH git@ with .git", "git@github.com:owner/repo.git"},
		{"SSH protocol", "ssh://git@github.com/owner/repo"},
		{"SSH protocol with .git", "ssh://git@github.com/owner/repo.git"},
		{"Shorthand", "owner/repo"},
	}

	for _, tt := range urlVariations {
		t.Run(tt.name, func(t *testing.T) {
			foundProfile, err := pm.FindProfileBySourceURL(tt.url)
			if err != nil {
				t.Errorf("FindProfileBySourceURL(%s) returned error: %v", tt.url, err)
			}
			if foundProfile != profileName {
				t.Errorf("FindProfileBySourceURL(%s) = %s, want %s", tt.url, foundProfile, profileName)
			}
		})
	}
}

// TestProfileReuse_DifferentReposDifferentProfiles tests that different repositories
// result in different profiles (no false positives)
func TestProfileReuse_DifferentReposDifferentProfiles(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	pm := &ProfileManager{
		profilesDir: tempDir,
		linker:      nil,
	}

	// Create first profile for owner1/repo1
	profile1 := "profile-1"
	profileDir1 := filepath.Join(tempDir, profile1)
	if err := os.MkdirAll(filepath.Join(profileDir1, "skills"), 0755); err != nil {
		t.Fatalf("Failed to create profile1 directory: %v", err)
	}
	if err := pm.SaveProfileMetadata(profile1, "https://github.com/owner1/repo1"); err != nil {
		t.Fatalf("Failed to save profile1 metadata: %v", err)
	}

	// Create second profile for owner2/repo2
	profile2 := "profile-2"
	profileDir2 := filepath.Join(tempDir, profile2)
	if err := os.MkdirAll(filepath.Join(profileDir2, "skills"), 0755); err != nil {
		t.Fatalf("Failed to create profile2 directory: %v", err)
	}
	if err := pm.SaveProfileMetadata(profile2, "https://github.com/owner2/repo2"); err != nil {
		t.Fatalf("Failed to save profile2 metadata: %v", err)
	}

	// Verify finding profile1 doesn't match profile2's URL
	foundProfile, err := pm.FindProfileBySourceURL("https://github.com/owner2/repo2")
	if err != nil {
		t.Fatalf("FindProfileBySourceURL returned error: %v", err)
	}
	if foundProfile != profile2 {
		t.Errorf("FindProfileBySourceURL(owner2/repo2) = %s, want %s", foundProfile, profile2)
	}

	// Verify finding profile1 URL returns profile1, not profile2
	foundProfile, err = pm.FindProfileBySourceURL("owner1/repo1")
	if err != nil {
		t.Fatalf("FindProfileBySourceURL returned error: %v", err)
	}
	if foundProfile != profile1 {
		t.Errorf("FindProfileBySourceURL(owner1/repo1) = %s, want %s", foundProfile, profile1)
	}

	// Verify non-existent repo returns empty string
	foundProfile, err = pm.FindProfileBySourceURL("https://github.com/owner3/repo3")
	if err != nil {
		t.Fatalf("FindProfileBySourceURL returned error: %v", err)
	}
	if foundProfile != "" {
		t.Errorf("FindProfileBySourceURL(owner3/repo3) = %s, want empty string", foundProfile)
	}
}

// TestProfileReuse_UpdateExistingProfileMetadata tests that updating metadata
// for an existing profile works correctly
func TestProfileReuse_UpdateExistingProfileMetadata(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	pm := &ProfileManager{
		profilesDir: tempDir,
		linker:      nil,
	}

	// Create a test profile
	profileName := "test-profile"
	profileDir := filepath.Join(tempDir, profileName)
	if err := os.MkdirAll(filepath.Join(profileDir, "skills"), 0755); err != nil {
		t.Fatalf("Failed to create profile directory: %v", err)
	}

	// Save initial metadata
	initialURL := "https://github.com/owner/repo"
	if err := pm.SaveProfileMetadata(profileName, initialURL); err != nil {
		t.Fatalf("Failed to save initial metadata: %v", err)
	}

	// Verify initial metadata is correct
	metadata, err := pm.LoadProfileMetadata(profileName)
	if err != nil {
		t.Fatalf("Failed to load initial metadata: %v", err)
	}
	if metadata.SourceURL != initialURL {
		t.Errorf("Initial metadata URL = %s, want %s", metadata.SourceURL, initialURL)
	}

	// Update metadata with SSH URL (should normalize to same URL)
	updatedURL := "git@github.com:owner/repo.git"
	if err := pm.SaveProfileMetadata(profileName, updatedURL); err != nil {
		t.Fatalf("Failed to update metadata: %v", err)
	}

	// Verify metadata is still normalized to the same URL
	metadata, err = pm.LoadProfileMetadata(profileName)
	if err != nil {
		t.Fatalf("Failed to load updated metadata: %v", err)
	}
	if metadata.SourceURL != initialURL {
		t.Errorf("Updated metadata URL = %s, want %s (normalized)", metadata.SourceURL, initialURL)
	}
}

// TestProfileReuse_PreventDuplicateProfiles tests that the system prevents
// creating duplicate profiles for the same source repository
func TestProfileReuse_PreventDuplicateProfiles(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	pm := &ProfileManager{
		profilesDir: tempDir,
		linker:      nil,
	}

	// Create first profile for a repository
	profile1 := "profile-1"
	profileDir1 := filepath.Join(tempDir, profile1)
	if err := os.MkdirAll(filepath.Join(profileDir1, "skills"), 0755); err != nil {
		t.Fatalf("Failed to create profile1 directory: %v", err)
	}
	repoURL := "https://github.com/owner/repo"
	if err := pm.SaveProfileMetadata(profile1, repoURL); err != nil {
		t.Fatalf("Failed to save profile1 metadata: %v", err)
	}

	// Verify we can find the existing profile
	foundProfile, err := pm.FindProfileBySourceURL(repoURL)
	if err != nil {
		t.Fatalf("FindProfileBySourceURL returned error: %v", err)
	}
	if foundProfile != profile1 {
		t.Errorf("FindProfileBySourceURL = %s, want %s", foundProfile, profile1)
	}

	// Verify finding by different URL formats returns the same profile
	foundProfile, err = pm.FindProfileBySourceURL("git@github.com:owner/repo.git")
	if err != nil {
		t.Fatalf("FindProfileBySourceURL (SSH) returned error: %v", err)
	}
	if foundProfile != profile1 {
		t.Errorf("FindProfileBySourceURL (SSH) = %s, want %s", foundProfile, profile1)
	}

	// If we try to create a second profile for the same repo, we should detect it
	// This is the responsibility of the caller (install command) to check FindProfileBySourceURL first
	foundProfile, err = pm.FindProfileBySourceURL("owner/repo")
	if err != nil {
		t.Fatalf("FindProfileBySourceURL (shorthand) returned error: %v", err)
	}
	if foundProfile != profile1 {
		t.Errorf("Duplicate detection failed: FindProfileBySourceURL(owner/repo) = %s, want %s", foundProfile, profile1)
	}
}

// TestProfileReuse_MetadataIntegrity tests that profile metadata remains intact
// across multiple operations
func TestProfileReuse_MetadataIntegrity(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	pm := &ProfileManager{
		profilesDir: tempDir,
		linker:      nil,
	}

	// Create a test profile
	profileName := "test-profile"
	profileDir := filepath.Join(tempDir, profileName)
	if err := os.MkdirAll(filepath.Join(profileDir, "skills"), 0755); err != nil {
		t.Fatalf("Failed to create profile directory: %v", err)
	}

	// Save metadata
	repoURL := "https://github.com/owner/repo"
	if err := pm.SaveProfileMetadata(profileName, repoURL); err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Load metadata multiple times - should be consistent
	for i := 0; i < 3; i++ {
		metadata, err := pm.LoadProfileMetadata(profileName)
		if err != nil {
			t.Fatalf("Load %d: Failed to load metadata: %v", i, err)
		}
		if metadata.SourceURL != repoURL {
			t.Errorf("Load %d: metadata URL = %s, want %s", i, metadata.SourceURL, repoURL)
		}
	}

	// Verify FindProfileBySourceURL still works after multiple loads
	foundProfile, err := pm.FindProfileBySourceURL(repoURL)
	if err != nil {
		t.Fatalf("FindProfileBySourceURL returned error: %v", err)
	}
	if foundProfile != profileName {
		t.Errorf("FindProfileBySourceURL = %s, want %s", foundProfile, profileName)
	}
}

// TestProfileReuse_BackwardCompatibilityNoMetadata tests that profiles without
// metadata files are skipped during duplicate detection (backward compatibility)
func TestProfileReuse_BackwardCompatibilityNoMetadata(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	pm := &ProfileManager{
		profilesDir: tempDir,
		linker:      nil,
	}

	// Create legacy profile without metadata
	legacyProfile := "legacy-profile"
	legacyProfileDir := filepath.Join(tempDir, legacyProfile)
	if err := os.MkdirAll(filepath.Join(legacyProfileDir, "skills"), 0755); err != nil {
		t.Fatalf("Failed to create legacy profile directory: %v", err)
	}
	// Note: NO metadata file created for legacy profile

	// Create new profile with metadata for same repository
	newProfile := "new-profile"
	newProfileDir := filepath.Join(tempDir, newProfile)
	if err := os.MkdirAll(filepath.Join(newProfileDir, "skills"), 0755); err != nil {
		t.Fatalf("Failed to create new profile directory: %v", err)
	}
	repoURL := "https://github.com/owner/repo"
	if err := pm.SaveProfileMetadata(newProfile, repoURL); err != nil {
		t.Fatalf("Failed to save new profile metadata: %v", err)
	}

	// Search for profile by URL - should find the new profile, not the legacy one
	foundProfile, err := pm.FindProfileBySourceURL(repoURL)
	if err != nil {
		t.Fatalf("FindProfileBySourceURL returned error: %v", err)
	}
	if foundProfile != newProfile {
		t.Errorf("FindProfileBySourceURL = %s, want %s (legacy profile should be skipped)", foundProfile, newProfile)
	}

	// Verify legacy profile metadata returns nil without error
	legacyMetadata, err := pm.LoadProfileMetadata(legacyProfile)
	if err != nil {
		t.Errorf("LoadProfileMetadata for legacy profile returned error: %v, want nil", err)
	}
	if legacyMetadata != nil {
		t.Errorf("LoadProfileMetadata for legacy profile = %v, want nil", legacyMetadata)
	}
}

// TestProfileReuse_NormalizedURLStorage tests that URLs are normalized before
// being stored in metadata
func TestProfileReuse_NormalizedURLStorage(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	pm := &ProfileManager{
		profilesDir: tempDir,
		linker:      nil,
	}

	tests := []struct {
		name        string
		inputURL    string
		expectedURL string
	}{
		{
			name:        "SSH URL normalized to HTTPS",
			inputURL:    "git@github.com:owner/repo.git",
			expectedURL: "https://github.com/owner/repo",
		},
		{
			name:        "HTTPS with trailing slash",
			inputURL:    "https://github.com/owner/repo/",
			expectedURL: "https://github.com/owner/repo",
		},
		{
			name:        "HTTPS with .git extension",
			inputURL:    "https://github.com/owner/repo.git",
			expectedURL: "https://github.com/owner/repo",
		},
		{
			name:        "HTTP upgraded to HTTPS",
			inputURL:    "http://github.com/owner/repo",
			expectedURL: "https://github.com/owner/repo",
		},
		{
			name:        "Shorthand expanded",
			inputURL:    "owner/repo",
			expectedURL: "https://github.com/owner/repo",
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a unique profile for each test
			profileName := filepath.Join("test-profile", string(rune('a'+i)))
			profileDir := filepath.Join(tempDir, profileName)
			if err := os.MkdirAll(filepath.Join(profileDir, "skills"), 0755); err != nil {
				t.Fatalf("Failed to create profile directory: %v", err)
			}

			// Save metadata with input URL
			if err := pm.SaveProfileMetadata(profileName, tt.inputURL); err != nil {
				t.Fatalf("Failed to save metadata: %v", err)
			}

			// Load metadata and verify it's normalized
			metadata, err := pm.LoadProfileMetadata(profileName)
			if err != nil {
				t.Fatalf("Failed to load metadata: %v", err)
			}

			if metadata.SourceURL != tt.expectedURL {
				t.Errorf("Normalized URL = %s, want %s", metadata.SourceURL, tt.expectedURL)
			}
		})
	}
}
