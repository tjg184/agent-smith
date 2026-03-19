package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type TargetType string

const (
	TargetOpenCode   TargetType = "opencode"
	TargetClaudeCode TargetType = "claudecode"
	TargetCopilot    TargetType = "copilot"
	TargetUniversal  TargetType = "universal"
)

// builtInTargetDef holds constructor functions for a single built-in target.
// Adding a new built-in target requires only a new entry here plus a *_target.go file.
type builtInTargetDef struct {
	targetType         TargetType
	constructor        func() (Target, error)
	projectConstructor func(dir string) Target
	isUniversal        bool
}

// builtInTargetDefs is the single authoritative registry of all built-in targets.
// Order determines detection priority: OpenCode > Claude Code > Copilot > Universal.
var builtInTargetDefs = []builtInTargetDef{
	{
		targetType:         TargetOpenCode,
		constructor:        func() (Target, error) { return NewOpencodeTarget() },
		projectConstructor: func(dir string) Target { return NewOpencodeTargetWithDir(dir) },
	},
	{
		targetType:         TargetClaudeCode,
		constructor:        func() (Target, error) { return NewClaudeCodeTarget() },
		projectConstructor: func(dir string) Target { return NewClaudeCodeTargetWithDir(dir) },
	},
	{
		targetType:         TargetCopilot,
		constructor:        func() (Target, error) { return NewCopilotTarget() },
		projectConstructor: func(dir string) Target { return NewCopilotTargetWithDir(dir) },
	},
	{
		targetType:         TargetUniversal,
		constructor:        func() (Target, error) { return NewUniversalTarget() },
		projectConstructor: func(dir string) Target { return NewUniversalTargetWithDir(dir) },
		isUniversal:        true,
	},
}

func builtInTargetNames() []string {
	names := make([]string, 0, len(builtInTargetDefs))
	for _, def := range builtInTargetDefs {
		names = append(names, string(def.targetType))
	}
	return names
}

func GetTargetFromEnv() string {
	return os.Getenv("AGENT_SMITH_TARGET")
}

func NewTarget(targetType string) (Target, error) {
	if targetType == "" {
		targetType = string(TargetOpenCode)
	}

	for _, def := range builtInTargetDefs {
		if string(def.targetType) == targetType {
			return def.constructor()
		}
	}

	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	for _, customTargetConfig := range config.CustomTargets {
		if customTargetConfig.Name == targetType {
			return NewCustomTarget(customTargetConfig)
		}
	}

	return nil, fmt.Errorf("unknown target type: %s (valid options: opencode, claudecode, copilot, universal, or custom targets from config)", targetType)
}

func NewTargetForProject(targetType, projectRoot string) (Target, error) {
	if targetType == "" {
		targetType = string(TargetOpenCode)
	}

	for _, def := range builtInTargetDefs {
		if string(def.targetType) == targetType {
			tmp, err := def.constructor()
			if err != nil {
				return nil, fmt.Errorf("failed to create %s target: %w", targetType, err)
			}
			return def.projectConstructor(filepath.Join(projectRoot, tmp.GetProjectDirName())), nil
		}
	}

	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	for _, customTargetConfig := range config.CustomTargets {
		if customTargetConfig.Name == targetType {
			customTargetConfig.BaseDir = filepath.Join(projectRoot, customTargetConfig.ProjectDir)
			return NewCustomTarget(customTargetConfig)
		}
	}

	return nil, fmt.Errorf("unknown target type: %s (valid options: opencode, claudecode, copilot, universal, or custom targets from config)", targetType)
}

// Priority: AGENT_SMITH_TARGET env var > auto-detection > default to OpenCode.
func DetectTarget() (Target, error) {
	envTarget := GetTargetFromEnv()
	if envTarget != "" {
		return NewTarget(envTarget)
	}

	// Auto-detect by checking which directories exist; skip universal (opt-in only).
	for _, def := range builtInTargetDefs {
		if def.isUniversal {
			continue
		}
		target, err := def.constructor()
		if err != nil {
			continue
		}
		baseDir, _ := target.GetGlobalBaseDir()
		if _, err := os.Stat(baseDir); err == nil {
			return target, nil
		}
	}

	return NewOpencodeTarget()
}

func GetAvailableTargets() []string {
	var names []string
	for _, def := range builtInTargetDefs {
		if !def.isUniversal {
			names = append(names, string(def.targetType))
		}
	}
	return names
}

func GetAllTargetTypes() []string {
	names := make([]string, 0, len(builtInTargetDefs))
	for _, def := range builtInTargetDefs {
		names = append(names, string(def.targetType))
	}
	return names
}

func DetectAllTargets() ([]Target, error) {
	var targets []Target

	for _, def := range builtInTargetDefs {
		target, err := def.constructor()
		if err != nil {
			continue
		}
		baseDir, _ := target.GetGlobalBaseDir()
		if _, err := os.Stat(baseDir); err == nil {
			targets = append(targets, target)
		}
	}

	config, err := LoadConfig()
	if err == nil && config != nil {
		for _, customTargetConfig := range config.CustomTargets {
			customTarget, err := NewCustomTarget(customTargetConfig)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to load custom target %s: %v\n", customTargetConfig.Name, err)
				continue
			}
			baseDir, _ := customTarget.GetGlobalBaseDir()
			if _, err := os.Stat(baseDir); err == nil {
				targets = append(targets, customTarget)
			}
		}
	}

	if len(targets) == 0 {
		opencodeTarget, err := NewOpencodeTarget()
		if err != nil {
			return nil, err
		}
		targets = append(targets, opencodeTarget)
	}

	return targets, nil
}

func GetAllTargetProjectDirNames() []string {
	names := make([]string, 0, len(builtInTargetDefs))
	for _, def := range builtInTargetDefs {
		t, err := def.constructor()
		if err != nil {
			continue
		}
		names = append(names, t.GetProjectDirName())
	}
	return names
}

func GetAllTargets() ([]Target, error) {
	var targets []Target
	for _, def := range builtInTargetDefs {
		if def.isUniversal {
			continue
		}
		target, err := def.constructor()
		if err != nil {
			return nil, fmt.Errorf("failed to create %s target: %w", def.targetType, err)
		}
		targets = append(targets, target)
	}
	return targets, nil
}
