package config

import (
	"path/filepath"
	"testing"

	"github.com/tjg184/agent-smith/pkg/paths"
)

func TestUniversalTarget_GetGlobalBaseDir(t *testing.T) {
	target, err := NewUniversalTarget()
	if err != nil {
		t.Fatalf("NewUniversalTarget() error = %v", err)
	}

	_, err = target.GetGlobalBaseDir()
	if err == nil {
		t.Error("GetGlobalBaseDir() expected error for universal target without base dir, got nil")
	}
}

func TestUniversalTarget_GetGlobalSkillsDir(t *testing.T) {
	testDir := "/test/project/.agents"
	target := NewUniversalTargetWithDir(testDir)

	got, err := target.GetGlobalSkillsDir()
	if err != nil {
		t.Fatalf("GetGlobalSkillsDir() error = %v", err)
	}

	want := filepath.Join(testDir, paths.SkillsSubDir)
	if got != want {
		t.Errorf("GetGlobalSkillsDir() = %v, want %v", got, want)
	}
}

func TestUniversalTarget_GetGlobalAgentsDir(t *testing.T) {
	testDir := "/test/project/.agents"
	target := NewUniversalTargetWithDir(testDir)

	got, err := target.GetGlobalAgentsDir()
	if err != nil {
		t.Fatalf("GetGlobalAgentsDir() error = %v", err)
	}

	want := filepath.Join(testDir, paths.AgentsSubDir)
	if got != want {
		t.Errorf("GetGlobalAgentsDir() = %v, want %v", got, want)
	}
}

func TestUniversalTarget_GetGlobalCommandsDir(t *testing.T) {
	testDir := "/test/project/.agents"
	target := NewUniversalTargetWithDir(testDir)

	got, err := target.GetGlobalCommandsDir()
	if err != nil {
		t.Fatalf("GetGlobalCommandsDir() error = %v", err)
	}

	want := filepath.Join(testDir, paths.CommandsSubDir)
	if got != want {
		t.Errorf("GetGlobalCommandsDir() = %v, want %v", got, want)
	}
}

func TestUniversalTarget_GetGlobalComponentDir(t *testing.T) {
	testDir := "/test/project/.agents"
	target := NewUniversalTargetWithDir(testDir)

	tests := []struct {
		name          string
		componentType string
		want          string
		wantErr       bool
	}{
		{
			name:          "Skills",
			componentType: paths.SkillsSubDir,
			want:          filepath.Join(testDir, paths.SkillsSubDir),
			wantErr:       false,
		},
		{
			name:          "Agents",
			componentType: paths.AgentsSubDir,
			want:          filepath.Join(testDir, paths.AgentsSubDir),
			wantErr:       false,
		},
		{
			name:          "Commands",
			componentType: paths.CommandsSubDir,
			want:          filepath.Join(testDir, paths.CommandsSubDir),
			wantErr:       false,
		},
		{
			name:          "Invalid",
			componentType: "invalid",
			want:          "",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := target.GetGlobalComponentDir(tt.componentType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGlobalComponentDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetGlobalComponentDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUniversalTarget_GetDetectionConfigPath(t *testing.T) {
	testDir := "/test/project/.agents"
	target := NewUniversalTargetWithDir(testDir)

	got, err := target.GetDetectionConfigPath()
	if err != nil {
		t.Fatalf("GetDetectionConfigPath() error = %v", err)
	}

	want := filepath.Join(testDir, paths.DetectionConfigFile)
	if got != want {
		t.Errorf("GetDetectionConfigPath() = %v, want %v", got, want)
	}
}

func TestUniversalTarget_GetName(t *testing.T) {
	target, err := NewUniversalTarget()
	if err != nil {
		t.Fatalf("NewUniversalTarget() error = %v", err)
	}

	got := target.GetName()
	want := "universal"
	if got != want {
		t.Errorf("GetName() = %v, want %v", got, want)
	}
}

func TestUniversalTargetWithDir_CustomDirectory(t *testing.T) {
	customDir := "/custom/path/.agents"
	target := NewUniversalTargetWithDir(customDir)

	baseDir, err := target.GetGlobalBaseDir()
	if err != nil {
		t.Fatalf("GetGlobalBaseDir() error = %v", err)
	}

	if baseDir != customDir {
		t.Errorf("GetGlobalBaseDir() = %v, want %v", baseDir, customDir)
	}
}

func TestUniversalTarget_RequiresProjectContext(t *testing.T) {
	target, err := NewUniversalTarget()
	if err != nil {
		t.Fatalf("NewUniversalTarget() error = %v", err)
	}

	// All directory methods should return error when no base dir is set
	_, err = target.GetGlobalSkillsDir()
	if err == nil {
		t.Error("GetGlobalSkillsDir() expected error without project context, got nil")
	}

	_, err = target.GetGlobalAgentsDir()
	if err == nil {
		t.Error("GetGlobalAgentsDir() expected error without project context, got nil")
	}

	_, err = target.GetGlobalCommandsDir()
	if err == nil {
		t.Error("GetGlobalCommandsDir() expected error without project context, got nil")
	}

	_, err = target.GetDetectionConfigPath()
	if err == nil {
		t.Error("GetDetectionConfigPath() expected error without project context, got nil")
	}
}
