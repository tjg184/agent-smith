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

func ExpandHome(path string) (string, error) {
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
	return ExpandHome(AgentsDir)
}

func GetOpencodeDir() (string, error) {
	return ExpandHome(OpencodeDir)
}

func GetClaudeCodeDir() (string, error) {
	return ExpandHome(ClaudeCodeDir)
}

func GetCopilotDir() (string, error) {
	return ExpandHome(CopilotDir)
}

func GetUniversalDir() (string, error) {
	return ExpandHome(UniversalDir)
}

func GetSkillsDir() (string, error) {
	baseDir, err := GetAgentsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, SkillsSubDir), nil
}

func GetAgentsSubDir() (string, error) {
	baseDir, err := GetAgentsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, AgentsSubDir), nil
}

func GetCommandsDir() (string, error) {
	baseDir, err := GetAgentsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, CommandsSubDir), nil
}

func GetDetectionConfigPath() (string, error) {
	configDir, err := GetOpencodeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, DetectionConfigFile), nil
}

func GetComponentLockPath(baseDir, componentType string) string {
	return filepath.Join(baseDir, ComponentLockFile)
}

func GetComponentTypes() []string {
	return []string{SkillsSubDir, AgentsSubDir, CommandsSubDir}
}

func GetComponentTypeNames() []string {
	return []string{AgentsSubDir, CommandsSubDir, SkillsSubDir}
}

func GetProfilesDir() (string, error) {
	baseDir, err := GetAgentsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, ProfilesSubDir), nil
}

func GetProfileDir(profileName string) (string, error) {
	profilesDir, err := GetProfilesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(profilesDir, profileName), nil
}

func ResolveTargetDir(targetDir string) (string, error) {
	if targetDir == "" {
		return "", fmt.Errorf("target directory cannot be empty")
	}

	// First expand tilde if present
	expanded, err := ExpandHome(targetDir)
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
