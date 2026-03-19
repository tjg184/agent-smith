package config

import (
	"fmt"

	"github.com/tjg184/agent-smith/pkg/paths"
)

const opencodeProjectDirName = ".opencode"

// OpencodeTarget implements the Target interface for the opencode configuration system
type OpencodeTarget struct {
	baseTarget
}

// NewOpencodeTarget creates a new OpencodeTarget with the default opencode directory
func NewOpencodeTarget() (*OpencodeTarget, error) {
	baseDir, err := paths.GetOpencodeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get opencode directory: %w", err)
	}

	return &OpencodeTarget{baseTarget{baseDir: baseDir, projectDirName: opencodeProjectDirName}}, nil
}

// NewOpencodeTargetWithDir creates a new OpencodeTarget with a custom directory
// This is useful for testing or custom configurations
func NewOpencodeTargetWithDir(dir string) *OpencodeTarget {
	return &OpencodeTarget{baseTarget{baseDir: dir, projectDirName: opencodeProjectDirName}}
}

func (t *OpencodeTarget) GetName() string {
	return "opencode"
}

func (t *OpencodeTarget) GetDisplayName() string {
	return "OpenCode"
}

// IsUniversalTarget returns false for opencode (it's editor-specific)
func (t *OpencodeTarget) IsUniversalTarget() bool {
	return false
}
