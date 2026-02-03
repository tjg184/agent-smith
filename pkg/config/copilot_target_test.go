package config

import (
	"path/filepath"
	"testing"
)

func TestCopilotTarget_GetBaseDir(t *testing.T) {
	target, err := NewCopilotTarget()
	if err != nil {
		t.Fatalf("Failed to create copilot target: %v", err)
	}

	baseDir, err := target.GetBaseDir()
	if err != nil {
		t.Fatalf("Failed to get base dir: %v", err)
	}

	if baseDir == "" {
		t.Error("Base dir should not be empty")
	}
}

func TestCopilotTarget_GetSkillsDir(t *testing.T) {
	target, err := NewCopilotTarget()
	if err != nil {
		t.Fatalf("Failed to create copilot target: %v", err)
	}

	skillsDir, err := target.GetSkillsDir()
	if err != nil {
		t.Fatalf("Failed to get skills dir: %v", err)
	}

	if skillsDir == "" {
		t.Error("Skills dir should not be empty")
	}

	baseDir, _ := target.GetBaseDir()
	expectedSkillsDir := filepath.Join(baseDir, "skills")
	if skillsDir != expectedSkillsDir {
		t.Errorf("Expected skills dir %s, got %s", expectedSkillsDir, skillsDir)
	}
}

func TestCopilotTarget_GetAgentsDir(t *testing.T) {
	target, err := NewCopilotTarget()
	if err != nil {
		t.Fatalf("Failed to create copilot target: %v", err)
	}

	agentsDir, err := target.GetAgentsDir()
	if err != nil {
		t.Fatalf("Failed to get agents dir: %v", err)
	}

	if agentsDir == "" {
		t.Error("Agents dir should not be empty")
	}

	baseDir, _ := target.GetBaseDir()
	expectedAgentsDir := filepath.Join(baseDir, "agents")
	if agentsDir != expectedAgentsDir {
		t.Errorf("Expected agents dir %s, got %s", expectedAgentsDir, agentsDir)
	}
}

func TestCopilotTarget_GetCommandsDir(t *testing.T) {
	target, err := NewCopilotTarget()
	if err != nil {
		t.Fatalf("Failed to create copilot target: %v", err)
	}

	commandsDir, err := target.GetCommandsDir()
	if err != nil {
		t.Fatalf("Failed to get commands dir: %v", err)
	}

	if commandsDir == "" {
		t.Error("Commands dir should not be empty")
	}

	baseDir, _ := target.GetBaseDir()
	expectedCommandsDir := filepath.Join(baseDir, "commands")
	if commandsDir != expectedCommandsDir {
		t.Errorf("Expected commands dir %s, got %s", expectedCommandsDir, commandsDir)
	}
}

func TestCopilotTarget_GetComponentDir(t *testing.T) {
	target, err := NewCopilotTarget()
	if err != nil {
		t.Fatalf("Failed to create copilot target: %v", err)
	}

	tests := []struct {
		name          string
		componentType string
		shouldError   bool
	}{
		{"Skills", "skills", false},
		{"Agents", "agents", false},
		{"Commands", "commands", false},
		{"Invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := target.GetComponentDir(tt.componentType)
			if tt.shouldError {
				if err == nil {
					t.Error("Expected error for invalid component type")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if dir == "" {
					t.Error("Component dir should not be empty")
				}
			}
		})
	}
}

func TestCopilotTarget_GetDetectionConfigPath(t *testing.T) {
	target, err := NewCopilotTarget()
	if err != nil {
		t.Fatalf("Failed to create copilot target: %v", err)
	}

	configPath, err := target.GetDetectionConfigPath()
	if err != nil {
		t.Fatalf("Failed to get detection config path: %v", err)
	}

	if configPath == "" {
		t.Error("Detection config path should not be empty")
	}

	baseDir, _ := target.GetBaseDir()
	expectedPath := filepath.Join(baseDir, "detection-config.json")
	if configPath != expectedPath {
		t.Errorf("Expected config path %s, got %s", expectedPath, configPath)
	}
}

func TestCopilotTarget_GetName(t *testing.T) {
	target, err := NewCopilotTarget()
	if err != nil {
		t.Fatalf("Failed to create copilot target: %v", err)
	}

	name := target.GetName()
	if name != "copilot" {
		t.Errorf("Expected name 'copilot', got '%s'", name)
	}
}

func TestCopilotTargetWithDir_CustomDirectory(t *testing.T) {
	customDir := "/custom/copilot/path"
	target := NewCopilotTargetWithDir(customDir)

	baseDir, err := target.GetBaseDir()
	if err != nil {
		t.Fatalf("Failed to get base dir: %v", err)
	}

	if baseDir != customDir {
		t.Errorf("Expected base dir %s, got %s", customDir, baseDir)
	}

	// Test that subdirectories are correct
	skillsDir, _ := target.GetSkillsDir()
	expectedSkillsDir := filepath.Join(customDir, "skills")
	if skillsDir != expectedSkillsDir {
		t.Errorf("Expected skills dir %s, got %s", expectedSkillsDir, skillsDir)
	}
}
