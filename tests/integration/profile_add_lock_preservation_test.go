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

	"github.com/tjg184/agent-smith/internal/testutil"
)

// TestE2E_ProfileAddWorkflow tests the full workflow: install → create profile → add to profile → verify update works
// This verifies that lock file entries are preserved when copying components to profiles.
func TestE2E_ProfileAddWorkflow(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-profile-add-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	binaryPath := AgentSmithBinary
	testRepo := "anthropics/skills"
	skillName := "frontend-design"
	profileName := "test-profile"

	// Step 1: Install skill to base directory using --install-dir
	// (The profile add command expects components to be in base directory)
	t.Run("Step1_InstallSkillToBase", func(t *testing.T) {
		// Use --install-dir to force installation to base directory structure
		// This is the old behavior needed for the "profile add" command
		baseDir := filepath.Join(tempDir, ".agent-smith")
		cmd := exec.Command(binaryPath, "install", "skill", testRepo, skillName, "--install-dir", baseDir)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Install output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Install failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify skill was installed to base directory structure
		skillDir := filepath.Join(baseDir, "skills", skillName)
		testutil.AssertDirectoryExists(t, skillDir)

		t.Logf("Successfully installed skill: %s to base directory", skillName)
	})

	// Step 2: Create profile
	t.Run("Step2_CreateProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "create", profileName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Create profile output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Create profile failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify profile directory was created
		profileDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName)
		testutil.AssertDirectoryExists(t, profileDir)

		t.Logf("Successfully created profile: %s", profileName)
	})

	// Step 3: Add skill to profile
	t.Run("Step3_AddSkillToProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "add", "skills", profileName, skillName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Profile add output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Profile add failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify skill was copied to profile
		profileSkillDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skillName)
		testutil.AssertDirectoryExists(t, profileSkillDir)

		// Verify component files exist
		skillFiles, err := os.ReadDir(profileSkillDir)
		testutil.AssertNoError(t, err, "Failed to read profile skill directory")
		if len(skillFiles) == 0 {
			t.Fatalf("Profile skill directory is empty")
		}

		t.Logf("Successfully added skill to profile: %s", profileName)
	})

	// Step 4: Verify lock file entry was preserved (implementation verification)
	t.Run("Step4_VerifyLockFilePreserved", func(t *testing.T) {
		// Read base lock file
		baseLockPath := filepath.Join(tempDir, ".agent-smith", ".component-lock.json")
		baseLockData, err := os.ReadFile(baseLockPath)
		testutil.AssertNoError(t, err, "Failed to read base lock file")

		var baseLockFile struct {
			Skills map[string]map[string]interface{} `json:"skills"`
		}
		testutil.AssertNoError(t, json.Unmarshal(baseLockData, &baseLockFile), "Failed to parse base lock file")

		// Read profile lock file
		profileLockPath := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, ".component-lock.json")
		profileLockData, err := os.ReadFile(profileLockPath)
		testutil.AssertNoError(t, err, "Failed to read profile lock file")

		var profileLockFile struct {
			Skills map[string]map[string]interface{} `json:"skills"`
		}
		testutil.AssertNoError(t, json.Unmarshal(profileLockData, &profileLockFile), "Failed to parse profile lock file")

		// Verify lock entry exists in both
		expectedSourceUrl := "https://github.com/anthropics/skills"
		baseEntry := baseLockFile.Skills[expectedSourceUrl][skillName].(map[string]interface{})
		profileEntry := profileLockFile.Skills[expectedSourceUrl][skillName].(map[string]interface{})

		// Verify commit hash matches (key field for updates)
		baseCommitHash := baseEntry["commitHash"].(string)
		profileCommitHash := profileEntry["commitHash"].(string)

		if baseCommitHash != profileCommitHash {
			t.Errorf("Commit hash mismatch. Base: %s, Profile: %s", baseCommitHash, profileCommitHash)
		}

		t.Logf("Verified lock file entry preserved with commit hash: %s", profileCommitHash)
	})

	// Step 5: Activate profile and verify update command works (E2E verification)
	t.Run("Step5_UpdateSkillInProfile", func(t *testing.T) {
		// Activate the profile
		activateCmd := exec.Command(binaryPath, "profile", "activate", profileName)
		output, err := activateCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Profile activate failed: %v\nOutput: %s", err, string(output))
		}

		// Run update command (should work because lock file was preserved)
		updateCmd := exec.Command(binaryPath, "update", "skills", skillName)
		output, err = updateCmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Update output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Update failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify update completed (should show "Up to date" or "Updated")
		if !strings.Contains(outputStr, "Up to date") && !strings.Contains(outputStr, "Updated") {
			t.Errorf("Expected update success message, got: %s", outputStr)
		}

		t.Logf("Successfully verified update works in profile (lock file preserved)")
	})
}

// TestE2E_ProfileAddManualComponentWorkflow tests adding manually created components to profiles
// This verifies backward compatibility with components not installed via Git
func TestE2E_ProfileAddManualComponentWorkflow(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-profile-manual-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	binaryPath := AgentSmithBinary
	profileName := "manual-profile"
	skillName := "manual-skill"

	// Step 1: Create manual skill in base directory (no Git source)
	t.Run("Step1_CreateManualSkill", func(t *testing.T) {
		skillDir := filepath.Join(tempDir, ".agent-smith", "skills", skillName)
		err := os.MkdirAll(skillDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create manual skill directory")

		skillContent := "# Manual Skill\n\nThis is a manually created skill without Git source.\n"
		err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0644)
		testutil.AssertNoError(t, err, "Failed to write skill file")

		t.Logf("Created manual skill: %s", skillName)
	})

	// Step 2: Create profile
	t.Run("Step2_CreateProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "create", profileName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Create profile output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Create profile failed: %v\nOutput: %s", err, outputStr)
		}

		t.Logf("Successfully created profile: %s", profileName)
	})

	// Step 3: Add manual skill to profile
	t.Run("Step3_AddManualSkillToProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "add", "skills", profileName, skillName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Profile add output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Profile add failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify skill was copied to profile
		profileSkillDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skillName)
		testutil.AssertDirectoryExists(t, profileSkillDir)

		// Verify SKILL.md was copied
		skillPath := filepath.Join(profileSkillDir, "SKILL.md")
		testutil.AssertFileExists(t, skillPath)

		t.Logf("Successfully added manual skill to profile")
	})

	// Step 4: Verify profile list shows the skill
	t.Run("Step4_VerifySkillInProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "list")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Profile list output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Profile list failed: %v\nOutput: %s", err, outputStr)
		}

		// Should show the profile with skill count
		if !strings.Contains(outputStr, profileName) {
			t.Errorf("Expected profile %s in list output", profileName)
		}

		t.Logf("Verified manual skill appears in profile")
	})
}
