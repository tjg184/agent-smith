package sync

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
	"github.com/tjg184/agent-smith/pkg/config"
	"github.com/tjg184/agent-smith/pkg/errors"
	"github.com/tjg184/agent-smith/pkg/project"
	"github.com/tjg184/agent-smith/pkg/services"
)

// PostprocessContext carries the context passed to postprocessors.
// Mirrors the type in the parent package to avoid a circular import.
type PostprocessContext struct {
	ComponentType   string
	ComponentName   string
	FilesystemName  string
	Target          string
	TargetDir       string
	DestPath        string
	DryRun          bool
	Formatter       *formatter.Formatter
	SymlinkRegistry map[string]string
}

// PostprocessorRegistry is the minimal interface the sync package needs.
type PostprocessorRegistry interface {
	RunPostprocessors(ctx PostprocessContext) error
	RunCleanup(ctx PostprocessContext) error
}

// Deps is the set of dependencies the sync functions need from the parent service.
type Deps struct {
	Logger interface {
		Error(format string, args ...interface{})
	}
	Formatter *formatter.Formatter
	Registry  PostprocessorRegistry
	// GetSourceDir resolves the base directory and profile name for the given profile flag.
	GetSourceDir func(profile string) (baseDir string, sourceProfile string, err error)
}

type componentInfo struct {
	ComponentName string
	ComponentType string
	SourceUrl     string
}

func buildFilesystemNameMap(baseDir string) (map[string]componentInfo, error) {
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

func MaterializeComponent(d Deps, componentType, componentName string, opts services.MaterializeOptions) error {
	targetName := opts.Target
	if targetName == "" {
		targetName = config.GetTargetFromEnv()
	}
	if targetName == "" {
		fmt.Println(errors.NewMissingTargetFlagError("materialize " + componentType + " <name>").Format())
		return fmt.Errorf("target not specified")
	}

	if opts.DryRun {
		d.Formatter.Info("=== DRY RUN MODE ===")
		d.Formatter.Info("No changes will be made to the filesystem")
		d.Formatter.EmptyLine()
	}

	projectRoot, err := resolveProjectRoot(opts.ProjectDir)
	if err != nil {
		d.Logger.Error("Failed to determine project root: %v", err)
		return err
	}

	baseDir, sourceProfile, err := d.GetSourceDir(opts.Profile)
	if err != nil {
		d.Logger.Error("Failed to get source directory: %v", err)
		return err
	}

	var lockEntry *models.ComponentEntry

	if opts.Source != "" {
		var loadErr error
		lockEntry, loadErr = metadataPkg.LoadLockFileEntryBySource(baseDir, componentType, componentName, opts.Source)
		if loadErr != nil {
			d.Logger.Error("Failed to load component metadata from source %s: %v", opts.Source, loadErr)
			return fmt.Errorf("failed to load component metadata from source %s: %w", opts.Source, loadErr)
		}
	} else {
		var loadErr error
		lockEntry, loadErr = metadataPkg.LoadLockFileEntry(baseDir, componentType, componentName)
		if loadErr != nil {
			if containsStr(loadErr.Error(), "found in multiple sources") {
				sources, findErr := metadataPkg.FindComponentSources(baseDir, componentType, componentName)
				if findErr == nil && len(sources) > 0 {
					fmt.Println(errors.NewAmbiguousComponentError(componentType, componentName, sources).Format())
				}
				return fmt.Errorf("ambiguous component name")
			}
			d.Logger.Error("Failed to load component metadata: %v", loadErr)
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
				d.Formatter.Info("⊘ Would skip %s '%s' to %s (already exists and identical)", componentType, componentName, tgt)
			} else {
				d.Formatter.Info("⊘ Skipped %s '%s' to %s (already exists and identical)", componentType, componentName, tgt)
			}
			skipCount++
			continue
		}

		alreadyExists := recordedEntry != nil

		if alreadyExists {
			if !opts.Force {
				if opts.DryRun {
					d.Formatter.Info("⚠ Would fail: Component '%s' already exists in %s (use --force)", componentName, tgt)
					continue
				}
				return fmt.Errorf("component '%s' already exists in %s (use --force to overwrite)", componentName, tgt)
			}

			if opts.DryRun {
				d.Formatter.Info("⚠ Would overwrite %s '%s' in %s (--force)", componentType, componentName, tgt)
			} else {
				d.Formatter.Info("⚠ Overwriting %s '%s' in %s (--force)", componentType, componentName, tgt)

				cleanupCtx := PostprocessContext{
					ComponentType:  componentType,
					ComponentName:  componentName,
					FilesystemName: filesystemName,
					Target:         tgt,
					TargetDir:      targetDir,
					DestPath:       destPath,
					DryRun:         false,
					Formatter:      d.Formatter,
				}
				d.Registry.RunCleanup(cleanupCtx)

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
			d.Formatter.Info("%s Would materialize %s '%s' to %s", formatter.SymbolSuccess, componentType, componentName, tgt)
			d.Formatter.Info("  Source:      %s", componentSourceDir)
			if sourceProfile != "" {
				d.Formatter.Info("  From Profile: %s", sourceProfile)
			}
			d.Formatter.Info("  Destination: %s", destPath)
			if lockEntry.CommitHash != "" && len(lockEntry.CommitHash) >= 8 {
				d.Formatter.Info("  Provenance:  %s @ %s", lockEntry.SourceUrl, lockEntry.CommitHash[:8])
			} else if lockEntry.SourceUrl != "" {
				d.Formatter.Info("  Provenance:  %s", lockEntry.SourceUrl)
			}

			postprocessCtx := PostprocessContext{
				ComponentType:   componentType,
				ComponentName:   componentName,
				FilesystemName:  filesystemName,
				Target:          tgt,
				TargetDir:       targetDir,
				DestPath:        destPath,
				DryRun:          true,
				Formatter:       d.Formatter,
				SymlinkRegistry: symlinkRegistry,
			}
			if err := d.Registry.RunPostprocessors(postprocessCtx); err != nil {
				d.Formatter.WarningMsg("Postprocessor dry-run warning: %v", err)
			}

			successCount++
		} else {
			structureCreated, err := project.EnsureComponentDirectory(targetDir, componentType)
			if err != nil {
				return fmt.Errorf("failed to create target structure: %w", err)
			}
			if structureCreated {
				d.Formatter.Info("%s Created directory: %s/%s/", formatter.SymbolSuccess, targetDir, componentType)
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
				Formatter:       d.Formatter,
				SymlinkRegistry: symlinkRegistry,
			}
			if err := d.Registry.RunPostprocessors(postprocessCtx); err != nil {
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
				d.Formatter.Info("%s Materialized %s '%s' as '%s' to %s", formatter.SymbolSuccess, componentType, componentName, filesystemName, tgt)
			} else {
				d.Formatter.Info("%s Materialized %s '%s' to %s", formatter.SymbolSuccess, componentType, componentName, tgt)
			}
			d.Formatter.Info("  Source:      %s", componentSourceDir)
			if sourceProfile != "" {
				d.Formatter.Info("  From Profile: %s", sourceProfile)
			}
			d.Formatter.Info("  Destination: %s", destPath)
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

func MaterializeAll(d Deps, opts services.MaterializeOptions) error {
	baseDir, sourceProfile, err := d.GetSourceDir(opts.Profile)
	if err != nil {
		d.Logger.Error("Failed to get source directory: %v", err)
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
		d.Formatter.Info("No components found to materialize")
		if sourceProfile != "" {
			d.Formatter.Info("  Source: profile '%s'", sourceProfile)
		} else {
			d.Formatter.Info("  Source: ~/.agent-smith/")
		}
		return nil
	}

	for _, comp := range components {
		compOpts := opts
		compOpts.Source = comp.Source
		if err := MaterializeComponent(d, comp.Type, comp.Name, compOpts); err != nil {
			d.Formatter.WarningMsg("Failed to materialize %s '%s': %v", comp.Type, comp.Name, err)
		}
	}

	return nil
}

func MaterializeByType(d Deps, componentType string, opts services.MaterializeOptions) error {
	validTypes := map[string]bool{
		"skills":   true,
		"agents":   true,
		"commands": true,
	}
	if !validTypes[componentType] {
		return fmt.Errorf("invalid component type: %s (must be skills, agents, or commands)", componentType)
	}

	baseDir, sourceProfile, err := d.GetSourceDir(opts.Profile)
	if err != nil {
		d.Logger.Error("Failed to get source directory: %v", err)
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
		d.Formatter.Info("No %s found to materialize", componentType)
		if sourceProfile != "" {
			d.Formatter.Info("  Source: profile '%s' (~/.agent-smith/profiles/%s/%s/)", sourceProfile, sourceProfile, componentType)
		} else {
			d.Formatter.Info("  Source: ~/.agent-smith/%s/", componentType)
		}
		return nil
	}

	successCount := 0
	failureCount := 0
	for _, comp := range components {
		compOpts := opts
		compOpts.Source = comp.Source
		if err := MaterializeComponent(d, comp.Type, comp.Name, compOpts); err != nil {
			d.Formatter.WarningMsg("Failed to materialize %s '%s': %v", comp.Type, comp.Name, err)
			failureCount++
		} else {
			successCount++
		}
	}

	if successCount > 0 || failureCount > 0 {
		d.Formatter.EmptyLine()
		if failureCount == 0 {
			green := color.New(color.FgGreen).SprintFunc()
			d.Formatter.Info("%s Successfully materialized %d %s", green("✓"), successCount, componentType)
		} else {
			yellow := color.New(color.FgYellow).SprintFunc()
			d.Formatter.Info("%s Materialized %d %s, %d failed", yellow("⚠"), successCount, componentType, failureCount)
		}
	}

	return nil
}

func UpdateMaterialized(d Deps, opts services.MaterializeUpdateOptions) error {
	projectRoot, err := resolveProjectRoot(opts.ProjectDir)
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
			d.Logger.Error("Failed to create target %s: %v", targetName, err)
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

		localBaseDir, _, _ := d.GetSourceDir(opts.Profile)

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
			var matchErr error
			if useFlatCopy {
				match, matchErr = materializer.FlatMdFilesMatch(localSourceDir, destDir)
			} else {
				match, matchErr = materializer.DirectoriesMatch(localSourceDir, destDir)
			}
			if matchErr != nil {
				fmt.Printf("  %s %s (error comparing directories: %v)\n", red("✗"), comp.Name, matchErr)
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
				Formatter:       d.Formatter,
				SymlinkRegistry: symlinkRegistry,
			}
			if err := d.Registry.RunPostprocessors(postprocessCtx); err != nil {
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

			// Read commit hash from the local lock file so metadata stays consistent with local state.
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
		d.Formatter.Info("%s No components materialized yet", yellow(formatter.SymbolWarning))
		d.Formatter.EmptyLine()
		d.Formatter.Info("To materialize components:")
		d.Formatter.Info("  agent-smith materialize skill <name> --target opencode")
		d.Formatter.Info("  agent-smith materialize all --target claudecode")
		d.Formatter.Info("  agent-smith materialize agent <name> --target copilot")
		d.Formatter.Info("  agent-smith materialize skill <name> --target universal  # Target-agnostic storage")
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

func resolveProjectRoot(projectDir string) (string, error) {
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

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
