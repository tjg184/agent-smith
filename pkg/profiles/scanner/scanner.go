package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	metadataPkg "github.com/tjg184/agent-smith/internal/metadata"
	"github.com/tjg184/agent-smith/pkg/paths"
	"github.com/tjg184/agent-smith/pkg/services"
)

// Profile is a minimal representation of a profile for scanner operations.
// The parent profiles package maps these to its own Profile type.
type Profile struct {
	Name        string
	BasePath    string
	HasAgents   bool
	HasSkills   bool
	HasCommands bool
}

func (p *Profile) IsValid() bool {
	return p.HasAgents || p.HasSkills || p.HasCommands
}

// ScanProfiles discovers all valid profiles in profilesDir.
// Returns an empty slice if the directory doesn't exist.
func ScanProfiles(profilesDir string) ([]*Profile, error) {
	if _, err := os.Stat(profilesDir); os.IsNotExist(err) {
		return []*Profile{}, nil
	}

	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read profiles directory: %w", err)
	}

	var profiles []*Profile
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		fullPath := filepath.Join(profilesDir, entry.Name())
		info, err := os.Stat(fullPath)
		if err != nil || !info.IsDir() {
			continue
		}

		profile := LoadProfile(profilesDir, entry.Name())
		if profile.IsValid() {
			profiles = append(profiles, profile)
		}
	}

	return profiles, nil
}

func LoadProfile(profilesDir, name string) *Profile {
	basePath := filepath.Join(profilesDir, name)
	profile := &Profile{Name: name, BasePath: basePath}

	if _, err := os.Stat(filepath.Join(basePath, paths.AgentsSubDir)); err == nil {
		profile.HasAgents = true
	}
	if _, err := os.Stat(filepath.Join(basePath, paths.SkillsSubDir)); err == nil {
		profile.HasSkills = true
	}
	if _, err := os.Stat(filepath.Join(basePath, paths.CommandsSubDir)); err == nil {
		profile.HasCommands = true
	}

	return profile
}

func CountComponents(profile *Profile) (agents, skills, commands int) {
	if profile.HasAgents {
		agents = countDirs(filepath.Join(profile.BasePath, paths.AgentsSubDir))
	}
	if profile.HasSkills {
		components, err := metadataPkg.LoadAllComponents(profile.BasePath, "skills")
		if err == nil && len(components) > 0 {
			skills = len(components)
		} else {
			skills = countDirs(filepath.Join(profile.BasePath, paths.SkillsSubDir))
		}
	}
	if profile.HasCommands {
		commands = countDirs(filepath.Join(profile.BasePath, paths.CommandsSubDir))
	}
	return agents, skills, commands
}

func GetComponentNames(profile *Profile) (agents, skills, commands []string) {
	if profile.HasAgents {
		agents = listDirs(filepath.Join(profile.BasePath, paths.AgentsSubDir))
	}
	if profile.HasSkills {
		components, err := metadataPkg.LoadAllComponents(profile.BasePath, "skills")
		if err == nil && len(components) > 0 {
			for _, c := range components {
				name := c.Entry.FilesystemName
				if name == "" {
					name = c.Name
				}
				skills = append(skills, name)
			}
		} else {
			skills = listDirs(filepath.Join(profile.BasePath, paths.SkillsSubDir))
		}
	}
	if profile.HasCommands {
		commands = listDirs(filepath.Join(profile.BasePath, paths.CommandsSubDir))
	}
	return agents, skills, commands
}

func GetComponentSource(profile *Profile, lockService services.ComponentLockService, componentType, componentName string) string {
	lookupName := componentName
	if componentType == "skills" {
		// GetComponentNames returns FilesystemName values (e.g. "sdlc-pipeline/brainstorm-vision")
		// but FindComponentSources looks up by the short lock-file key ("brainstorm-vision").
		lookupName = filepath.Base(componentName)
	}
	sources, err := lockService.FindComponentSources(profile.BasePath, componentType, lookupName)
	if err != nil || len(sources) == 0 {
		return ""
	}
	return sources[0]
}

func countDirs(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			count++
		}
	}
	return count
}

func listDirs(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			names = append(names, e.Name())
		}
	}
	return names
}
