package link

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/tgaines/agent-smith/internal/detector"
	"github.com/tgaines/agent-smith/internal/formatter"
	"github.com/tgaines/agent-smith/internal/linker"
	"github.com/tgaines/agent-smith/pkg/config"
	"github.com/tgaines/agent-smith/pkg/logger"
	"github.com/tgaines/agent-smith/pkg/paths"
	"github.com/tgaines/agent-smith/pkg/profiles"
	"github.com/tgaines/agent-smith/pkg/services"
)

// Service implements the LinkService interface
type Service struct {
	profileManager *profiles.ProfileManager
	logger         *logger.Logger
	formatter      *formatter.Formatter
}

// NewService creates a new LinkService with the given dependencies
func NewService(pm *profiles.ProfileManager, logger *logger.Logger, formatter *formatter.Formatter) services.LinkService {
	return &Service{
		profileManager: pm,
		logger:         logger,
		formatter:      formatter,
	}
}

// showCurrentContext displays the current profile context at the start of link operations
func (s *Service) showCurrentContext(explicitProfile string) {
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	gray := color.New(color.FgHiBlack).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	s.formatter.EmptyLine()
	fmt.Printf("%s\n", bold("Current Context:"))

	if explicitProfile != "" {
		fmt.Printf("  Profile: %s (explicit)\n", cyan(explicitProfile))
	} else {
		activeProfile, err := s.profileManager.GetActiveProfile()
		if err != nil {
			fmt.Printf("  Profile: %s\n", gray("unknown (error checking)"))
		} else if activeProfile != "" {
			fmt.Printf("  Profile: %s\n", green(activeProfile))
		} else {
			fmt.Printf("  Profile: %s\n", gray("none (using base installation)"))
		}
	}

	s.formatter.EmptyLine()
}

// profileManagerAdapter adapts profiles.ProfileManager to linker.ProfileManager
type profileManagerAdapter struct {
	pm *profiles.ProfileManager
}

func (pma *profileManagerAdapter) ScanProfiles() ([]*linker.Profile, error) {
	profiles, err := pma.pm.ScanProfiles()
	if err != nil {
		return nil, err
	}

	// Convert profiles.Profile to linker.Profile
	result := make([]*linker.Profile, len(profiles))
	for i, p := range profiles {
		result[i] = &linker.Profile{
			Name:        p.Name,
			BasePath:    p.BasePath,
			HasAgents:   p.HasAgents,
			HasSkills:   p.HasSkills,
			HasCommands: p.HasCommands,
		}
	}
	return result, nil
}

func (pma *profileManagerAdapter) GetActiveProfile() (string, error) {
	return pma.pm.GetActiveProfile()
}

// createLinker creates a ComponentLinker with the default configuration
func (s *Service) createLinker() (*linker.ComponentLinker, error) {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}

	// Check if a profile is active and use its path instead
	activeProfile, err := s.profileManager.GetActiveProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to get active profile: %w", err)
	}

	// If a profile is active, use the profile's base path instead
	if activeProfile != "" {
		profilesDir, err := paths.GetProfilesDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get profiles directory: %w", err)
		}
		agentsDir = filepath.Join(profilesDir, activeProfile)
		s.formatter.Info("Using active profile: %s", activeProfile)
	}

	// Detect all available targets
	targets, err := config.DetectAllTargets()
	if err != nil {
		return nil, fmt.Errorf("failed to detect targets: %w", err)
	}

	det := detector.NewRepositoryDetector()
	if s.logger != nil {
		det.SetLogger(s.logger)
	}

	return linker.NewComponentLinker(agentsDir, targets, det, nil)
}

// createLinkerWithFilter creates a ComponentLinker with target filtering
func (s *Service) createLinkerWithFilter(targetFilter string) (*linker.ComponentLinker, error) {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}

	// Check if a profile is active
	activeProfile, err := s.profileManager.GetActiveProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to get active profile: %w", err)
	}

	if activeProfile != "" {
		profilesDir, err := paths.GetProfilesDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get profiles directory: %w", err)
		}
		agentsDir = filepath.Join(profilesDir, activeProfile)
	}

	// Detect targets and apply filter
	allTargets, err := config.DetectAllTargets()
	if err != nil {
		return nil, fmt.Errorf("failed to detect targets: %w", err)
	}

	targets := s.filterTargets(allTargets, targetFilter)

	det := detector.NewRepositoryDetector()
	if s.logger != nil {
		det.SetLogger(s.logger)
	}

	return linker.NewComponentLinker(agentsDir, targets, det, nil)
}

// createLinkerWithFilterAndProfile creates a ComponentLinker with target filtering and explicit profile
func (s *Service) createLinkerWithFilterAndProfile(targetFilter string, profile string) (*linker.ComponentLinker, error) {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}

	// Check if an explicit profile was specified
	if profile != "" {
		profilesDir, err := paths.GetProfilesDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get profiles directory: %w", err)
		}

		// Validate that the profile exists
		profilePath := filepath.Join(profilesDir, profile)
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			availableProfiles, scanErr := s.profileManager.ScanProfiles()
			if scanErr == nil && len(availableProfiles) > 0 {
				profileNames := make([]string, len(availableProfiles))
				for i, p := range availableProfiles {
					profileNames[i] = p.Name
				}
				return nil, fmt.Errorf("profile '%s' does not exist\n\nAvailable profiles:\n  - %s\n\nTo create this profile:\n  agent-smith profile create %s",
					profile, strings.Join(profileNames, "\n  - "), profile)
			}
			return nil, fmt.Errorf("profile '%s' does not exist\n\nTo create this profile:\n  agent-smith profile create %s\n\nTo list available profiles:\n  agent-smith profile list", profile, profile)
		}

		agentsDir = profilePath
		s.formatter.Info("Using profile: %s", profile)
	} else {
		// Use active profile logic
		activeProfile, err := s.profileManager.GetActiveProfile()
		if err != nil {
			return nil, fmt.Errorf("failed to get active profile: %w", err)
		}

		if activeProfile != "" {
			profilesDir, err := paths.GetProfilesDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get profiles directory: %w", err)
			}
			agentsDir = filepath.Join(profilesDir, activeProfile)
		}
	}

	// Detect targets and apply filter
	allTargets, err := config.DetectAllTargets()
	if err != nil {
		return nil, fmt.Errorf("failed to detect targets: %w", err)
	}

	targets := s.filterTargets(allTargets, targetFilter)

	det := detector.NewRepositoryDetector()
	if s.logger != nil {
		det.SetLogger(s.logger)
	}

	return linker.NewComponentLinker(agentsDir, targets, det, nil)
}

// createLinkerWithProfileManager creates a ComponentLinker with ProfileManager for multi-profile operations
func (s *Service) createLinkerWithProfileManager() (*linker.ComponentLinker, error) {
	// For multi-profile view, use base directory as the starting point
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}

	// Detect all available targets
	targets, err := config.DetectAllTargets()
	if err != nil {
		return nil, fmt.Errorf("failed to detect targets: %w", err)
	}

	det := detector.NewRepositoryDetector()
	if s.logger != nil {
		det.SetLogger(s.logger)
	}

	// Wrap the ProfileManager in an adapter
	adapter := &profileManagerAdapter{pm: s.profileManager}

	return linker.NewComponentLinker(agentsDir, targets, det, adapter)
}

// filterTargets filters targets based on the targetFilter string
func (s *Service) filterTargets(targets []config.Target, targetFilter string) []config.Target {
	if targetFilter == "" || targetFilter == "all" {
		return targets
	}

	var filtered []config.Target
	for _, target := range targets {
		if target.GetName() == targetFilter {
			filtered = append(filtered, target)
		}
	}
	return filtered
}

// LinkComponent links a single component to targets
func (s *Service) LinkComponent(componentType, componentName string, opts services.LinkOptions) error {
	s.showCurrentContext(opts.Profile)
	cl, err := s.createLinkerWithFilterAndProfile(opts.TargetFilter, opts.Profile)
	if err != nil {
		return fmt.Errorf("failed to create component linker: %w", err)
	}
	return cl.LinkComponent(componentType, componentName)
}

// LinkAll links all components to targets
func (s *Service) LinkAll(opts services.LinkOptions) error {
	// Validate flag combination
	if opts.AllProfiles && opts.Profile != "" {
		return fmt.Errorf("cannot use both --all-profiles and --profile flags together")
	}

	if opts.AllProfiles {
		// Link from all profiles
		profilesList, err := s.profileManager.ScanProfiles()
		if err != nil {
			s.formatter.EmptyLine()
			s.formatter.ErrorMsg("Failed to scan profiles")
			s.formatter.DetailItem("Error", err.Error())
			s.formatter.EmptyLine()
			s.formatter.InfoMsg("The --all-profiles flag requires at least one profile.")
			s.formatter.InfoMsg("Try running without --all-profiles, or create a profile first:")
			s.formatter.InfoMsg("  agent-smith profile create <name>")
			return err
		}

		if len(profilesList) == 0 {
			s.formatter.EmptyLine()
			s.formatter.InfoMsg("No profiles found")
			s.formatter.EmptyLine()
			s.formatter.InfoMsg("The --all-profiles flag requires at least one profile.")
			s.formatter.InfoMsg("Options:")
			s.formatter.InfoMsg("  1. Run without --all-profiles to link components from base installation")
			s.formatter.InfoMsg("  2. Create a profile first: agent-smith profile create <name>")
			return fmt.Errorf("no profiles found")
		}

		// Color helpers
		bold := color.New(color.Bold).SprintFunc()
		green := color.New(color.FgGreen, color.Bold).SprintFunc()
		cyan := color.New(color.FgCyan).SprintFunc()
		gray := color.New(color.FgHiBlack).SprintFunc()

		// Link from each profile
		fmt.Printf("\n%s\n", bold("Linking components from all profiles..."))
		fmt.Println()

		for _, profileItem := range profilesList {
			fmt.Printf("%s\n", cyan(fmt.Sprintf("Profile: %s", profileItem.Name)))

			cl, err := s.createLinkerWithFilterAndProfile(opts.TargetFilter, profileItem.Name)
			if err != nil {
				return fmt.Errorf("failed to create component linker for profile '%s': %w", profileItem.Name, err)
			}

			// Count components before linking to check if profile has any
			agents, skills, commands := s.profileManager.CountComponents(profileItem)
			totalComponents := agents + skills + commands

			if totalComponents == 0 {
				fmt.Printf("  %s\n\n", gray("(no components)"))
				continue
			}

			if err := cl.LinkAllComponents(); err != nil {
				return fmt.Errorf("failed to link components from profile '%s': %w", profileItem.Name, err)
			}
		}

		fmt.Printf("\n%s\n", green("✓ Successfully linked components from all profiles"))
		return nil
	}

	// Link from single profile (existing behavior)
	s.showCurrentContext(opts.Profile)
	cl, err := s.createLinkerWithFilterAndProfile(opts.TargetFilter, opts.Profile)
	if err != nil {
		return fmt.Errorf("failed to create component linker: %w", err)
	}
	return cl.LinkAllComponents()
}

// LinkByType links all components of a specific type to targets
func (s *Service) LinkByType(componentType string, opts services.LinkOptions) error {
	s.showCurrentContext(opts.Profile)
	cl, err := s.createLinkerWithFilterAndProfile(opts.TargetFilter, opts.Profile)
	if err != nil {
		return fmt.Errorf("failed to create component linker: %w", err)
	}
	return cl.LinkComponentsByType(componentType)
}

// UnlinkComponent unlinks a single component from targets
func (s *Service) UnlinkComponent(componentType, componentName string, opts services.UnlinkOptions) error {
	var cl *linker.ComponentLinker
	var err error

	if opts.Profile != "" {
		cl, err = s.createLinkerWithFilterAndProfile(opts.TargetFilter, opts.Profile)
	} else {
		cl, err = s.createLinkerWithFilter(opts.TargetFilter)
	}

	if err != nil {
		return fmt.Errorf("failed to create component linker: %w", err)
	}
	return cl.UnlinkComponent(componentType, componentName, opts.TargetFilter)
}

// UnlinkAll unlinks all components from targets
func (s *Service) UnlinkAll(opts services.UnlinkOptions) error {
	// Validate flag combination
	if opts.AllProfiles && opts.Profile != "" {
		return fmt.Errorf("cannot use both --all-profiles and --profile flags together")
	}

	var cl *linker.ComponentLinker
	var err error

	if opts.Profile != "" {
		cl, err = s.createLinkerWithFilterAndProfile(opts.TargetFilter, opts.Profile)
	} else {
		cl, err = s.createLinkerWithFilter(opts.TargetFilter)
	}

	if err != nil {
		return fmt.Errorf("failed to create component linker: %w", err)
	}
	return cl.UnlinkAllComponents(opts.TargetFilter, opts.Force, opts.AllProfiles)
}

// UnlinkByType unlinks all components of a specific type from targets
func (s *Service) UnlinkByType(componentType string, opts services.UnlinkOptions) error {
	var cl *linker.ComponentLinker
	var err error

	if opts.Profile != "" {
		cl, err = s.createLinkerWithFilterAndProfile(opts.TargetFilter, opts.Profile)
	} else {
		cl, err = s.createLinkerWithFilter(opts.TargetFilter)
	}

	if err != nil {
		return fmt.Errorf("failed to create component linker: %w", err)
	}
	return cl.UnlinkComponentsByType(componentType, opts.TargetFilter, opts.Force)
}

// AutoLinkRepositories automatically detects and links local repositories
func (s *Service) AutoLinkRepositories() error {
	cl, err := s.createLinker()
	if err != nil {
		return fmt.Errorf("failed to create component linker: %w", err)
	}
	return cl.DetectAndLinkLocalRepositories()
}

// ListLinked lists all currently linked components
func (s *Service) ListLinked() error {
	cl, err := s.createLinker()
	if err != nil {
		return fmt.Errorf("failed to create component linker: %w", err)
	}
	return cl.ListLinkedComponents()
}

// ShowStatus shows the link status of components
func (s *Service) ShowStatus(opts services.LinkStatusOptions) error {
	// Validate flags
	if len(opts.ProfileFilter) > 0 && !opts.AllProfiles {
		return fmt.Errorf("--profile flag requires --all-profiles")
	}

	if opts.AllProfiles {
		// Check if any profiles exist
		profilesList, err := s.profileManager.ScanProfiles()
		if err != nil {
			s.formatter.EmptyLine()
			s.formatter.ErrorMsg("Failed to scan profiles")
			s.formatter.DetailItem("Error", err.Error())
			s.formatter.EmptyLine()
			s.formatter.InfoMsg("The --all-profiles flag requires at least one profile.")
			s.formatter.InfoMsg("Try running without --all-profiles, or create a profile first:")
			s.formatter.InfoMsg("  agent-smith profile create <name>")
			return err
		}

		if len(profilesList) == 0 {
			s.formatter.EmptyLine()
			s.formatter.InfoMsg("No profiles found")
			s.formatter.EmptyLine()
			s.formatter.InfoMsg("The --all-profiles flag requires at least one profile.")
			s.formatter.InfoMsg("Options:")
			s.formatter.InfoMsg("  1. Run without --all-profiles to show components from base installation")
			s.formatter.InfoMsg("  2. Create a profile first: agent-smith profile create <name>")
			return fmt.Errorf("no profiles found")
		}

		cl, err := s.createLinkerWithProfileManager()
		if err != nil {
			return fmt.Errorf("failed to create component linker: %w", err)
		}

		return cl.ShowAllProfilesLinkStatus(opts.ProfileFilter)
	}

	// Standard single-profile view
	cl, err := s.createLinker()
	if err != nil {
		return fmt.Errorf("failed to create component linker: %w", err)
	}
	return cl.ShowLinkStatus()
}
