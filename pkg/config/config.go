package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tjg184/agent-smith/pkg/paths"
)

type Config struct {
	Version       int                  `json:"version"`
	CustomTargets []CustomTargetConfig `json:"customTargets"`
	Display       DisplaySettings      `json:"display"`
}

type CustomTargetConfig struct {
	Name        string `json:"name"`
	BaseDir     string `json:"baseDir"`
	ProjectDir  string `json:"projectDir"`
	SkillsDir   string `json:"skillsDir"`
	AgentsDir   string `json:"agentsDir"`
	CommandsDir string `json:"commandsDir"`
}

type DisplaySettings struct {
	Colors  string `json:"colors"`
	Unicode string `json:"unicode"`
}

const (
	ConfigVersion  = 1
	ConfigFileName = "config.json"
)

var targetNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func GetConfigPath() (string, error) {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return "", fmt.Errorf("failed to get agents directory: %w", err)
	}
	return filepath.Join(agentsDir, ConfigFileName), nil
}

func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{
			Version:       ConfigVersion,
			CustomTargets: []CustomTargetConfig{},
			Display: DisplaySettings{
				Colors:  "auto",
				Unicode: "auto",
			},
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config file: %w", err)
	}

	return &config, nil
}

func SaveConfig(config *Config) error {
	if err := validateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func validateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Version != ConfigVersion {
		return fmt.Errorf("unsupported config version %d (expected %d)", config.Version, ConfigVersion)
	}

	if config.CustomTargets == nil {
		config.CustomTargets = []CustomTargetConfig{}
	}

	if err := validateDisplaySettings(&config.Display); err != nil {
		return fmt.Errorf("display settings: %w", err)
	}

	seenNames := make(map[string]bool)

	for i, target := range config.CustomTargets {
		if err := validateCustomTargetConfig(&target); err != nil {
			return fmt.Errorf("custom target at index %d: %w", i, err)
		}

		nameLower := strings.ToLower(target.Name)
		if seenNames[nameLower] {
			return fmt.Errorf("duplicate target name: %s (names are case-insensitive)", target.Name)
		}
		seenNames[nameLower] = true

		for _, reserved := range builtInTargetNames() {
			if nameLower == reserved {
				return fmt.Errorf("target name %s conflicts with built-in target", target.Name)
			}
		}
	}

	return nil
}

func validateCustomTargetConfig(target *CustomTargetConfig) error {
	if target.Name == "" {
		return fmt.Errorf("target name cannot be empty")
	}
	if !targetNameRegex.MatchString(target.Name) {
		return fmt.Errorf("target name %q contains invalid characters (only alphanumeric, hyphens, and underscores allowed)", target.Name)
	}

	if target.BaseDir == "" {
		return fmt.Errorf("baseDir cannot be empty")
	}

	expandedBaseDir, err := paths.ExpandHome(target.BaseDir)
	if err != nil {
		return fmt.Errorf("invalid baseDir %q: %w", target.BaseDir, err)
	}

	absBaseDir, err := filepath.Abs(expandedBaseDir)
	if err != nil {
		return fmt.Errorf("invalid baseDir path %q: %w", target.BaseDir, err)
	}
	_ = absBaseDir

	if target.ProjectDir == "" {
		return fmt.Errorf("projectDir cannot be empty (required for materialize support)")
	}
	if err := validateSubdirectoryName(target.ProjectDir, "projectDir"); err != nil {
		return err
	}

	if err := validateSubdirectoryName(target.SkillsDir, "skillsDir"); err != nil {
		return err
	}
	if err := validateSubdirectoryName(target.AgentsDir, "agentsDir"); err != nil {
		return err
	}
	if err := validateSubdirectoryName(target.CommandsDir, "commandsDir"); err != nil {
		return err
	}

	return nil
}

func validateSubdirectoryName(name, fieldName string) error {
	if name == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}

	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("%s %q cannot contain path separators", fieldName, name)
	}

	if name == "." || name == ".." {
		return fmt.Errorf("%s cannot be %q", fieldName, name)
	}

	return nil
}

func validateDisplaySettings(settings *DisplaySettings) error {
	switch settings.Colors {
	case "auto", "always", "never":
	case "":
		settings.Colors = "auto"
	default:
		settings.Colors = "auto"
	}

	switch settings.Unicode {
	case "auto", "always", "ascii":
	case "":
		settings.Unicode = "auto"
	default:
		settings.Unicode = "auto"
	}

	return nil
}
