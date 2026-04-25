package update

import (
	"fmt"

	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/internal/updater"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/profiles"
	"github.com/tjg184/agent-smith/pkg/services"
)

type Service struct {
	profileManager *profiles.ProfileManager
	logger         *logger.Logger
	formatter      *formatter.Formatter
}

func NewService(
	pm *profiles.ProfileManager,
	logger *logger.Logger,
	formatter *formatter.Formatter,
) services.UpdateService {
	return &Service{
		profileManager: pm,
		logger:         logger,
		formatter:      formatter,
	}
}

func (s *Service) UpdateComponent(componentType, componentName string, opts services.UpdateOptions) error {
	detector, err := s.createUpdateDetector(opts.Profile)
	if err != nil {
		return fmt.Errorf("failed to create update detector: %w", err)
	}

	metadata, err := detector.LoadMetadata(componentType, componentName)
	if err != nil {
		return fmt.Errorf("failed to load component metadata: %w", err)
	}

	if err := detector.UpdateComponent(componentType, componentName, metadata.SourceUrl); err != nil {
		return fmt.Errorf("failed to update component: %w", err)
	}

	return nil
}

func (s *Service) UpdateAll(opts services.UpdateOptions) error {
	resolved, err := s.resolveProfile(opts)
	if err != nil {
		return err
	}

	detector, err := s.createUpdateDetector(resolved.Profile)
	if err != nil {
		return fmt.Errorf("failed to create update detector: %w", err)
	}

	if err := detector.UpdateAll(resolved.RepoURL); err != nil {
		return fmt.Errorf("failed to update components: %w", err)
	}

	return nil
}

// resolveProfile finds the profile that owns the given repo URL when no explicit
// profile is set, mirroring how materialize resolves its profile.
func (s *Service) resolveProfile(opts services.UpdateOptions) (services.UpdateOptions, error) {
	if opts.RepoURL == "" || opts.Profile != "" {
		return opts, nil
	}
	profileName, err := s.profileManager.FindProfileBySourceURL(opts.RepoURL)
	if err != nil {
		return opts, fmt.Errorf("failed to look up profile for repo '%s': %w", opts.RepoURL, err)
	}
	if profileName == "" {
		return opts, fmt.Errorf("no installed profile found for repository '%s'", opts.RepoURL)
	}
	opts.Profile = profileName
	opts.RepoURL = ""
	return opts, nil
}

func (s *Service) CheckForUpdates(opts services.UpdateOptions) ([]services.UpdateInfo, error) {
	// This would require extending UpdateDetector to return available updates
	// For now, we'll return a placeholder implementation
	// TODO: Implement actual update checking logic
	s.logger.Debug("[DEBUG] CheckForUpdates called with profile: %s", opts.Profile)

	return []services.UpdateInfo{}, fmt.Errorf("CheckForUpdates not yet implemented")
}

func (s *Service) createUpdateDetector(profile string) (*updater.UpdateDetector, error) {
	if profile != "" {
		s.logger.Debug("[DEBUG] Creating UpdateDetector with profile: %s", profile)
		return updater.NewUpdateDetectorWithProfile(profile)
	}
	s.logger.Debug("[DEBUG] Creating UpdateDetector with default profile")
	return updater.NewUpdateDetector()
}
