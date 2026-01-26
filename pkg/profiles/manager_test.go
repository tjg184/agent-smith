package profiles

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tgaines/agent-smith/pkg/paths"
)

func TestProfileManager_ScanProfiles(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	// Create profiles directory structure
	profilesDir := filepath.Join(tempDir, "profiles")

	// Create valid profile with agents directory
	codingProfile := filepath.Join(profilesDir, "coding")
	if err := os.MkdirAll(filepath.Join(codingProfile, paths.AgentsSubDir), 0755); err != nil {
		t.Fatalf("Failed to create coding profile: %v", err)
	}

	// Create valid profile with skills directory
	workProfile := filepath.Join(profilesDir, "work")
	if err := os.MkdirAll(filepath.Join(workProfile, paths.SkillsSubDir), 0755); err != nil {
		t.Fatalf("Failed to create work profile: %v", err)
	}

	// Create valid profile with all component directories
	fullProfile := filepath.Join(profilesDir, "full")
	if err := os.MkdirAll(filepath.Join(fullProfile, paths.AgentsSubDir), 0755); err != nil {
		t.Fatalf("Failed to create full profile agents: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(fullProfile, paths.SkillsSubDir), 0755); err != nil {
		t.Fatalf("Failed to create full profile skills: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(fullProfile, paths.CommandsSubDir), 0755); err != nil {
		t.Fatalf("Failed to create full profile commands: %v", err)
	}

	// Create invalid profile (no component directories)
	emptyProfile := filepath.Join(profilesDir, "empty")
	if err := os.MkdirAll(emptyProfile, 0755); err != nil {
		t.Fatalf("Failed to create empty profile: %v", err)
	}

	// Create hidden directory (should be ignored)
	hiddenProfile := filepath.Join(profilesDir, ".hidden")
	if err := os.MkdirAll(filepath.Join(hiddenProfile, paths.AgentsSubDir), 0755); err != nil {
		t.Fatalf("Failed to create hidden profile: %v", err)
	}

	// Create a file in profiles directory (should be ignored)
	testFile := filepath.Join(profilesDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create ProfileManager with custom profiles directory
	pm := &ProfileManager{profilesDir: profilesDir}

	// Scan profiles
	profiles, err := pm.ScanProfiles()
	if err != nil {
		t.Fatalf("ScanProfiles() error = %v", err)
	}

	// Should find exactly 3 valid profiles (coding, work, full)
	if len(profiles) != 3 {
		t.Errorf("Expected 3 valid profiles, got %d", len(profiles))
	}

	// Verify profiles are correctly loaded
	profileMap := make(map[string]*Profile)
	for _, p := range profiles {
		profileMap[p.Name] = p
	}

	// Check coding profile
	if coding, ok := profileMap["coding"]; !ok {
		t.Error("coding profile not found")
	} else {
		if !coding.HasAgents {
			t.Error("coding profile should have agents")
		}
		if coding.HasSkills || coding.HasCommands {
			t.Error("coding profile should not have skills or commands")
		}
	}

	// Check work profile
	if work, ok := profileMap["work"]; !ok {
		t.Error("work profile not found")
	} else {
		if !work.HasSkills {
			t.Error("work profile should have skills")
		}
		if work.HasAgents || work.HasCommands {
			t.Error("work profile should not have agents or commands")
		}
	}

	// Check full profile
	if full, ok := profileMap["full"]; !ok {
		t.Error("full profile not found")
	} else {
		if !full.HasAgents || !full.HasSkills || !full.HasCommands {
			t.Error("full profile should have all component directories")
		}
	}

	// Verify empty, hidden, and file are not included
	if _, ok := profileMap["empty"]; ok {
		t.Error("empty profile should not be included")
	}
	if _, ok := profileMap[".hidden"]; ok {
		t.Error("hidden profile should not be included")
	}
	if _, ok := profileMap["test.txt"]; ok {
		t.Error("test.txt file should not be included")
	}
}

func TestProfileManager_ScanProfiles_NoProfilesDirectory(t *testing.T) {
	// Create temporary directory but don't create profiles subdirectory
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")

	// Create ProfileManager with non-existent profiles directory
	pm := &ProfileManager{profilesDir: profilesDir}

	// Scan profiles should return empty list without error
	profiles, err := pm.ScanProfiles()
	if err != nil {
		t.Fatalf("ScanProfiles() error = %v", err)
	}

	if len(profiles) != 0 {
		t.Errorf("Expected 0 profiles, got %d", len(profiles))
	}
}

func TestProfileManager_loadProfile(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")

	// Create test profile with only skills
	testProfile := filepath.Join(profilesDir, "test")
	if err := os.MkdirAll(filepath.Join(testProfile, paths.SkillsSubDir), 0755); err != nil {
		t.Fatalf("Failed to create test profile: %v", err)
	}

	pm := &ProfileManager{profilesDir: profilesDir}

	// Load profile
	profile := pm.loadProfile("test")

	if profile.Name != "test" {
		t.Errorf("Expected profile name 'test', got '%s'", profile.Name)
	}

	if profile.BasePath != testProfile {
		t.Errorf("Expected base path '%s', got '%s'", testProfile, profile.BasePath)
	}

	if profile.HasAgents || profile.HasCommands {
		t.Error("Profile should not have agents or commands")
	}

	if !profile.HasSkills {
		t.Error("Profile should have skills")
	}

	if !profile.IsValid() {
		t.Error("Profile should be valid")
	}
}
