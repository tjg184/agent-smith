package linker

import (
	"fmt"
	"os"

	"github.com/tjg184/agent-smith/internal/detector"
	"github.com/tjg184/agent-smith/internal/fileutil"
	"github.com/tjg184/agent-smith/internal/formatter"
	linkerDisplay "github.com/tjg184/agent-smith/internal/linker/display"
	"github.com/tjg184/agent-smith/internal/linker/profilepicker"
	linkerSync "github.com/tjg184/agent-smith/internal/linker/sync"
	linkerUnlink "github.com/tjg184/agent-smith/internal/linker/unlink"
	"github.com/tjg184/agent-smith/pkg/config"
)

type ComponentLinker struct {
	agentsDir      string
	targets        []config.Target
	detector       *detector.RepositoryDetector
	profileManager ProfileManager // Optional - can be nil
	formatter      *formatter.Formatter
}

// ProfileManager prevents circular dependencies between linker and profiles packages.
type ProfileManager interface {
	ScanProfiles() ([]*Profile, error)
	GetActiveProfile() (string, error)
}

// targetDisplayNames builds a name→display-name map from the linker's target list.
func (cl *ComponentLinker) targetDisplayNames() map[string]string {
	m := make(map[string]string, len(cl.targets))
	for _, t := range cl.targets {
		m[t.GetName()] = t.GetDisplayName()
	}
	return m
}

// ProfileMatch represents a profile that contains a specific component.
// This type is exposed from this package for backward compatibility;
// the implementation lives in the profilepicker sub-package.
type ProfileMatch struct {
	ProfileName string
	ProfilePath string
	IsActive    bool
	SourceUrl   string
}

// Profile must match the Profile struct from pkg/profiles/profiles.go
type Profile struct {
	Name        string
	BasePath    string
	HasAgents   bool
	HasSkills   bool
	HasCommands bool
}

func NewComponentLinker(agentsDir string, targets []config.Target, det *detector.RepositoryDetector, pm ProfileManager) (*ComponentLinker, error) {
	if agentsDir == "" {
		return nil, fmt.Errorf("agentsDir cannot be empty")
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("at least one target must be provided")
	}
	if det == nil {
		return nil, fmt.Errorf("detector cannot be nil")
	}

	for _, target := range targets {
		targetDir, err := target.GetGlobalBaseDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get target base directory: %w", err)
		}

		if err := fileutil.CreateDirectoryWithPermissions(targetDir); err != nil {
			return nil, fmt.Errorf("failed to create target directory: %w", err)
		}
	}

	return &ComponentLinker{
		agentsDir:      agentsDir,
		targets:        targets,
		detector:       det,
		profileManager: pm,
		formatter:      formatter.New(),
	}, nil
}

func (cl *ComponentLinker) SetFormatter(f *formatter.Formatter) {
	cl.formatter = f
}

// filterTargets returns all targets when targetFilter is empty or "all", otherwise returns only the matching target.
func (cl *ComponentLinker) filterTargets(targetFilter string) []config.Target {
	if targetFilter == "" || targetFilter == "all" {
		return cl.targets
	}

	filtered := make([]config.Target, 0)
	for _, target := range cl.targets {
		if target.GetName() == targetFilter {
			filtered = append(filtered, target)
		}
	}

	return filtered
}

func (cl *ComponentLinker) createSymlink(src, dst string) error {
	return linkerSync.CreateSymlink(src, dst)
}

// createJunction creates a Windows junction, falling back to directory copy.
func (cl *ComponentLinker) createJunction(src, dst string) error {
	return fileutil.CopyDirectoryContents(src, dst)
}

func (cl *ComponentLinker) copyDirectory(src, dst string) error {
	return fileutil.CopyDirectoryContents(src, dst)
}

func (cl *ComponentLinker) copyFile(src, dst string) error {
	return fileutil.CopyFile(src, dst)
}

func (cl *ComponentLinker) LinkComponent(componentType, componentName string) error {
	return linkerSync.LinkComponent(cl.agentsDir, cl.targets, cl.formatter, componentType, componentName)
}

func (cl *ComponentLinker) LinkComponentsByType(componentType string) error {
	return linkerSync.LinkComponentsByType(cl.agentsDir, cl.targets, cl.formatter, componentType)
}

func (cl *ComponentLinker) LinkAllComponents() error {
	return linkerSync.LinkAllComponents(cl.agentsDir, cl.targets, cl.formatter)
}

func (cl *ComponentLinker) DetectAndLinkLocalRepositories() error {
	return linkerSync.DetectAndLinkLocalRepositories(cl.agentsDir, cl.targets, cl.formatter, cl.detector)
}

// ListLinkedComponents lists all components linked to the configured targets
func (cl *ComponentLinker) ListLinkedComponents() error {
	return linkerDisplay.ListLinkedComponents(cl.agentsDir, cl.targets, cl.formatter)
}

// ShowLinkStatus displays a matrix view of components and their status across all targets
func (cl *ComponentLinker) ShowLinkStatus(linkedOnly bool) error {
	return linkerDisplay.ShowLinkStatus(cl.agentsDir, cl.targets, cl.formatter, linkedOnly)
}

// ShowAllProfilesLinkStatus displays link status for components across all profiles
// profileFilter can filter to specific profiles, or empty to show all
func (cl *ComponentLinker) ShowAllProfilesLinkStatus(profileFilter []string, linkedOnly bool) error {
	if cl.profileManager == nil {
		return fmt.Errorf("profile manager not available - this operation requires a profile manager")
	}
	adapter := &displayProfileManagerAdapter{pm: cl.profileManager}
	return linkerDisplay.ShowAllProfilesLinkStatus(cl.agentsDir, cl.targets, cl.formatter, adapter, profileFilter, linkedOnly)
}

// UnlinkComponent removes a linked component from configured targets
func (cl *ComponentLinker) UnlinkComponent(componentType, componentName, targetFilter string) error {
	return linkerUnlink.UnlinkComponent(cl.agentsDir, cl.targets, cl.formatter, componentType, componentName, targetFilter)
}

// UnlinkComponentFromDir removes a linked component using sourceDir as the component
// root instead of the linker's default agentsDir. Required when the component lives
// in a profile directory rather than the global agents directory.
func (cl *ComponentLinker) UnlinkComponentFromDir(sourceDir, componentType, componentName, targetFilter string) error {
	return linkerUnlink.UnlinkComponent(sourceDir, cl.targets, cl.formatter, componentType, componentName, targetFilter)
}

// UnlinkComponentsByType removes all linked components of a specific type from configured targets
func (cl *ComponentLinker) UnlinkComponentsByType(componentType, targetFilter string, force bool) error {
	return linkerUnlink.UnlinkComponentsByType(cl.agentsDir, cl.targets, cl.formatter, componentType, targetFilter, force)
}

// UnlinkAllComponents removes all linked components from configured targets
func (cl *ComponentLinker) UnlinkAllComponents(targetFilter string, force bool, allProfiles bool) error {
	return linkerUnlink.UnlinkAllComponents(cl.agentsDir, cl.targets, cl.formatter, targetFilter, force, allProfiles)
}

func (cl *ComponentLinker) searchComponentInProfiles(componentType, componentName string) ([]ProfileMatch, error) {
	matches, err := profilepicker.SearchComponentInProfiles(componentType, componentName)
	if err != nil {
		return nil, err
	}
	result := make([]ProfileMatch, len(matches))
	for i, m := range matches {
		result[i] = ProfileMatch{
			ProfileName: m.ProfileName,
			ProfilePath: m.ProfilePath,
			IsActive:    m.IsActive,
			SourceUrl:   m.SourceUrl,
		}
	}
	return result, nil
}

func (cl *ComponentLinker) promptProfileSelection(componentType, componentName string, matches []ProfileMatch) (string, string, error) {
	ppMatches := make([]profilepicker.ProfileMatch, len(matches))
	for i, m := range matches {
		ppMatches[i] = profilepicker.ProfileMatch{
			ProfileName: m.ProfileName,
			ProfilePath: m.ProfilePath,
			IsActive:    m.IsActive,
			SourceUrl:   m.SourceUrl,
		}
	}
	return profilepicker.PromptProfileSelection(componentType, componentName, ppMatches, os.Stdin, os.Stdout)
}

func (cl *ComponentLinker) renderLinkSummary(successCount, failedCount int, failedComponents []string) {
	f := cl.formatter
	f.EmptyLine()
	f.CounterSummary(successCount+failedCount, successCount, failedCount, 0)

	if failedCount > 0 {
		f.EmptyLine()
		fmt.Println("Failed components:")
		for _, comp := range failedComponents {
			fmt.Printf("  • %s\n", comp)
		}
	}
}

// displayProfileManagerAdapter adapts linker.ProfileManager to linkerDisplay.DisplayProfileManager.
type displayProfileManagerAdapter struct {
	pm ProfileManager
}

func (a *displayProfileManagerAdapter) ScanProfiles() ([]*linkerDisplay.Profile, error) {
	profiles, err := a.pm.ScanProfiles()
	if err != nil {
		return nil, err
	}
	result := make([]*linkerDisplay.Profile, len(profiles))
	for i, p := range profiles {
		result[i] = &linkerDisplay.Profile{
			Name:        p.Name,
			BasePath:    p.BasePath,
			HasAgents:   p.HasAgents,
			HasSkills:   p.HasSkills,
			HasCommands: p.HasCommands,
		}
	}
	return result, nil
}

func (a *displayProfileManagerAdapter) GetActiveProfile() (string, error) {
	return a.pm.GetActiveProfile()
}

// loadComponentMetadata loads metadata for a component — kept for compatibility.
func (cl *ComponentLinker) loadComponentMetadata(componentType, componentName string) interface{} {
	return nil
}

// isSymlinkFromCurrentProfile, isSymlinkFromAgentSmith, anyProfilesExist are now internal
// to the unlink sub-package. These wrappers exist only if called from linker.go tests.

// linkStatusLegendItems is re-exported from display for callers in this package.
func linkStatusLegendItems() []formatter.LegendItem {
	return linkerDisplay.LinkStatusLegendItems()
}

// isSymlinkFromCurrentProfile delegates to the unlink sub-package.
func (cl *ComponentLinker) isSymlinkFromCurrentProfile(symlinkPath string) (bool, error) {
	return linkerUnlink.IsSymlinkFromCurrentProfile(cl.agentsDir, symlinkPath)
}
