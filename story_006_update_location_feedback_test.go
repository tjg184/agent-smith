//go:build integration
// +build integration

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestStory006_UpdateLocationFeedback tests that update commands provide
// clear feedback about which location is being updated.
// This is the acceptance test for Story-006 of the Profile-Aware Update Command PRD.
func TestStory006_UpdateLocationFeedback(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-story-006-*")
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

	// Test repository to install (using a small, fast-cloning repository)
	testRepo := "anthropics/skills"

	t.Run("ActiveProfileShowsLocationMessage", func(t *testing.T) {
		// Install components to create a profile
		cmd := exec.Command(binaryPath, "install", "all", testRepo)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to install components: %v\nOutput: %s", err, string(output))
		}
		t.Logf("Install output:\n%s", string(output))

		// Get the profile name by listing profiles
		cmd = exec.Command(binaryPath, "profile", "list")
		output, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to list profiles: %v\nOutput: %s", err, string(output))
		}
		outputStr := string(output)
		t.Logf("Profile list output:\n%s", outputStr)

		// Extract profile name from output (looking for "✓ skills [active]")
		lines := strings.Split(outputStr, "\n")
		var profileName string
		for _, line := range lines {
			if strings.Contains(line, "[active]") {
				// Extract profile name - format is "✓ profile-name [active]"
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					// Skip the "✓" and take the profile name
					profileName = strings.TrimSpace(parts[1])
					break
				}
			}
		}

		if profileName == "" {
			t.Fatalf("Could not determine active profile name from output:\n%s", outputStr)
		}
		t.Logf("Active profile: %s", profileName)

		// Construct expected profile path
		expectedPath := filepath.Join(testHome, ".agent-smith", "profiles", profileName)
		t.Logf("Expected profile path: %s", expectedPath)

		// Run update all - should show location message
		cmd = exec.Command(binaryPath, "update", "all")
		output, err = cmd.CombinedOutput()
		outputStr = string(output)
		t.Logf("Update all output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Update all failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify: Active profile message appears
		expectedActiveMsg := "Using active profile for updates: " + profileName
		if !strings.Contains(outputStr, expectedActiveMsg) {
			t.Errorf("Update output should contain active profile message '%s', got:\n%s", expectedActiveMsg, outputStr)
		}

		// Verify: Location message appears with full path
		expectedLocationMsg := "Updating components in: " + expectedPath
		if !strings.Contains(outputStr, expectedLocationMsg) {
			t.Errorf("Update output should contain location message '%s', got:\n%s", expectedLocationMsg, outputStr)
		}

		// Verify: Location message appears before checking components
		activeIdx := strings.Index(outputStr, expectedActiveMsg)
		locationIdx := strings.Index(outputStr, "Updating components in:")
		checkingIdx := strings.Index(outputStr, "Checking all components for updates")

		if activeIdx == -1 || locationIdx == -1 || checkingIdx == -1 {
			t.Error("Could not find expected messages in output")
		} else if locationIdx < activeIdx {
			t.Error("Location message should appear after active profile message")
		} else if checkingIdx < locationIdx {
			t.Error("Location message should appear before checking components message")
		}
	})

	t.Run("ExplicitProfileFlagShowsLocationMessage", func(t *testing.T) {
		// Verify profile exists from previous test
		profilesDir := filepath.Join(testHome, ".agent-smith", "profiles")
		entries, err := os.ReadDir(profilesDir)
		if err != nil || len(entries) == 0 {
			t.Skip("No profiles available for testing")
		}
		profileName := entries[0].Name()
		t.Logf("Testing with profile: %s", profileName)

		// Deactivate profile to test explicit flag
		cmd := exec.Command(binaryPath, "profile", "deactivate")
		_, _ = cmd.CombinedOutput() // Ignore errors

		// Construct expected profile path
		expectedPath := filepath.Join(testHome, ".agent-smith", "profiles", profileName)

		// Run update with explicit profile flag
		cmd = exec.Command(binaryPath, "update", "all", "--profile", profileName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Update with --profile output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Update with --profile failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify: Specified profile message appears
		expectedProfileMsg := "Using specified profile for updates: " + profileName
		if !strings.Contains(outputStr, expectedProfileMsg) {
			t.Errorf("Update output should contain specified profile message '%s', got:\n%s", expectedProfileMsg, outputStr)
		}

		// Verify: Location message appears with full path
		expectedLocationMsg := "Updating components in: " + expectedPath
		if !strings.Contains(outputStr, expectedLocationMsg) {
			t.Errorf("Update output should contain location message '%s', got:\n%s", expectedLocationMsg, outputStr)
		}
	})

	t.Run("BaseDirectoryHasNoLocationMessage", func(t *testing.T) {
		// Install components to base directory without creating a profile
		targetDir := filepath.Join(testHome, ".agent-smith-base")
		cmd := exec.Command(binaryPath, "install", "all", testRepo, "--target-dir", targetDir)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to install to base directory: %v\nOutput: %s", err, string(output))
		}
		t.Logf("Install to base directory output:\n%s", string(output))

		// Ensure no profile is active
		cmd = exec.Command(binaryPath, "profile", "deactivate")
		_, _ = cmd.CombinedOutput() // Ignore errors

		// Temporarily override HOME to use base directory
		os.Setenv("HOME", tempDir)
		// Create .agent-smith symlink to point to our test directory
		agentSmithDir := filepath.Join(tempDir, ".agent-smith")
		os.RemoveAll(agentSmithDir) // Remove if exists
		if err := os.Symlink(targetDir, agentSmithDir); err != nil {
			t.Fatalf("Failed to create symlink: %v", err)
		}

		// Run update all - should NOT show location message
		cmd = exec.Command(binaryPath, "update", "all")
		output, err = cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Update all (base dir) output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Update all failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify: No profile messages when operating on base directory
		if strings.Contains(outputStr, "Using active profile") {
			t.Errorf("Update output should not mention active profile when none is active, got:\n%s", outputStr)
		}

		if strings.Contains(outputStr, "Using specified profile") {
			t.Errorf("Update output should not mention specified profile when none was provided, got:\n%s", outputStr)
		}

		if strings.Contains(outputStr, "Updating components in:") {
			t.Errorf("Update output should not show location message when no profile is active, got:\n%s", outputStr)
		}
	})

	t.Run("SingleComponentUpdateShowsLocationForProfile", func(t *testing.T) {
		// Verify components exist in a profile
		profilesDir := filepath.Join(testHome, ".agent-smith", "profiles")
		entries, err := os.ReadDir(profilesDir)
		if err != nil || len(entries) == 0 {
			t.Skip("No profiles available for testing")
		}
		profileName := entries[0].Name()

		// Activate the profile
		cmd := exec.Command(binaryPath, "profile", "activate", profileName)
		_, _ = cmd.CombinedOutput() // Ignore errors

		// Get a skill from the profile
		skillsDir := filepath.Join(profilesDir, profileName, "skills")
		skillEntries, err := os.ReadDir(skillsDir)
		if err != nil || len(skillEntries) == 0 {
			t.Skip("No skills in profile for testing")
		}
		skillName := skillEntries[0].Name()
		t.Logf("Testing with skill: %s in profile: %s", skillName, profileName)

		// Construct expected profile path
		expectedPath := filepath.Join(testHome, ".agent-smith", "profiles", profileName)

		// Run update for single component
		cmd = exec.Command(binaryPath, "update", "skills", skillName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Update single component output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Update single component failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify: Active profile message appears
		expectedActiveMsg := "Using active profile for updates: " + profileName
		if !strings.Contains(outputStr, expectedActiveMsg) {
			t.Errorf("Update output should contain active profile message '%s', got:\n%s", expectedActiveMsg, outputStr)
		}

		// Verify: Location message appears with full path
		expectedLocationMsg := "Updating components in: " + expectedPath
		if !strings.Contains(outputStr, expectedLocationMsg) {
			t.Errorf("Update output should contain location message '%s', got:\n%s", expectedLocationMsg, outputStr)
		}

		// Verify: Component checking message appears after location
		expectedCheckMsg := "Checking for updates to skills/" + skillName
		if !strings.Contains(outputStr, expectedCheckMsg) {
			t.Errorf("Update output should contain component checking message '%s', got:\n%s", expectedCheckMsg, outputStr)
		}
	})
}
