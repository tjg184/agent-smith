package container

import (
	"log"
	"os"

	"github.com/tjg184/agent-smith/cmd"
	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/internal/linker"
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

	componentLinker, err := linker.Build(linker.BuildOptions{}, a.logger)
	if err != nil {
		log.Fatal("Failed to initialize component linker:", err)
	}

	installService := installsvc.NewService(profileManager, a.logger, a.formatter)
	updateService := updatesvc.NewService(a.logger, a.formatter)
	uninstallService := uninstallsvc.NewService(componentLinker, a.logger, a.formatter, profileManager)
	targetService := targetsvc.NewService(a.logger, a.formatter)
	statusService := statussvc.NewService(profileManager, a.logger, a.formatter)
	linkService := linksvc.NewService(profileManager, a.logger, a.formatter)
	profileService := profilesvc.NewService(profileManager, a.logger, a.formatter)
	materializeService := materializesvc.NewService(profileManager, a.logger, a.formatter)
	findService := findsvc.NewService(a.logger, a.formatter)

	cmd.Register(&cmd.Handlers{
		Install: cmd.InstallHandlers{
			AddSkill: func(repoURL, name, profile, installDir string, global bool) {
				opts := services.InstallOptions{Profile: profile, InstallDir: installDir, Global: global}
				if err := installService.InstallSkill(repoURL, name, opts); err != nil {
					log.Fatal("Failed to install skill:", err)
				}
			},
			AddAgent: func(repoURL, name, profile, installDir string, global bool) {
				opts := services.InstallOptions{Profile: profile, InstallDir: installDir, Global: global}
				if err := installService.InstallAgent(repoURL, name, opts); err != nil {
					log.Fatal("Failed to install agent:", err)
				}
			},
			AddCommand: func(repoURL, name, profile, installDir string, global bool) {
				opts := services.InstallOptions{Profile: profile, InstallDir: installDir, Global: global}
				if err := installService.InstallCommand(repoURL, name, opts); err != nil {
					log.Fatal("Failed to install command:", err)
				}
			},
			AddAll: func(repoURL, profile, installDir string, global bool) {
				opts := services.InstallOptions{Profile: profile, InstallDir: installDir, Global: global}
				if err := installService.InstallBulk(repoURL, opts); err != nil {
					log.Fatal("Failed to bulk install:", err)
				}
			},
		},
		Update: cmd.UpdateHandlers{
			Update: func(componentType, componentName, profile string) {
				opts := services.UpdateOptions{Profile: profile}
				if err := updateService.UpdateComponent(componentType, componentName, opts); err != nil {
					log.Fatal("Failed to update component:", err)
				}
			},
			UpdateAll: func(profile string) {
				opts := services.UpdateOptions{Profile: profile}
				if err := updateService.UpdateAll(opts); err != nil {
					log.Fatal("Failed to update all components:", err)
				}
			},
		},
		Link: cmd.LinkHandlers{
			Link: func(componentType, componentName, targetFilter, profile string) {
				opts := services.LinkOptions{TargetFilter: targetFilter, Profile: profile}
				if err := linkService.LinkComponent(componentType, componentName, opts); err != nil {
					log.Fatal("Failed to link component:", err)
				}
			},
			LinkAll: func(targetFilter, profile string, allProfiles bool) {
				opts := services.LinkOptions{TargetFilter: targetFilter, Profile: profile, AllProfiles: allProfiles}
				if err := linkService.LinkAll(opts); err != nil {
					log.Fatal("Failed to link all components:", err)
				}
			},
			LinkType: func(componentType, targetFilter, profile string) {
				opts := services.LinkOptions{TargetFilter: targetFilter, Profile: profile}
				if err := linkService.LinkByType(componentType, opts); err != nil {
					log.Fatal("Failed to link components:", err)
				}
			},
			AutoLink: func() {
				if err := linkService.AutoLinkRepositories(); err != nil {
					log.Fatal("Failed to auto-link repositories:", err)
				}
			},
			ListLinks: func() {
				if err := linkService.ListLinked(); err != nil {
					log.Fatal("Failed to list linked components:", err)
				}
			},
			LinkStatus: func(allProfiles bool, profileFilter []string, linkedOnly bool) {
				opts := services.LinkStatusOptions{AllProfiles: allProfiles, ProfileFilter: profileFilter, LinkedOnly: linkedOnly}
				if err := linkService.ShowStatus(opts); err != nil {
					log.Fatal("Failed to show link status:", err)
				}
			},
		},
		Unlink: cmd.UnlinkHandlers{
			Unlink: func(componentType, componentName, targetFilter string) {
				opts := services.UnlinkOptions{TargetFilter: targetFilter}
				if err := linkService.UnlinkComponent(componentType, componentName, opts); err != nil {
					log.Fatal("Failed to unlink component:", err)
				}
			},
			UnlinkWithProfile: func(componentType, componentName, targetFilter, profile string) {
				opts := services.UnlinkOptions{TargetFilter: targetFilter, Profile: profile}
				if err := linkService.UnlinkComponent(componentType, componentName, opts); err != nil {
					log.Fatal("Failed to unlink component:", err)
				}
			},
			UnlinkAll: func(targetFilter string, force bool, allProfiles bool) {
				opts := services.UnlinkOptions{TargetFilter: targetFilter, Force: force, AllProfiles: allProfiles}
				if err := linkService.UnlinkAll(opts); err != nil {
					log.Fatal("Failed to unlink all components:", err)
				}
			},
			UnlinkAllWithProfile: func(targetFilter string, force bool, allProfiles bool, profile string) {
				opts := services.UnlinkOptions{TargetFilter: targetFilter, Force: force, AllProfiles: allProfiles, Profile: profile}
				if err := linkService.UnlinkAll(opts); err != nil {
					log.Fatal("Failed to unlink all components:", err)
				}
			},
			UnlinkType: func(componentType, targetFilter string, force bool) {
				opts := services.UnlinkOptions{TargetFilter: targetFilter, Force: force}
				if err := linkService.UnlinkByType(componentType, opts); err != nil {
					log.Fatal("Failed to unlink components:", err)
				}
			},
			UnlinkTypeWithProfile: func(componentType, targetFilter string, force bool, profile string) {
				opts := services.UnlinkOptions{TargetFilter: targetFilter, Force: force, Profile: profile}
				if err := linkService.UnlinkByType(componentType, opts); err != nil {
					log.Fatal("Failed to unlink components:", err)
				}
			},
		},
		Uninstall: cmd.UninstallHandlers{
			Uninstall: func(componentType, componentName, profile, source string) {
				opts := services.UninstallOptions{Profile: profile, Source: source}
				if err := uninstallService.UninstallComponent(componentType, componentName, opts); err != nil {
					log.Fatal("Failed to uninstall component:", err)
				}
			},
			UninstallAll: func(repoURL string, force bool) {
				opts := services.UninstallOptions{Force: force}
				if err := uninstallService.UninstallAllFromSource(repoURL, opts); err != nil {
					log.Fatal("Failed to uninstall components:", err)
				}
			},
		},
		Profile: cmd.ProfileHandlers{
			List: func(profileFilter []string, activeOnly bool, typeFilter string) {
				opts := services.ListProfileOptions{ProfileFilter: profileFilter, ActiveOnly: activeOnly, TypeFilter: typeFilter}
				if err := profileService.ListProfiles(opts); err != nil {
					log.Fatal("Failed to list profiles:", err)
				}
			},
			Show: func(profileName string) {
				if profileName == "" {
					profileName = a.resolveActiveProfile()
				}
				if err := profileService.ShowProfile(profileName); err != nil {
					log.Fatal("Failed to show profile:", err)
				}
			},
			Create: func(profileName string) {
				if err := profileService.CreateProfile(profileName); err != nil {
					log.Fatal("Failed to create profile:", err)
				}
			},
			Delete: func(profileName string) {
				if err := profileService.DeleteProfile(profileName); err != nil {
					log.Fatal("Failed to delete profile:", err)
				}
			},
			Activate: func(profileName string) {
				if err := profileService.ActivateProfile(profileName); err != nil {
					log.Fatal("Failed to activate profile:", err)
				}
			},
			Deactivate: func() {
				if err := profileService.DeactivateProfile(); err != nil {
					log.Fatal("Failed to deactivate profile:", err)
				}
			},
			Add: func(componentType, profileName, componentName string) {
				if err := profileService.AddComponent(componentType, profileName, componentName); err != nil {
					log.Fatal("Failed to add component:", err)
				}
			},
			Copy: func(componentType, sourceProfile, targetProfile, componentName string) {
				if err := profileService.CopyComponent(sourceProfile, targetProfile, componentType, componentName); err != nil {
					log.Fatal("Failed to copy component:", err)
				}
			},
			Remove: func(componentType, profileName, componentName string) {
				if err := profileService.RemoveComponent(profileName, componentType, componentName); err != nil {
					log.Fatal("Failed to remove component:", err)
				}
			},
			CherryPick: func(targetProfile string, sourceProfiles []string) {
				if err := profileService.CherryPickComponents(targetProfile, sourceProfiles); err != nil {
					log.Fatal("Failed to cherry-pick components:", err)
				}
			},
			Share: func(profileName, outputFile string) {
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
			Rename: func(oldName, newName string) {
				if err := profileService.RenameProfile(oldName, newName); err != nil {
					log.Fatal("Failed to rename profile:", err)
				}
			},
		},
		Status: cmd.StatusHandlers{
			Status: func() {
				if err := statusService.ShowSystemStatus(); err != nil {
					log.Fatal("Failed to show system status:", err)
				}
			},
		},
		Target: cmd.TargetHandlers{
			Add: func(name, path string) {
				if err := targetService.AddCustomTarget(name, path); err != nil {
					log.Fatal("Failed to add custom target:", err)
				}
			},
			Remove: func(name string) {
				if err := targetService.RemoveCustomTarget(name); err != nil {
					log.Fatal("Failed to remove custom target:", err)
				}
			},
			List: func() {
				if err := targetService.ListTargets(); err != nil {
					log.Fatal("Failed to list targets:", err)
				}
			},
		},
		Materialize: cmd.MaterializeHandlers{
			Component: func(componentType, componentName, target, projectDir string, force, dryRun bool, profile, source string) {
				opts := services.MaterializeOptions{Target: target, ProjectDir: projectDir, Profile: profile, Source: source, Force: force, DryRun: dryRun}
				if err := materializeService.MaterializeComponent(componentType, componentName, opts); err != nil {
					os.Exit(1)
				}
			},
			Type: func(componentType, target, projectDir string, force, dryRun bool, profile string) {
				opts := services.MaterializeOptions{Target: target, ProjectDir: projectDir, Profile: profile, Force: force, DryRun: dryRun}
				if err := materializeService.MaterializeByType(componentType, opts); err != nil {
					os.Exit(1)
				}
			},
			All: func(target, projectDir string, force, dryRun bool, profile string) {
				opts := services.MaterializeOptions{Target: target, ProjectDir: projectDir, Profile: profile, Force: force, DryRun: dryRun}
				if err := materializeService.MaterializeAll(opts); err != nil {
					os.Exit(1)
				}
			},
			List: func(projectDir string) {
				opts := services.ListMaterializedOptions{ProjectDir: projectDir}
				if err := materializeService.ListMaterialized(opts); err != nil {
					log.Fatal("Failed to list materialized components:", err)
				}
			},
			Info: func(componentType, componentName, target, projectDir, source string) {
				opts := services.MaterializeInfoOptions{Target: target, ProjectDir: projectDir, Source: source}
				if err := materializeService.ShowComponentInfo(componentType, componentName, opts); err != nil {
					os.Exit(1)
				}
			},
			Status: func(target, projectDir string) {
				opts := services.MaterializeStatusOptions{Target: target, ProjectDir: projectDir}
				if err := materializeService.ShowStatus(opts); err != nil {
					os.Exit(1)
				}
			},
			Update: func(target, projectDir, source string, force, dryRun bool) {
				opts := services.MaterializeUpdateOptions{Target: target, ProjectDir: projectDir, Source: source, Force: force, DryRun: dryRun}
				if err := materializeService.UpdateMaterialized(opts); err != nil {
					os.Exit(1)
				}
			},
		},
		Find: cmd.FindHandlers{
			FindSkill: func(query string, limit int, jsonOutput bool) {
				opts := services.FindOptions{Limit: limit, JSON: jsonOutput}
				if err := findService.FindSkills(query, opts); err != nil {
					log.Fatal("Failed to search skills:", err)
				}
			},
		},
	})

	cmd.Execute()
}

// resolveActiveProfile returns the active profile name, fataling if none is set.
func (a *App) resolveActiveProfile() string {
	activeProfile, err := profiles.ResolveActiveProfile()
	if err != nil {
		log.Fatal("Failed to resolve active profile:", err)
	}
	if activeProfile == "" {
		log.Fatal("No profile specified and no active profile set")
	}
	return activeProfile
}
