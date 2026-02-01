//go:build integration
// +build integration

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestProfileAddPreservesLockFileEntries tests that the `profile add` command
// preserves lock file entries when copying components from base installation to a profile.
// This is the acceptance test for Story-004 from 20260201-0052-profile-component-copy.md
func TestProfileAddPreservesLockFileEntries(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-profile-add-lock-*")
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

	t.Run("ProfileAddPreservesLockEntry", func(t *testing.T) {
		// Step 1: Install a skill from Git to the base directory
		t.Log("Step 1: Installing skill from Git to base directory...")
		installCmd := exec.Command(binaryPath, "install", "skill", "anthropics/skills", "frontend-design")
		output, err := installCmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Install output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Failed to install skill: %v\nOutput: %s", err, outputStr)
		}

		// Step 2: Verify the lock file entry exists in base installation
		t.Log("Step 2: Verifying lock file entry in base installation...")
		baseLockPath := filepath.Join(testHome, ".agent-smith", ".skill-lock.json")
		baseLockData, err := os.ReadFile(baseLockPath)
		if err != nil {
			t.Fatalf("Failed to read base lock file: %v", err)
		}

		var baseLockFile struct {
			Version int                               `json:"version"`
			Skills  map[string]map[string]interface{} `json:"skills"`
		}
		if err := json.Unmarshal(baseLockData, &baseLockFile); err != nil {
			t.Fatalf("Failed to parse base lock file: %v", err)
		}

		// Verify frontend-design entry exists
		frontendDesignEntry, exists := baseLockFile.Skills["frontend-design"]
		if !exists {
			t.Fatalf("frontend-design entry not found in base lock file. Available entries: %v", getMapKeys(baseLockFile.Skills))
		}

		// Verify the entry has required fields
		requiredFields := []string{"sourceUrl", "commitHash", "installedAt"}
		for _, field := range requiredFields {
			if _, ok := frontendDesignEntry[field]; !ok {
				t.Errorf("Lock entry missing required field '%s'. Entry: %v", field, frontendDesignEntry)
			}
		}

		// Store the original commit hash for later comparison
		originalCommitHash, _ := frontendDesignEntry["commitHash"].(string)
		t.Logf("Base lock entry commitHash: %s", originalCommitHash)

		// Step 3: Create a profile
		t.Log("Step 3: Creating test profile...")
		createProfileCmd := exec.Command(binaryPath, "profile", "create", "test-profile")
		output, err = createProfileCmd.CombinedOutput()
		outputStr = string(output)
		t.Logf("Create profile output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Failed to create profile: %v\nOutput: %s", err, outputStr)
		}

		// Step 4: Add the skill to the profile using 'profile add'
		t.Log("Step 4: Adding skill to profile using 'profile add'...")
		addCmd := exec.Command(binaryPath, "profile", "add", "skills", "test-profile", "frontend-design")
		output, err = addCmd.CombinedOutput()
		outputStr = string(output)
		t.Logf("Profile add output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Failed to add component to profile: %v\nOutput: %s", err, outputStr)
		}

		// Step 5: Verify the lock file entry was copied to the profile
		t.Log("Step 5: Verifying lock file entry was copied to profile...")
		profileLockPath := filepath.Join(testHome, ".agent-smith", "profiles", "test-profile", ".skill-lock.json")
		profileLockData, err := os.ReadFile(profileLockPath)
		if err != nil {
			t.Fatalf("Failed to read profile lock file at %s: %v", profileLockPath, err)
		}

		var profileLockFile struct {
			Version int                               `json:"version"`
			Skills  map[string]map[string]interface{} `json:"skills"`
		}
		if err := json.Unmarshal(profileLockData, &profileLockFile); err != nil {
			t.Fatalf("Failed to parse profile lock file: %v", err)
		}

		// Verify frontend-design entry exists in profile
		profileEntry, exists := profileLockFile.Skills["frontend-design"]
		if !exists {
			t.Fatalf("frontend-design entry not found in profile lock file. Available entries: %v", getMapKeys(profileLockFile.Skills))
		}

		// Step 6: Verify all metadata fields were preserved
		t.Log("Step 6: Verifying metadata fields were preserved...")
		for _, field := range requiredFields {
			if _, ok := profileEntry[field]; !ok {
				t.Errorf("Profile lock entry missing required field '%s'. Entry: %v", field, profileEntry)
			}
		}

		// Verify commit hash matches
		profileCommitHash, _ := profileEntry["commitHash"].(string)
		if profileCommitHash != originalCommitHash {
			t.Errorf("Commit hash mismatch. Base: %s, Profile: %s", originalCommitHash, profileCommitHash)
		}

		// Step 7: Verify the component files were copied
		t.Log("Step 7: Verifying component files were copied...")
		profileComponentPath := filepath.Join(testHome, ".agent-smith", "profiles", "test-profile", "skills", "frontend-design")
		if _, err := os.Stat(profileComponentPath); os.IsNotExist(err) {
			t.Fatalf("Component directory not found in profile at %s", profileComponentPath)
		}

		// Check for SKILL.md file (or similar marker file)
		skillFiles, err := os.ReadDir(profileComponentPath)
		if err != nil {
			t.Fatalf("Failed to read component directory: %v", err)
		}
		if len(skillFiles) == 0 {
			t.Errorf("Component directory is empty at %s", profileComponentPath)
		}

		t.Log("✓ Profile add successfully preserved lock file entry")
		t.Logf("✓ Component is now updateable in profile 'test-profile'")
	})

	t.Run("ProfileAddHandlesMissingLockEntry", func(t *testing.T) {
		// Test that profile add handles components without lock entries gracefully
		// This tests backward compatibility with manually created components

		// Step 1: Create a profile
		t.Log("Step 1: Creating test profile for manual component...")
		createProfileCmd := exec.Command(binaryPath, "profile", "create", "manual-profile")
		output, err := createProfileCmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Create profile output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Failed to create profile: %v\nOutput: %s", err, outputStr)
		}

		// Step 2: Create a manual skill (no Git source)
		t.Log("Step 2: Creating manual skill...")
		manualSkillDir := filepath.Join(testHome, ".agent-smith", "skills", "manual-skill")
		if err := os.MkdirAll(manualSkillDir, 0755); err != nil {
			t.Fatalf("Failed to create manual skill directory: %v", err)
		}

		// Write a simple SKILL.md file
		skillContent := `# Manual Skill

This is a manually created skill without Git source.
`
		if err := os.WriteFile(filepath.Join(manualSkillDir, "SKILL.md"), []byte(skillContent), 0644); err != nil {
			t.Fatalf("Failed to write skill file: %v", err)
		}

		// Step 3: Add the manual skill to the profile
		t.Log("Step 3: Adding manual skill to profile...")
		addCmd := exec.Command(binaryPath, "profile", "add", "skills", "manual-profile", "manual-skill")
		output, err = addCmd.CombinedOutput()
		outputStr = string(output)
		t.Logf("Profile add output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Failed to add manual component to profile: %v\nOutput: %s", err, outputStr)
		}

		// Step 4: Verify the component was copied even without lock entry
		t.Log("Step 4: Verifying manual component was copied...")
		profileComponentPath := filepath.Join(testHome, ".agent-smith", "profiles", "manual-profile", "skills", "manual-skill")
		if _, err := os.Stat(profileComponentPath); os.IsNotExist(err) {
			t.Fatalf("Manual component directory not found in profile at %s", profileComponentPath)
		}

		// Check that SKILL.md was copied
		skillPath := filepath.Join(profileComponentPath, "SKILL.md")
		if _, err := os.Stat(skillPath); os.IsNotExist(err) {
			t.Errorf("SKILL.md file not found in profile component at %s", skillPath)
		}

		t.Log("✓ Profile add successfully handled manual component without lock entry")
	})
}

// Helper function to get keys from a map
func getMapKeys(m map[string]map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
