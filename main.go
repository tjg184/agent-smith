package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
	"github.com/tgaines/agent-smith/internal/uninstaller"
	"github.com/tgaines/agent-smith/internal/updater"
	"github.com/tgaines/agent-smith/pkg/config"
	"github.com/tgaines/agent-smith/pkg/errors"
	"github.com/tgaines/agent-smith/pkg/logger"
	"github.com/tgaines/agent-smith/pkg/paths"
	"github.com/tgaines/agent-smith/pkg/profiles"
	"github.com/tgaines/agent-smith/pkg/project"
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

	// Set up handlers for Cobra commands
	cmd.SetHandlers(
		func(repoURL, name, profile, targetDir string) {
			debugPrintf("[DEBUG] handleAddSkill called with repoURL=%s, name=%s, profile=%s, targetDir=%s\n", repoURL, name, profile, targetDir)

			if profile != "" && targetDir != "" {
				appLogger.FatalMsg(errors.NewInvalidFlagsError("--profile", "--target-dir"))
			}

			if targetDir != "" {
				// Install to custom target directory (isolated testing)
				debugPrintln("[DEBUG] Installing to custom target directory")
				resolvedPath, err := paths.ResolveTargetDir(targetDir)
				if err != nil {
					log.Fatal("Failed to resolve target directory:", err)
				}
				debugPrintf("[DEBUG] Resolved target directory: %s\n", resolvedPath)

				if err := fileutil.CreateDirectoryWithPermissions(resolvedPath); err != nil {
					log.Fatal("Failed to create target directory:", err)
				}

				infoPrintf("Installing to custom directory: %s\n", resolvedPath)
				dl := downloader.NewSkillDownloaderWithTargetDir(resolvedPath)
				if err := dl.DownloadSkill(repoURL, name); err != nil {
					appLogger.FatalMsg(errors.NewComponentDownloadError("skill", repoURL, err))
				}
				debugPrintln("[DEBUG] Skill download completed successfully")
			} else if profile != "" {
				// Install directly to profile
				debugPrintf("[DEBUG] Installing to profile: %s\n", profile)
				pm, err := profiles.NewProfileManager(nil)
				if err != nil {
					appLogger.FatalMsg(errors.NewProfileManagerError(err))
				}

				// Validate profile exists by scanning
				profilesList, err := pm.ScanProfiles()
				if err != nil {
					log.Fatal("Failed to scan profiles:", err)
				}
				debugPrintf("[DEBUG] Found %d profiles\n", len(profilesList))

				profileExists := false
				for _, p := range profilesList {
					if p.Name == profile {
						profileExists = true
						break
					}
				}

				if !profileExists {
					appLogger.FatalMsg(errors.NewProfileNotFoundError(profile))
				}

				dl := downloader.NewSkillDownloaderForProfile(profile)
				if err := dl.DownloadSkill(repoURL, name); err != nil {
					appLogger.FatalMsg(errors.NewComponentDownloadError("skill", repoURL, err))
				}
				debugPrintln("[DEBUG] Skill download to profile completed successfully")

				// Auto-activate profile if no profile is currently active
				activeProfile, err := pm.GetActiveProfile()
				if err != nil {
					log.Fatal("Failed to get active profile:", err)
				}
				if activeProfile == "" {
					debugPrintf("[DEBUG] No active profile detected, auto-activating profile: %s\n", profile)
					if err := pm.ActivateProfile(profile); err != nil {
						log.Fatal("Failed to auto-activate profile:", err)
					}
					infoPrintf("Profile '%s' has been automatically activated as your first profile.\n", profile)
					infoPrintln("Components from this profile are now ready to be linked.")
				}
			} else {
				// Standard installation to ~/.agent-smith/
				debugPrintln("[DEBUG] Installing to standard directory (~/.agent-smith/)")
				dl := downloader.NewSkillDownloader()
				if err := dl.DownloadSkill(repoURL, name); err != nil {
					appLogger.FatalMsg(errors.NewComponentDownloadError("skill", repoURL, err))
				}
				debugPrintln("[DEBUG] Skill download completed successfully")
			}
		},
		func(repoURL, name, profile, targetDir string) {
			if profile != "" && targetDir != "" {
				appLogger.FatalMsg(errors.NewInvalidFlagsError("--profile", "--target-dir"))
			}

			if targetDir != "" {
				// Install to custom target directory (isolated testing)
				resolvedPath, err := paths.ResolveTargetDir(targetDir)
				if err != nil {
					log.Fatal("Failed to resolve target directory:", err)
				}

				if err := fileutil.CreateDirectoryWithPermissions(resolvedPath); err != nil {
					log.Fatal("Failed to create target directory:", err)
				}

				infoPrintf("Installing to custom directory: %s\n", resolvedPath)
				dl := downloader.NewAgentDownloaderWithTargetDir(resolvedPath)
				if err := dl.DownloadAgent(repoURL, name); err != nil {
					log.Fatal("Failed to download agent:", err)
				}
			} else if profile != "" {
				// Install directly to profile
				pm, err := profiles.NewProfileManager(nil)
				if err != nil {
					log.Fatal("Failed to create profile manager:", err)
				}

				// Validate profile exists by scanning
				profilesList, err := pm.ScanProfiles()
				if err != nil {
					log.Fatal("Failed to scan profiles:", err)
				}

				profileExists := false
				for _, p := range profilesList {
					if p.Name == profile {
						profileExists = true
						break
					}
				}

				if !profileExists {
					log.Fatalf("Profile '%s' does not exist. Create it first with: agent-smith profiles create %s", profile, profile)
				}

				dl := downloader.NewAgentDownloaderForProfile(profile)
				if err := dl.DownloadAgent(repoURL, name); err != nil {
					log.Fatal("Failed to download agent:", err)
				}

				// Auto-activate profile if no profile is currently active
				activeProfile, err := pm.GetActiveProfile()
				if err != nil {
					log.Fatal("Failed to get active profile:", err)
				}
				if activeProfile == "" {
					debugPrintf("[DEBUG] No active profile detected, auto-activating profile: %s\n", profile)
					if err := pm.ActivateProfile(profile); err != nil {
						log.Fatal("Failed to auto-activate profile:", err)
					}
					infoPrintf("Profile '%s' has been automatically activated as your first profile.\n", profile)
					infoPrintln("Components from this profile are now ready to be linked.")
				}
			} else {
				// Standard installation to ~/.agent-smith/
				dl := downloader.NewAgentDownloader()
				if err := dl.DownloadAgent(repoURL, name); err != nil {
					log.Fatal("Failed to download agent:", err)
				}
			}
		},
		func(repoURL, name, profile, targetDir string) {
			if profile != "" && targetDir != "" {
				log.Fatal("Cannot specify both --profile and --target-dir flags")
			}

			if targetDir != "" {
				// Install to custom target directory (isolated testing)
				resolvedPath, err := paths.ResolveTargetDir(targetDir)
				if err != nil {
					log.Fatal("Failed to resolve target directory:", err)
				}

				if err := fileutil.CreateDirectoryWithPermissions(resolvedPath); err != nil {
					log.Fatal("Failed to create target directory:", err)
				}

				infoPrintf("Installing to custom directory: %s\n", resolvedPath)
				dl := downloader.NewCommandDownloaderWithTargetDir(resolvedPath)
				if err := dl.DownloadCommand(repoURL, name); err != nil {
					log.Fatal("Failed to download command:", err)
				}
			} else if profile != "" {
				// Install directly to profile
				pm, err := profiles.NewProfileManager(nil)
				if err != nil {
					log.Fatal("Failed to create profile manager:", err)
				}

				// Validate profile exists by scanning
				profilesList, err := pm.ScanProfiles()
				if err != nil {
					log.Fatal("Failed to scan profiles:", err)
				}

				profileExists := false
				for _, p := range profilesList {
					if p.Name == profile {
						profileExists = true
						break
					}
				}

				if !profileExists {
					log.Fatalf("Profile '%s' does not exist. Create it first with: agent-smith profiles create %s", profile, profile)
				}

				dl := downloader.NewCommandDownloaderForProfile(profile)
				if err := dl.DownloadCommand(repoURL, name); err != nil {
					log.Fatal("Failed to download command:", err)
				}

				// Auto-activate profile if no profile is currently active
				activeProfile, err := pm.GetActiveProfile()
				if err != nil {
					log.Fatal("Failed to get active profile:", err)
				}
				if activeProfile == "" {
					debugPrintf("[DEBUG] No active profile detected, auto-activating profile: %s\n", profile)
					if err := pm.ActivateProfile(profile); err != nil {
						log.Fatal("Failed to auto-activate profile:", err)
					}
					infoPrintf("Profile '%s' has been automatically activated as your first profile.\n", profile)
					infoPrintln("Components from this profile are now ready to be linked.")
				}
			} else {
				// Standard installation to ~/.agent-smith/
				dl := downloader.NewCommandDownloader()
				if err := dl.DownloadCommand(repoURL, name); err != nil {
					log.Fatal("Failed to download command:", err)
				}
			}
		},
		func(repoURL, profile, targetDir string) {
			var bulkDownloader *downloader.BulkDownloader

			if profile != "" && targetDir != "" {
				log.Fatal("Cannot specify both --profile and --target-dir flags")
			}

			if targetDir != "" {
				// Resolve the target directory path
				resolvedPath, err := paths.ResolveTargetDir(targetDir)
				if err != nil {
					log.Fatal("Failed to resolve target directory:", err)
				}

				// Create the target directory if it doesn't exist
				if err := fileutil.CreateDirectoryWithPermissions(resolvedPath); err != nil {
					log.Fatal("Failed to create target directory:", err)
				}

				infoPrintf("Installing to custom directory: %s\n", resolvedPath)
				bulkDownloader = downloader.NewBulkDownloaderWithTargetDir(resolvedPath)

				if err := bulkDownloader.AddAll(repoURL); err != nil {
					log.Fatal("Failed to bulk download components:", err)
				}
			} else {
				// Automatically create a profile from the repository URL
				pm, err := profiles.NewProfileManager(nil)
				if err != nil {
					log.Fatal("Failed to create profile manager:", err)
				}

				var profileName string

				if profile != "" {
					// Custom profile name provided via --profile flag
					// Check if profile with this name already exists
					profilesList, err := pm.ScanProfiles()
					if err != nil {
						log.Fatal("Failed to scan profiles:", err)
					}

					profileExists := false
					for _, p := range profilesList {
						if p.Name == profile {
							profileExists = true
							break
						}
					}

					if profileExists {
						log.Fatalf("Profile '%s' already exists. Please choose a different name or remove the --profile flag to update the existing profile.", profile)
					}

					profileName = profile
					infoPrintf("Creating profile: %s\n", profileName)

					// Create the profile with metadata
					if err := pm.CreateProfileWithMetadata(profileName, repoURL); err != nil {
						log.Fatal("Failed to create profile:", err)
					}
				} else {
					// No custom profile name - use auto-detection and reuse logic
					// Check if a profile already exists for this repository
					existingProfileName, err := pm.FindProfileBySourceURL(repoURL)
					if err != nil {
						log.Fatal("Failed to search for existing profile:", err)
					}

					if existingProfileName != "" {
						// Profile already exists, reuse it
						profileName = existingProfileName
						infoPrintf("Found existing profile for repository: %s\n", profileName)
						infoPrintf("Updating profile with latest components...\n")
					} else {
						// Get existing profiles for name generation
						existingProfiles, err := pm.ScanProfiles()
						if err != nil {
							log.Fatal("Failed to scan profiles:", err)
						}

						existingProfileNames := make([]string, len(existingProfiles))
						for i, p := range existingProfiles {
							existingProfileNames[i] = p.Name
						}

						// Generate a unique profile name
						profileName = profiles.GenerateProfileNameFromRepo(repoURL, existingProfileNames)
						infoPrintf("Creating profile: %s\n", profileName)

						// Create the profile with metadata
						if err := pm.CreateProfileWithMetadata(profileName, repoURL); err != nil {
							log.Fatal("Failed to create profile:", err)
						}
					}
				}

				// Install components to the profile
				infoPrintf("Installing components to profile: %s\n", profileName)
				bulkDownloader = downloader.NewBulkDownloaderForProfile(profileName)

				if err := bulkDownloader.AddAll(repoURL); err != nil {
					log.Fatal("Failed to bulk download components:", err)
				}

				// Auto-activate profile if no profile is currently active
				activeProfile, err := pm.GetActiveProfile()
				if err != nil {
					log.Fatal("Failed to get active profile:", err)
				}
				if activeProfile == "" {
					debugPrintf("[DEBUG] No active profile detected, auto-activating profile: %s\n", profileName)
					if err := pm.ActivateProfile(profileName); err != nil {
						log.Fatal("Failed to auto-activate profile:", err)
					}
					infoPrintf("Profile '%s' has been automatically activated as your first profile.\n", profileName)
					infoPrintln("Components from this profile are now ready to be linked.")
				} else if activeProfile != profileName {
					// Only show activation message if this is not the active profile
					appFormatter.EmptyLine()
					appFormatter.Info("Profile updated successfully!")
					appFormatter.Info("To activate this profile and use these components, run:")
					appFormatter.Info("  agent-smith profile activate %s", profileName)
					appFormatter.Info("  agent-smith link all")
				} else {
					// Profile is already active, components are ready
					infoPrintln("\nProfile updated successfully! Components are ready to be linked.")
				}
			}
		},
		func(componentType, componentName, profile string) {
			detector := NewUpdateDetectorWithProfile(profile)

			// Load metadata to get source URL
			metadata, err := detector.LoadMetadata(componentType, componentName)
			if err != nil {
				log.Fatal("Failed to load component metadata:", err)
			}

			if err := detector.UpdateComponent(componentType, componentName, metadata.Source); err != nil {
				log.Fatal("Failed to update component:", err)
			}
		},
		func(profile string) {
			detector := NewUpdateDetectorWithProfile(profile)
			if err := detector.UpdateAll(); err != nil {
				log.Fatal("Failed to update components:", err)
			}
		},
		func(componentType, componentName, targetFilter, profile string) {
			debugPrintf("[DEBUG] handleLink called with componentType=%s, componentName=%s, targetFilter=%s, profile=%s\n", componentType, componentName, targetFilter, profile)
			linker, err := NewComponentLinkerWithFilterAndProfile(targetFilter, profile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to create component linker\n\n%v\n", err)
				os.Exit(1)
			}
			debugPrintln("[DEBUG] Component linker created successfully")
			if err := linker.LinkComponent(componentType, componentName); err != nil {
				log.Fatal("Failed to link component:", err)
			}
			debugPrintln("[DEBUG] Component linked successfully")
		},
		func(targetFilter, profile string, allProfiles bool) {
			debugPrintf("[DEBUG] handleLinkAll called with targetFilter=%s, profile=%s, allProfiles=%v\n", targetFilter, profile, allProfiles)

			// Validate flag combination
			if allProfiles && profile != "" {
				log.Fatal("Cannot use both --all-profiles and --profile flags together")
			}

			if allProfiles {
				// Link from all profiles
				debugPrintln("[DEBUG] Linking from all profiles")

				// Create profile manager
				pm, err := profiles.NewProfileManager(nil)
				if err != nil {
					appFormatter.EmptyLine()
					appFormatter.ErrorMsg("Failed to initialize profile manager")
					appFormatter.DetailItem("Error", err.Error())
					appFormatter.EmptyLine()
					appFormatter.InfoMsg("The --all-profiles flag requires a working profile system.")
					appFormatter.InfoMsg("Please check your ~/.agent-smith/ directory permissions.")
					os.Exit(1)
				}

				// Get all profiles
				allProfilesList, err := pm.ScanProfiles()
				if err != nil {
					appFormatter.EmptyLine()
					appFormatter.ErrorMsg("Failed to scan profiles")
					appFormatter.DetailItem("Error", err.Error())
					appFormatter.EmptyLine()
					appFormatter.InfoMsg("The --all-profiles flag requires at least one profile.")
					appFormatter.InfoMsg("Try running without --all-profiles, or create a profile first:")
					appFormatter.InfoMsg("  agent-smith profile create <name>")
					os.Exit(1)
				}

				if len(allProfilesList) == 0 {
					appFormatter.EmptyLine()
					appFormatter.InfoMsg("No profiles found")
					appFormatter.EmptyLine()
					appFormatter.InfoMsg("The --all-profiles flag requires at least one profile.")
					appFormatter.InfoMsg("Options:")
					appFormatter.InfoMsg("  1. Run without --all-profiles to link components from base installation")
					appFormatter.InfoMsg("  2. Create a profile first: agent-smith profile create <name>")
					os.Exit(1)
				}

				// Color helpers
				bold := color.New(color.Bold).SprintFunc()
				green := color.New(color.FgGreen, color.Bold).SprintFunc()
				cyan := color.New(color.FgCyan).SprintFunc()
				gray := color.New(color.FgHiBlack).SprintFunc()

				// Link from each profile
				fmt.Printf("\n%s\n", bold("Linking components from all profiles..."))
				fmt.Println()

				for _, profileItem := range allProfilesList {
					fmt.Printf("%s\n", cyan(fmt.Sprintf("Profile: %s", profileItem.Name)))

					linker, err := NewComponentLinkerWithFilterAndProfile(targetFilter, profileItem.Name)
					if err != nil {
						log.Fatalf("Failed to create component linker for profile '%s': %v", profileItem.Name, err)
					}

					// Count components before linking to check if profile has any
					agents, skills, commands := pm.CountComponents(profileItem)
					totalComponents := agents + skills + commands

					if totalComponents == 0 {
						fmt.Printf("  %s\n\n", gray("(no components)"))
						continue
					}

					if err := linker.LinkAllComponents(); err != nil {
						log.Fatalf("Failed to link components from profile '%s': %v", profileItem.Name, err)
					}
				}

				fmt.Printf("\n%s\n", green("✓ Successfully linked components from all profiles"))
			} else {
				// Link from single profile (existing behavior)
				linker, err := NewComponentLinkerWithFilterAndProfile(targetFilter, profile)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: Failed to create component linker\n\n%v\n", err)
					os.Exit(1)
				}
				debugPrintln("[DEBUG] Component linker created successfully")
				if err := linker.LinkAllComponents(); err != nil {
					log.Fatal("Failed to link all components:", err)
				}
				debugPrintln("[DEBUG] All components linked successfully")
			}
		},
		func(componentType, targetFilter, profile string) {
			debugPrintf("[DEBUG] handleLinkType called with componentType=%s, targetFilter=%s, profile=%s\n", componentType, targetFilter, profile)
			linker, err := NewComponentLinkerWithFilterAndProfile(targetFilter, profile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to create component linker\n\n%v\n", err)
				os.Exit(1)
			}
			debugPrintln("[DEBUG] Component linker created successfully")
			if err := linker.LinkComponentsByType(componentType); err != nil {
				log.Fatal("Failed to link components:", err)
			}
			debugPrintf("[DEBUG] Components of type %s linked successfully\n", componentType)
		},
		func() {
			linker, err := NewComponentLinker()
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}
			if err := linker.DetectAndLinkLocalRepositories(); err != nil {
				log.Fatal("Failed to auto-link repositories:", err)
			}
		},
		func() {
			linker, err := NewComponentLinker()
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}
			if err := linker.ListLinkedComponents(); err != nil {
				log.Fatal("Failed to list linked components:", err)
			}
		},
		func(allProfiles bool, profileFilter []string) {
			// Validate flags
			if len(profileFilter) > 0 && !allProfiles {
				log.Fatal("--profile flag requires --all-profiles")
			}

			if allProfiles {
				// Create linker with ProfileManager for multi-profile view
				pm, err := profiles.NewProfileManager(nil)
				if err != nil {
					appFormatter.EmptyLine()
					appFormatter.ErrorMsg("Failed to initialize profile manager")
					appFormatter.DetailItem("Error", err.Error())
					appFormatter.EmptyLine()
					appFormatter.InfoMsg("The --all-profiles flag requires a working profile system.")
					appFormatter.InfoMsg("Please check your ~/.agent-smith/ directory permissions.")
					os.Exit(1)
				}

				// Check if any profiles exist
				profilesList, err := pm.ScanProfiles()
				if err != nil {
					appFormatter.EmptyLine()
					appFormatter.ErrorMsg("Failed to scan profiles")
					appFormatter.DetailItem("Error", err.Error())
					appFormatter.EmptyLine()
					appFormatter.InfoMsg("The --all-profiles flag requires at least one profile.")
					appFormatter.InfoMsg("Try running without --all-profiles, or create a profile first:")
					appFormatter.InfoMsg("  agent-smith profile create <name>")
					os.Exit(1)
				}

				if len(profilesList) == 0 {
					appFormatter.EmptyLine()
					appFormatter.InfoMsg("No profiles found")
					appFormatter.EmptyLine()
					appFormatter.InfoMsg("The --all-profiles flag requires at least one profile.")
					appFormatter.InfoMsg("Options:")
					appFormatter.InfoMsg("  1. Run without --all-profiles to show components from base installation")
					appFormatter.InfoMsg("  2. Create a profile first: agent-smith profile create <name>")
					os.Exit(1)
				}

				linker, err := NewComponentLinkerWithProfileManager(pm)
				if err != nil {
					log.Fatal("Failed to create component linker:", err)
				}

				if err := linker.ShowAllProfilesLinkStatus(profileFilter); err != nil {
					log.Fatal("Failed to show link status:", err)
				}
			} else {
				// Standard single-profile view (backward compatibility)
				linker, err := NewComponentLinker()
				if err != nil {
					log.Fatal("Failed to create component linker:", err)
				}
				if err := linker.ShowLinkStatus(); err != nil {
					log.Fatal("Failed to show link status:", err)
				}
			}
		},
		func(componentType, componentName, targetFilter string) {
			linker, err := NewComponentLinkerWithFilter(targetFilter)
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}
			if err := linker.UnlinkComponent(componentType, componentName, targetFilter); err != nil {
				log.Fatal("Failed to unlink component:", err)
			}
		},
		func(componentType, componentName, targetFilter, profile string) {
			linker, err := NewComponentLinkerWithFilterAndProfile(targetFilter, profile)
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}
			if err := linker.UnlinkComponent(componentType, componentName, targetFilter); err != nil {
				log.Fatal("Failed to unlink component:", err)
			}
		},
		func(targetFilter string, force bool, allProfiles bool) {
			linker, err := NewComponentLinkerWithFilter(targetFilter)
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}
			if err := linker.UnlinkAllComponents(targetFilter, force, allProfiles); err != nil {
				log.Fatal("Failed to unlink all components:", err)
			}
		},
		func(targetFilter string, force bool, allProfiles bool, profile string) {
			// Validate flag combination
			if allProfiles && profile != "" {
				log.Fatal("Cannot use both --all-profiles and --profile flags together")
			}

			linker, err := NewComponentLinkerWithFilterAndProfile(targetFilter, profile)
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}
			if err := linker.UnlinkAllComponents(targetFilter, force, allProfiles); err != nil {
				log.Fatal("Failed to unlink all components:", err)
			}
		},
		func(componentType, targetFilter string, force bool) {
			linker, err := NewComponentLinkerWithFilter(targetFilter)
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}
			if err := linker.UnlinkComponentsByType(componentType, targetFilter, force); err != nil {
				log.Fatal("Failed to unlink components:", err)
			}
		},
		func(componentType, targetFilter string, force bool, profile string) {
			linker, err := NewComponentLinkerWithFilterAndProfile(targetFilter, profile)
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}
			if err := linker.UnlinkComponentsByType(componentType, targetFilter, force); err != nil {
				log.Fatal("Failed to unlink components:", err)
			}
		},
		func(componentType, componentName, profile string) {
			// Determine base directory
			baseDir, err := paths.GetAgentsDir()
			if err != nil {
				log.Fatal("Failed to get agents directory:", err)
			}

			if profile != "" {
				// Use profile directory
				profilesDir, err := paths.GetProfilesDir()
				if err != nil {
					log.Fatal("Failed to get profiles directory:", err)
				}
				baseDir = filepath.Join(profilesDir, profile)

				// Validate profile exists
				if _, err := os.Stat(baseDir); os.IsNotExist(err) {
					log.Fatalf("Profile '%s' does not exist", profile)
				}
			}

			// Create linker for unlinking
			linker, err := NewComponentLinker()
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}

			// Create uninstaller
			uninstaller := uninstaller.NewUninstaller(baseDir, linker)

			// Uninstall component
			if err := uninstaller.UninstallComponent(componentType, componentName); err != nil {
				log.Fatal("Failed to uninstall component:", err)
			}
		},
		func(repoURL string, force bool) {
			// Get base directory (always ~/.agent-smith/ for bulk uninstall)
			baseDir, err := paths.GetAgentsDir()
			if err != nil {
				log.Fatal("Failed to get base directory:", err)
			}

			// Create linker for unlinking
			linker, err := NewComponentLinker()
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}

			// Create uninstaller
			uninstaller := uninstaller.NewUninstaller(baseDir, linker)

			// Uninstall all components from source
			if err := uninstaller.UninstallAllFromSource(repoURL, force); err != nil {
				log.Fatal("Failed to uninstall components:", err)
			}
		},
		func(profileFilter []string, activeOnly bool, typeFilter string) {
			pm, err := profiles.NewProfileManager(nil)
			if err != nil {
				log.Fatal("Failed to create profile manager:", err)
			}

			// Validate typeFilter if provided
			if typeFilter != "" && typeFilter != "repo" && typeFilter != "user" {
				log.Fatalf("Invalid type filter '%s'. Valid values are: repo, user", typeFilter)
			}

			profilesList, err := pm.ScanProfiles()
			if err != nil {
				log.Fatal("Failed to scan profiles:", err)
			}

			// Get active profile
			activeProfile, err := pm.GetActiveProfile()
			if err != nil {
				log.Fatal("Failed to get active profile:", err)
			}

			// Apply filters
			var filteredProfiles []*profiles.Profile

			// Filter by active-only if specified
			if activeOnly {
				for _, profile := range profilesList {
					if profile.Name == activeProfile {
						filteredProfiles = append(filteredProfiles, profile)
						break
					}
				}
			} else if len(profileFilter) > 0 {
				// Filter by specific profile names
				filterMap := make(map[string]bool)
				for _, name := range profileFilter {
					filterMap[name] = true
				}

				// Validate that all filter names exist
				profileMap := make(map[string]bool)
				for _, p := range profilesList {
					profileMap[p.Name] = true
				}

				for _, filterName := range profileFilter {
					if !profileMap[filterName] {
						log.Fatalf("Profile '%s' does not exist", filterName)
					}
				}

				// Apply filter
				for _, p := range profilesList {
					if filterMap[p.Name] {
						filteredProfiles = append(filteredProfiles, p)
					}
				}
			} else {
				// No filters, show all profiles
				filteredProfiles = profilesList
			}

			// Apply type filter if specified
			if typeFilter != "" {
				var typeFilteredProfiles []*profiles.Profile
				for _, profile := range filteredProfiles {
					profileType, err := pm.GetProfileType(profile.Name)
					if err != nil {
						// Log warning but continue
						continue
					}
					if profileType == typeFilter {
						typeFilteredProfiles = append(typeFilteredProfiles, profile)
					}
				}
				filteredProfiles = typeFilteredProfiles
			}

			// Display results
			if len(filteredProfiles) == 0 {
				if activeOnly {
					appFormatter.Info("No active profile set")
				} else if len(profileFilter) > 0 {
					appFormatter.Info("No matching profiles found")
				} else {
					appFormatter.Info("No profiles found in ~/.agent-smith/profiles/")
					appFormatter.EmptyLine()
					appFormatter.Info("To create a profile, run:")
					appFormatter.Info("  agent-smith profile create <profile-name>")
				}
				return
			}

			// Create table with box-drawing characters
			table := formatter.NewBoxTable(os.Stdout, []string{"Profile", "Components"})

			// Add rows to table
			for _, profile := range filteredProfiles {
				// Get profile type and metadata
				profileType, err := pm.GetProfileType(profile.Name)
				if err != nil {
					profileType = "unknown"
				}

				// Get metadata for repo profiles
				var sourceURL string
				if profileType == "repo" {
					metadata, err := pm.LoadProfileMetadata(profile.Name)
					if err == nil && metadata != nil {
						sourceURL = metadata.SourceURL
					}
				}

				// Count components
				agents, skills, commands := pm.CountComponents(profile)

				// Build component counts string with proper singular/plural handling
				var components []string
				if agents > 0 {
					if agents == 1 {
						components = append(components, "1 agent")
					} else {
						components = append(components, fmt.Sprintf("%d agents", agents))
					}
				}
				if skills > 0 {
					if skills == 1 {
						components = append(components, "1 skill")
					} else {
						components = append(components, fmt.Sprintf("%d skills", skills))
					}
				}
				if commands > 0 {
					if commands == 1 {
						components = append(components, "1 command")
					} else {
						components = append(components, fmt.Sprintf("%d commands", commands))
					}
				}

				componentStr := ""
				if len(components) > 0 {
					componentStr = fmt.Sprintf("(%s)", joinStrings(components, ", "))
				} else {
					componentStr = "(empty)"
				}

				// Build profile cell with active indicator and type emoji
				activeIndicator := " "
				if profile.Name == activeProfile {
					activeIndicator = formatter.ColoredSuccess()
				}

				// Add type emoji
				var typeEmoji string
				switch profileType {
				case "repo":
					typeEmoji = "📦"
				case "user":
					typeEmoji = "👤"
				default:
					typeEmoji = "❓"
				}

				// Build profile name with source URL for repo types
				profileName := profile.Name
				if profileType == "repo" && sourceURL != "" {
					profileName = fmt.Sprintf("%s (%s)", profile.Name, sourceURL)
				}

				profileCell := fmt.Sprintf("%s %s %s", activeIndicator, typeEmoji, profileName)

				// Add row to table
				table.AddRow([]string{profileCell, componentStr})
			}

			// Render the table
			table.Render()

			// Display legend
			appFormatter.EmptyLine()
			appFormatter.Info("Legend:")
			appFormatter.Info("  %s - Currently active profile", formatter.ColoredSuccess())
			appFormatter.Info("  📦 - Repository-sourced profile")
			appFormatter.Info("  👤 - User-created profile")

			// Display total count
			if len(profileFilter) > 0 || activeOnly || typeFilter != "" {
				appFormatter.Info("\nShowing: %d profile(s) (filtered from %d total)", len(filteredProfiles), len(profilesList))
			} else {
				appFormatter.Info("\nTotal: %d profile(s)", len(filteredProfiles))
			}
		},
		func(profileName string) {
			pm, err := profiles.NewProfileManager(nil)
			if err != nil {
				log.Fatal("Failed to create profile manager:", err)
			}

			// Load the profile
			profilesList, err := pm.ScanProfiles()
			if err != nil {
				log.Fatal("Failed to scan profiles:", err)
			}

			var targetProfile *profiles.Profile
			for _, p := range profilesList {
				if p.Name == profileName {
					targetProfile = p
					break
				}
			}

			if targetProfile == nil {
				log.Fatalf("Profile '%s' not found", profileName)
			}

			// Get active profile to show status
			activeProfile, err := pm.GetActiveProfile()
			if err != nil {
				log.Fatal("Failed to get active profile:", err)
			}

			// Display profile information
			appFormatter.Info("Profile: %s", targetProfile.Name)
			if targetProfile.Name == activeProfile {
				appFormatter.Info(" %s [active]", formatter.SymbolSuccess)
			}
			appFormatter.EmptyLine()
			appFormatter.Info("Location: %s", targetProfile.BasePath)
			appFormatter.EmptyLine()

			// Get component names
			agents, skills, commands := pm.GetComponentNames(targetProfile)

			// Display agents
			if len(agents) > 0 {
				appFormatter.Info("Agents (%d):", len(agents))
				for _, agent := range agents {
					sourceURL := pm.GetComponentSource(targetProfile, "agents", agent)
					if sourceURL != "" {
						appFormatter.Info("  - %s (%s)", agent, sourceURL)
					} else {
						appFormatter.Info("  - %s", agent)
					}
				}
				appFormatter.EmptyLine()
			}

			// Display skills
			if len(skills) > 0 {
				appFormatter.Info("Skills (%d):", len(skills))
				for _, skill := range skills {
					sourceURL := pm.GetComponentSource(targetProfile, "skills", skill)
					if sourceURL != "" {
						appFormatter.Info("  - %s (%s)", skill, sourceURL)
					} else {
						appFormatter.Info("  - %s", skill)
					}
				}
				appFormatter.EmptyLine()
			}

			// Display commands
			if len(commands) > 0 {
				appFormatter.Info("Commands (%d):", len(commands))
				for _, command := range commands {
					sourceURL := pm.GetComponentSource(targetProfile, "commands", command)
					if sourceURL != "" {
						appFormatter.Info("  - %s (%s)", command, sourceURL)
					} else {
						appFormatter.Info("  - %s", command)
					}
				}
				appFormatter.EmptyLine()
			}

			// Show empty state if no components
			if len(agents) == 0 && len(skills) == 0 && len(commands) == 0 {
				appFormatter.Info("This profile is empty.")
				appFormatter.EmptyLine()
				appFormatter.Info("Add components with:")
				appFormatter.Info("  agent-smith profiles add <type> %s <component-name>", profileName)
			} else if targetProfile.Name != activeProfile {
				// Show activation hint if not active
				appFormatter.Info("To activate this profile:")
				appFormatter.Info("  agent-smith profiles activate %s", profileName)
			}
		},
		func(profileName string) {
			pm, err := profiles.NewProfileManager(nil)
			if err != nil {
				log.Fatal("Failed to create profile manager:", err)
			}

			// Check if there's an active profile before creating
			activeProfile, err := pm.GetActiveProfile()
			if err != nil {
				log.Fatal("Failed to get active profile:", err)
			}

			if err := pm.CreateProfile(profileName); err != nil {
				log.Fatal("Failed to create profile:", err)
			}

			// Auto-activate profile if no profile was previously active
			if activeProfile == "" {
				debugPrintf("[DEBUG] No active profile detected, auto-activating profile: %s\n", profileName)
				if err := pm.ActivateProfile(profileName); err != nil {
					log.Fatal("Failed to auto-activate profile:", err)
				}
				fmt.Println()
				infoPrintf("✓ Profile '%s' has been automatically activated as your first profile.\n", profileName)
				infoPrintln("  You can now add components and link them with: agent-smith link all")
			}
		},
		func(profileName string) {
			// Create component linker for defensive unlinking
			componentLinker, err := NewComponentLinker()
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}

			pm, err := profiles.NewProfileManager(componentLinker)
			if err != nil {
				log.Fatal("Failed to create profile manager:", err)
			}

			if err := pm.DeleteProfile(profileName); err != nil {
				log.Fatal("Failed to delete profile:", err)
			}
		},
		func(profileName string) {
			pm, err := profiles.NewProfileManager(nil)
			if err != nil {
				log.Fatal("Failed to create profile manager:", err)
			}

			if err := pm.ActivateProfile(profileName); err != nil {
				log.Fatal("Failed to activate profile:", err)
			}
		},
		func() {
			pm, err := profiles.NewProfileManager(nil)
			if err != nil {
				log.Fatal("Failed to create profile manager:", err)
			}

			if err := pm.DeactivateProfile(); err != nil {
				log.Fatal("Failed to deactivate profile:", err)
			}
		},

		func(componentType, profileName, componentName string) {
			pm, err := profiles.NewProfileManager(nil)
			if err != nil {
				log.Fatal("Failed to create profile manager:", err)
			}

			if err := pm.AddComponentToProfile(profileName, componentType, componentName); err != nil {
				log.Fatal("Failed to add component to profile:", err)
			}
		},
		func(componentType, sourceProfile, targetProfile, componentName string) {
			pm, err := profiles.NewProfileManager(nil)
			if err != nil {
				log.Fatal("Failed to create profile manager:", err)
			}

			if err := pm.CopyComponentBetweenProfiles(sourceProfile, targetProfile, componentType, componentName); err != nil {
				log.Fatal("Failed to copy component between profiles:", err)
			}
		},
		func(componentType, profileName, componentName string) {
			// Create component linker to handle auto-unlinking
			componentLinker, err := NewComponentLinker()
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}

			pm, err := profiles.NewProfileManager(componentLinker)
			if err != nil {
				log.Fatal("Failed to create profile manager:", err)
			}

			if err := pm.RemoveComponentFromProfile(profileName, componentType, componentName); err != nil {
				log.Fatal("Failed to remove component from profile:", err)
			}
		},
		func(targetProfile string, sourceProfiles []string) {
			// Cherry-pick handler - interactively select components from profiles
			pm, err := profiles.NewProfileManager(nil)
			if err != nil {
				log.Fatal("Failed to create profile manager:", err)
			}

			// Check if target profile exists, if not create it
			allProfiles, err := pm.ScanProfiles()
			if err != nil {
				log.Fatal("Failed to scan profiles:", err)
			}

			profileExists := false
			for _, p := range allProfiles {
				if p.Name == targetProfile {
					profileExists = true
					break
				}
			}

			if !profileExists {
				appFormatter.Info("Target profile '%s' does not exist. Creating it...", targetProfile)
				appFormatter.EmptyLine()
				if err := pm.CreateProfile(targetProfile); err != nil {
					log.Fatal("Failed to create target profile:", err)
				}
				appFormatter.EmptyLine()
			}

			// Get all available components
			appFormatter.Info("Scanning for available components...")
			if len(sourceProfiles) > 0 {
				appFormatter.Info("Source profiles: %s", joinStrings(sourceProfiles, ", "))
				appFormatter.EmptyLine()
			} else {
				appFormatter.Info("Source profiles: All profiles")
				appFormatter.EmptyLine()
			}

			components, err := pm.GetAllAvailableComponents(sourceProfiles)
			if err != nil {
				log.Fatal("Failed to get available components:", err)
			}

			if len(components) == 0 {
				appFormatter.Info("No components found in source profiles.")
				if len(sourceProfiles) > 0 {
					appFormatter.EmptyLine()
					appFormatter.Info("The specified source profiles may be empty or not exist.")
				} else {
					appFormatter.EmptyLine()
					appFormatter.Info("Try installing some components first with:")
					appFormatter.Info("  agent-smith install skill <repo-url> <skill-name> --profile <profile-name>")
				}
				return
			}

			// Display available components grouped by type
			appFormatter.Info("Found %d component(s) available for cherry-picking:", len(components))
			appFormatter.EmptyLine()

			// Group by type
			agentItems := []profiles.ComponentItem{}
			skillItems := []profiles.ComponentItem{}
			commandItems := []profiles.ComponentItem{}

			for _, comp := range components {
				switch comp.Type {
				case "agents":
					agentItems = append(agentItems, comp)
				case "skills":
					skillItems = append(skillItems, comp)
				case "commands":
					commandItems = append(commandItems, comp)
				}
			}

			// Display agents
			if len(agentItems) > 0 {
				appFormatter.Info("Agents (%d):", len(agentItems))
				for i, comp := range agentItems {
					appFormatter.Info("  [%d] %s (from %s)", i+1, comp.Name, comp.SourceProfile)
				}
				appFormatter.EmptyLine()
			}

			// Display skills
			if len(skillItems) > 0 {
				appFormatter.Info("Skills (%d):", len(skillItems))
				for i, comp := range skillItems {
					appFormatter.Info("  [%d] %s (from %s)", i+len(agentItems)+1, comp.Name, comp.SourceProfile)
				}
				appFormatter.EmptyLine()
			}

			// Display commands
			if len(commandItems) > 0 {
				appFormatter.Info("Commands (%d):", len(commandItems))
				for i, comp := range commandItems {
					appFormatter.Info("  [%d] %s (from %s)", i+len(agentItems)+len(skillItems)+1, comp.Name, comp.SourceProfile)
				}
				appFormatter.EmptyLine()
			}

			// Interactive selection
			appFormatter.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			appFormatter.EmptyLine()
			appFormatter.Info("Select components to copy to the target profile.")
			appFormatter.Info("Enter component numbers separated by spaces (e.g., '1 3 5')")
			appFormatter.Info("Or enter 'all' to select all components, or 'quit' to cancel.")
			fmt.Print("\nSelection: ")

			// Read user input
			var input string
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				input = strings.TrimSpace(scanner.Text())
			}

			if input == "" || strings.ToLower(input) == "quit" {
				appFormatter.EmptyLine()
				appFormatter.Info("Cancelled.")
				return
			}

			// Parse selection
			var selectedComponents []profiles.ComponentItem

			if strings.ToLower(input) == "all" {
				selectedComponents = components
				appFormatter.Info("\nSelected all %d components.", len(components))
				appFormatter.EmptyLine()
			} else {
				// Parse individual numbers
				parts := strings.Fields(input)
				selectedIndices := make(map[int]bool)

				for _, part := range parts {
					idx, err := strconv.Atoi(part)
					if err != nil || idx < 1 || idx > len(components) {
						appFormatter.PlainWarning("Invalid selection '%s' (valid range: 1-%d)", part, len(components))
						continue
					}
					selectedIndices[idx-1] = true
				}

				if len(selectedIndices) == 0 {
					appFormatter.EmptyLine()
					appFormatter.Info("No valid selections made. Cancelled.")
					return
				}

				for idx := range selectedIndices {
					selectedComponents = append(selectedComponents, components[idx])
				}

				appFormatter.Info("\nSelected %d component(s).", len(selectedComponents))
				appFormatter.EmptyLine()
			}

			// Execute cherry-pick
			if err := pm.CherryPickComponents(targetProfile, selectedComponents); err != nil {
				log.Fatal("Cherry-pick failed:", err)
			}

			appFormatter.EmptyLine()
			appFormatter.Info("To activate this profile and use these components:")
			appFormatter.Info("  agent-smith profile activate %s", targetProfile)
			appFormatter.Info("  agent-smith link all")
		},
		func() {
			// Status handler - shows current system status
			debugPrintln("[DEBUG] handleStatus called")
			pm, err := profiles.NewProfileManager(nil)
			if err != nil {
				log.Fatal("Failed to create profile manager:", err)
			}

			// Get active profile
			activeProfile, err := pm.GetActiveProfile()
			if err != nil {
				log.Fatal("Failed to get active profile:", err)
			}
			debugPrintf("[DEBUG] Active profile: %s\n", activeProfile)

			// Detect all available targets
			debugPrintln("[DEBUG] Detecting targets")
			targets, err := config.DetectAllTargets()
			if err != nil {
				log.Fatal("Failed to detect targets:", err)
			}
			debugPrintf("[DEBUG] Detected %d target(s)\n", len(targets))

			// Get agents directory
			agentsDir, err := paths.GetAgentsDir()
			if err != nil {
				log.Fatal("Failed to get agents directory:", err)
			}
			debugPrintf("[DEBUG] Agents directory: %s\n", agentsDir)

			// Count components in ~/.agent-smith/
			agentsPath := filepath.Join(agentsDir, "agents")
			skillsPath := filepath.Join(agentsDir, "skills")
			commandsPath := filepath.Join(agentsDir, "commands")

			agentsCount := 0
			skillsCount := 0
			commandsCount := 0

			if entries, err := os.ReadDir(agentsPath); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						agentsCount++
					}
				}
			}

			if entries, err := os.ReadDir(skillsPath); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						skillsCount++
					}
				}
			}

			if entries, err := os.ReadDir(commandsPath); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						commandsCount++
					}
				}
			}

			// Display status with modern formatting
			f := formatter.New()
			f.SectionHeader("Agent Smith Status")

			// Show active profile
			if activeProfile != "" {
				green := color.New(color.FgGreen).SprintFunc()
				appFormatter.Info("  Active Profile:     %s %s", green(activeProfile), formatter.ColoredSuccess())
			} else {
				gray := color.New(color.FgHiBlack).SprintFunc()
				appFormatter.Info("  Active Profile:     %s", gray("None"))
			}

			// Show detected targets
			if len(targets) > 0 {
				var targetNames []string
				for _, target := range targets {
					targetNames = append(targetNames, target.GetName())
				}
				cyan := color.New(color.FgCyan).SprintFunc()
				appFormatter.Info("  Detected Targets:   %s", cyan(joinStrings(targetNames, ", ")))
			} else {
				gray := color.New(color.FgHiBlack).SprintFunc()
				appFormatter.Info("  Detected Targets:   %s", gray("None"))
			}

			// Show base components count
			appFormatter.EmptyLine()
			bold := color.New(color.Bold).SprintFunc()
			appFormatter.Info("%s", bold("Base Components (~/.agent-smith/)"))
			appFormatter.Info("  • Agents:           %d", agentsCount)
			appFormatter.Info("  • Skills:           %d", skillsCount)
			appFormatter.Info("  • Commands:         %d", commandsCount)

			// If there's an active profile, show its components
			if activeProfile != "" {
				profilesList, err := pm.ScanProfiles()
				if err == nil {
					for _, profile := range profilesList {
						if profile.Name == activeProfile {
							agents, skills, commands := pm.CountComponents(profile)
							appFormatter.EmptyLine()
							green := color.New(color.FgGreen, color.Bold).SprintFunc()
							appFormatter.Info("%s", green("Active Profile Components"))
							appFormatter.Info("  • Agents:           %d", agents)
							appFormatter.Info("  • Skills:           %d", skills)
							appFormatter.Info("  • Commands:         %d", commands)
							break
						}
					}
				}
			}

			// Show helpful commands
			appFormatter.EmptyLine()
			dim := color.New(color.Faint).SprintFunc()
			appFormatter.Info("%s", dim("Quick Actions:"))
			appFormatter.Info("  %s agent-smith link status     %s", dim("•"), dim("View component link status"))
			appFormatter.Info("  %s agent-smith profile list    %s", dim("•"), dim("List all profiles"))
			appFormatter.EmptyLine()
		},
		func(name, path string) {
			// Load existing config
			cfg, err := config.LoadConfig()
			if err != nil {
				log.Fatal("Failed to load config:", err)
			}

			// Validate that target name doesn't already exist
			for _, target := range cfg.CustomTargets {
				if target.Name == name {
					log.Fatalf("Target '%s' already exists in config", name)
				}
			}

			// Create new custom target config
			newTarget := config.CustomTargetConfig{
				Name:        name,
				BaseDir:     path,
				SkillsDir:   "skills",
				AgentsDir:   "agents",
				CommandsDir: "commands",
			}

			// Add to config
			cfg.CustomTargets = append(cfg.CustomTargets, newTarget)

			// Save config
			if err := config.SaveConfig(cfg); err != nil {
				log.Fatal("Failed to save config:", err)
			}

			infoPrintf("%s Successfully added custom target '%s'\n", formatter.SymbolSuccess, name)
			infoPrintf("  Base directory: %s\n", path)
			infoPrintln("\nSubdirectories:")
			infoPrintf("  Skills:   %s/skills\n", path)
			infoPrintf("  Agents:   %s/agents\n", path)
			infoPrintf("  Commands: %s/commands\n", path)
			infoPrintln("\nYou can now link components to this target:")
			infoPrintf("  agent-smith link all --target %s\n", name)
		},
		func(name string) {
			// Load existing config
			cfg, err := config.LoadConfig()
			if err != nil {
				log.Fatal("Failed to load config:", err)
			}

			// Check if target exists and is a custom target
			found := false
			targetIndex := -1
			for i, target := range cfg.CustomTargets {
				if target.Name == name {
					found = true
					targetIndex = i
					break
				}
			}

			if !found {
				log.Fatalf("Target '%s' not found in custom targets", name)
			}

			// Remove the target from the slice
			cfg.CustomTargets = append(cfg.CustomTargets[:targetIndex], cfg.CustomTargets[targetIndex+1:]...)

			// Save config
			if err := config.SaveConfig(cfg); err != nil {
				log.Fatal("Failed to save config:", err)
			}

			infoPrintf("%s Successfully removed custom target '%s'\n", formatter.SymbolSuccess, name)
			infoPrintln("\nNote: This only removes the target from configuration.")
			infoPrintln("Components linked to this target are not automatically unlinked.")
		},
		func() {
			// Load config to distinguish between built-in and custom targets
			cfg, err := config.LoadConfig()
			if err != nil {
				log.Fatal("Failed to load config:", err)
			}

			// Get all built-in targets (even if not detected)
			builtInNames := []string{"opencode", "claudecode"}

			// Create formatter instance
			f := formatter.New()
			green := color.New(color.FgGreen).SprintFunc()
			red := color.New(color.FgRed).SprintFunc()
			yellow := color.New(color.FgYellow).SprintFunc()

			// Section header
			f.SectionHeader("Available Targets")

			// Collect all target data
			type targetInfo struct {
				name     string
				baseDir  string
				exists   bool
				isCustom bool
				hasError bool
			}
			var allTargets []targetInfo

			// Collect built-in targets
			for _, name := range builtInNames {
				var target config.Target
				var err error

				if name == "opencode" {
					target, err = config.NewOpencodeTarget()
				} else if name == "claudecode" {
					target, err = config.NewClaudeCodeTarget()
				}

				if err != nil {
					allTargets = append(allTargets, targetInfo{
						name:     name,
						baseDir:  "error loading target",
						exists:   false,
						isCustom: false,
						hasError: true,
					})
					continue
				}

				baseDir, _ := target.GetBaseDir()
				exists := false
				if _, err := os.Stat(baseDir); err == nil {
					exists = true
				}

				allTargets = append(allTargets, targetInfo{
					name:     name,
					baseDir:  baseDir,
					exists:   exists,
					isCustom: false,
					hasError: false,
				})
			}

			// Collect custom targets
			for _, customTargetConfig := range cfg.CustomTargets {
				customTarget, err := config.NewCustomTarget(customTargetConfig)
				if err != nil {
					allTargets = append(allTargets, targetInfo{
						name:     customTargetConfig.Name,
						baseDir:  "error loading target",
						exists:   false,
						isCustom: true,
						hasError: true,
					})
					continue
				}

				baseDir, _ := customTarget.GetBaseDir()
				exists := false
				if _, err := os.Stat(baseDir); err == nil {
					exists = true
				}

				allTargets = append(allTargets, targetInfo{
					name:     customTargetConfig.Name,
					baseDir:  baseDir,
					exists:   exists,
					isCustom: true,
					hasError: false,
				})
			}

			// Create table with box-drawing characters
			table := formatter.NewBoxTable(os.Stdout, []string{"Status", "Target", "Type", "Location"})

			// Add rows to table
			availableCount := 0
			for _, target := range allTargets {
				var statusSymbol string
				var targetType string

				if target.hasError {
					statusSymbol = red(formatter.SymbolError)
				} else if target.exists {
					statusSymbol = green(formatter.SymbolSuccess)
					availableCount++
				} else {
					statusSymbol = yellow(formatter.SymbolNotLinked)
				}

				if target.isCustom {
					targetType = "Custom"
				} else {
					targetType = "Built-in"
				}

				table.AddRow([]string{statusSymbol, target.name, targetType, target.baseDir})
			}

			// Render the table
			table.Render()

			// Display summary
			appFormatter.EmptyLine()
			totalCount := len(allTargets)
			if availableCount == totalCount {
				appFormatter.Info("%s All %d target(s) detected and available", green(formatter.SymbolSuccess), totalCount)
			} else if availableCount > 0 {
				appFormatter.Info("%s %d of %d target(s) available", yellow(formatter.SymbolWarning), availableCount, totalCount)
			} else {
				appFormatter.Info("%s No targets currently available", red(formatter.SymbolError))
			}

			// Display legend
			appFormatter.EmptyLine()
			appFormatter.Info("Legend:")
			appFormatter.Info("  %s Available  %s Not found  %s Error",
				green(formatter.SymbolSuccess),
				yellow(formatter.SymbolNotLinked),
				red(formatter.SymbolError))
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
			infoPrintf("Materialized Components in %s:\n\n", projectRoot)

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
				infoPrintf("%s %s\n", green(formatter.SymbolSuccess), targetLabel)

				// Display skills
				if len(metadata.Skills) > 0 {
					infoPrintf("  Skills (%d):\n", len(metadata.Skills))
					for name, meta := range metadata.Skills {
						sourceInfo := meta.Source
						if meta.SourceProfile != "" {
							sourceInfo = fmt.Sprintf("%s (profile: %s)", meta.Source, meta.SourceProfile)
						}
						infoPrintf("    • %-30s (from %s)\n", name, sourceInfo)
					}
				}

				// Display agents
				if len(metadata.Agents) > 0 {
					infoPrintf("  Agents (%d):\n", len(metadata.Agents))
					for name, meta := range metadata.Agents {
						sourceInfo := meta.Source
						if meta.SourceProfile != "" {
							sourceInfo = fmt.Sprintf("%s (profile: %s)", meta.Source, meta.SourceProfile)
						}
						infoPrintf("    • %-30s (from %s)\n", name, sourceInfo)
					}
				}

				// Display commands
				if len(metadata.Commands) > 0 {
					infoPrintf("  Commands (%d):\n", len(metadata.Commands))
					for name, meta := range metadata.Commands {
						sourceInfo := meta.Source
						if meta.SourceProfile != "" {
							sourceInfo = fmt.Sprintf("%s (profile: %s)", meta.Source, meta.SourceProfile)
						}
						infoPrintf("    • %-30s (from %s)\n", name, sourceInfo)
					}
				}

				infoPrintln("")
			}

			if !foundAny {
				infoPrintf("%s No components materialized yet\n\n", yellow(formatter.SymbolWarning))
				infoPrintln("To materialize components:")
				infoPrintln("  agent-smith materialize skill <name> --target opencode")
				infoPrintln("  agent-smith materialize all --target opencode")
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
						infoPrintf("%s Component '%s' not found in %s target\n", red(formatter.SymbolError), componentName, targetName)
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

				infoPrintf("\n%s Provenance Information - %s\n\n", green(formatter.SymbolSuccess), bold(targetLabel))

				// Display component information
				infoPrintf("  %s: %s\n", cyan("Component"), componentName)
				infoPrintf("  %s: %s\n", cyan("Type"), componentType)
				infoPrintln("")

				// Display source information
				infoPrintf("  %s\n", bold("Source Information:"))
				infoPrintf("    %s: %s\n", cyan("Repository"), meta.Source)
				infoPrintf("    %s: %s\n", cyan("Source Type"), meta.SourceType)
				if meta.SourceProfile != "" {
					infoPrintf("    %s: %s\n", cyan("Profile"), meta.SourceProfile)
				}
				infoPrintf("    %s: %s\n", cyan("Commit Hash"), meta.CommitHash)
				infoPrintf("    %s: %s\n", cyan("Original Path"), meta.OriginalPath)
				infoPrintln("")

				// Display materialization information
				infoPrintf("  %s\n", bold("Materialization:"))
				infoPrintf("    %s: %s\n", cyan("Materialized At"), meta.MaterializedAt)
				infoPrintf("    %s: %s\n", cyan("Target Directory"), targetDir)
				infoPrintln("")

				// Display hash information for sync status
				infoPrintf("  %s\n", bold("Sync Status:"))
				infoPrintf("    %s: %s\n", cyan("Source Hash"), meta.SourceHash)

				// Recalculate current hash from the actual directory
				componentPath := filepath.Join(targetDir, componentType, componentName)
				actualCurrentHash, err := materializer.CalculateDirectoryHash(componentPath)
				if err != nil {
					debugPrintf("[DEBUG] Failed to calculate current hash: %v\n", err)
					// Fall back to stored hash
					actualCurrentHash = meta.CurrentHash
				}

				infoPrintf("    %s: %s\n", cyan("Current Hash"), actualCurrentHash)

				// Check if hashes match
				if meta.SourceHash == actualCurrentHash {
					infoPrintf("    %s: %s (component is unchanged)\n", cyan("Status"), green("In Sync"))
				} else {
					infoPrintf("    %s: %s (component has been modified)\n", cyan("Status"), yellow("Modified"))
				}

				infoPrintln("")
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
	)

	// Execute Cobra command
	cmd.Execute()
}
