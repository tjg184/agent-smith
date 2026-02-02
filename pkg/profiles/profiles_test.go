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
