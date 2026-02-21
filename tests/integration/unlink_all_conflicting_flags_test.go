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

// TestUnlinkAll_ConflictingFlags_ProfileAndAllProfiles tests Story-003:
// As a user, I want to be prevented from using conflicting flags so that I don't accidentally perform unintended operations.
//
// Acceptance Criteria:
// - Cannot use `--profile` and `--all-profiles` together
// - Error message: "Cannot use both --all-profiles and --profile flags together"
// - Validation matches `link all` flag validation behavior
// - No unlinking operation was performed
func TestUnlinkAll_ConflictingFlags_ProfileAndAllProfiles(t *testing.T) {
	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-unlink-conflicting-flags-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	// Build agent-smith binary
	binaryPath := filepath.Join(tempDir, "agent-smith")
	repoRoot := filepath.Join("..", "..")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build agent-smith: %v\nOutput: %s", err, string(output))
	}

	// Create base agent-smith directory structure
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	profilesDir := filepath.Join(agentSmithDir, "profiles")

	// Create test profile with components
	testProfilePath := filepath.Join(profilesDir, "test-profile")
	testSkillDir := filepath.Join(testProfilePath, "skills", "test-skill")
	if err := os.MkdirAll(testSkillDir, 0755); err != nil {
		t.Fatalf("Failed to create test skill directory: %v", err)
	}

	// Create test file in the skill
	testFile := filepath.Join(testSkillDir, "SKILL.md")
	if err := os.WriteFile(testFile, []byte("# Test Skill"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a mock target configuration
	targetDir := filepath.Join(tempDir, "test-target")
	targetSkillsDir := filepath.Join(targetDir, "skills")
	if err := os.MkdirAll(targetSkillsDir, 0755); err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}

	configDir := filepath.Join(tempDir, ".config", "agent-smith")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	configContent := `targets:
  - name: test-target
    agents_dir: ` + filepath.Join(targetDir, "agents") + `
    skills_dir: ` + filepath.Join(targetDir, "skills") + `
    commands_dir: ` + filepath.Join(targetDir, "commands") + `
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Create a symlink in the target (simulating previous link operation)
	sourceSkillPath := filepath.Join(testProfilePath, "skills", "test-skill")
	targetSkillPath := filepath.Join(targetSkillsDir, "test-skill")
	if err := os.Symlink(sourceSkillPath, targetSkillPath); err != nil {
		t.Fatalf("Failed to create test symlink: %v", err)
	}

	// Verify symlink exists before test
	if _, err := os.Lstat(targetSkillPath); err != nil {
		t.Fatalf("Symlink should exist before test: %v", err)
	}

	// Run `agent-smith unlink all --profile test-profile --all-profiles`
	// This should fail with a clear error message
	cmd = exec.Command(binaryPath, "unlink", "all", "--profile", "test-profile", "--all-profiles")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	t.Logf("Unlink all with conflicting flags output:\n%s", outputStr)

	// Should fail (non-zero exit code)
	if err == nil {
		t.Errorf("Command should have failed when using both --profile and --all-profiles flags")
	}

	// Verify error message matches expected format
	expectedErrorMsg := "Cannot use both --all-profiles and --profile flags together"
	if !strings.Contains(outputStr, expectedErrorMsg) {
		t.Errorf("Error message should contain: %q\nActual output:\n%s", expectedErrorMsg, outputStr)
	}

	// Verify no unlinking operation was performed (symlink should still exist)
	if _, err := os.Lstat(targetSkillPath); err != nil {
		t.Errorf("Symlink should still exist after failed command (no partial unlinking should occur): %v", err)
	}

	t.Log("Successfully prevented conflicting flags and provided clear error message")
}

// TestUnlinkAll_ConflictingFlags_OrderIndependent tests that the flag validation
// works regardless of the order the flags are specified
func TestUnlinkAll_ConflictingFlags_OrderIndependent(t *testing.T) {
	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-unlink-flags-order-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	// Build agent-smith binary
	binaryPath := filepath.Join(tempDir, "agent-smith")
	repoRoot := filepath.Join("..", "..")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build agent-smith: %v\nOutput: %s", err, string(output))
	}

	// Create minimal test structure
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	profilesDir := filepath.Join(agentSmithDir, "profiles")
	testProfilePath := filepath.Join(profilesDir, "test-profile")
	skillsDir := filepath.Join(testProfilePath, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Test both flag orders
	testCases := []struct {
		name string
		args []string
	}{
		{
			name: "profile_then_all-profiles",
			args: []string{"unlink", "all", "--profile", "test-profile", "--all-profiles"},
		},
		{
			name: "all-profiles_then_profile",
			args: []string{"unlink", "all", "--all-profiles", "--profile", "test-profile"},
		},
		{
			name: "short_flag_profile",
			args: []string{"unlink", "all", "-p", "test-profile", "--all-profiles"},
		},
	}

	expectedErrorMsg := "Cannot use both --all-profiles and --profile flags together"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tc.args...)
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			t.Logf("Test case %s output:\n%s", tc.name, outputStr)

			// Should fail
			if err == nil {
				t.Errorf("Command should have failed for conflicting flags")
			}

			// Should contain expected error message
			if !strings.Contains(outputStr, expectedErrorMsg) {
				t.Errorf("Error message should contain: %q\nActual output:\n%s", expectedErrorMsg, outputStr)
			}
		})
	}
}

// TestUnlinkAll_ProfileFlag_WithoutAllProfiles tests that --profile flag works correctly
// when --all-profiles is NOT specified (should succeed)
func TestUnlinkAll_ProfileFlag_WithoutAllProfiles(t *testing.T) {
	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-unlink-profile-only-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	// Build agent-smith binary
	binaryPath := filepath.Join(tempDir, "agent-smith")
	repoRoot := filepath.Join("..", "..")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build agent-smith: %v\nOutput: %s", err, string(output))
	}

	// Create base agent-smith directory structure
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	profilesDir := filepath.Join(agentSmithDir, "profiles")

	// Create test profile with components
	testProfilePath := filepath.Join(profilesDir, "test-profile")
	testSkillDir := filepath.Join(testProfilePath, "skills", "test-skill")
	if err := os.MkdirAll(testSkillDir, 0755); err != nil {
		t.Fatalf("Failed to create test skill directory: %v", err)
	}

	// Create test file in the skill
	testFile := filepath.Join(testSkillDir, "SKILL.md")
	if err := os.WriteFile(testFile, []byte("# Test Skill"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a mock target configuration
	targetDir := filepath.Join(tempDir, "test-target")
	targetSkillsDir := filepath.Join(targetDir, "skills")
	if err := os.MkdirAll(targetSkillsDir, 0755); err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}

	configDir := filepath.Join(tempDir, ".config", "agent-smith")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	configContent := `targets:
  - name: test-target
    agents_dir: ` + filepath.Join(targetDir, "agents") + `
    skills_dir: ` + filepath.Join(targetDir, "skills") + `
    commands_dir: ` + filepath.Join(targetDir, "commands") + `
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Create a symlink in the target (simulating previous link operation)
	sourceSkillPath := filepath.Join(testProfilePath, "skills", "test-skill")
	targetSkillPath := filepath.Join(targetSkillsDir, "test-skill")
	if err := os.Symlink(sourceSkillPath, targetSkillPath); err != nil {
		t.Fatalf("Failed to create test symlink: %v", err)
	}

	// Verify symlink exists before test
	if _, err := os.Lstat(targetSkillPath); err != nil {
		t.Fatalf("Symlink should exist before test: %v", err)
	}

	// Run `agent-smith unlink all --profile test-profile --force`
	// This should succeed (--profile without --all-profiles is valid)
	cmd = exec.Command(binaryPath, "unlink", "all", "--profile", "test-profile", "--force")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	t.Logf("Unlink all with --profile only output:\n%s", outputStr)

	// Should succeed
	if err != nil {
		t.Errorf("Command should succeed when using --profile without --all-profiles: %v\nOutput: %s", err, outputStr)
	}

	// Should NOT contain the conflicting flags error
	conflictingFlagsError := "Cannot use both --all-profiles and --profile flags together"
	if strings.Contains(outputStr, conflictingFlagsError) {
		t.Errorf("Should not show conflicting flags error when only --profile is used: %s", outputStr)
	}

	// Note: The actual unlinking behavior is tested elsewhere. This test focuses on
	// validating that the --profile flag is accepted without the conflicting flags error.

	t.Log("Successfully used --profile flag without --all-profiles")
}

// TestUnlinkAll_AllProfilesFlag_WithoutProfile tests that --all-profiles flag works correctly
// when --profile is NOT specified (should succeed)
func TestUnlinkAll_AllProfilesFlag_WithoutProfile(t *testing.T) {
	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-unlink-allprofiles-only-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	// Build agent-smith binary
	binaryPath := filepath.Join(tempDir, "agent-smith")
	repoRoot := filepath.Join("..", "..")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build agent-smith: %v\nOutput: %s", err, string(output))
	}

	// Create base agent-smith directory structure
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	profilesDir := filepath.Join(agentSmithDir, "profiles")

	// Create test profile with components
	testProfilePath := filepath.Join(profilesDir, "test-profile")
	testSkillDir := filepath.Join(testProfilePath, "skills", "test-skill")
	if err := os.MkdirAll(testSkillDir, 0755); err != nil {
		t.Fatalf("Failed to create test skill directory: %v", err)
	}

	// Create test file in the skill
	testFile := filepath.Join(testSkillDir, "SKILL.md")
	if err := os.WriteFile(testFile, []byte("# Test Skill"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a mock target configuration
	targetDir := filepath.Join(tempDir, "test-target")
	targetSkillsDir := filepath.Join(targetDir, "skills")
	if err := os.MkdirAll(targetSkillsDir, 0755); err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}

	configDir := filepath.Join(tempDir, ".config", "agent-smith")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	configContent := `targets:
  - name: test-target
    agents_dir: ` + filepath.Join(targetDir, "agents") + `
    skills_dir: ` + filepath.Join(targetDir, "skills") + `
    commands_dir: ` + filepath.Join(targetDir, "commands") + `
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Create a symlink in the target (simulating previous link operation)
	sourceSkillPath := filepath.Join(testProfilePath, "skills", "test-skill")
	targetSkillPath := filepath.Join(targetSkillsDir, "test-skill")
	if err := os.Symlink(sourceSkillPath, targetSkillPath); err != nil {
		t.Fatalf("Failed to create test symlink: %v", err)
	}

	// Verify symlink exists before test
	if _, err := os.Lstat(targetSkillPath); err != nil {
		t.Fatalf("Symlink should exist before test: %v", err)
	}

	// Run `agent-smith unlink all --all-profiles --force`
	// This should succeed (--all-profiles without --profile is valid)
	cmd = exec.Command(binaryPath, "unlink", "all", "--all-profiles", "--force")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	t.Logf("Unlink all with --all-profiles only output:\n%s", outputStr)

	// Should succeed
	if err != nil {
		t.Errorf("Command should succeed when using --all-profiles without --profile: %v\nOutput: %s", err, outputStr)
	}

	// Should NOT contain the conflicting flags error
	conflictingFlagsError := "Cannot use both --all-profiles and --profile flags together"
	if strings.Contains(outputStr, conflictingFlagsError) {
		t.Errorf("Should not show conflicting flags error when only --all-profiles is used: %s", outputStr)
	}

	t.Log("Successfully used --all-profiles flag without --profile")
}
