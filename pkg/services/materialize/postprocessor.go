package materialize

import (
	"github.com/tjg184/agent-smith/internal/formatter"
)

type ComponentPostprocessor interface {
	ShouldProcess(componentType, target string) bool
	Process(ctx PostprocessContext) error
	Cleanup(ctx PostprocessContext) error
	Name() string
}

type PostprocessContext struct {
	ComponentType   string
	ComponentName   string
	FilesystemName  string
	Target          string
	TargetDir       string
	DestPath        string
	DryRun          bool
	Formatter       *formatter.Formatter
	SymlinkRegistry map[string]string
}
