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
// Version 2+ uses nested structure: map[sourceURL]map[componentName]MaterializedComponentMetadata
type MaterializationMetadata struct {
	Version  int                                                 `json:"version"`
	Skills   map[string]map[string]MaterializedComponentMetadata `json:"skills"`
	Agents   map[string]map[string]MaterializedComponentMetadata `json:"agents"`
	Commands map[string]map[string]MaterializedComponentMetadata `json:"commands"`
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
	FilesystemName string `json:"filesystemName"` // Actual directory name on disk (may differ from component name)
}

// LoadMaterializationMetadata loads metadata from the target directory's .materializations.json
func LoadMaterializationMetadata(targetDir string) (*MaterializationMetadata, error) {
	metadataPath := filepath.Join(targetDir, ".materializations.json")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return empty metadata with v2 structure
			return &MaterializationMetadata{
				Version:  2,
				Skills:   make(map[string]map[string]MaterializedComponentMetadata),
				Agents:   make(map[string]map[string]MaterializedComponentMetadata),
				Commands: make(map[string]map[string]MaterializedComponentMetadata),
			}, nil
		}
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata MaterializationMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	// Ensure maps are initialized and update version
	metadata.Version = 2
	if metadata.Skills == nil {
		metadata.Skills = make(map[string]map[string]MaterializedComponentMetadata)
	}
	if metadata.Agents == nil {
		metadata.Agents = make(map[string]map[string]MaterializedComponentMetadata)
	}
	if metadata.Commands == nil {
		metadata.Commands = make(map[string]map[string]MaterializedComponentMetadata)
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
// Uses nested structure by source URL
func AddMaterializationEntry(metadata *MaterializationMetadata, componentType, componentName, source, sourceType, sourceProfile, commitHash, originalPath, sourceHash, currentHash, filesystemName string) {
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
		FilesystemName: filesystemName,
	}

	// Get or create the nested map for this component type
	var targetMap map[string]map[string]MaterializedComponentMetadata
	switch componentType {
	case "skills":
		targetMap = metadata.Skills
	case "agents":
		targetMap = metadata.Agents
	case "commands":
		targetMap = metadata.Commands
	default:
		return
	}

	// Initialize source map if it doesn't exist
	if targetMap[source] == nil {
		targetMap[source] = make(map[string]MaterializedComponentMetadata)
	}

	// Add or update the entry
	targetMap[source][componentName] = entry
}

// GetComponentMap returns the appropriate nested component map for the given component type
func (m *MaterializationMetadata) GetComponentMap(componentType string) map[string]map[string]MaterializedComponentMetadata {
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

// ResolveFilesystemName determines the actual filesystem name to use for a component
// If the exact component (sourceUrl + componentName) is already materialized, returns its existing filesystem name
// Otherwise, if componentName already exists, returns componentName-2, componentName-3, etc.
func ResolveFilesystemName(targetDir, componentType, componentName, sourceUrl string, metadata *MaterializationMetadata) string {
	// First, check if this exact component (sourceUrl + componentName) is already materialized
	// If so, reuse its existing filesystem name for idempotency
	if sourceUrl != "" {
		existingFilesystemName := findExistingFilesystemName(componentType, componentName, sourceUrl, metadata)
		if existingFilesystemName != "" {
			return existingFilesystemName
		}
	}

	// Check both filesystem and metadata for conflicts
	baseComponentDir := filepath.Join(targetDir, componentName)

	// If the base name doesn't exist on disk or in metadata, use it
	if !filesystemNameExists(baseComponentDir) && !metadataFilesystemNameExists(componentName, componentType, metadata) {
		return componentName
	}

	// Find the next available suffix
	suffix := 2
	for {
		candidateName := fmt.Sprintf("%s-%d", componentName, suffix)
		candidatePath := filepath.Join(targetDir, candidateName)

		if !filesystemNameExists(candidatePath) && !metadataFilesystemNameExists(candidateName, componentType, metadata) {
			return candidateName
		}

		suffix++

		// Safety check to prevent infinite loops
		if suffix > 1000 {
			return fmt.Sprintf("%s-%d", componentName, suffix)
		}
	}
}

// filesystemNameExists checks if a path exists on disk
func filesystemNameExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// metadataFilesystemNameExists checks if a filesystem name is already used in metadata
// Only checks within the specified component type since each type has its own directory
func metadataFilesystemNameExists(filesystemName string, componentType string, metadata *MaterializationMetadata) bool {
	var componentMap map[string]map[string]MaterializedComponentMetadata

	switch componentType {
	case "skills":
		componentMap = metadata.Skills
	case "agents":
		componentMap = metadata.Agents
	case "commands":
		componentMap = metadata.Commands
	default:
		return false
	}

	// Check only the specified component type
	for _, sourceComponents := range componentMap {
		for _, entry := range sourceComponents {
			if entry.FilesystemName == filesystemName {
				return true
			}
		}
	}

	return false
}

// findExistingFilesystemName checks if a component with the given sourceUrl and componentName
// is already materialized, and returns its filesystem name if found
func findExistingFilesystemName(componentType, componentName, sourceUrl string, metadata *MaterializationMetadata) string {
	var componentMap map[string]map[string]MaterializedComponentMetadata

	switch componentType {
	case "skills":
		componentMap = metadata.Skills
	case "agents":
		componentMap = metadata.Agents
	case "commands":
		componentMap = metadata.Commands
	default:
		return ""
	}

	// Check if this source has components
	sourceComponents, exists := componentMap[sourceUrl]
	if !exists {
		return ""
	}

	// Check if this specific component is materialized from this source
	if entry, exists := sourceComponents[componentName]; exists {
		return entry.FilesystemName
	}

	return ""
}

// ComponentInfo represents a single materialized component with its type and metadata
type ComponentInfo struct {
	Type     string
	Name     string
	Metadata MaterializedComponentMetadata
}

// GetAllMaterializedComponents returns a flat list of all materialized components
// Iterates through nested structure and flattens
func (m *MaterializationMetadata) GetAllMaterializedComponents() []ComponentInfo {
	var components []ComponentInfo

	for _, sourceComponents := range m.Skills {
		for name, metadata := range sourceComponents {
			components = append(components, ComponentInfo{
				Type:     "skills",
				Name:     name,
				Metadata: metadata,
			})
		}
	}

	for _, sourceComponents := range m.Agents {
		for name, metadata := range sourceComponents {
			components = append(components, ComponentInfo{
				Type:     "agents",
				Name:     name,
				Metadata: metadata,
			})
		}
	}

	for _, sourceComponents := range m.Commands {
		for name, metadata := range sourceComponents {
			components = append(components, ComponentInfo{
				Type:     "commands",
				Name:     name,
				Metadata: metadata,
			})
		}
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

// SyncCheckResult holds the result of a sync status check
type SyncCheckResult struct {
	Status SyncStatus
	Error  error
}

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

// CheckMultipleComponentsSyncStatusBatched checks sync status for multiple components
// by batching components from the same repository to reduce git clone operations.
// Returns a map of component key (type/name) -> sync check result
func CheckMultipleComponentsSyncStatusBatched(baseDir string, components []ComponentInfo) (map[string]SyncCheckResult, error) {
	results := make(map[string]SyncCheckResult)

	// Group components by source repository
	componentsByRepo := make(map[string][]ComponentInfo)
	for _, comp := range components {
		if comp.Metadata.Source == "" {
			// Handle components with missing source URL
			key := fmt.Sprintf("%s/%s", comp.Type, comp.Name)
			results[key] = SyncCheckResult{
				Status: "",
				Error:  fmt.Errorf("component metadata missing source URL"),
			}
			continue
		}

		if comp.Metadata.CommitHash == "" {
			// Old metadata format without commit hash - treat as out of sync
			key := fmt.Sprintf("%s/%s", comp.Type, comp.Name)
			results[key] = SyncCheckResult{
				Status: SyncStatusOutOfSync,
				Error:  nil,
			}
			continue
		}

		componentsByRepo[comp.Metadata.Source] = append(componentsByRepo[comp.Metadata.Source], comp)
	}

	// Create updater for checking repositories
	ud := updater.NewUpdateDetectorWithBaseDir(baseDir)

	// Process each repository batch
	for repoURL, repoComponents := range componentsByRepo {
		// Fetch the current commit hash from GitHub once for this repository
		currentCommit, err := ud.GetCurrentRepoSHA(repoURL)

		if err != nil {
			// Check for specific error types
			errMsg := err.Error()

			// Repository not found or deleted
			if strings.Contains(errMsg, "repository not found") ||
				strings.Contains(errMsg, "not found") ||
				strings.Contains(errMsg, "404") {
				// Mark all components from this repo as source missing
				for _, comp := range repoComponents {
					key := fmt.Sprintf("%s/%s", comp.Type, comp.Name)
					results[key] = SyncCheckResult{
						Status: SyncStatusSourceMissing,
						Error:  nil,
					}
				}
				continue
			}

			// Authentication required
			if strings.Contains(errMsg, "authentication required") ||
				strings.Contains(errMsg, "authentication failed") ||
				strings.Contains(errMsg, "401") ||
				strings.Contains(errMsg, "403") {
				// Mark all components from this repo with auth error
				authErr := fmt.Errorf("authentication required: set GITHUB_TOKEN environment variable for private repositories")
				for _, comp := range repoComponents {
					key := fmt.Sprintf("%s/%s", comp.Type, comp.Name)
					results[key] = SyncCheckResult{
						Status: "",
						Error:  authErr,
					}
				}
				continue
			}

			// Network or other errors
			networkErr := fmt.Errorf("failed to check GitHub repository: %w", err)
			for _, comp := range repoComponents {
				key := fmt.Sprintf("%s/%s", comp.Type, comp.Name)
				results[key] = SyncCheckResult{
					Status: "",
					Error:  networkErr,
				}
			}
			continue
		}

		// Compare each component's stored commit hash with current GitHub commit
		for _, comp := range repoComponents {
			key := fmt.Sprintf("%s/%s", comp.Type, comp.Name)

			if currentCommit == comp.Metadata.CommitHash {
				results[key] = SyncCheckResult{
					Status: SyncStatusInSync,
					Error:  nil,
				}
			} else {
				results[key] = SyncCheckResult{
					Status: SyncStatusOutOfSync,
					Error:  nil,
				}
			}
		}
	}

	return results, nil
}

// UpdateMaterializationEntry updates an existing materialization entry with new hashes and timestamp
// Searches across all sources and updates all matches
func UpdateMaterializationEntry(metadata *MaterializationMetadata, baseDir, componentType, componentName, newSourceHash, newCurrentHash string) error {
	// Get the nested component map
	componentMap := metadata.GetComponentMap(componentType)
	if componentMap == nil {
		return fmt.Errorf("invalid component type: %s", componentType)
	}

	// Find all instances of this component across sources
	found := false
	for sourceUrl, components := range componentMap {
		entry, exists := components[componentName]
		if !exists {
			continue
		}
		found = true

		// Re-read lock file to get latest commit hash for this source
		lockEntry, err := metadataPkg.LoadLockFileEntryBySource(baseDir, componentType, componentName, sourceUrl)
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
		components[componentName] = entry
	}

	if !found {
		return fmt.Errorf("component not found in metadata: %s/%s", componentType, componentName)
	}

	return nil
}
