package linkerDisplay

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/internal/linker/linkutil"
	"github.com/tjg184/agent-smith/internal/linker/profilepicker"
	"github.com/tjg184/agent-smith/pkg/colors"
	"github.com/tjg184/agent-smith/pkg/config"
	"github.com/tjg184/agent-smith/pkg/paths"
	"github.com/tjg184/agent-smith/pkg/profiles/profilemeta"
)

// profileLabel returns "<name> (<url>)" for repo profiles, or just the name.
func profileLabel(profileName string) string {
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		return profileName
	}
	meta, err := profilemeta.Load(filepath.Join(profilesDir, profileName))
	if err != nil || meta == nil || meta.SourceURL == "" {
		return profileName
	}
	return fmt.Sprintf("%s (%s)", profileName, meta.SourceURL)
}

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

type LinkStatus struct {
	Name       string
	Type       string
	LinkType   string
	Target     string
	Valid      bool
	TargetPath string
	Profile    string
}

type ComponentInfo struct {
	Name     string
	Type     string
	Profile  string
	BasePath string
}

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
				linkType, targetPath, valid := linkutil.AnalyzeLinkStatus(fullPath)

				profile := linkutil.ProfileFromPath(targetPath)

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

func ShowLinkStatus(agentsDir string, targets []config.Target, f *formatter.Formatter, linkedOnly bool) error {
	componentTypes := paths.GetComponentTypes()

	allComponents := make([]ComponentInfo, 0)

	for _, componentType := range componentTypes {
		sourceDir := filepath.Join(agentsDir, componentType)
		if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
			continue
		}

		if componentType == "skills" {
			allComponents = append(allComponents, collectLeafSkills(sourceDir, "", agentsDir, componentType, resolveProfileForPath)...)
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
					profile = linkutil.ProfileFromPath(componentPath)
				}
			} else {
				profile = linkutil.ProfileFromPath(componentPath)
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
				if linkutil.IsFlatMdLinked(comp.Name, componentTypeDir, componentDir) {
					status.Targets[target.GetName()] = colors.Success("✓")
				} else {
					status.Targets[target.GetName()] = colors.Muted("-")
				}
				continue
			}

			expectedSource := filepath.Join(comp.BasePath, comp.Type, comp.Name)
			symbol := skillLinkSymbol(componentDir, comp.Name, expectedSource)
			status.Targets[target.GetName()] = symbol
		}

		statuses = append(statuses, status)
	}

	targetNames := make([]string, 0, len(targets))
	displayNames := targetDisplayNames(targets)
	for _, target := range targets {
		targetNames = append(targetNames, target.GetName())
	}

	f.EmptyLine()
	f.InfoMsg("%s", getSourceDescription(agentsDir))
	f.EmptyLine()

	headers := []string{"Component", "Profile / Repo"}
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
			row := []string{componentName, profileLabel(status.Component.Profile)}

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

func ShowAllProfilesLinkStatus(agentsDir string, targets []config.Target, f *formatter.Formatter, pm DisplayProfileManager, profileFilter []string, linkedOnly bool) error {
	componentTypes := paths.GetComponentTypes()

	allComponents := make([]ComponentInfo, 0)

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

			if componentType == "skills" {
				fixedProfile := profile.Name
				profileComponents = append(profileComponents, collectLeafSkills(sourceDir, "", profile.BasePath, componentType, func(_ string) string { return fixedProfile })...)
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
				if linkutil.IsFlatMdLinked(comp.Name, componentTypeDir, componentDir) {
					status.Targets[target.GetName()] = colors.Success("✓")
				} else {
					status.Targets[target.GetName()] = colors.Muted("-")
				}
				continue
			}

			expectedSource := filepath.Join(comp.BasePath, comp.Type, comp.Name)
			symbol := skillLinkSymbol(componentDir, comp.Name, expectedSource)
			status.Targets[target.GetName()] = symbol
		}

		statuses = append(statuses, status)
	}

	targetNames := make([]string, 0, len(targets))
	displayNames := targetDisplayNames(targets)
	for _, target := range targets {
		targetNames = append(targetNames, target.GetName())
	}

	f.EmptyLine()

	headers := []string{"Component", "Type", "Profile / Repo"}
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
			row := []string{componentName, status.Component.Type, profileLabel(status.Component.Profile)}

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

	profileCountStr := fmt.Sprintf("%d", len(filteredProfiles))
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
	return fmt.Sprintf("Source: %s", agentsDir)
}

func targetDisplayNames(targets []config.Target) map[string]string {
	m := make(map[string]string, len(targets))
	for _, t := range targets {
		m[t.GetName()] = t.GetDisplayName()
	}
	return m
}

// resolveProfile is called with the absolute path of each leaf to determine its
// profile name — pass a closure over a fixed string for profile-scoped calls, or
// resolveProfileForPath when the profile must be inferred from disk.
func collectLeafSkills(dir, relPrefix, basePath, componentType string, resolveProfile func(string) string) []ComponentInfo {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var results []ComponentInfo
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") || !entry.IsDir() {
			continue
		}

		relName := entry.Name()
		if relPrefix != "" {
			relName = relPrefix + string(filepath.Separator) + entry.Name()
		}

		entryPath := filepath.Join(dir, entry.Name())
		skillMD := filepath.Join(entryPath, "SKILL.md")
		if _, err := os.Stat(skillMD); err == nil {
			results = append(results, ComponentInfo{
				Name:     relName,
				Type:     componentType,
				Profile:  resolveProfile(entryPath),
				BasePath: basePath,
			})
		} else {
			results = append(results, collectLeafSkills(entryPath, relName, basePath, componentType, resolveProfile)...)
		}
	}
	return results
}

func resolveProfileForPath(path string) string {
	info, err := os.Lstat(path)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		profile := profilepicker.GetProfileNameFromSymlink(path)
		if profile != "" {
			return profile
		}
		return linkutil.ProfileFromPath(path)
	}
	return linkutil.ProfileFromPath(path)
}

// expectedSource is the absolute path of the skill in the profile
// (e.g. ~/.agent-smith/profiles/foo/skills/sdlc-pipeline/record-completion).
func skillLinkSymbol(componentDir, compName, expectedSource string) string {
	parts := strings.SplitN(compName, string(filepath.Separator), 2)
	categoryPath := filepath.Join(componentDir, parts[0])

	// Check category-level symlink first. If it exists as a symlink the whole
	// category directory is managed — verify it points to the right source.
	if info, err := os.Lstat(categoryPath); err == nil && info.Mode()&os.ModeSymlink != 0 {
		categorySource := filepath.Join(filepath.Dir(expectedSource), parts[0])
		return symlinkMatchSymbol(categoryPath, categorySource)
	}

	// No category symlink — check for a leaf symlink inside a real category dir.
	leafPath := filepath.Join(componentDir, compName)
	if info, err := os.Lstat(leafPath); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return symlinkMatchSymbol(leafPath, expectedSource)
		}
		// Real directory at the leaf level — genuinely copied.
		if info.IsDir() {
			return colors.Success("◆")
		}
	}

	return colors.Muted("-")
}

func symlinkMatchSymbol(linkPath, expectedSource string) string {
	target, err := os.Readlink(linkPath)
	if err != nil {
		return colors.Error("✗")
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(linkPath), target)
	}
	if _, err := os.Stat(target); err != nil {
		return colors.Error("✗")
	}
	resolvedTarget, err1 := filepath.EvalSymlinks(target)
	resolvedExpected, err2 := filepath.EvalSymlinks(expectedSource)
	if err1 != nil || err2 != nil {
		return colors.Success("✓")
	}
	if resolvedTarget != resolvedExpected {
		return colors.Muted("-")
	}
	return colors.Success("✓")
}
