package downloader

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjg184/agent-smith/internal/detector"
	"github.com/tjg184/agent-smith/internal/fileutil"
	"github.com/tjg184/agent-smith/internal/formatter"
	gitpkg "github.com/tjg184/agent-smith/internal/git"
	metadataPkg "github.com/tjg184/agent-smith/internal/metadata"
	"github.com/tjg184/agent-smith/internal/models"
	"github.com/tjg184/agent-smith/pkg/paths"
)

// SkillDownloader handles downloading skill components
type SkillDownloader struct {
	baseDir   string
	detector  *detector.RepositoryDetector
	cloner    gitpkg.Cloner
	formatter *formatter.Formatter
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
		baseDir:   baseDir,
		detector:  detector.NewRepositoryDetector(),
		cloner:    gitpkg.NewDefaultCloner(),
		formatter: formatter.New(),
	}
}

// NewSkillDownloaderForProfile creates a new SkillDownloader instance that installs to a profile
func NewSkillDownloaderForProfile(profileName string) *SkillDownloader {
	// Get profiles directory
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		log.Fatal("Failed to get profiles directory:", err)
	}

	// Build path to profile's skills directory
	baseDir := filepath.Join(profilesDir, profileName, "skills")

	// Create base directory if it doesn't exist
	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create profile skills directory:", err)
	}

	return &SkillDownloader{
		baseDir:   baseDir,
		detector:  detector.NewRepositoryDetector(),
		cloner:    gitpkg.NewDefaultCloner(),
		formatter: formatter.New(),
	}
}

// NewSkillDownloaderWithTargetDir creates a new SkillDownloader instance that installs to a custom target directory
func NewSkillDownloaderWithTargetDir(targetDir string) *SkillDownloader {
	// Build path to target directory's skills subdirectory
	baseDir := filepath.Join(targetDir, "skills")

	// Create base directory if it doesn't exist
	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create target skills directory:", err)
	}

	return &SkillDownloader{
		baseDir:   baseDir,
		detector:  detector.NewRepositoryDetector(),
		cloner:    gitpkg.NewDefaultCloner(),
		formatter: formatter.New(),
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
	var commitHashFromRepo string // Store commit hash from clone
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
		repo, err := gitpkg.CloneShallow(sd.cloner, tempDir, fullURL)
		if err != nil {
			return fmt.Errorf("failed to clone repository for detection: %w", err)
		}
		repoPath = tempDir

		// Get commit hash from the cloned repository
		if repo != nil {
			ref, err := repo.Head()
			if err == nil {
				commitHashFromRepo = ref.Hash().String()
			}
		}
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

	// Resolve filesystem name before creating directory to handle conflicts
	lockBaseDir := filepath.Dir(sd.baseDir)
	filesystemName, err := metadataPkg.ResolveInstallFilesystemName(lockBaseDir, "skills", skillName, fullURL)
	if err != nil {
		sd.formatter.Warning("failed to resolve filesystem name, using skill name: %v", err)
		filesystemName = skillName
	}

	// Create skill directory with resolved name
	skillDir := filepath.Join(sd.baseDir, filesystemName)
	if err := fileutil.CreateDirectoryWithPermissions(skillDir); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	// Set up cleanup on error
	shouldCleanup := true
	defer func() {
		if shouldCleanup {
			os.RemoveAll(skillDir)
		}
	}()

	// Check if the requested skillName matches one of the detected components
	var matchingComponent *models.DetectedComponent
	for _, comp := range skillComponents {
		if comp.Name == skillName {
			matchingComponent = &comp
			break
		}
	}

	// Additional check: if we found a matching component but it's part of a larger directory structure,
	// prefer components that have their own directory (more specific)
	if matchingComponent != nil && len(skillComponents) > 1 {
		for _, comp := range skillComponents {
			if comp.Name == skillName && comp.Path != matchingComponent.Path {
				// Found a more specific version (different path)
				matchingComponent = &comp
				break
			}
		}
	}

	// If only one skill component found, copy its contents
	if len(skillComponents) == 1 {
		component := skillComponents[0]

		// Copy component files (non-recursive) using FilePath to skill directory
		err = fileutil.CopyComponentFiles(repoPath, component, skillDir)
		if err != nil {
			return fmt.Errorf("failed to copy skill files: %w", err)
		}
	} else if matchingComponent != nil {
		// Downloading a specific component from a multi-component directory
		// Use heuristic to determine proper folder name to avoid nested monorepo directories
		destFolderName := DetermineDestinationFolderName(matchingComponent.FilePath)

		// If heuristic name differs from resolved filesystem name, update it
		// (This can happen when the heuristic produces a different name than what was requested)
		if destFolderName != filesystemName {
			// Remove the originally created directory
			os.RemoveAll(skillDir)

			// Recreate with heuristic name
			filesystemName = destFolderName
			skillDir = filepath.Join(sd.baseDir, filesystemName)
			if err := fileutil.CreateDirectoryWithPermissions(skillDir); err != nil {
				return fmt.Errorf("failed to create skill directory: %w", err)
			}
		}

		// Copy component files (non-recursive) using FilePath to skill directory
		err = fileutil.CopyComponentFiles(repoPath, *matchingComponent, skillDir)
		if err != nil {
			return fmt.Errorf("failed to copy skill files: %w", err)
		}
	} else {
		// Multiple skills found but none match the requested skill name
		// Return error with list of available skills
		var skillNames []string
		for _, comp := range skillComponents {
			skillNames = append(skillNames, comp.Name)
		}
		return fmt.Errorf("skill '%s' not found in repository. Available skills: %s", skillName, strings.Join(skillNames, ", "))
	}

	// Save to lock file
	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	// Get commit hash from repository (already retrieved during clone for remote repos)
	var commitHash string
	if hasProvidedPath || sd.detector.DetectProvider(repoURL) == "local" {
		// For provided path or local repos, try to get hash from path
		if hash, err := gitpkg.GetCommitHashFromPath(sd.cloner, repoPath); err == nil {
			commitHash = hash
		} else {
			sd.formatter.Warning("failed to get commit hash: %v", err)
		}
	} else {
		// For remote repos, use the hash we got during clone
		commitHash = commitHashFromRepo
	}

	// Determine detection type and original path for lock file
	detectionType := "recursive"
	originalPath := ""
	if matchingComponent != nil && len(skillComponents) > 1 {
		// Single skill from multi-skill repo
		detectionType = "single"
		originalPath = matchingComponent.FilePath
	}

	if err := sd.saveLockFile(skillName, filesystemName, fullURL, sourceType, fullURL, commitHash, len(skillComponents), detectionType, originalPath); err != nil {
		sd.formatter.Warning("failed to save lock file: %v", err)
	}

	// Clean up git clone if it exists
	if _, err := os.Stat(skillDir + ".git"); err == nil {
		os.RemoveAll(skillDir + ".git")
	}

	// Success - don't clean up the directory
	shouldCleanup = false

	sd.formatter.Success("skill", skillName)

	return nil
}

func (sd *SkillDownloader) downloadSkillDirect(fullURL, skillName, repoURL string) error {
	// Resolve filesystem name before creating directory to handle conflicts
	lockBaseDir := filepath.Dir(sd.baseDir)
	filesystemName, err := metadataPkg.ResolveInstallFilesystemName(lockBaseDir, "skills", skillName, fullURL)
	if err != nil {
		sd.formatter.Warning("failed to resolve filesystem name, using skill name: %v", err)
		filesystemName = skillName
	}

	// Create skill directory with resolved name
	skillDir := filepath.Join(sd.baseDir, filesystemName)
	if err := fileutil.CreateDirectoryWithPermissions(skillDir); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	// Set up cleanup on error
	shouldCleanup := true
	defer func() {
		if shouldCleanup {
			os.RemoveAll(skillDir)
		}
	}()

	// Clone repository for local or remote
	var cloneErr error

	// Handle local vs remote repositories
	if sd.detector.DetectProvider(repoURL) == "local" {
		// For local repositories, copy directory contents directly
		cloneErr = fileutil.CopyDirectoryContents(fullURL, skillDir)
		if cloneErr != nil {
			return fmt.Errorf("failed to copy local repository: %w", cloneErr)
		}
	} else {
		// For remote repositories, clone directly
		_, cloneErr = gitpkg.CloneShallow(sd.cloner, skillDir, fullURL)
		if cloneErr != nil {
			return fmt.Errorf("failed to clone repository: %w", cloneErr)
		}
	}

	// Save to lock file
	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	// Get commit hash from the repository
	var commitHash string
	if hash, hashErr := gitpkg.GetCommitHashFromPath(sd.cloner, skillDir); hashErr == nil {
		commitHash = hash
	} else {
		sd.formatter.Warning("failed to get commit hash: %v", hashErr)
	}

	if err := sd.saveLockFile(skillName, filesystemName, fullURL, sourceType, fullURL, commitHash, 1, "direct", ""); err != nil {
		sd.formatter.Warning("failed to save lock file: %v", err)
	}

	// Create SKILL.md if it doesn't exist
	skillFile := filepath.Join(skillDir, "SKILL.md")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		if err := sd.createSkillFile(skillFile, skillName, fullURL); err != nil {
			sd.formatter.Warning("failed to create SKILL.md: %v", err)
		}
	}

	// Success - don't clean up the directory
	shouldCleanup = false

	return nil
}

// saveLockFile saves component lock entry in agent-smith install compatible format
func (sd *SkillDownloader) saveLockFile(skillName, filesystemName, source, sourceType, sourceUrl, commitHash string, components int, detection, originalPath string) error {
	// Use the parent directory of baseDir for lock file
	// baseDir is the skills directory (e.g., ~/.agent-smith/skills)
	// We want the lock file in the parent (e.g., ~/.agent-smith)
	lockBaseDir := filepath.Dir(sd.baseDir)

	if err := fileutil.CreateDirectoryWithPermissions(lockBaseDir); err != nil {
		return fmt.Errorf("failed to create lock file directory: %w", err)
	}

	// Calculate hashes for drift detection
	// Both sourceHash and currentHash use local filesystem hashing
	// They should match at install time (no modifications yet)
	var sourceHash, currentHash string
	skillDir := filepath.Join(sd.baseDir, filesystemName)

	if hash, err := metadataPkg.ComputeLocalFolderHash(skillDir); err == nil {
		sourceHash = hash
		currentHash = hash
	} else {
		// Only warn if we can't hash at all (rare - filesystem issue)
		sd.formatter.Warning("failed to compute hash: %v", err)
	}

	return metadataPkg.SaveComponentEntry(lockBaseDir, "skills", skillName, source, sourceType, sourceUrl, commitHash, originalPath, metadataPkg.ComponentEntryOptions{
		UpdatedAt:      "", // Will be set by SaveComponentEntry
		Components:     components,
		Detection:      detection,
		SourceHash:     sourceHash,
		CurrentHash:    currentHash,
		FilesystemName: filesystemName,
	})
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

	// Set up cleanup on error
	shouldCleanup := true
	defer func() {
		if shouldCleanup {
			os.RemoveAll(skillDir)
		}
	}()

	// Copy the entire component directory recursively
	err := fileutil.CopyComponentFiles(repoPath, *targetComponent, skillDir)
	if err != nil {
		return fmt.Errorf("failed to copy skill files: %w", err)
	}

	// Save to lock file
	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	// Get commit hash from the repository
	var commitHash string
	if hash, err := gitpkg.GetCommitHashFromPath(sd.cloner, repoPath); err == nil {
		commitHash = hash
	} else {
		sd.formatter.Warning("failed to get commit hash: %v", err)
	}

	if err := sd.saveLockFile(skillName, destFolderName, fullURL, sourceType, fullURL, commitHash, 1, "single", targetComponent.FilePath); err != nil {
		sd.formatter.Warning("failed to save lock file: %v", err)
	}

	// Clean up git clone if it exists
	if _, err := os.Stat(skillDir + ".git"); err == nil {
		os.RemoveAll(skillDir + ".git")
	}

	// Success - don't clean up the directory
	shouldCleanup = false

	sd.formatter.Success("skill", skillName)

	return nil
}
