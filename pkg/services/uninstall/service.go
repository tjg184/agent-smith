package uninstall

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/internal/linker"
	"github.com/tjg184/agent-smith/internal/uninstaller"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/paths"
	"github.com/tjg184/agent-smith/pkg/services"
)

// Service implements the UninstallService interface
type Service struct {
	linker    *linker.ComponentLinker
	logger    *logger.Logger
	formatter *formatter.Formatter
}

// NewService creates a new UninstallService with dependencies injected
func NewService(
	linker *linker.ComponentLinker,
	logger *logger.Logger,
	formatter *formatter.Formatter,
) services.UninstallService {
	return &Service{
		linker:    linker,
		logger:    logger,
		formatter: formatter,
	}
}

// UninstallComponent uninstalls a single component
func (s *Service) UninstallComponent(componentType, componentName string, opts services.UninstallOptions) error {
	// Determine base directory
	baseDir, err := paths.GetAgentsDir()
	if err != nil {
		return fmt.Errorf("failed to get agents directory: %w", err)
	}

	if opts.Profile != "" {
		// Use profile directory
		profilesDir, err := paths.GetProfilesDir()
		if err != nil {
			return fmt.Errorf("failed to get profiles directory: %w", err)
		}
		baseDir = filepath.Join(profilesDir, opts.Profile)

		// Validate profile exists
		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			return fmt.Errorf("profile '%s' does not exist", opts.Profile)
		}
	}

	uninstallerService := uninstaller.NewUninstaller(baseDir, s.linker)

	if err := uninstallerService.UninstallComponent(componentType, componentName, opts.Source); err != nil {
		return fmt.Errorf("failed to uninstall component: %w", err)
	}

	return nil
}

// UninstallAllFromSource uninstalls all components from a specific repository source
func (s *Service) UninstallAllFromSource(repoURL string, opts services.UninstallOptions) error {
	// Get base directory (always ~/.agent-smith/ for bulk uninstall)
	baseDir, err := paths.GetAgentsDir()
	if err != nil {
		return fmt.Errorf("failed to get base directory: %w", err)
	}

	uninstallerService := uninstaller.NewUninstaller(baseDir, s.linker)

	// Uninstall all components from source
	if err := uninstallerService.UninstallAllFromSource(repoURL, opts.Force); err != nil {
		return fmt.Errorf("failed to uninstall components: %w", err)
	}

	return nil
}
