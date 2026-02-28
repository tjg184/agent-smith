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

// TestShowLinkStatus_LinkedOnly verifies that linkedOnly=true filters out unlinked components
func TestShowLinkStatus_LinkedOnly(t *testing.T) {
	// Create test environment
	tempDir, err := os.MkdirTemp("", "agent-smith-linked-only-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup source directory structure
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

	// Create test files
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

	// Create symlinks only for linked components
	symlinks := []struct {
		source string
		target string
	}{
		{
			source: filepath.Join(sourceDir, "agents", "linked-agent"),
			target: filepath.Join(targetDir, "agents", "linked-agent"),
		},
		{
			source: filepath.Join(sourceDir, "skills", "linked-skill"),
			target: filepath.Join(targetDir, "skills", "linked-skill"),
		},
	}

	for _, link := range symlinks {
		if err := os.Symlink(link.source, link.target); err != nil {
			t.Fatalf("Failed to create symlink from %s to %s: %v", link.target, link.source, err)
		}
	}

	// Create mock target
	targets := []config.Target{
		&mockTarget{name: "TARGET", baseDir: targetDir},
	}

	// Create detector
	det := detector.NewRepositoryDetector()

	// Create ComponentLinker
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create ComponentLinker: %v", err)
	}

	// Test with linkedOnly=false (should show all components)
	var bufAll bytes.Buffer
	linker.SetFormatter(formatter.NewWithWriter(&bufAll))
	err = linker.ShowLinkStatus(false)
	if err != nil {
		t.Fatalf("ShowLinkStatus(false) returned error: %v", err)
	}

	outputAll := bufAll.String()

	// Should contain both linked and unlinked components
	if !strings.Contains(outputAll, "linked-agent") {
		t.Errorf("Output should contain linked-agent")
	}
	if !strings.Contains(outputAll, "unlinked-agent") {
		t.Errorf("Output should contain unlinked-agent")
	}
	if !strings.Contains(outputAll, "linked-skill") {
		t.Errorf("Output should contain linked-skill")
	}
	if !strings.Contains(outputAll, "unlinked-skill") {
		t.Errorf("Output should contain unlinked-skill")
	}

	// Test with linkedOnly=true (should show only linked components)
	var bufLinked bytes.Buffer
	linker.SetFormatter(formatter.NewWithWriter(&bufLinked))
	err = linker.ShowLinkStatus(true)
	if err != nil {
		t.Fatalf("ShowLinkStatus(true) returned error: %v", err)
	}

	outputLinked := bufLinked.String()

	// Should contain only linked components
	if !strings.Contains(outputLinked, "linked-agent") {
		t.Errorf("Output with linkedOnly=true should contain linked-agent")
	}
	if strings.Contains(outputLinked, "unlinked-agent") {
		t.Errorf("Output with linkedOnly=true should NOT contain unlinked-agent")
	}
	if !strings.Contains(outputLinked, "linked-skill") {
		t.Errorf("Output with linkedOnly=true should contain linked-skill")
	}
	if strings.Contains(outputLinked, "unlinked-skill") {
		t.Errorf("Output with linkedOnly=true should NOT contain unlinked-skill")
	}

	// Verify summary still shows correct total (including filtered items)
	if !strings.Contains(outputLinked, "Summary") {
		t.Errorf("Output should contain summary section")
	}
}

// TestShowLinkStatus_LinkedOnlyAllUnlinked verifies behavior when all components are unlinked
func TestShowLinkStatus_LinkedOnlyAllUnlinked(t *testing.T) {
	// Create test environment
	tempDir, err := os.MkdirTemp("", "agent-smith-linked-only-none-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup source directory structure
	sourceDir := filepath.Join(tempDir, "source")
	targetDir := filepath.Join(tempDir, "target")

	dirs := []string{
		filepath.Join(sourceDir, "skills", "unlinked-skill"),
		filepath.Join(targetDir, "skills"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test file
	skillFile := filepath.Join(sourceDir, "skills", "unlinked-skill", "SKILL.md")
	if err := os.WriteFile(skillFile, []byte("# Unlinked Skill"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create mock target (no links created)
	targets := []config.Target{
		&mockTarget{name: "TARGET", baseDir: targetDir},
	}

	// Create detector
	det := detector.NewRepositoryDetector()

	// Create ComponentLinker
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create ComponentLinker: %v", err)
	}

	// Test with linkedOnly=true (should show empty table)
	var buf bytes.Buffer
	linker.SetFormatter(formatter.NewWithWriter(&buf))
	err = linker.ShowLinkStatus(true)
	if err != nil {
		t.Fatalf("ShowLinkStatus(true) returned error: %v", err)
	}

	output := buf.String()

	// Should not contain the unlinked component
	if strings.Contains(output, "unlinked-skill") {
		t.Errorf("Output with linkedOnly=true should NOT contain unlinked-skill when all are unlinked")
	}

	// Should still show the table structure
	if !strings.Contains(output, "Link Status") {
		t.Errorf("Output should contain status header")
	}
	if !strings.Contains(output, "Legend") {
		t.Errorf("Output should contain legend")
	}
	if !strings.Contains(output, "Summary") {
		t.Errorf("Output should contain summary")
	}
}

// TestShowLinkStatus_LinkedOnlyMixedStatuses verifies filtering with various link statuses
func TestShowLinkStatus_LinkedOnlyMixedStatuses(t *testing.T) {
	// Create test environment
	tempDir, err := os.MkdirTemp("", "agent-smith-linked-only-mixed-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup source directory structure
	sourceDir := filepath.Join(tempDir, "source")
	target1Dir := filepath.Join(tempDir, "target1")
	target2Dir := filepath.Join(tempDir, "target2")

	dirs := []string{
		filepath.Join(sourceDir, "skills", "partially-linked"),
		filepath.Join(sourceDir, "skills", "fully-unlinked"),
		filepath.Join(target1Dir, "skills"),
		filepath.Join(target2Dir, "skills"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test files
	testFiles := map[string]string{
		filepath.Join(sourceDir, "skills", "partially-linked", "SKILL.md"): "# Partially Linked",
		filepath.Join(sourceDir, "skills", "fully-unlinked", "SKILL.md"):   "# Fully Unlinked",
	}

	for file, content := range testFiles {
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Create symlink only to target1 for partially-linked
	if err := os.Symlink(
		filepath.Join(sourceDir, "skills", "partially-linked"),
		filepath.Join(target1Dir, "skills", "partially-linked"),
	); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Create mock targets
	targets := []config.Target{
		&mockTarget{name: "TARGET1", baseDir: target1Dir},
		&mockTarget{name: "TARGET2", baseDir: target2Dir},
	}

	// Create detector
	det := detector.NewRepositoryDetector()

	// Create ComponentLinker
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create ComponentLinker: %v", err)
	}

	// Test with linkedOnly=true
	var buf bytes.Buffer
	linker.SetFormatter(formatter.NewWithWriter(&buf))
	err = linker.ShowLinkStatus(true)
	if err != nil {
		t.Fatalf("ShowLinkStatus(true) returned error: %v", err)
	}

	output := buf.String()

	// Should contain partially-linked (linked to at least one target)
	if !strings.Contains(output, "partially-linked") {
		t.Errorf("Output should contain partially-linked (linked to TARGET1)")
	}

	// Should NOT contain fully-unlinked (not linked to any target)
	if strings.Contains(output, "fully-unlinked") {
		t.Errorf("Output should NOT contain fully-unlinked")
	}
}
