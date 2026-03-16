package downloader

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjg184/agent-smith/internal/fileutil"
	gitpkg "github.com/tjg184/agent-smith/internal/git"
	metadataPkg "github.com/tjg184/agent-smith/internal/metadata"
	"github.com/tjg184/agent-smith/internal/models"
	"github.com/tjg184/agent-smith/pkg/paths"
)

// CommandDownloader handles downloading command components
type CommandDownloader struct {
	baseDownloader
}

func NewCommandDownloader() *CommandDownloader {
	baseDir, err := paths.GetCommandsDir()
	if err != nil {
		log.Fatal("Failed to get commands directory:", err)
	}

	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create commands directory:", err)
	}

	return &CommandDownloader{newBaseDownloader(baseDir)}
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

	return &CommandDownloader{newBaseDownloader(baseDir)}
}

func NewCommandDownloaderWithTargetDir(targetDir string) *CommandDownloader {
	baseDir := filepath.Join(targetDir, "commands")

	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create target commands directory:", err)
	}

	return &CommandDownloader{newBaseDownloader(baseDir)}
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

		_, err = gitpkg.CloneShallow(cd.cloner, tempDir, fullURL)
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
		return cd.downloadCommandDirect(fullURL, commandName)
	}

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

	sourceType := cd.detectSourceType(fullURL)

	var commitHash string
	if hash, err := gitpkg.GetCommitHashFromPath(cd.cloner, repoPath); err == nil {
		commitHash = hash
	} else {
		cd.formatter.Warning("failed to get commit hash: %v", err)
	}

	detectionType := "recursive"
	originalPath := ""
	if matchingComponent != nil && len(commandComponents) > 1 {
		detectionType = "single"
		originalPath = matchingComponent.FilePath
	}

	if err := cd.saveLockFile("commands", commandName, filesystemName, fullURL, sourceType, fullURL, commitHash, len(commandComponents), detectionType, originalPath); err != nil {
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

	_, cloneErr := gitpkg.CloneShallow(cd.cloner, commandDir, fullURL)
	if cloneErr != nil {
		return fmt.Errorf("failed to clone repository: %w", cloneErr)
	}

	sourceType := cd.detectSourceType(fullURL)

	var commitHash string
	if hash, hashErr := gitpkg.GetCommitHashFromPath(cd.cloner, commandDir); hashErr == nil {
		commitHash = hash
	} else {
		cd.formatter.Warning("failed to get commit hash: %v", hashErr)
	}

	if err := cd.saveLockFile("commands", commandName, filesystemName, fullURL, sourceType, fullURL, commitHash, 1, "direct", ""); err != nil {
		cd.formatter.Warning("failed to save lock file: %v", err)
	}

	commandFile := filepath.Join(commandDir, commandName+".md")
	if _, err := os.Stat(commandFile); os.IsNotExist(err) {
		if err := cd.createComponentMarkdownFile(commandFile, "command", commandName, fullURL); err != nil {
			cd.formatter.Warning("failed to create %s.md: %v", commandName, err)
		}
	}

	shouldCleanup = false

	return nil
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

	sourceType := cd.detectSourceType(fullURL)

	var commitHash string
	if hash, err := gitpkg.GetCommitHashFromPath(cd.cloner, repoPath); err == nil {
		commitHash = hash
	} else {
		cd.formatter.Warning("failed to get commit hash: %v", err)
	}

	if err := cd.saveLockFile("commands", commandName, destFolderName, fullURL, sourceType, fullURL, commitHash, 1, "single", targetComponent.FilePath); err != nil {
		cd.formatter.Warning("failed to save lock file: %v", err)
	}

	if _, err := os.Stat(commandDir + ".git"); err == nil {
		os.RemoveAll(commandDir + ".git")
	}

	shouldCleanup = false

	return nil
}
