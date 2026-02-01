// Package testutil provides shared test utilities for consistent testing across packages
package testutil

import (
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
