package config

import (
	"fmt"
	"path/filepath"

	"github.com/tjg184/agent-smith/pkg/paths"
)

// baseTarget provides shared field storage and method implementations for all Target types.
// Concrete targets embed this struct and only need to implement GetName() and IsUniversalTarget().
type baseTarget struct {
	baseDir        string
	projectDirName string
}

func (t *baseTarget) GetGlobalBaseDir() (string, error) {
	return t.baseDir, nil
}

func (t *baseTarget) GetGlobalSkillsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.SkillsSubDir), nil
}

func (t *baseTarget) GetGlobalAgentsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.AgentsSubDir), nil
}

func (t *baseTarget) GetGlobalCommandsDir() (string, error) {
	return filepath.Join(t.baseDir, paths.CommandsSubDir), nil
}

func (t *baseTarget) GetGlobalComponentDir(componentType string) (string, error) {
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

func (t *baseTarget) GetDetectionConfigPath() (string, error) {
	return filepath.Join(t.baseDir, paths.DetectionConfigFile), nil
}

func (t *baseTarget) GetProjectDirName() string {
	return t.projectDirName
}

func (t *baseTarget) GetProjectBaseDir(projectRoot string) string {
	return filepath.Join(projectRoot, t.projectDirName)
}

func (t *baseTarget) GetProjectComponentDir(projectRoot, componentType string) (string, error) {
	return filepath.Join(projectRoot, t.projectDirName, componentType), nil
}

// GetDisplayName provides a default title-cased display name for targets that don't override it.
// Concrete built-in targets should override this with their canonical display name.
func (t *baseTarget) GetDisplayName() string {
	// baseTarget has no name field; concrete types embed baseTarget and implement GetName().
	// This default is intentionally unreachable for built-in targets (they all override it),
	// and serves as a fallback only if a future embedded type forgets to implement GetDisplayName.
	return ""
}
