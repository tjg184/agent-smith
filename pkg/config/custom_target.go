package config

import (
	"fmt"
	"path/filepath"
)

// CustomTarget implements the Target interface for custom user-defined targets
type CustomTarget struct {
	name        string
	baseDir     string
	skillsDir   string
	agentsDir   string
	commandsDir string
}

// NewCustomTarget creates a new CustomTarget from a CustomTargetConfig
func NewCustomTarget(config CustomTargetConfig) (*CustomTarget, error) {
	// Validate the config
	if err := validateCustomTargetConfig(&config); err != nil {
		return nil, err
	}

	// Expand base directory path
	expandedBaseDir, err := expandHomePath(config.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to expand base directory: %w", err)
	}

	// Convert to absolute path
	absBaseDir, err := filepath.Abs(expandedBaseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base directory: %w", err)
	}

	return &CustomTarget{
		name:        config.Name,
		baseDir:     absBaseDir,
		skillsDir:   config.SkillsDir,
		agentsDir:   config.AgentsDir,
		commandsDir: config.CommandsDir,
	}, nil
}

// GetBaseDir returns the base directory for this target
func (t *CustomTarget) GetBaseDir() (string, error) {
	return t.baseDir, nil
}

// GetSkillsDir returns the directory where skills should be linked
func (t *CustomTarget) GetSkillsDir() (string, error) {
	return filepath.Join(t.baseDir, t.skillsDir), nil
}

// GetAgentsDir returns the directory where agents should be linked
func (t *CustomTarget) GetAgentsDir() (string, error) {
	return filepath.Join(t.baseDir, t.agentsDir), nil
}

// GetCommandsDir returns the directory where commands should be linked
func (t *CustomTarget) GetCommandsDir() (string, error) {
	return filepath.Join(t.baseDir, t.commandsDir), nil
}

// GetComponentDir returns the directory for a specific component type
func (t *CustomTarget) GetComponentDir(componentType string) (string, error) {
	switch componentType {
	case "skills":
		return t.GetSkillsDir()
	case "agents":
		return t.GetAgentsDir()
	case "commands":
		return t.GetCommandsDir()
	default:
		return "", fmt.Errorf("unknown component type: %s", componentType)
	}
}

// GetDetectionConfigPath returns the path to the detection config file
func (t *CustomTarget) GetDetectionConfigPath() (string, error) {
	return filepath.Join(t.baseDir, "detection-config.json"), nil
}

// GetName returns the human-readable name of this target
func (t *CustomTarget) GetName() string {
	return t.name
}

// IsCustom returns true to indicate this is a custom target (not built-in)
func (t *CustomTarget) IsCustom() bool {
	return true
}
