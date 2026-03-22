package sync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/pkg/services"
)

// Each entry in skills maps shortName → filesystemName.
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
		Version  int                         `json:"version"`
		Skills   map[string]map[string]entry `json:"skills"`
		Agents   map[string]map[string]entry `json:"agents"`
		Commands map[string]map[string]entry `json:"commands"`
	}

	const sourceURL = "git@github.com:test/skills"
	sourceSkills := make(map[string]entry, len(skills))
	for name, fsName := range skills {
		sourceSkills[name] = entry{
			Source:         sourceURL,
			SourceType:     "git",
			SourceUrl:      sourceURL,
			CommitHash:     "abc123",
			FilesystemName: fsName,
			Components:     1,
			Detection:      "single",
			Version:        1,
		}
	}

	lf := lockFile{
		Version:  1,
		Skills:   map[string]map[string]entry{sourceURL: sourceSkills},
		Agents:   map[string]map[string]entry{},
		Commands: map[string]map[string]entry{},
	}

	data, err := json.Marshal(lf)
	if err != nil {
		t.Fatalf("marshal lock file: %v", err)
	}

	if err := os.WriteFile(filepath.Join(baseDir, ".component-lock.json"), data, 0644); err != nil {
		t.Fatalf("write lock file: %v", err)
	}
}

func TestResolveNestedComponentName(t *testing.T) {
	baseDir := t.TempDir()
	writeLockFile(t, baseDir, map[string]string{
		"record-completion":  "sdlc-pipeline/record-completion",
		"draft-architecture": "sdlc-pipeline/draft-architecture",
		"prd":                "planning/prd",
	})

	cases := []struct {
		filesystemName string
		wantName       string
	}{
		{"sdlc-pipeline/record-completion", "record-completion"},
		{"sdlc-pipeline/draft-architecture", "draft-architecture"},
		{"planning/prd", "prd"},
	}

	for _, tc := range cases {
		info, err := resolveNestedComponentName(baseDir, "skills", tc.filesystemName)
		if err != nil {
			t.Errorf("resolveNestedComponentName(%q): %v", tc.filesystemName, err)
			continue
		}
		if info.ComponentName != tc.wantName {
			t.Errorf("resolveNestedComponentName(%q): got %q, want %q", tc.filesystemName, info.ComponentName, tc.wantName)
		}
	}
}

func TestResolveNestedComponentName_NotFound(t *testing.T) {
	baseDir := t.TempDir()
	writeLockFile(t, baseDir, map[string]string{
		"record-completion": "sdlc-pipeline/record-completion",
	})

	_, err := resolveNestedComponentName(baseDir, "skills", "unknown/skill")
	if err == nil {
		t.Fatal("expected error for unknown filesystem name, got nil")
	}
}

// noopPostprocessorRegistry satisfies the PostprocessorRegistry interface for tests.
type noopRegistry struct{}

func (n *noopRegistry) RunPostprocessors(ctx PostprocessContext) error { return nil }
func (n *noopRegistry) RunCleanup(ctx PostprocessContext) error        { return nil }

func TestMaterializeComponent_NestedSkillName(t *testing.T) {
	baseDir := t.TempDir()
	projectDir := t.TempDir()

	// Set up skill source.
	skillSrc := filepath.Join(baseDir, "skills", "sdlc-pipeline", "record-completion")
	if err := os.MkdirAll(skillSrc, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillSrc, "SKILL.md"), []byte("# skill"), 0644); err != nil {
		t.Fatal(err)
	}

	writeLockFile(t, baseDir, map[string]string{
		"record-completion": "sdlc-pipeline/record-completion",
	})

	// Set up a fake target project directory so materialize has somewhere to copy.
	targetProjectDir := filepath.Join(projectDir, ".opencode")
	if err := os.MkdirAll(targetProjectDir, 0755); err != nil {
		t.Fatal(err)
	}

	f := formatter.New()
	deps := Deps{
		Logger:    &testLogger{t},
		Formatter: f,
		Registry:  &noopRegistry{},
		GetSourceDir: func(_ string) (string, string, error) {
			return baseDir, "", nil
		},
	}

	// Register a minimal target that points at projectDir.
	t.Setenv("AGENT_SMITH_TARGET", "opencode")

	opts := services.MaterializeOptions{
		Target:     "opencode",
		ProjectDir: projectDir,
	}

	err := MaterializeComponent(deps, "skills", "sdlc-pipeline/record-completion", opts)
	if err != nil {
		t.Fatalf("MaterializeComponent with nested name: %v", err)
	}

	// Verify the skill was materialized under the nested path.
	dest := filepath.Join(targetProjectDir, "skills", "sdlc-pipeline", "record-completion")
	if _, err := os.Stat(dest); err != nil {
		t.Errorf("expected materialized skill at %s: %v", dest, err)
	}
}

type testLogger struct{ t *testing.T }

func (l *testLogger) Error(format string, args ...interface{}) {
	l.t.Logf("ERROR: "+format, args...)
}
