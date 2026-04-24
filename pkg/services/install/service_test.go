package install

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/profiles"
	"github.com/tjg184/agent-smith/pkg/services"
	locksvc "github.com/tjg184/agent-smith/pkg/services/lock"
)

// newTestService builds a Service backed by a real ProfileManager rooted at a
// temp HOME dir so no real ~/.agent-smith/ is touched.
func newTestService(t *testing.T) (*Service, string) {
	t.Helper()

	tmpHome, err := os.MkdirTemp("", "agent-smith-install-svc-test-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpHome) })

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	log := logger.New(logger.LevelError)
	lockService := locksvc.NewService(log)

	pm, err := profiles.NewProfileManager(nil, lockService)
	if err != nil {
		t.Fatalf("failed to create profile manager: %v", err)
	}

	svc := &Service{
		profileManager: pm,
		logger:         log,
		formatter:      formatter.New(),
	}
	return svc, tmpHome
}

// profilesDir returns the profiles subdirectory inside the fake HOME.
func profilesDir(home string) string {
	return filepath.Join(home, ".agent-smith", "profiles")
}

func TestValidateInstallOptions_MutuallyExclusiveFlags(t *testing.T) {
	svc, _ := newTestService(t)

	cases := []struct {
		name    string
		opts    services.InstallOptions
		wantErr bool
	}{
		{
			name:    "profile and install-dir both set",
			opts:    services.InstallOptions{Profile: "myprofile", InstallDir: "/tmp/foo"},
			wantErr: true,
		},
		{
			name:    "only profile set",
			opts:    services.InstallOptions{Profile: "myprofile"},
			wantErr: false,
		},
		{
			name:    "only install-dir set",
			opts:    services.InstallOptions{InstallDir: "/tmp/foo"},
			wantErr: false,
		},
		{
			name:    "nothing set",
			opts:    services.InstallOptions{},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.validateInstallOptions(tc.opts)
			if tc.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

// TestInstallSkill_CreatesRepoProfile verifies that InstallSkill always creates a
// repository-sourced profile rather than installing to the base directory.
func TestInstallSkill_CreatesRepoProfile(t *testing.T) {
	svc, tmpHome := newTestService(t)

	repoPath := createLocalSkillRepo(t)

	err := svc.InstallSkill(repoPath, "test-skill", services.InstallOptions{})
	if err != nil {
		t.Fatalf("InstallSkill failed: %v", err)
	}

	profilesPath := profilesDir(tmpHome)
	entries, readErr := os.ReadDir(profilesPath)
	if readErr != nil {
		t.Fatalf("unexpected error reading profiles dir: %v", readErr)
	}
	if len(entries) == 0 {
		t.Error("expected a repo profile to be created, but profiles dir is empty")
	}
}

// TestInstallAgent_CreatesRepoProfile mirrors the skill test for agents.
func TestInstallAgent_CreatesRepoProfile(t *testing.T) {
	svc, tmpHome := newTestService(t)

	repoPath := createLocalAgentRepo(t)

	err := svc.InstallAgent(repoPath, "test-agent", services.InstallOptions{})
	if err != nil {
		t.Fatalf("InstallAgent failed: %v", err)
	}

	profilesPath := profilesDir(tmpHome)
	entries, readErr := os.ReadDir(profilesPath)
	if readErr != nil {
		t.Fatalf("unexpected error reading profiles dir: %v", readErr)
	}
	if len(entries) == 0 {
		t.Error("expected a repo profile to be created, but profiles dir is empty")
	}
}

// TestInstallCommand_CreatesRepoProfile mirrors the skill test for commands.
func TestInstallCommand_CreatesRepoProfile(t *testing.T) {
	svc, tmpHome := newTestService(t)

	repoPath := createLocalCommandRepo(t)

	err := svc.InstallCommand(repoPath, "test-command", services.InstallOptions{})
	if err != nil {
		t.Fatalf("InstallCommand failed: %v", err)
	}

	profilesPath := profilesDir(tmpHome)
	entries, readErr := os.ReadDir(profilesPath)
	if readErr != nil {
		t.Fatalf("unexpected error reading profiles dir: %v", readErr)
	}
	if len(entries) == 0 {
		t.Error("expected a repo profile to be created, but profiles dir is empty")
	}
}

// TestInstallBulk_CreatesRepoProfile verifies InstallBulk always creates a repo profile.
func TestInstallBulk_CreatesRepoProfile(t *testing.T) {
	svc, tmpHome := newTestService(t)

	repoPath := createLocalSkillRepo(t)

	err := svc.InstallBulk(repoPath, services.InstallOptions{})
	if err != nil {
		t.Fatalf("InstallBulk failed: %v", err)
	}

	profilesPath := profilesDir(tmpHome)
	entries, readErr := os.ReadDir(profilesPath)
	if readErr != nil {
		t.Fatalf("unexpected error reading profiles dir: %v", readErr)
	}
	if len(entries) == 0 {
		t.Error("expected a repo profile to be created, but profiles dir is empty")
	}
}

// createLocalSkillRepo creates a minimal local git repo containing a skill.
func createLocalSkillRepo(t *testing.T) string {
	t.Helper()
	return createLocalRepo(t, "skill-repo", map[string]string{
		"skills/test-skill/SKILL.md": "# test-skill\nA test skill.",
	})
}

// createLocalAgentRepo creates a minimal local git repo containing an agent.
func createLocalAgentRepo(t *testing.T) string {
	t.Helper()
	return createLocalRepo(t, "agent-repo", map[string]string{
		"agents/test-agent.md": "# test-agent\nA test agent.",
	})
}

// createLocalCommandRepo creates a minimal local git repo containing a command.
func createLocalCommandRepo(t *testing.T) string {
	t.Helper()
	return createLocalRepo(t, "command-repo", map[string]string{
		"commands/test-command.md": "# test-command\nA test command.",
	})
}

// createLocalRepo initialises a bare git repo with the given files committed.
func createLocalRepo(t *testing.T, name string, files map[string]string) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "agent-smith-repo-"+name+"-*")
	if err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	for relPath, content := range files {
		full := filepath.Join(dir, relPath)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("failed to create dirs for %s: %v", relPath, err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", relPath, err)
		}
	}

	run := func(args ...string) {
		t.Helper()
		// #nosec G204 — test-only, no user input
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	run("git", "init")
	run("git", "add", ".")
	run("git", "commit", "-m", "init")

	return dir
}
