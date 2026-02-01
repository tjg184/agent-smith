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
	metadataPkg "github.com/tgaines/agent-smith/internal/metadata"
	"github.com/tgaines/agent-smith/internal/models"
	"github.com/tgaines/agent-smith/pkg/colors"
	"github.com/tgaines/agent-smith/pkg/paths"
	"github.com/tgaines/agent-smith/pkg/profiles"
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

// getCurrentRepoSHA fetches the current HEAD commit SHA from a repository
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
	currentSHA, err := ud.getCurrentRepoSHA(repoURL)
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
		fmt.Printf("%s Updating components in: %s\n\n", colors.Info("→"), ud.baseDir)
	} else {
		fmt.Printf("%s Checking for updates...\n\n", colors.Info("→"))
	}

	// Check for updates
	fmt.Printf("Checking %s/%s... ", colors.Muted(componentType), componentName)

	hasUpdates, err := ud.HasUpdates(componentType, componentName, repoURL)
	if err != nil {
		fmt.Printf("%s Failed\n", colors.Error(formatter.SymbolError))
		fmt.Printf("  %s %v\n\n", colors.Muted("└─"), err)
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !hasUpdates {
		fmt.Printf("%s Up to date\n\n", colors.Success(formatter.SymbolSuccess))
		return nil
	}

	// Component needs updating
	fmt.Printf("%s Updating\n", colors.Warning(formatter.SymbolUpdating))

	// Remove old component directory to ensure clean re-clone
	componentDir := filepath.Join(ud.baseDir, componentType, componentName)
	if _, err := os.Stat(componentDir); err == nil {
		if err := os.RemoveAll(componentDir); err != nil {
			fmt.Printf("  %s Failed to remove old version: %v\n\n", colors.Error(formatter.SymbolError), err)
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
		fmt.Printf("  %s Update failed: %v\n\n", colors.Error(formatter.SymbolError), downloadErr)
		return downloadErr
	}

	fmt.Printf("  %s Updated successfully\n\n", colors.Success(formatter.SymbolSuccess))
	return nil
}

// UpdateAll iterates through all installed components and updates them
func (ud *UpdateDetector) UpdateAll() error {
	// Show location header
	if ud.profileName != "" {
		fmt.Printf("%s Updating components in: %s\n\n", colors.Info("→"), ud.baseDir)
	} else {
		fmt.Printf("%s Checking all components for updates...\n\n", colors.Info("→"))
	}

	componentTypes := paths.GetComponentTypes()

	// Track update statistics
	var totalChecked, upToDate, updated, failed int

	// First, count total components to check
	var totalComponents int
	for _, componentType := range componentTypes {
		typeDir := filepath.Join(ud.baseDir, componentType)
		if _, err := os.Stat(typeDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(typeDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				totalComponents++
			}
		}
	}

	if totalComponents == 0 {
		fmt.Println("No components found to update.")
		return nil
	}

	for _, componentType := range componentTypes {
		typeDir := filepath.Join(ud.baseDir, componentType)
		if _, err := os.Stat(typeDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(typeDir)
		if err != nil {
			fmt.Printf("%s Failed to read %s directory: %v\n", colors.Error(formatter.SymbolError), componentType, err)
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				componentName := entry.Name()
				totalChecked++

				fmt.Printf("[%d/%d] %s/%s... ", totalChecked, totalComponents, colors.Muted(componentType), componentName)

				// Load metadata to get source URL
				metadata, err := ud.loadMetadata(componentType, componentName)
				if err != nil {
					fmt.Printf("%s Failed\n", colors.Error(formatter.SymbolError))
					fmt.Printf("  %s %v\n", colors.Muted("└─"), err)
					failed++
					continue
				}

				// Check if updates are available
				hasUpdates, err := ud.HasUpdates(componentType, componentName, metadata.Source)
				if err != nil {
					fmt.Printf("%s Failed\n", colors.Error(formatter.SymbolError))
					fmt.Printf("  %s %v\n", colors.Muted("└─"), err)
					failed++
					continue
				}

				if !hasUpdates {
					fmt.Printf("%s Up to date\n", colors.Success(formatter.SymbolSuccess))
					upToDate++
					continue
				}

				// Apply the update
				fmt.Printf("%s Updating\n", colors.Warning(formatter.SymbolUpdating))

				// Remove old component directory to ensure clean re-clone
				componentDir := filepath.Join(ud.baseDir, componentType, componentName)
				if _, err := os.Stat(componentDir); err == nil {
					if err := os.RemoveAll(componentDir); err != nil {
						fmt.Printf("  %s %v\n", colors.Error(formatter.SymbolError), err)
						failed++
						continue
					}
				}

				// Re-download the component with the latest changes
				var downloadErr error
				if ud.profileName != "" {
					// Use profile-aware downloaders
					switch componentType {
					case "skills":
						dl := downloader.NewSkillDownloaderForProfile(ud.profileName)
						downloadErr = dl.DownloadSkill(metadata.Source, componentName)
					case "agents":
						dl := downloader.NewAgentDownloaderForProfile(ud.profileName)
						downloadErr = dl.DownloadAgent(metadata.Source, componentName)
					case "commands":
						dl := downloader.NewCommandDownloaderForProfile(ud.profileName)
						downloadErr = dl.DownloadCommand(metadata.Source, componentName)
					default:
						downloadErr = fmt.Errorf("unknown component type: %s", componentType)
					}
				} else {
					// Use standard downloaders
					switch componentType {
					case "skills":
						dl := downloader.NewSkillDownloader()
						downloadErr = dl.DownloadSkill(metadata.Source, componentName)
					case "agents":
						dl := downloader.NewAgentDownloader()
						downloadErr = dl.DownloadAgent(metadata.Source, componentName)
					case "commands":
						dl := downloader.NewCommandDownloader()
						downloadErr = dl.DownloadCommand(metadata.Source, componentName)
					default:
						downloadErr = fmt.Errorf("unknown component type: %s", componentType)
					}
				}

				if downloadErr != nil {
					fmt.Printf("  %s %v\n", colors.Error(formatter.SymbolError), downloadErr)
					failed++
				} else {
					fmt.Printf("  %s Updated successfully\n", colors.Success(formatter.SymbolSuccess))
					updated++
				}
			}
		}
	}

	// Print summary with box drawing
	fmt.Println()
	fmt.Println("┌────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│                            Update Summary                                  │")
	fmt.Println("├────────────────────────────────────────────────────────────────────────────┤")
	fmt.Printf("│ Total components checked:     %-45d│\n", totalChecked)
	fmt.Printf("│ %s Already up to date:         %-45d│\n", colors.Success(formatter.SymbolSuccess), upToDate)
	fmt.Printf("│ %s Successfully updated:       %-45d│\n", colors.Success(formatter.SymbolSuccess), updated)
	if failed > 0 {
		fmt.Printf("│ %s Failed:                     %-45d│\n", colors.Error(formatter.SymbolError), failed)
	}
	fmt.Println("└────────────────────────────────────────────────────────────────────────────┘")
	fmt.Println()

	return nil
}
