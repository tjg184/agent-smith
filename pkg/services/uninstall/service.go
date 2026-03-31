package uninstall

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tjg184/agent-smith/internal/detector"
	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/internal/linker"
	"github.com/tjg184/agent-smith/internal/models"
	"github.com/tjg184/agent-smith/internal/uninstaller"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/paths"
	"github.com/tjg184/agent-smith/pkg/profiles"
	"github.com/tjg184/agent-smith/pkg/profiles/profilemeta"
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

// lockFileIsEmpty returns true when the lock file at dir contains no component entries.
func lockFileIsEmpty(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, paths.ComponentLockFile))
	if err != nil {
		return true
	}
	var lf models.ComponentLockFile
	if err := json.Unmarshal(data, &lf); err != nil {
		return false
	}
	for _, byName := range lf.Skills {
		if len(byName) > 0 {
			return false
		}
	}
	for _, byName := range lf.Agents {
		if len(byName) > 0 {
			return false
		}
	}
	for _, byName := range lf.Commands {
		if len(byName) > 0 {
			return false
		}
	}
	return true
}

type Service struct {
	linker         *linker.ComponentLinker
	logger         *logger.Logger
	formatter      *formatter.Formatter
	profileManager *profiles.ProfileManager
}

func NewService(
	componentLinker *linker.ComponentLinker,
	logger *logger.Logger,
	formatter *formatter.Formatter,
	profileManager *profiles.ProfileManager,
) services.UninstallService {
	return &Service{
		linker:         componentLinker,
		logger:         logger,
		formatter:      formatter,
		profileManager: profileManager,
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
	profileDirs := collectProfileDirs(profilesDir)

	uninstallerService := uninstaller.NewUninstaller(baseDir, s.linker)

	if err := uninstallerService.UninstallAllFromSourceAcrossDirs(repoURL, profileDirs, opts.Force); err != nil {
		return fmt.Errorf("failed to uninstall components: %w", err)
	}

	s.cleanupEmptyRepoProfiles(profileDirs, repoURL)

	return nil
}

// cleanupEmptyRepoProfiles deletes profile directories whose source URL matches repoURL
// and whose lock file is now empty. Warnings are printed but never fatal — the uninstall
// already succeeded; cleanup is best-effort.
func (s *Service) cleanupEmptyRepoProfiles(profileDirs []string, repoURL string) {
	if s.profileManager == nil {
		return
	}

	rd := detector.NewRepositoryDetector()
	normalizedURL, err := rd.NormalizeURL(repoURL)
	if err != nil {
		normalizedURL = repoURL
	}

	for _, dir := range profileDirs {
		meta, err := profilemeta.Load(dir)
		if err != nil || meta == nil {
			continue
		}
		if meta.Type != "repo" || meta.SourceURL != normalizedURL {
			continue
		}
		if !lockFileIsEmpty(dir) {
			continue
		}
		profileName := filepath.Base(dir)
		if err := s.profileManager.DeleteProfile(profileName); err != nil {
			fmt.Printf("Warning: could not remove empty profile '%s': %v\n", profileName, err)
		}
	}
}
