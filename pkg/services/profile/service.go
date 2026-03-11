package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/paths"
	"github.com/tjg184/agent-smith/pkg/profiles"
	"github.com/tjg184/agent-smith/pkg/services"
)

// Service implements the ProfileService interface
type Service struct {
	profileManager *profiles.ProfileManager
	logger         *logger.Logger
	formatter      *formatter.Formatter
}

// NewService creates a new ProfileService with the given dependencies
func NewService(pm *profiles.ProfileManager, logger *logger.Logger, formatter *formatter.Formatter) services.ProfileService {
	return &Service{
		profileManager: pm,
		logger:         logger,
		formatter:      formatter,
	}
}

// ListProfiles lists profiles with optional filtering
func (s *Service) ListProfiles(opts services.ListProfileOptions) error {
	// Validate typeFilter if provided
	if opts.TypeFilter != "" && opts.TypeFilter != "repo" && opts.TypeFilter != "user" {
		return fmt.Errorf("invalid type filter '%s'. Valid values are: repo, user", opts.TypeFilter)
	}

	profilesList, err := s.profileManager.ScanProfiles()
	if err != nil {
		return fmt.Errorf("failed to scan profiles: %w", err)
	}

	baseAgentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return fmt.Errorf("failed to get base agents directory: %w", err)
	}

	baseProfile := s.scanBaseInstallation(baseAgentsDir)

	// Prepend base to the list if it has components
	allProfiles := []*profiles.Profile{}
	if baseProfile != nil {
		allProfiles = append(allProfiles, baseProfile)
	}
	allProfiles = append(allProfiles, profilesList...)
	profilesList = allProfiles

	// Get active profile
	activeProfile, err := s.profileManager.GetActiveProfile()
	if err != nil {
		return fmt.Errorf("failed to get active profile: %w", err)
	}

	// Apply filters
	var filteredProfiles []*profiles.Profile

	// Filter by active-only if specified
	if opts.ActiveOnly {
		for _, profile := range profilesList {
			if profile.Name == activeProfile {
				filteredProfiles = append(filteredProfiles, profile)
				break
			}
		}
	} else if len(opts.ProfileFilter) > 0 {
		// Filter by specific profile names
		filterMap := make(map[string]bool)
		for _, name := range opts.ProfileFilter {
			filterMap[name] = true
		}

		// Validate that all filter names exist
		profileMap := make(map[string]bool)
		for _, p := range profilesList {
			profileMap[p.Name] = true
		}

		for _, filterName := range opts.ProfileFilter {
			if !profileMap[filterName] {
				return fmt.Errorf("profile '%s' does not exist", filterName)
			}
		}

		// Apply filter
		for _, p := range profilesList {
			if filterMap[p.Name] {
				filteredProfiles = append(filteredProfiles, p)
			}
		}
	} else {
		// No filters, show all profiles
		filteredProfiles = profilesList
	}

	// Apply type filter if specified
	if opts.TypeFilter != "" {
		var typeFilteredProfiles []*profiles.Profile
		for _, profile := range filteredProfiles {
			profileType, err := s.profileManager.GetProfileType(profile.Name)
			if err != nil {
				// Log warning but continue
				continue
			}
			if profileType == opts.TypeFilter {
				typeFilteredProfiles = append(typeFilteredProfiles, profile)
			}
		}
		filteredProfiles = typeFilteredProfiles
	}

	// Display results
	if len(filteredProfiles) == 0 {
		if opts.ActiveOnly {
			s.formatter.Info("No active profile set")
		} else if len(opts.ProfileFilter) > 0 {
			s.formatter.Info("No matching profiles found")
		} else {
			s.formatter.Info("No profiles found in ~/.agent-smith/profiles/")
			s.formatter.EmptyLine()
			s.formatter.Info("To create a profile, run:")
			s.formatter.Info("  agent-smith profile create <profile-name>")
		}
		return nil
	}

	// Create table with box-drawing characters
	table := formatter.NewBoxTable(os.Stdout, []string{"Profile", "Components"})

	// Add rows to table
	for _, profile := range filteredProfiles {
		// Get profile type and metadata
		profileType, err := s.profileManager.GetProfileType(profile.Name)
		if err != nil {
			profileType = "unknown"
		}

		// Get metadata for repo profiles (skip for base)
		var sourceURL string
		if profile.Name != paths.BaseProfileName && profileType == "repo" {
			metadata, err := s.profileManager.LoadProfileMetadata(profile.Name)
			if err == nil && metadata != nil {
				sourceURL = metadata.SourceURL
			}
		}

		// Count components
		agents, skills, commands := s.profileManager.CountComponents(profile)

		// Build component counts string with proper singular/plural handling
		var components []string
		if agents > 0 {
			if agents == 1 {
				components = append(components, "1 agent")
			} else {
				components = append(components, fmt.Sprintf("%d agents", agents))
			}
		}
		if skills > 0 {
			if skills == 1 {
				components = append(components, "1 skill")
			} else {
				components = append(components, fmt.Sprintf("%d skills", skills))
			}
		}
		if commands > 0 {
			if commands == 1 {
				components = append(components, "1 command")
			} else {
				components = append(components, fmt.Sprintf("%d commands", commands))
			}
		}

		componentStr := ""
		if len(components) > 0 {
			componentStr = fmt.Sprintf("(%s)", joinStrings(components, ", "))
		} else {
			componentStr = "(empty)"
		}

		// Build profile cell with active indicator and type emoji
		// Never show active indicator for base
		activeIndicator := " "
		if profile.Name != paths.BaseProfileName && profile.Name == activeProfile {
			activeIndicator = formatter.ColoredSuccess()
		}

		// Add type emoji
		var typeEmoji string
		switch profileType {
		case "repo":
			typeEmoji = "📦"
		case "user":
			typeEmoji = "👤"
		case "base":
			typeEmoji = "⊙"
		default:
			typeEmoji = "❓"
		}

		// Build profile name with source URL for repo types
		profileName := profile.Name
		if profileType == "repo" && sourceURL != "" {
			profileName = fmt.Sprintf("%s (%s)", profile.Name, sourceURL)
		}

		profileCell := fmt.Sprintf("%s %s %s", activeIndicator, typeEmoji, profileName)

		// Add row to table
		table.AddRow([]string{profileCell, componentStr})
	}

	// Render the table
	table.Render()

	// Display legend
	s.formatter.EmptyLine()
	s.formatter.Info("Legend:")
	s.formatter.Info("  %s - Currently active profile", formatter.ColoredSuccess())
	s.formatter.Info("  📦 - Repository-sourced profile")
	s.formatter.Info("  👤 - User-created profile")
	s.formatter.Info("  ⊙ - Base installation (no profile)")

	// Count base separately from profiles
	baseCount := 0
	profileCount := len(filteredProfiles)

	for _, p := range filteredProfiles {
		if p.Name == paths.BaseProfileName {
			baseCount = 1
			profileCount--
			break
		}
	}

	// Display appropriate count string
	s.formatter.EmptyLine()
	if opts.ProfileFilter != nil || opts.ActiveOnly || opts.TypeFilter != "" {
		// For filtered views, just show count
		if baseCount > 0 && profileCount > 0 {
			s.formatter.Info("Showing: %d profile(s) + base installation", profileCount)
		} else if baseCount > 0 {
			s.formatter.Info("Showing: base installation only")
		} else {
			s.formatter.Info("Showing: %d profile(s)", profileCount)
		}
	} else {
		// For unfiltered view, show total
		if baseCount > 0 && profileCount > 0 {
			s.formatter.Info("Total: %d profile(s) + base installation", profileCount)
		} else if baseCount > 0 {
			s.formatter.Info("Total: base installation only")
		} else {
			s.formatter.Info("Total: %d profile(s)", profileCount)
		}
	}

	return nil
}

// ShowProfile displays detailed information about a specific profile
func (s *Service) ShowProfile(name string) error {
	// Load the profile
	profilesList, err := s.profileManager.ScanProfiles()
	if err != nil {
		return fmt.Errorf("failed to scan profiles: %w", err)
	}

	var targetProfile *profiles.Profile
	for _, p := range profilesList {
		if p.Name == name {
			targetProfile = p
			break
		}
	}

	if targetProfile == nil {
		return fmt.Errorf("profile '%s' not found", name)
	}

	// Get active profile to show status
	activeProfile, err := s.profileManager.GetActiveProfile()
	if err != nil {
		return fmt.Errorf("failed to get active profile: %w", err)
	}

	// Display profile information
	s.formatter.Info("Profile: %s", targetProfile.Name)
	if targetProfile.Name == activeProfile {
		s.formatter.Info(" %s [active]", formatter.SymbolSuccess)
	}
	s.formatter.EmptyLine()
	s.formatter.Info("Location: %s", targetProfile.BasePath)
	s.formatter.EmptyLine()

	// Get component names
	agents, skills, commands := s.profileManager.GetComponentNames(targetProfile)

	// Display agents
	if len(agents) > 0 {
		s.formatter.Info("Agents (%d):", len(agents))
		for _, agent := range agents {
			sourceURL := s.profileManager.GetComponentSource(targetProfile, "agents", agent)
			if sourceURL != "" {
				s.formatter.Info("  - %s (%s)", agent, sourceURL)
			} else {
				s.formatter.Info("  - %s", agent)
			}
		}
		s.formatter.EmptyLine()
	}

	// Display skills
	if len(skills) > 0 {
		s.formatter.Info("Skills (%d):", len(skills))
		for _, skill := range skills {
			sourceURL := s.profileManager.GetComponentSource(targetProfile, "skills", skill)
			if sourceURL != "" {
				s.formatter.Info("  - %s (%s)", skill, sourceURL)
			} else {
				s.formatter.Info("  - %s", skill)
			}
		}
		s.formatter.EmptyLine()
	}

	// Display commands
	if len(commands) > 0 {
		s.formatter.Info("Commands (%d):", len(commands))
		for _, command := range commands {
			sourceURL := s.profileManager.GetComponentSource(targetProfile, "commands", command)
			if sourceURL != "" {
				s.formatter.Info("  - %s (%s)", command, sourceURL)
			} else {
				s.formatter.Info("  - %s", command)
			}
		}
		s.formatter.EmptyLine()
	}

	// If profile is empty, show helpful message
	if len(agents) == 0 && len(skills) == 0 && len(commands) == 0 {
		s.formatter.Info("This profile is empty.")
		s.formatter.EmptyLine()
		s.formatter.Info("To add components:")
		s.formatter.Info("  agent-smith install <repo-url> --profile %s", name)
		s.formatter.EmptyLine()
	}

	return nil
}

// CreateProfile creates a new profile
func (s *Service) CreateProfile(name string) error {
	if err := s.profileManager.CreateProfile(name); err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}

	s.formatter.Info("%s Successfully created profile '%s'", formatter.SymbolSuccess, name)
	s.formatter.EmptyLine()
	s.formatter.Info("To add components to this profile:")
	s.formatter.Info("  agent-smith install <repo-url> --profile %s", name)
	s.formatter.EmptyLine()
	s.formatter.Info("To activate this profile:")
	s.formatter.Info("  agent-smith profile activate %s", name)

	return nil
}

// DeleteProfile deletes a profile
func (s *Service) DeleteProfile(name string) error {
	// Check if profile is currently active
	activeProfile, err := s.profileManager.GetActiveProfile()
	if err != nil {
		return fmt.Errorf("failed to get active profile: %w", err)
	}

	if name == activeProfile {
		s.formatter.EmptyLine()
		s.formatter.WarningMsg("Cannot delete active profile")
		s.formatter.EmptyLine()
		s.formatter.InfoMsg("To delete this profile:")
		s.formatter.InfoMsg("  1. Deactivate it first:")
		s.formatter.InfoMsg("     agent-smith profile deactivate")
		s.formatter.InfoMsg("  2. Then delete it:")
		s.formatter.InfoMsg("     agent-smith profile delete %s", name)
		return fmt.Errorf("cannot delete active profile '%s'", name)
	}

	if err := s.profileManager.DeleteProfile(name); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	s.formatter.Info("%s Successfully deleted profile '%s'", formatter.SymbolSuccess, name)
	return nil
}

// ActivateProfile activates a profile
func (s *Service) ActivateProfile(name string) error {
	result, err := s.profileManager.ActivateProfileWithResult(name)
	if err != nil {
		return fmt.Errorf("failed to activate profile: %w", err)
	}

	// Display appropriate message based on whether we switched or it was already active
	if result.Switched {
		s.formatter.Info("%s Switched profile: %s → %s", formatter.ColoredSuccess(), result.PreviousProfile, result.NewProfile)
	} else if result.PreviousProfile == result.NewProfile {
		s.formatter.Info("%s Profile '%s' is already active", formatter.ColoredSuccess(), name)
	} else {
		s.formatter.Info("%s Profile '%s' activated", formatter.ColoredSuccess(), name)
	}

	s.formatter.EmptyLine()
	s.formatter.Info("Components from this profile are now ready to be linked:")
	s.formatter.Info("  agent-smith link all")

	return nil
}

// DeactivateProfile deactivates the current profile
func (s *Service) DeactivateProfile() error {
	// Check if there's an active profile
	activeProfile, err := s.profileManager.GetActiveProfile()
	if err != nil {
		return fmt.Errorf("failed to get active profile: %w", err)
	}

	if activeProfile == "" {
		s.formatter.Info("No active profile to deactivate")
		return nil
	}

	if err := s.profileManager.DeactivateProfile(); err != nil {
		return fmt.Errorf("failed to deactivate profile: %w", err)
	}

	s.formatter.Info("%s Profile '%s' deactivated", formatter.ColoredSuccess(), activeProfile)
	s.formatter.EmptyLine()
	s.formatter.Info("To activate a profile:")
	s.formatter.Info("  agent-smith profile activate <profile-name>")

	return nil
}

// AddComponent adds a component to a profile
func (s *Service) AddComponent(componentType, profileName, componentName string) error {
	if err := s.profileManager.AddComponentToProfile(profileName, componentType, componentName); err != nil {
		return fmt.Errorf("failed to add component: %w", err)
	}

	s.formatter.Info("%s Added %s '%s' to profile '%s'", formatter.SymbolSuccess, componentType, componentName, profileName)
	return nil
}

// CopyComponent copies a component from one profile to another
func (s *Service) CopyComponent(sourceProfile, targetProfile, componentType, componentName string) error {
	if err := s.profileManager.CopyComponentBetweenProfiles(sourceProfile, targetProfile, componentType, componentName); err != nil {
		return fmt.Errorf("failed to copy component: %w", err)
	}

	s.formatter.Info("%s Copied %s '%s' from '%s' to '%s'", formatter.SymbolSuccess, componentType, componentName, sourceProfile, targetProfile)
	return nil
}

// RemoveComponent removes a component from a profile
func (s *Service) RemoveComponent(profileName, componentType, componentName string) error {
	if err := s.profileManager.RemoveComponentFromProfile(profileName, componentType, componentName); err != nil {
		return fmt.Errorf("failed to remove component: %w", err)
	}

	s.formatter.Info("%s Removed %s '%s' from profile '%s'", formatter.SymbolSuccess, componentType, componentName, profileName)
	return nil
}

// CherryPickComponents allows selecting components from multiple source profiles
func (s *Service) CherryPickComponents(targetProfile string, sourceProfiles []string) error {
	components, err := s.profileManager.GetAllAvailableComponents(sourceProfiles)
	if err != nil {
		return fmt.Errorf("failed to get available components: %w", err)
	}

	if len(components) == 0 {
		return fmt.Errorf("no components found in source profile(s)")
	}

	selectedComponents, err := s.profileManager.PromptComponentSelection(components)
	if err != nil {
		return err
	}

	if len(selectedComponents) == 0 {
		return fmt.Errorf("no components selected")
	}

	if err := s.profileManager.CherryPickComponents(targetProfile, selectedComponents); err != nil {
		return fmt.Errorf("failed to cherry-pick components: %w", err)
	}

	s.formatter.Info("%s Cherry-picked %d component(s) into profile '%s'", formatter.SymbolSuccess, len(selectedComponents), targetProfile)
	return nil
}

// joinStrings joins strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// scanBaseInstallation creates a pseudo-profile for base installation
// Returns nil if base has no components
func (s *Service) scanBaseInstallation(baseDir string) *profiles.Profile {
	baseProfile := &profiles.Profile{
		Name:     paths.BaseProfileName,
		BasePath: baseDir,
	}

	hasComponents := false

	// Check for skills directory
	skillsDir := filepath.Join(baseDir, "skills")
	if entries, err := os.ReadDir(skillsDir); err == nil && len(entries) > 0 {
		// Count non-hidden directories
		for _, entry := range entries {
			if entry.IsDir() && !isHidden(entry.Name()) {
				baseProfile.HasSkills = true
				hasComponents = true
				break
			}
		}
	}

	// Check for agents directory
	agentsDir := filepath.Join(baseDir, "agents")
	if entries, err := os.ReadDir(agentsDir); err == nil && len(entries) > 0 {
		// Count non-hidden directories
		for _, entry := range entries {
			if entry.IsDir() && !isHidden(entry.Name()) {
				baseProfile.HasAgents = true
				hasComponents = true
				break
			}
		}
	}

	// Check for commands directory
	commandsDir := filepath.Join(baseDir, "commands")
	if entries, err := os.ReadDir(commandsDir); err == nil && len(entries) > 0 {
		// Count non-hidden directories
		for _, entry := range entries {
			if entry.IsDir() && !isHidden(entry.Name()) {
				baseProfile.HasCommands = true
				hasComponents = true
				break
			}
		}
	}

	if !hasComponents {
		return nil
	}

	return baseProfile
}

// isHidden returns true if the filename starts with a dot
func isHidden(name string) bool {
	return len(name) > 0 && name[0] == '.'
}

// ShareProfile generates commands to recreate a profile
func (s *Service) ShareProfile(profileName, outputPath string) error {
	// Validate profile exists
	profilesList, err := s.profileManager.ScanProfiles()
	if err != nil {
		return fmt.Errorf("failed to scan profiles: %w", err)
	}

	// Check if it's the base profile
	isBase := profileName == paths.BaseProfileName
	var targetProfile *profiles.Profile

	if isBase {
		// Get base installation directory
		baseAgentsDir, err := paths.GetAgentsDir()
		if err != nil {
			return fmt.Errorf("failed to get base directory: %w", err)
		}
		targetProfile = s.scanBaseInstallation(baseAgentsDir)
		if targetProfile == nil {
			return fmt.Errorf("base installation is empty - no components to share")
		}
	} else {
		// Find the named profile
		for _, p := range profilesList {
			if p.Name == profileName {
				targetProfile = p
				break
			}
		}

		if targetProfile == nil {
			return fmt.Errorf("profile '%s' not found", profileName)
		}
	}

	// Generate commands
	commands, err := s.generateProfileCommands(targetProfile, isBase)
	if err != nil {
		return fmt.Errorf("failed to generate commands: %w", err)
	}

	// Output to file or stdout
	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(commands), 0644); err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
		s.formatter.Info("%s Commands saved to: %s", formatter.SymbolSuccess, outputPath)
	} else {
		fmt.Print(commands)
	}

	return nil
}

// generateProfileCommands creates the full command output for a profile
func (s *Service) generateProfileCommands(profile *profiles.Profile, isBase bool) (string, error) {
	var buf strings.Builder
	now := time.Now().Format("2006-01-02")

	// Header
	buf.WriteString(fmt.Sprintf("# Agent Smith Profile: %s\n", profile.Name))
	buf.WriteString(fmt.Sprintf("# Generated on: %s\n", now))
	buf.WriteString("#\n")
	buf.WriteString("# To recreate this profile, copy and run these commands:\n\n")

	// Profile creation (skip for base)
	if !isBase {
		buf.WriteString(fmt.Sprintf("agent-smith profile create %s\n", profile.Name))
		buf.WriteString(fmt.Sprintf("agent-smith profile activate %s\n\n", profile.Name))
	}

	// Get components from lock files
	skillCommands, skillShareable, skillTotal := s.generateComponentCommands(profile, "skills", isBase)
	agentCommands, agentShareable, agentTotal := s.generateComponentCommands(profile, "agents", isBase)
	commandCommands, commandShareable, commandTotal := s.generateComponentCommands(profile, "commands", isBase)

	totalShareable := skillShareable + agentShareable + commandShareable
	totalComponents := skillTotal + agentTotal + commandTotal

	if totalShareable == 0 {
		buf.WriteString("# This profile has no shareable components\n")
		if totalComponents > 0 {
			buf.WriteString(fmt.Sprintf("# Note: Found %d local-only component(s) that cannot be shared:\n", totalComponents))
			if skillTotal > 0 {
				buf.WriteString(fmt.Sprintf("#   - %d skill(s)\n", skillTotal))
			}
			if agentTotal > 0 {
				buf.WriteString(fmt.Sprintf("#   - %d agent(s)\n", agentTotal))
			}
			if commandTotal > 0 {
				buf.WriteString(fmt.Sprintf("#   - %d command(s)\n", commandTotal))
			}
			buf.WriteString("# Local components must be installed from a repository to be shareable\n")
		}
		return buf.String(), nil
	}

	// Add skill commands
	if skillShareable > 0 {
		buf.WriteString(fmt.Sprintf("# Install skills (%d components)\n", skillShareable))
		buf.WriteString(skillCommands)
		buf.WriteString("\n")
	}

	// Add agent commands
	if agentShareable > 0 {
		buf.WriteString(fmt.Sprintf("# Install agents (%d components)\n", agentShareable))
		buf.WriteString(agentCommands)
		buf.WriteString("\n")
	}

	// Add command commands
	if commandShareable > 0 {
		buf.WriteString(fmt.Sprintf("# Install commands (%d components)\n", commandShareable))
		buf.WriteString(commandCommands)
		buf.WriteString("\n")
	}

	// Footer
	buf.WriteString("# Link to your editor (optional)\n")
	buf.WriteString("agent-smith link all\n\n")

	// Summary
	buf.WriteString(fmt.Sprintf("# Total: %d shareable components (", totalShareable))
	parts := []string{}
	if skillShareable > 0 {
		parts = append(parts, fmt.Sprintf("%d skills", skillShareable))
	}
	if agentShareable > 0 {
		parts = append(parts, fmt.Sprintf("%d agents", agentShareable))
	}
	if commandShareable > 0 {
		parts = append(parts, fmt.Sprintf("%d commands", commandShareable))
	}
	buf.WriteString(strings.Join(parts, ", "))
	buf.WriteString(")\n")

	// Note about local components if any were skipped
	localCount := totalComponents - totalShareable
	if localCount > 0 {
		buf.WriteString(fmt.Sprintf("# Note: %d local-only component(s) not included\n", localCount))
	}

	return buf.String(), nil
}

// generateComponentCommands generates install commands for a component type
// Returns: (commands string, shareable count, total count)
func (s *Service) generateComponentCommands(profile *profiles.Profile, componentType string, isBase bool) (string, int, int) {
	var buf strings.Builder
	shareableCount := 0

	// Get component names from filesystem (source of truth)
	agents, skills, commands := s.profileManager.GetComponentNames(profile)

	// Select the appropriate component list based on type
	var componentNames []string
	switch componentType {
	case "skills":
		componentNames = skills
	case "agents":
		componentNames = agents
	case "commands":
		componentNames = commands
	default:
		return "", 0, 0
	}

	totalCount := len(componentNames)

	// Generate install commands for each component
	singularType := strings.TrimSuffix(componentType, "s") // "skills" -> "skill"
	for _, componentName := range componentNames {
		// Get source URL from lock file (for enrichment)
		sourceURL := s.profileManager.GetComponentSource(profile, componentType, componentName)

		// Skip components without source URLs (can't be shared) or with local paths
		if sourceURL == "" || isLocalPath(sourceURL) {
			continue
		}

		if isBase {
			// Base installation doesn't use --profile flag
			buf.WriteString(fmt.Sprintf("agent-smith install %s %s %s\n",
				singularType,
				sourceURL,
				componentName))
		} else {
			// Named profile uses --profile flag
			buf.WriteString(fmt.Sprintf("agent-smith install %s %s %s --profile %s\n",
				singularType,
				sourceURL,
				componentName,
				profile.Name))
		}
		shareableCount++
	}

	return buf.String(), shareableCount, totalCount
}

// isLocalPath checks if a path is a local file path
func isLocalPath(path string) bool {
	return strings.HasPrefix(path, "/") ||
		strings.HasPrefix(path, "file://") ||
		strings.HasPrefix(path, "~/") ||
		strings.HasPrefix(path, ".") ||
		(len(path) > 1 && path[1] == ':') // Windows drive letter
}

// RenameProfile renames a user-created profile.
func (s *Service) RenameProfile(oldName, newName string) error {
	activeProfile, err := s.profileManager.GetActiveProfile()
	if err != nil {
		return fmt.Errorf("failed to get active profile: %w", err)
	}

	wasActive := activeProfile == oldName

	if wasActive {
		fmt.Printf("Profile '%s' is currently active. Rename it? [y/N]: ", oldName)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("Rename cancelled.")
			return nil
		}
	}

	if err := s.profileManager.RenameProfile(oldName, newName); err != nil {
		return fmt.Errorf("failed to rename profile: %w", err)
	}

	s.formatter.Info("%s Renamed profile: %s → %s", formatter.ColoredSuccess(), oldName, newName)

	if wasActive {
		s.formatter.InfoMsg("Profile '%s' is now active. Run 'agent-smith link all' to restore links.", newName)
	}

	return nil
}
