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

// TestStory004_UpdateCommandsWorkOnBaseDirectory tests that update commands
// work correctly on base directory components when no profile is active.
// This is the acceptance test for Story-004 of the Profile-Aware Update Command PRD.
func TestStory004_UpdateCommandsWorkOnBaseDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-story-004-*")
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

	t.Run("UpdateAllChecksBaseDirectoryWhenNoProfile", func(t *testing.T) {
		// Install components to base directory without creating a profile
		// Use --target-dir to bypass profile creation
		targetDir := filepath.Join(testHome, ".agent-smith")
		cmd := exec.Command(binaryPath, "install", "all", testRepo, "--target-dir", targetDir)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to install components: %v\nOutput: %s", err, string(output))
		}
		t.Logf("Install output:\n%s", string(output))

		// Verify components are in base directory
		baseSkillsDir := filepath.Join(testHome, ".agent-smith", "skills")
		entries, err := os.ReadDir(baseSkillsDir)
		if err != nil || len(entries) == 0 {
			t.Fatalf("No skills found in base directory after install: %v", err)
		}
		t.Logf("Found %d skills in base directory", len(entries))

		// Ensure no profile is active
		cmd = exec.Command(binaryPath, "profile", "deactivate")
		_, _ = cmd.CombinedOutput() // Ignore errors

		// Run update all - should check base directory
		cmd = exec.Command(binaryPath, "update", "all")
		output, err = cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Update all output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Update all failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify: No profile-related messages
		if strings.Contains(outputStr, "Using active profile") {
			t.Errorf("Update output should not mention active profile when none is active, got:\n%s", outputStr)
		}

		if strings.Contains(outputStr, "Using specified profile") {
			t.Errorf("Update output should not mention specified profile when none was provided, got:\n%s", outputStr)
		}

		// Verify: Components were checked
		if strings.Contains(outputStr, "Total components checked: 0") {
			t.Errorf("Update should have checked components in base directory, got:\n%s", outputStr)
		}

		// Verify: Update summary shows correct component counts
		if !strings.Contains(outputStr, "Update Summary") {
			t.Errorf("Update output should include update summary, got:\n%s", outputStr)
		}
	})

	t.Run("UpdateSingleComponentChecksBaseDirectoryWhenNoProfile", func(t *testing.T) {
		// Verify components exist in base directory
		baseSkillsDir := filepath.Join(testHome, ".agent-smith", "skills")
		entries, err := os.ReadDir(baseSkillsDir)
		if err != nil || len(entries) == 0 {
			t.Skip("No skills in base directory for testing")
		}
		skillName := entries[0].Name()
		t.Logf("Testing with skill: %s", skillName)

		// Ensure no profile is active
		cmd := exec.Command(binaryPath, "profile", "deactivate")
		_, _ = cmd.CombinedOutput() // Ignore errors

		// Update single component - should check base directory
		cmd = exec.Command(binaryPath, "update", "skills", skillName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Update single component output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Update single component failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify: No profile-related messages
		if strings.Contains(outputStr, "Using active profile") {
			t.Errorf("Update output should not mention active profile when none is active, got:\n%s", outputStr)
		}

		if strings.Contains(outputStr, "Using specified profile") {
			t.Errorf("Update output should not mention specified profile when none was provided, got:\n%s", outputStr)
		}

		// Verify: Component was checked
		expectedMsg := "Checking for updates to skills/" + skillName
		if !strings.Contains(outputStr, expectedMsg) {
			t.Errorf("Update output should mention checking component %s, got:\n%s", skillName, outputStr)
		}
	})

	t.Run("UpdateSummaryShowsBaseDirectoryComponents", func(t *testing.T) {
		// Count components in base directory
		baseSkillsDir := filepath.Join(testHome, ".agent-smith", "skills")
		skillEntries, err := os.ReadDir(baseSkillsDir)
		if err != nil {
			t.Skip("No skills directory for testing")
		}
		expectedSkillCount := len(skillEntries)
		t.Logf("Expected %d skills in base directory", expectedSkillCount)

		// Ensure no profile is active
		cmd := exec.Command(binaryPath, "profile", "deactivate")
		_, _ = cmd.CombinedOutput() // Ignore errors

		// Run update all
		cmd = exec.Command(binaryPath, "update", "all")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Update all output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Update all failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify: Summary shows correct component count
		if expectedSkillCount > 0 && strings.Contains(outputStr, "Total components checked: 0") {
			t.Errorf("Update summary should show %d components, got:\n%s", expectedSkillCount, outputStr)
		}
	})

	t.Run("BackwardCompatibilityWithoutProfiles", func(t *testing.T) {
		// This test verifies that users not using profiles have the same experience as before

		// Verify no profiles directory exists
		profilesDir := filepath.Join(testHome, ".agent-smith", "profiles")
		if _, err := os.Stat(profilesDir); err == nil {
			// If profiles dir exists from previous tests, ensure it's empty or no active profile
			cmd := exec.Command(binaryPath, "profile", "deactivate")
			_, _ = cmd.CombinedOutput()
		}

		// Verify components exist in base directory
		baseSkillsDir := filepath.Join(testHome, ".agent-smith", "skills")
		entries, err := os.ReadDir(baseSkillsDir)
		if err != nil || len(entries) == 0 {
			t.Skip("No skills in base directory for testing")
		}

		// Update all should work seamlessly
		cmd := exec.Command(binaryPath, "update", "all")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Update all output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Update all failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify: Clean output without profile messages
		if strings.Contains(outputStr, "profile") || strings.Contains(outputStr, "Profile") {
			t.Errorf("Update output should not mention profiles for users not using them, got:\n%s", outputStr)
		}

		// Verify: Functional update summary
		if !strings.Contains(outputStr, "Update Summary") {
			t.Errorf("Update output should include summary, got:\n%s", outputStr)
		}
	})
}
