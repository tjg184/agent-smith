# Colored Error Messages with Context

This feature provides colored, contextual error messages to help users quickly understand what went wrong and how to fix it.

## Features

- **Colored Output**: Error messages use colors to distinguish different parts (error icon, context, suggestions)
- **Contextual Information**: Errors include context about what operation failed
- **Helpful Suggestions**: Each error provides actionable suggestions on how to fix the issue
- **Example Commands**: Where applicable, errors show example commands to resolve the issue

## Usage in Code

### Using Pre-built Error Helpers

```go
import (
    "github.com/tgaines/agent-smith/pkg/errors"
    "github.com/tgaines/agent-smith/pkg/logger"
)

// Create logger
log := logger.Default(false, false)

// Use pre-built error helpers
log.FatalMsg(errors.NewProfileNotFoundError("my-profile"))
log.FatalMsg(errors.NewInvalidFlagsError("--profile", "--target-dir"))
log.FatalMsg(errors.NewComponentDownloadError("skill", repoURL, err))
```

### Creating Custom Error Messages

```go
// Simple error
errMsg := errors.New("Failed to parse configuration file")
log.ErrorMsg(errMsg)

// Error with context
errMsg := errors.New("Failed to parse configuration file").
    WithContext("The configuration file contains invalid YAML syntax")
log.ErrorMsg(errMsg)

// Full-featured error
errMsg := errors.New("Failed to parse configuration file").
    WithContext("The configuration file contains invalid YAML syntax").
    WithDetails(
        "Line 15: unexpected character '{'",
        "Expected a valid YAML mapping or sequence",
    ).
    WithSuggestion("Check your YAML syntax using a validator").
    WithExample("yamllint ~/.agent-smith/config.yaml")
log.FatalMsg(errMsg)
```

### Simple Colored Messages

For simple one-line errors without additional context:

```go
// Colored error (automatically adds ✗ icon)
log.Error("Failed to download skill: repository not found")

// Colored warning (automatically adds ⚠ icon)
log.Warn("Profile directory already exists, using existing profile")
```

## Available Error Helpers

The following pre-built error helpers are available in `pkg/errors/helpers.go`:

- `NewProfileNotFoundError(profileName)` - Profile doesn't exist
- `NewInvalidFlagsError(flag1, flag2)` - Conflicting flags
- `NewDirectoryCreationError(dir, cause)` - Directory creation failures
- `NewComponentDownloadError(componentType, repoURL, cause)` - Component download failures
- `NewProfileManagerError(cause)` - Profile manager initialization failures
- `NewComponentLinkerError(componentType, target, cause)` - Component linking failures
- `NewTargetNotFoundError(targetName)` - Target directory not found
- `NewInvalidComponentTypeError(componentType, validTypes)` - Invalid component types
- `NewNoActiveProfileError()` - No profile is active
- `NewFileOperationError(operation, path, cause)` - Generic file operation failures
- `NewGitOperationError(operation, repoURL, cause)` - Git operation failures
- `NewValidationError(field, reason)` - Input validation failures
- `NewConfigurationError(setting, cause)` - Configuration issues

## Example Output

```
✗ Profile 'my-project' does not exist

Context: Profiles must be created before they can be used

Suggestion: Create the profile first

  $ agent-smith profile create my-project
```

## Color Support

Colors are automatically enabled for TTY output and disabled for piped output or when `NO_COLOR` environment variable is set.

To manually control colors:

```go
// Disable colors
errors.Disable()

// Enable colors
errors.Enable()
```
