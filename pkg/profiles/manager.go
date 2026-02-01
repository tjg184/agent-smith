package profiles

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tgaines/agent-smith/internal/detector"
	"github.com/tgaines/agent-smith/internal/linker"
	"github.com/tgaines/agent-smith/pkg/paths"
)

// ProfileManager handles profile discovery and management
type ProfileManager struct {
	profilesDir string
	linker      *linker.ComponentLinker // Optional - can be nil
}

// ProfileMetadata stores metadata about a profile's source
type ProfileMetadata struct {
	SourceURL string `json:"source_url"`
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

// SaveProfileMetadata saves metadata about a profile's source URL.
// The source URL is normalized before being saved to ensure consistency across different URL formats.
// This enables duplicate detection when installing from the same repository using different URL formats.
// Returns an error if the metadata file cannot be written.
func (pm *ProfileManager) SaveProfileMetadata(profileName, sourceURL string) error {
	profileDir := filepath.Join(pm.profilesDir, profileName)
	metadataPath := filepath.Join(profileDir, ".profile-metadata")

	fmt.Printf("Updating profile metadata for '%s'...\n", profileName)

	// Normalize the URL before saving
	rd := detector.NewRepositoryDetector()
	normalizedURL, err := rd.NormalizeURL(sourceURL)
	if err != nil {
		// If normalization fails, save the original URL
		normalizedURL = sourceURL
	}

	metadata := ProfileMetadata{
		SourceURL: normalizedURL,
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	fmt.Printf("✓ Profile metadata saved successfully\n")
	return nil
}

// LoadProfileMetadata loads metadata for a profile.
// Returns nil if the metadata file does not exist (backward compatibility with legacy profiles).
// Returns an error if the metadata file exists but cannot be read or parsed.
// The returned ProfileMetadata contains the source URL and other profile information.
func (pm *ProfileManager) LoadProfileMetadata(profileName string) (*ProfileMetadata, error) {
	profileDir := filepath.Join(pm.profilesDir, profileName)
	metadataPath := filepath.Join(profileDir, ".profile-metadata")

	// Check if metadata file exists
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return nil, nil // No metadata file, return nil without error
	}

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata ProfileMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

// FindProfileBySourceURL finds a profile that matches the given source URL.
// The input URL is normalized before comparison to match different URL formats (HTTPS, SSH, shorthand).
// Returns the profile name if found, empty string if not found.
// Returns an error only if there's a problem scanning the profiles directory.
// Profiles without metadata files are skipped (backward compatibility).
func (pm *ProfileManager) FindProfileBySourceURL(repoURL string) (string, error) {
	// Normalize the input URL
	rd := detector.NewRepositoryDetector()
	normalizedURL, err := rd.NormalizeURL(repoURL)
	if err != nil {
		// If normalization fails, use the original URL
		normalizedURL = repoURL
	}

	// Scan all profiles
	profiles, err := pm.ScanProfiles()
	if err != nil {
		return "", fmt.Errorf("failed to scan profiles: %w", err)
	}

	// Check each profile's metadata
	for _, profile := range profiles {
		metadata, err := pm.LoadProfileMetadata(profile.Name)
		if err != nil {
			// Skip profiles with metadata errors
			continue
		}

		if metadata != nil && metadata.SourceURL == normalizedURL {
			return profile.Name, nil
		}
	}

	return "", nil // Not found
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

	fmt.Printf("Activating profile '%s'...\n", profileName)

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
	fmt.Printf("Updating active profile state...\n")
	activeProfilePath := filepath.Join(agentsDir, ".active-profile")
	if err := os.WriteFile(activeProfilePath, []byte(profileName), 0644); err != nil {
		return fmt.Errorf("failed to write active profile state: %w", err)
	}

	// Count components for informational output
	agents, skills, commands := pm.CountComponents(profile)
	totalComponents := agents + skills + commands

	fmt.Printf("\n✓ Successfully activated profile '%s'\n", profileName)
	fmt.Printf("Profile contains %d components (%d agents, %d skills, %d commands)\n", totalComponents, agents, skills, commands)
	fmt.Println("\nTo apply this profile to your editor, run:")
	fmt.Println("  agent-smith link all")
	return nil
}

// copyComponentWithMetadata is a helper that copies a component directory and its lock file entry
// from sourceBaseDir to targetBaseDir. This ensures the component remains updateable after copying.
func (pm *ProfileManager) copyComponentWithMetadata(
	sourceBaseDir, targetBaseDir, componentType, componentName string,
) error {
	srcDir := filepath.Join(sourceBaseDir, componentType, componentName)
	dstDir := filepath.Join(targetBaseDir, componentType, componentName)

	// Check if source component exists
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("component '%s' does not exist in source profile (expected at: %s)", componentName, srcDir)
	} else if err != nil {
		return fmt.Errorf("failed to access source component '%s' at %s: %w", componentName, srcDir, err)
	}

	// Check if component already exists in target
	if _, err := os.Stat(dstDir); err == nil {
		return fmt.Errorf("component '%s' already exists in target profile at: %s\n\nTo overwrite, first remove the existing component:\n  agent-smith remove %s %s", componentName, dstDir, componentType, componentName)
	}

	fmt.Printf("Copying component files...\n")

	// Create destination directory
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Copy all files and directories
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

	// Copy lock file entry if it exists
	// Import the internal/metadata package at the top of the file
	fmt.Printf("Copying metadata...\n")

	// We need to import the internal/metadata package to use these functions
	// Use reflection to avoid circular imports by using the metadata package directly
	// Actually, we can just import it - let's check if there's a conflict

	// Try to load the lock entry from source
	sourceLockPath := filepath.Join(sourceBaseDir, fmt.Sprintf(".%s-lock.json", componentType[:len(componentType)-1]))

	// Read source lock file
	lockData, err := os.ReadFile(sourceLockPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No lock file in source - this is OK for manually created components
			fmt.Printf("Note: No lock file entry found in source (manual component)\n")
			return nil
		}
		// Log warning but don't fail
		fmt.Printf("Warning: Failed to read source lock file: %v\n", err)
		return nil
	}

	// Parse the lock file
	var sourceLockFile struct {
		Version  int                               `json:"version"`
		Skills   map[string]map[string]interface{} `json:"skills,omitempty"`
		Agents   map[string]map[string]interface{} `json:"agents,omitempty"`
		Commands map[string]map[string]interface{} `json:"commands,omitempty"`
	}

	if err := json.Unmarshal(lockData, &sourceLockFile); err != nil {
		fmt.Printf("Warning: Failed to parse source lock file: %v\n", err)
		return nil
	}

	// Extract the entry for this component
	var entryMap map[string]interface{}
	switch componentType {
	case "skills":
		entryMap = sourceLockFile.Skills[componentName]
	case "agents":
		entryMap = sourceLockFile.Agents[componentName]
	case "commands":
		entryMap = sourceLockFile.Commands[componentName]
	}

	if entryMap == nil {
		// No entry found - this is OK for manually created components
		fmt.Printf("Note: No lock entry found for component (manual component)\n")
		return nil
	}

	// Read or create target lock file
	targetLockPath := filepath.Join(targetBaseDir, fmt.Sprintf(".%s-lock.json", componentType[:len(componentType)-1]))

	var targetLockFile struct {
		Version  int                               `json:"version"`
		Skills   map[string]map[string]interface{} `json:"skills"`
		Agents   map[string]map[string]interface{} `json:"agents"`
		Commands map[string]map[string]interface{} `json:"commands"`
	}

	targetData, err := os.ReadFile(targetLockPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create new lock file
			targetLockFile.Version = 3
			targetLockFile.Skills = make(map[string]map[string]interface{})
			targetLockFile.Agents = make(map[string]map[string]interface{})
			targetLockFile.Commands = make(map[string]map[string]interface{})
		} else {
			fmt.Printf("Warning: Failed to read target lock file: %v\n", err)
			return nil
		}
	} else {
		if err := json.Unmarshal(targetData, &targetLockFile); err != nil {
			// Create new lock file if parsing fails
			targetLockFile.Version = 3
			targetLockFile.Skills = make(map[string]map[string]interface{})
			targetLockFile.Agents = make(map[string]map[string]interface{})
			targetLockFile.Commands = make(map[string]map[string]interface{})
		}
		// Ensure maps exist
		if targetLockFile.Skills == nil {
			targetLockFile.Skills = make(map[string]map[string]interface{})
		}
		if targetLockFile.Agents == nil {
			targetLockFile.Agents = make(map[string]map[string]interface{})
		}
		if targetLockFile.Commands == nil {
			targetLockFile.Commands = make(map[string]map[string]interface{})
		}
	}

	// Add the entry to target lock file
	switch componentType {
	case "skills":
		targetLockFile.Skills[componentName] = entryMap
	case "agents":
		targetLockFile.Agents[componentName] = entryMap
	case "commands":
		targetLockFile.Commands[componentName] = entryMap
	}

	// Write target lock file
	jsonData, err := json.MarshalIndent(targetLockFile, "", "  ")
	if err != nil {
		fmt.Printf("Warning: Failed to marshal target lock file: %v\n", err)
		return nil
	}

	if err := os.WriteFile(targetLockPath, jsonData, 0644); err != nil {
		fmt.Printf("Warning: Failed to write target lock file: %v\n", err)
		return nil
	}

	fmt.Printf("✓ Metadata copied successfully\n")
	return nil
}

// CopyComponentBetweenProfiles copies a component from one profile to another
func (pm *ProfileManager) CopyComponentBetweenProfiles(
	sourceProfile, targetProfile, componentType, componentName string,
) error {
	// Validate profile names
	if err := validateProfileName(sourceProfile); err != nil {
		return fmt.Errorf("invalid source profile name: %w", err)
	}
	if err := validateProfileName(targetProfile); err != nil {
		return fmt.Errorf("invalid target profile name: %w", err)
	}

	// Validate component type
	if componentType != "skills" && componentType != "agents" && componentType != "commands" {
		return fmt.Errorf("invalid component type '%s'\n\nValid component types:\n  - skills\n  - agents\n  - commands\n\nExample:\n  agent-smith profile copy skills work-profile personal-profile api-design", componentType)
	}

	fmt.Printf("Copying %s '%s' from profile '%s' to profile '%s'...\n", componentType, componentName, sourceProfile, targetProfile)

	// Validate that source profile exists
	srcProfile := pm.loadProfile(sourceProfile)
	if !srcProfile.IsValid() {
		return fmt.Errorf("source profile '%s' does not exist or has no components", sourceProfile)
	}

	// Validate that target profile exists
	dstProfile := pm.loadProfile(targetProfile)
	if !dstProfile.IsValid() {
		return fmt.Errorf("target profile '%s' does not exist or has no components", targetProfile)
	}

	// Check if component exists in source profile
	componentPath := filepath.Join(srcProfile.BasePath, componentType, componentName)
	if _, err := os.Stat(componentPath); os.IsNotExist(err) {
		// List available components in source profile
		availableComponents := []string{}
		componentDir := filepath.Join(srcProfile.BasePath, componentType)
		if entries, err := os.ReadDir(componentDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
					availableComponents = append(availableComponents, entry.Name())
				}
			}
		}

		if len(availableComponents) > 0 {
			return fmt.Errorf("component '%s' not found in source profile '%s'\n\nAvailable %s in '%s':\n  - %s\n\nExample:\n  agent-smith profile copy %s %s %s %s",
				componentName, sourceProfile, componentType, sourceProfile, strings.Join(availableComponents, "\n  - "),
				componentType, sourceProfile, targetProfile, availableComponents[0])
		}
		return fmt.Errorf("component '%s' not found in source profile '%s'", componentName, sourceProfile)
	}

	// Copy component with metadata
	if err := pm.copyComponentWithMetadata(srcProfile.BasePath, dstProfile.BasePath, componentType, componentName); err != nil {
		return err
	}

	// Get component details for success message
	var sourceURL string
	sourceLockPath := filepath.Join(srcProfile.BasePath, fmt.Sprintf(".%s-lock.json", componentType[:len(componentType)-1]))
	if lockData, err := os.ReadFile(sourceLockPath); err == nil {
		var lockFile struct {
			Skills map[string]struct {
				SourceUrl string `json:"sourceUrl"`
			} `json:"skills,omitempty"`
			Agents map[string]struct {
				SourceUrl string `json:"sourceUrl"`
			} `json:"agents,omitempty"`
			Commands map[string]struct {
				SourceUrl string `json:"sourceUrl"`
			} `json:"commands,omitempty"`
		}
		if json.Unmarshal(lockData, &lockFile) == nil {
			switch componentType {
			case "skills":
				if entry, ok := lockFile.Skills[componentName]; ok {
					sourceURL = entry.SourceUrl
				}
			case "agents":
				if entry, ok := lockFile.Agents[componentName]; ok {
					sourceURL = entry.SourceUrl
				}
			case "commands":
				if entry, ok := lockFile.Commands[componentName]; ok {
					sourceURL = entry.SourceUrl
				}
			}
		}
	}

	componentSingular := strings.TrimSuffix(componentType, "s")
	fmt.Printf("\n✓ Successfully copied %s '%s' from '%s' to '%s'\n", componentSingular, componentName, sourceProfile, targetProfile)
	fmt.Printf("\nComponent details:\n")
	fmt.Printf("  Type: %s\n", componentType)
	fmt.Printf("  Name: %s\n", componentName)
	if sourceURL != "" {
		fmt.Printf("  Source: %s\n", sourceURL)
	}
	fmt.Printf("  Location: %s\n", filepath.Join(dstProfile.BasePath, componentType, componentName))
	fmt.Printf("\nBoth profiles can now update this component independently.\n")

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

	fmt.Printf("Adding %s '%s' to profile '%s'...\n", componentType, componentName, profileName)

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

	// Copy component with metadata using the helper function
	// This will copy both the files and the lock file entry
	if err := pm.copyComponentWithMetadata(agentsDir, profile.BasePath, componentType, componentName); err != nil {
		return err
	}

	fmt.Printf("✓ Successfully added %s '%s' to profile '%s'\n", componentType, componentName, profileName)
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

	fmt.Printf("Removing %s '%s' from profile '%s'...\n", componentType, componentName, profileName)

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
		fmt.Printf("Unlinking component from active profile...\n")
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

	fmt.Printf("✓ Successfully removed %s '%s' from profile '%s'\n", componentType, componentName, profileName)
	return nil
}

// copyDirectory recursively copies a directory
func copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("failed to determine relative path for %s: %w", path, err)
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dstPath, err)
			}
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("permission denied reading file %s: %w", path, err)
			}
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("permission denied writing file %s: %w", dstPath, err)
			}
			return fmt.Errorf("failed to write file %s: %w", dstPath, err)
		}

		return nil
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

	fmt.Printf("Deactivating profile '%s'...\n", currentActive)

	// Clear the active profile state file
	activeProfilePath := filepath.Join(agentsDir, ".active-profile")
	if err := os.Remove(activeProfilePath); err != nil {
		return fmt.Errorf("failed to clear active profile state: %w", err)
	}

	fmt.Printf("\n✓ Successfully deactivated profile '%s'\n", currentActive)
	fmt.Println("\nTo apply this change to your editor, run:")
	fmt.Println("  agent-smith link all")
	return nil
}

// CreateProfile creates a new empty profile with the standard directory structure
func (pm *ProfileManager) CreateProfile(profileName string) error {
	// Validate profile name
	if err := validateProfileName(profileName); err != nil {
		return err
	}

	fmt.Printf("Creating profile '%s'...\n", profileName)

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

	fmt.Printf("Creating component directories...\n")
	for _, dir := range componentDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create component directory %s: %w", dir, err)
		}
	}

	fmt.Printf("\n✓ Created profile: %s\n", profileName)
	fmt.Printf("  Location: %s\n", profileDir)
	fmt.Println("\nComponent directories created:")
	fmt.Printf("  - %s\n", paths.AgentsSubDir)
	fmt.Printf("  - %s\n", paths.SkillsSubDir)
	fmt.Printf("  - %s\n", paths.CommandsSubDir)
	fmt.Println("\nYou can now add components to this profile and activate it with:")
	fmt.Printf("  agent-smith profiles activate %s\n", profileName)

	return nil
}

// CreateProfileWithMetadata creates a new profile with source URL metadata
func (pm *ProfileManager) CreateProfileWithMetadata(profileName, sourceURL string) error {
	// Create the profile first
	if err := pm.CreateProfile(profileName); err != nil {
		return err
	}

	// Save metadata
	if sourceURL != "" {
		if err := pm.SaveProfileMetadata(profileName, sourceURL); err != nil {
			// Log warning but don't fail the operation
			fmt.Printf("Warning: Failed to save profile metadata: %v\n", err)
		}
	}

	return nil
}

// DeleteProfile deletes a profile and all its contents
// Returns an error if the profile is currently active or doesn't exist
func (pm *ProfileManager) DeleteProfile(profileName string) error {
	// Validate profile name
	if err := validateProfileName(profileName); err != nil {
		return err
	}

	fmt.Printf("Deleting profile '%s'...\n", profileName)

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

	fmt.Printf("Cleaning up components...\n")

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

	fmt.Printf("✓ Successfully deleted profile '%s'\n", profileName)
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

// ComponentItem represents a component available for cherry-picking
type ComponentItem struct {
	Type          string // "skills", "agents", or "commands"
	Name          string
	SourceProfile string
}

// GetAllAvailableComponents returns all components from specified profiles
// If sourceProfiles is empty, returns components from all profiles
func (pm *ProfileManager) GetAllAvailableComponents(sourceProfiles []string) ([]ComponentItem, error) {
	var items []ComponentItem

	// Get list of profiles to scan
	var profilesToScan []*Profile
	if len(sourceProfiles) == 0 {
		// Scan all profiles
		allProfiles, err := pm.ScanProfiles()
		if err != nil {
			return nil, fmt.Errorf("failed to scan profiles: %w", err)
		}
		profilesToScan = allProfiles
	} else {
		// Scan only specified profiles
		for _, profileName := range sourceProfiles {
			profile := pm.loadProfile(profileName)
			if !profile.IsValid() {
				return nil, fmt.Errorf("source profile '%s' does not exist or has no components", profileName)
			}
			profilesToScan = append(profilesToScan, profile)
		}
	}

	// Collect components from each profile
	for _, profile := range profilesToScan {
		agents, skills, commands := pm.GetComponentNames(profile)

		for _, name := range agents {
			items = append(items, ComponentItem{
				Type:          "agents",
				Name:          name,
				SourceProfile: profile.Name,
			})
		}

		for _, name := range skills {
			items = append(items, ComponentItem{
				Type:          "skills",
				Name:          name,
				SourceProfile: profile.Name,
			})
		}

		for _, name := range commands {
			items = append(items, ComponentItem{
				Type:          "commands",
				Name:          name,
				SourceProfile: profile.Name,
			})
		}
	}

	return items, nil
}

// CherryPickComponents copies selected components from source profiles to target profile
func (pm *ProfileManager) CherryPickComponents(targetProfile string, components []ComponentItem) error {
	// Validate target profile
	if err := validateProfileName(targetProfile); err != nil {
		return fmt.Errorf("invalid target profile name: %w", err)
	}

	profile := pm.loadProfile(targetProfile)
	if !profile.IsValid() {
		return fmt.Errorf("target profile '%s' does not exist or has no components", targetProfile)
	}

	fmt.Printf("Cherry-picking %d component(s) to profile '%s'...\n\n", len(components), targetProfile)

	// Copy each component
	successCount := 0
	skipCount := 0
	errorCount := 0

	for _, component := range components {
		fmt.Printf("Copying %s '%s' from '%s'...\n", component.Type, component.Name, component.SourceProfile)

		err := pm.CopyComponentBetweenProfiles(component.SourceProfile, targetProfile, component.Type, component.Name)
		if err != nil {
			// Check if error is due to component already existing
			if strings.Contains(err.Error(), "already exists") {
				fmt.Printf("  ⊘ Skipped (already exists)\n\n")
				skipCount++
			} else {
				fmt.Printf("  ✗ Error: %v\n\n", err)
				errorCount++
			}
		} else {
			fmt.Printf("  ✓ Success\n\n")
			successCount++
		}
	}

	// Print summary
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("\nCherry-pick Summary:\n")
	fmt.Printf("  ✓ Successfully copied: %d\n", successCount)
	if skipCount > 0 {
		fmt.Printf("  ⊘ Skipped (existing):  %d\n", skipCount)
	}
	if errorCount > 0 {
		fmt.Printf("  ✗ Failed:              %d\n", errorCount)
	}
	fmt.Printf("\nTotal components in '%s': ", targetProfile)

	// Count final components
	finalProfile := pm.loadProfile(targetProfile)
	agents, skills, commands := pm.CountComponents(finalProfile)
	total := agents + skills + commands
	fmt.Printf("%d (%d agents, %d skills, %d commands)\n", total, agents, skills, commands)

	if errorCount > 0 {
		return fmt.Errorf("some components failed to copy")
	}

	return nil
}
