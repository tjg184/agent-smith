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

// TestStory003_UpdateSingleComponentRespectsActiveProfile tests that updating
// a single component respects the active profile setting.
// This is the acceptance test for Story-003 of the Profile-Aware Update Command PRD.
func TestStory003_UpdateSingleComponentRespectsActiveProfile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-story-003-*")
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

	t.Run("UpdateSingleSkillFromActiveProfile", func(t *testing.T) {
		// Step 1: Install components to create a profile
		cmd := exec.Command(binaryPath, "install", "all", testRepo)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to install components: %v\nOutput: %s", err, string(output))
		}
		t.Logf("Install output:\n%s", string(output))

		// Step 2: Get the profile name
		profilesDir := filepath.Join(testHome, ".agent-smith", "profiles")
		entries, err := os.ReadDir(profilesDir)
		if err != nil {
			t.Fatalf("Failed to read profiles directory: %v", err)
		}
		if len(entries) == 0 {
			t.Fatal("No profiles created after install")
		}
		profileName := entries[0].Name()
		t.Logf("Created profile: %s", profileName)

		// Note: install all automatically activates the profile, so we don't need to activate it again

		// Step 3: List skills in the profile to find one to update
		profileSkillsDir := filepath.Join(profilesDir, profileName, "skills")
		skillEntries, err := os.ReadDir(profileSkillsDir)
		if err != nil || len(skillEntries) == 0 {
			t.Fatalf("No skills found in profile: %v", err)
		}
		skillName := skillEntries[0].Name()
		t.Logf("Testing with skill: %s", skillName)

		// Step 4: Update a single skill - should respect active profile
		cmd = exec.Command(binaryPath, "update", "skills", skillName)
		output, err = cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Update output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Failed to update skill: %v\nOutput: %s", err, outputStr)
		}

		// Verify: Output should indicate it's using the active profile
		if !strings.Contains(outputStr, "Using active profile for updates: "+profileName) {
			t.Errorf("Update output should mention active profile '%s', but got:\n%s", profileName, outputStr)
		}

		// Verify: The skill should still exist in the profile directory
		skillPath := filepath.Join(profileSkillsDir, skillName)
		if _, err := os.Stat(skillPath); os.IsNotExist(err) {
			t.Errorf("Skill %s should still exist in profile directory after update", skillName)
		}
	})

	t.Run("UpdateSingleAgentFromActiveProfile", func(t *testing.T) {
		// Similar test for agents, if any exist in the test profile
		profilesDir := filepath.Join(testHome, ".agent-smith", "profiles")
		entries, err := os.ReadDir(profilesDir)
		if err != nil || len(entries) == 0 {
			t.Skip("No profile available for testing")
		}
		profileName := entries[0].Name()

		profileAgentsDir := filepath.Join(profilesDir, profileName, "agents")
		agentEntries, err := os.ReadDir(profileAgentsDir)
		if err != nil || len(agentEntries) == 0 {
			t.Skip("No agents found in profile for testing")
		}
		agentName := agentEntries[0].Name()
		t.Logf("Testing with agent: %s", agentName)

		// Update a single agent - should respect active profile
		cmd := exec.Command(binaryPath, "update", "agents", agentName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Update output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Failed to update agent: %v\nOutput: %s", err, outputStr)
		}

		// Verify: Output should indicate it's using the active profile
		if !strings.Contains(outputStr, "Using active profile for updates: "+profileName) {
			t.Errorf("Update output should mention active profile '%s', but got:\n%s", profileName, outputStr)
		}
	})

	t.Run("UpdateSingleCommandFromActiveProfile", func(t *testing.T) {
		// Similar test for commands, if any exist in the test profile
		profilesDir := filepath.Join(testHome, ".agent-smith", "profiles")
		entries, err := os.ReadDir(profilesDir)
		if err != nil || len(entries) == 0 {
			t.Skip("No profile available for testing")
		}
		profileName := entries[0].Name()

		profileCommandsDir := filepath.Join(profilesDir, profileName, "commands")
		commandEntries, err := os.ReadDir(profileCommandsDir)
		if err != nil || len(commandEntries) == 0 {
			t.Skip("No commands found in profile for testing")
		}
		commandName := commandEntries[0].Name()
		t.Logf("Testing with command: %s", commandName)

		// Update a single command - should respect active profile
		cmd := exec.Command(binaryPath, "update", "commands", commandName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Update output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Failed to update command: %v\nOutput: %s", err, outputStr)
		}

		// Verify: Output should indicate it's using the active profile
		if !strings.Contains(outputStr, "Using active profile for updates: "+profileName) {
			t.Errorf("Update output should mention active profile '%s', but got:\n%s", profileName, outputStr)
		}
	})

	t.Run("UpdateWithExplicitProfileFlag", func(t *testing.T) {
		// Test that --profile flag works for single component updates
		profilesDir := filepath.Join(testHome, ".agent-smith", "profiles")
		entries, err := os.ReadDir(profilesDir)
		if err != nil || len(entries) == 0 {
			t.Skip("No profile available for testing")
		}
		profileName := entries[0].Name()

		// Deactivate the profile first
		cmd := exec.Command(binaryPath, "profile", "deactivate")
		_, _ = cmd.CombinedOutput() // Ignore errors

		profileSkillsDir := filepath.Join(profilesDir, profileName, "skills")
		skillEntries, err := os.ReadDir(profileSkillsDir)
		if err != nil || len(skillEntries) == 0 {
			t.Skip("No skills found in profile for testing")
		}
		skillName := skillEntries[0].Name()

		// Update with explicit --profile flag
		cmd = exec.Command(binaryPath, "update", "skills", skillName, "--profile", profileName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Update output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Failed to update with --profile flag: %v\nOutput: %s", err, outputStr)
		}

		// Verify: Output should indicate it's using the specified profile
		if !strings.Contains(outputStr, "Using specified profile for updates: "+profileName) {
			t.Errorf("Update output should mention specified profile '%s', but got:\n%s", profileName, outputStr)
		}
	})

	t.Run("UpdateWithoutActiveProfileUsesBaseDirectory", func(t *testing.T) {
		// This test verifies backward compatibility when no profile is active
		// We'll manually create a skill in the base directory for testing

		// Deactivate any active profile first
		cmd := exec.Command(binaryPath, "profile", "deactivate")
		_, _ = cmd.CombinedOutput() // Ignore errors if no profile was active

		// Manually create a test skill in base directory
		baseSkillsDir := filepath.Join(testHome, ".agent-smith", "skills")
		testSkillDir := filepath.Join(baseSkillsDir, "test-skill")
		if err := os.MkdirAll(testSkillDir, 0755); err != nil {
			t.Fatalf("Failed to create test skill directory: %v", err)
		}

		// Create a minimal lock file entry for the test skill
		lockFilePath := filepath.Join(testHome, ".agent-smith", ".skill-lock.json")
		lockData := `{
  "test-skill": {
    "name": "test-skill",
    "source": "https://github.com/anthropics/skills",
    "sourceType": "github",
    "sourceUrl": "https://github.com/anthropics/skills",
    "commitHash": "abc123",
    "components": 1,
    "detection": "single"
  }
}`
		if err := os.WriteFile(lockFilePath, []byte(lockData), 0644); err != nil {
			t.Fatalf("Failed to create lock file: %v", err)
		}

		// Update should work without profile
		cmd = exec.Command(binaryPath, "update", "skills", "test-skill")
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Update output:\n%s", outputStr)

		// The update might fail to fetch (expected for fake data), but we can still verify it's using base directory
		// Verify: Output should NOT mention any profile
		if strings.Contains(outputStr, "Using active profile") {
			t.Errorf("Update output should not mention active profile when none is active, but got:\n%s", outputStr)
		}

		if strings.Contains(outputStr, "Using specified profile") {
			t.Errorf("Update output should not mention specified profile when none was provided, but got:\n%s", outputStr)
		}
	})
}

// Note: The UpdateDetector profile handling is tested through the integration tests above.
// The unit tests for internal behavior are in the updater package tests.
