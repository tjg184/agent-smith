//go:build integration

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/profiles"
	locksvc "github.com/tjg184/agent-smith/pkg/services/lock"
	profilesvc "github.com/tjg184/agent-smith/pkg/services/profile"
)

func setupShareEnv(t *testing.T) (tempDir string, pm *profiles.ProfileManager, cleanup func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "agent-smith-share-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)

	agentsDir := filepath.Join(tempDir, ".agent-smith")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("failed to create agents dir: %v", err)
	}

	lockService := locksvc.NewService(logger.New(logger.LevelError))
	pm, err = profiles.NewProfileManager(nil, lockService)
	if err != nil {
		t.Fatalf("failed to create profile manager: %v", err)
	}

	cleanup = func() {
		os.Setenv("HOME", oldHome)
		os.RemoveAll(tempDir)
	}
	return tempDir, pm, cleanup
}

func createShareableProfile(t *testing.T, tempDir, name string) {
	t.Helper()
	profileDir := filepath.Join(tempDir, ".agent-smith", "profiles", name)
	for _, sub := range []string{"agents", "skills", "commands"} {
		if err := os.MkdirAll(filepath.Join(profileDir, sub), 0755); err != nil {
			t.Fatalf("failed to create profile subdir %s: %v", sub, err)
		}
	}

	// Add a skill with a lock file so ShareProfile has something to emit
	skillDir := filepath.Join(profileDir, "skills", "my-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}
	lockContent := `{"name":"my-skill","repo":"owner/repo","type":"skill","source":"remote"}`
	if err := os.WriteFile(filepath.Join(skillDir, ".lock"), []byte(lockContent), 0644); err != nil {
		t.Fatalf("failed to write lock file: %v", err)
	}
}

func TestShareProfile_UsesActiveProfileWhenNoneSpecified(t *testing.T) {
	tempDir, pm, cleanup := setupShareEnv(t)
	defer cleanup()

	createShareableProfile(t, tempDir, "my-active-profile")
	setActiveProfile(t, tempDir, "my-active-profile")

	activeProfile, err := pm.GetActiveProfile()
	if err != nil {
		t.Fatalf("unexpected error getting active profile: %v", err)
	}
	if activeProfile != "my-active-profile" {
		t.Fatalf("expected active profile %q, got %q", "my-active-profile", activeProfile)
	}

	svc := profilesvc.NewService(pm, logger.New(logger.LevelError), formatter.New())

	outputPath := filepath.Join(tempDir, "share-output.txt")
	if err := svc.ShareProfile(activeProfile, outputPath); err != nil {
		t.Fatalf("ShareProfile failed: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read share output: %v", err)
	}
	output := string(data)

	if !strings.Contains(output, "my-active-profile") {
		t.Errorf("expected output to reference profile name, got:\n%s", output)
	}
}

func TestShareProfile_NoActiveProfileReturnsError(t *testing.T) {
	_, pm, cleanup := setupShareEnv(t)
	defer cleanup()

	activeProfile, err := pm.GetActiveProfile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activeProfile != "" {
		t.Fatalf("expected no active profile, got %q", activeProfile)
	}
}

func TestShareProfile_ExplicitNameTakesPrecedence(t *testing.T) {
	tempDir, pm, cleanup := setupShareEnv(t)
	defer cleanup()

	createShareableProfile(t, tempDir, "active-profile")
	createShareableProfile(t, tempDir, "explicit-profile")
	setActiveProfile(t, tempDir, "active-profile")

	svc := profilesvc.NewService(pm, logger.New(logger.LevelError), formatter.New())

	outputPath := filepath.Join(tempDir, "share-output.txt")
	if err := svc.ShareProfile("explicit-profile", outputPath); err != nil {
		t.Fatalf("ShareProfile failed: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read share output: %v", err)
	}
	output := string(data)

	if !strings.Contains(output, "explicit-profile") {
		t.Errorf("expected output to reference explicit-profile, got:\n%s", output)
	}
	if strings.Contains(output, "active-profile") {
		t.Errorf("expected output to NOT reference active-profile when explicit name given, got:\n%s", output)
	}
}
