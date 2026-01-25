package config

// Target defines the interface for component linking targets
// A target specifies where components (skills, agents, commands) should be linked
type Target interface {
	// GetBaseDir returns the base directory for this target
	GetBaseDir() (string, error)

	// GetSkillsDir returns the directory where skills should be linked
	GetSkillsDir() (string, error)

	// GetAgentsDir returns the directory where agents should be linked
	GetAgentsDir() (string, error)

	// GetCommandsDir returns the directory where commands should be linked
	GetCommandsDir() (string, error)

	// GetComponentDir returns the directory for a specific component type
	GetComponentDir(componentType string) (string, error)

	// GetDetectionConfigPath returns the path to the detection config file
	GetDetectionConfigPath() (string, error)

	// GetName returns the human-readable name of this target
	GetName() string
}
