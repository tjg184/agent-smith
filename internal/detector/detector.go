package detector

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tjg184/agent-smith/internal/fileutil"
	gitpkg "github.com/tjg184/agent-smith/internal/git"
	"github.com/tjg184/agent-smith/internal/models"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/paths"
)

// RepositoryDetector maintains repository patterns and component detection
type RepositoryDetector struct {
	patterns        map[string]string
	detectionConfig *models.DetectionConfig
	logger          *logger.Logger
}

// NewRepositoryDetector creates a new RepositoryDetector with default config
func NewRepositoryDetector() *RepositoryDetector {
	return NewRepositoryDetectorWithConfig("")
}

// NewRepositoryDetectorWithConfig creates a new RepositoryDetector with custom config path
func NewRepositoryDetectorWithConfig(configPath string) *RepositoryDetector {
	rd := &RepositoryDetector{
		patterns: map[string]string{
			"github":     `^https?://(?:www\.)?github\.com/[^/]+/[^/]+$`,
			"github_git": `^(git@|ssh://)git@github\.com:[^/]+/[^/]+\.git$`,
			"github_api": `^https?://api\.github\.com/repos/[^/]+/[^/]+$`,

			"gitlab":     `^https?://(?:www\.)?gitlab\.com/[^/]+/[^/]+$`,
			"gitlab_git": `^(git@|ssh://)git@gitlab\.com:[^/]+/[^/]+\.git$`,
			"gitlab_api": `^https?://gitlab\.com/api/v4/projects/[^/]+$`,

			"bitbucket":     `^https?://(?:www\.)?bitbucket\.org/[^/]+/[^/]+$`,
			"bitbucket_git": `^(git@|ssh://)git@bitbucket\.org:[^/]+/[^/]+\.git$`,
			"bitbucket_api": `^https?://api\.bitbucket\.org/2\.0/repositories/[^/]+/[^/]+$`,

			"git_http": `^https?://(?!.*(?:github\.com|gitlab\.com|bitbucket\.org)).+$`,
			"git_ssh":  `^(ssh://|git@).+$`,
			"git":      `^(https?://|git@|ssh://).+\.git$`,
		},
		logger: nil,
	}

	if configPath == "" {
		configPath = getDetectionConfigPath()
	}

	if err := rd.loadDetectionConfig(configPath); err != nil {
		rd.detectionConfig = createDefaultDetectionConfig()
	}

	return rd
}

// SetLogger sets the logger for this detector
func (rd *RepositoryDetector) SetLogger(l *logger.Logger) {
	rd.logger = l
}

// createDefaultDetectionConfig returns the default component detection patterns
func createDefaultDetectionConfig() *models.DetectionConfig {
	return &models.DetectionConfig{
		Components: map[string]models.ComponentDetectionPattern{
			string(models.ComponentSkill): {
				Name:        "skill",
				ExactFiles:  []string{paths.SkillMarkdownFile},
				IgnorePaths: paths.IgnoredPaths,
			},
			string(models.ComponentAgent): {
				Name:           "agent",
				PathPatterns:   []string{paths.AgentsPathPattern, paths.AgentsSubDir},
				FileExtensions: []string{".md"},
				IgnorePaths:    paths.IgnoredPaths,
			},
			string(models.ComponentCommand): {
				Name:           "command",
				PathPatterns:   []string{paths.CommandsPathPattern, paths.CommandsSubDir},
				FileExtensions: []string{".md"},
				IgnorePaths:    paths.IgnoredPaths,
			},
		},
	}
}

// loadDetectionConfig loads detection configuration from a JSON file
func (rd *RepositoryDetector) loadDetectionConfig(configPath string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		rd.detectionConfig = createDefaultDetectionConfig()
		return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read detection config file %s: %v", configPath, err)
	}

	var config models.DetectionConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse detection config file %s: %v", configPath, err)
	}

	rd.detectionConfig = &config
	return nil
}

// SaveDetectionConfig saves detection configuration to a JSON file
func (rd *RepositoryDetector) SaveDetectionConfig(configPath string) error {
	if rd.detectionConfig == nil {
		return fmt.Errorf("no detection config to save")
	}

	data, err := json.MarshalIndent(rd.detectionConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal detection config: %v", err)
	}

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, fileutil.GetCrossPlatformPermissions()); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	return fileutil.CreateFileWithPermissions(configPath, data)
}

// getDetectionConfigPath returns the default path for the detection configuration file
func getDetectionConfigPath() string {
	configPath, err := paths.GetDetectionConfigPath()
	if err != nil {
		return "./detection-config.json"
	}
	return configPath
}

// IsLocalPath checks if a path is a local filesystem path
func (rd *RepositoryDetector) IsLocalPath(path string) bool {
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

// DetectProvider detects the git provider from a repository URL
func (rd *RepositoryDetector) DetectProvider(repoURL string) string {
	repoURL = strings.TrimSpace(repoURL)

	// Check for local paths first (most specific)
	if rd.IsLocalPath(repoURL) {
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

// NormalizeURL normalizes a repository URL
// NormalizeURL normalizes a repository URL
func (rd *RepositoryDetector) NormalizeURL(repoURL string) (string, error) {
	repoURL = strings.TrimSpace(repoURL)

	// Handle local paths
	if rd.IsLocalPath(repoURL) {
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

	// Delegate to git package's NormalizeURL for remote repositories
	return gitpkg.NormalizeURL(repoURL)
}

// ValidateRepository validates a repository URL
func (rd *RepositoryDetector) ValidateRepository(repoURL string) error {
	provider := rd.DetectProvider(repoURL)

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
