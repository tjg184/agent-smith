package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// TargetType represents the type of target environment
type TargetType string

const (
	// TargetOpenCode represents the OpenCode environment
	TargetOpenCode TargetType = "opencode"
	// TargetClaudeCode represents the Claude Code environment
	TargetClaudeCode TargetType = "claudecode"
	// TargetCopilot represents the GitHub Copilot environment
	TargetCopilot TargetType = "copilot"
	// TargetUniversal represents the universal (.agents) environment
	TargetUniversal TargetType = "universal"
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

// builtInTargetNames returns the machine name of every registered built-in target.
func builtInTargetNames() []string {
	names := make([]string, 0, len(builtInTargetDefs))
	for _, def := range builtInTargetDefs {
		names = append(names, string(def.targetType))
	}
	return names
}

// GetTargetFromEnv returns the target specified by the AGENT_SMITH_TARGET environment variable.
// Returns empty string if not set.
func GetTargetFromEnv() string {
	return os.Getenv("AGENT_SMITH_TARGET")
}

// NewTarget creates a new Target based on the specified target type.
// If targetType is empty, defaults to OpenCode.
// Also checks custom targets from config file.
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

// NewTargetForProject creates a Target configured for a specific project.
// The returned Target will use project-relative paths based on the provided projectRoot.
// For built-in targets, this returns a Target with the appropriate project directory name.
// For custom targets, this validates that ProjectDir is configured.
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

// DetectTarget attempts to detect which target environment is available.
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

// GetAvailableTargets returns the machine names of all non-universal built-in targets.
func GetAvailableTargets() []string {
	var names []string
	for _, def := range builtInTargetDefs {
		if !def.isUniversal {
			names = append(names, string(def.targetType))
		}
	}
	return names
}

// GetAllTargetTypes returns the machine names of all registered built-in targets,
// including universal.
func GetAllTargetTypes() []string {
	names := make([]string, 0, len(builtInTargetDefs))
	for _, def := range builtInTargetDefs {
		names = append(names, string(def.targetType))
	}
	return names
}

// DetectAllTargets returns all detected target environments that exist on the system.
// Also includes custom targets from the config file.
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

// GetAllTargetProjectDirNames returns the project directory name (e.g. ".opencode")
// for every registered built-in target, including universal.
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

// GetAllTargets returns all non-universal built-in targets regardless of whether
// their directories exist. Used for operations like "materialize --target all".
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
