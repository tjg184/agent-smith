package downloader

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/tgaines/agent-smith/internal/detector"
	"github.com/tgaines/agent-smith/internal/fileutil"
	metadataPkg "github.com/tgaines/agent-smith/internal/metadata"
	"github.com/tgaines/agent-smith/internal/models"
	"github.com/tgaines/agent-smith/pkg/paths"
)

// AgentDownloader handles downloading agent components
type AgentDownloader struct {
	baseDir  string
	detector *detector.RepositoryDetector
}

// NewAgentDownloader creates a new AgentDownloader instance
func NewAgentDownloader() *AgentDownloader {
	baseDir, err := paths.GetAgentsSubDir()
	if err != nil {
		log.Fatal("Failed to get agents directory:", err)
	}

	// Create base directory if it doesn't exist
	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create agents directory:", err)
	}

	return &AgentDownloader{
		baseDir:  baseDir,
		detector: detector.NewRepositoryDetector(),
	}
}

// NewAgentDownloaderWithParams creates a new AgentDownloader with custom parameters (for testing)
func NewAgentDownloaderWithParams(baseDir string, detect *detector.RepositoryDetector) *AgentDownloader {
	return &AgentDownloader{
		baseDir:  baseDir,
		detector: detect,
	}
}

func (ad *AgentDownloader) parseRepoURL(repoURL string) (string, error) {
	// Normalize URL first (handles GitHub shorthand, etc.)
	normalizedURL, err := ad.detector.NormalizeURL(repoURL)
	if err != nil {
		return "", err
	}

	// Validate normalized repository
	if err := ad.detector.ValidateRepository(normalizedURL); err != nil {
		return "", fmt.Errorf("repository validation failed: %w", err)
	}

	return normalizedURL, nil
}

// DownloadAgent downloads an agent from the repository
func (ad *AgentDownloader) DownloadAgent(repoURL, agentName string, providedRepoPath ...string) error {
	fullURL, err := ad.parseRepoURL(repoURL)
	if err != nil {
		return err
	}

	var repoPath string
	hasProvidedPath := len(providedRepoPath) > 0 && providedRepoPath[0] != ""

	// Use provided repo path if available, otherwise clone for detection
	if hasProvidedPath {
		repoPath = providedRepoPath[0]
	} else if ad.detector.DetectProvider(repoURL) == "local" {
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

	// Detect components in the repository
	components, err := ad.detector.DetectComponentsInRepo(repoPath)
	if err != nil {
		return fmt.Errorf("failed to detect components: %w", err)
	}

	// Filter for agent components
	var agentComponents []models.DetectedComponent
	for _, comp := range components {
		if comp.Type == models.ComponentAgent {
			agentComponents = append(agentComponents, comp)
		}
	}

	if len(agentComponents) == 0 {
		// No agent components detected, fall back to original behavior
		return ad.downloadAgentDirect(fullURL, agentName)
	}

	// Create agent directory
	agentDir := filepath.Join(ad.baseDir, agentName)
	if err := fileutil.CreateDirectoryWithPermissions(agentDir); err != nil {
		return fmt.Errorf("failed to create agent directory: %w", err)
	}

	// Check if the requested agentName matches one of the detected components
	var matchingComponent *models.DetectedComponent
	for _, comp := range agentComponents {
		if comp.Name == agentName {
			matchingComponent = &comp
			break
		}
	}

	// Additional check: if we found a matching component but it's part of a larger directory structure,
	// prefer components that have their own directory (more specific)
	if matchingComponent != nil && len(agentComponents) > 1 {
		for _, comp := range agentComponents {
			if comp.Name == agentName && comp.Path != matchingComponent.Path {
				// Found a more specific version (different path)
				matchingComponent = &comp
				break
			}
		}
	}

	// If only one agent component found, copy its contents
	if len(agentComponents) == 1 {
		component := agentComponents[0]

		// Copy component files (non-recursive) using FilePath to agent directory
		err = fileutil.CopyComponentFiles(repoPath, component, agentDir)
		if err != nil {
			os.RemoveAll(agentDir)
			return fmt.Errorf("failed to copy agent files: %w", err)
		}
	} else if matchingComponent != nil {
		// Downloading a specific component from a multi-component directory
		// Use direct copy to avoid double nesting

		// Copy component files (non-recursive) using FilePath to agent directory
		err = fileutil.CopyComponentFiles(repoPath, *matchingComponent, agentDir)
		if err != nil {
			os.RemoveAll(agentDir)
			return fmt.Errorf("failed to copy agent files: %w", err)
		}
	} else {
		// Multiple agents found, create a monorepo structure
		for _, component := range agentComponents {
			componentDir := filepath.Join(agentDir, component.Name)

			err = fileutil.CreateDirectoryWithPermissions(componentDir)
			if err != nil {
				continue
			}

			// Copy component files (non-recursive) using FilePath
			err = fileutil.CopyComponentFiles(repoPath, component, componentDir)
			if err != nil {
				log.Printf("Warning: failed to copy agent %s: %v", component.Name, err)
			}
		}
	}

	// Save to lock file
	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	// Compute folder hash if it's a GitHub repo
	var folderHash string
	if sourceType != "github" {
		if hash, err := ComputeLocalFolderHash(agentDir); err == nil {
			folderHash = hash
		}
	}

	if err := ad.saveLockFile(agentName, fullURL, sourceType, fullURL, folderHash, len(agentComponents), "recursive", ""); err != nil {
		log.Printf("Warning: failed to save lock file: %v", err)
	}

	// Clean up git clone if it exists
	if _, err := os.Stat(agentDir + ".git"); err == nil {
		os.RemoveAll(agentDir + ".git")
	}

	fmt.Printf("Successfully downloaded agent '%s' from %s\n", agentName, fullURL)
	fmt.Printf("Agent stored in: %s\n", agentDir)
	fmt.Printf("Components detected: %d\n", len(agentComponents))

	return nil
}

func (ad *AgentDownloader) downloadAgentDirect(fullURL, agentName string) error {
	// Create agent directory
	agentDir := filepath.Join(ad.baseDir, agentName)
	if err := fileutil.CreateDirectoryWithPermissions(agentDir); err != nil {
		return fmt.Errorf("failed to create agent directory: %w", err)
	}

	// Clone repository directly
	_, err := git.PlainClone(agentDir, false, &git.CloneOptions{
		URL:           fullURL,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	})
	if err != nil {
		os.RemoveAll(agentDir)
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Save to lock file
	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	// Compute folder hash if it's a GitHub repo
	var folderHash string
	if sourceType != "github" {
		if hash, err := ComputeLocalFolderHash(agentDir); err == nil {
			folderHash = hash
		}
	}

	if err := ad.saveLockFile(agentName, fullURL, sourceType, fullURL, folderHash, 1, "direct", ""); err != nil {
		log.Printf("Warning: failed to save lock file: %v", err)
	}

	// Create {name}.md if it doesn't exist
	agentFile := filepath.Join(agentDir, agentName+".md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		if err := ad.createAgentFile(agentFile, agentName, fullURL); err != nil {
			log.Printf("Warning: failed to create %s.md: %v", agentName, err)
		}
	}

	return nil
}

// saveLockFile saves agent lock entry in agent-smith install compatible format
func (ad *AgentDownloader) saveLockFile(agentName string, source string, sourceType string, sourceUrl string, skillFolderHash string, components int, detection string, originalPath string) error {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return fmt.Errorf("failed to get agents directory: %w", err)
	}

	if err := fileutil.CreateDirectoryWithPermissions(agentsDir); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	return metadataPkg.SaveLockFileEntry(agentsDir, "agents", agentName, source, sourceType, sourceUrl, skillFolderHash, components, detection, originalPath)
}

func (ad *AgentDownloader) createAgentFile(filePath, agentName, source string) error {
	content := fmt.Sprintf(`# %s

Downloaded from: %s

## Description

This agent was automatically downloaded by Agent Smith.

## Usage

Add usage instructions here.

---
*Auto-generated by Agent Smith*
`, agentName, source)

	return fileutil.CreateFileWithPermissions(filePath, []byte(content))
}

// DownloadAgentWithRepo downloads an agent with repo path provided
func (ad *AgentDownloader) DownloadAgentWithRepo(fullURL, agentName, repoURL string, repoPath string, components []models.DetectedComponent) error {
	// Find the specific agent component with matching name
	var targetComponent *models.DetectedComponent
	for _, comp := range components {
		if comp.Type == models.ComponentAgent && comp.Name == agentName {
			targetComponent = &comp
			break
		}
	}

	if targetComponent == nil {
		// Agent component not found in provided components, fall back to original behavior
		return ad.downloadAgentDirect(fullURL, agentName)
	}

	// Determine destination folder name using heuristic
	destFolderName := DetermineDestinationFolderName(targetComponent.FilePath)

	// Create agent directory with heuristic name
	agentDir := filepath.Join(ad.baseDir, destFolderName)
	if err := fileutil.CreateDirectoryWithPermissions(agentDir); err != nil {
		return fmt.Errorf("failed to create agent directory: %w", err)
	}

	// Copy the entire component directory recursively
	err := fileutil.CopyComponentFiles(repoPath, *targetComponent, agentDir)
	if err != nil {
		os.RemoveAll(agentDir)
		return fmt.Errorf("failed to copy agent files: %w", err)
	}

	// Save to lock file
	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	// Compute folder hash if it's a GitHub repo
	var folderHash string
	if sourceType != "github" {
		if hash, err := ComputeLocalFolderHash(agentDir); err == nil {
			folderHash = hash
		}
	}

	if err := ad.saveLockFile(destFolderName, fullURL, sourceType, fullURL, folderHash, 1, "single", targetComponent.FilePath); err != nil {
		log.Printf("Warning: failed to save lock file: %v", err)
	}

	// Clean up git clone if it exists
	if _, err := os.Stat(agentDir + ".git"); err == nil {
		os.RemoveAll(agentDir + ".git")
	}

	fmt.Printf("Successfully downloaded agent '%s' from %s\n", agentName, fullURL)
	fmt.Printf("Agent stored in: %s\n", agentDir)

	return nil
}
