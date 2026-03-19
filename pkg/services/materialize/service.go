package materialize

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/internal/materializer"
	metadataPkg "github.com/tjg184/agent-smith/internal/metadata"
	"github.com/tjg184/agent-smith/internal/models"
	"github.com/tjg184/agent-smith/internal/updater"
	"github.com/tjg184/agent-smith/pkg/config"
	"github.com/tjg184/agent-smith/pkg/errors"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/paths"
	"github.com/tjg184/agent-smith/pkg/profiles"
	"github.com/tjg184/agent-smith/pkg/project"
	"github.com/tjg184/agent-smith/pkg/services"
)

type Service struct {
	profileManager        *profiles.ProfileManager
	logger                *logger.Logger
	formatter             *formatter.Formatter
	postprocessorRegistry *PostprocessorRegistry
}

func NewService(pm *profiles.ProfileManager, logger *logger.Logger, formatter *formatter.Formatter) services.MaterializeService {
	return &Service{
		profileManager:        pm,
		logger:                logger,
		formatter:             formatter,
		postprocessorRegistry: NewPostprocessorRegistry(),
	}
}

func (s *Service) getSourceDir(profile string) (string, string, error) {
	baseDir, err := paths.GetAgentsDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to get agent-smith directory: %w", err)
	}

	if profile != "" {
		if profile == "base" {
			return baseDir, "", nil
		}

		profilesList, err := s.profileManager.ScanProfiles()
		if err != nil {
			return "", "", fmt.Errorf("failed to scan profiles: %w", err)
		}

		profileExists := false
		for _, p := range profilesList {
			if p.Name == profile {
				profileExists = true
				break
			}
		}

		if !profileExists {
			return "", "", fmt.Errorf("profile '%s' not found", profile)
		}

		profilesDir, err := paths.GetProfilesDir()
		if err != nil {
			return "", "", fmt.Errorf("failed to get profiles directory: %w", err)
		}
		return filepath.Join(profilesDir, profile), profile, nil
	}

	activeProfile, err := s.profileManager.GetActiveProfile()
	if err != nil {
		return "", "", fmt.Errorf("failed to check active profile: %w", err)
	}

	if activeProfile != "" {
		profilesDir, err := paths.GetProfilesDir()
		if err != nil {
			return "", "", fmt.Errorf("failed to get profiles directory: %w", err)
		}
		return filepath.Join(profilesDir, activeProfile), activeProfile, nil
	}

	return baseDir, "", nil
}

type componentInfo struct {
	ComponentName string
	ComponentType string
	SourceUrl     string
}

func (s *Service) buildFilesystemNameMap(baseDir string) (map[string]componentInfo, error) {
	lockFile, err := metadataPkg.LoadLockFile(baseDir)
	if err != nil {
		return nil, err
	}

	mapping := make(map[string]componentInfo)

	addEntries := func(componentType string, sourceMap map[string]map[string]models.ComponentEntry) {
		for sourceUrl, components := range sourceMap {
			for componentName, entry := range components {
				key := componentType + "/" + entry.FilesystemName
				mapping[key] = componentInfo{
					ComponentName: componentName,
					ComponentType: componentType,
					SourceUrl:     sourceUrl,
				}
			}
		}
	}

	addEntries("skills", lockFile.Skills)
	addEntries("agents", lockFile.Agents)
	addEntries("commands", lockFile.Commands)

	return mapping, nil
}

func (s *Service) MaterializeComponent(componentType, componentName string, opts services.MaterializeOptions) error {
	targetName := opts.Target
	if targetName == "" {
		targetName = config.GetTargetFromEnv()
	}
	if targetName == "" {
		fmt.Println(errors.NewMissingTargetFlagError("materialize " + componentType + " <name>").Format())
		return fmt.Errorf("target not specified")
	}

	if opts.DryRun {
		s.formatter.Info("=== DRY RUN MODE ===")
		s.formatter.Info("No changes will be made to the filesystem")
		s.formatter.EmptyLine()
	}

	projectRoot, err := s.getProjectRoot(opts.ProjectDir)
	if err != nil {
		s.logger.Error("Failed to determine project root: %v", err)
		return err
	}

	baseDir, sourceProfile, err := s.getSourceDir(opts.Profile)
	if err != nil {
		s.logger.Error("Failed to get source directory: %v", err)
		return err
	}

	var lockEntry *models.ComponentEntry

	if opts.Source != "" {
		var loadErr error
		lockEntry, loadErr = metadataPkg.LoadLockFileEntryBySource(baseDir, componentType, componentName, opts.Source)
		if loadErr != nil {
			s.logger.Error("Failed to load component metadata from source %s: %v", opts.Source, loadErr)
			return fmt.Errorf("failed to load component metadata from source %s: %w", opts.Source, loadErr)
		}
	} else {
		var loadErr error
		lockEntry, loadErr = metadataPkg.LoadLockFileEntry(baseDir, componentType, componentName)
		if loadErr != nil {
			if strings.Contains(loadErr.Error(), "found in multiple sources") {
				sources, findErr := metadataPkg.FindComponentSources(baseDir, componentType, componentName)
				if findErr == nil && len(sources) > 0 {
					fmt.Println(errors.NewAmbiguousComponentError(componentType, componentName, sources).Format())
				}
				return fmt.Errorf("ambiguous component name")
			}
			s.logger.Error("Failed to load component metadata: %v", loadErr)
			return fmt.Errorf("failed to load component metadata: %w", loadErr)
		}
	}

	dirName := componentName
	if lockEntry != nil && lockEntry.FilesystemName != "" {
		dirName = lockEntry.FilesystemName
	}

	componentSourceDir := filepath.Join(baseDir, componentType, dirName)
	if _, err := os.Stat(componentSourceDir); os.IsNotExist(err) {
		var sourcePath string
		if sourceProfile != "" {
			sourcePath = fmt.Sprintf("profile '%s' (%s/%s/)", sourceProfile, sourceProfile, componentType)
		} else {
			sourcePath = fmt.Sprintf("~/.agent-smith/%s/", componentType)
		}
		fmt.Println(errors.NewComponentNotInstalledError(componentType, componentName, sourcePath).Format())
		return fmt.Errorf("component not found")
	}

	sourceHash, err := materializer.CalculateDirectoryHash(componentSourceDir)
	if err != nil {
		return fmt.Errorf("failed to calculate source hash: %w", err)
	}

	var targets []string
	if targetName == "all" {
		targets = config.GetAvailableTargets()
	} else {
		targets = []string{targetName}
	}

	successCount := 0
	skipCount := 0

	symlinkRegistry := make(map[string]string)

	for _, tgt := range targets {
		target, err := config.NewTargetForProject(tgt, projectRoot)
		if err != nil {
			fmt.Println(errors.NewInvalidTargetError(tgt).Format())
			return fmt.Errorf("invalid target: %s", tgt)
		}
		targetDir := target.GetProjectBaseDir(projectRoot)

		matMetadata, err := project.LoadMaterializationMetadata(targetDir)
		if err != nil {
			return fmt.Errorf("failed to load materialization metadata: %w", err)
		}

		sourceUrl := ""
		if lockEntry != nil {
			sourceUrl = lockEntry.SourceUrl
		}

		// When the lock entry carries a category prefix (e.g. "kotlin/convert-groovy-kotlin"),
		// use it directly to preserve the source hierarchy. ResolveFilesystemName only knows
		// the leaf name and would silently drop the prefix.
		var filesystemName string
		if dirName != componentName {
			filesystemName = dirName
		} else {
			filesystemName = project.ResolveFilesystemName(filepath.Join(targetDir, componentType), componentType, componentName, sourceUrl, matMetadata)
		}

		// Agents and commands are expected as flat .md files directly in the component type dir
		// (e.g. .opencode/agents/architect.md), not wrapped in a subdirectory. This mirrors
		// what `link` produces and what editors actually load.
		useFlatCopy := componentType == "agents" || componentType == "commands"

		componentTypeDir := filepath.Join(targetDir, componentType)
		var destPath string
		if useFlatCopy {
			destPath = componentTypeDir
		} else {
			destPath = filepath.Join(componentTypeDir, filesystemName)
		}

		componentMap := project.GetMaterializationComponentMap(matMetadata, componentType)
		var recordedEntry *models.ComponentEntry
		if componentMap != nil {
			for _, components := range componentMap {
				if entry, exists := components[componentName]; exists {
					recordedEntry = &entry
					break
				}
			}
		}

		alreadyMaterialized := recordedEntry != nil && recordedEntry.SourceHash == sourceHash && !opts.Force

		if alreadyMaterialized {
			if opts.DryRun {
				s.formatter.Info("⊘ Would skip %s '%s' to %s (already exists and identical)", componentType, componentName, tgt)
			} else {
				s.formatter.Info("⊘ Skipped %s '%s' to %s (already exists and identical)", componentType, componentName, tgt)
			}
			skipCount++
			continue
		}

		alreadyExists := recordedEntry != nil

		if alreadyExists {
			if !opts.Force {
				if opts.DryRun {
					s.formatter.Info("⚠ Would fail: Component '%s' already exists in %s (use --force)", componentName, tgt)
					continue
				}
				return fmt.Errorf("component '%s' already exists in %s (use --force to overwrite)", componentName, tgt)
			}

			if opts.DryRun {
				s.formatter.Info("⚠ Would overwrite %s '%s' in %s (--force)", componentType, componentName, tgt)
			} else {
				s.formatter.Info("⚠ Overwriting %s '%s' in %s (--force)", componentType, componentName, tgt)

				cleanupCtx := PostprocessContext{
					ComponentType:  componentType,
					ComponentName:  componentName,
					FilesystemName: filesystemName,
					Target:         tgt,
					TargetDir:      targetDir,
					DestPath:       destPath,
					DryRun:         false,
					Formatter:      s.formatter,
				}
				s.postprocessorRegistry.RunCleanup(cleanupCtx)

				if useFlatCopy {
					if err := materializer.RemoveFlatMdFiles(componentSourceDir, destPath); err != nil {
						return fmt.Errorf("failed to remove existing component files: %w", err)
					}
				} else if err := os.RemoveAll(destPath); err != nil {
					return fmt.Errorf("failed to remove existing component: %w", err)
				}
			}
		}

		if opts.DryRun {
			s.formatter.Info("%s Would materialize %s '%s' to %s", formatter.SymbolSuccess, componentType, componentName, tgt)
			s.formatter.Info("  Source:      %s", componentSourceDir)
			if sourceProfile != "" {
				s.formatter.Info("  From Profile: %s", sourceProfile)
			}
			s.formatter.Info("  Destination: %s", destPath)
			if lockEntry.CommitHash != "" && len(lockEntry.CommitHash) >= 8 {
				s.formatter.Info("  Provenance:  %s @ %s", lockEntry.SourceUrl, lockEntry.CommitHash[:8])
			} else if lockEntry.SourceUrl != "" {
				s.formatter.Info("  Provenance:  %s", lockEntry.SourceUrl)
			}

			postprocessCtx := PostprocessContext{
				ComponentType:   componentType,
				ComponentName:   componentName,
				FilesystemName:  filesystemName,
				Target:          tgt,
				TargetDir:       targetDir,
				DestPath:        destPath,
				DryRun:          true,
				Formatter:       s.formatter,
				SymlinkRegistry: symlinkRegistry,
			}
			if err := s.postprocessorRegistry.RunPostprocessors(postprocessCtx); err != nil {
				s.formatter.WarningMsg("Postprocessor dry-run warning: %v", err)
			}

			successCount++
		} else {
			structureCreated, err := project.EnsureComponentDirectory(targetDir, componentType)
			if err != nil {
				return fmt.Errorf("failed to create target structure: %w", err)
			}
			if structureCreated {
				s.formatter.Info("%s Created directory: %s/%s/", formatter.SymbolSuccess, targetDir, componentType)
			}

			if useFlatCopy {
				if err := materializer.CopyFlatMdFiles(componentSourceDir, destPath); err != nil {
					return fmt.Errorf("failed to copy component: %w", err)
				}
			} else if err := materializer.CopyDirectory(componentSourceDir, destPath); err != nil {
				return fmt.Errorf("failed to copy component: %w", err)
			}

			postprocessCtx := PostprocessContext{
				ComponentType:   componentType,
				ComponentName:   componentName,
				FilesystemName:  filesystemName,
				Target:          tgt,
				TargetDir:       targetDir,
				DestPath:        destPath,
				DryRun:          false,
				Formatter:       s.formatter,
				SymlinkRegistry: symlinkRegistry,
			}
			if err := s.postprocessorRegistry.RunPostprocessors(postprocessCtx); err != nil {
				return fmt.Errorf("postprocessing failed: %w", err)
			}

			project.AddMaterializationEntry(
				matMetadata,
				componentType,
				componentName,
				lockEntry.SourceUrl,
				lockEntry.SourceType,
				sourceProfile,
				lockEntry.CommitHash,
				lockEntry.OriginalPath,
				sourceHash,
				sourceHash,
				filesystemName,
			)

			if err := project.SaveMaterializationMetadata(targetDir, matMetadata); err != nil {
				return fmt.Errorf("failed to save materialization metadata: %w", err)
			}

			if filesystemName != componentName {
				s.formatter.Info("%s Materialized %s '%s' as '%s' to %s", formatter.SymbolSuccess, componentType, componentName, filesystemName, tgt)
			} else {
				s.formatter.Info("%s Materialized %s '%s' to %s", formatter.SymbolSuccess, componentType, componentName, tgt)
			}
			s.formatter.Info("  Source:      %s", componentSourceDir)
			if sourceProfile != "" {
				s.formatter.Info("  From Profile: %s", sourceProfile)
			}
			s.formatter.Info("  Destination: %s", destPath)
			successCount++
		}
	}

	if len(targets) > 0 {
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Println()
		if successCount > 0 {
			fmt.Printf("%s %d component(s) materialized", green("✓"), successCount)
			if skipCount > 0 {
				fmt.Printf(", %d skipped", skipCount)
			}
			fmt.Println()
		}
	}

	return nil
}

func (s *Service) MaterializeAll(opts services.MaterializeOptions) error {
	baseDir, sourceProfile, err := s.getSourceDir(opts.Profile)
	if err != nil {
		s.logger.Error("Failed to get source directory: %v", err)
		return err
	}

	lockFile, err := metadataPkg.LoadLockFile(baseDir)
	if err != nil {
		return fmt.Errorf("failed to load lock file: %w", err)
	}

	var components []struct {
		Type   string
		Name   string
		Source string
	}

	typeMap := map[string]map[string]map[string]models.ComponentEntry{
		"skills":   lockFile.Skills,
		"agents":   lockFile.Agents,
		"commands": lockFile.Commands,
	}

	for _, componentType := range []string{"skills", "agents", "commands"} {
		for sourceURL, componentsByName := range typeMap[componentType] {
			for componentName := range componentsByName {
				components = append(components, struct {
					Type   string
					Name   string
					Source string
				}{componentType, componentName, sourceURL})
			}
		}
	}

	if len(components) == 0 {
		s.formatter.Info("No components found to materialize")
		if sourceProfile != "" {
			s.formatter.Info("  Source: profile '%s'", sourceProfile)
		} else {
			s.formatter.Info("  Source: ~/.agent-smith/")
		}
		return nil
	}

	for _, comp := range components {
		compOpts := opts
		compOpts.Source = comp.Source
		if err := s.MaterializeComponent(comp.Type, comp.Name, compOpts); err != nil {
			s.formatter.WarningMsg("Failed to materialize %s '%s': %v", comp.Type, comp.Name, err)
		}
	}

	return nil
}

func (s *Service) MaterializeByType(componentType string, opts services.MaterializeOptions) error {
	validTypes := map[string]bool{
		"skills":   true,
		"agents":   true,
		"commands": true,
	}
	if !validTypes[componentType] {
		return fmt.Errorf("invalid component type: %s (must be skills, agents, or commands)", componentType)
	}

	baseDir, sourceProfile, err := s.getSourceDir(opts.Profile)
	if err != nil {
		s.logger.Error("Failed to get source directory: %v", err)
		return err
	}

	lockFile, err := metadataPkg.LoadLockFile(baseDir)
	if err != nil {
		return fmt.Errorf("failed to load lock file: %w", err)
	}

	typeMap := map[string]map[string]map[string]models.ComponentEntry{
		"skills":   lockFile.Skills,
		"agents":   lockFile.Agents,
		"commands": lockFile.Commands,
	}

	var components []struct {
		Type   string
		Name   string
		Source string
	}

	for sourceURL, componentsByName := range typeMap[componentType] {
		for componentName := range componentsByName {
			components = append(components, struct {
				Type   string
				Name   string
				Source string
			}{componentType, componentName, sourceURL})
		}
	}

	if len(components) == 0 {
		s.formatter.Info("No %s found to materialize", componentType)
		if sourceProfile != "" {
			s.formatter.Info("  Source: profile '%s' (~/.agent-smith/profiles/%s/%s/)", sourceProfile, sourceProfile, componentType)
		} else {
			s.formatter.Info("  Source: ~/.agent-smith/%s/", componentType)
		}
		return nil
	}

	successCount := 0
	failureCount := 0
	for _, comp := range components {
		compOpts := opts
		compOpts.Source = comp.Source
		if err := s.MaterializeComponent(comp.Type, comp.Name, compOpts); err != nil {
			s.formatter.WarningMsg("Failed to materialize %s '%s': %v", comp.Type, comp.Name, err)
			failureCount++
		} else {
			successCount++
		}
	}

	if successCount > 0 || failureCount > 0 {
		s.formatter.EmptyLine()
		if failureCount == 0 {
			green := color.New(color.FgGreen).SprintFunc()
			s.formatter.Info("%s Successfully materialized %d %s", green("✓"), successCount, componentType)
		} else {
			yellow := color.New(color.FgYellow).SprintFunc()
			s.formatter.Info("%s Materialized %d %s, %d failed", yellow("⚠"), successCount, componentType, failureCount)
		}
	}

	return nil
}

func (s *Service) ListMaterialized(opts services.ListMaterializedOptions) error {
	projectRoot, err := s.getProjectRoot(opts.ProjectDir)
	if err != nil {
		return err
	}

	s.formatter.Info("Materialized Components in %s:", projectRoot)
	s.formatter.EmptyLine()

	foundAny := false

	for _, targetName := range config.GetAvailableTargets() {
		target, err := config.NewTargetForProject(targetName, projectRoot)
		if err != nil {
			s.logger.Debug("Failed to create target %s: %v", targetName, err)
			continue
		}
		targetDir := target.GetProjectBaseDir(projectRoot)

		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			continue
		}

		metadata, err := project.LoadMaterializationMetadata(targetDir)
		if err != nil {
			s.logger.Debug("Failed to load metadata for %s: %v", targetName, err)
			continue
		}

		totalComponents := len(metadata.Skills) + len(metadata.Agents) + len(metadata.Commands)

		if totalComponents == 0 {
			continue
		}

		foundAny = true

		targetLabel := fmt.Sprintf("%s (%s/)", target.GetDisplayName(), target.GetProjectDirName())
		green := color.New(color.FgGreen).SprintFunc()
		s.formatter.Info("%s %s", green(formatter.SymbolSuccess), targetLabel)

		totalSkills := 0
		for _, components := range metadata.Skills {
			totalSkills += len(components)
		}
		if totalSkills > 0 {
			s.formatter.Info("  Skills (%d):", totalSkills)
			for _, components := range metadata.Skills {
				for name, meta := range components {
					sourceInfo := meta.Source
					if meta.SourceProfile != "" {
						sourceInfo = fmt.Sprintf("%s (profile: %s)", meta.Source, meta.SourceProfile)
					}
					displayName := name
					if meta.FilesystemName != name {
						displayName = fmt.Sprintf("%s (as %s)", name, meta.FilesystemName)
					}
					s.formatter.Info("    • %-30s (from %s)", displayName, sourceInfo)
				}
			}
		}

		totalAgents := 0
		for _, components := range metadata.Agents {
			totalAgents += len(components)
		}
		if totalAgents > 0 {
			s.formatter.Info("  Agents (%d):", totalAgents)
			for _, components := range metadata.Agents {
				for name, meta := range components {
					sourceInfo := meta.Source
					if meta.SourceProfile != "" {
						sourceInfo = fmt.Sprintf("%s (profile: %s)", meta.Source, meta.SourceProfile)
					}
					displayName := name
					if meta.FilesystemName != name {
						displayName = fmt.Sprintf("%s (as %s)", name, meta.FilesystemName)
					}
					s.formatter.Info("    • %-30s (from %s)", displayName, sourceInfo)
				}
			}
		}

		totalCommands := 0
		for _, components := range metadata.Commands {
			totalCommands += len(components)
		}
		if totalCommands > 0 {
			s.formatter.Info("  Commands (%d):", totalCommands)
			for _, components := range metadata.Commands {
				for name, meta := range components {
					sourceInfo := meta.Source
					if meta.SourceProfile != "" {
						sourceInfo = fmt.Sprintf("%s (profile: %s)", meta.Source, meta.SourceProfile)
					}
					displayName := name
					if meta.FilesystemName != name {
						displayName = fmt.Sprintf("%s (as %s)", name, meta.FilesystemName)
					}
					s.formatter.Info("    • %-30s (from %s)", displayName, sourceInfo)
				}
			}
		}

		s.formatter.EmptyLine()
	}

	if !foundAny {
		yellow := color.New(color.FgYellow).SprintFunc()
		s.formatter.Info("%s No components materialized yet", yellow(formatter.SymbolWarning))
		s.formatter.EmptyLine()
		s.formatter.Info("To materialize components:")
		s.formatter.Info("  agent-smith materialize skill <name> --target opencode")
		s.formatter.Info("  agent-smith materialize all --target claudecode")
		s.formatter.Info("  agent-smith materialize agent <name> --target copilot")
		s.formatter.Info("  agent-smith materialize skill <name> --target universal  # Target-agnostic storage")
	}

	return nil
}

func (s *Service) ShowComponentInfo(componentType, componentName string, opts services.MaterializeInfoOptions) error {
	projectRoot, err := s.getProjectRoot(opts.ProjectDir)
	if err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	foundInAnyTarget := false

	var targetsToCheck []string
	if opts.Target != "" {
		targetsToCheck = []string{opts.Target}
	} else {
		targetsToCheck = config.GetAvailableTargets()
	}

	for _, targetName := range targetsToCheck {
		target, err := config.NewTargetForProject(targetName, projectRoot)
		if err != nil {
			s.logger.Debug("Failed to create target %s: %v", targetName, err)
			continue
		}
		targetDir := target.GetProjectBaseDir(projectRoot)

		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			if opts.Target != "" {
				fmt.Println(errors.NewTargetDirectoryNotFoundError(target).Format())
				return fmt.Errorf("target directory not found")
			}
			continue
		}

		metadata, err := project.LoadMaterializationMetadata(targetDir)
		if err != nil {
			s.logger.Debug("Failed to load metadata for %s: %v", targetName, err)
			continue
		}

		componentMap := project.GetMaterializationComponentMap(metadata, componentType)
		if componentMap == nil {
			return fmt.Errorf("invalid component type: %s (must be skills, agents, or commands)", componentType)
		}

		var foundMeta *models.ComponentEntry
		for _, components := range componentMap {
			if meta, exists := components[componentName]; exists {
				foundMeta = &meta
				break
			}
		}

		if foundMeta == nil {
			if opts.Target != "" {
				s.formatter.Info("%s Component '%s' not found in %s target", red(formatter.SymbolError), componentName, targetName)
			}
			continue
		}

		foundInAnyTarget = true
		meta := *foundMeta

		targetLabel := fmt.Sprintf("%s (%s/)", target.GetDisplayName(), target.GetProjectDirName())

		s.formatter.EmptyLine()
		s.formatter.Info("%s Provenance Information - %s", green(formatter.SymbolSuccess), bold(targetLabel))
		s.formatter.EmptyLine()

		s.formatter.Info("  %s: %s", cyan("Component"), componentName)
		s.formatter.Info("  %s: %s", cyan("Type"), componentType)
		s.formatter.EmptyLine()

		s.formatter.Info("  %s", bold("Source Information:"))
		s.formatter.Info("    %s: %s", cyan("Repository"), meta.Source)
		s.formatter.Info("    %s: %s", cyan("Source Type"), meta.SourceType)
		if meta.SourceProfile != "" {
			s.formatter.Info("    %s: %s", cyan("Profile"), meta.SourceProfile)
		}
		s.formatter.Info("    %s: %s", cyan("Commit Hash"), meta.CommitHash)
		s.formatter.Info("    %s: %s", cyan("Original Path"), meta.OriginalPath)
		s.formatter.EmptyLine()

		s.formatter.Info("  %s", bold("Materialization:"))
		s.formatter.Info("    %s: %s", cyan("Materialized At"), meta.MaterializedAt)
		s.formatter.Info("    %s: %s", cyan("Target Directory"), targetDir)
		s.formatter.EmptyLine()

		s.formatter.Info("  %s", bold("Sync Status:"))
		s.formatter.Info("    %s: %s", cyan("Source Hash"), meta.SourceHash)

		componentPath := filepath.Join(targetDir, componentType, componentName)
		actualCurrentHash, err := materializer.CalculateDirectoryHash(componentPath)
		if err != nil {
			s.logger.Debug("Failed to calculate current hash: %v", err)
			actualCurrentHash = meta.CurrentHash
		}

		s.formatter.Info("    %s: %s", cyan("Current Hash"), actualCurrentHash)

		if meta.SourceHash == actualCurrentHash {
			s.formatter.Info("    %s: %s (component is unchanged)", cyan("Status"), green("In Sync"))
		} else {
			s.formatter.Info("    %s: %s (component has been modified)", cyan("Status"), yellow("Modified"))
		}

		s.formatter.EmptyLine()
	}

	if !foundInAnyTarget {
		if opts.Target != "" {
			if t, err := config.NewTarget(opts.Target); err == nil {
				fmt.Println(errors.NewTargetDirectoryNotFoundError(t).Format())
			} else {
				fmt.Println(errors.NewInvalidTargetError(opts.Target).Format())
			}
		} else {
			var availableComponents []string
			for _, targetName := range config.GetAvailableTargets() {
				target, err := config.NewTargetForProject(targetName, projectRoot)
				if err != nil {
					continue
				}
				targetDir := target.GetProjectBaseDir(projectRoot)
				if _, err := os.Stat(targetDir); os.IsNotExist(err) {
					continue
				}
				metadata, err := project.LoadMaterializationMetadata(targetDir)
				if err != nil {
					continue
				}
				componentMap := project.GetMaterializationComponentMap(metadata, componentType)
				if componentMap != nil {
					for _, components := range componentMap {
						for compName := range components {
							availableComponents = append(availableComponents, compName)
						}
					}
				}
			}
			fmt.Println(errors.NewComponentNotFoundInProjectError(componentType, componentName, availableComponents).Format())
		}
		return fmt.Errorf("component not found")
	}

	return nil
}

func (s *Service) ShowStatus(opts services.MaterializeStatusOptions) error {
	projectRoot, err := s.getProjectRoot(opts.ProjectDir)
	if err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	fmt.Printf("\n%s %s\n\n", bold("Project:"), projectRoot)

	var targetsToCheck []string
	if opts.Target != "" {
		targetsToCheck = []string{opts.Target}
	} else {
		targetsToCheck = config.GetAvailableTargets()
	}

	totalInSync := 0
	totalOutOfSync := 0
	totalMissing := 0
	foundAny := false

	for _, targetName := range targetsToCheck {
		target, err := config.NewTargetForProject(targetName, projectRoot)
		if err != nil {
			s.logger.Debug("Failed to create target %s: %v", targetName, err)
			continue
		}
		targetDir := target.GetProjectBaseDir(projectRoot)

		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			if opts.Target != "" {
				fmt.Println(errors.NewTargetDirectoryNotFoundError(target).Format())
				return fmt.Errorf("target directory not found")
			}
			continue
		}

		metadata, err := project.LoadMaterializationMetadata(targetDir)
		if err != nil {
			return fmt.Errorf("failed to load materialization metadata: %w", err)
		}

		components := project.GetAllMaterializedComponents(metadata)
		if len(components) == 0 {
			continue
		}

		foundAny = true

		targetLabel := fmt.Sprintf("%s (%s/)", target.GetDisplayName(), target.GetProjectDirName())
		fmt.Printf("%s %s\n\n", bold("Target:"), targetLabel)

		baseDir, _ := paths.GetAgentsDir()
		syncResults, err := project.CheckMultipleComponentsSyncStatusBatched(baseDir, components)
		if err != nil {
			return fmt.Errorf("failed to check sync status: %w", err)
		}

		componentsByType := make(map[string][]project.ComponentInfo)
		for _, comp := range components {
			componentsByType[comp.Type] = append(componentsByType[comp.Type], comp)
		}

		for _, componentType := range []string{"skills", "agents", "commands"} {
			comps := componentsByType[componentType]
			if len(comps) == 0 {
				continue
			}

			typeLabel := strings.Title(componentType)
			fmt.Printf("%s:\n", typeLabel)

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

				shortHash := comp.Metadata.CommitHash
				if len(shortHash) > 7 {
					shortHash = shortHash[:7]
				}

				switch result.Status {
				case project.SyncStatusInSync:
					fmt.Printf("  %s %s (in sync - %s)\n", green("✓"), comp.Name, shortHash)
					totalInSync++
				case project.SyncStatusOutOfSync:
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
		s.formatter.Info("%s No components materialized yet", yellow(formatter.SymbolWarning))
		s.formatter.EmptyLine()
		s.formatter.Info("To materialize components:")
		s.formatter.Info("  agent-smith materialize skill <name> --target opencode")
		s.formatter.Info("  agent-smith materialize all --target claudecode")
		s.formatter.Info("  agent-smith materialize agent <name> --target copilot")
		s.formatter.Info("  agent-smith materialize skill <name> --target universal  # Target-agnostic storage")
		return nil
	}

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

	return nil
}

func (s *Service) UpdateMaterialized(opts services.MaterializeUpdateOptions) error {
	projectRoot, err := s.getProjectRoot(opts.ProjectDir)
	if err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	if opts.DryRun {
		fmt.Printf("\n%s Previewing updates in: %s\n\n", bold("[DRY RUN]"), projectRoot)
	} else {
		fmt.Printf("\nUpdating materialized components in: %s\n\n", projectRoot)
	}

	var targetsToUpdate []string
	if opts.Target != "" {
		targetsToUpdate = []string{opts.Target}
	} else {
		targetsToUpdate = config.GetAvailableTargets()
	}

	totalUpdated := 0
	totalSkippedInSync := 0
	totalSkippedMissing := 0
	foundAny := false

	symlinkRegistry := make(map[string]string)

	for _, targetName := range targetsToUpdate {
		target, err := config.NewTargetForProject(targetName, projectRoot)
		if err != nil {
			s.logger.Debug("Failed to create target %s: %v", targetName, err)
			continue
		}
		targetDir := target.GetProjectBaseDir(projectRoot)

		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			if opts.Target != "" {
				fmt.Println(errors.NewTargetDirectoryNotFoundError(target).Format())
				return fmt.Errorf("target directory not found")
			}
			continue
		}

		metadata, err := project.LoadMaterializationMetadata(targetDir)
		if err != nil {
			return fmt.Errorf("failed to load materialization metadata: %w", err)
		}

		components := project.GetAllMaterializedComponents(metadata)
		if len(components) == 0 {
			continue
		}

		foundAny = true

		targetLabel := fmt.Sprintf("%s (%s/)", target.GetDisplayName(), target.GetProjectDirName())
		fmt.Printf("%s %s\n\n", bold("Target:"), targetLabel)

		localBaseDir, _, _ := s.getSourceDir(opts.Profile)

		for _, comp := range components {
			filesystemName := comp.Name
			if comp.Metadata.FilesystemName != "" {
				filesystemName = comp.Metadata.FilesystemName
			}

			useFlatCopy := comp.Type == "agents" || comp.Type == "commands"
			componentTypeDir := filepath.Join(targetDir, comp.Type)

			var destDir string
			if useFlatCopy {
				destDir = componentTypeDir
			} else {
				destDir = filepath.Join(componentTypeDir, filesystemName)
			}

			localSourceDir := filepath.Join(localBaseDir, comp.Type, filesystemName)
			localExists := false
			if _, statErr := os.Stat(localSourceDir); statErr == nil {
				localExists = true
			}

			if !localExists {
				fmt.Printf("  %s Skipped %s (source no longer installed)\n", yellow("⚠"), comp.Name)
				totalSkippedMissing++
				continue
			}

			var match bool
			var err error
			if useFlatCopy {
				match, err = materializer.FlatMdFilesMatch(localSourceDir, destDir)
			} else {
				match, err = materializer.DirectoriesMatch(localSourceDir, destDir)
			}
			if err != nil {
				fmt.Printf("  %s %s (error comparing directories: %v)\n", red("✗"), comp.Name, err)
				continue
			}

			if match && !opts.Force {
				fmt.Printf("  %s Skipped %s (already in sync)\n", green("⊘"), comp.Name)
				totalSkippedInSync++
				continue
			}

			if opts.DryRun {
				fmt.Printf("  %s Would update %s\n", green("→"), comp.Name)
				totalUpdated++
				continue
			}

			if useFlatCopy {
				if err := materializer.RemoveFlatMdFiles(localSourceDir, destDir); err != nil {
					fmt.Printf("  %s Failed to remove existing %s: %v\n", red("✗"), comp.Name, err)
					continue
				}
				if err := materializer.CopyFlatMdFiles(localSourceDir, destDir); err != nil {
					fmt.Printf("  %s Failed to copy %s: %v\n", red("✗"), comp.Name, err)
					continue
				}
			} else {
				if err := os.RemoveAll(destDir); err != nil {
					fmt.Printf("  %s Failed to remove existing %s: %v\n", red("✗"), comp.Name, err)
					continue
				}
				if err := materializer.CopyDirectory(localSourceDir, destDir); err != nil {
					fmt.Printf("  %s Failed to copy %s: %v\n", red("✗"), comp.Name, err)
					continue
				}
			}

			postprocessCtx := PostprocessContext{
				ComponentType:   comp.Type,
				ComponentName:   comp.Name,
				FilesystemName:  filesystemName,
				Target:          targetName,
				TargetDir:       targetDir,
				DestPath:        destDir,
				DryRun:          false,
				Formatter:       s.formatter,
				SymlinkRegistry: symlinkRegistry,
			}
			if err := s.postprocessorRegistry.RunPostprocessors(postprocessCtx); err != nil {
				fmt.Printf("  %s Postprocessing failed for %s: %v\n", red("✗"), comp.Name, err)
				continue
			}

			newSourceHash, err := materializer.CalculateDirectoryHash(localSourceDir)
			if err != nil {
				fmt.Printf("  %s Failed to calculate source hash for %s: %v\n", red("✗"), comp.Name, err)
				continue
			}

			// For flat copies the files land in the shared type dir alongside other components,
			// so we use the source hash as the current hash (consistent with MaterializeComponent).
			var newCurrentHash string
			if useFlatCopy {
				newCurrentHash = newSourceHash
			} else {
				newCurrentHash, err = materializer.CalculateDirectoryHash(destDir)
				if err != nil {
					fmt.Printf("  %s Failed to calculate current hash for %s: %v\n", red("✗"), comp.Name, err)
					continue
				}
			}

			// Read commit hash from the local lock file so metadata stays consistent with local state
			var newCommitHash string
			if lockEntry, lockErr := metadataPkg.LoadLockFileEntry(localBaseDir, comp.Type, comp.Name); lockErr == nil && lockEntry != nil {
				newCommitHash = lockEntry.CommitHash
			} else {
				newCommitHash = comp.Metadata.CommitHash
			}

			comp.Metadata.CommitHash = newCommitHash
			comp.Metadata.SourceHash = newSourceHash
			comp.Metadata.CurrentHash = newCurrentHash
			comp.Metadata.MaterializedAt = time.Now().Format(time.RFC3339)

			sourceURL := comp.Metadata.Source
			switch comp.Type {
			case "skills":
				if metadata.Skills[sourceURL] == nil {
					metadata.Skills[sourceURL] = make(map[string]models.ComponentEntry)
				}
				metadata.Skills[sourceURL][comp.Name] = comp.Metadata
			case "agents":
				if metadata.Agents[sourceURL] == nil {
					metadata.Agents[sourceURL] = make(map[string]models.ComponentEntry)
				}
				metadata.Agents[sourceURL][comp.Name] = comp.Metadata
			case "commands":
				if metadata.Commands[sourceURL] == nil {
					metadata.Commands[sourceURL] = make(map[string]models.ComponentEntry)
				}
				metadata.Commands[sourceURL][comp.Name] = comp.Metadata
			}

			fmt.Printf("  %s Updated %s\n", green("✓"), comp.Name)
			totalUpdated++
		}

		if !opts.DryRun && totalUpdated > 0 {
			if err := project.SaveMaterializationMetadata(targetDir, metadata); err != nil {
				return fmt.Errorf("failed to save materialization metadata: %w", err)
			}
		}

		fmt.Println()
	}

	if !foundAny {
		s.formatter.Info("%s No components materialized yet", yellow(formatter.SymbolWarning))
		s.formatter.EmptyLine()
		s.formatter.Info("To materialize components:")
		s.formatter.Info("  agent-smith materialize skill <name> --target opencode")
		s.formatter.Info("  agent-smith materialize all --target claudecode")
		s.formatter.Info("  agent-smith materialize agent <name> --target copilot")
		s.formatter.Info("  agent-smith materialize skill <name> --target universal  # Target-agnostic storage")
		return nil
	}

	fmt.Printf("%s: ", bold("Summary"))
	var parts []string
	if totalUpdated > 0 {
		if opts.DryRun {
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

	return nil
}

func (s *Service) getProjectRoot(projectDir string) (string, error) {
	if projectDir != "" {
		abs, err := filepath.Abs(projectDir)
		if err != nil {
			return "", fmt.Errorf("failed to resolve project directory: %w", err)
		}
		return abs, nil
	}

	root, err := project.FindProjectRoot()
	if err != nil {
		return "", fmt.Errorf("failed to find project root: %w", err)
	}
	return root, nil
}
