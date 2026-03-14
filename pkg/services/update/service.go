package update

import (
	"fmt"

	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/internal/updater"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/services"
)

// Service implements the UpdateService interface
type Service struct {
	logger    *logger.Logger
	formatter *formatter.Formatter
}

func NewService(
	logger *logger.Logger,
	formatter *formatter.Formatter,
) services.UpdateService {
	return &Service{
		logger:    logger,
		formatter: formatter,
	}
}

func (s *Service) UpdateComponent(componentType, componentName string, opts services.UpdateOptions) error {
	detector := s.createUpdateDetector(opts.Profile)

	// Load metadata to get source URL
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
	detector := s.createUpdateDetector(opts.Profile)

	if err := detector.UpdateAll(); err != nil {
		return fmt.Errorf("failed to update components: %w", err)
	}

	return nil
}

func (s *Service) CheckForUpdates(opts services.UpdateOptions) ([]services.UpdateInfo, error) {
	// This would require extending UpdateDetector to return available updates
	// For now, we'll return a placeholder implementation
	// TODO: Implement actual update checking logic
	s.logger.Debug("[DEBUG] CheckForUpdates called with profile: %s", opts.Profile)

	return []services.UpdateInfo{}, fmt.Errorf("CheckForUpdates not yet implemented")
}

// createUpdateDetector creates an UpdateDetector with the appropriate profile
func (s *Service) createUpdateDetector(profile string) *updater.UpdateDetector {
	if profile != "" {
		s.logger.Debug("[DEBUG] Creating UpdateDetector with profile: %s", profile)
		return updater.NewUpdateDetectorWithProfile(profile)
	}
	s.logger.Debug("[DEBUG] Creating UpdateDetector with default profile")
	return updater.NewUpdateDetector()
}
