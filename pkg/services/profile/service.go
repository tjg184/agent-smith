package profile

import (
	"fmt"
	"os"

	"github.com/tgaines/agent-smith/internal/formatter"
	"github.com/tgaines/agent-smith/pkg/logger"
	"github.com/tgaines/agent-smith/pkg/profiles"
	"github.com/tgaines/agent-smith/pkg/services"
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

		// Get metadata for repo profiles
		var sourceURL string
		if profileType == "repo" {
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
		activeIndicator := " "
		if profile.Name == activeProfile {
			activeIndicator = formatter.ColoredSuccess()
		}

		// Add type emoji
		var typeEmoji string
		switch profileType {
		case "repo":
			typeEmoji = "📦"
		case "user":
			typeEmoji = "👤"
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

	// Display total count
	if len(opts.ProfileFilter) > 0 || opts.ActiveOnly || opts.TypeFilter != "" {
		s.formatter.Info("\nShowing: %d profile(s) (filtered from %d total)", len(filteredProfiles), len(profilesList))
	} else {
		s.formatter.Info("\nTotal: %d profile(s)", len(filteredProfiles))
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
	// Need to convert sourceProfiles to ComponentItems
	// This method is complex and interactive, so we delegate to ProfileManager
	// which will handle the interactive selection
	var components []profiles.ComponentItem // Empty for now, actual implementation would gather from sourceProfiles
	if err := s.profileManager.CherryPickComponents(targetProfile, components); err != nil {
		return fmt.Errorf("failed to cherry-pick components: %w", err)
	}

	s.formatter.Info("%s Cherry-picked components into profile '%s'", formatter.SymbolSuccess, targetProfile)
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
