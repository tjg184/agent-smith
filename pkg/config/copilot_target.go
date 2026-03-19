package config

import (
	"fmt"

	"github.com/tjg184/agent-smith/pkg/paths"
)

const copilotProjectDirName = ".github"

type CopilotTarget struct {
	baseTarget
}

func NewCopilotTarget() (*CopilotTarget, error) {
	baseDir, err := paths.GetCopilotDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get copilot directory: %w", err)
	}

	return &CopilotTarget{baseTarget{baseDir: baseDir, projectDirName: copilotProjectDirName}}, nil
}

func NewCopilotTargetWithDir(dir string) *CopilotTarget {
	return &CopilotTarget{baseTarget{baseDir: dir, projectDirName: copilotProjectDirName}}
}

func (t *CopilotTarget) GetName() string {
	return "copilot"
}

func (t *CopilotTarget) GetDisplayName() string {
	return "GitHub Copilot"
}

func (t *CopilotTarget) IsUniversalTarget() bool {
	return false
}
