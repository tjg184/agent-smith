package config

import (
	"fmt"
	"path/filepath"

	"github.com/tjg184/agent-smith/pkg/paths"
)

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

// GetBaseDir returns the base copilot directory
func (t *CopilotTarget) GetBaseDir() (string, error) {
	return t.baseDir, nil
}

// GetSkillsDir returns the directory where skills should be linked
func (t *CopilotTarget) GetSkillsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.SkillsSubDir), nil
}

// GetAgentsDir returns the directory where agents should be linked
func (t *CopilotTarget) GetAgentsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.AgentsSubDir), nil
}

// GetCommandsDir returns the directory where commands should be linked
func (t *CopilotTarget) GetCommandsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.CommandsSubDir), nil
}

// GetComponentDir returns the directory for a specific component type
func (t *CopilotTarget) GetComponentDir(componentType string) (string, error) {
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
func (t *CopilotTarget) GetDetectionConfigPath() (string, error) {
	return filepath.Join(t.baseDir, paths.DetectionConfigFile), nil
}

// GetName returns the human-readable name of this target
func (t *CopilotTarget) GetName() string {
	return "copilot"
}
