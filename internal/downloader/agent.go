package downloader

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjg184/agent-smith/internal/detector"
	"github.com/tjg184/agent-smith/internal/fileutil"
	gitpkg "github.com/tjg184/agent-smith/internal/git"
	metadataPkg "github.com/tjg184/agent-smith/internal/metadata"
	"github.com/tjg184/agent-smith/internal/models"
	"github.com/tjg184/agent-smith/pkg/paths"
)

// AgentDownloader handles downloading agent components
type AgentDownloader struct {
	baseDownloader
}

func NewAgentDownloader() *AgentDownloader {
	baseDir, err := paths.GetAgentsSubDir()
	if err != nil {
		log.Fatal("Failed to get agents directory:", err)
	}

	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create agents directory:", err)
	}

	return &AgentDownloader{newBaseDownloader(baseDir)}
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

	return &AgentDownloader{newBaseDownloader(baseDir)}
}

func NewAgentDownloaderWithTargetDir(targetDir string) *AgentDownloader {
	baseDir := filepath.Join(targetDir, "agents")

	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create target agents directory:", err)
	}

	return &AgentDownloader{newBaseDownloader(baseDir)}
}

func NewAgentDownloaderWithParams(baseDir string, detect *detector.RepositoryDetector) *AgentDownloader {
	return &AgentDownloader{newBaseDownloader(baseDir)}
}

// DownloadAgent downloads an agent from the repository
func (ad *AgentDownloader) DownloadAgent(repoURL, agentName string, providedRepoPath ...string) error {
	fullURL, err := ad.parseRepoURL(repoURL)
	if err != nil {
		return err
	}

	var repoPath string
	var commitHashFromRepo string
	hasProvidedPath := len(providedRepoPath) > 0 && providedRepoPath[0] != ""

	if hasProvidedPath {
		repoPath = providedRepoPath[0]
	} else if ad.detector.DetectProvider(repoURL) == "local" {
		repoPath = fullURL
	} else {
		tempDir, err := os.MkdirTemp("", "agent-smith-detect-*")
		if err != nil {
			return fmt.Errorf("failed to create temporary directory: %w", err)
		}
		defer os.RemoveAll(tempDir)

		repo, err := gitpkg.CloneShallow(ad.cloner, tempDir, fullURL)
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

	components, err := ad.detector.DetectComponentsInRepo(repoPath)
	if err != nil {
		return fmt.Errorf("failed to detect components: %w", err)
	}

	var agentComponents []models.DetectedComponent
	for _, comp := range components {
		if comp.Type == models.ComponentAgent {
			agentComponents = append(agentComponents, comp)
		}
	}

	if len(agentComponents) == 0 {
		return ad.downloadAgentDirect(fullURL, agentName)
	}

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

	if len(agentComponents) == 1 {
		component := agentComponents[0]

		err = fileutil.CopyComponentFiles(repoPath, component, agentDir)
		if err != nil {
			return fmt.Errorf("failed to copy agent files: %w", err)
		}
	} else if matchingComponent != nil {
		destFolderName := DetermineDestinationFolderName(matchingComponent.FilePath)

		if destFolderName != filesystemName {
			os.RemoveAll(agentDir)

			filesystemName = destFolderName
			agentDir = filepath.Join(ad.baseDir, filesystemName)
			if err := fileutil.CreateDirectoryWithPermissions(agentDir); err != nil {
				return fmt.Errorf("failed to create agent directory: %w", err)
			}
		}

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

	sourceType := ad.detectSourceType(fullURL)

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

	detectionType := "recursive"
	originalPath := ""
	if matchingComponent != nil && len(agentComponents) > 1 {
		detectionType = "single"
		originalPath = matchingComponent.FilePath
	}

	if err := ad.saveLockFile("agents", agentName, filesystemName, fullURL, sourceType, fullURL, commitHash, len(agentComponents), detectionType, originalPath); err != nil {
		ad.formatter.Warning("failed to save lock file: %v", err)
	}

	if _, err := os.Stat(agentDir + ".git"); err == nil {
		os.RemoveAll(agentDir + ".git")
	}

	shouldCleanup = false

	ad.formatter.Success("agent", agentName)

	return nil
}

func (ad *AgentDownloader) downloadAgentDirect(fullURL, agentName string) error {
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

	_, cloneErr := gitpkg.CloneShallow(ad.cloner, agentDir, fullURL)
	if cloneErr != nil {
		return fmt.Errorf("failed to clone repository: %w", cloneErr)
	}

	sourceType := ad.detectSourceType(fullURL)

	var commitHash string
	if hash, hashErr := gitpkg.GetCommitHashFromPath(ad.cloner, agentDir); hashErr == nil {
		commitHash = hash
	} else {
		ad.formatter.Warning("failed to get commit hash: %v", hashErr)
	}

	if err := ad.saveLockFile("agents", agentName, filesystemName, fullURL, sourceType, fullURL, commitHash, 1, "direct", ""); err != nil {
		ad.formatter.Warning("failed to save lock file: %v", err)
	}

	agentFile := filepath.Join(agentDir, agentName+".md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		if err := ad.createComponentMarkdownFile(agentFile, "agent", agentName, fullURL); err != nil {
			ad.formatter.Warning("failed to create %s.md: %v", agentName, err)
		}
	}

	shouldCleanup = false

	return nil
}

// DownloadAgentWithRepo downloads an agent with repo path provided
func (ad *AgentDownloader) DownloadAgentWithRepo(fullURL, agentName, repoURL string, repoPath string, components []models.DetectedComponent) error {
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

	sourceType := ad.detectSourceType(fullURL)

	var commitHash string
	if hash, err := gitpkg.GetCommitHashFromPath(ad.cloner, repoPath); err == nil {
		commitHash = hash
	} else {
		ad.formatter.Warning("failed to get commit hash: %v", err)
	}

	if err := ad.saveLockFile("agents", agentName, destFolderName, fullURL, sourceType, fullURL, commitHash, 1, "single", targetComponent.FilePath); err != nil {
		ad.formatter.Warning("failed to save lock file: %v", err)
	}

	if _, err := os.Stat(agentDir + ".git"); err == nil {
		os.RemoveAll(agentDir + ".git")
	}

	shouldCleanup = false

	return nil
}
