package lock

import (
	"fmt"

	"github.com/tjg184/agent-smith/internal/metadata"
	"github.com/tjg184/agent-smith/internal/models"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/services"
)

type Service struct {
	logger *logger.Logger
}

func NewService(logger *logger.Logger) services.ComponentLockService {
	return &Service{
		logger: logger,
	}
}

func (s *Service) LoadEntry(baseDir, componentType, componentName string) (*models.ComponentEntry, error) {
	s.logger.Debug("[ComponentLockService] LoadEntry: baseDir=%s, type=%s, name=%s", baseDir, componentType, componentName)

	if baseDir == "" {
		return nil, fmt.Errorf("baseDir is required")
	}
	if componentType == "" {
		return nil, fmt.Errorf("componentType is required")
	}
	if componentName == "" {
		return nil, fmt.Errorf("componentName is required")
	}

	entry, err := metadata.LoadLockFileEntry(baseDir, componentType, componentName)
	if err != nil {
		return nil, fmt.Errorf("failed to load entry: %w", err)
	}

	return entry, nil
}

func (s *Service) LoadEntryBySource(baseDir, componentType, componentName, sourceURL string) (*models.ComponentEntry, error) {
	s.logger.Debug("[ComponentLockService] LoadEntryBySource: baseDir=%s, type=%s, name=%s, source=%s",
		baseDir, componentType, componentName, sourceURL)

	if baseDir == "" {
		return nil, fmt.Errorf("baseDir is required")
	}
	if componentType == "" {
		return nil, fmt.Errorf("componentType is required")
	}
	if componentName == "" {
		return nil, fmt.Errorf("componentName is required")
	}
	if sourceURL == "" {
		return nil, fmt.Errorf("sourceURL is required")
	}

	entry, err := metadata.LoadLockFileEntryBySource(baseDir, componentType, componentName, sourceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to load entry by source: %w", err)
	}

	return entry, nil
}

func (s *Service) GetAllComponentNames(baseDir, componentType string) ([]string, error) {
	s.logger.Debug("[ComponentLockService] GetAllComponentNames: baseDir=%s, type=%s", baseDir, componentType)

	if baseDir == "" {
		return nil, fmt.Errorf("baseDir is required")
	}
	if componentType == "" {
		return nil, fmt.Errorf("componentType is required")
	}

	names, err := metadata.GetAllComponentNames(baseDir, componentType)
	if err != nil {
		return nil, fmt.Errorf("failed to get component names: %w", err)
	}

	return names, nil
}

func (s *Service) FindComponentSources(baseDir, componentType, componentName string) ([]string, error) {
	s.logger.Debug("[ComponentLockService] FindComponentSources: baseDir=%s, type=%s, name=%s",
		baseDir, componentType, componentName)

	if baseDir == "" {
		return nil, fmt.Errorf("baseDir is required")
	}
	if componentType == "" {
		return nil, fmt.Errorf("componentType is required")
	}
	if componentName == "" {
		return nil, fmt.Errorf("componentName is required")
	}

	sources, err := metadata.FindComponentSources(baseDir, componentType, componentName)
	if err != nil {
		return nil, fmt.Errorf("failed to find component sources: %w", err)
	}

	return sources, nil
}

func (s *Service) FindAllInstances(baseDir, componentType, componentName string) ([]*models.ComponentEntry, error) {
	s.logger.Debug("[ComponentLockService] FindAllInstances: baseDir=%s, type=%s, name=%s",
		baseDir, componentType, componentName)

	if baseDir == "" {
		return nil, fmt.Errorf("baseDir is required")
	}
	if componentType == "" {
		return nil, fmt.Errorf("componentType is required")
	}
	if componentName == "" {
		return nil, fmt.Errorf("componentName is required")
	}

	sources, err := metadata.FindAllComponentInstances(baseDir, componentType, componentName)
	if err != nil {
		return nil, fmt.Errorf("failed to find all instances: %w", err)
	}

	entries := make([]*models.ComponentEntry, len(sources))
	for i, src := range sources {
		entryCopy := src.Entry
		entries[i] = &entryCopy
	}

	return entries, nil
}

func (s *Service) SaveEntry(baseDir, componentType, componentName string, entry *models.ComponentEntry) error {
	s.logger.Debug("[ComponentLockService] SaveEntry: baseDir=%s, type=%s, name=%s",
		baseDir, componentType, componentName)

	if baseDir == "" {
		return fmt.Errorf("baseDir is required")
	}
	if componentType == "" {
		return fmt.Errorf("componentType is required")
	}
	if componentName == "" {
		return fmt.Errorf("componentName is required")
	}
	if entry == nil {
		return fmt.Errorf("entry is required")
	}

	if entry.SourceUrl == "" {
		return fmt.Errorf("entry.SourceUrl is required")
	}
	if entry.CommitHash == "" {
		return fmt.Errorf("entry.CommitHash is required")
	}

	opts := metadata.ComponentEntryOptions{
		InstalledAt:    entry.InstalledAt,
		MaterializedAt: entry.MaterializedAt,
		UpdatedAt:      entry.UpdatedAt,
		SourceHash:     entry.SourceHash,
		CurrentHash:    entry.CurrentHash,
		FilesystemName: entry.FilesystemName,
		SourceProfile:  entry.SourceProfile,
		Components:     entry.Components,
		Detection:      entry.Detection,
	}

	err := metadata.SaveComponentEntry(
		baseDir,
		componentType,
		componentName,
		entry.Source,
		entry.SourceType,
		entry.SourceUrl,
		entry.CommitHash,
		entry.OriginalPath,
		opts,
	)
	if err != nil {
		return fmt.Errorf("failed to save entry: %w", err)
	}

	s.logger.Debug("[ComponentLockService] Successfully saved entry for %s/%s", componentType, componentName)
	return nil
}

func (s *Service) RemoveEntry(baseDir, componentType, componentName string) error {
	s.logger.Debug("[ComponentLockService] RemoveEntry: baseDir=%s, type=%s, name=%s",
		baseDir, componentType, componentName)

	if baseDir == "" {
		return fmt.Errorf("baseDir is required")
	}
	if componentType == "" {
		return fmt.Errorf("componentType is required")
	}
	if componentName == "" {
		return fmt.Errorf("componentName is required")
	}

	err := metadata.RemoveComponentEntry(baseDir, componentType, componentName)
	if err != nil {
		return fmt.Errorf("failed to remove entry: %w", err)
	}

	s.logger.Debug("[ComponentLockService] Successfully removed entry for %s/%s", componentType, componentName)
	return nil
}

func (s *Service) RemoveEntryBySource(baseDir, componentType, componentName, sourceURL string) error {
	s.logger.Debug("[ComponentLockService] RemoveEntryBySource: baseDir=%s, type=%s, name=%s, source=%s",
		baseDir, componentType, componentName, sourceURL)

	if baseDir == "" {
		return fmt.Errorf("baseDir is required")
	}
	if componentType == "" {
		return fmt.Errorf("componentType is required")
	}
	if componentName == "" {
		return fmt.Errorf("componentName is required")
	}
	if sourceURL == "" {
		return fmt.Errorf("sourceURL is required")
	}

	err := metadata.RemoveComponentEntryBySource(baseDir, componentType, componentName, sourceURL)
	if err != nil {
		return fmt.Errorf("failed to remove entry by source: %w", err)
	}

	s.logger.Debug("[ComponentLockService] Successfully removed entry for %s/%s from source %s",
		componentType, componentName, sourceURL)
	return nil
}

func (s *Service) ResolveFilesystemName(baseDir, componentType, desiredName, sourceURL string) (string, error) {
	s.logger.Debug("[ComponentLockService] ResolveFilesystemName: baseDir=%s, type=%s, desiredName=%s, sourceURL=%s",
		baseDir, componentType, desiredName, sourceURL)

	if baseDir == "" {
		return "", fmt.Errorf("baseDir is required")
	}
	if componentType == "" {
		return "", fmt.Errorf("componentType is required")
	}
	if desiredName == "" {
		return "", fmt.Errorf("desiredName is required")
	}

	resolvedName, err := metadata.ResolveInstallFilesystemName(baseDir, componentType, desiredName, sourceURL)
	if err != nil {
		return "", fmt.Errorf("failed to resolve filesystem name: %w", err)
	}

	s.logger.Debug("[ComponentLockService] Resolved filesystem name: %s -> %s", desiredName, resolvedName)
	return resolvedName, nil
}

func (s *Service) HasConflict(baseDir, componentType, componentName string) (bool, error) {
	s.logger.Debug("[ComponentLockService] HasConflict: baseDir=%s, type=%s, name=%s",
		baseDir, componentType, componentName)

	if baseDir == "" {
		return false, fmt.Errorf("baseDir is required")
	}
	if componentType == "" {
		return false, fmt.Errorf("componentType is required")
	}
	if componentName == "" {
		return false, fmt.Errorf("componentName is required")
	}

	sources, err := metadata.FindComponentSources(baseDir, componentType, componentName)
	if err != nil {
		return false, fmt.Errorf("failed to check for conflicts: %w", err)
	}

	hasConflict := len(sources) > 1
	s.logger.Debug("[ComponentLockService] Conflict check result: %v (sources: %d)", hasConflict, len(sources))
	return hasConflict, nil
}
