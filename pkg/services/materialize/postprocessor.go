package materialize

import (
	"github.com/tgaines/agent-smith/internal/formatter"
)

// ComponentPostprocessor handles post-materialization processing for specific
// component type and target combinations. Postprocessors run after a component
// has been copied to its destination and can perform additional operations
// like creating symlinks, generating indexes, or validating structure.
type ComponentPostprocessor interface {
	// ShouldProcess returns true if this postprocessor should run for the given
	// component type and target combination
	ShouldProcess(componentType, target string) bool

	// Process performs post-materialization operations on a component
	// Returns error only for fatal issues; warnings should be logged via Formatter
	Process(ctx PostprocessContext) error

	// Cleanup removes postprocessor artifacts before re-materialization
	// Should never return fatal errors - log warnings instead
	Cleanup(ctx PostprocessContext) error

	// Name returns the name of this postprocessor for logging
	Name() string
}

// PostprocessContext contains all information needed for postprocessing operations
type PostprocessContext struct {
	// ComponentType is the type of component being processed ("skills", "agents", "commands")
	ComponentType string

	// ComponentName is the name of the component (e.g., "my-agent")
	ComponentName string

	// Target is the materialization target (e.g., "copilot", "opencode", "claudecode")
	Target string

	// TargetDir is the target directory path (e.g., "/project/.github")
	TargetDir string

	// DestPath is the full destination path of the materialized component
	// (e.g., "/project/.github/agents/my-agent")
	DestPath string

	// DryRun indicates whether this is a dry-run operation
	// Postprocessors should not modify the filesystem when DryRun is true
	DryRun bool

	// Formatter is used for outputting messages and warnings
	Formatter *formatter.Formatter

	// SymlinkRegistry tracks symlinks created during this materialization run
	// to detect and warn about name conflicts. Key: filename, Value: componentName
	// This is shared across all postprocessors in a single materialization operation
	SymlinkRegistry map[string]string
}
