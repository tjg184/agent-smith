package linker

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tjg184/agent-smith/internal/detector"
	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/pkg/config"
)

// TestShowLinkStatus_DefaultBehavior verifies that ShowLinkStatus() maintains backward compatibility
// This test ensures Story-004: Default link status command behavior remains unchanged
func TestShowLinkStatus_DefaultBehavior(t *testing.T) {
	// Create test environment
	tempDir, err := os.MkdirTemp("", "agent-smith-link-status-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup source directory structure
	sourceDir := filepath.Join(tempDir, "source")
	target1Dir := filepath.Join(tempDir, "target1")
	target2Dir := filepath.Join(tempDir, "target2")

	dirs := []string{
		filepath.Join(sourceDir, "agents", "backend-dev"),
		filepath.Join(sourceDir, "skills", "api-design"),
		filepath.Join(sourceDir, "commands", "docker-helper"),
		filepath.Join(target1Dir, "agents"),
		filepath.Join(target1Dir, "skills"),
		filepath.Join(target1Dir, "commands"),
		filepath.Join(target2Dir, "agents"),
		filepath.Join(target2Dir, "skills"),
		filepath.Join(target2Dir, "commands"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test files in source components
	testFiles := map[string]string{
		filepath.Join(sourceDir, "agents", "backend-dev", "README.md"):        "# Backend Dev Agent",
		filepath.Join(sourceDir, "skills", "api-design", "SKILL.md"):          "# API Design Skill",
		filepath.Join(sourceDir, "commands", "docker-helper", "commands.yml"): "# Docker Helper Commands",
	}

	for file, content := range testFiles {
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Create symlinks from target1 to source
	symlinks := []struct {
		source string
		target string
	}{
		{
			source: filepath.Join(sourceDir, "agents", "backend-dev"),
			target: filepath.Join(target1Dir, "agents", "backend-dev"),
		},
		{
			source: filepath.Join(sourceDir, "skills", "api-design"),
			target: filepath.Join(target1Dir, "skills", "api-design"),
		},
	}

	for _, link := range symlinks {
		if err := os.Symlink(link.source, link.target); err != nil {
			t.Fatalf("Failed to create symlink from %s to %s: %v", link.target, link.source, err)
		}
	}

	// Create mock targets
	targets := []config.Target{
		&mockTarget{name: "TARGET1", baseDir: target1Dir},
		&mockTarget{name: "TARGET2", baseDir: target2Dir},
	}

	// Create detector
	det := detector.NewRepositoryDetector()

	// Create ComponentLinker WITHOUT ProfileManager (backward compatibility)
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create ComponentLinker: %v", err)
	}

	// Create a buffer to capture output
	var buf bytes.Buffer
	linker.SetFormatter(formatter.NewWithWriter(&buf))

	// Execute ShowLinkStatus
	err = linker.ShowLinkStatus(false)
	if err != nil {
		t.Fatalf("ShowLinkStatus() returned error: %v", err)
	}

	// Get output
	output := buf.String()

	// Verify output contains expected elements
	expectedStrings := []string{
		"=== Link Status Across All Targets ===",
		"Component",
		"Profile",
		"TARGET1",
		"TARGET2",
		"Skills:",
		"api-design",
		"Agents:",
		"backend-dev",
		"Commands:",
		"docker-helper",
		"--- Legend ---",
		"✓", "Valid symlink",
		"◆", "Copied directory",
		"✗", "Broken link",
		"-", "Not linked",
		"?", "Unknown status",
		"--- Summary ---",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Output missing expected string: %s\nOutput:\n%s", expected, output)
		}
	}

	// Verify that linked components show ✓ symbol
	lines := strings.Split(output, "\n")
	var foundBackendDev bool
	var foundApiDesign bool
	var foundDockerHelper bool

	for _, line := range lines {
		if strings.Contains(line, "backend-dev") {
			foundBackendDev = true
			// Should have ✓ for TARGET1
			if !strings.Contains(line, "✓") {
				t.Errorf("backend-dev should show ✓ for linked target, got: %s", line)
			}
		}
		if strings.Contains(line, "api-design") {
			foundApiDesign = true
			// Should have ✓ for TARGET1
			if !strings.Contains(line, "✓") {
				t.Errorf("api-design should show ✓ for linked target, got: %s", line)
			}
		}
		if strings.Contains(line, "docker-helper") {
			foundDockerHelper = true
			// Should have - for both targets (not linked)
			if !strings.Contains(line, "-") {
				t.Errorf("docker-helper should show - for unlinked targets, got: %s", line)
			}
		}
	}

	if !foundBackendDev {
		t.Error("Output should contain backend-dev component")
	}
	if !foundApiDesign {
		t.Error("Output should contain api-design component")
	}
	if !foundDockerHelper {
		t.Error("Output should contain docker-helper component")
	}
}

// TestShowLinkStatus_WithoutComponents verifies graceful handling when no components exist
func TestShowLinkStatus_WithoutComponents(t *testing.T) {
	// Create empty test environment
	tempDir, err := os.MkdirTemp("", "agent-smith-link-status-empty-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sourceDir := filepath.Join(tempDir, "source")
	targetDir := filepath.Join(tempDir, "target")

	// Create empty directories
	dirs := []string{
		filepath.Join(sourceDir, "agents"),
		filepath.Join(sourceDir, "skills"),
		filepath.Join(sourceDir, "commands"),
		filepath.Join(targetDir, "agents"),
		filepath.Join(targetDir, "skills"),
		filepath.Join(targetDir, "commands"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create mock target
	targets := []config.Target{
		&mockTarget{name: "TARGET1", baseDir: targetDir},
	}

	// Create detector
	det := detector.NewRepositoryDetector()

	// Create ComponentLinker
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create ComponentLinker: %v", err)
	}

	// Create a buffer to capture output
	var buf bytes.Buffer
	linker.SetFormatter(formatter.NewWithWriter(&buf))

	// Execute ShowLinkStatus with linkedOnly=false (default behavior)
	err = linker.ShowLinkStatus(false)
	if err != nil {
		t.Fatalf("ShowLinkStatus() returned error: %v", err)
	}

	// Get output
	output := buf.String()

	// Should display "No components found" message
	if !strings.Contains(output, "No components found") {
		t.Errorf("Expected 'No components found' message, got: %s", output)
	}
}

// TestShowLinkStatus_ProfileDetection verifies that profile information is displayed correctly
func TestShowLinkStatus_ProfileDetection(t *testing.T) {
	// Create test environment with base and profile components
	tempDir, err := os.MkdirTemp("", "agent-smith-link-status-profile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup source directory with both base and profile components
	baseDir := filepath.Join(tempDir, "base")
	profileDir := filepath.Join(tempDir, "profiles", "work-profile")
	targetDir := filepath.Join(tempDir, "target")

	dirs := []string{
		filepath.Join(baseDir, "agents", "base-agent"),
		filepath.Join(profileDir, "agents", "profile-agent"),
		filepath.Join(targetDir, "agents"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test files
	testFiles := map[string]string{
		filepath.Join(baseDir, "agents", "base-agent", "README.md"):       "# Base Agent",
		filepath.Join(profileDir, "agents", "profile-agent", "README.md"): "# Profile Agent",
	}

	for file, content := range testFiles {
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Create symlink from target to base (only base component is linked)
	sourceLink := filepath.Join(baseDir, "agents", "base-agent")
	targetLink := filepath.Join(targetDir, "agents", "base-agent")
	if err := os.Symlink(sourceLink, targetLink); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Create mock target
	targets := []config.Target{
		&mockTarget{name: "TARGET1", baseDir: targetDir},
	}

	// Create detector
	det := detector.NewRepositoryDetector()

	// Create ComponentLinker pointing to base directory
	linker, err := NewComponentLinker(baseDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create ComponentLinker: %v", err)
	}

	// Create a buffer to capture output
	var buf bytes.Buffer
	linker.SetFormatter(formatter.NewWithWriter(&buf))

	// Execute ShowLinkStatus
	err = linker.ShowLinkStatus(false)
	if err != nil {
		t.Fatalf("ShowLinkStatus() returned error: %v", err)
	}

	// Get output
	output := buf.String()

	// Verify output contains base-agent
	if !strings.Contains(output, "base-agent") {
		t.Errorf("Output should contain base-agent component: %s", output)
	}

	// Verify output does NOT contain profile-agent (it's in a different directory)
	if strings.Contains(output, "profile-agent") {
		t.Errorf("Output should not contain profile-agent (outside source directory): %s", output)
	}

	// Verify Profile column exists
	if !strings.Contains(output, "Profile") {
		t.Errorf("Output should contain Profile column header: %s", output)
	}
}

// TestShowLinkStatus_BackwardCompatibility verifies that the method works without ProfileManager
func TestShowLinkStatus_BackwardCompatibility(t *testing.T) {
	// Create test environment
	tempDir, err := os.MkdirTemp("", "agent-smith-link-status-compat-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sourceDir := filepath.Join(tempDir, "source")
	targetDir := filepath.Join(tempDir, "target")

	// Create minimal directory structure
	dirs := []string{
		filepath.Join(sourceDir, "agents", "test-agent"),
		filepath.Join(targetDir, "agents"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test file
	testFile := filepath.Join(sourceDir, "agents", "test-agent", "README.md")
	if err := os.WriteFile(testFile, []byte("# Test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create mock target
	targets := []config.Target{
		&mockTarget{name: "TARGET1", baseDir: targetDir},
	}

	// Create detector
	det := detector.NewRepositoryDetector()

	// Test 1: Create ComponentLinker with nil ProfileManager
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create ComponentLinker with nil ProfileManager: %v", err)
	}

	// Verify ProfileManager is nil
	if linker.profileManager != nil {
		t.Error("ProfileManager should be nil for backward compatibility")
	}

	// Test 2: ShowLinkStatus should work without ProfileManager
	// Create a buffer to capture output
	var buf bytes.Buffer
	linker.SetFormatter(formatter.NewWithWriter(&buf))

	err = linker.ShowLinkStatus(false)
	if err != nil {
		t.Fatalf("ShowLinkStatus() should work without ProfileManager: %v", err)
	}

	// Get output
	output := buf.String()

	// Verify basic output structure
	if !strings.Contains(output, "test-agent") {
		t.Errorf("Output should contain test-agent: %s", output)
	}

	if !strings.Contains(output, "--- Legend ---") {
		t.Errorf("Output should contain Legend section: %s", output)
	}
}

// TestShowLinkStatus_LinkedOnly verifies that linkedOnly=true filters out unlinked components
func TestShowLinkStatus_LinkedOnly(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agent-smith-linked-only-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sourceDir := filepath.Join(tempDir, "source")
	targetDir := filepath.Join(tempDir, "target")

	dirs := []string{
		filepath.Join(sourceDir, "agents", "linked-agent"),
		filepath.Join(sourceDir, "agents", "unlinked-agent"),
		filepath.Join(sourceDir, "skills", "linked-skill"),
		filepath.Join(sourceDir, "skills", "unlinked-skill"),
		filepath.Join(targetDir, "agents"),
		filepath.Join(targetDir, "skills"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	testFiles := map[string]string{
		filepath.Join(sourceDir, "agents", "linked-agent", "README.md"):   "# Linked Agent",
		filepath.Join(sourceDir, "agents", "unlinked-agent", "README.md"): "# Unlinked Agent",
		filepath.Join(sourceDir, "skills", "linked-skill", "SKILL.md"):    "# Linked Skill",
		filepath.Join(sourceDir, "skills", "unlinked-skill", "SKILL.md"):  "# Unlinked Skill",
	}

	for file, content := range testFiles {
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	symlinks := []struct{ source, target string }{
		{filepath.Join(sourceDir, "agents", "linked-agent"), filepath.Join(targetDir, "agents", "linked-agent")},
		{filepath.Join(sourceDir, "skills", "linked-skill"), filepath.Join(targetDir, "skills", "linked-skill")},
	}

	for _, link := range symlinks {
		if err := os.Symlink(link.source, link.target); err != nil {
			t.Fatalf("Failed to create symlink from %s to %s: %v", link.target, link.source, err)
		}
	}

	targets := []config.Target{&mockTarget{name: "TARGET", baseDir: targetDir}}
	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create ComponentLinker: %v", err)
	}

	var bufAll bytes.Buffer
	linker.SetFormatter(formatter.NewWithWriter(&bufAll))
	if err := linker.ShowLinkStatus(false); err != nil {
		t.Fatalf("ShowLinkStatus(false) returned error: %v", err)
	}

	outputAll := bufAll.String()
	for _, name := range []string{"linked-agent", "unlinked-agent", "linked-skill", "unlinked-skill"} {
		if !strings.Contains(outputAll, name) {
			t.Errorf("Output should contain %s", name)
		}
	}

	var bufLinked bytes.Buffer
	linker.SetFormatter(formatter.NewWithWriter(&bufLinked))
	if err := linker.ShowLinkStatus(true); err != nil {
		t.Fatalf("ShowLinkStatus(true) returned error: %v", err)
	}

	outputLinked := bufLinked.String()
	if !strings.Contains(outputLinked, "linked-agent") {
		t.Error("Output with linkedOnly=true should contain linked-agent")
	}
	if strings.Contains(outputLinked, "unlinked-agent") {
		t.Error("Output with linkedOnly=true should NOT contain unlinked-agent")
	}
	if !strings.Contains(outputLinked, "linked-skill") {
		t.Error("Output with linkedOnly=true should contain linked-skill")
	}
	if strings.Contains(outputLinked, "unlinked-skill") {
		t.Error("Output with linkedOnly=true should NOT contain unlinked-skill")
	}
	if !strings.Contains(outputLinked, "Summary") {
		t.Error("Output should contain summary section")
	}
}

// TestShowLinkStatus_LinkedOnlyAllUnlinked verifies behavior when all components are unlinked
func TestShowLinkStatus_LinkedOnlyAllUnlinked(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agent-smith-linked-only-none-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sourceDir := filepath.Join(tempDir, "source")
	targetDir := filepath.Join(tempDir, "target")

	for _, dir := range []string{
		filepath.Join(sourceDir, "skills", "unlinked-skill"),
		filepath.Join(targetDir, "skills"),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	if err := os.WriteFile(
		filepath.Join(sourceDir, "skills", "unlinked-skill", "SKILL.md"),
		[]byte("# Unlinked Skill"), 0644,
	); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	targets := []config.Target{&mockTarget{name: "TARGET", baseDir: targetDir}}
	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create ComponentLinker: %v", err)
	}

	var buf bytes.Buffer
	linker.SetFormatter(formatter.NewWithWriter(&buf))
	if err := linker.ShowLinkStatus(true); err != nil {
		t.Fatalf("ShowLinkStatus(true) returned error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "unlinked-skill") {
		t.Error("Output with linkedOnly=true should NOT contain unlinked-skill when all are unlinked")
	}
	for _, expected := range []string{"Link Status", "Legend", "Summary"} {
		if !strings.Contains(output, expected) {
			t.Errorf("Output should contain %s", expected)
		}
	}
}

// TestShowLinkStatus_LinkedOnlyMixedStatuses verifies filtering with various link statuses
func TestShowLinkStatus_LinkedOnlyMixedStatuses(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agent-smith-linked-only-mixed-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sourceDir := filepath.Join(tempDir, "source")
	target1Dir := filepath.Join(tempDir, "target1")
	target2Dir := filepath.Join(tempDir, "target2")

	for _, dir := range []string{
		filepath.Join(sourceDir, "skills", "partially-linked"),
		filepath.Join(sourceDir, "skills", "fully-unlinked"),
		filepath.Join(target1Dir, "skills"),
		filepath.Join(target2Dir, "skills"),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	for file, content := range map[string]string{
		filepath.Join(sourceDir, "skills", "partially-linked", "SKILL.md"): "# Partially Linked",
		filepath.Join(sourceDir, "skills", "fully-unlinked", "SKILL.md"):   "# Fully Unlinked",
	} {
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	if err := os.Symlink(
		filepath.Join(sourceDir, "skills", "partially-linked"),
		filepath.Join(target1Dir, "skills", "partially-linked"),
	); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	targets := []config.Target{
		&mockTarget{name: "TARGET1", baseDir: target1Dir},
		&mockTarget{name: "TARGET2", baseDir: target2Dir},
	}
	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create ComponentLinker: %v", err)
	}

	var buf bytes.Buffer
	linker.SetFormatter(formatter.NewWithWriter(&buf))
	if err := linker.ShowLinkStatus(true); err != nil {
		t.Fatalf("ShowLinkStatus(true) returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "partially-linked") {
		t.Error("Output should contain partially-linked (linked to TARGET1)")
	}
	if strings.Contains(output, "fully-unlinked") {
		t.Error("Output should NOT contain fully-unlinked")
	}
}

// TestShowLinkStatus_OutputFormat verifies the exact format of output remains unchanged
func TestShowLinkStatus_OutputFormat(t *testing.T) {
	// Create test environment
	tempDir, err := os.MkdirTemp("", "agent-smith-link-status-format-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sourceDir := filepath.Join(tempDir, "source")
	targetDir := filepath.Join(tempDir, "target")

	// Create structure with one component
	dirs := []string{
		filepath.Join(sourceDir, "skills", "test-skill"),
		filepath.Join(targetDir, "skills"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test file
	testFile := filepath.Join(sourceDir, "skills", "test-skill", "SKILL.md")
	if err := os.WriteFile(testFile, []byte("# Test Skill"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create mock target
	targets := []config.Target{
		&mockTarget{name: "OPENCODE", baseDir: targetDir},
	}

	// Create detector and linker
	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create ComponentLinker: %v", err)
	}

	// Create a buffer to capture output
	var buf bytes.Buffer
	linker.SetFormatter(formatter.NewWithWriter(&buf))

	err = linker.ShowLinkStatus(false)
	if err != nil {
		t.Fatalf("ShowLinkStatus() failed: %v", err)
	}

	// Get output
	output := buf.String()

	// Verify exact format structure (these are critical for backward compatibility)
	requiredSections := []string{
		"=== Link Status Across All Targets ===",
		"Skills:",
		"--- Legend ---",
		"--- Summary ---",
	}

	for _, section := range requiredSections {
		if !strings.Contains(output, section) {
			t.Errorf("Output format changed - missing required section: %s\nOutput:\n%s", section, output)
		}
	}

	// Verify legend entries haven't changed (check separately for symbol and text)
	legendSymbols := []string{"✓", "◆", "✗", "-", "?"}
	legendTexts := []string{"Valid symlink", "Copied directory", "Broken link", "Not linked", "Unknown status"}

	for _, symbol := range legendSymbols {
		if !strings.Contains(output, symbol) {
			t.Errorf("Legend format changed - missing symbol: %s\nOutput:\n%s", symbol, output)
		}
	}

	for _, text := range legendTexts {
		if !strings.Contains(output, text) {
			t.Errorf("Legend format changed - missing text: %s\nOutput:\n%s", text, output)
		}
	}
}
