package install

import (
	"fmt"

	"github.com/tgaines/agent-smith/internal/downloader"
	"github.com/tgaines/agent-smith/internal/fileutil"
	"github.com/tgaines/agent-smith/internal/formatter"
	"github.com/tgaines/agent-smith/pkg/logger"
	"github.com/tgaines/agent-smith/pkg/paths"
	"github.com/tgaines/agent-smith/pkg/profiles"
	"github.com/tgaines/agent-smith/pkg/services"
)

// Service implements the InstallService interface
type Service struct {
	profileManager *profiles.ProfileManager
	logger         *logger.Logger
	formatter      *formatter.Formatter
}

// NewService creates a new InstallService with dependencies injected
func NewService(
	profileManager *profiles.ProfileManager,
	logger *logger.Logger,
	formatter *formatter.Formatter,
) services.InstallService {
	return &Service{
		profileManager: profileManager,
		logger:         logger,
		formatter:      formatter,
	}
}

// InstallSkill installs a skill component from a repository
func (s *Service) InstallSkill(repoURL, name string, opts services.InstallOptions) error {
	s.logger.Debug("[DEBUG] InstallSkill called with repoURL=%s, name=%s, profile=%s, targetDir=%s", repoURL, name, opts.Profile, opts.TargetDir)

	// Validate options
	if err := s.validateInstallOptions(opts); err != nil {
		return err
	}

	// Handle different installation modes
	if opts.TargetDir != "" {
		return s.installSkillToTargetDir(repoURL, name, opts.TargetDir)
	} else if opts.Profile != "" {
		return s.installSkillToProfile(repoURL, name, opts.Profile)
	}

	// Standard installation to ~/.agent-smith/
	return s.installSkillToBase(repoURL, name)
}

// InstallAgent installs an agent component from a repository
func (s *Service) InstallAgent(repoURL, name string, opts services.InstallOptions) error {
	s.logger.Debug("[DEBUG] InstallAgent called with repoURL=%s, name=%s, profile=%s, targetDir=%s", repoURL, name, opts.Profile, opts.TargetDir)

	// Validate options
	if err := s.validateInstallOptions(opts); err != nil {
		return err
	}

	// Handle different installation modes
	if opts.TargetDir != "" {
		return s.installAgentToTargetDir(repoURL, name, opts.TargetDir)
	} else if opts.Profile != "" {
		return s.installAgentToProfile(repoURL, name, opts.Profile)
	}

	// Standard installation to ~/.agent-smith/
	return s.installAgentToBase(repoURL, name)
}

// InstallCommand installs a command component from a repository
func (s *Service) InstallCommand(repoURL, name string, opts services.InstallOptions) error {
	s.logger.Debug("[DEBUG] InstallCommand called with repoURL=%s, name=%s, profile=%s, targetDir=%s", repoURL, name, opts.Profile, opts.TargetDir)

	// Validate options
	if err := s.validateInstallOptions(opts); err != nil {
		return err
	}

	// Handle different installation modes
	if opts.TargetDir != "" {
		return s.installCommandToTargetDir(repoURL, name, opts.TargetDir)
	} else if opts.Profile != "" {
		return s.installCommandToProfile(repoURL, name, opts.Profile)
	}

	// Standard installation to ~/.agent-smith/
	return s.installCommandToBase(repoURL, name)
}

// InstallBulk installs all components from a repository
func (s *Service) InstallBulk(repoURL string, opts services.InstallOptions) error {
	s.logger.Debug("[DEBUG] InstallBulk called with repoURL=%s, profile=%s, targetDir=%s", repoURL, opts.Profile, opts.TargetDir)

	// Validate options
	if err := s.validateInstallOptions(opts); err != nil {
		return err
	}

	// Handle different installation modes
	if opts.TargetDir != "" {
		return s.installBulkToTargetDir(repoURL, opts.TargetDir)
	}

	// Profile-based bulk installation (with auto-creation and reuse)
	return s.installBulkToProfile(repoURL, opts.Profile)
}

// validateInstallOptions validates that conflicting options aren't specified
func (s *Service) validateInstallOptions(opts services.InstallOptions) error {
	if opts.Profile != "" && opts.TargetDir != "" {
		return fmt.Errorf("cannot specify both --profile and --target-dir flags")
	}
	return nil
}

// installSkillToTargetDir installs a skill to a custom target directory
func (s *Service) installSkillToTargetDir(repoURL, name, targetDir string) error {
	s.logger.Debug("[DEBUG] Installing skill to custom target directory")
	resolvedPath, err := paths.ResolveTargetDir(targetDir)
	if err != nil {
		return fmt.Errorf("failed to resolve target directory: %w", err)
	}
	s.logger.Debug("[DEBUG] Resolved target directory: %s", resolvedPath)

	if err := fileutil.CreateDirectoryWithPermissions(resolvedPath); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	s.logger.Info("Installing to custom directory: %s", resolvedPath)
	dl := downloader.NewSkillDownloaderWithTargetDir(resolvedPath)
	if err := dl.DownloadSkill(repoURL, name); err != nil {
		return fmt.Errorf("failed to download skill: %w", err)
	}
	s.logger.Debug("[DEBUG] Skill download completed successfully")

	return nil
}

// installSkillToProfile installs a skill to a profile
func (s *Service) installSkillToProfile(repoURL, name, profile string) error {
	s.logger.Debug("[DEBUG] Installing skill to profile: %s", profile)

	// Validate profile exists
	if err := s.validateProfileExists(profile); err != nil {
		return err
	}

	dl := downloader.NewSkillDownloaderForProfile(profile)
	if err := dl.DownloadSkill(repoURL, name); err != nil {
		return fmt.Errorf("failed to download skill: %w", err)
	}
	s.logger.Debug("[DEBUG] Skill download to profile completed successfully")

	// Auto-activate profile if no profile is currently active
	return s.maybeAutoActivateProfile(profile)
}

// installSkillToBase installs a skill to ~/.agent-smith/
func (s *Service) installSkillToBase(repoURL, name string) error {
	s.logger.Debug("[DEBUG] Installing skill to standard directory (~/.agent-smith/)")
	dl := downloader.NewSkillDownloader()
	if err := dl.DownloadSkill(repoURL, name); err != nil {
		return fmt.Errorf("failed to download skill: %w", err)
	}
	s.logger.Debug("[DEBUG] Skill download completed successfully")

	return nil
}

// installAgentToTargetDir installs an agent to a custom target directory
func (s *Service) installAgentToTargetDir(repoURL, name, targetDir string) error {
	resolvedPath, err := paths.ResolveTargetDir(targetDir)
	if err != nil {
		return fmt.Errorf("failed to resolve target directory: %w", err)
	}

	if err := fileutil.CreateDirectoryWithPermissions(resolvedPath); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	s.logger.Info("Installing to custom directory: %s", resolvedPath)
	dl := downloader.NewAgentDownloaderWithTargetDir(resolvedPath)
	if err := dl.DownloadAgent(repoURL, name); err != nil {
		return fmt.Errorf("failed to download agent: %w", err)
	}

	return nil
}

// installAgentToProfile installs an agent to a profile
func (s *Service) installAgentToProfile(repoURL, name, profile string) error {
	// Validate profile exists
	if err := s.validateProfileExists(profile); err != nil {
		return err
	}

	dl := downloader.NewAgentDownloaderForProfile(profile)
	if err := dl.DownloadAgent(repoURL, name); err != nil {
		return fmt.Errorf("failed to download agent: %w", err)
	}

	// Auto-activate profile if no profile is currently active
	return s.maybeAutoActivateProfile(profile)
}

// installAgentToBase installs an agent to ~/.agent-smith/
func (s *Service) installAgentToBase(repoURL, name string) error {
	dl := downloader.NewAgentDownloader()
	if err := dl.DownloadAgent(repoURL, name); err != nil {
		return fmt.Errorf("failed to download agent: %w", err)
	}

	return nil
}

// installCommandToTargetDir installs a command to a custom target directory
func (s *Service) installCommandToTargetDir(repoURL, name, targetDir string) error {
	resolvedPath, err := paths.ResolveTargetDir(targetDir)
	if err != nil {
		return fmt.Errorf("failed to resolve target directory: %w", err)
	}

	if err := fileutil.CreateDirectoryWithPermissions(resolvedPath); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	s.logger.Info("Installing to custom directory: %s", resolvedPath)
	dl := downloader.NewCommandDownloaderWithTargetDir(resolvedPath)
	if err := dl.DownloadCommand(repoURL, name); err != nil {
		return fmt.Errorf("failed to download command: %w", err)
	}

	return nil
}

// installCommandToProfile installs a command to a profile
func (s *Service) installCommandToProfile(repoURL, name, profile string) error {
	// Validate profile exists
	if err := s.validateProfileExists(profile); err != nil {
		return err
	}

	dl := downloader.NewCommandDownloaderForProfile(profile)
	if err := dl.DownloadCommand(repoURL, name); err != nil {
		return fmt.Errorf("failed to download command: %w", err)
	}

	// Auto-activate profile if no profile is currently active
	return s.maybeAutoActivateProfile(profile)
}

// installCommandToBase installs a command to ~/.agent-smith/
func (s *Service) installCommandToBase(repoURL, name string) error {
	dl := downloader.NewCommandDownloader()
	if err := dl.DownloadCommand(repoURL, name); err != nil {
		return fmt.Errorf("failed to download command: %w", err)
	}

	return nil
}

// installBulkToTargetDir installs all components to a custom target directory
func (s *Service) installBulkToTargetDir(repoURL, targetDir string) error {
	// Resolve the target directory path
	resolvedPath, err := paths.ResolveTargetDir(targetDir)
	if err != nil {
		return fmt.Errorf("failed to resolve target directory: %w", err)
	}

	// Create the target directory if it doesn't exist
	if err := fileutil.CreateDirectoryWithPermissions(resolvedPath); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	s.logger.Info("Installing to custom directory: %s", resolvedPath)
	bulkDownloader := downloader.NewBulkDownloaderWithTargetDir(resolvedPath)

	if err := bulkDownloader.AddAll(repoURL); err != nil {
		return fmt.Errorf("failed to bulk download components: %w", err)
	}

	return nil
}

// installBulkToProfile installs all components to a profile (with auto-creation and reuse)
func (s *Service) installBulkToProfile(repoURL, profile string) error {
	var profileName string

	if profile != "" {
		// Custom profile name provided via --profile flag
		// Check if profile with this name already exists
		profilesList, err := s.profileManager.ScanProfiles()
		if err != nil {
			return fmt.Errorf("failed to scan profiles: %w", err)
		}

		profileExists := false
		for _, p := range profilesList {
			if p.Name == profile {
				profileExists = true
				break
			}
		}

		if profileExists {
			return fmt.Errorf("profile '%s' already exists. Please choose a different name or remove the --profile flag to update the existing profile", profile)
		}

		profileName = profile
		s.logger.Info("Creating profile: %s", profileName)

		// Create the profile with metadata
		if err := s.profileManager.CreateProfileWithMetadata(profileName, repoURL); err != nil {
			return fmt.Errorf("failed to create profile: %w", err)
		}
	} else {
		// No custom profile name - use auto-detection and reuse logic
		// Check if a profile already exists for this repository
		existingProfileName, err := s.profileManager.FindProfileBySourceURL(repoURL)
		if err != nil {
			return fmt.Errorf("failed to search for existing profile: %w", err)
		}

		if existingProfileName != "" {
			// Profile already exists, reuse it
			profileName = existingProfileName
			s.logger.Info("Found existing profile for repository: %s", profileName)
			s.logger.Info("Updating profile with latest components...")
		} else {
			// Get existing profiles for name generation
			existingProfiles, err := s.profileManager.ScanProfiles()
			if err != nil {
				return fmt.Errorf("failed to scan profiles: %w", err)
			}

			existingProfileNames := make([]string, len(existingProfiles))
			for i, p := range existingProfiles {
				existingProfileNames[i] = p.Name
			}

			// Generate a unique profile name
			profileName = profiles.GenerateProfileNameFromRepo(repoURL, existingProfileNames)
			s.logger.Info("Creating profile: %s", profileName)

			// Create the profile with metadata
			if err := s.profileManager.CreateProfileWithMetadata(profileName, repoURL); err != nil {
				return fmt.Errorf("failed to create profile: %w", err)
			}
		}
	}

	// Install components to the profile
	s.logger.Info("Installing components to profile: %s", profileName)
	bulkDownloader := downloader.NewBulkDownloaderForProfile(profileName)

	if err := bulkDownloader.AddAll(repoURL); err != nil {
		return fmt.Errorf("failed to bulk download components: %w", err)
	}

	// Auto-activate profile after successful installation
	s.logger.Debug("[DEBUG] Auto-activating profile after install all: %s", profileName)
	result, err := s.profileManager.ActivateProfileWithResult(profileName)
	if err != nil {
		// Don't fail the installation if activation fails, just warn
		s.logger.Warn("⚠ Profile created but activation failed: %v", err)
		s.formatter.EmptyLine()
		s.formatter.Info("To manually activate this profile, run:")
		s.formatter.Info("  agent-smith profile activate %s", profileName)
		return nil
	}

	// Display activation result with clear messaging
	s.formatter.EmptyLine()
	if result.Switched {
		s.formatter.SuccessMsg("Switched profile: %s → %s", result.PreviousProfile, result.NewProfile)
	} else if result.PreviousProfile == result.NewProfile {
		// Profile was already active - just confirm it's ready
		s.formatter.SuccessMsg("Profile '%s' is active and ready", result.NewProfile)
	} else {
		// First activation
		s.formatter.SuccessMsg("Profile activated: %s", result.NewProfile)
	}

	// Display next step hint
	s.formatter.EmptyLine()
	s.formatter.Info("Next: Run 'agent-smith link all' to apply changes to your editor(s)")

	return nil
}

// validateProfileExists validates that a profile exists
func (s *Service) validateProfileExists(profile string) error {
	profilesList, err := s.profileManager.ScanProfiles()
	if err != nil {
		return fmt.Errorf("failed to scan profiles: %w", err)
	}
	s.logger.Debug("[DEBUG] Found %d profiles", len(profilesList))

	profileExists := false
	for _, p := range profilesList {
		if p.Name == profile {
			profileExists = true
			break
		}
	}

	if !profileExists {
		return fmt.Errorf("profile '%s' not found", profile)
	}

	return nil
}

// maybeAutoActivateProfile auto-activates a profile if no profile is currently active
func (s *Service) maybeAutoActivateProfile(profile string) error {
	activeProfile, err := s.profileManager.GetActiveProfile()
	if err != nil {
		return fmt.Errorf("failed to get active profile: %w", err)
	}

	if activeProfile == "" {
		s.logger.Debug("[DEBUG] No active profile detected, auto-activating profile: %s", profile)
		if err := s.profileManager.ActivateProfile(profile); err != nil {
			return fmt.Errorf("failed to auto-activate profile: %w", err)
		}
		s.logger.Info("Profile '%s' has been automatically activated as your first profile.", profile)
		s.logger.Info("Components from this profile are now ready to be linked.")
	}

	return nil
}
