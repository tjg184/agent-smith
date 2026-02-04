package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/tgaines/agent-smith/internal/models"
	"github.com/tgaines/agent-smith/pkg/paths"
)

// ComponentLockFile represents the lock file structure for component metadata
// Version 4+ uses nested structure: map[sourceURL]map[componentName]ComponentLockEntry
type ComponentLockFile struct {
	Version  int                                      `json:"version"`
	Skills   map[string]map[string]ComponentLockEntry `json:"skills"`
	Agents   map[string]map[string]ComponentLockEntry `json:"agents,omitempty"`
	Commands map[string]map[string]ComponentLockEntry `json:"commands,omitempty"`
}

// ComponentLockEntry represents a single component entry in the lock file
type ComponentLockEntry struct {
	Source       string `json:"source"`
	SourceType   string `json:"sourceType"`
	SourceUrl    string `json:"sourceUrl"`
	OriginalPath string `json:"originalPath,omitempty"`
	CommitHash   string `json:"commitHash,omitempty"`
	InstalledAt  string `json:"installedAt"`
	UpdatedAt    string `json:"updatedAt"`
	Version      int    `json:"version"`
	Components   int    `json:"components,omitempty"`
	Detection    string `json:"detection,omitempty"`
}

// SaveLockFileEntry saves a component lock entry in agent-smith install compatible format
// Version 4 uses nested structure by source URL
func SaveLockFileEntry(baseDir, componentType, componentName, source, sourceType, sourceUrl, commitHash string, components int, detection, originalPath string) error {
	lockFilePath := paths.GetComponentLockPath(baseDir, componentType)

	// Read existing lock file or create new one
	var lockFile ComponentLockFile
	lockData, err := os.ReadFile(lockFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			lockFile = ComponentLockFile{
				Version:  4, // Version 4 uses nested structure
				Skills:   make(map[string]map[string]ComponentLockEntry),
				Agents:   make(map[string]map[string]ComponentLockEntry),
				Commands: make(map[string]map[string]ComponentLockEntry),
			}
		} else {
			return fmt.Errorf("failed to read lock file: %w", err)
		}
	} else {
		if err := json.Unmarshal(lockData, &lockFile); err != nil {
			// If lock file is corrupted, create new one
			lockFile = ComponentLockFile{
				Version:  4,
				Skills:   make(map[string]map[string]ComponentLockEntry),
				Agents:   make(map[string]map[string]ComponentLockEntry),
				Commands: make(map[string]map[string]ComponentLockEntry),
			}
		}
		// Ensure version is current and maps are initialized
		lockFile.Version = 4
		if lockFile.Skills == nil {
			lockFile.Skills = make(map[string]map[string]ComponentLockEntry)
		}
		if lockFile.Agents == nil {
			lockFile.Agents = make(map[string]map[string]ComponentLockEntry)
		}
		if lockFile.Commands == nil {
			lockFile.Commands = make(map[string]map[string]ComponentLockEntry)
		}
	}

	now := time.Now().Format(time.RFC3339)

	// Get the appropriate nested map for this component type
	var targetMap map[string]map[string]ComponentLockEntry
	switch componentType {
	case "skills":
		targetMap = lockFile.Skills
	case "agents":
		targetMap = lockFile.Agents
	case "commands":
		targetMap = lockFile.Commands
	default:
		return fmt.Errorf("unknown component type: %s", componentType)
	}

	// Initialize source map if it doesn't exist
	if targetMap[sourceUrl] == nil {
		targetMap[sourceUrl] = make(map[string]ComponentLockEntry)
	}

	// Check if entry exists to preserve installedAt
	existingEntry, exists := targetMap[sourceUrl][componentName]
	if !exists {
		existingEntry.InstalledAt = now
	}

	// Update or add the component entry
	targetMap[sourceUrl][componentName] = ComponentLockEntry{
		Source:       source,
		SourceType:   sourceType,
		SourceUrl:    sourceUrl,
		OriginalPath: originalPath,
		CommitHash:   commitHash,
		InstalledAt:  existingEntry.InstalledAt,
		UpdatedAt:    now,
		Version:      4,
		Components:   components,
		Detection:    detection,
	}

	// Write back to file
	jsonData, err := json.MarshalIndent(lockFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}

	return os.WriteFile(lockFilePath, jsonData, 0644)
}

// LoadFromLockFile loads metadata from lock file
func LoadFromLockFile(baseDir, componentType, componentName string) (*models.ComponentMetadata, error) {
	entry, err := LoadLockFileEntry(baseDir, componentType, componentName)
	if err != nil {
		return nil, err
	}

	// Convert lock entry to metadata
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
func LoadLockFileEntry(baseDir, componentType, componentName string) (*models.ComponentLockEntry, error) {
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
	var targetMap map[string]map[string]models.ComponentLockEntry
	switch componentType {
	case "skills":
		targetMap = lockFile.Skills
	case "agents":
		targetMap = lockFile.Agents
	case "commands":
		targetMap = lockFile.Commands
	default:
		return nil, fmt.Errorf("unknown component type: %s", componentType)
	}

	// Search across all sources
	var foundEntry *models.ComponentLockEntry
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
func LoadLockFileEntryBySource(baseDir, componentType, componentName, sourceUrl string) (*models.ComponentLockEntry, error) {
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
	var targetMap map[string]map[string]models.ComponentLockEntry
	switch componentType {
	case "skills":
		targetMap = lockFile.Skills
	case "agents":
		targetMap = lockFile.Agents
	case "commands":
		targetMap = lockFile.Commands
	default:
		return nil, fmt.Errorf("unknown component type: %s", componentType)
	}

	// Check if source exists
	components, sourceExists := targetMap[sourceUrl]
	if !sourceExists {
		return nil, fmt.Errorf("source %s not found in lock file", sourceUrl)
	}

	// Check if component exists in this source
	entry, exists := components[componentName]
	if !exists {
		return nil, fmt.Errorf("component %s not found in source %s", componentName, sourceUrl)
	}

	return &entry, nil
}

// RemoveLockFileEntry removes a component entry from the lock file
// Searches across all sources and removes all matches
func RemoveLockFileEntry(baseDir, componentType, componentName string) error {
	lockFilePath := paths.GetComponentLockPath(baseDir, componentType)

	// Read existing lock file
	lockData, err := os.ReadFile(lockFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Lock file doesn't exist, nothing to remove
			return nil
		}
		return fmt.Errorf("failed to read lock file: %w", err)
	}

	var lockFile ComponentLockFile
	if err := json.Unmarshal(lockData, &lockFile); err != nil {
		return fmt.Errorf("failed to unmarshal lock file: %w", err)
	}

	// Get the appropriate nested map for this component type
	var targetMap map[string]map[string]ComponentLockEntry
	switch componentType {
	case "skills":
		targetMap = lockFile.Skills
	case "agents":
		targetMap = lockFile.Agents
	case "commands":
		targetMap = lockFile.Commands
	default:
		return fmt.Errorf("unknown component type: %s", componentType)
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

	// Write back to file
	jsonData, err := json.MarshalIndent(lockFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}

	return os.WriteFile(lockFilePath, jsonData, 0644)
}

// RemoveLockFileEntryBySource removes a component entry from a specific source
func RemoveLockFileEntryBySource(baseDir, componentType, componentName, sourceUrl string) error {
	lockFilePath := paths.GetComponentLockPath(baseDir, componentType)

	// Read existing lock file
	lockData, err := os.ReadFile(lockFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Lock file doesn't exist, nothing to remove
			return nil
		}
		return fmt.Errorf("failed to read lock file: %w", err)
	}

	var lockFile ComponentLockFile
	if err := json.Unmarshal(lockData, &lockFile); err != nil {
		return fmt.Errorf("failed to unmarshal lock file: %w", err)
	}

	// Get the appropriate nested map for this component type
	var targetMap map[string]map[string]ComponentLockEntry
	switch componentType {
	case "skills":
		targetMap = lockFile.Skills
	case "agents":
		targetMap = lockFile.Agents
	case "commands":
		targetMap = lockFile.Commands
	default:
		return fmt.Errorf("unknown component type: %s", componentType)
	}

	// Check if source exists
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

	// Write back to file
	jsonData, err := json.MarshalIndent(lockFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}

	return os.WriteFile(lockFilePath, jsonData, 0644)
}

// GetAllComponentNames returns all component names from the lock file for a given type
// Aggregates names across all sources
func GetAllComponentNames(baseDir, componentType string) ([]string, error) {
	lockFilePath := paths.GetComponentLockPath(baseDir, componentType)

	// Read lock file
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
	var targetMap map[string]map[string]models.ComponentLockEntry
	switch componentType {
	case "skills":
		targetMap = lockFile.Skills
	case "agents":
		targetMap = lockFile.Agents
	case "commands":
		targetMap = lockFile.Commands
	default:
		return nil, fmt.Errorf("unknown component type: %s", componentType)
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

	// Read lock file
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
	var targetMap map[string]map[string]models.ComponentLockEntry
	switch componentType {
	case "skills":
		targetMap = lockFile.Skills
	case "agents":
		targetMap = lockFile.Agents
	case "commands":
		targetMap = lockFile.Commands
	default:
		return nil, fmt.Errorf("unknown component type: %s", componentType)
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
	Entry     models.ComponentLockEntry
}

// FindAllComponentInstances returns all instances of a component across all sources
func FindAllComponentInstances(baseDir, componentType, componentName string) ([]ComponentSource, error) {
	lockFilePath := paths.GetComponentLockPath(baseDir, componentType)

	// Read lock file
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
	var targetMap map[string]map[string]models.ComponentLockEntry
	switch componentType {
	case "skills":
		targetMap = lockFile.Skills
	case "agents":
		targetMap = lockFile.Agents
	case "commands":
		targetMap = lockFile.Commands
	default:
		return nil, fmt.Errorf("unknown component type: %s", componentType)
	}

	// Find all instances
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
