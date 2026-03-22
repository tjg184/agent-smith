package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/tjg184/agent-smith/internal/models"
	"github.com/tjg184/agent-smith/pkg/services"
)

// writeLockFile creates a minimal .component-lock.json mapping short skill names
// to their filesystemName values inside baseDir.
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

	if err := os.WriteFile(filepath.Join(baseDir, ".component-lock.json"), data, 0644); err != nil {
		t.Fatalf("write lock file: %v", err)
	}
}

// stubLockService implements services.ComponentLockService for use in tests.
// Only FindComponentSources has meaningful behaviour; everything else is a no-op.
type stubLockService struct {
	sources map[string][]string // keyed by short component name
}

var _ services.ComponentLockService = (*stubLockService)(nil)

func (s *stubLockService) FindComponentSources(baseDir, componentType, componentName string) ([]string, error) {
	return s.sources[componentName], nil
}
func (s *stubLockService) LoadEntry(baseDir, componentType, componentName string) (*models.ComponentEntry, error) {
	return nil, nil
}
func (s *stubLockService) LoadEntryBySource(baseDir, componentType, componentName, sourceURL string) (*models.ComponentEntry, error) {
	return nil, nil
}
func (s *stubLockService) GetAllComponentNames(baseDir, componentType string) ([]string, error) {
	return nil, nil
}
func (s *stubLockService) FindAllInstances(baseDir, componentType, componentName string) ([]*models.ComponentEntry, error) {
	return nil, nil
}
func (s *stubLockService) SaveEntry(baseDir, componentType, componentName string, entry *models.ComponentEntry) error {
	return nil
}
func (s *stubLockService) RemoveEntry(baseDir, componentType, componentName string) error {
	return nil
}
func (s *stubLockService) RemoveEntryBySource(baseDir, componentType, componentName, sourceURL string) error {
	return nil
}
func (s *stubLockService) ResolveFilesystemName(baseDir, componentType, desiredName, sourceURL string) (string, error) {
	return desiredName, nil
}
func (s *stubLockService) HasConflict(baseDir, componentType, componentName string) (bool, error) {
	return false, nil
}

// TestGetComponentNames_NestedSkills verifies that GetComponentNames returns the
// full FilesystemName for nested skills (e.g. "sdlc-pipeline/brainstorm-vision").
func TestGetComponentNames_NestedSkills(t *testing.T) {
	basePath := t.TempDir()

	// Create the skills directory so HasSkills is detected.
	skillsDir := filepath.Join(basePath, "skills")
	if err := os.MkdirAll(filepath.Join(skillsDir, "sdlc-pipeline", "brainstorm-vision"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(skillsDir, "planning", "prd"), 0755); err != nil {
		t.Fatal(err)
	}

	writeLockFile(t, basePath, map[string]string{
		"brainstorm-vision": "sdlc-pipeline/brainstorm-vision",
		"prd":               "planning/prd",
	})

	profile := LoadProfile(filepath.Dir(basePath), filepath.Base(basePath))

	_, skills, _ := GetComponentNames(profile)

	sort.Strings(skills)
	expected := []string{"planning/prd", "sdlc-pipeline/brainstorm-vision"}
	if len(skills) != len(expected) {
		t.Fatalf("got skills %v, want %v", skills, expected)
	}
	for i, s := range expected {
		if skills[i] != s {
			t.Errorf("skills[%d] = %q, want %q", i, skills[i], s)
		}
	}
}

// TestCountComponents_NestedSkills verifies that CountComponents counts individual
// skills from the lock file, not category directories on disk.
func TestCountComponents_NestedSkills(t *testing.T) {
	basePath := t.TempDir()

	skillsDir := filepath.Join(basePath, "skills")
	// Two skills in one category dir.
	if err := os.MkdirAll(filepath.Join(skillsDir, "sdlc-pipeline", "brainstorm-vision"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(skillsDir, "sdlc-pipeline", "draft-architecture"), 0755); err != nil {
		t.Fatal(err)
	}

	writeLockFile(t, basePath, map[string]string{
		"brainstorm-vision":  "sdlc-pipeline/brainstorm-vision",
		"draft-architecture": "sdlc-pipeline/draft-architecture",
	})

	profile := LoadProfile(filepath.Dir(basePath), filepath.Base(basePath))

	_, skills, _ := CountComponents(profile)
	if skills != 2 {
		t.Errorf("CountComponents skills = %d, want 2", skills)
	}
}

// TestGetComponentSource_NestedSkill verifies that GetComponentSource resolves the
// source URL correctly when given a full FilesystemName like "sdlc-pipeline/prd".
func TestGetComponentSource_NestedSkill(t *testing.T) {
	basePath := t.TempDir()
	if err := os.MkdirAll(filepath.Join(basePath, "skills"), 0755); err != nil {
		t.Fatal(err)
	}

	profile := &Profile{Name: "test", BasePath: basePath, HasSkills: true}

	svc := &stubLockService{
		sources: map[string][]string{
			"prd": {"git@github.com:test/skills"},
		},
	}

	got := GetComponentSource(profile, svc, "skills", "planning/prd")
	want := "git@github.com:test/skills"
	if got != want {
		t.Errorf("GetComponentSource = %q, want %q", got, want)
	}
}

// TestGetComponentSource_FlatSkill verifies GetComponentSource still works for
// flat (non-nested) skill names.
func TestGetComponentSource_FlatSkill(t *testing.T) {
	basePath := t.TempDir()
	if err := os.MkdirAll(filepath.Join(basePath, "skills"), 0755); err != nil {
		t.Fatal(err)
	}

	profile := &Profile{Name: "test", BasePath: basePath, HasSkills: true}

	svc := &stubLockService{
		sources: map[string][]string{
			"my-skill": {"git@github.com:test/skills"},
		},
	}

	got := GetComponentSource(profile, svc, "skills", "my-skill")
	want := "git@github.com:test/skills"
	if got != want {
		t.Errorf("GetComponentSource = %q, want %q", got, want)
	}
}
