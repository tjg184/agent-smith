package paths

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetAgentsDir(t *testing.T) {
	agentsDir, err := GetAgentsDir()
	if err != nil {
		t.Fatalf("GetAgentsDir() error = %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".agent-smith")

	if agentsDir != expected {
		t.Errorf("GetAgentsDir() = %v, want %v", agentsDir, expected)
	}

	// Should not contain ~
	if strings.Contains(agentsDir, "~") {
		t.Errorf("GetAgentsDir() contains unexpanded ~: %v", agentsDir)
	}
}

func TestGetOpencodeDir(t *testing.T) {
	opencodeDir, err := GetOpencodeDir()
	if err != nil {
		t.Fatalf("GetOpencodeDir() error = %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "opencode")

	if opencodeDir != expected {
		t.Errorf("GetOpencodeDir() = %v, want %v", opencodeDir, expected)
	}

	// Should not contain ~
	if strings.Contains(opencodeDir, "~") {
		t.Errorf("GetOpencodeDir() contains unexpanded ~: %v", opencodeDir)
	}
}

func TestGetSkillsDir(t *testing.T) {
	skillsDir, err := GetSkillsDir()
	if err != nil {
		t.Fatalf("GetSkillsDir() error = %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".agent-smith", "skills")

	if skillsDir != expected {
		t.Errorf("GetSkillsDir() = %v, want %v", skillsDir, expected)
	}
}

func TestGetAgentsSubDir(t *testing.T) {
	agentsSubDir, err := GetAgentsSubDir()
	if err != nil {
		t.Fatalf("GetAgentsSubDir() error = %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".agent-smith", "agents")

	if agentsSubDir != expected {
		t.Errorf("GetAgentsSubDir() = %v, want %v", agentsSubDir, expected)
	}
}

func TestGetCommandsDir(t *testing.T) {
	commandsDir, err := GetCommandsDir()
	if err != nil {
		t.Fatalf("GetCommandsDir() error = %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".agent-smith", "commands")

	if commandsDir != expected {
		t.Errorf("GetCommandsDir() = %v, want %v", commandsDir, expected)
	}
}

func TestGetDetectionConfigPath(t *testing.T) {
	configPath, err := GetDetectionConfigPath()
	if err != nil {
		t.Fatalf("GetDetectionConfigPath() error = %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "opencode", "detection-config.json")

	if configPath != expected {
		t.Errorf("GetDetectionConfigPath() = %v, want %v", configPath, expected)
	}
}

func TestGetComponentLockPath(t *testing.T) {
	tests := []struct {
		name          string
		baseDir       string
		componentType string
		want          string
	}{
		{
			name:          "skill lock",
			baseDir:       "/test/base",
			componentType: "skills",
			want:          filepath.Join("/test/base", ".skill-lock.json"),
		},
		{
			name:          "agent lock",
			baseDir:       "/test/base",
			componentType: "agents",
			want:          filepath.Join("/test/base", ".agent-lock.json"),
		},
		{
			name:          "command lock",
			baseDir:       "/test/base",
			componentType: "commands",
			want:          filepath.Join("/test/base", ".command-lock.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetComponentLockPath(tt.baseDir, tt.componentType)
			if got != tt.want {
				t.Errorf("GetComponentLockPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetComponentTypes(t *testing.T) {
	types := GetComponentTypes()
	expected := []string{"skills", "agents", "commands"}

	if len(types) != len(expected) {
		t.Errorf("GetComponentTypes() returned %d types, want %d", len(types), len(expected))
	}

	for i, typ := range types {
		if typ != expected[i] {
			t.Errorf("GetComponentTypes()[%d] = %v, want %v", i, typ, expected[i])
		}
	}
}

func TestGetComponentTypeNames(t *testing.T) {
	types := GetComponentTypeNames()
	expected := []string{"agents", "commands", "skills"}

	if len(types) != len(expected) {
		t.Errorf("GetComponentTypeNames() returned %d types, want %d", len(types), len(expected))
	}

	for i, typ := range types {
		if typ != expected[i] {
			t.Errorf("GetComponentTypeNames()[%d] = %v, want %v", i, typ, expected[i])
		}
	}
}
