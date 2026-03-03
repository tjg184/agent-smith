package downloader

import (
	"path/filepath"
	"strings"

	"github.com/tjg184/agent-smith/internal/metadata"
	"github.com/tjg184/agent-smith/internal/models"
)

// Re-export types for backward compatibility
// These now reference models.ComponentLockFile and models.ComponentEntry
type ComponentLockFile = models.ComponentLockFile
type ComponentLockEntry = models.ComponentEntry

// Re-export functions for backward compatibility
var ComputeLocalFolderHash = metadata.ComputeLocalFolderHash

// DetermineDestinationFolderName determines the destination folder name using hierarchy heuristic
// Walks up from component file directory, skipping component-type names (agents/commands/skills)
// Returns first non-component-type directory name for preserving optional hierarchy
func DetermineDestinationFolderName(componentFilePath string) string {
	componentTypeNames := []string{"skills", "agents", "commands"}

	fileName := filepath.Base(componentFilePath)
	fileExt := filepath.Ext(fileName)

	if fileExt == ".md" &&
		fileName != "SKILL.md" &&
		fileName != "AGENT.md" &&
		fileName != "COMMAND.md" {
		return strings.TrimSuffix(fileName, fileExt)
	}

	currentDir := filepath.Dir(componentFilePath)

	for {
		dirName := filepath.Base(currentDir)

		isComponentType := false
		for _, typeName := range componentTypeNames {
			if dirName == typeName {
				isComponentType = true
				break
			}
		}

		if !isComponentType && dirName != "." && dirName != "" {
			return dirName
		}

		parentDir := filepath.Dir(currentDir)

		if parentDir == currentDir || parentDir == "." || parentDir == "/" || dirName == "" {
			return "root"
		}

		currentDir = parentDir
	}
}

// extractGitHubOwnerRepo extracts "owner/repo" from a GitHub URL
// Returns empty string if not a GitHub URL or parsing fails
func extractGitHubOwnerRepo(url string) string {
	url = strings.TrimSpace(url)

	if strings.HasPrefix(url, "git@github.com:") {
		path := strings.TrimPrefix(url, "git@github.com:")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			return parts[0] + "/" + parts[1]
		}
		return ""
	}

	if strings.Contains(url, "github.com/") {
		idx := strings.Index(url, "github.com/")
		path := url[idx+len("github.com/"):]
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			return parts[0] + "/" + parts[1]
		}
	}

	return ""
}
