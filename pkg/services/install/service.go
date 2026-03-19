package install

import (
	"fmt"
	"os"

	"github.com/tjg184/agent-smith/internal/downloader"
	"github.com/tjg184/agent-smith/internal/fileutil"
	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/internal/models"
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

func (s *Service) InstallSkill(repoURL, name string, opts services.InstallOptions) error {
	s.logger.Debug("[DEBUG] InstallSkill called with repoURL=%s, name=%s, profile=%s, installDir=%s", repoURL, name, opts.Profile, opts.InstallDir)

	if err := s.validateInstallOptions(opts); err != nil {
		return err
	}

	if opts.InstallDir != "" {
		return s.installToTargetDir(models.ComponentSkill, repoURL, name, opts.InstallDir)
	} else if opts.Profile != "" {
		return s.installToProfile(models.ComponentSkill, repoURL, name, opts.Profile)
	}

	profile, err := s.getOrCreateRepoProfile(repoURL)
	if err != nil {
		return fmt.Errorf("failed to determine profile for repository: %w", err)
	}

	if profile != "" {
		return s.installToProfile(models.ComponentSkill, repoURL, name, profile)
	}

	return s.installToBase(models.ComponentSkill, repoURL, name)
}

func (s *Service) InstallAgent(repoURL, name string, opts services.InstallOptions) error {
	s.logger.Debug("[DEBUG] InstallAgent called with repoURL=%s, name=%s, profile=%s, installDir=%s", repoURL, name, opts.Profile, opts.InstallDir)

	if err := s.validateInstallOptions(opts); err != nil {
		return err
	}

	if opts.InstallDir != "" {
		return s.installToTargetDir(models.ComponentAgent, repoURL, name, opts.InstallDir)
	} else if opts.Profile != "" {
		return s.installToProfile(models.ComponentAgent, repoURL, name, opts.Profile)
	}

	profile, err := s.getOrCreateRepoProfile(repoURL)
	if err != nil {
		return fmt.Errorf("failed to determine profile for repository: %w", err)
	}

	if profile != "" {
		return s.installToProfile(models.ComponentAgent, repoURL, name, profile)
	}

	return s.installToBase(models.ComponentAgent, repoURL, name)
}

func (s *Service) InstallCommand(repoURL, name string, opts services.InstallOptions) error {
	s.logger.Debug("[DEBUG] InstallCommand called with repoURL=%s, name=%s, profile=%s, installDir=%s", repoURL, name, opts.Profile, opts.InstallDir)

	if err := s.validateInstallOptions(opts); err != nil {
		return err
	}

	if opts.InstallDir != "" {
		return s.installToTargetDir(models.ComponentCommand, repoURL, name, opts.InstallDir)
	} else if opts.Profile != "" {
		return s.installToProfile(models.ComponentCommand, repoURL, name, opts.Profile)
	}

	profile, err := s.getOrCreateRepoProfile(repoURL)
	if err != nil {
		return fmt.Errorf("failed to determine profile for repository: %w", err)
	}

	if profile != "" {
		return s.installToProfile(models.ComponentCommand, repoURL, name, profile)
	}

	return s.installToBase(models.ComponentCommand, repoURL, name)
}

func (s *Service) InstallBulk(repoURL string, opts services.InstallOptions) error {
	s.logger.Debug("[DEBUG] InstallBulk called with repoURL=%s, profile=%s, installDir=%s", repoURL, opts.Profile, opts.InstallDir)

	if err := s.validateInstallOptions(opts); err != nil {
		return err
	}

	if opts.InstallDir != "" {
		return s.installBulkToTargetDir(repoURL, opts.InstallDir)
	}

	return s.installBulkToProfile(repoURL, opts.Profile)
}

func (s *Service) validateInstallOptions(opts services.InstallOptions) error {
	if opts.Profile != "" && opts.InstallDir != "" {
		return fmt.Errorf("cannot specify both --profile and --install-dir flags")
	}
	return nil
}

func (s *Service) installToTargetDir(ct models.ComponentType, repoURL, name, targetDir string) error {
	resolvedPath, err := paths.ResolveTargetDir(targetDir)
	if err != nil {
		return fmt.Errorf("failed to resolve target directory: %w", err)
	}

	if err := fileutil.CreateDirectoryWithPermissions(resolvedPath); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	s.logger.Info("Installing to custom directory: %s", resolvedPath)
	dl, err := downloader.ForTypeWithTargetDir(ct, resolvedPath)
	if err != nil {
		return fmt.Errorf("failed to create downloader: %w", err)
	}
	if err := dl.Download(repoURL, name); err != nil {
		return fmt.Errorf("failed to download %s: %w", ct, err)
	}
	return nil
}

func (s *Service) installToProfile(ct models.ComponentType, repoURL, name, profile string) error {
	if err := s.validateProfileExists(profile); err != nil {
		return err
	}
	dl, err := downloader.ForTypeWithProfile(ct, profile)
	if err != nil {
		return fmt.Errorf("failed to create downloader: %w", err)
	}
	if err := dl.Download(repoURL, name); err != nil {
		return fmt.Errorf("failed to download %s: %w", ct, err)
	}
	return s.activateProfileWithFeedback(profile)
}

func (s *Service) installToBase(ct models.ComponentType, repoURL, name string) error {
	dl, err := downloader.ForType(ct)
	if err != nil {
		return fmt.Errorf("failed to create downloader: %w", err)
	}
	if err := dl.Download(repoURL, name); err != nil {
		return fmt.Errorf("failed to download %s: %w", ct, err)
	}
	return nil
}

func (s *Service) installBulkToTargetDir(repoURL, targetDir string) error {
	resolvedPath, err := paths.ResolveTargetDir(targetDir)
	if err != nil {
		return fmt.Errorf("failed to resolve target directory: %w", err)
	}

	if err := fileutil.CreateDirectoryWithPermissions(resolvedPath); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	s.logger.Info("Installing to custom directory: %s", resolvedPath)
	bulkDownloader, err := downloader.NewBulkDownloaderWithTargetDir(resolvedPath)
	if err != nil {
		return fmt.Errorf("failed to create downloader: %w", err)
	}

	if err := bulkDownloader.AddAll(repoURL); err != nil {
		return fmt.Errorf("failed to bulk download components: %w", err)
	}

	return nil
}

func (s *Service) installBulkToProfile(repoURL, profile string) error {
	s.logger.Info("Validating repository: %s", repoURL)
	validationDownloader, err := downloader.NewBulkDownloader()
	if err != nil {
		return fmt.Errorf("failed to create downloader: %w", err)
	}
	tempDir, components, err := validationDownloader.ValidateRepo(repoURL)
	if err != nil {
		return fmt.Errorf("repository validation failed: %w", err)
	}

	var profileName string
	var isNewProfile bool

	if profile != "" {
		profilesList, err := s.profileManager.ScanProfiles()
		if err != nil {
			os.RemoveAll(tempDir)
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
			os.RemoveAll(tempDir)
			return fmt.Errorf("profile '%s' already exists. Please choose a different name or remove the --profile flag to update the existing profile", profile)
		}

		profileName = profile
		isNewProfile = true
	} else {
		existingProfileName, err := s.profileManager.FindProfileBySourceURL(repoURL)
		if err != nil {
			os.RemoveAll(tempDir)
			return fmt.Errorf("failed to search for existing profile: %w", err)
		}

		if existingProfileName != "" {
			profileName = existingProfileName
			isNewProfile = false
			s.logger.Info("Found existing profile for repository: %s", profileName)
			s.logger.Info("Updating profile with latest components...")
		} else {
			existingProfiles, err := s.profileManager.ScanProfiles()
			if err != nil {
				os.RemoveAll(tempDir)
				return fmt.Errorf("failed to scan profiles: %w", err)
			}

			existingProfileNames := make([]string, len(existingProfiles))
			for i, p := range existingProfiles {
				existingProfileNames[i] = p.Name
			}

			profileName = profiles.GenerateProfileNameFromRepo(repoURL, existingProfileNames)
			isNewProfile = true
		}
	}

	if isNewProfile {
		s.logger.Info("Creating profile: %s", profileName)
		if err := s.profileManager.CreateProfileWithMetadata(profileName, repoURL); err != nil {
			os.RemoveAll(tempDir)
			return fmt.Errorf("failed to create profile: %w", err)
		}
	}

	s.logger.Info("Installing components to profile: %s", profileName)
	bulkDownloader, err := downloader.NewBulkDownloaderForProfile(profileName)
	if err != nil {
		os.RemoveAll(tempDir)
		return fmt.Errorf("failed to create downloader: %w", err)
	}

	if err := bulkDownloader.AddAllFromTemp(repoURL, components, tempDir); err != nil {
		if isNewProfile {
			s.logger.Debug("[DEBUG] Installation failed, cleaning up newly created profile: %s", profileName)
			if cleanupErr := s.profileManager.DeleteProfile(profileName); cleanupErr != nil {
				s.logger.Warn("Failed to clean up profile after installation failure: %v", cleanupErr)
			}
		}
		return fmt.Errorf("failed to bulk download components: %w", err)
	}

	s.logger.Debug("[DEBUG] Auto-activating profile after install all: %s", profileName)
	return s.activateProfileWithFeedback(profileName)
}

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

func (s *Service) activateProfileWithFeedback(profile string) error {
	result, err := s.profileManager.ActivateProfileWithResult(profile)
	if err != nil {
		s.logger.Warn("Profile created but activation failed: %v", err)
		s.formatter.EmptyLine()
		s.formatter.Info("To manually activate this profile, run:")
		s.formatter.Info("  agent-smith profile activate %s", profile)
		return nil
	}

	s.formatter.EmptyLine()
	if result.Switched {
		s.formatter.SuccessMsg("Switched profile: %s → %s", result.PreviousProfile, result.NewProfile)
	} else if result.PreviousProfile == result.NewProfile {
		s.formatter.SuccessMsg("Profile '%s' is active and ready", result.NewProfile)
	} else {
		s.formatter.SuccessMsg("Profile activated: %s", result.NewProfile)
	}

	s.formatter.EmptyLine()
	s.formatter.Info("Next: Run 'agent-smith link all' to apply changes to your editor(s)")

	return nil
}

func (s *Service) getOrCreateRepoProfile(repoURL string) (string, error) {
	existingProfileName, err := s.profileManager.FindProfileBySourceURL(repoURL)
	if err != nil {
		return "", fmt.Errorf("failed to search for existing profile: %w", err)
	}

	if existingProfileName != "" {
		s.logger.Debug("[DEBUG] Found existing profile for repository: %s", existingProfileName)
		return existingProfileName, nil
	}

	existingProfiles, err := s.profileManager.ScanProfiles()
	if err != nil {
		return "", fmt.Errorf("failed to scan profiles: %w", err)
	}

	existingProfileNames := make([]string, len(existingProfiles))
	for i, p := range existingProfiles {
		existingProfileNames[i] = p.Name
	}

	profileName := profiles.GenerateProfileNameFromRepo(repoURL, existingProfileNames)

	s.logger.Info("Creating profile: %s", profileName)
	if err := s.profileManager.CreateProfileWithMetadata(profileName, repoURL); err != nil {
		return "", fmt.Errorf("failed to create profile: %w", err)
	}

	return profileName, nil
}
