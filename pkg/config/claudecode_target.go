package config

import (
	"fmt"

	"github.com/tjg184/agent-smith/pkg/paths"
)

const claudeCodeProjectDirName = ".claude"

type ClaudeCodeTarget struct {
	baseTarget
}

func NewClaudeCodeTarget() (*ClaudeCodeTarget, error) {
	baseDir, err := paths.GetClaudeCodeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get claudecode directory: %w", err)
	}

	return &ClaudeCodeTarget{baseTarget{baseDir: baseDir, projectDirName: claudeCodeProjectDirName}}, nil
}

func NewClaudeCodeTargetWithDir(dir string) *ClaudeCodeTarget {
	return &ClaudeCodeTarget{baseTarget{baseDir: dir, projectDirName: claudeCodeProjectDirName}}
}

func (t *ClaudeCodeTarget) GetName() string {
	return "claudecode"
}

func (t *ClaudeCodeTarget) GetDisplayName() string {
	return "Claude Code"
}

func (t *ClaudeCodeTarget) IsUniversalTarget() bool {
	return false
}
