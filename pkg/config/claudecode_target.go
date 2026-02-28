package config

import (
	"fmt"
	"path/filepath"

	"github.com/tjg184/agent-smith/pkg/paths"
)

const claudeCodeProjectDirName = ".claude"

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

// GetGlobalBaseDir returns the base claudecode directory
func (t *ClaudeCodeTarget) GetGlobalBaseDir() (string, error) {
	return t.baseDir, nil
}

// GetGlobalSkillsDir returns the directory where skills should be linked
func (t *ClaudeCodeTarget) GetGlobalSkillsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.SkillsSubDir), nil
}

// GetGlobalAgentsDir returns the directory where agents should be linked
func (t *ClaudeCodeTarget) GetGlobalAgentsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.AgentsSubDir), nil
}

// GetGlobalCommandsDir returns the directory where commands should be linked
func (t *ClaudeCodeTarget) GetGlobalCommandsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.CommandsSubDir), nil
}

// GetGlobalComponentDir returns the directory for a specific component type
func (t *ClaudeCodeTarget) GetGlobalComponentDir(componentType string) (string, error) {
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
func (t *ClaudeCodeTarget) GetDetectionConfigPath() (string, error) {
	return filepath.Join(t.baseDir, paths.DetectionConfigFile), nil
}

// GetName returns the human-readable name of this target
func (t *ClaudeCodeTarget) GetName() string {
	return "claudecode"
}

// GetProjectDirName returns the directory name used in projects
func (t *ClaudeCodeTarget) GetProjectDirName() string {
	return claudeCodeProjectDirName
}

// GetProjectBaseDir returns the base directory within a project
func (t *ClaudeCodeTarget) GetProjectBaseDir(projectRoot string) string {
	return filepath.Join(projectRoot, claudeCodeProjectDirName)
}

// GetProjectComponentDir returns the component directory within a project
func (t *ClaudeCodeTarget) GetProjectComponentDir(projectRoot, componentType string) (string, error) {
	return filepath.Join(projectRoot, claudeCodeProjectDirName, componentType), nil
}

// IsUniversalTarget returns false for claudecode (it's editor-specific)
func (t *ClaudeCodeTarget) IsUniversalTarget() bool {
	return false
}
