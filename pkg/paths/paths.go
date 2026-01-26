package paths

import (
	"os"
	"path/filepath"
)

// Base directory constants
const (
	AgentsDir     = "~/.agents"
	OpencodeDir   = "~/.config/opencode"
	ClaudeCodeDir = "~/.config/claudecode"
)

// Component subdirectory names
const (
	SkillsSubDir   = "skills"
	AgentsSubDir   = "agents"
	CommandsSubDir = "commands"
)

// Metadata file patterns
const (
	SkillMetadataFile   = ".skill-metadata.json"
	AgentMetadataFile   = ".agent-metadata.json"
	CommandMetadataFile = ".command-metadata.json"
)

// Lock file patterns
const (
	SkillLockFile   = ".skill-lock.json"
	AgentLockFile   = ".agent-lock.json"
	CommandLockFile = ".command-lock.json"
)

// Special files
const (
	SkillMarkdownFile   = "SKILL.md"
	DetectionConfigFile = "detection-config.json"
)

// Path patterns for detection
const (
	AgentsPathPattern   = "/agents/"
	CommandsPathPattern = "/commands/"
)

// Ignored paths
var IgnoredPaths = []string{
	".git",
	"node_modules",
	".vscode",
	".idea",
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

// GetAgentsDir returns the expanded agents directory path
func GetAgentsDir() (string, error) {
	return expandHome(AgentsDir)
}

// GetOpencodeDir returns the expanded opencode config directory path
func GetOpencodeDir() (string, error) {
	return expandHome(OpencodeDir)
}

// GetClaudeCodeDir returns the expanded claudecode config directory path
func GetClaudeCodeDir() (string, error) {
	return expandHome(ClaudeCodeDir)
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

// GetComponentMetadataPath returns the full path to a component's metadata file
func GetComponentMetadataPath(baseDir, componentType, componentName string) string {
	var metadataFile string
	var subDir string

	switch componentType {
	case "skills":
		subDir = SkillsSubDir
		metadataFile = SkillMetadataFile
	case "agents":
		subDir = AgentsSubDir
		metadataFile = AgentMetadataFile
	case "commands":
		subDir = CommandsSubDir
		metadataFile = CommandMetadataFile
	default:
		return ""
	}

	return filepath.Join(baseDir, subDir, componentName, metadataFile)
}

// GetComponentLockPath returns the full path to a component type's lock file
func GetComponentLockPath(baseDir, componentType string) string {
	var lockFile string

	switch componentType {
	case "skills":
		lockFile = SkillLockFile
	case "agents":
		lockFile = AgentLockFile
	case "commands":
		lockFile = CommandLockFile
	default:
		return ""
	}

	return filepath.Join(baseDir, lockFile)
}

// GetComponentTypes returns the list of valid component types
func GetComponentTypes() []string {
	return []string{SkillsSubDir, AgentsSubDir, CommandsSubDir}
}

// GetComponentTypeNames returns the list of component type names for display
func GetComponentTypeNames() []string {
	return []string{AgentsSubDir, CommandsSubDir, SkillsSubDir}
}
