package downloader

import (
	"path/filepath"
	"strings"
)

// DetermineDestinationFolderName determines the destination folder name, preserving any
// intermediate hierarchy between the component-type directory (skills/agents/commands) and
// the component file. For example:
//   - agents/category/my-agent.md      → category/my-agent
//   - skills/category/skill-name/SKILL.md → category/skill-name
//   - skills/skill-name/SKILL.md       → skill-name
func DetermineDestinationFolderName(componentFilePath string) string {
	componentTypeNames := []string{"skills", "agents", "commands"}

	fileName := filepath.Base(componentFilePath)
	fileExt := filepath.Ext(fileName)

	isSingleFile := fileExt == ".md" &&
		fileName != "SKILL.md" &&
		fileName != "AGENT.md" &&
		fileName != "COMMAND.md"

	// Start walking up from the file's directory (or the file itself for single-file components).
	// Collect all path segments until we hit a component-type directory.
	currentDir := filepath.Dir(componentFilePath)

	var segments []string

	if isSingleFile {
		segments = append(segments, strings.TrimSuffix(fileName, fileExt))
	}

	for {
		dirName := filepath.Base(currentDir)

		isComponentType := false
		for _, typeName := range componentTypeNames {
			if dirName == typeName {
				isComponentType = true
				break
			}
		}

		if isComponentType || dirName == "." || dirName == "" {
			break
		}

		segments = append(segments, dirName)

		parentDir := filepath.Dir(currentDir)

		if parentDir == currentDir || parentDir == "." || parentDir == "/" {
			break
		}

		currentDir = parentDir
	}

	if len(segments) == 0 {
		return "root"
	}

	// Segments were collected bottom-up (or with file stem prepended); reverse for top-down order.
	for i, j := 0, len(segments)-1; i < j; i, j = i+1, j-1 {
		segments[i], segments[j] = segments[j], segments[i]
	}

	return filepath.Join(segments...)
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
