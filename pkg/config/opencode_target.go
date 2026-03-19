package config

import (
	"fmt"

	"github.com/tjg184/agent-smith/pkg/paths"
)

const opencodeProjectDirName = ".opencode"

type OpencodeTarget struct {
	baseTarget
}

func NewOpencodeTarget() (*OpencodeTarget, error) {
	baseDir, err := paths.GetOpencodeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get opencode directory: %w", err)
	}

	return &OpencodeTarget{baseTarget{baseDir: baseDir, projectDirName: opencodeProjectDirName}}, nil
}

func NewOpencodeTargetWithDir(dir string) *OpencodeTarget {
	return &OpencodeTarget{baseTarget{baseDir: dir, projectDirName: opencodeProjectDirName}}
}

func (t *OpencodeTarget) GetName() string {
	return "opencode"
}

func (t *OpencodeTarget) GetDisplayName() string {
	return "OpenCode"
}

func (t *OpencodeTarget) IsUniversalTarget() bool {
	return false
}
