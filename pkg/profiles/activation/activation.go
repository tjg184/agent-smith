package activation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjg184/agent-smith/pkg/paths"
)

type ProfileActivationResult struct {
	PreviousProfile string // empty if no profile was active
	NewProfile      string
	Switched        bool // true if switching from another profile
}

// GetActiveProfile returns the name of the currently active profile, or empty string if none.
func GetActiveProfile() (string, error) {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return "", fmt.Errorf("failed to get agents directory: %w", err)
	}

	activeProfilePath := filepath.Join(agentsDir, ".active-profile")

	if _, err := os.Stat(activeProfilePath); os.IsNotExist(err) {
		return "", nil
	}

	data, err := os.ReadFile(activeProfilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read active profile file: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}

func ActivateProfile(profileName string, profileIsValid bool) error {
	_, err := ActivateProfileWithResult(profileName, profileIsValid)
	return err
}

func ActivateProfileWithResult(profileName string, profileIsValid bool) (*ProfileActivationResult, error) {
	if !profileIsValid {
		return nil, fmt.Errorf("profile '%s' does not exist or has no components", profileName)
	}

	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}

	currentActive, err := GetActiveProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to check current active profile: %w", err)
	}

	if currentActive == profileName {
		return &ProfileActivationResult{
			PreviousProfile: currentActive,
			NewProfile:      profileName,
			Switched:        false,
		}, nil
	}

	activeProfilePath := filepath.Join(agentsDir, ".active-profile")
	if err := os.WriteFile(activeProfilePath, []byte(profileName), 0644); err != nil {
		return nil, fmt.Errorf("failed to write active profile state: %w", err)
	}

	return &ProfileActivationResult{
		PreviousProfile: currentActive,
		NewProfile:      profileName,
		Switched:        currentActive != "",
	}, nil
}

func DeactivateProfile() error {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return fmt.Errorf("failed to get agents directory: %w", err)
	}

	currentActive, err := GetActiveProfile()
	if err != nil {
		return fmt.Errorf("failed to check current active profile: %w", err)
	}

	if currentActive == "" {
		return fmt.Errorf("no profile is currently active")
	}

	activeProfilePath := filepath.Join(agentsDir, ".active-profile")
	if err := os.Remove(activeProfilePath); err != nil {
		return fmt.Errorf("failed to clear active profile state: %w", err)
	}

	return nil
}
