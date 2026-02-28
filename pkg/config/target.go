package config

// Target defines the interface for component linking targets
// A target specifies where components (skills, agents, commands) should be linked
type Target interface {
	// GetGlobalBaseDir returns the global/base directory for this target (e.g., ~/.config/opencode)
	GetGlobalBaseDir() (string, error)

	// GetGlobalSkillsDir returns the global directory where skills should be linked
	GetGlobalSkillsDir() (string, error)

	// GetGlobalAgentsDir returns the global directory where agents should be linked
	GetGlobalAgentsDir() (string, error)

	// GetGlobalCommandsDir returns the global directory where commands should be linked
	GetGlobalCommandsDir() (string, error)

	// GetGlobalComponentDir returns the global directory for a specific component type
	GetGlobalComponentDir(componentType string) (string, error)

	// GetDetectionConfigPath returns the path to the detection config file
	GetDetectionConfigPath() (string, error)

	// GetName returns the human-readable name of this target
	GetName() string

	// GetProjectDirName returns the directory name used in projects (e.g., ".opencode", ".claude")
	GetProjectDirName() string

	// GetProjectBaseDir returns the base directory within a project
	GetProjectBaseDir(projectRoot string) string

	// GetProjectComponentDir returns the component directory within a project
	GetProjectComponentDir(projectRoot, componentType string) (string, error)

	// IsUniversalTarget returns true for target-agnostic storage (.agents)
	IsUniversalTarget() bool
}
