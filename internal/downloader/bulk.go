package downloader

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/schollz/progressbar/v3"
	"github.com/tgaines/agent-smith/internal/detector"
	"github.com/tgaines/agent-smith/internal/models"
)

// BulkDownloader handles bulk downloading of all components from a repository
type BulkDownloader struct {
	skillDownloader   *SkillDownloader
	agentDownloader   *AgentDownloader
	commandDownloader *CommandDownloader
	detector          *detector.RepositoryDetector
}

// NewBulkDownloader creates a new BulkDownloader instance
func NewBulkDownloader() *BulkDownloader {
	return &BulkDownloader{
		skillDownloader:   NewSkillDownloader(),
		agentDownloader:   NewAgentDownloader(),
		commandDownloader: NewCommandDownloader(),
		detector:          detector.NewRepositoryDetector(),
	}
}

// AddAll downloads all components from a repository
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

// InstallResult represents the result of a single component installation
type InstallResult struct {
	Name    string
	Type    string
	Success bool
	Error   string
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

	var results []InstallResult

	// Download skills using optimized method with shared repository
	for _, comp := range skillComponents {
		result := InstallResult{
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
		result := InstallResult{
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
		result := InstallResult{
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
	bd.displaySummaryTable(results, len(skillComponents), len(agentComponents), len(commandComponents))

	return nil
}

// displaySummaryTable shows a formatted table of installation results
func (bd *BulkDownloader) displaySummaryTable(results []InstallResult, skillCount, agentCount, commandCount int) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("Installation Summary")
	fmt.Println(strings.Repeat("=", 80))

	// Group results by type
	skillResults := []InstallResult{}
	agentResults := []InstallResult{}
	commandResults := []InstallResult{}

	for _, result := range results {
		switch result.Type {
		case "skill":
			skillResults = append(skillResults, result)
		case "agent":
			agentResults = append(agentResults, result)
		case "command":
			commandResults = append(commandResults, result)
		}
	}

	// Display each type section
	if len(skillResults) > 0 {
		bd.displayTypeSection("Skills", skillResults)
	}
	if len(agentResults) > 0 {
		bd.displayTypeSection("Agents", agentResults)
	}
	if len(commandResults) > 0 {
		bd.displayTypeSection("Commands", commandResults)
	}

	// Calculate summary statistics
	successCount := 0
	failureCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			failureCount++
		}
	}

	// Display summary
	fmt.Println("\n" + strings.Repeat("-", 80))
	fmt.Printf("Successfully installed: %d/%d components\n", successCount, len(results))
	if failureCount > 0 {
		fmt.Printf("Failed: %d components\n", failureCount)
	}
	fmt.Println(strings.Repeat("=", 80))
}

// displayTypeSection displays a section of results for a specific component type
func (bd *BulkDownloader) displayTypeSection(typeName string, results []InstallResult) {
	fmt.Printf("\n%s:\n", typeName)
	fmt.Println(strings.Repeat("-", 80))

	for _, result := range results {
		status := "✓"
		statusText := "Success"
		if !result.Success {
			status = "✗"
			statusText = "Failed"
		}

		fmt.Printf("  %s  %-40s  %s\n", status, result.Name, statusText)
		if !result.Success && result.Error != "" {
			fmt.Printf("      Error: %s\n", result.Error)
		}
	}
}
