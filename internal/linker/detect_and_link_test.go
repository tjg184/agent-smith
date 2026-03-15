package linker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tjg184/agent-smith/internal/detector"
	"github.com/tjg184/agent-smith/pkg/config"
)

// TestDetectAndLinkLocalRepositories verifies that link auto correctly detects
// and links both skills (SKILL.md) and commands (.md files) from a local repo,
// including the case where component.Path is a file path (commands) rather than
// a directory path (skills).
func TestDetectAndLinkLocalRepositories_LinksCommandsAndSkills(t *testing.T) {
	// Build a fake repo: commands/commit.md and skills/my-skill/SKILL.md
	repoDir := t.TempDir()
	must(t, os.MkdirAll(filepath.Join(repoDir, "commands"), 0755))
	must(t, os.WriteFile(filepath.Join(repoDir, "commands", "commit.md"), []byte("# commit"), 0644))
	must(t, os.MkdirAll(filepath.Join(repoDir, "skills", "my-skill"), 0755))
	must(t, os.WriteFile(filepath.Join(repoDir, "skills", "my-skill", "SKILL.md"), []byte("# my skill"), 0644))

	agentsDir := t.TempDir()
	targetDir := t.TempDir()

	targets := []config.Target{
		&mockTarget{name: "test-target", baseDir: targetDir},
	}

	det := detector.NewRepositoryDetector()
	l, err := NewComponentLinker(agentsDir, targets, det, nil)
	if err != nil {
		t.Fatalf("NewComponentLinker: %v", err)
	}

	// Change into the fake repo so DetectAndLinkLocalRepositories picks it up
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer os.Chdir(orig) //nolint:errcheck
	must(t, os.Chdir(repoDir))

	if err := l.DetectAndLinkLocalRepositories(); err != nil {
		t.Fatalf("DetectAndLinkLocalRepositories: %v", err)
	}

	// --- skill assertions ---
	// Store symlink should point to the skill directory, not the SKILL.md file.
	skillStore := filepath.Join(agentsDir, "skills", "auto-detected-my-skill")
	skillInfo, err := os.Stat(skillStore)
	if err != nil {
		t.Fatalf("skill store entry missing: %v", err)
	}
	if !skillInfo.IsDir() {
		t.Errorf("skill store entry should resolve to a directory, got mode %v", skillInfo.Mode())
	}

	// Target symlink for skill should exist.
	skillTarget := filepath.Join(targetDir, "skills", "auto-detected-my-skill")
	if _, err := os.Lstat(skillTarget); os.IsNotExist(err) {
		t.Error("skill not linked to target")
	}

	// --- command assertions ---
	// Store entry: the symlink target should be the commands/ directory (not the .md file).
	commandStore := filepath.Join(agentsDir, "commands", "auto-detected-commit")
	commandInfo, err := os.Stat(commandStore)
	if err != nil {
		t.Fatalf("command store entry missing: %v", err)
	}
	if !commandInfo.IsDir() {
		t.Errorf("command store entry should resolve to a directory, got mode %v", commandInfo.Mode())
	}

	// Target: commit.md should be linked flat into the target commands dir.
	commitTarget := filepath.Join(targetDir, "commands", "commit.md")
	if _, err := os.Lstat(commitTarget); os.IsNotExist(err) {
		t.Error("commands/commit.md not linked to target")
	}

	// Verify commit.md symlink resolves to the real file.
	resolved, err := filepath.EvalSymlinks(commitTarget)
	if err != nil {
		t.Fatalf("commit.md symlink broken: %v", err)
	}
	expected, err := filepath.EvalSymlinks(filepath.Join(repoDir, "commands", "commit.md"))
	if err != nil {
		t.Fatalf("EvalSymlinks on expected path: %v", err)
	}
	if resolved != expected {
		t.Errorf("commit.md symlink wrong target\nwant: %s\n got: %s", expected, resolved)
	}
}

// TestDetectAndLinkLocalRepositories_SymlinkedAgentsDir verifies that linking
// works correctly when agentsDir is itself behind a symlink — the original bug
// that caused "component does not exist in any profile" errors.
func TestDetectAndLinkLocalRepositories_SymlinkedAgentsDir(t *testing.T) {
	repoDir := t.TempDir()
	must(t, os.MkdirAll(filepath.Join(repoDir, "skills", "my-skill"), 0755))
	must(t, os.WriteFile(filepath.Join(repoDir, "skills", "my-skill", "SKILL.md"), []byte("# my skill"), 0644))

	// realAgentsDir is the actual directory; agentsDirLink is a symlink to it,
	// simulating ~/.agent-smith being a symlink into a dotfiles repo.
	realAgentsDir := t.TempDir()
	symlinkParent := t.TempDir()
	agentsDirLink := filepath.Join(symlinkParent, "agent-smith")
	must(t, os.Symlink(realAgentsDir, agentsDirLink))

	targetDir := t.TempDir()
	targets := []config.Target{
		&mockTarget{name: "test-target", baseDir: targetDir},
	}

	det := detector.NewRepositoryDetector()
	l, err := NewComponentLinker(agentsDirLink, targets, det, nil)
	if err != nil {
		t.Fatalf("NewComponentLinker: %v", err)
	}

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer os.Chdir(orig) //nolint:errcheck
	must(t, os.Chdir(repoDir))

	if err := l.DetectAndLinkLocalRepositories(); err != nil {
		t.Fatalf("DetectAndLinkLocalRepositories: %v", err)
	}

	// Skill should be linked to target despite agentsDir being a symlink.
	skillTarget := filepath.Join(targetDir, "skills", "auto-detected-my-skill")
	if _, err := os.Lstat(skillTarget); os.IsNotExist(err) {
		t.Error("skill not linked to target when agentsDir is a symlink")
	}

	resolved, err := filepath.EvalSymlinks(skillTarget)
	if err != nil {
		t.Fatalf("skill symlink broken: %v", err)
	}
	expected, err := filepath.EvalSymlinks(filepath.Join(repoDir, "skills", "my-skill"))
	if err != nil {
		t.Fatalf("EvalSymlinks on expected path: %v", err)
	}
	if resolved != expected {
		t.Errorf("skill symlink wrong target\nwant: %s\n got: %s", expected, resolved)
	}
}
