package materialize

import (
	"testing"

	"github.com/tgaines/agent-smith/internal/formatter"
)

// MockPostprocessor is a test postprocessor for testing the registry
type MockPostprocessor struct {
	name            string
	shouldProcessFn func(componentType, target string) bool
	processCalled   bool
	cleanupCalled   bool
	processError    error
	cleanupError    error
}

func (m *MockPostprocessor) Name() string {
	return m.name
}

func (m *MockPostprocessor) ShouldProcess(componentType, target string) bool {
	if m.shouldProcessFn != nil {
		return m.shouldProcessFn(componentType, target)
	}
	return true
}

func (m *MockPostprocessor) Process(ctx PostprocessContext) error {
	m.processCalled = true
	return m.processError
}

func (m *MockPostprocessor) Cleanup(ctx PostprocessContext) error {
	m.cleanupCalled = true
	return m.cleanupError
}

func TestPostprocessorRegistry_RunPostprocessors(t *testing.T) {
	tests := []struct {
		name          string
		postprocessor *MockPostprocessor
		ctx           PostprocessContext
		wantCalled    bool
		wantError     bool
	}{
		{
			name: "postprocessor should run when ShouldProcess returns true",
			postprocessor: &MockPostprocessor{
				name: "TestProcessor",
				shouldProcessFn: func(componentType, target string) bool {
					return componentType == "agents" && target == "copilot"
				},
			},
			ctx: PostprocessContext{
				ComponentType: "agents",
				Target:        "copilot",
			},
			wantCalled: true,
			wantError:  false,
		},
		{
			name: "postprocessor should not run when ShouldProcess returns false",
			postprocessor: &MockPostprocessor{
				name: "TestProcessor",
				shouldProcessFn: func(componentType, target string) bool {
					return componentType == "agents" && target == "copilot"
				},
			},
			ctx: PostprocessContext{
				ComponentType: "skills",
				Target:        "opencode",
			},
			wantCalled: false,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := &PostprocessorRegistry{
				postprocessors: []ComponentPostprocessor{tt.postprocessor},
			}

			err := registry.RunPostprocessors(tt.ctx)

			if (err != nil) != tt.wantError {
				t.Errorf("RunPostprocessors() error = %v, wantError %v", err, tt.wantError)
			}

			if tt.postprocessor.processCalled != tt.wantCalled {
				t.Errorf("Process() called = %v, want %v", tt.postprocessor.processCalled, tt.wantCalled)
			}
		})
	}
}

func TestPostprocessorRegistry_RunCleanup(t *testing.T) {
	mock := &MockPostprocessor{
		name: "TestProcessor",
		shouldProcessFn: func(componentType, target string) bool {
			return true
		},
	}

	registry := &PostprocessorRegistry{
		postprocessors: []ComponentPostprocessor{mock},
	}

	ctx := PostprocessContext{
		ComponentType: "agents",
		Target:        "copilot",
		Formatter:     formatter.New(),
	}

	err := registry.RunCleanup(ctx)

	// Cleanup should never return error (errors are logged)
	if err != nil {
		t.Errorf("RunCleanup() should never return error, got: %v", err)
	}

	if !mock.cleanupCalled {
		t.Error("Cleanup() was not called")
	}
}

func TestPostprocessorRegistry_MultiplePostprocessors(t *testing.T) {
	mock1 := &MockPostprocessor{
		name: "Processor1",
		shouldProcessFn: func(componentType, target string) bool {
			return componentType == "agents"
		},
	}

	mock2 := &MockPostprocessor{
		name: "Processor2",
		shouldProcessFn: func(componentType, target string) bool {
			return target == "copilot"
		},
	}

	registry := &PostprocessorRegistry{
		postprocessors: []ComponentPostprocessor{mock1, mock2},
	}

	ctx := PostprocessContext{
		ComponentType: "agents",
		Target:        "copilot",
	}

	err := registry.RunPostprocessors(ctx)

	if err != nil {
		t.Errorf("RunPostprocessors() error = %v", err)
	}

	// Both should be called
	if !mock1.processCalled {
		t.Error("Processor1 Process() was not called")
	}
	if !mock2.processCalled {
		t.Error("Processor2 Process() was not called")
	}
}
