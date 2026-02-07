package downloader

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/schollz/progressbar/v3"
	"github.com/tgaines/agent-smith/internal/detector"
	"github.com/tgaines/agent-smith/internal/formatter"
	gitpkg "github.com/tgaines/agent-smith/internal/git"
	"github.com/tgaines/agent-smith/internal/models"
)

// BulkDownloader handles bulk downloading of all components from a repository
type BulkDownloader struct {
	skillDownloader   *SkillDownloader
	agentDownloader   *AgentDownloader
	commandDownloader *CommandDownloader
	detector          *detector.RepositoryDetector
	formatter         *formatter.Formatter
}

// NewBulkDownloader creates a new BulkDownloader instance
func NewBulkDownloader() *BulkDownloader {
	return &BulkDownloader{
		skillDownloader:   NewSkillDownloader(),
		agentDownloader:   NewAgentDownloader(),
		commandDownloader: NewCommandDownloader(),
		detector:          detector.NewRepositoryDetector(),
		formatter:         formatter.New(),
	}
}

// NewBulkDownloaderWithTargetDir creates a new BulkDownloader instance that installs to a custom target directory
func NewBulkDownloaderWithTargetDir(targetDir string) *BulkDownloader {
	return &BulkDownloader{
		skillDownloader:   NewSkillDownloaderWithTargetDir(targetDir),
		agentDownloader:   NewAgentDownloaderWithTargetDir(targetDir),
		commandDownloader: NewCommandDownloaderWithTargetDir(targetDir),
		detector:          detector.NewRepositoryDetector(),
		formatter:         formatter.New(),
	}
}

// NewBulkDownloaderForProfile creates a new BulkDownloader instance that installs to a profile
func NewBulkDownloaderForProfile(profileName string) *BulkDownloader {
	return &BulkDownloader{
		skillDownloader:   NewSkillDownloaderForProfile(profileName),
		agentDownloader:   NewAgentDownloaderForProfile(profileName),
		commandDownloader: NewCommandDownloaderForProfile(profileName),
		detector:          detector.NewRepositoryDetector(),
		formatter:         formatter.New(),
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

	// Create temporary directory for repository detection
	tempDir, err = os.MkdirTemp("", "agent-smith-bulk-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

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
		os.RemoveAll(tempDir)
		return "", nil, fmt.Errorf("failed to clone repository for bulk detection: %w", err)
	}

	// Detect all components in the repository from root
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

	// Display installation header
	bd.formatter.SectionHeader(fmt.Sprintf("Installing components from %s", repoURL))

	// Clean up temp directory when done
	defer os.RemoveAll(tempDir)

	return bd.processComponents(components, fullURL, repoURL, tempDir)
}

// AddAll downloads all components from a repository
func (bd *BulkDownloader) AddAll(repoURL string) error {
	// Validate repository and get components
	tempDir, components, err := bd.ValidateRepo(repoURL)
	if err != nil {
		return err
	}

	// Install from the temp directory (ValidateRepo already created it)
	return bd.AddAllFromTemp(repoURL, components, tempDir)
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

	totalComponents := len(components)
	fmt.Printf("\nInstalling %d components...\n", totalComponents)

	// Create progress bar
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

	// Download skills using optimized method with shared repository
	for _, comp := range skillComponents {
		result := formatter.InstallResult{
			Name:    comp.Name,
			Type:    "skill",
			Success: true,
		}
		if err := bd.skillDownloader.DownloadSkillWithRepo(fullURL, comp.Name, repoURL, tempDir, components); err != nil {
			result.Success = false
			result.Error = err.Error()
		}
		results = append(results, result)
		bar.Add(1)
	}

	// Download agents using optimized method with shared repository
	for _, comp := range agentComponents {
		result := formatter.InstallResult{
			Name:    comp.Name,
			Type:    "agent",
			Success: true,
		}
		if err := bd.agentDownloader.DownloadAgentWithRepo(fullURL, comp.Name, repoURL, tempDir, components); err != nil {
			result.Success = false
			result.Error = err.Error()
		}
		results = append(results, result)
		bar.Add(1)
	}

	// Download commands using optimized method with shared repository
	for _, comp := range commandComponents {
		result := formatter.InstallResult{
			Name:    comp.Name,
			Type:    "command",
			Success: true,
		}
		if err := bd.commandDownloader.DownloadCommandWithRepo(fullURL, comp.Name, repoURL, tempDir, components); err != nil {
			result.Success = false
			result.Error = err.Error()
		}
		results = append(results, result)
		bar.Add(1)
	}

	// Finish the progress bar
	bar.Finish()

	// Display summary table
	bd.formatter.DisplaySummaryTable(results, len(skillComponents), len(agentComponents), len(commandComponents))

	return nil
}
