package updater

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/tgaines/agent-smith/internal/detector"
	"github.com/tgaines/agent-smith/internal/downloader"
	"github.com/tgaines/agent-smith/internal/formatter"
	gitpkg "github.com/tgaines/agent-smith/internal/git"
	metadataPkg "github.com/tgaines/agent-smith/internal/metadata"
	"github.com/tgaines/agent-smith/internal/models"
	"github.com/tgaines/agent-smith/pkg/colors"
	"github.com/tgaines/agent-smith/pkg/paths"
	"github.com/tgaines/agent-smith/pkg/profiles"
	"github.com/tgaines/agent-smith/pkg/styles"
)

// UpdateDetector provides functionality to detect and apply updates to components
type UpdateDetector struct {
	baseDir     string
	detector    *detector.RepositoryDetector
	profileName string // If non-empty, we're working with a profile
}

// NewUpdateDetector creates a new UpdateDetector instance
// If an active profile exists, it will use that profile's directory
// Otherwise, it will use the default ~/.agent-smith/ directory
func NewUpdateDetector() *UpdateDetector {
	baseDir, err := paths.GetAgentsDir()
	if err != nil {
		panic(fmt.Sprintf("Failed to get agents directory: %v", err))
	}

	var profileName string

	// Check if a profile is active
	pm, err := profiles.NewProfileManager(nil)
	if err == nil {
		activeProfile, err := pm.GetActiveProfile()
		if err == nil && activeProfile != "" {
			// Use the active profile directory instead
			profilesDir, err := paths.GetProfilesDir()
			if err == nil {
				baseDir = filepath.Join(profilesDir, activeProfile)
				profileName = activeProfile
				fmt.Printf("Using active profile for updates: %s\n", activeProfile)
			}
		}
	}

	return &UpdateDetector{
		baseDir:     baseDir,
		detector:    detector.NewRepositoryDetector(),
		profileName: profileName,
	}
}

// NewUpdateDetectorWithProfile creates a new UpdateDetector instance for a specific profile
// If profile is empty, it falls back to the active profile or base directory
func NewUpdateDetectorWithProfile(profile string) *UpdateDetector {
	baseDir, err := paths.GetAgentsDir()
	if err != nil {
		panic(fmt.Sprintf("Failed to get agents directory: %v", err))
	}

	var profileName string

	if profile != "" {
		// Use explicit profile (bypasses active profile logic)
		profilesDir, err := paths.GetProfilesDir()
		if err != nil {
			panic(fmt.Sprintf("Failed to get profiles directory: %v", err))
		}
		baseDir = filepath.Join(profilesDir, profile)
		profileName = profile
		fmt.Printf("Using specified profile for updates: %s\n", profile)
	} else {
		// Check if a profile is active
		pm, err := profiles.NewProfileManager(nil)
		if err == nil {
			activeProfile, err := pm.GetActiveProfile()
			if err == nil && activeProfile != "" {
				// Use the active profile directory instead
				profilesDir, err := paths.GetProfilesDir()
				if err == nil {
					baseDir = filepath.Join(profilesDir, activeProfile)
					profileName = activeProfile
					fmt.Printf("Using active profile for updates: %s\n", activeProfile)
				}
			}
		}
	}

	return &UpdateDetector{
		baseDir:     baseDir,
		detector:    detector.NewRepositoryDetector(),
		profileName: profileName,
	}
}

// NewUpdateDetectorWithBaseDir creates a new UpdateDetector instance with an explicit base directory
// This allows the caller to specify exactly which directory to use, bypassing all profile detection logic
// The profileName is left empty since the caller is managing the directory directly
func NewUpdateDetectorWithBaseDir(baseDir string) *UpdateDetector {
	return &UpdateDetector{
		baseDir:     baseDir,
		detector:    detector.NewRepositoryDetector(),
		profileName: "", // No profile name since we're using an explicit directory
	}
}

// LoadMetadata loads component metadata from lock files only
func (ud *UpdateDetector) LoadMetadata(componentType, componentName string) (*models.ComponentMetadata, error) {
	return ud.loadMetadata(componentType, componentName)
}

// loadMetadata loads component metadata from lock files only
func (ud *UpdateDetector) loadMetadata(componentType, componentName string) (*models.ComponentMetadata, error) {
	// Load from lock files
	entry, err := metadataPkg.LoadLockFileEntry(ud.baseDir, componentType, componentName)
	if err != nil {
		return nil, fmt.Errorf("failed to load lock file entry: %w", err)
	}

	// Convert to legacy format for compatibility
	return &models.ComponentMetadata{
		Name:   componentName,
		Source: entry.SourceUrl,
		Commit: entry.CommitHash,
	}, nil
}

// loadFromLockFile loads component metadata from lock files
func (ud *UpdateDetector) loadFromLockFile(componentType, componentName string) (*models.ComponentLockEntry, error) {
	return metadataPkg.LoadLockFileEntry(ud.baseDir, componentType, componentName)
}

// GetCurrentRepoSHA fetches the current HEAD commit SHA from a repository
// This method is public to allow other packages (like materialization) to check for updates
func (ud *UpdateDetector) GetCurrentRepoSHA(repoURL string) (string, error) {
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

	repo, err := git.PlainClone(tempDir, true, cloneOpts)
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

// HasUpdates checks if a component has updates available
func (ud *UpdateDetector) HasUpdates(componentType, componentName, repoURL string) (bool, error) {
	// Load existing metadata
	metadata, err := ud.loadMetadata(componentType, componentName)
	if err != nil {
		return false, fmt.Errorf("failed to load metadata: %w", err)
	}

	// If no commit hash is stored (old lock file format), assume update is needed
	if metadata.Commit == "" {
		fmt.Printf("Warning: %s/%s has no commit hash stored, will re-download to update lock file\n", componentType, componentName)
		return true, nil
	}

	// Get current repository SHA
	currentSHA, err := ud.GetCurrentRepoSHA(repoURL)
	if err != nil {
		return false, fmt.Errorf("failed to get current repository SHA: %w", err)
	}

	// Compare stored SHA with current SHA
	return metadata.Commit != currentSHA, nil
}

// UpdateComponent updates a single component if updates are detected
func (ud *UpdateDetector) UpdateComponent(componentType, componentName, repoURL string) error {
	// Show location header
	if ud.profileName != "" {
		fmt.Printf("%s\n\n", styles.InfoArrowFormat(fmt.Sprintf("Updating components in: %s", ud.baseDir)))
	} else {
		fmt.Printf("%s\n\n", styles.InfoArrowFormat("Checking for updates..."))
	}

	// Check for updates
	fmt.Print(styles.ProgressCheckingFormat(componentType, componentName))

	hasUpdates, err := ud.HasUpdates(componentType, componentName, repoURL)
	if err != nil {
		fmt.Printf("%s\n", styles.StatusFailedFormat())
		fmt.Printf("%s\n\n", styles.IndentedErrorFormat(err.Error()))
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !hasUpdates {
		fmt.Printf("%s\n\n", styles.StatusUpToDateFormat())
		return nil
	}

	// Component needs updating
	fmt.Printf("%s\n", styles.StatusUpdatingFormat())

	// Remove old component directory to ensure clean re-clone
	componentDir := filepath.Join(ud.baseDir, componentType, componentName)
	if _, err := os.Stat(componentDir); err == nil {
		if err := os.RemoveAll(componentDir); err != nil {
			fmt.Printf("  %s\n\n", styles.IndentedErrorFormat(fmt.Sprintf("Failed to remove old version: %v", err)))
			return fmt.Errorf("failed to remove old component directory: %w", err)
		}
	}

	// Re-download the component with the latest changes
	var downloadErr error
	if ud.profileName != "" {
		// Use profile-aware downloaders
		switch componentType {
		case "skills":
			dl := downloader.NewSkillDownloaderForProfile(ud.profileName)
			downloadErr = dl.DownloadSkill(repoURL, componentName)
		case "agents":
			dl := downloader.NewAgentDownloaderForProfile(ud.profileName)
			downloadErr = dl.DownloadAgent(repoURL, componentName)
		case "commands":
			dl := downloader.NewCommandDownloaderForProfile(ud.profileName)
			downloadErr = dl.DownloadCommand(repoURL, componentName)
		default:
			return fmt.Errorf("unknown component type: %s", componentType)
		}
	} else {
		// Use standard downloaders
		switch componentType {
		case "skills":
			dl := downloader.NewSkillDownloader()
			downloadErr = dl.DownloadSkill(repoURL, componentName)
		case "agents":
			dl := downloader.NewAgentDownloader()
			downloadErr = dl.DownloadAgent(repoURL, componentName)
		case "commands":
			dl := downloader.NewCommandDownloader()
			downloadErr = dl.DownloadCommand(repoURL, componentName)
		default:
			return fmt.Errorf("unknown component type: %s", componentType)
		}
	}

	if downloadErr != nil {
		fmt.Printf("  %s\n\n", styles.IndentedErrorFormat(fmt.Sprintf("Update failed: %v", downloadErr)))
		return downloadErr
	}

	fmt.Printf("  %s\n\n", styles.StatusUpdatedSuccessfullyFormat())
	return nil
}

// componentUpdateInfo holds information about a component that needs checking/updating
type componentUpdateInfo struct {
	Type     string
	Name     string
	Metadata *models.ComponentMetadata
}

// UpdateAll iterates through all installed components and updates them
// Optimized to batch components by repository to reduce git clone operations
func (ud *UpdateDetector) UpdateAll() error {
	// Show location header
	if ud.profileName != "" {
		fmt.Printf("%s\n\n", styles.InfoArrowFormat(fmt.Sprintf("Updating components in: %s", ud.baseDir)))
	} else {
		fmt.Printf("%s\n\n", styles.InfoArrowFormat("Checking all components for updates..."))
	}

	// Step 1: Scan all components and group by repository
	componentsByRepo, totalComponents, err := ud.groupComponentsByRepository()
	if err != nil {
		return err
	}

	if totalComponents == 0 {
		fmt.Println("No components found to update.")
		return nil
	}

	// Track update statistics
	var totalChecked, upToDate, updated, failed int

	// Step 2: Process each repository batch
	for repoURL, components := range componentsByRepo {
		// Create temporary directory for this repository
		tempDir, err := os.MkdirTemp("", "agent-smith-update-batch-*")
		if err != nil {
			// If we can't create temp dir, mark all components as failed
			for _, comp := range components {
				totalChecked++
				fmt.Print(styles.ComponentProgressFormat(totalChecked, totalComponents, comp.Type, comp.Name))
				fmt.Printf("%s\n", styles.StatusFailedFormat())
				fmt.Printf("%s\n", styles.IndentedErrorFormat(fmt.Sprintf("Failed to create temp directory: %v", err)))
				failed++
			}
			continue
		}

		// Clone repository once for this batch
		fullURL, normalizeErr := ud.detector.NormalizeURL(repoURL)
		if normalizeErr != nil {
			os.RemoveAll(tempDir)
			// Mark all components from this repo as failed
			for _, comp := range components {
				totalChecked++
				fmt.Print(styles.ComponentProgressFormat(totalChecked, totalComponents, comp.Type, comp.Name))
				fmt.Printf("%s\n", styles.StatusFailedFormat())
				fmt.Printf("%s\n", styles.IndentedErrorFormat(fmt.Sprintf("Failed to normalize URL: %v", normalizeErr)))
				failed++
			}
			continue
		}

		// Clone repository (shallow clone)
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

		repo, cloneErr := git.PlainClone(tempDir, false, cloneOpts)
		if cloneErr != nil {
			os.RemoveAll(tempDir)
			// Mark all components from this repo as failed
			for _, comp := range components {
				totalChecked++
				fmt.Print(styles.ComponentProgressFormat(totalChecked, totalComponents, comp.Type, comp.Name))
				fmt.Printf("%s\n", styles.StatusFailedFormat())
				fmt.Printf("%s\n", styles.IndentedErrorFormat(fmt.Sprintf("Failed to clone repository: %v", cloneErr)))
				failed++
			}
			continue
		}

		// Get current HEAD commit hash once for this repository
		ref, err := repo.Head()
		var currentSHA string
		if err != nil {
			os.RemoveAll(tempDir)
			// Mark all components from this repo as failed
			for _, comp := range components {
				totalChecked++
				fmt.Print(styles.ComponentProgressFormat(totalChecked, totalComponents, comp.Type, comp.Name))
				fmt.Printf("%s\n", styles.StatusFailedFormat())
				fmt.Printf("%s\n", styles.IndentedErrorFormat(fmt.Sprintf("Failed to get HEAD reference: %v", err)))
				failed++
			}
			continue
		}
		currentSHA = ref.Hash().String()

		// Detect all components in this repository for *WithRepo methods
		allDetectedComponents, detectErr := ud.detector.DetectComponentsInRepo(tempDir)
		if detectErr != nil {
			// If detection fails, we can still update components individually
			// but we won't have the full component list for *WithRepo methods
			allDetectedComponents = []models.DetectedComponent{}
		}

		// Step 3: Check and update each component from this repository
		for _, comp := range components {
			totalChecked++
			fmt.Print(styles.ComponentProgressFormat(totalChecked, totalComponents, comp.Type, comp.Name))

			// Check if update is needed by comparing commit hashes
			if comp.Metadata.Commit == "" {
				// No commit hash stored, assume update needed
				fmt.Printf("%s\n", styles.StatusUpdatingFormat())
			} else if comp.Metadata.Commit == currentSHA {
				// Already up to date
				fmt.Printf("%s\n", styles.StatusUpToDateFormat())
				upToDate++
				continue
			} else {
				// Update needed
				fmt.Printf("%s\n", styles.StatusUpdatingFormat())
			}

			// Remove old component directory
			componentDir := filepath.Join(ud.baseDir, comp.Type, comp.Name)
			if _, err := os.Stat(componentDir); err == nil {
				if err := os.RemoveAll(componentDir); err != nil {
					fmt.Printf("%s\n", styles.IndentedErrorFormat(fmt.Sprintf("Failed to remove old version: %v", err)))
					failed++
					continue
				}
			}

			// Download using *WithRepo methods to reuse the cloned repository
			var downloadErr error
			if ud.profileName != "" {
				// Use profile-aware downloaders
				downloadErr = ud.downloadComponentWithRepoForProfile(comp.Type, comp.Name, fullURL, repoURL, tempDir, allDetectedComponents)
			} else {
				// Use standard downloaders
				downloadErr = ud.downloadComponentWithRepo(comp.Type, comp.Name, fullURL, repoURL, tempDir, allDetectedComponents)
			}

			if downloadErr != nil {
				fmt.Printf("%s\n", styles.IndentedErrorFormat(downloadErr.Error()))
				failed++
			} else {
				fmt.Printf("  %s\n", styles.StatusUpdatedSuccessfullyFormat())
				updated++
			}
		}

		// Clean up temporary directory for this repository
		os.RemoveAll(tempDir)
	}

	// Print summary with box drawing using styles package
	fmt.Println()
	table := styles.SummaryTableFormat("Update Summary", 80)
	table.AddRow("Total components checked:", totalChecked)
	table.AddRowWithSymbol(colors.Success(formatter.SymbolSuccess), "Already up to date:", upToDate)
	table.AddRowWithSymbol(colors.Success(formatter.SymbolSuccess), "Successfully updated:", updated)
	if failed > 0 {
		table.AddRowWithSymbol(colors.Error(formatter.SymbolError), "Failed:", failed)
	}
	fmt.Println(table.Build())
	fmt.Println()

	return nil
}

// groupComponentsByRepository scans all installed components and groups them by source repository
// Returns a map of repository URL -> components, total component count, and any error
func (ud *UpdateDetector) groupComponentsByRepository() (map[string][]componentUpdateInfo, int, error) {
	componentsByRepo := make(map[string][]componentUpdateInfo)
	totalComponents := 0

	componentTypes := paths.GetComponentTypes()

	for _, componentType := range componentTypes {
		typeDir := filepath.Join(ud.baseDir, componentType)
		if _, err := os.Stat(typeDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(typeDir)
		if err != nil {
			fmt.Printf("%s\n", styles.IndentedErrorFormat(fmt.Sprintf("Failed to read %s directory: %v", componentType, err)))
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				componentName := entry.Name()
				totalComponents++

				// Load metadata to get source URL
				metadata, err := ud.loadMetadata(componentType, componentName)
				if err != nil {
					// Skip components with missing metadata, they'll be handled as errors during update
					continue
				}

				// Group by source repository URL
				repoURL := metadata.Source
				componentsByRepo[repoURL] = append(componentsByRepo[repoURL], componentUpdateInfo{
					Type:     componentType,
					Name:     componentName,
					Metadata: metadata,
				})
			}
		}
	}

	return componentsByRepo, totalComponents, nil
}

// downloadComponentWithRepo downloads a component using the *WithRepo methods to reuse a cloned repository
func (ud *UpdateDetector) downloadComponentWithRepo(componentType, componentName, fullURL, repoURL, tempDir string, detectedComponents []models.DetectedComponent) error {
	switch componentType {
	case "skills":
		dl := downloader.NewSkillDownloader()
		return dl.DownloadSkillWithRepo(fullURL, componentName, repoURL, tempDir, detectedComponents)
	case "agents":
		dl := downloader.NewAgentDownloader()
		return dl.DownloadAgentWithRepo(fullURL, componentName, repoURL, tempDir, detectedComponents)
	case "commands":
		dl := downloader.NewCommandDownloader()
		return dl.DownloadCommandWithRepo(fullURL, componentName, repoURL, tempDir, detectedComponents)
	default:
		return fmt.Errorf("unknown component type: %s", componentType)
	}
}

// downloadComponentWithRepoForProfile downloads a component using profile-aware downloaders with *WithRepo methods
func (ud *UpdateDetector) downloadComponentWithRepoForProfile(componentType, componentName, fullURL, repoURL, tempDir string, detectedComponents []models.DetectedComponent) error {
	switch componentType {
	case "skills":
		dl := downloader.NewSkillDownloaderForProfile(ud.profileName)
		return dl.DownloadSkillWithRepo(fullURL, componentName, repoURL, tempDir, detectedComponents)
	case "agents":
		dl := downloader.NewAgentDownloaderForProfile(ud.profileName)
		return dl.DownloadAgentWithRepo(fullURL, componentName, repoURL, tempDir, detectedComponents)
	case "commands":
		dl := downloader.NewCommandDownloaderForProfile(ud.profileName)
		return dl.DownloadCommandWithRepo(fullURL, componentName, repoURL, tempDir, detectedComponents)
	default:
		return fmt.Errorf("unknown component type: %s", componentType)
	}
}
