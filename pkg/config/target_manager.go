package config

import (
	"fmt"
	"os"
)

// TargetType represents the type of target environment
type TargetType string

const (
	// TargetOpenCode represents the OpenCode environment
	TargetOpenCode TargetType = "opencode"
	// TargetClaudeCode represents the Claude Code environment
	TargetClaudeCode TargetType = "claudecode"
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

	return nil, fmt.Errorf("unknown target type: %s (valid options: opencode, claudecode, or custom targets from config)", targetType)
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
	// Priority: OpenCode > Claude Code (since OpenCode is the primary target)
	opencodeTarget, err := NewOpencodeTarget()
	if err == nil {
		baseDir, _ := opencodeTarget.GetBaseDir()
		if _, err := os.Stat(baseDir); err == nil {
			return opencodeTarget, nil
		}
	}

	claudeCodeTarget, err := NewClaudeCodeTarget()
	if err == nil {
		baseDir, _ := claudeCodeTarget.GetBaseDir()
		if _, err := os.Stat(baseDir); err == nil {
			return claudeCodeTarget, nil
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
		baseDir, _ := opencodeTarget.GetBaseDir()
		if _, err := os.Stat(baseDir); err == nil {
			targets = append(targets, opencodeTarget)
		}
	}

	// Check Claude Code
	claudeCodeTarget, err := NewClaudeCodeTarget()
	if err == nil {
		baseDir, _ := claudeCodeTarget.GetBaseDir()
		if _, err := os.Stat(baseDir); err == nil {
			targets = append(targets, claudeCodeTarget)
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
			baseDir, _ := customTarget.GetBaseDir()
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
