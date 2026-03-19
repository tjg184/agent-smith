package config

import (
	"fmt"

	"github.com/tjg184/agent-smith/pkg/paths"
)

const claudeCodeProjectDirName = ".claude"

// ClaudeCodeTarget implements the Target interface for the claudecode configuration system
type ClaudeCodeTarget struct {
	baseTarget
}

// NewClaudeCodeTarget creates a new ClaudeCodeTarget with the default claudecode directory
func NewClaudeCodeTarget() (*ClaudeCodeTarget, error) {
	baseDir, err := paths.GetClaudeCodeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get claudecode directory: %w", err)
	}

	return &ClaudeCodeTarget{baseTarget{baseDir: baseDir, projectDirName: claudeCodeProjectDirName}}, nil
}

// NewClaudeCodeTargetWithDir creates a new ClaudeCodeTarget with a custom directory
// This is useful for testing or custom configurations
func NewClaudeCodeTargetWithDir(dir string) *ClaudeCodeTarget {
	return &ClaudeCodeTarget{baseTarget{baseDir: dir, projectDirName: claudeCodeProjectDirName}}
}

func (t *ClaudeCodeTarget) GetName() string {
	return "claudecode"
}

func (t *ClaudeCodeTarget) GetDisplayName() string {
	return "Claude Code"
}

// IsUniversalTarget returns false for claudecode (it's editor-specific)
func (t *ClaudeCodeTarget) IsUniversalTarget() bool {
	return false
}
