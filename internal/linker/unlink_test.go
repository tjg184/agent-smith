package linker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tgaines/agent-smith/internal/detector"
	"github.com/tgaines/agent-smith/pkg/config"
)

// mockTarget implements config.Target for testing
type mockTarget struct {
	name    string
	baseDir string
}

func (m *mockTarget) GetName() string {
	return m.name
}

func (m *mockTarget) GetBaseDir() (string, error) {
	return m.baseDir, nil
}

func (m *mockTarget) GetSkillsDir() (string, error) {
	return filepath.Join(m.baseDir, "skills"), nil
}

func (m *mockTarget) GetAgentsDir() (string, error) {
	return filepath.Join(m.baseDir, "agents"), nil
}

func (m *mockTarget) GetCommandsDir() (string, error) {
	return filepath.Join(m.baseDir, "commands"), nil
}

func (m *mockTarget) GetComponentDir(componentType string) (string, error) {
	return filepath.Join(m.baseDir, componentType), nil
}

func (m *mockTarget) GetDetectionConfigPath() (string, error) {
	return filepath.Join(m.baseDir, ".detection-config.json"), nil
}

// setupTestEnvironment creates a test directory structure with source and targets
func setupTestEnvironment(t *testing.T) (string, string, string, []config.Target, func()) {
	t.Helper()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "agent-smith-unlink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	sourceDir := filepath.Join(tempDir, "source")
	target1Dir := filepath.Join(tempDir, "target1")
	target2Dir := filepath.Join(tempDir, "target2")

	// Create directory structure
	dirs := []string{
		filepath.Join(sourceDir, "agents", "test-agent"),
		filepath.Join(sourceDir, "skills", "test-skill"),
		filepath.Join(sourceDir, "commands", "test-command"),
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

	// Create test files in source
	testFiles := []string{
		filepath.Join(sourceDir, "agents", "test-agent", "README.md"),
		filepath.Join(sourceDir, "skills", "test-skill", "SKILL.md"),
		filepath.Join(sourceDir, "commands", "test-command", "README.md"),
	}

	for _, file := range testFiles {
		if err := os.WriteFile(file, []byte("# Test Component"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Create mock targets
	targets := []config.Target{
		&mockTarget{name: "target1", baseDir: target1Dir},
		&mockTarget{name: "target2", baseDir: target2Dir},
	}

	// Cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return sourceDir, target1Dir, target2Dir, targets, cleanup
}

// createSymlink creates a symlink from src to dst
func createSymlink(t *testing.T, src, dst string) {
	t.Helper()

	// Remove existing destination if it exists
	os.Remove(dst)

	// Create relative path for the symlink
	relPath, err := filepath.Rel(filepath.Dir(dst), src)
	if err != nil {
		t.Fatalf("Failed to create relative path: %v", err)
	}

	if err := os.Symlink(relPath, dst); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}
}

func TestUnlinkComponent_SingleTarget(t *testing.T) {
	sourceDir, target1Dir, _, targets, cleanup := setupTestEnvironment(t)
	defer cleanup()

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Create symlinks in both targets
	createSymlink(t, filepath.Join(sourceDir, "agents", "test-agent"),
		filepath.Join(target1Dir, "agents", "test-agent"))

	// Verify symlink exists
	if _, err := os.Lstat(filepath.Join(target1Dir, "agents", "test-agent")); err != nil {
		t.Fatalf("Symlink should exist before unlinking: %v", err)
	}

	// Unlink from target1 only
	err = linker.UnlinkComponent("agents", "test-agent", "target1")
	if err != nil {
		t.Fatalf("UnlinkComponent failed: %v", err)
	}

	// Verify symlink is removed from target1
	if _, err := os.Lstat(filepath.Join(target1Dir, "agents", "test-agent")); !os.IsNotExist(err) {
		t.Errorf("Symlink should be removed from target1")
	}

	// Verify source still exists
	if _, err := os.Stat(filepath.Join(sourceDir, "agents", "test-agent")); err != nil {
		t.Errorf("Source directory should still exist: %v", err)
	}
}

func TestUnlinkComponent_AllTargets(t *testing.T) {
	sourceDir, target1Dir, target2Dir, targets, cleanup := setupTestEnvironment(t)
	defer cleanup()

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Create symlinks in both targets
	createSymlink(t, filepath.Join(sourceDir, "agents", "test-agent"),
		filepath.Join(target1Dir, "agents", "test-agent"))
	createSymlink(t, filepath.Join(sourceDir, "agents", "test-agent"),
		filepath.Join(target2Dir, "agents", "test-agent"))

	// Verify symlinks exist
	if _, err := os.Lstat(filepath.Join(target1Dir, "agents", "test-agent")); err != nil {
		t.Fatalf("Symlink should exist in target1 before unlinking: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(target2Dir, "agents", "test-agent")); err != nil {
		t.Fatalf("Symlink should exist in target2 before unlinking: %v", err)
	}

	// Unlink from all targets (empty targetFilter)
	err = linker.UnlinkComponent("agents", "test-agent", "")
	if err != nil {
		t.Fatalf("UnlinkComponent failed: %v", err)
	}

	// Verify symlinks are removed from both targets
	if _, err := os.Lstat(filepath.Join(target1Dir, "agents", "test-agent")); !os.IsNotExist(err) {
		t.Errorf("Symlink should be removed from target1")
	}
	if _, err := os.Lstat(filepath.Join(target2Dir, "agents", "test-agent")); !os.IsNotExist(err) {
		t.Errorf("Symlink should be removed from target2")
	}

	// Verify source still exists
	if _, err := os.Stat(filepath.Join(sourceDir, "agents", "test-agent")); err != nil {
		t.Errorf("Source directory should still exist: %v", err)
	}
}

func TestUnlinkComponent_NonExistentTarget(t *testing.T) {
	sourceDir, target1Dir, _, targets, cleanup := setupTestEnvironment(t)
	defer cleanup()

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Create symlink in target1
	createSymlink(t, filepath.Join(sourceDir, "agents", "test-agent"),
		filepath.Join(target1Dir, "agents", "test-agent"))

	// Try to unlink from non-existent target
	err = linker.UnlinkComponent("agents", "test-agent", "nonexistent")
	if err == nil {
		t.Errorf("Expected error when unlinking from non-existent target")
	}

	// Verify symlink still exists in target1 (should not be affected)
	if _, err := os.Lstat(filepath.Join(target1Dir, "agents", "test-agent")); err != nil {
		t.Errorf("Symlink should still exist in target1: %v", err)
	}
}

func TestUnlinkComponent_ComponentNotLinked(t *testing.T) {
	sourceDir, _, _, targets, cleanup := setupTestEnvironment(t)
	defer cleanup()

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Try to unlink a component that's not linked
	err = linker.UnlinkComponent("agents", "test-agent", "target1")
	if err == nil {
		t.Errorf("Expected error when unlinking component that's not linked")
	}
}

func TestUnlinkComponent_InvalidComponentType(t *testing.T) {
	sourceDir, _, _, targets, cleanup := setupTestEnvironment(t)
	defer cleanup()

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Try to unlink with invalid component type
	err = linker.UnlinkComponent("invalid", "test-agent", "target1")
	if err == nil {
		t.Errorf("Expected error when using invalid component type")
	}
}

func TestUnlinkComponent_SkillsAndCommands(t *testing.T) {
	sourceDir, target1Dir, _, targets, cleanup := setupTestEnvironment(t)
	defer cleanup()

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Create symlinks for skills and commands
	createSymlink(t, filepath.Join(sourceDir, "skills", "test-skill"),
		filepath.Join(target1Dir, "skills", "test-skill"))
	createSymlink(t, filepath.Join(sourceDir, "commands", "test-command"),
		filepath.Join(target1Dir, "commands", "test-command"))

	// Unlink skill
	err = linker.UnlinkComponent("skills", "test-skill", "target1")
	if err != nil {
		t.Fatalf("Failed to unlink skill: %v", err)
	}

	// Verify skill symlink is removed
	if _, err := os.Lstat(filepath.Join(target1Dir, "skills", "test-skill")); !os.IsNotExist(err) {
		t.Errorf("Skill symlink should be removed")
	}

	// Unlink command
	err = linker.UnlinkComponent("commands", "test-command", "target1")
	if err != nil {
		t.Fatalf("Failed to unlink command: %v", err)
	}

	// Verify command symlink is removed
	if _, err := os.Lstat(filepath.Join(target1Dir, "commands", "test-command")); !os.IsNotExist(err) {
		t.Errorf("Command symlink should be removed")
	}
}

func TestUnlinkComponentsByType_SpecificTarget(t *testing.T) {
	sourceDir, target1Dir, target2Dir, targets, cleanup := setupTestEnvironment(t)
	defer cleanup()

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Create multiple command symlinks in both targets
	createSymlink(t, filepath.Join(sourceDir, "commands", "test-command"),
		filepath.Join(target1Dir, "commands", "test-command"))

	// Create a second command
	secondCommandDir := filepath.Join(sourceDir, "commands", "test-command-2")
	if err := os.MkdirAll(secondCommandDir, 0755); err != nil {
		t.Fatalf("Failed to create second command dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(secondCommandDir, "README.md"), []byte("# Test Command 2"), 0644); err != nil {
		t.Fatalf("Failed to create second command file: %v", err)
	}

	createSymlink(t, secondCommandDir,
		filepath.Join(target1Dir, "commands", "test-command-2"))
	createSymlink(t, filepath.Join(sourceDir, "commands", "test-command"),
		filepath.Join(target2Dir, "commands", "test-command"))
	createSymlink(t, secondCommandDir,
		filepath.Join(target2Dir, "commands", "test-command-2"))

	// Verify symlinks exist in both targets
	if _, err := os.Lstat(filepath.Join(target1Dir, "commands", "test-command")); err != nil {
		t.Fatalf("Symlink should exist in target1 before unlinking: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(target1Dir, "commands", "test-command-2")); err != nil {
		t.Fatalf("Symlink should exist in target1 before unlinking: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(target2Dir, "commands", "test-command")); err != nil {
		t.Fatalf("Symlink should exist in target2 before unlinking: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(target2Dir, "commands", "test-command-2")); err != nil {
		t.Fatalf("Symlink should exist in target2 before unlinking: %v", err)
	}

	// Unlink all commands from target1 only (with force=true to skip confirmation)
	err = linker.UnlinkComponentsByType("commands", "target1", true)
	if err != nil {
		t.Fatalf("UnlinkComponentsByType failed: %v", err)
	}

	// Verify symlinks are removed from target1
	if _, err := os.Lstat(filepath.Join(target1Dir, "commands", "test-command")); !os.IsNotExist(err) {
		t.Errorf("Symlink should be removed from target1")
	}
	if _, err := os.Lstat(filepath.Join(target1Dir, "commands", "test-command-2")); !os.IsNotExist(err) {
		t.Errorf("Second symlink should be removed from target1")
	}

	// Verify symlinks still exist in target2 (should not be affected)
	if _, err := os.Lstat(filepath.Join(target2Dir, "commands", "test-command")); err != nil {
		t.Errorf("Symlink should still exist in target2: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(target2Dir, "commands", "test-command-2")); err != nil {
		t.Errorf("Second symlink should still exist in target2: %v", err)
	}

	// Verify source still exists
	if _, err := os.Stat(filepath.Join(sourceDir, "commands", "test-command")); err != nil {
		t.Errorf("Source directory should still exist: %v", err)
	}
	if _, err := os.Stat(secondCommandDir); err != nil {
		t.Errorf("Second source directory should still exist: %v", err)
	}
}

func TestUnlinkAllComponents_SpecificTarget(t *testing.T) {
	sourceDir, target1Dir, target2Dir, targets, cleanup := setupTestEnvironment(t)
	defer cleanup()

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Create symlinks for all component types in both targets
	createSymlink(t, filepath.Join(sourceDir, "agents", "test-agent"),
		filepath.Join(target1Dir, "agents", "test-agent"))
	createSymlink(t, filepath.Join(sourceDir, "skills", "test-skill"),
		filepath.Join(target1Dir, "skills", "test-skill"))
	createSymlink(t, filepath.Join(sourceDir, "commands", "test-command"),
		filepath.Join(target1Dir, "commands", "test-command"))
	createSymlink(t, filepath.Join(sourceDir, "agents", "test-agent"),
		filepath.Join(target2Dir, "agents", "test-agent"))
	createSymlink(t, filepath.Join(sourceDir, "skills", "test-skill"),
		filepath.Join(target2Dir, "skills", "test-skill"))
	createSymlink(t, filepath.Join(sourceDir, "commands", "test-command"),
		filepath.Join(target2Dir, "commands", "test-command"))

	// Unlink all components from target1 only (with force=true to skip confirmation)
	err = linker.UnlinkAllComponents("target1", true)
	if err != nil {
		t.Fatalf("UnlinkAllComponents failed: %v", err)
	}

	// Verify symlinks are removed from target1
	if _, err := os.Lstat(filepath.Join(target1Dir, "agents", "test-agent")); !os.IsNotExist(err) {
		t.Errorf("Agent symlink should be removed from target1")
	}
	if _, err := os.Lstat(filepath.Join(target1Dir, "skills", "test-skill")); !os.IsNotExist(err) {
		t.Errorf("Skill symlink should be removed from target1")
	}
	if _, err := os.Lstat(filepath.Join(target1Dir, "commands", "test-command")); !os.IsNotExist(err) {
		t.Errorf("Command symlink should be removed from target1")
	}

	// Verify symlinks still exist in target2 (should not be affected)
	if _, err := os.Lstat(filepath.Join(target2Dir, "agents", "test-agent")); err != nil {
		t.Errorf("Agent symlink should still exist in target2: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(target2Dir, "skills", "test-skill")); err != nil {
		t.Errorf("Skill symlink should still exist in target2: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(target2Dir, "commands", "test-command")); err != nil {
		t.Errorf("Command symlink should still exist in target2: %v", err)
	}

	// Verify source still exists
	if _, err := os.Stat(filepath.Join(sourceDir, "agents", "test-agent")); err != nil {
		t.Errorf("Agent source directory should still exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(sourceDir, "skills", "test-skill")); err != nil {
		t.Errorf("Skill source directory should still exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(sourceDir, "commands", "test-command")); err != nil {
		t.Errorf("Command source directory should still exist: %v", err)
	}
}

func TestUnlinkAllComponents_AllTargets(t *testing.T) {
	sourceDir, target1Dir, target2Dir, targets, cleanup := setupTestEnvironment(t)
	defer cleanup()

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Create symlinks for all component types in both targets
	createSymlink(t, filepath.Join(sourceDir, "agents", "test-agent"),
		filepath.Join(target1Dir, "agents", "test-agent"))
	createSymlink(t, filepath.Join(sourceDir, "skills", "test-skill"),
		filepath.Join(target1Dir, "skills", "test-skill"))
	createSymlink(t, filepath.Join(sourceDir, "commands", "test-command"),
		filepath.Join(target1Dir, "commands", "test-command"))
	createSymlink(t, filepath.Join(sourceDir, "agents", "test-agent"),
		filepath.Join(target2Dir, "agents", "test-agent"))
	createSymlink(t, filepath.Join(sourceDir, "skills", "test-skill"),
		filepath.Join(target2Dir, "skills", "test-skill"))
	createSymlink(t, filepath.Join(sourceDir, "commands", "test-command"),
		filepath.Join(target2Dir, "commands", "test-command"))

	// Unlink all components from all targets (with force=true to skip confirmation)
	err = linker.UnlinkAllComponents("", true)
	if err != nil {
		t.Fatalf("UnlinkAllComponents failed: %v", err)
	}

	// Verify symlinks are removed from both targets
	if _, err := os.Lstat(filepath.Join(target1Dir, "agents", "test-agent")); !os.IsNotExist(err) {
		t.Errorf("Agent symlink should be removed from target1")
	}
	if _, err := os.Lstat(filepath.Join(target1Dir, "skills", "test-skill")); !os.IsNotExist(err) {
		t.Errorf("Skill symlink should be removed from target1")
	}
	if _, err := os.Lstat(filepath.Join(target1Dir, "commands", "test-command")); !os.IsNotExist(err) {
		t.Errorf("Command symlink should be removed from target1")
	}
	if _, err := os.Lstat(filepath.Join(target2Dir, "agents", "test-agent")); !os.IsNotExist(err) {
		t.Errorf("Agent symlink should be removed from target2")
	}
	if _, err := os.Lstat(filepath.Join(target2Dir, "skills", "test-skill")); !os.IsNotExist(err) {
		t.Errorf("Skill symlink should be removed from target2")
	}
	if _, err := os.Lstat(filepath.Join(target2Dir, "commands", "test-command")); !os.IsNotExist(err) {
		t.Errorf("Command symlink should be removed from target2")
	}

	// Verify source still exists
	if _, err := os.Stat(filepath.Join(sourceDir, "agents", "test-agent")); err != nil {
		t.Errorf("Agent source directory should still exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(sourceDir, "skills", "test-skill")); err != nil {
		t.Errorf("Skill source directory should still exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(sourceDir, "commands", "test-command")); err != nil {
		t.Errorf("Command source directory should still exist: %v", err)
	}
}

func TestUnlinkAllComponents_NonExistentTarget(t *testing.T) {
	sourceDir, _, _, targets, cleanup := setupTestEnvironment(t)
	defer cleanup()

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Try to unlink from non-existent target
	err = linker.UnlinkAllComponents("nonexistent", true)
	if err == nil {
		t.Errorf("Expected error when unlinking from non-existent target")
	}
}

func TestFilterTargets(t *testing.T) {
	sourceDir, _, _, targets, cleanup := setupTestEnvironment(t)
	defer cleanup()

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	tests := []struct {
		name          string
		filter        string
		expectedCount int
		expectedNames []string
	}{
		{
			name:          "empty filter returns all targets",
			filter:        "",
			expectedCount: 2,
			expectedNames: []string{"target1", "target2"},
		},
		{
			name:          "all filter returns all targets",
			filter:        "all",
			expectedCount: 2,
			expectedNames: []string{"target1", "target2"},
		},
		{
			name:          "specific target filter",
			filter:        "target1",
			expectedCount: 1,
			expectedNames: []string{"target1"},
		},
		{
			name:          "non-existent target returns empty",
			filter:        "nonexistent",
			expectedCount: 0,
			expectedNames: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := linker.filterTargets(tt.filter)
			if len(filtered) != tt.expectedCount {
				t.Errorf("Expected %d targets, got %d", tt.expectedCount, len(filtered))
			}

			// Check target names
			for i, target := range filtered {
				if i < len(tt.expectedNames) && target.GetName() != tt.expectedNames[i] {
					t.Errorf("Expected target name %s, got %s", tt.expectedNames[i], target.GetName())
				}
			}
		})
	}
}
