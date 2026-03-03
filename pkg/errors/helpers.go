// Package errors provides colored error formatting with context for better user experience.
package errors

import (
	"fmt"
)

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

// NewAgentsDirectoryError creates an error for when the agents directory cannot be accessed.
func NewAgentsDirectoryError(cause error) *ErrorMessage {
	msg := New("Failed to access agents directory")

	if cause != nil {
		msg.WithContext(fmt.Sprintf("Error: %v", cause))

		causeStr := cause.Error()
		if contains(causeStr, "permission") {
			msg.WithSuggestion("Check that you have read/write permissions to ~/.agent-smith/")
		} else if contains(causeStr, "not found") || contains(causeStr, "no such file") {
			msg.WithSuggestion("The agents directory may not be initialized. Try installing a component first").
				WithExample("agent-smith install skill <repo-url> <name>")
		}
	}

	return msg
}

// NewTargetDetectionError creates an error for target detection failures.
func NewTargetDetectionError(cause error) *ErrorMessage {
	msg := New("Failed to detect installation targets")

	if cause != nil {
		msg.WithContext(fmt.Sprintf("Error: %v", cause))
		msg.WithSuggestion("Ensure OpenCode, Claude Code, or a custom target is configured properly")
	}

	return msg
}

// NewActiveProfileError creates an error for profile activation issues.
func NewActiveProfileError(cause error) *ErrorMessage {
	msg := New("Failed to get active profile")

	if cause != nil {
		msg.WithContext(fmt.Sprintf("Error: %v", cause))
		msg.WithSuggestion("Check profile configuration or create a new profile").
			WithExample("agent-smith profile list")
	}

	return msg
}

// NewMissingArgumentsError creates an error for missing command arguments.
func NewMissingArgumentsError(command, usage string) *ErrorMessage {
	return New(fmt.Sprintf("Missing required arguments for '%s'", command)).
		WithSuggestion(fmt.Sprintf("Usage: %s", usage)).
		WithExample(fmt.Sprintf("%s --help", command))
}

// NewTooManyArgumentsError creates an error for too many command arguments.
func NewTooManyArgumentsError(command, usage string) *ErrorMessage {
	return New(fmt.Sprintf("Too many arguments provided for '%s'", command)).
		WithSuggestion(fmt.Sprintf("Usage: %s", usage)).
		WithExample(fmt.Sprintf("%s --help", command))
}

// NewUnknownComponentTypeError creates an error for unknown component types.
func NewUnknownComponentTypeError(componentType string) *ErrorMessage {
	return New(fmt.Sprintf("Unknown component type: %s", componentType)).
		WithContext("Component type must be one of: skills, agents, commands").
		WithSuggestion("Check the component type spelling").
		WithExample("agent-smith install skill <repo-url> <name>")
}

// NewLockFileError creates an error for lock file operations.
func NewLockFileError(operation, componentType string, cause error) *ErrorMessage {
	msg := New(fmt.Sprintf("Failed to %s lock file for %s", operation, componentType))

	if cause != nil {
		msg.WithContext(fmt.Sprintf("Error: %v", cause))

		causeStr := cause.Error()
		if contains(causeStr, "permission") {
			msg.WithSuggestion("Check file permissions in the agents directory")
		} else if contains(causeStr, "not found") {
			msg.WithSuggestion(fmt.Sprintf("The %s may not be installed yet", componentType))
		} else if contains(causeStr, "invalid") || contains(causeStr, "unmarshal") {
			msg.WithSuggestion("The lock file may be corrupted. Try reinstalling the component")
		}
	}

	return msg
}

// NewProjectDetectionError creates an error for project detection failures.
func NewProjectDetectionError(cause error) *ErrorMessage {
	msg := New("No project directory found")

	if cause != nil {
		msg.WithContext(fmt.Sprintf("Error: %v", cause))
	}

	return msg.WithContext("A project directory (.opencode/ or .claude/) is required for this operation").
		WithSuggestion("Navigate to a project directory or create one").
		WithExample("mkdir -p .opencode/ && cd .opencode")
}

// NewMaterializationError creates an error for component materialization failures.
func NewMaterializationError(componentType, componentName string, cause error) *ErrorMessage {
	msg := New(fmt.Sprintf("Failed to materialize %s '%s'", componentType, componentName))

	if cause != nil {
		msg.WithContext(fmt.Sprintf("Error: %v", cause))

		causeStr := cause.Error()
		if contains(causeStr, "not found") || contains(causeStr, "does not exist") {
			msg.WithSuggestion(fmt.Sprintf("Ensure the %s is installed before materializing", componentType)).
				WithExample(fmt.Sprintf("agent-smith install %s <repo-url> %s", componentType, componentName))
		} else if contains(causeStr, "permission") {
			msg.WithSuggestion("Check that you have write permissions to the project directory")
		} else if contains(causeStr, "already exists") {
			msg.WithSuggestion("Use --force to overwrite existing components").
				WithExample(fmt.Sprintf("agent-smith materialize %s %s --force", componentType, componentName))
		}
	}

	return msg
}

// NewComponentNotFoundInProjectError creates an error for when a component is not found in the project.
func NewComponentNotFoundInProjectError(componentType, componentName string, availableComponents []string) *ErrorMessage {
	msg := New(fmt.Sprintf("Component '%s' (%s) not materialized in current project", componentName, componentType)).
		WithContext("The component must be materialized before you can view its information")

	if len(availableComponents) > 0 {
		msg.WithSuggestion("Materialize the component or choose from available components:")
		for _, comp := range availableComponents {
			msg.WithDetails(comp)
		}
	} else {
		msg.WithSuggestion("Materialize a component from your installed collection")
	}

	msg.WithExample(fmt.Sprintf("agent-smith materialize %s %s --target opencode", componentType, componentName))

	return msg
}

// NewComponentNotInstalledError creates an error for when a component is not installed.
func NewComponentNotInstalledError(componentType, componentName, source string) *ErrorMessage {
	msg := New(fmt.Sprintf("Component '%s' not found in %s", componentName, source))

	msg.WithContext(fmt.Sprintf("The %s must be installed before it can be materialized", componentType)).
		WithSuggestion(fmt.Sprintf("Install the %s first", componentType)).
		WithExample(fmt.Sprintf("agent-smith install %s <repo-url> %s", componentType, componentName))

	return msg
}

// NewMissingTargetFlagError creates an error for when the --target flag is missing.
func NewMissingTargetFlagError(command string) *ErrorMessage {
	return New("Target must be specified with --target flag or AGENT_SMITH_TARGET environment variable").
		WithContext("Valid targets: opencode, claudecode, copilot, universal, all").
		WithSuggestion("Specify a target using the --target flag or set AGENT_SMITH_TARGET").
		WithExample(fmt.Sprintf("agent-smith %s --target opencode", command)).
		WithDetails("export AGENT_SMITH_TARGET=opencode  # Set default target")
}

// NewInvalidTargetError creates an error for when an invalid target is specified.
func NewInvalidTargetError(targetName string) *ErrorMessage {
	return New(fmt.Sprintf("Invalid target: %s", targetName)).
		WithContext("Valid targets are: opencode, claudecode, copilot, universal, all").
		WithSuggestion("Use one of the valid target names").
		WithExample("agent-smith materialize skill my-skill --target opencode")
}

// NewTargetDirectoryNotFoundError creates an error for when a target directory doesn't exist.
func NewTargetDirectoryNotFoundError(targetName string) *ErrorMessage {
	var dirName string
	switch targetName {
	case "opencode":
		dirName = ".opencode/"
	case "claudecode":
		dirName = ".claude/"
	case "copilot":
		dirName = ".github/"
	case "universal":
		dirName = ".agents/"
	default:
		dirName = targetName
	}

	return New(fmt.Sprintf("Target directory not found: %s", targetName)).
		WithContext(fmt.Sprintf("The %s directory does not exist in this project", dirName)).
		WithSuggestion("Create the directory or materialize to a different target").
		WithExample(fmt.Sprintf("mkdir -p %s", dirName))
}

// NewAmbiguousComponentError creates an error for when a component name exists in multiple sources.
func NewAmbiguousComponentError(componentType, componentName string, sourceURLs []string) *ErrorMessage {
	msg := New(fmt.Sprintf("Component '%s' found in multiple sources", componentName)).
		WithContext("The same component name exists in multiple repositories")

	if len(sourceURLs) > 0 {
		msg.WithDetails("Available sources:")
		for _, url := range sourceURLs {
			msg.WithDetails(fmt.Sprintf("  - %s", url))
		}
	}

	msg.WithSuggestion("Specify the source explicitly using the --source flag").
		WithExample(fmt.Sprintf("agent-smith materialize %s %s --source <source-url>", componentType, componentName))

	return msg
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
