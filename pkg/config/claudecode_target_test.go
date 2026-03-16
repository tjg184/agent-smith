package config

import (
	"path/filepath"
	"testing"
)

func TestClaudeCodeTarget_GetGlobalBaseDir(t *testing.T) {
	target := NewClaudeCodeTargetWithDir("/test/claudecode")

	baseDir, err := target.GetGlobalBaseDir()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := "/test/claudecode"
	if baseDir != expected {
		t.Errorf("Expected base dir %s, got %s", expected, baseDir)
	}
}

func TestClaudeCodeTarget_GetGlobalSkillsDir(t *testing.T) {
	target := NewClaudeCodeTargetWithDir("/test/claudecode")

	skillsDir, err := target.GetGlobalSkillsDir()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := filepath.Join("/test/claudecode", "skills")
	if skillsDir != expected {
		t.Errorf("Expected skills dir %s, got %s", expected, skillsDir)
	}
}

func TestClaudeCodeTarget_GetGlobalAgentsDir(t *testing.T) {
	target := NewClaudeCodeTargetWithDir("/test/claudecode")

	agentsDir, err := target.GetGlobalAgentsDir()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := filepath.Join("/test/claudecode", "agents")
	if agentsDir != expected {
		t.Errorf("Expected agents dir %s, got %s", expected, agentsDir)
	}
}

func TestClaudeCodeTarget_GetGlobalCommandsDir(t *testing.T) {
	target := NewClaudeCodeTargetWithDir("/test/claudecode")

	commandsDir, err := target.GetGlobalCommandsDir()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := filepath.Join("/test/claudecode", "commands")
	if commandsDir != expected {
		t.Errorf("Expected commands dir %s, got %s", expected, commandsDir)
	}
}

func TestClaudeCodeTarget_GetGlobalComponentDir(t *testing.T) {
	target := NewClaudeCodeTargetWithDir("/test/claudecode")

	tests := []struct {
		name          string
		componentType string
		expected      string
		shouldError   bool
	}{
		{
			name:          "skills component type",
			componentType: "skills",
			expected:      filepath.Join("/test/claudecode", "skills"),
			shouldError:   false,
		},
		{
			name:          "agents component type",
			componentType: "agents",
			expected:      filepath.Join("/test/claudecode", "agents"),
			shouldError:   false,
		},
		{
			name:          "commands component type",
			componentType: "commands",
			expected:      filepath.Join("/test/claudecode", "commands"),
			shouldError:   false,
		},
		{
			name:          "unknown component type",
			componentType: "unknown",
			shouldError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := target.GetGlobalComponentDir(tt.componentType)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for component type %s, got nil", tt.componentType)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
				if dir != tt.expected {
					t.Errorf("Expected component dir %s, got %s", tt.expected, dir)
				}
			}
		})
	}
}

func TestClaudeCodeTarget_GetDetectionConfigPath(t *testing.T) {
	target := NewClaudeCodeTargetWithDir("/test/claudecode")

	configPath, err := target.GetDetectionConfigPath()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := filepath.Join("/test/claudecode", "detection-config.json")
	if configPath != expected {
		t.Errorf("Expected config path %s, got %s", expected, configPath)
	}
}

func TestClaudeCodeTarget_GetName(t *testing.T) {
	target := NewClaudeCodeTargetWithDir("/test/claudecode")

	name := target.GetName()
	expected := "claudecode"

	if name != expected {
		t.Errorf("Expected name %s, got %s", expected, name)
	}
}

func TestNewClaudeCodeTarget(t *testing.T) {
	target, err := NewClaudeCodeTarget()
	if err != nil {
		t.Fatalf("Expected no error creating default target, got %v", err)
	}

	if target == nil {
		t.Fatal("Expected target to be non-nil")
	}

	baseDir, err := target.GetGlobalBaseDir()
	if err != nil {
		t.Fatalf("Expected no error getting base dir, got %v", err)
	}

	if baseDir == "" {
		t.Error("Expected base dir to be set, got empty string")
	}

	if target.GetName() != "claudecode" {
		t.Errorf("Expected name 'claudecode', got %s", target.GetName())
	}
}
