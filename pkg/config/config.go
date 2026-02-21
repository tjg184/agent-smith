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

// Config represents the agent-smith configuration file structure.
//
// The configuration file is stored at ~/.agent-smith/config.json and contains
// settings for custom targets and other global preferences.
//
// Example config.json:
//
//	{
//	  "version": 1,
//	  "customTargets": [
//	    {
//	      "name": "cursor",
//	      "baseDir": "~/.cursor",
//	      "skillsDir": "skills",
//	      "agentsDir": "agents",
//	      "commandsDir": "commands"
//	    }
//	  ],
//	  "display": {
//	    "colors": "auto",
//	    "unicode": "auto"
//	  }
//	}
//
// Schema:
//   - version (int): Configuration file format version. Currently must be 1.
//   - customTargets (array): List of custom target configurations for linking
//     components to additional editors or tools.
//   - display (object): Display settings for CLI output formatting and colors.
type Config struct {
	Version       int                  `json:"version"`       // Configuration schema version (currently 1)
	CustomTargets []CustomTargetConfig `json:"customTargets"` // List of custom target configurations
	Display       DisplaySettings      `json:"display"`       // Display settings for output formatting
}

// CustomTargetConfig represents a custom target configuration in the config file.
//
// Custom targets allow you to link agent-smith components (skills, agents, commands)
// to additional editors or tools beyond the built-in OpenCode and ClaudeCode targets.
//
// Field descriptions:
//   - name: Unique identifier for the target (alphanumeric, hyphens, underscores only).
//     Must not conflict with built-in targets "opencode" or "claudecode".
//   - baseDir: Root directory where the target stores its configuration.
//     Supports tilde (~) expansion for home directory.
//   - skillsDir: Subdirectory name (relative to baseDir) for skills.
//     Must be a simple directory name without path separators.
//   - agentsDir: Subdirectory name (relative to baseDir) for agents.
//     Must be a simple directory name without path separators.
//   - commandsDir: Subdirectory name (relative to baseDir) for commands.
//     Must be a simple directory name without path separators.
//
// Example:
//
//	{
//	  "name": "vscode-insiders",
//	  "baseDir": "~/.vscode-insiders",
//	  "skillsDir": "skills",
//	  "agentsDir": "agents",
//	  "commandsDir": "commands"
//	}
//
// Validation rules:
//   - name: Cannot be empty, must match ^[a-zA-Z0-9_-]+$, case-insensitive unique
//   - baseDir: Cannot be empty, must be a valid path (absolute or with ~)
//   - skillsDir/agentsDir/commandsDir: Cannot be empty, no path separators, not "." or ".."
type CustomTargetConfig struct {
	Name        string `json:"name"`        // Unique target identifier
	BaseDir     string `json:"baseDir"`     // Root directory for target configuration
	SkillsDir   string `json:"skillsDir"`   // Skills subdirectory name
	AgentsDir   string `json:"agentsDir"`   // Agents subdirectory name
	CommandsDir string `json:"commandsDir"` // Commands subdirectory name
}

// DisplaySettings represents display configuration for CLI output.
//
// Controls how agent-smith formats and colors its output in the terminal.
//
// Field descriptions:
//   - colors: Controls ANSI color output behavior
//   - "auto": Automatically detect TTY and use colors when appropriate (default)
//   - "always": Force colors on, even for non-TTY outputs
//   - "never": Force colors off, even for TTY outputs
//   - unicode: Controls Unicode character usage for formatting
//   - "auto": Automatically detect Unicode support and use when appropriate (default)
//   - "always": Force Unicode characters (box-drawing, symbols)
//   - "ascii": Use ASCII-only characters for maximum compatibility
//
// Example:
//
//	{
//	  "colors": "auto",
//	  "unicode": "auto"
//	}
//
// Validation rules:
//   - colors: Must be "auto", "always", or "never" (defaults to "auto" if invalid)
//   - unicode: Must be "auto", "always", or "ascii" (defaults to "auto" if invalid)
type DisplaySettings struct {
	Colors  string `json:"colors"`  // Color output mode: "auto", "always", or "never"
	Unicode string `json:"unicode"` // Unicode character mode: "auto", "always", or "ascii"
}

const (
	// ConfigVersion is the current version of the config file format.
	// This version number should be incremented when making breaking changes
	// to the configuration schema. The application will reject config files
	// with unsupported version numbers.
	ConfigVersion = 1

	// ConfigFileName is the name of the config file stored in the agents directory.
	// Full path: ~/.agent-smith/config.json
	ConfigFileName = "config.json"
)

var (
	// targetNameRegex validates target names (alphanumeric, hyphens, underscores only).
	// This ensures target names are filesystem-safe and URL-safe.
	// Examples of valid names: "cursor", "vscode-insiders", "my_editor"
	// Examples of invalid names: "my.editor", "editor#1", "editor/name"
	targetNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// GetConfigPath returns the path to the config file.
//
// The config file is stored at ~/.agent-smith/config.json by default.
// This location is determined by the GetAgentsDir() function.
//
// Returns:
//   - string: Absolute path to the config file
//   - error: If the agents directory cannot be determined
func GetConfigPath() (string, error) {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return "", fmt.Errorf("failed to get agents directory: %w", err)
	}
	return filepath.Join(agentsDir, ConfigFileName), nil
}

// LoadConfig loads the configuration from the config file.
//
// If the config file doesn't exist, this is not considered an error.
// Instead, an empty config with default values is returned.
//
// The config file is loaded from ~/.agent-smith/config.json and must:
//   - Be valid JSON
//   - Match the ConfigVersion (currently 1)
//   - Pass all validation rules (unique target names, valid paths, etc.)
//
// Returns:
//   - *Config: Loaded configuration (or empty config if file doesn't exist)
//   - error: If the file exists but cannot be read, parsed, or is invalid
//
// Example usage:
//
//	config, err := LoadConfig()
//	if err != nil {
//	    log.Fatalf("Failed to load config: %v", err)
//	}
//	for _, target := range config.CustomTargets {
//	    fmt.Printf("Target: %s at %s\n", target.Name, target.BaseDir)
//	}
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
			Display: DisplaySettings{
				Colors:  "auto",
				Unicode: "auto",
			},
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

// SaveConfig saves the configuration to the config file.
//
// The configuration is validated before saving to ensure it meets all requirements.
// The config file is written to ~/.agent-smith/config.json with indented JSON formatting
// for readability.
//
// The parent directory (~/.agent-smith) is automatically created if it doesn't exist.
//
// Parameters:
//   - config: The configuration to save (must not be nil and must be valid)
//
// Returns:
//   - error: If validation fails, the directory cannot be created, or writing fails
//
// Example usage:
//
//	config := &Config{
//	    Version: ConfigVersion,
//	    CustomTargets: []CustomTargetConfig{
//	        {
//	            Name:        "cursor",
//	            BaseDir:     "~/.cursor",
//	            SkillsDir:   "skills",
//	            AgentsDir:   "agents",
//	            CommandsDir: "commands",
//	        },
//	    },
//	}
//	if err := SaveConfig(config); err != nil {
//	    log.Fatalf("Failed to save config: %v", err)
//	}
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

// validateConfig validates the configuration structure.
//
// This function performs comprehensive validation including:
//   - Config is not nil
//   - Version matches ConfigVersion (currently 1)
//   - Each custom target passes validation
//   - No duplicate target names (case-insensitive)
//   - No conflicts with built-in target names (opencode, claudecode)
//
// Returns:
//   - error: Describes the first validation failure encountered, or nil if valid
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

	// Apply defaults and validate display settings
	if err := validateDisplaySettings(&config.Display); err != nil {
		return fmt.Errorf("display settings: %w", err)
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

// validateCustomTargetConfig validates a single custom target configuration.
//
// Validation rules:
//   - name: Cannot be empty, must match pattern ^[a-zA-Z0-9_-]+$
//   - baseDir: Cannot be empty, must be a valid path (supports ~ expansion)
//   - skillsDir, agentsDir, commandsDir: Cannot be empty, must be simple directory
//     names without path separators, cannot be "." or ".."
//
// Returns:
//   - error: Describes the validation failure, or nil if valid
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

// validateSubdirectoryName validates a subdirectory name.
//
// Subdirectory names must be simple directory names that will be joined with
// baseDir to create the full path. They cannot contain path separators or
// special directory names.
//
// Validation rules:
//   - Cannot be empty
//   - Cannot contain "/" or "\" (path separators)
//   - Cannot be "." or ".." (special directory names)
//
// Parameters:
//   - name: The subdirectory name to validate
//   - fieldName: The field name for error messages (e.g., "skillsDir")
//
// Returns:
//   - error: Describes the validation failure, or nil if valid
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

// expandHomePath expands ~ to the user's home directory.
//
// This function handles tilde expansion for paths like:
//   - "~" -> user's home directory
//   - "~/Documents" -> user's home directory + /Documents
//   - "/absolute/path" -> unchanged
//   - "relative/path" -> unchanged
//
// Parameters:
//   - path: The path to expand
//
// Returns:
//   - string: The expanded path
//   - error: If the home directory cannot be determined
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

// validateDisplaySettings validates and applies defaults to display settings.
//
// This function ensures that display settings have valid values and applies
// defaults if the settings are missing or invalid.
//
// Validation rules:
//   - colors: Must be "auto", "always", or "never". Defaults to "auto" if empty or invalid.
//   - unicode: Must be "auto", "always", or "ascii". Defaults to "auto" if empty or invalid.
//
// Parameters:
//   - settings: The display settings to validate (may be modified to apply defaults)
//
// Returns:
//   - error: Always returns nil (defaults are applied for invalid values)
func validateDisplaySettings(settings *DisplaySettings) error {
	// Validate and apply defaults for colors setting
	switch settings.Colors {
	case "auto", "always", "never":
		// Valid value, keep as-is
	case "":
		// Empty, apply default
		settings.Colors = "auto"
	default:
		// Invalid value, apply default
		settings.Colors = "auto"
	}

	// Validate and apply defaults for unicode setting
	switch settings.Unicode {
	case "auto", "always", "ascii":
		// Valid value, keep as-is
	case "":
		// Empty, apply default
		settings.Unicode = "auto"
	default:
		// Invalid value, apply default
		settings.Unicode = "auto"
	}

	return nil
}
