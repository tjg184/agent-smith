package materialize

import (
	"fmt"
)

type PostprocessorRegistry struct {
	postprocessors []ComponentPostprocessor
}

func NewPostprocessorRegistry() *PostprocessorRegistry {
	return &PostprocessorRegistry{
		postprocessors: []ComponentPostprocessor{},
	}
}

func (r *PostprocessorRegistry) RunPostprocessors(ctx PostprocessContext) error {
	for _, processor := range r.postprocessors {
		if processor.ShouldProcess(ctx.ComponentType, ctx.Target) {
			if err := processor.Process(ctx); err != nil {
				return fmt.Errorf("%s failed: %w", processor.Name(), err)
			}
		}
	}
	return nil
}

func (r *PostprocessorRegistry) RunCleanup(ctx PostprocessContext) error {
	for _, processor := range r.postprocessors {
		if processor.ShouldProcess(ctx.ComponentType, ctx.Target) {
			if err := processor.Cleanup(ctx); err != nil {
				ctx.Formatter.WarningMsg("%s cleanup warning: %v", processor.Name(), err)
			}
		}
	}
	return nil
}
