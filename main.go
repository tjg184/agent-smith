package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjg184/agent-smith/cmd"
	"github.com/tjg184/agent-smith/internal/detector"
	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/internal/linker"
	"github.com/tjg184/agent-smith/pkg/config"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/paths"
	"github.com/tjg184/agent-smith/pkg/profiles"
	"github.com/tjg184/agent-smith/pkg/services"
	findsvc "github.com/tjg184/agent-smith/pkg/services/find"
	installsvc "github.com/tjg184/agent-smith/pkg/services/install"
	linksvc "github.com/tjg184/agent-smith/pkg/services/link"
	locksvc "github.com/tjg184/agent-smith/pkg/services/lock"
	materializesvc "github.com/tjg184/agent-smith/pkg/services/materialize"
	profilesvc "github.com/tjg184/agent-smith/pkg/services/profile"
	statussvc "github.com/tjg184/agent-smith/pkg/services/status"
	targetsvc "github.com/tjg184/agent-smith/pkg/services/target"
	uninstallsvc "github.com/tjg184/agent-smith/pkg/services/uninstall"
	updatesvc "github.com/tjg184/agent-smith/pkg/services/update"
)

var appLogger *logger.Logger
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
		appLogger.Info("%s", fmt.Sprint(a...))
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
		appLogger.Debug("%s", fmt.Sprint(a...))
	}
}

func NewLockService() services.ComponentLockService {
	return locksvc.NewService(appLogger)
}

func NewComponentLinker() (*linker.ComponentLinker, error) {
	debugPrintln("[DEBUG] NewComponentLinker: Creating component linker")
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}
	debugPrintf("[DEBUG] NewComponentLinker: Base agents directory: %s\n", agentsDir)

	profileManager, err := profiles.NewProfileManager(nil, NewLockService())
	if err != nil {
		return nil, fmt.Errorf("failed to create profile manager: %w", err)
	}

	activeProfile, err := profileManager.GetActiveProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to get active profile: %w", err)
	}

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

func NewComponentLinkerWithProfileManager(pm *profiles.ProfileManager) (*linker.ComponentLinker, error) {
	debugPrintln("[DEBUG] NewComponentLinkerWithProfileManager: Creating component linker with profile manager")

	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}
	debugPrintf("[DEBUG] NewComponentLinkerWithProfileManager: Base agents directory: %s\n", agentsDir)

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

	if profile != "" {
		debugPrintf("[DEBUG] NewComponentLinkerWithFilterAndProfile: Explicit profile specified: %s\n", profile)
		profilesDir, err := paths.GetProfilesDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get profiles directory: %w", err)
		}

		profilePath := filepath.Join(profilesDir, profile)
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			pm, pmErr := profiles.NewProfileManager(nil, NewLockService())
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
			return nil, fmt.Errorf("profile '%s' does not exist\n\nTo create this profile:\n  agent-smith profile create %s\n\nTo list available profiles:\n  agent-smith profile list", profile, profile)
		}

		profileObj := &profiles.Profile{
			Name:        profile,
			BasePath:    profilePath,
			HasAgents:   false,
			HasSkills:   false,
			HasCommands: false,
		}

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
		profileManager, err := profiles.NewProfileManager(nil, NewLockService())
		if err != nil {
			return nil, fmt.Errorf("failed to create profile manager: %w", err)
		}

		activeProfile, err := profileManager.GetActiveProfile()
		if err != nil {
			return nil, fmt.Errorf("failed to get active profile: %w", err)
		}

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

	debugPrintln("[DEBUG] NewComponentLinkerWithFilterAndProfile: Detecting all targets")
	allTargets, err := config.DetectAllTargets()
	if err != nil {
		return nil, fmt.Errorf("failed to detect targets: %w", err)
	}
	debugPrintf("[DEBUG] NewComponentLinkerWithFilterAndProfile: Detected %d target(s)\n", len(allTargets))

	if targetFilter == "" || targetFilter == "all" {
		debugPrintln("[DEBUG] NewComponentLinkerWithFilterAndProfile: Using all detected targets")
		targets = allTargets
	} else {
		debugPrintf("[DEBUG] NewComponentLinkerWithFilterAndProfile: Filtering for target: %s\n", targetFilter)
		for _, target := range allTargets {
			if target.GetName() == targetFilter {
				targets = append(targets, target)
				debugPrintf("[DEBUG] NewComponentLinkerWithFilterAndProfile: Found matching target: %s\n", targetFilter)
				break
			}
		}
		if len(targets) == 0 {
			debugPrintf("[DEBUG] NewComponentLinkerWithFilterAndProfile: Target '%s' not found\n", targetFilter)
			return nil, fmt.Errorf("target '%s' not found. Available targets: %v", targetFilter, getTargetNames(allTargets))
		}
	}
	debugPrintf("[DEBUG] NewComponentLinkerWithFilterAndProfile: Using %d target(s)\n", len(targets))

	det := detector.NewRepositoryDetector()
	if appLogger != nil {
		det.SetLogger(appLogger)
	}

	return linker.NewComponentLinker(agentsDir, targets, det, nil)
}

func NewComponentLinkerWithFilter(targetFilter string) (*linker.ComponentLinker, error) {
	debugPrintf("[DEBUG] NewComponentLinkerWithFilter: Creating component linker with filter=%s\n", targetFilter)
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}
	debugPrintf("[DEBUG] NewComponentLinkerWithFilter: Base agents directory: %s\n", agentsDir)

	profileManager, err := profiles.NewProfileManager(nil, NewLockService())
	if err != nil {
		return nil, fmt.Errorf("failed to create profile manager: %w", err)
	}

	activeProfile, err := profileManager.GetActiveProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to get active profile: %w", err)
	}

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

	debugPrintln("[DEBUG] NewComponentLinkerWithFilter: Detecting all targets")
	allTargets, err := config.DetectAllTargets()
	if err != nil {
		return nil, fmt.Errorf("failed to detect targets: %w", err)
	}
	debugPrintf("[DEBUG] NewComponentLinkerWithFilter: Detected %d target(s)\n", len(allTargets))

	if targetFilter == "" || targetFilter == "all" {
		debugPrintln("[DEBUG] NewComponentLinkerWithFilter: Using all detected targets")
		targets = allTargets
	} else {
		debugPrintf("[DEBUG] NewComponentLinkerWithFilter: Filtering for target: %s\n", targetFilter)
		for _, target := range allTargets {
			if target.GetName() == targetFilter {
				targets = append(targets, target)
				debugPrintf("[DEBUG] NewComponentLinkerWithFilter: Found matching target: %s\n", targetFilter)
				break
			}
		}
		if len(targets) == 0 {
			debugPrintf("[DEBUG] NewComponentLinkerWithFilter: Target '%s' not found\n", targetFilter)
			return nil, fmt.Errorf("target '%s' not found. Available targets: %v", targetFilter, getTargetNames(allTargets))
		}
	}
	debugPrintf("[DEBUG] NewComponentLinkerWithFilter: Using %d target(s)\n", len(targets))

	det := detector.NewRepositoryDetector()
	if appLogger != nil {
		det.SetLogger(appLogger)
	}

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

func main() {
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

	appLogger = logger.Default(debugMode, verboseMode)
	appLogger.SetShowTags(false)

	appFormatter = formatter.New()

	if debugMode {
		SetDebugMode(true)
	} else if verboseMode {
		SetVerboseMode(true)
	}

	profileManager, err := profiles.NewProfileManager(nil, NewLockService())
	if err != nil {
		log.Fatal("Failed to initialize profile manager:", err)
	}

	componentLinker, err := NewComponentLinker()
	if err != nil {
		log.Fatal("Failed to initialize component linker:", err)
	}

	installService := installsvc.NewService(profileManager, appLogger, appFormatter)
	updateService := updatesvc.NewService(appLogger, appFormatter)
	uninstallService := uninstallsvc.NewService(componentLinker, appLogger, appFormatter)
	targetService := targetsvc.NewService(appLogger, appFormatter)
	statusService := statussvc.NewService(profileManager, appLogger, appFormatter)
	linkService := linksvc.NewService(profileManager, appLogger, appFormatter)
	profileService := profilesvc.NewService(profileManager, appLogger, appFormatter)
	materializeService := materializesvc.NewService(profileManager, appLogger, appFormatter)
	findService := findsvc.NewService(appLogger, appFormatter)

	// Set up handlers for Cobra commands
	cmd.SetHandlers(
		func(repoURL, name, profile, installDir string) {
			opts := services.InstallOptions{
				Profile:    profile,
				InstallDir: installDir,
			}
			if err := installService.InstallSkill(repoURL, name, opts); err != nil {
				log.Fatal("Failed to install skill:", err)
			}
		},
		func(repoURL, name, profile, installDir string) {
			opts := services.InstallOptions{
				Profile:    profile,
				InstallDir: installDir,
			}
			if err := installService.InstallAgent(repoURL, name, opts); err != nil {
				log.Fatal("Failed to install agent:", err)
			}
		},
		func(repoURL, name, profile, installDir string) {
			opts := services.InstallOptions{
				Profile:    profile,
				InstallDir: installDir,
			}
			if err := installService.InstallCommand(repoURL, name, opts); err != nil {
				log.Fatal("Failed to install command:", err)
			}
		},
		func(repoURL, profile, installDir string) {
			opts := services.InstallOptions{
				Profile:    profile,
				InstallDir: installDir,
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
		func(allProfiles bool, profileFilter []string, linkedOnly bool) {
			opts := services.LinkStatusOptions{
				AllProfiles:   allProfiles,
				ProfileFilter: profileFilter,
				LinkedOnly:    linkedOnly,
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
		func(componentType, componentName, profile, source string) {
			opts := services.UninstallOptions{
				Profile: profile,
				Source:  source,
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
			// If no profile name provided, use active profile
			if profileName == "" {
				pm, err := profiles.NewProfileManager(nil, NewLockService())
				if err != nil {
					log.Fatal("Failed to initialize profile manager:", err)
				}
				activeProfile, err := pm.GetActiveProfile()
				if err != nil {
					log.Fatal("Failed to get active profile:", err)
				}
				if activeProfile == "" {
					log.Fatal("No profile specified and no active profile set")
				}
				profileName = activeProfile
			}

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
		func(profileName, outputFile string) {
			if profileName == "" {
				pm, err := profiles.NewProfileManager(nil, NewLockService())
				if err != nil {
					log.Fatal("Failed to initialize profile manager:", err)
				}
				activeProfile, err := pm.GetActiveProfile()
				if err != nil {
					log.Fatal("Failed to get active profile:", err)
				}
				if activeProfile == "" {
					log.Fatal("No profile specified and no active profile set")
				}
				profileName = activeProfile
			}
			if profileName == "base" {
				profileName = paths.BaseProfileName
			}
			if err := profileService.ShareProfile(profileName, outputFile); err != nil {
				log.Fatal("Failed to share profile:", err)
			}
		},
		func(oldName, newName string) {
			if err := profileService.RenameProfile(oldName, newName); err != nil {
				log.Fatal("Failed to rename profile:", err)
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
		func(componentType, componentName, target, projectDir string, force, dryRun bool, profile, source string) {
			opts := services.MaterializeOptions{
				Target:     target,
				ProjectDir: projectDir,
				Profile:    profile,
				Source:     source,
				Force:      force,
				DryRun:     dryRun,
			}
			if err := materializeService.MaterializeComponent(componentType, componentName, opts); err != nil {
				// Error already logged/displayed by service
				os.Exit(1)
			}
		},
		func(componentType, target, projectDir string, force, dryRun bool, profile string) {
			opts := services.MaterializeOptions{
				Target:     target,
				ProjectDir: projectDir,
				Profile:    profile,
				Force:      force,
				DryRun:     dryRun,
			}
			if err := materializeService.MaterializeByType(componentType, opts); err != nil {
				// Error already logged/displayed by service
				os.Exit(1)
			}
		},
		func(target, projectDir string, force, dryRun bool, profile string) {
			opts := services.MaterializeOptions{
				Target:     target,
				ProjectDir: projectDir,
				Profile:    profile,
				Force:      force,
				DryRun:     dryRun,
			}
			if err := materializeService.MaterializeAll(opts); err != nil {
				// Error already logged/displayed by service
				os.Exit(1)
			}
		},
		func(projectDir string) {
			opts := services.ListMaterializedOptions{
				ProjectDir: projectDir,
			}
			if err := materializeService.ListMaterialized(opts); err != nil {
				log.Fatal("Failed to list materialized components:", err)
			}
		},
		func(componentType, componentName, target, projectDir, source string) {
			opts := services.MaterializeInfoOptions{
				Target:     target,
				ProjectDir: projectDir,
				Source:     source,
			}
			if err := materializeService.ShowComponentInfo(componentType, componentName, opts); err != nil {
				// Error already logged/displayed by service
				os.Exit(1)
			}
		},
		func(target, projectDir string) {
			opts := services.MaterializeStatusOptions{
				Target:     target,
				ProjectDir: projectDir,
			}
			if err := materializeService.ShowStatus(opts); err != nil {
				// Error already logged/displayed by service
				os.Exit(1)
			}
		},
		func(target, projectDir, source string, force, dryRun bool) {
			opts := services.MaterializeUpdateOptions{
				Target:     target,
				ProjectDir: projectDir,
				Source:     source,
				Force:      force,
				DryRun:     dryRun,
			}
			if err := materializeService.UpdateMaterialized(opts); err != nil {
				// Error already logged/displayed by service
				os.Exit(1)
			}
		},
		func(query string, limit int, jsonOutput bool) {
			opts := services.FindOptions{
				Limit: limit,
				JSON:  jsonOutput,
			}
			if err := findService.FindSkills(query, opts); err != nil {
				log.Fatal("Failed to search skills:", err)
			}
		},
	)

	// Execute Cobra command
	cmd.Execute()
}
