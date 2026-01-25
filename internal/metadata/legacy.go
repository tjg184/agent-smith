package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/tgaines/agent-smith/internal/fileutil"
	"github.com/tgaines/agent-smith/internal/models"
	"github.com/tgaines/agent-smith/pkg/paths"
)

// SaveLegacyMetadata saves legacy metadata JSON files for backward compatibility
func SaveLegacyMetadata(filePath string, metadata map[string]interface{}) error {
	metadata["downloaded"] = time.Now().Format(time.RFC3339)

	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return fileutil.CreateFileWithPermissions(filePath, jsonData)
}

// LoadLegacyMetadata loads metadata from legacy metadata files
func LoadLegacyMetadata(baseDir, componentType, componentName string) (*models.ComponentMetadata, error) {
	var metadataFile string

	switch componentType {
	case "skills":
		metadataFile = paths.GetComponentMetadataPath(baseDir, componentType, componentName)
	case "agents":
		metadataFile = paths.GetComponentMetadataPath(baseDir, componentType, componentName)
	case "commands":
		metadataFile = paths.GetComponentMetadataPath(baseDir, componentType, componentName)
	default:
		return nil, fmt.Errorf("unknown component type: %s", componentType)
	}

	data, err := os.ReadFile(metadataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata models.ComponentMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

// LoadComponentMetadata loads metadata for a component from lock files or legacy metadata files
// Tries lock file first, then falls back to legacy metadata files
func LoadComponentMetadata(baseDir, componentType, componentName string) *models.ComponentMetadata {
	// Try lock file first
	if metadata, err := LoadFromLockFile(baseDir, componentType, componentName); err == nil {
		return metadata
	}

	// Try legacy metadata file
	metadata, err := LoadLegacyMetadata(baseDir, componentType, componentName)
	if err != nil {
		return nil
	}

	return metadata
}
