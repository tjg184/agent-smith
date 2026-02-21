package config

import (
	"path/filepath"
	"testing"

	"github.com/tjg184/agent-smith/pkg/paths"
)

func TestUniversalTarget_GetBaseDir(t *testing.T) {
	target, err := NewUniversalTarget()
	if err != nil {
		t.Fatalf("NewUniversalTarget() error = %v", err)
	}

	_, err = target.GetBaseDir()
	if err == nil {
		t.Error("GetBaseDir() expected error for universal target without base dir, got nil")
	}
}

func TestUniversalTarget_GetSkillsDir(t *testing.T) {
	testDir := "/test/project/.agents"
	target := NewUniversalTargetWithDir(testDir)

	got, err := target.GetSkillsDir()
	if err != nil {
		t.Fatalf("GetSkillsDir() error = %v", err)
	}

	want := filepath.Join(testDir, paths.SkillsSubDir)
	if got != want {
		t.Errorf("GetSkillsDir() = %v, want %v", got, want)
	}
}

func TestUniversalTarget_GetAgentsDir(t *testing.T) {
	testDir := "/test/project/.agents"
	target := NewUniversalTargetWithDir(testDir)

	got, err := target.GetAgentsDir()
	if err != nil {
		t.Fatalf("GetAgentsDir() error = %v", err)
	}

	want := filepath.Join(testDir, paths.AgentsSubDir)
	if got != want {
		t.Errorf("GetAgentsDir() = %v, want %v", got, want)
	}
}

func TestUniversalTarget_GetCommandsDir(t *testing.T) {
	testDir := "/test/project/.agents"
	target := NewUniversalTargetWithDir(testDir)

	got, err := target.GetCommandsDir()
	if err != nil {
		t.Fatalf("GetCommandsDir() error = %v", err)
	}

	want := filepath.Join(testDir, paths.CommandsSubDir)
	if got != want {
		t.Errorf("GetCommandsDir() = %v, want %v", got, want)
	}
}

func TestUniversalTarget_GetComponentDir(t *testing.T) {
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
			got, err := target.GetComponentDir(tt.componentType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetComponentDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetComponentDir() = %v, want %v", got, tt.want)
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

	baseDir, err := target.GetBaseDir()
	if err != nil {
		t.Fatalf("GetBaseDir() error = %v", err)
	}

	if baseDir != customDir {
		t.Errorf("GetBaseDir() = %v, want %v", baseDir, customDir)
	}
}

func TestUniversalTarget_RequiresProjectContext(t *testing.T) {
	target, err := NewUniversalTarget()
	if err != nil {
		t.Fatalf("NewUniversalTarget() error = %v", err)
	}

	// All directory methods should return error when no base dir is set
	_, err = target.GetSkillsDir()
	if err == nil {
		t.Error("GetSkillsDir() expected error without project context, got nil")
	}

	_, err = target.GetAgentsDir()
	if err == nil {
		t.Error("GetAgentsDir() expected error without project context, got nil")
	}

	_, err = target.GetCommandsDir()
	if err == nil {
		t.Error("GetCommandsDir() expected error without project context, got nil")
	}

	_, err = target.GetDetectionConfigPath()
	if err == nil {
		t.Error("GetDetectionConfigPath() expected error without project context, got nil")
	}
}
