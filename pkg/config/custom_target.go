package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/tjg184/agent-smith/pkg/paths"
)

// CustomTarget implements the Target interface for custom user-defined targets
type CustomTarget struct {
	name        string
	baseDir     string
	skillsDir   string
	agentsDir   string
	commandsDir string
	projectDir  string
}

// NewCustomTarget creates a new CustomTarget from a CustomTargetConfig
func NewCustomTarget(config CustomTargetConfig) (*CustomTarget, error) {
	// Validate the config
	if err := validateCustomTargetConfig(&config); err != nil {
		return nil, err
	}

	// Expand base directory path
	expandedBaseDir, err := paths.ExpandHome(config.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to expand base directory: %w", err)
	}

	// Convert to absolute path
	absBaseDir, err := filepath.Abs(expandedBaseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base directory: %w", err)
	}

	// Validate projectDir if provided (required for materialize support)
	projectDir := config.ProjectDir
	if projectDir == "" {
		return nil, fmt.Errorf("custom target '%s' must have ProjectDir configured for materialize support", config.Name)
	}

	return &CustomTarget{
		name:        config.Name,
		baseDir:     absBaseDir,
		skillsDir:   config.SkillsDir,
		agentsDir:   config.AgentsDir,
		commandsDir: config.CommandsDir,
		projectDir:  projectDir,
	}, nil
}

// GetGlobalBaseDir returns the base directory for this target
func (t *CustomTarget) GetGlobalBaseDir() (string, error) {
	return t.baseDir, nil
}

// GetGlobalSkillsDir returns the directory where skills should be linked
func (t *CustomTarget) GetGlobalSkillsDir() (string, error) {
	return filepath.Join(t.baseDir, t.skillsDir), nil
}

// GetGlobalAgentsDir returns the directory where agents should be linked
func (t *CustomTarget) GetGlobalAgentsDir() (string, error) {
	return filepath.Join(t.baseDir, t.agentsDir), nil
}

// GetGlobalCommandsDir returns the directory where commands should be linked
func (t *CustomTarget) GetGlobalCommandsDir() (string, error) {
	return filepath.Join(t.baseDir, t.commandsDir), nil
}

// GetGlobalComponentDir returns the directory for a specific component type
func (t *CustomTarget) GetGlobalComponentDir(componentType string) (string, error) {
	switch componentType {
	case "skills":
		return t.GetGlobalSkillsDir()
	case "agents":
		return t.GetGlobalAgentsDir()
	case "commands":
		return t.GetGlobalCommandsDir()
	default:
		return "", fmt.Errorf("unknown component type: %s", componentType)
	}
}

// GetDetectionConfigPath returns the path to the detection config file
func (t *CustomTarget) GetDetectionConfigPath() (string, error) {
	return filepath.Join(t.baseDir, "detection-config.json"), nil
}

func (t *CustomTarget) GetName() string {
	return t.name
}

// GetDisplayName returns a title-cased display name derived from the custom target's machine name.
func (t *CustomTarget) GetDisplayName() string {
	replaced := strings.ReplaceAll(t.name, "-", " ")
	replaced = strings.ReplaceAll(replaced, "_", " ")
	words := strings.Fields(replaced)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

// GetProjectDirName returns the directory name used in projects
func (t *CustomTarget) GetProjectDirName() string {
	return t.projectDir
}

// GetProjectBaseDir returns the base directory within a project
func (t *CustomTarget) GetProjectBaseDir(projectRoot string) string {
	return filepath.Join(projectRoot, t.projectDir)
}

// GetProjectComponentDir returns the component directory within a project
func (t *CustomTarget) GetProjectComponentDir(projectRoot, componentType string) (string, error) {
	return filepath.Join(projectRoot, t.projectDir, componentType), nil
}

// IsUniversalTarget returns false for custom targets (they are editor-specific)
func (t *CustomTarget) IsUniversalTarget() bool {
	return false
}

// IsCustom returns true to indicate this is a custom target (not built-in)
func (t *CustomTarget) IsCustom() bool {
	return true
}
