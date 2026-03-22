package linkerDisplay

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/pkg/config"
)

// stubTarget implements config.Target backed by a temp directory.
type stubTarget struct {
	baseDir string
}

func (s *stubTarget) GetGlobalBaseDir() (string, error) { return s.baseDir, nil }
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
	return filepath.Join(s.baseDir, "config.json"), nil
}
func (s *stubTarget) GetName() string           { return "stub" }
func (s *stubTarget) GetDisplayName() string    { return "Stub" }
func (s *stubTarget) GetProjectDirName() string { return ".stub" }
func (s *stubTarget) GetProjectBaseDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".stub")
}
func (s *stubTarget) GetProjectComponentDir(projectRoot, componentType string) (string, error) {
	return filepath.Join(projectRoot, ".stub", componentType), nil
}
func (s *stubTarget) IsUniversalTarget() bool { return false }

var _ config.Target = (*stubTarget)(nil)

// makeSkill creates a skill directory with a SKILL.md at the given path.
func makeSkill(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("makeSkill MkdirAll %s: %v", path, err)
	}
	if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte("# skill"), 0644); err != nil {
		t.Fatalf("makeSkill WriteFile: %v", err)
	}
}

// TestCollectLeafSkills_Flat verifies that flat (non-nested) skills are collected correctly.
func TestCollectLeafSkills_Flat(t *testing.T) {
	tmp := t.TempDir()
	skillsDir := filepath.Join(tmp, "skills")

	makeSkill(t, filepath.Join(skillsDir, "alpha"))
	makeSkill(t, filepath.Join(skillsDir, "beta"))

	got := collectLeafSkills(skillsDir, "", tmp, "myprofile", "skills")

	if len(got) != 2 {
		t.Fatalf("expected 2 skills, got %d: %v", len(got), got)
	}
	names := map[string]bool{}
	for _, c := range got {
		names[c.Name] = true
	}
	if !names["alpha"] || !names["beta"] {
		t.Errorf("unexpected skill names: %v", names)
	}
}

// TestCollectLeafSkills_Nested verifies that skills nested under a category directory
// are collected with their full relative path as Name.
func TestCollectLeafSkills_Nested(t *testing.T) {
	tmp := t.TempDir()
	skillsDir := filepath.Join(tmp, "skills")

	makeSkill(t, filepath.Join(skillsDir, "sdlc-pipeline", "review-architecture"))
	makeSkill(t, filepath.Join(skillsDir, "sdlc-pipeline", "draft-architecture"))
	makeSkill(t, filepath.Join(skillsDir, "planning", "prd"))

	got := collectLeafSkills(skillsDir, "", tmp, "myprofile", "skills")

	if len(got) != 3 {
		t.Fatalf("expected 3 skills, got %d: %v", len(got), got)
	}
	names := map[string]bool{}
	for _, c := range got {
		names[c.Name] = true
	}

	wantNames := []string{
		filepath.Join("sdlc-pipeline", "review-architecture"),
		filepath.Join("sdlc-pipeline", "draft-architecture"),
		filepath.Join("planning", "prd"),
	}
	for _, want := range wantNames {
		if !names[want] {
			t.Errorf("expected skill name %q, have: %v", want, names)
		}
	}
}

// TestCollectLeafSkills_Mixed verifies that flat and nested skills coexist correctly.
func TestCollectLeafSkills_Mixed(t *testing.T) {
	tmp := t.TempDir()
	skillsDir := filepath.Join(tmp, "skills")

	makeSkill(t, filepath.Join(skillsDir, "flat-skill"))
	makeSkill(t, filepath.Join(skillsDir, "category", "nested-skill"))

	got := collectLeafSkills(skillsDir, "", tmp, "myprofile", "skills")

	if len(got) != 2 {
		t.Fatalf("expected 2 skills, got %d: %v", len(got), got)
	}
	names := map[string]bool{}
	for _, c := range got {
		names[c.Name] = true
	}
	if !names["flat-skill"] {
		t.Errorf("expected flat-skill in names: %v", names)
	}
	if !names[filepath.Join("category", "nested-skill")] {
		t.Errorf("expected category/nested-skill in names: %v", names)
	}
}

// TestShowLinkStatus_LeafSymlinkLayout verifies that ShowLinkStatus reports nested
// skills as linked when the target has a real category directory containing a leaf symlink
// (the layout produced by `link skill sdlc-pipeline/record-completion`).
func TestShowLinkStatus_LeafSymlinkLayout(t *testing.T) {
	tmp := t.TempDir()

	agentsDir := filepath.Join(tmp, ".agent-smith")
	makeSkill(t, filepath.Join(agentsDir, "skills", "sdlc-pipeline", "record-completion"))

	targetSkillsDir := filepath.Join(tmp, ".stub", "skills")
	if err := os.MkdirAll(targetSkillsDir, 0755); err != nil {
		t.Fatalf("MkdirAll target: %v", err)
	}

	// Real category dir + leaf symlink (what `link skill sdlc-pipeline/record-completion` creates).
	categoryDir := filepath.Join(targetSkillsDir, "sdlc-pipeline")
	if err := os.MkdirAll(categoryDir, 0755); err != nil {
		t.Fatalf("MkdirAll category: %v", err)
	}
	leafSrc := filepath.Join(agentsDir, "skills", "sdlc-pipeline", "record-completion")
	leafDst := filepath.Join(categoryDir, "record-completion")
	if err := os.Symlink(leafSrc, leafDst); err != nil {
		t.Fatalf("Symlink leaf: %v", err)
	}

	target := &stubTarget{baseDir: filepath.Join(tmp, ".stub")}
	targets := []config.Target{target}

	var buf bytes.Buffer
	f := formatter.NewWithWriter(&buf)

	if err := ShowLinkStatus(agentsDir, targets, f, false); err != nil {
		t.Fatalf("ShowLinkStatus: %v", err)
	}

	output := buf.String()
	t.Logf("output:\n%s", output)

	if !strings.Contains(output, "record-completion") {
		t.Errorf("expected record-completion in output:\n%s", output)
	}

	legendIdx := strings.Index(output, "--- Legend ---")
	dataSection := output
	if legendIdx >= 0 {
		dataSection = output[:legendIdx]
	}

	if !strings.Contains(dataSection, "✓") {
		t.Errorf("expected ✓ (linked) symbol in data rows:\n%s", dataSection)
	}
}

// TestShowLinkStatus_NestedSkillsShowAsLinked verifies that ShowLinkStatus reports nested
// skills as linked when the category-level symlink in the target points to the correct source.
func TestShowLinkStatus_NestedSkillsShowAsLinked(t *testing.T) {
	tmp := t.TempDir()

	// Source: ~/.agent-smith/skills/sdlc-pipeline/{review-architecture,draft-architecture}
	agentsDir := filepath.Join(tmp, ".agent-smith")
	makeSkill(t, filepath.Join(agentsDir, "skills", "sdlc-pipeline", "review-architecture"))
	makeSkill(t, filepath.Join(agentsDir, "skills", "sdlc-pipeline", "draft-architecture"))
	makeSkill(t, filepath.Join(agentsDir, "skills", "flat-skill"))

	// Target: ~/.stub/skills/sdlc-pipeline → symlink to source sdlc-pipeline dir
	targetSkillsDir := filepath.Join(tmp, ".stub", "skills")
	if err := os.MkdirAll(targetSkillsDir, 0755); err != nil {
		t.Fatalf("MkdirAll target: %v", err)
	}

	sdlcSource := filepath.Join(agentsDir, "skills", "sdlc-pipeline")
	sdlcLink := filepath.Join(targetSkillsDir, "sdlc-pipeline")
	if err := os.Symlink(sdlcSource, sdlcLink); err != nil {
		t.Fatalf("Symlink sdlc-pipeline: %v", err)
	}

	flatSource := filepath.Join(agentsDir, "skills", "flat-skill")
	flatLink := filepath.Join(targetSkillsDir, "flat-skill")
	if err := os.Symlink(flatSource, flatLink); err != nil {
		t.Fatalf("Symlink flat-skill: %v", err)
	}

	target := &stubTarget{baseDir: filepath.Join(tmp, ".stub")}
	targets := []config.Target{target}

	var buf bytes.Buffer
	f := formatter.NewWithWriter(&buf)

	if err := ShowLinkStatus(agentsDir, targets, f, false); err != nil {
		t.Fatalf("ShowLinkStatus: %v", err)
	}

	output := buf.String()
	t.Logf("output:\n%s", output)

	// All three individual skills must appear.
	wantNames := []string{
		"review-architecture",
		"draft-architecture",
		"flat-skill",
	}
	for _, name := range wantNames {
		if !strings.Contains(output, name) {
			t.Errorf("expected %q in output:\n%s", name, output)
		}
	}

	// The category folder "sdlc-pipeline" must NOT appear as its own row.
	// It should only appear as a prefix within the skill names.
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "sdlc-pipeline" {
			t.Errorf("sdlc-pipeline should not appear as a standalone row, but found: %q", line)
		}
	}
}

// TestShowLinkStatus_NestedSkillsShowAsUnlinked verifies that nested skills that have no
// corresponding symlink in the target are reported as unlinked.
func TestShowLinkStatus_NestedSkillsShowAsUnlinked(t *testing.T) {
	tmp := t.TempDir()

	agentsDir := filepath.Join(tmp, ".agent-smith")
	makeSkill(t, filepath.Join(agentsDir, "skills", "sdlc-pipeline", "review-architecture"))

	// Target exists but has no sdlc-pipeline symlink.
	targetSkillsDir := filepath.Join(tmp, ".stub", "skills")
	if err := os.MkdirAll(targetSkillsDir, 0755); err != nil {
		t.Fatalf("MkdirAll target: %v", err)
	}

	target := &stubTarget{baseDir: filepath.Join(tmp, ".stub")}
	targets := []config.Target{target}

	var buf bytes.Buffer
	f := formatter.NewWithWriter(&buf)

	if err := ShowLinkStatus(agentsDir, targets, f, false); err != nil {
		t.Fatalf("ShowLinkStatus: %v", err)
	}

	output := buf.String()
	t.Logf("output:\n%s", output)

	if !strings.Contains(output, "review-architecture") {
		t.Errorf("expected review-architecture in output:\n%s", output)
	}

	// The skill row should show as unlinked ("-"). Extract the data rows only
	// (before the legend section) to avoid false positives from legend symbols.
	legendIdx := strings.Index(output, "--- Legend ---")
	dataSection := output
	if legendIdx >= 0 {
		dataSection = output[:legendIdx]
	}

	if strings.Contains(dataSection, "✓") {
		t.Errorf("expected no ✓ (linked) symbol in data rows when skill is unlinked:\n%s", dataSection)
	}
	if !strings.Contains(dataSection, "-") {
		t.Errorf("expected - (unlinked) symbol in data rows:\n%s", dataSection)
	}
}
