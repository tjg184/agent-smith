package profiles

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjg184/agent-smith/internal/linker"
	"github.com/tjg184/agent-smith/pkg/paths"
	"github.com/tjg184/agent-smith/pkg/profiles/activation"
	profileCopy "github.com/tjg184/agent-smith/pkg/profiles/copy"
	"github.com/tjg184/agent-smith/pkg/profiles/profilemeta"
	"github.com/tjg184/agent-smith/pkg/profiles/scanner"
	"github.com/tjg184/agent-smith/pkg/services"
)

// ProfileManager handles profile discovery and management.
type ProfileManager struct {
	profilesDir string
	linker      *linker.ComponentLinker       // Optional - can be nil
	lockService services.ComponentLockService // Required for lock file operations
}

// ProfileMetadata stores metadata about a profile's source.
// Delegates to profilemeta.ProfileMetadata.
type ProfileMetadata = profilemeta.ProfileMetadata

// ProfileActivationResult contains information about a profile activation operation.
// Delegates to activation.ProfileActivationResult.
type ProfileActivationResult = activation.ProfileActivationResult

// ComponentItem represents a component available for cherry-picking.
type ComponentItem struct {
	Type          string // "skills", "agents", or "commands"
	Name          string
	SourceProfile string
}

// NewProfileManager creates a new ProfileManager instance.
// The linker parameter is optional (pass nil if not needed for unlinking functionality).
// The lockService parameter is required for full ProfileManager functionality.
func NewProfileManager(componentLinker *linker.ComponentLinker, lockService services.ComponentLockService) (*ProfileManager, error) {
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get profiles directory: %w", err)
	}
	return &ProfileManager{
		profilesDir: profilesDir,
		linker:      componentLinker,
		lockService: lockService,
	}, nil
}

// SaveProfileMetadata saves metadata about a profile's source URL.
func (pm *ProfileManager) SaveProfileMetadata(profileName, sourceURL string) error {
	profileDir := filepath.Join(pm.profilesDir, profileName)
	fmt.Printf("Updating profile metadata for '%s'...\n", profileName)
	if err := profilemeta.Save(profileDir, sourceURL); err != nil {
		return err
	}
	fmt.Printf("✓ Profile metadata saved successfully\n")
	return nil
}

// SaveUserProfileMetadata saves metadata for a user-created profile.
func (pm *ProfileManager) SaveUserProfileMetadata(profileName string) error {
	profileDir := filepath.Join(pm.profilesDir, profileName)
	return profilemeta.SaveUser(profileDir)
}

// LoadProfileMetadata loads metadata for a profile.
func (pm *ProfileManager) LoadProfileMetadata(profileName string) (*ProfileMetadata, error) {
	profileDir := filepath.Join(pm.profilesDir, profileName)
	return profilemeta.Load(profileDir)
}

// GetProfileType returns the type of a profile ("repo", "user", or "unknown").
func (pm *ProfileManager) GetProfileType(profileName string) (string, error) {
	profileDir := filepath.Join(pm.profilesDir, profileName)
	return profilemeta.GetProfileType(profileDir)
}

// FindProfileBySourceURL finds a profile that matches the given source URL.
func (pm *ProfileManager) FindProfileBySourceURL(repoURL string) (string, error) {
	return profilemeta.FindBySourceURL(pm.profilesDir, repoURL)
}

// GenerateProfileNameFromRepo generates a unique profile name from a repository URL.
func GenerateProfileNameFromRepo(repoURL string, existingProfiles []string) string {
	return profilemeta.GenerateNameFromRepo(repoURL, existingProfiles)
}

// ScanProfiles discovers all valid profiles in the profiles directory.
func (pm *ProfileManager) ScanProfiles() ([]*Profile, error) {
	scanProfiles, err := scanner.ScanProfiles(pm.profilesDir)
	if err != nil {
		return nil, err
	}
	result := make([]*Profile, len(scanProfiles))
	for i, sp := range scanProfiles {
		result[i] = &Profile{
			Name:        sp.Name,
			BasePath:    sp.BasePath,
			HasAgents:   sp.HasAgents,
			HasSkills:   sp.HasSkills,
			HasCommands: sp.HasCommands,
		}
	}
	return result, nil
}

// loadProfile loads a profile from the profiles directory.
func (pm *ProfileManager) loadProfile(name string) *Profile {
	sp := scanner.LoadProfile(pm.profilesDir, name)
	return &Profile{
		Name:        sp.Name,
		BasePath:    sp.BasePath,
		HasAgents:   sp.HasAgents,
		HasSkills:   sp.HasSkills,
		HasCommands: sp.HasCommands,
	}
}

// GetActiveProfile reads the active profile from the state file.
func (pm *ProfileManager) GetActiveProfile() (string, error) {
	return activation.GetActiveProfile()
}

// CountComponents counts the number of component directories in a profile.
func (pm *ProfileManager) CountComponents(profile *Profile) (agents, skills, commands int) {
	sp := &scanner.Profile{
		Name:        profile.Name,
		BasePath:    profile.BasePath,
		HasAgents:   profile.HasAgents,
		HasSkills:   profile.HasSkills,
		HasCommands: profile.HasCommands,
	}
	return scanner.CountComponents(sp)
}

// GetComponentNames returns sorted lists of component names in a profile.
func (pm *ProfileManager) GetComponentNames(profile *Profile) (agents, skills, commands []string) {
	sp := &scanner.Profile{
		Name:        profile.Name,
		BasePath:    profile.BasePath,
		HasAgents:   profile.HasAgents,
		HasSkills:   profile.HasSkills,
		HasCommands: profile.HasCommands,
	}
	return scanner.GetComponentNames(sp)
}

// GetComponentSource returns the source URL for a component from its lock file.
func (pm *ProfileManager) GetComponentSource(profile *Profile, componentType, componentName string) string {
	sp := &scanner.Profile{
		Name:        profile.Name,
		BasePath:    profile.BasePath,
		HasAgents:   profile.HasAgents,
		HasSkills:   profile.HasSkills,
		HasCommands: profile.HasCommands,
	}
	return scanner.GetComponentSource(sp, pm.lockService, componentType, componentName)
}

// ActivateProfile activates a profile by updating the active profile state.
func (pm *ProfileManager) ActivateProfile(profileName string) error {
	if err := validateProfileName(profileName); err != nil {
		return err
	}
	profile := pm.loadProfile(profileName)
	return activation.ActivateProfile(profileName, profile.IsValid())
}

// ActivateProfileWithResult sets the given profile as the active profile and returns detailed result.
func (pm *ProfileManager) ActivateProfileWithResult(profileName string) (*ProfileActivationResult, error) {
	if err := validateProfileName(profileName); err != nil {
		return nil, err
	}
	profile := pm.loadProfile(profileName)
	return activation.ActivateProfileWithResult(profileName, profile.IsValid())
}

// DeactivateProfile deactivates the currently active profile.
func (pm *ProfileManager) DeactivateProfile() error {
	return activation.DeactivateProfile()
}

// CopyComponentBetweenProfiles copies a component from one profile to another.
func (pm *ProfileManager) CopyComponentBetweenProfiles(
	sourceProfile, targetProfile, componentType, componentName string,
) error {
	if err := validateProfileName(sourceProfile); err != nil {
		return fmt.Errorf("invalid source profile name: %w", err)
	}
	if err := validateProfileName(targetProfile); err != nil {
		return fmt.Errorf("invalid target profile name: %w", err)
	}

	if componentType != "skills" && componentType != "agents" && componentType != "commands" {
		return fmt.Errorf("invalid component type '%s'\n\nValid component types:\n  - skills\n  - agents\n  - commands\n\nExample:\n  agent-smith profile copy skills work-profile personal-profile api-design", componentType)
	}

	fmt.Printf("Copying %s '%s' from profile '%s' to profile '%s'...\n", componentType, componentName, sourceProfile, targetProfile)

	srcProfile := pm.loadProfile(sourceProfile)
	if !srcProfile.IsValid() {
		return fmt.Errorf("source profile '%s' does not exist or has no components", sourceProfile)
	}

	dstProfile := pm.loadProfile(targetProfile)
	if !dstProfile.IsValid() {
		return fmt.Errorf("target profile '%s' does not exist or has no components", targetProfile)
	}

	componentPath := filepath.Join(srcProfile.BasePath, componentType, componentName)
	if _, err := os.Stat(componentPath); os.IsNotExist(err) {
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

	return profileCopy.CopyComponentBetweenProfiles(srcProfile.BasePath, dstProfile.BasePath, componentType, componentName, pm.lockService)
}

// AddComponentToProfile copies an existing component from ~/.agent-smith/ to a profile.
func (pm *ProfileManager) AddComponentToProfile(profileName, componentType, componentName string) error {
	if err := validateProfileName(profileName); err != nil {
		return err
	}

	if componentType != "skills" && componentType != "agents" && componentType != "commands" {
		return fmt.Errorf("invalid component type '%s': must be 'skills', 'agents', or 'commands'", componentType)
	}

	fmt.Printf("Adding %s '%s' to profile '%s'...\n", componentType, componentName, profileName)

	profile := pm.loadProfile(profileName)
	if !profile.IsValid() {
		return fmt.Errorf("profile '%s' does not exist or has no components", profileName)
	}

	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return fmt.Errorf("failed to get agents directory: %w", err)
	}

	if err := profileCopy.AddComponentToProfile(agentsDir, profile.BasePath, componentType, componentName, pm.lockService); err != nil {
		return err
	}

	fmt.Printf("✓ Successfully added %s '%s' to profile '%s'\n", componentType, componentName, profileName)
	return nil
}

// RemoveComponentFromProfile removes a component from a profile.
func (pm *ProfileManager) RemoveComponentFromProfile(profileName, componentType, componentName string) error {
	if err := validateProfileName(profileName); err != nil {
		return err
	}

	if componentType != "skills" && componentType != "agents" && componentType != "commands" {
		return fmt.Errorf("invalid component type '%s': must be 'skills', 'agents', or 'commands'", componentType)
	}

	fmt.Printf("Removing %s '%s' from profile '%s'...\n", componentType, componentName, profileName)

	profile := pm.loadProfile(profileName)
	if !profile.IsValid() {
		return fmt.Errorf("profile '%s' does not exist or has no components", profileName)
	}

	activeProfile, err := pm.GetActiveProfile()

	var unlinkFn func() error
	if err == nil && activeProfile == profileName && pm.linker != nil {
		fmt.Printf("Unlinking component from active profile...\n")
		linker := pm.linker
		unlinkFn = func() error {
			return linker.UnlinkComponent(componentType, componentName, "")
		}
	}

	if err := profileCopy.RemoveComponentFromProfile(profile.BasePath, componentType, componentName, unlinkFn); err != nil {
		return err
	}

	// Remove lock file entry
	if removeErr := pm.lockService.RemoveEntry(profile.BasePath, componentType, componentName); removeErr != nil {
		fmt.Printf("Warning: Could not update lock file: %v\n", removeErr)
	}

	fmt.Printf("✓ Successfully removed %s '%s' from profile '%s'\n", componentType, componentName, profileName)
	return nil
}

// CreateProfile creates a new empty profile with the standard directory structure.
func (pm *ProfileManager) CreateProfile(profileName string) error {
	if err := validateProfileName(profileName); err != nil {
		return err
	}

	fmt.Printf("Creating profile '%s'...\n", profileName)

	profile := pm.loadProfile(profileName)
	if profile.IsValid() {
		return fmt.Errorf("profile '%s' already exists", profileName)
	}

	profileDir := filepath.Join(pm.profilesDir, profileName)
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}

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

	if err := pm.SaveUserProfileMetadata(profileName); err != nil {
		fmt.Printf("Warning: Failed to save profile metadata: %v\n", err)
	}

	return nil
}

// CreateProfileWithMetadata creates a new profile with source URL metadata.
func (pm *ProfileManager) CreateProfileWithMetadata(profileName, sourceURL string) error {
	if err := pm.CreateProfile(profileName); err != nil {
		return err
	}

	if sourceURL != "" {
		if err := pm.SaveProfileMetadata(profileName, sourceURL); err != nil {
			fmt.Printf("Warning: Failed to save profile metadata: %v\n", err)
		}
	}

	return nil
}

// DeleteProfile removes a profile and its components.
func (pm *ProfileManager) DeleteProfile(profileName string) error {
	if err := validateProfileName(profileName); err != nil {
		return err
	}

	fmt.Printf("Deleting profile '%s'...\n", profileName)

	profile := pm.loadProfile(profileName)
	if !profile.IsValid() {
		return fmt.Errorf("profile '%s' does not exist", profileName)
	}

	activeProfile, err := pm.GetActiveProfile()
	if err != nil {
		return fmt.Errorf("failed to check active profile: %w", err)
	}

	if activeProfile == profileName {
		if err := pm.DeactivateProfile(); err != nil {
			return fmt.Errorf("failed to deactivate profile before deletion: %w", err)
		}
	}

	fmt.Printf("Cleaning up components...\n")

	if pm.linker != nil {
		componentTypes := []string{"skills", "agents", "commands"}
		for _, componentType := range componentTypes {
			componentDir := filepath.Join(profile.BasePath, componentType)
			if _, err := os.Stat(componentDir); !os.IsNotExist(err) {
				entries, err := os.ReadDir(componentDir)
				if err == nil {
					for _, entry := range entries {
						if !strings.HasPrefix(entry.Name(), ".") {
							_ = pm.linker.UnlinkComponent(componentType, entry.Name(), "")
						}
					}
				}
			}
		}
	}

	profileDir := filepath.Join(pm.profilesDir, profileName)
	if err := os.RemoveAll(profileDir); err != nil {
		return fmt.Errorf("failed to delete profile directory: %w", err)
	}

	fmt.Printf("✓ Successfully deleted profile '%s'\n", profileName)
	return nil
}

// RenameProfile renames a user-created profile.
func (pm *ProfileManager) RenameProfile(oldName, newName string) error {
	if err := validateProfileName(newName); err != nil {
		return err
	}

	oldProfile := pm.loadProfile(oldName)
	if !oldProfile.IsValid() {
		return fmt.Errorf("profile '%s' does not exist", oldName)
	}

	meta, err := pm.LoadProfileMetadata(oldName)
	if err == nil && meta != nil && meta.Type == "repo" {
		return fmt.Errorf("cannot rename repo profile '%s': only user-created profiles can be renamed", oldName)
	}

	newProfile := pm.loadProfile(newName)
	if newProfile.IsValid() {
		return fmt.Errorf("profile '%s' already exists", newName)
	}

	activeProfile, err := pm.GetActiveProfile()
	if err != nil {
		return fmt.Errorf("failed to check active profile: %w", err)
	}

	wasActive := activeProfile == oldName

	if wasActive && pm.linker != nil {
		componentTypes := []string{paths.AgentsSubDir, paths.SkillsSubDir, paths.CommandsSubDir}
		for _, componentType := range componentTypes {
			componentDir := filepath.Join(oldProfile.BasePath, componentType)
			if _, err := os.Stat(componentDir); !os.IsNotExist(err) {
				entries, err := os.ReadDir(componentDir)
				if err == nil {
					for _, entry := range entries {
						if !strings.HasPrefix(entry.Name(), ".") {
							_ = pm.linker.UnlinkComponent(componentType, entry.Name(), "")
						}
					}
				}
			}
		}
	}

	oldPath := filepath.Join(pm.profilesDir, oldName)
	newPath := filepath.Join(pm.profilesDir, newName)

	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename profile directory: %w", err)
	}

	if wasActive {
		agentsDir, err := paths.GetAgentsDir()
		if err != nil {
			return fmt.Errorf("failed to get agents directory after rename: %w", err)
		}
		activeProfilePath := filepath.Join(agentsDir, ".active-profile")
		if err := os.WriteFile(activeProfilePath, []byte(newName), 0644); err != nil {
			return fmt.Errorf("failed to update active profile state: %w", err)
		}

		if pm.linker != nil {
			if err := pm.linker.LinkAllComponents(); err != nil {
				return fmt.Errorf("failed to re-link components after rename: %w", err)
			}
		}
	}

	return nil
}

// GetAllAvailableComponents returns all components from specified profiles.
// If sourceProfiles is empty, returns components from all profiles.
func (pm *ProfileManager) GetAllAvailableComponents(sourceProfiles []string) ([]ComponentItem, error) {
	var items []ComponentItem

	var profilesToScan []*Profile
	if len(sourceProfiles) == 0 {
		allProfiles, err := pm.ScanProfiles()
		if err != nil {
			return nil, fmt.Errorf("failed to scan profiles: %w", err)
		}
		profilesToScan = allProfiles
	} else {
		for _, profileName := range sourceProfiles {
			profile := pm.loadProfile(profileName)
			if !profile.IsValid() {
				return nil, fmt.Errorf("source profile '%s' does not exist or has no components", profileName)
			}
			profilesToScan = append(profilesToScan, profile)
		}
	}

	for _, profile := range profilesToScan {
		agents, skills, commands := pm.GetComponentNames(profile)

		for _, name := range agents {
			items = append(items, ComponentItem{Type: "agents", Name: name, SourceProfile: profile.Name})
		}
		for _, name := range skills {
			items = append(items, ComponentItem{Type: "skills", Name: name, SourceProfile: profile.Name})
		}
		for _, name := range commands {
			items = append(items, ComponentItem{Type: "commands", Name: name, SourceProfile: profile.Name})
		}
	}

	return items, nil
}

// PromptComponentSelection displays an interactive UI for selecting components.
func (pm *ProfileManager) PromptComponentSelection(components []ComponentItem) ([]ComponentItem, error) {
	if len(components) == 0 {
		return nil, fmt.Errorf("no components available for selection")
	}

	var skills, agents, commands []ComponentItem
	for _, c := range components {
		switch c.Type {
		case "skills":
			skills = append(skills, c)
		case "agents":
			agents = append(agents, c)
		case "commands":
			commands = append(commands, c)
		}
	}

	fmt.Println("\nAvailable Components:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if len(skills) > 0 {
		fmt.Printf("\nSkills (%d):\n", len(skills))
		for i, c := range skills {
			fmt.Printf("  [%d] %s  (from: %s)\n", i+1, c.Name, c.SourceProfile)
		}
	}

	if len(agents) > 0 {
		fmt.Printf("\nAgents (%d):\n", len(agents))
		for i, c := range agents {
			fmt.Printf("  [%d] %s  (from: %s)\n", len(skills)+i+1, c.Name, c.SourceProfile)
		}
	}

	if len(commands) > 0 {
		fmt.Printf("\nCommands (%d):\n", len(commands))
		for i, c := range commands {
			fmt.Printf("  [%d] %s  (from: %s)\n", len(skills)+len(agents)+i+1, c.Name, c.SourceProfile)
		}
	}

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("\nSelect components:")
	fmt.Println("  Enter numbers (e.g., 1,3,5 or 1-3)")
	fmt.Println("  Keywords: a=all, s=skills, g=agents, c=commands")
	fmt.Println("  q=quit")
	fmt.Printf("\nSelection: ")

	var response string
	fmt.Scanln(&response) //nolint:errcheck
	response = strings.TrimSpace(response)

	if strings.ToLower(response) == "q" {
		return nil, fmt.Errorf("selection cancelled")
	}

	allComponents := append(append(skills, agents...), commands...)
	indexMap := make(map[int]ComponentItem)
	for i, c := range allComponents {
		indexMap[i+1] = c
	}

	selected := parseSelection(response, indexMap)

	if len(selected) == 0 {
		return nil, fmt.Errorf("no components selected")
	}

	fmt.Println("\nSelected:")
	for _, c := range selected {
		fmt.Printf("  ✓ %s (%s) from %s\n", c.Name, c.Type, c.SourceProfile)
	}

	fmt.Printf("\nCopy to profile? [y/n]: ")
	var confirm string
	fmt.Scanln(&confirm) //nolint:errcheck
	if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
		return nil, fmt.Errorf("cancelled")
	}

	return selected, nil
}

// CherryPickComponents copies selected components from source profiles to target profile.
func (pm *ProfileManager) CherryPickComponents(targetProfile string, components []ComponentItem) error {
	if err := validateProfileName(targetProfile); err != nil {
		return fmt.Errorf("invalid target profile name: %w", err)
	}

	deps := &cherryPickAdapter{pm: pm}
	cherryItems := make([]cherryPickItem, len(components))
	for i, c := range components {
		cherryItems[i] = cherryPickItem{Type: c.Type, Name: c.Name, SourceProfile: c.SourceProfile}
	}

	return runCherryPick(deps, targetProfile, cherryItems)
}

// cherryPickItem mirrors ComponentItem for internal use without creating a dependency cycle.
type cherryPickItem = ComponentItem

// cherryPickAdapter implements the cherry-pick dependencies using ProfileManager.
type cherryPickAdapter struct {
	pm *ProfileManager
}

func (a *cherryPickAdapter) CopyComponentBetweenProfiles(sourceProfile, targetProfile, componentType, componentName string) error {
	return a.pm.CopyComponentBetweenProfiles(sourceProfile, targetProfile, componentType, componentName)
}

func (a *cherryPickAdapter) CreateProfile(profileName string) error {
	return a.pm.CreateProfile(profileName)
}

func (a *cherryPickAdapter) ProfileExists(profileName string) bool {
	return a.pm.loadProfile(profileName).IsValid()
}

func (a *cherryPickAdapter) CountComponents(profileName string) (agents, skills, commands int) {
	profile := a.pm.loadProfile(profileName)
	return a.pm.CountComponents(profile)
}

// runCherryPick is the implementation of cherry-picking without the io.Reader dependency.
func runCherryPick(deps *cherryPickAdapter, targetProfile string, components []ComponentItem) error {
	if !deps.ProfileExists(targetProfile) {
		fmt.Printf("Creating new profile '%s'...\n", targetProfile)
		if err := deps.CreateProfile(targetProfile); err != nil {
			return fmt.Errorf("failed to create target profile: %w", err)
		}
	}

	fmt.Printf("Cherry-picking %d component(s) to profile '%s'...\n\n", len(components), targetProfile)

	successCount := 0
	skipCount := 0
	errorCount := 0

	for _, component := range components {
		fmt.Printf("Copying %s '%s' from '%s'...\n", component.Type, component.Name, component.SourceProfile)

		err := deps.CopyComponentBetweenProfiles(component.SourceProfile, targetProfile, component.Type, component.Name)
		if err != nil {
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

	agents, skills, commands := deps.CountComponents(targetProfile)
	total := agents + skills + commands
	fmt.Printf("%d (%d agents, %d skills, %d commands)\n", total, agents, skills, commands)

	if errorCount > 0 {
		return fmt.Errorf("some components failed to copy")
	}

	return nil
}

// unlinkAllComponents removes all symlinks from the agents directory component folders.
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

// validateProfileName delegates to profilemeta for name validation.
func validateProfileName(name string) error {
	return profilemeta.ValidateProfileName(name)
}

// parseSelection parses user input like "1,3,5" or "1-3" into ComponentItems.
func parseSelection(response string, indexMap map[int]ComponentItem) []ComponentItem {
	selected := make(map[int]ComponentItem)

	for _, part := range strings.Split(response, ",") {
		part = strings.TrimSpace(part)

		switch strings.ToLower(part) {
		case "a", "all":
			for i, c := range indexMap {
				selected[i] = c
			}
			continue
		case "s", "skills":
			for i, c := range indexMap {
				if c.Type == "skills" {
					selected[i] = c
				}
			}
			continue
		case "g", "agents":
			for i, c := range indexMap {
				if c.Type == "agents" {
					selected[i] = c
				}
			}
			continue
		case "c", "commands":
			for i, c := range indexMap {
				if c.Type == "commands" {
					selected[i] = c
				}
			}
			continue
		}

		if strings.Contains(part, "-") {
			rangeParts := strings.SplitN(part, "-", 2)
			if len(rangeParts) == 2 {
				start, err1 := parseInt(rangeParts[0])
				end, err2 := parseInt(rangeParts[1])
				if err1 == nil && err2 == nil && start <= end {
					for i := start; i <= end; i++ {
						if c, ok := indexMap[i]; ok {
							selected[i] = c
						}
					}
				}
			}
			continue
		}

		if num, err := parseInt(part); err == nil {
			if c, ok := indexMap[num]; ok {
				selected[num] = c
			}
		}
	}

	result := make([]ComponentItem, 0, len(selected))
	for _, c := range selected {
		result = append(result, c)
	}
	return result
}

func parseInt(s string) (int, error) {
	s = strings.TrimSpace(s)
	n := 0
	if len(s) == 0 {
		return 0, fmt.Errorf("empty string")
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("not a number: %s", s)
		}
		n = n*10 + int(ch-'0')
	}
	return n, nil
}

// removeComponentEntry removes a lock file entry; thin wrapper around metadata package.
