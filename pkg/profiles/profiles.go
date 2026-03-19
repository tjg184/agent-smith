package profiles

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tjg184/agent-smith/pkg/paths"
)

// Profile represents a user profile with agents, skills, and commands
type Profile struct {
	Name        string
	BasePath    string
	HasAgents   bool
	HasSkills   bool
	HasCommands bool
}

// IsValid checks if the profile has at least one component directory
func (p *Profile) IsValid() bool {
	return p.HasAgents || p.HasSkills || p.HasCommands
}

// GetAgentsDir returns the full path to the profile's agents directory
func (p *Profile) GetAgentsDir() string {
	return filepath.Join(p.BasePath, paths.AgentsSubDir)
}

// GetSkillsDir returns the full path to the profile's skills directory
func (p *Profile) GetSkillsDir() string {
	return filepath.Join(p.BasePath, paths.SkillsSubDir)
}

// GetCommandsDir returns the full path to the profile's commands directory
func (p *Profile) GetCommandsDir() string {
	return filepath.Join(p.BasePath, paths.CommandsSubDir)
}

// GetProfileNameFromSymlink determines which profile a symlink belongs to by
// resolving its target path and extracting the profile name.
// Returns "base" if the symlink points to the base installation (~/.agent-smith/),
// or the profile name if it points to a profile directory (~/.agent-smith/profiles/<name>/).
// Returns an error if the path is not a symlink or if the symlink cannot be read.
//
// This method provides shared logic for profile-aware operations that need to
// identify which profile a component belongs to based on its symlink target.
func GetProfileNameFromSymlink(symlinkPath string) (string, error) {
	info, err := os.Lstat(symlinkPath)
	if err != nil {
		return "", fmt.Errorf("failed to stat path: %w", err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return "", fmt.Errorf("path is not a symlink: %s", symlinkPath)
	}

	target, err := os.Readlink(symlinkPath)
	if err != nil {
		return "", fmt.Errorf("failed to read symlink: %w", err)
	}

	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(symlinkPath), target)
	}

	target = filepath.Clean(target)

	dir := filepath.Dir(target)
	for {
		parent := filepath.Dir(dir)
		if filepath.Base(parent) == "profiles" {
			return filepath.Base(dir), nil
		}
		if parent == dir || parent == "." || parent == "/" {
			return "base", nil
		}
		dir = parent
	}
}
