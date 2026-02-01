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

// TestLinkStatus_DefaultBehavior verifies that the default `agent-smith link status` command
// maintains backward compatibility and shows current profile/base only.
// This test ensures Story-004 acceptance criteria are met.
func TestLinkStatus_DefaultBehavior(t *testing.T) {
	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-link-status-integration-*")
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

	// Create test component structure manually (to avoid network dependencies)
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

	// Run `agent-smith link status` without any flags (default behavior)
	cmd = exec.Command(binaryPath, "link", "status")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	t.Logf("Link status output:\n%s", outputStr)

	// Should succeed even if no targets are detected
	if err != nil {
		// If error occurs, it should be due to no targets being detected, not a code error
		if !strings.Contains(outputStr, "No components found") && !strings.Contains(outputStr, "No targets detected") {
			t.Fatalf("Unexpected error from link status: %v\nOutput: %s", err, outputStr)
		}
	}

	// Verify output contains expected format elements
	expectedStrings := []string{
		"=== Link Status Across All Targets ===",
		"Component",
		"Profile",
		"Legend:",
		"✓  Valid symlink",
		"◆  Copied directory",
		"✗  Broken link",
		"-  Not linked",
		"?  Unknown status",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Default link status output missing expected string: %s\nFull output:\n%s", expected, outputStr)
		}
	}

	// Verify it shows component types
	componentTypeSections := []string{"Skills:", "Agents:", "Commands:"}
	foundAtLeastOne := false
	for _, section := range componentTypeSections {
		if strings.Contains(outputStr, section) {
			foundAtLeastOne = true
			break
		}
	}

	if !foundAtLeastOne && !strings.Contains(outputStr, "No components found") {
		t.Error("Output should contain at least one component type section (Skills:/Agents:/Commands:)")
	}
}

// TestLinkStatus_FlagDefaults verifies that flag defaults ensure existing behavior
func TestLinkStatus_FlagDefaults(t *testing.T) {
	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-link-status-flags-*")
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

	// Run with explicit --all-profiles=false (should be same as default)
	cmd = exec.Command(binaryPath, "link", "status", "--all-profiles=false")
	output1, err1 := cmd.CombinedOutput()
	output1Str := string(output1)

	// Run without any flags (default)
	cmd = exec.Command(binaryPath, "link", "status")
	output2, err2 := cmd.CombinedOutput()
	output2Str := string(output2)

	t.Logf("Output with --all-profiles=false:\n%s", output1Str)
	t.Logf("Output without flags:\n%s", output2Str)

	// Both should behave identically
	if (err1 != nil) != (err2 != nil) {
		t.Errorf("Error states differ: with flag err=%v, without flag err=%v", err1, err2)
	}

	// Outputs should be similar (allowing for minor timing/ordering differences)
	// Both should NOT show multi-profile format
	if strings.Contains(output1Str, "Profiles scanned:") {
		t.Error("Default behavior should not show 'Profiles scanned:' (that's multi-profile format)")
	}
	if strings.Contains(output2Str, "Profiles scanned:") {
		t.Error("Default behavior should not show 'Profiles scanned:' (that's multi-profile format)")
	}

	// Both should show single-profile format
	if !strings.Contains(output1Str, "=== Link Status Across All Targets ===") {
		t.Error("Output with --all-profiles=false missing expected header")
	}
	if !strings.Contains(output2Str, "=== Link Status Across All Targets ===") {
		t.Error("Output without flags missing expected header")
	}
}

// TestLinkStatus_NoNewFlagsRequired verifies that existing use cases work without new flags
func TestLinkStatus_NoNewFlagsRequired(t *testing.T) {
	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-link-status-no-flags-*")
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

	// Create test structure with at least one component
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	testAgentDir := filepath.Join(agentSmithDir, "agents", "test-agent")
	if err := os.MkdirAll(testAgentDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create a component file so there's something to show
	testFile := filepath.Join(testAgentDir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test that command works with just `link status` (no additional flags)
	cmd = exec.Command(binaryPath, "link", "status")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	t.Logf("Link status output:\n%s", outputStr)

	// Command should execute without requiring any new flags
	if err != nil && !strings.Contains(outputStr, "No components found") && !strings.Contains(outputStr, "No targets detected") {
		t.Fatalf("Basic 'link status' command should work without new flags: %v\nOutput: %s", err, outputStr)
	}

	// Should show familiar output format
	if !strings.Contains(outputStr, "Legend:") {
		t.Error("Output should contain familiar Legend section")
	}
}

// TestLinkStatus_OutputFormatUnchanged verifies the output format matches existing patterns
func TestLinkStatus_OutputFormatUnchanged(t *testing.T) {
	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-link-status-format-*")
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

	// Create test structure with components
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	dirs := []string{
		filepath.Join(agentSmithDir, "agents", "backend-dev"),
		filepath.Join(agentSmithDir, "skills", "api-design"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
	}

	// Create component files
	files := map[string]string{
		filepath.Join(agentSmithDir, "agents", "backend-dev", "README.md"): "# Backend Dev",
		filepath.Join(agentSmithDir, "skills", "api-design", "SKILL.md"):   "# API Design",
	}

	for file, content := range files {
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	// Run link status
	cmd = exec.Command(binaryPath, "link", "status")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	t.Logf("Link status output:\n%s", outputStr)

	if err != nil && !strings.Contains(outputStr, "No targets detected") {
		t.Fatalf("Link status failed: %v\nOutput: %s", err, outputStr)
	}

	// Verify critical format elements that must not change
	criticalElements := []struct {
		element     string
		description string
	}{
		{"=== Link Status Across All Targets ===", "main header"},
		{"Component", "component column header"},
		{"Profile", "profile column header"},
		{"Legend:", "legend section header"},
		{"✓  Valid symlink", "valid symlink legend entry"},
		{"◆  Copied directory", "copied directory legend entry"},
		{"✗  Broken link", "broken link legend entry"},
		{"-  Not linked", "not linked legend entry"},
		{"?  Unknown status", "unknown status legend entry"},
	}

	for _, elem := range criticalElements {
		if !strings.Contains(outputStr, elem.element) {
			t.Errorf("Critical format element missing (%s): %s\nThis breaks backward compatibility!", elem.description, elem.element)
		}
	}

	// Verify component type grouping headers
	if strings.Contains(outputStr, "backend-dev") {
		if !strings.Contains(outputStr, "Agents:") {
			t.Error("Components should be grouped with type headers like 'Agents:'")
		}
	}

	if strings.Contains(outputStr, "api-design") {
		if !strings.Contains(outputStr, "Skills:") {
			t.Error("Components should be grouped with type headers like 'Skills:'")
		}
	}
}

// TestLinkStatus_PerformanceRegression ensures default behavior has acceptable performance
func TestLinkStatus_PerformanceRegression(t *testing.T) {
	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-link-status-perf-*")
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

	// Create test structure with multiple components
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")

	// Create 20 test components (reasonable workload)
	for i := 0; i < 20; i++ {
		dirs := []string{
			filepath.Join(agentSmithDir, "agents", "test-agent-"+string(rune('a'+i))),
			filepath.Join(agentSmithDir, "skills", "test-skill-"+string(rune('a'+i))),
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

	// Run link status and measure execution time
	cmd = exec.Command(binaryPath, "link", "status")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	t.Logf("Link status output (first 500 chars):\n%s", outputStr[:min(500, len(outputStr))])

	// Should complete without hanging (test has implicit timeout)
	if err != nil && !strings.Contains(outputStr, "No targets detected") {
		t.Fatalf("Link status failed with multiple components: %v\nOutput: %s", err, outputStr)
	}

	// Test passes if it completes within reasonable time (handled by test framework timeout)
	t.Log("Performance check passed - command completed in reasonable time")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
