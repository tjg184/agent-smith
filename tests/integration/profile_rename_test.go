package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tjg184/agent-smith/internal/testutil"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/profiles"
	locksvc "github.com/tjg184/agent-smith/pkg/services/lock"
)

func setupRenameEnv(t *testing.T) (tempDir string, pm *profiles.ProfileManager, cleanup func()) {
	t.Helper()
	tempDir = testutil.CreateTempDir(t, "agent-smith-rename-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)

	agentsDir := filepath.Join(tempDir, ".agent-smith")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("failed to create agents dir: %v", err)
	}

	lockService := locksvc.NewService(logger.New(logger.LevelError))
	pm, err := profiles.NewProfileManager(nil, lockService)
	if err != nil {
		t.Fatalf("failed to create profile manager: %v", err)
	}

	cleanup = func() {
		os.Setenv("HOME", oldHome)
		os.RemoveAll(tempDir)
	}
	return tempDir, pm, cleanup
}

func createUserProfile(t *testing.T, tempDir, name string) {
	t.Helper()
	profileDir := filepath.Join(tempDir, ".agent-smith", "profiles", name)
	for _, sub := range []string{"agents", "skills", "commands"} {
		if err := os.MkdirAll(filepath.Join(profileDir, sub), 0755); err != nil {
			t.Fatalf("failed to create profile dir %s: %v", sub, err)
		}
	}
	metadata := map[string]string{"type": "user"}
	data, _ := json.Marshal(metadata)
	if err := os.WriteFile(filepath.Join(profileDir, ".profile-metadata"), data, 0644); err != nil {
		t.Fatalf("failed to write profile metadata: %v", err)
	}
}

func createRepoProfile(t *testing.T, tempDir, name string) {
	t.Helper()
	profileDir := filepath.Join(tempDir, ".agent-smith", "profiles", name)
	for _, sub := range []string{"agents", "skills", "commands"} {
		if err := os.MkdirAll(filepath.Join(profileDir, sub), 0755); err != nil {
			t.Fatalf("failed to create profile dir %s: %v", sub, err)
		}
	}
	metadata := map[string]string{"type": "repo", "source_url": "https://github.com/owner/repo"}
	data, _ := json.Marshal(metadata)
	if err := os.WriteFile(filepath.Join(profileDir, ".profile-metadata"), data, 0644); err != nil {
		t.Fatalf("failed to write profile metadata: %v", err)
	}
}

func setActiveProfile(t *testing.T, tempDir, name string) {
	t.Helper()
	activeProfilePath := filepath.Join(tempDir, ".agent-smith", ".active-profile")
	if err := os.WriteFile(activeProfilePath, []byte(name), 0644); err != nil {
		t.Fatalf("failed to set active profile: %v", err)
	}
}

func readActiveProfile(t *testing.T, tempDir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(tempDir, ".agent-smith", ".active-profile"))
	if os.IsNotExist(err) {
		return ""
	}
	if err != nil {
		t.Fatalf("failed to read active profile: %v", err)
	}
	return string(data)
}

func TestRenameProfile_InactiveProfile(t *testing.T) {
	tempDir, pm, cleanup := setupRenameEnv(t)
	defer cleanup()

	createUserProfile(t, tempDir, "my-profile")

	if err := pm.RenameProfile("my-profile", "renamed-profile"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	oldPath := filepath.Join(tempDir, ".agent-smith", "profiles", "my-profile")
	newPath := filepath.Join(tempDir, ".agent-smith", "profiles", "renamed-profile")

	testutil.AssertDirectoryExists(t, newPath)

	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Errorf("expected old profile directory to not exist after rename")
	}
	if active := readActiveProfile(t, tempDir); active != "" {
		t.Errorf("expected no active profile, got %q", active)
	}
}

func TestRenameProfile_ActiveProfile(t *testing.T) {
	tempDir, pm, cleanup := setupRenameEnv(t)
	defer cleanup()

	createUserProfile(t, tempDir, "active-profile")
	setActiveProfile(t, tempDir, "active-profile")

	if err := pm.RenameProfile("active-profile", "new-active-profile"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	oldPath := filepath.Join(tempDir, ".agent-smith", "profiles", "active-profile")
	newPath := filepath.Join(tempDir, ".agent-smith", "profiles", "new-active-profile")

	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Errorf("expected old profile directory to not exist after rename")
	}
	testutil.AssertDirectoryExists(t, newPath)

	active := readActiveProfile(t, tempDir)
	if active != "new-active-profile" {
		t.Errorf("expected active profile to be %q, got %q", "new-active-profile", active)
	}
}

func TestRenameProfile_OldProfileNotFound(t *testing.T) {
	_, pm, cleanup := setupRenameEnv(t)
	defer cleanup()

	err := pm.RenameProfile("does-not-exist", "new-name")
	if err == nil {
		t.Fatal("expected error renaming non-existent profile, got nil")
	}
}

func TestRenameProfile_NewNameAlreadyExists(t *testing.T) {
	tempDir, pm, cleanup := setupRenameEnv(t)
	defer cleanup()

	createUserProfile(t, tempDir, "profile-a")
	createUserProfile(t, tempDir, "profile-b")

	err := pm.RenameProfile("profile-a", "profile-b")
	if err == nil {
		t.Fatal("expected error when new name already exists, got nil")
	}
}

func TestRenameProfile_InvalidNewName(t *testing.T) {
	tempDir, pm, cleanup := setupRenameEnv(t)
	defer cleanup()

	createUserProfile(t, tempDir, "my-profile")

	cases := []string{"bad name", "bad/name", "bad.name", "", ".."}
	for _, name := range cases {
		if err := pm.RenameProfile("my-profile", name); err == nil {
			t.Errorf("expected validation error for new name %q, got nil", name)
		}
	}
}

func TestRenameProfile_RepoProfileRejected(t *testing.T) {
	tempDir, pm, cleanup := setupRenameEnv(t)
	defer cleanup()

	createRepoProfile(t, tempDir, "repo-profile")

	err := pm.RenameProfile("repo-profile", "new-name")
	if err == nil {
		t.Fatal("expected error renaming repo profile, got nil")
	}
}

func TestRenameProfile_MetadataPreserved(t *testing.T) {
	tempDir, pm, cleanup := setupRenameEnv(t)
	defer cleanup()

	createUserProfile(t, tempDir, "source-profile")

	if err := pm.RenameProfile("source-profile", "dest-profile"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	metadataPath := filepath.Join(tempDir, ".agent-smith", "profiles", "dest-profile", ".profile-metadata")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatalf("metadata file not found after rename: %v", err)
	}

	var metadata map[string]string
	if err := json.Unmarshal(data, &metadata); err != nil {
		t.Fatalf("failed to parse metadata: %v", err)
	}

	if metadata["type"] != "user" {
		t.Errorf("expected type %q, got %q", "user", metadata["type"])
	}
}

func TestRenameProfile_ComponentsPreserved(t *testing.T) {
	tempDir, pm, cleanup := setupRenameEnv(t)
	defer cleanup()

	createUserProfile(t, tempDir, "source-profile")

	skillDir := filepath.Join(tempDir, ".agent-smith", "profiles", "source-profile", "skills", "my-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("failed to create test skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# My Skill"), 0644); err != nil {
		t.Fatalf("failed to write skill file: %v", err)
	}

	if err := pm.RenameProfile("source-profile", "dest-profile"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	renamedSkillDir := filepath.Join(tempDir, ".agent-smith", "profiles", "dest-profile", "skills", "my-skill")
	testutil.AssertDirectoryExists(t, renamedSkillDir)

	skillFile := filepath.Join(renamedSkillDir, "SKILL.md")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		t.Error("expected SKILL.md to exist after rename, but it was not found")
	}
}
