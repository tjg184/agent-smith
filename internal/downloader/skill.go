package downloader

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	gitpkg "github.com/tjg184/agent-smith/internal/git"
	metadataPkg "github.com/tjg184/agent-smith/internal/metadata"
	"github.com/tjg184/agent-smith/internal/fileutil"
	"github.com/tjg184/agent-smith/internal/models"
	"github.com/tjg184/agent-smith/pkg/paths"
)

// SkillDownloader handles downloading skill components
type SkillDownloader struct {
	baseDownloader
}

func NewSkillDownloader() *SkillDownloader {
	baseDir, err := paths.GetSkillsDir()
	if err != nil {
		log.Fatal("Failed to get skills directory:", err)
	}

	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create skills directory:", err)
	}

	return &SkillDownloader{newBaseDownloader(baseDir)}
}

func NewSkillDownloaderForProfile(profileName string) *SkillDownloader {
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		log.Fatal("Failed to get profiles directory:", err)
	}

	baseDir := filepath.Join(profilesDir, profileName, "skills")

	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create profile skills directory:", err)
	}

	return &SkillDownloader{newBaseDownloader(baseDir)}
}

func NewSkillDownloaderWithTargetDir(targetDir string) *SkillDownloader {
	baseDir := filepath.Join(targetDir, "skills")

	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create target skills directory:", err)
	}

	return &SkillDownloader{newBaseDownloader(baseDir)}
}

// DownloadSkill downloads a skill from the repository
func (sd *SkillDownloader) DownloadSkill(repoURL, skillName string, providedRepoPath ...string) error {
	fullURL, err := sd.parseRepoURL(repoURL)
	if err != nil {
		return err
	}

	var repoPath string
	var commitHashFromRepo string
	hasProvidedPath := len(providedRepoPath) > 0 && providedRepoPath[0] != ""

	if hasProvidedPath {
		repoPath = providedRepoPath[0]
	} else if sd.detector.DetectProvider(repoURL) == "local" {
		repoPath = fullURL
	} else {
		tempDir, err := os.MkdirTemp("", "agent-smith-detect-*")
		if err != nil {
			return fmt.Errorf("failed to create temporary directory: %w", err)
		}
		defer os.RemoveAll(tempDir)

		repo, err := gitpkg.CloneShallow(sd.cloner, tempDir, fullURL)
		if err != nil {
			return fmt.Errorf("failed to clone repository for detection: %w", err)
		}
		repoPath = tempDir

		if repo != nil {
			ref, err := repo.Head()
			if err == nil {
				commitHashFromRepo = ref.Hash().String()
			}
		}
	}

	components, err := sd.detector.DetectComponentsInRepo(repoPath)
	if err != nil {
		return fmt.Errorf("failed to detect components: %w", err)
	}

	var skillComponents []models.DetectedComponent
	for _, comp := range components {
		if comp.Type == models.ComponentSkill {
			skillComponents = append(skillComponents, comp)
		}
	}

	if len(skillComponents) == 0 {
		return sd.downloadSkillDirect(fullURL, skillName, repoURL)
	}

	lockBaseDir := filepath.Dir(sd.baseDir)
	filesystemName, err := metadataPkg.ResolveInstallFilesystemName(lockBaseDir, "skills", skillName, fullURL)
	if err != nil {
		sd.formatter.Warning("failed to resolve filesystem name, using skill name: %v", err)
		filesystemName = skillName
	}

	skillDir := filepath.Join(sd.baseDir, filesystemName)
	if err := fileutil.CreateDirectoryWithPermissions(skillDir); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	shouldCleanup := true
	defer func() {
		if shouldCleanup {
			os.RemoveAll(skillDir)
		}
	}()

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
				matchingComponent = &comp
				break
			}
		}
	}

	if len(skillComponents) == 1 {
		component := skillComponents[0]

		err = fileutil.CopyComponentFiles(repoPath, component, skillDir)
		if err != nil {
			return fmt.Errorf("failed to copy skill files: %w", err)
		}
	} else if matchingComponent != nil {
		destFolderName := DetermineDestinationFolderName(matchingComponent.FilePath)

		if destFolderName != filesystemName {
			os.RemoveAll(skillDir)

			filesystemName = destFolderName
			skillDir = filepath.Join(sd.baseDir, filesystemName)
			if err := fileutil.CreateDirectoryWithPermissions(skillDir); err != nil {
				return fmt.Errorf("failed to create skill directory: %w", err)
			}
		}

		err = fileutil.CopyComponentFiles(repoPath, *matchingComponent, skillDir)
		if err != nil {
			return fmt.Errorf("failed to copy skill files: %w", err)
		}
	} else {
		var skillNames []string
		for _, comp := range skillComponents {
			skillNames = append(skillNames, comp.Name)
		}
		return fmt.Errorf("skill '%s' not found in repository. Available skills: %s", skillName, strings.Join(skillNames, ", "))
	}

	sourceType := sd.detectSourceType(fullURL)

	var commitHash string
	if hasProvidedPath || sd.detector.DetectProvider(repoURL) == "local" {
		if hash, err := gitpkg.GetCommitHashFromPath(sd.cloner, repoPath); err == nil {
			commitHash = hash
		} else {
			sd.formatter.Warning("failed to get commit hash: %v", err)
		}
	} else {
		commitHash = commitHashFromRepo
	}

	detectionType := "recursive"
	originalPath := ""
	if matchingComponent != nil && len(skillComponents) > 1 {
		detectionType = "single"
		originalPath = matchingComponent.FilePath
	}

	if err := sd.saveLockFile("skills", skillName, filesystemName, fullURL, sourceType, fullURL, commitHash, len(skillComponents), detectionType, originalPath); err != nil {
		sd.formatter.Warning("failed to save lock file: %v", err)
	}

	if _, err := os.Stat(skillDir + ".git"); err == nil {
		os.RemoveAll(skillDir + ".git")
	}

	shouldCleanup = false

	sd.formatter.Success("skill", skillName)

	return nil
}

func (sd *SkillDownloader) downloadSkillDirect(fullURL, skillName, repoURL string) error {
	lockBaseDir := filepath.Dir(sd.baseDir)
	filesystemName, err := metadataPkg.ResolveInstallFilesystemName(lockBaseDir, "skills", skillName, fullURL)
	if err != nil {
		sd.formatter.Warning("failed to resolve filesystem name, using skill name: %v", err)
		filesystemName = skillName
	}

	skillDir := filepath.Join(sd.baseDir, filesystemName)
	if err := fileutil.CreateDirectoryWithPermissions(skillDir); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	shouldCleanup := true
	defer func() {
		if shouldCleanup {
			os.RemoveAll(skillDir)
		}
	}()

	var cloneErr error

	if sd.detector.DetectProvider(repoURL) == "local" {
		cloneErr = fileutil.CopyDirectoryContents(fullURL, skillDir)
		if cloneErr != nil {
			return fmt.Errorf("failed to copy local repository: %w", cloneErr)
		}
	} else {
		_, cloneErr = gitpkg.CloneShallow(sd.cloner, skillDir, fullURL)
		if cloneErr != nil {
			return fmt.Errorf("failed to clone repository: %w", cloneErr)
		}
	}

	sourceType := sd.detectSourceType(fullURL)

	var commitHash string
	if hash, hashErr := gitpkg.GetCommitHashFromPath(sd.cloner, skillDir); hashErr == nil {
		commitHash = hash
	} else {
		sd.formatter.Warning("failed to get commit hash: %v", hashErr)
	}

	if err := sd.saveLockFile("skills", skillName, filesystemName, fullURL, sourceType, fullURL, commitHash, 1, "direct", ""); err != nil {
		sd.formatter.Warning("failed to save lock file: %v", err)
	}

	skillFile := filepath.Join(skillDir, "SKILL.md")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		if err := sd.createComponentMarkdownFile(skillFile, "skill", skillName, fullURL); err != nil {
			sd.formatter.Warning("failed to create SKILL.md: %v", err)
		}
	}

	shouldCleanup = false

	return nil
}

func (sd *SkillDownloader) DownloadSkillWithRepo(fullURL, skillName, repoURL string, repoPath string, components []models.DetectedComponent) error {
	var targetComponent *models.DetectedComponent
	for _, comp := range components {
		if comp.Type == models.ComponentSkill && comp.Name == skillName {
			targetComponent = &comp
			break
		}
	}

	if targetComponent == nil {
		return sd.downloadSkillDirect(fullURL, skillName, repoURL)
	}

	destFolderName := DetermineDestinationFolderName(targetComponent.FilePath)

	skillDir := filepath.Join(sd.baseDir, destFolderName)
	if err := fileutil.CreateDirectoryWithPermissions(skillDir); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	shouldCleanup := true
	defer func() {
		if shouldCleanup {
			os.RemoveAll(skillDir)
		}
	}()

	err := fileutil.CopyComponentFiles(repoPath, *targetComponent, skillDir)
	if err != nil {
		return fmt.Errorf("failed to copy skill files: %w", err)
	}

	sourceType := sd.detectSourceType(fullURL)

	var commitHash string
	if hash, err := gitpkg.GetCommitHashFromPath(sd.cloner, repoPath); err == nil {
		commitHash = hash
	} else {
		sd.formatter.Warning("failed to get commit hash: %v", err)
	}

	if err := sd.saveLockFile("skills", skillName, destFolderName, fullURL, sourceType, fullURL, commitHash, 1, "single", targetComponent.FilePath); err != nil {
		sd.formatter.Warning("failed to save lock file: %v", err)
	}

	if _, err := os.Stat(skillDir + ".git"); err == nil {
		os.RemoveAll(skillDir + ".git")
	}

	shouldCleanup = false

	return nil
}
