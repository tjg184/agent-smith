package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tjg184/agent-smith/pkg/paths"
)

type TargetType string

const (
	TargetOpenCode   TargetType = "opencode"
	TargetClaudeCode TargetType = "claudecode"
	TargetCopilot    TargetType = "copilot"
	TargetUniversal  TargetType = "universal"
)

type builtInTargetDef struct {
	targetType         TargetType
	constructor        func() (Target, error)
	projectConstructor func(dir string) Target
	isUniversal        bool
}

func newTargetFromSpec(spec targetSpec) (Target, error) {
	dir, err := paths.ExpandHome(spec.globalDir)
	if err != nil {
		return nil, fmt.Errorf("failed to expand %s directory: %w", spec.name, err)
	}
	return &genericTarget{
		baseTarget:  baseTarget{baseDir: dir, projectDirName: spec.projectDir},
		name:        spec.name,
		displayName: spec.displayName,
		isUniversal: spec.isUniversal,
	}, nil
}

func newTargetFromSpecWithDir(spec targetSpec, dir string) Target {
	return &genericTarget{
		baseTarget:  baseTarget{baseDir: dir, projectDirName: spec.projectDir},
		name:        spec.name,
		displayName: spec.displayName,
		isUniversal: spec.isUniversal,
	}
}

// builtInTargetDefs is derived from builtInTargetSpecs.
// To add a new built-in target, add a line to target_specs.go only.
var builtInTargetDefs = func() []builtInTargetDef {
	defs := make([]builtInTargetDef, len(builtInTargetSpecs))
	for i, spec := range builtInTargetSpecs {
		s := spec
		defs[i] = builtInTargetDef{
			targetType:         TargetType(s.name),
			constructor:        func() (Target, error) { return newTargetFromSpec(s) },
			projectConstructor: func(dir string) Target { return newTargetFromSpecWithDir(s, dir) },
			isUniversal:        s.isUniversal,
		}
	}
	return defs
}()

// NewUniversalTarget is provided for callers that need the universal target explicitly.
func NewUniversalTarget() (Target, error) {
	return NewTarget(string(TargetUniversal))
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

	return NewTarget(string(TargetOpenCode))
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
	seenProjectDirs := make(map[string]bool)

	for _, def := range builtInTargetDefs {
		target, err := def.constructor()
		if err != nil {
			continue
		}
		projectDir := target.GetProjectDirName()
		if seenProjectDirs[projectDir] {
			continue
		}
		baseDir, _ := target.GetGlobalBaseDir()
		if _, err := os.Stat(baseDir); err == nil {
			targets = append(targets, target)
			seenProjectDirs[projectDir] = true
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
			projectDir := customTarget.GetProjectDirName()
			if seenProjectDirs[projectDir] {
				continue
			}
			baseDir, _ := customTarget.GetGlobalBaseDir()
			if _, err := os.Stat(baseDir); err == nil {
				targets = append(targets, customTarget)
				seenProjectDirs[projectDir] = true
			}
		}
	}

	if len(targets) == 0 {
		opencodeTarget, err := NewTarget(string(TargetOpenCode))
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
