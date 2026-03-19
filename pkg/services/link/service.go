package link

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/tjg184/agent-smith/internal/detector"
	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/internal/linker"
	"github.com/tjg184/agent-smith/pkg/config"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/paths"
	"github.com/tjg184/agent-smith/pkg/profiles"
	"github.com/tjg184/agent-smith/pkg/services"
)

type Service struct {
	profileManager *profiles.ProfileManager
	logger         *logger.Logger
	formatter      *formatter.Formatter
}

func NewService(pm *profiles.ProfileManager, logger *logger.Logger, formatter *formatter.Formatter) services.LinkService {
	return &Service{
		profileManager: pm,
		logger:         logger,
		formatter:      formatter,
	}
}

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

type profileManagerAdapter struct {
	pm *profiles.ProfileManager
}

func (pma *profileManagerAdapter) ScanProfiles() ([]*linker.Profile, error) {
	profiles, err := pma.pm.ScanProfiles()
	if err != nil {
		return nil, err
	}

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

func (s *Service) createLinker() (*linker.ComponentLinker, error) {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}

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
		s.formatter.Info("Using active profile: %s", activeProfile)
	}

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

func (s *Service) createLinkerWithFilter(targetFilter string) (*linker.ComponentLinker, error) {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}

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

func (s *Service) createLinkerWithFilterAndProfile(targetFilter string, profile string) (*linker.ComponentLinker, error) {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}

	if profile != "" {
		profilesDir, err := paths.GetProfilesDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get profiles directory: %w", err)
		}

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

	allTargets, err := config.DetectAllTargets()
	if err != nil {
		return nil, fmt.Errorf("failed to detect targets: %w", err)
	}

	if targetFilter == string(config.TargetUniversal) {
		alreadyPresent := false
		for _, t := range allTargets {
			if t.GetName() == string(config.TargetUniversal) {
				alreadyPresent = true
				break
			}
		}
		if !alreadyPresent {
			universalTarget, err := config.NewUniversalTarget()
			if err != nil {
				return nil, fmt.Errorf("failed to create universal target: %w", err)
			}
			allTargets = append(allTargets, universalTarget)
		}
	}

	targets := s.filterTargets(allTargets, targetFilter)

	det := detector.NewRepositoryDetector()
	if s.logger != nil {
		det.SetLogger(s.logger)
	}

	return linker.NewComponentLinker(agentsDir, targets, det, nil)
}

func (s *Service) createLinkerWithProfileManager() (*linker.ComponentLinker, error) {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}

	targets, err := config.DetectAllTargets()
	if err != nil {
		return nil, fmt.Errorf("failed to detect targets: %w", err)
	}

	det := detector.NewRepositoryDetector()
	if s.logger != nil {
		det.SetLogger(s.logger)
	}

	adapter := &profileManagerAdapter{pm: s.profileManager}

	return linker.NewComponentLinker(agentsDir, targets, det, adapter)
}

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

func (s *Service) LinkComponent(componentType, componentName string, opts services.LinkOptions) error {
	s.showCurrentContext(opts.Profile)
	cl, err := s.createLinkerWithFilterAndProfile(opts.TargetFilter, opts.Profile)
	if err != nil {
		return fmt.Errorf("failed to create component linker: %w", err)
	}
	return cl.LinkComponent(componentType, componentName)
}

func (s *Service) LinkAll(opts services.LinkOptions) error {
	if opts.AllProfiles && opts.Profile != "" {
		return fmt.Errorf("cannot use both --all-profiles and --profile flags together")
	}

	if opts.AllProfiles {
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

		bold := color.New(color.Bold).SprintFunc()
		green := color.New(color.FgGreen, color.Bold).SprintFunc()
		cyan := color.New(color.FgCyan).SprintFunc()
		gray := color.New(color.FgHiBlack).SprintFunc()

		fmt.Printf("\n%s\n", bold("Linking components from all profiles..."))
		fmt.Println()

		for _, profileItem := range profilesList {
			fmt.Printf("%s\n", cyan(fmt.Sprintf("Profile: %s", profileItem.Name)))

			cl, err := s.createLinkerWithFilterAndProfile(opts.TargetFilter, profileItem.Name)
			if err != nil {
				return fmt.Errorf("failed to create component linker for profile '%s': %w", profileItem.Name, err)
			}

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

	s.showCurrentContext(opts.Profile)
	cl, err := s.createLinkerWithFilterAndProfile(opts.TargetFilter, opts.Profile)
	if err != nil {
		return fmt.Errorf("failed to create component linker: %w", err)
	}
	return cl.LinkAllComponents()
}

func (s *Service) LinkByType(componentType string, opts services.LinkOptions) error {
	s.showCurrentContext(opts.Profile)
	cl, err := s.createLinkerWithFilterAndProfile(opts.TargetFilter, opts.Profile)
	if err != nil {
		return fmt.Errorf("failed to create component linker: %w", err)
	}
	return cl.LinkComponentsByType(componentType)
}

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

func (s *Service) UnlinkAll(opts services.UnlinkOptions) error {
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

func (s *Service) AutoLinkRepositories() error {
	cl, err := s.createLinker()
	if err != nil {
		return fmt.Errorf("failed to create component linker: %w", err)
	}
	return cl.DetectAndLinkLocalRepositories()
}

func (s *Service) ListLinked() error {
	cl, err := s.createLinker()
	if err != nil {
		return fmt.Errorf("failed to create component linker: %w", err)
	}
	return cl.ListLinkedComponents()
}

func (s *Service) ShowStatus(opts services.LinkStatusOptions) error {
	if opts.AllProfiles {
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

		return cl.ShowAllProfilesLinkStatus(opts.ProfileFilter, opts.LinkedOnly)
	}

	var profileName string
	if len(opts.ProfileFilter) > 0 {
		profileName = opts.ProfileFilter[0]
	}

	cl, err := s.createLinkerWithFilterAndProfile("", profileName)
	if err != nil {
		return fmt.Errorf("failed to create component linker: %w", err)
	}
	return cl.ShowLinkStatus(opts.LinkedOnly)
}
