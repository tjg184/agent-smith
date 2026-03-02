// Package testutil provides shared test utilities for consistent testing across packages
package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// CreateTempDir creates a temporary directory for testing and returns its path.
// The directory is automatically cleaned up when the test completes.
func CreateTempDir(t testing.TB, pattern string) string {
	t.Helper()

	tempDir, err := os.MkdirTemp("", pattern)
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	return tempDir
}

// CreateTestFile creates a test file with the given content at the specified path.
// It creates parent directories as needed.
func CreateTestFile(t testing.TB, path string, content string) {
	t.Helper()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create file %s: %v", path, err)
	}
}

// CreateTestFiles creates multiple test files from a map of relative paths to content.
// The baseDir is the root directory where files will be created.
func CreateTestFiles(t testing.TB, baseDir string, files map[string]string) {
	t.Helper()

	for relPath, content := range files {
		fullPath := filepath.Join(baseDir, relPath)
		CreateTestFile(t, fullPath, content)
	}
}

// AssertFileExists checks that a file exists at the given path.
func AssertFileExists(t testing.TB, path string) {
	t.Helper()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file to exist: %s", path)
	}
}

// AssertFileNotExists checks that a file does not exist at the given path.
func AssertFileNotExists(t testing.TB, path string) {
	t.Helper()

	if _, err := os.Stat(path); err == nil {
		t.Errorf("expected file to not exist: %s", path)
	}
}

// AssertFileContent checks that a file contains the expected content.
func AssertFileContent(t testing.TB, path string, expectedContent string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", path, err)
	}

	if string(content) != expectedContent {
		t.Errorf("file content mismatch for %s:\nExpected: %s\nActual: %s",
			path, expectedContent, string(content))
	}
}

// AssertDirectoryExists checks that a directory exists at the given path.
func AssertDirectoryExists(t testing.TB, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Errorf("expected directory to exist: %s (error: %v)", path, err)
		return
	}

	if !info.IsDir() {
		t.Errorf("path exists but is not a directory: %s", path)
	}
}

// AssertError checks that an error occurred.
func AssertError(t testing.TB, err error, msgAndArgs ...interface{}) {
	t.Helper()

	if err == nil {
		if len(msgAndArgs) > 0 {
			t.Errorf("expected error: %v", msgAndArgs[0])
		} else {
			t.Error("expected error, got nil")
		}
	}
}

// AssertNoError checks that no error occurred.
func AssertNoError(t testing.TB, err error, msgAndArgs ...interface{}) {
	t.Helper()

	if err != nil {
		if len(msgAndArgs) > 0 {
			t.Errorf("unexpected error: %v - %v", msgAndArgs[0], err)
		} else {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

// AssertEqual checks that two values are equal.
func AssertEqual(t testing.TB, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	if expected != actual {
		if len(msgAndArgs) > 0 {
			t.Errorf("%v: expected %v, got %v", msgAndArgs[0], expected, actual)
		} else {
			t.Errorf("expected %v, got %v", expected, actual)
		}
	}
}

// AssertNotEqual checks that two values are not equal.
func AssertNotEqual(t testing.TB, notExpected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	if notExpected == actual {
		if len(msgAndArgs) > 0 {
			t.Errorf("%v: expected values to be different, but both are %v", msgAndArgs[0], actual)
		} else {
			t.Errorf("expected values to be different, but both are %v", actual)
		}
	}
}

// AssertTrue checks that a condition is true.
func AssertTrue(t testing.TB, condition bool, msgAndArgs ...interface{}) {
	t.Helper()

	if !condition {
		if len(msgAndArgs) > 0 {
			t.Errorf("expected condition to be true: %v", msgAndArgs[0])
		} else {
			t.Error("expected condition to be true")
		}
	}
}

// AssertFalse checks that a condition is false.
func AssertFalse(t testing.TB, condition bool, msgAndArgs ...interface{}) {
	t.Helper()

	if condition {
		if len(msgAndArgs) > 0 {
			t.Errorf("expected condition to be false: %v", msgAndArgs[0])
		} else {
			t.Error("expected condition to be false")
		}
	}
}

// AssertContains checks that a slice contains an element.
func AssertContains(t testing.TB, slice []string, element string, msgAndArgs ...interface{}) {
	t.Helper()

	for _, item := range slice {
		if item == element {
			return
		}
	}

	if len(msgAndArgs) > 0 {
		t.Errorf("%v: slice does not contain element %q: %v", msgAndArgs[0], element, slice)
	} else {
		t.Errorf("slice does not contain element %q: %v", element, slice)
	}
}

// CreateGitRepo creates a minimal git repository structure for testing.
// It creates a .git directory to make the path look like a git repository.
func CreateGitRepo(t testing.TB, path string) string {
	t.Helper()

	gitDir := filepath.Join(path, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	return path
}

// CreateComponentLockFile creates a component lock file with the correct nested structure (version 5).
// This helper ensures tests use the correct lock file format that matches the code's expectations:
//
//	{
//	  "version": 5,
//	  "skills": {
//	    "https://github.com/...": {
//	      "component-name": { ... entry ... }
//	    }
//	  }
//	}
func CreateComponentLockFile(t testing.TB, lockFilePath, componentType, componentName, sourceUrl string, entry map[string]interface{}) {
	t.Helper()

	// Ensure the entry has the minimum required fields
	if entry["sourceUrl"] == nil {
		entry["sourceUrl"] = sourceUrl
	}
	if entry["version"] == nil {
		entry["version"] = 5
	}

	// Build the nested structure: map[componentType]map[sourceUrl]map[componentName]entry
	lockData := map[string]interface{}{
		"version": 5,
		componentType: map[string]interface{}{
			sourceUrl: map[string]interface{}{
				componentName: entry,
			},
		},
	}

	lockJSON, err := json.MarshalIndent(lockData, "", "  ")
	AssertNoError(t, err, "Failed to marshal lock data")

	// Create parent directory if it doesn't exist
	dir := filepath.Dir(lockFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}

	err = os.WriteFile(lockFilePath, lockJSON, 0644)
	AssertNoError(t, err, "Failed to write lock file")
}

// AddComponentToLockFile adds a component to an existing lock file, or creates it if it doesn't exist.
// This maintains the correct nested structure for version 5 lock files.
func AddComponentToLockFile(t testing.TB, lockFilePath, componentType, componentName, sourceUrl string, entry map[string]interface{}) {
	t.Helper()

	// Ensure the entry has the minimum required fields
	if entry["sourceUrl"] == nil {
		entry["sourceUrl"] = sourceUrl
	}
	if entry["version"] == nil {
		entry["version"] = 5
	}

	// Read existing lock file or create new one
	var lockData map[string]interface{}
	lockBytes, err := os.ReadFile(lockFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create new lock file structure
			lockData = map[string]interface{}{
				"version": 5,
			}
		} else {
			t.Fatalf("failed to read lock file: %v", err)
		}
	} else {
		if err := json.Unmarshal(lockBytes, &lockData); err != nil {
			t.Fatalf("failed to unmarshal lock file: %v", err)
		}
	}

	// Ensure component type map exists
	if lockData[componentType] == nil {
		lockData[componentType] = make(map[string]interface{})
	}

	componentTypeMap := lockData[componentType].(map[string]interface{})

	// Ensure source URL map exists
	if componentTypeMap[sourceUrl] == nil {
		componentTypeMap[sourceUrl] = make(map[string]interface{})
	}

	sourceUrlMap := componentTypeMap[sourceUrl].(map[string]interface{})

	// Add the component entry
	sourceUrlMap[componentName] = entry

	// Write back
	lockJSON, err := json.MarshalIndent(lockData, "", "  ")
	AssertNoError(t, err, "Failed to marshal lock data")

	// Create parent directory if it doesn't exist
	dir := filepath.Dir(lockFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}

	err = os.WriteFile(lockFilePath, lockJSON, 0644)
	AssertNoError(t, err, "Failed to write lock file")
}
