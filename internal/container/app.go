package container

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

// App holds application-wide dependencies and wires the CLI together.
type App struct {
	logger    *logger.Logger
	formatter *formatter.Formatter
}

// New creates an App, initialising logger and formatter from os.Args flags.
func New() *App {
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

	appLogger := logger.Default(debugMode, verboseMode)
	appLogger.SetShowTags(false)

	if debugMode {
		appLogger.SetLevel(logger.LevelDebug)
	} else if verboseMode {
		appLogger.SetLevel(logger.LevelInfo)
	}

	return &App{
		logger:    appLogger,
		formatter: formatter.New(),
	}
}

// Run constructs all services, wires command handlers, and executes the CLI.
func (a *App) Run() {
	lockService := locksvc.NewService(a.logger)

	profileManager, err := profiles.NewProfileManager(nil, lockService)
	if err != nil {
		log.Fatal("Failed to initialize profile manager:", err)
	}

	componentLinker, err := a.newComponentLinker()
	if err != nil {
		log.Fatal("Failed to initialize component linker:", err)
	}

	installService := installsvc.NewService(profileManager, a.logger, a.formatter)
	updateService := updatesvc.NewService(a.logger, a.formatter)
	uninstallService := uninstallsvc.NewService(componentLinker, a.logger, a.formatter)
	targetService := targetsvc.NewService(a.logger, a.formatter)
	statusService := statussvc.NewService(profileManager, a.logger, a.formatter)
	linkService := linksvc.NewService(profileManager, a.logger, a.formatter)
	profileService := profilesvc.NewService(profileManager, a.logger, a.formatter)
	materializeService := materializesvc.NewService(profileManager, a.logger, a.formatter)
	findService := findsvc.NewService(a.logger, a.formatter)

	cmd.SetHandlers(
		func(repoURL, name, profile, installDir string) {
			opts := services.InstallOptions{Profile: profile, InstallDir: installDir}
			if err := installService.InstallSkill(repoURL, name, opts); err != nil {
				log.Fatal("Failed to install skill:", err)
			}
		},
		func(repoURL, name, profile, installDir string) {
			opts := services.InstallOptions{Profile: profile, InstallDir: installDir}
			if err := installService.InstallAgent(repoURL, name, opts); err != nil {
				log.Fatal("Failed to install agent:", err)
			}
		},
		func(repoURL, name, profile, installDir string) {
			opts := services.InstallOptions{Profile: profile, InstallDir: installDir}
			if err := installService.InstallCommand(repoURL, name, opts); err != nil {
				log.Fatal("Failed to install command:", err)
			}
		},
		func(repoURL, profile, installDir string) {
			opts := services.InstallOptions{Profile: profile, InstallDir: installDir}
			if err := installService.InstallBulk(repoURL, opts); err != nil {
				log.Fatal("Failed to bulk install:", err)
			}
		},
		func(componentType, componentName, profile string) {
			opts := services.UpdateOptions{Profile: profile}
			if err := updateService.UpdateComponent(componentType, componentName, opts); err != nil {
				log.Fatal("Failed to update component:", err)
			}
		},
		func(profile string) {
			opts := services.UpdateOptions{Profile: profile}
			if err := updateService.UpdateAll(opts); err != nil {
				log.Fatal("Failed to update all components:", err)
			}
		},
		func(componentType, componentName, targetFilter, profile string) {
			opts := services.LinkOptions{TargetFilter: targetFilter, Profile: profile}
			if err := linkService.LinkComponent(componentType, componentName, opts); err != nil {
				log.Fatal("Failed to link component:", err)
			}
		},
		func(targetFilter, profile string, allProfiles bool) {
			opts := services.LinkOptions{TargetFilter: targetFilter, Profile: profile, AllProfiles: allProfiles}
			if err := linkService.LinkAll(opts); err != nil {
				log.Fatal("Failed to link all components:", err)
			}
		},
		func(componentType, targetFilter, profile string) {
			opts := services.LinkOptions{TargetFilter: targetFilter, Profile: profile}
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
			opts := services.LinkStatusOptions{AllProfiles: allProfiles, ProfileFilter: profileFilter, LinkedOnly: linkedOnly}
			if err := linkService.ShowStatus(opts); err != nil {
				log.Fatal("Failed to show link status:", err)
			}
		},
		func(componentType, componentName, targetFilter string) {
			opts := services.UnlinkOptions{TargetFilter: targetFilter}
			if err := linkService.UnlinkComponent(componentType, componentName, opts); err != nil {
				log.Fatal("Failed to unlink component:", err)
			}
		},
		func(componentType, componentName, targetFilter, profile string) {
			opts := services.UnlinkOptions{TargetFilter: targetFilter, Profile: profile}
			if err := linkService.UnlinkComponent(componentType, componentName, opts); err != nil {
				log.Fatal("Failed to unlink component:", err)
			}
		},
		func(targetFilter string, force bool, allProfiles bool) {
			opts := services.UnlinkOptions{TargetFilter: targetFilter, Force: force, AllProfiles: allProfiles}
			if err := linkService.UnlinkAll(opts); err != nil {
				log.Fatal("Failed to unlink all components:", err)
			}
		},
		func(targetFilter string, force bool, allProfiles bool, profile string) {
			opts := services.UnlinkOptions{TargetFilter: targetFilter, Force: force, AllProfiles: allProfiles, Profile: profile}
			if err := linkService.UnlinkAll(opts); err != nil {
				log.Fatal("Failed to unlink all components:", err)
			}
		},
		func(componentType, targetFilter string, force bool) {
			opts := services.UnlinkOptions{TargetFilter: targetFilter, Force: force}
			if err := linkService.UnlinkByType(componentType, opts); err != nil {
				log.Fatal("Failed to unlink components:", err)
			}
		},
		func(componentType, targetFilter string, force bool, profile string) {
			opts := services.UnlinkOptions{TargetFilter: targetFilter, Force: force, Profile: profile}
			if err := linkService.UnlinkByType(componentType, opts); err != nil {
				log.Fatal("Failed to unlink components:", err)
			}
		},
		func(componentType, componentName, profile, source string) {
			opts := services.UninstallOptions{Profile: profile, Source: source}
			if err := uninstallService.UninstallComponent(componentType, componentName, opts); err != nil {
				log.Fatal("Failed to uninstall component:", err)
			}
		},
		func(repoURL string, force bool) {
			opts := services.UninstallOptions{Force: force}
			if err := uninstallService.UninstallAllFromSource(repoURL, opts); err != nil {
				log.Fatal("Failed to uninstall components:", err)
			}
		},
		func(profileFilter []string, activeOnly bool, typeFilter string) {
			opts := services.ListProfileOptions{ProfileFilter: profileFilter, ActiveOnly: activeOnly, TypeFilter: typeFilter}
			if err := profileService.ListProfiles(opts); err != nil {
				log.Fatal("Failed to list profiles:", err)
			}
		},
		func(profileName string) {
			if profileName == "" {
				profileName = a.resolveActiveProfile()
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
				profileName = a.resolveActiveProfile()
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
			opts := services.MaterializeOptions{Target: target, ProjectDir: projectDir, Profile: profile, Source: source, Force: force, DryRun: dryRun}
			if err := materializeService.MaterializeComponent(componentType, componentName, opts); err != nil {
				os.Exit(1)
			}
		},
		func(componentType, target, projectDir string, force, dryRun bool, profile string) {
			opts := services.MaterializeOptions{Target: target, ProjectDir: projectDir, Profile: profile, Force: force, DryRun: dryRun}
			if err := materializeService.MaterializeByType(componentType, opts); err != nil {
				os.Exit(1)
			}
		},
		func(target, projectDir string, force, dryRun bool, profile string) {
			opts := services.MaterializeOptions{Target: target, ProjectDir: projectDir, Profile: profile, Force: force, DryRun: dryRun}
			if err := materializeService.MaterializeAll(opts); err != nil {
				os.Exit(1)
			}
		},
		func(projectDir string) {
			opts := services.ListMaterializedOptions{ProjectDir: projectDir}
			if err := materializeService.ListMaterialized(opts); err != nil {
				log.Fatal("Failed to list materialized components:", err)
			}
		},
		func(componentType, componentName, target, projectDir, source string) {
			opts := services.MaterializeInfoOptions{Target: target, ProjectDir: projectDir, Source: source}
			if err := materializeService.ShowComponentInfo(componentType, componentName, opts); err != nil {
				os.Exit(1)
			}
		},
		func(target, projectDir string) {
			opts := services.MaterializeStatusOptions{Target: target, ProjectDir: projectDir}
			if err := materializeService.ShowStatus(opts); err != nil {
				os.Exit(1)
			}
		},
		func(target, projectDir, source string, force, dryRun bool) {
			opts := services.MaterializeUpdateOptions{Target: target, ProjectDir: projectDir, Source: source, Force: force, DryRun: dryRun}
			if err := materializeService.UpdateMaterialized(opts); err != nil {
				os.Exit(1)
			}
		},
		func(query string, limit int, jsonOutput bool) {
			opts := services.FindOptions{Limit: limit, JSON: jsonOutput}
			if err := findService.FindSkills(query, opts); err != nil {
				log.Fatal("Failed to search skills:", err)
			}
		},
	)

	cmd.Execute()
}

// resolveActiveProfile returns the active profile name, fataling if none is set.
func (a *App) resolveActiveProfile() string {
	lockService := locksvc.NewService(a.logger)
	pm, err := profiles.NewProfileManager(nil, lockService)
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
	return activeProfile
}

// newComponentLinker builds a ComponentLinker for the active profile (or base dir if none).
func (a *App) newComponentLinker() (*linker.ComponentLinker, error) {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}

	lockService := locksvc.NewService(a.logger)
	profileManager, err := profiles.NewProfileManager(nil, lockService)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile manager: %w", err)
	}

	activeProfile, err := profileManager.GetActiveProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to get active profile: %w", err)
	}

	if activeProfile != "" {
		profilesDir, err := paths.GetProfilesDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get profiles directory: %w", err)
		}
		agentsDir = filepath.Join(profilesDir, activeProfile)
	}

	targets, err := config.DetectAllTargets()
	if err != nil {
		return nil, fmt.Errorf("failed to detect targets: %w", err)
	}

	det := detector.NewRepositoryDetector()
	det.SetLogger(a.logger)

	return linker.NewComponentLinker(agentsDir, targets, det, nil)
}

// getTargetNames returns target names for error messages.
func getTargetNames(targets []config.Target) []string {
	names := make([]string, len(targets))
	for i, t := range targets {
		names[i] = t.GetName()
	}
	return names
}

// newComponentLinkerWithFilterAndProfile builds a ComponentLinker with an optional target
// filter and explicit profile override. This mirrors the logic previously in main.go and
// is kept here so the container owns all linker construction.
func (a *App) newComponentLinkerWithFilterAndProfile(targetFilter string, profile string) (*linker.ComponentLinker, error) {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}

	if profile != "" {
		profilesDir, err := paths.GetProfilesDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get profiles directory: %w", err)
		}

		profilePath := filepath.Join(profilesDir, profile)
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			lockService := locksvc.NewService(a.logger)
			pm, pmErr := profiles.NewProfileManager(nil, lockService)
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
			Name:     profile,
			BasePath: profilePath,
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
	} else {
		lockService := locksvc.NewService(a.logger)
		profileManager, err := profiles.NewProfileManager(nil, lockService)
		if err != nil {
			return nil, fmt.Errorf("failed to create profile manager: %w", err)
		}
		activeProfile, err := profileManager.GetActiveProfile()
		if err != nil {
			return nil, fmt.Errorf("failed to get active profile: %w", err)
		}
		if activeProfile != "" {
			profilesDir, err := paths.GetProfilesDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get profiles directory: %w", err)
			}
			agentsDir = filepath.Join(profilesDir, activeProfile)
		}
	}

	allTargets, err := config.DetectAllTargets()
	if err != nil {
		return nil, fmt.Errorf("failed to detect targets: %w", err)
	}

	var targets []config.Target
	if targetFilter == "" || targetFilter == "all" {
		targets = allTargets
	} else {
		for _, t := range allTargets {
			if t.GetName() == targetFilter {
				targets = append(targets, t)
				break
			}
		}
		if len(targets) == 0 {
			return nil, fmt.Errorf("target '%s' not found. Available targets: %v", targetFilter, getTargetNames(allTargets))
		}
	}

	det := detector.NewRepositoryDetector()
	det.SetLogger(a.logger)

	return linker.NewComponentLinker(agentsDir, targets, det, nil)
}
