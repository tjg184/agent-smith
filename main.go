package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/tgaines/agent-smith/cmd"
	"github.com/tgaines/agent-smith/internal/detector"
	"github.com/tgaines/agent-smith/internal/fileutil"
	"github.com/tgaines/agent-smith/internal/models"
	"github.com/tgaines/agent-smith/pkg/paths"
)

type SkillDownloader struct {
	baseDir  string
	detector *detector.RepositoryDetector
}

type AgentDownloader struct {
	baseDir  string
	detector *detector.RepositoryDetector
}

type CommandDownloader struct {
	baseDir  string
	detector *detector.RepositoryDetector
}

type ComponentLinker struct {
	agentsDir   string
	opencodeDir string
	detector    *detector.RepositoryDetector
}

type BulkDownloader struct {
	skillDownloader   *SkillDownloader
	agentDownloader   *AgentDownloader
	commandDownloader *CommandDownloader
	detector          *detector.RepositoryDetector
}

type UpdateDetector struct {
	baseDir  string
	detector *detector.RepositoryDetector
}

type ComponentLockFile struct {
	Version  int                                  `json:"version"`
	Skills   map[string]models.ComponentLockEntry `json:"skills"`
	Agents   map[string]models.ComponentLockEntry `json:"agents,omitempty"`
	Commands map[string]models.ComponentLockEntry `json:"commands,omitempty"`
}

// Cross-platform helper functions
func getCrossPlatformPermissions() os.FileMode {
	return fileutil.GetCrossPlatformPermissions()
}

func getCrossPlatformFilePermissions() os.FileMode {
	return fileutil.GetCrossPlatformFilePermissions()
}

func createDirectoryWithPermissions(path string) error {
	return fileutil.CreateDirectoryWithPermissions(path)
}

func createFileWithPermissions(path string, data []byte) error {
	return fileutil.CreateFileWithPermissions(path, data)
}

// parseFrontmatter extracts YAML frontmatter from a markdown file
// Frontmatter must be delimited by "---" at the start of the file
// Returns nil if no frontmatter is found (not an error)
func parseFrontmatter(filePath string) (*models.ComponentFrontmatter, error) {
	return fileutil.ParseFrontmatter(filePath)
}

// determineComponentName determines the component name using frontmatter or filename
// Priority: frontmatter.name > filename (without extension)
// Special files (README.md, index.md, main.md) are skipped
func determineComponentName(frontmatter *models.ComponentFrontmatter, fileName string) string {
	return fileutil.DetermineComponentName(frontmatter, fileName)
}

// determineDestinationFolderName determines the destination folder name using hierarchy heuristic
// Walks up from component file directory, skipping component-type names (agents/commands/skills)
// Returns first non-component-type directory name for preserving optional hierarchy
func determineDestinationFolderName(componentFilePath string) string {
	componentTypeNames := paths.GetComponentTypeNames()

	// Get directory containing the component file
	currentDir := filepath.Dir(componentFilePath)

	// Walk up the directory tree
	for {
		dirName := filepath.Base(currentDir)

		// Check if current directory name is a component type
		isComponentType := false
		for _, typeName := range componentTypeNames {
			if dirName == typeName {
				isComponentType = true
				break
			}
		}

		// If not a component type name, use it
		if !isComponentType && dirName != "." && dirName != "" {
			return dirName
		}

		// Go up one directory
		parentDir := filepath.Dir(currentDir)

		// Check if we've reached the root
		if parentDir == currentDir || parentDir == "." || parentDir == "/" || dirName == "" {
			// Reached root, fall back to "root"
			return "root"
		}

		currentDir = parentDir
	}
}

func NewSkillDownloader() *SkillDownloader {
	baseDir, err := paths.GetSkillsDir()
	if err != nil {
		log.Fatal("Failed to get skills directory:", err)
	}

	// Create base directory if it doesn't exist
	if err := createDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create skills directory:", err)
	}

	return &SkillDownloader{
		baseDir:  baseDir,
		detector: detector.NewRepositoryDetector(),
	}
}

func NewAgentDownloader() *AgentDownloader {
	baseDir, err := paths.GetAgentsSubDir()
	if err != nil {
		log.Fatal("Failed to get agents directory:", err)
	}

	// Create base directory if it doesn't exist
	if err := createDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create agents directory:", err)
	}

	return &AgentDownloader{
		baseDir:  baseDir,
		detector: detector.NewRepositoryDetector(),
	}
}

func NewCommandDownloader() *CommandDownloader {
	baseDir, err := paths.GetCommandsDir()
	if err != nil {
		log.Fatal("Failed to get commands directory:", err)
	}

	// Create base directory if it doesn't exist
	if err := createDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create commands directory:", err)
	}

	return &CommandDownloader{
		baseDir:  baseDir,
		detector: detector.NewRepositoryDetector(),
	}
}

func NewComponentLinker() *ComponentLinker {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		log.Fatal("Failed to get agents directory:", err)
	}

	opencodeDir, err := paths.GetOpencodeDir()
	if err != nil {
		log.Fatal("Failed to get opencode directory:", err)
	}

	// Create opencode directory if it doesn't exist
	if err := createDirectoryWithPermissions(opencodeDir); err != nil {
		log.Fatal("Failed to create opencode directory:", err)
	}

	return &ComponentLinker{
		agentsDir:   agentsDir,
		opencodeDir: opencodeDir,
		detector:    detector.NewRepositoryDetector(),
	}
}

func NewUpdateDetector() *UpdateDetector {
	baseDir, err := paths.GetAgentsDir()
	if err != nil {
		log.Fatal("Failed to get agents directory:", err)
	}

	return &UpdateDetector{
		baseDir:  baseDir,
		detector: detector.NewRepositoryDetector(),
	}
}

func NewBulkDownloader() *BulkDownloader {
	return &BulkDownloader{
		skillDownloader:   NewSkillDownloader(),
		agentDownloader:   NewAgentDownloader(),
		commandDownloader: NewCommandDownloader(),
		detector:          detector.NewRepositoryDetector(),
	}
}

// Compute GitHub tree SHA for skill folder hash (npx add-skill compatible)
func computeGitHubTreeSHA(ownerRepo string, skillPath string) (string, error) {
	// Normalize skill path - remove SKILL.md suffix to get folder path
	folderPath := skillPath
	if strings.HasSuffix(folderPath, "/SKILL.md") {
		folderPath = folderPath[:len(folderPath)-9]
	} else if strings.HasSuffix(folderPath, "SKILL.md") {
		folderPath = folderPath[:len(folderPath)-8]
	}
	if strings.HasSuffix(folderPath, "/") {
		folderPath = folderPath[:len(folderPath)-1]
	}

	branches := []string{"main", "master"}

	for _, branch := range branches {
		url := fmt.Sprintf("https://api.github.com/repos/%s/git/trees/%s?recursive=1", ownerRepo, branch)
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		var treeData struct {
			Tree []struct {
				Path string `json:"path"`
				Type string `json:"type"`
				SHA  string `json:"sha"`
			} `json:"tree"`
		}

		if err := json.Unmarshal(body, &treeData); err != nil {
			continue
		}

		// Find tree entry for skill folder
		for _, entry := range treeData.Tree {
			if entry.Type == "tree" && entry.Path == folderPath {
				return entry.SHA, nil
			}
		}
	}

	return "", fmt.Errorf("skill folder not found in GitHub API")
}

// Compute local content hash for skill folder
func computeLocalFolderHash(folderPath string) (string, error) {
	hasher := sha256.New()

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(folderPath, path)
		if err != nil {
			return err
		}

		// Write relative path and file info to hash
		hasher.Write([]byte(relPath))
		hasher.Write([]byte(info.Mode().String()))
		hasher.Write([]byte(info.ModTime().Format(time.RFC3339)))

		// Write file content
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		hasher.Write(data)
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to compute folder hash: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
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

func (sd *SkillDownloader) downloadSkill(repoURL, skillName string, providedRepoPath ...string) error {
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
	if err := createDirectoryWithPermissions(skillDir); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	// If only one skill component found, copy its files
	if len(skillComponents) == 1 {
		component := skillComponents[0]

		// Copy component files (non-recursive) using FilePath to skill directory
		err = sd.copyComponentFiles(repoPath, component, skillDir)
		if err != nil {
			os.RemoveAll(skillDir)
			return fmt.Errorf("failed to copy skill files: %w", err)
		}
	} else {
		// Multiple skills found, create a monorepo structure
		for _, component := range skillComponents {
			componentDir := filepath.Join(skillDir, component.Name)

			err = createDirectoryWithPermissions(componentDir)
			if err != nil {
				continue
			}

			// Copy component files (non-recursive) using FilePath
			err = sd.copyComponentFiles(repoPath, component, componentDir)
			if err != nil {
				log.Printf("Warning: failed to copy skill %s: %v", component.Name, err)
			}
		}
	}

	var commitHash string
	var repo *git.Repository

	// Handle metadata differently for local vs remote repositories
	if sd.detector.DetectProvider(repoURL) == "local" {
		// For local repositories, open the repository directly
		repo, err = git.PlainOpen(fullURL)
		if err != nil {
			// Non-fatal, continue without git metadata
			log.Printf("Warning: failed to open local repository for metadata: %v", err)
		}
	} else {
		// For remote repositories, clone to get git history for metadata
		repo, err = git.PlainClone(skillDir+".git", true, &git.CloneOptions{
			URL:           fullURL,
			Depth:         1,
			ReferenceName: plumbing.HEAD,
			SingleBranch:  true,
		})
		if err != nil {
			// Non-fatal, continue without git metadata
			log.Printf("Warning: failed to clone repository for metadata: %v", err)
		}
	}

	if repo != nil {
		if ref, err := repo.Head(); err == nil {
			commitHash = ref.Hash().String()
		}
	}

	// Create metadata
	metadata := map[string]interface{}{
		"name":       skillName,
		"source":     fullURL,
		"commit":     commitHash,
		"downloaded": "now",
		"components": len(skillComponents),
		"detection":  "recursive",
	}

	// Save legacy metadata file for backward compatibility
	metadataFile := filepath.Join(skillDir, ".skill-metadata.json")
	if err := sd.saveMetadata(metadataFile, metadata); err != nil {
		log.Printf("Warning: failed to save metadata: %v", err)
	}

	// Save to npx add-skill compatible lock file
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
			if hash, err := computeGitHubTreeSHA(ownerRepo, "SKILL.md"); err == nil {
				folderHash = hash
			}
		}
	} else {
		// For non-GitHub repos, compute local hash
		if hash, err := computeLocalFolderHash(skillDir); err == nil {
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

	fmt.Printf("Successfully downloaded skill '%s' from %s\n", skillName, fullURL)
	fmt.Printf("Skill stored in: %s\n", skillDir)
	fmt.Printf("Components detected: %d\n", len(skillComponents))

	return nil
}

func (sd *SkillDownloader) downloadSkillDirect(fullURL, skillName, repoURL string) error {
	// Create skill directory
	skillDir := filepath.Join(sd.baseDir, skillName)
	if err := createDirectoryWithPermissions(skillDir); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	var repo *git.Repository
	var err error

	// Handle local vs remote repositories
	if sd.detector.DetectProvider(repoURL) == "local" {
		// For local repositories, copy directory contents directly
		err = sd.copyDirectoryContents(fullURL, skillDir)
		if err != nil {
			os.RemoveAll(skillDir)
			return fmt.Errorf("failed to copy local repository: %w", err)
		}

		// Open local repository for metadata
		repo, err = git.PlainOpen(fullURL)
		if err != nil {
			log.Printf("Warning: failed to open local repository for metadata: %v", err)
		}
	} else {
		// For remote repositories, clone directly
		repo, err = git.PlainClone(skillDir, false, &git.CloneOptions{
			URL:           fullURL,
			Depth:         1,
			ReferenceName: plumbing.HEAD,
			SingleBranch:  true,
		})
		if err != nil {
			os.RemoveAll(skillDir)
			return fmt.Errorf("failed to clone repository: %w", err)
		}
	}

	// Get repository info for metadata
	var commitHash string
	if repo != nil {
		if ref, err := repo.Head(); err == nil {
			commitHash = ref.Hash().String()
		}
	}

	// Create metadata
	metadata := map[string]interface{}{
		"name":       skillName,
		"source":     fullURL,
		"commit":     commitHash,
		"downloaded": "now",
		"detection":  "direct",
	}

	// Save legacy metadata file for backward compatibility
	metadataFile := filepath.Join(skillDir, ".skill-metadata.json")
	if err := sd.saveMetadata(metadataFile, metadata); err != nil {
		log.Printf("Warning: failed to save metadata: %v", err)
	}

	// Save to npx add-skill compatible lock file
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
			if hash, err := computeGitHubTreeSHA(ownerRepo, "SKILL.md"); err == nil {
				folderHash = hash
			}
		}
	} else {
		// For non-GitHub repos, compute local hash
		if hash, err := computeLocalFolderHash(skillDir); err == nil {
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

func (sd *SkillDownloader) copyDirectoryContents(src, dst string) error {
	return fileutil.CopyDirectoryContents(src, dst)
}

// copyComponentFiles copies the entire directory containing the component file (recursive)
// Uses FilePath to determine the component's directory and copies all contents recursively
func (sd *SkillDownloader) copyComponentFiles(repoPath string, component models.DetectedComponent, dst string) error {
	return fileutil.CopyComponentFiles(repoPath, component, dst)
}

func (sd *SkillDownloader) copyFile(src, dst string) error {
	return fileutil.CopyFile(src, dst)
}

func (sd *SkillDownloader) saveMetadata(filePath string, metadata map[string]interface{}) error {
	metadata["downloaded"] = time.Now().Format(time.RFC3339)

	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return createFileWithPermissions(filePath, jsonData)
}

// Save component lock entry in npx add-skill compatible format
func (sd *SkillDownloader) saveLockFile(skillName string, source string, sourceType string, sourceUrl string, skillFolderHash string, components int, detection string, originalPath string) error {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return fmt.Errorf("failed to get agents directory: %w", err)
	}

	if err := createDirectoryWithPermissions(agentsDir); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	lockFilePath := paths.GetComponentLockPath(agentsDir, "skills")

	// Read existing lock file or create new one
	var lockFile ComponentLockFile
	lockData, err := os.ReadFile(lockFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			lockFile = ComponentLockFile{
				Version: 3, // Current version matching npx add-skill
				Skills:  make(map[string]models.ComponentLockEntry),
			}
		} else {
			return fmt.Errorf("failed to read lock file: %w", err)
		}
	} else {
		if err := json.Unmarshal(lockData, &lockFile); err != nil {
			// If lock file is corrupted, create new one
			lockFile = ComponentLockFile{
				Version: 3,
				Skills:  make(map[string]models.ComponentLockEntry),
			}
		}
		// Ensure version is current
		if lockFile.Version < 3 {
			lockFile.Version = 3
			if lockFile.Skills == nil {
				lockFile.Skills = make(map[string]models.ComponentLockEntry)
			}
		}
	}

	now := time.Now().Format(time.RFC3339)

	// Check if entry exists to preserve installedAt
	existingEntry, exists := lockFile.Skills[skillName]
	if !exists {
		existingEntry.InstalledAt = now
	}

	// Update or add the skill entry
	lockFile.Skills[skillName] = models.ComponentLockEntry{
		Source:          source,
		SourceType:      sourceType,
		SourceUrl:       sourceUrl,
		OriginalPath:    originalPath, // Track original path in repo
		SkillFolderHash: skillFolderHash,
		InstalledAt:     existingEntry.InstalledAt,
		UpdatedAt:       now,
		Version:         3,
		Components:      components,
		Detection:       detection,
	}

	// Write back to file
	jsonData, err := json.MarshalIndent(lockFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}

	return os.WriteFile(lockFilePath, jsonData, 0644)
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

	return createFileWithPermissions(filePath, []byte(content))
}

func (sd *SkillDownloader) downloadSkillWithRepo(fullURL, skillName, repoURL string, repoPath string, components []models.DetectedComponent) error {
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
	destFolderName := determineDestinationFolderName(targetComponent.FilePath)

	// Create skill directory with heuristic name
	skillDir := filepath.Join(sd.baseDir, destFolderName)
	if err := createDirectoryWithPermissions(skillDir); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	// Copy the entire component directory recursively
	err := sd.copyComponentFiles(repoPath, *targetComponent, skillDir)
	if err != nil {
		os.RemoveAll(skillDir)
		return fmt.Errorf("failed to copy skill files: %w", err)
	}

	var commitHash string
	var repo *git.Repository

	// Handle metadata differently for local vs remote repositories
	if sd.detector.DetectProvider(repoURL) == "local" {
		// For local repositories, open the repository directly
		repo, err = git.PlainOpen(repoPath)
		if err != nil {
			// Non-fatal, continue without git metadata
			log.Printf("Warning: failed to open local repository for metadata: %v", err)
		}
	} else {
		// For remote repositories, use the already-cloned repository at repoPath
		repo, err = git.PlainOpen(repoPath)
		if err != nil {
			// Non-fatal, continue without git metadata
			log.Printf("Warning: failed to open repository for metadata: %v", err)
		}
	}

	if repo != nil {
		if ref, err := repo.Head(); err == nil {
			commitHash = ref.Hash().String()
		}
	}

	// Create metadata
	metadata := map[string]interface{}{
		"name":         skillName,
		"source":       fullURL,
		"commit":       commitHash,
		"downloaded":   "now",
		"components":   1,
		"detection":    "single",
		"originalPath": targetComponent.FilePath,
	}

	// Save legacy metadata file for backward compatibility
	metadataFile := filepath.Join(skillDir, ".skill-metadata.json")
	if err := sd.saveMetadata(metadataFile, metadata); err != nil {
		log.Printf("Warning: failed to save metadata: %v", err)
	}

	// Save to npx add-skill compatible lock file
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
			if hash, err := computeGitHubTreeSHA(ownerRepo, targetComponent.SourceFile); err == nil {
				folderHash = hash
			}
		}
	} else {
		// For non-GitHub repos, compute local hash
		if hash, err := computeLocalFolderHash(skillDir); err == nil {
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

	fmt.Printf("Successfully downloaded skill '%s' from %s\n", skillName, fullURL)
	fmt.Printf("Skill stored in: %s\n", skillDir)

	return nil
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

func (cd *CommandDownloader) downloadCommand(repoURL, commandName string, providedRepoPath ...string) error {
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
	if err := createDirectoryWithPermissions(commandDir); err != nil {
		return fmt.Errorf("failed to create command directory: %w", err)
	}

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
		err = cd.copyComponentFiles(repoPath, component, commandDir)
		if err != nil {
			os.RemoveAll(commandDir)
			return fmt.Errorf("failed to copy command files: %w", err)
		}
	} else if matchingComponent != nil {
		// Downloading a specific component from a multi-component directory
		// Use direct copy to avoid double nesting

		// Copy component files (non-recursive) using FilePath to command directory
		err = cd.copyComponentFiles(repoPath, *matchingComponent, commandDir)
		if err != nil {
			os.RemoveAll(commandDir)
			return fmt.Errorf("failed to copy command files: %w", err)
		}
	} else {
		// Multiple commands found, create a monorepo structure
		for _, component := range commandComponents {
			componentDir := filepath.Join(commandDir, component.Name)

			err = createDirectoryWithPermissions(componentDir)
			if err != nil {
				continue
			}

			// Copy component files (non-recursive) using FilePath
			err = cd.copyComponentFiles(repoPath, component, componentDir)
			if err != nil {
				log.Printf("Warning: failed to copy command %s: %v", component.Name, err)
			}
		}
	}

	// Clone the repository again to get proper git history for metadata
	repo, err := git.PlainClone(commandDir+".git", true, &git.CloneOptions{
		URL:           fullURL,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	})
	if err != nil {
		// Non-fatal, continue without git metadata
		log.Printf("Warning: failed to clone repository for metadata: %v", err)
	}

	var commitHash string
	if repo != nil {
		if ref, err := repo.Head(); err == nil {
			commitHash = ref.Hash().String()
		}
	}

	// Create metadata
	metadata := map[string]interface{}{
		"name":       commandName,
		"source":     fullURL,
		"commit":     commitHash,
		"downloaded": "now",
		"components": len(commandComponents),
		"detection":  "recursive",
	}

	// Save legacy metadata file for backward compatibility
	legacyMetadataFile := filepath.Join(commandDir, ".command-metadata.json")
	if err := cd.saveMetadata(legacyMetadataFile, metadata); err != nil {
		log.Printf("Warning: failed to save legacy metadata: %v", err)
	}

	// Save to npx add-skill compatible lock file
	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	// Compute folder hash if it's a GitHub repo
	var folderHash string
	if sourceType != "github" {
		if hash, err := computeLocalFolderHash(commandDir); err == nil {
			folderHash = hash
		}
	}

	if err := cd.saveLockFile(commandName, fullURL, sourceType, fullURL, folderHash, len(commandComponents), "recursive", ""); err != nil {
		log.Printf("Warning: failed to save lock file: %v", err)
	}

	// Clean up git clone if it exists
	if _, err := os.Stat(commandDir + ".git"); err == nil {
		os.RemoveAll(commandDir + ".git")
	}

	fmt.Printf("Successfully downloaded command '%s' from %s\n", commandName, fullURL)
	fmt.Printf("Command stored in: %s\n", commandDir)
	fmt.Printf("Components detected: %d\n", len(commandComponents))

	return nil
}

func (cd *CommandDownloader) downloadCommandDirect(fullURL, commandName string) error {
	// Create command directory
	commandDir := filepath.Join(cd.baseDir, commandName)
	if err := createDirectoryWithPermissions(commandDir); err != nil {
		return fmt.Errorf("failed to create command directory: %w", err)
	}

	// Clone repository directly
	repo, err := git.PlainClone(commandDir, false, &git.CloneOptions{
		URL:           fullURL,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	})
	if err != nil {
		os.RemoveAll(commandDir)
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Get repository info for metadata
	var commitHash string
	if ref, err := repo.Head(); err == nil {
		commitHash = ref.Hash().String()
	}

	// Create metadata
	metadata := map[string]interface{}{
		"name":       commandName,
		"source":     fullURL,
		"commit":     commitHash,
		"downloaded": "now",
		"detection":  "direct",
	}

	// Save legacy metadata file for backward compatibility
	legacyMetadataFile := filepath.Join(commandDir, ".command-metadata.json")
	if err := cd.saveMetadata(legacyMetadataFile, metadata); err != nil {
		log.Printf("Warning: failed to save legacy metadata: %v", err)
	}

	// Save to npx add-skill compatible lock file
	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	// Compute folder hash if it's a GitHub repo
	var folderHash string
	if sourceType != "github" {
		if hash, err := computeLocalFolderHash(commandDir); err == nil {
			folderHash = hash
		}
	}

	if err := cd.saveLockFile(commandName, fullURL, sourceType, fullURL, folderHash, 1, "direct", ""); err != nil {
		log.Printf("Warning: failed to save lock file: %v", err)
	}

	// Create {name}.md if it doesn't exist
	commandFile := filepath.Join(commandDir, commandName+".md")
	if _, err := os.Stat(commandFile); os.IsNotExist(err) {
		if err := cd.createCommandFile(commandFile, commandName, fullURL); err != nil {
			log.Printf("Warning: failed to create %s.md: %v", commandName, err)
		}
	}

	return nil
}

func (cd *CommandDownloader) copyDirectoryContents(src, dst string) error {
	return fileutil.CopyDirectoryContents(src, dst)
}

// copyComponentFiles copies the entire directory containing the component file (recursive)
// Uses FilePath to determine the component's directory and copies all contents recursively
func (cd *CommandDownloader) copyComponentFiles(repoPath string, component models.DetectedComponent, dst string) error {
	return fileutil.CopyComponentFiles(repoPath, component, dst)
}

func (cd *CommandDownloader) copyFile(src, dst string) error {
	return fileutil.CopyFile(src, dst)
}

func (cd *CommandDownloader) saveMetadata(filePath string, metadata map[string]interface{}) error {
	metadata["downloaded"] = time.Now().Format(time.RFC3339)

	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return createFileWithPermissions(filePath, jsonData)
}

// Save command lock entry in npx add-skill compatible format
func (cd *CommandDownloader) saveLockFile(commandName string, source string, sourceType string, sourceUrl string, skillFolderHash string, components int, detection string, originalPath string) error {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return fmt.Errorf("failed to get agents directory: %w", err)
	}

	if err := createDirectoryWithPermissions(agentsDir); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	lockFilePath := paths.GetComponentLockPath(agentsDir, "commands")

	// Read existing lock file or create new one
	var lockFile ComponentLockFile
	lockData, err := os.ReadFile(lockFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			lockFile = ComponentLockFile{
				Version:  3,
				Commands: make(map[string]models.ComponentLockEntry),
			}
		} else {
			return fmt.Errorf("failed to read lock file: %w", err)
		}
	} else {
		if err := json.Unmarshal(lockData, &lockFile); err != nil {
			lockFile = ComponentLockFile{
				Version:  3,
				Commands: make(map[string]models.ComponentLockEntry),
			}
		}
		if lockFile.Version < 3 {
			lockFile.Version = 3
			if lockFile.Commands == nil {
				lockFile.Commands = make(map[string]models.ComponentLockEntry)
			}
		}
	}

	now := time.Now().Format(time.RFC3339)

	existingEntry, exists := lockFile.Commands[commandName]
	if !exists {
		existingEntry.InstalledAt = now
	}

	lockFile.Commands[commandName] = models.ComponentLockEntry{
		Source:          source,
		SourceType:      sourceType,
		SourceUrl:       sourceUrl,
		OriginalPath:    originalPath, // Track original path in repo
		SkillFolderHash: skillFolderHash,
		InstalledAt:     existingEntry.InstalledAt,
		UpdatedAt:       now,
		Version:         3,
		Components:      components,
		Detection:       detection,
	}

	jsonData, err := json.MarshalIndent(lockFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}

	return os.WriteFile(lockFilePath, jsonData, 0644)
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

	return createFileWithPermissions(filePath, []byte(content))
}

func (cd *CommandDownloader) downloadCommandWithRepo(fullURL, commandName, repoURL string, repoPath string, components []models.DetectedComponent) error {
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
	destFolderName := determineDestinationFolderName(targetComponent.FilePath)

	// Create command directory with heuristic name
	commandDir := filepath.Join(cd.baseDir, destFolderName)
	if err := createDirectoryWithPermissions(commandDir); err != nil {
		return fmt.Errorf("failed to create command directory: %w", err)
	}

	// Copy the entire component directory recursively
	err := cd.copyComponentFiles(repoPath, *targetComponent, commandDir)
	if err != nil {
		os.RemoveAll(commandDir)
		return fmt.Errorf("failed to copy command files: %w", err)
	}

	var commitHash string
	var repo *git.Repository

	// Handle metadata differently for local vs remote repositories
	if cd.detector.DetectProvider(repoURL) == "local" {
		// For local repositories, open the repository directly
		var err error
		repo, err = git.PlainOpen(fullURL)
		if err != nil {
			// Non-fatal, continue without git metadata
			log.Printf("Warning: failed to open local repository for metadata: %v", err)
		}
	} else {
		// For remote repositories, use the already-cloned repository at repoPath
		var err error
		repo, err = git.PlainOpen(repoPath)
		if err != nil {
			// Non-fatal, continue without git metadata
			log.Printf("Warning: failed to open repository for metadata: %v", err)
		}
	}

	if repo != nil {
		if ref, err := repo.Head(); err == nil {
			commitHash = ref.Hash().String()
		}
	}

	// Create metadata
	metadata := map[string]interface{}{
		"name":         commandName,
		"source":       fullURL,
		"commit":       commitHash,
		"downloaded":   "now",
		"components":   1,
		"detection":    "single",
		"originalPath": targetComponent.FilePath,
	}

	// Save legacy metadata file for backward compatibility
	legacyMetadataFile := filepath.Join(commandDir, ".command-metadata.json")
	if err := cd.saveMetadata(legacyMetadataFile, metadata); err != nil {
		log.Printf("Warning: failed to save legacy metadata: %v", err)
	}

	// Save to npx add-skill compatible lock file
	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	// Compute folder hash if it's a GitHub repo
	var folderHash string
	if sourceType != "github" {
		if hash, err := computeLocalFolderHash(commandDir); err == nil {
			folderHash = hash
		}
	}

	if err := cd.saveLockFile(destFolderName, fullURL, sourceType, fullURL, folderHash, 1, "single", targetComponent.FilePath); err != nil {
		log.Printf("Warning: failed to save lock file: %v", err)
	}

	// Clean up git clone if it exists
	if _, err := os.Stat(commandDir + ".git"); err == nil {
		os.RemoveAll(commandDir + ".git")
	}

	fmt.Printf("Successfully downloaded command '%s' from %s\n", commandName, fullURL)
	fmt.Printf("Command stored in: %s\n", commandDir)

	return nil
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

func (ad *AgentDownloader) downloadAgent(repoURL, agentName string, providedRepoPath ...string) error {
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
	if err := createDirectoryWithPermissions(agentDir); err != nil {
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
		err = ad.copyComponentFiles(repoPath, component, agentDir)
		if err != nil {
			os.RemoveAll(agentDir)
			return fmt.Errorf("failed to copy agent files: %w", err)
		}
	} else if matchingComponent != nil {
		// Downloading a specific component from a multi-component directory
		// Use direct copy to avoid double nesting

		// Copy component files (non-recursive) using FilePath to agent directory
		err = ad.copyComponentFiles(repoPath, *matchingComponent, agentDir)
		if err != nil {
			os.RemoveAll(agentDir)
			return fmt.Errorf("failed to copy agent files: %w", err)
		}
	} else {
		// Multiple agents found, create a monorepo structure
		for _, component := range agentComponents {
			componentDir := filepath.Join(agentDir, component.Name)

			err = createDirectoryWithPermissions(componentDir)
			if err != nil {
				continue
			}

			// Copy component files (non-recursive) using FilePath
			err = ad.copyComponentFiles(repoPath, component, componentDir)
			if err != nil {
				log.Printf("Warning: failed to copy agent %s: %v", component.Name, err)
			}
		}
	}

	// Clone the repository again to get proper git history for metadata
	repo, err := git.PlainClone(agentDir+".git", true, &git.CloneOptions{
		URL:           fullURL,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	})
	if err != nil {
		// Non-fatal, continue without git metadata
		log.Printf("Warning: failed to clone repository for metadata: %v", err)
	}

	var commitHash string
	if repo != nil {
		if ref, err := repo.Head(); err == nil {
			commitHash = ref.Hash().String()
		}
	}

	// Create metadata
	metadata := map[string]interface{}{
		"name":       agentName,
		"source":     fullURL,
		"commit":     commitHash,
		"downloaded": "now",
		"components": len(agentComponents),
		"detection":  "recursive",
	}

	// Save legacy metadata file for backward compatibility
	legacyMetadataFile := filepath.Join(agentDir, ".agent-metadata.json")
	if err := ad.saveMetadata(legacyMetadataFile, metadata); err != nil {
		log.Printf("Warning: failed to save legacy metadata: %v", err)
	}

	// Save to npx add-skill compatible lock file
	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	// Compute folder hash if it's a GitHub repo
	var folderHash string
	if sourceType != "github" {
		if hash, err := computeLocalFolderHash(agentDir); err == nil {
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
	if err := createDirectoryWithPermissions(agentDir); err != nil {
		return fmt.Errorf("failed to create agent directory: %w", err)
	}

	// Clone repository directly
	repo, err := git.PlainClone(agentDir, false, &git.CloneOptions{
		URL:           fullURL,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	})
	if err != nil {
		os.RemoveAll(agentDir)
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Get repository info for metadata
	var commitHash string
	if ref, err := repo.Head(); err == nil {
		commitHash = ref.Hash().String()
	}

	// Create metadata
	metadata := map[string]interface{}{
		"name":       agentName,
		"source":     fullURL,
		"commit":     commitHash,
		"downloaded": "now",
		"detection":  "direct",
	}

	// Save legacy metadata file for backward compatibility
	legacyMetadataFile := filepath.Join(agentDir, ".agent-metadata.json")
	if err := ad.saveMetadata(legacyMetadataFile, metadata); err != nil {
		log.Printf("Warning: failed to save legacy metadata: %v", err)
	}

	// Save to npx add-skill compatible lock file
	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	// Compute folder hash if it's a GitHub repo
	var folderHash string
	if sourceType != "github" {
		if hash, err := computeLocalFolderHash(agentDir); err == nil {
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

func (ad *AgentDownloader) copyDirectoryContents(src, dst string) error {
	return fileutil.CopyDirectoryContents(src, dst)
}

// copyComponentFiles copies the entire directory containing the component file (recursive)
// Uses FilePath to determine the component's directory and copies all contents recursively
func (ad *AgentDownloader) copyComponentFiles(repoPath string, component models.DetectedComponent, dst string) error {
	return fileutil.CopyComponentFiles(repoPath, component, dst)
}

func (ad *AgentDownloader) copyFile(src, dst string) error {
	return fileutil.CopyFile(src, dst)
}

func (ad *AgentDownloader) saveMetadata(filePath string, metadata map[string]interface{}) error {
	metadata["downloaded"] = time.Now().Format(time.RFC3339)

	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return createFileWithPermissions(filePath, jsonData)
}

// Save agent lock entry in npx add-skill compatible format
func (ad *AgentDownloader) saveLockFile(agentName string, source string, sourceType string, sourceUrl string, skillFolderHash string, components int, detection string, originalPath string) error {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return fmt.Errorf("failed to get agents directory: %w", err)
	}

	if err := createDirectoryWithPermissions(agentsDir); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	lockFilePath := paths.GetComponentLockPath(agentsDir, "agents")

	// Read existing lock file or create new one
	var lockFile ComponentLockFile
	lockData, err := os.ReadFile(lockFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			lockFile = ComponentLockFile{
				Version: 3,
				Agents:  make(map[string]models.ComponentLockEntry),
			}
		} else {
			return fmt.Errorf("failed to read lock file: %w", err)
		}
	} else {
		if err := json.Unmarshal(lockData, &lockFile); err != nil {
			lockFile = ComponentLockFile{
				Version: 3,
				Agents:  make(map[string]models.ComponentLockEntry),
			}
		}
		if lockFile.Version < 3 {
			lockFile.Version = 3
			if lockFile.Agents == nil {
				lockFile.Agents = make(map[string]models.ComponentLockEntry)
			}
		}
	}

	now := time.Now().Format(time.RFC3339)

	existingEntry, exists := lockFile.Agents[agentName]
	if !exists {
		existingEntry.InstalledAt = now
	}

	lockFile.Agents[agentName] = models.ComponentLockEntry{
		Source:          source,
		SourceType:      sourceType,
		SourceUrl:       sourceUrl,
		OriginalPath:    originalPath, // Track original directory path in repo
		SkillFolderHash: skillFolderHash,
		InstalledAt:     existingEntry.InstalledAt,
		UpdatedAt:       now,
		Version:         3,
		Components:      components,
		Detection:       detection,
	}

	jsonData, err := json.MarshalIndent(lockFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}

	return os.WriteFile(lockFilePath, jsonData, 0644)
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

	return createFileWithPermissions(filePath, []byte(content))
}

func (ad *AgentDownloader) downloadAgentWithRepo(fullURL, agentName, repoURL string, repoPath string, components []models.DetectedComponent) error {
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
	destFolderName := determineDestinationFolderName(targetComponent.FilePath)

	// Create agent directory with heuristic name
	agentDir := filepath.Join(ad.baseDir, destFolderName)
	if err := createDirectoryWithPermissions(agentDir); err != nil {
		return fmt.Errorf("failed to create agent directory: %w", err)
	}

	// Copy the entire component directory recursively
	err := ad.copyComponentFiles(repoPath, *targetComponent, agentDir)
	if err != nil {
		os.RemoveAll(agentDir)
		return fmt.Errorf("failed to copy agent files: %w", err)
	}

	var commitHash string
	var repo *git.Repository

	// Handle metadata differently for local vs remote repositories
	if ad.detector.DetectProvider(repoURL) == "local" {
		// For local repositories, open repository directly
		var err error
		repo, err = git.PlainOpen(fullURL)
		if err != nil {
			// Non-fatal, continue without git metadata
			log.Printf("Warning: failed to open local repository for metadata: %v", err)
		}
	} else {
		// For remote repositories, use the already-cloned repository at repoPath
		var err error
		repo, err = git.PlainOpen(repoPath)
		if err != nil {
			// Non-fatal, continue without git metadata
			log.Printf("Warning: failed to open repository for metadata: %v", err)
		}
	}

	if repo != nil {
		if ref, err := repo.Head(); err == nil {
			commitHash = ref.Hash().String()
		}
	}

	// Create metadata
	metadata := map[string]interface{}{
		"name":         agentName,
		"source":       fullURL,
		"commit":       commitHash,
		"downloaded":   "now",
		"components":   1,
		"detection":    "single",
		"originalPath": targetComponent.FilePath,
	}

	// Save legacy metadata file for backward compatibility
	legacyMetadataFile := filepath.Join(agentDir, ".agent-metadata.json")
	if err := ad.saveMetadata(legacyMetadataFile, metadata); err != nil {
		log.Printf("Warning: failed to save legacy metadata: %v", err)
	}

	// Save to npx add-skill compatible lock file
	sourceType := "github"
	if strings.Contains(fullURL, "gitlab") {
		sourceType = "gitlab"
	} else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") {
		sourceType = "git"
	}

	// Compute folder hash if it's a GitHub repo
	var folderHash string
	if sourceType != "github" {
		if hash, err := computeLocalFolderHash(agentDir); err == nil {
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

func (cl *ComponentLinker) createSymlink(src, dst string) error {
	// Remove existing destination if it exists
	if _, err := os.Lstat(dst); err == nil {
		os.Remove(dst)
	}

	// Create relative path for the symlink
	relPath, err := filepath.Rel(filepath.Dir(dst), src)
	if err != nil {
		return fmt.Errorf("failed to create relative path: %w", err)
	}

	// Create the symbolic link
	if err := os.Symlink(relPath, dst); err != nil {
		// Try fallback to junction on Windows
		if runtime.GOOS == "windows" {
			return cl.createJunction(src, dst)
		}
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

func (cl *ComponentLinker) createJunction(src, dst string) error {
	// For Windows, we would need to use Windows API calls for proper junctions
	// For now, fall back to copying the directory as cross-platform solution
	return cl.copyDirectory(src, dst)
}

func (cl *ComponentLinker) copyDirectory(src, dst string) error {
	return fileutil.CopyDirectoryContents(src, dst)
}

func (cl *ComponentLinker) copyFile(src, dst string) error {
	return fileutil.CopyFile(src, dst)
}

func (cl *ComponentLinker) linkComponent(componentType, componentName string) error {
	srcDir := filepath.Join(cl.agentsDir, componentType, componentName)
	dstDir := filepath.Join(cl.opencodeDir, componentType, componentName)

	// Check if source component exists
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("component %s/%s does not exist in %s", componentType, componentName, cl.agentsDir)
	}

	// All components are now stored type-based, no special plugin handling needed
	metadata := cl.loadComponentMetadata(componentType, componentName)
	_ = metadata // Keep metadata loading for potential future use

	// Create destination directory
	if err := createDirectoryWithPermissions(filepath.Dir(dstDir)); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Create symlink or copy
	if err := cl.createSymlink(srcDir, dstDir); err != nil {
		return fmt.Errorf("failed to link component: %w", err)
	}

	fmt.Printf("Successfully linked %s '%s' to opencode\n", componentType, componentName)
	fmt.Printf("Source: %s\n", srcDir)
	fmt.Printf("Target: %s\n", dstDir)

	return nil
}

// loadComponentMetadata loads metadata for a component from lock files or metadata files
func (cl *ComponentLinker) loadComponentMetadata(componentType, componentName string) *models.ComponentMetadata {
	// Try lock file first
	if metadata := cl.loadFromLockFile(componentType, componentName); metadata != nil {
		return metadata
	}

	// Try legacy metadata file
	var metadataFile string

	switch componentType {
	case "skills":
		metadataFile = paths.GetComponentMetadataPath(cl.agentsDir, componentType, componentName)
	case "agents":
		metadataFile = paths.GetComponentMetadataPath(cl.agentsDir, componentType, componentName)
	case "commands":
		metadataFile = paths.GetComponentMetadataPath(cl.agentsDir, componentType, componentName)
	default:
		return nil
	}
	data, err := os.ReadFile(metadataFile)
	if err != nil {
		return nil
	}

	var metadata models.ComponentMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil
	}

	return &metadata
}

// loadFromLockFile loads metadata from lock file
func (cl *ComponentLinker) loadFromLockFile(componentType, componentName string) *models.ComponentMetadata {
	var lockFilePath string
	var entries map[string]models.ComponentLockEntry

	switch componentType {
	case "skills":
		lockFilePath = paths.GetComponentLockPath(cl.agentsDir, componentType)
	case "agents":
		lockFilePath = paths.GetComponentLockPath(cl.agentsDir, componentType)
	case "commands":
		lockFilePath = paths.GetComponentLockPath(cl.agentsDir, componentType)
	default:
		return nil
	}

	lockData, err := os.ReadFile(lockFilePath)
	if err != nil {
		return nil
	}

	var lockFile ComponentLockFile
	if err := json.Unmarshal(lockData, &lockFile); err != nil {
		return nil
	}

	switch componentType {
	case "skills":
		entries = lockFile.Skills
	case "agents":
		entries = lockFile.Agents
	case "commands":
		entries = lockFile.Commands
	}

	entry, exists := entries[componentName]
	if !exists {
		return nil
	}

	// Convert lock entry to metadata
	return &models.ComponentMetadata{
		Name:         componentName,
		Source:       entry.SourceUrl,
		Commit:       entry.SkillFolderHash,
		OriginalPath: entry.OriginalPath,
		Components:   entry.Components,
		Detection:    entry.Detection,
	}
}

func (cl *ComponentLinker) linkAllComponents() error {
	componentTypes := paths.GetComponentTypes()

	for _, componentType := range componentTypes {
		typeDir := filepath.Join(cl.agentsDir, componentType)
		if _, err := os.Stat(typeDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(typeDir)
		if err != nil {
			fmt.Printf("Warning: failed to read %s directory: %v\n", componentType, err)
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				componentName := entry.Name()

				// Skip monorepo containers - they shouldn't be linked as individual components
				if cl.isMonorepoContainer(componentType, componentName) {
					continue
				}

				// Link as a regular single component
				if err := cl.linkComponent(componentType, componentName); err != nil {
					fmt.Printf("Warning: failed to link %s/%s: %v\n", componentType, componentName, err)
				}
			}
		}
	}

	return nil
}

// isMonorepoContainer checks if a component directory contains other component directories
// and should not be linked as a single component
func (cl *ComponentLinker) isMonorepoContainer(componentType, componentName string) bool {
	componentDir := filepath.Join(cl.agentsDir, componentType, componentName)

	// Check if this directory contains other component directories
	entries, err := os.ReadDir(componentDir)
	if err != nil {
		return false
	}

	// Determine possible marker files for this component type
	var markerFiles []string
	switch componentType {
	case "skills":
		markerFiles = []string{"SKILL.md"}
	case "agents":
		markerFiles = []string{componentName + ".md"}
	case "commands":
		markerFiles = []string{componentName + ".md"}
	default:
		return false
	}

	// Count how many subdirectories contain a marker file
	subComponentCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			subDir := filepath.Join(componentDir, entry.Name())
			for _, markerFile := range markerFiles {
				if _, err := os.Stat(filepath.Join(subDir, markerFile)); err == nil {
					subComponentCount++
					break
				}
			}
		}
	}

	// If there are multiple sub-components, this is a monorepo container
	return subComponentCount > 1
}

// linkMonorepoComponents links individual components from a monorepo container
func (cl *ComponentLinker) linkMonorepoComponents(componentType, repoName string) error {
	repoDir := filepath.Join(cl.agentsDir, componentType, repoName)

	entries, err := os.ReadDir(repoDir)
	if err != nil {
		return fmt.Errorf("failed to read monorepo directory: %w", err)
	}

	// Determine possible marker files for this component type
	var markerFiles []string
	switch componentType {
	case "skills":
		markerFiles = []string{"SKILL.md"}
	case "agents":
		markerFiles = []string{}
	case "commands":
		markerFiles = []string{}
	default:
		return fmt.Errorf("unknown component type: %s", componentType)
	}

	linkedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			subComponentName := entry.Name()
			subComponentDir := filepath.Join(repoDir, subComponentName)

			// Check if this subdirectory contains any marker file or a {name}.md file
			hasMarker := false
			for _, markerFile := range markerFiles {
				if _, err := os.Stat(filepath.Join(subComponentDir, markerFile)); err == nil {
					hasMarker = true
					break
				}
			}

			// Also check for {name}.md pattern
			if !hasMarker {
				if _, err := os.Stat(filepath.Join(subComponentDir, subComponentName+".md")); err == nil {
					hasMarker = true
				}
			}

			if hasMarker {
				// Link this sub-component using a unique name that includes the repo name
				linkName := fmt.Sprintf("%s-%s", repoName, subComponentName)

				// Create the link from the sub-component directory
				srcDir := subComponentDir
				dstDir := filepath.Join(cl.opencodeDir, componentType, linkName)

				// Create destination directory
				if err := createDirectoryWithPermissions(filepath.Dir(dstDir)); err != nil {
					fmt.Printf("Warning: failed to create destination directory for %s: %v\n", linkName, err)
					continue
				}

				// Create symlink
				if err := cl.createSymlink(srcDir, dstDir); err != nil {
					fmt.Printf("Warning: failed to link monorepo component %s: %v\n", linkName, err)
					continue
				}

				fmt.Printf("Successfully linked monorepo component %s from %s\n", linkName, repoName)
				linkedCount++
			}
		}
	}

	if linkedCount > 0 {
		fmt.Printf("Linked %d components from monorepo %s\n", linkedCount, repoName)
	}

	return nil
}

func (cl *ComponentLinker) detectAndLinkLocalRepositories() error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Check if current directory is a git repository
	if !cl.detector.IsLocalPath(cwd) {
		return fmt.Errorf("current directory is not a git repository")
	}

	// Detect components in the current repository
	components, err := cl.detector.DetectComponentsInRepo(cwd)
	if err != nil {
		return fmt.Errorf("failed to detect components in repository: %w", err)
	}

	if len(components) == 0 {
		fmt.Println("No components detected in current repository")
		return nil
	}

	fmt.Printf("Detected %d components in current repository:\n", len(components))
	for _, component := range components {
		fmt.Printf("  - %s: %s (%s)\n", component.Type, component.Name, component.Path)
	}

	// Link each detected component
	for _, component := range components {
		componentTypeStr := string(component.Type)
		componentPath := filepath.Join(cwd, component.Path)

		// Create a temporary link to the detected component
		tempLinkName := fmt.Sprintf("auto-detected-%s", component.Name)
		tempLinkPath := filepath.Join(cl.agentsDir, componentTypeStr, tempLinkName)

		// Create destination directory
		if err := createDirectoryWithPermissions(filepath.Dir(tempLinkPath)); err != nil {
			fmt.Printf("Warning: failed to create directory for %s: %v\n", component.Name, err)
			continue
		}

		// Create symlink to the detected component
		if err := cl.createSymlink(componentPath, tempLinkPath); err != nil {
			fmt.Printf("Warning: failed to link component %s: %v\n", component.Name, err)
			continue
		}

		// Now link it to opencode
		if err := cl.linkComponent(componentTypeStr, tempLinkName); err != nil {
			fmt.Printf("Warning: failed to link %s to opencode: %v\n", component.Name, err)
			continue
		}

		fmt.Printf("✓ Automatically linked %s '%s' from current repository\n", component.Type, component.Name)
	}

	return nil
}

// LinkStatus represents the status of a linked component
type LinkStatus struct {
	Name       string
	Type       string
	LinkType   string // "symlink", "copied", "broken", "missing"
	Target     string
	Valid      bool
	TargetPath string
}

// listLinkedComponents lists all components linked to opencode
func (cl *ComponentLinker) listLinkedComponents() error {
	componentTypes := paths.GetComponentTypes()

	allLinks := make(map[string][]LinkStatus)
	totalCount := 0
	validCount := 0
	brokenCount := 0

	for _, componentType := range componentTypes {
		typeDir := filepath.Join(cl.opencodeDir, componentType)

		// Check if directory exists
		if _, err := os.Stat(typeDir); os.IsNotExist(err) {
			allLinks[componentType] = []LinkStatus{}
			continue
		}

		// Read directory entries
		entries, err := os.ReadDir(typeDir)
		if err != nil {
			return fmt.Errorf("failed to read %s directory: %w", componentType, err)
		}

		links := []LinkStatus{}
		for _, entry := range entries {
			// Skip hidden files
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			fullPath := filepath.Join(typeDir, entry.Name())
			linkType, target, valid := cl.analyzeLinkStatus(fullPath)

			status := LinkStatus{
				Name:       entry.Name(),
				Type:       componentType,
				LinkType:   linkType,
				Target:     target,
				Valid:      valid,
				TargetPath: fullPath,
			}

			links = append(links, status)
			totalCount++

			if valid {
				validCount++
			} else if linkType == "broken" {
				brokenCount++
			}
		}

		allLinks[componentType] = links
	}

	// Display results
	if totalCount == 0 {
		fmt.Println("No components are currently linked to opencode.")
		fmt.Printf("Link location: %s\n", cl.opencodeDir)
		return nil
	}

	// Display by type
	for _, componentType := range componentTypes {
		links := allLinks[componentType]
		if len(links) == 0 {
			continue
		}

		// Capitalize first letter for display
		displayType := strings.Title(componentType)
		fmt.Printf("\n%s (%d):\n", displayType, len(links))

		for _, link := range links {
			var symbol, statusMsg string

			switch link.LinkType {
			case "symlink":
				if link.Valid {
					symbol = "✓"
					statusMsg = fmt.Sprintf("→ %s", link.Target)
				} else {
					symbol = "✗"
					statusMsg = "[broken link]"
				}
			case "copied":
				symbol = "◆"
				statusMsg = "[copied directory]"
			case "broken":
				symbol = "✗"
				statusMsg = "[broken link]"
			case "missing":
				symbol = "?"
				statusMsg = "[unknown state]"
			default:
				symbol = "?"
				statusMsg = "[unknown]"
			}

			fmt.Printf("  %s %s %s\n", symbol, link.Name, statusMsg)
		}
	}

	// Summary
	fmt.Printf("\nTotal: %d components", totalCount)
	if brokenCount > 0 {
		fmt.Printf(" (%d valid, %d broken)", validCount, brokenCount)
	}
	fmt.Println()

	return nil
}

// analyzeLinkStatus analyzes the status of a link/directory
func (cl *ComponentLinker) analyzeLinkStatus(path string) (linkType string, target string, valid bool) {
	info, err := os.Lstat(path)
	if err != nil {
		return "missing", "", false
	}

	// Check if it's a symlink
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return "broken", "", false
		}

		// Resolve relative paths
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(path), target)
		}

		// Check if target exists
		if _, err := os.Stat(target); err == nil {
			return "symlink", target, true
		}
		return "broken", target, false
	}

	// If it's a directory, it's a copied component
	if info.IsDir() {
		return "copied", path, true
	}

	return "unknown", "", false
}

// unlinkComponent removes a linked component from opencode
func (cl *ComponentLinker) unlinkComponent(componentType, componentName string) error {
	// Validate component type
	if componentType != "skills" && componentType != "agents" && componentType != "commands" {
		return fmt.Errorf("invalid component type: %s (must be skills, agents, or commands)", componentType)
	}

	linkPath := filepath.Join(cl.opencodeDir, componentType, componentName)

	// Check if link exists
	if _, err := os.Lstat(linkPath); os.IsNotExist(err) {
		return fmt.Errorf("component %s/%s is not linked to opencode", componentType, componentName)
	}

	// Analyze what we're removing
	linkType, target, _ := cl.analyzeLinkStatus(linkPath)

	// For copied directories, ask for confirmation
	if linkType == "copied" {
		fmt.Printf("Warning: '%s' is a copied directory, not a symlink.\n", componentName)
		fmt.Printf("This will permanently delete: %s\n", linkPath)
		fmt.Print("Continue? [y/N]: ")

		var response string
		fmt.Scanln(&response)

		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("Unlink cancelled.")
			return nil
		}
	}

	// Remove the link or directory
	if linkType == "copied" {
		if err := os.RemoveAll(linkPath); err != nil {
			return fmt.Errorf("failed to remove copied directory: %w", err)
		}
	} else {
		// For symlinks and broken links
		if err := os.Remove(linkPath); err != nil {
			return fmt.Errorf("failed to remove link: %w", err)
		}
	}

	fmt.Printf("Successfully unlinked %s '%s' from opencode\n", componentType, componentName)

	if linkType == "symlink" && target != "" {
		fmt.Printf("Source still available at: %s\n", target)
	}

	return nil
}

// unlinkAllComponents removes all linked components from opencode
func (cl *ComponentLinker) unlinkAllComponents(force bool) error {
	componentTypes := paths.GetComponentTypes()

	// First, collect all symlinks (skip copied directories)
	totalLinks := 0
	copiedDirs := 0

	for _, componentType := range componentTypes {
		typeDir := filepath.Join(cl.opencodeDir, componentType)

		if _, err := os.Stat(typeDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(typeDir)
		if err != nil {
			return fmt.Errorf("failed to read %s directory: %w", componentType, err)
		}

		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			fullPath := filepath.Join(typeDir, entry.Name())
			linkType, _, _ := cl.analyzeLinkStatus(fullPath)

			if linkType == "copied" {
				copiedDirs++
				continue // Skip copied directories
			}
			totalLinks++
		}
	}

	if totalLinks == 0 && copiedDirs == 0 {
		fmt.Println("No linked components found.")
		return nil
	}

	// Require force flag or confirmation
	if !force {
		if totalLinks > 0 {
			fmt.Printf("This will unlink %d symlinked components from opencode", totalLinks)
			fmt.Println()
		}
		if copiedDirs > 0 {
			fmt.Printf("Note: %d copied directories will be skipped (not deleted)\n", copiedDirs)
		}
		if totalLinks == 0 {
			fmt.Println("No symlinked components to unlink (only copied directories found).")
			return nil
		}
		fmt.Print("Continue? [y/N]: ")

		var response string
		fmt.Scanln(&response)

		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("Unlink cancelled.")
			return nil
		}
	}

	// Remove all symlinks (skip copied directories)
	removedCount := 0
	skippedCount := 0
	errorCount := 0

	for _, componentType := range componentTypes {
		typeDir := filepath.Join(cl.opencodeDir, componentType)

		if _, err := os.Stat(typeDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(typeDir)
		if err != nil {
			fmt.Printf("Warning: failed to read %s directory: %v\n", componentType, err)
			continue
		}

		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			fullPath := filepath.Join(typeDir, entry.Name())
			linkType, _, _ := cl.analyzeLinkStatus(fullPath)

			// Skip copied directories - don't delete them
			if linkType == "copied" {
				skippedCount++
				continue
			}

			// Remove symlinks and broken links
			var err error
			err = os.Remove(fullPath)

			if err != nil {
				fmt.Printf("Warning: failed to unlink %s/%s: %v\n", componentType, entry.Name(), err)
				errorCount++
			} else {
				removedCount++
			}
		}
	}

	fmt.Printf("\nSuccessfully unlinked %d components", removedCount)
	if skippedCount > 0 {
		fmt.Printf(" (%d copied directories skipped)", skippedCount)
	}
	if errorCount > 0 {
		fmt.Printf(" (%d errors)", errorCount)
	}
	fmt.Println()

	return nil
}

func (ud *UpdateDetector) loadMetadata(componentType, componentName string) (*models.ComponentMetadata, error) {
	// First try to load from npx add-skill compatible lock files
	if metadata, err := ud.loadFromLockFile(componentType, componentName); err == nil {
		// Convert to legacy format for compatibility
		return &models.ComponentMetadata{
			Name:   componentName,
			Source: metadata.SourceUrl,
			Commit: metadata.SkillFolderHash,
		}, nil
	}

	// Fall back to legacy metadata files
	var metadataFile string
	switch componentType {
	case "skills":
		metadataFile = paths.GetComponentMetadataPath(ud.baseDir, componentType, componentName)
	case "agents":
		metadataFile = paths.GetComponentMetadataPath(ud.baseDir, componentType, componentName)
	case "commands":
		metadataFile = paths.GetComponentMetadataPath(ud.baseDir, componentType, componentName)
	default:
		return nil, fmt.Errorf("unknown component type: %s", componentType)
	}

	data, err := os.ReadFile(metadataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata models.ComponentMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

func (ud *UpdateDetector) loadFromLockFile(componentType, componentName string) (*models.ComponentLockEntry, error) {
	var lockFilePath string
	var entries map[string]models.ComponentLockEntry

	switch componentType {
	case "skills":
		lockFilePath = paths.GetComponentLockPath(ud.baseDir, componentType)
	case "agents":
		lockFilePath = paths.GetComponentLockPath(ud.baseDir, componentType)
	case "commands":
		lockFilePath = paths.GetComponentLockPath(ud.baseDir, componentType)
	default:
		return nil, fmt.Errorf("unknown component type: %s", componentType)
	}

	lockData, err := os.ReadFile(lockFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	var lockFile ComponentLockFile
	if err := json.Unmarshal(lockData, &lockFile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lock file: %w", err)
	}

	switch componentType {
	case "skills":
		entries = lockFile.Skills
	case "agents":
		entries = lockFile.Agents
	case "commands":
		entries = lockFile.Commands
	}

	entry, exists := entries[componentName]
	if !exists {
		return nil, fmt.Errorf("component %s not found in lock file", componentName)
	}

	return &entry, nil
}

func (ud *UpdateDetector) getCurrentRepoSHA(repoURL string) (string, error) {
	fullURL, err := ud.detector.NormalizeURL(repoURL)
	if err != nil {
		return "", fmt.Errorf("failed to normalize URL: %w", err)
	}

	// Create temporary directory for checking current state
	tempDir, err := os.MkdirTemp("", "agent-smith-check-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Clone repository to get current HEAD
	repo, err := git.PlainClone(tempDir, true, &git.CloneOptions{
		URL:           fullURL,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	// Get HEAD commit hash
	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	return ref.Hash().String(), nil
}

func (ud *UpdateDetector) HasUpdates(componentType, componentName, repoURL string) (bool, error) {
	// Load existing metadata
	metadata, err := ud.loadMetadata(componentType, componentName)
	if err != nil {
		return false, fmt.Errorf("failed to load metadata: %w", err)
	}

	// Get current repository SHA
	currentSHA, err := ud.getCurrentRepoSHA(repoURL)
	if err != nil {
		return false, fmt.Errorf("failed to get current repository SHA: %w", err)
	}

	// Compare stored SHA with current SHA
	return metadata.Commit != currentSHA, nil
}

func (ud *UpdateDetector) UpdateComponent(componentType, componentName, repoURL string) error {
	hasUpdates, err := ud.HasUpdates(componentType, componentName, repoURL)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !hasUpdates {
		fmt.Printf("Component %s/%s is already up to date\n", componentType, componentName)
		return nil
	}

	fmt.Printf("Updates detected for %s/%s, downloading new version...\n", componentType, componentName)

	// Remove old component directory to ensure clean re-clone
	componentDir := filepath.Join(ud.baseDir, componentType, componentName)
	if _, err := os.Stat(componentDir); err == nil {
		fmt.Printf("Removing old %s/%s directory...\n", componentType, componentName)
		if err := os.RemoveAll(componentDir); err != nil {
			return fmt.Errorf("failed to remove old component directory: %w", err)
		}
	}

	// Re-download the component with the latest changes
	switch componentType {
	case "skills":
		downloader := NewSkillDownloader()
		return downloader.downloadSkill(repoURL, componentName)
	case "agents":
		downloader := NewAgentDownloader()
		return downloader.downloadAgent(repoURL, componentName)
	case "commands":
		downloader := NewCommandDownloader()
		return downloader.downloadCommand(repoURL, componentName)
	default:
		return fmt.Errorf("unknown component type: %s", componentType)
	}
}

func (ud *UpdateDetector) UpdateAll() error {
	componentTypes := paths.GetComponentTypes()

	for _, componentType := range componentTypes {
		typeDir := filepath.Join(ud.baseDir, componentType)
		if _, err := os.Stat(typeDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(typeDir)
		if err != nil {
			fmt.Printf("Warning: failed to read %s directory: %v\n", componentType, err)
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				componentName := entry.Name()

				// Load metadata to get source URL
				metadata, err := ud.loadMetadata(componentType, componentName)
				if err != nil {
					fmt.Printf("Warning: failed to load metadata for %s/%s: %v\n", componentType, componentName, err)
					continue
				}

				if err := ud.UpdateComponent(componentType, componentName, metadata.Source); err != nil {
					fmt.Printf("Warning: failed to update %s/%s: %v\n", componentType, componentName, err)
				}
			}
		}
	}

	return nil
}

func (bd *BulkDownloader) AddAll(repoURL string) error {
	fullURL, err := bd.detector.NormalizeURL(repoURL)
	if err != nil {
		return fmt.Errorf("failed to normalize repository URL: %w", err)
	}

	// Create temporary directory for repository detection
	tempDir, err := os.MkdirTemp("", "agent-smith-bulk-*")
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
		return fmt.Errorf("failed to clone repository for bulk detection: %w", err)
	}

	// Detect all components in the repository from root
	components, err := bd.detector.DetectComponentsInRepo(tempDir)
	if err != nil {
		return fmt.Errorf("failed to detect components: %w", err)
	}

	if len(components) == 0 {
		return fmt.Errorf("no components (skills, agents, or commands) detected in repository")
	}

	return bd.processComponents(components, fullURL, repoURL, tempDir)
}

// processComponents handles downloading components from the repository
func (bd *BulkDownloader) processComponents(components []models.DetectedComponent, fullURL, repoURL, tempDir string) error {
	// Group components by type
	skillComponents := []models.DetectedComponent{}
	agentComponents := []models.DetectedComponent{}
	commandComponents := []models.DetectedComponent{}

	for _, comp := range components {
		switch comp.Type {
		case models.ComponentSkill:
			skillComponents = append(skillComponents, comp)
		case models.ComponentAgent:
			agentComponents = append(agentComponents, comp)
		case models.ComponentCommand:
			commandComponents = append(commandComponents, comp)
		}
	}

	// Download skills using optimized method with shared repository
	for _, comp := range skillComponents {
		fmt.Printf("Downloading: %s\n", comp.Name)
		if err := bd.skillDownloader.downloadSkillWithRepo(fullURL, comp.Name, repoURL, tempDir, components); err != nil {
			fmt.Printf("Warning: failed to download skill %s: %v\n", comp.Name, err)
		} else {
			fmt.Printf("Successfully downloaded skill: %s\n", comp.Name)
		}
	}

	// Download agents using optimized method with shared repository
	for _, comp := range agentComponents {
		fmt.Printf("Downloading: %s\n", comp.Name)
		if err := bd.agentDownloader.downloadAgentWithRepo(fullURL, comp.Name, repoURL, tempDir, components); err != nil {
			fmt.Printf("Warning: failed to download agent %s: %v\n", comp.Name, err)
		} else {
			fmt.Printf("Successfully downloaded agent: %s\n", comp.Name)
		}
	}

	// Download commands using optimized method with shared repository
	for _, comp := range commandComponents {
		fmt.Printf("Downloading: %s\n", comp.Name)
		if err := bd.commandDownloader.downloadCommandWithRepo(fullURL, comp.Name, repoURL, tempDir, components); err != nil {
			fmt.Printf("Warning: failed to download command %s: %v\n", comp.Name, err)
		} else {
			fmt.Printf("Successfully downloaded command: %s\n", comp.Name)
		}
	}

	totalComponents := len(skillComponents) + len(agentComponents) + len(commandComponents)
	fmt.Printf("Bulk download completed. Processed %d components:\n", totalComponents)
	fmt.Printf("  Skills: %d\n", len(skillComponents))
	fmt.Printf("  Agents: %d\n", len(agentComponents))
	fmt.Printf("  Commands: %d\n", len(commandComponents))

	return nil
}

// ComponentExecutor handles npx-like execution of components
type ComponentExecutor struct {
	detector   *detector.RepositoryDetector
	skillDir   string
	agentDir   string
	commandDir string
}

func NewComponentExecutor() *ComponentExecutor {
	skillDir, err := paths.GetSkillsDir()
	if err != nil {
		log.Fatal("Failed to get skills directory:", err)
	}

	agentDir, err := paths.GetAgentsSubDir()
	if err != nil {
		log.Fatal("Failed to get agents directory:", err)
	}

	commandDir, err := paths.GetCommandsDir()
	if err != nil {
		log.Fatal("Failed to get commands directory:", err)
	}

	return &ComponentExecutor{
		detector:   detector.NewRepositoryDetector(),
		skillDir:   skillDir,
		agentDir:   agentDir,
		commandDir: commandDir,
	}
}

// executeComponent provides npx-like functionality to run components without explicit installation
func executeComponent(target string, args []string) error {
	executor := NewComponentExecutor()

	// First, check if it's already installed locally
	if component, componentType, found := executor.findLocalComponent(target); found {
		return executor.runLocalComponent(component, componentType, args)
	}

	// If not found locally, try to interpret as a repository and install temporarily
	if strings.Contains(target, "/") {
		return executor.runFromRepository(target, args)
	}

	// If it's a simple name without "/", try to resolve as a known package
	return executor.resolveAndRunPackage(target, args)
}

func (ce *ComponentExecutor) findLocalComponent(name string) (string, string, bool) {
	// Check skills first
	skillPath := filepath.Join(ce.skillDir, name)
	if _, err := os.Stat(skillPath); err == nil {
		return skillPath, "skill", true
	}

	// Check agents
	agentPath := filepath.Join(ce.agentDir, name)
	if _, err := os.Stat(agentPath); err == nil {
		return agentPath, "agent", true
	}

	// Check commands
	commandPath := filepath.Join(ce.commandDir, name)
	if _, err := os.Stat(commandPath); err == nil {
		return commandPath, "command", true
	}

	return "", "", false
}

func (ce *ComponentExecutor) runLocalComponent(path, componentType string, args []string) error {
	// Look for executable files in the component directory
	executables, err := ce.findExecutables(path)
	if err != nil {
		return fmt.Errorf("failed to find executables in %s: %w", path, err)
	}

	if len(executables) == 0 {
		return fmt.Errorf("no executable found in component at %s", path)
	}

	// Prefer specific executable names based on component type
	var preferredExe string
	switch componentType {
	case "skill":
		preferredExe = ce.findExecutable(executables, []string{"skill", "run", "main", "index"})
	case "agent":
		preferredExe = ce.findExecutable(executables, []string{"agent", "run", "main", "index"})
	case "command":
		preferredExe = ce.findExecutable(executables, []string{"command", "run", "main", "index"})
	}

	if preferredExe == "" {
		preferredExe = executables[0] // Use first found if no preferred match
	}

	return ce.executeFile(preferredExe, args)
}

func (ce *ComponentExecutor) findExecutables(dir string) ([]string, error) {
	var executables []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if file is executable
		if runtime.GOOS != "windows" && info.Mode().Perm()&0111 != 0 {
			executables = append(executables, path)
			return nil
		}

		// On Windows or for scripts, check extensions
		ext := strings.ToLower(filepath.Ext(path))
		scriptExts := []string{".sh", ".py", ".js", ".go", ".ts"}
		for _, scriptExt := range scriptExts {
			if ext == scriptExt {
				executables = append(executables, path)
				break
			}
		}

		return nil
	})

	return executables, err
}

func (ce *ComponentExecutor) findExecutable(candidates []string, preferredNames []string) string {
	// Convert to lowercase for comparison
	preferredLower := make([]string, len(preferredNames))
	for i, name := range preferredNames {
		preferredLower[i] = strings.ToLower(name)
	}

	for _, candidate := range candidates {
		baseName := strings.ToLower(filepath.Base(candidate))
		baseNameNoExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))

		for _, preferred := range preferredLower {
			if baseNameNoExt == preferred {
				return candidate
			}
		}
	}

	return ""
}

func (ce *ComponentExecutor) executeFile(exePath string, args []string) error {
	ext := strings.ToLower(filepath.Ext(exePath))

	var cmdArgs []string

	switch ext {
	case ".sh":
		cmdArgs = append([]string{"bash", exePath}, args...)
	case ".py":
		cmdArgs = append([]string{"python3", exePath}, args...)
	case ".js":
		cmdArgs = append([]string{"node", exePath}, args...)
	case ".go":
		// For Go files, we need to compile and run
		return ce.compileAndRunGo(exePath, args)
	case ".ts":
		cmdArgs = append([]string{"npx", "tsx", exePath}, args...)
	default:
		// Direct execution for binaries
		cmdArgs = append([]string{exePath}, args...)
	}

	if len(cmdArgs) < 1 {
		return fmt.Errorf("no command to execute")
	}

	// Create and execute the command
	return ce.runCommand(cmdArgs[0], cmdArgs[1:]...)
}

func (ce *ComponentExecutor) compileAndRunGo(goFile string, args []string) error {
	// Create temporary directory for compilation
	tempDir, err := os.MkdirTemp("", "agent-smith-go-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Compile the Go file
	exePath := filepath.Join(tempDir, "run")
	cmd := exec.Command("go", "build", "-o", exePath, goFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to compile Go file: %w", err)
	}

	// Run the compiled binary
	return ce.runCommand(exePath, args...)
}

func (ce *ComponentExecutor) runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (ce *ComponentExecutor) runFromRepository(repoURL string, args []string) error {
	// Normalize repository URL
	fullURL, err := ce.detector.NormalizeURL(repoURL)
	if err != nil {
		return fmt.Errorf("invalid repository URL: %w", err)
	}

	// Create temporary directory for cloning
	tempDir, err := os.MkdirTemp("", "agent-smith-npx-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Clone repository
	_, err = git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:           fullURL,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Detect components in the repository
	components, err := ce.detector.DetectComponentsInRepo(tempDir)
	if err != nil {
		return fmt.Errorf("failed to detect components: %w", err)
	}

	if len(components) == 0 {
		return fmt.Errorf("no components found in repository %s", repoURL)
	}

	// Find the main/root component or use the first one
	var mainComponent *models.DetectedComponent
	for _, comp := range components {
		if comp.Name == "root-skill" || comp.Name == "root-agent" || comp.Name == "root-command" {
			mainComponent = &comp
			break
		}
	}

	if mainComponent == nil {
		mainComponent = &components[0] // Use first component if no root found
	}

	// Get the component path
	componentPath := filepath.Join(tempDir, mainComponent.Path)

	// Run the component
	switch mainComponent.Type {
	case models.ComponentSkill:
		return ce.runLocalComponent(componentPath, "skill", args)
	case models.ComponentAgent:
		return ce.runLocalComponent(componentPath, "agent", args)
	case models.ComponentCommand:
		return ce.runLocalComponent(componentPath, "command", args)
	default:
		return fmt.Errorf("unknown component type: %s", mainComponent.Type)
	}
}

func (ce *ComponentExecutor) resolveAndRunPackage(name string, args []string) error {
	// For now, try common GitHub prefixes for popular packages
	prefixes := []string{
		"agent-smith/",
		"opencode/",
		"npx/",
	}

	for _, prefix := range prefixes {
		repo := prefix + name
		err := ce.runFromRepository(repo, args)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("package '%s' not found locally and couldn't be resolved from common repositories", name)
}

func main() {
	// Set up handlers for Cobra commands
	cmd.SetHandlers(
		func(repoURL, name string) {
			downloader := NewSkillDownloader()
			if err := downloader.downloadSkill(repoURL, name); err != nil {
				log.Fatal("Failed to download skill:", err)
			}
		},
		func(repoURL, name string) {
			downloader := NewAgentDownloader()
			if err := downloader.downloadAgent(repoURL, name); err != nil {
				log.Fatal("Failed to download agent:", err)
			}
		},
		func(repoURL, name string) {
			downloader := NewCommandDownloader()
			if err := downloader.downloadCommand(repoURL, name); err != nil {
				log.Fatal("Failed to download command:", err)
			}
		},
		func(repoURL string) {
			bulkDownloader := NewBulkDownloader()
			if err := bulkDownloader.AddAll(repoURL); err != nil {
				log.Fatal("Failed to bulk download components:", err)
			}
		},
		func(target string, args []string) {
			if err := executeComponent(target, args); err != nil {
				log.Fatal("Failed to execute component:", err)
			}
		},
		func(componentType, componentName string) {
			// Validate component type
			if componentType != "skills" && componentType != "agents" && componentType != "commands" {
				log.Fatal("Invalid component type. Use: skills, agents, or commands")
			}

			detector := NewUpdateDetector()

			// Load metadata to get source URL
			metadata, err := detector.loadMetadata(componentType, componentName)
			if err != nil {
				log.Fatal("Failed to load component metadata:", err)
			}

			if err := detector.UpdateComponent(componentType, componentName, metadata.Source); err != nil {
				log.Fatal("Failed to update component:", err)
			}
		},
		func() {
			detector := NewUpdateDetector()
			if err := detector.UpdateAll(); err != nil {
				log.Fatal("Failed to update components:", err)
			}
		},
		func(componentType, componentName string) {
			linker := NewComponentLinker()
			if err := linker.linkComponent(componentType, componentName); err != nil {
				log.Fatal("Failed to link component:", err)
			}
		},
		func() {
			linker := NewComponentLinker()
			if err := linker.linkAllComponents(); err != nil {
				log.Fatal("Failed to link all components:", err)
			}
		},
		func() {
			linker := NewComponentLinker()
			if err := linker.detectAndLinkLocalRepositories(); err != nil {
				log.Fatal("Failed to auto-link repositories:", err)
			}
		},
		func() {
			linker := NewComponentLinker()
			if err := linker.listLinkedComponents(); err != nil {
				log.Fatal("Failed to list linked components:", err)
			}
		},
		func(componentType, componentName string) {
			linker := NewComponentLinker()
			if err := linker.unlinkComponent(componentType, componentName); err != nil {
				log.Fatal("Failed to unlink component:", err)
			}
		},
		func(force bool) {
			linker := NewComponentLinker()
			if err := linker.unlinkAllComponents(force); err != nil {
				log.Fatal("Failed to unlink all components:", err)
			}
		},
	)

	// Execute Cobra command
	cmd.Execute()
}
