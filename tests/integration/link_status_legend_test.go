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

// TestLinkStatusLegend_SingleProfile verifies that the legend is displayed in a box table format
// for the single-profile view (default link status command).
func TestLinkStatusLegend_SingleProfile(t *testing.T) {
	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-legend-single-*")
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
	dirs := []string{
		filepath.Join(agentSmithDir, "agents", "test-agent"),
		filepath.Join(agentSmithDir, "skills", "test-skill"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test files
	testFiles := map[string]string{
		filepath.Join(agentSmithDir, "agents", "test-agent", "README.md"): "# Test Agent",
		filepath.Join(agentSmithDir, "skills", "test-skill", "SKILL.md"):  "# Test Skill",
	}

	for file, content := range testFiles {
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Run `agent-smith link status`
	cmd = exec.Command(binaryPath, "link", "status")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	t.Logf("Link status output:\n%s", outputStr)

	// Should succeed or gracefully handle no targets
	if err != nil {
		if !strings.Contains(outputStr, "No components found") && !strings.Contains(outputStr, "No targets detected") {
			t.Fatalf("Unexpected error from link status: %v\nOutput: %s", err, outputStr)
		}
	}

	// Verify box table structure in legend
	// Box-drawing characters for table borders
	boxChars := []string{
		"┌", // Top-left corner
		"┐", // Top-right corner
		"└", // Bottom-left corner
		"┘", // Bottom-right corner
		"├", // Left T-junction
		"┤", // Right T-junction
		"┬", // Top T-junction
		"┴", // Bottom T-junction
		"─", // Horizontal line
		"│", // Vertical line
	}

	// Check that legend contains box-drawing characters
	foundBoxChars := false
	for _, char := range boxChars {
		if strings.Contains(outputStr, char) {
			foundBoxChars = true
			break
		}
	}

	if !foundBoxChars {
		t.Errorf("Legend should contain box-drawing characters for table formatting, got:\n%s", outputStr)
	}

	// Verify legend has proper headers
	if !strings.Contains(outputStr, "Symbol") {
		t.Errorf("Legend should contain 'Symbol' header, got:\n%s", outputStr)
	}
	if !strings.Contains(outputStr, "Meaning") {
		t.Errorf("Legend should contain 'Meaning' header, got:\n%s", outputStr)
	}

	// Verify all symbols are present
	expectedSymbols := map[string]string{
		"✓": "Valid symlink",
		"◆": "Copied directory",
		"✗": "Broken link",
		"-": "Not linked",
		"?": "Unknown status",
	}

	for symbol, description := range expectedSymbols {
		if !strings.Contains(outputStr, symbol) {
			t.Errorf("Legend should contain symbol '%s', got:\n%s", symbol, outputStr)
		}
		if !strings.Contains(outputStr, description) {
			t.Errorf("Legend should contain description '%s', got:\n%s", description, outputStr)
		}
	}

	// Verify legend appears after the Legend header
	if !strings.Contains(outputStr, "--- Legend ---") {
		t.Errorf("Legend section should have header '--- Legend ---', got:\n%s", outputStr)
	}

	t.Log("✓ Single-profile legend displays as box table with all required symbols")
}

// TestLinkStatusLegend_AllProfiles verifies that the legend is displayed in a box table format
// for the all-profiles view (link status --all-profiles command).
func TestLinkStatusLegend_AllProfiles(t *testing.T) {
	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-legend-all-profiles-*")
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

	// Create test component structure in base
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	dirs := []string{
		filepath.Join(agentSmithDir, "agents", "test-agent"),
		filepath.Join(agentSmithDir, "skills", "test-skill"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test files
	testFiles := map[string]string{
		filepath.Join(agentSmithDir, "agents", "test-agent", "README.md"): "# Test Agent",
		filepath.Join(agentSmithDir, "skills", "test-skill", "SKILL.md"):  "# Test Skill",
	}

	for file, content := range testFiles {
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Create a test profile with components so --all-profiles doesn't error
	profileDir := filepath.Join(agentSmithDir, "profiles", "test-profile")
	profileDirs := []string{
		filepath.Join(profileDir, "agents", "profile-agent"),
	}

	for _, dir := range profileDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create profile directory %s: %v", dir, err)
		}
	}

	// Create profile component file
	profileFile := filepath.Join(profileDir, "agents", "profile-agent", "README.md")
	if err := os.WriteFile(profileFile, []byte("# Profile Agent"), 0644); err != nil {
		t.Fatalf("Failed to create profile test file %s: %v", profileFile, err)
	}

	// Create profile metadata
	metadataFile := filepath.Join(profileDir, ".profile-metadata.json")
	metadataContent := `{"name":"test-profile","createdAt":"2024-01-01T00:00:00Z","sourceType":"custom"}`
	if err := os.WriteFile(metadataFile, []byte(metadataContent), 0644); err != nil {
		t.Fatalf("Failed to create profile metadata %s: %v", metadataFile, err)
	}

	// Run `agent-smith link status --all-profiles`
	cmd = exec.Command(binaryPath, "link", "status", "--all-profiles")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	t.Logf("Link status --all-profiles output:\n%s", outputStr)

	// Should succeed or gracefully handle no targets
	if err != nil {
		if !strings.Contains(outputStr, "No components found") && !strings.Contains(outputStr, "No targets detected") {
			t.Fatalf("Unexpected error from link status --all-profiles: %v\nOutput: %s", err, outputStr)
		}
	}

	// Verify box table structure in legend
	boxChars := []string{
		"┌", "┐", "└", "┘", "├", "┤", "┬", "┴", "─", "│",
	}

	foundBoxChars := false
	for _, char := range boxChars {
		if strings.Contains(outputStr, char) {
			foundBoxChars = true
			break
		}
	}

	if !foundBoxChars {
		t.Errorf("Legend should contain box-drawing characters for table formatting, got:\n%s", outputStr)
	}

	// Verify legend has proper headers
	if !strings.Contains(outputStr, "Symbol") {
		t.Errorf("Legend should contain 'Symbol' header, got:\n%s", outputStr)
	}
	if !strings.Contains(outputStr, "Meaning") {
		t.Errorf("Legend should contain 'Meaning' header, got:\n%s", outputStr)
	}

	// Verify all symbols are present
	expectedSymbols := map[string]string{
		"✓": "Valid symlink",
		"◆": "Copied directory",
		"✗": "Broken link",
		"-": "Not linked",
		"?": "Unknown status",
	}

	for symbol, description := range expectedSymbols {
		if !strings.Contains(outputStr, symbol) {
			t.Errorf("Legend should contain symbol '%s', got:\n%s", symbol, outputStr)
		}
		if !strings.Contains(outputStr, description) {
			t.Errorf("Legend should contain description '%s', got:\n%s", description, outputStr)
		}
	}

	// Verify legend appears after the Legend header
	if !strings.Contains(outputStr, "--- Legend ---") {
		t.Errorf("Legend section should have header '--- Legend ---', got:\n%s", outputStr)
	}

	t.Log("✓ All-profiles legend displays as box table with all required symbols")
}

// TestLinkStatusLegend_ConsistentFormatting verifies that both single-profile and all-profiles
// views have consistent legend table formatting.
func TestLinkStatusLegend_ConsistentFormatting(t *testing.T) {
	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-legend-consistency-*")
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
	dirs := []string{
		filepath.Join(agentSmithDir, "agents", "test-agent"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test file
	testFile := filepath.Join(agentSmithDir, "agents", "test-agent", "README.md")
	if err := os.WriteFile(testFile, []byte("# Test Agent"), 0644); err != nil {
		t.Fatalf("Failed to create test file %s: %v", testFile, err)
	}

	// Create a test profile for --all-profiles command
	profileDir := filepath.Join(agentSmithDir, "profiles", "test-profile")
	profileAgentDir := filepath.Join(profileDir, "agents", "profile-agent")
	if err := os.MkdirAll(profileAgentDir, 0755); err != nil {
		t.Fatalf("Failed to create profile directory %s: %v", profileAgentDir, err)
	}

	profileFile := filepath.Join(profileAgentDir, "README.md")
	if err := os.WriteFile(profileFile, []byte("# Profile Agent"), 0644); err != nil {
		t.Fatalf("Failed to create profile test file %s: %v", profileFile, err)
	}

	metadataFile := filepath.Join(profileDir, ".profile-metadata.json")
	metadataContent := `{"name":"test-profile","createdAt":"2024-01-01T00:00:00Z","sourceType":"custom"}`
	if err := os.WriteFile(metadataFile, []byte(metadataContent), 0644); err != nil {
		t.Fatalf("Failed to create profile metadata %s: %v", metadataFile, err)
	}

	// Run both commands and compare legend formatting
	cmd1 := exec.Command(binaryPath, "link", "status")
	output1, _ := cmd1.CombinedOutput()
	outputStr1 := string(output1)

	cmd2 := exec.Command(binaryPath, "link", "status", "--all-profiles")
	output2, _ := cmd2.CombinedOutput()
	outputStr2 := string(output2)

	t.Logf("Single-profile output:\n%s\n\nAll-profiles output:\n%s", outputStr1, outputStr2)

	// Extract legend sections from both outputs
	extractLegend := func(output string) string {
		lines := strings.Split(output, "\n")
		legendStart := -1
		legendEnd := -1

		for i, line := range lines {
			if strings.Contains(line, "--- Legend ---") {
				legendStart = i
			}
			if legendStart >= 0 && strings.Contains(line, "--- Summary ---") {
				legendEnd = i
				break
			}
		}

		if legendStart >= 0 && legendEnd > legendStart {
			return strings.Join(lines[legendStart:legendEnd], "\n")
		}
		return ""
	}

	legend1 := extractLegend(outputStr1)
	legend2 := extractLegend(outputStr2)

	// Both legends should have the same symbols and structure
	if legend1 == "" {
		t.Error("Could not extract legend from single-profile output")
	}
	if legend2 == "" {
		t.Error("Could not extract legend from all-profiles output")
	}

	// Verify both contain the same symbols
	expectedSymbols := []string{"✓", "◆", "✗", "-", "?"}
	for _, symbol := range expectedSymbols {
		if !strings.Contains(legend1, symbol) {
			t.Errorf("Single-profile legend missing symbol '%s'", symbol)
		}
		if !strings.Contains(legend2, symbol) {
			t.Errorf("All-profiles legend missing symbol '%s'", symbol)
		}
	}

	t.Log("✓ Both single-profile and all-profiles views have consistent legend formatting")
}
