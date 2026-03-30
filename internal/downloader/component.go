package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjg184/agent-smith/internal/fileutil"
	gitpkg "github.com/tjg184/agent-smith/internal/git"
	metadataPkg "github.com/tjg184/agent-smith/internal/metadata"
	"github.com/tjg184/agent-smith/internal/models"
	"github.com/tjg184/agent-smith/pkg/paths"
)

// Downloader is the unified interface for all component downloaders.
type Downloader interface {
	Download(repoURL, name string, providedRepoPath ...string) error
	DownloadWithRepo(fullURL, name, repoURL, repoPath string, components []models.DetectedComponent) error
}

type componentMeta struct {
	dir          string
	fallbackFile func(name string) string
}

var componentMetaTable = map[models.ComponentType]componentMeta{
	models.ComponentSkill:   {"skills", func(_ string) string { return "SKILL.md" }},
	models.ComponentAgent:   {"agents", func(n string) string { return n + ".md" }},
	models.ComponentCommand: {"commands", func(n string) string { return n + ".md" }},
}

type componentDownloader struct {
	baseDownloader
	ct   models.ComponentType
	meta componentMeta
}

// ForType creates a downloader for the default (base) installation directory.
func ForType(ct models.ComponentType) (Downloader, error) {
	meta := componentMetaTable[ct]
	baseDir, err := baseDirForType(ct)
	if err != nil {
		return nil, err
	}
	return &componentDownloader{
		baseDownloader: newBaseDownloader(baseDir),
		ct:             ct,
		meta:           meta,
	}, nil
}

// ForTypeWithProfile creates a downloader that installs into a named profile.
func ForTypeWithProfile(ct models.ComponentType, profile string) (Downloader, error) {
	meta := componentMetaTable[ct]
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get profiles directory: %w", err)
	}
	baseDir := filepath.Join(profilesDir, profile, meta.dir)
	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		return nil, fmt.Errorf("failed to create profile component directory: %w", err)
	}
	return &componentDownloader{
		baseDownloader: newBaseDownloader(baseDir),
		ct:             ct,
		meta:           meta,
	}, nil
}

// ForTypeWithTargetDir creates a downloader that installs into a custom target directory.
func ForTypeWithTargetDir(ct models.ComponentType, targetDir string) (Downloader, error) {
	meta := componentMetaTable[ct]
	baseDir := filepath.Join(targetDir, meta.dir)
	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		return nil, fmt.Errorf("failed to create target component directory: %w", err)
	}
	return &componentDownloader{
		baseDownloader: newBaseDownloader(baseDir),
		ct:             ct,
		meta:           meta,
	}, nil
}

func baseDirForType(ct models.ComponentType) (string, error) {
	var baseDir string
	var err error
	switch ct {
	case models.ComponentSkill:
		baseDir, err = paths.GetSkillsDir()
	case models.ComponentAgent:
		baseDir, err = paths.GetAgentsSubDir()
	case models.ComponentCommand:
		baseDir, err = paths.GetCommandsDir()
	default:
		return "", fmt.Errorf("unknown component type: %s", ct)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get directory for %s: %w", ct, err)
	}
	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		return "", fmt.Errorf("failed to create directory for %s: %w", ct, err)
	}
	return baseDir, nil
}

// Download implements Downloader.
func (cd *componentDownloader) Download(repoURL, name string, providedRepoPath ...string) error {
	fullURL, err := cd.parseRepoURL(repoURL)
	if err != nil {
		return err
	}

	var repoPath string
	var commitHashFromRepo string
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

		repo, err := gitpkg.CloneShallow(cd.cloner, tempDir, fullURL)
		if err != nil {
			return fmt.Errorf("failed to clone repository for detection: %w", err)
		}
		repoPath = tempDir

		// Capture commit hash from the shallow clone (silent on error: matches skill.go behavior).
		if repo != nil {
			ref, err := repo.Head()
			if err == nil {
				commitHashFromRepo = ref.Hash().String()
			}
		}
	}

	components, err := cd.detector.DetectComponentsInRepo(repoPath)
	if err != nil {
		return fmt.Errorf("failed to detect components: %w", err)
	}

	var matching []models.DetectedComponent
	for _, comp := range components {
		if comp.Type == cd.ct {
			matching = append(matching, comp)
		}
	}

	if len(matching) == 0 {
		return cd.downloadDirect(fullURL, name, repoURL)
	}

	// Resolve the matching component first so we can use its canonical Name as the
	// lock-file key. The user-supplied name is a selector and may be either the short
	// detected name ("visual-explainer") or a path-style name ("plugins/visual-explainer");
	// both must resolve to the same canonical entry.
	var matchingComponent *models.DetectedComponent
	if len(matching) == 1 {
		matchingComponent = &matching[0]
	} else {
		matchingComponent = findComponentByName(matching, name)
	}

	if matchingComponent == nil && len(matching) > 1 {
		var names []string
		for _, comp := range matching {
			names = append(names, comp.Name)
		}
		return fmt.Errorf("%s '%s' not found in repository. Available %ss: %s",
			cd.ct, name, cd.ct, strings.Join(names, ", "))
	}

	// Use the canonical detected name as the lock key so that both "install all"
	// (which stores entries under comp.Name) and "install skill/agent/command"
	// (which receives a user-supplied selector) produce the same lock entry.
	canonicalName := name
	if matchingComponent != nil {
		canonicalName = matchingComponent.Name
	}

	lockBaseDir := filepath.Dir(cd.baseDir)
	filesystemName, err := metadataPkg.ResolveInstallFilesystemName(lockBaseDir, cd.meta.dir, canonicalName, fullURL)
	if err != nil {
		cd.formatter.Warning("failed to resolve filesystem name, using component name: %v", err)
		filesystemName = canonicalName
	}

	componentDir := filepath.Join(cd.baseDir, filesystemName)
	if err := fileutil.CreateDirectoryWithPermissions(componentDir); err != nil {
		return fmt.Errorf("failed to create component directory: %w", err)
	}

	shouldCleanup := true
	defer func() {
		if shouldCleanup {
			os.RemoveAll(componentDir)
		}
	}()

	if matchingComponent == nil || len(matching) == 1 {
		comp := matching[0]
		if err = fileutil.CopyComponentFiles(repoPath, comp, componentDir); err != nil {
			return fmt.Errorf("failed to copy component files: %w", err)
		}
	} else {
		destFolderName := DetermineDestinationFolderName(matchingComponent.FilePath)

		if destFolderName != filesystemName {
			os.RemoveAll(componentDir)
			filesystemName = destFolderName
			componentDir = filepath.Join(cd.baseDir, filesystemName)
			if err := fileutil.CreateDirectoryWithPermissions(componentDir); err != nil {
				return fmt.Errorf("failed to create component directory: %w", err)
			}
		}

		if err = fileutil.CopyComponentFiles(repoPath, *matchingComponent, componentDir); err != nil {
			return fmt.Errorf("failed to copy component files: %w", err)
		}
	}

	sourceType := cd.detectSourceType(fullURL)

	var commitHash string
	if hasProvidedPath || cd.detector.DetectProvider(repoURL) == "local" {
		if hash, err := gitpkg.GetCommitHashFromPath(cd.cloner, repoPath); err == nil {
			commitHash = hash
		} else {
			cd.formatter.Warning("failed to get commit hash: %v", err)
		}
	} else {
		commitHash = commitHashFromRepo
	}

	detectionType := "recursive"
	originalPath := ""
	if matchingComponent != nil && len(matching) > 1 {
		detectionType = "single"
		originalPath = matchingComponent.FilePath
	}

	if err := cd.saveLockFile(cd.meta.dir, canonicalName, filesystemName, fullURL, sourceType, fullURL, commitHash, len(matching), detectionType, originalPath); err != nil {
		cd.formatter.Warning("failed to save lock file: %v", err)
	}

	if _, err := os.Stat(componentDir + ".git"); err == nil {
		os.RemoveAll(componentDir + ".git")
	}

	shouldCleanup = false
	cd.formatter.Success(string(cd.ct), name)
	return nil
}

func (cd *componentDownloader) downloadDirect(fullURL, name, repoURL string) error {
	lockBaseDir := filepath.Dir(cd.baseDir)
	filesystemName, err := metadataPkg.ResolveInstallFilesystemName(lockBaseDir, cd.meta.dir, name, fullURL)
	if err != nil {
		cd.formatter.Warning("failed to resolve filesystem name, using component name: %v", err)
		filesystemName = name
	}

	componentDir := filepath.Join(cd.baseDir, filesystemName)
	if err := fileutil.CreateDirectoryWithPermissions(componentDir); err != nil {
		return fmt.Errorf("failed to create component directory: %w", err)
	}

	shouldCleanup := true
	defer func() {
		if shouldCleanup {
			os.RemoveAll(componentDir)
		}
	}()

	if cd.detector.DetectProvider(repoURL) == "local" {
		if err := fileutil.CopyDirectoryContents(fullURL, componentDir); err != nil {
			return fmt.Errorf("failed to copy local repository: %w", err)
		}
	} else {
		if _, err := gitpkg.CloneShallow(cd.cloner, componentDir, fullURL); err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
	}

	sourceType := cd.detectSourceType(fullURL)

	var commitHash string
	if hash, err := gitpkg.GetCommitHashFromPath(cd.cloner, componentDir); err == nil {
		commitHash = hash
	} else {
		cd.formatter.Warning("failed to get commit hash: %v", err)
	}

	if err := cd.saveLockFile(cd.meta.dir, name, filesystemName, fullURL, sourceType, fullURL, commitHash, 1, "direct", ""); err != nil {
		cd.formatter.Warning("failed to save lock file: %v", err)
	}

	fallbackFile := filepath.Join(componentDir, cd.meta.fallbackFile(name))
	if _, err := os.Stat(fallbackFile); os.IsNotExist(err) {
		if err := cd.createComponentMarkdownFile(fallbackFile, string(cd.ct), name, fullURL); err != nil {
			cd.formatter.Warning("failed to create fallback file: %v", err)
		}
	}

	shouldCleanup = false
	return nil
}

// DownloadWithRepo implements Downloader.
func (cd *componentDownloader) DownloadWithRepo(fullURL, name, repoURL, repoPath string, components []models.DetectedComponent) error {
	var target *models.DetectedComponent
	for _, comp := range components {
		if comp.Type == cd.ct && comp.Name == name {
			target = &comp
			break
		}
	}

	if target == nil {
		return cd.downloadDirect(fullURL, name, repoURL)
	}

	// Check whether this component is already installed (idempotency guard).
	// "install all" may be re-run, or "install skill/agent/command" may be called
	// after "install all" for the same repo — both must produce a single entry.
	lockBaseDir := filepath.Dir(cd.baseDir)
	existingFilesystemName, err := metadataPkg.ResolveInstallFilesystemName(lockBaseDir, cd.meta.dir, target.Name, fullURL)
	if err == nil && existingFilesystemName != "" {
		existingDir := filepath.Join(cd.baseDir, existingFilesystemName)
		if _, statErr := os.Stat(existingDir); statErr == nil {
			// Already on disk under the resolved name — overwrite in place rather than
			// creating a duplicate directory.
			if err := fileutil.CopyComponentFiles(repoPath, *target, existingDir); err != nil {
				return fmt.Errorf("failed to update component files: %w", err)
			}
			sourceType := cd.detectSourceType(fullURL)
			var commitHash string
			if hash, err := gitpkg.GetCommitHashFromPath(cd.cloner, repoPath); err == nil {
				commitHash = hash
			} else {
				cd.formatter.Warning("failed to get commit hash: %v", err)
			}
			if err := cd.saveLockFile(cd.meta.dir, target.Name, existingFilesystemName, fullURL, sourceType, fullURL, commitHash, 1, "single", target.FilePath); err != nil {
				cd.formatter.Warning("failed to save lock file: %v", err)
			}
			if _, err := os.Stat(existingDir + ".git"); err == nil {
				os.RemoveAll(existingDir + ".git")
			}
			return nil
		}
	}

	destFolderName := DetermineDestinationFolderName(target.FilePath)
	componentDir := filepath.Join(cd.baseDir, destFolderName)
	if err := fileutil.CreateDirectoryWithPermissions(componentDir); err != nil {
		return fmt.Errorf("failed to create component directory: %w", err)
	}

	shouldCleanup := true
	defer func() {
		if shouldCleanup {
			os.RemoveAll(componentDir)
		}
	}()

	if err := fileutil.CopyComponentFiles(repoPath, *target, componentDir); err != nil {
		return fmt.Errorf("failed to copy component files: %w", err)
	}

	sourceType := cd.detectSourceType(fullURL)

	var commitHash string
	if hash, err := gitpkg.GetCommitHashFromPath(cd.cloner, repoPath); err == nil {
		commitHash = hash
	} else {
		cd.formatter.Warning("failed to get commit hash: %v", err)
	}

	// Use target.Name (canonical detected name) as the lock key so it matches
	// the entry written by "install skill/agent/command".
	if err := cd.saveLockFile(cd.meta.dir, target.Name, destFolderName, fullURL, sourceType, fullURL, commitHash, 1, "single", target.FilePath); err != nil {
		cd.formatter.Warning("failed to save lock file: %v", err)
	}

	if _, err := os.Stat(componentDir + ".git"); err == nil {
		os.RemoveAll(componentDir + ".git")
	}

	shouldCleanup = false
	return nil
}

// findComponentByName finds a component by matching the user-supplied name against both the
// canonical detected name and the path-style destination folder name. This lets users pass
// either "visual-explainer" or "plugins/visual-explainer" to refer to the same component.
func findComponentByName(components []models.DetectedComponent, name string) *models.DetectedComponent {
	// First pass: exact match on canonical detected name.
	for i := range components {
		if components[i].Name == name {
			return &components[i]
		}
	}
	// Second pass: match on path-style destination folder name (e.g. "plugins/visual-explainer").
	for i := range components {
		if DetermineDestinationFolderName(components[i].FilePath) == name {
			return &components[i]
		}
	}
	return nil
}
