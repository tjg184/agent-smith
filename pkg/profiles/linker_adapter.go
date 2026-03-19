package profiles

import "github.com/tjg184/agent-smith/internal/linker"

// NewLinkerAdapter wraps a ProfileManager so it satisfies the linker.ProfileManager
// interface. This adapter is the single canonical implementation — callers in the
// container and link service both use this instead of maintaining their own copies.
func NewLinkerAdapter(pm *ProfileManager) linker.ProfileManager {
	return &linkerAdapter{pm: pm}
}

type linkerAdapter struct {
	pm *ProfileManager
}

func (a *linkerAdapter) ScanProfiles() ([]*linker.Profile, error) {
	scanned, err := a.pm.ScanProfiles()
	if err != nil {
		return nil, err
	}
	result := make([]*linker.Profile, len(scanned))
	for i, p := range scanned {
		result[i] = &linker.Profile{
			Name:        p.Name,
			BasePath:    p.BasePath,
			HasAgents:   p.HasAgents,
			HasSkills:   p.HasSkills,
			HasCommands: p.HasCommands,
		}
	}
	return result, nil
}

func (a *linkerAdapter) GetActiveProfile() (string, error) {
	return a.pm.GetActiveProfile()
}
