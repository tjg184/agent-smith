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
func NewTarget(targetType string) (Target, error) {
	// Default to opencode if not specified
	if targetType == "" {
		targetType = string(TargetOpenCode)
	}

	switch TargetType(targetType) {
	case TargetOpenCode:
		return NewOpencodeTarget()
	case TargetClaudeCode:
		return NewClaudeCodeTarget()
	default:
		return nil, fmt.Errorf("unknown target type: %s (valid options: opencode, claudecode)", targetType)
	}
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
