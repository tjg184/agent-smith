package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tjg184/agent-smith/pkg/paths"
)

const opencodeProjectDirName = ".opencode"

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

// GetGlobalBaseDir returns the base opencode directory
func (t *OpencodeTarget) GetGlobalBaseDir() (string, error) {
	return t.baseDir, nil
}

// GetGlobalSkillsDir returns the directory where skills should be linked
func (t *OpencodeTarget) GetGlobalSkillsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.SkillsSubDir), nil
}

// GetGlobalAgentsDir returns the directory where agents should be linked
func (t *OpencodeTarget) GetGlobalAgentsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.AgentsSubDir), nil
}

// GetGlobalCommandsDir returns the directory where commands should be linked
func (t *OpencodeTarget) GetGlobalCommandsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.CommandsSubDir), nil
}

// GetGlobalComponentDir returns the directory for a specific component type
func (t *OpencodeTarget) GetGlobalComponentDir(componentType string) (string, error) {
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
func (t *OpencodeTarget) GetDetectionConfigPath() (string, error) {
	return filepath.Join(t.baseDir, paths.DetectionConfigFile), nil
}

// GetName returns the human-readable name of this target
func (t *OpencodeTarget) GetName() string {
	return "opencode"
}

// GetProjectDirName returns the directory name used in projects
func (t *OpencodeTarget) GetProjectDirName() string {
	return opencodeProjectDirName
}

// GetProjectBaseDir returns the base directory within a project
func (t *OpencodeTarget) GetProjectBaseDir(projectRoot string) string {
	return filepath.Join(projectRoot, opencodeProjectDirName)
}

// GetProjectComponentDir returns the component directory within a project
func (t *OpencodeTarget) GetProjectComponentDir(projectRoot, componentType string) (string, error) {
	return filepath.Join(projectRoot, opencodeProjectDirName, componentType), nil
}

// IsUniversalTarget returns false for opencode (it's editor-specific)
func (t *OpencodeTarget) IsUniversalTarget() bool {
	return false
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
