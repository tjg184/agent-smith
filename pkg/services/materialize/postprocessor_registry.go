package materialize

import (
	"fmt"
)

// PostprocessorRegistry manages all available component postprocessors
// and coordinates their execution during materialization operations
type PostprocessorRegistry struct {
	postprocessors []ComponentPostprocessor
}

// NewPostprocessorRegistry creates a new registry with all available postprocessors
func NewPostprocessorRegistry() *PostprocessorRegistry {
	return &PostprocessorRegistry{
		postprocessors: []ComponentPostprocessor{
			NewAgentFlattenPostprocessor(),
			// Future postprocessors can be added here:
			// NewSkillIndexPostprocessor(),
			// NewCommandValidationPostprocessor(),
		},
	}
}

// RunPostprocessors executes all applicable postprocessors for the given context
// Postprocessors are executed in registration order
func (r *PostprocessorRegistry) RunPostprocessors(ctx PostprocessContext) error {
	for _, processor := range r.postprocessors {
		if processor.ShouldProcess(ctx.ComponentType, ctx.Target) {
			if err := processor.Process(ctx); err != nil {
				// Fatal error from postprocessor
				return fmt.Errorf("%s failed: %w", processor.Name(), err)
			}
		}
	}
	return nil
}

// RunCleanup executes cleanup for all applicable postprocessors
// Cleanup is run before re-materializing a component (e.g., with --force flag)
// Cleanup errors are logged as warnings but never fail the operation
func (r *PostprocessorRegistry) RunCleanup(ctx PostprocessContext) error {
	for _, processor := range r.postprocessors {
		if processor.ShouldProcess(ctx.ComponentType, ctx.Target) {
			if err := processor.Cleanup(ctx); err != nil {
				// Log but don't fail on cleanup errors
				ctx.Formatter.WarningMsg("%s cleanup warning: %v", processor.Name(), err)
			}
		}
	}
	return nil
}
