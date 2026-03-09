package lock

import (
	"fmt"

	"github.com/tjg184/agent-smith/internal/metadata"
	"github.com/tjg184/agent-smith/internal/models"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/services"
)

// Service implements the ComponentLockService interface
type Service struct {
	logger *logger.Logger
}

// NewService creates a new ComponentLockService
func NewService(logger *logger.Logger) services.ComponentLockService {
	return &Service{
		logger: logger,
	}
}

// LoadEntry loads a component entry from the lock file
// Returns error if component exists in multiple sources (ambiguous)
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

// LoadEntryBySource loads a component entry from a specific source URL
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

// GetAllComponentNames returns all component names for a given type
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

// FindComponentSources returns all source URLs that contain the given component
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

// FindAllInstances returns all instances of a component across all sources
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

	// Convert ComponentSource to ComponentEntry pointers
	entries := make([]*models.ComponentEntry, len(sources))
	for i, src := range sources {
		entryCopy := src.Entry
		entries[i] = &entryCopy
	}

	return entries, nil
}

// SaveEntry saves a component entry to the lock file
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

	// Validate required entry fields
	if entry.SourceUrl == "" {
		return fmt.Errorf("entry.SourceUrl is required")
	}
	if entry.CommitHash == "" {
		return fmt.Errorf("entry.CommitHash is required")
	}

	// Convert ComponentEntry to SaveComponentEntry parameters
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

// RemoveEntry removes a component entry from the lock file
// Removes from all sources if component exists in multiple
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

// RemoveEntryBySource removes a component entry from a specific source
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

// ResolveFilesystemName resolves a filesystem name, adding suffixes if conflicts exist
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
	// sourceURL can be empty for non-source-aware resolution

	resolvedName, err := metadata.ResolveInstallFilesystemName(baseDir, componentType, desiredName, sourceURL)
	if err != nil {
		return "", fmt.Errorf("failed to resolve filesystem name: %w", err)
	}

	s.logger.Debug("[ComponentLockService] Resolved filesystem name: %s -> %s", desiredName, resolvedName)
	return resolvedName, nil
}

// HasConflict checks if a component name exists in multiple sources
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
