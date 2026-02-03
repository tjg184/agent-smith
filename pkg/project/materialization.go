package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	metadataPkg "github.com/tgaines/agent-smith/internal/metadata"
	"github.com/tgaines/agent-smith/internal/updater"
	"github.com/tgaines/agent-smith/pkg/paths"
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

// ComponentInfo represents a single materialized component with its type and metadata
type ComponentInfo struct {
	Type     string
	Name     string
	Metadata MaterializedComponentMetadata
}

// GetAllMaterializedComponents returns a flat list of all materialized components
func (m *MaterializationMetadata) GetAllMaterializedComponents() []ComponentInfo {
	var components []ComponentInfo

	for name, metadata := range m.Skills {
		components = append(components, ComponentInfo{
			Type:     "skills",
			Name:     name,
			Metadata: metadata,
		})
	}

	for name, metadata := range m.Agents {
		components = append(components, ComponentInfo{
			Type:     "agents",
			Name:     name,
			Metadata: metadata,
		})
	}

	for name, metadata := range m.Commands {
		components = append(components, ComponentInfo{
			Type:     "commands",
			Name:     name,
			Metadata: metadata,
		})
	}

	return components
}

// SyncStatus represents the sync status of a materialized component
type SyncStatus string

const (
	SyncStatusInSync        SyncStatus = "in_sync"
	SyncStatusOutOfSync     SyncStatus = "out_of_sync"
	SyncStatusSourceMissing SyncStatus = "source_missing"
)

// CheckComponentSyncStatus checks if a materialized component is in sync with its GitHub source
// Returns the sync status and any error encountered
func CheckComponentSyncStatus(componentType, componentName string, metadata MaterializedComponentMetadata) (SyncStatus, error) {
	// Check if we have a valid source URL and commit hash
	if metadata.Source == "" {
		return "", fmt.Errorf("component metadata missing source URL")
	}

	if metadata.CommitHash == "" {
		// Old metadata format without commit hash - treat as out of sync
		return SyncStatusOutOfSync, nil
	}

	// Create an updater to check GitHub
	// We use a minimal updater instance just for accessing GetCurrentRepoSHA
	baseDir, err := paths.GetAgentsDir()
	if err != nil {
		return "", fmt.Errorf("failed to get agents directory: %w", err)
	}

	ud := updater.NewUpdateDetectorWithBaseDir(baseDir)

	// Fetch the current commit hash from GitHub
	currentCommit, err := ud.GetCurrentRepoSHA(metadata.Source)
	if err != nil {
		// Check for specific error types
		errMsg := err.Error()

		// Repository not found or deleted
		if strings.Contains(errMsg, "repository not found") ||
			strings.Contains(errMsg, "not found") ||
			strings.Contains(errMsg, "404") {
			return SyncStatusSourceMissing, nil
		}

		// Authentication required
		if strings.Contains(errMsg, "authentication required") ||
			strings.Contains(errMsg, "authentication failed") ||
			strings.Contains(errMsg, "401") ||
			strings.Contains(errMsg, "403") {
			return "", fmt.Errorf("authentication required: set GITHUB_TOKEN environment variable for private repositories")
		}

		// Network or other errors
		return "", fmt.Errorf("failed to check GitHub repository: %w", err)
	}

	// Compare stored commit hash with current GitHub commit
	if currentCommit == metadata.CommitHash {
		return SyncStatusInSync, nil
	}

	return SyncStatusOutOfSync, nil
}

// UpdateMaterializationEntry updates an existing materialization entry with new hashes and timestamp
func UpdateMaterializationEntry(metadata *MaterializationMetadata, baseDir, componentType, componentName, newSourceHash, newCurrentHash string) error {
	// Get the existing entry
	componentMap := metadata.GetComponentMap(componentType)
	if componentMap == nil {
		return fmt.Errorf("invalid component type: %s", componentType)
	}

	entry, exists := componentMap[componentName]
	if !exists {
		return fmt.Errorf("component not found in metadata: %s/%s", componentType, componentName)
	}

	// Re-read lock file to get latest commit hash
	lockEntry, err := metadataPkg.LoadLockFileEntry(baseDir, componentType, componentName)
	if err != nil {
		// If lock file can't be read, preserve existing commit hash
		// This can happen if the component was uninstalled
		lockEntry = nil
	}

	// Update fields
	entry.SourceHash = newSourceHash
	entry.CurrentHash = newCurrentHash
	entry.MaterializedAt = time.Now().Format(time.RFC3339)

	// Update commit hash if we successfully read the lock file
	if lockEntry != nil {
		entry.CommitHash = lockEntry.CommitHash
	}

	// Save updated entry back to metadata
	switch componentType {
	case "skills":
		metadata.Skills[componentName] = entry
	case "agents":
		metadata.Agents[componentName] = entry
	case "commands":
		metadata.Commands[componentName] = entry
	}

	return nil
}
