package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tgaines/agent-smith/pkg/paths"
)

// Config represents the agent-smith configuration file structure
type Config struct {
	Version       int                  `json:"version"`
	CustomTargets []CustomTargetConfig `json:"customTargets"`
}

// CustomTargetConfig represents a custom target configuration in the config file
type CustomTargetConfig struct {
	Name        string `json:"name"`
	BaseDir     string `json:"baseDir"`
	SkillsDir   string `json:"skillsDir"`
	AgentsDir   string `json:"agentsDir"`
	CommandsDir string `json:"commandsDir"`
}

const (
	// ConfigVersion is the current version of the config file format
	ConfigVersion = 1
	// ConfigFileName is the name of the config file
	ConfigFileName = "config.json"
)

var (
	// targetNameRegex validates target names (alphanumeric, hyphens, underscores only)
	targetNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return "", fmt.Errorf("failed to get agents directory: %w", err)
	}
	return filepath.Join(agentsDir, ConfigFileName), nil
}

// LoadConfig loads the configuration from the config file
// Returns an empty config if the file doesn't exist (not an error)
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// If config file doesn't exist, return empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{
			Version:       ConfigVersion,
			CustomTargets: []CustomTargetConfig{},
		}, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate config
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration to the config file
func SaveConfig(config *Config) error {
	// Validate config before saving
	if err := validateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Ensure the parent directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// validateConfig validates the configuration structure
func validateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Check version
	if config.Version != ConfigVersion {
		return fmt.Errorf("unsupported config version %d (expected %d)", config.Version, ConfigVersion)
	}

	// Validate custom targets
	if config.CustomTargets == nil {
		config.CustomTargets = []CustomTargetConfig{}
	}

	// Track target names to check for duplicates
	seenNames := make(map[string]bool)

	for i, target := range config.CustomTargets {
		if err := validateCustomTargetConfig(&target); err != nil {
			return fmt.Errorf("custom target at index %d: %w", i, err)
		}

		// Check for duplicate names (case-insensitive)
		nameLower := strings.ToLower(target.Name)
		if seenNames[nameLower] {
			return fmt.Errorf("duplicate target name: %s (names are case-insensitive)", target.Name)
		}
		seenNames[nameLower] = true

		// Check for conflict with built-in target names
		if nameLower == "opencode" || nameLower == "claudecode" {
			return fmt.Errorf("target name %s conflicts with built-in target", target.Name)
		}
	}

	return nil
}

// validateCustomTargetConfig validates a single custom target configuration
func validateCustomTargetConfig(target *CustomTargetConfig) error {
	// Validate name
	if target.Name == "" {
		return fmt.Errorf("target name cannot be empty")
	}
	if !targetNameRegex.MatchString(target.Name) {
		return fmt.Errorf("target name %q contains invalid characters (only alphanumeric, hyphens, and underscores allowed)", target.Name)
	}

	// Validate baseDir
	if target.BaseDir == "" {
		return fmt.Errorf("baseDir cannot be empty")
	}

	// Expand and validate base directory path
	expandedBaseDir, err := expandHomePath(target.BaseDir)
	if err != nil {
		return fmt.Errorf("invalid baseDir %q: %w", target.BaseDir, err)
	}

	// Convert to absolute path for validation
	absBaseDir, err := filepath.Abs(expandedBaseDir)
	if err != nil {
		return fmt.Errorf("invalid baseDir path %q: %w", target.BaseDir, err)
	}
	_ = absBaseDir // We've validated it, but we don't modify the original value

	// Validate subdirectory names
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

// validateSubdirectoryName validates a subdirectory name
func validateSubdirectoryName(name, fieldName string) error {
	if name == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}

	// Check for invalid characters (no slashes or path separators)
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("%s %q cannot contain path separators", fieldName, name)
	}

	// Check for special directory names
	if name == "." || name == ".." {
		return fmt.Errorf("%s cannot be %q", fieldName, name)
	}

	return nil
}

// expandHomePath expands ~ to the user's home directory
func expandHomePath(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	if len(path) == 1 {
		return home, nil
	}

	if path[1] == filepath.Separator || path[1] == '/' {
		return filepath.Join(home, path[2:]), nil
	}

	return path, nil
}
