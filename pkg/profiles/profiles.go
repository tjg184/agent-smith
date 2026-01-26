package profiles

import (
	"path/filepath"

	"github.com/tgaines/agent-smith/pkg/paths"
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
