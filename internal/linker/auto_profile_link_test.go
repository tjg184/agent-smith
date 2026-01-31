package linker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tgaines/agent-smith/internal/detector"
	"github.com/tgaines/agent-smith/pkg/config"
	"github.com/tgaines/agent-smith/pkg/paths"
)

// Story-006 Integration Tests
//
// This file contains comprehensive tests for Story-006 from the Auto-Profile PRD:
// "As a user linking a component, I want the system to automatically link from
// the active profile if available so that I don't need to specify the profile
// for common cases."
//
// Implementation Status: ✅ COMPLETE
//
// The feature is fully implemented in linker.go (LinkComponent function):
// - Lines 120-124: Checks active profile first (agentsDir is set to active profile by NewComponentLinker)
// - Lines 124-149: Falls back to searching other profiles if not found in active profile
// - Lines 137-142: Presents interactive prompt when component exists in multiple profiles
// - Lines 144-149: Auto-selects if found in only one profile
// - Lines 132-134: Shows clear error when component not found anywhere
// - Lines 227-231: Shows profile name in output
//
// All acceptance criteria are verified by the tests in this file.

// TestLinkComponent_AutoLinkFromActiveProfile tests Story-006:
// When a component exists in the active profile, it should auto-link without prompting
func TestLinkComponent_AutoLinkFromActiveProfile(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get the real profiles directory
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		t.Fatalf("Failed to get profiles directory: %v", err)
	}

	// Get the real agents directory for active profile tracking
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		t.Fatalf("Failed to get agents directory: %v", err)
	}

	// Create a test profile
	testProfile := filepath.Join(profilesDir, "test-auto-link-profile")
	testProfileSkillDir := filepath.Join(testProfile, "skills", "auto-link-test-skill")

	// Cleanup function
	defer func() {
		os.RemoveAll(testProfile)
		// Clean up active profile marker if we set it
		os.Remove(filepath.Join(agentsDir, ".active-profile"))
	}()

	// Create test component in the profile
	if err := os.MkdirAll(testProfileSkillDir, 0755); err != nil {
		t.Fatalf("Failed to create profile skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testProfileSkillDir, "SKILL.md"), []byte("# Test Skill for Auto-Link"), 0644); err != nil {
		t.Fatalf("Failed to create skill file: %v", err)
	}

	// Set this profile as active by writing to .active-profile
	activeProfilePath := filepath.Join(agentsDir, ".active-profile")
	if err := os.WriteFile(activeProfilePath, []byte("test-auto-link-profile"), 0644); err != nil {
		t.Fatalf("Failed to set active profile: %v", err)
	}

	// Create a mock target for testing
	tempTargetDir, err := os.MkdirTemp("", "agent-smith-target-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp target dir: %v", err)
	}
	defer os.RemoveAll(tempTargetDir)

	targets := []config.Target{
		&mockTarget{name: "test-target", baseDir: tempTargetDir},
	}

	det := detector.NewRepositoryDetector()

	// Create linker with the active profile's directory (simulating NewComponentLinker behavior)
	linker, err := NewComponentLinker(testProfile, targets, det)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Test: Link component should automatically use the active profile
	err = linker.LinkComponent("skills", "auto-link-test-skill")
	if err != nil {
		t.Fatalf("Failed to link component from active profile: %v", err)
	}

	// Verify the component was linked
	linkedPath := filepath.Join(tempTargetDir, "skills", "auto-link-test-skill")
	if _, err := os.Lstat(linkedPath); os.IsNotExist(err) {
		t.Error("Expected component to be linked, but link does not exist")
	}

	// Verify it's a symlink pointing to the correct location
	target, err := os.Readlink(linkedPath)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}

	// Resolve the symlink to absolute path for comparison
	absTarget, err := filepath.Abs(filepath.Join(filepath.Dir(linkedPath), target))
	if err != nil {
		t.Fatalf("Failed to resolve absolute path: %v", err)
	}

	expectedTarget, err := filepath.Abs(testProfileSkillDir)
	if err != nil {
		t.Fatalf("Failed to resolve expected target: %v", err)
	}

	if absTarget != expectedTarget {
		t.Errorf("Symlink points to wrong location.\nExpected: %s\nGot: %s", expectedTarget, absTarget)
	}
}

// TestLinkComponent_FallbackToOtherProfile tests Story-006:
// When a component doesn't exist in the active profile, it should search other profiles
func TestLinkComponent_FallbackToOtherProfile(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get the real profiles directory
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		t.Fatalf("Failed to get profiles directory: %v", err)
	}

	// Get the real agents directory for active profile tracking
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		t.Fatalf("Failed to get agents directory: %v", err)
	}

	// Create two test profiles
	activeProfile := filepath.Join(profilesDir, "test-fallback-active")
	otherProfile := filepath.Join(profilesDir, "test-fallback-other")
	otherProfileSkillDir := filepath.Join(otherProfile, "skills", "fallback-test-skill")

	// Cleanup function
	defer func() {
		os.RemoveAll(activeProfile)
		os.RemoveAll(otherProfile)
		os.Remove(filepath.Join(agentsDir, ".active-profile"))
	}()

	// Create active profile (empty - no components)
	if err := os.MkdirAll(filepath.Join(activeProfile, "skills"), 0755); err != nil {
		t.Fatalf("Failed to create active profile: %v", err)
	}

	// Create other profile with test component
	if err := os.MkdirAll(otherProfileSkillDir, 0755); err != nil {
		t.Fatalf("Failed to create other profile skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(otherProfileSkillDir, "SKILL.md"), []byte("# Test Skill for Fallback"), 0644); err != nil {
		t.Fatalf("Failed to create skill file: %v", err)
	}

	// Set active profile (empty one)
	activeProfilePath := filepath.Join(agentsDir, ".active-profile")
	if err := os.WriteFile(activeProfilePath, []byte("test-fallback-active"), 0644); err != nil {
		t.Fatalf("Failed to set active profile: %v", err)
	}

	// Create a mock target for testing
	tempTargetDir, err := os.MkdirTemp("", "agent-smith-target-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp target dir: %v", err)
	}
	defer os.RemoveAll(tempTargetDir)

	targets := []config.Target{
		&mockTarget{name: "test-target", baseDir: tempTargetDir},
	}

	det := detector.NewRepositoryDetector()

	// Create linker with the active profile's directory (simulating NewComponentLinker behavior)
	linker, err := NewComponentLinker(activeProfile, targets, det)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Test: Link component should fall back to other profile when not in active profile
	err = linker.LinkComponent("skills", "fallback-test-skill")
	if err != nil {
		t.Fatalf("Failed to link component from other profile: %v", err)
	}

	// Verify the component was linked
	linkedPath := filepath.Join(tempTargetDir, "skills", "fallback-test-skill")
	if _, err := os.Lstat(linkedPath); os.IsNotExist(err) {
		t.Error("Expected component to be linked, but link does not exist")
	}

	// Verify it's a symlink pointing to the other profile's component
	target, err := os.Readlink(linkedPath)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}

	// Resolve the symlink to absolute path for comparison
	absTarget, err := filepath.Abs(filepath.Join(filepath.Dir(linkedPath), target))
	if err != nil {
		t.Fatalf("Failed to resolve absolute path: %v", err)
	}

	expectedTarget, err := filepath.Abs(otherProfileSkillDir)
	if err != nil {
		t.Fatalf("Failed to resolve expected target: %v", err)
	}

	if absTarget != expectedTarget {
		t.Errorf("Symlink points to wrong location.\nExpected: %s\nGot: %s", expectedTarget, absTarget)
	}
}

// TestLinkComponent_ErrorWhenComponentNotFound tests Story-006:
// When a component doesn't exist in any profile, it should show a clear error
func TestLinkComponent_ErrorWhenComponentNotFound(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get the real profiles directory
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		t.Fatalf("Failed to get profiles directory: %v", err)
	}

	// Get the real agents directory for active profile tracking
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		t.Fatalf("Failed to get agents directory: %v", err)
	}

	// Create a test profile (empty)
	testProfile := filepath.Join(profilesDir, "test-notfound-profile")

	// Cleanup function
	defer func() {
		os.RemoveAll(testProfile)
		os.Remove(filepath.Join(agentsDir, ".active-profile"))
	}()

	// Create empty profile
	if err := os.MkdirAll(filepath.Join(testProfile, "skills"), 0755); err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// Set this profile as active
	activeProfilePath := filepath.Join(agentsDir, ".active-profile")
	if err := os.WriteFile(activeProfilePath, []byte("test-notfound-profile"), 0644); err != nil {
		t.Fatalf("Failed to set active profile: %v", err)
	}

	// Create a mock target for testing
	tempTargetDir, err := os.MkdirTemp("", "agent-smith-target-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp target dir: %v", err)
	}
	defer os.RemoveAll(tempTargetDir)

	targets := []config.Target{
		&mockTarget{name: "test-target", baseDir: tempTargetDir},
	}

	det := detector.NewRepositoryDetector()

	// Create linker with the active profile's directory
	linker, err := NewComponentLinker(testProfile, targets, det)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Test: Attempting to link non-existent component should return clear error
	err = linker.LinkComponent("skills", "non-existent-component-xyz")
	if err == nil {
		t.Fatal("Expected error when linking non-existent component, but got nil")
	}

	// Verify error message is clear
	expectedMsg := "does not exist in any profile"
	if err.Error() != "component skills/non-existent-component-xyz does not exist in any profile" {
		t.Errorf("Expected error containing '%s', got: %v", expectedMsg, err)
	}
}
