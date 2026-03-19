package materialize

import (
	"fmt"
	"path/filepath"

	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/paths"
	"github.com/tjg184/agent-smith/pkg/profiles"
	"github.com/tjg184/agent-smith/pkg/services"
	matstatus "github.com/tjg184/agent-smith/pkg/services/materialize/status"
	matsync "github.com/tjg184/agent-smith/pkg/services/materialize/sync"
)

type Service struct {
	profileManager        *profiles.ProfileManager
	logger                *logger.Logger
	formatter             *formatter.Formatter
	postprocessorRegistry *PostprocessorRegistry
}

func NewService(pm *profiles.ProfileManager, logger *logger.Logger, formatter *formatter.Formatter) services.MaterializeService {
	return &Service{
		profileManager:        pm,
		logger:                logger,
		formatter:             formatter,
		postprocessorRegistry: NewPostprocessorRegistry(),
	}
}

// registryAdapter bridges between the parent's PostprocessorRegistry (which uses the parent's
// PostprocessContext) and the sync package's PostprocessorRegistry interface (which uses
// sync.PostprocessContext). Both context types are structurally identical.
type registryAdapter struct {
	registry *PostprocessorRegistry
}

func (a *registryAdapter) RunPostprocessors(ctx matsync.PostprocessContext) error {
	return a.registry.RunPostprocessors(PostprocessContext(ctx))
}

func (a *registryAdapter) RunCleanup(ctx matsync.PostprocessContext) error {
	return a.registry.RunCleanup(PostprocessContext(ctx))
}

func (s *Service) syncDeps() matsync.Deps {
	return matsync.Deps{
		Logger:       s.logger,
		Formatter:    s.formatter,
		Registry:     &registryAdapter{registry: s.postprocessorRegistry},
		GetSourceDir: s.getSourceDir,
	}
}

func (s *Service) statusDeps() matstatus.Deps {
	return matstatus.Deps{
		Logger:    s.logger,
		Formatter: s.formatter,
	}
}

func (s *Service) getSourceDir(profile string) (string, string, error) {
	baseDir, err := paths.GetAgentsDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to get agent-smith directory: %w", err)
	}

	if profile != "" {
		if profile == "base" {
			return baseDir, "", nil
		}

		profilesList, err := s.profileManager.ScanProfiles()
		if err != nil {
			return "", "", fmt.Errorf("failed to scan profiles: %w", err)
		}

		profileExists := false
		for _, p := range profilesList {
			if p.Name == profile {
				profileExists = true
				break
			}
		}

		if !profileExists {
			return "", "", fmt.Errorf("profile '%s' not found", profile)
		}

		profilesDir, err := paths.GetProfilesDir()
		if err != nil {
			return "", "", fmt.Errorf("failed to get profiles directory: %w", err)
		}
		return filepath.Join(profilesDir, profile), profile, nil
	}

	activeProfile, err := s.profileManager.GetActiveProfile()
	if err != nil {
		return "", "", fmt.Errorf("failed to check active profile: %w", err)
	}

	if activeProfile != "" {
		profilesDir, err := paths.GetProfilesDir()
		if err != nil {
			return "", "", fmt.Errorf("failed to get profiles directory: %w", err)
		}
		return filepath.Join(profilesDir, activeProfile), activeProfile, nil
	}

	return baseDir, "", nil
}

func (s *Service) MaterializeComponent(componentType, componentName string, opts services.MaterializeOptions) error {
	return matsync.MaterializeComponent(s.syncDeps(), componentType, componentName, opts)
}

func (s *Service) MaterializeAll(opts services.MaterializeOptions) error {
	return matsync.MaterializeAll(s.syncDeps(), opts)
}

func (s *Service) MaterializeByType(componentType string, opts services.MaterializeOptions) error {
	return matsync.MaterializeByType(s.syncDeps(), componentType, opts)
}

func (s *Service) ListMaterialized(opts services.ListMaterializedOptions) error {
	return matstatus.ListMaterialized(s.statusDeps(), opts)
}

func (s *Service) ShowComponentInfo(componentType, componentName string, opts services.MaterializeInfoOptions) error {
	return matstatus.ShowComponentInfo(s.statusDeps(), componentType, componentName, opts)
}

func (s *Service) ShowStatus(opts services.MaterializeStatusOptions) error {
	return matstatus.ShowStatus(s.statusDeps(), opts)
}

func (s *Service) UpdateMaterialized(opts services.MaterializeUpdateOptions) error {
	return matsync.UpdateMaterialized(s.syncDeps(), opts)
}
