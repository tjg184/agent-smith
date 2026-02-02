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

		// Validate that the profile exists
		profilePath := filepath.Join(profilesDir, profile)
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("profile '%s' does not exist", profile)
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

	return linker.NewComponentLinker(agentsDir, targets, det, nil)
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
				log.Fatal("Failed to create component linker:", err)
			}
			debugPrintln("[DEBUG] Component linker created successfully")
			if err := linker.LinkComponent(componentType, componentName); err != nil {
				log.Fatal("Failed to link component:", err)
			}
			debugPrintln("[DEBUG] Component linked successfully")
		},
		func(targetFilter, profile string) {
			debugPrintf("[DEBUG] handleLinkAll called with targetFilter=%s, profile=%s\n", targetFilter, profile)
			linker, err := NewComponentLinkerWithFilterAndProfile(targetFilter, profile)
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}
			debugPrintln("[DEBUG] Component linker created successfully")
			if err := linker.LinkAllComponents(); err != nil {
				log.Fatal("Failed to link all components:", err)
			}
			debugPrintln("[DEBUG] All components linked successfully")
		},
		func(componentType, targetFilter, profile string) {
			debugPrintf("[DEBUG] handleLinkType called with componentType=%s, targetFilter=%s, profile=%s\n", componentType, targetFilter, profile)
			linker, err := NewComponentLinkerWithFilterAndProfile(targetFilter, profile)
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
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
					log.Fatal("Failed to create profile manager:", err)
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
		func(targetFilter string, force bool) {
			linker, err := NewComponentLinkerWithFilter(targetFilter)
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}
			if err := linker.UnlinkAllComponents(targetFilter, force); err != nil {
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
		func(profileFilter []string, activeOnly bool) {
			pm, err := profiles.NewProfileManager(nil)
			if err != nil {
				log.Fatal("Failed to create profile manager:", err)
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

			// Print top border with title - centered "Available Profiles" header
			appFormatter.Info("┌────────────────────────────────────────────────────────────────────────────┐")
			appFormatter.Info("│                            Available Profiles                              │")
			appFormatter.Info("├────────────────────────────────────────────────────────────────────────────┤")

			for _, profile := range filteredProfiles {
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

				// Check if this is the active profile
				activeIndicator := " "
				if profile.Name == activeProfile {
					activeIndicator = formatter.ColoredSuccess()
				}

				// Format the line: "│ ✓ profile-name                (X agents, Y skills, Z commands)     │"
				// Total width: 80 characters (including borders)
				// Inner width: 76 characters (80 - 2 for borders - 2 for padding)
				// Layout: " [indicator] [name with padding] [component counts with padding] "

				// Available space: 76 - 1 (indicator) - 2 (spaces around indicator) = 73
				availableSpace := 73

				// Calculate how much space we need
				nameLen := len(profile.Name)
				countLen := len(componentStr)

				// We want to pad between name and count
				totalContentLen := nameLen + countLen
				padding := availableSpace - totalContentLen
				if padding < 2 {
					padding = 2 // Minimum 2 spaces between name and count
				}

				// Truncate name if too long
				displayName := profile.Name
				if nameLen > 40 {
					displayName = displayName[:37] + "..."
					nameLen = 40
					padding = availableSpace - nameLen - countLen
					if padding < 2 {
						padding = 2
					}
				}

				// Calculate final padding to fill the line
				rightPadding := availableSpace - nameLen - countLen - padding

				appFormatter.Info("│ %s %s%s%s%s │",
					activeIndicator,
					displayName,
					strings.Repeat(" ", padding),
					componentStr,
					strings.Repeat(" ", rightPadding))
			}

			// Print bottom border
			appFormatter.Info("└────────────────────────────────────────────────────────────────────────────┘")

			// Display legend
			appFormatter.EmptyLine()
			appFormatter.Info("Legend:")
			appFormatter.Info("  %s - Currently active profile", formatter.ColoredSuccess())

			// Display total count
			if len(profileFilter) > 0 || activeOnly {
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
		func(componentType, componentName, target, projectDir string) {
			// Import necessary packages at the top of main.go
			// - "github.com/tgaines/agent-smith/pkg/project"
			// - "github.com/tgaines/agent-smith/internal/materializer"
			// - "github.com/tgaines/agent-smith/internal/metadata"

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
					log.Fatal(err)
				}
			}

			// Get source directory (from ~/.agent-smith/)
			baseDir, err := paths.GetAgentsDir()
			if err != nil {
				log.Fatalf("Failed to get agent-smith directory: %v", err)
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
				log.Fatalf("Component '%s' not found in ~/.agent-smith/%s/\n\nInstall it first with: agent-smith install %s <repo> %s",
					componentName, componentType, componentType, componentName)
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

			// Materialize to each target
			for _, targetName := range targets {
				targetDir := project.GetTargetDirectory(projectRoot, targetName)
				if targetDir == "" {
					log.Fatalf("Invalid target: %s", targetName)
				}

				// Ensure target structure exists
				if err := project.EnsureTargetStructure(targetDir); err != nil {
					log.Fatalf("Failed to create target structure: %v", err)
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
						infoPrintf("⊘ Skipped %s '%s' to %s (already exists and identical)\n", componentType, componentName, targetName)
						continue
					} else {
						log.Fatalf("Component '%s' already exists in %s and differs.\n\nUse --force to overwrite (not yet implemented in this version)", componentName, targetName)
					}
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
					"", // sourceProfile (not implemented yet)
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
				infoPrintf("  Destination: %s\n", destPath)
			}
		},
	)

	// Execute Cobra command
	cmd.Execute()
}
