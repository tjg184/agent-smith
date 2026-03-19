package downloader

import (
	"fmt"
	"os"

	"github.com/schollz/progressbar/v3"
	"github.com/tjg184/agent-smith/internal/detector"
	"github.com/tjg184/agent-smith/internal/formatter"
	gitpkg "github.com/tjg184/agent-smith/internal/git"
	"github.com/tjg184/agent-smith/internal/models"
)

// BulkDownloader handles bulk downloading of all components from a repository
type BulkDownloader struct {
	downloaders map[models.ComponentType]Downloader
	detector    *detector.RepositoryDetector
	formatter   *formatter.Formatter
}

// NewBulkDownloader creates a new BulkDownloader instance
func NewBulkDownloader() *BulkDownloader {
	return &BulkDownloader{
		downloaders: map[models.ComponentType]Downloader{
			models.ComponentSkill:   ForType(models.ComponentSkill),
			models.ComponentAgent:   ForType(models.ComponentAgent),
			models.ComponentCommand: ForType(models.ComponentCommand),
		},
		detector:  detector.NewRepositoryDetector(),
		formatter: formatter.New(),
	}
}

// NewBulkDownloaderWithTargetDir creates a new BulkDownloader instance that installs to a custom target directory
func NewBulkDownloaderWithTargetDir(targetDir string) *BulkDownloader {
	return &BulkDownloader{
		downloaders: map[models.ComponentType]Downloader{
			models.ComponentSkill:   ForTypeWithTargetDir(models.ComponentSkill, targetDir),
			models.ComponentAgent:   ForTypeWithTargetDir(models.ComponentAgent, targetDir),
			models.ComponentCommand: ForTypeWithTargetDir(models.ComponentCommand, targetDir),
		},
		detector:  detector.NewRepositoryDetector(),
		formatter: formatter.New(),
	}
}

// NewBulkDownloaderForProfile creates a new BulkDownloader instance that installs to a profile
func NewBulkDownloaderForProfile(profileName string) *BulkDownloader {
	return &BulkDownloader{
		downloaders: map[models.ComponentType]Downloader{
			models.ComponentSkill:   ForTypeWithProfile(models.ComponentSkill, profileName),
			models.ComponentAgent:   ForTypeWithProfile(models.ComponentAgent, profileName),
			models.ComponentCommand: ForTypeWithProfile(models.ComponentCommand, profileName),
		},
		detector:  detector.NewRepositoryDetector(),
		formatter: formatter.New(),
	}
}

// ValidateRepo clones a repository and detects components without installing.
// Returns the temp directory path (caller must clean up), detected components, and any error.
// This allows validation before creating profiles or other state.
func (bd *BulkDownloader) ValidateRepo(repoURL string) (tempDir string, components []models.DetectedComponent, err error) {
	fullURL, err := bd.detector.NormalizeURL(repoURL)
	if err != nil {
		return "", nil, fmt.Errorf("failed to normalize repository URL: %w", err)
	}

	tempDir, err = os.MkdirTemp("", "agent-smith-bulk-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	if _, err = gitpkg.CloneShallow(gitpkg.NewDefaultCloner(), tempDir, fullURL); err != nil {
		os.RemoveAll(tempDir)
		return "", nil, fmt.Errorf("failed to clone repository for bulk detection: %w", err)
	}

	components, err = bd.detector.DetectComponentsInRepo(tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", nil, fmt.Errorf("failed to detect components: %w", err)
	}

	if len(components) == 0 {
		os.RemoveAll(tempDir)
		return "", nil, fmt.Errorf("no components (skills, agents, or commands) detected in repository")
	}

	return tempDir, components, nil
}

// AddAllFromTemp installs components from a pre-cloned repository.
// The tempDir should contain a cloned repository with the components to install.
// This is used after ValidateRepo to avoid double-cloning.
func (bd *BulkDownloader) AddAllFromTemp(repoURL string, components []models.DetectedComponent, tempDir string) error {
	fullURL, err := bd.detector.NormalizeURL(repoURL)
	if err != nil {
		return fmt.Errorf("failed to normalize repository URL: %w", err)
	}

	bd.formatter.SectionHeader(fmt.Sprintf("Installing components from %s", repoURL))

	defer os.RemoveAll(tempDir)

	return bd.processComponents(components, fullURL, repoURL, tempDir)
}

// AddAll downloads all components from a repository
func (bd *BulkDownloader) AddAll(repoURL string) error {
	tempDir, components, err := bd.ValidateRepo(repoURL)
	if err != nil {
		return err
	}

	return bd.AddAllFromTemp(repoURL, components, tempDir)
}

// processComponents handles downloading components from the repository
func (bd *BulkDownloader) processComponents(components []models.DetectedComponent, fullURL, repoURL, tempDir string) error {
	totalComponents := len(components)
	fmt.Printf("\nInstalling %d components...\n", totalComponents)

	bar := progressbar.NewOptions(totalComponents,
		progressbar.OptionSetDescription("Progress"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	var results []formatter.InstallResult
	typeCounts := map[models.ComponentType]int{}

	for _, comp := range components {
		dl, ok := bd.downloaders[comp.Type]
		if !ok {
			continue
		}
		result := formatter.InstallResult{
			Name:    comp.Name,
			Type:    string(comp.Type),
			Success: true,
		}
		if err := dl.DownloadWithRepo(fullURL, comp.Name, repoURL, tempDir, components); err != nil {
			result.Success = false
			result.Error = err.Error()
		}
		results = append(results, result)
		typeCounts[comp.Type]++
		bar.Add(1)
	}

	bar.Finish()

	bd.formatter.DisplaySummaryTable(results,
		typeCounts[models.ComponentSkill],
		typeCounts[models.ComponentAgent],
		typeCounts[models.ComponentCommand],
	)
	return nil
}
