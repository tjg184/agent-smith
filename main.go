package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/tgaines/agent-smith/cmd"
	"github.com/tgaines/agent-smith/internal/detector"
	"github.com/tgaines/agent-smith/internal/downloader"
	"github.com/tgaines/agent-smith/internal/fileutil"
	"github.com/tgaines/agent-smith/internal/formatter"
	"github.com/tgaines/agent-smith/internal/linker"
	"github.com/tgaines/agent-smith/internal/materializer"
	metadataPkg "github.com/tgaines/agent-smith/internal/metadata"
	"github.com/tgaines/agent-smith/internal/models"
	"github.com/tgaines/agent-smith/internal/updater"
	"github.com/tgaines/agent-smith/pkg/config"
	"github.com/tgaines/agent-smith/pkg/errors"
	"github.com/tgaines/agent-smith/pkg/logger"
	"github.com/tgaines/agent-smith/pkg/paths"
	"github.com/tgaines/agent-smith/pkg/profiles"
	"github.com/tgaines/agent-smith/pkg/project"
	"github.com/tgaines/agent-smith/pkg/services"
	installsvc "github.com/tgaines/agent-smith/pkg/services/install"
	linksvc "github.com/tgaines/agent-smith/pkg/services/link"
	profilesvc "github.com/tgaines/agent-smith/pkg/services/profile"
	statussvc "github.com/tgaines/agent-smith/pkg/services/status"
	targetsvc "github.com/tgaines/agent-smith/pkg/services/target"
	uninstallsvc "github.com/tgaines/agent-smith/pkg/services/uninstall"
	updatesvc "github.com/tgaines/agent-smith/pkg/services/update"
)

// appLogger is the global logger instance used throughout the application
var appLogger *logger.Logger

// appFormatter is the global formatter instance used for consistent output formatting
var appFormatter *formatter.Formatter

// SetVerboseMode enables informational output
// Deprecated: Use appLogger.SetLevel(logger.LevelInfo) instead
func SetVerboseMode(verbose bool) {
	if appLogger != nil {
		if verbose {
			appLogger.SetLevel(logger.LevelInfo)
		} else {
			appLogger.SetLevel(logger.LevelWarn)
		}
	}
}

// SetDebugMode enables debug output
// When debug mode is enabled, it also enables verbose mode
// Deprecated: Use appLogger.SetLevel(logger.LevelDebug) instead
func SetDebugMode(debug bool) {
	if appLogger != nil {
		if debug {
			appLogger.SetLevel(logger.LevelDebug)
		} else {
			appLogger.SetLevel(logger.LevelWarn)
		}
	}
}

// infoPrintf prints informational messages that can be suppressed
// Deprecated: Use appLogger.Info() instead
func infoPrintf(format string, a ...interface{}) {
	if appLogger != nil {
		appLogger.Info(format, a...)
	}
}

// infoPrintln prints informational messages that can be suppressed
// Deprecated: Use appLogger.Info() instead
func infoPrintln(a ...interface{}) {
	if appLogger != nil {
		appLogger.Info(fmt.Sprint(a...))
	}
}

// debugPrintf prints debug messages that can be enabled with --debug flag
// Deprecated: Use appLogger.Debug() instead
func debugPrintf(format string, a ...interface{}) {
	if appLogger != nil {
		appLogger.Debug(format, a...)
	}
}

// debugPrintln prints debug messages that can be enabled with --debug flag
// Deprecated: Use appLogger.Debug() instead
func debugPrintln(a ...interface{}) {
	if appLogger != nil {
		appLogger.Debug(fmt.Sprint(a...))
	}
}

type BulkDownloader = downloader.BulkDownloader

// Re-export types for backward compatibility
type UpdateDetector = updater.UpdateDetector
type ComponentLockFile = metadataPkg.ComponentLockFile

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

func NewUpdateDetector() *UpdateDetector {
	return updater.NewUpdateDetector()
}

func NewUpdateDetectorWithProfile(profile string) *UpdateDetector {
	return updater.NewUpdateDetectorWithProfile(profile)
}

func NewUpdateDetectorWithBaseDir(baseDir string) *UpdateDetector {
	return updater.NewUpdateDetectorWithBaseDir(baseDir)
}

func NewBulkDownloader() *BulkDownloader {
	return downloader.NewBulkDownloader()
}

// NewComponentLinker creates a new ComponentLinker with dependencies injected
func NewComponentLinker() (*linker.ComponentLinker, error) {
	debugPrintln("[DEBUG] NewComponentLinker: Creating component linker")
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}
	debugPrintf("[DEBUG] NewComponentLinker: Base agents directory: %s\n", agentsDir)

	// Check if a profile is active and use its path instead
	profileManager, err := profiles.NewProfileManager(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile manager: %w", err)
	}

	activeProfile, err := profileManager.GetActiveProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to get active profile: %w", err)
	}

	// If a profile is active, use the profile's base path instead
	if activeProfile != "" {
		debugPrintf("[DEBUG] NewComponentLinker: Active profile detected: %s\n", activeProfile)
		profilesDir, err := paths.GetProfilesDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get profiles directory: %w", err)
		}
		agentsDir = filepath.Join(profilesDir, activeProfile)
		debugPrintf("[DEBUG] NewComponentLinker: Using profile directory: %s\n", agentsDir)
		infoPrintf("Using active profile: %s\n", activeProfile)
	} else {
		debugPrintln("[DEBUG] NewComponentLinker: No active profile")
	}

	// Detect all available targets
	debugPrintln("[DEBUG] NewComponentLinker: Detecting available targets")
	targets, err := config.DetectAllTargets()
	if err != nil {
		return nil, fmt.Errorf("failed to detect targets: %w", err)
	}
	debugPrintf("[DEBUG] NewComponentLinker: Detected %d target(s)\n", len(targets))
	for i, target := range targets {
		debugPrintf("[DEBUG] NewComponentLinker: Target %d: %s\n", i+1, target.GetName())
	}

	det := detector.NewRepositoryDetector()
	// Pass the logger to the detector so it uses consistent logging
	if appLogger != nil {
		det.SetLogger(appLogger)
	}

	return linker.NewComponentLinker(agentsDir, targets, det, nil)
}

// profileManagerAdapter adapts profiles.ProfileManager to linker.ProfileManager
type profileManagerAdapter struct {
	pm *profiles.ProfileManager
}

func (pma *profileManagerAdapter) ScanProfiles() ([]*linker.Profile, error) {
	profiles, err := pma.pm.ScanProfiles()
	if err != nil {
		return nil, err
	}

	// Convert profiles.Profile to linker.Profile
	result := make([]*linker.Profile, len(profiles))
	for i, p := range profiles {
		result[i] = &linker.Profile{
			Name:        p.Name,
			BasePath:    p.BasePath,
			HasAgents:   p.HasAgents,
			HasSkills:   p.HasSkills,
			HasCommands: p.HasCommands,
		}
	}
	return result, nil
}

func (pma *profileManagerAdapter) GetActiveProfile() (string, error) {
	return pma.pm.GetActiveProfile()
}

// NewComponentLinkerWithProfileManager creates a new ComponentLinker with ProfileManager for multi-profile operations
func NewComponentLinkerWithProfileManager(pm *profiles.ProfileManager) (*linker.ComponentLinker, error) {
	debugPrintln("[DEBUG] NewComponentLinkerWithProfileManager: Creating component linker with profile manager")

	// For multi-profile view, use base directory as the starting point
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}
	debugPrintf("[DEBUG] NewComponentLinkerWithProfileManager: Base agents directory: %s\n", agentsDir)

	// Detect all available targets
	debugPrintln("[DEBUG] NewComponentLinkerWithProfileManager: Detecting available targets")
	targets, err := config.DetectAllTargets()
	if err != nil {
		return nil, fmt.Errorf("failed to detect targets: %w", err)
	}
	debugPrintf("[DEBUG] NewComponentLinkerWithProfileManager: Detected %d target(s)\n", len(targets))

	det := detector.NewRepositoryDetector()
	if appLogger != nil {
		det.SetLogger(appLogger)
	}

	// Wrap the ProfileManager in an adapter
	adapter := &profileManagerAdapter{pm: pm}

	return linker.NewComponentLinker(agentsDir, targets, det, adapter)
}

// NewComponentLinkerWithFilterAndProfile creates a new ComponentLinker with filtered targets and optional explicit profile
// targetFilter can be: "opencode", "claudecode", "all", or "" (defaults to all)
// profile can be: a specific profile name (bypasses active profile), or "" (uses active profile logic)
func NewComponentLinkerWithFilterAndProfile(targetFilter string, profile string) (*linker.ComponentLinker, error) {
	debugPrintf("[DEBUG] NewComponentLinkerWithFilterAndProfile: Creating component linker with filter=%s, profile=%s\n", targetFilter, profile)
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}
	debugPrintf("[DEBUG] NewComponentLinkerWithFilterAndProfile: Base agents directory: %s\n", agentsDir)

	// Check if an explicit profile was specified
	if profile != "" {
		// Use the explicitly specified profile (bypass active profile logic)
		debugPrintf("[DEBUG] NewComponentLinkerWithFilterAndProfile: Explicit profile specified: %s\n", profile)
		profilesDir, err := paths.GetProfilesDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get profiles directory: %w", err)
		}

		// Validate that the profile exists and is valid
		profilePath := filepath.Join(profilesDir, profile)
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			// Profile directory doesn't exist - provide helpful error with available profiles
			pm, pmErr := profiles.NewProfileManager(nil)
			if pmErr == nil {
				availableProfiles, scanErr := pm.ScanProfiles()
				if scanErr == nil && len(availableProfiles) > 0 {
					profileNames := make([]string, len(availableProfiles))
					for i, p := range availableProfiles {
						profileNames[i] = p.Name
					}
					return nil, fmt.Errorf("profile '%s' does not exist\n\nAvailable profiles:\n  - %s\n\nTo create this profile:\n  agent-smith profile create %s",
						profile, strings.Join(profileNames, "\n  - "), profile)
				}
			}
			// Fallback if we can't list profiles
			return nil, fmt.Errorf("profile '%s' does not exist\n\nTo create this profile:\n  agent-smith profile create %s\n\nTo list available profiles:\n  agent-smith profile list", profile, profile)
		}

		// Verify the profile is valid (has at least one component directory)
		profileObj := &profiles.Profile{
			Name:        profile,
			BasePath:    profilePath,
			HasAgents:   false,
			HasSkills:   false,
			HasCommands: false,
		}

		// Check which component directories exist
		if _, err := os.Stat(filepath.Join(profilePath, paths.AgentsSubDir)); err == nil {
			profileObj.HasAgents = true
		}
		if _, err := os.Stat(filepath.Join(profilePath, paths.SkillsSubDir)); err == nil {
			profileObj.HasSkills = true
		}
		if _, err := os.Stat(filepath.Join(profilePath, paths.CommandsSubDir)); err == nil {
			profileObj.HasCommands = true
		}

		if !profileObj.HasAgents && !profileObj.HasSkills && !profileObj.HasCommands {
			return nil, fmt.Errorf("profile '%s' exists but has no components\n\nThe profile directory is empty. To add components to this profile:\n  agent-smith install skill <repo-url> <name> --profile %s\n  agent-smith install agent <repo-url> <name> --profile %s\n  agent-smith install command <repo-url> <name> --profile %s",
				profile, profile, profile, profile)
		}

		agentsDir = profilePath
		debugPrintf("[DEBUG] NewComponentLinkerWithFilterAndProfile: Using explicit profile directory: %s\n", agentsDir)
		infoPrintf("Using explicit profile: %s\n", profile)
	} else {
		// No explicit profile, use active profile logic
		profileManager, err := profiles.NewProfileManager(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create profile manager: %w", err)
		}

		activeProfile, err := profileManager.GetActiveProfile()
		if err != nil {
			return nil, fmt.Errorf("failed to get active profile: %w", err)
		}

		// If a profile is active, use the profile's base path instead
		if activeProfile != "" {
			debugPrintf("[DEBUG] NewComponentLinkerWithFilterAndProfile: Active profile detected: %s\n", activeProfile)
			profilesDir, err := paths.GetProfilesDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get profiles directory: %w", err)
			}
			agentsDir = filepath.Join(profilesDir, activeProfile)
			debugPrintf("[DEBUG] NewComponentLinkerWithFilterAndProfile: Using active profile directory: %s\n", agentsDir)
			infoPrintf("Using active profile: %s\n", activeProfile)
		} else {
			debugPrintln("[DEBUG] NewComponentLinkerWithFilterAndProfile: No active profile")
		}
	}

	var targets []config.Target

	// Detect all targets first
	debugPrintln("[DEBUG] NewComponentLinkerWithFilterAndProfile: Detecting all targets")
	allTargets, err := config.DetectAllTargets()
	if err != nil {
		return nil, fmt.Errorf("failed to detect targets: %w", err)
	}
	debugPrintf("[DEBUG] NewComponentLinkerWithFilterAndProfile: Detected %d target(s)\n", len(allTargets))

	// Filter targets based on targetFilter parameter
	if targetFilter == "" || targetFilter == "all" {
		// No filter or "all" - use all detected targets
		debugPrintln("[DEBUG] NewComponentLinkerWithFilterAndProfile: Using all detected targets")
		targets = allTargets
	} else {
		// Filter for specific target
		debugPrintf("[DEBUG] NewComponentLinkerWithFilterAndProfile: Filtering for target: %s\n", targetFilter)
		for _, target := range allTargets {
			if target.GetName() == targetFilter {
				targets = append(targets, target)
				debugPrintf("[DEBUG] NewComponentLinkerWithFilterAndProfile: Found matching target: %s\n", targetFilter)
				break
			}
		}
		// If no matching target found, return error
		if len(targets) == 0 {
			debugPrintf("[DEBUG] NewComponentLinkerWithFilterAndProfile: Target '%s' not found\n", targetFilter)
			return nil, fmt.Errorf("target '%s' not found. Available targets: %v", targetFilter, getTargetNames(allTargets))
		}
	}
	debugPrintf("[DEBUG] NewComponentLinkerWithFilterAndProfile: Using %d target(s)\n", len(targets))

	det := detector.NewRepositoryDetector()
	// Pass the logger to the detector so it uses consistent logging
	if appLogger != nil {
		det.SetLogger(appLogger)
	}

	return linker.NewComponentLinker(agentsDir, targets, det, nil)
}

// NewComponentLinkerWithFilter creates a new ComponentLinker with filtered targets
// targetFilter can be: "opencode", "claudecode", "all", or "" (defaults to all)
func NewComponentLinkerWithFilter(targetFilter string) (*linker.ComponentLinker, error) {
	debugPrintf("[DEBUG] NewComponentLinkerWithFilter: Creating component linker with filter=%s\n", targetFilter)
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}
	debugPrintf("[DEBUG] NewComponentLinkerWithFilter: Base agents directory: %s\n", agentsDir)

	// Check if a profile is active and use its path instead
	profileManager, err := profiles.NewProfileManager(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile manager: %w", err)
	}

	activeProfile, err := profileManager.GetActiveProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to get active profile: %w", err)
	}

	// If a profile is active, use the profile's base path instead
	if activeProfile != "" {
		debugPrintf("[DEBUG] NewComponentLinkerWithFilter: Active profile detected: %s\n", activeProfile)
		profilesDir, err := paths.GetProfilesDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get profiles directory: %w", err)
		}
		agentsDir = filepath.Join(profilesDir, activeProfile)
		debugPrintf("[DEBUG] NewComponentLinkerWithFilter: Using profile directory: %s\n", agentsDir)
		infoPrintf("Using active profile: %s\n", activeProfile)
	} else {
		debugPrintln("[DEBUG] NewComponentLinkerWithFilter: No active profile")
	}

	var targets []config.Target

	// Detect all targets first
	debugPrintln("[DEBUG] NewComponentLinkerWithFilter: Detecting all targets")
	allTargets, err := config.DetectAllTargets()
	if err != nil {
		return nil, fmt.Errorf("failed to detect targets: %w", err)
	}
	debugPrintf("[DEBUG] NewComponentLinkerWithFilter: Detected %d target(s)\n", len(allTargets))

	// Filter targets based on targetFilter parameter
	if targetFilter == "" || targetFilter == "all" {
		// No filter or "all" - use all detected targets
		debugPrintln("[DEBUG] NewComponentLinkerWithFilter: Using all detected targets")
		targets = allTargets
	} else {
		// Filter for specific target
		debugPrintf("[DEBUG] NewComponentLinkerWithFilter: Filtering for target: %s\n", targetFilter)
		for _, target := range allTargets {
			if target.GetName() == targetFilter {
				targets = append(targets, target)
				debugPrintf("[DEBUG] NewComponentLinkerWithFilter: Found matching target: %s\n", targetFilter)
				break
			}
		}
		// If no matching target found, return error
		if len(targets) == 0 {
			debugPrintf("[DEBUG] NewComponentLinkerWithFilter: Target '%s' not found\n", targetFilter)
			return nil, fmt.Errorf("target '%s' not found. Available targets: %v", targetFilter, getTargetNames(allTargets))
		}
	}
	debugPrintf("[DEBUG] NewComponentLinkerWithFilter: Using %d target(s)\n", len(targets))

	det := detector.NewRepositoryDetector()
	// Pass the logger to the detector so it uses consistent logging
	if appLogger != nil {
		det.SetLogger(appLogger)
	}

	// Create a wrapper to adapt profiles.ProfileManager to linker.ProfileManager
	var linkerPM linker.ProfileManager
	if profileManager != nil {
		linkerPM = &profileManagerAdapter{pm: profileManager}
	}

	return linker.NewComponentLinker(agentsDir, targets, det, linkerPM)
}

// getTargetNames returns a slice of target names for error reporting
func getTargetNames(targets []config.Target) []string {
	names := make([]string, len(targets))
	for i, target := range targets {
		names[i] = target.GetName()
	}
	return names
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strings []string, separator string) string {
	if len(strings) == 0 {
		return ""
	}
	result := strings[0]
	for i := 1; i < len(strings); i++ {
		result += separator + strings[i]
	}
	return result
}

func main() {
	// Check for --debug flag before setting up handlers
	// Debug mode takes precedence and enables verbose mode automatically
	debugMode := false
	verboseMode := false
	for _, arg := range os.Args {
		if arg == "--debug" {
			debugMode = true
			break
		}
		if arg == "--verbose" {
			verboseMode = true
		}
	}

	// Initialize the global logger with appropriate level
	appLogger = logger.Default(debugMode, verboseMode)
	// Disable log level tags to maintain clean output format
	appLogger.SetShowTags(false)

	// Initialize the global formatter
	appFormatter = formatter.New()

	if debugMode {
		SetDebugMode(true)
	} else if verboseMode {
		SetVerboseMode(true)
	}

	// Initialize services with dependency injection
	profileManager, err := profiles.NewProfileManager(nil)
	if err != nil {
		log.Fatal("Failed to initialize profile manager:", err)
	}

	// Create component linker for services that need it
	componentLinker, err := NewComponentLinker()
	if err != nil {
		log.Fatal("Failed to initialize component linker:", err)
	}

	// Initialize all services
	installService := installsvc.NewService(profileManager, appLogger, appFormatter)
	updateService := updatesvc.NewService(appLogger, appFormatter)
	uninstallService := uninstallsvc.NewService(componentLinker, appLogger, appFormatter)
	targetService := targetsvc.NewService(appLogger, appFormatter)
	statusService := statussvc.NewService(profileManager, appLogger, appFormatter)
	linkService := linksvc.NewService(profileManager, appLogger, appFormatter)
	profileService := profilesvc.NewService(profileManager, appLogger, appFormatter)

	// Set up handlers for Cobra commands
	cmd.SetHandlers(
		func(repoURL, name, profile, targetDir string) {
			opts := services.InstallOptions{
				Profile:   profile,
				TargetDir: targetDir,
			}
			if err := installService.InstallSkill(repoURL, name, opts); err != nil {
				log.Fatal("Failed to install skill:", err)
			}
		},
		func(repoURL, name, profile, targetDir string) {
			opts := services.InstallOptions{
				Profile:   profile,
				TargetDir: targetDir,
			}
			if err := installService.InstallAgent(repoURL, name, opts); err != nil {
				log.Fatal("Failed to install agent:", err)
			}
		},
		func(repoURL, name, profile, targetDir string) {
			opts := services.InstallOptions{
				Profile:   profile,
				TargetDir: targetDir,
			}
			if err := installService.InstallCommand(repoURL, name, opts); err != nil {
				log.Fatal("Failed to install command:", err)
			}
		},
		func(repoURL, profile, targetDir string) {
			opts := services.InstallOptions{
				Profile:   profile,
				TargetDir: targetDir,
			}
			if err := installService.InstallBulk(repoURL, opts); err != nil {
				log.Fatal("Failed to bulk install:", err)
			}
		},
		func(componentType, componentName, profile string) {
			opts := services.UpdateOptions{
				Profile: profile,
			}
			if err := updateService.UpdateComponent(componentType, componentName, opts); err != nil {
				log.Fatal("Failed to update component:", err)
			}
		},
		func(profile string) {
			opts := services.UpdateOptions{
				Profile: profile,
			}
			if err := updateService.UpdateAll(opts); err != nil {
				log.Fatal("Failed to update all components:", err)
			}
		},
		func(componentType, componentName, targetFilter, profile string) {
			opts := services.LinkOptions{
				TargetFilter: targetFilter,
				Profile:      profile,
			}
			if err := linkService.LinkComponent(componentType, componentName, opts); err != nil {
				log.Fatal("Failed to link component:", err)
			}
		},
		func(targetFilter, profile string, allProfiles bool) {
			opts := services.LinkOptions{
				TargetFilter: targetFilter,
				Profile:      profile,
				AllProfiles:  allProfiles,
			}
			if err := linkService.LinkAll(opts); err != nil {
				log.Fatal("Failed to link all components:", err)
			}
		},
		func(componentType, targetFilter, profile string) {
			opts := services.LinkOptions{
				TargetFilter: targetFilter,
				Profile:      profile,
			}
			if err := linkService.LinkByType(componentType, opts); err != nil {
				log.Fatal("Failed to link components:", err)
			}
		},
		func() {
			if err := linkService.AutoLinkRepositories(); err != nil {
				log.Fatal("Failed to auto-link repositories:", err)
			}
		},
		func() {
			if err := linkService.ListLinked(); err != nil {
				log.Fatal("Failed to list linked components:", err)
			}
		},
		func(allProfiles bool, profileFilter []string) {
			opts := services.LinkStatusOptions{
				AllProfiles:   allProfiles,
				ProfileFilter: profileFilter,
			}
			if err := linkService.ShowStatus(opts); err != nil {
				log.Fatal("Failed to show link status:", err)
			}
		},
		func(componentType, componentName, targetFilter string) {
			opts := services.UnlinkOptions{
				TargetFilter: targetFilter,
			}
			if err := linkService.UnlinkComponent(componentType, componentName, opts); err != nil {
				log.Fatal("Failed to unlink component:", err)
			}
		},
		func(componentType, componentName, targetFilter, profile string) {
			opts := services.UnlinkOptions{
				TargetFilter: targetFilter,
				Profile:      profile,
			}
			if err := linkService.UnlinkComponent(componentType, componentName, opts); err != nil {
				log.Fatal("Failed to unlink component:", err)
			}
		},
		func(targetFilter string, force bool, allProfiles bool) {
			opts := services.UnlinkOptions{
				TargetFilter: targetFilter,
				Force:        force,
				AllProfiles:  allProfiles,
			}
			if err := linkService.UnlinkAll(opts); err != nil {
				log.Fatal("Failed to unlink all components:", err)
			}
		},
		func(targetFilter string, force bool, allProfiles bool, profile string) {
			opts := services.UnlinkOptions{
				TargetFilter: targetFilter,
				Force:        force,
				AllProfiles:  allProfiles,
				Profile:      profile,
			}
			if err := linkService.UnlinkAll(opts); err != nil {
				log.Fatal("Failed to unlink all components:", err)
			}
		},
		func(componentType, targetFilter string, force bool) {
			opts := services.UnlinkOptions{
				TargetFilter: targetFilter,
				Force:        force,
			}
			if err := linkService.UnlinkByType(componentType, opts); err != nil {
				log.Fatal("Failed to unlink components:", err)
			}
		},
		func(componentType, targetFilter string, force bool, profile string) {
			opts := services.UnlinkOptions{
				TargetFilter: targetFilter,
				Force:        force,
				Profile:      profile,
			}
			if err := linkService.UnlinkByType(componentType, opts); err != nil {
				log.Fatal("Failed to unlink components:", err)
			}
		},
		func(componentType, componentName, profile string) {
			opts := services.UninstallOptions{
				Profile: profile,
			}
			if err := uninstallService.UninstallComponent(componentType, componentName, opts); err != nil {
				log.Fatal("Failed to uninstall component:", err)
			}
		},
		func(repoURL string, force bool) {
			opts := services.UninstallOptions{
				Force: force,
			}
			if err := uninstallService.UninstallAllFromSource(repoURL, opts); err != nil {
				log.Fatal("Failed to uninstall components:", err)
			}
		},
		func(profileFilter []string, activeOnly bool, typeFilter string) {
			opts := services.ListProfileOptions{
				ProfileFilter: profileFilter,
				ActiveOnly:    activeOnly,
				TypeFilter:    typeFilter,
			}
			if err := profileService.ListProfiles(opts); err != nil {
				log.Fatal("Failed to list profiles:", err)
			}
		},
		func(profileName string) {
			if err := profileService.ShowProfile(profileName); err != nil {
				log.Fatal("Failed to show profile:", err)
			}
		},
		func(profileName string) {
			if err := profileService.CreateProfile(profileName); err != nil {
				log.Fatal("Failed to create profile:", err)
			}
		},
		func(profileName string) {
			if err := profileService.DeleteProfile(profileName); err != nil {
				log.Fatal("Failed to delete profile:", err)
			}
		},
		func(profileName string) {
			if err := profileService.ActivateProfile(profileName); err != nil {
				log.Fatal("Failed to activate profile:", err)
			}
		},
		func() {
			if err := profileService.DeactivateProfile(); err != nil {
				log.Fatal("Failed to deactivate profile:", err)
			}
		},
		func(componentType, profileName, componentName string) {
			if err := profileService.AddComponent(componentType, profileName, componentName); err != nil {
				log.Fatal("Failed to add component:", err)
			}
		},
		func(componentType, sourceProfile, targetProfile, componentName string) {
			if err := profileService.CopyComponent(sourceProfile, targetProfile, componentType, componentName); err != nil {
				log.Fatal("Failed to copy component:", err)
			}
		},
		func(componentType, profileName, componentName string) {
			if err := profileService.RemoveComponent(profileName, componentType, componentName); err != nil {
				log.Fatal("Failed to remove component:", err)
			}
		},
		func(targetProfile string, sourceProfiles []string) {
			if err := profileService.CherryPickComponents(targetProfile, sourceProfiles); err != nil {
				log.Fatal("Failed to cherry-pick components:", err)
			}
		},
		func() {
			if err := statusService.ShowSystemStatus(); err != nil {
				log.Fatal("Failed to show system status:", err)
			}
		},
		func(name, path string) {
			if err := targetService.AddCustomTarget(name, path); err != nil {
				log.Fatal("Failed to add custom target:", err)
			}
		},
		func(name string) {
			if err := targetService.RemoveCustomTarget(name); err != nil {
				log.Fatal("Failed to remove custom target:", err)
			}
		},
		func() {
			if err := targetService.ListTargets(); err != nil {
				log.Fatal("Failed to list targets:", err)
			}
		},
		func(componentType, componentName, target, projectDir string, force, dryRun bool, profile string) {
			// Import necessary packages at the top of main.go
			// - "github.com/tgaines/agent-smith/pkg/project"
			// - "github.com/tgaines/agent-smith/internal/materializer"
			// - "github.com/tgaines/agent-smith/internal/metadata"

			// If target is not provided via flag, check environment variable
			if target == "" {
				target = config.GetTargetFromEnv()
				if target != "" {
					debugPrintf("[DEBUG] Using target from AGENT_SMITH_TARGET environment variable: %s\n", target)
				}
			}

			// Validate target is provided
			if target == "" {
				fmt.Println(errors.NewMissingTargetFlagError("materialize skill <name>").Format())
				os.Exit(1)
			}

			// Show dry-run header if enabled
			if dryRun {
				infoPrintln("=== DRY RUN MODE ===")
				infoPrintln("No changes will be made to the filesystem")
				infoPrintln("")
			}

			// Determine project root
			var projectRoot string
			var err error
			if projectDir != "" {
				// Use specified project directory
				projectRoot, err = filepath.Abs(projectDir)
				if err != nil {
					log.Fatalf("Failed to resolve project directory: %v", err)
				}
			} else {
				// Auto-detect project root
				projectRoot, err = project.FindProjectRoot()
				if err != nil {
					log.Fatalf("Failed to find project root: %v", err)
				}
			}

			// Get source directory (from ~/.agent-smith/ or profile)
			baseDir, err := paths.GetAgentsDir()
			if err != nil {
				log.Fatalf("Failed to get agent-smith directory: %v", err)
			}

			// Check if there's an active profile or --profile flag
			profileMgr, err := profiles.NewProfileManager(nil)
			if err != nil {
				log.Fatalf("Failed to initialize profile manager: %v", err)
			}

			// Determine source profile based on --profile flag or active profile
			var sourceProfile string
			if profile != "" {
				// --profile flag is specified
				if profile == "base" {
					// Special value "base" means use ~/.agent-smith/
					sourceProfile = ""
					debugPrintln("[DEBUG] Using base directory (~/.agent-smith/) as source via --profile base")
				} else {
					// Validate that the specified profile exists
					profilesList, err := profileMgr.ScanProfiles()
					if err != nil {
						log.Fatalf("Failed to scan profiles: %v", err)
					}

					profileExists := false
					for _, p := range profilesList {
						if p.Name == profile {
							profileExists = true
							break
						}
					}

					if !profileExists {
						// Build list of available profiles for error message
						var availableProfiles []string
						for _, p := range profilesList {
							availableProfiles = append(availableProfiles, p.Name)
						}
						fmt.Println(errors.NewProfileNotFoundError(profile).Format())
						if len(availableProfiles) > 0 {
							infoPrintln("\nAvailable profiles:")
							for _, name := range availableProfiles {
								infoPrintf("  - %s\n", name)
							}
						} else {
							infoPrintln("\nNo profiles found. Create one with: agent-smith profile create <name>")
						}
						os.Exit(1)
					}

					// Use the specified profile
					profilesDir, err := paths.GetProfilesDir()
					if err != nil {
						log.Fatalf("Failed to get profiles directory: %v", err)
					}
					baseDir = filepath.Join(profilesDir, profile)
					sourceProfile = profile
					debugPrintf("[DEBUG] Using specified profile '%s' as source via --profile\n", profile)
				}
			} else {
				// No --profile flag, check active profile
				activeProfile, err := profileMgr.GetActiveProfile()
				if err != nil {
					log.Fatalf("Failed to check active profile: %v", err)
				}

				if activeProfile != "" {
					// Use active profile directory as source
					profilesDir, err := paths.GetProfilesDir()
					if err != nil {
						log.Fatalf("Failed to get profiles directory: %v", err)
					}
					baseDir = filepath.Join(profilesDir, activeProfile)
					sourceProfile = activeProfile
					debugPrintf("[DEBUG] Using active profile '%s' as source\n", activeProfile)
				} else {
					// No active profile, use base directory
					debugPrintln("[DEBUG] No active profile, using base directory (~/.agent-smith/) as source")
				}
			}

			var componentSourceDir string
			switch componentType {
			case "skills":
				componentSourceDir = filepath.Join(baseDir, "skills", componentName)
			case "agents":
				componentSourceDir = filepath.Join(baseDir, "agents", componentName)
			case "commands":
				componentSourceDir = filepath.Join(baseDir, "commands", componentName)
			default:
				log.Fatalf("Invalid component type: %s", componentType)
			}

			// Check if component exists
			if _, err := os.Stat(componentSourceDir); os.IsNotExist(err) {
				var sourcePath string
				if sourceProfile != "" {
					sourcePath = fmt.Sprintf("profile '%s' (%s/%s/)", sourceProfile, sourceProfile, componentType)
				} else {
					sourcePath = fmt.Sprintf("~/.agent-smith/%s/", componentType)
				}
				fmt.Println(errors.NewComponentNotInstalledError(componentType, componentName, sourcePath).Format())
				os.Exit(1)
			}

			// Get lock file entry for provenance
			lockEntry, err := metadataPkg.LoadLockFileEntry(baseDir, componentType, componentName)
			if err != nil {
				log.Fatalf("Failed to load component metadata: %v", err)
			}

			// Calculate source hash
			sourceHash, err := materializer.CalculateDirectoryHash(componentSourceDir)
			if err != nil {
				log.Fatalf("Failed to calculate source hash: %v", err)
			}

			// Determine which targets to materialize to
			var targets []string
			if target == "all" {
				targets = []string{"opencode", "claudecode"}
			} else {
				targets = []string{target}
			}

			// Track counts for summary
			successCount := 0
			skipCount := 0

			// Materialize to each target
			for _, targetName := range targets {
				targetDir := project.GetTargetDirectory(projectRoot, targetName)
				if targetDir == "" {
					fmt.Println(errors.NewInvalidTargetError(targetName).Format())
					os.Exit(1)
				}

				// Determine destination path
				destPath := filepath.Join(targetDir, componentType, componentName)

				// Check if component already exists
				if _, err := os.Stat(destPath); err == nil {
					// Component exists, check if it's identical
					match, err := materializer.DirectoriesMatch(componentSourceDir, destPath)
					if err != nil {
						log.Fatalf("Failed to compare directories: %v", err)
					}
					if match {
						if dryRun {
							infoPrintf("⊘ Would skip %s '%s' to %s (already exists and identical)\n", componentType, componentName, targetName)
						} else {
							infoPrintf("⊘ Skipped %s '%s' to %s (already exists and identical)\n", componentType, componentName, targetName)
						}
						skipCount++
						continue
					} else {
						// Component exists and differs
						if !force {
							if dryRun {
								infoPrintf("⚠ Would fail: Component '%s' already exists in %s and differs (use --force to overwrite)\n", componentName, targetName)
								continue
							}
							log.Fatalf("Component '%s' already exists in %s and differs.\n\nUse --force to overwrite", componentName, targetName)
						}
						// Force flag is set, would remove existing component before copying
						if dryRun {
							infoPrintf("⚠ Would overwrite %s '%s' in %s (--force)\n", componentType, componentName, targetName)
						} else {
							infoPrintf("⚠ Overwriting %s '%s' in %s (--force)\n", componentType, componentName, targetName)
							if err := os.RemoveAll(destPath); err != nil {
								log.Fatalf("Failed to remove existing component: %v", err)
							}
						}
					}
				} else {
					// Component doesn't exist
					if dryRun {
						// Check if target structure needs to be created
						if _, err := os.Stat(targetDir); os.IsNotExist(err) {
							infoPrintf("%s Would create project structure: %s/ (skills/, agents/, commands/)\n", formatter.SymbolSuccess, targetDir)
						}
					}
				}

				if dryRun {
					// In dry-run mode, just show what would happen
					infoPrintf("%s Would materialize %s '%s' to %s\n", formatter.SymbolSuccess, componentType, componentName, targetName)
					infoPrintf("  Source:      %s\n", componentSourceDir)
					if sourceProfile != "" {
						infoPrintf("  From Profile: %s\n", sourceProfile)
					}
					infoPrintf("  Destination: %s\n", destPath)
					infoPrintf("  Provenance:  %s @ %s\n", lockEntry.SourceUrl, lockEntry.CommitHash[:8])
					successCount++
				} else {
					// Ensure target structure exists
					structureCreated, err := project.EnsureTargetStructure(targetDir)
					if err != nil {
						log.Fatalf("Failed to create target structure: %v", err)
					}
					if structureCreated {
						infoPrintf("%s Created project structure: %s/ (skills/, agents/, commands/)\n", formatter.SymbolSuccess, targetDir)
					}

					// Copy the component
					if err := materializer.CopyDirectory(componentSourceDir, destPath); err != nil {
						log.Fatalf("Failed to copy component: %v", err)
					}

					// Calculate current hash (should match source hash immediately after copy)
					currentHash := sourceHash

					// Load or create materialization metadata
					matMetadata, err := project.LoadMaterializationMetadata(targetDir)
					if err != nil {
						log.Fatalf("Failed to load materialization metadata: %v", err)
					}

					// Add entry to metadata
					project.AddMaterializationEntry(
						matMetadata,
						componentType,
						componentName,
						lockEntry.SourceUrl,
						lockEntry.SourceType,
						sourceProfile, // sourceProfile from active profile
						lockEntry.CommitHash,
						lockEntry.OriginalPath,
						sourceHash,
						currentHash,
					)

					// Save metadata
					if err := project.SaveMaterializationMetadata(targetDir, matMetadata); err != nil {
						log.Fatalf("Failed to save materialization metadata: %v", err)
					}

					infoPrintf("%s Materialized %s '%s' to %s\n", formatter.SymbolSuccess, componentType, componentName, targetName)
					infoPrintf("  Source:      %s\n", componentSourceDir)
					if sourceProfile != "" {
						infoPrintf("  From Profile: %s\n", sourceProfile)
					}
					infoPrintf("  Destination: %s\n", destPath)
					successCount++
				}
			}

			// Print summary (always shown, even without --verbose)
			if len(targets) > 0 {
				green := color.New(color.FgGreen).SprintFunc()
				fmt.Println()

				// Build summary message
				if dryRun {
					if successCount > 0 {
						msg := fmt.Sprintf("%s Would materialize to %d target(s)", green(formatter.SymbolSuccess), successCount)
						if skipCount > 0 {
							msg += fmt.Sprintf(" (%d skipped)", skipCount)
						}
						fmt.Println(msg)
					} else if skipCount > 0 {
						fmt.Printf("⊘ Would skip %d target(s) (already exists and identical)\n", skipCount)
					}
				} else {
					if successCount > 0 {
						msg := fmt.Sprintf("%s Successfully materialized to %d target(s)", green(formatter.SymbolSuccess), successCount)
						if skipCount > 0 {
							msg += fmt.Sprintf(" (%d skipped)", skipCount)
						}
						fmt.Println(msg)
					} else if skipCount > 0 {
						fmt.Printf("⊘ Skipped %d target(s) (already exists and identical)\n", skipCount)
					}
				}
			}

			if dryRun {
				infoPrintln("")
				infoPrintln("=== DRY RUN COMPLETE ===")
				infoPrintln("Run without --dry-run to apply these changes")
			}
		},
		func(target, projectDir string, force, dryRun bool, profile string) {
			// Define color functions
			green := color.New(color.FgGreen).SprintFunc()

			// Show dry-run header if enabled
			if dryRun {
				infoPrintln("=== DRY RUN MODE ===")
				infoPrintln("No changes will be made to the filesystem")
				infoPrintln("")
			}

			// If target is not provided via flag, check environment variable
			if target == "" {
				target = config.GetTargetFromEnv()
				if target != "" {
					debugPrintf("[DEBUG] Using target from AGENT_SMITH_TARGET environment variable: %s\n", target)
				}
			}

			// Validate target is provided
			if target == "" {
				fmt.Println(errors.NewMissingTargetFlagError("materialize all").Format())
				os.Exit(1)
			}

			// Determine project root
			var projectRoot string
			var err error
			if projectDir != "" {
				// Use specified project directory
				projectRoot, err = filepath.Abs(projectDir)
				if err != nil {
					log.Fatalf("Failed to resolve project directory: %v", err)
				}
			} else {
				// Auto-detect project root
				projectRoot, err = project.FindProjectRoot()
				if err != nil {
					log.Fatalf("Failed to find project root: %v", err)
				}
			}

			// Get source directory (from ~/.agent-smith/ or profile)
			baseDir, err := paths.GetAgentsDir()
			if err != nil {
				log.Fatalf("Failed to get agent-smith directory: %v", err)
			}

			// Check if there's an active profile or --profile flag
			profileMgr, err := profiles.NewProfileManager(nil)
			if err != nil {
				log.Fatalf("Failed to initialize profile manager: %v", err)
			}

			// Determine source profile based on --profile flag or active profile
			var sourceProfile string
			if profile != "" {
				// --profile flag is specified
				if profile == "base" {
					// Special value "base" means use ~/.agent-smith/
					sourceProfile = ""
					debugPrintln("[DEBUG] Using base directory (~/.agent-smith/) as source via --profile base")
				} else {
					// Validate that the specified profile exists
					profilesList, err := profileMgr.ScanProfiles()
					if err != nil {
						log.Fatalf("Failed to scan profiles: %v", err)
					}

					profileExists := false
					for _, p := range profilesList {
						if p.Name == profile {
							profileExists = true
							break
						}
					}

					if !profileExists {
						// Build list of available profiles for error message
						var availableProfiles []string
						for _, p := range profilesList {
							availableProfiles = append(availableProfiles, p.Name)
						}
						fmt.Println(errors.NewProfileNotFoundError(profile).Format())
						if len(availableProfiles) > 0 {
							infoPrintln("\nAvailable profiles:")
							for _, name := range availableProfiles {
								infoPrintf("  - %s\n", name)
							}
						} else {
							infoPrintln("\nNo profiles found. Create one with: agent-smith profile create <name>")
						}
						os.Exit(1)
					}

					// Use the specified profile
					profilesDir, err := paths.GetProfilesDir()
					if err != nil {
						log.Fatalf("Failed to get profiles directory: %v", err)
					}
					baseDir = filepath.Join(profilesDir, profile)
					sourceProfile = profile
					debugPrintf("[DEBUG] Using specified profile '%s' as source via --profile\n", profile)
				}
			} else {
				// No --profile flag, check active profile
				activeProfile, err := profileMgr.GetActiveProfile()
				if err != nil {
					log.Fatalf("Failed to check active profile: %v", err)
				}

				if activeProfile != "" {
					// Use active profile directory as source
					profilesDir, err := paths.GetProfilesDir()
					if err != nil {
						log.Fatalf("Failed to get profiles directory: %v", err)
					}
					baseDir = filepath.Join(profilesDir, activeProfile)
					sourceProfile = activeProfile
					debugPrintf("[DEBUG] Using active profile '%s' as source\n", activeProfile)
				} else {
					// No active profile, use base directory
					debugPrintln("[DEBUG] No active profile, using base directory (~/.agent-smith/) as source")
				}
			}

			// Determine which targets to materialize to
			var targets []string
			if target == "all" {
				targets = []string{"opencode", "claudecode"}
			} else {
				targets = []string{target}
			}

			// Track counts for summary
			totalComponents := 0
			successCount := 0
			skipCount := 0
			errorCount := 0
			var errorMessages []string

			// Materialize all component types
			componentTypes := []string{"skills", "agents", "commands"}
			for _, componentType := range componentTypes {
				// Get all component names from lock file
				componentNames, err := metadataPkg.GetAllComponentNames(baseDir, componentType)
				if err != nil {
					errorMsg := fmt.Sprintf("Failed to load %s from lock file: %v", componentType, err)
					errorMessages = append(errorMessages, errorMsg)
					errorCount++
					continue
				}

				if len(componentNames) == 0 {
					debugPrintf("[DEBUG] No %s found in lock file\n", componentType)
					continue
				}

				totalComponents += len(componentNames)

				// Materialize each component
				for _, componentName := range componentNames {
					componentSourceDir := filepath.Join(baseDir, componentType, componentName)

					// Check if component exists
					if _, err := os.Stat(componentSourceDir); os.IsNotExist(err) {
						var errorMsg string
						if sourceProfile != "" {
							errorMsg = fmt.Sprintf("Component '%s' (%s) not found in profile '%s' at %s/%s/", componentName, componentType, sourceProfile, sourceProfile, componentType)
						} else {
							errorMsg = fmt.Sprintf("Component '%s' (%s) not found in ~/.agent-smith/%s/", componentName, componentType, componentType)
						}
						errorMessages = append(errorMessages, errorMsg)
						errorCount++
						continue
					}

					// Get lock file entry for provenance
					lockEntry, err := metadataPkg.LoadLockFileEntry(baseDir, componentType, componentName)
					if err != nil {
						errorMsg := fmt.Sprintf("Failed to load metadata for %s '%s': %v", componentType, componentName, err)
						errorMessages = append(errorMessages, errorMsg)
						errorCount++
						continue
					}

					// Calculate source hash
					sourceHash, err := materializer.CalculateDirectoryHash(componentSourceDir)
					if err != nil {
						errorMsg := fmt.Sprintf("Failed to calculate hash for %s '%s': %v", componentType, componentName, err)
						errorMessages = append(errorMessages, errorMsg)
						errorCount++
						continue
					}

					// Materialize to each target
					for _, targetName := range targets {
						targetDir := project.GetTargetDirectory(projectRoot, targetName)
						if targetDir == "" {
							errorMsg := fmt.Sprintf("Invalid target: %s", targetName)
							errorMessages = append(errorMessages, errorMsg)
							errorCount++
							continue
						}

						// Determine destination path
						destPath := filepath.Join(targetDir, componentType, componentName)

						// Check if component already exists
						componentSkipped := false
						if _, err := os.Stat(destPath); err == nil {
							// Component exists, check if it's identical
							match, err := materializer.DirectoriesMatch(componentSourceDir, destPath)
							if err != nil {
								errorMsg := fmt.Sprintf("Failed to compare directories for %s '%s': %v", componentType, componentName, err)
								errorMessages = append(errorMessages, errorMsg)
								errorCount++
								continue
							}
							if match {
								if dryRun {
									infoPrintf("⊘ Would skip %s '%s' to %s (already exists and identical)\n", componentType, componentName, targetName)
								} else {
									infoPrintf("⊘ Skipped %s '%s' to %s (already exists and identical)\n", componentType, componentName, targetName)
								}
								skipCount++
								componentSkipped = true
								continue
							} else {
								// Component exists and differs
								if !force {
									var errorMsg string
									if dryRun {
										errorMsg = fmt.Sprintf("Would fail: Component '%s' (%s) already exists in %s and differs (use --force to overwrite)", componentName, componentType, targetName)
									} else {
										errorMsg = fmt.Sprintf("Component '%s' (%s) already exists in %s and differs. Use --force to overwrite", componentName, componentType, targetName)
									}
									errorMessages = append(errorMessages, errorMsg)
									errorCount++
									continue
								}
								// Force flag is set
								if dryRun {
									infoPrintf("⚠ Would overwrite %s '%s' in %s (--force)\n", componentType, componentName, targetName)
								} else {
									infoPrintf("⚠ Overwriting %s '%s' in %s (--force)\n", componentType, componentName, targetName)
									if err := os.RemoveAll(destPath); err != nil {
										errorMsg := fmt.Sprintf("Failed to remove existing %s '%s': %v", componentType, componentName, err)
										errorMessages = append(errorMessages, errorMsg)
										errorCount++
										continue
									}
								}
							}
						} else {
							// Component doesn't exist
							if dryRun {
								// Check if target structure needs to be created
								if _, err := os.Stat(targetDir); os.IsNotExist(err) {
									infoPrintf("%s Would create project structure: %s/ (skills/, agents/, commands/)\n", formatter.SymbolSuccess, targetDir)
								}
							}
						}

						if componentSkipped {
							continue
						}

						if dryRun {
							// In dry-run mode, just show what would happen
							infoPrintf("%s Would materialize %s '%s' to %s\n", formatter.SymbolSuccess, componentType, componentName, targetName)
							successCount++
						} else {
							// Ensure target structure exists
							structureCreated, err := project.EnsureTargetStructure(targetDir)
							if err != nil {
								errorMsg := fmt.Sprintf("Failed to create target structure: %v", err)
								errorMessages = append(errorMessages, errorMsg)
								errorCount++
								continue
							}
							if structureCreated {
								infoPrintf("%s Created project structure: %s/ (skills/, agents/, commands/)\n", formatter.SymbolSuccess, targetDir)
							}

							// Copy the component
							if err := materializer.CopyDirectory(componentSourceDir, destPath); err != nil {
								errorMsg := fmt.Sprintf("Failed to copy %s '%s': %v", componentType, componentName, err)
								errorMessages = append(errorMessages, errorMsg)
								errorCount++
								continue
							}

							// Calculate current hash (should match source hash immediately after copy)
							currentHash := sourceHash

							// Load or create materialization metadata
							matMetadata, err := project.LoadMaterializationMetadata(targetDir)
							if err != nil {
								errorMsg := fmt.Sprintf("Failed to load materialization metadata: %v", err)
								errorMessages = append(errorMessages, errorMsg)
								errorCount++
								continue
							}

							// Add entry to metadata
							project.AddMaterializationEntry(
								matMetadata,
								componentType,
								componentName,
								lockEntry.SourceUrl,
								lockEntry.SourceType,
								sourceProfile, // sourceProfile from active profile
								lockEntry.CommitHash,
								lockEntry.OriginalPath,
								sourceHash,
								currentHash,
							)

							// Save metadata
							if err := project.SaveMaterializationMetadata(targetDir, matMetadata); err != nil {
								errorMsg := fmt.Sprintf("Failed to save materialization metadata: %v", err)
								errorMessages = append(errorMessages, errorMsg)
								errorCount++
								continue
							}

							infoPrintf("%s Materialized %s '%s' to %s\n", formatter.SymbolSuccess, componentType, componentName, targetName)
							successCount++
						}
					}
				}
			}

			// Print summary (always shown, even without --verbose)
			fmt.Println()

			// Build concise summary message
			if dryRun {
				if successCount > 0 || skipCount > 0 || errorCount > 0 {
					msg := fmt.Sprintf("%s Would materialize %d of %d component(s)", green(formatter.SymbolSuccess), successCount, totalComponents)

					// Add skip/error info inline
					var details []string
					if skipCount > 0 {
						details = append(details, fmt.Sprintf("%d skipped", skipCount))
					}
					if errorCount > 0 {
						details = append(details, fmt.Sprintf("%d errors", errorCount))
					}
					if len(details) > 0 {
						msg += fmt.Sprintf(" (%s)", strings.Join(details, ", "))
					}
					fmt.Println(msg)
				}
			} else {
				if successCount > 0 {
					msg := fmt.Sprintf("%s Successfully materialized %d component(s)", green(formatter.SymbolSuccess), successCount)

					// Add skip/error info inline
					var details []string
					if skipCount > 0 {
						details = append(details, fmt.Sprintf("%d skipped", skipCount))
					}
					if errorCount > 0 {
						details = append(details, fmt.Sprintf("%d errors", errorCount))
					}
					if len(details) > 0 {
						msg += fmt.Sprintf(" (%s)", strings.Join(details, ", "))
					}
					fmt.Println(msg)
				} else if skipCount > 0 {
					fmt.Printf("⊘ All %d component(s) already materialized and identical\n", skipCount)
				}
			}

			// Show error details if any
			if errorCount > 0 {
				fmt.Println("\nErrors:")
				for _, errorMsg := range errorMessages {
					fmt.Printf("  - %s\n", errorMsg)
				}
			}

			if dryRun {
				infoPrintln("")
				infoPrintln("=== DRY RUN COMPLETE ===")
				infoPrintln("Run without --dry-run to apply these changes")
			}

			if errorCount > 0 {
				os.Exit(1)
			}
		},
		func(projectDir string) {
			// Handle materialize list command
			green := color.New(color.FgGreen).SprintFunc()
			yellow := color.New(color.FgYellow).SprintFunc()

			var projectRoot string
			var err error
			if projectDir != "" {
				// Use specified project directory
				projectRoot, err = filepath.Abs(projectDir)
				if err != nil {
					log.Fatalf("Failed to resolve project directory: %v", err)
				}
			} else {
				// Auto-detect project root
				projectRoot, err = project.FindProjectRoot()
				if err != nil {
					log.Fatalf("Failed to find project root: %v", err)
				}
			}

			// Display project information
			appFormatter.Info("Materialized Components in %s:", projectRoot)
			appFormatter.EmptyLine()

			// Track if any components were found
			foundAny := false

			// Check each target
			for _, targetName := range []string{"opencode", "claudecode"} {
				targetDir := project.GetTargetDirectory(projectRoot, targetName)

				// Check if target directory exists
				if _, err := os.Stat(targetDir); os.IsNotExist(err) {
					continue
				}

				// Load materialization metadata
				metadata, err := project.LoadMaterializationMetadata(targetDir)
				if err != nil {
					debugPrintf("[DEBUG] Failed to load metadata for %s: %v\n", targetName, err)
					continue
				}

				// Count total components
				totalComponents := len(metadata.Skills) + len(metadata.Agents) + len(metadata.Commands)

				if totalComponents == 0 {
					continue
				}

				foundAny = true

				// Display target header
				var targetLabel string
				if targetName == "opencode" {
					targetLabel = "OpenCode (.opencode/)"
				} else {
					targetLabel = "Claude Code (.claude/)"
				}
				appFormatter.Info("%s %s", green(formatter.SymbolSuccess), targetLabel)

				// Display skills
				if len(metadata.Skills) > 0 {
					appFormatter.Info("  Skills (%d):", len(metadata.Skills))
					for name, meta := range metadata.Skills {
						sourceInfo := meta.Source
						if meta.SourceProfile != "" {
							sourceInfo = fmt.Sprintf("%s (profile: %s)", meta.Source, meta.SourceProfile)
						}
						appFormatter.Info("    • %-30s (from %s)", name, sourceInfo)
					}
				}

				// Display agents
				if len(metadata.Agents) > 0 {
					appFormatter.Info("  Agents (%d):", len(metadata.Agents))
					for name, meta := range metadata.Agents {
						sourceInfo := meta.Source
						if meta.SourceProfile != "" {
							sourceInfo = fmt.Sprintf("%s (profile: %s)", meta.Source, meta.SourceProfile)
						}
						appFormatter.Info("    • %-30s (from %s)", name, sourceInfo)
					}
				}

				// Display commands
				if len(metadata.Commands) > 0 {
					appFormatter.Info("  Commands (%d):", len(metadata.Commands))
					for name, meta := range metadata.Commands {
						sourceInfo := meta.Source
						if meta.SourceProfile != "" {
							sourceInfo = fmt.Sprintf("%s (profile: %s)", meta.Source, meta.SourceProfile)
						}
						appFormatter.Info("    • %-30s (from %s)", name, sourceInfo)
					}
				}

				appFormatter.EmptyLine()
			}

			if !foundAny {
				appFormatter.Info("%s No components materialized yet", yellow(formatter.SymbolWarning))
				appFormatter.EmptyLine()
				appFormatter.Info("To materialize components:")
				appFormatter.Info("  agent-smith materialize skill <name> --target opencode")
				appFormatter.Info("  agent-smith materialize all --target opencode")
			}
		},
		func(componentType, componentName, target, projectDir string) {
			// Handle materialize info command
			green := color.New(color.FgGreen).SprintFunc()
			yellow := color.New(color.FgYellow).SprintFunc()
			red := color.New(color.FgRed).SprintFunc()
			cyan := color.New(color.FgCyan).SprintFunc()
			bold := color.New(color.Bold).SprintFunc()

			var projectRoot string
			var err error
			if projectDir != "" {
				// Use specified project directory
				projectRoot, err = filepath.Abs(projectDir)
				if err != nil {
					log.Fatalf("Failed to resolve project directory: %v", err)
				}
			} else {
				// Auto-detect project root
				projectRoot, err = project.FindProjectRoot()
				if err != nil {
					log.Fatalf("Failed to find project root: %v", err)
				}
			}

			// Track if we found the component in any target
			foundInAnyTarget := false

			// Determine which targets to check
			var targetsToCheck []string
			if target != "" {
				targetsToCheck = []string{target}
			} else {
				targetsToCheck = []string{"opencode", "claudecode"}
			}

			// Check each target
			for _, targetName := range targetsToCheck {
				targetDir := project.GetTargetDirectory(projectRoot, targetName)

				// Check if target directory exists
				if _, err := os.Stat(targetDir); os.IsNotExist(err) {
					if target != "" {
						// User specified a target that doesn't exist
						fmt.Println(errors.NewTargetDirectoryNotFoundError(targetName).Format())
						os.Exit(1)
					}
					continue
				}

				// Load materialization metadata
				metadata, err := project.LoadMaterializationMetadata(targetDir)
				if err != nil {
					debugPrintf("[DEBUG] Failed to load metadata for %s: %v\n", targetName, err)
					continue
				}

				// Get the component map for the given type
				componentMap := metadata.GetComponentMap(componentType)
				if componentMap == nil {
					log.Fatalf("Invalid component type: %s (must be skills, agents, or commands)", componentType)
				}

				// Look up the component
				meta, exists := componentMap[componentName]
				if !exists {
					if target != "" {
						// User specified a target but component not found
						appFormatter.Info("%s Component '%s' not found in %s target", red(formatter.SymbolError), componentName, targetName)
					}
					continue
				}

				foundInAnyTarget = true

				// Display target header
				var targetLabel string
				if targetName == "opencode" {
					targetLabel = "OpenCode (.opencode/)"
				} else {
					targetLabel = "Claude Code (.claude/)"
				}

				appFormatter.EmptyLine()
				appFormatter.Info("%s Provenance Information - %s", green(formatter.SymbolSuccess), bold(targetLabel))
				appFormatter.EmptyLine()

				// Display component information
				appFormatter.Info("  %s: %s", cyan("Component"), componentName)
				appFormatter.Info("  %s: %s", cyan("Type"), componentType)
				appFormatter.EmptyLine()

				// Display source information
				appFormatter.Info("  %s", bold("Source Information:"))
				appFormatter.Info("    %s: %s", cyan("Repository"), meta.Source)
				appFormatter.Info("    %s: %s", cyan("Source Type"), meta.SourceType)
				if meta.SourceProfile != "" {
					appFormatter.Info("    %s: %s", cyan("Profile"), meta.SourceProfile)
				}
				appFormatter.Info("    %s: %s", cyan("Commit Hash"), meta.CommitHash)
				appFormatter.Info("    %s: %s", cyan("Original Path"), meta.OriginalPath)
				appFormatter.EmptyLine()

				// Display materialization information
				appFormatter.Info("  %s", bold("Materialization:"))
				appFormatter.Info("    %s: %s", cyan("Materialized At"), meta.MaterializedAt)
				appFormatter.Info("    %s: %s", cyan("Target Directory"), targetDir)
				appFormatter.EmptyLine()

				// Display hash information for sync status
				appFormatter.Info("  %s", bold("Sync Status:"))
				appFormatter.Info("    %s: %s", cyan("Source Hash"), meta.SourceHash)

				// Recalculate current hash from the actual directory
				componentPath := filepath.Join(targetDir, componentType, componentName)
				actualCurrentHash, err := materializer.CalculateDirectoryHash(componentPath)
				if err != nil {
					debugPrintf("[DEBUG] Failed to calculate current hash: %v\n", err)
					// Fall back to stored hash
					actualCurrentHash = meta.CurrentHash
				}

				appFormatter.Info("    %s: %s", cyan("Current Hash"), actualCurrentHash)

				// Check if hashes match
				if meta.SourceHash == actualCurrentHash {
					appFormatter.Info("    %s: %s (component is unchanged)", cyan("Status"), green("In Sync"))
				} else {
					appFormatter.Info("    %s: %s (component has been modified)", cyan("Status"), yellow("Modified"))
				}

				appFormatter.EmptyLine()
			}

			if !foundInAnyTarget {
				if target != "" {
					// Specific target was requested but component not found
					fmt.Println(errors.NewTargetDirectoryNotFoundError(target).Format())
				} else {
					// No target specified and component not found in any target
					// Collect available components from all targets
					var availableComponents []string
					for _, targetName := range []string{"opencode", "claudecode"} {
						targetDir := project.GetTargetDirectory(projectRoot, targetName)
						if _, err := os.Stat(targetDir); os.IsNotExist(err) {
							continue
						}
						metadata, err := project.LoadMaterializationMetadata(targetDir)
						if err != nil {
							continue
						}
						componentMap := metadata.GetComponentMap(componentType)
						if componentMap != nil {
							for compName := range componentMap {
								availableComponents = append(availableComponents, compName)
							}
						}
					}
					fmt.Println(errors.NewComponentNotFoundInProjectError(componentType, componentName, availableComponents).Format())
				}
				os.Exit(1)
			}
		},
		func(target, projectDir string) {
			// Handle materialize status command
			green := color.New(color.FgGreen).SprintFunc()
			yellow := color.New(color.FgYellow).SprintFunc()
			red := color.New(color.FgRed).SprintFunc()
			bold := color.New(color.Bold).SprintFunc()

			var projectRoot string
			var err error
			if projectDir != "" {
				// Use specified project directory
				projectRoot, err = filepath.Abs(projectDir)
				if err != nil {
					log.Fatalf("Failed to resolve project directory: %v", err)
				}
			} else {
				// Auto-detect project root
				projectRoot, err = project.FindProjectRoot()
				if err != nil {
					log.Fatalf("Failed to find project root: %v", err)
				}
			}

			fmt.Printf("\n%s %s\n\n", bold("Project:"), projectRoot)

			// Determine which targets to check
			var targetsToCheck []string
			if target != "" {
				targetsToCheck = []string{target}
			} else {
				targetsToCheck = []string{"opencode", "claudecode"}
			}

			// Track overall statistics
			totalInSync := 0
			totalOutOfSync := 0
			totalMissing := 0
			foundAny := false

			// Check each target
			for _, targetName := range targetsToCheck {
				targetDir := project.GetTargetDirectory(projectRoot, targetName)

				// Check if target directory exists
				if _, err := os.Stat(targetDir); os.IsNotExist(err) {
					if target != "" {
						// User specified a target that doesn't exist
						fmt.Println(errors.NewTargetDirectoryNotFoundError(targetName).Format())
						os.Exit(1)
					}
					continue
				}

				// Load materialization metadata
				metadata, err := project.LoadMaterializationMetadata(targetDir)
				if err != nil {
					log.Fatalf("Failed to load materialization metadata: %v", err)
				}

				// Get all components
				components := metadata.GetAllMaterializedComponents()
				if len(components) == 0 {
					continue
				}

				foundAny = true

				// Display target header
				var targetLabel string
				if targetName == "opencode" {
					targetLabel = "OpenCode (.opencode/)"
				} else {
					targetLabel = "Claude Code (.claude/)"
				}
				fmt.Printf("%s %s\n\n", bold("Target:"), targetLabel)

				// Use batched sync check for better performance (one clone per repo instead of per component)
				baseDir, _ := paths.GetAgentsDir()
				syncResults, err := project.CheckMultipleComponentsSyncStatusBatched(baseDir, components)
				if err != nil {
					log.Fatalf("Failed to check sync status: %v", err)
				}

				// Group by component type
				componentsByType := make(map[string][]project.ComponentInfo)
				for _, comp := range components {
					componentsByType[comp.Type] = append(componentsByType[comp.Type], comp)
				}

				// Display each type
				for _, componentType := range []string{"skills", "agents", "commands"} {
					comps := componentsByType[componentType]
					if len(comps) == 0 {
						continue
					}

					// Display type header
					typeLabel := strings.Title(componentType)
					fmt.Printf("%s:\n", typeLabel)

					// Display sync status for each component
					for _, comp := range comps {
						key := fmt.Sprintf("%s/%s", comp.Type, comp.Name)
						result, ok := syncResults[key]
						if !ok {
							fmt.Printf("  %s %s (error: status not available)\n", red("✗"), comp.Name)
							continue
						}

						if result.Error != nil {
							fmt.Printf("  %s %s (error: %v)\n", red("✗"), comp.Name, result.Error)
							continue
						}

						// Truncate commit hash to first 7 characters for display
						shortHash := comp.Metadata.CommitHash
						if len(shortHash) > 7 {
							shortHash = shortHash[:7]
						}

						switch result.Status {
						case project.SyncStatusInSync:
							fmt.Printf("  %s %s (in sync - %s)\n", green("✓"), comp.Name, shortHash)
							totalInSync++
						case project.SyncStatusOutOfSync:
							// Get current remote commit hash to show the change
							// We need to fetch this separately since CheckMultipleComponentsSyncStatusBatched
							// doesn't return the current SHA (only the status)
							ud := updater.NewUpdateDetectorWithBaseDir(baseDir)
							currentCommit, err := ud.GetCurrentRepoSHA(comp.Metadata.Source)
							shortCurrent := currentCommit
							if err == nil && len(shortCurrent) > 7 {
								shortCurrent = shortCurrent[:7]
							}

							if err == nil && currentCommit != comp.Metadata.CommitHash {
								fmt.Printf("  %s %s (out of sync - %s → %s)\n", yellow("⚠"), comp.Name, shortHash, shortCurrent)
							} else {
								fmt.Printf("  %s %s (out of sync)\n", yellow("⚠"), comp.Name)
							}
							totalOutOfSync++
						case project.SyncStatusSourceMissing:
							fmt.Printf("  %s %s (repository not found)\n", red("✗"), comp.Name)
							totalMissing++
						}
					}
					fmt.Println()
				}
			}

			if !foundAny {
				appFormatter.Info("%s No components materialized yet", yellow(formatter.SymbolWarning))
				appFormatter.EmptyLine()
				appFormatter.Info("To materialize components:")
				appFormatter.Info("  agent-smith materialize skill <name> --target opencode")
				appFormatter.Info("  agent-smith materialize all --target opencode")
				return
			}

			// Display summary
			fmt.Printf("%s: ", bold("Summary"))
			var parts []string
			if totalInSync > 0 {
				parts = append(parts, fmt.Sprintf("%s %d in sync", green("✓"), totalInSync))
			}
			if totalOutOfSync > 0 {
				parts = append(parts, fmt.Sprintf("%s %d out of sync", yellow("⚠"), totalOutOfSync))
			}
			if totalMissing > 0 {
				parts = append(parts, fmt.Sprintf("%s %d source missing", red("✗"), totalMissing))
			}
			fmt.Printf("%s\n\n", strings.Join(parts, ", "))
		},
		func(target, projectDir string, force, dryRun bool) {
			// Handle materialize update command
			green := color.New(color.FgGreen).SprintFunc()
			yellow := color.New(color.FgYellow).SprintFunc()
			red := color.New(color.FgRed).SprintFunc()
			bold := color.New(color.Bold).SprintFunc()

			var projectRoot string
			var err error
			if projectDir != "" {
				// Use specified project directory
				projectRoot, err = filepath.Abs(projectDir)
				if err != nil {
					log.Fatalf("Failed to resolve project directory: %v", err)
				}
			} else {
				// Auto-detect project root
				projectRoot, err = project.FindProjectRoot()
				if err != nil {
					log.Fatalf("Failed to find project root: %v", err)
				}
			}

			if dryRun {
				fmt.Printf("\n%s Previewing updates in: %s\n\n", bold("[DRY RUN]"), projectRoot)
			} else {
				fmt.Printf("\nUpdating materialized components in: %s\n\n", projectRoot)
			}

			// Determine which targets to update
			var targetsToUpdate []string
			if target != "" {
				targetsToUpdate = []string{target}
			} else {
				targetsToUpdate = []string{"opencode", "claudecode"}
			}

			// Track overall statistics
			totalUpdated := 0
			totalSkippedInSync := 0
			totalSkippedMissing := 0
			foundAny := false

			// Process each target
			for _, targetName := range targetsToUpdate {
				targetDir := project.GetTargetDirectory(projectRoot, targetName)

				// Check if target directory exists
				if _, err := os.Stat(targetDir); os.IsNotExist(err) {
					if target != "" {
						// User specified a target that doesn't exist
						fmt.Println(errors.NewTargetDirectoryNotFoundError(targetName).Format())
						os.Exit(1)
					}
					continue
				}

				// Load materialization metadata
				metadata, err := project.LoadMaterializationMetadata(targetDir)
				if err != nil {
					log.Fatalf("Failed to load materialization metadata: %v", err)
				}

				// Get all components
				components := metadata.GetAllMaterializedComponents()
				if len(components) == 0 {
					continue
				}

				foundAny = true

				// Display target header
				var targetLabel string
				if targetName == "opencode" {
					targetLabel = "OpenCode (.opencode/)"
				} else {
					targetLabel = "Claude Code (.claude/)"
				}
				fmt.Printf("%s %s\n\n", bold("Target:"), targetLabel)

				// Use batched sync check for better performance (one clone per repo instead of per component)
				baseDir, _ := paths.GetAgentsDir()
				syncResults, err := project.CheckMultipleComponentsSyncStatusBatched(baseDir, components)
				if err != nil {
					log.Fatalf("Failed to check sync status: %v", err)
				}

				// Process each component
				for _, comp := range components {
					key := fmt.Sprintf("%s/%s", comp.Type, comp.Name)
					result, ok := syncResults[key]

					// Check sync status
					if !ok {
						fmt.Printf("  %s %s (error: status not available)\n", red("✗"), comp.Name)
						continue
					}

					if result.Error != nil {
						fmt.Printf("  %s %s (error checking status: %v)\n", red("✗"), comp.Name, result.Error)
						continue
					}

					// Handle source missing
					if result.Status == project.SyncStatusSourceMissing {
						fmt.Printf("  %s Skipped %s (source no longer installed)\n", yellow("⚠"), comp.Name)
						totalSkippedMissing++
						continue
					}

					// Skip if in sync and not force mode
					if result.Status == project.SyncStatusInSync && !force {
						fmt.Printf("  %s Skipped %s (already in sync)\n", green("⊘"), comp.Name)
						totalSkippedInSync++
						continue
					}

					// Component needs updating
					if dryRun {
						fmt.Printf("  %s Would update %s\n", green("→"), comp.Name)
						totalUpdated++
						continue
					}

					// Download from GitHub to temp directory
					tempDir, err := os.MkdirTemp("", "materialize-update-*")
					if err != nil {
						fmt.Printf("  %s Failed to create temp directory for %s: %v\n", red("✗"), comp.Name, err)
						continue
					}
					defer os.RemoveAll(tempDir)

					// Download component from GitHub using downloader
					var downloadErr error
					switch comp.Type {
					case "skills":
						dl := downloader.NewSkillDownloaderWithTargetDir(tempDir)
						downloadErr = dl.DownloadSkill(comp.Metadata.Source, comp.Name)
					case "agents":
						dl := downloader.NewAgentDownloaderWithTargetDir(tempDir)
						downloadErr = dl.DownloadAgent(comp.Metadata.Source, comp.Name)
					case "commands":
						dl := downloader.NewCommandDownloaderWithTargetDir(tempDir)
						downloadErr = dl.DownloadCommand(comp.Metadata.Source, comp.Name)
					default:
						downloadErr = fmt.Errorf("unknown component type: %s", comp.Type)
					}

					if downloadErr != nil {
						fmt.Printf("  %s Failed to download %s from GitHub: %v\n", red("✗"), comp.Name, downloadErr)
						continue
					}

					// Source is in temp directory
					sourceDir := filepath.Join(tempDir, comp.Type, comp.Name)
					destDir := filepath.Join(targetDir, comp.Type, comp.Name)

					// Remove existing materialized component
					if err := os.RemoveAll(destDir); err != nil {
						fmt.Printf("  %s Failed to remove existing %s: %v\n", red("✗"), comp.Name, err)
						continue
					}

					// Copy from temp to target
					if err := materializer.CopyDirectory(sourceDir, destDir); err != nil {
						fmt.Printf("  %s Failed to copy %s: %v\n", red("✗"), comp.Name, err)
						continue
					}

					// Calculate new hashes
					newSourceHash, err := materializer.CalculateDirectoryHash(sourceDir)
					if err != nil {
						fmt.Printf("  %s Failed to calculate source hash for %s: %v\n", red("✗"), comp.Name, err)
						continue
					}

					newCurrentHash, err := materializer.CalculateDirectoryHash(destDir)
					if err != nil {
						fmt.Printf("  %s Failed to calculate current hash for %s: %v\n", red("✗"), comp.Name, err)
						continue
					}

					// Get the latest commit hash from what we just downloaded
					// Read the lock file entry from temp directory to get the updated commit hash
					lockEntry, err := metadataPkg.LoadLockFileEntry(tempDir, comp.Type, comp.Name)
					var newCommitHash string
					if err == nil && lockEntry != nil {
						newCommitHash = lockEntry.CommitHash
					} else {
						// Fallback: fetch current commit from GitHub
						baseDir, _ := paths.GetAgentsDir()
						ud := updater.NewUpdateDetectorWithBaseDir(baseDir)
						newCommitHash, _ = ud.GetCurrentRepoSHA(comp.Metadata.Source)
					}

					// Update metadata entry with new commit hash
					comp.Metadata.CommitHash = newCommitHash
					comp.Metadata.SourceHash = newSourceHash
					comp.Metadata.CurrentHash = newCurrentHash
					comp.Metadata.MaterializedAt = time.Now().Format(time.RFC3339)

					// Save updated metadata back to the metadata struct
					switch comp.Type {
					case "skills":
						metadata.Skills[comp.Name] = comp.Metadata
					case "agents":
						metadata.Agents[comp.Name] = comp.Metadata
					case "commands":
						metadata.Commands[comp.Name] = comp.Metadata
					}

					fmt.Printf("  %s Updated %s\n", green("✓"), comp.Name)
					totalUpdated++
				}

				// Save metadata
				if !dryRun && totalUpdated > 0 {
					if err := project.SaveMaterializationMetadata(targetDir, metadata); err != nil {
						log.Fatalf("Failed to save materialization metadata: %v", err)
					}
				}

				fmt.Println()
			}

			if !foundAny {
				appFormatter.Info("%s No components materialized yet", yellow(formatter.SymbolWarning))
				appFormatter.EmptyLine()
				appFormatter.Info("To materialize components:")
				appFormatter.Info("  agent-smith materialize skill <name> --target opencode")
				appFormatter.Info("  agent-smith materialize all --target opencode")
				return
			}

			// Display summary
			fmt.Printf("%s: ", bold("Summary"))
			var parts []string
			if totalUpdated > 0 {
				if dryRun {
					parts = append(parts, fmt.Sprintf("%d would be updated", totalUpdated))
				} else {
					parts = append(parts, fmt.Sprintf("%s %d updated", green("✓"), totalUpdated))
				}
			}
			if totalSkippedInSync > 0 {
				parts = append(parts, fmt.Sprintf("%d already in sync", totalSkippedInSync))
			}
			if totalSkippedMissing > 0 {
				parts = append(parts, fmt.Sprintf("%s %d skipped (source missing)", yellow("⚠"), totalSkippedMissing))
			}
			fmt.Printf("%s\n\n", strings.Join(parts, ", "))
		},
	)

	// Execute Cobra command
	cmd.Execute()
}
