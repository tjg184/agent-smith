package config

import (
	"fmt"

	"github.com/tjg184/agent-smith/pkg/paths"
)

const copilotProjectDirName = ".github"

// CopilotTarget implements the Target interface for GitHub Copilot
type CopilotTarget struct {
	baseTarget
}

// NewCopilotTarget creates a new CopilotTarget with the default copilot directory
func NewCopilotTarget() (*CopilotTarget, error) {
	baseDir, err := paths.GetCopilotDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get copilot directory: %w", err)
	}

	return &CopilotTarget{baseTarget{baseDir: baseDir, projectDirName: copilotProjectDirName}}, nil
}

// NewCopilotTargetWithDir creates a new CopilotTarget with a custom directory
// This is useful for testing or custom configurations
func NewCopilotTargetWithDir(dir string) *CopilotTarget {
	return &CopilotTarget{baseTarget{baseDir: dir, projectDirName: copilotProjectDirName}}
}

func (t *CopilotTarget) GetName() string {
	return "copilot"
}

func (t *CopilotTarget) GetDisplayName() string {
	return "GitHub Copilot"
}

// IsUniversalTarget returns false for copilot (it's editor-specific)
func (t *CopilotTarget) IsUniversalTarget() bool {
	return false
}
