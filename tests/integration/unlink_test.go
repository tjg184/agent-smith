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

// TestE2E_InstallLinkUnlinkWorkflow tests the full lifecycle: install → link → unlink → verify removal
func TestE2E_InstallLinkUnlinkWorkflow(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-unlink-*")
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

		t.Logf("Successfully installed skill: %s", skillName)
	})

	// Step 2: Link the skill to a target
	t.Run("Step2_LinkSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "skill", skillName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Link skill output:\n%s", outputStr)

		// Link may fail if no targets found (expected in test environment)
		if err != nil {
			if strings.Contains(outputStr, "No supported targets found") {
				t.Logf("Link correctly detected no targets (expected in test environment)")
				t.Skip("Skipping remaining steps - no targets available")
			} else {
				t.Fatalf("Link failed unexpectedly: %v\nOutput: %s", err, outputStr)
			}
		}

		t.Logf("Successfully linked skill: %s", skillName)
	})

	// Step 3: Unlink the skill
	t.Run("Step3_UnlinkSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "unlink", "skill", skillName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Unlink skill output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Unlink failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify output shows unlinking
		if !strings.Contains(outputStr, "Unlinking") {
			t.Errorf("Output should contain 'Unlinking', got: %s", outputStr)
		}

		t.Logf("Successfully unlinked skill: %s", skillName)
	})

	// Step 4: Verify source files still exist
	t.Run("Step4_VerifySourcePreserved", func(t *testing.T) {
		skillDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skillName)
		testutil.AssertDirectoryExists(t, skillDir)

		t.Logf("Verified source directory preserved: %s", skillDir)
	})
}

// TestE2E_InstallAllLinkAllUnlinkAllWorkflow tests: install all → link all → unlink all
func TestE2E_InstallAllLinkAllUnlinkAllWorkflow(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-unlink-all-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	binaryPath := AgentSmithBinary
	testRepo := "anthropics/skills"
	profileName := "anthropics-skills"

	// Step 1: Install all components
	t.Run("Step1_InstallAll", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "install", "all", testRepo)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Install all output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Install all failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify profile was created with components
		profilesDir := filepath.Join(tempDir, ".agent-smith", "profiles")
		testutil.AssertDirectoryExists(t, profilesDir)

		entries, err := os.ReadDir(profilesDir)
		testutil.AssertNoError(t, err, "Failed to read profiles directory")

		if len(entries) == 0 {
			t.Fatal("Expected at least one profile after install")
		}

		t.Logf("Successfully installed components to profile")
	})

	// Step 2: Link all components
	t.Run("Step2_LinkAll", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "all")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Link all output:\n%s", outputStr)

		// Link may fail if no targets found (expected in test environment)
		if err != nil {
			if strings.Contains(outputStr, "No supported targets found") {
				t.Logf("Link correctly detected no targets (expected)")
				t.Skip("Skipping remaining steps - no targets available")
			} else {
				t.Fatalf("Link all failed unexpectedly: %v\nOutput: %s", err, outputStr)
			}
		}

		t.Logf("Successfully linked all components")
	})

	// Step 3: Unlink all components
	t.Run("Step3_UnlinkAll", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "unlink", "all", "--force")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Unlink all output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Unlink all failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify output shows unlinking
		if !strings.Contains(outputStr, "Unlinking") {
			t.Errorf("Output should contain 'Unlinking', got: %s", outputStr)
		}

		t.Logf("Successfully unlinked all components")
	})

	// Step 4: Verify profile and components still exist
	t.Run("Step4_VerifyComponentsPreserved", func(t *testing.T) {
		profilesDir := filepath.Join(tempDir, ".agent-smith", "profiles")
		entries, err := os.ReadDir(profilesDir)
		testutil.AssertNoError(t, err, "Failed to read profiles directory")

		if len(entries) == 0 {
			t.Error("Profile should still exist after unlinking")
		}

		// Verify at least one skill directory still exists
		skillsDir := filepath.Join(profilesDir, profileName, "skills")
		skills, err := os.ReadDir(skillsDir)
		testutil.AssertNoError(t, err, "Failed to read skills directory")

		if len(skills) == 0 {
			t.Error("Skills should still exist after unlinking")
		}

		t.Logf("Verified all components preserved after unlinking")
	})
}

// TestE2E_UnlinkTypeWorkflow tests: install multiple → link → unlink by type (all skills, all agents, etc.)
func TestE2E_UnlinkTypeWorkflow(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-unlink-type-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	binaryPath := AgentSmithBinary
	testRepo := "anthropics/skills"
	profileName := "anthropics-skills"

	// Step 1: Install all components (creates profile with skills)
	t.Run("Step1_InstallAll", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "install", "all", testRepo)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Install all output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Install all failed: %v\nOutput: %s", err, outputStr)
		}

		t.Logf("Successfully installed all components")
	})

	// Step 2: Link all components
	t.Run("Step2_LinkAll", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "all")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Link all output:\n%s", outputStr)

		// Link may fail if no targets found (expected in test environment)
		if err != nil {
			if strings.Contains(outputStr, "No supported targets found") {
				t.Logf("Link correctly detected no targets (expected)")
				t.Skip("Skipping remaining steps - no targets available")
			} else {
				t.Fatalf("Link all failed unexpectedly: %v\nOutput: %s", err, outputStr)
			}
		}

		t.Logf("Successfully linked all components")
	})

	// Step 3: Unlink all skills (but not agents/commands)
	t.Run("Step3_UnlinkAllSkills", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "unlink", "skills", "--force")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Unlink all skills output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Unlink skills failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify output shows unlinking skills
		if !strings.Contains(outputStr, "skills") {
			t.Errorf("Output should mention skills, got: %s", outputStr)
		}

		t.Logf("Successfully unlinked all skills")
	})

	// Step 4: Verify skills still exist in profile
	t.Run("Step4_VerifySkillsPreserved", func(t *testing.T) {
		profilesDir := filepath.Join(tempDir, ".agent-smith", "profiles")
		entries, err := os.ReadDir(profilesDir)
		testutil.AssertNoError(t, err, "Failed to read profiles directory")

		if len(entries) == 0 {
			t.Fatal("Profile should exist")
		}

		skillsDir := filepath.Join(profilesDir, profileName, "skills")
		skills, err := os.ReadDir(skillsDir)
		testutil.AssertNoError(t, err, "Failed to read skills directory")

		if len(skills) == 0 {
			t.Error("Skills should still exist in profile after unlinking")
		}

		t.Logf("Verified skills preserved in profile after unlinking")
	})
}

// TestE2E_UnlinkWithTargetFilterWorkflow tests: install → link to specific target → unlink from that target
func TestE2E_UnlinkWithTargetFilterWorkflow(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-unlink-target-*")
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

	// Step 2: Link to specific target (opencode)
	t.Run("Step2_LinkToTarget", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "skill", skillName, "--to", "opencode")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Link to opencode output:\n%s", outputStr)

		// Link may fail if no targets found (expected in test environment)
		if err != nil {
			if strings.Contains(outputStr, "No supported targets found") ||
				strings.Contains(outputStr, "does not exist") {
				t.Logf("Link correctly detected no target (expected)")
				t.Skip("Skipping remaining steps - no targets available")
			} else {
				t.Fatalf("Link failed unexpectedly: %v\nOutput: %s", err, outputStr)
			}
		}

		t.Logf("Successfully linked to opencode")
	})

	// Step 3: Unlink from that specific target
	t.Run("Step3_UnlinkFromTarget", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "unlink", "skill", skillName, "--target", "opencode")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Unlink from opencode output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Unlink failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify output mentions the target
		if !strings.Contains(outputStr, "opencode") {
			t.Errorf("Output should mention opencode, got: %s", outputStr)
		}

		t.Logf("Successfully unlinked from opencode")
	})

	// Step 4: Verify source still exists
	t.Run("Step4_VerifySourcePreserved", func(t *testing.T) {
		skillDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skillName)
		testutil.AssertDirectoryExists(t, skillDir)

		t.Logf("Verified source directory preserved")
	})
}

// TestE2E_UnlinkNonExistentComponentWorkflow tests error handling when unlinking non-existent component
func TestE2E_UnlinkNonExistentComponentWorkflow(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-unlink-nonexistent-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	binaryPath := AgentSmithBinary

	// Step 1: Try to unlink a component that was never installed
	t.Run("Step1_UnlinkNonExistent", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "unlink", "skill", "non-existent-skill")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Unlink non-existent output:\n%s", outputStr)

		// Should fail with appropriate error
		if err == nil {
			t.Error("Expected error when unlinking non-existent component")
		}

		// Error message should indicate component is not linked
		if !strings.Contains(outputStr, "not linked") {
			t.Errorf("Error message should indicate component is not linked, got: %s", outputStr)
		}

		t.Logf("Correctly handled non-existent component")
	})
}

// TestE2E_UnlinkAfterUpdateWorkflow tests: install → link → update → unlink (verify unlink works after update)
func TestE2E_UnlinkAfterUpdateWorkflow(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-unlink-after-update-*")
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
				t.Skip("Skipping remaining steps - no targets available")
			} else {
				t.Fatalf("Link failed unexpectedly: %v\nOutput: %s", err, outputStr)
			}
		}

		t.Logf("Successfully linked skill")
	})

	// Step 3: Update the skill
	t.Run("Step3_UpdateSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "update", "skills", skillName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Update skill output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Update skill failed: %v\nOutput: %s", err, outputStr)
		}

		t.Logf("Successfully updated skill")
	})

	// Step 4: Unlink the skill (should work even after update)
	t.Run("Step4_UnlinkAfterUpdate", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "unlink", "skill", skillName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Unlink skill output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Unlink failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify output shows unlinking
		if !strings.Contains(outputStr, "Unlinking") {
			t.Errorf("Output should contain 'Unlinking', got: %s", outputStr)
		}

		t.Logf("Successfully unlinked skill after update")
	})

	// Step 5: Verify source still exists
	t.Run("Step5_VerifySourcePreserved", func(t *testing.T) {
		skillDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skillName)
		testutil.AssertDirectoryExists(t, skillDir)

		t.Logf("Verified source directory preserved after unlink")
	})
}
