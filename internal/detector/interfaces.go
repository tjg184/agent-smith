package detector

import (
	"github.com/tjg184/agent-smith/internal/models"
)

// Detector defines the interface for repository and component detection
type Detector interface {
	// Repository detection methods
	IsLocalPath(path string) bool
	DetectProvider(repoURL string) string
	NormalizeURL(repoURL string) (string, error)
	ValidateRepository(repoURL string) error

	// Pattern matching methods
	ShouldIgnorePath(relPath string, ignorePaths []string) bool
	MatchesExactFile(fileName string, exactFiles []string) bool
	MatchesPathPattern(relPath string, pathPatterns []string) bool
	MatchesFileExtension(fileName string, fileExtensions []string) bool

	// Component detection methods
	DetectComponentForPattern(fileName, relPath, fullRelPath, repoPath string, pattern models.ComponentDetectionPattern, componentType models.ComponentType) (string, string, bool)
	DetectComponentsInRepo(repoPath string) ([]models.DetectedComponent, error)

	// Configuration methods
	SaveDetectionConfig(configPath string) error
}

// Ensure RepositoryDetector implements Detector interface
var _ Detector = (*RepositoryDetector)(nil)
