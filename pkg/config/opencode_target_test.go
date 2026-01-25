package config

import (
	"path/filepath"
	"testing"
)

func TestOpencodeTarget_GetBaseDir(t *testing.T) {
	target := NewOpencodeTargetWithDir("/test/opencode")

	baseDir, err := target.GetBaseDir()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := "/test/opencode"
	if baseDir != expected {
		t.Errorf("Expected base dir %s, got %s", expected, baseDir)
	}
}

func TestOpencodeTarget_GetSkillsDir(t *testing.T) {
	target := NewOpencodeTargetWithDir("/test/opencode")

	skillsDir, err := target.GetSkillsDir()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := filepath.Join("/test/opencode", "skills")
	if skillsDir != expected {
		t.Errorf("Expected skills dir %s, got %s", expected, skillsDir)
	}
}

func TestOpencodeTarget_GetAgentsDir(t *testing.T) {
	target := NewOpencodeTargetWithDir("/test/opencode")

	agentsDir, err := target.GetAgentsDir()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := filepath.Join("/test/opencode", "agents")
	if agentsDir != expected {
		t.Errorf("Expected agents dir %s, got %s", expected, agentsDir)
	}
}

func TestOpencodeTarget_GetCommandsDir(t *testing.T) {
	target := NewOpencodeTargetWithDir("/test/opencode")

	commandsDir, err := target.GetCommandsDir()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := filepath.Join("/test/opencode", "commands")
	if commandsDir != expected {
		t.Errorf("Expected commands dir %s, got %s", expected, commandsDir)
	}
}

func TestOpencodeTarget_GetComponentDir(t *testing.T) {
	target := NewOpencodeTargetWithDir("/test/opencode")

	tests := []struct {
		name          string
		componentType string
		expected      string
		shouldError   bool
	}{
		{
			name:          "skills component type",
			componentType: "skills",
			expected:      filepath.Join("/test/opencode", "skills"),
			shouldError:   false,
		},
		{
			name:          "agents component type",
			componentType: "agents",
			expected:      filepath.Join("/test/opencode", "agents"),
			shouldError:   false,
		},
		{
			name:          "commands component type",
			componentType: "commands",
			expected:      filepath.Join("/test/opencode", "commands"),
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
			dir, err := target.GetComponentDir(tt.componentType)

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

func TestOpencodeTarget_GetDetectionConfigPath(t *testing.T) {
	target := NewOpencodeTargetWithDir("/test/opencode")

	configPath, err := target.GetDetectionConfigPath()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := filepath.Join("/test/opencode", "detection-config.json")
	if configPath != expected {
		t.Errorf("Expected config path %s, got %s", expected, configPath)
	}
}

func TestOpencodeTarget_GetName(t *testing.T) {
	target := NewOpencodeTargetWithDir("/test/opencode")

	name := target.GetName()
	expected := "opencode"

	if name != expected {
		t.Errorf("Expected name %s, got %s", expected, name)
	}
}

func TestNewOpencodeTarget(t *testing.T) {
	target, err := NewOpencodeTarget()
	if err != nil {
		t.Fatalf("Expected no error creating default target, got %v", err)
	}

	if target == nil {
		t.Fatal("Expected target to be non-nil")
	}

	// Verify it has a base dir set
	baseDir, err := target.GetBaseDir()
	if err != nil {
		t.Fatalf("Expected no error getting base dir, got %v", err)
	}

	if baseDir == "" {
		t.Error("Expected base dir to be set, got empty string")
	}

	// Verify the name is correct
	if target.GetName() != "opencode" {
		t.Errorf("Expected name 'opencode', got %s", target.GetName())
	}
}
