package downloader

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tjg184/agent-smith/internal/detector"
	"github.com/tjg184/agent-smith/internal/models"
)

// countLockEntries reads the lock file at baseDir and returns the total number of entries
// stored under the given componentType across all source URLs.
func countLockEntries(t *testing.T, baseDir, componentType string) int {
	t.Helper()
	lockPath := filepath.Join(baseDir, ".component-lock.json")
	data, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("failed to read lock file: %v", err)
	}
	var lf models.ComponentLockFile
	if err := json.Unmarshal(data, &lf); err != nil {
		t.Fatalf("failed to unmarshal lock file: %v", err)
	}
	var m map[string]map[string]models.ComponentEntry
	switch componentType {
	case "skills":
		m = lf.Skills
	case "agents":
		m = lf.Agents
	case "commands":
		m = lf.Commands
	}
	total := 0
	for _, entries := range m {
		total += len(entries)
	}
	return total
}

// countDirsInDir counts immediate child directories inside dir.
func countDirsInDir(t *testing.T, dir string) int {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read dir %s: %v", dir, err)
	}
	n := 0
	for _, e := range entries {
		if e.IsDir() {
			n++
		}
	}
	return n
}

// createSubdirSkillRepo builds a local git repo that mirrors nicobailon/visual-explainer:
// a single skill living at plugins/<skillName>/SKILL.md.
func createSubdirSkillRepo(h *TestHelper, skillName string) string {
	return h.CreateMockRepo("subdir-skill-repo", map[string]string{
		"plugins/" + skillName + "/SKILL.md": "---\nname: " + skillName + "\n---\n# " + skillName,
	})
}

// createSubdirAgentRepo builds a repo with an agent nested under plugins/<agentName>/<agentName>.md.
func createSubdirAgentRepo(h *TestHelper, agentName string) string {
	return h.CreateMockRepo("subdir-agent-repo", map[string]string{
		"agents/plugins/" + agentName + ".md": "---\nname: " + agentName + "\n---\n# " + agentName,
	})
}

// createSubdirCommandRepo builds a repo with a command nested under plugins/<commandName>/<commandName>.md.
func createSubdirCommandRepo(h *TestHelper, commandName string) string {
	return h.CreateMockRepo("subdir-command-repo", map[string]string{
		"commands/plugins/" + commandName + ".md": "---\nname: " + commandName + "\n---\n# " + commandName,
	})
}

// ---------------------------------------------------------------------------
// TestInstallAllThenInstallSkillNoDuplicate
//
// Regression for: running "install all" followed by "install skill … plugins/X"
// used to produce two copies of the same skill on disk and two lock entries.
// ---------------------------------------------------------------------------

func TestInstallAllThenInstallSkillNoDuplicate(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	const skillName = "visual-explainer"
	repoPath := createSubdirSkillRepo(h, skillName)
	installDir := h.CreateInstallDir()
	skillsDir := filepath.Join(installDir, "skills")

	det := detector.NewRepositoryDetector()
	components, err := det.DetectComponentsInRepo(repoPath)
	if err != nil {
		t.Fatalf("detection failed: %v", err)
	}

	dl, err := ForTypeWithTargetDir(models.ComponentSkill, installDir)
	if err != nil {
		t.Fatalf("failed to create downloader: %v", err)
	}

	// Simulate "install all": installs every detected component via DownloadWithRepo.
	for _, comp := range components {
		if comp.Type != models.ComponentSkill {
			continue
		}
		if err := dl.DownloadWithRepo(repoPath, comp.Name, repoPath, repoPath, components); err != nil {
			t.Fatalf("DownloadWithRepo(%q) failed: %v", comp.Name, err)
		}
	}

	// Simulate "install skill … plugins/visual-explainer" (path-style selector).
	if err := dl.Download(repoPath, "plugins/"+skillName, repoPath); err != nil {
		t.Fatalf("Download with path-style name failed: %v", err)
	}

	// Exactly one directory must exist under skills/.
	if got := countDirsInDir(t, skillsDir); got != 1 {
		t.Errorf("expected 1 skill directory after install-all + install-skill, got %d", got)
	}

	// Exactly one lock entry must exist.
	if got := countLockEntries(t, installDir, "skills"); got != 1 {
		t.Errorf("expected 1 lock entry after install-all + install-skill, got %d", got)
	}
}

func TestInstallAllThenInstallAgentNoDuplicate(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	const agentName = "my-agent"
	repoPath := createSubdirAgentRepo(h, agentName)
	installDir := h.CreateInstallDir()
	agentsDir := filepath.Join(installDir, "agents")

	det := detector.NewRepositoryDetector()
	components, err := det.DetectComponentsInRepo(repoPath)
	if err != nil {
		t.Fatalf("detection failed: %v", err)
	}

	dl, err := ForTypeWithTargetDir(models.ComponentAgent, installDir)
	if err != nil {
		t.Fatalf("failed to create downloader: %v", err)
	}

	for _, comp := range components {
		if comp.Type != models.ComponentAgent {
			continue
		}
		if err := dl.DownloadWithRepo(repoPath, comp.Name, repoPath, repoPath, components); err != nil {
			t.Fatalf("DownloadWithRepo(%q) failed: %v", comp.Name, err)
		}
	}

	// Path-style selector: "plugins/my-agent"
	if err := dl.Download(repoPath, "plugins/"+agentName, repoPath); err != nil {
		t.Fatalf("Download with path-style name failed: %v", err)
	}

	if got := countDirsInDir(t, agentsDir); got != 1 {
		t.Errorf("expected 1 agent directory, got %d", got)
	}
	if got := countLockEntries(t, installDir, "agents"); got != 1 {
		t.Errorf("expected 1 lock entry, got %d", got)
	}
}

func TestInstallAllThenInstallCommandNoDuplicate(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	const commandName = "my-command"
	repoPath := createSubdirCommandRepo(h, commandName)
	installDir := h.CreateInstallDir()
	commandsDir := filepath.Join(installDir, "commands")

	det := detector.NewRepositoryDetector()
	components, err := det.DetectComponentsInRepo(repoPath)
	if err != nil {
		t.Fatalf("detection failed: %v", err)
	}

	dl, err := ForTypeWithTargetDir(models.ComponentCommand, installDir)
	if err != nil {
		t.Fatalf("failed to create downloader: %v", err)
	}

	for _, comp := range components {
		if comp.Type != models.ComponentCommand {
			continue
		}
		if err := dl.DownloadWithRepo(repoPath, comp.Name, repoPath, repoPath, components); err != nil {
			t.Fatalf("DownloadWithRepo(%q) failed: %v", comp.Name, err)
		}
	}

	if err := dl.Download(repoPath, "plugins/"+commandName, repoPath); err != nil {
		t.Fatalf("Download with path-style name failed: %v", err)
	}

	if got := countDirsInDir(t, commandsDir); got != 1 {
		t.Errorf("expected 1 command directory, got %d", got)
	}
	if got := countLockEntries(t, installDir, "commands"); got != 1 {
		t.Errorf("expected 1 lock entry, got %d", got)
	}
}

// ---------------------------------------------------------------------------
// TestFindComponentByName
//
// Unit tests for findComponentByName: both short-name and path-style selectors
// must resolve to the correct component; unknown names must return nil.
// ---------------------------------------------------------------------------

func TestFindComponentByName(t *testing.T) {
	components := []models.DetectedComponent{
		{Type: models.ComponentSkill, Name: "visual-explainer", FilePath: "plugins/visual-explainer/SKILL.md"},
		{Type: models.ComponentSkill, Name: "another-skill", FilePath: "skills/another-skill/SKILL.md"},
		{Type: models.ComponentAgent, Name: "my-agent", FilePath: "agents/plugins/my-agent.md"},
	}

	skills := func() []models.DetectedComponent {
		var out []models.DetectedComponent
		for _, c := range components {
			if c.Type == models.ComponentSkill {
				out = append(out, c)
			}
		}
		return out
	}()

	tests := []struct {
		name       string
		selector   string
		candidates []models.DetectedComponent
		wantName   string
		wantNil    bool
	}{
		{
			name:       "short name matches canonical",
			selector:   "visual-explainer",
			candidates: skills,
			wantName:   "visual-explainer",
		},
		{
			name:       "path-style name resolves via DetermineDestinationFolderName",
			selector:   "plugins/visual-explainer",
			candidates: skills,
			wantName:   "visual-explainer",
		},
		{
			name:       "flat skill short name",
			selector:   "another-skill",
			candidates: skills,
			wantName:   "another-skill",
		},
		{
			name:       "unknown name returns nil",
			selector:   "does-not-exist",
			candidates: skills,
			wantNil:    true,
		},
		{
			name:       "empty candidates returns nil",
			selector:   "visual-explainer",
			candidates: nil,
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findComponentByName(tt.candidates, tt.selector)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got component with Name=%q", got.Name)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected component with Name=%q, got nil", tt.wantName)
			}
			if got.Name != tt.wantName {
				t.Errorf("Name: got %q, want %q", got.Name, tt.wantName)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestDownloadWithRepoIdempotency
//
// Calling DownloadWithRepo twice for the same component (same repo, same name)
// must result in exactly one directory on disk and one lock entry — not two.
// ---------------------------------------------------------------------------

func TestDownloadWithRepoIdempotency(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	const skillName = "visual-explainer"
	repoPath := createSubdirSkillRepo(h, skillName)
	installDir := h.CreateInstallDir()
	skillsDir := filepath.Join(installDir, "skills")

	det := detector.NewRepositoryDetector()
	components, err := det.DetectComponentsInRepo(repoPath)
	if err != nil {
		t.Fatalf("detection failed: %v", err)
	}

	dl, err := ForTypeWithTargetDir(models.ComponentSkill, installDir)
	if err != nil {
		t.Fatalf("failed to create downloader: %v", err)
	}

	var skillComp *models.DetectedComponent
	for i := range components {
		if components[i].Type == models.ComponentSkill {
			skillComp = &components[i]
			break
		}
	}
	if skillComp == nil {
		t.Fatal("no skill component detected in repo")
	}

	// First install.
	if err := dl.DownloadWithRepo(repoPath, skillComp.Name, repoPath, repoPath, components); err != nil {
		t.Fatalf("first DownloadWithRepo failed: %v", err)
	}

	// Second install (idempotent re-run of "install all").
	if err := dl.DownloadWithRepo(repoPath, skillComp.Name, repoPath, repoPath, components); err != nil {
		t.Fatalf("second DownloadWithRepo failed: %v", err)
	}

	if got := countDirsInDir(t, skillsDir); got != 1 {
		t.Errorf("expected 1 skill directory after two DownloadWithRepo calls, got %d", got)
	}
	if got := countLockEntries(t, installDir, "skills"); got != 1 {
		t.Errorf("expected 1 lock entry after two DownloadWithRepo calls, got %d", got)
	}
}

// ---------------------------------------------------------------------------
// TestInstallSkillShortNameAfterInstallAll
//
// After "install all", calling "install skill" with the short canonical name
// (not the path-style form) must also be idempotent.
// ---------------------------------------------------------------------------

func TestInstallSkillShortNameAfterInstallAll(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	const skillName = "visual-explainer"
	repoPath := createSubdirSkillRepo(h, skillName)
	installDir := h.CreateInstallDir()
	skillsDir := filepath.Join(installDir, "skills")

	det := detector.NewRepositoryDetector()
	components, err := det.DetectComponentsInRepo(repoPath)
	if err != nil {
		t.Fatalf("detection failed: %v", err)
	}

	dl, err := ForTypeWithTargetDir(models.ComponentSkill, installDir)
	if err != nil {
		t.Fatalf("failed to create downloader: %v", err)
	}

	for _, comp := range components {
		if comp.Type != models.ComponentSkill {
			continue
		}
		if err := dl.DownloadWithRepo(repoPath, comp.Name, repoPath, repoPath, components); err != nil {
			t.Fatalf("DownloadWithRepo failed: %v", err)
		}
	}

	// "install skill … visual-explainer" (short canonical name).
	if err := dl.Download(repoPath, skillName, repoPath); err != nil {
		t.Fatalf("Download with short name failed: %v", err)
	}

	if got := countDirsInDir(t, skillsDir); got != 1 {
		t.Errorf("expected 1 skill directory, got %d", got)
	}
	if got := countLockEntries(t, installDir, "skills"); got != 1 {
		t.Errorf("expected 1 lock entry, got %d", got)
	}
}

// ---------------------------------------------------------------------------
// TestLockKeyUsesCanonicalName
//
// Regardless of whether the user installs via short name or path-style name,
// the lock file must store the entry under the canonical detected component
// name (e.g. "visual-explainer"), not the raw user-supplied selector.
// ---------------------------------------------------------------------------

func TestLockKeyUsesCanonicalName(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	const skillName = "visual-explainer"
	repoPath := createSubdirSkillRepo(h, skillName)
	installDir := h.CreateInstallDir()

	dl, err := ForTypeWithTargetDir(models.ComponentSkill, installDir)
	if err != nil {
		t.Fatalf("failed to create downloader: %v", err)
	}

	// Install via path-style selector.
	if err := dl.Download(repoPath, "plugins/"+skillName, repoPath); err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	lockPath := filepath.Join(installDir, ".component-lock.json")
	data, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("failed to read lock file: %v", err)
	}
	var lf models.ComponentLockFile
	if err := json.Unmarshal(data, &lf); err != nil {
		t.Fatalf("failed to unmarshal lock file: %v", err)
	}

	for _, entries := range lf.Skills {
		if _, ok := entries[skillName]; ok {
			return // canonical name found — pass
		}
		// Fail if the path-style name leaked into the lock key.
		if _, bad := entries["plugins/"+skillName]; bad {
			t.Errorf("lock key should be %q but found path-style key %q", skillName, "plugins/"+skillName)
		}
	}

	// If we reach here without returning, the expected key was not found.
	t.Errorf("lock entry with key %q not found in lock file", skillName)
}
