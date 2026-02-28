package config

import (
	"fmt"
	"path/filepath"

	"github.com/tjg184/agent-smith/pkg/paths"
)

const copilotProjectDirName = ".github"

// CopilotTarget implements the Target interface for GitHub Copilot
type CopilotTarget struct {
	baseDir string
}

// NewCopilotTarget creates a new CopilotTarget with the default copilot directory
func NewCopilotTarget() (*CopilotTarget, error) {
	baseDir, err := paths.GetCopilotDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get copilot directory: %w", err)
	}

	return &CopilotTarget{
		baseDir: baseDir,
	}, nil
}

// NewCopilotTargetWithDir creates a new CopilotTarget with a custom directory
// This is useful for testing or custom configurations
func NewCopilotTargetWithDir(dir string) *CopilotTarget {
	return &CopilotTarget{
		baseDir: dir,
	}
}

// GetGlobalBaseDir returns the base copilot directory
func (t *CopilotTarget) GetGlobalBaseDir() (string, error) {
	return t.baseDir, nil
}

// GetGlobalSkillsDir returns the directory where skills should be linked
func (t *CopilotTarget) GetGlobalSkillsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.SkillsSubDir), nil
}

// GetGlobalAgentsDir returns the directory where agents should be linked
func (t *CopilotTarget) GetGlobalAgentsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.AgentsSubDir), nil
}

// GetGlobalCommandsDir returns the directory where commands should be linked
func (t *CopilotTarget) GetGlobalCommandsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.CommandsSubDir), nil
}

// GetGlobalComponentDir returns the directory for a specific component type
func (t *CopilotTarget) GetGlobalComponentDir(componentType string) (string, error) {
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
func (t *CopilotTarget) GetDetectionConfigPath() (string, error) {
	return filepath.Join(t.baseDir, paths.DetectionConfigFile), nil
}

// GetName returns the human-readable name of this target
func (t *CopilotTarget) GetName() string {
	return "copilot"
}

// GetProjectDirName returns the directory name used in projects
func (t *CopilotTarget) GetProjectDirName() string {
	return copilotProjectDirName
}

// GetProjectBaseDir returns the base directory within a project
func (t *CopilotTarget) GetProjectBaseDir(projectRoot string) string {
	return filepath.Join(projectRoot, copilotProjectDirName)
}

// GetProjectComponentDir returns the component directory within a project
func (t *CopilotTarget) GetProjectComponentDir(projectRoot, componentType string) (string, error) {
	return filepath.Join(projectRoot, copilotProjectDirName, componentType), nil
}

// IsUniversalTarget returns false for copilot (it's editor-specific)
func (t *CopilotTarget) IsUniversalTarget() bool {
	return false
}
