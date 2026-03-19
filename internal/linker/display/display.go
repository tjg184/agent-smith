package linkerDisplay

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/internal/linker/profilepicker"
	"github.com/tjg184/agent-smith/pkg/colors"
	"github.com/tjg184/agent-smith/pkg/config"
	"github.com/tjg184/agent-smith/pkg/paths"
)

// DisplayProfileManager is the minimal interface needed for display operations.
type DisplayProfileManager interface {
	ScanProfiles() ([]*Profile, error)
	GetActiveProfile() (string, error)
}

// Profile mirrors the linker.Profile struct to avoid circular imports.
type Profile struct {
	Name        string
	BasePath    string
	HasAgents   bool
	HasSkills   bool
	HasCommands bool
}

// LinkStatus represents the status of a linked component for display purposes.
type LinkStatus struct {
	Name       string
	Type       string
	LinkType   string
	Target     string
	Valid      bool
	TargetPath string
	Profile    string
}

// ListLinkedComponents lists all components linked to the configured targets.
func ListLinkedComponents(agentsDir string, targets []config.Target, f *formatter.Formatter) error {
	componentTypes := paths.GetComponentTypes()

	for _, target := range targets {
		allLinks := make(map[string][]LinkStatus)
		totalCount := 0
		validCount := 0
		brokenCount := 0

		for _, componentType := range componentTypes {
			componentDir, err := target.GetGlobalComponentDir(componentType)
			if err != nil {
				return fmt.Errorf("failed to get target component directory: %w", err)
			}

			if _, err := os.Stat(componentDir); os.IsNotExist(err) {
				allLinks[componentType] = []LinkStatus{}
				continue
			}

			entries, err := os.ReadDir(componentDir)
			if err != nil {
				return fmt.Errorf("failed to read %s directory: %w", componentType, err)
			}

			links := []LinkStatus{}
			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), ".") {
					continue
				}

				fullPath := filepath.Join(componentDir, entry.Name())
				linkType, targetPath, valid := analyzeLinkStatus(fullPath)

				profile := paths.BaseProfileName
				if targetPath != "" {
					profile = getProfileFromPath(targetPath)
				}

				status := LinkStatus{
					Name:       entry.Name(),
					Type:       componentType,
					LinkType:   linkType,
					Target:     targetPath,
					Valid:      valid,
					TargetPath: fullPath,
					Profile:    profile,
				}

				links = append(links, status)
				totalCount++

				if valid {
					validCount++
				} else if linkType == "broken" {
					brokenCount++
				}
			}

			allLinks[componentType] = links
		}

		targetName := target.GetName()
		targetDir, _ := target.GetGlobalBaseDir()

		fmt.Printf("\n=== %s ===\n", target.GetDisplayName())
		fmt.Printf("%s\n", getSourceDescription(agentsDir))

		if totalCount == 0 {
			fmt.Printf("No components are currently linked to %s.\n", targetName)
			fmt.Printf("Link location: %s\n", targetDir)
			continue
		}

		for _, componentType := range componentTypes {
			links := allLinks[componentType]
			if len(links) == 0 {
				continue
			}

			displayType := strings.Title(componentType)
			fmt.Printf("\n%s (%d):\n", displayType, len(links))

			for _, link := range links {
				var symbol, statusMsg string

				switch link.LinkType {
				case "symlink":
					if link.Valid {
						symbol = "✓"
						statusMsg = fmt.Sprintf("→ %s", link.Target)
					} else {
						symbol = "✗"
						statusMsg = "[broken link]"
					}
				case "copied":
					symbol = "◆"
					statusMsg = "[copied directory]"
				case "broken":
					symbol = "✗"
					statusMsg = "[broken link]"
				case "missing":
					symbol = "?"
					statusMsg = "[unknown state]"
				default:
					symbol = "?"
					statusMsg = "[unknown]"
				}

				profileDisplay := fmt.Sprintf("(%s)", link.Profile)
				fmt.Printf("  %s %s %s %s\n", symbol, link.Name, profileDisplay, statusMsg)
			}
		}

		fmt.Printf("\nTotal: %d components", totalCount)
		if brokenCount > 0 {
			fmt.Printf(" (%d valid, %d broken)", validCount, brokenCount)
		}
		fmt.Println()
	}

	return nil
}

// ShowLinkStatus displays a matrix view of components and their status across all targets.
func ShowLinkStatus(agentsDir string, targets []config.Target, f *formatter.Formatter, linkedOnly bool) error {
	componentTypes := paths.GetComponentTypes()

	type ComponentInfo struct {
		Name     string
		Type     string
		Profile  string
		BasePath string
	}
	allComponents := make([]ComponentInfo, 0)

	for _, componentType := range componentTypes {
		sourceDir := filepath.Join(agentsDir, componentType)
		if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(sourceDir)
		if err != nil {
			return fmt.Errorf("failed to read %s directory: %w", componentType, err)
		}

		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			componentPath := filepath.Join(sourceDir, entry.Name())

			var profile string
			info, err := os.Lstat(componentPath)
			if err == nil && info.Mode()&os.ModeSymlink != 0 {
				profile = profilepicker.GetProfileNameFromSymlink(componentPath)
				if profile == "" {
					profile = getProfileFromPath(componentPath)
				}
			} else {
				profile = getProfileFromPath(componentPath)
			}

			allComponents = append(allComponents, ComponentInfo{
				Name:     entry.Name(),
				Type:     componentType,
				Profile:  profile,
				BasePath: agentsDir,
			})
		}
	}

	if len(allComponents) == 0 {
		fmt.Fprintf(f.Writer(), "No components found in ~/.agent-smith/\n")
		return nil
	}

	type ComponentStatus struct {
		Component ComponentInfo
		Targets   map[string]string
	}

	statuses := make([]ComponentStatus, 0)

	for _, comp := range allComponents {
		status := ComponentStatus{
			Component: comp,
			Targets:   make(map[string]string),
		}

		for _, target := range targets {
			componentDir, err := target.GetGlobalComponentDir(comp.Type)
			if err != nil {
				status.Targets[target.GetName()] = colors.Warning("?")
				continue
			}

			if comp.Type == "commands" || comp.Type == "agents" {
				componentTypeDir := filepath.Join(comp.BasePath, comp.Type)
				if isFlatMdLinked(comp.Name, componentTypeDir, componentDir) {
					status.Targets[target.GetName()] = colors.Success("✓")
				} else {
					status.Targets[target.GetName()] = colors.Muted("-")
				}
				continue
			}

			linkPath := filepath.Join(componentDir, comp.Name)

			if _, err := os.Lstat(linkPath); os.IsNotExist(err) {
				status.Targets[target.GetName()] = colors.Muted("-")
				continue
			}

			linkType, targetPath, valid := analyzeLinkStatus(linkPath)

			expectedSource := filepath.Join(comp.BasePath, comp.Type, comp.Name)

			if linkType == "symlink" && valid {
				expectedSource, _ = filepath.EvalSymlinks(expectedSource)
				targetPath, _ = filepath.EvalSymlinks(targetPath)

				if expectedSource != targetPath {
					status.Targets[target.GetName()] = colors.Muted("-")
					continue
				}
			}

			status.Targets[target.GetName()] = linkStatusSymbol(linkType, valid)
		}

		statuses = append(statuses, status)
	}

	targetNames := make([]string, 0, len(targets))
	displayNames := targetDisplayNames(targets)
	for _, target := range targets {
		targetNames = append(targetNames, target.GetName())
	}

	f.EmptyLine()
	f.SectionHeader("Link Status Across All Targets")
	f.InfoMsg("%s", getSourceDescription(agentsDir))
	f.EmptyLine()

	headers := []string{"Component", "Profile"}
	for _, targetName := range targetNames {
		headers = append(headers, displayNames[targetName])
	}

	table := formatter.NewBoxTable(f.Writer(), headers)

	byType := make(map[string][]ComponentStatus)
	for _, status := range statuses {
		byType[status.Component.Type] = append(byType[status.Component.Type], status)
	}

	for _, componentType := range componentTypes {
		components := byType[componentType]
		if len(components) == 0 {
			continue
		}

		sectionRow := []string{strings.Title(componentType) + ":", ""}
		for range targetNames {
			sectionRow = append(sectionRow, "")
		}
		table.AddRow(sectionRow)

		for _, status := range components {
			if linkedOnly {
				hasAnyLink := false
				for _, symbol := range status.Targets {
					if symbol != colors.Muted("-") {
						hasAnyLink = true
						break
					}
				}
				if !hasAnyLink {
					continue
				}
			}

			componentName := fmt.Sprintf("  %s", status.Component.Name)
			row := []string{componentName, status.Component.Profile}

			for _, targetName := range targetNames {
				symbol := status.Targets[targetName]
				row = append(row, symbol)
			}
			table.AddRow(row)
		}
	}

	table.Render()

	f.EmptyLine()
	f.SubsectionHeader("Legend")
	f.DisplayLegendTable(LinkStatusLegendItems())

	f.EmptyLine()
	f.SubsectionHeader("Summary")
	for _, targetName := range targetNames {
		linkedCount := 0
		for _, status := range statuses {
			symbol := status.Targets[targetName]
			if symbol == colors.Success("✓") || symbol == colors.Success("◆") {
				linkedCount++
			}
		}
		f.ListItem("%s: %d/%d components linked", displayNames[targetName], linkedCount, len(statuses))
	}

	return nil
}

// ShowAllProfilesLinkStatus displays link status for components across all profiles.
func ShowAllProfilesLinkStatus(agentsDir string, targets []config.Target, f *formatter.Formatter, pm DisplayProfileManager, profileFilter []string, linkedOnly bool) error {
	componentTypes := paths.GetComponentTypes()

	type ComponentInfo struct {
		Name     string
		Type     string
		Profile  string
		BasePath string
	}
	allComponents := make([]ComponentInfo, 0)

	baseDir, err := paths.GetAgentsDir()
	if err != nil {
		return fmt.Errorf("failed to get base directory: %w", err)
	}

	baseComponents := make([]ComponentInfo, 0)
	for _, componentType := range componentTypes {
		sourceDir := filepath.Join(baseDir, componentType)
		if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(sourceDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			if !entry.IsDir() {
				continue
			}

			componentPath := filepath.Join(sourceDir, entry.Name())

			var profile string
			info, err := os.Lstat(componentPath)
			if err == nil && info.Mode()&os.ModeSymlink != 0 {
				profile = profilepicker.GetProfileNameFromSymlink(componentPath)
				if profile == "" {
					profile = paths.BaseProfileName
				}
			} else {
				profile = paths.BaseProfileName
			}

			baseComponents = append(baseComponents, ComponentInfo{
				Name:     entry.Name(),
				Type:     componentType,
				Profile:  profile,
				BasePath: baseDir,
			})
		}
	}

	profiles, err := pm.ScanProfiles()
	if err != nil {
		return fmt.Errorf("failed to scan profiles: %w", err)
	}

	var filteredProfiles []*Profile
	if len(profileFilter) > 0 {
		filterMap := make(map[string]bool)
		for _, name := range profileFilter {
			filterMap[name] = true
		}

		profileMap := make(map[string]bool)
		for _, p := range profiles {
			profileMap[p.Name] = true
		}

		for _, filterName := range profileFilter {
			if !profileMap[filterName] {
				return fmt.Errorf("profile '%s' does not exist", filterName)
			}
		}

		for _, p := range profiles {
			if filterMap[p.Name] {
				filteredProfiles = append(filteredProfiles, p)
			}
		}
	} else {
		filteredProfiles = profiles
	}

	profileComponents := make([]ComponentInfo, 0)
	for _, profile := range filteredProfiles {
		for _, componentType := range componentTypes {
			var sourceDir string
			switch componentType {
			case "agents":
				if !profile.HasAgents {
					continue
				}
				sourceDir = filepath.Join(profile.BasePath, "agents")
			case "skills":
				if !profile.HasSkills {
					continue
				}
				sourceDir = filepath.Join(profile.BasePath, "skills")
			case "commands":
				if !profile.HasCommands {
					continue
				}
				sourceDir = filepath.Join(profile.BasePath, "commands")
			default:
				continue
			}

			if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
				continue
			}

			entries, err := os.ReadDir(sourceDir)
			if err != nil {
				continue
			}
			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), ".") {
					continue
				}
				if !entry.IsDir() {
					continue
				}
				profileComponents = append(profileComponents, ComponentInfo{
					Name:     entry.Name(),
					Type:     componentType,
					Profile:  profile.Name,
					BasePath: profile.BasePath,
				})
			}
		}
	}

	if len(profileFilter) == 0 {
		allComponents = append(allComponents, baseComponents...)
	}
	allComponents = append(allComponents, profileComponents...)

	if len(allComponents) == 0 {
		if len(profileFilter) > 0 {
			fmt.Fprintf(f.Writer(), "No components found in the specified profiles\n")
		} else {
			fmt.Fprintf(f.Writer(), "No components found\n")
		}
		return nil
	}

	type ComponentStatus struct {
		Component ComponentInfo
		Targets   map[string]string
	}

	statuses := make([]ComponentStatus, 0)

	for _, comp := range allComponents {
		status := ComponentStatus{
			Component: comp,
			Targets:   make(map[string]string),
		}

		for _, target := range targets {
			componentDir, err := target.GetGlobalComponentDir(comp.Type)
			if err != nil {
				status.Targets[target.GetName()] = colors.Warning("?")
				continue
			}

			if comp.Type == "commands" || comp.Type == "agents" {
				componentTypeDir := filepath.Join(comp.BasePath, comp.Type)
				if isFlatMdLinked(comp.Name, componentTypeDir, componentDir) {
					status.Targets[target.GetName()] = colors.Success("✓")
				} else {
					status.Targets[target.GetName()] = colors.Muted("-")
				}
				continue
			}

			linkPath := filepath.Join(componentDir, comp.Name)

			if _, err := os.Lstat(linkPath); os.IsNotExist(err) {
				status.Targets[target.GetName()] = colors.Muted("-")
				continue
			}

			linkType, linkTarget, valid := analyzeLinkStatus(linkPath)

			if linkType == "symlink" && valid {
				expectedSource := filepath.Join(comp.BasePath, comp.Type, comp.Name)

				absTarget, err1 := filepath.Abs(linkTarget)
				absExpected, err2 := filepath.Abs(expectedSource)

				if err1 == nil && err2 == nil {
					absTarget = filepath.Clean(absTarget)
					absExpected = filepath.Clean(absExpected)

					if absTarget != absExpected {
						status.Targets[target.GetName()] = colors.Muted("-")
						continue
					}
				}
			}

			status.Targets[target.GetName()] = linkStatusSymbol(linkType, valid)
		}

		statuses = append(statuses, status)
	}

	targetNames := make([]string, 0, len(targets))
	displayNames := targetDisplayNames(targets)
	for _, target := range targets {
		targetNames = append(targetNames, target.GetName())
	}

	f.EmptyLine()
	f.SectionHeader("Link Status Across All Profiles")
	f.EmptyLine()

	headers := []string{"Component", "Type", "Profile"}
	for _, targetName := range targetNames {
		headers = append(headers, displayNames[targetName])
	}

	table := formatter.NewBoxTable(f.Writer(), headers)

	byType := make(map[string][]ComponentStatus)
	for _, status := range statuses {
		byType[status.Component.Type] = append(byType[status.Component.Type], status)
	}

	for _, componentType := range componentTypes {
		components := byType[componentType]
		if len(components) == 0 {
			continue
		}

		sectionRow := []string{strings.Title(componentType) + ":", "", ""}
		for range targetNames {
			sectionRow = append(sectionRow, "")
		}
		table.AddRow(sectionRow)

		for _, status := range components {
			if linkedOnly {
				hasAnyLink := false
				for _, targetName := range targetNames {
					symbol := status.Targets[targetName]
					if symbol != colors.Muted("-") {
						hasAnyLink = true
						break
					}
				}
				if !hasAnyLink {
					continue
				}
			}

			componentName := fmt.Sprintf("  %s", status.Component.Name)
			row := []string{componentName, status.Component.Type, status.Component.Profile}

			for _, targetName := range targetNames {
				symbol := status.Targets[targetName]
				row = append(row, symbol)
			}
			table.AddRow(row)
		}
	}

	table.Render()

	f.EmptyLine()
	f.SubsectionHeader("Legend")
	f.DisplayLegendTable(LinkStatusLegendItems())

	f.EmptyLine()
	f.SubsectionHeader("Summary")

	profileCount := len(filteredProfiles)
	if len(profileFilter) == 0 {
		profileCount++
	}
	profileCountStr := fmt.Sprintf("%d", profileCount)
	if len(profileFilter) == 0 {
		if len(filteredProfiles) == 0 {
			profileCountStr = fmt.Sprintf("1 (%s only)", paths.BaseProfileName)
		} else {
			profileCountStr = fmt.Sprintf("%d (%s + %d custom)", profileCount, paths.BaseProfileName, len(filteredProfiles))
		}
	}
	f.ListItem("Profiles scanned: %s", profileCountStr)
	f.ListItem("Total components: %d", len(statuses))

	for _, targetName := range targetNames {
		linkedCount := 0
		for _, status := range statuses {
			symbol := status.Targets[targetName]
			if symbol == colors.Success("✓") || symbol == colors.Success("◆") {
				linkedCount++
			}
		}
		percentage := 0
		if len(statuses) > 0 {
			percentage = (linkedCount * 100) / len(statuses)
		}
		f.ListItem("%s: %d/%d linked (%d%%)", displayNames[targetName], linkedCount, len(statuses), percentage)
	}

	activeProfile, err := pm.GetActiveProfile()
	if err == nil && activeProfile != "" {
		f.EmptyLine()
		f.InfoMsg("Active Profile: %s", activeProfile)
	}

	return nil
}

// LinkStatusLegendItems returns legend items for link status tables.
func LinkStatusLegendItems() []formatter.LegendItem {
	return []formatter.LegendItem{
		{Symbol: colors.Success("✓"), Description: colors.Success("Valid symlink")},
		{Symbol: colors.Success("◆"), Description: colors.Success("Copied directory")},
		{Symbol: colors.Error("✗"), Description: colors.Error("Broken link")},
		{Symbol: colors.Muted("-"), Description: colors.Muted("Not linked")},
		{Symbol: colors.Warning("?"), Description: colors.Warning("Unknown status")},
	}
}

func linkStatusSymbol(linkType string, valid bool) string {
	switch linkType {
	case "symlink":
		if valid {
			return colors.Success("✓")
		}
		return colors.Error("✗")
	case "copied":
		return colors.Success("◆")
	case "broken":
		return colors.Error("✗")
	default:
		return colors.Warning("?")
	}
}

func getSourceDescription(agentsDir string) string {
	if filepath.Base(filepath.Dir(agentsDir)) == "profiles" {
		profileName := filepath.Base(agentsDir)
		return fmt.Sprintf("Source: %s (profile '%s')", agentsDir, profileName)
	}
	return fmt.Sprintf("Source: %s (base installation)", agentsDir)
}

func getProfileFromPath(path string) string {
	path = filepath.Clean(path)

	parent := filepath.Dir(path)
	if filepath.Base(parent) == "profiles" {
		return filepath.Base(path)
	}

	dir := parent
	for {
		grandparent := filepath.Dir(dir)
		if filepath.Base(grandparent) == "profiles" {
			return filepath.Base(dir)
		}
		if grandparent == dir || grandparent == "." || grandparent == "/" {
			return paths.BaseProfileName
		}
		dir = grandparent
	}
}

func targetDisplayNames(targets []config.Target) map[string]string {
	m := make(map[string]string, len(targets))
	for _, t := range targets {
		m[t.GetName()] = t.GetDisplayName()
	}
	return m
}

func analyzeLinkStatus(path string) (linkType string, target string, valid bool) {
	info, err := os.Lstat(path)
	if err != nil {
		return "missing", "", false
	}

	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return "broken", "", false
		}

		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(path), target)
		}

		if _, err := os.Stat(target); err == nil {
			return "symlink", target, true
		}
		return "broken", target, false
	}

	if info.IsDir() {
		return "copied", path, true
	}

	return "unknown", "", false
}

func isFlatMdLinked(componentName, componentTypeDir, targetBaseDir string) bool {
	componentRoot := filepath.Clean(filepath.Join(componentTypeDir, componentName))
	expectedPrefix := componentRoot + string(filepath.Separator)

	found := false
	_ = filepath.WalkDir(targetBaseDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || found {
			return nil
		}

		info, err := os.Lstat(path)
		if err != nil || info.Mode()&os.ModeSymlink == 0 {
			return nil
		}

		target, err := os.Readlink(path)
		if err != nil {
			return nil
		}

		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(path), target)
		}
		target = filepath.Clean(target)

		if target == componentRoot || strings.HasPrefix(target, expectedPrefix) {
			found = true
		}

		return nil
	})

	return found
}
