package config

import (
	"fmt"

	"github.com/tjg184/agent-smith/pkg/paths"
)

const universalProjectDirName = ".agents"

// UniversalTarget implements the Target interface for the universal (.agents) directory.
// This provides a target-agnostic location for materialized components that can be used
// by any AI coding assistant.
type UniversalTarget struct {
	baseTarget
}

func NewUniversalTarget() (*UniversalTarget, error) {
	baseDir, err := paths.GetUniversalDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get universal directory: %w", err)
	}

	return &UniversalTarget{baseTarget{baseDir: baseDir, projectDirName: universalProjectDirName}}, nil
}

func NewUniversalTargetWithDir(dir string) *UniversalTarget {
	return &UniversalTarget{baseTarget{baseDir: dir, projectDirName: universalProjectDirName}}
}

func (t *UniversalTarget) GetName() string {
	return "universal"
}

func (t *UniversalTarget) IsUniversalTarget() bool {
	return true
}
