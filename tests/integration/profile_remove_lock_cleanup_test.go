//go:build integration
// +build integration

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/tjg184/agent-smith/internal/testutil"
)

// TestE2E_ProfileRemoveLockCleanup tests that removing a component from a profile also removes its lock file entry
func TestE2E_ProfileRemoveLockCleanup(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-profile-remove-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	binaryPath := AgentSmithBinary
	testRepo := "anthropics/skills"
	skillName := "frontend-design"
	profileName := "test-remove-profile"

	// Step 1: Install skill to base directory
	t.Run("Step1_InstallSkillToBase", func(t *testing.T) {
		baseDir := filepath.Join(tempDir, ".agent-smith")
		cmd := exec.Command(binaryPath, "install", "skill", testRepo, skillName, "--install-dir", baseDir)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Install output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Install failed: %v\nOutput: %s", err, outputStr)
		}

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

		profileSkillDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skillName)
		testutil.AssertDirectoryExists(t, profileSkillDir)

		t.Logf("Successfully added skill to profile: %s", profileName)
	})

	// Step 4: Verify lock file entry exists before removal
	t.Run("Step4_VerifyLockFileExists", func(t *testing.T) {
		profileLockPath := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, ".component-lock.json")
		profileLockData, err := os.ReadFile(profileLockPath)
		testutil.AssertNoError(t, err, "Failed to read profile lock file")

		var profileLockFile struct {
			Skills map[string]map[string]interface{} `json:"skills"`
		}
		testutil.AssertNoError(t, json.Unmarshal(profileLockData, &profileLockFile), "Failed to parse profile lock file")

		expectedSourceUrl := "https://github.com/anthropics/skills"
		if _, exists := profileLockFile.Skills[expectedSourceUrl]; !exists {
			t.Fatalf("Expected source URL %s not found in lock file", expectedSourceUrl)
		}

		if _, exists := profileLockFile.Skills[expectedSourceUrl][skillName]; !exists {
			t.Fatalf("Expected skill %s not found in lock file", skillName)
		}

		t.Logf("Verified lock file entry exists for skill: %s", skillName)
	})

	// Step 5: Remove skill from profile
	t.Run("Step5_RemoveSkillFromProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "remove", "skills", profileName, skillName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Profile remove output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Profile remove failed: %v\nOutput: %s", err, outputStr)
		}

		profileSkillDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skillName)
		if _, err := os.Stat(profileSkillDir); !os.IsNotExist(err) {
			t.Fatalf("Expected skill directory to be removed, but it still exists")
		}

		t.Logf("Successfully removed skill from profile: %s", profileName)
	})

	// Step 6: Verify lock file entry was removed
	t.Run("Step6_VerifyLockFileEntryRemoved", func(t *testing.T) {
		profileLockPath := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, ".component-lock.json")
		profileLockData, err := os.ReadFile(profileLockPath)
		testutil.AssertNoError(t, err, "Failed to read profile lock file")

		var profileLockFile struct {
			Skills map[string]map[string]interface{} `json:"skills"`
		}
		testutil.AssertNoError(t, json.Unmarshal(profileLockData, &profileLockFile), "Failed to parse profile lock file")

		expectedSourceUrl := "https://github.com/anthropics/skills"

		// Check if the skill entry was removed
		if sourceMap, exists := profileLockFile.Skills[expectedSourceUrl]; exists {
			if _, skillExists := sourceMap[skillName]; skillExists {
				t.Fatalf("Expected skill %s to be removed from lock file, but it still exists", skillName)
			}
		}

		t.Logf("Verified lock file entry was removed for skill: %s", skillName)
	})
}

// TestE2E_ProfileRemoveMultipleComponents tests that removing multiple components cleans up lock entries correctly
func TestE2E_ProfileRemoveMultipleComponents(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-profile-remove-multi-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	binaryPath := AgentSmithBinary
	testRepo := "anthropics/skills"
	skill1 := "frontend-design"
	skill2 := "skill-creator"
	profileName := "test-multi-remove"

	// Step 1: Install skills
	t.Run("Step1_InstallSkills", func(t *testing.T) {
		baseDir := filepath.Join(tempDir, ".agent-smith")

		for _, skillName := range []string{skill1, skill2} {
			cmd := exec.Command(binaryPath, "install", "skill", testRepo, skillName, "--install-dir", baseDir)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Fatalf("Install failed for %s: %v\nOutput: %s", skillName, err, string(output))
			}

			t.Logf("Installed skill: %s", skillName)
		}
	})

	// Step 2: Create profile and add skills
	t.Run("Step2_CreateProfileAndAddSkills", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "create", profileName)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Fatalf("Create profile failed: %v\nOutput: %s", err, string(output))
		}

		for _, skillName := range []string{skill1, skill2} {
			cmd := exec.Command(binaryPath, "profile", "add", "skills", profileName, skillName)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Fatalf("Profile add failed for %s: %v\nOutput: %s", skillName, err, string(output))
			}

			t.Logf("Added skill to profile: %s", skillName)
		}
	})

	// Step 3: Remove first skill
	t.Run("Step3_RemoveFirstSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "remove", "skills", profileName, skill1)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Fatalf("Profile remove failed: %v\nOutput: %s", err, string(output))
		}

		t.Logf("Removed skill: %s", skill1)
	})

	// Step 4: Verify first skill removed, second skill remains
	t.Run("Step4_VerifyPartialRemoval", func(t *testing.T) {
		profileLockPath := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, ".component-lock.json")
		profileLockData, err := os.ReadFile(profileLockPath)
		testutil.AssertNoError(t, err, "Failed to read profile lock file")

		var profileLockFile struct {
			Skills map[string]map[string]interface{} `json:"skills"`
		}
		testutil.AssertNoError(t, json.Unmarshal(profileLockData, &profileLockFile), "Failed to parse profile lock file")

		expectedSourceUrl := "https://github.com/anthropics/skills"
		sourceMap := profileLockFile.Skills[expectedSourceUrl]

		// First skill should be removed
		if _, exists := sourceMap[skill1]; exists {
			t.Fatalf("Expected skill %s to be removed from lock file", skill1)
		}

		// Second skill should still exist
		if _, exists := sourceMap[skill2]; !exists {
			t.Fatalf("Expected skill %s to remain in lock file", skill2)
		}

		t.Logf("Verified partial removal: %s removed, %s remains", skill1, skill2)
	})

	// Step 5: Remove second skill
	t.Run("Step5_RemoveSecondSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "remove", "skills", profileName, skill2)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Fatalf("Profile remove failed: %v\nOutput: %s", err, string(output))
		}

		t.Logf("Removed skill: %s", skill2)
	})

	// Step 6: Verify source URL removed from lock file when all components removed
	t.Run("Step6_VerifySourceUrlCleanup", func(t *testing.T) {
		profileLockPath := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, ".component-lock.json")
		profileLockData, err := os.ReadFile(profileLockPath)
		testutil.AssertNoError(t, err, "Failed to read profile lock file")

		var profileLockFile struct {
			Skills map[string]map[string]interface{} `json:"skills"`
		}
		testutil.AssertNoError(t, json.Unmarshal(profileLockData, &profileLockFile), "Failed to parse profile lock file")

		expectedSourceUrl := "https://github.com/anthropics/skills"

		// Source URL should be removed when all components from that source are removed
		if _, exists := profileLockFile.Skills[expectedSourceUrl]; exists {
			t.Fatalf("Expected source URL %s to be removed from lock file when all components removed", expectedSourceUrl)
		}

		t.Logf("Verified source URL cleanup after all components removed")
	})
}
