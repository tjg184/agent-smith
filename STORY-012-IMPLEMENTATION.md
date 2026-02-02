# Story-012: Helpful Error Messages Implementation

## Overview
This story implements comprehensive, developer-friendly error messages throughout the agent-smith CLI tool. The error handling system provides clear context, actionable suggestions, and examples to help developers quickly resolve issues.

## Implementation Details

### 1. Error Message Framework (`pkg/errors/`)

The error handling system is built around a structured `ErrorMessage` type that supports:

- **Main error message**: Clear, concise description of what went wrong
- **Context**: Additional information about why the error occurred
- **Suggestions**: Actionable steps to resolve the issue
- **Examples**: Concrete command examples showing how to fix the problem
- **Details**: Additional bullet points with relevant information
- **Color-coded output**: Visual distinction between errors, warnings, info, and suggestions

#### Core Components

**File: `pkg/errors/errors.go`**
- `ErrorMessage` struct with formatting support
- Builder pattern methods (`WithContext`, `WithSuggestion`, `WithExample`, `WithDetails`)
- Colored output using the `pkg/colors` package
- Support for both errors and warnings

**File: `pkg/errors/helpers.go`**
- Pre-built error constructors for common scenarios:
  - `NewProfileNotFoundError` - Profile doesn't exist
  - `NewComponentDownloadError` - Component download failures with smart suggestions
  - `NewComponentLinkerError` - Linking failures with permission/existence checks
  - `NewMaterializationError` - Component materialization issues
  - `NewLockFileError` - Lock file corruption or access issues
  - `NewTargetNotFoundError` - Invalid or missing targets
  - `NewGitOperationError` - Git operation failures
  - And many more...

### 2. File Operation Error Messages (`internal/fileutil/`)

**File: `internal/fileutil/fileutil.go`**

Enhanced error messages for file operations:

```go
// Before (generic):
return fmt.Errorf("failed to read file: %w", err)

// After (helpful):
if os.IsNotExist(err) {
    return fmt.Errorf("cannot copy file: source file does not exist: %s", src)
}
if os.IsPermission(err) {
    return fmt.Errorf("cannot copy file: permission denied reading source file: %s", src)
}
```

Key improvements:
- **Specific error types**: Distinguishes between "not found", "permission denied", "corrupted file", etc.
- **File paths included**: Always shows the problematic path
- **Actionable messages**: Tells users exactly what's wrong and where

### 3. Error Message Testing

**File: `pkg/errors/helpers_test.go`**
- Comprehensive test coverage for all error helper functions
- Validates that errors contain expected phrases
- Tests both colored and non-colored output
- Ensures all errors provide actionable information

**File: `internal/fileutil/error_messages_test.go`**
- Tests file operation error messages
- Validates context is provided
- Ensures file paths are included in error messages

## Error Message Examples

### Example 1: Profile Not Found

```bash
$ agent-smith profile activate work
```

**Output:**
```
✗ Profile 'work' does not exist

Context: Profiles must be created before they can be used

Suggestion: Create the profile first

  $ agent-smith profile create work
```

### Example 2: Component Download Failure

```bash
$ agent-smith install skill owner/private-repo skill-name
```

**Output:**
```
✗ Failed to download skill from repository

Repository: https://github.com/owner/private-repo

  • Error: authentication failed

Suggestion: Check your Git credentials and repository access permissions
```

### Example 3: Component Not Found

```bash
$ agent-smith link skill nonexistent-skill
```

**Output:**
```
✗ component skills/nonexistent-skill does not exist in any profile

Suggestion: Install the skill first

  $ agent-smith install skill <repo-url> nonexistent-skill
```

### Example 4: Target Directory Not Found

```bash
$ agent-smith materialize skill my-skill --target opencode
```

**Output:**
```
✗ Target directory not found: opencode

Context: The .opencode/ directory does not exist in this project

Suggestion: Create the directory or materialize to a different target

  $ mkdir -p .opencode/
```

### Example 5: File Copy Error

```go
// Code that triggers the error:
err := fileutil.CopyFile("/nonexistent/file.txt", "/dest/file.txt")
```

**Output:**
```
cannot copy file: source file does not exist: /nonexistent/file.txt
```

### Example 6: Lock File Corruption

```bash
$ agent-smith link all
```

**Output:**
```
✗ Failed to read lock file for skills

Error: invalid YAML: unmarshal error

Suggestion: The lock file may be corrupted. Try reinstalling the component
```

## Best Practices for Error Messages

### 1. Always Include Context
```go
// ❌ Bad
return fmt.Errorf("operation failed")

// ✅ Good
return errors.New("Failed to link component").
    WithContext("The component directory does not exist").
    WithSuggestion("Install the component before linking")
```

### 2. Provide Actionable Suggestions
```go
// ❌ Bad - No suggestion
errors.New("Target not found")

// ✅ Good - With actionable suggestion
errors.NewTargetNotFoundError("opencode").
    WithSuggestion("Create the target directory first").
    WithExample("mkdir -p .opencode/")
```

### 3. Show Concrete Examples
```go
// ❌ Bad - Vague help
WithSuggestion("Use the install command")

// ✅ Good - Concrete example
WithSuggestion("Install the component first").
    WithExample("agent-smith install skill github.com/user/repo skill-name")
```

### 4. Include Relevant Details
```go
// When multiple items are involved:
err.WithDetails(
    "Profile: work",
    "Component: my-skill",
    "Target: opencode"
)
```

### 5. Use Specific Error Types
```go
// Use pre-built helpers for common scenarios:
NewProfileNotFoundError("work")
NewComponentDownloadError("skill", repoURL, cause)
NewTargetNotFoundError("opencode")
NewGitOperationError("clone", repoURL, cause)
```

## Testing Error Messages

All error messages should be tested to ensure they:
1. Contain the expected phrases
2. Include relevant context (paths, names, etc.)
3. Provide actionable suggestions
4. Include examples where appropriate

**Example Test:**
```go
func TestNewProfileNotFoundError(t *testing.T) {
    err := errors.NewProfileNotFoundError("work")
    errMsg := err.Format()
    
    expectedPhrases := []string{
        "work",
        "does not exist",
        "Create the profile first",
        "agent-smith profile create",
    }
    
    for _, phrase := range expectedPhrases {
        if !strings.Contains(errMsg, phrase) {
            t.Errorf("Expected error to contain %q", phrase)
        }
    }
}
```

## Error Handling Coverage

### Fully Covered Areas ✅
- ✅ Profile management (create, activate, not found)
- ✅ Component installation (download, link, materialize)
- ✅ File operations (copy, permissions, not found)
- ✅ Git operations (clone, authentication)
- ✅ Lock file operations (read, write, corruption)
- ✅ Target management (detection, directory not found)
- ✅ Configuration errors

### File Operation Errors ✅
- ✅ Source file/directory does not exist
- ✅ Destination directory does not exist
- ✅ Permission denied (read/write)
- ✅ Source is not a directory
- ✅ Component file does not exist

### Network/Repository Errors ✅
- ✅ Authentication failures
- ✅ Network timeouts
- ✅ 404 Not Found
- ✅ SSH key issues
- ✅ Invalid repository URLs

## Usage Examples

### In Application Code

```go
// Download component
components, err := downloader.DownloadComponents(repoURL)
if err != nil {
    appLogger.FatalMsg(errors.NewComponentDownloadError("skill", repoURL, err))
}

// Profile not found
if !profileExists(name) {
    fmt.Println(errors.NewProfileNotFoundError(name).Format())
    os.Exit(1)
}

// Link component
if err := linker.LinkComponent(componentType, name); err != nil {
    appLogger.FatalMsg(errors.NewComponentLinkerError(componentType, target, err))
}
```

### Creating Custom Errors

```go
// Simple error
err := errors.New("Operation failed")

// Error with context
err := errors.New("Failed to process request").
    WithContext("The server returned an unexpected status code")

// Full error with suggestion and example
err := errors.New("Configuration invalid").
    WithContext("The config file is missing required fields").
    WithSuggestion("Add the missing fields to your config file").
    WithExample("agent-smith config edit").
    WithDetails(
        "Missing: api_key",
        "Missing: target_dir"
    )
```

## Implementation Status

- ✅ Error message framework (`pkg/errors/`)
- ✅ Pre-built error helpers (`pkg/errors/helpers.go`)
- ✅ File operation errors (`internal/fileutil/fileutil.go`)
- ✅ Comprehensive test coverage
- ✅ Color-coded output
- ✅ Documentation and examples

## Benefits

1. **Faster Issue Resolution**: Developers can quickly understand and fix problems
2. **Better User Experience**: Clear, actionable error messages reduce frustration
3. **Reduced Support Burden**: Self-service problem solving through helpful suggestions
4. **Consistent Error Handling**: Standardized approach across the codebase
5. **Easier Debugging**: Context and details help diagnose issues quickly

## Testing

Run all error message tests:
```bash
# Test error package
go test ./pkg/errors/... -v

# Test file utility errors
go test ./internal/fileutil/... -v -run "ErrorMessages"

# Test all packages
go test ./... -v
```

All tests pass ✅

## Conclusion

Story-012 has been successfully implemented with a comprehensive, developer-friendly error handling system. The implementation includes:

- ✅ Structured error messages with context, suggestions, and examples
- ✅ Pre-built helpers for common error scenarios
- ✅ Enhanced file operation error messages
- ✅ Comprehensive test coverage
- ✅ Color-coded output for better readability
- ✅ Clear documentation and usage examples

Developers can now quickly identify and resolve issues thanks to helpful, actionable error messages throughout the application.
