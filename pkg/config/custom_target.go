package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/tjg184/agent-smith/pkg/paths"
)

type CustomTarget struct {
	name        string
	baseDir     string
	skillsDir   string
	agentsDir   string
	commandsDir string
	projectDir  string
}

func NewCustomTarget(config CustomTargetConfig) (*CustomTarget, error) {
	if err := validateCustomTargetConfig(&config); err != nil {
		return nil, err
	}

	expandedBaseDir, err := paths.ExpandHome(config.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to expand base directory: %w", err)
	}

	absBaseDir, err := filepath.Abs(expandedBaseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base directory: %w", err)
	}

	projectDir := config.ProjectDir
	if projectDir == "" {
		return nil, fmt.Errorf("custom target '%s' must have ProjectDir configured for materialize support", config.Name)
	}

	_ = absBaseDir

	return &CustomTarget{
		name:        config.Name,
		baseDir:     absBaseDir,
		skillsDir:   config.SkillsDir,
		agentsDir:   config.AgentsDir,
		commandsDir: config.CommandsDir,
		projectDir:  projectDir,
	}, nil
}

func (t *CustomTarget) GetGlobalBaseDir() (string, error) {
	return t.baseDir, nil
}

func (t *CustomTarget) GetGlobalSkillsDir() (string, error) {
	return filepath.Join(t.baseDir, t.skillsDir), nil
}

func (t *CustomTarget) GetGlobalAgentsDir() (string, error) {
	return filepath.Join(t.baseDir, t.agentsDir), nil
}

func (t *CustomTarget) GetGlobalCommandsDir() (string, error) {
	return filepath.Join(t.baseDir, t.commandsDir), nil
}

func (t *CustomTarget) GetGlobalComponentDir(componentType string) (string, error) {
	switch componentType {
	case "skills":
		return t.GetGlobalSkillsDir()
	case "agents":
		return t.GetGlobalAgentsDir()
	case "commands":
		return t.GetGlobalCommandsDir()
	default:
		return "", fmt.Errorf("unknown component type: %s", componentType)
	}
}

func (t *CustomTarget) GetDetectionConfigPath() (string, error) {
	return filepath.Join(t.baseDir, "detection-config.json"), nil
}

func (t *CustomTarget) GetName() string {
	return t.name
}

func (t *CustomTarget) GetDisplayName() string {
	replaced := strings.ReplaceAll(t.name, "-", " ")
	replaced = strings.ReplaceAll(replaced, "_", " ")
	words := strings.Fields(replaced)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

func (t *CustomTarget) GetProjectDirName() string {
	return t.projectDir
}

func (t *CustomTarget) GetProjectBaseDir(projectRoot string) string {
	return filepath.Join(projectRoot, t.projectDir)
}

func (t *CustomTarget) GetProjectComponentDir(projectRoot, componentType string) (string, error) {
	return filepath.Join(projectRoot, t.projectDir, componentType), nil
}

func (t *CustomTarget) IsUniversalTarget() bool {
	return false
}

func (t *CustomTarget) IsCustom() bool {
	return true
}
