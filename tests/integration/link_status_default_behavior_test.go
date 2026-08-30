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

// TestE2E_LinkStatusWorkflow verifies the link status command shows correct information across targets
// Tests the full workflow: install → link → check status → verify output format
func TestE2E_LinkStatusWorkflow(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-link-status-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	binaryPath := AgentSmithBinary
	testRepo := "anthropics/skills"
	skillName := "web-artifacts-builder"

	// Step 1: Install skill
	t.Run("Step1_InstallSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "install", "skill", testRepo, skillName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Install output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Install failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify skill was installed to auto-derived profile
		expectedProfileName := "anthropics-skills"
		profileDir := filepath.Join(tempDir, ".agent-smith", "profiles", expectedProfileName)
		skillDir := filepath.Join(profileDir, "skills", skillName)
		testutil.AssertDirectoryExists(t, skillDir)

		t.Logf("Successfully installed skill: %s", skillName)
	})

	// Step 2: Check link status before linking (should show not linked)
	t.Run("Step2_LinkStatusBeforeLinking", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "status")
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Link status output:\n%s", outputStr)

		// Command may succeed or fail depending on whether targets are detected
		// Focus on verifying the output format is correct

		// Verify output contains expected format elements
		expectedStrings := []string{
			"Component",
			"Profile / Repo",
			"--- Legend ---",
			"Symbol",
			"Meaning",
		}

		for _, expected := range expectedStrings {
			if !strings.Contains(outputStr, expected) {
				t.Errorf("Link status output missing expected string: %s\nFull output:\n%s", expected, outputStr)
			}
		}

		// Verify legend symbols are present
		legendSymbols := []string{"✓", "◆", "✗", "-", "?"}
		for _, symbol := range legendSymbols {
			if !strings.Contains(outputStr, symbol) {
				t.Errorf("Legend missing symbol: %s", symbol)
			}
		}

		t.Logf("Verified link status output format")
	})

	// Step 3: Link skill
	t.Run("Step3_LinkSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "skill", skillName)
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Link output:\n%s", outputStr)

		// Link may succeed or warn about no targets, both are acceptable
		t.Logf("Link command completed")
	})

	// Step 4: Check link status after linking
	t.Run("Step4_LinkStatusAfterLinking", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "status")
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Link status output after linking:\n%s", outputStr)

		// Verify the skill appears in the output
		if !strings.Contains(outputStr, skillName) && !strings.Contains(outputStr, "No components found") {
			t.Errorf("Expected skill %s in link status output", skillName)
		}

		t.Logf("Verified link status shows component status")
	})
}

// TestE2E_LinkStatusProfileWorkflow verifies link status correctly shows profile information
// Tests the workflow: create profile → install to profile → activate → link → check status
func TestE2E_LinkStatusProfileWorkflow(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-link-status-profile-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	binaryPath := AgentSmithBinary
	testRepo := "anthropics/skills"
	skillName := "brand-guidelines"
	profileName := "work"

	// Step 1: Create profile
	t.Run("Step1_CreateProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "create", profileName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Create profile output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Create profile failed: %v\nOutput: %s", err, outputStr)
		}

		t.Logf("Successfully created profile: %s", profileName)
	})

	// Step 2: Install skill to profile
	t.Run("Step2_InstallToProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "install", "skill", testRepo, skillName, "--profile", profileName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Install to profile output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Install to profile failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify skill was installed to profile
		skillDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skillName)
		testutil.AssertDirectoryExists(t, skillDir)

		t.Logf("Successfully installed skill to profile")
	})

	// Step 3: Activate profile
	t.Run("Step3_ActivateProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "activate", profileName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Activate profile output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Activate profile failed: %v\nOutput: %s", err, outputStr)
		}

		t.Logf("Successfully activated profile: %s", profileName)
	})

	// Step 4: Link skill from profile
	t.Run("Step4_LinkFromProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "skill", skillName)
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Link output:\n%s", outputStr)

		// Link may succeed or warn about no targets
		t.Logf("Link command completed")
	})

	// Step 5: Check link status shows profile information
	t.Run("Step5_LinkStatusShowsProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "status")
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Link status output:\n%s", outputStr)

		// Verify output shows profile information
		if !strings.Contains(outputStr, "No components found") {
			// If components are found, verify profile name appears
			if !strings.Contains(outputStr, profileName) {
				t.Errorf("Expected profile name %s in link status output", profileName)
			}

			// Verify skill name appears
			if !strings.Contains(outputStr, skillName) {
				t.Errorf("Expected skill %s in link status output", skillName)
			}
		}

		t.Logf("Verified link status shows profile information")
	})
}

// TestE2E_LinkStatusAllProfilesWorkflow verifies each link status scope.
func TestE2E_LinkStatusAllProfilesWorkflow(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-link-status-all-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	binaryPath := AgentSmithBinary
	testRepo := "anthropics/skills"

	// Step 1: Create two profiles with different skills
	t.Run("Step1_CreateProfilesWithSkills", func(t *testing.T) {
		// Profile 1
		createCmd := exec.Command(binaryPath, "profile", "create", "profile1")
		output, err := createCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Create profile1 failed: %v\nOutput: %s", err, string(output))
		}

		installCmd := exec.Command(binaryPath, "install", "skill", testRepo, "docx", "--profile", "profile1")
		output, err = installCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Install to profile1 failed: %v\nOutput: %s", err, string(output))
		}

		// Profile 2
		createCmd = exec.Command(binaryPath, "profile", "create", "profile2")
		output, err = createCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Create profile2 failed: %v\nOutput: %s", err, string(output))
		}

		installCmd = exec.Command(binaryPath, "install", "skill", testRepo, "pdf", "--profile", "profile2")
		output, err = installCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Install to profile2 failed: %v\nOutput: %s", err, string(output))
		}

		t.Logf("Created two profiles with different skills")
	})

	// Step 2: Activate profile1 so it becomes the default status scope.
	t.Run("Step2_ActivateProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "activate", "profile1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Activate profile1 failed: %v\nOutput: %s", err, string(output))
		}
	})

	// Step 3: Check the default status shows only the active profile.
	t.Run("Step3_LinkStatusActiveProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "status")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Link status for active profile failed: %v\nOutput: %s", err, string(output))
		}
		outputStr := string(output)

		t.Logf("Link status output:\n%s", outputStr)

		if !strings.Contains(outputStr, "No components found") {
			if !strings.Contains(outputStr, "docx") {
				t.Errorf("Expected active profile component docx in link status output")
			}
			if strings.Contains(outputStr, "pdf") {
				t.Errorf("Did not expect inactive profile component pdf in link status output")
			}
		}
	})

	// Step 4: An explicit profile overrides the active profile.
	t.Run("Step4_LinkStatusExplicitProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "status", "--profile", "profile2")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Link status for explicit profile failed: %v\nOutput: %s", err, string(output))
		}
		outputStr := string(output)

		if !strings.Contains(outputStr, "No components found") {
			if !strings.Contains(outputStr, "pdf") {
				t.Errorf("Expected explicit profile component pdf in link status output")
			}
			if strings.Contains(outputStr, "docx") {
				t.Errorf("Did not expect active profile component docx in explicit profile output")
			}
		}
	})

	// Step 5: All profiles is an explicit global overview.
	t.Run("Step5_LinkStatusAllProfiles", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "status", "--all-profiles")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Link status for all profiles failed: %v\nOutput: %s", err, string(output))
		}
		outputStr := string(output)

		if !strings.Contains(outputStr, "No components found") {
			for _, expected := range []string{"profile1", "profile2", "docx", "pdf"} {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected %s in all-profiles link status output", expected)
				}
			}
		}
	})

	// Step 6: Conflicting explicit scopes fail.
	t.Run("Step6_LinkStatusConflictingScopes", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "status", "--profile", "profile1", "--all-profiles")
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("Expected conflicting scopes to fail\nOutput: %s", string(output))
		}
		if !strings.Contains(string(output), "cannot use both --all-profiles and --profile flags together") {
			t.Errorf("Expected conflicting scope error, got: %s", string(output))
		}
	})

	// Step 7: No active profile requires an explicit scope.
	t.Run("Step7_LinkStatusWithoutActiveProfile", func(t *testing.T) {
		deactivateCmd := exec.Command(binaryPath, "profile", "deactivate")
		output, err := deactivateCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Deactivate profile failed: %v\nOutput: %s", err, string(output))
		}

		cmd := exec.Command(binaryPath, "link", "status")
		output, err = cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("Expected status without active profile to fail\nOutput: %s", string(output))
		}
		if !strings.Contains(string(output), "no active profile; activate one or use --profile or --all-profiles") {
			t.Errorf("Expected missing active profile error, got: %s", string(output))
		}
	})
}
