package linker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tgaines/agent-smith/internal/detector"
	"github.com/tgaines/agent-smith/pkg/config"
	"github.com/tgaines/agent-smith/pkg/paths"
)

// TestProfileCollisionHandling_Integration is an integration test that verifies
// the profile collision handling feature works end-to-end
func TestProfileCollisionHandling_Integration(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get the real agents directory
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		t.Fatalf("Failed to get agents directory: %v", err)
	}

	// Get the real profiles directory
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		t.Fatalf("Failed to get profiles directory: %v", err)
	}

	// Create test profiles with the same component
	testProfile1 := filepath.Join(profilesDir, "test-profile-collision-1")
	testProfile2 := filepath.Join(profilesDir, "test-profile-collision-2")

	// Cleanup function
	defer func() {
		os.RemoveAll(testProfile1)
		os.RemoveAll(testProfile2)
	}()

	// Create test component in profile1
	profile1SkillDir := filepath.Join(testProfile1, "skills", "collision-test-skill")
	if err := os.MkdirAll(profile1SkillDir, 0755); err != nil {
		t.Fatalf("Failed to create profile1 skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(profile1SkillDir, "SKILL.md"), []byte("# Test Skill from Profile 1"), 0644); err != nil {
		t.Fatalf("Failed to create profile1 skill file: %v", err)
	}

	// Create lock file for profile1 with source URL
	lockFile1Content := `{
  "version": 4,
  "skills": {
    "https://github.com/test/repo1": {
      "collision-test-skill": {
        "source": "test-source-1",
        "sourceType": "git",
        "sourceUrl": "https://github.com/test/repo1",
        "commitHash": "abc123",
        "installedAt": "2024-01-01T00:00:00Z",
        "updatedAt": "2024-01-01T00:00:00Z",
        "version": 3
      }
    }
  }
}`
	lockFile1Path := filepath.Join(testProfile1, ".component-lock.json")
	if err := os.WriteFile(lockFile1Path, []byte(lockFile1Content), 0644); err != nil {
		t.Fatalf("Failed to create profile1 lock file: %v", err)
	}

	// Create test component in profile2
	profile2SkillDir := filepath.Join(testProfile2, "skills", "collision-test-skill")
	if err := os.MkdirAll(profile2SkillDir, 0755); err != nil {
		t.Fatalf("Failed to create profile2 skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(profile2SkillDir, "SKILL.md"), []byte("# Test Skill from Profile 2"), 0644); err != nil {
		t.Fatalf("Failed to create profile2 skill file: %v", err)
	}

	// Create lock file for profile2 with source URL
	lockFile2Content := `{
  "version": 4,
  "skills": {
    "https://github.com/test/repo2": {
      "collision-test-skill": {
        "source": "test-source-2",
        "sourceType": "git",
        "sourceUrl": "https://github.com/test/repo2",
        "commitHash": "def456",
        "installedAt": "2024-01-01T00:00:00Z",
        "updatedAt": "2024-01-01T00:00:00Z",
        "version": 3
      }
    }
  }
}`
	lockFile2Path := filepath.Join(testProfile2, ".component-lock.json")
	if err := os.WriteFile(lockFile2Path, []byte(lockFile2Content), 0644); err != nil {
		t.Fatalf("Failed to create profile2 lock file: %v", err)
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
	linker, err := NewComponentLinker(agentsDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Test 1: Search for component in multiple profiles
	matches, err := linker.searchComponentInProfiles("skills", "collision-test-skill")
	if err != nil {
		t.Fatalf("searchComponentInProfiles failed: %v", err)
	}

	// Verify we found exactly two matches
	if len(matches) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(matches))
	}

	// Verify both profiles are in the results
	foundProfile1 := false
	foundProfile2 := false
	var profile1SourceUrl, profile2SourceUrl string

	for _, match := range matches {
		if match.ProfileName == "test-profile-collision-1" {
			foundProfile1 = true
			profile1SourceUrl = match.SourceUrl
		}
		if match.ProfileName == "test-profile-collision-2" {
			foundProfile2 = true
			profile2SourceUrl = match.SourceUrl
		}
	}

	if !foundProfile1 || !foundProfile2 {
		t.Error("Expected to find both test profiles in matches")
	}

	// Verify source URLs are loaded correctly
	if profile1SourceUrl != "https://github.com/test/repo1" {
		t.Errorf("Expected profile1 source URL 'https://github.com/test/repo1', got '%s'", profile1SourceUrl)
	}
	if profile2SourceUrl != "https://github.com/test/repo2" {
		t.Errorf("Expected profile2 source URL 'https://github.com/test/repo2', got '%s'", profile2SourceUrl)
	}

	// Test 2: Verify promptProfileSelection exists and has the right signature
	// (We can't test the interactive prompt without mocking stdin, but we can verify it exists)
	_, _, err = linker.promptProfileSelection("skills", "collision-test-skill", matches)
	// We expect an error because we can't provide interactive input in tests
	if err == nil {
		t.Log("promptProfileSelection returned without error (unexpected in non-interactive test)")
	}
}

// TestSearchComponentInProfiles_NoMatches tests searching for a non-existent component
func TestSearchComponentInProfiles_NoMatches(t *testing.T) {
	// Get the real agents directory
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		t.Fatalf("Failed to get agents directory: %v", err)
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
	linker, err := NewComponentLinker(agentsDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Search for a non-existent component
	matches, err := linker.searchComponentInProfiles("skills", "this-component-definitely-does-not-exist-12345")
	if err != nil {
		t.Fatalf("searchComponentInProfiles failed: %v", err)
	}

	// Verify we found no matches
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches, got %d", len(matches))
	}
}

// TestLinkComponent_ComponentNotInAnyProfile tests error when component doesn't exist
func TestLinkComponent_ComponentNotInAnyProfile(t *testing.T) {
	// Get the real agents directory
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		t.Fatalf("Failed to get agents directory: %v", err)
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
	linker, err := NewComponentLinker(agentsDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Try to link a non-existent component
	err = linker.LinkComponent("agents", "this-agent-definitely-does-not-exist-12345")
	if err == nil {
		t.Fatal("Expected error when linking non-existent component")
	}

	// Verify error message mentions the component doesn't exist
	expectedError := "does not exist in any profile"
	if err.Error() != "component agents/this-agent-definitely-does-not-exist-12345 does not exist in any profile" {
		t.Errorf("Expected error containing '%s', got: %v", expectedError, err)
	}
}

// TestProfileMatch_ActiveFlag tests the ProfileMatch struct correctly tracks active status
func TestProfileMatch_ActiveFlag(t *testing.T) {
	// Test the ProfileMatch struct
	match1 := ProfileMatch{
		ProfileName: "profile1",
		ProfilePath: "/path/to/profile1",
		IsActive:    true,
		SourceUrl:   "https://github.com/user/repo1",
	}

	match2 := ProfileMatch{
		ProfileName: "profile2",
		ProfilePath: "/path/to/profile2",
		IsActive:    false,
		SourceUrl:   "https://github.com/user/repo2",
	}

	if !match1.IsActive {
		t.Error("match1 should be marked as active")
	}

	if match2.IsActive {
		t.Error("match2 should not be marked as active")
	}

	if match1.ProfileName != "profile1" {
		t.Errorf("Expected ProfileName 'profile1', got '%s'", match1.ProfileName)
	}

	if match1.SourceUrl != "https://github.com/user/repo1" {
		t.Errorf("Expected SourceUrl 'https://github.com/user/repo1', got '%s'", match1.SourceUrl)
	}

	if match2.SourceUrl != "https://github.com/user/repo2" {
		t.Errorf("Expected SourceUrl 'https://github.com/user/repo2', got '%s'", match2.SourceUrl)
	}
}
