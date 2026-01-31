package profiles

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tgaines/agent-smith/internal/linker"
	"github.com/tgaines/agent-smith/pkg/paths"
)

// ProfileManager handles profile discovery and management
type ProfileManager struct {
	profilesDir string
	linker      *linker.ComponentLinker // Optional - can be nil
}

// NewProfileManager creates a new ProfileManager instance
// The linker parameter is optional - pass nil if unlinking functionality is not needed
func NewProfileManager(componentLinker *linker.ComponentLinker) (*ProfileManager, error) {
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get profiles directory: %w", err)
	}
	return &ProfileManager{
		profilesDir: profilesDir,
		linker:      componentLinker,
	}, nil
}

// GenerateProfileNameFromRepo generates a unique profile name from a repository URL
// The format will be: owner-repo or sanitized-url for non-standard URLs
// If the profile already exists, it appends a short hash suffix
func GenerateProfileNameFromRepo(repoURL string, existingProfiles []string) string {
	// Remove trailing slashes and .git suffix
	repoURL = strings.TrimRight(repoURL, "/")
	repoURL = strings.TrimSuffix(repoURL, ".git")

	var baseName string

	// Try to extract owner/repo from different URL formats
	if strings.Contains(repoURL, "github.com") || strings.Contains(repoURL, "gitlab.com") || strings.Contains(repoURL, "bitbucket.org") {
		// Handle URLs like https://github.com/owner/repo or git@github.com:owner/repo
		parts := strings.Split(repoURL, "/")
		if len(parts) >= 2 {
			owner := parts[len(parts)-2]
			repo := parts[len(parts)-1]

			// Clean up owner name (remove protocol, domain, colon)
			if strings.Contains(owner, ":") {
				owner = strings.Split(owner, ":")[1]
			}

			baseName = fmt.Sprintf("%s-%s", sanitizeForProfileName(owner), sanitizeForProfileName(repo))
		}
	} else if !strings.Contains(repoURL, "/") {
		// Already in owner/repo shorthand format
		baseName = sanitizeForProfileName(strings.ReplaceAll(repoURL, "/", "-"))
	} else if filepath.IsAbs(repoURL) || strings.HasPrefix(repoURL, "./") || strings.HasPrefix(repoURL, "../") {
		// Local path - use the directory name
		baseName = sanitizeForProfileName(filepath.Base(repoURL))
	} else {
		// Fallback for other formats - use last part of URL
		parts := strings.Split(repoURL, "/")
		if len(parts) > 0 {
			baseName = sanitizeForProfileName(parts[len(parts)-1])
		} else {
			baseName = "repo"
		}
	}

	// Ensure baseName is not empty
	if baseName == "" {
		baseName = "repo"
	}

	// Check if profile name already exists
	profileName := baseName
	existsMap := make(map[string]bool)
	for _, p := range existingProfiles {
		existsMap[p] = true
	}

	if !existsMap[profileName] {
		return profileName
	}

	// If profile exists, append a short hash of the full URL
	hash := sha256.Sum256([]byte(repoURL))
	shortHash := hex.EncodeToString(hash[:])[:6]
	profileName = fmt.Sprintf("%s-%s", baseName, shortHash)

	// If still exists (unlikely), append incrementing number
	counter := 2
	originalProfileName := profileName
	for existsMap[profileName] {
		profileName = fmt.Sprintf("%s-%d", originalProfileName, counter)
		counter++
	}

	return profileName
}

// sanitizeForProfileName removes or replaces invalid characters for profile names
// Profile names must match: ^[a-zA-Z0-9-]+$
func sanitizeForProfileName(input string) string {
	// Replace invalid characters with hyphens
	reg := regexp.MustCompile(`[^a-zA-Z0-9-]+`)
	sanitized := reg.ReplaceAllString(input, "-")

	// Remove leading/trailing hyphens
	sanitized = strings.Trim(sanitized, "-")

	// Ensure not empty
	if sanitized == "" {
		return "repo"
	}

	return sanitized
}

// validateProfileName validates a profile name to prevent file system issues
// Returns an error if the name is invalid
func validateProfileName(name string) error {
	// Check for empty name
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	// Check for path traversal attempts (check before hidden directory check)
	if strings.Contains(name, "..") || strings.Contains(name, "./") {
		return fmt.Errorf("profile name cannot contain path traversal patterns (.. or ./)")
	}

	// Check for hidden directories (names starting with .)
	if strings.HasPrefix(name, ".") {
		return fmt.Errorf("profile name cannot start with '.' (hidden directories not allowed)")
	}

	// Check for path separators
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("profile name cannot contain path separators (/ or \\)")
	}

	// Validate against regex pattern: only alphanumeric and hyphens
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	if !validPattern.MatchString(name) {
		return fmt.Errorf("profile name must contain only letters, numbers, and hyphens (got '%s')", name)
	}

	return nil
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

	activeProfilePath := filepath.Join(agentsDir, ".active-profile")

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

// GetComponentNames returns sorted lists of component names in a profile
// Returns three slices: agent names, skill names, and command names
func (pm *ProfileManager) GetComponentNames(profile *Profile) (agents, skills, commands []string) {
	// Get agent names
	if profile.HasAgents {
		agentsPath := filepath.Join(profile.BasePath, paths.AgentsSubDir)
		if entries, err := os.ReadDir(agentsPath); err == nil {
			for _, entry := range entries {
				if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
					agents = append(agents, entry.Name())
				}
			}
		}
	}

	// Get skill names
	if profile.HasSkills {
		skillsPath := filepath.Join(profile.BasePath, paths.SkillsSubDir)
		if entries, err := os.ReadDir(skillsPath); err == nil {
			for _, entry := range entries {
				if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
					skills = append(skills, entry.Name())
				}
			}
		}
	}

	// Get command names
	if profile.HasCommands {
		commandsPath := filepath.Join(profile.BasePath, paths.CommandsSubDir)
		if entries, err := os.ReadDir(commandsPath); err == nil {
			for _, entry := range entries {
				if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
					commands = append(commands, entry.Name())
				}
			}
		}
	}

	return agents, skills, commands
}

// ActivateProfile activates a profile by updating the active profile state
// This does not immediately affect the editor - use 'agent-smith link all' to apply changes
func (pm *ProfileManager) ActivateProfile(profileName string) error {
	// Validate profile name
	if err := validateProfileName(profileName); err != nil {
		return err
	}

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

	// Check if trying to activate already active profile
	if currentActive == profileName {
		return fmt.Errorf("profile '%s' is already active", profileName)
	}

	// Update the active profile state file
	activeProfilePath := filepath.Join(agentsDir, ".active-profile")
	if err := os.WriteFile(activeProfilePath, []byte(profileName), 0644); err != nil {
		return fmt.Errorf("failed to write active profile state: %w", err)
	}

	// Count components for informational output
	agents, skills, commands := pm.CountComponents(profile)
	totalComponents := agents + skills + commands

	fmt.Printf("Successfully activated profile '%s'\n", profileName)
	fmt.Printf("Profile contains %d components (%d agents, %d skills, %d commands)\n", totalComponents, agents, skills, commands)
	fmt.Println("\nTo apply this profile to your editor, run:")
	fmt.Println("  agent-smith link all")
	return nil
}

// AddComponentToProfile copies an existing component from ~/.agent-smith/ to a profile
func (pm *ProfileManager) AddComponentToProfile(profileName, componentType, componentName string) error {
	// Validate profile name
	if err := validateProfileName(profileName); err != nil {
		return err
	}

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
		return fmt.Errorf("component '%s' not found in ~/.agent-smith/%s/", componentName, componentType)
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

// RemoveComponentFromProfile removes a component from a profile
func (pm *ProfileManager) RemoveComponentFromProfile(profileName, componentType, componentName string) error {
	// Validate profile name
	if err := validateProfileName(profileName); err != nil {
		return err
	}

	// Validate component type
	if componentType != "skills" && componentType != "agents" && componentType != "commands" {
		return fmt.Errorf("invalid component type '%s': must be 'skills', 'agents', or 'commands'", componentType)
	}

	// Validate that the profile exists
	profile := pm.loadProfile(profileName)
	if !profile.IsValid() {
		return fmt.Errorf("profile '%s' does not exist or has no components", profileName)
	}

	// Get component path in profile
	componentPath := filepath.Join(profile.BasePath, componentType, componentName)

	// Check if component exists in profile
	if _, err := os.Stat(componentPath); os.IsNotExist(err) {
		return fmt.Errorf("component '%s' not found in profile '%s'", componentName, profileName)
	}

	// Check if this profile is currently active
	activeProfile, err := pm.GetActiveProfile()
	if err == nil && activeProfile == profileName {
		// Component is linked via active profile, need to unlink it
		if pm.linker != nil {
			// Auto-unlink component from all targets (silent if not linked)
			// Pass empty targetFilter to unlink from all targets
			_ = pm.linker.UnlinkComponent(componentType, componentName, "")
		}
	}

	// Remove component directory
	if err := os.RemoveAll(componentPath); err != nil {
		return fmt.Errorf("failed to remove component: %w", err)
	}

	fmt.Printf("Successfully removed %s '%s' from profile '%s'\n", componentType, componentName, profileName)
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

// DeactivateProfile deactivates the currently active profile
// This only updates the state - use 'agent-smith link all' to apply changes
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

	// Clear the active profile state file
	activeProfilePath := filepath.Join(agentsDir, ".active-profile")
	if err := os.Remove(activeProfilePath); err != nil {
		return fmt.Errorf("failed to clear active profile state: %w", err)
	}

	fmt.Printf("Successfully deactivated profile '%s'\n", currentActive)
	fmt.Println("\nTo apply this change to your editor, run:")
	fmt.Println("  agent-smith link all")
	return nil
}

// SwitchProfile switches to a different profile and immediately applies the changes
// This combines ActivateProfile with automatic linking
func (pm *ProfileManager) SwitchProfile(profileName string) error {
	// Validate profile name
	if err := validateProfileName(profileName); err != nil {
		return err
	}

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

	// Check if trying to switch to already active profile
	if currentActive == profileName {
		return fmt.Errorf("profile '%s' is already active", profileName)
	}

	// Update the active profile state file
	activeProfilePath := filepath.Join(agentsDir, ".active-profile")
	if err := os.WriteFile(activeProfilePath, []byte(profileName), 0644); err != nil {
		return fmt.Errorf("failed to write active profile state: %w", err)
	}

	// Count components for informational output
	agents, skills, commands := pm.CountComponents(profile)
	totalComponents := agents + skills + commands

	fmt.Printf("Switched to profile '%s'\n", profileName)
	fmt.Printf("Profile contains %d components (%d agents, %d skills, %d commands)\n", totalComponents, agents, skills, commands)
	fmt.Println("\nNote: You must run 'agent-smith link all' to apply this profile to your editor.")

	return nil
}

// CreateProfile creates a new empty profile with the standard directory structure
func (pm *ProfileManager) CreateProfile(profileName string) error {
	// Validate profile name
	if err := validateProfileName(profileName); err != nil {
		return err
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

// DeleteProfile deletes a profile and all its contents
// Returns an error if the profile is currently active or doesn't exist
func (pm *ProfileManager) DeleteProfile(profileName string) error {
	// Validate profile name
	if err := validateProfileName(profileName); err != nil {
		return err
	}

	// Check if profile exists
	profile := pm.loadProfile(profileName)
	if !profile.IsValid() {
		return fmt.Errorf("profile '%s' does not exist", profileName)
	}

	// Check if profile is currently active
	activeProfile, err := pm.GetActiveProfile()
	if err != nil {
		return fmt.Errorf("failed to check active profile: %w", err)
	}

	if activeProfile == profileName {
		return fmt.Errorf("cannot delete active profile '%s'. Deactivate it first with: agent-smith profiles deactivate", profileName)
	}

	// Defensive unlinking: in case profile somehow has linked components
	// This is a safety net and should not normally be needed since active profiles
	// must be deactivated first (which handles unlinking)
	if pm.linker != nil {
		componentTypes := []string{"skills", "agents", "commands"}
		for _, componentType := range componentTypes {
			componentDir := filepath.Join(profile.BasePath, componentType)
			if _, err := os.Stat(componentDir); !os.IsNotExist(err) {
				entries, err := os.ReadDir(componentDir)
				if err == nil {
					for _, entry := range entries {
						if !strings.HasPrefix(entry.Name(), ".") {
							// Silently attempt to unlink (may not be linked, which is fine)
							// Pass empty targetFilter to unlink from all targets
							_ = pm.linker.UnlinkComponent(componentType, entry.Name(), "")
						}
					}
				}
			}
		}
	}

	// Delete the profile directory
	profileDir := filepath.Join(pm.profilesDir, profileName)
	if err := os.RemoveAll(profileDir); err != nil {
		return fmt.Errorf("failed to delete profile directory: %w", err)
	}

	fmt.Printf("Successfully deleted profile '%s'\n", profileName)
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
