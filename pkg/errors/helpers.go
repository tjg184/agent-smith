// Package errors provides colored error formatting with context for better user experience.
package errors

import (
	"fmt"
)

// Common error scenarios with helpful context and suggestions

// NewProfileNotFoundError creates an error for when a profile doesn't exist.
func NewProfileNotFoundError(profileName string) *ErrorMessage {
	return New(fmt.Sprintf("Profile '%s' does not exist", profileName)).
		WithContext("Profiles must be created before they can be used").
		WithSuggestion("Create the profile first").
		WithExample(fmt.Sprintf("agent-smith profile create %s", profileName))
}

// NewInvalidFlagsError creates an error for conflicting flags.
func NewInvalidFlagsError(flag1, flag2 string) *ErrorMessage {
	return New(fmt.Sprintf("Cannot specify both %s and %s flags", flag1, flag2)).
		WithContext("These flags are mutually exclusive").
		WithSuggestion(fmt.Sprintf("Use either %s or %s, but not both", flag1, flag2))
}

// NewDirectoryCreationError creates an error for directory creation failures.
func NewDirectoryCreationError(dir string, cause error) *ErrorMessage {
	msg := New(fmt.Sprintf("Failed to create directory: %s", dir)).
		WithContext(fmt.Sprintf("Underlying error: %v", cause))

	if cause != nil {
		causeStr := cause.Error()
		if contains(causeStr, "permission denied") {
			msg.WithSuggestion("Check that you have write permissions to the parent directory")
		} else if contains(causeStr, "no such file") {
			msg.WithSuggestion("Ensure the parent directory exists")
		}
	}

	return msg
}

// NewComponentDownloadError creates an error for component download failures.
func NewComponentDownloadError(componentType, repoURL string, cause error) *ErrorMessage {
	msg := New(fmt.Sprintf("Failed to download %s from repository", componentType)).
		WithContext(fmt.Sprintf("Repository: %s", repoURL))

	if cause != nil {
		msg.WithDetails(fmt.Sprintf("Error: %v", cause))

		causeStr := cause.Error()
		if contains(causeStr, "authentication") || contains(causeStr, "credentials") {
			msg.WithSuggestion("Check your Git credentials and repository access permissions")
		} else if contains(causeStr, "not found") || contains(causeStr, "404") {
			msg.WithSuggestion("Verify the repository URL is correct and accessible")
		} else if contains(causeStr, "network") || contains(causeStr, "timeout") {
			msg.WithSuggestion("Check your network connection and try again")
		} else if contains(causeStr, "no "+componentType) {
			msg.WithSuggestion(fmt.Sprintf("Ensure the repository contains a valid %s", componentType))
		}
	}

	return msg
}

// NewProfileManagerError creates an error for profile manager initialization failures.
func NewProfileManagerError(cause error) *ErrorMessage {
	msg := New("Failed to initialize profile manager")

	if cause != nil {
		msg.WithContext(fmt.Sprintf("Error: %v", cause))

		causeStr := cause.Error()
		if contains(causeStr, "permission") {
			msg.WithSuggestion("Check that you have read/write permissions to ~/.agent-smith/profiles/")
		}
	}

	return msg
}

// NewComponentLinkerError creates an error for component linking failures.
func NewComponentLinkerError(componentType, target string, cause error) *ErrorMessage {
	msg := New(fmt.Sprintf("Failed to link %s to %s", componentType, target))

	if cause != nil {
		msg.WithContext(fmt.Sprintf("Error: %v", cause))

		causeStr := cause.Error()
		if contains(causeStr, "permission") {
			msg.WithSuggestion("Check that you have write permissions to the target directory")
		} else if contains(causeStr, "already exists") {
			msg.WithSuggestion("The component may already be linked. Try unlinking first").
				WithExample(fmt.Sprintf("agent-smith unlink %s <component-name>", componentType))
		} else if contains(causeStr, "not found") {
			msg.WithSuggestion(fmt.Sprintf("Ensure the %s is installed before linking", componentType)).
				WithExample(fmt.Sprintf("agent-smith install %s <repo-url> <name>", componentType))
		}
	}

	return msg
}

// NewTargetNotFoundError creates an error for when a target directory is not found.
func NewTargetNotFoundError(targetName string) *ErrorMessage {
	return New(fmt.Sprintf("Target '%s' not found or not configured", targetName)).
		WithContext("Supported targets: OpenCode, Claude Code, or custom targets").
		WithSuggestion("List available targets or create a custom target").
		WithExample("agent-smith target list")
}

// NewInvalidComponentTypeError creates an error for invalid component types.
func NewInvalidComponentTypeError(componentType string, validTypes []string) *ErrorMessage {
	msg := New(fmt.Sprintf("Invalid component type: %s", componentType)).
		WithContext("Valid component types are:")

	for _, t := range validTypes {
		msg.WithDetails(t)
	}

	return msg
}

// NewNoActiveProfileError creates an error for when no profile is active.
func NewNoActiveProfileError() *ErrorMessage {
	return New("No active profile found").
		WithContext("You need an active profile to link components").
		WithSuggestion("Create and activate a profile first").
		WithExample("agent-smith profile create my-profile && agent-smith profile activate my-profile")
}

// NewFileOperationError creates an error for generic file operation failures.
func NewFileOperationError(operation, path string, cause error) *ErrorMessage {
	msg := New(fmt.Sprintf("Failed to %s: %s", operation, path))

	if cause != nil {
		msg.WithContext(fmt.Sprintf("Error: %v", cause))
	}

	return msg
}

// NewGitOperationError creates an error for git operation failures.
func NewGitOperationError(operation, repoURL string, cause error) *ErrorMessage {
	msg := New(fmt.Sprintf("Git operation failed: %s", operation)).
		WithContext(fmt.Sprintf("Repository: %s", repoURL))

	if cause != nil {
		msg.WithDetails(fmt.Sprintf("Error: %v", cause))

		causeStr := cause.Error()
		if contains(causeStr, "authentication") {
			msg.WithSuggestion("Configure your Git credentials or use HTTPS authentication")
		} else if contains(causeStr, "ssh") {
			msg.WithSuggestion("Check your SSH key configuration or use HTTPS instead")
		}
	}

	return msg
}

// NewValidationError creates an error for input validation failures.
func NewValidationError(field, reason string) *ErrorMessage {
	return New(fmt.Sprintf("Invalid %s: %s", field, reason)).
		WithContext("Input validation failed")
}

// NewConfigurationError creates an error for configuration issues.
func NewConfigurationError(setting string, cause error) *ErrorMessage {
	msg := New(fmt.Sprintf("Configuration error: %s", setting))

	if cause != nil {
		msg.WithContext(fmt.Sprintf("Error: %v", cause))
	}

	return msg.WithSuggestion("Check your configuration files and settings")
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && containsIgnoreCase(s, substr)))
}

func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}
