package linkerSync

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tjg184/agent-smith/internal/detector"
	"github.com/tjg184/agent-smith/internal/fileutil"
	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/internal/linker/linkutil"
	"github.com/tjg184/agent-smith/internal/linker/profilepicker"
	metadataPkg "github.com/tjg184/agent-smith/internal/metadata"
	"github.com/tjg184/agent-smith/internal/models"
	"github.com/tjg184/agent-smith/pkg/colors"
	"github.com/tjg184/agent-smith/pkg/config"
	"github.com/tjg184/agent-smith/pkg/paths"
	"github.com/tjg184/agent-smith/pkg/styles"
)

// CreateSymlink creates a relative symlink from dst pointing to src.
func CreateSymlink(src, dst string) error {
	if info, err := os.Lstat(dst); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			os.Remove(dst)
		} else {
			// dst is a real directory — do not replace it with a symlink.
			// It may contain individually-linked leaf skills.
			return nil
		}
	}

	dstDir := filepath.Dir(dst)
	if realDir, err := filepath.EvalSymlinks(dstDir); err == nil {
		dstDir = realDir
	}

	relPath, err := filepath.Rel(dstDir, src)
	if err != nil {
		return fmt.Errorf("failed to create relative path: %w", err)
	}

	if err := os.Symlink(relPath, dst); err != nil {
		if runtime.GOOS == "windows" {
			return fileutil.CopyDirectoryContents(src, dst)
		}
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

// LinkComponent links a single component to all provided targets.
func LinkComponent(agentsDir string, targets []config.Target, f *formatter.Formatter, componentType, componentName string) error {
	return linkComponentInternal(agentsDir, targets, f, componentType, componentName, true)
}

func linkComponentInternal(agentsDir string, targets []config.Target, f *formatter.Formatter, componentType, componentName string, verbose bool) error {
	effectiveName := componentName
	srcDir := filepath.Join(agentsDir, componentType, effectiveName)
	selectedProfileName := ""

	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		matches, searchErr := profilepicker.SearchComponentInProfiles(componentType, componentName)
		if searchErr != nil {
			return fmt.Errorf("failed to search profiles: %w", searchErr)
		}

		if len(matches) == 0 {
			return fmt.Errorf("component %s/%s does not exist in any profile", componentType, componentName)
		}

		if len(matches) > 1 {
			profilePath, profileName, err := profilepicker.PromptProfileSelection(componentType, componentName, matches, os.Stdin, os.Stdout)
			if err != nil {
				return err
			}
			selectedFilesystemName := effectiveName
			for _, m := range matches {
				if m.ProfilePath == profilePath && componentType == "skills" && m.FilesystemName != "" {
					selectedFilesystemName = m.FilesystemName
					break
				}
			}
			effectiveName = selectedFilesystemName
			srcDir = filepath.Join(profilePath, componentType, effectiveName)
			selectedProfileName = profileName
		} else {
			if componentType == "skills" && matches[0].FilesystemName != "" {
				effectiveName = matches[0].FilesystemName
			}
			srcDir = filepath.Join(matches[0].ProfilePath, componentType, effectiveName)
			selectedProfileName = matches[0].ProfileName
			fmt.Printf("  %s Component found in profile: %s\n", colors.Muted("→"), selectedProfileName)
		}
	} else {
		selectedProfileName = linkutil.ProfileFromPath(srcDir)
	}

	_ = loadComponentMetadata(agentsDir, componentType, componentName)

	type linkResult struct {
		name    string
		path    string
		success bool
		errMsg  string
	}
	var linkResults []linkResult

	for _, target := range targets {
		targetName := target.GetName()

		componentDir, err := target.GetGlobalComponentDir(componentType)
		if err != nil {
			linkResults = append(linkResults, linkResult{
				name:    targetName,
				success: false,
				errMsg:  fmt.Sprintf("failed to get target component directory: %v", err),
			})
			continue
		}
		if componentType == "commands" || componentType == "agents" {
			if err := fileutil.CreateDirectoryWithPermissions(componentDir); err != nil {
				linkResults = append(linkResults, linkResult{
					name:    targetName,
					success: false,
					errMsg:  fmt.Sprintf("failed to create destination directory: %v", err),
				})
				continue
			}

			linked, err := linkFlatMdFiles(srcDir, componentDir)
			if err != nil || len(linked) == 0 {
				msg := fmt.Sprintf("failed to link: %v", err)
				if err == nil {
					msg = "no .md files found to link"
				}
				linkResults = append(linkResults, linkResult{
					name:    targetName,
					success: false,
					errMsg:  msg,
				})
				continue
			}

			linkResults = append(linkResults, linkResult{
				name:    targetName,
				path:    componentDir,
				success: true,
			})
			continue
		}

		dstDir := filepath.Join(componentDir, effectiveName)

		if err := fileutil.CreateDirectoryWithPermissions(filepath.Dir(dstDir)); err != nil {
			linkResults = append(linkResults, linkResult{
				name:    targetName,
				path:    dstDir,
				success: false,
				errMsg:  fmt.Sprintf("failed to create destination directory: %v", err),
			})
			continue
		}

		if err := CreateSymlink(srcDir, dstDir); err != nil {
			linkResults = append(linkResults, linkResult{
				name:    targetName,
				path:    dstDir,
				success: false,
				errMsg:  fmt.Sprintf("failed to link: %v", err),
			})
			continue
		}

		linkResults = append(linkResults, linkResult{
			name:    targetName,
			path:    dstDir,
			success: true,
		})
	}

	if len(linkResults) > 0 && verbose {
		hasSuccess := false
		for _, result := range linkResults {
			if result.success {
				hasSuccess = true
				break
			}
		}

		if hasSuccess {
			profileNote := styles.ProfileNoteFormat(selectedProfileName)
			fmt.Printf("%s%s\n", styles.InlineSuccessFormat("Linking", componentType, componentName), profileNote)

			for _, result := range linkResults {
				if result.success {
					fmt.Printf("%s\n", styles.IndentedDetailFormat(result.name, result.path))
				}
			}
		} else {
			fmt.Printf("%s\n", styles.InlineFailedFormat("Linking", componentType, componentName))

			for _, result := range linkResults {
				if !result.success {
					fmt.Printf("%s\n", styles.IndentedDetailFormat(result.name, result.errMsg))
				}
			}
			return fmt.Errorf("failed to link to target")
		}
	}

	if len(linkResults) > 0 {
		hasSuccess := false
		for _, result := range linkResults {
			if result.success {
				hasSuccess = true
				break
			}
		}
		if !hasSuccess {
			return fmt.Errorf("failed to link to any target")
		}
	}

	return nil
}

func LinkComponentsByType(agentsDir string, targets []config.Target, f *formatter.Formatter, componentType string) error {
	typeDir := filepath.Join(agentsDir, componentType)

	if _, err := os.Stat(typeDir); os.IsNotExist(err) {
		fmt.Printf("No %s found in %s\n", componentType, agentsDir)
		return nil
	}

	var successCount, failedCount int
	var failedComponents []string

	targetNames := make([]string, len(targets))
	for i, target := range targets {
		targetNames[i] = target.GetName()
	}
	targetList := strings.Join(targetNames, ", ")

	f.EmptyLine()
	fmt.Printf("%s\n", colors.InfoBold(fmt.Sprintf("Linking %s to: %s", componentType, targetList)))
	f.EmptyLine()

	componentNames, err := collectComponentNames(typeDir, componentType)
	if err != nil {
		return fmt.Errorf("failed to read %s directory: %w", componentType, err)
	}

	for _, componentName := range componentNames {
		fmt.Printf("Linking %s: %s... ", componentType, componentName)

		if err := linkComponentInternal(agentsDir, targets, f, componentType, componentName, false); err != nil {
			fmt.Printf("%s\n", colors.Error(formatter.SymbolError+" Failed"))
			fmt.Printf("  %s %v\n", colors.Muted("→"), err)
			failedCount++
			failedComponents = append(failedComponents, fmt.Sprintf("%s/%s", componentType, componentName))
		} else {
			fmt.Printf("%s\n", colors.Success(formatter.SymbolSuccess+" Done"))
			successCount++
		}
	}

	renderLinkSummary(f, successCount, failedCount, failedComponents)

	return nil
}

func LinkAllComponents(agentsDir string, targets []config.Target, f *formatter.Formatter) error {
	componentTypes := paths.GetComponentTypes()

	var successCount, failedCount int
	var failedComponents []string

	targetNames := make([]string, len(targets))
	for i, target := range targets {
		targetNames[i] = target.GetName()
	}
	targetList := strings.Join(targetNames, ", ")

	f.EmptyLine()
	fmt.Printf("%s\n", colors.InfoBold("Linking components to: "+targetList))
	f.EmptyLine()

	for _, componentType := range componentTypes {
		typeDir := filepath.Join(agentsDir, componentType)
		if _, err := os.Stat(typeDir); os.IsNotExist(err) {
			continue
		}

		componentNames, err := collectComponentNames(typeDir, componentType)
		if err != nil {
			f.WarningMsg("Failed to read %s directory: %v", componentType, err)
			continue
		}

		for _, componentName := range componentNames {
			fmt.Printf("Linking %s: %s... ", componentType, componentName)

			if err := linkComponentInternal(agentsDir, targets, f, componentType, componentName, false); err != nil {
				fmt.Printf("%s\n", colors.Error(formatter.SymbolError+" Failed"))
				fmt.Printf("  %s %v\n", colors.Muted("→"), err)
				failedCount++
				failedComponents = append(failedComponents, fmt.Sprintf("%s/%s", componentType, componentName))
			} else {
				fmt.Printf("%s\n", colors.Success(formatter.SymbolSuccess+" Done"))
				successCount++
			}
		}
	}

	renderLinkSummary(f, successCount, failedCount, failedComponents)

	return nil
}

func DetectAndLinkLocalRepositories(agentsDir string, targets []config.Target, f *formatter.Formatter, det *detector.RepositoryDetector) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	if !det.IsLocalPath(cwd) {
		return fmt.Errorf("current directory is not a git repository")
	}

	components, err := det.DetectComponentsInRepo(cwd)
	if err != nil {
		return fmt.Errorf("failed to detect components in repository: %w", err)
	}

	if len(components) == 0 {
		fmt.Println("No components detected in current repository")
		return nil
	}

	fmt.Printf("Detected %d components in current repository:\n", len(components))
	for _, component := range components {
		fmt.Printf("  - %s: %s (%s)\n", component.Type, component.Name, component.Path)
	}

	for _, component := range components {
		componentTypeStr := string(component.Type) + "s"
		componentPath := filepath.Join(cwd, component.Path)
		if info, err := os.Stat(componentPath); err == nil && !info.IsDir() {
			componentPath = filepath.Dir(componentPath)
		}

		tempLinkName := fmt.Sprintf("auto-detected-%s", component.Name)
		tempLinkPath := filepath.Join(agentsDir, componentTypeStr, tempLinkName)

		if err := fileutil.CreateDirectoryWithPermissions(filepath.Dir(tempLinkPath)); err != nil {
			fmt.Printf("Warning: failed to create directory for %s: %v\n", component.Name, err)
			continue
		}

		if err := CreateSymlink(componentPath, tempLinkPath); err != nil {
			fmt.Printf("Warning: failed to link component %s: %v\n", component.Name, err)
			continue
		}

		if err := LinkComponent(agentsDir, targets, f, componentTypeStr, tempLinkName); err != nil {
			fmt.Printf("Warning: failed to link %s to opencode: %v\n", component.Name, err)
			continue
		}

		fmt.Printf("✓ Automatically linked %s '%s' from current repository\n", component.Type, component.Name)
	}

	return nil
}

func loadComponentMetadata(agentsDir, componentType, componentName string) *models.ComponentEntry {
	entry, err := metadataPkg.LoadLockFileEntry(agentsDir, componentType, componentName)
	if err != nil {
		return nil
	}
	return entry
}

func renderLinkSummary(f *formatter.Formatter, successCount, failedCount int, failedComponents []string) {
	f.EmptyLine()

	table := formatter.NewBoxTable(f.Writer(), []string{"Status", "Count"})
	table.AddRow([]string{colors.Success(formatter.SymbolSuccess + " Success"), fmt.Sprintf("%d", successCount)})
	if failedCount > 0 {
		table.AddRow([]string{colors.Error(formatter.SymbolError + " Failed"), fmt.Sprintf("%d", failedCount)})
	}
	table.Render()

	if failedCount > 0 {
		f.EmptyLine()
		fmt.Println("Failed components:")
		for _, comp := range failedComponents {
			fmt.Printf("  • %s\n", comp)
		}
	}
}

// linkFlatMdFiles creates a flat symlink in targetBaseDir for each .md file in srcDir.
func linkFlatMdFiles(srcDir, targetBaseDir string) ([]string, error) {
	var linked []string

	if resolved, err := filepath.EvalSymlinks(srcDir); err == nil {
		srcDir = resolved
	}

	if err := os.MkdirAll(targetBaseDir, 0755); err != nil {
		return nil, err
	}

	err := filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		flatName := strings.ReplaceAll(rel, string(filepath.Separator), "-")
		dst := filepath.Join(targetBaseDir, flatName)

		if _, err := os.Lstat(dst); err == nil {
			if err := os.Remove(dst); err != nil {
				return err
			}
		}

		dstDir := targetBaseDir
		if realDir, err := filepath.EvalSymlinks(dstDir); err == nil {
			dstDir = realDir
		}

		relSymlink, err := filepath.Rel(dstDir, path)
		if err != nil {
			return err
		}

		if err := os.Symlink(relSymlink, dst); err != nil {
			return err
		}

		linked = append(linked, dst)
		return nil
	})

	return linked, err
}

// collectComponentNames returns the names to use when iterating a component type
// directory for bulk-link operations. For skills, it recursively finds all leaf
// skill directories (those containing SKILL.md) and returns their relative paths
// (e.g. "sdlc-pipeline/record-completion"). For other types, it returns the
// names of the top-level subdirectories.
func collectComponentNames(typeDir, componentType string) ([]string, error) {
	if componentType == "skills" {
		return collectLeafSkillNames(typeDir, "")
	}

	entries, err := os.ReadDir(typeDir)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// prefix is the accumulated relative path from typeDir (empty at the top level).
func collectLeafSkillNames(typeDir, prefix string) ([]string, error) {
	searchDir := typeDir
	if prefix != "" {
		searchDir = filepath.Join(typeDir, prefix)
	}

	entries, err := os.ReadDir(searchDir)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		relName := e.Name()
		if prefix != "" {
			relName = prefix + "/" + e.Name()
		}

		skillMd := filepath.Join(typeDir, relName, "SKILL.md")
		if _, err := os.Stat(skillMd); err == nil {
			names = append(names, relName)
			continue
		}

		// Not a leaf — recurse into it.
		nested, err := collectLeafSkillNames(typeDir, relName)
		if err != nil {
			return nil, err
		}
		names = append(names, nested...)
	}
	return names, nil
}
