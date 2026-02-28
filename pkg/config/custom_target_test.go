package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCustomTarget_GetGlobalBaseDir(t *testing.T) {
	config := CustomTargetConfig{
		Name:        "test-target",
		BaseDir:     "/test/custom",
		ProjectDir:  ".custom",
		SkillsDir:   "skills",
		AgentsDir:   "agents",
		CommandsDir: "commands",
	}

	target, err := NewCustomTarget(config)
	if err != nil {
		t.Fatalf("Expected no error creating target, got %v", err)
	}

	baseDir, err := target.GetGlobalBaseDir()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should be converted to absolute path
	expected, _ := filepath.Abs("/test/custom")
	if baseDir != expected {
		t.Errorf("Expected base dir %s, got %s", expected, baseDir)
	}
}

func TestCustomTarget_GetGlobalSkillsDir(t *testing.T) {
	config := CustomTargetConfig{
		Name:        "test-target",
		BaseDir:     "/test/custom",
		ProjectDir:  ".custom",
		SkillsDir:   "my-skills",
		AgentsDir:   "agents",
		CommandsDir: "commands",
	}

	target, err := NewCustomTarget(config)
	if err != nil {
		t.Fatalf("Expected no error creating target, got %v", err)
	}

	skillsDir, err := target.GetGlobalSkillsDir()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	baseDir, _ := filepath.Abs("/test/custom")
	expected := filepath.Join(baseDir, "my-skills")
	if skillsDir != expected {
		t.Errorf("Expected skills dir %s, got %s", expected, skillsDir)
	}
}

func TestCustomTarget_GetGlobalAgentsDir(t *testing.T) {
	config := CustomTargetConfig{
		Name:        "test-target",
		BaseDir:     "/test/custom",
		ProjectDir:  ".custom",
		SkillsDir:   "skills",
		AgentsDir:   "my-agents",
		CommandsDir: "commands",
	}

	target, err := NewCustomTarget(config)
	if err != nil {
		t.Fatalf("Expected no error creating target, got %v", err)
	}

	agentsDir, err := target.GetGlobalAgentsDir()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	baseDir, _ := filepath.Abs("/test/custom")
	expected := filepath.Join(baseDir, "my-agents")
	if agentsDir != expected {
		t.Errorf("Expected agents dir %s, got %s", expected, agentsDir)
	}
}

func TestCustomTarget_GetGlobalCommandsDir(t *testing.T) {
	config := CustomTargetConfig{
		Name:        "test-target",
		BaseDir:     "/test/custom",
		ProjectDir:  ".custom",
		SkillsDir:   "skills",
		AgentsDir:   "agents",
		CommandsDir: "my-commands",
	}

	target, err := NewCustomTarget(config)
	if err != nil {
		t.Fatalf("Expected no error creating target, got %v", err)
	}

	commandsDir, err := target.GetGlobalCommandsDir()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	baseDir, _ := filepath.Abs("/test/custom")
	expected := filepath.Join(baseDir, "my-commands")
	if commandsDir != expected {
		t.Errorf("Expected commands dir %s, got %s", expected, commandsDir)
	}
}

func TestCustomTarget_GetGlobalComponentDir(t *testing.T) {
	config := CustomTargetConfig{
		Name:        "test-target",
		BaseDir:     "/test/custom",
		ProjectDir:  ".custom",
		SkillsDir:   "skills",
		AgentsDir:   "agents",
		CommandsDir: "commands",
	}

	target, err := NewCustomTarget(config)
	if err != nil {
		t.Fatalf("Expected no error creating target, got %v", err)
	}

	baseDir, _ := filepath.Abs("/test/custom")

	tests := []struct {
		name          string
		componentType string
		expected      string
		shouldError   bool
	}{
		{
			name:          "skills component type",
			componentType: "skills",
			expected:      filepath.Join(baseDir, "skills"),
			shouldError:   false,
		},
		{
			name:          "agents component type",
			componentType: "agents",
			expected:      filepath.Join(baseDir, "agents"),
			shouldError:   false,
		},
		{
			name:          "commands component type",
			componentType: "commands",
			expected:      filepath.Join(baseDir, "commands"),
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

func TestCustomTarget_GetName(t *testing.T) {
	config := CustomTargetConfig{
		Name:        "my-custom-target",
		BaseDir:     "/test/custom",
		ProjectDir:  ".custom",
		SkillsDir:   "skills",
		AgentsDir:   "agents",
		CommandsDir: "commands",
	}

	target, err := NewCustomTarget(config)
	if err != nil {
		t.Fatalf("Expected no error creating target, got %v", err)
	}

	name := target.GetName()
	expected := "my-custom-target"

	if name != expected {
		t.Errorf("Expected name %s, got %s", expected, name)
	}
}

func TestCustomTarget_TildeExpansion(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory, skipping tilde expansion test")
	}

	config := CustomTargetConfig{
		Name:        "test-target",
		BaseDir:     "~/test/custom",
		ProjectDir:  ".custom",
		SkillsDir:   "skills",
		AgentsDir:   "agents",
		CommandsDir: "commands",
	}

	target, err := NewCustomTarget(config)
	if err != nil {
		t.Fatalf("Expected no error creating target, got %v", err)
	}

	baseDir, err := target.GetGlobalBaseDir()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedBase := filepath.Join(homeDir, "test/custom")
	expected, _ := filepath.Abs(expectedBase)
	if baseDir != expected {
		t.Errorf("Expected base dir %s, got %s", expected, baseDir)
	}
}

func TestCustomTarget_InvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config CustomTargetConfig
	}{
		{
			name: "empty name",
			config: CustomTargetConfig{
				Name:        "",
				BaseDir:     "/test/custom",
				ProjectDir:  ".custom",
				SkillsDir:   "skills",
				AgentsDir:   "agents",
				CommandsDir: "commands",
			},
		},
		{
			name: "invalid name characters",
			config: CustomTargetConfig{
				Name:        "test/target",
				BaseDir:     "/test/custom",
				ProjectDir:  ".custom",
				SkillsDir:   "skills",
				AgentsDir:   "agents",
				CommandsDir: "commands",
			},
		},
		{
			name: "empty base dir",
			config: CustomTargetConfig{
				Name:        "test-target",
				BaseDir:     "",
				ProjectDir:  ".custom",
				SkillsDir:   "skills",
				AgentsDir:   "agents",
				CommandsDir: "commands",
			},
		},
		{
			name: "empty skills dir",
			config: CustomTargetConfig{
				Name:        "test-target",
				BaseDir:     "/test/custom",
				ProjectDir:  ".custom",
				SkillsDir:   "",
				AgentsDir:   "agents",
				CommandsDir: "commands",
			},
		},
		{
			name: "skills dir with slash",
			config: CustomTargetConfig{
				Name:        "test-target",
				BaseDir:     "/test/custom",
				ProjectDir:  ".custom",
				SkillsDir:   "my/skills",
				AgentsDir:   "agents",
				CommandsDir: "commands",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCustomTarget(tt.config)
			if err == nil {
				t.Errorf("Expected error for invalid config, got nil")
			}
		})
	}
}
