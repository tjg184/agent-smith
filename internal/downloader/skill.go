package downloader

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/tgaines/agent-smith/internal/detector"
	"github.com/tgaines/agent-smith/internal/fileutil"
	gitpkg "github.com/tgaines/agent-smith/internal/git"
	metadataPkg "github.com/tgaines/agent-smith/internal/metadata"
	"github.com/tgaines/agent-smith/internal/models"
	"github.com/tgaines/agent-smith/pkg/paths"
)

// SkillDownloader handles downloading skill components
type SkillDownloader struct {
	baseDir  string
	detector *detector.RepositoryDetector
	cloner   gitpkg.Cloner
}

// NewSkillDownloader creates a new SkillDownloader instance
func NewSkillDownloader() *SkillDownloader {
	baseDir, err := paths.GetSkillsDir()
	if err != nil {
		log.Fatal("Failed to get skills directory:", err)
	}

	// Create base directory if it doesn't exist
	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create skills directory:", err)
	}

	return &SkillDownloader{
		baseDir:  baseDir,
		detector: detector.NewRepositoryDetector(),
		cloner:   gitpkg.NewDefaultCloner(),
	}
}

func (sd *SkillDownloader) parseRepoURL(repoURL string) (string, error) {
	// Normalize URL first (handles GitHub shorthand, etc.)
	normalizedURL, err := sd.detector.NormalizeURL(repoURL)
	if err != nil {
		return "", err
	}

	// Validate the normalized repository
	if err := sd.detector.ValidateRepository(normalizedURL); err != nil {
		return "", fmt.Errorf("repository validation failed: %w", err)
	}

	return normalizedURL, nil
}

// DownloadSkill downloads a skill from the repository
func (sd *SkillDownloader) DownloadSkill(repoURL, skillName string, providedRepoPath ...string) error {
	fullURL, err := sd.parseRepoURL(repoURL)
	if err != nil {
		return err
	}

	var repoPath string
	hasProvidedPath := len(providedRepoPath) > 0 && providedRepoPath[0] != ""

	// Use provided repo path if available, otherwise clone for detection
	if hasProvidedPath {
		repoPath = providedRepoPath[0]
	} else if sd.detector.DetectProvider(repoURL) == "local" {
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
		_, err = gitpkg.CloneShallow(sd.cloner, tempDir, fullURL)
		if err != nil {
			return fmt.Errorf("failed to clone repository for detection: %w", err)
		}
		repoPath = tempDir
	}

	// Detect components in the repository
	components, err := sd.detector.DetectComponentsInRepo(repoPath)
	if err != nil {
		return fmt.Errorf("failed to detect components: %w", err)
	}

	// Filter for skill components
	var skillComponents []models.DetectedComponent
	for _, comp := range components {
		if comp.Type == models.ComponentSkill {
			skillComponents = append(skillComponents, comp)
		}
	}

	if len(skillComponents) == 0 {
		// No skill components detected, fall back to original behavior
		return sd.downloadSkillDirect(fullURL, skillName, repoURL)
	}

	// Create skill directory
	skillDir := filepath.Join(sd.baseDir, skillName)
	if err := fileutil.CreateDirectoryWithPermissions(skillDir); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	// If only one skill component found, copy its files
	if len(skillComponents) == 1 {
		component := skillComponents[0]

		// Copy component files (non-recursive) using FilePath to skill directory
		err = fileutil.CopyComponentFiles(repoPath, component, skillDir)
		if err != nil {
			os.RemoveAll(skillDir)
			return fmt.Errorf("failed to copy skill files: %w", err)
		}
	} else {
		// Multiple skills found, create a monorepo structure
		for _, component := range skillComponents {
			componentDir := filepath.Join(skillDir, component.Name)

			err = fileutil.CreateDirectoryWithPermissions(componentDir)
			if err != nil {
				continue
			}

			// Copy component files (non-recursive) using FilePath
			err = fileutil.CopyComponentFiles(repoPath, component, componentDir)
			if err != nil {
				log.Printf("Warning: failed to copy skill %s: %v", component.Name, err)
			}
		}
	}

	// Save to lock file
	var sourceType string
	if sd.detector.DetectProvider(repoURL) == "local" {
		sourceType = "local"
	} else if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	} else {
		sourceType = "github"
	}

	// Compute folder hash if it's a GitHub repo
	var folderHash string
	if sourceType == "github" {
		// Extract owner/repo from URL
		if strings.HasPrefix(fullURL, "https://github.com/") {
			ownerRepo := strings.TrimPrefix(fullURL, "https://github.com/")
			ownerRepo = strings.TrimSuffix(ownerRepo, ".git")
			if hash, err := ComputeGitHubTreeSHA(ownerRepo, "SKILL.md"); err == nil {
				folderHash = hash
			}
		}
	} else {
		// For non-GitHub repos, compute local hash
		if hash, err := ComputeLocalFolderHash(skillDir); err == nil {
			folderHash = hash
		}
	}

	if err := sd.saveLockFile(skillName, fullURL, sourceType, fullURL, folderHash, len(skillComponents), "recursive", ""); err != nil {
		log.Printf("Warning: failed to save lock file: %v", err)
	}

	// Clean up git clone only for remote repositories
	if sd.detector.DetectProvider(repoURL) != "local" {
		if _, err := os.Stat(skillDir + ".git"); err == nil {
			os.RemoveAll(skillDir + ".git")
		}
	}

	fmt.Printf("Installed: %s ✓\n", skillName)

	return nil
}

func (sd *SkillDownloader) downloadSkillDirect(fullURL, skillName, repoURL string) error {
	// Create skill directory
	skillDir := filepath.Join(sd.baseDir, skillName)
	if err := fileutil.CreateDirectoryWithPermissions(skillDir); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	// Clone repository for local or remote
	var err error

	// Handle local vs remote repositories
	if sd.detector.DetectProvider(repoURL) == "local" {
		// For local repositories, copy directory contents directly
		err = fileutil.CopyDirectoryContents(fullURL, skillDir)
		if err != nil {
			os.RemoveAll(skillDir)
			return fmt.Errorf("failed to copy local repository: %w", err)
		}
	} else {
		// For remote repositories, clone directly
		_, err = gitpkg.CloneShallow(sd.cloner, skillDir, fullURL)
		if err != nil {
			os.RemoveAll(skillDir)
			return fmt.Errorf("failed to clone repository: %w", err)
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
	if sourceType == "github" {
		// Extract owner/repo from URL
		if strings.HasPrefix(fullURL, "https://github.com/") {
			ownerRepo := strings.TrimPrefix(fullURL, "https://github.com/")
			ownerRepo = strings.TrimSuffix(ownerRepo, ".git")
			if hash, err := ComputeGitHubTreeSHA(ownerRepo, "SKILL.md"); err == nil {
				folderHash = hash
			}
		}
	} else {
		// For non-GitHub repos, compute local hash
		if hash, err := ComputeLocalFolderHash(skillDir); err == nil {
			folderHash = hash
		}
	}

	if err := sd.saveLockFile(skillName, fullURL, sourceType, fullURL, folderHash, 1, "direct", ""); err != nil {
		log.Printf("Warning: failed to save lock file: %v", err)
	}

	// Create SKILL.md if it doesn't exist
	skillFile := filepath.Join(skillDir, "SKILL.md")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		if err := sd.createSkillFile(skillFile, skillName, fullURL); err != nil {
			log.Printf("Warning: failed to create SKILL.md: %v", err)
		}
	}

	return nil
}

// saveLockFile saves component lock entry in agent-smith install compatible format
func (sd *SkillDownloader) saveLockFile(skillName string, source string, sourceType string, sourceUrl string, skillFolderHash string, components int, detection string, originalPath string) error {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return fmt.Errorf("failed to get agents directory: %w", err)
	}

	if err := fileutil.CreateDirectoryWithPermissions(agentsDir); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	return metadataPkg.SaveLockFileEntry(agentsDir, "skills", skillName, source, sourceType, sourceUrl, skillFolderHash, components, detection, originalPath)
}

func (sd *SkillDownloader) createSkillFile(filePath, skillName, source string) error {
	content := fmt.Sprintf(`# %s

Downloaded from: %s

## Description

This skill was automatically downloaded by Agent Smith.

## Usage

Add usage instructions here.

---
*Auto-generated by Agent Smith*
`, skillName, source)

	return fileutil.CreateFileWithPermissions(filePath, []byte(content))
}

// DownloadSkillWithRepo downloads a skill with repo path provided
func (sd *SkillDownloader) DownloadSkillWithRepo(fullURL, skillName, repoURL string, repoPath string, components []models.DetectedComponent) error {
	// Find the specific skill component with matching name
	var targetComponent *models.DetectedComponent
	for _, comp := range components {
		if comp.Type == models.ComponentSkill && comp.Name == skillName {
			targetComponent = &comp
			break
		}
	}

	if targetComponent == nil {
		// Skill component not found in provided components, fall back to original behavior
		return sd.downloadSkillDirect(fullURL, skillName, repoURL)
	}

	// Determine destination folder name using heuristic
	destFolderName := DetermineDestinationFolderName(targetComponent.FilePath)

	// Create skill directory with heuristic name
	skillDir := filepath.Join(sd.baseDir, destFolderName)
	if err := fileutil.CreateDirectoryWithPermissions(skillDir); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	// Copy the entire component directory recursively
	err := fileutil.CopyComponentFiles(repoPath, *targetComponent, skillDir)
	if err != nil {
		os.RemoveAll(skillDir)
		return fmt.Errorf("failed to copy skill files: %w", err)
	}

	// Save to lock file
	var sourceType string
	if sd.detector.DetectProvider(repoURL) == "local" {
		sourceType = "local"
	} else if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	} else {
		sourceType = "github"
	}

	// Compute folder hash if it's a GitHub repo
	var folderHash string
	if sourceType == "github" {
		// Extract owner/repo from URL
		if strings.HasPrefix(fullURL, "https://github.com/") {
			ownerRepo := strings.TrimPrefix(fullURL, "https://github.com/")
			ownerRepo = strings.TrimSuffix(ownerRepo, ".git")
			if hash, err := ComputeGitHubTreeSHA(ownerRepo, targetComponent.SourceFile); err == nil {
				folderHash = hash
			}
		}
	} else {
		// For non-GitHub repos, compute local hash
		if hash, err := ComputeLocalFolderHash(skillDir); err == nil {
			folderHash = hash
		}
	}

	if err := sd.saveLockFile(destFolderName, fullURL, sourceType, fullURL, folderHash, 1, "single", targetComponent.FilePath); err != nil {
		log.Printf("Warning: failed to save lock file: %v", err)
	}

	// Clean up git clone only for remote repositories
	if sd.detector.DetectProvider(repoURL) != "local" {
		if _, err := os.Stat(skillDir + ".git"); err == nil {
			os.RemoveAll(skillDir + ".git")
		}
	}

	fmt.Printf("Installed: %s ✓\n", skillName)

	return nil
}
