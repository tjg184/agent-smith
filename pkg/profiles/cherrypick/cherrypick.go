package cherrypick

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type ComponentItem struct {
	Type          string // "skills", "agents", or "commands"
	Name          string
	SourceProfile string
}

// CherryPickDeps is the interface the caller must satisfy so this sub-package
// can copy components without importing pkg/profiles (which would be circular).
type CherryPickDeps interface {
	CopyComponentBetweenProfiles(sourceProfile, targetProfile, componentType, componentName string) error
	CreateProfile(profileName string) error
	ProfileExists(profileName string) bool
	CountComponents(profileName string) (agents, skills, commands int)
}

// PromptComponentSelection displays an interactive UI and returns the selected components.
func PromptComponentSelection(components []ComponentItem, in io.Reader, out io.Writer) ([]ComponentItem, error) {
	if len(components) == 0 {
		return nil, fmt.Errorf("no components available for selection")
	}

	var skills, agents, commands []ComponentItem
	for _, c := range components {
		switch c.Type {
		case "skills":
			skills = append(skills, c)
		case "agents":
			agents = append(agents, c)
		case "commands":
			commands = append(commands, c)
		}
	}

	fmt.Fprintln(out, "\nAvailable Components:")
	fmt.Fprintln(out, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if len(skills) > 0 {
		fmt.Fprintf(out, "\nSkills (%d):\n", len(skills))
		for i, c := range skills {
			fmt.Fprintf(out, "  [%d] %s  (from: %s)\n", i+1, c.Name, c.SourceProfile)
		}
	}

	if len(agents) > 0 {
		fmt.Fprintf(out, "\nAgents (%d):\n", len(agents))
		for i, c := range agents {
			fmt.Fprintf(out, "  [%d] %s  (from: %s)\n", len(skills)+i+1, c.Name, c.SourceProfile)
		}
	}

	if len(commands) > 0 {
		fmt.Fprintf(out, "\nCommands (%d):\n", len(commands))
		for i, c := range commands {
			fmt.Fprintf(out, "  [%d] %s  (from: %s)\n", len(skills)+len(agents)+i+1, c.Name, c.SourceProfile)
		}
	}

	fmt.Fprintln(out, "\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Fprintln(out, "\nSelect components:")
	fmt.Fprintln(out, "  Enter numbers (e.g., 1,3,5 or 1-3)")
	fmt.Fprintln(out, "  Keywords: a=all, s=skills, g=agents, c=commands")
	fmt.Fprintln(out, "  q=quit")
	fmt.Fprint(out, "\nSelection: ")

	reader := bufio.NewReader(in)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)

	if strings.ToLower(response) == "q" {
		return nil, fmt.Errorf("selection cancelled")
	}

	indexMap := buildIndexMap(skills, agents, commands)
	selected := parseSelection(response, indexMap)

	if len(selected) == 0 {
		return nil, fmt.Errorf("no components selected")
	}

	fmt.Fprintln(out, "\nSelected:")
	for _, c := range selected {
		fmt.Fprintf(out, "  ✓ %s (%s) from %s\n", c.Name, c.Type, c.SourceProfile)
	}

	fmt.Fprint(out, "\nCopy to profile? [y/n]: ")
	confirm, _ := reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
		return nil, fmt.Errorf("cancelled")
	}

	return selected, nil
}

func CherryPickComponents(deps CherryPickDeps, targetProfile string, selectedComponents []ComponentItem) error {
	if !deps.ProfileExists(targetProfile) {
		fmt.Printf("Creating new profile '%s'...\n", targetProfile)
		if err := deps.CreateProfile(targetProfile); err != nil {
			return fmt.Errorf("failed to create target profile: %w", err)
		}
	}

	fmt.Printf("Cherry-picking %d component(s) to profile '%s'...\n\n", len(selectedComponents), targetProfile)

	successCount := 0
	skipCount := 0
	errorCount := 0

	for _, component := range selectedComponents {
		fmt.Printf("Copying %s '%s' from '%s'...\n", component.Type, component.Name, component.SourceProfile)

		err := deps.CopyComponentBetweenProfiles(component.SourceProfile, targetProfile, component.Type, component.Name)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				fmt.Printf("  ⊘ Skipped (already exists)\n\n")
				skipCount++
			} else {
				fmt.Printf("  ✗ Error: %v\n\n", err)
				errorCount++
			}
		} else {
			fmt.Printf("  ✓ Success\n\n")
			successCount++
		}
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("\nCherry-pick Summary:\n")
	fmt.Printf("  ✓ Successfully copied: %d\n", successCount)
	if skipCount > 0 {
		fmt.Printf("  ⊘ Skipped (existing):  %d\n", skipCount)
	}
	if errorCount > 0 {
		fmt.Printf("  ✗ Failed:              %d\n", errorCount)
	}

	fmt.Printf("\nTotal components in '%s': ", targetProfile)
	agents, skills, commands := deps.CountComponents(targetProfile)
	total := agents + skills + commands
	fmt.Printf("%d (%d agents, %d skills, %d commands)\n", total, agents, skills, commands)

	if errorCount > 0 {
		return fmt.Errorf("some components failed to copy")
	}

	return nil
}

func buildIndexMap(skills, agents, commands []ComponentItem) map[int]ComponentItem {
	m := make(map[int]ComponentItem)
	idx := 1
	for _, groups := range [][]ComponentItem{skills, agents, commands} {
		for _, c := range groups {
			m[idx] = c
			idx++
		}
	}
	return m
}

func parseSelection(response string, indexMap map[int]ComponentItem) []ComponentItem {
	selected := make(map[int]ComponentItem)

	for _, part := range strings.Split(response, ",") {
		part = strings.TrimSpace(part)

		switch strings.ToLower(part) {
		case "a", "all":
			for i, c := range indexMap {
				selected[i] = c
			}
			continue
		case "s", "skills":
			for i, c := range indexMap {
				if c.Type == "skills" {
					selected[i] = c
				}
			}
			continue
		case "g", "agents":
			for i, c := range indexMap {
				if c.Type == "agents" {
					selected[i] = c
				}
			}
			continue
		case "c", "commands":
			for i, c := range indexMap {
				if c.Type == "commands" {
					selected[i] = c
				}
			}
			continue
		}

		if strings.Contains(part, "-") {
			rangeParts := strings.SplitN(part, "-", 2)
			start, err1 := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err1 == nil && err2 == nil && start <= end {
				for i := start; i <= end; i++ {
					if c, ok := indexMap[i]; ok {
						selected[i] = c
					}
				}
			}
			continue
		}

		if num, err := strconv.Atoi(part); err == nil {
			if c, ok := indexMap[num]; ok {
				selected[num] = c
			}
		}
	}

	result := make([]ComponentItem, 0, len(selected))
	for _, c := range selected {
		result = append(result, c)
	}
	return result
}
