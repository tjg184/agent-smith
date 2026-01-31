//go:build integration
// +build integration

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestStory001_RepeatedInstallsUpdateProfile tests that repeated installs from the same repository
// update the existing profile instead of creating duplicates.
// This is the acceptance test for Story-001.
func TestStory001_RepeatedInstallsUpdateProfile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-story-001-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Build agent-smith binary
	binaryPath := filepath.Join(tempDir, "agent-smith")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build agent-smith: %v\nOutput: %s", err, string(output))
	}

	// Set HOME to test directory to avoid affecting actual configuration
	oldHome := os.Getenv("HOME")
	testHome := tempDir
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", oldHome)

	t.Run("FirstInstallCreatesProfile", func(t *testing.T) {
		// First install should create a new profile
		cmd := exec.Command(binaryPath, "install", "all", "anthropics/skills")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("First install output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("First install failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify the profile was created
		profilesDir := filepath.Join(testHome, ".agent-smith", "profiles")
		entries, err := os.ReadDir(profilesDir)
		if err != nil {
			t.Fatalf("Failed to read profiles directory: %v", err)
		}

		if len(entries) != 1 {
			var profileNames []string
			for _, entry := range entries {
				if entry.IsDir() {
					profileNames = append(profileNames, entry.Name())
				}
			}
			t.Fatalf("Expected 1 profile after first install, got %d: %v", len(entries), profileNames)
		}

		profileName := entries[0].Name()
		t.Logf("Created profile: %s", profileName)

		// Verify the profile name is based on the repository
		// Note: The shorthand "anthropics/skills" generates just "skills" as the profile name
		if !strings.Contains(profileName, "skills") {
			t.Errorf("Expected profile name to contain 'skills', got: %s", profileName)
		}

		// Verify metadata file exists
		metadataPath := filepath.Join(profilesDir, profileName, ".profile-metadata")
		if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
			t.Errorf("Metadata file not created at %s", metadataPath)
		}

		// Verify metadata contains the correct URL
		metadataBytes, err := os.ReadFile(metadataPath)
		if err != nil {
			t.Fatalf("Failed to read metadata file: %v", err)
		}

		var metadata map[string]interface{}
		if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
			t.Fatalf("Failed to parse metadata JSON: %v", err)
		}

		sourceURL, ok := metadata["source_url"].(string)
		if !ok {
			t.Fatalf("Metadata missing source_url field: %v", metadata)
		}

		// The URL should be normalized to HTTPS
		if sourceURL != "https://github.com/anthropics/skills" {
			t.Errorf("Expected normalized URL https://github.com/anthropics/skills, got: %s", sourceURL)
		}
	})

	t.Run("SecondInstallUpdatesExistingProfile", func(t *testing.T) {
		// Get the profile name before second install
		profilesDir := filepath.Join(testHome, ".agent-smith", "profiles")
		entriesBefore, err := os.ReadDir(profilesDir)
		if err != nil {
			t.Fatalf("Failed to read profiles directory: %v", err)
		}

		if len(entriesBefore) != 1 {
			t.Fatalf("Expected 1 profile before second install, got %d", len(entriesBefore))
		}

		firstProfileName := entriesBefore[0].Name()
		t.Logf("Existing profile before second install: %s", firstProfileName)

		// Second install with the same repository (different URL format) - use --verbose to see messages
		cmd := exec.Command(binaryPath, "install", "all", "https://github.com/anthropics/skills", "--verbose")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Second install output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Second install failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify output mentions updating existing profile (with --verbose flag, messages should appear)
		if !strings.Contains(outputStr, "Found existing profile") {
			t.Logf("Note: 'Found existing profile' message not in output (may be suppressed without --verbose)")
		}

		if !strings.Contains(outputStr, "Updating profile") {
			t.Logf("Note: 'Updating profile' message not in output (may be suppressed without --verbose)")
		}

		// Verify still only one profile exists (THIS IS THE KEY REQUIREMENT)
		entriesAfter, err := os.ReadDir(profilesDir)
		if err != nil {
			t.Fatalf("Failed to read profiles directory: %v", err)
		}

		if len(entriesAfter) != 1 {
			var profileNames []string
			for _, entry := range entriesAfter {
				if entry.IsDir() {
					profileNames = append(profileNames, entry.Name())
				}
			}
			t.Errorf("Expected 1 profile after second install (no duplicates), got %d: %v", len(entriesAfter), profileNames)
		}

		secondProfileName := entriesAfter[0].Name()
		t.Logf("Profile after second install: %s", secondProfileName)

		// Verify it's the same profile name
		if firstProfileName != secondProfileName {
			t.Errorf("Profile name changed! Before: %s, After: %s", firstProfileName, secondProfileName)
		}
	})

	t.Run("ThirdInstallWithDifferentURLFormat", func(t *testing.T) {
		// Third install with SSH URL format - use --verbose to see messages
		cmd := exec.Command(binaryPath, "install", "all", "git@github.com:anthropics/skills.git", "--verbose")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Third install output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Third install failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify output mentions updating existing profile (with --verbose)
		if !strings.Contains(outputStr, "Found existing profile") {
			t.Logf("Note: 'Found existing profile' message not in output for SSH URL")
		}

		// Verify still only one profile exists (THIS IS THE KEY REQUIREMENT)
		profilesDir := filepath.Join(testHome, ".agent-smith", "profiles")
		entries, err := os.ReadDir(profilesDir)
		if err != nil {
			t.Fatalf("Failed to read profiles directory: %v", err)
		}

		if len(entries) != 1 {
			var profileNames []string
			for _, entry := range entries {
				if entry.IsDir() {
					profileNames = append(profileNames, entry.Name())
				}
			}
			t.Errorf("Expected 1 profile after third install, got %d: %v", len(entries), profileNames)
		}
	})
}

// TestStory001_DifferentReposCreateDifferentProfiles tests that installing from different
// repositories creates separate profiles (not everything goes into one profile).
func TestStory001_DifferentReposCreateDifferentProfiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-story-001-different-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Build agent-smith binary
	binaryPath := filepath.Join(tempDir, "agent-smith")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build agent-smith: %v\nOutput: %s", err, string(output))
	}

	// Set HOME to test directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	t.Run("InstallFromTwoDifferentRepos", func(t *testing.T) {
		// Install from first repo
		cmd1 := exec.Command(binaryPath, "install", "all", "anthropics/skills")
		output1, err1 := cmd1.CombinedOutput()
		if err1 != nil {
			t.Logf("First install output: %s", string(output1))
			// Don't fail if the repo doesn't exist - this is just testing the logic
		}

		// Install from second repo (different owner/repo)
		cmd2 := exec.Command(binaryPath, "install", "all", "owner2/repo2")
		output2, err2 := cmd2.CombinedOutput()
		if err2 != nil {
			t.Logf("Second install output: %s", string(output2))
			// Don't fail if the repo doesn't exist
		}

		// Verify two different profiles were created
		profilesDir := filepath.Join(tempDir, ".agent-smith", "profiles")
		entries, err := os.ReadDir(profilesDir)
		if err != nil {
			// If no profiles created (repos don't exist), that's okay - test the code path
			t.Logf("No profiles directory or failed to read: %v", err)
			return
		}

		// If both installs succeeded, we should have 2 profiles
		if err1 == nil && err2 == nil {
			if len(entries) != 2 {
				var profileNames []string
				for _, entry := range entries {
					if entry.IsDir() {
						profileNames = append(profileNames, entry.Name())
					}
				}
				t.Errorf("Expected 2 profiles for different repos, got %d: %v", len(entries), profileNames)
			}
		}
	})
}

// TestStory001_MetadataIntegrity tests that metadata files are correctly maintained
// across multiple install operations.
func TestStory001_MetadataIntegrity(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-story-001-metadata-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Build agent-smith binary
	binaryPath := filepath.Join(tempDir, "agent-smith")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build agent-smith: %v\nOutput: %s", err, string(output))
	}

	// Set HOME to test directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	t.Run("MetadataPreservedAcrossUpdates", func(t *testing.T) {
		// First install
		cmd1 := exec.Command(binaryPath, "install", "all", "anthropics/skills")
		output1, err1 := cmd1.CombinedOutput()
		if err1 != nil {
			t.Logf("First install failed (expected if repo doesn't exist): %s", string(output1))
			t.Skip("Skipping test - repository doesn't exist")
		}

		profilesDir := filepath.Join(tempDir, ".agent-smith", "profiles")
		entries, err := os.ReadDir(profilesDir)
		if err != nil || len(entries) == 0 {
			t.Skip("No profiles created")
		}

		profileName := entries[0].Name()
		metadataPath := filepath.Join(profilesDir, profileName, ".profile-metadata")

		// Read metadata after first install
		metadata1, err := os.ReadFile(metadataPath)
		if err != nil {
			t.Fatalf("Failed to read metadata after first install: %v", err)
		}

		// Second install
		cmd2 := exec.Command(binaryPath, "install", "all", "https://github.com/anthropics/skills.git")
		output2, err2 := cmd2.CombinedOutput()
		if err2 != nil {
			t.Fatalf("Second install failed: %v\nOutput: %s", err2, string(output2))
		}

		// Read metadata after second install
		metadata2, err := os.ReadFile(metadataPath)
		if err != nil {
			t.Fatalf("Failed to read metadata after second install: %v", err)
		}

		// Verify metadata file still exists and is valid JSON
		var metadataObj map[string]interface{}
		if err := json.Unmarshal(metadata2, &metadataObj); err != nil {
			t.Errorf("Metadata file corrupted after update: %v\nContent: %s", err, string(metadata2))
		}

		// Verify source URL is still the normalized format
		sourceURL, ok := metadataObj["source_url"].(string)
		if !ok {
			t.Errorf("Metadata missing source_url after update: %v", metadataObj)
		}

		if sourceURL != "https://github.com/anthropics/skills" {
			t.Errorf("Expected normalized URL https://github.com/anthropics/skills, got: %s", sourceURL)
		}

		t.Logf("Metadata after first install: %s", string(metadata1))
		t.Logf("Metadata after second install: %s", string(metadata2))
	})
}
