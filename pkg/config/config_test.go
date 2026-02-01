package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateCustomTargetConfig_Valid(t *testing.T) {
	validConfigs := []CustomTargetConfig{
		{
			Name:        "cursor",
			BaseDir:     "~/.cursor",
			SkillsDir:   "skills",
			AgentsDir:   "agents",
			CommandsDir: "commands",
		},
		{
			Name:        "vscode-123",
			BaseDir:     "/opt/vscode/agent-smith",
			SkillsDir:   "skills",
			AgentsDir:   "agents",
			CommandsDir: "commands",
		},
		{
			Name:        "my_custom_target",
			BaseDir:     "./relative/path",
			SkillsDir:   "skills",
			AgentsDir:   "agents",
			CommandsDir: "commands",
		},
	}

	for _, config := range validConfigs {
		t.Run(config.Name, func(t *testing.T) {
			err := validateCustomTargetConfig(&config)
			if err != nil {
				t.Errorf("Expected valid config for %s, got error: %v", config.Name, err)
			}
		})
	}
}

func TestValidateCustomTargetConfig_Invalid(t *testing.T) {
	tests := []struct {
		name   string
		config CustomTargetConfig
	}{
		{
			name: "empty name",
			config: CustomTargetConfig{
				Name:        "",
				BaseDir:     "/test",
				SkillsDir:   "skills",
				AgentsDir:   "agents",
				CommandsDir: "commands",
			},
		},
		{
			name: "name with slash",
			config: CustomTargetConfig{
				Name:        "test/target",
				BaseDir:     "/test",
				SkillsDir:   "skills",
				AgentsDir:   "agents",
				CommandsDir: "commands",
			},
		},
		{
			name: "name with spaces",
			config: CustomTargetConfig{
				Name:        "test target",
				BaseDir:     "/test",
				SkillsDir:   "skills",
				AgentsDir:   "agents",
				CommandsDir: "commands",
			},
		},
		{
			name: "empty baseDir",
			config: CustomTargetConfig{
				Name:        "test",
				BaseDir:     "",
				SkillsDir:   "skills",
				AgentsDir:   "agents",
				CommandsDir: "commands",
			},
		},
		{
			name: "empty skillsDir",
			config: CustomTargetConfig{
				Name:        "test",
				BaseDir:     "/test",
				SkillsDir:   "",
				AgentsDir:   "agents",
				CommandsDir: "commands",
			},
		},
		{
			name: "skillsDir with slash",
			config: CustomTargetConfig{
				Name:        "test",
				BaseDir:     "/test",
				SkillsDir:   "my/skills",
				AgentsDir:   "agents",
				CommandsDir: "commands",
			},
		},
		{
			name: "agentsDir is dot",
			config: CustomTargetConfig{
				Name:        "test",
				BaseDir:     "/test",
				SkillsDir:   "skills",
				AgentsDir:   ".",
				CommandsDir: "commands",
			},
		},
		{
			name: "commandsDir is dot-dot",
			config: CustomTargetConfig{
				Name:        "test",
				BaseDir:     "/test",
				SkillsDir:   "skills",
				AgentsDir:   "agents",
				CommandsDir: "..",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCustomTargetConfig(&tt.config)
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
		})
	}
}

func TestValidateConfig_Valid(t *testing.T) {
	config := &Config{
		Version: 1,
		CustomTargets: []CustomTargetConfig{
			{
				Name:        "cursor",
				BaseDir:     "~/.cursor",
				SkillsDir:   "skills",
				AgentsDir:   "agents",
				CommandsDir: "commands",
			},
			{
				Name:        "vscode",
				BaseDir:     "~/.vscode/agent-smith",
				SkillsDir:   "skills",
				AgentsDir:   "agents",
				CommandsDir: "commands",
			},
		},
	}

	err := validateConfig(config)
	if err != nil {
		t.Errorf("Expected valid config, got error: %v", err)
	}
}

func TestValidateConfig_DuplicateNames(t *testing.T) {
	config := &Config{
		Version: 1,
		CustomTargets: []CustomTargetConfig{
			{
				Name:        "cursor",
				BaseDir:     "~/.cursor",
				SkillsDir:   "skills",
				AgentsDir:   "agents",
				CommandsDir: "commands",
			},
			{
				Name:        "Cursor", // Same name, different case
				BaseDir:     "~/.cursor2",
				SkillsDir:   "skills",
				AgentsDir:   "agents",
				CommandsDir: "commands",
			},
		},
	}

	err := validateConfig(config)
	if err == nil {
		t.Errorf("Expected error for duplicate names, got nil")
	}
}

func TestValidateConfig_BuiltinNameConflict(t *testing.T) {
	tests := []struct {
		name       string
		targetName string
	}{
		{"opencode", "opencode"},
		{"opencode uppercase", "OpenCode"},
		{"claudecode", "claudecode"},
		{"claudecode uppercase", "ClaudeCode"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Version: 1,
				CustomTargets: []CustomTargetConfig{
					{
						Name:        tt.targetName,
						BaseDir:     "/test",
						SkillsDir:   "skills",
						AgentsDir:   "agents",
						CommandsDir: "commands",
					},
				},
			}

			err := validateConfig(config)
			if err == nil {
				t.Errorf("Expected error for built-in name conflict with %s, got nil", tt.targetName)
			}
		})
	}
}

func TestValidateConfig_WrongVersion(t *testing.T) {
	config := &Config{
		Version:       999,
		CustomTargets: []CustomTargetConfig{},
	}

	err := validateConfig(config)
	if err == nil {
		t.Errorf("Expected error for wrong version, got nil")
	}
}

func TestLoadConfig_NonExistent(t *testing.T) {
	// This test will use a non-existent path in the agents dir
	// which should return an empty config without error
	origHome := os.Getenv("HOME")
	tempDir, err := os.MkdirTemp("", "agent-smith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	defer os.Setenv("HOME", origHome)

	os.Setenv("HOME", tempDir)

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Expected no error for non-existent config, got %v", err)
	}

	if config == nil {
		t.Fatal("Expected non-nil config")
	}

	if config.Version != ConfigVersion {
		t.Errorf("Expected version %d, got %d", ConfigVersion, config.Version)
	}

	if len(config.CustomTargets) != 0 {
		t.Errorf("Expected empty custom targets, got %d", len(config.CustomTargets))
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Create a temporary home directory for testing
	origHome := os.Getenv("HOME")
	tempHome, err := os.MkdirTemp("", "agent-smith-home-*")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tempHome)
	defer os.Setenv("HOME", origHome)

	os.Setenv("HOME", tempHome)

	// Create .agent-smith directory
	agentsDir := filepath.Join(tempHome, ".agent-smith")
	err = os.MkdirAll(agentsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .agent-smith dir: %v", err)
	}

	// Create a test config
	testConfig := &Config{
		Version: ConfigVersion,
		CustomTargets: []CustomTargetConfig{
			{
				Name:        "cursor",
				BaseDir:     "~/.cursor",
				SkillsDir:   "skills",
				AgentsDir:   "agents",
				CommandsDir: "commands",
			},
			{
				Name:        "vscode",
				BaseDir:     "~/.vscode/agent-smith",
				SkillsDir:   "my-skills",
				AgentsDir:   "my-agents",
				CommandsDir: "my-commands",
			},
		},
	}

	// Save the config
	err = SaveConfig(testConfig)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify the file exists
	configPath := filepath.Join(agentsDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file was not created")
	}

	// Load the config back
	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify the loaded config matches
	if loadedConfig.Version != testConfig.Version {
		t.Errorf("Expected version %d, got %d", testConfig.Version, loadedConfig.Version)
	}

	if len(loadedConfig.CustomTargets) != len(testConfig.CustomTargets) {
		t.Fatalf("Expected %d custom targets, got %d", len(testConfig.CustomTargets), len(loadedConfig.CustomTargets))
	}

	for i, expected := range testConfig.CustomTargets {
		actual := loadedConfig.CustomTargets[i]
		if actual.Name != expected.Name {
			t.Errorf("Target %d: expected name %s, got %s", i, expected.Name, actual.Name)
		}
		if actual.BaseDir != expected.BaseDir {
			t.Errorf("Target %d: expected baseDir %s, got %s", i, expected.BaseDir, actual.BaseDir)
		}
		if actual.SkillsDir != expected.SkillsDir {
			t.Errorf("Target %d: expected skillsDir %s, got %s", i, expected.SkillsDir, actual.SkillsDir)
		}
		if actual.AgentsDir != expected.AgentsDir {
			t.Errorf("Target %d: expected agentsDir %s, got %s", i, expected.AgentsDir, actual.AgentsDir)
		}
		if actual.CommandsDir != expected.CommandsDir {
			t.Errorf("Target %d: expected commandsDir %s, got %s", i, expected.CommandsDir, actual.CommandsDir)
		}
	}
}

func TestSaveConfig_InvalidConfig(t *testing.T) {
	// Create a temporary home directory for testing
	origHome := os.Getenv("HOME")
	tempHome, err := os.MkdirTemp("", "agent-smith-home-*")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tempHome)
	defer os.Setenv("HOME", origHome)

	os.Setenv("HOME", tempHome)

	// Create .agent-smith directory
	agentsDir := filepath.Join(tempHome, ".agent-smith")
	err = os.MkdirAll(agentsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .agent-smith dir: %v", err)
	}

	// Try to save an invalid config
	invalidConfig := &Config{
		Version: ConfigVersion,
		CustomTargets: []CustomTargetConfig{
			{
				Name:        "", // Invalid: empty name
				BaseDir:     "/test",
				SkillsDir:   "skills",
				AgentsDir:   "agents",
				CommandsDir: "commands",
			},
		},
	}

	err = SaveConfig(invalidConfig)
	if err == nil {
		t.Errorf("Expected error saving invalid config, got nil")
	}
}

func TestLoadConfig_MalformedJSON(t *testing.T) {
	// Create a temporary home directory for testing
	origHome := os.Getenv("HOME")
	tempHome, err := os.MkdirTemp("", "agent-smith-home-*")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tempHome)
	defer os.Setenv("HOME", origHome)

	os.Setenv("HOME", tempHome)

	// Create .agent-smith directory
	agentsDir := filepath.Join(tempHome, ".agent-smith")
	err = os.MkdirAll(agentsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .agent-smith dir: %v", err)
	}

	// Write malformed JSON
	configPath := filepath.Join(agentsDir, "config.json")
	malformedJSON := `{"version": 1, "customTargets": [invalid json}`
	err = os.WriteFile(configPath, []byte(malformedJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write malformed JSON: %v", err)
	}

	// Try to load it
	_, err = LoadConfig()
	if err == nil {
		t.Errorf("Expected error loading malformed JSON, got nil")
	}
}

func TestExpandHomePath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory, skipping test")
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "tilde only",
			input:    "~",
			expected: homeDir,
		},
		{
			name:     "tilde with path",
			input:    "~/.cursor",
			expected: filepath.Join(homeDir, ".cursor"),
		},
		{
			name:     "absolute path",
			input:    "/opt/test",
			expected: "/opt/test",
		},
		{
			name:     "relative path",
			input:    "./test",
			expected: "./test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandHomePath(tt.input)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestConfigJSONFormat(t *testing.T) {
	// Test that the config marshals to the expected JSON format
	config := &Config{
		Version: 1,
		CustomTargets: []CustomTargetConfig{
			{
				Name:        "cursor",
				BaseDir:     "~/.cursor",
				SkillsDir:   "skills",
				AgentsDir:   "agents",
				CommandsDir: "commands",
			},
		},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Unmarshal it back and verify
	var loadedConfig Config
	err = json.Unmarshal(data, &loadedConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if loadedConfig.Version != config.Version {
		t.Errorf("Expected version %d, got %d", config.Version, loadedConfig.Version)
	}

	if len(loadedConfig.CustomTargets) != len(config.CustomTargets) {
		t.Fatalf("Expected %d targets, got %d", len(config.CustomTargets), len(loadedConfig.CustomTargets))
	}
}

func TestValidateDisplaySettings_ValidValues(t *testing.T) {
	tests := []struct {
		name     string
		settings DisplaySettings
		expected DisplaySettings
	}{
		{
			name:     "auto colors and unicode",
			settings: DisplaySettings{Colors: "auto", Unicode: "auto"},
			expected: DisplaySettings{Colors: "auto", Unicode: "auto"},
		},
		{
			name:     "always colors and unicode",
			settings: DisplaySettings{Colors: "always", Unicode: "always"},
			expected: DisplaySettings{Colors: "always", Unicode: "always"},
		},
		{
			name:     "never colors",
			settings: DisplaySettings{Colors: "never", Unicode: "auto"},
			expected: DisplaySettings{Colors: "never", Unicode: "auto"},
		},
		{
			name:     "ascii unicode",
			settings: DisplaySettings{Colors: "auto", Unicode: "ascii"},
			expected: DisplaySettings{Colors: "auto", Unicode: "ascii"},
		},
		{
			name:     "empty values default to auto",
			settings: DisplaySettings{Colors: "", Unicode: ""},
			expected: DisplaySettings{Colors: "auto", Unicode: "auto"},
		},
		{
			name:     "invalid colors defaults to auto",
			settings: DisplaySettings{Colors: "invalid", Unicode: "auto"},
			expected: DisplaySettings{Colors: "auto", Unicode: "auto"},
		},
		{
			name:     "invalid unicode defaults to auto",
			settings: DisplaySettings{Colors: "auto", Unicode: "invalid"},
			expected: DisplaySettings{Colors: "auto", Unicode: "auto"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDisplaySettings(&tt.settings)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.settings.Colors != tt.expected.Colors {
				t.Errorf("Expected colors %q, got %q", tt.expected.Colors, tt.settings.Colors)
			}

			if tt.settings.Unicode != tt.expected.Unicode {
				t.Errorf("Expected unicode %q, got %q", tt.expected.Unicode, tt.settings.Unicode)
			}
		})
	}
}

func TestLoadConfig_DisplayDefaults(t *testing.T) {
	// Create a temporary home directory for testing
	origHome := os.Getenv("HOME")
	tempHome, err := os.MkdirTemp("", "agent-smith-home-*")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tempHome)
	defer os.Setenv("HOME", origHome)

	os.Setenv("HOME", tempHome)

	// Load config when file doesn't exist - should have defaults
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.Display.Colors != "auto" {
		t.Errorf("Expected default colors 'auto', got %q", config.Display.Colors)
	}

	if config.Display.Unicode != "auto" {
		t.Errorf("Expected default unicode 'auto', got %q", config.Display.Unicode)
	}
}

func TestSaveAndLoadConfig_WithDisplay(t *testing.T) {
	// Create a temporary home directory for testing
	origHome := os.Getenv("HOME")
	tempHome, err := os.MkdirTemp("", "agent-smith-home-*")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tempHome)
	defer os.Setenv("HOME", origHome)

	os.Setenv("HOME", tempHome)

	// Create .agent-smith directory
	agentsDir := filepath.Join(tempHome, ".agent-smith")
	err = os.MkdirAll(agentsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .agent-smith dir: %v", err)
	}

	// Create a test config with custom display settings
	testConfig := &Config{
		Version:       ConfigVersion,
		CustomTargets: []CustomTargetConfig{},
		Display: DisplaySettings{
			Colors:  "always",
			Unicode: "ascii",
		},
	}

	// Save the config
	err = SaveConfig(testConfig)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load the config back
	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify the display settings were preserved
	if loadedConfig.Display.Colors != "always" {
		t.Errorf("Expected colors 'always', got %q", loadedConfig.Display.Colors)
	}

	if loadedConfig.Display.Unicode != "ascii" {
		t.Errorf("Expected unicode 'ascii', got %q", loadedConfig.Display.Unicode)
	}
}

func TestValidateConfig_AppliesDisplayDefaults(t *testing.T) {
	config := &Config{
		Version:       ConfigVersion,
		CustomTargets: []CustomTargetConfig{},
		Display: DisplaySettings{
			Colors:  "",
			Unicode: "",
		},
	}

	err := validateConfig(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Validate should have applied defaults
	if config.Display.Colors != "auto" {
		t.Errorf("Expected colors 'auto', got %q", config.Display.Colors)
	}

	if config.Display.Unicode != "auto" {
		t.Errorf("Expected unicode 'auto', got %q", config.Display.Unicode)
	}
}

func TestConfigJSONFormat_WithDisplay(t *testing.T) {
	// Test that the config marshals to the expected JSON format with display settings
	config := &Config{
		Version:       1,
		CustomTargets: []CustomTargetConfig{},
		Display: DisplaySettings{
			Colors:  "always",
			Unicode: "ascii",
		},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Unmarshal it back and verify
	var loadedConfig Config
	err = json.Unmarshal(data, &loadedConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if loadedConfig.Display.Colors != config.Display.Colors {
		t.Errorf("Expected colors %q, got %q", config.Display.Colors, loadedConfig.Display.Colors)
	}

	if loadedConfig.Display.Unicode != config.Display.Unicode {
		t.Errorf("Expected unicode %q, got %q", config.Display.Unicode, loadedConfig.Display.Unicode)
	}
}
