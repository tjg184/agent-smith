package status

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/internal/materializer"
	"github.com/tjg184/agent-smith/internal/models"
	"github.com/tjg184/agent-smith/internal/updater"
	"github.com/tjg184/agent-smith/pkg/config"
	"github.com/tjg184/agent-smith/pkg/errors"
	"github.com/tjg184/agent-smith/pkg/paths"
	"github.com/tjg184/agent-smith/pkg/project"
	"github.com/tjg184/agent-smith/pkg/services"
)

type Deps struct {
	Logger interface {
		Debug(format string, args ...interface{})
		Error(format string, args ...interface{})
	}
	Formatter *formatter.Formatter
}

func ListMaterialized(d Deps, opts services.ListMaterializedOptions) error {
	projectRoot, err := resolveProjectRoot(opts.ProjectDir)
	if err != nil {
		return err
	}

	d.Formatter.Info("Materialized Components in %s:", projectRoot)
	d.Formatter.EmptyLine()

	foundAny := false

	for _, targetName := range config.GetAvailableTargets() {
		target, err := config.NewTargetForProject(targetName, projectRoot)
		if err != nil {
			d.Logger.Debug("Failed to create target %s: %v", targetName, err)
			continue
		}
		targetDir := target.GetProjectBaseDir(projectRoot)

		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			continue
		}

		metadata, err := project.LoadMaterializationMetadata(targetDir)
		if err != nil {
			d.Logger.Debug("Failed to load metadata for %s: %v", targetName, err)
			continue
		}

		totalComponents := len(metadata.Skills) + len(metadata.Agents) + len(metadata.Commands)

		if totalComponents == 0 {
			continue
		}

		foundAny = true

		targetLabel := fmt.Sprintf("%s (%s/)", target.GetDisplayName(), target.GetProjectDirName())
		green := color.New(color.FgGreen).SprintFunc()
		d.Formatter.Info("%s %s", green(formatter.SymbolSuccess), targetLabel)

		totalSkills := 0
		for _, components := range metadata.Skills {
			totalSkills += len(components)
		}
		if totalSkills > 0 {
			d.Formatter.Info("  Skills (%d):", totalSkills)
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
					d.Formatter.Info("    • %-30s (from %s)", displayName, sourceInfo)
				}
			}
		}

		totalAgents := 0
		for _, components := range metadata.Agents {
			totalAgents += len(components)
		}
		if totalAgents > 0 {
			d.Formatter.Info("  Agents (%d):", totalAgents)
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
					d.Formatter.Info("    • %-30s (from %s)", displayName, sourceInfo)
				}
			}
		}

		totalCommands := 0
		for _, components := range metadata.Commands {
			totalCommands += len(components)
		}
		if totalCommands > 0 {
			d.Formatter.Info("  Commands (%d):", totalCommands)
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
					d.Formatter.Info("    • %-30s (from %s)", displayName, sourceInfo)
				}
			}
		}

		d.Formatter.EmptyLine()
	}

	if !foundAny {
		yellow := color.New(color.FgYellow).SprintFunc()
		d.Formatter.Info("%s No components materialized yet", yellow(formatter.SymbolWarning))
		d.Formatter.EmptyLine()
		d.Formatter.Info("To materialize components:")
		d.Formatter.Info("  agent-smith materialize skill <name> --target opencode")
		d.Formatter.Info("  agent-smith materialize all --target claudecode")
		d.Formatter.Info("  agent-smith materialize agent <name> --target copilot")
		d.Formatter.Info("  agent-smith materialize skill <name> --target universal  # Target-agnostic storage")
	}

	return nil
}

func ShowComponentInfo(d Deps, componentType, componentName string, opts services.MaterializeInfoOptions) error {
	projectRoot, err := resolveProjectRoot(opts.ProjectDir)
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
			d.Logger.Debug("Failed to create target %s: %v", targetName, err)
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
			d.Logger.Debug("Failed to load metadata for %s: %v", targetName, err)
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
				d.Formatter.Info("%s Component '%s' not found in %s target", red(formatter.SymbolError), componentName, targetName)
			}
			continue
		}

		foundInAnyTarget = true
		meta := *foundMeta

		targetLabel := fmt.Sprintf("%s (%s/)", target.GetDisplayName(), target.GetProjectDirName())

		d.Formatter.EmptyLine()
		d.Formatter.Info("%s Provenance Information - %s", green(formatter.SymbolSuccess), bold(targetLabel))
		d.Formatter.EmptyLine()

		d.Formatter.Info("  %s: %s", cyan("Component"), componentName)
		d.Formatter.Info("  %s: %s", cyan("Type"), componentType)
		d.Formatter.EmptyLine()

		d.Formatter.Info("  %s", bold("Source Information:"))
		d.Formatter.Info("    %s: %s", cyan("Repository"), meta.Source)
		d.Formatter.Info("    %s: %s", cyan("Source Type"), meta.SourceType)
		if meta.SourceProfile != "" {
			d.Formatter.Info("    %s: %s", cyan("Profile"), meta.SourceProfile)
		}
		d.Formatter.Info("    %s: %s", cyan("Commit Hash"), meta.CommitHash)
		d.Formatter.Info("    %s: %s", cyan("Original Path"), meta.OriginalPath)
		d.Formatter.EmptyLine()

		d.Formatter.Info("  %s", bold("Materialization:"))
		d.Formatter.Info("    %s: %s", cyan("Materialized At"), meta.MaterializedAt)
		d.Formatter.Info("    %s: %s", cyan("Target Directory"), targetDir)
		d.Formatter.EmptyLine()

		d.Formatter.Info("  %s", bold("Sync Status:"))
		d.Formatter.Info("    %s: %s", cyan("Source Hash"), meta.SourceHash)

		componentPath := filepath.Join(targetDir, componentType, componentName)
		actualCurrentHash, err := materializer.CalculateDirectoryHash(componentPath)
		if err != nil {
			d.Logger.Debug("Failed to calculate current hash: %v", err)
			actualCurrentHash = meta.CurrentHash
		}

		d.Formatter.Info("    %s: %s", cyan("Current Hash"), actualCurrentHash)

		if meta.SourceHash == actualCurrentHash {
			d.Formatter.Info("    %s: %s (component is unchanged)", cyan("Status"), green("In Sync"))
		} else {
			d.Formatter.Info("    %s: %s (component has been modified)", cyan("Status"), yellow("Modified"))
		}

		d.Formatter.EmptyLine()
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

func ShowStatus(d Deps, opts services.MaterializeStatusOptions) error {
	projectRoot, err := resolveProjectRoot(opts.ProjectDir)
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
			d.Logger.Debug("Failed to create target %s: %v", targetName, err)
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
