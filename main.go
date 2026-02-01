package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/tgaines/agent-smith/cmd"
	"github.com/tgaines/agent-smith/internal/detector"
	"github.com/tgaines/agent-smith/internal/downloader"
	"github.com/tgaines/agent-smith/internal/fileutil"
	"github.com/tgaines/agent-smith/internal/formatter"
	"github.com/tgaines/agent-smith/internal/linker"
	metadataPkg "github.com/tgaines/agent-smith/internal/metadata"
	"github.com/tgaines/agent-smith/internal/models"
	"github.com/tgaines/agent-smith/internal/uninstaller"
	"github.com/tgaines/agent-smith/internal/updater"
	"github.com/tgaines/agent-smith/pkg/config"
	"github.com/tgaines/agent-smith/pkg/logger"
	"github.com/tgaines/agent-smith/pkg/paths"
	"github.com/tgaines/agent-smith/pkg/profiles"
)

// appLogger is the global logger instance used throughout the application
var appLogger *logger.Logger

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

	return linker.NewComponentLinker(agentsDir, targets, det)
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

	return linker.NewComponentLinker(agentsDir, targets, det)
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

	return linker.NewComponentLinker(agentsDir, targets, det)
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
				log.Fatal("Cannot specify both --profile and --target-dir flags")
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
					log.Fatal("Failed to download skill:", err)
				}
				debugPrintln("[DEBUG] Skill download completed successfully")
			} else if profile != "" {
				// Install directly to profile
				debugPrintf("[DEBUG] Installing to profile: %s\n", profile)
				pm, err := profiles.NewProfileManager(nil)
				if err != nil {
					log.Fatal("Failed to create profile manager:", err)
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
					log.Fatalf("Profile '%s' does not exist. Create it first with: agent-smith profiles create %s", profile, profile)
				}

				dl := downloader.NewSkillDownloaderForProfile(profile)
				if err := dl.DownloadSkill(repoURL, name); err != nil {
					log.Fatal("Failed to download skill:", err)
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
					log.Fatal("Failed to download skill:", err)
				}
				debugPrintln("[DEBUG] Skill download completed successfully")
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
					fmt.Println("\nProfile updated successfully!")
					fmt.Printf("To activate this profile and use these components, run:\n")
					fmt.Printf("  agent-smith profile activate %s\n", profileName)
					fmt.Printf("  agent-smith link all\n")
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
		func() {
			linker, err := NewComponentLinker()
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}
			if err := linker.ShowLinkStatus(); err != nil {
				log.Fatal("Failed to show link status:", err)
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
		func() {
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

			// Display results
			if len(profilesList) == 0 {
				fmt.Println("No profiles found in ~/.agent-smith/profiles/")
				fmt.Println("\nTo create a profile, run:")
				fmt.Println("  ./agent-smith profile create <profile-name>")
				return
			}

			fmt.Println("Available Profiles:")
			fmt.Println()

			for _, profile := range profilesList {
				// Count components
				agents, skills, commands := pm.CountComponents(profile)

				// Build component counts string
				var components []string
				if agents > 0 {
					components = append(components, fmt.Sprintf("%d agent(s)", agents))
				}
				if skills > 0 {
					components = append(components, fmt.Sprintf("%d skill(s)", skills))
				}
				if commands > 0 {
					components = append(components, fmt.Sprintf("%d command(s)", commands))
				}

				componentStr := ""
				if len(components) > 0 {
					componentStr = fmt.Sprintf(" (%s)", joinStrings(components, ", "))
				}

				// Check if this is the active profile
				activeIndicator := "  "
				activeLabel := ""
				if profile.Name == activeProfile {
					activeIndicator = fmt.Sprintf("%s ", formatter.SymbolSuccess)
					activeLabel = " [active]"
				}

				fmt.Printf("%s%-15s%s%s\n", activeIndicator, profile.Name, activeLabel, componentStr)
			}

			// Display legend
			fmt.Println("\nLegend:")
			fmt.Printf("  %s - Currently active profile\n", formatter.SymbolSuccess)

			// Display total count
			fmt.Printf("\nTotal: %d profile(s)\n", len(profilesList))
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
			fmt.Printf("Profile: %s", targetProfile.Name)
			if targetProfile.Name == activeProfile {
				fmt.Printf(" %s [active]", formatter.SymbolSuccess)
			}
			fmt.Println()
			fmt.Printf("Location: %s\n", targetProfile.BasePath)
			fmt.Println()

			// Get component names
			agents, skills, commands := pm.GetComponentNames(targetProfile)

			// Display agents
			if len(agents) > 0 {
				fmt.Printf("Agents (%d):\n", len(agents))
				for _, agent := range agents {
					fmt.Printf("  - %s\n", agent)
				}
				fmt.Println()
			}

			// Display skills
			if len(skills) > 0 {
				fmt.Printf("Skills (%d):\n", len(skills))
				for _, skill := range skills {
					fmt.Printf("  - %s\n", skill)
				}
				fmt.Println()
			}

			// Display commands
			if len(commands) > 0 {
				fmt.Printf("Commands (%d):\n", len(commands))
				for _, command := range commands {
					fmt.Printf("  - %s\n", command)
				}
				fmt.Println()
			}

			// Show empty state if no components
			if len(agents) == 0 && len(skills) == 0 && len(commands) == 0 {
				fmt.Println("This profile is empty.")
				fmt.Println("\nAdd components with:")
				fmt.Printf("  agent-smith profiles add <type> %s <component-name>\n", profileName)
			} else if targetProfile.Name != activeProfile {
				// Show activation hint if not active
				fmt.Println("To activate this profile:")
				fmt.Printf("  agent-smith profiles activate %s\n", profileName)
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

			// Display status - use fmt.Println to always show output without flags
			fmt.Println("Current Configuration:")
			fmt.Println()

			// Show active profile
			if activeProfile != "" {
				fmt.Printf("  Active Profile: %s %s\n", activeProfile, formatter.SymbolSuccess)
			} else {
				fmt.Println("  Active Profile: None")
			}

			// Show detected targets
			if len(targets) > 0 {
				var targetNames []string
				for _, target := range targets {
					targetNames = append(targetNames, target.GetName())
				}
				fmt.Printf("  Detected Targets: %s\n", joinStrings(targetNames, ", "))
			} else {
				fmt.Println("  Detected Targets: None")
			}

			fmt.Println()
			fmt.Println("Components in ~/.agent-smith/:")
			fmt.Printf("  Agents: %d\n", agentsCount)
			fmt.Printf("  Skills: %d\n", skillsCount)
			fmt.Printf("  Commands: %d\n", commandsCount)

			// If there's an active profile, show its components
			if activeProfile != "" {
				profilesList, err := pm.ScanProfiles()
				if err == nil {
					for _, profile := range profilesList {
						if profile.Name == activeProfile {
							agents, skills, commands := pm.CountComponents(profile)
							fmt.Println()
							fmt.Printf("Active Profile (%s):\n", activeProfile)
							fmt.Printf("  Agents: %d\n", agents)
							fmt.Printf("  Skills: %d\n", skills)
							fmt.Printf("  Commands: %d\n", commands)
							break
						}
					}
				}
			}

			fmt.Println()
			fmt.Println("For more details:")
			fmt.Println("  - Run 'agent-smith link status' for link information")
			fmt.Println("  - Run 'agent-smith profile list' to see all profiles")
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

			fmt.Println("Available Targets:")
			fmt.Println()

			// Display built-in targets
			fmt.Println("Built-in Targets:")
			for _, name := range builtInNames {
				var target config.Target
				var err error

				if name == "opencode" {
					target, err = config.NewOpencodeTarget()
				} else if name == "claudecode" {
					target, err = config.NewClaudeCodeTarget()
				}

				if err != nil {
					continue
				}

				baseDir, _ := target.GetBaseDir()
				exists := false
				if _, err := os.Stat(baseDir); err == nil {
					exists = true
				}

				symbol := formatter.SymbolNotLinked
				status := "[not found]"
				if exists {
					symbol = formatter.SymbolSuccess
					status = "[detected]"
				}

				fmt.Printf("  %s %-15s %-30s %s\n", symbol, name, baseDir, status)
			}

			// Display custom targets
			if len(cfg.CustomTargets) > 0 {
				fmt.Println()
				fmt.Println("Custom Targets:")
				for _, customTargetConfig := range cfg.CustomTargets {
					customTarget, err := config.NewCustomTarget(customTargetConfig)
					if err != nil {
						fmt.Printf("  %s %-15s <error loading target>\n", formatter.SymbolError, customTargetConfig.Name)
						continue
					}

					baseDir, _ := customTarget.GetBaseDir()
					exists := false
					if _, err := os.Stat(baseDir); err == nil {
						exists = true
					}

					symbol := formatter.SymbolNotLinked
					status := "[not found]"
					if exists {
						symbol = formatter.SymbolSuccess
						status = "[detected]"
					}

					fmt.Printf("  %s %-15s %-30s %s\n", symbol, customTargetConfig.Name, baseDir, status)
				}
			}

			// Display legend
			fmt.Println()
			fmt.Println("Legend:")
			fmt.Printf("  %s - Target directory exists\n", formatter.SymbolSuccess)
			fmt.Printf("  %s - Target directory not found\n", formatter.SymbolNotLinked)

			// Count available targets
			availableCount := 0
			totalCount := len(builtInNames) + len(cfg.CustomTargets)

			for _, name := range builtInNames {
				var target config.Target
				var err error

				if name == "opencode" {
					target, err = config.NewOpencodeTarget()
				} else if name == "claudecode" {
					target, err = config.NewClaudeCodeTarget()
				}

				if err == nil {
					baseDir, _ := target.GetBaseDir()
					if _, err := os.Stat(baseDir); err == nil {
						availableCount++
					}
				}
			}

			for _, customTargetConfig := range cfg.CustomTargets {
				customTarget, err := config.NewCustomTarget(customTargetConfig)
				if err == nil {
					baseDir, _ := customTarget.GetBaseDir()
					if _, err := os.Stat(baseDir); err == nil {
						availableCount++
					}
				}
			}

			fmt.Println()
			fmt.Printf("Total: %d target(s) (%d available)\n", totalCount, availableCount)
		},
	)

	// Execute Cobra command
	cmd.Execute()
}
