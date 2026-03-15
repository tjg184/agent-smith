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

// TestE2E_LinkSkill_AutoLinkFromActiveProfile verifies that `link skill <name>` auto-links
// from the active profile when the component exists there.
func TestE2E_LinkSkill_AutoLinkFromActiveProfile(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-auto-link-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	binaryPath := AgentSmithBinary

	// Create OpenCode target dir so the target detector finds it
	opencodeDir := filepath.Join(tempDir, ".config", "opencode")
	if err := os.MkdirAll(opencodeDir, 0755); err != nil {
		t.Fatalf("failed to create opencode dir: %v", err)
	}

	// Create a profile with a skill
	profileName := "test-auto-link-profile"
	skillName := "auto-link-test-skill"
	skillDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills", skillName)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Auto Link Test"), 0644); err != nil {
		t.Fatalf("failed to write skill file: %v", err)
	}

	// Activate the profile
	t.Run("Step1_ActivateProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "activate", profileName)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("profile activate failed: %v\nOutput: %s", err, output)
		}
		t.Logf("Output:\n%s", output)
	})

	// Link skill — should auto-link from the active profile
	t.Run("Step2_LinkSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "skill", skillName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Output:\n%s", outputStr)
		if err != nil {
			t.Fatalf("link skill failed: %v\nOutput: %s", err, outputStr)
		}
	})

	// Verify symlink exists in OpenCode target
	t.Run("Step3_VerifySymlink", func(t *testing.T) {
		linkedPath := filepath.Join(opencodeDir, "skills", skillName)
		info, err := os.Lstat(linkedPath)
		if err != nil {
			t.Fatalf("expected symlink at %s, got error: %v", linkedPath, err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("expected a symlink at %s, got mode %v", linkedPath, info.Mode())
		}

		target, err := os.Readlink(linkedPath)
		if err != nil {
			t.Fatalf("failed to read symlink: %v", err)
		}

		// Resolve relative symlink to absolute for comparison
		absTarget := target
		if !filepath.IsAbs(target) {
			absTarget = filepath.Clean(filepath.Join(filepath.Dir(linkedPath), target))
		}

		expectedTarget, err := filepath.EvalSymlinks(skillDir)
		if err != nil {
			// skillDir may not be a symlink itself — use clean abs path
			expectedTarget = filepath.Clean(skillDir)
		}

		if absTarget != expectedTarget {
			t.Errorf("symlink target mismatch\nexpected: %s\ngot:      %s", expectedTarget, absTarget)
		}
	})
}

// TestE2E_LinkSkill_FallbackToOtherProfile verifies that `link skill <name>` falls back to
// a non-active profile when the component is not in the active profile.
func TestE2E_LinkSkill_FallbackToOtherProfile(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-fallback-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	binaryPath := AgentSmithBinary

	opencodeDir := filepath.Join(tempDir, ".config", "opencode")
	if err := os.MkdirAll(opencodeDir, 0755); err != nil {
		t.Fatalf("failed to create opencode dir: %v", err)
	}

	// Create an empty active profile
	activeProfileName := "test-fallback-active"
	activeProfileDir := filepath.Join(tempDir, ".agent-smith", "profiles", activeProfileName, "skills")
	if err := os.MkdirAll(activeProfileDir, 0755); err != nil {
		t.Fatalf("failed to create active profile: %v", err)
	}

	// Create another profile that has the skill
	otherProfileName := "test-fallback-other"
	skillName := "fallback-test-skill"
	skillDir := filepath.Join(tempDir, ".agent-smith", "profiles", otherProfileName, "skills", skillName)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("failed to create other profile skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Fallback Skill"), 0644); err != nil {
		t.Fatalf("failed to write skill file: %v", err)
	}

	// Activate the empty profile
	t.Run("Step1_ActivateEmptyProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "activate", activeProfileName)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("profile activate failed: %v\nOutput: %s", err, output)
		}
		t.Logf("Output:\n%s", output)
	})

	// Link skill — should fall back to the other profile
	t.Run("Step2_LinkSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "skill", skillName)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Output:\n%s", outputStr)
		if err != nil {
			t.Fatalf("link skill failed: %v\nOutput: %s", err, outputStr)
		}

		// Output should mention the other profile name or auto-selection
		if strings.Contains(outputStr, "does not exist") {
			t.Errorf("unexpected 'does not exist' in output; fallback should have succeeded:\n%s", outputStr)
		}
	})

	// Verify symlink exists in OpenCode target pointing to the other profile
	t.Run("Step3_VerifySymlinkFromOtherProfile", func(t *testing.T) {
		linkedPath := filepath.Join(opencodeDir, "skills", skillName)
		if _, err := os.Lstat(linkedPath); err != nil {
			t.Fatalf("expected symlink at %s, got error: %v", linkedPath, err)
		}
	})
}

// TestE2E_LinkSkill_ErrorWhenNotFound verifies that `link skill <name>` returns a clear error
// when the component does not exist in any profile.
func TestE2E_LinkSkill_ErrorWhenNotFound(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-notfound-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	binaryPath := AgentSmithBinary

	// Create a profile (empty) and activate it
	profileName := "test-notfound-profile"
	profileSkillsDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName, "skills")
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

	t.Run("Step2_LinkNonExistentSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "skill", "non-existent-skill-xyz-12345")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Output:\n%s", outputStr)

		if err == nil {
			t.Fatal("expected error when linking non-existent skill, got nil")
		}

		if !strings.Contains(outputStr, "does not exist") {
			t.Errorf("expected 'does not exist' in error output, got:\n%s", outputStr)
		}
	})
}
