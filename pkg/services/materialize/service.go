package materialize

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/tjg184/agent-smith/internal/downloader"
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

// Service implements the MaterializeService interface
type Service struct {
	profileManager        *profiles.ProfileManager
	logger                *logger.Logger
	formatter             *formatter.Formatter
	postprocessorRegistry *PostprocessorRegistry
}

// NewService creates a new MaterializeService with the given dependencies
func NewService(pm *profiles.ProfileManager, logger *logger.Logger, formatter *formatter.Formatter) services.MaterializeService {
	return &Service{
		profileManager:        pm,
		logger:                logger,
		formatter:             formatter,
		postprocessorRegistry: NewPostprocessorRegistry(),
	}
}

// getSourceDir determines the source directory based on profile settings
func (s *Service) getSourceDir(profile string) (string, string, error) {
	baseDir, err := paths.GetAgentsDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to get agent-smith directory: %w", err)
	}

	if profile != "" {
		if profile == "base" {
			return baseDir, "", nil
		}

		// Validate profile exists
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

	// Check active profile
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

// componentInfo holds information about a component from the lock file
type componentInfo struct {
	ComponentName string
	ComponentType string
	SourceUrl     string
}

// buildFilesystemNameMap creates a mapping from filesystem names to component info
// This is needed because filesystemName can differ from componentName due to conflicts
func (s *Service) buildFilesystemNameMap(baseDir string) (map[string]componentInfo, error) {
	lockFilePath := filepath.Join(baseDir, ".component-lock.json")

	// Check if lock file exists
	if _, err := os.Stat(lockFilePath); os.IsNotExist(err) {
		// No lock file, return empty map
		return make(map[string]componentInfo), nil
	}

	lockData, err := os.ReadFile(lockFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	var lockFile models.ComponentLockFile
	if err := json.Unmarshal(lockData, &lockFile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lock file: %w", err)
	}

	mapping := make(map[string]componentInfo)

	// Helper to add entries from a component type map
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

	// Add all component types
	if lockFile.Skills != nil {
		addEntries("skills", lockFile.Skills)
	}
	if lockFile.Agents != nil {
		addEntries("agents", lockFile.Agents)
	}
	if lockFile.Commands != nil {
		addEntries("commands", lockFile.Commands)
	}

	return mapping, nil
}

// MaterializeComponent materializes a single component to a target
func (s *Service) MaterializeComponent(componentType, componentName string, opts services.MaterializeOptions) error {
	// Validate target
	targetName := opts.Target
	if targetName == "" {
		targetName = config.GetTargetFromEnv()
	}
	if targetName == "" {
		fmt.Println(errors.NewMissingTargetFlagError("materialize " + componentType + " <name>").Format())
		return fmt.Errorf("target not specified")
	}

	// Show dry-run header if enabled
	if opts.DryRun {
		s.formatter.Info("=== DRY RUN MODE ===")
		s.formatter.Info("No changes will be made to the filesystem")
		s.formatter.EmptyLine()
	}

	// Determine project root
	projectRoot, err := s.getProjectRoot(opts.ProjectDir)
	if err != nil {
		s.logger.Error("Failed to determine project root: %v", err)
		return err
	}

	// Get source directory
	baseDir, sourceProfile, err := s.getSourceDir(opts.Profile)
	if err != nil {
		s.logger.Error("Failed to get source directory: %v", err)
		return err
	}

	// Get lock file entry for provenance FIRST (we need the FilesystemName)
	var lockEntry *models.ComponentLockEntry

	if opts.Source != "" {
		// Use specific source if provided
		var loadErr error
		lockEntry, loadErr = metadataPkg.LoadLockFileEntryBySource(baseDir, componentType, componentName, opts.Source)
		if loadErr != nil {
			s.logger.Error("Failed to load component metadata from source %s: %v", opts.Source, loadErr)
			return fmt.Errorf("failed to load component metadata from source %s: %w", opts.Source, loadErr)
		}
	} else {
		// Try to load from any source
		var loadErr error
		lockEntry, loadErr = metadataPkg.LoadLockFileEntry(baseDir, componentType, componentName)
		if loadErr != nil {
			// Check if it's an ambiguous component error
			if strings.Contains(loadErr.Error(), "found in multiple sources") {
				// Extract source URLs from error and show nice disambiguation message
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

	// Determine the actual directory name to use
	// Use FilesystemName from lock file if available, otherwise use componentName
	dirName := componentName
	if lockEntry != nil && lockEntry.FilesystemName != "" {
		dirName = lockEntry.FilesystemName
	}

	// Get component source path using the filesystem name
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

	// Calculate source hash
	sourceHash, err := materializer.CalculateDirectoryHash(componentSourceDir)
	if err != nil {
		return fmt.Errorf("failed to calculate source hash: %w", err)
	}

	// Determine targets
	var targets []string
	if targetName == "all" {
		targets = []string{"opencode", "claudecode", "copilot"}
	} else {
		targets = []string{targetName}
	}

	// Materialize to each target
	successCount := 0
	skipCount := 0

	// Initialize symlink registry for tracking conflicts across all targets
	symlinkRegistry := make(map[string]string)

	for _, tgt := range targets {
		target, err := config.NewTargetForProject(tgt, projectRoot)
		if err != nil {
			fmt.Println(errors.NewInvalidTargetError(tgt).Format())
			return fmt.Errorf("invalid target: %s", tgt)
		}
		targetDir := target.GetProjectBaseDir(projectRoot)

		// Load materialization metadata to check for filesystem name conflicts
		matMetadata, err := project.LoadMaterializationMetadata(targetDir)
		if err != nil {
			return fmt.Errorf("failed to load materialization metadata: %w", err)
		}

		// Get source URL for idempotency check
		sourceUrl := ""
		if lockEntry != nil {
			sourceUrl = lockEntry.SourceUrl
		}

		// Resolve the actual filesystem name (handles conflicts with auto-suffixing)
		// Will reuse existing filesystem name if this exact component is already materialized
		filesystemName := project.ResolveFilesystemName(filepath.Join(targetDir, componentType), componentType, componentName, sourceUrl, matMetadata)
		destPath := filepath.Join(targetDir, componentType, filesystemName)

		// Check if exists
		if _, err := os.Stat(destPath); err == nil {
			match, err := materializer.DirectoriesMatch(componentSourceDir, destPath)
			if err != nil {
				return fmt.Errorf("failed to compare directories: %w", err)
			}
			if match {
				if opts.DryRun {
					s.formatter.Info("⊘ Would skip %s '%s' to %s (already exists and identical)", componentType, componentName, tgt)
				} else {
					s.formatter.Info("⊘ Skipped %s '%s' to %s (already exists and identical)", componentType, componentName, tgt)
				}
				skipCount++
				continue
			}

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

				// Run cleanup postprocessors before removing
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

				if err := os.RemoveAll(destPath); err != nil {
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

			// Run postprocessors in dry-run mode
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
			// Ensure component directory exists (only create the specific component type directory)
			structureCreated, err := project.EnsureComponentDirectory(targetDir, componentType)
			if err != nil {
				return fmt.Errorf("failed to create target structure: %w", err)
			}
			if structureCreated {
				s.formatter.Info("%s Created directory: %s/%s/", formatter.SymbolSuccess, targetDir, componentType)
			}

			// Copy component
			if err := materializer.CopyDirectory(componentSourceDir, destPath); err != nil {
				return fmt.Errorf("failed to copy component: %w", err)
			}

			// Run postprocessors after copying
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

			// Load/update metadata (already loaded earlier, reuse it)
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

			// Display success message
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

	// Print summary
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

// MaterializeAll materializes all components to a target
func (s *Service) MaterializeAll(opts services.MaterializeOptions) error {
	// Get source directory
	baseDir, sourceProfile, err := s.getSourceDir(opts.Profile)
	if err != nil {
		return err
	}

	// Build filesystem name mapping from lock file
	fsNameMap, err := s.buildFilesystemNameMap(baseDir)
	if err != nil {
		return fmt.Errorf("failed to build filesystem name mapping: %w", err)
	}

	// Get all components
	var components []struct {
		Type string
		Name string
	}

	for _, componentType := range []string{"skills", "agents", "commands"} {
		dir := filepath.Join(baseDir, componentType)
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("failed to read %s directory: %w", componentType, err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				// Map filesystem name to component name using lock file
				key := componentType + "/" + entry.Name()
				if info, exists := fsNameMap[key]; exists {
					// Use the component name from lock file
					components = append(components, struct {
						Type string
						Name string
					}{componentType, info.ComponentName})
				} else {
					// Directory not in lock file - warn and skip
					s.formatter.WarningMsg("Skipping untracked directory: %s/%s (not found in lock file)", componentType, entry.Name())
				}
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

	// Materialize each component
	for _, comp := range components {
		if err := s.MaterializeComponent(comp.Type, comp.Name, opts); err != nil {
			s.formatter.WarningMsg("Failed to materialize %s '%s': %v", comp.Type, comp.Name, err)
		}
	}

	return nil
}

// MaterializeByType materializes all components of a specific type to a target
func (s *Service) MaterializeByType(componentType string, opts services.MaterializeOptions) error {
	// Validate component type
	validTypes := map[string]bool{
		"skills":   true,
		"agents":   true,
		"commands": true,
	}
	if !validTypes[componentType] {
		return fmt.Errorf("invalid component type: %s (must be skills, agents, or commands)", componentType)
	}

	// Get source directory
	baseDir, sourceProfile, err := s.getSourceDir(opts.Profile)
	if err != nil {
		return err
	}

	// Build filesystem name mapping from lock file
	fsNameMap, err := s.buildFilesystemNameMap(baseDir)
	if err != nil {
		return fmt.Errorf("failed to build filesystem name mapping: %w", err)
	}

	// Get all components of the specified type
	var components []struct {
		Type string
		Name string
	}

	dir := filepath.Join(baseDir, componentType)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			s.formatter.Info("No %s found to materialize", componentType)
			if sourceProfile != "" {
				s.formatter.Info("  Source: profile '%s' (~/.agent-smith/profiles/%s/%s/)", sourceProfile, sourceProfile, componentType)
			} else {
				s.formatter.Info("  Source: ~/.agent-smith/%s/", componentType)
			}
			return nil
		}
		return fmt.Errorf("failed to read %s directory: %w", componentType, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Map filesystem name to component name using lock file
			key := componentType + "/" + entry.Name()
			if info, exists := fsNameMap[key]; exists {
				// Use the component name from lock file
				components = append(components, struct {
					Type string
					Name string
				}{componentType, info.ComponentName})
			} else {
				// Directory not in lock file - warn and skip
				s.formatter.WarningMsg("Skipping untracked directory: %s/%s (not found in lock file)", componentType, entry.Name())
			}
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

	// Materialize each component
	successCount := 0
	failureCount := 0
	for _, comp := range components {
		if err := s.MaterializeComponent(comp.Type, comp.Name, opts); err != nil {
			s.formatter.WarningMsg("Failed to materialize %s '%s': %v", comp.Type, comp.Name, err)
			failureCount++
		} else {
			successCount++
		}
	}

	// Display summary if we processed any components
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

// ListMaterialized lists all materialized components in a project
func (s *Service) ListMaterialized(opts services.ListMaterializedOptions) error {
	projectRoot, err := s.getProjectRoot(opts.ProjectDir)
	if err != nil {
		return err
	}

	// Display project information
	s.formatter.Info("Materialized Components in %s:", projectRoot)
	s.formatter.EmptyLine()

	// Track if any components were found
	foundAny := false

	// Check each target
	for _, targetName := range []string{"opencode", "claudecode", "copilot"} {
		target, err := config.NewTargetForProject(targetName, projectRoot)
		if err != nil {
			s.logger.Debug("Failed to create target %s: %v", targetName, err)
			continue
		}
		targetDir := target.GetProjectBaseDir(projectRoot)

		// Check if target directory exists
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			continue
		}

		// Load materialization metadata
		metadata, err := project.LoadMaterializationMetadata(targetDir)
		if err != nil {
			s.logger.Debug("Failed to load metadata for %s: %v", targetName, err)
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
		} else if targetName == "claudecode" {
			targetLabel = "Claude Code (.claude/)"
		} else if targetName == "copilot" {
			targetLabel = "GitHub Copilot (.github/)"
		} else {
			targetLabel = "Universal (.agents/)"
		}
		green := color.New(color.FgGreen).SprintFunc()
		s.formatter.Info("%s %s", green(formatter.SymbolSuccess), targetLabel)

		// Display skills
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

		// Display agents
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

		// Display commands
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

// ShowComponentInfo shows information about a materialized component
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

	// Track if we found the component in any target
	foundInAnyTarget := false

	// Determine which targets to check
	var targetsToCheck []string
	if opts.Target != "" {
		targetsToCheck = []string{opts.Target}
	} else {
		targetsToCheck = []string{"opencode", "claudecode", "copilot"}
	}

	// Check each target
	for _, targetName := range targetsToCheck {
		target, err := config.NewTargetForProject(targetName, projectRoot)
		if err != nil {
			s.logger.Debug("Failed to create target %s: %v", targetName, err)
			continue
		}
		targetDir := target.GetProjectBaseDir(projectRoot)

		// Check if target directory exists
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			if opts.Target != "" {
				// User specified a target that doesn't exist
				fmt.Println(errors.NewTargetDirectoryNotFoundError(targetName).Format())
				return fmt.Errorf("target directory not found")
			}
			continue
		}

		// Load materialization metadata
		metadata, err := project.LoadMaterializationMetadata(targetDir)
		if err != nil {
			s.logger.Debug("Failed to load metadata for %s: %v", targetName, err)
			continue
		}

		// Get the component map for the given type
		componentMap := project.GetMaterializationComponentMap(metadata, componentType)
		if componentMap == nil {
			return fmt.Errorf("invalid component type: %s (must be skills, agents, or commands)", componentType)
		}

		// Look up the component across all sources
		var foundMeta *project.MaterializedComponentMetadata
		for _, components := range componentMap {
			if meta, exists := components[componentName]; exists {
				foundMeta = &meta
				break
			}
		}

		if foundMeta == nil {
			if opts.Target != "" {
				// User specified a target but component not found
				s.formatter.Info("%s Component '%s' not found in %s target", red(formatter.SymbolError), componentName, targetName)
			}
			continue
		}

		foundInAnyTarget = true
		meta := *foundMeta

		// Display target header
		var targetLabel string
		if targetName == "opencode" {
			targetLabel = "OpenCode (.opencode/)"
		} else if targetName == "claudecode" {
			targetLabel = "Claude Code (.claude/)"
		} else if targetName == "copilot" {
			targetLabel = "GitHub Copilot (.github/)"
		} else {
			targetLabel = "Universal (.agents/)"
		}

		s.formatter.EmptyLine()
		s.formatter.Info("%s Provenance Information - %s", green(formatter.SymbolSuccess), bold(targetLabel))
		s.formatter.EmptyLine()

		// Display component information
		s.formatter.Info("  %s: %s", cyan("Component"), componentName)
		s.formatter.Info("  %s: %s", cyan("Type"), componentType)
		s.formatter.EmptyLine()

		// Display source information
		s.formatter.Info("  %s", bold("Source Information:"))
		s.formatter.Info("    %s: %s", cyan("Repository"), meta.Source)
		s.formatter.Info("    %s: %s", cyan("Source Type"), meta.SourceType)
		if meta.SourceProfile != "" {
			s.formatter.Info("    %s: %s", cyan("Profile"), meta.SourceProfile)
		}
		s.formatter.Info("    %s: %s", cyan("Commit Hash"), meta.CommitHash)
		s.formatter.Info("    %s: %s", cyan("Original Path"), meta.OriginalPath)
		s.formatter.EmptyLine()

		// Display materialization information
		s.formatter.Info("  %s", bold("Materialization:"))
		s.formatter.Info("    %s: %s", cyan("Materialized At"), meta.MaterializedAt)
		s.formatter.Info("    %s: %s", cyan("Target Directory"), targetDir)
		s.formatter.EmptyLine()

		// Display hash information for sync status
		s.formatter.Info("  %s", bold("Sync Status:"))
		s.formatter.Info("    %s: %s", cyan("Source Hash"), meta.SourceHash)

		// Recalculate current hash from the actual directory
		componentPath := filepath.Join(targetDir, componentType, componentName)
		actualCurrentHash, err := materializer.CalculateDirectoryHash(componentPath)
		if err != nil {
			s.logger.Debug("Failed to calculate current hash: %v", err)
			// Fall back to stored hash
			actualCurrentHash = meta.CurrentHash
		}

		s.formatter.Info("    %s: %s", cyan("Current Hash"), actualCurrentHash)

		// Check if hashes match
		if meta.SourceHash == actualCurrentHash {
			s.formatter.Info("    %s: %s (component is unchanged)", cyan("Status"), green("In Sync"))
		} else {
			s.formatter.Info("    %s: %s (component has been modified)", cyan("Status"), yellow("Modified"))
		}

		s.formatter.EmptyLine()
	}

	if !foundInAnyTarget {
		if opts.Target != "" {
			// Specific target was requested but component not found
			fmt.Println(errors.NewTargetDirectoryNotFoundError(opts.Target).Format())
		} else {
			// No target specified and component not found in any target
			// Collect available components from all targets
			var availableComponents []string
			for _, targetName := range []string{"opencode", "claudecode", "copilot"} {
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

// ShowStatus shows the materialization status of a project
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

	// Determine which targets to check
	var targetsToCheck []string
	if opts.Target != "" {
		targetsToCheck = []string{opts.Target}
	} else {
		targetsToCheck = []string{"opencode", "claudecode", "copilot"}
	}

	// Track overall statistics
	totalInSync := 0
	totalOutOfSync := 0
	totalMissing := 0
	foundAny := false

	// Check each target
	for _, targetName := range targetsToCheck {
		target, err := config.NewTargetForProject(targetName, projectRoot)
		if err != nil {
			s.logger.Debug("Failed to create target %s: %v", targetName, err)
			continue
		}
		targetDir := target.GetProjectBaseDir(projectRoot)

		// Check if target directory exists
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			if opts.Target != "" {
				// User specified a target that doesn't exist
				fmt.Println(errors.NewTargetDirectoryNotFoundError(targetName).Format())
				return fmt.Errorf("target directory not found")
			}
			continue
		}

		// Load materialization metadata
		metadata, err := project.LoadMaterializationMetadata(targetDir)
		if err != nil {
			return fmt.Errorf("failed to load materialization metadata: %w", err)
		}

		// Get all components
		components := project.GetAllMaterializedComponents(metadata)
		if len(components) == 0 {
			continue
		}

		foundAny = true

		// Display target header
		var targetLabel string
		if targetName == "opencode" {
			targetLabel = "OpenCode (.opencode/)"
		} else if targetName == "claudecode" {
			targetLabel = "Claude Code (.claude/)"
		} else if targetName == "copilot" {
			targetLabel = "GitHub Copilot (.github/)"
		} else {
			targetLabel = "Universal (.agents/)"
		}
		fmt.Printf("%s %s\n\n", bold("Target:"), targetLabel)

		// Use batched sync check for better performance (one clone per repo instead of per component)
		baseDir, _ := paths.GetAgentsDir()
		syncResults, err := project.CheckMultipleComponentsSyncStatusBatched(baseDir, components)
		if err != nil {
			return fmt.Errorf("failed to check sync status: %w", err)
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

	return nil
}

// UpdateMaterialized updates materialized components
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

	// Determine which targets to update
	var targetsToUpdate []string
	if opts.Target != "" {
		targetsToUpdate = []string{opts.Target}
	} else {
		targetsToUpdate = []string{"opencode", "claudecode", "copilot"}
	}

	// Track overall statistics
	totalUpdated := 0
	totalSkippedInSync := 0
	totalSkippedMissing := 0
	foundAny := false

	// Initialize symlink registry for tracking conflicts across all targets
	symlinkRegistry := make(map[string]string)

	// Update each target
	for _, targetName := range targetsToUpdate {
		target, err := config.NewTargetForProject(targetName, projectRoot)
		if err != nil {
			s.logger.Debug("Failed to create target %s: %v", targetName, err)
			continue
		}
		targetDir := target.GetProjectBaseDir(projectRoot)

		// Check if target directory exists
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			if opts.Target != "" {
				// User specified a target that doesn't exist
				fmt.Println(errors.NewTargetDirectoryNotFoundError(targetName).Format())
				return fmt.Errorf("target directory not found")
			}
			continue
		}

		// Load materialization metadata
		metadata, err := project.LoadMaterializationMetadata(targetDir)
		if err != nil {
			return fmt.Errorf("failed to load materialization metadata: %w", err)
		}

		// Get all components
		components := project.GetAllMaterializedComponents(metadata)
		if len(components) == 0 {
			continue
		}

		foundAny = true

		// Display target header
		var targetLabel string
		if targetName == "opencode" {
			targetLabel = "OpenCode (.opencode/)"
		} else if targetName == "claudecode" {
			targetLabel = "Claude Code (.claude/)"
		} else if targetName == "copilot" {
			targetLabel = "GitHub Copilot (.github/)"
		} else {
			targetLabel = "Universal (.agents/)"
		}
		fmt.Printf("%s %s\n\n", bold("Target:"), targetLabel)

		// Use batched sync check for better performance (one clone per repo instead of per component)
		baseDir, _ := paths.GetAgentsDir()
		syncResults, err := project.CheckMultipleComponentsSyncStatusBatched(baseDir, components)
		if err != nil {
			return fmt.Errorf("failed to check sync status: %w", err)
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
			if result.Status == project.SyncStatusInSync && !opts.Force {
				fmt.Printf("  %s Skipped %s (already in sync)\n", green("⊘"), comp.Name)
				totalSkippedInSync++
				continue
			}

			// Component needs updating
			if opts.DryRun {
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

			// Use FilesystemName from metadata if available (handles auto-suffixing)
			filesystemName := comp.Name
			if comp.Metadata.FilesystemName != "" {
				filesystemName = comp.Metadata.FilesystemName
			}
			destDir := filepath.Join(targetDir, comp.Type, filesystemName)

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

			// Run postprocessors after updating component
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
				ud := updater.NewUpdateDetectorWithBaseDir(baseDir)
				newCommitHash, _ = ud.GetCurrentRepoSHA(comp.Metadata.Source)
			}

			// Update metadata entry with new commit hash
			comp.Metadata.CommitHash = newCommitHash
			comp.Metadata.SourceHash = newSourceHash
			comp.Metadata.CurrentHash = newCurrentHash
			comp.Metadata.MaterializedAt = time.Now().Format(time.RFC3339)

			// Save updated metadata back to the nested metadata struct
			// Use the source URL from the metadata to determine the correct nested location
			sourceURL := comp.Metadata.Source
			switch comp.Type {
			case "skills":
				if metadata.Skills[sourceURL] == nil {
					metadata.Skills[sourceURL] = make(map[string]project.MaterializedComponentMetadata)
				}
				metadata.Skills[sourceURL][comp.Name] = comp.Metadata
			case "agents":
				if metadata.Agents[sourceURL] == nil {
					metadata.Agents[sourceURL] = make(map[string]project.MaterializedComponentMetadata)
				}
				metadata.Agents[sourceURL][comp.Name] = comp.Metadata
			case "commands":
				if metadata.Commands[sourceURL] == nil {
					metadata.Commands[sourceURL] = make(map[string]project.MaterializedComponentMetadata)
				}
				metadata.Commands[sourceURL][comp.Name] = comp.Metadata
			}

			fmt.Printf("  %s Updated %s\n", green("✓"), comp.Name)
			totalUpdated++
		}

		// Save metadata
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

	// Display summary
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

// getProjectRoot determines the project root directory
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
