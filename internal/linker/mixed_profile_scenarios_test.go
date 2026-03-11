package linker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tjg184/agent-smith/internal/detector"
	"github.com/tjg184/agent-smith/pkg/config"
	"github.com/tjg184/agent-smith/pkg/paths"
)

// Story-008 Integration Tests: Mixed-Profile Scenarios
//
// This file contains comprehensive tests for Story-008 from the Profile-Aware Link/Unlink PRD:
// "As a developer, I want comprehensive tests for mixed-profile scenarios so that edge cases
// are properly handled."
//
// Test Coverage:
// - Active profile with other profiles' symlinks present
// - Broken symlinks pointing to deleted profiles
// - Manually created symlinks outside agent-smith
// - Empty profiles (no components)
// - Profile with only some component types

// TestMixedProfiles_ActiveWithOtherProfilesSymlinks tests Story-008 AC #1:
// When profile A is active but symlinks from profiles A, B, and C exist,
// only profile A's components should be unlinked by default
func TestMixedProfiles_ActiveWithOtherProfilesSymlinks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test environment
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		t.Fatalf("Failed to get profiles directory: %v", err)
	}

	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		t.Fatalf("Failed to get agents directory: %v", err)
	}

	// Create three test profiles
	profileA := filepath.Join(profilesDir, "test-mixed-profile-a")
	profileB := filepath.Join(profilesDir, "test-mixed-profile-b")
	profileC := filepath.Join(profilesDir, "test-mixed-profile-c")

	defer func() {
		os.RemoveAll(profileA)
		os.RemoveAll(profileB)
		os.RemoveAll(profileC)
		os.Remove(filepath.Join(agentsDir, ".active-profile"))
	}()

	// Create components in each profile
	setupProfileWithSkill(t, profileA, "skill-from-a")
	setupProfileWithSkill(t, profileB, "skill-from-b")
	setupProfileWithSkill(t, profileC, "skill-from-c")

	// Create a temp target directory
	tempTargetDir, err := os.MkdirTemp("", "agent-smith-mixed-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp target dir: %v", err)
	}
	defer os.RemoveAll(tempTargetDir)

	targets := []config.Target{
		&mockTarget{name: "test-target", baseDir: tempTargetDir},
	}

	det := detector.NewRepositoryDetector()

	// Set profile A as active
	activeProfilePath := filepath.Join(agentsDir, ".active-profile")
	if err := os.WriteFile(activeProfilePath, []byte("test-mixed-profile-a"), 0644); err != nil {
		t.Fatalf("Failed to set active profile: %v", err)
	}

	// Link components from all three profiles
	linkerA, _ := NewComponentLinker(profileA, targets, det, nil)
	if err := linkerA.LinkComponent("skills", "skill-from-a"); err != nil {
		t.Fatalf("Failed to link from profile A: %v", err)
	}

	linkerB, _ := NewComponentLinker(profileB, targets, det, nil)
	if err := linkerB.LinkComponent("skills", "skill-from-b"); err != nil {
		t.Fatalf("Failed to link from profile B: %v", err)
	}

	linkerC, _ := NewComponentLinker(profileC, targets, det, nil)
	if err := linkerC.LinkComponent("skills", "skill-from-c"); err != nil {
		t.Fatalf("Failed to link from profile C: %v", err)
	}

	// Verify all three symlinks exist
	skillAPath := filepath.Join(tempTargetDir, "skills", "skill-from-a")
	skillBPath := filepath.Join(tempTargetDir, "skills", "skill-from-b")
	skillCPath := filepath.Join(tempTargetDir, "skills", "skill-from-c")

	if _, err := os.Lstat(skillAPath); os.IsNotExist(err) {
		t.Error("skill-from-a should exist")
	}
	if _, err := os.Lstat(skillBPath); os.IsNotExist(err) {
		t.Error("skill-from-b should exist")
	}
	if _, err := os.Lstat(skillCPath); os.IsNotExist(err) {
		t.Error("skill-from-c should exist")
	}

	// Unlink all components with profile A active (allProfiles=false)
	if err := linkerA.UnlinkAllComponents("", true, false); err != nil {
		t.Fatalf("UnlinkAllComponents failed: %v", err)
	}

	// Verify: skill-from-a should be removed, skill-from-b and skill-from-c should remain
	if _, err := os.Lstat(skillAPath); !os.IsNotExist(err) {
		t.Error("skill-from-a should be removed (from active profile)")
	}
	if _, err := os.Lstat(skillBPath); os.IsNotExist(err) {
		t.Error("skill-from-b should still exist (from different profile)")
	}
	if _, err := os.Lstat(skillCPath); os.IsNotExist(err) {
		t.Error("skill-from-c should still exist (from different profile)")
	}

	// Now test with allProfiles=true - should remove all
	if err := linkerA.UnlinkAllComponents("", true, true); err != nil {
		t.Fatalf("UnlinkAllComponents with allProfiles failed: %v", err)
	}

	// Verify: all symlinks should be removed
	if _, err := os.Lstat(skillBPath); !os.IsNotExist(err) {
		t.Error("skill-from-b should be removed with --all-profiles")
	}
	if _, err := os.Lstat(skillCPath); !os.IsNotExist(err) {
		t.Error("skill-from-c should be removed with --all-profiles")
	}
}

// TestMixedProfiles_BrokenSymlinksToDeletedProfile tests Story-008 AC #2:
// Broken symlinks pointing to deleted profiles should be handled gracefully
func TestMixedProfiles_BrokenSymlinksToDeletedProfile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		t.Fatalf("Failed to get profiles directory: %v", err)
	}

	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		t.Fatalf("Failed to get agents directory: %v", err)
	}

	// Create a test profile that we'll delete later
	profileToDelete := filepath.Join(profilesDir, "test-profile-to-delete")
	activeProfile := filepath.Join(profilesDir, "test-active-profile")

	defer func() {
		os.RemoveAll(profileToDelete)
		os.RemoveAll(activeProfile)
		os.Remove(filepath.Join(agentsDir, ".active-profile"))
	}()

	setupProfileWithSkill(t, profileToDelete, "skill-will-be-broken")
	setupProfileWithSkill(t, activeProfile, "skill-active")

	tempTargetDir, err := os.MkdirTemp("", "agent-smith-broken-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp target dir: %v", err)
	}
	defer os.RemoveAll(tempTargetDir)

	targets := []config.Target{
		&mockTarget{name: "test-target", baseDir: tempTargetDir},
	}

	det := detector.NewRepositoryDetector()

	// Link component from profile that will be deleted
	linkerDelete, _ := NewComponentLinker(profileToDelete, targets, det, nil)
	if err := linkerDelete.LinkComponent("skills", "skill-will-be-broken"); err != nil {
		t.Fatalf("Failed to link from profile-to-delete: %v", err)
	}

	// Set active profile
	activeProfilePath := filepath.Join(agentsDir, ".active-profile")
	if err := os.WriteFile(activeProfilePath, []byte("test-active-profile"), 0644); err != nil {
		t.Fatalf("Failed to set active profile: %v", err)
	}

	// Link component from active profile
	linkerActive, _ := NewComponentLinker(activeProfile, targets, det, nil)
	if err := linkerActive.LinkComponent("skills", "skill-active"); err != nil {
		t.Fatalf("Failed to link from active profile: %v", err)
	}

	// Delete the profile directory, creating a broken symlink
	os.RemoveAll(profileToDelete)

	// Verify the symlink is now broken
	brokenSkillPath := filepath.Join(tempTargetDir, "skills", "skill-will-be-broken")
	if _, err := os.Stat(brokenSkillPath); !os.IsNotExist(err) {
		t.Error("Expected broken symlink (stat should fail)")
	}
	if _, err := os.Lstat(brokenSkillPath); os.IsNotExist(err) {
		t.Error("Symlink should still exist (lstat should succeed)")
	}

	// Unlink all components - should handle broken symlink gracefully
	err = linkerActive.UnlinkAllComponents("", true, false)
	if err != nil {
		t.Fatalf("UnlinkAllComponents should handle broken symlinks: %v", err)
	}

	// The broken symlink should be skipped (it's from a different profile)
	if _, err := os.Lstat(brokenSkillPath); os.IsNotExist(err) {
		t.Error("Broken symlink from other profile should be skipped, not removed")
	}

	// But active profile's component should be removed
	activeSkillPath := filepath.Join(tempTargetDir, "skills", "skill-active")
	if _, err := os.Lstat(activeSkillPath); !os.IsNotExist(err) {
		t.Error("skill-active should be removed")
	}

	// Now unlink with allProfiles=true - should remove broken symlink too
	err = linkerActive.UnlinkAllComponents("", true, true)
	if err != nil {
		t.Fatalf("UnlinkAllComponents with allProfiles should handle broken symlinks: %v", err)
	}

	// Broken symlink should now be removed
	if _, err := os.Lstat(brokenSkillPath); !os.IsNotExist(err) {
		t.Error("Broken symlink should be removed with --all-profiles")
	}
}

// TestMixedProfiles_ManualSymlinksOutsideAgentSmith tests Story-008 AC #3:
// Manually created symlinks outside agent-smith should be preserved
func TestMixedProfiles_ManualSymlinksOutsideAgentSmith(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		t.Fatalf("Failed to get profiles directory: %v", err)
	}

	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		t.Fatalf("Failed to get agents directory: %v", err)
	}

	// Create a test profile
	testProfile := filepath.Join(profilesDir, "test-manual-profile")

	defer func() {
		os.RemoveAll(testProfile)
		os.Remove(filepath.Join(agentsDir, ".active-profile"))
	}()

	setupProfileWithSkill(t, testProfile, "skill-from-profile")

	tempTargetDir, err := os.MkdirTemp("", "agent-smith-manual-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp target dir: %v", err)
	}
	defer os.RemoveAll(tempTargetDir)

	targets := []config.Target{
		&mockTarget{name: "test-target", baseDir: tempTargetDir},
	}

	det := detector.NewRepositoryDetector()

	// Set active profile
	activeProfilePath := filepath.Join(agentsDir, ".active-profile")
	if err := os.WriteFile(activeProfilePath, []byte("test-manual-profile"), 0644); err != nil {
		t.Fatalf("Failed to set active profile: %v", err)
	}

	// Link a real agent-smith component
	linker, _ := NewComponentLinker(testProfile, targets, det, nil)
	if err := linker.LinkComponent("skills", "skill-from-profile"); err != nil {
		t.Fatalf("Failed to link from profile: %v", err)
	}

	// Create a manual symlink pointing to some arbitrary location (not agent-smith)
	// We'll use a location outside the target directory completely
	manualSkillDir, err := os.MkdirTemp("", "manual-skill-*")
	if err != nil {
		t.Fatalf("Failed to create manual skill dir: %v", err)
	}
	defer os.RemoveAll(manualSkillDir)

	if err := os.WriteFile(filepath.Join(manualSkillDir, "SKILL.md"), []byte("# Manual Skill"), 0644); err != nil {
		t.Fatalf("Failed to create manual skill file: %v", err)
	}

	skillsDir := filepath.Join(tempTargetDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	manualSymlinkPath := filepath.Join(skillsDir, "manual-skill")
	if err := os.Symlink(manualSkillDir, manualSymlinkPath); err != nil {
		t.Fatalf("Failed to create manual symlink: %v", err)
	}

	// Verify both symlinks exist
	profileSkillPath := filepath.Join(tempTargetDir, "skills", "skill-from-profile")
	if _, err := os.Lstat(profileSkillPath); os.IsNotExist(err) {
		t.Error("Profile skill should exist")
	}
	if _, err := os.Lstat(manualSymlinkPath); os.IsNotExist(err) {
		t.Error("Manual symlink should exist")
	}

	// Unlink all components
	if err := linker.UnlinkAllComponents("", true, false); err != nil {
		t.Fatalf("UnlinkAllComponents failed: %v", err)
	}

	// Verify: profile skill should be removed, manual symlink should be preserved
	if _, err := os.Lstat(profileSkillPath); !os.IsNotExist(err) {
		t.Error("Profile skill should be removed")
	}
	if _, err := os.Lstat(manualSymlinkPath); os.IsNotExist(err) {
		t.Error("Manual symlink should be preserved (not from agent-smith)")
	}

	// Verify manual symlink still works
	target, err := os.Readlink(manualSymlinkPath)
	if err != nil {
		t.Fatalf("Failed to read manual symlink: %v", err)
	}

	// For absolute symlinks, just compare directly
	// For relative symlinks, resolve to absolute
	var absTarget string
	if filepath.IsAbs(target) {
		absTarget = target
	} else {
		absTarget = filepath.Clean(filepath.Join(filepath.Dir(manualSymlinkPath), target))
	}

	expectedTarget := filepath.Clean(manualSkillDir)
	if absTarget != expectedTarget {
		t.Errorf("Manual symlink target changed. Expected: %s, Got: %s", expectedTarget, absTarget)
	}
}

// TestMixedProfiles_EmptyProfile tests Story-008 AC #4:
// Empty profiles (with no components) should be handled gracefully
func TestMixedProfiles_EmptyProfile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		t.Fatalf("Failed to get profiles directory: %v", err)
	}

	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		t.Fatalf("Failed to get agents directory: %v", err)
	}

	// Create two profiles: one empty, one with components
	emptyProfile := filepath.Join(profilesDir, "test-empty-profile")
	populatedProfile := filepath.Join(profilesDir, "test-populated-profile")

	defer func() {
		os.RemoveAll(emptyProfile)
		os.RemoveAll(populatedProfile)
		os.Remove(filepath.Join(agentsDir, ".active-profile"))
	}()

	// Create empty profile (just directory structure, no components)
	if err := os.MkdirAll(filepath.Join(emptyProfile, "skills"), 0755); err != nil {
		t.Fatalf("Failed to create empty profile: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(emptyProfile, "agents"), 0755); err != nil {
		t.Fatalf("Failed to create empty profile agents dir: %v", err)
	}

	// Create populated profile
	setupProfileWithSkill(t, populatedProfile, "skill-populated")

	tempTargetDir, err := os.MkdirTemp("", "agent-smith-empty-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp target dir: %v", err)
	}
	defer os.RemoveAll(tempTargetDir)

	targets := []config.Target{
		&mockTarget{name: "test-target", baseDir: tempTargetDir},
	}

	det := detector.NewRepositoryDetector()

	// Set empty profile as active
	activeProfilePath := filepath.Join(agentsDir, ".active-profile")
	if err := os.WriteFile(activeProfilePath, []byte("test-empty-profile"), 0644); err != nil {
		t.Fatalf("Failed to set active profile: %v", err)
	}

	// Link from populated profile (while empty is active)
	linkerPopulated, _ := NewComponentLinker(populatedProfile, targets, det, nil)
	if err := linkerPopulated.LinkComponent("skills", "skill-populated"); err != nil {
		t.Fatalf("Failed to link from populated profile: %v", err)
	}

	// Create linker with empty profile active
	linkerEmpty, _ := NewComponentLinker(emptyProfile, targets, det, nil)

	// Try to link from empty profile - should fail gracefully
	err = linkerEmpty.LinkComponent("skills", "non-existent-skill")
	if err == nil {
		t.Error("Expected error when linking from empty profile")
	}

	// Unlink all with empty profile active - should skip populated profile's component
	if err := linkerEmpty.UnlinkAllComponents("", true, false); err != nil {
		t.Fatalf("UnlinkAllComponents from empty profile failed: %v", err)
	}

	// Verify: populated profile's component should still exist
	populatedSkillPath := filepath.Join(tempTargetDir, "skills", "skill-populated")
	if _, err := os.Lstat(populatedSkillPath); os.IsNotExist(err) {
		t.Error("Populated profile's skill should be preserved")
	}

	// Unlink with allProfiles=true should remove it
	if err := linkerEmpty.UnlinkAllComponents("", true, true); err != nil {
		t.Fatalf("UnlinkAllComponents with allProfiles failed: %v", err)
	}

	if _, err := os.Lstat(populatedSkillPath); !os.IsNotExist(err) {
		t.Error("Populated profile's skill should be removed with --all-profiles")
	}
}

// TestMixedProfiles_PartialProfile tests Story-008 AC #5:
// Profiles with only some component types should be handled correctly
func TestMixedProfiles_PartialProfile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		t.Fatalf("Failed to get profiles directory: %v", err)
	}

	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		t.Fatalf("Failed to get agents directory: %v", err)
	}

	// Create two profiles: one with only skills, one with only agents
	skillsOnlyProfile := filepath.Join(profilesDir, "test-skills-only-profile")
	agentsOnlyProfile := filepath.Join(profilesDir, "test-agents-only-profile")

	defer func() {
		os.RemoveAll(skillsOnlyProfile)
		os.RemoveAll(agentsOnlyProfile)
		os.Remove(filepath.Join(agentsDir, ".active-profile"))
	}()

	// Create profile with only skills
	setupProfileWithSkill(t, skillsOnlyProfile, "only-skill")

	// Create profile with only agents
	agentDir := filepath.Join(agentsOnlyProfile, "agents", "only-agent")
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		t.Fatalf("Failed to create agent dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentDir, "AGENT.md"), []byte("# Test Agent"), 0644); err != nil {
		t.Fatalf("Failed to create agent file: %v", err)
	}

	tempTargetDir, err := os.MkdirTemp("", "agent-smith-partial-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp target dir: %v", err)
	}
	defer os.RemoveAll(tempTargetDir)

	targets := []config.Target{
		&mockTarget{name: "test-target", baseDir: tempTargetDir},
	}

	det := detector.NewRepositoryDetector()

	// Link skill from skills-only profile
	linkerSkills, _ := NewComponentLinker(skillsOnlyProfile, targets, det, nil)
	if err := linkerSkills.LinkComponent("skills", "only-skill"); err != nil {
		t.Fatalf("Failed to link skill: %v", err)
	}

	// Link agent from agents-only profile
	linkerAgents, _ := NewComponentLinker(agentsOnlyProfile, targets, det, nil)
	if err := linkerAgents.LinkComponent("agents", "only-agent"); err != nil {
		t.Fatalf("Failed to link agent: %v", err)
	}

	// Set skills-only profile as active
	activeProfilePath := filepath.Join(agentsDir, ".active-profile")
	if err := os.WriteFile(activeProfilePath, []byte("test-skills-only-profile"), 0644); err != nil {
		t.Fatalf("Failed to set active profile: %v", err)
	}

	// Verify both components are linked
	skillPath := filepath.Join(tempTargetDir, "skills", "only-skill")
	agentPath := filepath.Join(tempTargetDir, "agents", "AGENT.md")

	if _, err := os.Lstat(skillPath); os.IsNotExist(err) {
		t.Error("Skill should be linked")
	}
	if _, err := os.Lstat(agentPath); os.IsNotExist(err) {
		t.Error("Agent should be linked")
	}

	// Unlink all with skills-only profile active
	if err := linkerSkills.UnlinkAllComponents("", true, false); err != nil {
		t.Fatalf("UnlinkAllComponents failed: %v", err)
	}

	// Verify: only skill should be removed (from active profile)
	if _, err := os.Lstat(skillPath); !os.IsNotExist(err) {
		t.Error("Skill should be removed (from active profile)")
	}
	if _, err := os.Lstat(agentPath); os.IsNotExist(err) {
		t.Error("Agent should be preserved (from different profile)")
	}

	// Switch to agents-only profile
	if err := os.WriteFile(activeProfilePath, []byte("test-agents-only-profile"), 0644); err != nil {
		t.Fatalf("Failed to switch active profile: %v", err)
	}

	// Unlink all with agents-only profile active
	if err := linkerAgents.UnlinkAllComponents("", true, false); err != nil {
		t.Fatalf("UnlinkAllComponents with agents profile failed: %v", err)
	}

	// Verify: agent should be removed now
	if _, err := os.Lstat(agentPath); !os.IsNotExist(err) {
		t.Error("Agent should be removed (from now-active profile)")
	}
}

// TestMixedProfiles_AllEdgeCasesCombined tests all edge cases together
func TestMixedProfiles_AllEdgeCasesCombined(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		t.Fatalf("Failed to get profiles directory: %v", err)
	}

	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		t.Fatalf("Failed to get agents directory: %v", err)
	}

	// Create multiple profiles with various characteristics
	activeProfile := filepath.Join(profilesDir, "test-combined-active")
	emptyProfile := filepath.Join(profilesDir, "test-combined-empty")
	partialProfile := filepath.Join(profilesDir, "test-combined-partial")
	deletedProfile := filepath.Join(profilesDir, "test-combined-deleted")

	defer func() {
		os.RemoveAll(activeProfile)
		os.RemoveAll(emptyProfile)
		os.RemoveAll(partialProfile)
		os.RemoveAll(deletedProfile)
		os.Remove(filepath.Join(agentsDir, ".active-profile"))
	}()

	// Setup profiles
	setupProfileWithSkill(t, activeProfile, "skill-active")
	os.MkdirAll(filepath.Join(emptyProfile, "skills"), 0755)
	setupProfileWithSkill(t, partialProfile, "skill-partial")
	setupProfileWithSkill(t, deletedProfile, "skill-will-break")

	tempTargetDir, err := os.MkdirTemp("", "agent-smith-combined-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp target dir: %v", err)
	}
	defer os.RemoveAll(tempTargetDir)

	targets := []config.Target{
		&mockTarget{name: "test-target", baseDir: tempTargetDir},
	}

	det := detector.NewRepositoryDetector()

	// Link components from various profiles
	linkerActive, _ := NewComponentLinker(activeProfile, targets, det, nil)
	linkerPartial, _ := NewComponentLinker(partialProfile, targets, det, nil)
	linkerDeleted, _ := NewComponentLinker(deletedProfile, targets, det, nil)

	if err := linkerActive.LinkComponent("skills", "skill-active"); err != nil {
		t.Fatalf("Failed to link from active: %v", err)
	}
	if err := linkerPartial.LinkComponent("skills", "skill-partial"); err != nil {
		t.Fatalf("Failed to link from partial: %v", err)
	}
	if err := linkerDeleted.LinkComponent("skills", "skill-will-break"); err != nil {
		t.Fatalf("Failed to link from deleted: %v", err)
	}

	// Create manual symlink
	manualSkillDir, err := os.MkdirTemp("", "manual-skill-*")
	if err != nil {
		t.Fatalf("Failed to create manual skill dir: %v", err)
	}
	defer os.RemoveAll(manualSkillDir)

	os.WriteFile(filepath.Join(manualSkillDir, "SKILL.md"), []byte("# Manual"), 0644)

	skillsDir := filepath.Join(tempTargetDir, "skills")
	os.MkdirAll(skillsDir, 0755)
	manualSymlink := filepath.Join(skillsDir, "manual-skill")
	os.Symlink(manualSkillDir, manualSymlink)

	// Delete profile to create broken symlink
	os.RemoveAll(deletedProfile)

	// Set active profile
	activeProfilePath := filepath.Join(agentsDir, ".active-profile")
	if err := os.WriteFile(activeProfilePath, []byte("test-combined-active"), 0644); err != nil {
		t.Fatalf("Failed to set active profile: %v", err)
	}

	// Count initial components
	activePath := filepath.Join(tempTargetDir, "skills", "skill-active")
	partialPath := filepath.Join(tempTargetDir, "skills", "skill-partial")
	brokenPath := filepath.Join(tempTargetDir, "skills", "skill-will-break")

	if _, err := os.Lstat(activePath); os.IsNotExist(err) {
		t.Error("Active skill should exist")
	}
	if _, err := os.Lstat(partialPath); os.IsNotExist(err) {
		t.Error("Partial skill should exist")
	}
	if _, err := os.Lstat(brokenPath); os.IsNotExist(err) {
		t.Error("Broken symlink should still exist")
	}
	if _, err := os.Lstat(manualSymlink); os.IsNotExist(err) {
		t.Error("Manual symlink should exist")
	}

	// Unlink with active profile only
	if err := linkerActive.UnlinkAllComponents("", true, false); err != nil {
		t.Fatalf("UnlinkAllComponents failed: %v", err)
	}

	// Verify results:
	// - skill-active: removed (from active profile)
	// - skill-partial: preserved (from different profile)
	// - skill-will-break: preserved (broken, from different profile)
	// - manual-skill: preserved (not from agent-smith)

	if _, err := os.Lstat(activePath); !os.IsNotExist(err) {
		t.Error("Active skill should be removed")
	}
	if _, err := os.Lstat(partialPath); os.IsNotExist(err) {
		t.Error("Partial skill should be preserved (different profile)")
	}
	if _, err := os.Lstat(brokenPath); os.IsNotExist(err) {
		t.Error("Broken symlink should be preserved (different profile)")
	}
	if _, err := os.Lstat(manualSymlink); os.IsNotExist(err) {
		t.Error("Manual symlink should be preserved (not agent-smith)")
	}

	// Unlink with --all-profiles
	if err := linkerActive.UnlinkAllComponents("", true, true); err != nil {
		t.Fatalf("UnlinkAllComponents with allProfiles failed: %v", err)
	}

	// Now everything except manual symlink should be removed
	if _, err := os.Lstat(partialPath); !os.IsNotExist(err) {
		t.Error("Partial skill should be removed with --all-profiles")
	}
	if _, err := os.Lstat(brokenPath); !os.IsNotExist(err) {
		t.Error("Broken symlink should be removed with --all-profiles")
	}
	if _, err := os.Lstat(manualSymlink); os.IsNotExist(err) {
		t.Error("Manual symlink should still be preserved")
	}
}

// Helper function to setup a profile with a test skill
func setupProfileWithSkill(t *testing.T, profilePath, skillName string) {
	skillDir := filepath.Join(profilePath, "skills", skillName)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create skill dir: %v", err)
	}
	skillContent := []byte("# Test Skill: " + skillName)
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), skillContent, 0644); err != nil {
		t.Fatalf("Failed to create skill file: %v", err)
	}
}
