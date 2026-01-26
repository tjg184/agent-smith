package profiles

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tgaines/agent-smith/pkg/paths"
)

// ProfileManager handles profile discovery and management
type ProfileManager struct {
	profilesDir string
}

// NewProfileManager creates a new ProfileManager instance
func NewProfileManager() (*ProfileManager, error) {
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get profiles directory: %w", err)
	}
	return &ProfileManager{profilesDir: profilesDir}, nil
}

// ScanProfiles discovers all valid profiles in the profiles directory
// Returns an empty slice if the profiles directory doesn't exist
// Invalid profiles (those without any component directories) are silently ignored
func (pm *ProfileManager) ScanProfiles() ([]*Profile, error) {
	// Check if profiles directory exists
	if _, err := os.Stat(pm.profilesDir); os.IsNotExist(err) {
		return []*Profile{}, nil // No profiles yet, return empty list
	}

	entries, err := os.ReadDir(pm.profilesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read profiles directory: %w", err)
	}

	var profiles []*Profile
	for _, entry := range entries {
		// Skip files and hidden directories
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		profile := pm.loadProfile(entry.Name())
		if profile.IsValid() {
			profiles = append(profiles, profile)
		}
		// Silently skip invalid profiles (graceful handling)
	}

	return profiles, nil
}

// loadProfile loads a profile from a directory and checks which component directories exist
func (pm *ProfileManager) loadProfile(name string) *Profile {
	basePath := filepath.Join(pm.profilesDir, name)

	profile := &Profile{
		Name:     name,
		BasePath: basePath,
	}

	// Check which component directories exist
	if _, err := os.Stat(filepath.Join(basePath, paths.AgentsSubDir)); err == nil {
		profile.HasAgents = true
	}
	if _, err := os.Stat(filepath.Join(basePath, paths.SkillsSubDir)); err == nil {
		profile.HasSkills = true
	}
	if _, err := os.Stat(filepath.Join(basePath, paths.CommandsSubDir)); err == nil {
		profile.HasCommands = true
	}

	return profile
}

// GetActiveProfile reads the active profile from the state file
// Returns empty string if no profile is active or state file doesn't exist
func (pm *ProfileManager) GetActiveProfile() (string, error) {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return "", fmt.Errorf("failed to get agents directory: %w", err)
	}

	activeProfilePath := filepath.Join(agentsDir, ".active_profile")

	// Check if state file exists
	if _, err := os.Stat(activeProfilePath); os.IsNotExist(err) {
		return "", nil // No active profile yet
	}

	// Read the file
	data, err := os.ReadFile(activeProfilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read active profile file: %w", err)
	}

	// Trim whitespace and return profile name
	profileName := strings.TrimSpace(string(data))
	return profileName, nil
}

// CountComponents counts the number of component directories in a profile
// Returns counts for agents, skills, and commands
func (pm *ProfileManager) CountComponents(profile *Profile) (agents, skills, commands int) {
	// Count agents
	if profile.HasAgents {
		agentsPath := filepath.Join(profile.BasePath, paths.AgentsSubDir)
		if entries, err := os.ReadDir(agentsPath); err == nil {
			for _, entry := range entries {
				if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
					agents++
				}
			}
		}
	}

	// Count skills
	if profile.HasSkills {
		skillsPath := filepath.Join(profile.BasePath, paths.SkillsSubDir)
		if entries, err := os.ReadDir(skillsPath); err == nil {
			for _, entry := range entries {
				if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
					skills++
				}
			}
		}
	}

	// Count commands
	if profile.HasCommands {
		commandsPath := filepath.Join(profile.BasePath, paths.CommandsSubDir)
		if entries, err := os.ReadDir(commandsPath); err == nil {
			for _, entry := range entries {
				if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
					commands++
				}
			}
		}
	}

	return agents, skills, commands
}
