package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MaterializationMetadata represents the metadata file structure
// stored in .opencode/.materializations.json or .claude/.materializations.json
type MaterializationMetadata struct {
	Version  int                                      `json:"version"`
	Skills   map[string]MaterializedComponentMetadata `json:"skills"`
	Agents   map[string]MaterializedComponentMetadata `json:"agents"`
	Commands map[string]MaterializedComponentMetadata `json:"commands"`
}

// MaterializedComponentMetadata represents metadata for a single materialized component
type MaterializedComponentMetadata struct {
	Source         string `json:"source"`
	SourceType     string `json:"sourceType"`
	SourceProfile  string `json:"sourceProfile,omitempty"`
	CommitHash     string `json:"commitHash"`
	OriginalPath   string `json:"originalPath"`
	MaterializedAt string `json:"materializedAt"`
	SourceHash     string `json:"sourceHash"`
	CurrentHash    string `json:"currentHash"`
}

// LoadMaterializationMetadata loads metadata from the target directory's .materializations.json
func LoadMaterializationMetadata(targetDir string) (*MaterializationMetadata, error) {
	metadataPath := filepath.Join(targetDir, ".materializations.json")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return empty metadata
			return &MaterializationMetadata{
				Version:  1,
				Skills:   make(map[string]MaterializedComponentMetadata),
				Agents:   make(map[string]MaterializedComponentMetadata),
				Commands: make(map[string]MaterializedComponentMetadata),
			}, nil
		}
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata MaterializationMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	// Ensure maps are initialized
	if metadata.Skills == nil {
		metadata.Skills = make(map[string]MaterializedComponentMetadata)
	}
	if metadata.Agents == nil {
		metadata.Agents = make(map[string]MaterializedComponentMetadata)
	}
	if metadata.Commands == nil {
		metadata.Commands = make(map[string]MaterializedComponentMetadata)
	}

	return &metadata, nil
}

// SaveMaterializationMetadata saves metadata to the target directory's .materializations.json
func SaveMaterializationMetadata(targetDir string, metadata *MaterializationMetadata) error {
	metadataPath := filepath.Join(targetDir, ".materializations.json")

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// AddMaterializationEntry adds or updates a materialization entry in the metadata
func AddMaterializationEntry(metadata *MaterializationMetadata, componentType, componentName, source, sourceType, sourceProfile, commitHash, originalPath, sourceHash, currentHash string) {
	now := time.Now().Format(time.RFC3339)

	entry := MaterializedComponentMetadata{
		Source:         source,
		SourceType:     sourceType,
		SourceProfile:  sourceProfile,
		CommitHash:     commitHash,
		OriginalPath:   originalPath,
		MaterializedAt: now,
		SourceHash:     sourceHash,
		CurrentHash:    currentHash,
	}

	switch componentType {
	case "skills":
		metadata.Skills[componentName] = entry
	case "agents":
		metadata.Agents[componentName] = entry
	case "commands":
		metadata.Commands[componentName] = entry
	}
}

// GetComponentMap returns the appropriate component map for the given component type
func (m *MaterializationMetadata) GetComponentMap(componentType string) map[string]MaterializedComponentMetadata {
	switch componentType {
	case "skills":
		return m.Skills
	case "agents":
		return m.Agents
	case "commands":
		return m.Commands
	default:
		return nil
	}
}
