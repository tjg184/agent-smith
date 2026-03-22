package linkerUnlink

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/pkg/config"
)

type stubTarget struct{ baseDir string }

func (s *stubTarget) GetName() string        { return "test-target" }
func (s *stubTarget) GetDisplayName() string { return "Test Target" }
func (s *stubTarget) GetGlobalBaseDir() (string, error) {
	return s.baseDir, nil
}
func (s *stubTarget) GetGlobalSkillsDir() (string, error) {
	return filepath.Join(s.baseDir, "skills"), nil
}
func (s *stubTarget) GetGlobalAgentsDir() (string, error) {
	return filepath.Join(s.baseDir, "agents"), nil
}
func (s *stubTarget) GetGlobalCommandsDir() (string, error) {
	return filepath.Join(s.baseDir, "commands"), nil
}
func (s *stubTarget) GetGlobalComponentDir(componentType string) (string, error) {
	return filepath.Join(s.baseDir, componentType), nil
}
func (s *stubTarget) GetDetectionConfigPath() (string, error) {
	return filepath.Join(s.baseDir, ".detection-config.json"), nil
}
func (s *stubTarget) GetProjectDirName() string { return ".stub" }
func (s *stubTarget) GetProjectBaseDir(root string) string {
	return filepath.Join(root, ".stub")
}
func (s *stubTarget) GetProjectComponentDir(projectRoot, componentType string) (string, error) {
	return filepath.Join(projectRoot, ".stub", componentType), nil
}
func (s *stubTarget) IsUniversalTarget() bool { return false }

var _ config.Target = (*stubTarget)(nil)

func makeSymlink(t *testing.T, src, dst string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(src, dst); err != nil {
		t.Fatal(err)
	}
}

func TestIsManagedCategoryDir_DeeplyNested(t *testing.T) {
	dir := t.TempDir()
	target := t.TempDir()

	leaf := filepath.Join(dir, "a", "b", "c", "d", "e")
	makeSymlink(t, target, leaf)

	if !isManagedCategoryDir(dir) {
		t.Error("expected deeply nested symlink dir to be managed")
	}
}

func TestIsManagedCategoryDir_ContainsRegularFile(t *testing.T) {
	dir := t.TempDir()
	target := t.TempDir()

	leaf := filepath.Join(dir, "a", "b", "skill")
	makeSymlink(t, target, leaf)

	regular := filepath.Join(dir, "a", "README.md")
	if err := os.WriteFile(regular, []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}

	if isManagedCategoryDir(dir) {
		t.Error("expected dir with regular file to NOT be managed")
	}
}

func TestIsManagedCategoryDir_Empty(t *testing.T) {
	dir := t.TempDir()
	if isManagedCategoryDir(dir) {
		t.Error("expected empty dir to NOT be managed")
	}
}

func TestCountManagedLeafSymlinks_DeeplyNested(t *testing.T) {
	categoryDir := t.TempDir()
	target := t.TempDir()

	symlinkPaths := []string{
		"a/b/c/skill1",
		"a/b/skill2",
		"x/y/z/w/skill3",
	}
	for _, p := range symlinkPaths {
		makeSymlink(t, target, filepath.Join(categoryDir, p))
	}

	count, err := countManagedLeafSymlinks(categoryDir, "")
	if err != nil {
		t.Fatalf("countManagedLeafSymlinks: %v", err)
	}
	if count != 3 {
		t.Errorf("got %d, want 3", count)
	}
}

func TestRemoveManagedLeafSymlinks_DeeplyNested(t *testing.T) {
	categoryDir := t.TempDir()
	skillSrc := t.TempDir()

	leafPaths := []string{
		"test/test2/test3/test4/test5",
		"test/test2/other",
	}
	for _, p := range leafPaths {
		makeSymlink(t, skillSrc, filepath.Join(categoryDir, p))
	}

	f := formatter.New()
	removed, skipped := removeManagedLeafSymlinks(f, categoryDir, "", true, "test-target", "skills", false, nil)

	if removed != 2 {
		t.Errorf("removed = %d, want 2", removed)
	}
	if skipped != 0 {
		t.Errorf("skipped = %d, want 0", skipped)
	}

	// All intermediate dirs and the category dir itself should be gone.
	if _, err := os.Stat(categoryDir); !os.IsNotExist(err) {
		t.Errorf("expected category dir to be removed, but it still exists")
	}
}

func TestRemoveManagedLeafSymlinks_CleanupPartial(t *testing.T) {
	categoryDir := t.TempDir()
	skillSrc := t.TempDir()

	// One managed leaf and one regular file — the regular file prevents full cleanup.
	makeSymlink(t, skillSrc, filepath.Join(categoryDir, "sub", "skill"))
	if err := os.WriteFile(filepath.Join(categoryDir, "README.md"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}

	f := formatter.New()
	removed, _ := removeManagedLeafSymlinks(f, categoryDir, "", true, "test-target", "skills", false, nil)

	if removed != 1 {
		t.Errorf("removed = %d, want 1", removed)
	}

	// categoryDir must still exist because README.md is there.
	if _, err := os.Stat(categoryDir); err != nil {
		t.Errorf("categoryDir should still exist: %v", err)
	}
}

func TestUnlinkAllComponents_DeeplyNestedSkill(t *testing.T) {
	agentsDir := t.TempDir()
	targetDir := t.TempDir()

	// Source skill.
	skillSrc := filepath.Join(agentsDir, "skills", "sdlc-pipeline", "test", "test2", "test3", "test4", "test5")
	if err := os.MkdirAll(skillSrc, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillSrc, "SKILL.md"), []byte("# skill"), 0644); err != nil {
		t.Fatal(err)
	}

	// Simulate what `link all` creates: real category dirs + leaf symlink at the bottom.
	leafDst := filepath.Join(targetDir, "skills", "sdlc-pipeline", "test", "test2", "test3", "test4", "test5")
	if err := os.MkdirAll(filepath.Dir(leafDst), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(skillSrc, leafDst); err != nil {
		t.Fatal(err)
	}

	target := &stubTarget{baseDir: targetDir}
	f := formatter.New()

	if err := UnlinkAllComponents(agentsDir, []config.Target{target}, f, "all", true, false); err != nil {
		t.Fatalf("UnlinkAllComponents: %v", err)
	}

	// The leaf symlink and all empty intermediate dirs should be gone.
	if _, err := os.Lstat(leafDst); !os.IsNotExist(err) {
		t.Errorf("leaf symlink should be removed")
	}

	categoryRoot := filepath.Join(targetDir, "skills", "sdlc-pipeline")
	if _, err := os.Lstat(categoryRoot); !os.IsNotExist(err) {
		t.Errorf("empty category root should be removed")
	}
}
