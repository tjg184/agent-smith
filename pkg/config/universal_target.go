package config

import (
	"fmt"
	"path/filepath"

	"github.com/tjg184/agent-smith/pkg/paths"
)

// UniversalTarget implements the Target interface for the universal (.agents) directory
// This provides a target-agnostic location for materialized components that can be used
// by any AI coding assistant.
type UniversalTarget struct {
	baseDir string
}

// NewUniversalTarget creates a new UniversalTarget
// Note: Universal target is project-local only (no global directory)
func NewUniversalTarget() (*UniversalTarget, error) {
	return &UniversalTarget{
		baseDir: "", // No global directory for universal target
	}, nil
}

// NewUniversalTargetWithDir creates a new UniversalTarget with a custom directory
// This is useful for testing or custom configurations
func NewUniversalTargetWithDir(dir string) *UniversalTarget {
	return &UniversalTarget{
		baseDir: dir,
	}
}

// GetBaseDir returns the base universal directory
// For universal target, this is always empty as it's project-local only
func (t *UniversalTarget) GetBaseDir() (string, error) {
	if t.baseDir == "" {
		return "", fmt.Errorf("universal target has no global directory (project-local only)")
	}
	return t.baseDir, nil
}

// GetSkillsDir returns the directory where skills should be linked
func (t *UniversalTarget) GetSkillsDir() (string, error) {
	if t.baseDir == "" {
		return "", fmt.Errorf("universal target requires project context")
	}
	return filepath.Join(t.baseDir, paths.SkillsSubDir), nil
}

// GetAgentsDir returns the directory where agents should be linked
func (t *UniversalTarget) GetAgentsDir() (string, error) {
	if t.baseDir == "" {
		return "", fmt.Errorf("universal target requires project context")
	}
	return filepath.Join(t.baseDir, paths.AgentsSubDir), nil
}

// GetCommandsDir returns the directory where commands should be linked
func (t *UniversalTarget) GetCommandsDir() (string, error) {
	if t.baseDir == "" {
		return "", fmt.Errorf("universal target requires project context")
	}
	return filepath.Join(t.baseDir, paths.CommandsSubDir), nil
}

// GetComponentDir returns the directory for a specific component type
func (t *UniversalTarget) GetComponentDir(componentType string) (string, error) {
	switch componentType {
	case paths.SkillsSubDir:
		return t.GetSkillsDir()
	case paths.AgentsSubDir:
		return t.GetAgentsDir()
	case paths.CommandsSubDir:
		return t.GetCommandsDir()
	default:
		return "", fmt.Errorf("unknown component type: %s", componentType)
	}
}

// GetDetectionConfigPath returns the path to the detection config file
func (t *UniversalTarget) GetDetectionConfigPath() (string, error) {
	if t.baseDir == "" {
		return "", fmt.Errorf("universal target requires project context")
	}
	return filepath.Join(t.baseDir, paths.DetectionConfigFile), nil
}

// GetName returns the human-readable name of this target
func (t *UniversalTarget) GetName() string {
	return "universal"
}
