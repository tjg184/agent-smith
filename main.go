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
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/tgaines/agent-smith/cmd"
)

const (
	agentsDir    = "~/.agents"
	skillsDir    = agentsDir + string(filepath.Separator) + "skills"
	agentsSubDir = agentsDir + string(filepath.Separator) + "agents"
	commandsDir  = agentsDir + string(filepath.Separator) + "commands"
	opencodeDir  = "~/.config" + string(filepath.Separator) + "opencode"
)

// ComponentDetectionPattern defines how to detect a component type
type ComponentDetectionPattern struct {
	Name           string   `json:"name"`
	ExactFiles     []string `json:"exactFiles"`     // Files that must match exactly (e.g., "SKILL.md")
	PathPatterns   []string `json:"pathPatterns"`   // Path patterns (e.g., "/agents/", "*/docs/*")
	FileExtensions []string `json:"fileExtensions"` // File extensions to match (e.g., ".md")
	IgnorePaths    []string `json:"ignorePaths"`    // Paths to ignore during detection
}

// DetectionConfig holds all component detection patterns
type DetectionConfig struct {
	Components map[string]ComponentDetectionPattern `json:"components"`
}

// RepositoryDetector maintains repository patterns and component detection
type RepositoryDetector struct {
	patterns        map[string]string
	detectionConfig *DetectionConfig
}

// createDefaultDetectionConfig returns the default component detection patterns
func createDefaultDetectionConfig() *DetectionConfig {
	return &DetectionConfig{
		Components: map[string]ComponentDetectionPattern{
			string(ComponentSkill): {
				Name:       "skill",
				ExactFiles: []string{"SKILL.md"},
				IgnorePaths: []string{
					".git",
					"node_modules",
					".vscode",
					".idea",
					"target",
					"build",
					"dist",
				},
			},
			string(ComponentAgent): {
				Name:           "agent",
				ExactFiles:     []string{"AGENT.md"},
				PathPatterns:   []string{"/agents/", "agents"},
				FileExtensions: []string{".md"},
				IgnorePaths: []string{
					".git",
					"node_modules",
					".vscode",
					".idea",
					"target",
					"build",
					"dist",
				},
			},
			string(ComponentCommand): {
				Name:           "command",
				ExactFiles:     []string{"COMMAND.md"},
				PathPatterns:   []string{"/commands/", "commands"},
				FileExtensions: []string{".md"},
				IgnorePaths: []string{
					".git",
					"node_modules",
					".vscode",
					".idea",
					"target",
					"build",
					"dist",
				},
			},
		},
	}
}

type ComponentType string

const (
	ComponentSkill   ComponentType = "skill"
	ComponentAgent   ComponentType = "agent"
	ComponentCommand ComponentType = "command"
)

type DetectedComponent struct {
	Type       ComponentType
	Name       string
	Path       string
	SourceFile string
}

type SkillDownloader struct {
	baseDir  string
	detector *RepositoryDetector
}

type AgentDownloader struct {
	baseDir  string
	detector *RepositoryDetector
}

type CommandDownloader struct {
	baseDir  string
	detector *RepositoryDetector
}

type ComponentLinker struct {
	agentsDir   string
	opencodeDir string
	detector    *RepositoryDetector
}

type BulkDownloader struct {
	skillDownloader   *SkillDownloader
	agentDownloader   *AgentDownloader
	commandDownloader *CommandDownloader
	detector          *RepositoryDetector
}

type UpdateDetector struct {
	baseDir  string
	detector *RepositoryDetector
}

type ComponentLockEntry struct {
	Source          string `json:"source"`
	SourceType      string `json:"sourceType"`
	SourceUrl       string `json:"sourceUrl"`
	SkillPath       string `json:"skillPath,omitempty"`
	SkillFolderHash string `json:"skillFolderHash"`
	InstalledAt     string `json:"installedAt"`
	UpdatedAt       string `json:"updatedAt"`
	Version         int    `json:"version"`
	Components      int    `json:"components,omitempty"`
	Detection       string `json:"detection,omitempty"`
}

type ComponentLockFile struct {
	Version  int                           `json:"version"`
	Skills   map[string]ComponentLockEntry `json:"skills"`
	Agents   map[string]ComponentLockEntry `json:"agents,omitempty"`
	Commands map[string]ComponentLockEntry `json:"commands,omitempty"`
}

// Legacy metadata structure for backward compatibility
type ComponentMetadata struct {
	Name       string `json:"name"`
	Source     string `json:"source"`
	Commit     string `json:"commit"`
	Downloaded string `json:"downloaded"`
	Components int    `json:"components,omitempty"`
	Detection  string `json:"detection,omitempty"`
}

// Cross-platform helper functions
func getCrossPlatformPermissions() os.FileMode {
	if runtime.GOOS == "windows" {
		return 0666 // Windows has less granular permissions
	}
	return 0755 // Unix-like systems
}

func getCrossPlatformFilePermissions() os.FileMode {
	if runtime.GOOS == "windows" {
		return 0644 // Windows has less granular permissions
	}
	return 0644 // Unix-like systems
}

func createDirectoryWithPermissions(path string) error {
	perm := getCrossPlatformPermissions()
	return os.MkdirAll(path, perm)
}

func createFileWithPermissions(path string, data []byte) error {
	perm := getCrossPlatformFilePermissions()
	return os.WriteFile(path, data, perm)
}

// loadDetectionConfig loads detection configuration from a JSON file
func (rd *RepositoryDetector) loadDetectionConfig(configPath string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file doesn't exist, use defaults
		rd.detectionConfig = createDefaultDetectionConfig()
		return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read detection config file %s: %v", configPath, err)
	}

	var config DetectionConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse detection config file %s: %v", configPath, err)
	}

	rd.detectionConfig = &config
	return nil
}

// saveDetectionConfig saves detection configuration to a JSON file
func (rd *RepositoryDetector) saveDetectionConfig(configPath string) error {
	if rd.detectionConfig == nil {
		return fmt.Errorf("no detection config to save")
	}

	data, err := json.MarshalIndent(rd.detectionConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal detection config: %v", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, getCrossPlatformPermissions()); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	return createFileWithPermissions(configPath, data)
}

// getDetectionConfigPath returns the default path for the detection configuration file
func getDetectionConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./detection-config.json"
	}
	return filepath.Join(homeDir, ".config", "opencode", "detection-config.json")
}

func NewRepositoryDetector() *RepositoryDetector {
	return NewRepositoryDetectorWithConfig("")
}

func NewRepositoryDetectorWithConfig(configPath string) *RepositoryDetector {
	rd := &RepositoryDetector{
		patterns: map[string]string{
			// GitHub patterns (most specific first)
			"github":     `^https?://(?:www\.)?github\.com/[^/]+/[^/]+$`,
			"github_git": `^(git@|ssh://)git@github\.com:[^/]+/[^/]+\.git$`,
			"github_api": `^https?://api\.github\.com/repos/[^/]+/[^/]+$`,

			// GitLab patterns
			"gitlab":     `^https?://(?:www\.)?gitlab\.com/[^/]+/[^/]+$`,
			"gitlab_git": `^(git@|ssh://)git@gitlab\.com:[^/]+/[^/]+\.git$`,
			"gitlab_api": `^https?://gitlab\.com/api/v4/projects/[^/]+$`,

			// Bitbucket patterns
			"bitbucket":     `^https?://(?:www\.)?bitbucket\.org/[^/]+/[^/]+$`,
			"bitbucket_git": `^(git@|ssh://)git@bitbucket\.org:[^/]+/[^/]+\.git$`,
			"bitbucket_api": `^https?://api\.bitbucket\.org/2\.0/repositories/[^/]+/[^/]+$`,

			// Generic git patterns (most generic last)
			"git_http": `^https?://(?!.*(?:github\.com|gitlab\.com|bitbucket\.org)).+$`,
			"git_ssh":  `^(ssh://|git@).+$`,
			"git":      `^(https?://|git@|ssh://).+\.git$`,
		},
	}

	// Load detection configuration
	if configPath == "" {
		configPath = getDetectionConfigPath()
	}

	if err := rd.loadDetectionConfig(configPath); err != nil {
		// If loading fails, use default config
		rd.detectionConfig = createDefaultDetectionConfig()
	}

	return rd
}

func (rd *RepositoryDetector) isLocalPath(path string) bool {
	path = strings.TrimSpace(path)

	// Check for absolute Unix paths
	if strings.HasPrefix(path, "/") {
		// Verify it looks like a valid path and exists
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	// Check for Windows paths
	if len(path) > 1 && path[1] == ':' {
		// C:\... or C:/... format
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	// Check for Windows UNC paths
	if strings.HasPrefix(path, "\\\\") {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	// Check for relative paths that might exist locally
	if !strings.Contains(path, "://") && !strings.HasPrefix(path, "git@") {
		// Try expanding to absolute path
		if absPath, err := filepath.Abs(path); err == nil {
			if _, err := os.Stat(absPath); err == nil {
				return true
			}
		}
	}

	return false
}

func (rd *RepositoryDetector) detectProvider(repoURL string) string {
	repoURL = strings.TrimSpace(repoURL)

	// Check for local paths first (most specific)
	if rd.isLocalPath(repoURL) {
		return "local"
	}

	// Check for specific providers in order of specificity
	for provider, pattern := range rd.patterns {
		if matched, _ := regexp.MatchString(pattern, repoURL); matched {
			// Normalize provider names (remove suffixes like _git, _api)
			if strings.HasSuffix(provider, "_git") {
				return strings.TrimSuffix(provider, "_git")
			}
			if strings.HasSuffix(provider, "_api") {
				return strings.TrimSuffix(provider, "_api")
			}
			if strings.Contains(provider, "_") {
				parts := strings.Split(provider, "_")
				return parts[0]
			}
			return provider
		}
	}

	return "generic"
}

func (rd *RepositoryDetector) normalizeURL(repoURL string) (string, error) {
	repoURL = strings.TrimSpace(repoURL)

	// Handle local paths
	if rd.isLocalPath(repoURL) {
		absPath, err := filepath.Abs(repoURL)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path for local repository: %w", err)
		}

		// Verify it's a valid git repository
		if _, err := os.Stat(filepath.Join(absPath, ".git")); os.IsNotExist(err) {
			return "", fmt.Errorf("local path is not a git repository: %s", absPath)
		}

		return absPath, nil
	}

	// If it's already a full URL or SSH/Git format, validate and return as-is
	if strings.HasPrefix(repoURL, "http://") || strings.HasPrefix(repoURL, "https://") ||
		strings.HasPrefix(repoURL, "git@") || strings.HasPrefix(repoURL, "ssh://") {

		// Basic URL validation
		if strings.HasPrefix(repoURL, "http") {
			if !strings.Contains(repoURL, "://") {
				return "", fmt.Errorf("invalid URL format: %s", repoURL)
			}
		}

		return repoURL, nil
	}

	// Handle GitHub shorthand (owner/repo)
	if !strings.Contains(repoURL, "/") {
		return "", fmt.Errorf("invalid repository format: %s", repoURL)
	}

	parts := strings.Split(repoURL, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid repository format: %s", repoURL)
	}

	// Validate shorthand format
	if parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("invalid repository format: %s", repoURL)
	}

	// Default to GitHub for shorthand notation
	return fmt.Sprintf("https://github.com/%s", repoURL), nil
}

func (rd *RepositoryDetector) validateRepository(repoURL string) error {
	provider := rd.detectProvider(repoURL)

	switch provider {
	case "local":
		// For local paths, check if it's a valid git repository
		absPath, err := filepath.Abs(repoURL)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}

		if _, err := os.Stat(filepath.Join(absPath, ".git")); os.IsNotExist(err) {
			return fmt.Errorf("local path is not a git repository: %s", absPath)
		}

		// Check if directory is accessible
		if _, err := os.Stat(absPath); err != nil {
			return fmt.Errorf("cannot access local repository: %w", err)
		}

	case "github", "gitlab", "bitbucket":
		// For known providers, validate the URL format
		if !strings.Contains(repoURL, "/") {
			return fmt.Errorf("invalid repository URL format: %s", repoURL)
		}

		// Additional validation for HTTP/HTTPS URLs
		if strings.HasPrefix(repoURL, "http") {
			if !strings.HasPrefix(repoURL, "http://") && !strings.HasPrefix(repoURL, "https://") {
				return fmt.Errorf("invalid protocol in URL: %s", repoURL)
			}
		}

	case "generic", "git":
		// For generic git URLs, do basic validation
		if !strings.HasPrefix(repoURL, "http://") && !strings.HasPrefix(repoURL, "https://") &&
			!strings.HasPrefix(repoURL, "git@") && !strings.HasPrefix(repoURL, "ssh://") &&
			!strings.HasSuffix(repoURL, ".git") {
			return fmt.Errorf("unrecognized repository format: %s", repoURL)
		}

	default:
		return fmt.Errorf("unsupported repository type: %s", provider)
	}

	return nil
}

// shouldIgnorePath checks if a path should be ignored during detection
func (rd *RepositoryDetector) shouldIgnorePath(relPath string, ignorePaths []string) bool {
	for _, ignorePath := range ignorePaths {
		if strings.Contains(relPath, ignorePath) {
			return true
		}
	}
	return false
}

// matchesExactFile checks if the filename matches any exact file patterns
func (rd *RepositoryDetector) matchesExactFile(fileName string, exactFiles []string) bool {
	for _, exactFile := range exactFiles {
		if fileName == exactFile {
			return true
		}
	}
	return false
}

// matchesPathPattern checks if the relative path matches any path patterns
func (rd *RepositoryDetector) matchesPathPattern(relPath string, pathPatterns []string) bool {
	for _, pattern := range pathPatterns {
		if strings.Contains(relPath, pattern) || strings.HasSuffix(relPath, pattern) {
			return true
		}
	}
	return false
}

// matchesFileExtension checks if the file has any of the specified extensions
func (rd *RepositoryDetector) matchesFileExtension(fileName string, fileExtensions []string) bool {
	for _, ext := range fileExtensions {
		if strings.HasSuffix(fileName, ext) {
			return true
		}
	}
	return false
}

// detectComponentForPattern checks if a file matches a component detection pattern
func (rd *RepositoryDetector) detectComponentForPattern(fileName, relPath string, pattern ComponentDetectionPattern, componentType ComponentType) (string, bool) {
	// Check if path should be ignored
	if rd.shouldIgnorePath(relPath, pattern.IgnorePaths) {
		return "", false
	}

	// Check exact file matches first (highest priority)
	if rd.matchesExactFile(fileName, pattern.ExactFiles) {
		return filepath.Base(filepath.Dir(relPath)), true
	}

	// Check path patterns with file extensions (medium priority)
	if len(pattern.PathPatterns) > 0 && len(pattern.FileExtensions) > 0 {
		if rd.matchesPathPattern(relPath, pattern.PathPatterns) && rd.matchesFileExtension(fileName, pattern.FileExtensions) {
			return strings.TrimSuffix(fileName, filepath.Ext(fileName)), true
		}
	}

	// Check just path patterns (lower priority)
	if len(pattern.PathPatterns) > 0 && rd.matchesPathPattern(relPath, pattern.PathPatterns) {
		return strings.TrimSuffix(fileName, filepath.Ext(fileName)), true
	}

	return "", false
}

func (rd *RepositoryDetector) detectComponentsInRepo(repoPath string) ([]DetectedComponent, error) {
	var components []DetectedComponent
	seenComponents := make(map[string]bool) // Prevent duplicates

	// Walk the repository to detect components
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fileName := filepath.Base(path)
		parentDir := filepath.Dir(path)
		relPath, err := filepath.Rel(repoPath, parentDir)
		if err != nil {
			return err
		}

		// Full relative path including filename for path-based detection
		fullRelPath, err := filepath.Rel(repoPath, path)
		if err != nil {
			return err
		}

		// Check each component type using its detection pattern
		for componentTypeStr, pattern := range rd.detectionConfig.Components {
			componentType := ComponentType(componentTypeStr)

			if componentName, matched := rd.detectComponentForPattern(fileName, relPath, pattern, componentType); matched {
				// Handle default component names
				if componentName == "" || componentName == "." {
					componentName = fmt.Sprintf("root-%s", pattern.Name)
				}

				componentKey := fmt.Sprintf("%s-%s", pattern.Name, componentName)
				if !seenComponents[componentKey] {
					components = append(components, DetectedComponent{
						Type:       componentType,
						Name:       componentName,
						Path:       relPath,
						SourceFile: fileName,
					})
					seenComponents[componentKey] = true
				}
			}
		}

		// Additional agent detection for .md files in /agents/ paths
		if strings.HasSuffix(fileName, ".md") && strings.Contains(fullRelPath, "/agents/") {
			componentName := strings.TrimSuffix(fileName, ".md")
			if componentName == "" || componentName == "." {
				componentName = "root-agent"
			}
			componentKey := fmt.Sprintf("agent-%s", componentName)
			if !seenComponents[componentKey] {
				components = append(components, DetectedComponent{
					Type:       ComponentAgent,
					Name:       componentName,
					Path:       relPath,
					SourceFile: fileName,
				})
				seenComponents[componentKey] = true
			}
		}

		// Additional command detection for .md files in /commands/ paths
		if strings.HasSuffix(fileName, ".md") && strings.Contains(fullRelPath, "/commands/") {
			componentName := strings.TrimSuffix(fileName, ".md")
			if componentName == "" || componentName == "." {
				componentName = "root-command"
			}
			componentKey := fmt.Sprintf("command-%s", componentName)
			if !seenComponents[componentKey] {
				components = append(components, DetectedComponent{
					Type:       ComponentCommand,
					Name:       componentName,
					Path:       relPath,
					SourceFile: fileName,
				})
				seenComponents[componentKey] = true
			}
		}

		return nil
	})

	return components, err
}

func NewSkillDownloader() *SkillDownloader {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get user home directory:", err)
	}

	baseDir := filepath.Join(home, ".agents", "skills")

	// Create base directory if it doesn't exist
	if err := createDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create skills directory:", err)
	}

	return &SkillDownloader{
		baseDir:  baseDir,
		detector: NewRepositoryDetector(),
	}
}

func NewAgentDownloader() *AgentDownloader {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get user home directory:", err)
	}

	baseDir := filepath.Join(home, ".agents", "agents")

	// Create base directory if it doesn't exist
	if err := createDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create agents directory:", err)
	}

	return &AgentDownloader{
		baseDir:  baseDir,
		detector: NewRepositoryDetector(),
	}
}

func NewCommandDownloader() *CommandDownloader {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get user home directory:", err)
	}

	baseDir := filepath.Join(home, ".agents", "commands")

	// Create base directory if it doesn't exist
	if err := createDirectoryWithPermissions(baseDir); err != nil {
		log.Fatal("Failed to create commands directory:", err)
	}

	return &CommandDownloader{
		baseDir:  baseDir,
		detector: NewRepositoryDetector(),
	}
}

func NewComponentLinker() *ComponentLinker {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get user home directory:", err)
	}

	agentsDir := filepath.Join(home, ".agents")
	opencodeDir := filepath.Join(home, ".config", "opencode")

	// Create opencode directory if it doesn't exist
	if err := createDirectoryWithPermissions(opencodeDir); err != nil {
		log.Fatal("Failed to create opencode directory:", err)
	}

	return &ComponentLinker{
		agentsDir:   agentsDir,
		opencodeDir: opencodeDir,
		detector:    NewRepositoryDetector(),
	}
}

func NewUpdateDetector() *UpdateDetector {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get user home directory:", err)
	}

	baseDir := filepath.Join(home, ".agents")

	return &UpdateDetector{
		baseDir:  baseDir,
		detector: NewRepositoryDetector(),
	}
}

func NewBulkDownloader() *BulkDownloader {
	return &BulkDownloader{
		skillDownloader:   NewSkillDownloader(),
		agentDownloader:   NewAgentDownloader(),
		commandDownloader: NewCommandDownloader(),
		detector:          NewRepositoryDetector(),
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
	normalizedURL, err := sd.detector.normalizeURL(repoURL)
	if err != nil {
		return "", err
	}

	// Validate the normalized repository
	if err := sd.detector.validateRepository(normalizedURL); err != nil {
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
	} else if sd.detector.detectProvider(repoURL) == "local" {
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
	components, err := sd.detector.detectComponentsInRepo(repoPath)
	if err != nil {
		return fmt.Errorf("failed to detect components: %w", err)
	}

	// Filter for skill components
	var skillComponents []DetectedComponent
	for _, comp := range components {
		if comp.Type == ComponentSkill {
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

	// If only one skill component found, copy its contents
	if len(skillComponents) == 1 {
		component := skillComponents[0]
		componentPath := filepath.Join(repoPath, component.Path)

		// Copy component contents to skill directory
		err = sd.copyDirectoryContents(componentPath, skillDir)
		if err != nil {
			os.RemoveAll(skillDir)
			return fmt.Errorf("failed to copy skill contents: %w", err)
		}
	} else {
		// Multiple skills found, create a monorepo structure
		for _, component := range skillComponents {
			componentDir := filepath.Join(skillDir, component.Name)
			componentPath := filepath.Join(repoPath, component.Path)

			err = createDirectoryWithPermissions(componentDir)
			if err != nil {
				continue
			}

			err = sd.copyDirectoryContents(componentPath, componentDir)
			if err != nil {
				log.Printf("Warning: failed to copy skill %s: %v", component.Name, err)
			}
		}
	}

	var commitHash string
	var repo *git.Repository

	// Handle metadata differently for local vs remote repositories
	if sd.detector.detectProvider(repoURL) == "local" {
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
	if sd.detector.detectProvider(repoURL) == "local" {
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

	if err := sd.saveLockFile(skillName, fullURL, sourceType, fullURL, folderHash, len(skillComponents), "recursive"); err != nil {
		log.Printf("Warning: failed to save lock file: %v", err)
	}

	// Clean up git clone only for remote repositories
	if sd.detector.detectProvider(repoURL) != "local" {
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
	if sd.detector.detectProvider(repoURL) == "local" {
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

	if err := sd.saveLockFile(skillName, fullURL, sourceType, fullURL, folderHash, 1, "direct"); err != nil {
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
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return createDirectoryWithPermissions(dstPath)
		}

		return sd.copyFile(path, dstPath)
	})
}

func (sd *SkillDownloader) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return createFileWithPermissions(dst, data)
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
func (sd *SkillDownloader) saveLockFile(skillName string, source string, sourceType string, sourceUrl string, skillFolderHash string, components int, detection string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	agentsDir := filepath.Join(home, ".agents")
	if err := createDirectoryWithPermissions(agentsDir); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	lockFilePath := filepath.Join(agentsDir, ".skill-lock.json")

	// Read existing lock file or create new one
	var lockFile ComponentLockFile
	lockData, err := os.ReadFile(lockFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			lockFile = ComponentLockFile{
				Version: 3, // Current version matching npx add-skill
				Skills:  make(map[string]ComponentLockEntry),
			}
		} else {
			return fmt.Errorf("failed to read lock file: %w", err)
		}
	} else {
		if err := json.Unmarshal(lockData, &lockFile); err != nil {
			// If lock file is corrupted, create new one
			lockFile = ComponentLockFile{
				Version: 3,
				Skills:  make(map[string]ComponentLockEntry),
			}
		}
		// Ensure version is current
		if lockFile.Version < 3 {
			lockFile.Version = 3
			if lockFile.Skills == nil {
				lockFile.Skills = make(map[string]ComponentLockEntry)
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
	lockFile.Skills[skillName] = ComponentLockEntry{
		Source:          source,
		SourceType:      sourceType,
		SourceUrl:       sourceUrl,
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

func (sd *SkillDownloader) downloadSkillWithRepo(fullURL, skillName, repoURL string, repoPath string, components []DetectedComponent) error {
	// Filter for skill components from provided components
	var skillComponents []DetectedComponent
	for _, comp := range components {
		if comp.Type == ComponentSkill {
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

	// If only one skill component found, copy its contents
	if len(skillComponents) == 1 {
		component := skillComponents[0]
		componentPath := filepath.Join(repoPath, component.Path)

		// Copy component contents to skill directory
		err := sd.copyDirectoryContents(componentPath, skillDir)
		if err != nil {
			os.RemoveAll(skillDir)
			return fmt.Errorf("failed to copy skill contents: %w", err)
		}
	} else {
		// Multiple skills found, create a monorepo structure
		for _, component := range skillComponents {
			componentDir := filepath.Join(skillDir, component.Name)
			componentPath := filepath.Join(repoPath, component.Path)

			err := createDirectoryWithPermissions(componentDir)
			if err != nil {
				continue
			}

			err = sd.copyDirectoryContents(componentPath, componentDir)
			if err != nil {
				log.Printf("Warning: failed to copy skill %s: %v", component.Name, err)
			}
		}
	}

	var commitHash string
	var repo *git.Repository

	// Handle metadata differently for local vs remote repositories
	if sd.detector.detectProvider(repoURL) == "local" {
		// For local repositories, open the repository directly
		var err error
		repo, err = git.PlainOpen(fullURL)
		if err != nil {
			// Non-fatal, continue without git metadata
			log.Printf("Warning: failed to open local repository for metadata: %v", err)
		}
	} else {
		// For remote repositories, clone to get git history for metadata
		var err error
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
	if sd.detector.detectProvider(repoURL) == "local" {
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

	if err := sd.saveLockFile(skillName, fullURL, sourceType, fullURL, folderHash, len(skillComponents), "recursive"); err != nil {
		log.Printf("Warning: failed to save lock file: %v", err)
	}

	// Clean up git clone only for remote repositories
	if sd.detector.detectProvider(repoURL) != "local" {
		if _, err := os.Stat(skillDir + ".git"); err == nil {
			os.RemoveAll(skillDir + ".git")
		}
	}

	fmt.Printf("Successfully downloaded skill '%s' from %s\n", skillName, fullURL)
	fmt.Printf("Skill stored in: %s\n", skillDir)
	fmt.Printf("Components detected: %d\n", len(skillComponents))

	return nil
}

func (cd *CommandDownloader) parseRepoURL(repoURL string) (string, error) {
	// Normalize URL first (handles GitHub shorthand, etc.)
	normalizedURL, err := cd.detector.normalizeURL(repoURL)
	if err != nil {
		return "", err
	}

	// Validate normalized repository
	if err := cd.detector.validateRepository(normalizedURL); err != nil {
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
	} else if cd.detector.detectProvider(repoURL) == "local" {
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
	components, err := cd.detector.detectComponentsInRepo(repoPath)
	if err != nil {
		return fmt.Errorf("failed to detect components: %w", err)
	}

	// Filter for command components
	var commandComponents []DetectedComponent
	for _, comp := range components {
		if comp.Type == ComponentCommand {
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

	// If only one command component found, copy its contents
	if len(commandComponents) == 1 {
		component := commandComponents[0]
		componentPath := filepath.Join(repoPath, component.Path)

		// Copy component contents to command directory
		err = cd.copyDirectoryContents(componentPath, commandDir)
		if err != nil {
			os.RemoveAll(commandDir)
			return fmt.Errorf("failed to copy command contents: %w", err)
		}
	} else {
		// Multiple commands found, create a monorepo structure
		for _, component := range commandComponents {
			componentDir := filepath.Join(commandDir, component.Name)
			componentPath := filepath.Join(repoPath, component.Path)

			err = createDirectoryWithPermissions(componentDir)
			if err != nil {
				continue
			}

			err = cd.copyDirectoryContents(componentPath, componentDir)
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
	if sourceType == "github" {
		if strings.HasPrefix(fullURL, "https://github.com/") {
			ownerRepo := strings.TrimPrefix(fullURL, "https://github.com/")
			ownerRepo = strings.TrimSuffix(ownerRepo, ".git")
			if hash, err := computeGitHubTreeSHA(ownerRepo, "COMMAND.md"); err == nil {
				folderHash = hash
			}
		}
	} else {
		if hash, err := computeLocalFolderHash(commandDir); err == nil {
			folderHash = hash
		}
	}

	if err := cd.saveLockFile(commandName, fullURL, sourceType, fullURL, folderHash, len(commandComponents), "recursive"); err != nil {
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
	if sourceType == "github" {
		if strings.HasPrefix(fullURL, "https://github.com/") {
			ownerRepo := strings.TrimPrefix(fullURL, "https://github.com/")
			ownerRepo = strings.TrimSuffix(ownerRepo, ".git")
			if hash, err := computeGitHubTreeSHA(ownerRepo, "COMMAND.md"); err == nil {
				folderHash = hash
			}
		}
	} else {
		if hash, err := computeLocalFolderHash(commandDir); err == nil {
			folderHash = hash
		}
	}

	if err := cd.saveLockFile(commandName, fullURL, sourceType, fullURL, folderHash, 1, "direct"); err != nil {
		log.Printf("Warning: failed to save lock file: %v", err)
	}

	// Create COMMAND.md if it doesn't exist
	commandFile := filepath.Join(commandDir, "COMMAND.md")
	if _, err := os.Stat(commandFile); os.IsNotExist(err) {
		if err := cd.createCommandFile(commandFile, commandName, fullURL); err != nil {
			log.Printf("Warning: failed to create COMMAND.md: %v", err)
		}
	}

	return nil
}

func (cd *CommandDownloader) copyDirectoryContents(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return createDirectoryWithPermissions(dstPath)
		}

		return cd.copyFile(path, dstPath)
	})
}

func (cd *CommandDownloader) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return createFileWithPermissions(dst, data)
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
func (cd *CommandDownloader) saveLockFile(commandName string, source string, sourceType string, sourceUrl string, skillFolderHash string, components int, detection string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	agentsDir := filepath.Join(home, ".agents")
	if err := createDirectoryWithPermissions(agentsDir); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	lockFilePath := filepath.Join(agentsDir, ".command-lock.json")

	// Read existing lock file or create new one
	var lockFile ComponentLockFile
	lockData, err := os.ReadFile(lockFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			lockFile = ComponentLockFile{
				Version:  3,
				Commands: make(map[string]ComponentLockEntry),
			}
		} else {
			return fmt.Errorf("failed to read lock file: %w", err)
		}
	} else {
		if err := json.Unmarshal(lockData, &lockFile); err != nil {
			lockFile = ComponentLockFile{
				Version:  3,
				Commands: make(map[string]ComponentLockEntry),
			}
		}
		if lockFile.Version < 3 {
			lockFile.Version = 3
			if lockFile.Commands == nil {
				lockFile.Commands = make(map[string]ComponentLockEntry)
			}
		}
	}

	now := time.Now().Format(time.RFC3339)

	existingEntry, exists := lockFile.Commands[commandName]
	if !exists {
		existingEntry.InstalledAt = now
	}

	lockFile.Commands[commandName] = ComponentLockEntry{
		Source:          source,
		SourceType:      sourceType,
		SourceUrl:       sourceUrl,
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

func (cd *CommandDownloader) downloadCommandWithRepo(fullURL, commandName, repoURL string, repoPath string, components []DetectedComponent) error {
	// Filter for command components from provided components
	var commandComponents []DetectedComponent
	for _, comp := range components {
		if comp.Type == ComponentCommand {
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

	// If only one command component found, copy its contents
	if len(commandComponents) == 1 {
		component := commandComponents[0]
		componentPath := filepath.Join(repoPath, component.Path)

		// Copy component contents to command directory
		err := cd.copyDirectoryContents(componentPath, commandDir)
		if err != nil {
			os.RemoveAll(commandDir)
			return fmt.Errorf("failed to copy command contents: %w", err)
		}
	} else {
		// Multiple commands found, create a monorepo structure
		for _, component := range commandComponents {
			componentDir := filepath.Join(commandDir, component.Name)
			componentPath := filepath.Join(repoPath, component.Path)

			err := createDirectoryWithPermissions(componentDir)
			if err != nil {
				continue
			}

			err = cd.copyDirectoryContents(componentPath, componentDir)
			if err != nil {
				log.Printf("Warning: failed to copy command %s: %v", component.Name, err)
			}
		}
	}

	var commitHash string
	var repo *git.Repository

	// For remote repositories, clone to get proper git history for metadata
	var err error
	repo, err = git.PlainClone(commandDir+".git", true, &git.CloneOptions{
		URL:           fullURL,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	})
	if err != nil {
		// Non-fatal, continue without git metadata
		log.Printf("Warning: failed to clone repository for metadata: %v", err)
	}

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
	if sourceType == "github" {
		if strings.HasPrefix(fullURL, "https://github.com/") {
			ownerRepo := strings.TrimPrefix(fullURL, "https://github.com/")
			ownerRepo = strings.TrimSuffix(ownerRepo, ".git")
			if hash, err := computeGitHubTreeSHA(ownerRepo, "COMMAND.md"); err == nil {
				folderHash = hash
			}
		}
	} else {
		if hash, err := computeLocalFolderHash(commandDir); err == nil {
			folderHash = hash
		}
	}

	if err := cd.saveLockFile(commandName, fullURL, sourceType, fullURL, folderHash, len(commandComponents), "recursive"); err != nil {
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

func (ad *AgentDownloader) parseRepoURL(repoURL string) (string, error) {
	// Normalize URL first (handles GitHub shorthand, etc.)
	normalizedURL, err := ad.detector.normalizeURL(repoURL)
	if err != nil {
		return "", err
	}

	// Validate normalized repository
	if err := ad.detector.validateRepository(normalizedURL); err != nil {
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
	} else if ad.detector.detectProvider(repoURL) == "local" {
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
	components, err := ad.detector.detectComponentsInRepo(repoPath)
	if err != nil {
		return fmt.Errorf("failed to detect components: %w", err)
	}

	// Filter for agent components
	var agentComponents []DetectedComponent
	for _, comp := range components {
		if comp.Type == ComponentAgent {
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

	// If only one agent component found, copy its contents
	if len(agentComponents) == 1 {
		component := agentComponents[0]
		componentPath := filepath.Join(repoPath, component.Path)

		// Copy component contents to agent directory
		err = ad.copyDirectoryContents(componentPath, agentDir)
		if err != nil {
			os.RemoveAll(agentDir)
			return fmt.Errorf("failed to copy agent contents: %w", err)
		}
	} else {
		// Multiple agents found, create a monorepo structure
		for _, component := range agentComponents {
			componentDir := filepath.Join(agentDir, component.Name)
			componentPath := filepath.Join(repoPath, component.Path)

			err = createDirectoryWithPermissions(componentDir)
			if err != nil {
				continue
			}

			err = ad.copyDirectoryContents(componentPath, componentDir)
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
	if sourceType == "github" {
		if strings.HasPrefix(fullURL, "https://github.com/") {
			ownerRepo := strings.TrimPrefix(fullURL, "https://github.com/")
			ownerRepo = strings.TrimSuffix(ownerRepo, ".git")
			if hash, err := computeGitHubTreeSHA(ownerRepo, "AGENT.md"); err == nil {
				folderHash = hash
			}
		}
	} else {
		if hash, err := computeLocalFolderHash(agentDir); err == nil {
			folderHash = hash
		}
	}

	if err := ad.saveLockFile(agentName, fullURL, sourceType, fullURL, folderHash, len(agentComponents), "recursive"); err != nil {
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
	if sourceType == "github" {
		if strings.HasPrefix(fullURL, "https://github.com/") {
			ownerRepo := strings.TrimPrefix(fullURL, "https://github.com/")
			ownerRepo = strings.TrimSuffix(ownerRepo, ".git")
			if hash, err := computeGitHubTreeSHA(ownerRepo, "AGENT.md"); err == nil {
				folderHash = hash
			}
		}
	} else {
		if hash, err := computeLocalFolderHash(agentDir); err == nil {
			folderHash = hash
		}
	}

	if err := ad.saveLockFile(agentName, fullURL, sourceType, fullURL, folderHash, 1, "direct"); err != nil {
		log.Printf("Warning: failed to save lock file: %v", err)
	}

	// Create AGENT.md if it doesn't exist
	agentFile := filepath.Join(agentDir, "AGENT.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		if err := ad.createAgentFile(agentFile, agentName, fullURL); err != nil {
			log.Printf("Warning: failed to create AGENT.md: %v", err)
		}
	}

	return nil
}

func (ad *AgentDownloader) copyDirectoryContents(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return createDirectoryWithPermissions(dstPath)
		}

		return ad.copyFile(path, dstPath)
	})
}

func (ad *AgentDownloader) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return createFileWithPermissions(dst, data)
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
func (ad *AgentDownloader) saveLockFile(agentName string, source string, sourceType string, sourceUrl string, skillFolderHash string, components int, detection string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	agentsDir := filepath.Join(home, ".agents")
	if err := createDirectoryWithPermissions(agentsDir); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	lockFilePath := filepath.Join(agentsDir, ".agent-lock.json")

	// Read existing lock file or create new one
	var lockFile ComponentLockFile
	lockData, err := os.ReadFile(lockFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			lockFile = ComponentLockFile{
				Version: 3,
				Agents:  make(map[string]ComponentLockEntry),
			}
		} else {
			return fmt.Errorf("failed to read lock file: %w", err)
		}
	} else {
		if err := json.Unmarshal(lockData, &lockFile); err != nil {
			lockFile = ComponentLockFile{
				Version: 3,
				Agents:  make(map[string]ComponentLockEntry),
			}
		}
		if lockFile.Version < 3 {
			lockFile.Version = 3
			if lockFile.Agents == nil {
				lockFile.Agents = make(map[string]ComponentLockEntry)
			}
		}
	}

	now := time.Now().Format(time.RFC3339)

	existingEntry, exists := lockFile.Agents[agentName]
	if !exists {
		existingEntry.InstalledAt = now
	}

	lockFile.Agents[agentName] = ComponentLockEntry{
		Source:          source,
		SourceType:      sourceType,
		SourceUrl:       sourceUrl,
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

func (ad *AgentDownloader) downloadAgentWithRepo(fullURL, agentName, repoURL string, repoPath string, components []DetectedComponent) error {
	// Filter for agent components from provided components
	var agentComponents []DetectedComponent
	for _, comp := range components {
		if comp.Type == ComponentAgent {
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

	// If only one agent component found, copy its contents
	if len(agentComponents) == 1 {
		component := agentComponents[0]
		componentPath := filepath.Join(repoPath, component.Path)

		// Copy component contents to agent directory
		err := ad.copyDirectoryContents(componentPath, agentDir)
		if err != nil {
			os.RemoveAll(agentDir)
			return fmt.Errorf("failed to copy agent contents: %w", err)
		}
	} else {
		// Multiple agents found, create a monorepo structure
		for _, component := range agentComponents {
			componentDir := filepath.Join(agentDir, component.Name)
			componentPath := filepath.Join(repoPath, component.Path)

			err := createDirectoryWithPermissions(componentDir)
			if err != nil {
				continue
			}

			err = ad.copyDirectoryContents(componentPath, componentDir)
			if err != nil {
				log.Printf("Warning: failed to copy agent %s: %v", component.Name, err)
			}
		}
	}

	var commitHash string
	var repo *git.Repository

	// Handle metadata differently for local vs remote repositories
	if ad.detector.detectProvider(repoURL) == "local" {
		// For local repositories, open repository directly
		var err error
		repo, err = git.PlainOpen(fullURL)
		if err != nil {
			// Non-fatal, continue without git metadata
			log.Printf("Warning: failed to open local repository for metadata: %v", err)
		}
	} else {
		// For remote repositories, clone to get git history for metadata
		var err error
		repo, err = git.PlainClone(agentDir+".git", true, &git.CloneOptions{
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
	if sourceType == "github" {
		if strings.HasPrefix(fullURL, "https://github.com/") {
			ownerRepo := strings.TrimPrefix(fullURL, "https://github.com/")
			ownerRepo = strings.TrimSuffix(ownerRepo, ".git")
			if hash, err := computeGitHubTreeSHA(ownerRepo, "AGENT.md"); err == nil {
				folderHash = hash
			}
		}
	} else {
		if hash, err := computeLocalFolderHash(agentDir); err == nil {
			folderHash = hash
		}
	}

	if err := ad.saveLockFile(agentName, fullURL, sourceType, fullURL, folderHash, len(agentComponents), "recursive"); err != nil {
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
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return createDirectoryWithPermissions(dstPath)
		}

		return cl.copyFile(path, dstPath)
	})
}

func (cl *ComponentLinker) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return createFileWithPermissions(dst, data)
}

func (cl *ComponentLinker) linkComponent(componentType, componentName string) error {
	srcDir := filepath.Join(cl.agentsDir, componentType, componentName)
	dstDir := filepath.Join(cl.opencodeDir, componentType, componentName)

	// Check if source component exists
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("component %s/%s does not exist in %s", componentType, componentName, cl.agentsDir)
	}

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

func (cl *ComponentLinker) linkAllComponents() error {
	componentTypes := []string{"skills", "agents", "commands"}

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
				if err := cl.linkComponent(componentType, entry.Name()); err != nil {
					fmt.Printf("Warning: failed to link %s/%s: %v\n", componentType, entry.Name(), err)
				}
			}
		}
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
	if !cl.detector.isLocalPath(cwd) {
		return fmt.Errorf("current directory is not a git repository")
	}

	// Detect components in the current repository
	components, err := cl.detector.detectComponentsInRepo(cwd)
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

func (ud *UpdateDetector) loadMetadata(componentType, componentName string) (*ComponentMetadata, error) {
	// First try to load from npx add-skill compatible lock files
	if metadata, err := ud.loadFromLockFile(componentType, componentName); err == nil {
		// Convert to legacy format for compatibility
		return &ComponentMetadata{
			Name:   componentName,
			Source: metadata.SourceUrl,
			Commit: metadata.SkillFolderHash,
		}, nil
	}

	// Fall back to legacy metadata files
	var metadataFile string
	switch componentType {
	case "skills":
		metadataFile = filepath.Join(ud.baseDir, "skills", componentName, ".skill-metadata.json")
	case "agents":
		metadataFile = filepath.Join(ud.baseDir, "agents", componentName, ".agent-metadata.json")
	case "commands":
		metadataFile = filepath.Join(ud.baseDir, "commands", componentName, ".command-metadata.json")
	default:
		return nil, fmt.Errorf("unknown component type: %s", componentType)
	}

	data, err := os.ReadFile(metadataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata ComponentMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

func (ud *UpdateDetector) loadFromLockFile(componentType, componentName string) (*ComponentLockEntry, error) {
	var lockFilePath string
	var entries map[string]ComponentLockEntry

	switch componentType {
	case "skills":
		lockFilePath = filepath.Join(ud.baseDir, ".skill-lock.json")
	case "agents":
		lockFilePath = filepath.Join(ud.baseDir, ".agent-lock.json")
	case "commands":
		lockFilePath = filepath.Join(ud.baseDir, ".command-lock.json")
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
	fullURL, err := ud.detector.normalizeURL(repoURL)
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
	componentTypes := []string{"skills", "agents", "commands"}

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
	fullURL, err := bd.detector.normalizeURL(repoURL)
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

	// Detect all components in the repository
	components, err := bd.detector.detectComponentsInRepo(tempDir)
	if err != nil {
		return fmt.Errorf("failed to detect components: %w", err)
	}

	if len(components) == 0 {
		return fmt.Errorf("no components (skills, agents, or commands) detected in repository")
	}

	// Group components by type
	skillComponents := []DetectedComponent{}
	agentComponents := []DetectedComponent{}
	commandComponents := []DetectedComponent{}

	for _, comp := range components {
		switch comp.Type {
		case ComponentSkill:
			skillComponents = append(skillComponents, comp)
		case ComponentAgent:
			agentComponents = append(agentComponents, comp)
		case ComponentCommand:
			commandComponents = append(commandComponents, comp)
		}
	}

	// Download skills using optimized method
	for _, comp := range skillComponents {
		fmt.Printf("Downloading skill: %s\n", comp.Name)
		if err := bd.skillDownloader.downloadSkill(fullURL, comp.Name, tempDir); err != nil {
			fmt.Printf("Warning: failed to download skill %s: %v\n", comp.Name, err)
		} else {
			fmt.Printf("Successfully downloaded skill: %s\n", comp.Name)
		}
	}

	// Download agents using optimized method
	for _, comp := range agentComponents {
		fmt.Printf("Downloading agent: %s\n", comp.Name)
		if err := bd.agentDownloader.downloadAgent(fullURL, comp.Name, tempDir); err != nil {
			fmt.Printf("Warning: failed to download agent %s: %v\n", comp.Name, err)
		} else {
			fmt.Printf("Successfully downloaded agent: %s\n", comp.Name)
		}
	}

	// Download commands using optimized method
	for _, comp := range commandComponents {
		fmt.Printf("Downloading command: %s\n", comp.Name)
		if err := bd.commandDownloader.downloadCommand(fullURL, comp.Name, tempDir); err != nil {
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
	detector   *RepositoryDetector
	skillDir   string
	agentDir   string
	commandDir string
}

func NewComponentExecutor() *ComponentExecutor {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get user home directory:", err)
	}

	return &ComponentExecutor{
		detector:   NewRepositoryDetector(),
		skillDir:   filepath.Join(home, ".agents", "skills"),
		agentDir:   filepath.Join(home, ".agents", "agents"),
		commandDir: filepath.Join(home, ".agents", "commands"),
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
	fullURL, err := ce.detector.normalizeURL(repoURL)
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
	components, err := ce.detector.detectComponentsInRepo(tempDir)
	if err != nil {
		return fmt.Errorf("failed to detect components: %w", err)
	}

	if len(components) == 0 {
		return fmt.Errorf("no components found in repository %s", repoURL)
	}

	// Find the main/root component or use the first one
	var mainComponent *DetectedComponent
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
	case ComponentSkill:
		return ce.runLocalComponent(componentPath, "skill", args)
	case ComponentAgent:
		return ce.runLocalComponent(componentPath, "agent", args)
	case ComponentCommand:
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
	)

	// Execute Cobra command
	cmd.Execute()
}
