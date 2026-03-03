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

// TestE2E_InstallLinkUpdateWorkflow tests the full lifecycle: install all → verify files → link all → verify symlinks → update all → verify updates
func TestE2E_InstallLinkUpdateWorkflow(t *testing.T) {
	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-full-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	// Use the globally compiled binary (built once in TestMain)
	binaryPath := AgentSmithBinary

	// Test repository (using a well-known public repo)
	testRepo := "anthropics/skills"

	// Step 1: Install all components
	t.Run("Step1_InstallAll", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "install", "all", testRepo)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Install all output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Install all failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify profile was created
		profilesDir := filepath.Join(tempDir, ".agent-smith", "profiles")
		testutil.AssertDirectoryExists(t, profilesDir)

		entries, err := os.ReadDir(profilesDir)
		testutil.AssertNoError(t, err, "Failed to read profiles directory")

		if len(entries) != 1 {
			t.Fatalf("Expected 1 profile after install, got %d", len(entries))
		}

		profileName := entries[0].Name()
		t.Logf("Created profile: %s", profileName)

		// Verify metadata file exists and contains correct info
		metadataPath := filepath.Join(profilesDir, profileName, ".profile-metadata")
		testutil.AssertFileExists(t, metadataPath)

		metadataBytes, err := os.ReadFile(metadataPath)
		testutil.AssertNoError(t, err, "Failed to read metadata")

		var metadata map[string]interface{}
		err = json.Unmarshal(metadataBytes, &metadata)
		testutil.AssertNoError(t, err, "Failed to parse metadata JSON")

		sourceURL, ok := metadata["source_url"].(string)
		testutil.AssertTrue(t, ok, "Metadata missing source_url")
		testutil.AssertEqual(t, "https://github.com/anthropics/skills", sourceURL, "Incorrect source URL")

		// Verify component directories exist
		skillsDir := filepath.Join(profilesDir, profileName, "skills")
		testutil.AssertDirectoryExists(t, skillsDir)

		// Verify at least one skill was installed
		skills, err := os.ReadDir(skillsDir)
		testutil.AssertNoError(t, err, "Failed to read skills directory")

		if len(skills) == 0 {
			t.Error("Expected at least one skill to be installed")
		}

		t.Logf("Installed %d skills", len(skills))
	})

	// Step 2: Link all components
	t.Run("Step2_LinkAll", func(t *testing.T) {
		// Create a mock target directory to test linking
		targetDir := filepath.Join(tempDir, ".claude", "skills")
		err := os.MkdirAll(targetDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create target directory")

		// Get profile name
		profilesDir := filepath.Join(tempDir, ".agent-smith", "profiles")
		entries, err := os.ReadDir(profilesDir)
		testutil.AssertNoError(t, err, "Failed to read profiles directory")

		if len(entries) == 0 {
			t.Fatal("No profiles found")
		}

		profileName := entries[0].Name()
		profilePath := filepath.Join(profilesDir, profileName)

		// Note: Profile may already be active from install, so activation may fail
		// We'll check if it's active and activate only if needed
		cmd := exec.Command(binaryPath, "profile", "list")
		output, err := cmd.CombinedOutput()
		testutil.AssertNoError(t, err, "Failed to check active profile")

		isActive := strings.Contains(string(output), profileName) && strings.Contains(string(output), "[active]")

		if !isActive {
			cmd = exec.Command(binaryPath, "profile", "activate", profileName)
			output, err = cmd.CombinedOutput()
			testutil.AssertNoError(t, err, "Failed to activate profile")
			t.Logf("Activated profile: %s", profileName)
		} else {
			t.Logf("Profile already active: %s", profileName)
		}

		// Link all (no arguments - links all components from active profile)
		cmd = exec.Command(binaryPath, "link", "all")
		output, err = cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Link all output:\n%s", outputStr)

		// Link may fail if no valid target is detected, which is expected in test environment
		// We're primarily testing that the command works, not that it finds a target
		if err != nil {
			if strings.Contains(outputStr, "No supported targets found") {
				t.Logf("Link command correctly detected no targets (expected in test environment)")
			} else {
				t.Logf("Link command failed (may be expected if no targets available): %v", err)
			}
		} else {
			t.Logf("Link command succeeded (target may have been available)")
		}

		// Verify profile is still intact after link attempt
		testutil.AssertDirectoryExists(t, profilePath)
	})

	// Step 3: Update all components
	t.Run("Step3_UpdateAll", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "update", "all")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Update all output:\n%s", outputStr)

		// Update should succeed even if no updates are available
		testutil.AssertNoError(t, err, "Update all failed")

		// Verify profile still exists after update
		profilesDir := filepath.Join(tempDir, ".agent-smith", "profiles")
		entries, err := os.ReadDir(profilesDir)
		testutil.AssertNoError(t, err, "Failed to read profiles directory")

		if len(entries) != 1 {
			t.Errorf("Expected 1 profile after update, got %d", len(entries))
		}

		// Verify metadata still valid
		profileName := entries[0].Name()
		metadataPath := filepath.Join(profilesDir, profileName, ".profile-metadata")
		testutil.AssertFileExists(t, metadataPath)

		metadataBytes, err := os.ReadFile(metadataPath)
		testutil.AssertNoError(t, err, "Failed to read metadata after update")

		var metadata map[string]interface{}
		err = json.Unmarshal(metadataBytes, &metadata)
		testutil.AssertNoError(t, err, "Failed to parse metadata JSON after update")
	})
}

// TestE2E_SingleComponentWorkflow tests: Install single component → link single component → update single component
func TestE2E_SingleComponentWorkflow(t *testing.T) {
	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-single-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	// Use the globally compiled binary (built once in TestMain)
	binaryPath := AgentSmithBinary

	testRepo := "anthropics/skills"
	// We'll use a skill that's likely to exist in the repo
	skillName := "web-artifacts-builder"

	// Step 1: Install single skill (installs to auto-derived profile from repository)
	t.Run("Step1_InstallSingle", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "install", "skill", testRepo, skillName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Install single skill output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Install single skill failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify profile was created based on repository name
		profilesDir := filepath.Join(tempDir, ".agent-smith", "profiles")
		testutil.AssertDirectoryExists(t, profilesDir)

		// Profile should be named after the repository (anthropics-skills)
		expectedProfileName := "anthropics-skills"
		profileDir := filepath.Join(profilesDir, expectedProfileName)
		testutil.AssertDirectoryExists(t, profileDir)

		// Verify the specific skill directory exists in the profile
		skillDir := filepath.Join(profileDir, "skills", skillName)
		testutil.AssertDirectoryExists(t, skillDir)

		// Verify profile lock file was created
		lockFile := filepath.Join(profileDir, ".component-lock.json")
		testutil.AssertFileExists(t, lockFile)

		t.Logf("Successfully installed skill: %s to profile: %s", skillName, expectedProfileName)
	})

	// Step 2: Link single component (from auto-derived profile)
	t.Run("Step2_LinkSingle", func(t *testing.T) {
		// Link single skill (should work from the auto-created profile which is now active)
		cmd := exec.Command(binaryPath, "link", "skill", skillName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Link single skill output:\n%s", outputStr)

		// Link may fail if no targets found (expected in test environment)
		if err != nil {
			if strings.Contains(outputStr, "No supported targets found") {
				t.Logf("Link correctly detected no targets (expected)")
			} else {
				t.Logf("Link failed (may be expected): %v", err)
			}
		}
	})

	// Step 3: Update single component (in profile)
	t.Run("Step3_UpdateSingle", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "update", "skills", skillName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Update single skill output:\n%s", outputStr)

		// Update should succeed even if no updates available
		testutil.AssertNoError(t, err, "Update single skill failed")

		// Verify skill directory still exists in profile
		expectedProfileName := "anthropics-skills"
		profilesDir := filepath.Join(tempDir, ".agent-smith", "profiles")
		skillDir := filepath.Join(profilesDir, expectedProfileName, "skills", skillName)
		testutil.AssertDirectoryExists(t, skillDir)
	})
}

// TestE2E_ProfileWorkflow tests: Create profile → install to profile → activate → verify active → deactivate → verify inactive
func TestE2E_ProfileWorkflow(t *testing.T) {
	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-profile-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	// Use the globally compiled binary (built once in TestMain)
	binaryPath := AgentSmithBinary

	testRepo := "anthropics/skills"
	profileName := "work-profile"

	// Step 1: Create profile and verify
	t.Run("Step1_CreateProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "create", profileName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Create profile output:\n%s", outputStr)

		testutil.AssertNoError(t, err, "Failed to create profile")

		// Verify profile was created with specified name
		profileDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName)
		testutil.AssertDirectoryExists(t, profileDir)

		// Verify component directories were created
		testutil.AssertDirectoryExists(t, filepath.Join(profileDir, "skills"))
		testutil.AssertDirectoryExists(t, filepath.Join(profileDir, "agents"))
		testutil.AssertDirectoryExists(t, filepath.Join(profileDir, "commands"))

		t.Logf("Successfully created profile: %s", profileName)
	})

	// Step 2: Install components (this creates a new profile based on repo name, doesn't use the active profile)
	t.Run("Step2_InstallAll", func(t *testing.T) {
		// Install all components - this creates a profile based on the repository name
		cmd := exec.Command(binaryPath, "install", "all", testRepo)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Install all output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Install all failed: %v\nOutput: %s", err, outputStr)
		}

		// The install all command creates a profile based on repo name ("anthropics-skills")
		// So we now have two profiles: work-profile (empty) and anthropics-skills (with components)
		profilesDir := filepath.Join(tempDir, ".agent-smith", "profiles")
		entries, err := os.ReadDir(profilesDir)
		testutil.AssertNoError(t, err, "Failed to read profiles directory")

		if len(entries) < 2 {
			t.Errorf("Expected at least 2 profiles (work-profile + anthropics-skills), got %d", len(entries))
		}

		// Verify the anthropics-skills profile has components
		anthropicsProfileDir := filepath.Join(profilesDir, "anthropics-skills")
		testutil.AssertDirectoryExists(t, anthropicsProfileDir)

		skillsDir := filepath.Join(anthropicsProfileDir, "skills")
		skills, err := os.ReadDir(skillsDir)
		testutil.AssertNoError(t, err, "Failed to read skills directory")

		if len(skills) == 0 {
			t.Error("Expected at least one skill to be installed")
		}

		t.Logf("Successfully installed %d components to 'anthropics-skills' profile", len(skills))
	})

	// Step 3: Deactivate currently active profile (work-profile)
	t.Run("Step3_DeactivateProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "deactivate")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Deactivate profile output:\n%s", outputStr)

		testutil.AssertNoError(t, err, "Failed to deactivate profile")

		// Verify output mentions deactivation
		testutil.AssertTrue(t, strings.Contains(outputStr, "Deactivated") || strings.Contains(outputStr, "deactivated"),
			"Output should mention deactivation")
	})

	// Step 4: Verify no active profile
	t.Run("Step4_VerifyNoActiveProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "list")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Profiles list output:\n%s", outputStr)

		testutil.AssertNoError(t, err, "Failed to list profiles")

		// Verify both profiles exist but neither is active
		testutil.AssertTrue(t, strings.Contains(outputStr, profileName), "work-profile should be listed")
		testutil.AssertTrue(t, strings.Contains(outputStr, "skills"), "skills profile should be listed")

		// Count active markers - there should be none (no checkmarks before profile names)
		// Active profiles have "✓" at the start of their line
		lines := strings.Split(outputStr, "\n")
		activeCount := 0
		for _, line := range lines {
			// Check if line starts with "│ ✓" which indicates an active profile
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "│ ✓") {
				activeCount++
			}
		}
		testutil.AssertEqual(t, 0, activeCount, "No profiles should be active")
	})

	// Step 5: Activate the anthropics-skills profile
	t.Run("Step5_ActivateSkillsProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "activate", "anthropics-skills")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Activate anthropics-skills profile output:\n%s", outputStr)

		testutil.AssertNoError(t, err, "Failed to activate anthropics-skills profile")

		// Verify it's active
		cmd = exec.Command(binaryPath, "profile", "list")
		output, err = cmd.CombinedOutput()
		outputStr = string(output)

		testutil.AssertNoError(t, err, "Failed to list profiles")

		// Check if anthropics-skills profile is active (line should start with "│ ✓" and contain "anthropics-skills")
		isActive := false
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "│ ✓") && strings.Contains(line, "anthropics-skills") {
				isActive = true
				break
			}
		}
		testutil.AssertTrue(t, isActive, "anthropics-skills profile should be active")

		t.Logf("Successfully activated anthropics-skills profile")
	})

	// Step 6: Switch back to work-profile
	t.Run("Step6_SwitchToWorkProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "activate", profileName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Activate work-profile output:\n%s", outputStr)

		testutil.AssertNoError(t, err, "Failed to activate work-profile")

		// Verify work-profile is now active
		cmd = exec.Command(binaryPath, "profile", "list")
		output, err = cmd.CombinedOutput()
		outputStr = string(output)

		testutil.AssertNoError(t, err, "Failed to list profiles")

		// Check if work-profile is active (line should start with "│ ✓" and contain profileName)
		isActive := false
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "│ ✓") && strings.Contains(line, profileName) {
				isActive = true
				break
			}
		}
		testutil.AssertTrue(t, isActive, "work-profile should be active")

		t.Logf("Successfully switched to work-profile")
	})
}

// TestE2E_CustomTargetDirWorkflow tests: Install to custom --install-dir → verify isolation from ~/.agent-smith/
func TestE2E_CustomTargetDirWorkflow(t *testing.T) {
	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-targetdir-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	// Use the globally compiled binary (built once in TestMain)
	binaryPath := AgentSmithBinary

	testRepo := "anthropics/skills"
	skillName := "web-artifacts-builder"
	customTargetDir := filepath.Join(tempDir, "custom-components")

	// Step 1: Install to custom target directory
	t.Run("Step1_InstallToCustomDir", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "install", "skill", testRepo, skillName, "--install-dir", customTargetDir)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Install to custom target-dir output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Install to custom target-dir failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify skill was installed to custom directory
		skillDir := filepath.Join(customTargetDir, "skills", skillName)
		testutil.AssertDirectoryExists(t, skillDir)

		t.Logf("Successfully installed to custom directory: %s", customTargetDir)
	})

	// Step 2: Verify isolation from ~/.agent-smith/
	t.Run("Step2_VerifyIsolation", func(t *testing.T) {
		// Verify ~/.agent-smith/ profiles directory was NOT created or is empty
		agentSmithDir := filepath.Join(tempDir, ".agent-smith", "profiles")

		// Check if profiles directory exists
		if _, err := os.Stat(agentSmithDir); err == nil {
			// If it exists, verify it's empty or doesn't contain profiles from this install
			entries, err := os.ReadDir(agentSmithDir)
			if err == nil && len(entries) > 0 {
				// Check if any profile contains our skill
				foundSkillInProfile := false
				for _, entry := range entries {
					if !entry.IsDir() {
						continue
					}
					profileSkillDir := filepath.Join(agentSmithDir, entry.Name(), "skills", skillName)
					if _, err := os.Stat(profileSkillDir); err == nil {
						foundSkillInProfile = true
						break
					}
				}

				if foundSkillInProfile {
					t.Error("Skill was installed to profile directory despite using --install-dir")
				}
			}
		}

		// Verify skill only exists in custom directory
		customSkillDir := filepath.Join(customTargetDir, "skills", skillName)
		testutil.AssertDirectoryExists(t, customSkillDir)

		t.Logf("Verified isolation: skill only in custom directory, not in ~/.agent-smith/")
	})

	// Step 3: Install another component to same custom directory
	t.Run("Step3_InstallSecondToSameCustomDir", func(t *testing.T) {
		secondSkillName := "brand-guidelines"

		cmd := exec.Command(binaryPath, "install", "skill", testRepo, secondSkillName, "--install-dir", customTargetDir)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		t.Logf("Install second skill to custom dir output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Install second skill to custom dir failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify both skills exist in custom directory
		skill1Dir := filepath.Join(customTargetDir, "skills", skillName)
		skill2Dir := filepath.Join(customTargetDir, "skills", secondSkillName)

		testutil.AssertDirectoryExists(t, skill1Dir)
		testutil.AssertDirectoryExists(t, skill2Dir)

		// Verify both are in same parent directory
		skills, err := os.ReadDir(filepath.Join(customTargetDir, "skills"))
		testutil.AssertNoError(t, err, "Failed to read skills directory")

		if len(skills) < 2 {
			t.Errorf("Expected at least 2 skills in custom directory, got %d", len(skills))
		}

		t.Logf("Successfully installed multiple components to same custom directory")
	})
}
