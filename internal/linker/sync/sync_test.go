package linkerSync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/pkg/config"
)

// stubTarget implements config.Target for testing.
type stubTarget struct {
	baseDir string
}

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

// writeLockFile writes a minimal .component-lock.json for the given skills
// into baseDir. Each skill maps short name → filesystemName.
func writeLockFile(t *testing.T, baseDir string, skills map[string]string) {
	t.Helper()

	type entry struct {
		Source         string `json:"source"`
		SourceType     string `json:"sourceType"`
		SourceUrl      string `json:"sourceUrl"`
		CommitHash     string `json:"commitHash"`
		FilesystemName string `json:"filesystemName,omitempty"`
		Components     int    `json:"components"`
		Detection      string `json:"detection"`
		Version        int    `json:"version"`
	}

	type lockFile struct {
		Version int                         `json:"version"`
		Skills  map[string]map[string]entry `json:"skills"`
	}

	source := "git@github.com:test/skills"
	sourceSkills := make(map[string]entry, len(skills))
	for name, fsName := range skills {
		sourceSkills[name] = entry{
			Source:         source,
			SourceType:     "git",
			SourceUrl:      source,
			CommitHash:     "abc123",
			FilesystemName: fsName,
			Components:     1,
			Detection:      "single",
			Version:        1,
		}
	}

	lf := lockFile{
		Version: 1,
		Skills:  map[string]map[string]entry{source: sourceSkills},
	}

	data, err := json.Marshal(lf)
	if err != nil {
		t.Fatalf("marshal lock file: %v", err)
	}

	lockPath := filepath.Join(baseDir, ".component-lock.json")
	if err := os.WriteFile(lockPath, data, 0644); err != nil {
		t.Fatalf("write lock file: %v", err)
	}
}

// TestLinkComponent_NestedSkillCreatesLeafSymlink verifies that
// LinkComponent("skills", "sdlc-pipeline/record-completion") creates a real
// category directory and a leaf symlink inside it, preserving hierarchy.
func TestLinkComponent_NestedSkillCreatesLeafSymlink(t *testing.T) {
	agentsDir := t.TempDir()
	targetDir := t.TempDir()

	skillDir := filepath.Join(agentsDir, "skills", "sdlc-pipeline", "record-completion")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# skill"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(targetDir, "skills"), 0755); err != nil {
		t.Fatal(err)
	}

	target := &stubTarget{baseDir: targetDir}
	f := formatter.New()

	if err := LinkComponent(agentsDir, []config.Target{target}, f, "skills", "sdlc-pipeline/record-completion"); err != nil {
		t.Fatalf("LinkComponent: %v", err)
	}

	// Category dir must be a real directory, not a symlink.
	categoryPath := filepath.Join(targetDir, "skills", "sdlc-pipeline")
	info, err := os.Lstat(categoryPath)
	if err != nil {
		t.Fatalf("expected category dir at %s: %v", categoryPath, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("%s should be a real directory, got symlink", categoryPath)
	}
	if !info.IsDir() {
		t.Fatalf("%s should be a directory", categoryPath)
	}

	// Leaf symlink must exist inside the category dir.
	leafPath := filepath.Join(categoryPath, "record-completion")
	leafInfo, err := os.Lstat(leafPath)
	if err != nil {
		t.Fatalf("expected leaf symlink at %s: %v", leafPath, err)
	}
	if leafInfo.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%s should be a symlink", leafPath)
	}

	// Leaf symlink must resolve to the source skill directory.
	resolved, err := filepath.EvalSymlinks(leafPath)
	if err != nil {
		t.Fatalf("EvalSymlinks %s: %v", leafPath, err)
	}
	expectedResolved, err := filepath.EvalSymlinks(skillDir)
	if err != nil {
		t.Fatalf("EvalSymlinks expected %s: %v", skillDir, err)
	}
	if resolved != expectedResolved {
		t.Errorf("leaf symlink resolves to %q, want %q", resolved, expectedResolved)
	}
}

// TestLinkComponent_NestedSkillIdempotent verifies that running LinkComponent
// twice for the same nested skill does not error (idempotent).
func TestLinkComponent_NestedSkillIdempotent(t *testing.T) {
	agentsDir := t.TempDir()
	targetDir := t.TempDir()

	skillDir := filepath.Join(agentsDir, "skills", "sdlc-pipeline", "record-completion")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# skill"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(targetDir, "skills"), 0755); err != nil {
		t.Fatal(err)
	}

	target := &stubTarget{baseDir: targetDir}
	f := formatter.New()

	for i := range 2 {
		if err := LinkComponent(agentsDir, []config.Target{target}, f, "skills", "sdlc-pipeline/record-completion"); err != nil {
			t.Fatalf("LinkComponent call %d: %v", i+1, err)
		}
	}

	leafPath := filepath.Join(targetDir, "skills", "sdlc-pipeline", "record-completion")
	if _, err := os.Lstat(leafPath); err != nil {
		t.Fatalf("leaf symlink missing after second call: %v", err)
	}
}

// TestLinkComponent_ShortNameCreatesLeafSymlink verifies that a short lock-file
// key like "record-completion" resolves to the full filesystem path via the lock
// file and creates a category dir + leaf symlink.
func TestLinkComponent_ShortNameCreatesLeafSymlink(t *testing.T) {
	agentsDir := t.TempDir()
	targetDir := t.TempDir()

	skillDir := filepath.Join(agentsDir, "skills", "sdlc-pipeline", "record-completion")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# skill"), 0644); err != nil {
		t.Fatal(err)
	}

	writeLockFile(t, agentsDir, map[string]string{
		"record-completion": "sdlc-pipeline/record-completion",
	})

	if err := os.MkdirAll(filepath.Join(targetDir, "skills"), 0755); err != nil {
		t.Fatal(err)
	}

	target := &stubTarget{baseDir: targetDir}
	f := formatter.New()

	if err := LinkComponent(agentsDir, []config.Target{target}, f, "skills", "record-completion"); err != nil {
		t.Fatalf("LinkComponent: %v", err)
	}

	categoryPath := filepath.Join(targetDir, "skills", "sdlc-pipeline")
	info, err := os.Lstat(categoryPath)
	if err != nil {
		t.Fatalf("expected category dir at %s: %v", categoryPath, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("%s should be a real directory, not a symlink", categoryPath)
	}

	leafPath := filepath.Join(categoryPath, "record-completion")
	leafInfo, err := os.Lstat(leafPath)
	if err != nil {
		t.Fatalf("expected leaf symlink at %s: %v", leafPath, err)
	}
	if leafInfo.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%s should be a symlink", leafPath)
	}
}

// TestLinkAll_RespectsCategoryRealDir verifies that LinkAllComponents does not
// replace an existing real category directory (from previous leaf-level linking)
// with a category-level symlink.
func TestLinkAll_RespectsCategoryRealDir(t *testing.T) {
	agentsDir := t.TempDir()
	targetDir := t.TempDir()

	// Source: category with one skill.
	skillDir := filepath.Join(agentsDir, "skills", "sdlc-pipeline", "record-completion")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# skill"), 0644); err != nil {
		t.Fatal(err)
	}

	// Pre-create real category dir + leaf symlink (simulating previous `link skill`).
	categoryDst := filepath.Join(targetDir, "skills", "sdlc-pipeline")
	if err := os.MkdirAll(categoryDst, 0755); err != nil {
		t.Fatal(err)
	}
	leafDst := filepath.Join(categoryDst, "record-completion")
	if err := os.Symlink(skillDir, leafDst); err != nil {
		t.Fatal(err)
	}

	target := &stubTarget{baseDir: targetDir}
	f := formatter.New()

	if err := LinkAllComponents(agentsDir, []config.Target{target}, f); err != nil {
		t.Fatalf("LinkAllComponents: %v", err)
	}

	// Category path must still be a real directory, not replaced by a symlink.
	info, err := os.Lstat(categoryDst)
	if err != nil {
		t.Fatalf("category path missing: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("%s was replaced by a symlink; real dir should be preserved", categoryDst)
	}
	if !info.IsDir() {
		t.Fatalf("%s should still be a directory", categoryDst)
	}
}

// TestLinkAll_LeafSkillsLinkedIndividually verifies that LinkAllComponents links
// each leaf skill in a nested structure individually (not by category symlink).
func TestLinkAll_LeafSkillsLinkedIndividually(t *testing.T) {
	agentsDir := t.TempDir()
	targetDir := t.TempDir()

	skills := []string{
		"sdlc-pipeline/record-completion",
		"sdlc-pipeline/draft-architecture",
		"planning/prd",
	}
	for _, s := range skills {
		dir := filepath.Join(agentsDir, "skills", s)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# skill"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	target := &stubTarget{baseDir: targetDir}
	f := formatter.New()

	if err := LinkAllComponents(agentsDir, []config.Target{target}, f); err != nil {
		t.Fatalf("LinkAllComponents: %v", err)
	}

	for _, s := range skills {
		leafPath := filepath.Join(targetDir, "skills", s)
		info, err := os.Lstat(leafPath)
		if err != nil {
			t.Errorf("expected leaf at %s: %v", leafPath, err)
			continue
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("%s should be a symlink", leafPath)
		}
	}

	// Category dirs must be real directories, not symlinks.
	for _, cat := range []string{"sdlc-pipeline", "planning"} {
		catPath := filepath.Join(targetDir, "skills", cat)
		info, err := os.Lstat(catPath)
		if err != nil {
			t.Errorf("expected category dir at %s: %v", catPath, err)
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			t.Errorf("%s should be a real directory, not a symlink", catPath)
		}
	}
}

// TestCollectLeafSkillNames verifies the helper returns only leaf skill paths.
func TestCollectLeafSkillNames(t *testing.T) {
	typeDir := t.TempDir()

	mkSkill := func(relPath string) {
		t.Helper()
		dir := filepath.Join(typeDir, relPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# skill"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	mkSkill("sdlc-pipeline/record-completion")
	mkSkill("sdlc-pipeline/draft-architecture")
	mkSkill("planning/prd")
	mkSkill("flat-skill")

	names, err := collectLeafSkillNames(typeDir, "")
	if err != nil {
		t.Fatalf("collectLeafSkillNames: %v", err)
	}

	want := map[string]bool{
		"sdlc-pipeline/record-completion":  true,
		"sdlc-pipeline/draft-architecture": true,
		"planning/prd":                     true,
		"flat-skill":                       true,
	}

	if len(names) != len(want) {
		t.Errorf("got %d names, want %d: %v", len(names), len(want), names)
	}
	for _, n := range names {
		if !want[n] {
			t.Errorf("unexpected name: %q", n)
		}
	}
}
