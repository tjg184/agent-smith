package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	AgentsDir     = "~/.agent-smith"
	OpencodeDir   = "~/.config/opencode"
	ClaudeCodeDir = "~/.claude"
	CopilotDir    = "~/.copilot"
	UniversalDir  = "~/.agents"
)

const (
	SkillsSubDir   = "skills"
	AgentsSubDir   = "agents"
	CommandsSubDir = "commands"
	ProfilesSubDir = "profiles"
)

const (
	ComponentLockFile = ".component-lock.json"
)

const (
	SkillMarkdownFile   = "SKILL.md"
	DetectionConfigFile = "detection-config.json"
)

const (
	BaseProfileName = "(no profile)"
)

const (
	AgentsPathPattern   = "/agents/"
	CommandsPathPattern = "/commands/"
)

var IgnoredPaths = []string{
	".git",
	"node_modules",
	".vscode",
	".idea",
}

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

func GetAgentsDir() (string, error) {
	return expandHome(AgentsDir)
}

func GetOpencodeDir() (string, error) {
	return expandHome(OpencodeDir)
}

func GetClaudeCodeDir() (string, error) {
	return expandHome(ClaudeCodeDir)
}

func GetCopilotDir() (string, error) {
	return expandHome(CopilotDir)
}

func GetUniversalDir() (string, error) {
	return expandHome(UniversalDir)
}

// GetSkillsDir returns the full path to the skills directory
func GetSkillsDir() (string, error) {
	baseDir, err := GetAgentsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, SkillsSubDir), nil
}

// GetAgentsSubDir returns the full path to the agents subdirectory
func GetAgentsSubDir() (string, error) {
	baseDir, err := GetAgentsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, AgentsSubDir), nil
}

// GetCommandsDir returns the full path to the commands directory
func GetCommandsDir() (string, error) {
	baseDir, err := GetAgentsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, CommandsSubDir), nil
}

// GetDetectionConfigPath returns the full path to the detection config file
func GetDetectionConfigPath() (string, error) {
	configDir, err := GetOpencodeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, DetectionConfigFile), nil
}

// GetComponentLockPath returns the full path to the unified component lock file
func GetComponentLockPath(baseDir, componentType string) string {
	return filepath.Join(baseDir, ComponentLockFile)
}

// GetComponentTypes returns the list of valid component types
func GetComponentTypes() []string {
	return []string{SkillsSubDir, AgentsSubDir, CommandsSubDir}
}

// GetComponentTypeNames returns the list of component type names for display
func GetComponentTypeNames() []string {
	return []string{AgentsSubDir, CommandsSubDir, SkillsSubDir}
}

// GetProfilesDir returns the full path to the profiles directory
func GetProfilesDir() (string, error) {
	baseDir, err := GetAgentsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, ProfilesSubDir), nil
}

// GetProfileDir returns the full path to a specific profile directory
func GetProfileDir(profileName string) (string, error) {
	profilesDir, err := GetProfilesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(profilesDir, profileName), nil
}

// ResolveTargetDir resolves a custom target directory path
// Supports relative paths, absolute paths, and tilde expansion
func ResolveTargetDir(targetDir string) (string, error) {
	if targetDir == "" {
		return "", fmt.Errorf("target directory cannot be empty")
	}

	// First expand tilde if present
	expanded, err := expandHome(targetDir)
	if err != nil {
		return "", fmt.Errorf("failed to expand home directory: %w", err)
	}

	// Convert to absolute path if relative
	absPath, err := filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	return absPath, nil
}
