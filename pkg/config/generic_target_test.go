package config

import (
	"path/filepath"
	"testing"
)

func TestBuiltInTargetSpecs_Metadata(t *testing.T) {
	tests := []struct {
		name        string
		displayName string
		projectDir  string
		isUniversal bool
	}{
		{"opencode", "OpenCode", ".opencode", false},
		{"claudecode", "Claude Code", ".claude", false},
		{"copilot", "GitHub Copilot", ".github", false},
		{"universal", "Universal", ".agents", true},
	}

	if len(builtInTargetSpecs) != len(tests) {
		t.Fatalf("expected %d specs, got %d", len(tests), len(builtInTargetSpecs))
	}

	for i, tt := range tests {
		spec := builtInTargetSpecs[i]
		t.Run(tt.name, func(t *testing.T) {
			target := newTargetFromSpecWithDir(spec, "/test/"+tt.name)

			if got := target.GetName(); got != tt.name {
				t.Errorf("GetName() = %q, want %q", got, tt.name)
			}
			if got := target.GetDisplayName(); got != tt.displayName {
				t.Errorf("GetDisplayName() = %q, want %q", got, tt.displayName)
			}
			if got := target.GetProjectDirName(); got != tt.projectDir {
				t.Errorf("GetProjectDirName() = %q, want %q", got, tt.projectDir)
			}
			if got := target.IsUniversalTarget(); got != tt.isUniversal {
				t.Errorf("IsUniversalTarget() = %v, want %v", got, tt.isUniversal)
			}
		})
	}
}

func TestGenericTarget_Paths(t *testing.T) {
	spec := builtInTargetSpecs[0] // opencode
	target := newTargetFromSpecWithDir(spec, "/test/opencode")

	t.Run("GetGlobalBaseDir", func(t *testing.T) {
		dir, err := target.GetGlobalBaseDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dir != "/test/opencode" {
			t.Errorf("got %q, want %q", dir, "/test/opencode")
		}
	})

	t.Run("GetGlobalSkillsDir", func(t *testing.T) {
		dir, err := target.GetGlobalSkillsDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if want := filepath.Join("/test/opencode", "skills"); dir != want {
			t.Errorf("got %q, want %q", dir, want)
		}
	})

	t.Run("GetGlobalAgentsDir", func(t *testing.T) {
		dir, err := target.GetGlobalAgentsDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if want := filepath.Join("/test/opencode", "agents"); dir != want {
			t.Errorf("got %q, want %q", dir, want)
		}
	})

	t.Run("GetGlobalCommandsDir", func(t *testing.T) {
		dir, err := target.GetGlobalCommandsDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if want := filepath.Join("/test/opencode", "commands"); dir != want {
			t.Errorf("got %q, want %q", dir, want)
		}
	})

	t.Run("GetGlobalComponentDir_invalid", func(t *testing.T) {
		_, err := target.GetGlobalComponentDir("unknown")
		if err == nil {
			t.Error("expected error for unknown component type, got nil")
		}
	})

	t.Run("GetDetectionConfigPath", func(t *testing.T) {
		path, err := target.GetDetectionConfigPath()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if want := filepath.Join("/test/opencode", "detection-config.json"); path != want {
			t.Errorf("got %q, want %q", path, want)
		}
	})

	t.Run("GetProjectBaseDir", func(t *testing.T) {
		dir := target.GetProjectBaseDir("/my/project")
		if want := filepath.Join("/my/project", ".opencode"); dir != want {
			t.Errorf("got %q, want %q", dir, want)
		}
	})

	t.Run("GetProjectComponentDir", func(t *testing.T) {
		dir, err := target.GetProjectComponentDir("/my/project", "skills")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if want := filepath.Join("/my/project", ".opencode", "skills"); dir != want {
			t.Errorf("got %q, want %q", dir, want)
		}
	})
}

func TestNewTarget_BuiltIns(t *testing.T) {
	for _, spec := range builtInTargetSpecs {
		t.Run(spec.name, func(t *testing.T) {
			target, err := NewTarget(spec.name)
			if err != nil {
				t.Fatalf("NewTarget(%q) error: %v", spec.name, err)
			}
			if target == nil {
				t.Fatal("expected non-nil target")
			}
			if got := target.GetName(); got != spec.name {
				t.Errorf("GetName() = %q, want %q", got, spec.name)
			}
			baseDir, err := target.GetGlobalBaseDir()
			if err != nil {
				t.Fatalf("GetGlobalBaseDir() error: %v", err)
			}
			if baseDir == "" {
				t.Error("expected non-empty base dir")
			}
		})
	}
}
