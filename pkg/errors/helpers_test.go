package errors

import (
	"fmt"
	"strings"
	"testing"

	"github.com/tjg184/agent-smith/pkg/config"
)

// TestNewAgentsDirectoryError verifies helpful error messages for agents directory issues
func TestNewAgentsDirectoryError(t *testing.T) {
	tests := []struct {
		name            string
		cause           error
		expectedPhrases []string
	}{
		{
			name:  "permission denied",
			cause: fmt.Errorf("permission denied"),
			expectedPhrases: []string{
				"Failed to access agents directory",
				"permission",
				"~/.agent-smith/",
			},
		},
		{
			name:  "not found",
			cause: fmt.Errorf("no such file or directory"),
			expectedPhrases: []string{
				"Failed to access agents directory",
				"not be initialized",
				"install",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAgentsDirectoryError(tt.cause)
			errMsg := err.Format()

			for _, phrase := range tt.expectedPhrases {
				if !strings.Contains(errMsg, phrase) {
					t.Errorf("Expected error to contain %q, but got:\n%s", phrase, errMsg)
				}
			}
		})
	}
}

// TestNewTargetDetectionError verifies helpful error messages for target detection
func TestNewTargetDetectionError(t *testing.T) {
	cause := fmt.Errorf("no targets found")
	err := NewTargetDetectionError(cause)
	errMsg := err.Format()

	expectedPhrases := []string{
		"Failed to detect installation targets",
		"OpenCode",
		"Claude Code",
		"custom target",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(errMsg, phrase) {
			t.Errorf("Expected error to contain %q, but got:\n%s", phrase, errMsg)
		}
	}
}

// TestNewProfileNotFoundError verifies error includes profile name and helpful suggestion
func TestNewProfileNotFoundError(t *testing.T) {
	profileName := "my-project"
	err := NewProfileNotFoundError(profileName)
	errMsg := err.Format()

	expectedPhrases := []string{
		"my-project",
		"does not exist",
		"Create the profile first",
		"agent-smith profile create",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(errMsg, phrase) {
			t.Errorf("Expected error to contain %q, but got:\n%s", phrase, errMsg)
		}
	}
}

// TestNewComponentDownloadError verifies helpful suggestions for download failures
func TestNewComponentDownloadError(t *testing.T) {
	tests := []struct {
		name            string
		cause           error
		expectedPhrases []string
	}{
		{
			name:  "authentication failure",
			cause: fmt.Errorf("authentication failed"),
			expectedPhrases: []string{
				"Failed to download",
				"Git credentials",
				"repository access permissions",
			},
		},
		{
			name:  "404 not found",
			cause: fmt.Errorf("repository not found: 404"),
			expectedPhrases: []string{
				"Failed to download",
				"repository URL is correct",
				"accessible",
			},
		},
		{
			name:  "network timeout",
			cause: fmt.Errorf("network timeout"),
			expectedPhrases: []string{
				"Failed to download",
				"network connection",
				"try again",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewComponentDownloadError("skill", "https://github.com/user/repo", tt.cause)
			errMsg := err.Format()

			for _, phrase := range tt.expectedPhrases {
				if !strings.Contains(errMsg, phrase) {
					t.Errorf("Expected error to contain %q, but got:\n%s", phrase, errMsg)
				}
			}
		})
	}
}

// TestNewComponentLinkerError verifies helpful suggestions for linking failures
func TestNewComponentLinkerError(t *testing.T) {
	tests := []struct {
		name            string
		cause           error
		expectedPhrases []string
	}{
		{
			name:  "permission denied",
			cause: fmt.Errorf("permission denied"),
			expectedPhrases: []string{
				"Failed to link",
				"write permissions",
				"target directory",
			},
		},
		{
			name:  "already exists",
			cause: fmt.Errorf("file already exists"),
			expectedPhrases: []string{
				"Failed to link",
				"may already be linked",
				"unlinking first",
				"agent-smith unlink",
			},
		},
		{
			name:  "component not found",
			cause: fmt.Errorf("component not found"),
			expectedPhrases: []string{
				"Failed to link",
				"installed before linking",
				"agent-smith install",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewComponentLinkerError("skill", "OpenCode", tt.cause)
			errMsg := err.Format()

			for _, phrase := range tt.expectedPhrases {
				if !strings.Contains(errMsg, phrase) {
					t.Errorf("Expected error to contain %q, but got:\n%s", phrase, errMsg)
				}
			}
		})
	}
}

// TestNewMaterializationError verifies materialization errors are helpful
func TestNewMaterializationError(t *testing.T) {
	tests := []struct {
		name            string
		cause           error
		expectedPhrases []string
	}{
		{
			name:  "component not found",
			cause: fmt.Errorf("component does not exist"),
			expectedPhrases: []string{
				"Failed to materialize",
				"my-skill",
				"installed before materializing",
				"agent-smith install",
			},
		},
		{
			name:  "permission denied",
			cause: fmt.Errorf("permission denied"),
			expectedPhrases: []string{
				"Failed to materialize",
				"write permissions",
				"project directory",
			},
		},
		{
			name:  "already exists",
			cause: fmt.Errorf("already exists"),
			expectedPhrases: []string{
				"Failed to materialize",
				"--force",
				"overwrite",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewMaterializationError("skill", "my-skill", tt.cause)
			errMsg := err.Format()

			for _, phrase := range tt.expectedPhrases {
				if !strings.Contains(errMsg, phrase) {
					t.Errorf("Expected error to contain %q, but got:\n%s", phrase, errMsg)
				}
			}
		})
	}
}

// TestNewLockFileError verifies lock file errors include helpful recovery steps
func TestNewLockFileError(t *testing.T) {
	tests := []struct {
		name            string
		cause           error
		expectedPhrases []string
	}{
		{
			name:  "corrupted lock file",
			cause: fmt.Errorf("invalid YAML: unmarshal error"),
			expectedPhrases: []string{
				"Failed to",
				"lock file",
				"corrupted",
				"reinstalling",
			},
		},
		{
			name:  "permission denied",
			cause: fmt.Errorf("permission denied"),
			expectedPhrases: []string{
				"Failed to",
				"lock file",
				"file permissions",
				"agents directory",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewLockFileError("read", "skills", tt.cause)
			errMsg := err.Format()

			for _, phrase := range tt.expectedPhrases {
				if !strings.Contains(errMsg, phrase) {
					t.Errorf("Expected error to contain %q, but got:\n%s", phrase, errMsg)
				}
			}
		})
	}
}

// TestNewProjectDetectionError verifies project detection errors are clear
func TestNewProjectDetectionError(t *testing.T) {
	cause := fmt.Errorf("no .opencode or .claude directory found")
	err := NewProjectDetectionError(cause)
	errMsg := err.Format()

	expectedPhrases := []string{
		"No project directory found",
		".opencode/",
		".claude/",
		"Navigate to a project directory",
		"mkdir",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(errMsg, phrase) {
			t.Errorf("Expected error to contain %q, but got:\n%s", phrase, errMsg)
		}
	}
}

// TestErrorFormattingWithoutColors verifies error messages work without colors
func TestErrorFormattingWithoutColors(t *testing.T) {
	// Disable colors for testing
	Disable()
	defer Enable()

	err := NewProfileNotFoundError("test-profile")
	errMsg := err.Format()

	// Should still have the content even without colors
	if !strings.Contains(errMsg, "Profile 'test-profile' does not exist") {
		t.Errorf("Error message missing main content: %s", errMsg)
	}
	if !strings.Contains(errMsg, "Context:") {
		t.Errorf("Error message missing context section: %s", errMsg)
	}
	if !strings.Contains(errMsg, "Suggestion:") {
		t.Errorf("Error message missing suggestion section: %s", errMsg)
	}
	if !strings.Contains(errMsg, "agent-smith profile create") {
		t.Errorf("Error message missing example: %s", errMsg)
	}
}

// TestErrorMessagesContainActionableInformation verifies all errors provide clear next steps
func TestErrorMessagesContainActionableInformation(t *testing.T) {
	errorFuncs := []struct {
		name string
		fn   func() *ErrorMessage
	}{
		{
			name: "NewProfileNotFoundError",
			fn:   func() *ErrorMessage { return NewProfileNotFoundError("test") },
		},
		{
			name: "NewNoActiveProfileError",
			fn:   func() *ErrorMessage { return NewNoActiveProfileError() },
		},
		{
			name: "NewTargetNotFoundError",
			fn:   func() *ErrorMessage { return NewTargetNotFoundError("test") },
		},
		{
			name: "NewAgentsDirectoryError",
			fn:   func() *ErrorMessage { return NewAgentsDirectoryError(fmt.Errorf("test error")) },
		},
		{
			name: "NewTargetDetectionError",
			fn:   func() *ErrorMessage { return NewTargetDetectionError(fmt.Errorf("test error")) },
		},
		{
			name: "NewProjectDetectionError",
			fn:   func() *ErrorMessage { return NewProjectDetectionError(fmt.Errorf("test error")) },
		},
	}

	for _, tt := range errorFuncs {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			errMsg := err.Format()

			// All errors should have:
			// 1. A clear message
			if err.Message == "" {
				t.Error("Error message is empty")
			}

			// 2. Context or suggestion (at least one)
			if err.Context == "" && err.Suggestion == "" {
				t.Error("Error has neither context nor suggestion - not helpful for developers")
			}

			// 3. The formatted message should have content
			if len(errMsg) < 50 {
				t.Errorf("Formatted error message seems too short to be helpful: %s", errMsg)
			}
		})
	}
}

// TestNewComponentNotFoundInProjectError verifies component not found error is helpful
func TestNewComponentNotFoundInProjectError(t *testing.T) {
	tests := []struct {
		name                string
		componentType       string
		componentName       string
		availableComponents []string
		expectedPhrases     []string
	}{
		{
			name:                "with available components",
			componentType:       "skill",
			componentName:       "my-skill",
			availableComponents: []string{"skill-a", "skill-b", "skill-c"},
			expectedPhrases: []string{
				"my-skill",
				"not materialized",
				"available components",
				"skill-a",
				"agent-smith materialize",
			},
		},
		{
			name:                "without available components",
			componentType:       "agent",
			componentName:       "my-agent",
			availableComponents: []string{},
			expectedPhrases: []string{
				"my-agent",
				"not materialized",
				"Materialize a component",
				"agent-smith materialize",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewComponentNotFoundInProjectError(tt.componentType, tt.componentName, tt.availableComponents)
			errMsg := err.Format()

			for _, phrase := range tt.expectedPhrases {
				if !strings.Contains(errMsg, phrase) {
					t.Errorf("Expected error to contain %q, but got:\n%s", phrase, errMsg)
				}
			}
		})
	}
}

// TestNewComponentNotInstalledError verifies component not installed error is helpful
func TestNewComponentNotInstalledError(t *testing.T) {
	err := NewComponentNotInstalledError("skill", "my-skill", "~/.agent-smith/skills/")
	errMsg := err.Format()

	expectedPhrases := []string{
		"my-skill",
		"not found",
		"must be installed",
		"agent-smith install",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(errMsg, phrase) {
			t.Errorf("Expected error to contain %q, but got:\n%s", phrase, errMsg)
		}
	}
}

// TestNewMissingTargetFlagError verifies missing target flag error is helpful
func TestNewMissingTargetFlagError(t *testing.T) {
	err := NewMissingTargetFlagError("materialize skill my-skill")
	errMsg := err.Format()

	expectedPhrases := []string{
		"--target flag",
		"AGENT_SMITH_TARGET",
		"opencode",
		"claudecode",
		"agent-smith",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(errMsg, phrase) {
			t.Errorf("Expected error to contain %q, but got:\n%s", phrase, errMsg)
		}
	}
}

// TestNewInvalidTargetError verifies invalid target error is helpful
func TestNewInvalidTargetError(t *testing.T) {
	err := NewInvalidTargetError("invalid-target")
	errMsg := err.Format()

	expectedPhrases := []string{
		"Invalid target",
		"invalid-target",
		"opencode",
		"claudecode",
		"all",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(errMsg, phrase) {
			t.Errorf("Expected error to contain %q, but got:\n%s", phrase, errMsg)
		}
	}
}

// TestNewTargetDirectoryNotFoundError verifies target directory not found error is helpful
func TestNewTargetDirectoryNotFoundError(t *testing.T) {
	tests := []struct {
		name            string
		targetName      string
		expectedPhrases []string
	}{
		{
			name:       "opencode target",
			targetName: "opencode",
			expectedPhrases: []string{
				"Target directory not found",
				"opencode",
				".opencode/",
				"mkdir",
			},
		},
		{
			name:       "claudecode target",
			targetName: "claudecode",
			expectedPhrases: []string{
				"Target directory not found",
				"Claude Code",
				".claude/",
				"mkdir",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, err := config.NewTarget(tt.targetName)
			if err != nil {
				t.Fatalf("failed to create target %q: %v", tt.targetName, err)
			}
			em := NewTargetDirectoryNotFoundError(target)
			errMsg := em.Format()

			for _, phrase := range tt.expectedPhrases {
				if !strings.Contains(errMsg, phrase) {
					t.Errorf("Expected error to contain %q, but got:\n%s", phrase, errMsg)
				}
			}
		})
	}
}
