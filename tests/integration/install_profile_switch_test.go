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

// TestE2E_InstallSkillSwitchesProfile verifies that installing a single skill
// unconditionally switches the active profile, even when one is already active.
func TestE2E_InstallSkillSwitchesProfile(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-install-switch-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	binaryPath := AgentSmithBinary
	activeProfileFile := filepath.Join(tempDir, ".agent-smith", ".active-profile")

	// Step 1: Install first skill — creates and activates "anthropics-skills" profile
	t.Run("Step1_InstallFirstSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "install", "skill", "anthropics/skills", "web-artifacts-builder")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("install skill failed: %v\nOutput: %s", err, output)
		}
		t.Logf("Output:\n%s", output)

		data, err := os.ReadFile(activeProfileFile)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "anthropics-skills", strings.TrimSpace(string(data)))
	})

	// Step 2: Install a skill from a different repo — should switch active profile
	t.Run("Step2_InstallSkillFromDifferentRepo", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "install", "skill", "spillwavesolutions/confluence-skill", "confluence")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("install skill failed: %v\nOutput: %s", err, output)
		}
		outputStr := string(output)
		t.Logf("Output:\n%s", outputStr)

		// Active profile should have switched
		data, err := os.ReadFile(activeProfileFile)
		testutil.AssertNoError(t, err)
		activeProfile := strings.TrimSpace(string(data))
		if activeProfile == "anthropics-skills" {
			t.Errorf("expected active profile to switch away from anthropics-skills, but it did not")
		}

		// Output should mention the switch
		if !strings.Contains(outputStr, "Switched profile") && !strings.Contains(outputStr, "Profile activated") {
			t.Errorf("expected output to mention profile switch/activation, got:\n%s", outputStr)
		}
	})
}
