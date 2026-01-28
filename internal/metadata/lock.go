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
type ComponentLockFile struct {
	Version  int                           `json:"version"`
	Skills   map[string]ComponentLockEntry `json:"skills"`
	Agents   map[string]ComponentLockEntry `json:"agents,omitempty"`
	Commands map[string]ComponentLockEntry `json:"commands,omitempty"`
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
func SaveLockFileEntry(baseDir, componentType, componentName, source, sourceType, sourceUrl, commitHash string, components int, detection, originalPath string) error {
	lockFilePath := paths.GetComponentLockPath(baseDir, componentType)

	// Read existing lock file or create new one
	var lockFile ComponentLockFile
	lockData, err := os.ReadFile(lockFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			lockFile = ComponentLockFile{
				Version:  3, // Current version matching agent-smith install
				Skills:   make(map[string]ComponentLockEntry),
				Agents:   make(map[string]ComponentLockEntry),
				Commands: make(map[string]ComponentLockEntry),
			}
		} else {
			return fmt.Errorf("failed to read lock file: %w", err)
		}
	} else {
		if err := json.Unmarshal(lockData, &lockFile); err != nil {
			// If lock file is corrupted, create new one
			lockFile = ComponentLockFile{
				Version:  3,
				Skills:   make(map[string]ComponentLockEntry),
				Agents:   make(map[string]ComponentLockEntry),
				Commands: make(map[string]ComponentLockEntry),
			}
		}
		// Ensure version is current
		if lockFile.Version < 3 {
			lockFile.Version = 3
			if lockFile.Skills == nil {
				lockFile.Skills = make(map[string]ComponentLockEntry)
			}
			if lockFile.Agents == nil {
				lockFile.Agents = make(map[string]ComponentLockEntry)
			}
			if lockFile.Commands == nil {
				lockFile.Commands = make(map[string]ComponentLockEntry)
			}
		}
	}

	now := time.Now().Format(time.RFC3339)

	// Get the appropriate map for this component type
	var targetMap map[string]ComponentLockEntry
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

	// Check if entry exists to preserve installedAt
	existingEntry, exists := targetMap[componentName]
	if !exists {
		existingEntry.InstalledAt = now
	}

	// Update or add the component entry
	targetMap[componentName] = ComponentLockEntry{
		Source:       source,
		SourceType:   sourceType,
		SourceUrl:    sourceUrl,
		OriginalPath: originalPath,
		CommitHash:   commitHash,
		InstalledAt:  existingEntry.InstalledAt,
		UpdatedAt:    now,
		Version:      3,
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
func LoadLockFileEntry(baseDir, componentType, componentName string) (*models.ComponentLockEntry, error) {
	var lockFilePath string
	var entries map[string]models.ComponentLockEntry

	switch componentType {
	case "skills":
		lockFilePath = paths.GetComponentLockPath(baseDir, componentType)
	case "agents":
		lockFilePath = paths.GetComponentLockPath(baseDir, componentType)
	case "commands":
		lockFilePath = paths.GetComponentLockPath(baseDir, componentType)
	default:
		return nil, fmt.Errorf("unknown component type: %s", componentType)
	}

	lockData, err := os.ReadFile(lockFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	var lockFile models.ComponentLockFile
	if err := json.Unmarshal(lockData, &lockFile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lock file: %w", err)
	}

	switch componentType {
	case "skills":
		entries = lockFile.Skills
	case "agents":
		entries = lockFile.Agents
	case "commands":
		entries = lockFile.Commands
	}

	entry, exists := entries[componentName]
	if !exists {
		return nil, fmt.Errorf("component %s not found in lock file", componentName)
	}

	return &entry, nil
}
