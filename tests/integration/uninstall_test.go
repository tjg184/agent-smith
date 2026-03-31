//go:build integration
// +build integration

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tjg184/agent-smith/internal/testutil"
)

// TestE2E_InstallUninstallWorkflow tests the full lifecycle: install → verify → uninstall → verify removed
func TestE2E_InstallUninstallWorkflow(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-uninstall-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	binaryPath := AgentSmithBinary
	testRepo := "anthropics/skills"
	skillName := "web-artifacts-builder"
	profileName := "anthropics-skills"

	// Step 1: Install a single skill
	t.Run("Step1_InstallSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "install", "skill", testRepo, skillName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Install skill output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Install skill failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify skill was installed to profile
		skillDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skillName)
		testutil.AssertDirectoryExists(t, skillDir)

		// Verify profile lock file was created
		lockFile := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, ".component-lock.json")
		testutil.AssertFileExists(t, lockFile)

		t.Logf("Successfully installed skill: %s", skillName)
	})

	// Step 2: Uninstall the skill
	t.Run("Step2_UninstallSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "uninstall", "skill", skillName, "--profile", profileName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Uninstall skill output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Uninstall failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify output shows uninstalling
		if !strings.Contains(outputStr, "Uninstalling") {
			t.Errorf("Output should contain 'Uninstalling', got: %s", outputStr)
		}
		if !strings.Contains(outputStr, "Removed") {
			t.Errorf("Output should contain 'Removed', got: %s", outputStr)
		}

		t.Logf("Successfully uninstalled skill: %s", skillName)
	})

	// Step 3: Verify component directory was removed
	t.Run("Step3_VerifyRemoved", func(t *testing.T) {
		skillDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skillName)
		testutil.AssertFileNotExists(t, skillDir)

		t.Logf("Verified skill directory was removed: %s", skillDir)
	})
}

// TestE2E_InstallLinkUninstallWorkflow tests: install → link → uninstall (should auto-unlink)
func TestE2E_InstallLinkUninstallWorkflow(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-uninstall-linked-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	binaryPath := AgentSmithBinary
	testRepo := "anthropics/skills"
	skillName := "web-artifacts-builder"
	profileName := "anthropics-skills"

	// Step 1: Install a skill
	t.Run("Step1_InstallSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "install", "skill", testRepo, skillName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Install skill output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Install skill failed: %v\nOutput: %s", err, outputStr)
		}

		t.Logf("Successfully installed skill: %s", skillName)
	})

	// Step 2: Link the skill
	t.Run("Step2_LinkSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "skill", skillName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Link skill output:\n%s", outputStr)

		// Link may fail if no targets found (expected in test environment)
		if err != nil {
			if strings.Contains(outputStr, "No supported targets found") {
				t.Logf("Link correctly detected no targets (expected)")
				t.Skip("Skipping remaining steps - no targets available for linking test")
			} else {
				t.Fatalf("Link failed unexpectedly: %v\nOutput: %s", err, outputStr)
			}
		}

		t.Logf("Successfully linked skill: %s", skillName)
	})

	// Step 3: Uninstall the skill (should automatically unlink)
	t.Run("Step3_UninstallLinkedSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "uninstall", "skill", skillName, "--profile", profileName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Uninstall linked skill output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Uninstall failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify output shows both unlinking and removal
		if !strings.Contains(outputStr, "Unlinking") && !strings.Contains(outputStr, "Removing") {
			t.Errorf("Output should show unlinking or removal steps, got: %s", outputStr)
		}

		t.Logf("Successfully uninstalled linked skill")
	})

	// Step 4: Verify component directory was removed
	t.Run("Step4_VerifyRemovedCompletely", func(t *testing.T) {
		skillDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skillName)
		testutil.AssertFileNotExists(t, skillDir)

		t.Logf("Verified skill completely removed")
	})
}

// TestE2E_UninstallAllFromRepoWorkflow tests: install all → uninstall all from repo
// Components installed via 'install skill' go into a profile directory. 'uninstall all'
// must find and remove them there.
func TestE2E_UninstallAllFromRepoWorkflow(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-uninstall-all-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	binaryPath := AgentSmithBinary
	testRepo := "anthropics/skills"
	skillName := "web-artifacts-builder"
	profileName := "anthropics-skills"

	// Step 1: Install a skill to profile (automatic profile creation)
	t.Run("Step1_InstallToProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "install", "skill", testRepo, skillName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Install skill output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Install failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify skill was installed to profile directory
		skillsDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skillName)
		testutil.AssertDirectoryExists(t, skillsDir)

		t.Logf("Successfully installed %s to profile directory", skillName)
	})

	// Step 2: Uninstall all components from the repo (should find them in the profile)
	t.Run("Step2_UninstallAll", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "uninstall", "all", testRepo, "--force")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Uninstall all output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Uninstall all failed: %v\nOutput: %s", err, outputStr)
		}

		// Should show the uninstall preview and removal, not "no components found"
		if strings.Contains(outputStr, "No components found") {
			t.Errorf("Uninstall all should have found components in profiles, got: %s", outputStr)
		}

		t.Logf("Successfully uninstalled all components from profiles")
	})

	// Step 3: Verify component was removed from profile
	t.Run("Step3_VerifyRemovedFromProfile", func(t *testing.T) {
		skillsDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skillName)
		testutil.AssertFileNotExists(t, skillsDir)

		t.Logf("Verified component was removed from profile")
	})

	// Step 4: Verify the profile directory itself was cleaned up (empty repo profile)
	t.Run("Step4_VerifyProfileDirRemoved", func(t *testing.T) {
		profileDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName)
		testutil.AssertFileNotExists(t, profileDir)

		t.Logf("Verified empty repo profile directory was removed")
	})
}

// TestE2E_UninstallNonExistentComponentWorkflow tests error handling when uninstalling non-existent component
func TestE2E_UninstallNonExistentComponentWorkflow(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-uninstall-nonexistent-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	binaryPath := AgentSmithBinary

	// Step 1: Try to uninstall a component that was never installed
	t.Run("Step1_UninstallNonExistent", func(t *testing.T) {
		// Create .agent-smith directory so it's not a "no installation" error
		agentSmithDir := filepath.Join(tempDir, ".agent-smith")
		if err := os.MkdirAll(agentSmithDir, 0755); err != nil {
			t.Fatalf("Failed to create .agent-smith directory: %v", err)
		}

		cmd := exec.Command(binaryPath, "uninstall", "skill", "non-existent-skill")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Uninstall non-existent output:\n%s", outputStr)

		// Should fail with appropriate error
		if err == nil {
			t.Error("Expected error when uninstalling non-existent component")
		}

		// Error message should indicate component is not installed
		if !strings.Contains(outputStr, "not installed") {
			t.Errorf("Error message should indicate component is not installed, got: %s", outputStr)
		}

		t.Logf("Correctly handled non-existent component")
	})
}

// TestE2E_InstallMultipleUninstallOneWorkflow tests: install multiple → uninstall one → verify others preserved
func TestE2E_InstallMultipleUninstallOneWorkflow(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-uninstall-selective-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	binaryPath := AgentSmithBinary
	testRepo := "anthropics/skills"
	skill1 := "web-artifacts-builder"
	skill2 := "brand-guidelines"
	profileName := "anthropics-skills"

	// Step 1: Install first skill
	t.Run("Step1_InstallFirstSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "install", "skill", testRepo, skill1)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Install first skill output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Install first skill failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify first skill was installed to profile
		skillDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skill1)
		testutil.AssertDirectoryExists(t, skillDir)

		t.Logf("Successfully installed first skill: %s", skill1)
	})

	// Step 2: Install second skill
	t.Run("Step2_InstallSecondSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "install", "skill", testRepo, skill2)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Install second skill output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Install second skill failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify second skill was installed to profile
		skillDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skill2)
		testutil.AssertDirectoryExists(t, skillDir)

		t.Logf("Successfully installed second skill: %s", skill2)
	})

	// Step 3: Verify both skills exist
	t.Run("Step3_VerifyBothExist", func(t *testing.T) {
		skill1Dir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skill1)
		skill2Dir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skill2)

		testutil.AssertDirectoryExists(t, skill1Dir)
		testutil.AssertDirectoryExists(t, skill2Dir)

		t.Logf("Verified both skills exist")
	})

	// Step 4: Uninstall only the first skill
	t.Run("Step4_UninstallFirstSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "uninstall", "skill", skill1, "--profile", profileName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Uninstall first skill output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Uninstall first skill failed: %v\nOutput: %s", err, outputStr)
		}

		t.Logf("Successfully uninstalled first skill: %s", skill1)
	})

	// Step 5: Verify first skill removed, second skill preserved
	t.Run("Step5_VerifySelectiveRemoval", func(t *testing.T) {
		skill1Dir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skill1)
		skill2Dir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skill2)

		// First skill should be removed
		testutil.AssertFileNotExists(t, skill1Dir)

		// Second skill should still exist
		testutil.AssertDirectoryExists(t, skill2Dir)

		t.Logf("Verified selective removal: %s removed, %s preserved", skill1, skill2)
	})
}

// TestE2E_UninstallAfterUpdateWorkflow tests: install → update → uninstall (verify uninstall works after update)
func TestE2E_UninstallAfterUpdateWorkflow(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-uninstall-after-update-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	binaryPath := AgentSmithBinary
	testRepo := "anthropics/skills"
	skillName := "web-artifacts-builder"
	profileName := "anthropics-skills"

	// Step 1: Install a skill
	t.Run("Step1_InstallSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "install", "skill", testRepo, skillName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Install skill output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Install skill failed: %v\nOutput: %s", err, outputStr)
		}

		t.Logf("Successfully installed skill: %s", skillName)
	})

	// Step 2: Update the skill
	t.Run("Step2_UpdateSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "update", "skills", skillName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Update skill output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Update skill failed: %v\nOutput: %s", err, outputStr)
		}

		t.Logf("Successfully updated skill")
	})

	// Step 3: Uninstall the skill (should work even after update)
	t.Run("Step3_UninstallAfterUpdate", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "uninstall", "skill", skillName, "--profile", profileName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Uninstall skill output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Uninstall failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify output shows uninstalling
		if !strings.Contains(outputStr, "Uninstalling") {
			t.Errorf("Output should contain 'Uninstalling', got: %s", outputStr)
		}

		t.Logf("Successfully uninstalled skill after update")
	})

	// Step 4: Verify component directory was removed
	t.Run("Step4_VerifyRemoved", func(t *testing.T) {
		skillDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skillName)
		testutil.AssertFileNotExists(t, skillDir)

		t.Logf("Verified skill directory removed after uninstall")
	})
}

// TestE2E_UninstallFromProfileWorkflow tests: create profile → install to profile → uninstall from profile
func TestE2E_UninstallFromProfileWorkflow(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-uninstall-profile-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	binaryPath := AgentSmithBinary
	testRepo := "anthropics/skills"
	skillName := "web-artifacts-builder"
	profileName := "work"

	// Step 1: Create profile and install to it
	t.Run("Step1_CreateProfileAndInstall", func(t *testing.T) {
		// Create profile
		createCmd := exec.Command(binaryPath, "profile", "create", profileName)
		output, err := createCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Profile create failed: %v\nOutput: %s", err, string(output))
		}

		// Install skill to profile
		cmd := exec.Command(binaryPath, "install", "skill", testRepo, skillName, "--profile", profileName)
		output, err = cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Install to profile output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Install to profile failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify skill was installed to profile
		skillDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skillName)
		testutil.AssertDirectoryExists(t, skillDir)

		t.Logf("Successfully installed to profile: %s", profileName)
	})

	// Step 2: Uninstall from profile
	t.Run("Step2_UninstallFromProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "uninstall", "skill", skillName, "--profile", profileName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Uninstall from profile output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Uninstall from profile failed: %v\nOutput: %s", err, outputStr)
		}

		t.Logf("Successfully uninstalled from profile")
	})

	// Step 3: Verify component removed from profile
	t.Run("Step3_VerifyRemovedFromProfile", func(t *testing.T) {
		skillDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skillName)
		testutil.AssertFileNotExists(t, skillDir)

		t.Logf("Verified skill removed from profile")
	})
}

// TestUninstallSkillWithSourceFlag verifies that --source disambiguates when the same
// component name is installed from multiple repositories. Only the specified source's
// entry is removed; the other remains intact.
func TestUninstallSkillWithSourceFlag(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-uninstall-source-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	binaryPath := AgentSmithBinary

	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	sourceA := "https://github.com/github/awesome-copilot"
	sourceB := "https://github.com/marcelorodrigo/agent-skills"

	for _, source := range []string{sourceA, sourceB} {
		skillDir := filepath.Join(agentSmithDir, "skills", "conventional-commit")
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			t.Fatalf("Failed to create skill directory: %v", err)
		}
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Conventional Commit"), 0644); err != nil {
			t.Fatalf("Failed to write SKILL.md: %v", err)
		}

		lockFilePath := filepath.Join(agentSmithDir, ".component-lock.json")
		testutil.AddComponentToLockFile(t, lockFilePath, "skills", "conventional-commit", source, map[string]interface{}{
			"sourceType":     "github",
			"sourceUrl":      source,
			"commitHash":     "abc123",
			"filesystemName": "conventional-commit",
		})
	}

	t.Run("FailsWithoutSourceWhenAmbiguous", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "uninstall", "skill", "conventional-commit")
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("Expected failure for ambiguous component, but command succeeded:\n%s", string(output))
		}
	})

	t.Run("SucceedsWithSourceFlag", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "uninstall", "skill", "conventional-commit", "--source", sourceB)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Output:\n%s", outputStr)
		if err != nil {
			t.Fatalf("Uninstall with --source failed: %v\nOutput: %s", err, outputStr)
		}
	})

	t.Run("OtherSourceEntryStillPresent", func(t *testing.T) {
		lockFilePath := filepath.Join(agentSmithDir, ".component-lock.json")
		data, err := os.ReadFile(lockFilePath)
		if err != nil {
			t.Fatalf("Failed to read lock file: %v", err)
		}
		if !strings.Contains(string(data), sourceA) {
			t.Errorf("Expected source %s to remain in lock file after targeted uninstall, but it was removed", sourceA)
		}
		if strings.Contains(string(data), sourceB) {
			t.Errorf("Expected source %s to be removed from lock file, but it is still present", sourceB)
		}
	})

	t.Run("SharedDirectoryPreserved", func(t *testing.T) {
		skillDir := filepath.Join(agentSmithDir, "skills", "conventional-commit")
		if _, err := os.Stat(skillDir); os.IsNotExist(err) {
			t.Errorf("Expected skill directory to be preserved (still referenced by %s), but it was deleted", sourceA)
		}
	})
}
