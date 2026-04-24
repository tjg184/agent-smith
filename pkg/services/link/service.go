package link

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/internal/linker"
	"github.com/tjg184/agent-smith/pkg/config"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/paths"
	"github.com/tjg184/agent-smith/pkg/profiles"
	"github.com/tjg184/agent-smith/pkg/profiles/profilemeta"
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

	profileName := explicitProfile
	if profileName == "" {
		activeProfile, err := s.profileManager.GetActiveProfile()
		if err != nil {
			fmt.Printf("  Repo:    %s\n", gray("unknown (error checking)"))
			s.formatter.EmptyLine()
			return
		}
		profileName = activeProfile
	}

	if profileName == "" {
		fmt.Printf("  Repo:    %s\n", gray("none"))
	} else {
		repoURL := s.sourceURLForProfile(profileName)
		if repoURL != "" {
			fmt.Printf("  Repo:    %s\n", cyan(repoURL))
		} else {
			fmt.Printf("  Profile: %s\n", green(profileName))
		}
	}

	s.formatter.EmptyLine()
}

// sourceURLForProfile returns the source repo URL for the profile, or "" if none (user-created profiles).
func (s *Service) sourceURLForProfile(profileName string) string {
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		return ""
	}
	meta, err := profilemeta.Load(filepath.Join(profilesDir, profileName))
	if err != nil || meta == nil {
		return ""
	}
	return meta.SourceURL
}

func (s *Service) createLinker() (*linker.ComponentLinker, error) {
	activeProfile, err := s.profileManager.GetActiveProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to get active profile: %w", err)
	}
	if activeProfile != "" {
		s.formatter.Info("Using active profile: %s", activeProfile)
	}
	return linker.Build(linker.BuildOptions{ActiveProfile: activeProfile}, s.logger)
}

func (s *Service) createLinkerWithFilter(targetFilter string) (*linker.ComponentLinker, error) {
	activeProfile, err := s.profileManager.GetActiveProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to get active profile: %w", err)
	}
	return linker.Build(linker.BuildOptions{ActiveProfile: activeProfile, TargetFilter: targetFilter}, s.logger)
}

func (s *Service) createLinkerWithFilterAndProfile(targetFilter string, profile string) (*linker.ComponentLinker, error) {
	opts := linker.BuildOptions{TargetFilter: targetFilter}

	if profile != "" {
		if err := s.validateProfileExists(profile); err != nil {
			return nil, err
		}
		s.formatter.Info("Using profile: %s", profile)
		opts.ExplicitProfile = profile
	} else {
		activeProfile, err := s.profileManager.GetActiveProfile()
		if err != nil {
			return nil, fmt.Errorf("failed to get active profile: %w", err)
		}
		opts.ActiveProfile = activeProfile
	}

	// The universal target is not auto-detected; inject it when explicitly requested.
	if targetFilter == string(config.TargetUniversal) {
		allTargets, err := config.DetectAllTargets()
		if err != nil {
			return nil, fmt.Errorf("failed to detect targets: %w", err)
		}
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
		opts.Targets = allTargets
	}

	return linker.Build(opts, s.logger)
}

func (s *Service) createLinkerWithProfileManager() (*linker.ComponentLinker, error) {
	return linker.Build(linker.BuildOptions{ProfileManager: profiles.NewLinkerAdapter(s.profileManager)}, s.logger)
}

// validateProfileExists returns a descriptive error if the named profile directory
// does not exist, listing available profiles when possible.
func (s *Service) validateProfileExists(profile string) error {
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		return fmt.Errorf("failed to get profiles directory: %w", err)
	}
	profilePath := filepath.Join(profilesDir, profile)
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		availableProfiles, scanErr := s.profileManager.ScanProfiles()
		if scanErr == nil && len(availableProfiles) > 0 {
			profileNames := make([]string, len(availableProfiles))
			for i, p := range availableProfiles {
				profileNames[i] = p.Name
			}
			return fmt.Errorf("profile '%s' does not exist\n\nAvailable profiles:\n  - %s\n\nTo create this profile:\n  agent-smith profile create %s",
				profile, strings.Join(profileNames, "\n  - "), profile)
		}
		return fmt.Errorf("profile '%s' does not exist\n\nTo create this profile:\n  agent-smith profile create %s\n\nTo list available profiles:\n  agent-smith profile list", profile, profile)
	}
	return nil
}

// resolveProfileForRepo finds the profile name for a given repo URL.
// Returns an error if no matching profile is found.
func (s *Service) resolveProfileForRepo(repoURL string) (string, error) {
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		return "", fmt.Errorf("failed to get profiles directory: %w", err)
	}

	profileName, err := profilemeta.FindBySourceURL(profilesDir, repoURL)
	if err != nil {
		return "", fmt.Errorf("failed to search profiles: %w", err)
	}

	if profileName == "" {
		availableProfiles, scanErr := s.profileManager.ScanProfiles()
		if scanErr == nil && len(availableProfiles) > 0 {
			profileNames := make([]string, len(availableProfiles))
			for i, p := range availableProfiles {
				profileNames[i] = p.Name
			}
			return "", fmt.Errorf("no installed components found for repo '%s'\n\nAvailable profiles:\n  - %s", repoURL, strings.Join(profileNames, "\n  - "))
		}
		return "", fmt.Errorf("no installed components found for repo '%s'", repoURL)
	}

	return profileName, nil
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

	if opts.RepoURL != "" {
		profileName, err := s.resolveProfileForRepo(opts.RepoURL)
		if err != nil {
			return err
		}
		s.showCurrentContext(profileName)
		cl, err := s.createLinkerWithFilterAndProfile(opts.TargetFilter, profileName)
		if err != nil {
			return fmt.Errorf("failed to create component linker: %w", err)
		}
		return cl.LinkAllComponents()
	}

	if opts.AllProfiles {
		profilesList, err := s.profileManager.ScanProfiles()
		if err != nil {
			s.formatter.EmptyLine()
			s.formatter.ErrorMsg("Failed to scan profiles")
			s.formatter.DetailItem("Error", err.Error())
			s.formatter.EmptyLine()
			return err
		}

		if len(profilesList) == 0 {
			s.formatter.EmptyLine()
			s.formatter.InfoMsg("No installed repos found. Run 'agent-smith install all <repo>' to get started.")
			s.formatter.EmptyLine()
			return fmt.Errorf("no profiles found")
		}

		bold := color.New(color.Bold).SprintFunc()
		green := color.New(color.FgGreen, color.Bold).SprintFunc()
		cyan := color.New(color.FgCyan).SprintFunc()
		gray := color.New(color.FgHiBlack).SprintFunc()

		fmt.Printf("\n%s\n", bold("Linking components from all repos..."))
		fmt.Println()

		for _, profileItem := range profilesList {
			label := s.sourceURLForProfile(profileItem.Name)
			if label == "" {
				label = profileItem.Name
			}
			fmt.Printf("%s\n", cyan(label))

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

		fmt.Printf("\n%s\n", green("✓ Successfully linked components from all repos"))
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

	if opts.RepoURL != "" {
		profileName, err := s.resolveProfileForRepo(opts.RepoURL)
		if err != nil {
			return err
		}
		cl, err := s.createLinkerWithFilterAndProfile(opts.TargetFilter, profileName)
		if err != nil {
			return fmt.Errorf("failed to create component linker: %w", err)
		}
		return cl.UnlinkAllComponents(opts.TargetFilter, opts.Force, false)
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
	// When a profile filter is provided, scope to just that profile.
	if len(opts.ProfileFilter) > 0 {
		cl, err := s.createLinkerWithFilterAndProfile("", opts.ProfileFilter[0])
		if err != nil {
			return fmt.Errorf("failed to create component linker: %w", err)
		}
		return cl.ShowLinkStatus(opts.LinkedOnly)
	}

	if opts.AllProfiles {
		profilesList, err := s.profileManager.ScanProfiles()
		if err != nil {
			s.formatter.EmptyLine()
			s.formatter.ErrorMsg("Failed to scan profiles")
			s.formatter.DetailItem("Error", err.Error())
			s.formatter.EmptyLine()
			return err
		}

		if len(profilesList) == 0 {
			s.formatter.EmptyLine()
			s.formatter.InfoMsg("No installed repos found. Run 'agent-smith install all <repo>' to get started.")
			s.formatter.EmptyLine()
			return nil
		}

		cl, err := s.createLinkerWithProfileManager()
		if err != nil {
			return fmt.Errorf("failed to create component linker: %w", err)
		}

		return cl.ShowAllProfilesLinkStatus(opts.ProfileFilter, opts.LinkedOnly)
	}

	cl, err := s.createLinkerWithFilterAndProfile("", "")
	if err != nil {
		return fmt.Errorf("failed to create component linker: %w", err)
	}
	return cl.ShowLinkStatus(opts.LinkedOnly)
}
