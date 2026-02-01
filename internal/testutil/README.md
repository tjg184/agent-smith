# testutil

The `testutil` package provides shared test utilities for consistent testing across packages in agent-smith.

## Overview

This package eliminates code duplication in tests by providing common utilities for:
- Temporary directory management
- Test file creation
- Assertion helpers
- Git repository mocking

## Usage

### Temporary Directories

```go
func TestMyFeature(t *testing.T) {
    // Create a temporary directory that's automatically cleaned up
    tempDir := testutil.CreateTempDir(t, "mytest-*")
    
    // Use the directory in your test...
}
```

### Creating Test Files

```go
func TestFileOperations(t *testing.T) {
    tempDir := testutil.CreateTempDir(t, "test-*")
    
    // Create a single file
    testutil.CreateTestFile(t, filepath.Join(tempDir, "file.txt"), "content")
    
    // Create multiple files at once
    files := map[string]string{
        "file1.txt":        "content 1",
        "dir/file2.txt":    "content 2",
        "a/b/c/nested.txt": "nested content",
    }
    testutil.CreateTestFiles(t, tempDir, files)
}
```

### Assertion Helpers

All assertion helpers call `t.Helper()` to ensure correct line numbers in test failures.

#### File Assertions

```go
// Assert file exists
testutil.AssertFileExists(t, "/path/to/file")

// Assert file does not exist
testutil.AssertFileNotExists(t, "/path/to/file")

// Assert file content matches
testutil.AssertFileContent(t, "/path/to/file", "expected content")

// Assert directory exists
testutil.AssertDirectoryExists(t, "/path/to/dir")
```

#### Error Assertions

```go
// Assert an error occurred
testutil.AssertError(t, err, "optional message")

// Assert no error occurred
testutil.AssertNoError(t, err, "optional message")
```

#### Value Assertions

```go
// Assert values are equal
testutil.AssertEqual(t, expected, actual, "optional message")

// Assert values are not equal
testutil.AssertNotEqual(t, notExpected, actual, "optional message")

// Assert condition is true
testutil.AssertTrue(t, condition, "optional message")

// Assert condition is false
testutil.AssertFalse(t, condition, "optional message")

// Assert slice contains element
testutil.AssertContains(t, []string{"a", "b", "c"}, "b", "optional message")
```

### Git Repository Mocking

```go
func TestGitOperations(t *testing.T) {
    tempDir := testutil.CreateTempDir(t, "git-test-*")
    
    // Create a minimal git repository structure (creates .git directory)
    repoPath := testutil.CreateGitRepo(t, tempDir)
    
    // Now repoPath looks like a git repository
}
```

## Example Test

Here's a complete example showing how to use testutil:

```go
package mypackage

import (
    "path/filepath"
    "testing"
    
    "github.com/tgaines/agent-smith/internal/testutil"
)

func TestComponentCopy(t *testing.T) {
    // Setup: Create test environment
    srcDir := testutil.CreateTempDir(t, "src-*")
    dstDir := testutil.CreateTempDir(t, "dst-*")
    
    // Setup: Create test files
    files := map[string]string{
        "component.md":      "# Component",
        "resources/data.txt": "test data",
    }
    testutil.CreateTestFiles(t, srcDir, files)
    
    // Execute: Run the code being tested
    err := CopyDirectory(srcDir, dstDir)
    
    // Verify: Check results
    testutil.AssertNoError(t, err, "copy should succeed")
    
    for relPath, expectedContent := range files {
        dstPath := filepath.Join(dstDir, relPath)
        testutil.AssertFileExists(t, dstPath)
        testutil.AssertFileContent(t, dstPath, expectedContent)
    }
}
```

## Design Principles

1. **Automatic Cleanup**: All temporary resources are automatically cleaned up using `t.Cleanup()`
2. **Proper Error Reporting**: All helpers use `t.Helper()` to report errors at the correct line
3. **Consistent Interface**: All functions follow Go testing conventions with `t testing.TB` as the first parameter
4. **Fail Fast**: Helpers use `t.Fatalf()` for setup errors and `t.Errorf()` for assertion failures
5. **Composable**: Utilities can be combined to create complex test scenarios

## Benefits

Using testutil provides several benefits:

1. **Reduced Boilerplate**: Eliminates repetitive test setup code
2. **Consistency**: All tests use the same utilities and patterns
3. **Maintainability**: Changes to test utilities only need to be made in one place
4. **Readability**: Tests are more concise and focus on what's being tested
5. **Safety**: Automatic cleanup prevents test pollution

## Conventions

- All functions that fail the test use `t testing.TB` as the first parameter
- All assertion functions call `t.Helper()` for proper error reporting
- All temp directories are automatically cleaned up
- Optional messages can be provided to most assertions via variadic parameters
