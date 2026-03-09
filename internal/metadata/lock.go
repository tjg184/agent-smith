package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tjg184/agent-smith/internal/models"
	"github.com/tjg184/agent-smith/pkg/paths"
)

// ComponentEntryOptions holds optional parameters for SaveComponentEntry
type ComponentEntryOptions struct {
	// For install operations
	InstalledAt string
	UpdatedAt   string
	Components  int
	Detection   string

	// For materialize operations
	MaterializedAt string
	SourceProfile  string

	// For both
	SourceHash     string
	CurrentHash    string
	FilesystemName string
}

// SaveComponentEntry saves a unified component entry for both installs and materializations
// Version 5 unified format
func SaveComponentEntry(baseDir, componentType, componentName, source, sourceType, sourceUrl, commitHash, originalPath string, opts ComponentEntryOptions) error {
	lockFilePath := paths.GetComponentLockPath(baseDir, componentType)

	lockFile, err := loadOrCreateLockFile(lockFilePath)
	if err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339)

	targetMap, err := getTargetMap(&lockFile, componentType)
	if err != nil {
		return err
	}

	if targetMap[sourceUrl] == nil {
		targetMap[sourceUrl] = make(map[string]models.ComponentEntry)
	}

	existingEntry, exists := targetMap[sourceUrl][componentName]

	// Preserve installedAt if this is an update to an existing install
	installedAt := opts.InstalledAt
	if installedAt == "" && exists && existingEntry.InstalledAt != "" {
		installedAt = existingEntry.InstalledAt
	} else if installedAt == "" && opts.UpdatedAt != "" {
		// This is a new install operation
		installedAt = now
	}

	entry := models.ComponentEntry{
		Source:       source,
		SourceType:   sourceType,
		SourceUrl:    sourceUrl,
		OriginalPath: originalPath,
		CommitHash:   commitHash,
		Version:      5,

		// Timestamps
		InstalledAt:    installedAt,
		MaterializedAt: opts.MaterializedAt,
		UpdatedAt:      opts.UpdatedAt,

		// Drift detection
		SourceHash:  opts.SourceHash,
		CurrentHash: opts.CurrentHash,

		// Location/tracking
		FilesystemName: opts.FilesystemName,
		SourceProfile:  opts.SourceProfile,

		// Install-specific
		Components: opts.Components,
		Detection:  opts.Detection,
	}

	targetMap[sourceUrl][componentName] = entry

	return writeLockFile(lockFilePath, lockFile)
}

// SaveLockFileEntry is the legacy function for backward compatibility
// Calls SaveComponentEntry with install-specific parameters
func SaveLockFileEntry(baseDir, componentType, componentName, source, sourceType, sourceUrl, commitHash string, components int, detection, originalPath string) error {
	return SaveComponentEntry(baseDir, componentType, componentName, source, sourceType, sourceUrl, commitHash, originalPath, ComponentEntryOptions{
		UpdatedAt:  time.Now().Format(time.RFC3339),
		Components: components,
		Detection:  detection,
	})
}

// LoadFromLockFile loads metadata from lock file
func LoadFromLockFile(baseDir, componentType, componentName string) (*models.ComponentMetadata, error) {
	entry, err := LoadLockFileEntry(baseDir, componentType, componentName)
	if err != nil {
		return nil, err
	}

	return &models.ComponentMetadata{
		Name:         componentName,
		Source:       entry.SourceUrl,
		Commit:       entry.CommitHash,
		OriginalPath: entry.OriginalPath,
		Components:   entry.Components,
		Detection:    entry.Detection,
	}, nil
}

// LoadLockFileEntry loads a component lock entry from the lock file
// Searches across all sources and returns the first match
// Returns error if not found or multiple sources have the same component
func LoadLockFileEntry(baseDir, componentType, componentName string) (*models.ComponentEntry, error) {
	lockFilePath := paths.GetComponentLockPath(baseDir, componentType)

	lockData, err := os.ReadFile(lockFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	var lockFile models.ComponentLockFile
	if err := json.Unmarshal(lockData, &lockFile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lock file: %w", err)
	}

	targetMap, err := getTargetMap(&lockFile, componentType)
	if err != nil {
		return nil, err
	}

	var foundEntry *models.ComponentEntry
	var foundSources []string

	for sourceUrl, components := range targetMap {
		if entry, exists := components[componentName]; exists {
			foundSources = append(foundSources, sourceUrl)
			if foundEntry == nil {
				entryCopy := entry
				foundEntry = &entryCopy
			}
		}
	}

	if foundEntry == nil {
		return nil, fmt.Errorf("component %s not found in lock file", componentName)
	}

	if len(foundSources) > 1 {
		return nil, fmt.Errorf("component %s found in multiple sources: %v. Use LoadLockFileEntryBySource or specify --source flag", componentName, foundSources)
	}

	return foundEntry, nil
}

// LoadLockFileEntryBySource loads a component lock entry from a specific source
func LoadLockFileEntryBySource(baseDir, componentType, componentName, sourceUrl string) (*models.ComponentEntry, error) {
	lockFilePath := paths.GetComponentLockPath(baseDir, componentType)

	lockData, err := os.ReadFile(lockFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	var lockFile models.ComponentLockFile
	if err := json.Unmarshal(lockData, &lockFile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lock file: %w", err)
	}

	// Get the appropriate nested map for this component type
	targetMap, err := getTargetMap(&lockFile, componentType)
	if err != nil {
		return nil, err
	}

	components, sourceExists := targetMap[sourceUrl]
	if !sourceExists {
		return nil, fmt.Errorf("source %s not found in lock file", sourceUrl)
	}

	entry, exists := components[componentName]
	if !exists {
		return nil, fmt.Errorf("component %s not found in source %s", componentName, sourceUrl)
	}

	return &entry, nil
}

// RemoveComponentEntry removes component from lock file across all sources
func RemoveComponentEntry(baseDir, componentType, componentName string) error {
	lockFilePath := paths.GetComponentLockPath(baseDir, componentType)

	lockFile, err := loadOrCreateLockFile(lockFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Lock file doesn't exist, nothing to remove
			return nil
		}
		return err
	}

	// Get the appropriate nested map for this component type
	targetMap, err := getTargetMap(&lockFile, componentType)
	if err != nil {
		return err
	}

	// Remove from all sources
	found := false
	for sourceUrl, components := range targetMap {
		if _, exists := components[componentName]; exists {
			delete(components, componentName)
			found = true

			// If source map is now empty, remove it too
			if len(components) == 0 {
				delete(targetMap, sourceUrl)
			}
		}
	}

	if !found {
		// Entry doesn't exist, nothing to remove
		return nil
	}

	return writeLockFile(lockFilePath, lockFile)
}

// RemoveComponentEntryBySource removes component from lock file for specific source only
func RemoveComponentEntryBySource(baseDir, componentType, componentName, sourceUrl string) error {
	lockFilePath := paths.GetComponentLockPath(baseDir, componentType)

	lockFile, err := loadOrCreateLockFile(lockFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Lock file doesn't exist, nothing to remove
			return nil
		}
		return err
	}

	// Get the appropriate nested map for this component type
	targetMap, err := getTargetMap(&lockFile, componentType)
	if err != nil {
		return err
	}

	components, sourceExists := targetMap[sourceUrl]
	if !sourceExists {
		// Source doesn't exist, nothing to remove
		return nil
	}

	// Remove the component from this source
	delete(components, componentName)

	// If source map is now empty, remove it too
	if len(components) == 0 {
		delete(targetMap, sourceUrl)
	}

	return writeLockFile(lockFilePath, lockFile)
}

// GetAllComponentNames returns all component names from the lock file for a given type
// Aggregates names across all sources
func GetAllComponentNames(baseDir, componentType string) ([]string, error) {
	lockFilePath := paths.GetComponentLockPath(baseDir, componentType)

	lockData, err := os.ReadFile(lockFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Lock file doesn't exist, return empty list
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	var lockFile models.ComponentLockFile
	if err := json.Unmarshal(lockData, &lockFile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lock file: %w", err)
	}

	// Get the appropriate nested map for this component type
	targetMap, err := getTargetMap(&lockFile, componentType)
	if err != nil {
		return nil, err
	}

	// Extract component names from all sources (may have duplicates)
	nameSet := make(map[string]bool)
	for _, components := range targetMap {
		for name := range components {
			nameSet[name] = true
		}
	}

	// Convert set to slice
	names := make([]string, 0, len(nameSet))
	for name := range nameSet {
		names = append(names, name)
	}

	return names, nil
}

// FindComponentSources returns all source URLs that have the given component
func FindComponentSources(baseDir, componentType, componentName string) ([]string, error) {
	lockFilePath := paths.GetComponentLockPath(baseDir, componentType)

	lockData, err := os.ReadFile(lockFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Lock file doesn't exist, return empty list
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	var lockFile models.ComponentLockFile
	if err := json.Unmarshal(lockData, &lockFile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lock file: %w", err)
	}

	// Get the appropriate nested map for this component type
	targetMap, err := getTargetMap(&lockFile, componentType)
	if err != nil {
		return nil, err
	}

	// Find all sources that have this component
	var sources []string
	for sourceUrl, components := range targetMap {
		if _, exists := components[componentName]; exists {
			sources = append(sources, sourceUrl)
		}
	}

	return sources, nil
}

// ComponentSource represents a component's source information
type ComponentSource struct {
	SourceUrl string
	Entry     models.ComponentEntry
}

// FindAllComponentInstances returns all instances of a component across all sources
func FindAllComponentInstances(baseDir, componentType, componentName string) ([]ComponentSource, error) {
	lockFilePath := paths.GetComponentLockPath(baseDir, componentType)

	lockData, err := os.ReadFile(lockFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Lock file doesn't exist, return empty list
			return []ComponentSource{}, nil
		}
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	var lockFile models.ComponentLockFile
	if err := json.Unmarshal(lockData, &lockFile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lock file: %w", err)
	}

	// Get the appropriate nested map for this component type
	targetMap, err := getTargetMap(&lockFile, componentType)
	if err != nil {
		return nil, err
	}

	var instances []ComponentSource
	for sourceUrl, components := range targetMap {
		if entry, exists := components[componentName]; exists {
			instances = append(instances, ComponentSource{
				SourceUrl: sourceUrl,
				Entry:     entry,
			})
		}
	}

	return instances, nil
}

// Helper functions

func loadOrCreateLockFile(lockFilePath string) (models.ComponentLockFile, error) {
	lockData, err := os.ReadFile(lockFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return models.ComponentLockFile{
				Version:  5,
				Skills:   make(map[string]map[string]models.ComponentEntry),
				Agents:   make(map[string]map[string]models.ComponentEntry),
				Commands: make(map[string]map[string]models.ComponentEntry),
			}, nil
		}
		return models.ComponentLockFile{}, fmt.Errorf("failed to read lock file: %w", err)
	}

	var lockFile models.ComponentLockFile
	if err := json.Unmarshal(lockData, &lockFile); err != nil {
		// If lock file is corrupted, create new one
		return models.ComponentLockFile{
			Version:  5,
			Skills:   make(map[string]map[string]models.ComponentEntry),
			Agents:   make(map[string]map[string]models.ComponentEntry),
			Commands: make(map[string]map[string]models.ComponentEntry),
		}, nil
	}

	// Ensure version is current and maps are initialized
	lockFile.Version = 5
	if lockFile.Skills == nil {
		lockFile.Skills = make(map[string]map[string]models.ComponentEntry)
	}
	if lockFile.Agents == nil {
		lockFile.Agents = make(map[string]map[string]models.ComponentEntry)
	}
	if lockFile.Commands == nil {
		lockFile.Commands = make(map[string]map[string]models.ComponentEntry)
	}

	return lockFile, nil
}

func writeLockFile(lockFilePath string, lockFile models.ComponentLockFile) error {
	jsonData, err := json.MarshalIndent(lockFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}

	return os.WriteFile(lockFilePath, jsonData, 0644)
}

func getTargetMap(lockFile *models.ComponentLockFile, componentType string) (map[string]map[string]models.ComponentEntry, error) {
	switch componentType {
	case "skills":
		return lockFile.Skills, nil
	case "agents":
		return lockFile.Agents, nil
	case "commands":
		return lockFile.Commands, nil
	default:
		return nil, fmt.Errorf("unknown component type: %s", componentType)
	}
}

// ResolveInstallFilesystemName determines the actual filesystem name to use for a component during install
// If the exact component (sourceUrl + componentName) is already installed, returns its existing filesystem name
// Otherwise, if componentName already exists, returns componentName-2, componentName-3, etc.
func ResolveInstallFilesystemName(baseDir, componentType, componentName, sourceUrl string) (string, error) {
	// Load existing lock file
	lockFilePath := paths.GetComponentLockPath(baseDir, componentType)
	lockFile, err := loadOrCreateLockFile(lockFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to load lock file: %w", err)
	}

	// Get the component map for this type
	componentMap, err := getTargetMap(&lockFile, componentType)
	if err != nil {
		return "", err
	}

	// First, check if this exact component (sourceUrl + componentName) is already installed
	// If so, reuse its existing filesystem name for idempotency
	if sourceUrl != "" {
		if sourceComponents, exists := componentMap[sourceUrl]; exists {
			if entry, exists := sourceComponents[componentName]; exists && entry.FilesystemName != "" {
				return entry.FilesystemName, nil
			}
		}
	}

	// Get the base directory for this component type
	var componentBaseDir string
	switch componentType {
	case "skills":
		componentBaseDir, err = paths.GetSkillsDir()
	case "agents":
		componentBaseDir, err = paths.GetAgentsDir()
	case "commands":
		componentBaseDir, err = paths.GetCommandsDir()
	default:
		return "", fmt.Errorf("unknown component type: %s", componentType)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get component directory: %w", err)
	}

	// Check both filesystem and metadata for conflicts
	baseComponentPath := filepath.Join(componentBaseDir, componentName)

	// If the base name doesn't exist on disk or in metadata, use it
	if !fileExists(baseComponentPath) && !installFilesystemNameExists(componentName, componentMap) {
		return componentName, nil
	}

	suffix := 2
	for {
		candidateName := fmt.Sprintf("%s-%d", componentName, suffix)
		candidatePath := filepath.Join(componentBaseDir, candidateName)

		if !fileExists(candidatePath) && !installFilesystemNameExists(candidateName, componentMap) {
			return candidateName, nil
		}

		suffix++

		// Safety check to prevent infinite loops
		if suffix > 1000 {
			return fmt.Sprintf("%s-%d", componentName, suffix), nil
		}
	}
}

// fileExists checks if a path exists on disk
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// installFilesystemNameExists checks if a filesystem name is already used in the component map
func installFilesystemNameExists(filesystemName string, componentMap map[string]map[string]models.ComponentEntry) bool {
	for _, sourceComponents := range componentMap {
		for _, entry := range sourceComponents {
			if entry.FilesystemName == filesystemName {
				return true
			}
		}
	}
	return false
}
