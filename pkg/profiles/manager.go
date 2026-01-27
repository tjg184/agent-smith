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

// ActivateProfile activates a profile by creating symlinks from profile components to ~/.agents/
// If another profile is active, it will be deactivated first
func (pm *ProfileManager) ActivateProfile(profileName string) error {
	// Validate that the profile exists
	profile := pm.loadProfile(profileName)
	if !profile.IsValid() {
		return fmt.Errorf("profile '%s' does not exist or has no components", profileName)
	}

	// Get the agents directory
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return fmt.Errorf("failed to get agents directory: %w", err)
	}

	// Check if a profile is currently active
	currentActive, err := pm.GetActiveProfile()
	if err != nil {
		return fmt.Errorf("failed to check current active profile: %w", err)
	}

	// Deactivate current profile if one is active
	if currentActive != "" {
		if currentActive == profileName {
			return fmt.Errorf("profile '%s' is already active", profileName)
		}
		fmt.Printf("Deactivating current profile: %s\n", currentActive)
		if err := pm.unlinkAllComponents(agentsDir); err != nil {
			return fmt.Errorf("failed to deactivate current profile: %w", err)
		}
	}

	// Activate the new profile by creating symlinks
	fmt.Printf("Activating profile: %s\n", profileName)

	componentTypes := []struct {
		name      string
		hasType   bool
		getDir    func() string
		targetDir string
	}{
		{"agents", profile.HasAgents, profile.GetAgentsDir, filepath.Join(agentsDir, paths.AgentsSubDir)},
		{"skills", profile.HasSkills, profile.GetSkillsDir, filepath.Join(agentsDir, paths.SkillsSubDir)},
		{"commands", profile.HasCommands, profile.GetCommandsDir, filepath.Join(agentsDir, paths.CommandsSubDir)},
	}

	linkedCount := 0
	for _, ct := range componentTypes {
		if !ct.hasType {
			continue
		}

		sourceDir := ct.getDir()
		entries, err := os.ReadDir(sourceDir)
		if err != nil {
			return fmt.Errorf("failed to read %s directory: %w", ct.name, err)
		}

		// Ensure target directory exists
		if err := os.MkdirAll(ct.targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create target directory for %s: %w", ct.name, err)
		}

		for _, entry := range entries {
			if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			componentName := entry.Name()
			srcPath := filepath.Join(sourceDir, componentName)
			dstPath := filepath.Join(ct.targetDir, componentName)

			// Create relative symlink
			relPath, err := filepath.Rel(filepath.Dir(dstPath), srcPath)
			if err != nil {
				return fmt.Errorf("failed to create relative path for %s/%s: %w", ct.name, componentName, err)
			}

			// Remove existing destination if it exists
			if _, err := os.Lstat(dstPath); err == nil {
				if err := os.Remove(dstPath); err != nil {
					return fmt.Errorf("failed to remove existing %s/%s: %w", ct.name, componentName, err)
				}
			}

			// Create the symlink
			if err := os.Symlink(relPath, dstPath); err != nil {
				return fmt.Errorf("failed to create symlink for %s/%s: %w", ct.name, componentName, err)
			}

			linkedCount++
			fmt.Printf("  Linked %s/%s\n", ct.name, componentName)
		}
	}

	// Update the active profile state file
	activeProfilePath := filepath.Join(agentsDir, ".active_profile")
	if err := os.WriteFile(activeProfilePath, []byte(profileName), 0644); err != nil {
		return fmt.Errorf("failed to write active profile state: %w", err)
	}

	fmt.Printf("\nSuccessfully activated profile '%s' (%d components linked)\n", profileName, linkedCount)
	return nil
}

// AddComponentToProfile copies an existing component from ~/.agents/ to a profile
func (pm *ProfileManager) AddComponentToProfile(profileName, componentType, componentName string) error {
	// Validate component type
	if componentType != "skills" && componentType != "agents" && componentType != "commands" {
		return fmt.Errorf("invalid component type '%s': must be 'skills', 'agents', or 'commands'", componentType)
	}

	// Validate that the profile exists
	profile := pm.loadProfile(profileName)
	if !profile.IsValid() {
		return fmt.Errorf("profile '%s' does not exist or has no components", profileName)
	}

	// Get source directory based on component type
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return fmt.Errorf("failed to get agents directory: %w", err)
	}

	srcDir := filepath.Join(agentsDir, componentType, componentName)

	// Check if source component exists
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("component '%s' not found in ~/.agents/%s/", componentName, componentType)
	}

	// Check if component is a symlink (from active profile)
	info, err := os.Lstat(srcDir)
	if err != nil {
		return fmt.Errorf("failed to stat component: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("cannot add component '%s': it is a symlink from an active profile. Deactivate the profile first.", componentName)
	}

	// Get destination directory in profile
	dstDir := filepath.Join(profile.BasePath, componentType, componentName)

	// Check if component already exists in profile
	if _, err := os.Stat(dstDir); err == nil {
		return fmt.Errorf("component '%s' already exists in profile '%s'", componentName, profileName)
	}

	// Copy component to profile
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Copy files
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectories
			if err := copyDirectory(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy directory %s: %w", entry.Name(), err)
			}
		} else {
			// Copy file
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", entry.Name(), err)
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", entry.Name(), err)
			}
		}
	}

	fmt.Printf("Successfully added %s '%s' to profile '%s'\n", componentType, componentName, profileName)
	return nil
}

// copyDirectory recursively copies a directory
func copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(dstPath, data, 0644)
	})
}

// DeactivateProfile deactivates the currently active profile and returns to base state
// This removes all profile symlinks and clears the active profile state
func (pm *ProfileManager) DeactivateProfile() error {
	// Get the agents directory
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return fmt.Errorf("failed to get agents directory: %w", err)
	}

	// Check if a profile is currently active
	currentActive, err := pm.GetActiveProfile()
	if err != nil {
		return fmt.Errorf("failed to check current active profile: %w", err)
	}

	if currentActive == "" {
		return fmt.Errorf("no profile is currently active")
	}

	fmt.Printf("Deactivating profile: %s\n", currentActive)

	// Remove all profile symlinks
	if err := pm.unlinkAllComponents(agentsDir); err != nil {
		return fmt.Errorf("failed to remove profile components: %w", err)
	}

	// Clear the active profile state file
	activeProfilePath := filepath.Join(agentsDir, ".active_profile")
	if err := os.Remove(activeProfilePath); err != nil {
		return fmt.Errorf("failed to clear active profile state: %w", err)
	}

	fmt.Printf("\nSuccessfully deactivated profile '%s'\n", currentActive)
	fmt.Println("Returned to base state (no profile active)")
	return nil
}

// CreateProfile creates a new empty profile with the standard directory structure
func (pm *ProfileManager) CreateProfile(profileName string) error {
	// Validate profile name
	if profileName == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	// Check if profile already exists
	profile := pm.loadProfile(profileName)
	if profile.IsValid() {
		return fmt.Errorf("profile '%s' already exists", profileName)
	}

	// Create profile directory
	profileDir := filepath.Join(pm.profilesDir, profileName)
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}

	// Create component directories
	componentDirs := []string{
		filepath.Join(profileDir, paths.AgentsSubDir),
		filepath.Join(profileDir, paths.SkillsSubDir),
		filepath.Join(profileDir, paths.CommandsSubDir),
	}

	for _, dir := range componentDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create component directory %s: %w", dir, err)
		}
	}

	fmt.Printf("Created profile: %s\n", profileName)
	fmt.Printf("  Location: %s\n", profileDir)
	fmt.Println("\nComponent directories created:")
	fmt.Printf("  - %s\n", paths.AgentsSubDir)
	fmt.Printf("  - %s\n", paths.SkillsSubDir)
	fmt.Printf("  - %s\n", paths.CommandsSubDir)
	fmt.Println("\nYou can now add components to this profile and activate it with:")
	fmt.Printf("  agent-smith profiles activate %s\n", profileName)

	return nil
}

// unlinkAllComponents removes all symlinks from the agents directory component folders
func (pm *ProfileManager) unlinkAllComponents(agentsDir string) error {
	componentDirs := []string{
		filepath.Join(agentsDir, paths.AgentsSubDir),
		filepath.Join(agentsDir, paths.SkillsSubDir),
		filepath.Join(agentsDir, paths.CommandsSubDir),
	}

	for _, dir := range componentDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("failed to read directory %s: %w", dir, err)
		}

		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			entryPath := filepath.Join(dir, entry.Name())

			// Check if it's a symlink
			info, err := os.Lstat(entryPath)
			if err != nil {
				continue
			}

			if info.Mode()&os.ModeSymlink != 0 {
				if err := os.Remove(entryPath); err != nil {
					return fmt.Errorf("failed to remove symlink %s: %w", entryPath, err)
				}
			}
		}
	}

	return nil
}
