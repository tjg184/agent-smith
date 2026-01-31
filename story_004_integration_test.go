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

// TestStory004_ForceCreateNewProfileWithCustomName tests that the --profile flag
// allows users to force creation of a new profile with a custom name.
// This is the acceptance test for Story-004.
func TestStory004_ForceCreateNewProfileWithCustomName(t *testing.T) {
	t.Run("InstallWithCustomProfileName", func(t *testing.T) {
		// Create a temporary directory for testing
		tempDir, err := os.MkdirTemp("", "agent-smith-story-004-test1-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Build agent-smith binary
		binaryPath := filepath.Join(tempDir, "agent-smith")
		buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
		if output, err := buildCmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to build agent-smith: %v\nOutput: %s", err, string(output))
		}

		// Set HOME to test directory to avoid affecting actual configuration
		oldHome := os.Getenv("HOME")
		testHome := tempDir
		os.Setenv("HOME", testHome)
		defer os.Setenv("HOME", oldHome)

		// Install with custom profile name
		installCmd := exec.Command(binaryPath, "install", "all", "anthropics/skills", "--profile", "my-custom-profile")
		output, err := installCmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Install with --profile output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Install with --profile failed: %v\nOutput: %s", err, outputStr)
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
			t.Fatalf("Expected 1 profile after install with --profile, got %d: %v", len(entries), profileNames)
		}

		profileName := entries[0].Name()
		t.Logf("Created profile: %s", profileName)

		// Verify the profile name matches the custom name
		if profileName != "my-custom-profile" {
			t.Errorf("Expected profile name 'my-custom-profile', got: %s", profileName)
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
			t.Errorf("Expected normalized URL 'https://github.com/anthropics/skills', got: %s", sourceURL)
		}

		// Verify output mentions creating the profile
		if !strings.Contains(outputStr, "Creating profile: my-custom-profile") {
			t.Logf("Note: 'Creating profile' message not found in output")
		}
	})

	t.Run("InstallWithCustomProfileCreatesSecondProfile", func(t *testing.T) {
		// Create a temporary directory for testing
		tempDir, err := os.MkdirTemp("", "agent-smith-story-004-test2-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Build agent-smith binary
		binaryPath := filepath.Join(tempDir, "agent-smith")
		buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
		if output, err := buildCmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to build agent-smith: %v\nOutput: %s", err, string(output))
		}

		// Set HOME to test directory to avoid affecting actual configuration
		oldHome := os.Getenv("HOME")
		testHome := tempDir
		os.Setenv("HOME", testHome)
		defer os.Setenv("HOME", oldHome)

		// First install with default behavior (should create auto-named profile)
		installCmd1 := exec.Command(binaryPath, "install", "all", "anthropics/skills")
		output, err := installCmd1.CombinedOutput()
		outputStr := string(output)

		t.Logf("First install output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("First install failed: %v\nOutput: %s", err, outputStr)
		}

		// Second install with custom profile name from same repo
		installCmd2 := exec.Command(binaryPath, "install", "all", "anthropics/skills", "--profile", "work-profile")
		output, err = installCmd2.CombinedOutput()
		outputStr = string(output)

		t.Logf("Second install with --profile output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Second install with --profile failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify two profiles exist
		profilesDir := filepath.Join(testHome, ".agent-smith", "profiles")
		entries, err := os.ReadDir(profilesDir)
		if err != nil {
			t.Fatalf("Failed to read profiles directory: %v", err)
		}

		if len(entries) != 2 {
			var profileNames []string
			for _, entry := range entries {
				if entry.IsDir() {
					profileNames = append(profileNames, entry.Name())
				}
			}
			t.Fatalf("Expected 2 profiles (auto-named and custom), got %d: %v", len(entries), profileNames)
		}

		// Verify one profile is named "work-profile"
		hasWorkProfile := false
		for _, entry := range entries {
			if entry.Name() == "work-profile" {
				hasWorkProfile = true
				break
			}
		}

		if !hasWorkProfile {
			var profileNames []string
			for _, entry := range entries {
				if entry.IsDir() {
					profileNames = append(profileNames, entry.Name())
				}
			}
			t.Errorf("Expected to find 'work-profile', profiles found: %v", profileNames)
		}
	})

	t.Run("InstallWithExistingCustomNameShowsError", func(t *testing.T) {
		// Create a temporary directory for testing
		tempDir, err := os.MkdirTemp("", "agent-smith-story-004-test3-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Build agent-smith binary
		binaryPath := filepath.Join(tempDir, "agent-smith")
		buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
		if output, err := buildCmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to build agent-smith: %v\nOutput: %s", err, string(output))
		}

		// Set HOME to test directory to avoid affecting actual configuration
		oldHome := os.Getenv("HOME")
		testHome := tempDir
		os.Setenv("HOME", testHome)
		defer os.Setenv("HOME", oldHome)

		// First install with custom profile name
		installCmd1 := exec.Command(binaryPath, "install", "all", "anthropics/skills", "--profile", "test-profile")
		output, err := installCmd1.CombinedOutput()
		outputStr := string(output)

		t.Logf("First install with --profile output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("First install failed: %v\nOutput: %s", err, outputStr)
		}

		// Second install with same custom profile name (should fail)
		installCmd2 := exec.Command(binaryPath, "install", "all", "anthropics/skills", "--profile", "test-profile")
		output, err = installCmd2.CombinedOutput()
		outputStr = string(output)

		t.Logf("Second install with same --profile output:\n%s", outputStr)

		// Should fail with error
		if err == nil {
			t.Fatalf("Expected error when using existing profile name, but command succeeded")
		}

		// Verify error message mentions the profile already exists
		if !strings.Contains(outputStr, "already exists") {
			t.Errorf("Expected error message to mention 'already exists', got: %s", outputStr)
		}

		// Verify only one profile exists
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
			t.Fatalf("Expected 1 profile after failed duplicate name, got %d: %v", len(entries), profileNames)
		}
	})

	t.Run("ProfileFlagCannotBeUsedWithTargetDir", func(t *testing.T) {
		// Create a temporary directory for testing
		tempDir, err := os.MkdirTemp("", "agent-smith-story-004-test4-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Build agent-smith binary
		binaryPath := filepath.Join(tempDir, "agent-smith")
		buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
		if output, err := buildCmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to build agent-smith: %v\nOutput: %s", err, string(output))
		}

		// Set HOME to test directory to avoid affecting actual configuration
		oldHome := os.Getenv("HOME")
		testHome := tempDir
		os.Setenv("HOME", testHome)
		defer os.Setenv("HOME", oldHome)

		// Attempt to use both --profile and --target-dir flags (should fail)
		installCmd := exec.Command(binaryPath, "install", "all", "anthropics/skills", "--profile", "test", "--target-dir", "./custom")
		output, err := installCmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Install with both flags output:\n%s", outputStr)

		// Should fail with error
		if err == nil {
			t.Fatalf("Expected error when using both --profile and --target-dir, but command succeeded")
		}

		// Verify error message mentions the conflict
		if !strings.Contains(outputStr, "Cannot specify both") {
			t.Errorf("Expected error message about conflicting flags, got: %s", outputStr)
		}
	})

	t.Run("InstallWithoutProfileFlagStillReusesExistingProfile", func(t *testing.T) {
		// Create a temporary directory for testing
		tempDir, err := os.MkdirTemp("", "agent-smith-story-004-test5-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Build agent-smith binary
		binaryPath := filepath.Join(tempDir, "agent-smith")
		buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
		if output, err := buildCmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to build agent-smith: %v\nOutput: %s", err, string(output))
		}

		// Set HOME to test directory to avoid affecting actual configuration
		oldHome := os.Getenv("HOME")
		testHome := tempDir
		os.Setenv("HOME", testHome)
		defer os.Setenv("HOME", oldHome)

		// First install creates auto-named profile
		installCmd1 := exec.Command(binaryPath, "install", "all", "anthropics/skills")
		output, err := installCmd1.CombinedOutput()
		outputStr := string(output)

		t.Logf("First install output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("First install failed: %v\nOutput: %s", err, outputStr)
		}

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
		t.Logf("First profile name: %s", firstProfileName)

		// Second install without --profile flag (should reuse existing profile)
		installCmd2 := exec.Command(binaryPath, "install", "all", "anthropics/skills")
		output, err = installCmd2.CombinedOutput()
		outputStr = string(output)

		t.Logf("Second install output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Second install failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify still only one profile exists
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

		// Verify output mentions finding existing profile
		if !strings.Contains(outputStr, "Found existing profile") {
			t.Logf("Note: 'Found existing profile' message not in output")
		}
	})
}
