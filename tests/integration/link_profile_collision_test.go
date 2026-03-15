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

// TestE2E_LinkSkill_CollisionActiveProfileWins verifies that when the same component exists in
// multiple profiles and one is active, the active profile is used without prompting.
func TestE2E_LinkSkill_CollisionActiveProfileWins(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-collision-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	binaryPath := AgentSmithBinary

	opencodeDir := filepath.Join(tempDir, ".config", "opencode")
	if err := os.MkdirAll(opencodeDir, 0755); err != nil {
		t.Fatalf("failed to create opencode dir: %v", err)
	}

	skillName := "collision-test-skill"

	// Create the same skill in two profiles with distinct content so we can verify which was linked
	for _, profile := range []string{"collision-profile-1", "collision-profile-2"} {
		skillDir := filepath.Join(tempDir, ".agent-smith", "profiles", profile, "skills", skillName)
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			t.Fatalf("failed to create skill dir for %s: %v", profile, err)
		}
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Skill from "+profile), 0644); err != nil {
			t.Fatalf("failed to write skill file for %s: %v", profile, err)
		}
	}

	// Activate profile-1
	t.Run("Step1_ActivateProfile1", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "activate", "collision-profile-1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("profile activate failed: %v\nOutput: %s", err, output)
		}
		t.Logf("Output:\n%s", output)
	})

	// Link skill — active profile (profile-1) should win without prompting
	t.Run("Step2_LinkSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "skill", skillName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Output:\n%s", outputStr)
		if err != nil {
			t.Fatalf("link skill failed: %v\nOutput: %s", err, outputStr)
		}
	})

	// Verify symlink exists and points to profile-1 (the active one)
	t.Run("Step3_VerifyLinkedToActiveProfile", func(t *testing.T) {
		linkedPath := filepath.Join(opencodeDir, "skills", skillName)
		if _, err := os.Lstat(linkedPath); err != nil {
			t.Fatalf("expected symlink at %s: %v", linkedPath, err)
		}

		linkTarget, err := os.Readlink(linkedPath)
		if err != nil {
			t.Fatalf("failed to read symlink: %v", err)
		}

		absTarget := linkTarget
		if !filepath.IsAbs(linkTarget) {
			absTarget = filepath.Clean(filepath.Join(filepath.Dir(linkedPath), linkTarget))
		}

		expectedProfile1SkillDir := filepath.Clean(filepath.Join(
			tempDir, ".agent-smith", "profiles", "collision-profile-1", "skills", skillName,
		))

		if absTarget != expectedProfile1SkillDir {
			t.Errorf("symlink should point to active profile (collision-profile-1)\nexpected: %s\ngot:      %s",
				expectedProfile1SkillDir, absTarget)
		}
	})
}

// TestE2E_LinkSkill_NotFoundInAnyProfile verifies that `link skill <name>` returns a clear error
// when the component does not exist in any profile.
func TestE2E_LinkSkill_NotFoundInAnyProfile(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-noprofile-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	binaryPath := AgentSmithBinary

	// Create a profile with a different skill (so profiles dir exists but target skill is absent)
	profileName := "test-noprofile-profile"
	otherSkillDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", "other-skill")
	if err := os.MkdirAll(otherSkillDir, 0755); err != nil {
		t.Fatalf("failed to create profile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(otherSkillDir, "SKILL.md"), []byte("# Other"), 0644); err != nil {
		t.Fatalf("failed to write skill file: %v", err)
	}

	t.Run("Step1_ActivateProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "activate", profileName)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("profile activate failed: %v\nOutput: %s", err, output)
		}
	})

	t.Run("Step2_LinkAbsentSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "skill", "definitely-does-not-exist-xyz")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Output:\n%s", outputStr)

		if err == nil {
			t.Fatal("expected error when linking absent skill")
		}

		if !strings.Contains(outputStr, "does not exist") {
			t.Errorf("expected 'does not exist' in output, got:\n%s", outputStr)
		}
	})
}

// TestE2E_LinkAgent_NotFoundInAnyProfile verifies the same not-found error for agents.
func TestE2E_LinkAgent_NotFoundInAnyProfile(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-noagent-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	binaryPath := AgentSmithBinary

	profileName := "test-no-agent-profile"
	profileSkillsDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "agents")
	if err := os.MkdirAll(profileSkillsDir, 0755); err != nil {
		t.Fatalf("failed to create profile: %v", err)
	}

	t.Run("Step1_ActivateProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "activate", profileName)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("profile activate failed: %v\nOutput: %s", err, output)
		}
	})

	t.Run("Step2_LinkAbsentAgent", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "agent", "this-agent-does-not-exist-xyz")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Output:\n%s", outputStr)

		if err == nil {
			t.Fatal("expected error when linking absent agent")
		}

		if !strings.Contains(outputStr, "does not exist") {
			t.Errorf("expected 'does not exist' in output, got:\n%s", outputStr)
		}
	})
}
