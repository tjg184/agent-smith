package config

import (
	"fmt"
	"path/filepath"

	"github.com/tjg184/agent-smith/pkg/paths"
)

const universalProjectDirName = ".agents"

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

// GetGlobalBaseDir returns the base universal directory
// For universal target, this is always empty as it's project-local only
func (t *UniversalTarget) GetGlobalBaseDir() (string, error) {
	if t.baseDir == "" {
		return "", fmt.Errorf("universal target has no global directory (project-local only)")
	}
	return t.baseDir, nil
}

// GetGlobalSkillsDir returns the directory where skills should be linked
func (t *UniversalTarget) GetGlobalSkillsDir() (string, error) {
	if t.baseDir == "" {
		return "", fmt.Errorf("universal target requires project context")
	}
	return filepath.Join(t.baseDir, paths.SkillsSubDir), nil
}

// GetGlobalAgentsDir returns the directory where agents should be linked
func (t *UniversalTarget) GetGlobalAgentsDir() (string, error) {
	if t.baseDir == "" {
		return "", fmt.Errorf("universal target requires project context")
	}
	return filepath.Join(t.baseDir, paths.AgentsSubDir), nil
}

// GetGlobalCommandsDir returns the directory where commands should be linked
func (t *UniversalTarget) GetGlobalCommandsDir() (string, error) {
	if t.baseDir == "" {
		return "", fmt.Errorf("universal target requires project context")
	}
	return filepath.Join(t.baseDir, paths.CommandsSubDir), nil
}

// GetGlobalComponentDir returns the directory for a specific component type
func (t *UniversalTarget) GetGlobalComponentDir(componentType string) (string, error) {
	switch componentType {
	case paths.SkillsSubDir:
		return t.GetGlobalSkillsDir()
	case paths.AgentsSubDir:
		return t.GetGlobalAgentsDir()
	case paths.CommandsSubDir:
		return t.GetGlobalCommandsDir()
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

// GetProjectDirName returns the directory name used in projects
func (t *UniversalTarget) GetProjectDirName() string {
	return universalProjectDirName
}

// GetProjectBaseDir returns the base directory within a project
func (t *UniversalTarget) GetProjectBaseDir(projectRoot string) string {
	return filepath.Join(projectRoot, universalProjectDirName)
}

// GetProjectComponentDir returns the component directory within a project
func (t *UniversalTarget) GetProjectComponentDir(projectRoot, componentType string) (string, error) {
	return filepath.Join(projectRoot, universalProjectDirName, componentType), nil
}

// IsUniversalTarget returns true for universal target
func (t *UniversalTarget) IsUniversalTarget() bool {
	return true
}
