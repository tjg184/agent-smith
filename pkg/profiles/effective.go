package profiles

import (
	"github.com/tjg184/agent-smith/pkg/logger"
	locksvc "github.com/tjg184/agent-smith/pkg/services/lock"
)

// ResolveActiveProfile returns the name of the currently active profile,
// or an empty string if no profile is active. Returns an error only if
// the profile manager itself cannot be constructed.
func ResolveActiveProfile() (string, error) {
	pm, err := NewProfileManager(nil, locksvc.NewService(logger.New(logger.LevelError)))
	if err != nil {
		return "", err
	}
	activeProfile, err := pm.GetActiveProfile()
	if err != nil {
		return "", nil
	}
	return activeProfile, nil
}
