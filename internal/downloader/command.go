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

// CommandDownloader handles downloading command components
type CommandDownloader struct {
	baseDir   string
	detector  *detector.RepositoryDetector
	cloner    gitpkg.Cloner
	formatter *formatter.Formatter
}

func NewCommandDownloader() *CommandDownloader {
	baseDir, err := paths.GetCommandsDir()
	if err != nil {
		log.Fatal("Failed to get commands directory:", err)
	}

	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create commands directory:", err)
	}

	return &CommandDownloader{
		baseDir:   baseDir,
		detector:  detector.NewRepositoryDetector(),
		cloner:    gitpkg.NewDefaultCloner(),
		formatter: formatter.New(),
	}
}

func NewCommandDownloaderForProfile(profileName string) *CommandDownloader {
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		log.Fatal("Failed to get profiles directory:", err)
	}

	baseDir := filepath.Join(profilesDir, profileName, "commands")

	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create profile commands directory:", err)
	}

	return &CommandDownloader{
		baseDir:   baseDir,
		detector:  detector.NewRepositoryDetector(),
		cloner:    gitpkg.NewDefaultCloner(),
		formatter: formatter.New(),
	}
}

func NewCommandDownloaderWithTargetDir(targetDir string) *CommandDownloader {
	baseDir := filepath.Join(targetDir, "commands")

	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create target commands directory:", err)
	}

	return &CommandDownloader{
		baseDir:   baseDir,
		detector:  detector.NewRepositoryDetector(),
		cloner:    gitpkg.NewDefaultCloner(),
		formatter: formatter.New(),
	}
}

func (cd *CommandDownloader) parseRepoURL(repoURL string) (string, error) {
	normalizedURL, err := cd.detector.NormalizeURL(repoURL)
	if err != nil {
		return "", err
	}

	if err := cd.detector.ValidateRepository(normalizedURL); err != nil {
		return "", fmt.Errorf("repository validation failed: %w", err)
	}

	return normalizedURL, nil
}

func (cd *CommandDownloader) DownloadCommand(repoURL, commandName string, providedRepoPath ...string) error {
	fullURL, err := cd.parseRepoURL(repoURL)
	if err != nil {
		return err
	}

	var repoPath string
	hasProvidedPath := len(providedRepoPath) > 0 && providedRepoPath[0] != ""

	if hasProvidedPath {
		repoPath = providedRepoPath[0]
	} else if cd.detector.DetectProvider(repoURL) == "local" {
		repoPath = fullURL
	} else {
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

		_, err = git.PlainClone(tempDir, false, cloneOpts)
		if err != nil {
			return fmt.Errorf("failed to clone repository for detection: %w", err)
		}
		repoPath = tempDir
	}

	components, err := cd.detector.DetectComponentsInRepo(repoPath)
	if err != nil {
		return fmt.Errorf("failed to detect components: %w", err)
	}

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

	// Resolve filesystem name before creating directory to handle conflicts
	lockBaseDir := filepath.Dir(cd.baseDir)
	filesystemName, err := metadataPkg.ResolveInstallFilesystemName(lockBaseDir, "commands", commandName, fullURL)
	if err != nil {
		cd.formatter.Warning("failed to resolve filesystem name, using command name: %v", err)
		filesystemName = commandName
	}

	commandDir := filepath.Join(cd.baseDir, filesystemName)
	if err := fileutil.CreateDirectoryWithPermissions(commandDir); err != nil {
		return fmt.Errorf("failed to create command directory: %w", err)
	}

	shouldCleanup := true
	defer func() {
		if shouldCleanup {
			os.RemoveAll(commandDir)
		}
	}()

	var matchingComponent *models.DetectedComponent
	for _, comp := range commandComponents {
		if comp.Name == commandName {
			matchingComponent = &comp
			break
		}
	}

	if matchingComponent != nil && len(commandComponents) > 1 {
		for _, comp := range commandComponents {
			if comp.Name == commandName && comp.Path != matchingComponent.Path {
				matchingComponent = &comp
				break
			}
		}
	}

	if len(commandComponents) == 1 {
		component := commandComponents[0]

		err = fileutil.CopyComponentFiles(repoPath, component, commandDir)
		if err != nil {
			return fmt.Errorf("failed to copy command files: %w", err)
		}
	} else if matchingComponent != nil {
		// Use heuristic to determine proper folder name to avoid nested monorepo directories
		destFolderName := DetermineDestinationFolderName(matchingComponent.FilePath)

		if destFolderName != filesystemName {
			os.RemoveAll(commandDir)

			filesystemName = destFolderName
			commandDir = filepath.Join(cd.baseDir, filesystemName)
			if err := fileutil.CreateDirectoryWithPermissions(commandDir); err != nil {
				return fmt.Errorf("failed to create command directory: %w", err)
			}
		}

		err = fileutil.CopyComponentFiles(repoPath, *matchingComponent, commandDir)
		if err != nil {
			return fmt.Errorf("failed to copy command files: %w", err)
		}
	} else {
		var commandNames []string
		for _, comp := range commandComponents {
			commandNames = append(commandNames, comp.Name)
		}
		return fmt.Errorf("command '%s' not found in repository. Available commands: %s", commandName, strings.Join(commandNames, ", "))
	}

	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	var commitHash string
	if hash, err := gitpkg.GetCommitHashFromPath(cd.cloner, repoPath); err == nil {
		commitHash = hash
	} else {
		cd.formatter.Warning("failed to get commit hash: %v", err)
	}

	detectionType := "recursive"
	originalPath := ""
	if matchingComponent != nil && len(commandComponents) > 1 {
		// Single command from multi-command repo
		detectionType = "single"
		originalPath = matchingComponent.FilePath
	}

	if err := cd.saveLockFile(commandName, filesystemName, fullURL, sourceType, fullURL, commitHash, len(commandComponents), detectionType, originalPath); err != nil {
		cd.formatter.Warning("failed to save lock file: %v", err)
	}

	if _, err := os.Stat(commandDir + ".git"); err == nil {
		os.RemoveAll(commandDir + ".git")
	}

	shouldCleanup = false

	cd.formatter.Success("command", commandName)

	return nil
}

func (cd *CommandDownloader) downloadCommandDirect(fullURL, commandName string) error {
	lockBaseDir := filepath.Dir(cd.baseDir)
	filesystemName, err := metadataPkg.ResolveInstallFilesystemName(lockBaseDir, "commands", commandName, fullURL)
	if err != nil {
		cd.formatter.Warning("failed to resolve filesystem name, using command name: %v", err)
		filesystemName = commandName
	}

	commandDir := filepath.Join(cd.baseDir, filesystemName)
	if err := fileutil.CreateDirectoryWithPermissions(commandDir); err != nil {
		return fmt.Errorf("failed to create command directory: %w", err)
	}

	shouldCleanup := true
	defer func() {
		if shouldCleanup {
			os.RemoveAll(commandDir)
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

	_, cloneErr := git.PlainClone(commandDir, false, cloneOpts)
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
	if hash, hashErr := gitpkg.GetCommitHashFromPath(cd.cloner, commandDir); hashErr == nil {
		commitHash = hash
	} else {
		cd.formatter.Warning("failed to get commit hash: %v", hashErr)
	}

	if err := cd.saveLockFile(commandName, filesystemName, fullURL, sourceType, fullURL, commitHash, 1, "direct", ""); err != nil {
		cd.formatter.Warning("failed to save lock file: %v", err)
	}

	commandFile := filepath.Join(commandDir, commandName+".md")
	if _, err := os.Stat(commandFile); os.IsNotExist(err) {
		if err := cd.createCommandFile(commandFile, commandName, fullURL); err != nil {
			cd.formatter.Warning("failed to create %s.md: %v", commandName, err)
		}
	}

	shouldCleanup = false

	return nil
}

// saveLockFile saves command lock entry in agent-smith install compatible format
func (cd *CommandDownloader) saveLockFile(commandName, filesystemName, source, sourceType, sourceUrl, commitHash string, components int, detection, originalPath string) error {
	// Use the parent directory of baseDir for lock file
	// baseDir is the commands directory (e.g., ~/.agent-smith/commands)
	// We want the lock file in the parent (e.g., ~/.agent-smith)
	lockBaseDir := filepath.Dir(cd.baseDir)

	if err := fileutil.CreateDirectoryWithPermissions(lockBaseDir); err != nil {
		return fmt.Errorf("failed to create lock file directory: %w", err)
	}

	// Calculate hashes for drift detection
	// Both sourceHash and currentHash use local filesystem hashing
	// They should match at install time (no modifications yet)
	var sourceHash, currentHash string
	commandDir := filepath.Join(cd.baseDir, filesystemName)

	if hash, err := metadataPkg.ComputeLocalFolderHash(commandDir); err == nil {
		sourceHash = hash
		currentHash = hash
	} else {
		// Only warn if we can't hash at all (rare - filesystem issue)
		cd.formatter.Warning("failed to compute hash: %v", err)
	}

	return metadataPkg.SaveComponentEntry(lockBaseDir, "commands", commandName, source, sourceType, sourceUrl, commitHash, originalPath, metadataPkg.ComponentEntryOptions{
		UpdatedAt:      "", // Will be set by SaveComponentEntry
		Components:     components,
		Detection:      detection,
		SourceHash:     sourceHash,
		CurrentHash:    currentHash,
		FilesystemName: filesystemName,
	})
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

func (cd *CommandDownloader) DownloadCommandWithRepo(fullURL, commandName, repoURL string, repoPath string, components []models.DetectedComponent) error {
	var targetComponent *models.DetectedComponent
	for _, comp := range components {
		if comp.Type == models.ComponentCommand && comp.Name == commandName {
			targetComponent = &comp
			break
		}
	}

	if targetComponent == nil {
		return cd.downloadCommandDirect(fullURL, commandName)
	}

	// Use heuristic to determine proper folder name to avoid nested monorepo directories
	destFolderName := DetermineDestinationFolderName(targetComponent.FilePath)

	commandDir := filepath.Join(cd.baseDir, destFolderName)
	if err := fileutil.CreateDirectoryWithPermissions(commandDir); err != nil {
		return fmt.Errorf("failed to create command directory: %w", err)
	}

	shouldCleanup := true
	defer func() {
		if shouldCleanup {
			os.RemoveAll(commandDir)
		}
	}()

	err := fileutil.CopyComponentFiles(repoPath, *targetComponent, commandDir)
	if err != nil {
		return fmt.Errorf("failed to copy command files: %w", err)
	}

	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	var commitHash string
	if hash, err := gitpkg.GetCommitHashFromPath(cd.cloner, repoPath); err == nil {
		commitHash = hash
	} else {
		cd.formatter.Warning("failed to get commit hash: %v", err)
	}

	if err := cd.saveLockFile(commandName, destFolderName, fullURL, sourceType, fullURL, commitHash, 1, "single", targetComponent.FilePath); err != nil {
		cd.formatter.Warning("failed to save lock file: %v", err)
	}

	if _, err := os.Stat(commandDir + ".git"); err == nil {
		os.RemoveAll(commandDir + ".git")
	}

	shouldCleanup = false

	return nil
}
