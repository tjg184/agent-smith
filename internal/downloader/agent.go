package downloader

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/tjg184/agent-smith/internal/detector"
	"github.com/tjg184/agent-smith/internal/fileutil"
	"github.com/tjg184/agent-smith/internal/formatter"
	gitpkg "github.com/tjg184/agent-smith/internal/git"
	metadataPkg "github.com/tjg184/agent-smith/internal/metadata"
	"github.com/tjg184/agent-smith/internal/models"
	"github.com/tjg184/agent-smith/pkg/paths"
)

// AgentDownloader handles downloading agent components
type AgentDownloader struct {
	baseDir   string
	detector  *detector.RepositoryDetector
	cloner    gitpkg.Cloner
	formatter *formatter.Formatter
}

func NewAgentDownloader() *AgentDownloader {
	baseDir, err := paths.GetAgentsSubDir()
	if err != nil {
		log.Fatal("Failed to get agents directory:", err)
	}

	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create agents directory:", err)
	}

	return &AgentDownloader{
		baseDir:   baseDir,
		detector:  detector.NewRepositoryDetector(),
		cloner:    gitpkg.NewDefaultCloner(),
		formatter: formatter.New(),
	}
}

func NewAgentDownloaderForProfile(profileName string) *AgentDownloader {
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		log.Fatal("Failed to get profiles directory:", err)
	}

	baseDir := filepath.Join(profilesDir, profileName, "agents")

	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create profile agents directory:", err)
	}

	return &AgentDownloader{
		baseDir:   baseDir,
		detector:  detector.NewRepositoryDetector(),
		cloner:    gitpkg.NewDefaultCloner(),
		formatter: formatter.New(),
	}
}

func NewAgentDownloaderWithTargetDir(targetDir string) *AgentDownloader {
	baseDir := filepath.Join(targetDir, "agents")

	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create target agents directory:", err)
	}

	return &AgentDownloader{
		baseDir:   baseDir,
		detector:  detector.NewRepositoryDetector(),
		cloner:    gitpkg.NewDefaultCloner(),
		formatter: formatter.New(),
	}
}

func NewAgentDownloaderWithParams(baseDir string, detect *detector.RepositoryDetector) *AgentDownloader {
	return &AgentDownloader{
		baseDir:   baseDir,
		detector:  detector.NewRepositoryDetector(),
		cloner:    gitpkg.NewDefaultCloner(),
		formatter: formatter.New(),
	}
}

func (ad *AgentDownloader) parseRepoURL(repoURL string) (string, error) {
	// Normalize URL first (handles GitHub shorthand, etc.)
	normalizedURL, err := ad.detector.NormalizeURL(repoURL)
	if err != nil {
		return "", err
	}

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
	var commitHashFromRepo string // Store commit hash from clone
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

		cloneOpts := &git.CloneOptions{
			URL:           fullURL,
			Depth:         1,
			ReferenceName: plumbing.HEAD,
			SingleBranch:  true,
		}

		if auth, _ := gitpkg.GetAuthMethod(fullURL); auth != nil {
			cloneOpts.Auth = auth
		}

		repo, err := git.PlainClone(tempDir, false, cloneOpts)
		if err != nil {
			return fmt.Errorf("failed to clone repository for detection: %w", err)
		}
		repoPath = tempDir

		ref, err := repo.Head()
		if err != nil {
			return fmt.Errorf("failed to get HEAD reference: %w", err)
		}
		commitHashFromRepo = ref.Hash().String()
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

	// Resolve filesystem name before creating directory to handle conflicts
	lockBaseDir := filepath.Dir(ad.baseDir)
	filesystemName, err := metadataPkg.ResolveInstallFilesystemName(lockBaseDir, "agents", agentName, fullURL)
	if err != nil {
		ad.formatter.Warning("failed to resolve filesystem name, using agent name: %v", err)
		filesystemName = agentName
	}

	agentDir := filepath.Join(ad.baseDir, filesystemName)
	if err := fileutil.CreateDirectoryWithPermissions(agentDir); err != nil {
		return fmt.Errorf("failed to create agent directory: %w", err)
	}

	shouldCleanup := true
	defer func() {
		if shouldCleanup {
			os.RemoveAll(agentDir)
		}
	}()

	var matchingComponent *models.DetectedComponent
	for _, comp := range agentComponents {
		if comp.Name == agentName {
			matchingComponent = &comp
			break
		}
	}

	if matchingComponent != nil && len(agentComponents) > 1 {
		for _, comp := range agentComponents {
			if comp.Name == agentName && comp.Path != matchingComponent.Path {
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
			return fmt.Errorf("failed to copy agent files: %w", err)
		}
	} else if matchingComponent != nil {
		// Downloading a specific component from a multi-component directory
		// Use heuristic to determine proper folder name to avoid nested monorepo directories
		destFolderName := DetermineDestinationFolderName(matchingComponent.FilePath)

		// If heuristic name differs from resolved filesystem name, update it
		if destFolderName != filesystemName {
			// Remove the originally created directory
			os.RemoveAll(agentDir)

			// Recreate with heuristic name
			filesystemName = destFolderName
			agentDir = filepath.Join(ad.baseDir, filesystemName)
			if err := fileutil.CreateDirectoryWithPermissions(agentDir); err != nil {
				return fmt.Errorf("failed to create agent directory: %w", err)
			}
		}

		// Copy component files (non-recursive) using FilePath to agent directory
		err = fileutil.CopyComponentFiles(repoPath, *matchingComponent, agentDir)
		if err != nil {
			return fmt.Errorf("failed to copy agent files: %w", err)
		}
	} else {
		var agentNames []string
		for _, comp := range agentComponents {
			agentNames = append(agentNames, comp.Name)
		}
		return fmt.Errorf("agent '%s' not found in repository. Available agents: %s", agentName, strings.Join(agentNames, ", "))
	}

	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	var commitHash string
	if hasProvidedPath || ad.detector.DetectProvider(repoURL) == "local" {
		if hash, err := gitpkg.GetCommitHashFromPath(ad.cloner, repoPath); err == nil {
			commitHash = hash
		} else {
			ad.formatter.Warning("failed to get commit hash: %v", err)
		}
	} else {
		commitHash = commitHashFromRepo
	}

	// Determine detection type and original path for lock file
	detectionType := "recursive"
	originalPath := ""
	if matchingComponent != nil && len(agentComponents) > 1 {
		// Single agent from multi-agent repo
		detectionType = "single"
		originalPath = matchingComponent.FilePath
	}

	if err := ad.saveLockFile(agentName, filesystemName, fullURL, sourceType, fullURL, commitHash, len(agentComponents), detectionType, originalPath); err != nil {
		ad.formatter.Warning("failed to save lock file: %v", err)
	}

	// Clean up git clone if it exists
	if _, err := os.Stat(agentDir + ".git"); err == nil {
		os.RemoveAll(agentDir + ".git")
	}

	// Success - don't clean up the directory
	shouldCleanup = false

	ad.formatter.Success("agent", agentName)

	return nil
}

func (ad *AgentDownloader) downloadAgentDirect(fullURL, agentName string) error {
	// Resolve filesystem name before creating directory to handle conflicts
	lockBaseDir := filepath.Dir(ad.baseDir)
	filesystemName, err := metadataPkg.ResolveInstallFilesystemName(lockBaseDir, "agents", agentName, fullURL)
	if err != nil {
		ad.formatter.Warning("failed to resolve filesystem name, using agent name: %v", err)
		filesystemName = agentName
	}

	agentDir := filepath.Join(ad.baseDir, filesystemName)
	if err := fileutil.CreateDirectoryWithPermissions(agentDir); err != nil {
		return fmt.Errorf("failed to create agent directory: %w", err)
	}

	shouldCleanup := true
	defer func() {
		if shouldCleanup {
			os.RemoveAll(agentDir)
		}
	}()

	cloneOpts := &git.CloneOptions{
		URL:           fullURL,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	}

	if auth, _ := gitpkg.GetAuthMethod(fullURL); auth != nil {
		cloneOpts.Auth = auth
	}

	_, cloneErr := git.PlainClone(agentDir, false, cloneOpts)
	if cloneErr != nil {
		return fmt.Errorf("failed to clone repository: %w", cloneErr)
	}

	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	var commitHash string
	if hash, hashErr := gitpkg.GetCommitHashFromPath(ad.cloner, agentDir); hashErr == nil {
		commitHash = hash
	} else {
		ad.formatter.Warning("failed to get commit hash: %v", hashErr)
	}

	if err := ad.saveLockFile(agentName, filesystemName, fullURL, sourceType, fullURL, commitHash, 1, "direct", ""); err != nil {
		ad.formatter.Warning("failed to save lock file: %v", err)
	}

	agentFile := filepath.Join(agentDir, agentName+".md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		if err := ad.createAgentFile(agentFile, agentName, fullURL); err != nil {
			ad.formatter.Warning("failed to create %s.md: %v", agentName, err)
		}
	}

	shouldCleanup = false

	return nil
}

// saveLockFile saves agent lock entry in agent-smith install compatible format
func (ad *AgentDownloader) saveLockFile(agentName, filesystemName, source, sourceType, sourceUrl, commitHash string, components int, detection, originalPath string) error {
	// Use the parent directory of baseDir for lock file
	// baseDir is the agents directory (e.g., ~/.agent-smith/agents)
	// We want the lock file in the parent (e.g., ~/.agent-smith)
	lockBaseDir := filepath.Dir(ad.baseDir)

	if err := fileutil.CreateDirectoryWithPermissions(lockBaseDir); err != nil {
		return fmt.Errorf("failed to create lock file directory: %w", err)
	}

	// Calculate hashes for drift detection
	// Both sourceHash and currentHash use local filesystem hashing
	// They should match at install time (no modifications yet)
	var sourceHash, currentHash string
	agentDir := filepath.Join(ad.baseDir, filesystemName)

	if hash, err := metadataPkg.ComputeLocalFolderHash(agentDir); err == nil {
		sourceHash = hash
		currentHash = hash
	} else {
		// Only warn if we can't hash at all (rare - filesystem issue)
		ad.formatter.Warning("failed to compute hash: %v", err)
	}

	return metadataPkg.SaveComponentEntry(lockBaseDir, "agents", agentName, source, sourceType, sourceUrl, commitHash, originalPath, metadataPkg.ComponentEntryOptions{
		UpdatedAt:      "", // Will be set by SaveComponentEntry
		Components:     components,
		Detection:      detection,
		SourceHash:     sourceHash,
		CurrentHash:    currentHash,
		FilesystemName: filesystemName,
	})
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
		return ad.downloadAgentDirect(fullURL, agentName)
	}

	// Use heuristic to determine proper folder name to avoid nested monorepo directories
	destFolderName := DetermineDestinationFolderName(targetComponent.FilePath)

	agentDir := filepath.Join(ad.baseDir, destFolderName)
	if err := fileutil.CreateDirectoryWithPermissions(agentDir); err != nil {
		return fmt.Errorf("failed to create agent directory: %w", err)
	}

	shouldCleanup := true
	defer func() {
		if shouldCleanup {
			os.RemoveAll(agentDir)
		}
	}()

	if err := fileutil.CopyComponentFiles(repoPath, *targetComponent, agentDir); err != nil {
		return fmt.Errorf("failed to copy agent files: %w", err)
	}

	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	var commitHash string
	if hash, err := gitpkg.GetCommitHashFromPath(ad.cloner, repoPath); err == nil {
		commitHash = hash
	} else {
		ad.formatter.Warning("failed to get commit hash: %v", err)
	}

	if err := ad.saveLockFile(agentName, destFolderName, fullURL, sourceType, fullURL, commitHash, 1, "single", targetComponent.FilePath); err != nil {
		ad.formatter.Warning("failed to save lock file: %v", err)
	}

	// Clean up git clone if it exists
	if _, err := os.Stat(agentDir + ".git"); err == nil {
		os.RemoveAll(agentDir + ".git")
	}

	// Success - don't clean up the directory
	shouldCleanup = false

	return nil
}
