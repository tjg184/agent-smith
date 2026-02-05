package models

// ComponentType represents the type of component
type ComponentType string

const (
	ComponentSkill   ComponentType = "skill"
	ComponentAgent   ComponentType = "agent"
	ComponentCommand ComponentType = "command"
)

// ComponentDetectionPattern defines how to detect a component type
type ComponentDetectionPattern struct {
	Name           string   `json:"name"`
	ExactFiles     []string `json:"exactFiles"`     // Files that must match exactly (e.g., "SKILL.md")
	PathPatterns   []string `json:"pathPatterns"`   // Path patterns (e.g., "/agents/", "*/docs/*")
	FileExtensions []string `json:"fileExtensions"` // File extensions to match (e.g., ".md")
	IgnorePaths    []string `json:"ignorePaths"`    // Paths to ignore during detection
}

// DetectionConfig holds all component detection patterns
type DetectionConfig struct {
	Components map[string]ComponentDetectionPattern `json:"components"`
}

// DetectedComponent represents a detected component in a repository
type DetectedComponent struct {
	Type       ComponentType
	Name       string
	Path       string // Relative path to component directory
	SourceFile string // Source file name
	FilePath   string // Full relative path from repo root (including filename)
}

// ComponentFrontmatter represents YAML frontmatter metadata for agents/commands
type ComponentFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Model       string `yaml:"model"`
	Mode        string `yaml:"mode"`
}

// ComponentEntry represents a single component entry in unified lock files
// Used for both installs (~/.agent-smith/.component-lock.json) and
// materializations (project/.component-lock.json)
// Version 5+ unified format
type ComponentEntry struct {
	// Core metadata
	Source       string `json:"source"`
	SourceType   string `json:"sourceType"`
	SourceUrl    string `json:"sourceUrl"`
	OriginalPath string `json:"originalPath,omitempty"` // Original path in repo (e.g., "plugins/ui-design/agents/expert.md")
	CommitHash   string `json:"commitHash"`

	// Timestamps (semantics vary by context)
	InstalledAt    string `json:"installedAt,omitempty"`    // When installed to ~/.agent-smith
	MaterializedAt string `json:"materializedAt,omitempty"` // When copied to project
	UpdatedAt      string `json:"updatedAt,omitempty"`      // Last update time

	// Drift detection
	SourceHash  string `json:"sourceHash,omitempty"`  // Hash at install/materialize time
	CurrentHash string `json:"currentHash,omitempty"` // Current hash (detect modifications)

	// Location/tracking
	FilesystemName string `json:"filesystemName,omitempty"` // Actual directory name on disk (handles conflicts)
	SourceProfile  string `json:"sourceProfile,omitempty"`  // Which profile (for materialization)

	// Install-specific metadata
	Components int    `json:"components,omitempty"` // Component count
	Detection  string `json:"detection,omitempty"`  // How detected (auto/manual)

	Version int `json:"version"` // Entry version
}

// ComponentLockEntry is deprecated - use ComponentEntry instead
// Kept for backward compatibility
type ComponentLockEntry = ComponentEntry

// ComponentLockFile tracks all components (installs and materializations)
// Version 5+ uses unified ComponentEntry structure
type ComponentLockFile struct {
	Version  int                                  `json:"version"`
	Skills   map[string]map[string]ComponentEntry `json:"skills"`
	Agents   map[string]map[string]ComponentEntry `json:"agents,omitempty"`
	Commands map[string]map[string]ComponentEntry `json:"commands,omitempty"`
}

// ComponentMetadata is a legacy metadata structure for backward compatibility
type ComponentMetadata struct {
	Name         string `json:"name"`
	Source       string `json:"source"`
	Commit       string `json:"commit"`
	Downloaded   string `json:"downloaded"`
	Components   int    `json:"components,omitempty"`
	Detection    string `json:"detection,omitempty"`
	OriginalPath string `json:"originalPath,omitempty"` // Original path in repo
}
