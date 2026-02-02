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
	"github.com/tgaines/agent-smith/internal/formatter"
	gitpkg "github.com/tgaines/agent-smith/internal/git"
	metadataPkg "github.com/tgaines/agent-smith/internal/metadata"
	"github.com/tgaines/agent-smith/internal/models"
	"github.com/tgaines/agent-smith/pkg/paths"
)

// CommandDownloader handles downloading command components
type CommandDownloader struct {
	baseDir   string
	detector  *detector.RepositoryDetector
	cloner    gitpkg.Cloner
	formatter *formatter.Formatter
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
		baseDir:   baseDir,
		detector:  detector.NewRepositoryDetector(),
		cloner:    gitpkg.NewDefaultCloner(),
		formatter: formatter.New(),
	}
}

// NewCommandDownloaderForProfile creates a new CommandDownloader instance that installs to a profile
func NewCommandDownloaderForProfile(profileName string) *CommandDownloader {
	// Get profiles directory
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		log.Fatal("Failed to get profiles directory:", err)
	}

	// Build path to profile's commands directory
	baseDir := filepath.Join(profilesDir, profileName, "commands")

	// Create base directory if it doesn't exist
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

// NewCommandDownloaderWithTargetDir creates a new CommandDownloader instance that installs to a custom target directory
func NewCommandDownloaderWithTargetDir(targetDir string) *CommandDownloader {
	// Build path to target directory's commands subdirectory
	baseDir := filepath.Join(targetDir, "commands")

	// Create base directory if it doesn't exist
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
		cloneOpts := &git.CloneOptions{
			URL:           fullURL,
			Depth:         1,
			ReferenceName: plumbing.HEAD,
			SingleBranch:  true,
		}

		// Add authentication if needed
		if auth, _ := gitpkg.GetAuthMethod(fullURL); auth != nil {
			cloneOpts.Auth = auth
		}

		_, err = git.PlainClone(tempDir, false, cloneOpts)
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

	// Set up cleanup on error
	shouldCleanup := true
	defer func() {
		if shouldCleanup {
			os.RemoveAll(commandDir)
		}
	}()

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
			return fmt.Errorf("failed to copy command files: %w", err)
		}
	} else if matchingComponent != nil {
		// Downloading a specific component from a multi-component directory
		// Use heuristic to determine proper folder name to avoid nested monorepo directories
		destFolderName := DetermineDestinationFolderName(matchingComponent.FilePath)

		// If heuristic name differs from commandName, recreate directory with proper name
		if destFolderName != commandName {
			commandDir = filepath.Join(cd.baseDir, destFolderName)
			if err := fileutil.CreateDirectoryWithPermissions(commandDir); err != nil {
				return fmt.Errorf("failed to create command directory: %w", err)
			}
		}

		// Copy component files (non-recursive) using FilePath to command directory
		err = fileutil.CopyComponentFiles(repoPath, *matchingComponent, commandDir)
		if err != nil {
			return fmt.Errorf("failed to copy command files: %w", err)
		}

		// Update commandName to match the actual directory name for lock file
		commandName = destFolderName
	} else {
		// Multiple commands found but none match the requested command name
		// Return error with list of available commands
		var commandNames []string
		for _, comp := range commandComponents {
			commandNames = append(commandNames, comp.Name)
		}
		return fmt.Errorf("command '%s' not found in repository. Available commands: %s", commandName, strings.Join(commandNames, ", "))
	}

	// Save to lock file
	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	// Get commit hash from repository
	var commitHash string
	if hash, err := gitpkg.GetCommitHashFromPath(cd.cloner, repoPath); err == nil {
		commitHash = hash
	} else {
		cd.formatter.Warning("failed to get commit hash: %v", err)
	}

	// Determine detection type and original path for lock file
	detectionType := "recursive"
	originalPath := ""
	if matchingComponent != nil && len(commandComponents) > 1 {
		// Single command from multi-command repo
		detectionType = "single"
		originalPath = matchingComponent.FilePath
	}

	if err := cd.saveLockFile(commandName, fullURL, sourceType, fullURL, commitHash, len(commandComponents), detectionType, originalPath); err != nil {
		cd.formatter.Warning("failed to save lock file: %v", err)
	}

	// Clean up git clone if it exists
	if _, err := os.Stat(commandDir + ".git"); err == nil {
		os.RemoveAll(commandDir + ".git")
	}

	// Success - don't clean up the directory
	shouldCleanup = false

	cd.formatter.Success("command", commandName)

	return nil
}

func (cd *CommandDownloader) downloadCommandDirect(fullURL, commandName string) error {
	// Create command directory
	commandDir := filepath.Join(cd.baseDir, commandName)
	if err := fileutil.CreateDirectoryWithPermissions(commandDir); err != nil {
		return fmt.Errorf("failed to create command directory: %w", err)
	}

	// Set up cleanup on error
	shouldCleanup := true
	defer func() {
		if shouldCleanup {
			os.RemoveAll(commandDir)
		}
	}()

	// Clone repository directly
	cloneOpts := &git.CloneOptions{
		URL:           fullURL,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	}

	// Add authentication if needed
	if auth, _ := gitpkg.GetAuthMethod(fullURL); auth != nil {
		cloneOpts.Auth = auth
	}

	_, err := git.PlainClone(commandDir, false, cloneOpts)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Save to lock file
	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	// Get commit hash from repository
	var commitHash string
	if hash, err := gitpkg.GetCommitHashFromPath(cd.cloner, commandDir); err == nil {
		commitHash = hash
	} else {
		cd.formatter.Warning("failed to get commit hash: %v", err)
	}

	if err := cd.saveLockFile(commandName, fullURL, sourceType, fullURL, commitHash, 1, "direct", ""); err != nil {
		cd.formatter.Warning("failed to save lock file: %v", err)
	}

	// Create {name}.md if it doesn't exist
	commandFile := filepath.Join(commandDir, commandName+".md")
	if _, err := os.Stat(commandFile); os.IsNotExist(err) {
		if err := cd.createCommandFile(commandFile, commandName, fullURL); err != nil {
			cd.formatter.Warning("failed to create %s.md: %v", commandName, err)
		}
	}

	// Success - don't clean up the directory
	shouldCleanup = false

	return nil
}

// saveLockFile saves command lock entry in agent-smith install compatible format
func (cd *CommandDownloader) saveLockFile(commandName string, source string, sourceType string, sourceUrl string, commitHash string, components int, detection string, originalPath string) error {
	// Use the parent directory of baseDir for lock file
	// baseDir is the commands directory (e.g., ~/.agent-smith/commands)
	// We want the lock file in the parent (e.g., ~/.agent-smith)
	lockBaseDir := filepath.Dir(cd.baseDir)

	if err := fileutil.CreateDirectoryWithPermissions(lockBaseDir); err != nil {
		return fmt.Errorf("failed to create lock file directory: %w", err)
	}

	return metadataPkg.SaveLockFileEntry(lockBaseDir, "commands", commandName, source, sourceType, sourceUrl, commitHash, components, detection, originalPath)
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

	// Set up cleanup on error
	shouldCleanup := true
	defer func() {
		if shouldCleanup {
			os.RemoveAll(commandDir)
		}
	}()

	// Copy the entire component directory recursively
	err := fileutil.CopyComponentFiles(repoPath, *targetComponent, commandDir)
	if err != nil {
		return fmt.Errorf("failed to copy command files: %w", err)
	}

	// Save to lock file
	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	// Get commit hash from repository
	var commitHash string
	if hash, err := gitpkg.GetCommitHashFromPath(cd.cloner, repoPath); err == nil {
		commitHash = hash
	} else {
		cd.formatter.Warning("failed to get commit hash: %v", err)
	}

	if err := cd.saveLockFile(destFolderName, fullURL, sourceType, fullURL, commitHash, 1, "single", targetComponent.FilePath); err != nil {
		cd.formatter.Warning("failed to save lock file: %v", err)
	}

	// Clean up git clone if it exists
	if _, err := os.Stat(commandDir + ".git"); err == nil {
		os.RemoveAll(commandDir + ".git")
	}

	// Success - don't clean up the directory
	shouldCleanup = false

	cd.formatter.Success("command", commandName)

	return nil
}
