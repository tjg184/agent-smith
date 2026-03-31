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

func collectProfileDirs(profilesDir string) []string {
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return nil
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, filepath.Join(profilesDir, e.Name()))
		}
	}
	return dirs
}

type Service struct {
	linker    *linker.ComponentLinker
	logger    *logger.Logger
	formatter *formatter.Formatter
}

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

func (s *Service) UninstallComponent(componentType, componentName string, opts services.UninstallOptions) error {
	baseDir, err := paths.GetAgentsDir()
	if err != nil {
		return fmt.Errorf("failed to get agents directory: %w", err)
	}

	if opts.Profile != "" {
		profilesDir, err := paths.GetProfilesDir()
		if err != nil {
			return fmt.Errorf("failed to get profiles directory: %w", err)
		}
		baseDir = filepath.Join(profilesDir, opts.Profile)

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

func (s *Service) UninstallAllFromSource(repoURL string, opts services.UninstallOptions) error {
	baseDir, err := paths.GetAgentsDir()
	if err != nil {
		return fmt.Errorf("failed to get base directory: %w", err)
	}

	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		return fmt.Errorf("failed to get profiles directory: %w", err)
	}
	extraDirs := collectProfileDirs(profilesDir)

	uninstallerService := uninstaller.NewUninstaller(baseDir, s.linker)

	if err := uninstallerService.UninstallAllFromSourceAcrossDirs(repoURL, extraDirs, opts.Force); err != nil {
		return fmt.Errorf("failed to uninstall components: %w", err)
	}

	return nil
}
