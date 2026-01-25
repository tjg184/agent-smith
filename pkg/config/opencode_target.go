package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tgaines/agent-smith/pkg/paths"
)

// OpencodeTarget implements the Target interface for the opencode configuration system
type OpencodeTarget struct {
	baseDir string
}

// NewOpencodeTarget creates a new OpencodeTarget with the default opencode directory
func NewOpencodeTarget() (*OpencodeTarget, error) {
	baseDir, err := paths.GetOpencodeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get opencode directory: %w", err)
	}

	return &OpencodeTarget{
		baseDir: baseDir,
	}, nil
}

// NewOpencodeTargetWithDir creates a new OpencodeTarget with a custom directory
// This is useful for testing or custom configurations
func NewOpencodeTargetWithDir(dir string) *OpencodeTarget {
	return &OpencodeTarget{
		baseDir: dir,
	}
}

// GetBaseDir returns the base opencode directory
func (t *OpencodeTarget) GetBaseDir() (string, error) {
	return t.baseDir, nil
}

// GetSkillsDir returns the directory where skills should be linked
func (t *OpencodeTarget) GetSkillsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.SkillsSubDir), nil
}

// GetAgentsDir returns the directory where agents should be linked
func (t *OpencodeTarget) GetAgentsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.AgentsSubDir), nil
}

// GetCommandsDir returns the directory where commands should be linked
func (t *OpencodeTarget) GetCommandsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.CommandsSubDir), nil
}

// GetComponentDir returns the directory for a specific component type
func (t *OpencodeTarget) GetComponentDir(componentType string) (string, error) {
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
func (t *OpencodeTarget) GetDetectionConfigPath() (string, error) {
	return filepath.Join(t.baseDir, paths.DetectionConfigFile), nil
}

// GetName returns the human-readable name of this target
func (t *OpencodeTarget) GetName() string {
	return "opencode"
}

// expandHome expands ~ to the user's home directory
func expandHome(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if len(path) == 1 {
		return home, nil
	}

	if path[1] == filepath.Separator {
		return filepath.Join(home, path[2:]), nil
	}

	return path, nil
}
