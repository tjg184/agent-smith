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

// ComponentLockEntry represents a single entry in the lock file
type ComponentLockEntry struct {
	Source       string `json:"source"`
	SourceType   string `json:"sourceType"`
	SourceUrl    string `json:"sourceUrl"`
	SkillPath    string `json:"skillPath,omitempty"`
	OriginalPath string `json:"originalPath,omitempty"` // Original path in repo (e.g., "plugins/ui-design/agents/expert.md")
	CommitHash   string `json:"commitHash"`
	InstalledAt  string `json:"installedAt"`
	UpdatedAt    string `json:"updatedAt"`
	Version      int    `json:"version"`
	Components   int    `json:"components,omitempty"`
	Detection    string `json:"detection,omitempty"`
}

// ComponentLockFile tracks all installed components
// Version 4+ uses nested structure: map[sourceURL]map[componentName]ComponentLockEntry
type ComponentLockFile struct {
	Version  int                                      `json:"version"`
	Skills   map[string]map[string]ComponentLockEntry `json:"skills"`
	Agents   map[string]map[string]ComponentLockEntry `json:"agents,omitempty"`
	Commands map[string]map[string]ComponentLockEntry `json:"commands,omitempty"`
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
