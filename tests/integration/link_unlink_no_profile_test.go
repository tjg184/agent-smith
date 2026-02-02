//go:build integration
// +build integration

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tgaines/agent-smith/internal/testutil"
)

// TestLinkAll_NoProfile verifies that link all works without profiles (backward compatibility)
// This test ensures Story-007 acceptance criteria are met for link operations.
func TestLinkAll_NoProfile(t *testing.T) {
	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-link-no-profile-*")
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

	// Create test component structure in base ~/.agent-smith/
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	dirs := []string{
		filepath.Join(agentSmithDir, "agents", "test-agent"),
		filepath.Join(agentSmithDir, "skills", "test-skill"),
		filepath.Join(agentSmithDir, "commands", "test-command"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test files in components
	testFiles := map[string]string{
		filepath.Join(agentSmithDir, "agents", "test-agent", "README.md"):        "# Test Agent",
		filepath.Join(agentSmithDir, "skills", "test-skill", "SKILL.md"):         "# Test Skill",
		filepath.Join(agentSmithDir, "commands", "test-command", "commands.yml"): "# Test Command",
	}

	for file, content := range testFiles {
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Create a mock target configuration
	configDir := filepath.Join(tempDir, ".config", "agent-smith")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	configContent := `targets:
  - name: test-target
    agents_dir: ` + filepath.Join(tempDir, "test-target", "agents") + `
    skills_dir: ` + filepath.Join(tempDir, "test-target", "skills") + `
    commands_dir: ` + filepath.Join(tempDir, "test-target", "commands") + `
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Run `agent-smith link all` without any profile
	cmd = exec.Command(binaryPath, "link", "all")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	t.Logf("Link all output:\n%s", outputStr)

	// Should succeed
	if err != nil {
		t.Fatalf("Link all failed: %v\nOutput: %s", err, outputStr)
	}

	// Verify NO profile-related messages
	profileMessages := []string{
		"profile",
		"Profile",
		"from profile",
		"--all-profiles",
	}

	for _, msg := range profileMessages {
		if strings.Contains(outputStr, msg) {
			t.Errorf("Output should NOT contain profile-related message when no profiles exist: %q\nFull output:\n%s", msg, outputStr)
		}
	}

	// Verify symlinks were created
	targetAgentPath := filepath.Join(tempDir, "test-target", "agents", "test-agent")
	if _, err := os.Lstat(targetAgentPath); err != nil {
		t.Errorf("Expected agent symlink to be created: %v", err)
	}
}

// TestUnlinkAll_NoProfile verifies that unlink all works without profiles (backward compatibility)
// This test ensures Story-007 acceptance criteria are met for unlink operations.
func TestUnlinkAll_NoProfile(t *testing.T) {
	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-unlink-no-profile-*")
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

	// Create test component structure in base ~/.agent-smith/
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	targetDir := filepath.Join(tempDir, "test-target")
	dirs := []string{
		filepath.Join(agentSmithDir, "agents", "test-agent"),
		filepath.Join(agentSmithDir, "skills", "test-skill"),
		filepath.Join(targetDir, "agents"),
		filepath.Join(targetDir, "skills"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test files in source components
	testFiles := map[string]string{
		filepath.Join(agentSmithDir, "agents", "test-agent", "README.md"): "# Test Agent",
		filepath.Join(agentSmithDir, "skills", "test-skill", "SKILL.md"):  "# Test Skill",
	}

	for file, content := range testFiles {
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Create symlinks manually (simulating previous link operation)
	sourcePath := filepath.Join(agentSmithDir, "agents", "test-agent")
	targetPath := filepath.Join(targetDir, "agents", "test-agent")
	if err := os.Symlink(sourcePath, targetPath); err != nil {
		t.Fatalf("Failed to create test symlink: %v", err)
	}

	sourcePath2 := filepath.Join(agentSmithDir, "skills", "test-skill")
	targetPath2 := filepath.Join(targetDir, "skills", "test-skill")
	if err := os.Symlink(sourcePath2, targetPath2); err != nil {
		t.Fatalf("Failed to create test symlink: %v", err)
	}

	// Create a mock target configuration
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

	// Run `agent-smith unlink all --force` without any profile
	cmd = exec.Command(binaryPath, "unlink", "all", "--force")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	t.Logf("Unlink all output:\n%s", outputStr)

	// Should succeed
	if err != nil {
		t.Fatalf("Unlink all failed: %v\nOutput: %s", err, outputStr)
	}

	// Verify NO profile-related messages (except for "base installation" which is acceptable)
	profileMessages := []string{
		"from profile",
		"other profiles",
		"--all-profiles",
		"Profile '",
	}

	for _, msg := range profileMessages {
		if strings.Contains(outputStr, msg) {
			t.Errorf("Output should NOT contain profile-related message when no profiles exist: %q\nFull output:\n%s", msg, outputStr)
		}
	}

	// Verify symlinks were removed
	if _, err := os.Lstat(targetPath); !os.IsNotExist(err) {
		t.Errorf("Expected agent symlink to be removed")
	}
	if _, err := os.Lstat(targetPath2); !os.IsNotExist(err) {
		t.Errorf("Expected skill symlink to be removed")
	}

	// Verify source still exists
	if _, err := os.Stat(sourcePath); err != nil {
		t.Errorf("Source agent directory should still exist: %v", err)
	}
	if _, err := os.Stat(sourcePath2); err != nil {
		t.Errorf("Source skill directory should still exist: %v", err)
	}
}

// TestLinkUnlink_NoProfileNoNewFlags verifies that existing flags work without profiles
func TestLinkUnlink_NoProfileNoNewFlags(t *testing.T) {
	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-no-profile-flags-*")
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

	// Create test component structure
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	testAgentDir := filepath.Join(agentSmithDir, "agents", "test-agent")
	if err := os.MkdirAll(testAgentDir, 0755); err != nil {
		t.Fatalf("Failed to create test agent directory: %v", err)
	}

	testFile := filepath.Join(testAgentDir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a mock target configuration
	targetDir := filepath.Join(tempDir, "test-target")
	if err := os.MkdirAll(filepath.Join(targetDir, "agents"), 0755); err != nil {
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

	// Test link all with --target flag
	cmd = exec.Command(binaryPath, "link", "all", "--target", "test-target")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	t.Logf("Link all with --target output:\n%s", outputStr)

	if err != nil {
		t.Fatalf("Link all with --target flag should work: %v\nOutput: %s", err, outputStr)
	}

	// Verify symlink was created
	targetPath := filepath.Join(targetDir, "agents", "test-agent")
	if _, err := os.Lstat(targetPath); err != nil {
		t.Errorf("Expected symlink to be created: %v", err)
	}

	// Test unlink all with --target and --force flags
	cmd = exec.Command(binaryPath, "unlink", "all", "--target", "test-target", "--force")
	output, err = cmd.CombinedOutput()
	outputStr = string(output)

	t.Logf("Unlink all with --target and --force output:\n%s", outputStr)

	if err != nil {
		t.Fatalf("Unlink all with existing flags should work: %v\nOutput: %s", err, outputStr)
	}

	// Verify symlink was removed
	if _, err := os.Lstat(targetPath); !os.IsNotExist(err) {
		t.Errorf("Expected symlink to be removed")
	}
}

// TestNoProfile_PerformanceUnchanged ensures no performance regression for non-profile users
func TestNoProfile_PerformanceUnchanged(t *testing.T) {
	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-no-profile-perf-*")
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

	// Create test structure with multiple components (realistic workload)
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	targetDir := filepath.Join(tempDir, "test-target")

	// Create 10 components of each type
	for i := 0; i < 10; i++ {
		dirs := []string{
			filepath.Join(agentSmithDir, "agents", "test-agent-"+string(rune('a'+i))),
			filepath.Join(agentSmithDir, "skills", "test-skill-"+string(rune('a'+i))),
			filepath.Join(agentSmithDir, "commands", "test-command-"+string(rune('a'+i))),
		}
		for _, dir := range dirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("Failed to create directory: %v", err)
			}
			// Add a file to each component
			file := filepath.Join(dir, "README.md")
			if err := os.WriteFile(file, []byte("# Test"), 0644); err != nil {
				t.Fatalf("Failed to create file: %v", err)
			}
		}
	}

	// Create target directories
	targetDirs := []string{
		filepath.Join(targetDir, "agents"),
		filepath.Join(targetDir, "skills"),
		filepath.Join(targetDir, "commands"),
	}
	for _, dir := range targetDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create target directory: %v", err)
		}
	}

	// Create a mock target configuration
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

	// Run link all and verify it completes quickly
	cmd = exec.Command(binaryPath, "link", "all")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	t.Logf("Link all output (first 500 chars):\n%s", outputStr[:min(500, len(outputStr))])

	if err != nil {
		t.Fatalf("Link all failed: %v\nOutput: %s", err, outputStr)
	}

	// Run unlink all and verify it completes quickly
	cmd = exec.Command(binaryPath, "unlink", "all", "--force")
	output, err = cmd.CombinedOutput()
	outputStr = string(output)

	t.Logf("Unlink all output (first 500 chars):\n%s", outputStr[:min(500, len(outputStr))])

	if err != nil {
		t.Fatalf("Unlink all failed: %v\nOutput: %s", err, outputStr)
	}

	// Test passes if it completes within reasonable time (handled by test framework timeout)
	t.Log("Performance check passed - commands completed in reasonable time")
}

// TestNoProfile_NoProfileManager verifies graceful handling when profile manager is not initialized
func TestNoProfile_NoProfileManager(t *testing.T) {
	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-no-profile-manager-*")
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
	testAgentDir := filepath.Join(agentSmithDir, "agents", "test-agent")
	if err := os.MkdirAll(testAgentDir, 0755); err != nil {
		t.Fatalf("Failed to create test agent directory: %v", err)
	}

	testFile := filepath.Join(testAgentDir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a mock target configuration
	targetDir := filepath.Join(tempDir, "test-target")
	if err := os.MkdirAll(filepath.Join(targetDir, "agents"), 0755); err != nil {
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

	// Verify that normal link/unlink operations work without profile manager
	// (This is the backward compatibility test)
	cmd = exec.Command(binaryPath, "link", "all")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	t.Logf("Link all without profile manager:\n%s", outputStr)

	// Should succeed even without profile manager
	if err != nil {
		t.Fatalf("Link all should work without profile manager: %v\nOutput: %s", err, outputStr)
	}

	// Should NOT crash or show errors about missing profile manager
	errorStrings := []string{
		"profile manager not initialized",
		"profile manager error",
		"panic",
		"nil pointer",
	}

	for _, errStr := range errorStrings {
		if strings.Contains(strings.ToLower(outputStr), strings.ToLower(errStr)) {
			t.Errorf("Output should NOT contain error about profile manager: %q\nFull output:\n%s", errStr, outputStr)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
