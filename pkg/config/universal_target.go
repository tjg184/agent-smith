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

func NewUniversalTarget() (*UniversalTarget, error) {
	baseDir, err := paths.GetUniversalDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get universal directory: %w", err)
	}

	return &UniversalTarget{
		baseDir: baseDir,
	}, nil
}

func NewUniversalTargetWithDir(dir string) *UniversalTarget {
	return &UniversalTarget{
		baseDir: dir,
	}
}

func (t *UniversalTarget) GetGlobalBaseDir() (string, error) {
	return t.baseDir, nil
}

func (t *UniversalTarget) GetGlobalSkillsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.SkillsSubDir), nil
}

func (t *UniversalTarget) GetGlobalAgentsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.AgentsSubDir), nil
}

func (t *UniversalTarget) GetGlobalCommandsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.CommandsSubDir), nil
}

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

func (t *UniversalTarget) GetDetectionConfigPath() (string, error) {
	return filepath.Join(t.baseDir, paths.DetectionConfigFile), nil
}

func (t *UniversalTarget) GetName() string {
	return "universal"
}

func (t *UniversalTarget) GetProjectDirName() string {
	return universalProjectDirName
}

func (t *UniversalTarget) GetProjectBaseDir(projectRoot string) string {
	return filepath.Join(projectRoot, universalProjectDirName)
}

func (t *UniversalTarget) GetProjectComponentDir(projectRoot, componentType string) (string, error) {
	return filepath.Join(projectRoot, universalProjectDirName, componentType), nil
}

func (t *UniversalTarget) IsUniversalTarget() bool {
	return true
}
