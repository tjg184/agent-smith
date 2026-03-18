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

// GetTargetFromEnv returns the target specified by the AGENT_SMITH_TARGET environment variable
// Returns empty string if not set
func GetTargetFromEnv() string {
	return os.Getenv("AGENT_SMITH_TARGET")
}

// NewTarget creates a new Target based on the specified target type
// If targetType is empty, defaults to OpenCode
// Also checks custom targets from config file
func NewTarget(targetType string) (Target, error) {
	// Default to opencode if not specified
	if targetType == "" {
		targetType = string(TargetOpenCode)
	}

	// Check built-in targets first
	switch TargetType(targetType) {
	case TargetOpenCode:
		return NewOpencodeTarget()
	case TargetClaudeCode:
		return NewClaudeCodeTarget()
	case TargetCopilot:
		return NewCopilotTarget()
	case TargetUniversal:
		return NewUniversalTarget()
	}

	// Check custom targets from config
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

// NewTargetForProject creates a Target configured for a specific project
// The returned Target will use project-relative paths based on the provided projectRoot
// For built-in targets, this returns a Target with the appropriate project directory name
// For custom targets, this validates that ProjectDir is configured
func NewTargetForProject(targetType, projectRoot string) (Target, error) {
	if targetType == "" {
		targetType = string(TargetOpenCode)
	}

	// Handle built-in targets
	switch TargetType(targetType) {
	case TargetOpenCode:
		opencodeTarget := NewOpencodeTargetWithDir(filepath.Join(projectRoot, ".opencode"))
		return opencodeTarget, nil
	case TargetClaudeCode:
		claudeCodeTarget := NewClaudeCodeTargetWithDir(filepath.Join(projectRoot, ".claude"))
		return claudeCodeTarget, nil
	case TargetCopilot:
		copilotTarget := NewCopilotTargetWithDir(filepath.Join(projectRoot, ".github"))
		return copilotTarget, nil
	case TargetUniversal:
		universalTarget := NewUniversalTargetWithDir(filepath.Join(projectRoot, ".agents"))
		return universalTarget, nil
	}

	// Check custom targets from config
	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	for _, customTargetConfig := range config.CustomTargets {
		if customTargetConfig.Name == targetType {
			// For custom targets, we need to set the baseDir to the project-relative path
			customTargetConfig.BaseDir = filepath.Join(projectRoot, customTargetConfig.ProjectDir)
			return NewCustomTarget(customTargetConfig)
		}
	}

	return nil, fmt.Errorf("unknown target type: %s (valid options: opencode, claudecode, copilot, universal, or custom targets from config)", targetType)
}

// DetectTarget attempts to detect which target environment is available
// Priority: AGENT_SMITH_TARGET env var > auto-detection > default to OpenCode
func DetectTarget() (Target, error) {
	// Check environment variable first
	envTarget := GetTargetFromEnv()
	if envTarget != "" {
		return NewTarget(envTarget)
	}

	// Auto-detect by checking which directories exist
	// Priority: OpenCode > Claude Code > Copilot (since OpenCode is the primary target)
	opencodeTarget, err := NewOpencodeTarget()
	if err == nil {
		baseDir, _ := opencodeTarget.GetGlobalBaseDir()
		if _, err := os.Stat(baseDir); err == nil {
			return opencodeTarget, nil
		}
	}

	claudeCodeTarget, err := NewClaudeCodeTarget()
	if err == nil {
		baseDir, _ := claudeCodeTarget.GetGlobalBaseDir()
		if _, err := os.Stat(baseDir); err == nil {
			return claudeCodeTarget, nil
		}
	}

	copilotTarget, err := NewCopilotTarget()
	if err == nil {
		baseDir, _ := copilotTarget.GetGlobalBaseDir()
		if _, err := os.Stat(baseDir); err == nil {
			return copilotTarget, nil
		}
	}

	// Default to OpenCode if neither directory exists
	return NewOpencodeTarget()
}

// GetAvailableTargets returns a list of all available target types
func GetAvailableTargets() []string {
	return []string{
		string(TargetOpenCode),
		string(TargetClaudeCode),
		string(TargetCopilot),
	}
}

// GetAllTargetTypes returns all valid target type strings (including universal)
func GetAllTargetTypes() []string {
	return []string{
		string(TargetOpenCode),
		string(TargetClaudeCode),
		string(TargetCopilot),
		string(TargetUniversal),
	}
}

// DetectAllTargets returns all detected target environments that exist on the system
// This checks which target directories are present and returns Target instances for each
// Also includes custom targets from config file
func DetectAllTargets() ([]Target, error) {
	var targets []Target

	// Check OpenCode
	opencodeTarget, err := NewOpencodeTarget()
	if err == nil {
		baseDir, _ := opencodeTarget.GetGlobalBaseDir()
		if _, err := os.Stat(baseDir); err == nil {
			targets = append(targets, opencodeTarget)
		}
	}

	// Check Claude Code
	claudeCodeTarget, err := NewClaudeCodeTarget()
	if err == nil {
		baseDir, _ := claudeCodeTarget.GetGlobalBaseDir()
		if _, err := os.Stat(baseDir); err == nil {
			targets = append(targets, claudeCodeTarget)
		}
	}

	// Check Copilot
	copilotTarget, err := NewCopilotTarget()
	if err == nil {
		baseDir, _ := copilotTarget.GetGlobalBaseDir()
		if _, err := os.Stat(baseDir); err == nil {
			targets = append(targets, copilotTarget)
		}
	}

	// Check Universal
	universalTarget, err := NewUniversalTarget()
	if err == nil {
		baseDir, _ := universalTarget.GetGlobalBaseDir()
		if _, err := os.Stat(baseDir); err == nil {
			targets = append(targets, universalTarget)
		}
	}

	// Load custom targets from config
	config, err := LoadConfig()
	if err == nil && config != nil {
		for _, customTargetConfig := range config.CustomTargets {
			customTarget, err := NewCustomTarget(customTargetConfig)
			if err != nil {
				// Log warning but continue with other targets
				fmt.Fprintf(os.Stderr, "Warning: failed to load custom target %s: %v\n", customTargetConfig.Name, err)
				continue
			}
			// Check if the base directory exists
			baseDir, _ := customTarget.GetGlobalBaseDir()
			if _, err := os.Stat(baseDir); err == nil {
				targets = append(targets, customTarget)
			}
		}
	}

	// If no targets detected, default to OpenCode
	if len(targets) == 0 {
		opencodeTarget, err := NewOpencodeTarget()
		if err != nil {
			return nil, err
		}
		targets = append(targets, opencodeTarget)
	}

	return targets, nil
}

// GetAllTargets returns all built-in targets (regardless of whether they exist on the system)
// This is useful for operations like "materialize --target all" or "link --target all"
func GetAllTargets() ([]Target, error) {
	opencodeTarget, err := NewOpencodeTarget()
	if err != nil {
		return nil, fmt.Errorf("failed to create opencode target: %w", err)
	}

	claudeCodeTarget, err := NewClaudeCodeTarget()
	if err != nil {
		return nil, fmt.Errorf("failed to create claudecode target: %w", err)
	}

	copilotTarget, err := NewCopilotTarget()
	if err != nil {
		return nil, fmt.Errorf("failed to create copilot target: %w", err)
	}

	return []Target{opencodeTarget, claudeCodeTarget, copilotTarget}, nil
}
