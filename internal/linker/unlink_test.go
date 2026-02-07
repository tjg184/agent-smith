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
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
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
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
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
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
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
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
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
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
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
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
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
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
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
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
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

	// Unlink all components from target1 only (with force=true to skip confirmation, allProfiles=true for test)
	err = linker.UnlinkAllComponents("target1", true, true)
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
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
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

	// Unlink all components from all targets (with force=true to skip confirmation, allProfiles=true for test)
	err = linker.UnlinkAllComponents("", true, true)
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
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Try to unlink from non-existent target
	err = linker.UnlinkAllComponents("nonexistent", true, true)
	if err == nil {
		t.Errorf("Expected error when unlinking from non-existent target")
	}
}

func TestFilterTargets(t *testing.T) {
	sourceDir, _, _, targets, cleanup := setupTestEnvironment(t)
	defer cleanup()

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
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

// TestUnlinkComponent_PartialTargetLinking tests unlinking from one target when component is linked to multiple
func TestUnlinkComponent_PartialTargetLinking(t *testing.T) {
	sourceDir, target1Dir, target2Dir, targets, cleanup := setupTestEnvironment(t)
	defer cleanup()

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Create symlinks in both targets
	createSymlink(t, filepath.Join(sourceDir, "agents", "test-agent"),
		filepath.Join(target1Dir, "agents", "test-agent"))
	createSymlink(t, filepath.Join(sourceDir, "agents", "test-agent"),
		filepath.Join(target2Dir, "agents", "test-agent"))

	// Verify both symlinks exist
	if _, err := os.Lstat(filepath.Join(target1Dir, "agents", "test-agent")); err != nil {
		t.Fatalf("Symlink should exist in target1 before unlinking: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(target2Dir, "agents", "test-agent")); err != nil {
		t.Fatalf("Symlink should exist in target2 before unlinking: %v", err)
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

	// Verify symlink still exists in target2
	if _, err := os.Lstat(filepath.Join(target2Dir, "agents", "test-agent")); err != nil {
		t.Errorf("Symlink should still exist in target2: %v", err)
	}

	// Verify source still exists
	if _, err := os.Stat(filepath.Join(sourceDir, "agents", "test-agent")); err != nil {
		t.Errorf("Source directory should still exist: %v", err)
	}
}

// TestUnlinkComponent_BrokenSymlinkWithTarget tests unlinking a broken symlink from a specific target
func TestUnlinkComponent_BrokenSymlinkWithTarget(t *testing.T) {
	sourceDir, target1Dir, target2Dir, targets, cleanup := setupTestEnvironment(t)
	defer cleanup()

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Create symlinks in both targets
	createSymlink(t, filepath.Join(sourceDir, "agents", "test-agent"),
		filepath.Join(target1Dir, "agents", "test-agent"))
	createSymlink(t, filepath.Join(sourceDir, "agents", "test-agent"),
		filepath.Join(target2Dir, "agents", "test-agent"))

	// Remove source to create broken symlinks
	if err := os.RemoveAll(filepath.Join(sourceDir, "agents", "test-agent")); err != nil {
		t.Fatalf("Failed to remove source directory: %v", err)
	}

	// Unlink broken symlink from target1 only
	err = linker.UnlinkComponent("agents", "test-agent", "target1")
	if err != nil {
		t.Fatalf("UnlinkComponent should handle broken symlinks: %v", err)
	}

	// Verify symlink is removed from target1
	if _, err := os.Lstat(filepath.Join(target1Dir, "agents", "test-agent")); !os.IsNotExist(err) {
		t.Errorf("Broken symlink should be removed from target1")
	}

	// Verify broken symlink still exists in target2
	if _, err := os.Lstat(filepath.Join(target2Dir, "agents", "test-agent")); err != nil {
		t.Errorf("Broken symlink should still exist in target2: %v", err)
	}
}

// TestUnlinkComponent_AllTargetsExplicit tests unlinking from all targets using "all" filter
func TestUnlinkComponent_AllTargetsExplicit(t *testing.T) {
	sourceDir, target1Dir, target2Dir, targets, cleanup := setupTestEnvironment(t)
	defer cleanup()

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Create symlinks in both targets
	createSymlink(t, filepath.Join(sourceDir, "skills", "test-skill"),
		filepath.Join(target1Dir, "skills", "test-skill"))
	createSymlink(t, filepath.Join(sourceDir, "skills", "test-skill"),
		filepath.Join(target2Dir, "skills", "test-skill"))

	// Unlink from all targets using explicit "all" filter
	err = linker.UnlinkComponent("skills", "test-skill", "all")
	if err != nil {
		t.Fatalf("UnlinkComponent failed: %v", err)
	}

	// Verify symlinks are removed from both targets
	if _, err := os.Lstat(filepath.Join(target1Dir, "skills", "test-skill")); !os.IsNotExist(err) {
		t.Errorf("Symlink should be removed from target1")
	}
	if _, err := os.Lstat(filepath.Join(target2Dir, "skills", "test-skill")); !os.IsNotExist(err) {
		t.Errorf("Symlink should be removed from target2")
	}
}

// TestUnlinkComponent_OnlyInOneTarget tests unlinking when component is only in one target
func TestUnlinkComponent_OnlyInOneTarget(t *testing.T) {
	sourceDir, target1Dir, target2Dir, targets, cleanup := setupTestEnvironment(t)
	defer cleanup()

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Create symlink only in target2
	createSymlink(t, filepath.Join(sourceDir, "commands", "test-command"),
		filepath.Join(target2Dir, "commands", "test-command"))

	// Verify symlink exists only in target2
	if _, err := os.Lstat(filepath.Join(target1Dir, "commands", "test-command")); !os.IsNotExist(err) {
		t.Fatalf("Symlink should not exist in target1")
	}
	if _, err := os.Lstat(filepath.Join(target2Dir, "commands", "test-command")); err != nil {
		t.Fatalf("Symlink should exist in target2: %v", err)
	}

	// Try to unlink from target1 (should succeed but do nothing)
	err = linker.UnlinkComponent("commands", "test-command", "target1")
	if err == nil {
		t.Errorf("Expected error when component not linked to target1")
	}

	// Verify symlink still exists in target2
	if _, err := os.Lstat(filepath.Join(target2Dir, "commands", "test-command")); err != nil {
		t.Errorf("Symlink should still exist in target2: %v", err)
	}

	// Now unlink from target2 (should succeed)
	err = linker.UnlinkComponent("commands", "test-command", "target2")
	if err != nil {
		t.Fatalf("UnlinkComponent failed: %v", err)
	}

	// Verify symlink is removed from target2
	if _, err := os.Lstat(filepath.Join(target2Dir, "commands", "test-command")); !os.IsNotExist(err) {
		t.Errorf("Symlink should be removed from target2")
	}
}

// TestUnlinkComponentsByType_AllTargets tests unlinking all components of a type from all targets
func TestUnlinkComponentsByType_AllTargets(t *testing.T) {
	sourceDir, target1Dir, target2Dir, targets, cleanup := setupTestEnvironment(t)
	defer cleanup()

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Create multiple agent symlinks in both targets
	createSymlink(t, filepath.Join(sourceDir, "agents", "test-agent"),
		filepath.Join(target1Dir, "agents", "test-agent"))
	createSymlink(t, filepath.Join(sourceDir, "agents", "test-agent"),
		filepath.Join(target2Dir, "agents", "test-agent"))

	// Create a second agent
	secondAgentDir := filepath.Join(sourceDir, "agents", "test-agent-2")
	if err := os.MkdirAll(secondAgentDir, 0755); err != nil {
		t.Fatalf("Failed to create second agent dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(secondAgentDir, "README.md"), []byte("# Test Agent 2"), 0644); err != nil {
		t.Fatalf("Failed to create second agent file: %v", err)
	}

	createSymlink(t, secondAgentDir, filepath.Join(target1Dir, "agents", "test-agent-2"))
	createSymlink(t, secondAgentDir, filepath.Join(target2Dir, "agents", "test-agent-2"))

	// Unlink all agents from all targets (with force=true to skip confirmation)
	err = linker.UnlinkComponentsByType("agents", "", true)
	if err != nil {
		t.Fatalf("UnlinkComponentsByType failed: %v", err)
	}

	// Verify symlinks are removed from both targets
	if _, err := os.Lstat(filepath.Join(target1Dir, "agents", "test-agent")); !os.IsNotExist(err) {
		t.Errorf("Agent symlink should be removed from target1")
	}
	if _, err := os.Lstat(filepath.Join(target1Dir, "agents", "test-agent-2")); !os.IsNotExist(err) {
		t.Errorf("Second agent symlink should be removed from target1")
	}
	if _, err := os.Lstat(filepath.Join(target2Dir, "agents", "test-agent")); !os.IsNotExist(err) {
		t.Errorf("Agent symlink should be removed from target2")
	}
	if _, err := os.Lstat(filepath.Join(target2Dir, "agents", "test-agent-2")); !os.IsNotExist(err) {
		t.Errorf("Second agent symlink should be removed from target2")
	}

	// Verify source still exists
	if _, err := os.Stat(filepath.Join(sourceDir, "agents", "test-agent")); err != nil {
		t.Errorf("Source directory should still exist: %v", err)
	}
	if _, err := os.Stat(secondAgentDir); err != nil {
		t.Errorf("Second source directory should still exist: %v", err)
	}
}

// TestUnlinkComponentsByType_NonExistentTarget tests error handling for non-existent target
func TestUnlinkComponentsByType_NonExistentTarget(t *testing.T) {
	sourceDir, target1Dir, _, targets, cleanup := setupTestEnvironment(t)
	defer cleanup()

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Create symlink in target1
	createSymlink(t, filepath.Join(sourceDir, "skills", "test-skill"),
		filepath.Join(target1Dir, "skills", "test-skill"))

	// Try to unlink from non-existent target
	err = linker.UnlinkComponentsByType("skills", "nonexistent", true)
	if err == nil {
		t.Errorf("Expected error when unlinking from non-existent target")
	}

	// Verify symlink still exists in target1 (should not be affected)
	if _, err := os.Lstat(filepath.Join(target1Dir, "skills", "test-skill")); err != nil {
		t.Errorf("Symlink should still exist in target1: %v", err)
	}
}

// TestUnlinkComponentsByType_MixedLinksInTargets tests unlinking when different targets have different components
func TestUnlinkComponentsByType_MixedLinksInTargets(t *testing.T) {
	sourceDir, target1Dir, target2Dir, targets, cleanup := setupTestEnvironment(t)
	defer cleanup()

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Create different skills in different targets
	createSymlink(t, filepath.Join(sourceDir, "skills", "test-skill"),
		filepath.Join(target1Dir, "skills", "test-skill"))

	// Create a second skill
	secondSkillDir := filepath.Join(sourceDir, "skills", "test-skill-2")
	if err := os.MkdirAll(secondSkillDir, 0755); err != nil {
		t.Fatalf("Failed to create second skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(secondSkillDir, "SKILL.md"), []byte("# Test Skill 2"), 0644); err != nil {
		t.Fatalf("Failed to create second skill file: %v", err)
	}

	// Link first skill to target2 and second skill only to target2
	createSymlink(t, filepath.Join(sourceDir, "skills", "test-skill"),
		filepath.Join(target2Dir, "skills", "test-skill"))
	createSymlink(t, secondSkillDir, filepath.Join(target2Dir, "skills", "test-skill-2"))

	// Unlink all skills from target1 only
	err = linker.UnlinkComponentsByType("skills", "target1", true)
	if err != nil {
		t.Fatalf("UnlinkComponentsByType failed: %v", err)
	}

	// Verify first skill is removed from target1
	if _, err := os.Lstat(filepath.Join(target1Dir, "skills", "test-skill")); !os.IsNotExist(err) {
		t.Errorf("Skill symlink should be removed from target1")
	}

	// Verify both skills still exist in target2
	if _, err := os.Lstat(filepath.Join(target2Dir, "skills", "test-skill")); err != nil {
		t.Errorf("Skill symlink should still exist in target2: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(target2Dir, "skills", "test-skill-2")); err != nil {
		t.Errorf("Second skill symlink should still exist in target2: %v", err)
	}
}

// TestUnlinkAllComponents_MixedComponentTypes tests unlinking all component types from specific target
func TestUnlinkAllComponents_MixedComponentTypes(t *testing.T) {
	sourceDir, target1Dir, target2Dir, targets, cleanup := setupTestEnvironment(t)
	defer cleanup()

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Create one of each component type in both targets
	createSymlink(t, filepath.Join(sourceDir, "agents", "test-agent"),
		filepath.Join(target1Dir, "agents", "test-agent"))
	createSymlink(t, filepath.Join(sourceDir, "skills", "test-skill"),
		filepath.Join(target1Dir, "skills", "test-skill"))
	createSymlink(t, filepath.Join(sourceDir, "commands", "test-command"),
		filepath.Join(target1Dir, "commands", "test-command"))

	// Create additional components only in target2
	createSymlink(t, filepath.Join(sourceDir, "agents", "test-agent"),
		filepath.Join(target2Dir, "agents", "test-agent"))
	createSymlink(t, filepath.Join(sourceDir, "skills", "test-skill"),
		filepath.Join(target2Dir, "skills", "test-skill"))

	// Unlink all from target1 only
	err = linker.UnlinkAllComponents("target1", true, true)
	if err != nil {
		t.Fatalf("UnlinkAllComponents failed: %v", err)
	}

	// Verify all are removed from target1
	if _, err := os.Lstat(filepath.Join(target1Dir, "agents", "test-agent")); !os.IsNotExist(err) {
		t.Errorf("Agent symlink should be removed from target1")
	}
	if _, err := os.Lstat(filepath.Join(target1Dir, "skills", "test-skill")); !os.IsNotExist(err) {
		t.Errorf("Skill symlink should be removed from target1")
	}
	if _, err := os.Lstat(filepath.Join(target1Dir, "commands", "test-command")); !os.IsNotExist(err) {
		t.Errorf("Command symlink should be removed from target1")
	}

	// Verify target2 components still exist
	if _, err := os.Lstat(filepath.Join(target2Dir, "agents", "test-agent")); err != nil {
		t.Errorf("Agent symlink should still exist in target2: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(target2Dir, "skills", "test-skill")); err != nil {
		t.Errorf("Skill symlink should still exist in target2: %v", err)
	}
}

// TestUnlinkAllComponents_EmptyTarget tests unlinking when target has no components
func TestUnlinkAllComponents_EmptyTarget(t *testing.T) {
	sourceDir, _, target2Dir, targets, cleanup := setupTestEnvironment(t)
	defer cleanup()

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(sourceDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Create symlink only in target2
	createSymlink(t, filepath.Join(sourceDir, "agents", "test-agent"),
		filepath.Join(target2Dir, "agents", "test-agent"))

	// Try to unlink from target1 (which has no components)
	err = linker.UnlinkAllComponents("target1", true, true)
	if err != nil {
		t.Fatalf("UnlinkAllComponents should handle empty target: %v", err)
	}

	// Verify target2 component still exists
	if _, err := os.Lstat(filepath.Join(target2Dir, "agents", "test-agent")); err != nil {
		t.Errorf("Agent symlink should still exist in target2: %v", err)
	}
}

// TestIsSymlinkFromCurrentProfile_BaseInstallation tests that base installation only matches base symlinks
func TestIsSymlinkFromCurrentProfile_BaseInstallation(t *testing.T) {
	// Create temp directory structure
	tempDir, err := os.MkdirTemp("", "agent-smith-profile-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create base installation structure
	baseDir := filepath.Join(tempDir, ".agent-smith")
	baseSkillsDir := filepath.Join(baseDir, "skills", "test-skill")
	profileDir := filepath.Join(baseDir, "profiles", "work")
	profileSkillsDir := filepath.Join(profileDir, "skills", "profile-skill")
	targetDir := filepath.Join(tempDir, "target")

	for _, dir := range []string{baseSkillsDir, profileSkillsDir, targetDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test files
	if err := os.WriteFile(filepath.Join(baseSkillsDir, "SKILL.md"), []byte("# Base Skill"), 0644); err != nil {
		t.Fatalf("Failed to create base skill file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(profileSkillsDir, "SKILL.md"), []byte("# Profile Skill"), 0644); err != nil {
		t.Fatalf("Failed to create profile skill file: %v", err)
	}

	// Create target skills directory
	targetSkillsDir := filepath.Join(targetDir, "skills")
	if err := os.MkdirAll(targetSkillsDir, 0755); err != nil {
		t.Fatalf("Failed to create target skills dir: %v", err)
	}

	// Create symlinks: one to base, one to profile
	baseSymlink := filepath.Join(targetSkillsDir, "base-skill")
	profileSymlink := filepath.Join(targetSkillsDir, "profile-skill")

	// Create base symlink (relative path)
	baseRelPath, _ := filepath.Rel(targetSkillsDir, baseSkillsDir)
	if err := os.Symlink(baseRelPath, baseSymlink); err != nil {
		t.Fatalf("Failed to create base symlink: %v", err)
	}

	// Create profile symlink (relative path)
	profileRelPath, _ := filepath.Rel(targetSkillsDir, profileSkillsDir)
	if err := os.Symlink(profileRelPath, profileSymlink); err != nil {
		t.Fatalf("Failed to create profile symlink: %v", err)
	}

	// Create linker with base directory
	det := detector.NewRepositoryDetector()
	mockTargets := []config.Target{&mockTarget{name: "test-target", baseDir: targetDir}}
	linker, err := NewComponentLinker(baseDir, mockTargets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Test: base symlink should be recognized as from current profile (base)
	isBase, err := linker.isSymlinkFromCurrentProfile(baseSymlink)
	if err != nil {
		t.Fatalf("isSymlinkFromCurrentProfile failed for base symlink: %v", err)
	}
	if !isBase {
		t.Errorf("Base symlink should be recognized as from current profile (base), got false")
	}

	// Test: profile symlink should NOT be recognized as from current profile (base)
	isProfile, err := linker.isSymlinkFromCurrentProfile(profileSymlink)
	if err != nil {
		t.Fatalf("isSymlinkFromCurrentProfile failed for profile symlink: %v", err)
	}
	if isProfile {
		t.Errorf("Profile symlink should NOT be recognized as from current profile (base), got true")
	}
}

// TestIsSymlinkFromCurrentProfile_ProfileInstallation tests that profile installation only matches that profile's symlinks
func TestIsSymlinkFromCurrentProfile_ProfileInstallation(t *testing.T) {
	// Create temp directory structure
	tempDir, err := os.MkdirTemp("", "agent-smith-profile-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create base installation structure
	baseDir := filepath.Join(tempDir, ".agent-smith")
	workProfileDir := filepath.Join(baseDir, "profiles", "work")
	personalProfileDir := filepath.Join(baseDir, "profiles", "personal")
	targetDir := filepath.Join(tempDir, "target")

	workSkillDir := filepath.Join(workProfileDir, "skills", "work-skill")
	personalSkillDir := filepath.Join(personalProfileDir, "skills", "personal-skill")

	for _, dir := range []string{workSkillDir, personalSkillDir, targetDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test files
	for _, f := range []struct {
		path string
		data string
	}{
		{filepath.Join(workSkillDir, "SKILL.md"), "# Work Skill"},
		{filepath.Join(personalSkillDir, "SKILL.md"), "# Personal Skill"},
	} {
		if err := os.WriteFile(f.path, []byte(f.data), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", f.path, err)
		}
	}

	// Create target skills directory
	targetSkillsDir := filepath.Join(targetDir, "skills")
	if err := os.MkdirAll(targetSkillsDir, 0755); err != nil {
		t.Fatalf("Failed to create target skills dir: %v", err)
	}

	// Create symlinks: one to work profile, one to personal profile
	workSymlink := filepath.Join(targetSkillsDir, "work-skill")
	personalSymlink := filepath.Join(targetSkillsDir, "personal-skill")

	// Create work symlink (relative path)
	workRelPath, err := filepath.Rel(targetSkillsDir, workSkillDir)
	if err != nil {
		t.Fatalf("Failed to create relative path for work: %v", err)
	}
	if err := os.Symlink(workRelPath, workSymlink); err != nil {
		t.Fatalf("Failed to create work symlink: %v", err)
	}

	// Create personal symlink (relative path)
	personalRelPath, err := filepath.Rel(targetSkillsDir, personalSkillDir)
	if err != nil {
		t.Fatalf("Failed to create relative path for personal: %v", err)
	}
	if err := os.Symlink(personalRelPath, personalSymlink); err != nil {
		t.Fatalf("Failed to create personal symlink: %v", err)
	}

	// Create linker with work profile directory
	det := detector.NewRepositoryDetector()
	mockTargets := []config.Target{&mockTarget{name: "test-target", baseDir: targetDir}}
	linker, err := NewComponentLinker(workProfileDir, mockTargets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Test: work profile symlink should be recognized as from current profile
	isWork, err := linker.isSymlinkFromCurrentProfile(workSymlink)
	if err != nil {
		t.Fatalf("isSymlinkFromCurrentProfile failed for work symlink: %v", err)
	}
	if !isWork {
		t.Errorf("Work profile symlink should be recognized as from current profile, got false")
	}

	// Test: personal profile symlink should NOT be recognized as from current profile (work)
	isPersonal, err := linker.isSymlinkFromCurrentProfile(personalSymlink)
	if err != nil {
		t.Fatalf("isSymlinkFromCurrentProfile failed for personal symlink: %v", err)
	}
	if isPersonal {
		t.Errorf("Personal profile symlink should NOT be recognized as from current profile (work), got true")
	}
}
