package downloader

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
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
	Source          string `json:"source"`
	SourceType      string `json:"sourceType"`
	SourceUrl       string `json:"sourceUrl"`
	OriginalPath    string `json:"originalPath,omitempty"`
	SkillFolderHash string `json:"skillFolderHash,omitempty"`
	InstalledAt     string `json:"installedAt"`
	UpdatedAt       string `json:"updatedAt"`
	Version         int    `json:"version"`
	Components      int    `json:"components,omitempty"`
	Detection       string `json:"detection,omitempty"`
}

// ComputeGitHubTreeSHA computes the GitHub tree SHA for a skill folder hash (npx add-skill compatible)
func ComputeGitHubTreeSHA(ownerRepo string, skillPath string) (string, error) {
	// Normalize skill path - remove SKILL.md suffix to get folder path
	folderPath := skillPath
	if len(folderPath) >= 9 && folderPath[len(folderPath)-9:] == "/SKILL.md" {
		folderPath = folderPath[:len(folderPath)-9]
	} else if len(folderPath) >= 8 && folderPath[len(folderPath)-8:] == "SKILL.md" {
		folderPath = folderPath[:len(folderPath)-8]
	}
	if len(folderPath) > 0 && folderPath[len(folderPath)-1] == '/' {
		folderPath = folderPath[:len(folderPath)-1]
	}

	branches := []string{"main", "master"}

	for _, branch := range branches {
		url := fmt.Sprintf("https://api.github.com/repos/%s/git/trees/%s?recursive=1", ownerRepo, branch)
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		var treeData struct {
			Tree []struct {
				Path string `json:"path"`
				Type string `json:"type"`
				SHA  string `json:"sha"`
			} `json:"tree"`
		}

		if err := json.Unmarshal(body, &treeData); err != nil {
			continue
		}

		// Find tree entry for skill folder
		for _, entry := range treeData.Tree {
			if entry.Type == "tree" && entry.Path == folderPath {
				return entry.SHA, nil
			}
		}
	}

	return "", fmt.Errorf("skill folder not found in GitHub API")
}

// ComputeLocalFolderHash computes local content hash for skill folder
func ComputeLocalFolderHash(folderPath string) (string, error) {
	hasher := sha256.New()

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(folderPath, path)
		if err != nil {
			return err
		}

		// Write relative path and file info to hash
		hasher.Write([]byte(relPath))
		hasher.Write([]byte(info.Mode().String()))
		hasher.Write([]byte(info.ModTime().Format(time.RFC3339)))

		// Write file content
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		hasher.Write(data)
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to compute folder hash: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// DetermineDestinationFolderName determines the destination folder name using hierarchy heuristic
// Walks up from component file directory, skipping component-type names (agents/commands/skills)
// Returns first non-component-type directory name for preserving optional hierarchy
func DetermineDestinationFolderName(componentFilePath string) string {
	componentTypeNames := []string{"skills", "agents", "commands"}

	// Get directory containing the component file
	currentDir := filepath.Dir(componentFilePath)

	// Walk up the directory tree
	for {
		dirName := filepath.Base(currentDir)

		// Check if current directory name is a component type
		isComponentType := false
		for _, typeName := range componentTypeNames {
			if dirName == typeName {
				isComponentType = true
				break
			}
		}

		// If not a component type name, use it
		if !isComponentType && dirName != "." && dirName != "" {
			return dirName
		}

		// Go up one directory
		parentDir := filepath.Dir(currentDir)

		// Check if we've reached the root
		if parentDir == currentDir || parentDir == "." || parentDir == "/" || dirName == "" {
			// Reached root, fall back to "root"
			return "root"
		}

		currentDir = parentDir
	}
}
