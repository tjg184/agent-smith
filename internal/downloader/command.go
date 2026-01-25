package downloader

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/tgaines/agent-smith/internal/detector"
	"github.com/tgaines/agent-smith/internal/fileutil"
	"github.com/tgaines/agent-smith/internal/models"
	"github.com/tgaines/agent-smith/pkg/paths"
)

// CommandDownloader handles downloading command components
type CommandDownloader struct {
	baseDir  string
	detector *detector.RepositoryDetector
}

// NewCommandDownloader creates a new CommandDownloader instance
func NewCommandDownloader() *CommandDownloader {
	baseDir, err := paths.GetCommandsDir()
	if err != nil {
		log.Fatal("Failed to get commands directory:", err)
	}

	// Create base directory if it doesn't exist
	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create commands directory:", err)
	}

	return &CommandDownloader{
		baseDir:  baseDir,
		detector: detector.NewRepositoryDetector(),
	}
}

func (cd *CommandDownloader) parseRepoURL(repoURL string) (string, error) {
	// Normalize URL first (handles GitHub shorthand, etc.)
	normalizedURL, err := cd.detector.NormalizeURL(repoURL)
	if err != nil {
		return "", err
	}

	// Validate normalized repository
	if err := cd.detector.ValidateRepository(normalizedURL); err != nil {
		return "", fmt.Errorf("repository validation failed: %w", err)
	}

	return normalizedURL, nil
}

// DownloadCommand downloads a command from the repository
func (cd *CommandDownloader) DownloadCommand(repoURL, commandName string, providedRepoPath ...string) error {
	fullURL, err := cd.parseRepoURL(repoURL)
	if err != nil {
		return err
	}

	var repoPath string
	hasProvidedPath := len(providedRepoPath) > 0 && providedRepoPath[0] != ""

	// Use provided repo path if available, otherwise clone for detection
	if hasProvidedPath {
		repoPath = providedRepoPath[0]
	} else if cd.detector.DetectProvider(repoURL) == "local" {
		// For local repositories, use path directly
		repoPath = fullURL
	} else {
		// For remote repositories, create temporary directory for repository detection
		tempDir, err := os.MkdirTemp("", "agent-smith-detect-*")
		if err != nil {
			return fmt.Errorf("failed to create temporary directory: %w", err)
		}
		defer os.RemoveAll(tempDir)

		// Clone repository to temporary location for detection
		_, err = git.PlainClone(tempDir, false, &git.CloneOptions{
			URL:           fullURL,
			Depth:         1,
			ReferenceName: plumbing.HEAD,
			SingleBranch:  true,
		})
		if err != nil {
			return fmt.Errorf("failed to clone repository for detection: %w", err)
		}
		repoPath = tempDir
	}

	// Detect components in repository
	components, err := cd.detector.DetectComponentsInRepo(repoPath)
	if err != nil {
		return fmt.Errorf("failed to detect components: %w", err)
	}

	// Filter for command components
	var commandComponents []models.DetectedComponent
	for _, comp := range components {
		if comp.Type == models.ComponentCommand {
			commandComponents = append(commandComponents, comp)
		}
	}

	if len(commandComponents) == 0 {
		// No command components detected, fall back to original behavior
		return cd.downloadCommandDirect(fullURL, commandName)
	}

	// Create command directory
	commandDir := filepath.Join(cd.baseDir, commandName)
	if err := fileutil.CreateDirectoryWithPermissions(commandDir); err != nil {
		return fmt.Errorf("failed to create command directory: %w", err)
	}

	// Check if the requested commandName matches one of the detected components
	var matchingComponent *models.DetectedComponent
	for _, comp := range commandComponents {
		if comp.Name == commandName {
			matchingComponent = &comp
			break
		}
	}

	// Additional check: if we found a matching component but it's part of a larger directory structure,
	// prefer components that have their own directory (more specific)
	if matchingComponent != nil && len(commandComponents) > 1 {
		for _, comp := range commandComponents {
			if comp.Name == commandName && comp.Path != matchingComponent.Path {
				// Found a more specific version (different path)
				matchingComponent = &comp
				break
			}
		}
	}

	// If only one command component found, copy its contents
	if len(commandComponents) == 1 {
		component := commandComponents[0]

		// Copy component files (non-recursive) using FilePath to command directory
		err = fileutil.CopyComponentFiles(repoPath, component, commandDir)
		if err != nil {
			os.RemoveAll(commandDir)
			return fmt.Errorf("failed to copy command files: %w", err)
		}
	} else if matchingComponent != nil {
		// Downloading a specific component from a multi-component directory
		// Use direct copy to avoid double nesting

		// Copy component files (non-recursive) using FilePath to command directory
		err = fileutil.CopyComponentFiles(repoPath, *matchingComponent, commandDir)
		if err != nil {
			os.RemoveAll(commandDir)
			return fmt.Errorf("failed to copy command files: %w", err)
		}
	} else {
		// Multiple commands found, create a monorepo structure
		for _, component := range commandComponents {
			componentDir := filepath.Join(commandDir, component.Name)

			err = fileutil.CreateDirectoryWithPermissions(componentDir)
			if err != nil {
				continue
			}

			// Copy component files (non-recursive) using FilePath
			err = fileutil.CopyComponentFiles(repoPath, component, componentDir)
			if err != nil {
				log.Printf("Warning: failed to copy command %s: %v", component.Name, err)
			}
		}
	}

	// Clone the repository again to get proper git history for metadata
	repo, err := git.PlainClone(commandDir+".git", true, &git.CloneOptions{
		URL:           fullURL,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	})
	if err != nil {
		// Non-fatal, continue without git metadata
		log.Printf("Warning: failed to clone repository for metadata: %v", err)
	}

	var commitHash string
	if repo != nil {
		if ref, err := repo.Head(); err == nil {
			commitHash = ref.Hash().String()
		}
	}

	// Create metadata
	metadata := map[string]interface{}{
		"name":       commandName,
		"source":     fullURL,
		"commit":     commitHash,
		"downloaded": "now",
		"components": len(commandComponents),
		"detection":  "recursive",
	}

	// Save legacy metadata file for backward compatibility
	legacyMetadataFile := filepath.Join(commandDir, ".command-metadata.json")
	if err := cd.saveMetadata(legacyMetadataFile, metadata); err != nil {
		log.Printf("Warning: failed to save legacy metadata: %v", err)
	}

	// Save to npx add-skill compatible lock file
	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	// Compute folder hash if it's a GitHub repo
	var folderHash string
	if sourceType != "github" {
		if hash, err := ComputeLocalFolderHash(commandDir); err == nil {
			folderHash = hash
		}
	}

	if err := cd.saveLockFile(commandName, fullURL, sourceType, fullURL, folderHash, len(commandComponents), "recursive", ""); err != nil {
		log.Printf("Warning: failed to save lock file: %v", err)
	}

	// Clean up git clone if it exists
	if _, err := os.Stat(commandDir + ".git"); err == nil {
		os.RemoveAll(commandDir + ".git")
	}

	fmt.Printf("Successfully downloaded command '%s' from %s\n", commandName, fullURL)
	fmt.Printf("Command stored in: %s\n", commandDir)
	fmt.Printf("Components detected: %d\n", len(commandComponents))

	return nil
}

func (cd *CommandDownloader) downloadCommandDirect(fullURL, commandName string) error {
	// Create command directory
	commandDir := filepath.Join(cd.baseDir, commandName)
	if err := fileutil.CreateDirectoryWithPermissions(commandDir); err != nil {
		return fmt.Errorf("failed to create command directory: %w", err)
	}

	// Clone repository directly
	repo, err := git.PlainClone(commandDir, false, &git.CloneOptions{
		URL:           fullURL,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	})
	if err != nil {
		os.RemoveAll(commandDir)
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Get repository info for metadata
	var commitHash string
	if ref, err := repo.Head(); err == nil {
		commitHash = ref.Hash().String()
	}

	// Create metadata
	metadata := map[string]interface{}{
		"name":       commandName,
		"source":     fullURL,
		"commit":     commitHash,
		"downloaded": "now",
		"detection":  "direct",
	}

	// Save legacy metadata file for backward compatibility
	legacyMetadataFile := filepath.Join(commandDir, ".command-metadata.json")
	if err := cd.saveMetadata(legacyMetadataFile, metadata); err != nil {
		log.Printf("Warning: failed to save legacy metadata: %v", err)
	}

	// Save to npx add-skill compatible lock file
	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	// Compute folder hash if it's a GitHub repo
	var folderHash string
	if sourceType != "github" {
		if hash, err := ComputeLocalFolderHash(commandDir); err == nil {
			folderHash = hash
		}
	}

	if err := cd.saveLockFile(commandName, fullURL, sourceType, fullURL, folderHash, 1, "direct", ""); err != nil {
		log.Printf("Warning: failed to save lock file: %v", err)
	}

	// Create {name}.md if it doesn't exist
	commandFile := filepath.Join(commandDir, commandName+".md")
	if _, err := os.Stat(commandFile); os.IsNotExist(err) {
		if err := cd.createCommandFile(commandFile, commandName, fullURL); err != nil {
			log.Printf("Warning: failed to create %s.md: %v", commandName, err)
		}
	}

	return nil
}

func (cd *CommandDownloader) saveMetadata(filePath string, metadata map[string]interface{}) error {
	metadata["downloaded"] = time.Now().Format(time.RFC3339)

	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return fileutil.CreateFileWithPermissions(filePath, jsonData)
}

// saveLockFile saves command lock entry in npx add-skill compatible format
func (cd *CommandDownloader) saveLockFile(commandName string, source string, sourceType string, sourceUrl string, skillFolderHash string, components int, detection string, originalPath string) error {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return fmt.Errorf("failed to get agents directory: %w", err)
	}

	if err := fileutil.CreateDirectoryWithPermissions(agentsDir); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	lockFilePath := paths.GetComponentLockPath(agentsDir, "commands")

	// Read existing lock file or create new one
	var lockFile ComponentLockFile
	lockData, err := os.ReadFile(lockFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			lockFile = ComponentLockFile{
				Version:  3,
				Commands: make(map[string]ComponentLockEntry),
			}
		} else {
			return fmt.Errorf("failed to read lock file: %w", err)
		}
	} else {
		if err := json.Unmarshal(lockData, &lockFile); err != nil {
			lockFile = ComponentLockFile{
				Version:  3,
				Commands: make(map[string]ComponentLockEntry),
			}
		}
		if lockFile.Version < 3 {
			lockFile.Version = 3
			if lockFile.Commands == nil {
				lockFile.Commands = make(map[string]ComponentLockEntry)
			}
		}
	}

	now := time.Now().Format(time.RFC3339)

	existingEntry, exists := lockFile.Commands[commandName]
	if !exists {
		existingEntry.InstalledAt = now
	}

	lockFile.Commands[commandName] = ComponentLockEntry{
		Source:          source,
		SourceType:      sourceType,
		SourceUrl:       sourceUrl,
		OriginalPath:    originalPath, // Track original path in repo
		SkillFolderHash: skillFolderHash,
		InstalledAt:     existingEntry.InstalledAt,
		UpdatedAt:       now,
		Version:         3,
		Components:      components,
		Detection:       detection,
	}

	jsonData, err := json.MarshalIndent(lockFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}

	return os.WriteFile(lockFilePath, jsonData, 0644)
}

func (cd *CommandDownloader) createCommandFile(filePath, commandName, source string) error {
	content := fmt.Sprintf(`# %s

Downloaded from: %s

## Description

This command was automatically downloaded by Agent Smith.

## Usage

Add usage instructions here.

---
*Auto-generated by Agent Smith*
`, commandName, source)

	return fileutil.CreateFileWithPermissions(filePath, []byte(content))
}

// DownloadCommandWithRepo downloads a command with repo path provided
func (cd *CommandDownloader) DownloadCommandWithRepo(fullURL, commandName, repoURL string, repoPath string, components []models.DetectedComponent) error {
	// Find the specific command component with matching name
	var targetComponent *models.DetectedComponent
	for _, comp := range components {
		if comp.Type == models.ComponentCommand && comp.Name == commandName {
			targetComponent = &comp
			break
		}
	}

	if targetComponent == nil {
		// Command component not found in provided components, fall back to original behavior
		return cd.downloadCommandDirect(fullURL, commandName)
	}

	// Determine destination folder name using heuristic
	destFolderName := DetermineDestinationFolderName(targetComponent.FilePath)

	// Create command directory with heuristic name
	commandDir := filepath.Join(cd.baseDir, destFolderName)
	if err := fileutil.CreateDirectoryWithPermissions(commandDir); err != nil {
		return fmt.Errorf("failed to create command directory: %w", err)
	}

	// Copy the entire component directory recursively
	err := fileutil.CopyComponentFiles(repoPath, *targetComponent, commandDir)
	if err != nil {
		os.RemoveAll(commandDir)
		return fmt.Errorf("failed to copy command files: %w", err)
	}

	var commitHash string
	var repo *git.Repository

	// Handle metadata differently for local vs remote repositories
	if cd.detector.DetectProvider(repoURL) == "local" {
		// For local repositories, open the repository directly
		var err error
		repo, err = git.PlainOpen(fullURL)
		if err != nil {
			// Non-fatal, continue without git metadata
			log.Printf("Warning: failed to open local repository for metadata: %v", err)
		}
	} else {
		// For remote repositories, use the already-cloned repository at repoPath
		var err error
		repo, err = git.PlainOpen(repoPath)
		if err != nil {
			// Non-fatal, continue without git metadata
			log.Printf("Warning: failed to open repository for metadata: %v", err)
		}
	}

	if repo != nil {
		if ref, err := repo.Head(); err == nil {
			commitHash = ref.Hash().String()
		}
	}

	// Create metadata
	metadata := map[string]interface{}{
		"name":         commandName,
		"source":       fullURL,
		"commit":       commitHash,
		"downloaded":   "now",
		"components":   1,
		"detection":    "single",
		"originalPath": targetComponent.FilePath,
	}

	// Save legacy metadata file for backward compatibility
	legacyMetadataFile := filepath.Join(commandDir, ".command-metadata.json")
	if err := cd.saveMetadata(legacyMetadataFile, metadata); err != nil {
		log.Printf("Warning: failed to save legacy metadata: %v", err)
	}

	// Save to npx add-skill compatible lock file
	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	// Compute folder hash if it's a GitHub repo
	var folderHash string
	if sourceType != "github" {
		if hash, err := ComputeLocalFolderHash(commandDir); err == nil {
			folderHash = hash
		}
	}

	if err := cd.saveLockFile(destFolderName, fullURL, sourceType, fullURL, folderHash, 1, "single", targetComponent.FilePath); err != nil {
		log.Printf("Warning: failed to save lock file: %v", err)
	}

	// Clean up git clone if it exists
	if _, err := os.Stat(commandDir + ".git"); err == nil {
		os.RemoveAll(commandDir + ".git")
	}

	fmt.Printf("Successfully downloaded command '%s' from %s\n", commandName, fullURL)
	fmt.Printf("Command stored in: %s\n", commandDir)

	return nil
}
