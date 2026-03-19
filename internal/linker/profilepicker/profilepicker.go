package profilepicker

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	metadataPkg "github.com/tjg184/agent-smith/internal/metadata"
	"github.com/tjg184/agent-smith/pkg/paths"
)

// ProfileMatch represents a profile that contains a specific component.
type ProfileMatch struct {
	ProfileName string
	ProfilePath string
	IsActive    bool
	SourceUrl   string
}

// GetProfileNameFromSymlink reads the symlink target and extracts the profile name from its path.
func GetProfileNameFromSymlink(symlinkPath string) string {
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		return ""
	}

	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(symlinkPath), target)
	}

	return getProfileFromPath(target)
}

// getProfileFromPath extracts the profile name from a filesystem path.
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

// SearchComponentInProfiles searches all profiles for a component and returns matching profiles.
func SearchComponentInProfiles(componentType, componentName string) ([]ProfileMatch, error) {
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get profiles directory: %w", err)
	}

	if _, err := os.Stat(profilesDir); os.IsNotExist(err) {
		return []ProfileMatch{}, nil
	}

	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}
	activeProfilePath := filepath.Join(agentsDir, ".active-profile")
	activeProfileData, _ := os.ReadFile(activeProfilePath)
	activeProfile := strings.TrimSpace(string(activeProfileData))

	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read profiles directory: %w", err)
	}

	var matches []ProfileMatch
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		profileName := entry.Name()
		profilePath := filepath.Join(profilesDir, profileName)
		componentPath := filepath.Join(profilePath, componentType, componentName)

		if _, err := os.Stat(componentPath); err == nil {
			sourceUrl := ""
			lockEntry, err := metadataPkg.LoadLockFileEntry(profilePath, componentType, componentName)
			if err == nil && lockEntry != nil {
				sourceUrl = lockEntry.SourceUrl
			}

			matches = append(matches, ProfileMatch{
				ProfileName: profileName,
				ProfilePath: profilePath,
				IsActive:    profileName == activeProfile,
				SourceUrl:   sourceUrl,
			})
		}
	}

	return matches, nil
}

// PromptProfileSelection displays an interactive prompt for the user to select a profile.
// Returns the selected profile path and name. Reads from in, writes to out.
func PromptProfileSelection(componentType, componentName string, matches []ProfileMatch, in io.Reader, out io.Writer) (profilePath string, profileName string, err error) {
	if len(matches) == 0 {
		return "", "", fmt.Errorf("no profiles contain component %s", componentName)
	}

	fmt.Fprintf(out, "\n⚠️  Component \"%s\" found in multiple profiles:\n\n", componentName)

	for i, match := range matches {
		activeIndicator := ""
		if match.IsActive {
			activeIndicator = " (active)"
		}
		fmt.Fprintf(out, "  %d. %s%s\n", i+1, match.ProfileName, activeIndicator)

		if match.SourceUrl != "" {
			fmt.Fprintf(out, "     Source: %s\n", match.SourceUrl)
		}

		if i < len(matches)-1 {
			fmt.Fprintln(out)
		}
	}

	fmt.Fprintf(out, "\nSelect profile to link from [1-%d] (or 'c' to cancel): ", len(matches))

	reader := bufio.NewReader(in)
	line, _ := reader.ReadString('\n')
	response := strings.TrimSpace(strings.ToLower(line))

	if response == "c" || response == "" {
		return "", "", fmt.Errorf("profile selection cancelled")
	}

	var selection int
	_, scanErr := fmt.Sscanf(response, "%d", &selection)
	if scanErr != nil || selection < 1 || selection > len(matches) {
		return "", "", fmt.Errorf("invalid selection: %s", response)
	}

	selected := matches[selection-1]
	return selected.ProfilePath, selected.ProfileName, nil
}
