package updater

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tjg184/agent-smith/internal/detector"
	"github.com/tjg184/agent-smith/internal/downloader"
	"github.com/tjg184/agent-smith/internal/formatter"
	gitpkg "github.com/tjg184/agent-smith/internal/git"
	metadataPkg "github.com/tjg184/agent-smith/internal/metadata"
	"github.com/tjg184/agent-smith/internal/models"
	"github.com/tjg184/agent-smith/pkg/colors"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/paths"
	"github.com/tjg184/agent-smith/pkg/profiles"
	locksvc "github.com/tjg184/agent-smith/pkg/services/lock"
	"github.com/tjg184/agent-smith/pkg/styles"
)

type UpdateDetector struct {
	baseDir     string
	detector    *detector.RepositoryDetector
	profileName string // If non-empty, we're working with a profile
}

func NewUpdateDetector() (*UpdateDetector, error) {
	baseDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}

	var profileName string

	pm, err := profiles.NewProfileManager(nil, locksvc.NewService(logger.New(logger.LevelError)))
	if err == nil {
		activeProfile, err := pm.GetActiveProfile()
		if err == nil && activeProfile != "" {
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
		detector:    newDetector(),
		profileName: profileName,
	}, nil
}

func NewUpdateDetectorWithProfile(profile string) (*UpdateDetector, error) {
	baseDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}

	var profileName string

	if profile != "" {
		// Use explicit profile (bypasses active profile logic)
		profilesDir, err := paths.GetProfilesDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get profiles directory: %w", err)
		}
		baseDir = filepath.Join(profilesDir, profile)
		profileName = profile
		fmt.Printf("Using specified profile for updates: %s\n", profile)
	} else {
		pm, err := profiles.NewProfileManager(nil, locksvc.NewService(logger.New(logger.LevelError)))
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
		detector:    newDetector(),
		profileName: profileName,
	}, nil
}

func NewUpdateDetectorWithBaseDir(baseDir string) *UpdateDetector {
	return &UpdateDetector{
		baseDir:     baseDir,
		detector:    newDetector(),
		profileName: "", // No profile name since we're using an explicit directory
	}
}

func newDetector() *detector.RepositoryDetector {
	d := detector.NewRepositoryDetector()
	d.SuppressDuplicateWarning()
	return d
}

func (ud *UpdateDetector) LoadMetadata(componentType, componentName string) (*models.ComponentEntry, error) {
	return metadataPkg.LoadLockFileEntry(ud.baseDir, componentType, componentName)
}

func (ud *UpdateDetector) loadFromLockFile(componentType, componentName string) (*models.ComponentEntry, error) {
	return metadataPkg.LoadLockFileEntry(ud.baseDir, componentType, componentName)
}

func (ud *UpdateDetector) GetCurrentRepoSHA(repoURL string) (string, error) {
	fullURL, err := ud.detector.NormalizeURL(repoURL)
	if err != nil {
		return "", fmt.Errorf("failed to normalize URL: %w", err)
	}

	tempDir, err := os.MkdirTemp("", "agent-smith-check-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	repo, err := gitpkg.CloneBareShallow(gitpkg.NewDefaultCloner(), tempDir, fullURL)
	if err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	return ref.Hash().String(), nil
}

// HasUpdates checks if a component has updates available
func (ud *UpdateDetector) HasUpdates(componentType, componentName, repoURL string) (bool, error) {
	metadata, err := ud.LoadMetadata(componentType, componentName)
	if err != nil {
		return false, fmt.Errorf("failed to load metadata: %w", err)
	}

	if metadata.CommitHash == "" {
		fmt.Printf("Warning: %s/%s has no commit hash stored, will re-download to update lock file\n", componentType, componentName)
		return true, nil
	}

	currentSHA, err := ud.GetCurrentRepoSHA(repoURL)
	if err != nil {
		return false, fmt.Errorf("failed to get current repository SHA: %w", err)
	}

	return metadata.CommitHash != currentSHA, nil
}

func (ud *UpdateDetector) UpdateComponent(componentType, componentName, repoURL string) error {
	if ud.profileName != "" {
		fmt.Printf("%s\n\n", styles.InfoArrowFormat(fmt.Sprintf("Updating components in: %s", ud.baseDir)))
	} else {
		fmt.Printf("%s\n\n", styles.InfoArrowFormat("Checking for updates..."))
	}

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

	fmt.Printf("%s\n", styles.StatusUpdatingFormat())

	componentDir := filepath.Join(ud.baseDir, componentType, componentName)
	if _, err := os.Stat(componentDir); err == nil {
		if err := os.RemoveAll(componentDir); err != nil {
			fmt.Printf("  %s\n\n", styles.IndentedErrorFormat(fmt.Sprintf("Failed to remove old version: %v", err)))
			return fmt.Errorf("failed to remove old component directory: %w", err)
		}
	}

	ct, err := pluralToComponentType(componentType)
	if err != nil {
		return err
	}

	var dl downloader.Downloader
	var dlErr error
	if ud.profileName != "" {
		dl, dlErr = downloader.ForTypeWithProfile(ct, ud.profileName)
	} else {
		dl, dlErr = downloader.ForType(ct)
	}
	if dlErr != nil {
		return fmt.Errorf("failed to create downloader: %w", dlErr)
	}

	var downloadErr error
	downloadErr = dl.Download(repoURL, componentName)

	if downloadErr != nil {
		fmt.Printf("  %s\n\n", styles.IndentedErrorFormat(fmt.Sprintf("Update failed: %v", downloadErr)))
		return downloadErr
	}

	fmt.Printf("  %s\n\n", styles.StatusUpdatedSuccessfullyFormat())
	return nil
}

type componentUpdateInfo struct {
	Type     string
	Name     string
	Metadata *models.ComponentEntry
}

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

	var totalChecked, upToDate, updated, failed int

	for repoURL, components := range componentsByRepo {
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

		fullURL, normalizeErr := ud.detector.NormalizeURL(repoURL)
		if normalizeErr != nil {
			os.RemoveAll(tempDir)
			for _, comp := range components {
				totalChecked++
				fmt.Print(styles.ComponentProgressFormat(totalChecked, totalComponents, comp.Type, comp.Name))
				fmt.Printf("%s\n", styles.StatusFailedFormat())
				fmt.Printf("%s\n", styles.IndentedErrorFormat(fmt.Sprintf("Failed to normalize URL: %v", normalizeErr)))
				failed++
			}
			continue
		}

		repo, cloneErr := gitpkg.CloneShallow(gitpkg.NewDefaultCloner(), tempDir, fullURL)
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

			if comp.Metadata.CommitHash == "" {
				// No commit hash stored, assume update needed
				fmt.Printf("%s\n", styles.StatusUpdatingFormat())
			} else if comp.Metadata.CommitHash == currentSHA {
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

			// Download using DownloadWithRepo to reuse the cloned repository
			downloadErr := ud.downloadComponentWithRepo(comp.Type, comp.Name, fullURL, repoURL, tempDir, allDetectedComponents)

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

// groupComponentsByRepository reads the lock file and groups all installed components by
// source repository. This approach is authoritative: only components with lock entries are
// considered, which correctly handles components whose filesystemName contains subdirectory
// separators (e.g. "category/skill-name" installed from a monorepo).
func (ud *UpdateDetector) groupComponentsByRepository() (map[string][]componentUpdateInfo, int, error) {
	componentsByRepo := make(map[string][]componentUpdateInfo)
	totalComponents := 0

	for _, componentType := range paths.GetComponentTypes() {
		entries, err := metadataPkg.LoadAllComponents(ud.baseDir, componentType)
		if err != nil {
			fmt.Printf("%s\n", styles.IndentedErrorFormat(fmt.Sprintf("Failed to read %s lock entries: %v", componentType, err)))
			continue
		}

		totalComponents += len(entries)
		for _, entry := range entries {
			componentsByRepo[entry.SourceUrl] = append(componentsByRepo[entry.SourceUrl], componentUpdateInfo{
				Type:     componentType,
				Name:     entry.Name,
				Metadata: &entry.Entry,
			})
		}
	}

	return componentsByRepo, totalComponents, nil
}

// pluralToComponentType converts a lock-file plural key ("skills" / "agents" / "commands")
// to a models.ComponentType. Returns an error for unknown strings.
func pluralToComponentType(plural string) (models.ComponentType, error) {
	switch plural {
	case "skills":
		return models.ComponentSkill, nil
	case "agents":
		return models.ComponentAgent, nil
	case "commands":
		return models.ComponentCommand, nil
	default:
		return "", fmt.Errorf("unknown component type: %s", plural)
	}
}

// downloadComponentWithRepo downloads a component using DownloadWithRepo to reuse a cloned repository.
// Uses ForTypeWithProfile when a profile is active, otherwise ForType.
func (ud *UpdateDetector) downloadComponentWithRepo(componentType, componentName, fullURL, repoURL, tempDir string, detectedComponents []models.DetectedComponent) error {
	ct, err := pluralToComponentType(componentType)
	if err != nil {
		return err
	}

	var dl downloader.Downloader
	var dlErr error
	if ud.profileName != "" {
		dl, dlErr = downloader.ForTypeWithProfile(ct, ud.profileName)
	} else {
		dl, dlErr = downloader.ForType(ct)
	}
	if dlErr != nil {
		return fmt.Errorf("failed to create downloader: %w", dlErr)
	}

	return dl.DownloadWithRepo(fullURL, componentName, repoURL, tempDir, detectedComponents)
}
