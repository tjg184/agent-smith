//go:build integration
// +build integration

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/tjg184/agent-smith/internal/testutil"
)

// createProfileWithSkill creates a profile directory containing a single test skill.
func createProfileWithSkill(t *testing.T, baseDir, profileName, skillName string) string {
	t.Helper()
	skillDir := filepath.Join(baseDir, ".agent-smith", "profiles", profileName, "skills", skillName)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("failed to create skill dir for profile %s: %v", profileName, err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Skill: "+skillName), 0644); err != nil {
		t.Fatalf("failed to write skill file: %v", err)
	}
	return skillDir
}

// activateProfile runs `profile activate <name>` via the binary.
func activateProfile(t *testing.T, binaryPath, profileName string) {
	t.Helper()
	cmd := exec.Command(binaryPath, "profile", "activate", profileName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("profile activate %s failed: %v\nOutput: %s", profileName, err, output)
	}
}

// linkSkill runs `link skill <name>` via the binary.
func linkSkill(t *testing.T, binaryPath, skillName string) {
	t.Helper()
	cmd := exec.Command(binaryPath, "link", "skill", skillName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("link skill %s failed: %v\nOutput: %s", skillName, err, output)
	}
}

// assertSymlinkExists asserts that a symlink exists at path.
func assertSymlinkExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); err != nil {
		t.Errorf("expected symlink at %s: %v", path, err)
	}
}

// assertSymlinkAbsent asserts that no entry exists at path.
func assertSymlinkAbsent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Errorf("expected no entry at %s, but it exists", path)
	}
}

// TestE2E_MixedProfiles_ActiveWithOtherProfilesSymlinks verifies that `unlink all` (active profile
// only) removes only the active profile's components while preserving others'.
func TestE2E_MixedProfiles_ActiveWithOtherProfilesSymlinks(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-mixed-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	binaryPath := AgentSmithBinary

	opencodeDir := filepath.Join(tempDir, ".config", "opencode")
	if err := os.MkdirAll(opencodeDir, 0755); err != nil {
		t.Fatalf("failed to create opencode dir: %v", err)
	}

	createProfileWithSkill(t, tempDir, "mixed-profile-a", "skill-from-a")
	createProfileWithSkill(t, tempDir, "mixed-profile-b", "skill-from-b")
	createProfileWithSkill(t, tempDir, "mixed-profile-c", "skill-from-c")

	// Link skills from all three profiles (activate each before linking)
	for _, tc := range []struct{ profile, skill string }{
		{"mixed-profile-a", "skill-from-a"},
		{"mixed-profile-b", "skill-from-b"},
		{"mixed-profile-c", "skill-from-c"},
	} {
		activateProfile(t, binaryPath, tc.profile)
		linkSkill(t, binaryPath, tc.skill)
	}

	skillAPath := filepath.Join(opencodeDir, "skills", "skill-from-a")
	skillBPath := filepath.Join(opencodeDir, "skills", "skill-from-b")
	skillCPath := filepath.Join(opencodeDir, "skills", "skill-from-c")

	// Verify all three symlinks exist
	t.Run("Step1_VerifyAllLinked", func(t *testing.T) {
		assertSymlinkExists(t, skillAPath)
		assertSymlinkExists(t, skillBPath)
		assertSymlinkExists(t, skillCPath)
	})

	// Activate profile-a and unlink (active profile only)
	t.Run("Step2_UnlinkActiveProfileOnly", func(t *testing.T) {
		activateProfile(t, binaryPath, "mixed-profile-a")

		cmd := exec.Command(binaryPath, "unlink", "all", "--force")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("unlink all failed: %v\nOutput: %s", err, output)
		}
		t.Logf("Output:\n%s", output)

		assertSymlinkAbsent(t, skillAPath)
		assertSymlinkExists(t, skillBPath)
		assertSymlinkExists(t, skillCPath)
	})

	// Unlink all profiles
	t.Run("Step3_UnlinkAllProfiles", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "unlink", "all", "--force", "--all-profiles")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("unlink all --all-profiles failed: %v\nOutput: %s", err, output)
		}
		t.Logf("Output:\n%s", output)

		assertSymlinkAbsent(t, skillBPath)
		assertSymlinkAbsent(t, skillCPath)
	})
}

// TestE2E_MixedProfiles_BrokenSymlinksHandledGracefully verifies that broken symlinks left by
// deleted profiles are preserved (not removed) by `unlink all` (active profile only), and removed
// by `unlink all --all-profiles`.
func TestE2E_MixedProfiles_BrokenSymlinksHandledGracefully(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-broken-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	binaryPath := AgentSmithBinary

	opencodeDir := filepath.Join(tempDir, ".config", "opencode")
	if err := os.MkdirAll(opencodeDir, 0755); err != nil {
		t.Fatalf("failed to create opencode dir: %v", err)
	}

	createProfileWithSkill(t, tempDir, "profile-to-delete", "skill-will-be-broken")
	createProfileWithSkill(t, tempDir, "active-profile", "skill-active")

	// Link the skill that will become broken
	activateProfile(t, binaryPath, "profile-to-delete")
	linkSkill(t, binaryPath, "skill-will-be-broken")

	// Link the active skill
	activateProfile(t, binaryPath, "active-profile")
	linkSkill(t, binaryPath, "skill-active")

	brokenPath := filepath.Join(opencodeDir, "skills", "skill-will-be-broken")
	activePath := filepath.Join(opencodeDir, "skills", "skill-active")

	// Delete the profile to create a broken symlink
	if err := os.RemoveAll(filepath.Join(tempDir, ".agent-smith", "profiles", "profile-to-delete")); err != nil {
		t.Fatalf("failed to delete profile: %v", err)
	}

	t.Run("Step1_BrokenSymlinkStillExists", func(t *testing.T) {
		// Lstat works on broken symlinks; Stat would fail
		if _, err := os.Lstat(brokenPath); err != nil {
			t.Errorf("broken symlink should still exist at %s: %v", brokenPath, err)
		}
	})

	t.Run("Step2_UnlinkActiveOnlyPreservesBroken", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "unlink", "all", "--force")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("unlink all failed: %v\nOutput: %s", err, output)
		}
		t.Logf("Output:\n%s", output)

		// Active profile's skill should be removed
		assertSymlinkAbsent(t, activePath)
		// Broken symlink from deleted profile should be preserved
		if _, err := os.Lstat(brokenPath); err != nil {
			t.Errorf("broken symlink from other profile should be preserved, got: %v", err)
		}
	})

	t.Run("Step3_UnlinkAllProfilesRemovesBroken", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "unlink", "all", "--force", "--all-profiles")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("unlink all --all-profiles failed: %v\nOutput: %s", err, output)
		}
		t.Logf("Output:\n%s", output)

		assertSymlinkAbsent(t, brokenPath)
	})
}

// TestE2E_MixedProfiles_ManualSymlinksPreserved verifies that manually created symlinks
// (not from agent-smith) are never removed by `unlink all`.
func TestE2E_MixedProfiles_ManualSymlinksPreserved(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-manual-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	binaryPath := AgentSmithBinary

	opencodeDir := filepath.Join(tempDir, ".config", "opencode")
	if err := os.MkdirAll(opencodeDir, 0755); err != nil {
		t.Fatalf("failed to create opencode dir: %v", err)
	}

	createProfileWithSkill(t, tempDir, "test-profile", "skill-from-profile")
	activateProfile(t, binaryPath, "test-profile")
	linkSkill(t, binaryPath, "skill-from-profile")

	// Create a manual symlink pointing outside agent-smith
	manualTarget := testutil.CreateTempDir(t, "manual-skill-*")
	if err := os.WriteFile(filepath.Join(manualTarget, "SKILL.md"), []byte("# Manual"), 0644); err != nil {
		t.Fatalf("failed to write manual skill file: %v", err)
	}

	skillsDir := filepath.Join(opencodeDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("failed to create skills dir: %v", err)
	}

	manualSymlink := filepath.Join(skillsDir, "manual-skill")
	if err := os.Symlink(manualTarget, manualSymlink); err != nil {
		t.Fatalf("failed to create manual symlink: %v", err)
	}

	profileSkillPath := filepath.Join(opencodeDir, "skills", "skill-from-profile")

	t.Run("Step1_BothSymlinksExist", func(t *testing.T) {
		assertSymlinkExists(t, profileSkillPath)
		assertSymlinkExists(t, manualSymlink)
	})

	t.Run("Step2_UnlinkRemovesProfileSkillPreservesManual", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "unlink", "all", "--force")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("unlink all failed: %v\nOutput: %s", err, output)
		}
		t.Logf("Output:\n%s", output)

		assertSymlinkAbsent(t, profileSkillPath)
		assertSymlinkExists(t, manualSymlink)
	})
}

// TestE2E_MixedProfiles_EmptyProfileHandledGracefully verifies that unlinking with an empty
// active profile does not remove components from other profiles.
func TestE2E_MixedProfiles_EmptyProfileHandledGracefully(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-empty-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	binaryPath := AgentSmithBinary

	opencodeDir := filepath.Join(tempDir, ".config", "opencode")
	if err := os.MkdirAll(opencodeDir, 0755); err != nil {
		t.Fatalf("failed to create opencode dir: %v", err)
	}

	// Create populated profile and link its skill
	createProfileWithSkill(t, tempDir, "populated-profile", "skill-populated")
	activateProfile(t, binaryPath, "populated-profile")
	linkSkill(t, binaryPath, "skill-populated")

	// Create an empty profile
	emptyProfileDir := filepath.Join(tempDir, ".agent-smith", "profiles", "empty-profile", "skills")
	if err := os.MkdirAll(emptyProfileDir, 0755); err != nil {
		t.Fatalf("failed to create empty profile: %v", err)
	}

	// Switch to empty profile
	activateProfile(t, binaryPath, "empty-profile")

	populatedSkillPath := filepath.Join(opencodeDir, "skills", "skill-populated")

	t.Run("Step1_UnlinkWithEmptyActiveProfile", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "unlink", "all", "--force")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("unlink all failed: %v\nOutput: %s", err, output)
		}
		t.Logf("Output:\n%s", output)

		// Populated profile's skill should be preserved (different profile)
		assertSymlinkExists(t, populatedSkillPath)
	})

	t.Run("Step2_UnlinkAllProfilesRemovesIt", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "unlink", "all", "--force", "--all-profiles")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("unlink all --all-profiles failed: %v\nOutput: %s", err, output)
		}
		t.Logf("Output:\n%s", output)

		assertSymlinkAbsent(t, populatedSkillPath)
	})
}
