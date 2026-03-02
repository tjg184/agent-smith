package profiles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProfile_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		profile *Profile
		want    bool
	}{
		{
			name: "valid profile with agents only",
			profile: &Profile{
				Name:      "test",
				BasePath:  "/home/user/.agents/profiles/test",
				HasAgents: true,
			},
			want: true,
		},
		{
			name: "valid profile with skills only",
			profile: &Profile{
				Name:      "test",
				BasePath:  "/home/user/.agents/profiles/test",
				HasSkills: true,
			},
			want: true,
		},
		{
			name: "valid profile with commands only",
			profile: &Profile{
				Name:        "test",
				BasePath:    "/home/user/.agents/profiles/test",
				HasCommands: true,
			},
			want: true,
		},
		{
			name: "valid profile with all components",
			profile: &Profile{
				Name:        "test",
				BasePath:    "/home/user/.agents/profiles/test",
				HasAgents:   true,
				HasSkills:   true,
				HasCommands: true,
			},
			want: true,
		},
		{
			name: "invalid profile with no components",
			profile: &Profile{
				Name:     "test",
				BasePath: "/home/user/.agents/profiles/test",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.profile.IsValid(); got != tt.want {
				t.Errorf("Profile.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProfile_GetAgentsDir(t *testing.T) {
	profile := &Profile{
		Name:      "test",
		BasePath:  "/home/user/.agents/profiles/test",
		HasAgents: true,
	}

	expected := "/home/user/.agents/profiles/test/agents"
	if got := profile.GetAgentsDir(); got != expected {
		t.Errorf("Profile.GetAgentsDir() = %v, want %v", got, expected)
	}
}

func TestProfile_GetSkillsDir(t *testing.T) {
	profile := &Profile{
		Name:      "test",
		BasePath:  "/home/user/.agents/profiles/test",
		HasSkills: true,
	}

	expected := "/home/user/.agents/profiles/test/skills"
	if got := profile.GetSkillsDir(); got != expected {
		t.Errorf("Profile.GetSkillsDir() = %v, want %v", got, expected)
	}
}

func TestProfile_GetCommandsDir(t *testing.T) {
	profile := &Profile{
		Name:        "test",
		BasePath:    "/home/user/.agents/profiles/test",
		HasCommands: true,
	}

	expected := "/home/user/.agents/profiles/test/commands"
	if got := profile.GetCommandsDir(); got != expected {
		t.Errorf("Profile.GetCommandsDir() = %v, want %v", got, expected)
	}
}

func TestGetProfileNameFromSymlink(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create base installation directory
	baseDir := filepath.Join(tempDir, ".agent-smith")
	baseSkillsDir := filepath.Join(baseDir, "skills")
	if err := os.MkdirAll(baseSkillsDir, 0755); err != nil {
		t.Fatalf("Failed to create base skills dir: %v", err)
	}

	// Create a profile directory
	profilesDir := filepath.Join(baseDir, "profiles")
	workProfileDir := filepath.Join(profilesDir, "work")
	workSkillsDir := filepath.Join(workProfileDir, "skills")
	if err := os.MkdirAll(workSkillsDir, 0755); err != nil {
		t.Fatalf("Failed to create work profile skills dir: %v", err)
	}

	// Create another profile directory
	personalProfileDir := filepath.Join(profilesDir, "personal")
	personalSkillsDir := filepath.Join(personalProfileDir, "skills")
	if err := os.MkdirAll(personalSkillsDir, 0755); err != nil {
		t.Fatalf("Failed to create personal profile skills dir: %v", err)
	}

	// Create actual skill directories
	baseSkillDir := filepath.Join(baseSkillsDir, "base-skill")
	workSkillDir := filepath.Join(workSkillsDir, "work-skill")
	personalSkillDir := filepath.Join(personalSkillsDir, "personal-skill")

	for _, dir := range []string{baseSkillDir, workSkillDir, personalSkillDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create skill dir %s: %v", dir, err)
		}
	}

	// Create a target directory for symlinks
	targetDir := filepath.Join(tempDir, "target", "skills")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("Failed to create target dir: %v", err)
	}

	// Create symlinks
	baseSymlink := filepath.Join(targetDir, "base-skill")
	workSymlink := filepath.Join(targetDir, "work-skill")
	personalSymlink := filepath.Join(targetDir, "personal-skill")
	regularFile := filepath.Join(targetDir, "regular-file")

	if err := os.Symlink(baseSkillDir, baseSymlink); err != nil {
		t.Fatalf("Failed to create base symlink: %v", err)
	}
	if err := os.Symlink(workSkillDir, workSymlink); err != nil {
		t.Fatalf("Failed to create work symlink: %v", err)
	}
	if err := os.Symlink(personalSkillDir, personalSymlink); err != nil {
		t.Fatalf("Failed to create personal symlink: %v", err)
	}
	if err := os.WriteFile(regularFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create regular file: %v", err)
	}

	tests := []struct {
		name        string
		symlinkPath string
		want        string
		wantErr     bool
	}{
		{
			name:        "symlink pointing to base installation",
			symlinkPath: baseSymlink,
			want:        "base",
			wantErr:     false,
		},
		{
			name:        "symlink pointing to work profile",
			symlinkPath: workSymlink,
			want:        "work",
			wantErr:     false,
		},
		{
			name:        "symlink pointing to personal profile",
			symlinkPath: personalSymlink,
			want:        "personal",
			wantErr:     false,
		},
		{
			name:        "regular file (not a symlink)",
			symlinkPath: regularFile,
			want:        "",
			wantErr:     true,
		},
		{
			name:        "non-existent path",
			symlinkPath: filepath.Join(targetDir, "non-existent"),
			want:        "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetProfileNameFromSymlink(tt.symlinkPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProfileNameFromSymlink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetProfileNameFromSymlink() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestProfileValidation tests profile validation logic
func TestProfileValidation(t *testing.T) {
	tempDir := t.TempDir()

	// Set up test environment
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Create profiles directory
	profilesDir := filepath.Join(tempDir, ".agent-smith", "profiles")
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatalf("Failed to create profiles directory: %v", err)
	}

	t.Run("NonExistentProfile", func(t *testing.T) {
		// Create a valid profile for comparison
		validProfilePath := filepath.Join(profilesDir, "valid-profile")
		if err := os.MkdirAll(filepath.Join(validProfilePath, "skills"), 0755); err != nil {
			t.Fatalf("Failed to create valid profile: %v", err)
		}

		// Check if non-existent profile exists
		nonExistentProfile := "non-existent-profile"
		profilePath := filepath.Join(profilesDir, nonExistentProfile)

		_, err := os.Stat(profilePath)
		if !os.IsNotExist(err) {
			t.Errorf("Expected profile to not exist, but stat succeeded or returned different error: %v", err)
		}

		t.Logf("Profile '%s' does not exist (as expected)", nonExistentProfile)
	})

	t.Run("EmptyProfile", func(t *testing.T) {
		pm, err := NewProfileManager(nil, nil)
		if err != nil {
			t.Fatalf("Failed to create profile manager: %v", err)
		}

		// Create an empty profile (directory exists but no component subdirectories)
		emptyProfileName := "empty-profile"
		emptyProfilePath := filepath.Join(profilesDir, emptyProfileName)
		if err := os.MkdirAll(emptyProfilePath, 0755); err != nil {
			t.Fatalf("Failed to create empty profile directory: %v", err)
		}

		// Load the profile and verify it's not valid
		profile := &Profile{
			Name:        emptyProfileName,
			BasePath:    emptyProfilePath,
			HasAgents:   false,
			HasSkills:   false,
			HasCommands: false,
		}

		// Check which component directories exist
		if _, err := os.Stat(filepath.Join(emptyProfilePath, "agents")); err == nil {
			profile.HasAgents = true
		}
		if _, err := os.Stat(filepath.Join(emptyProfilePath, "skills")); err == nil {
			profile.HasSkills = true
		}
		if _, err := os.Stat(filepath.Join(emptyProfilePath, "commands")); err == nil {
			profile.HasCommands = true
		}

		if profile.HasAgents || profile.HasSkills || profile.HasCommands {
			t.Errorf("Expected profile to have no components, but HasAgents=%v, HasSkills=%v, HasCommands=%v",
				profile.HasAgents, profile.HasSkills, profile.HasCommands)
		}

		// Scan profiles - empty profile should not be listed as valid
		validProfiles, err := pm.ScanProfiles()
		if err != nil {
			t.Fatalf("Failed to scan profiles: %v", err)
		}

		for _, p := range validProfiles {
			if p.Name == emptyProfileName {
				t.Errorf("Empty profile '%s' should not be listed as a valid profile", emptyProfileName)
			}
		}
	})

	t.Run("ValidProfile", func(t *testing.T) {
		pm, err := NewProfileManager(nil, nil)
		if err != nil {
			t.Fatalf("Failed to create profile manager: %v", err)
		}

		// Create a valid profile with at least one component directory
		validProfileName := "working-profile"
		validProfilePath := filepath.Join(profilesDir, validProfileName)

		// Create skills directory to make it a valid profile
		skillsDir := filepath.Join(validProfilePath, "skills")
		if err := os.MkdirAll(skillsDir, 0755); err != nil {
			t.Fatalf("Failed to create skills directory: %v", err)
		}

		// Create a test skill
		testSkillPath := filepath.Join(skillsDir, "test-skill")
		if err := os.MkdirAll(testSkillPath, 0755); err != nil {
			t.Fatalf("Failed to create test skill directory: %v", err)
		}

		// Verify profile is valid
		validProfiles, err := pm.ScanProfiles()
		if err != nil {
			t.Fatalf("Failed to scan profiles: %v", err)
		}

		found := false
		for _, p := range validProfiles {
			if p.Name == validProfileName {
				found = true
				if !p.HasSkills {
					t.Errorf("Expected profile to have skills, but HasSkills=false")
				}
			}
		}

		if !found {
			t.Errorf("Valid profile '%s' should be listed in scan results", validProfileName)
		}
	})
}

// TestProfileScanningWithMultipleProfiles verifies that profile scanning correctly identifies all valid profiles
func TestProfileScanningWithMultipleProfiles(t *testing.T) {
	tempDir := t.TempDir()

	// Set HOME to tempDir for profile manager
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Create profiles directory with multiple valid profiles
	profilesDir := filepath.Join(tempDir, ".agent-smith", "profiles")
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatalf("Failed to create profiles directory: %v", err)
	}

	// Create multiple valid profiles
	validProfiles := []string{"work", "personal", "test"}
	for _, profileName := range validProfiles {
		profilePath := filepath.Join(profilesDir, profileName)
		skillsDir := filepath.Join(profilePath, "skills")
		if err := os.MkdirAll(skillsDir, 0755); err != nil {
			t.Fatalf("Failed to create profile %s: %v", profileName, err)
		}
	}

	pm, err := NewProfileManager(nil, nil)
	if err != nil {
		t.Fatalf("Failed to create profile manager: %v", err)
	}

	// Scan profiles
	scannedProfiles, err := pm.ScanProfiles()
	if err != nil {
		t.Fatalf("Failed to scan profiles: %v", err)
	}

	// Verify all valid profiles are found
	if len(scannedProfiles) != len(validProfiles) {
		t.Errorf("Expected %d profiles, got %d", len(validProfiles), len(scannedProfiles))
	}

	for _, expectedProfile := range validProfiles {
		found := false
		for _, scanned := range scannedProfiles {
			if scanned.Name == expectedProfile {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find profile '%s' in scanned results", expectedProfile)
		}
	}
}
