package config

import (
	"fmt"
	"path/filepath"

	"github.com/tgaines/agent-smith/pkg/paths"
)

// ClaudeCodeTarget implements the Target interface for the claudecode configuration system
type ClaudeCodeTarget struct {
	baseDir string
}

// NewClaudeCodeTarget creates a new ClaudeCodeTarget with the default claudecode directory
func NewClaudeCodeTarget() (*ClaudeCodeTarget, error) {
	baseDir, err := paths.GetClaudeCodeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get claudecode directory: %w", err)
	}

	return &ClaudeCodeTarget{
		baseDir: baseDir,
	}, nil
}

// NewClaudeCodeTargetWithDir creates a new ClaudeCodeTarget with a custom directory
// This is useful for testing or custom configurations
func NewClaudeCodeTargetWithDir(dir string) *ClaudeCodeTarget {
	return &ClaudeCodeTarget{
		baseDir: dir,
	}
}

// GetBaseDir returns the base claudecode directory
func (t *ClaudeCodeTarget) GetBaseDir() (string, error) {
	return t.baseDir, nil
}

// GetSkillsDir returns the directory where skills should be linked
func (t *ClaudeCodeTarget) GetSkillsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.SkillsSubDir), nil
}

// GetAgentsDir returns the directory where agents should be linked
func (t *ClaudeCodeTarget) GetAgentsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.AgentsSubDir), nil
}

// GetCommandsDir returns the directory where commands should be linked
func (t *ClaudeCodeTarget) GetCommandsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.CommandsSubDir), nil
}

// GetComponentDir returns the directory for a specific component type
func (t *ClaudeCodeTarget) GetComponentDir(componentType string) (string, error) {
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
func (t *ClaudeCodeTarget) GetDetectionConfigPath() (string, error) {
	return filepath.Join(t.baseDir, paths.DetectionConfigFile), nil
}

// GetName returns the human-readable name of this target
func (t *ClaudeCodeTarget) GetName() string {
	return "claudecode"
}
