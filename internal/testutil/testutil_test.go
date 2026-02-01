package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCreateTempDir tests temporary directory creation
func TestCreateTempDir(t *testing.T) {
	dir := CreateTempDir(t, "testutil-test-*")

	// Verify directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("expected temp directory to exist: %s", dir)
	}

	// Verify it's a directory
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("failed to stat temp directory: %v", err)
	}

	if !info.IsDir() {
		t.Errorf("expected path to be a directory: %s", dir)
	}
}

// TestCreateTestFile tests single file creation
func TestCreateTestFile(t *testing.T) {
	tempDir := CreateTempDir(t, "testutil-file-*")

	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "test content"

	CreateTestFile(t, testFile, testContent)

	// Verify file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Errorf("expected test file to exist: %s", testFile)
	}

	// Verify content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("expected content %q, got %q", testContent, string(content))
	}
}

// TestCreateTestFileWithNestedDirs tests file creation with nested directories
func TestCreateTestFileWithNestedDirs(t *testing.T) {
	tempDir := CreateTempDir(t, "testutil-nested-*")

	testFile := filepath.Join(tempDir, "a", "b", "c", "test.txt")
	testContent := "nested content"

	CreateTestFile(t, testFile, testContent)

	// Verify file exists
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read nested file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("expected content %q, got %q", testContent, string(content))
	}
}

// TestCreateTestFiles tests creating multiple files
func TestCreateTestFiles(t *testing.T) {
	tempDir := CreateTempDir(t, "testutil-multi-*")

	files := map[string]string{
		"file1.txt":        "content 1",
		"file2.txt":        "content 2",
		"subdir/file3.txt": "content 3",
		"a/b/c/nested.txt": "nested",
	}

	CreateTestFiles(t, tempDir, files)

	// Verify all files were created with correct content
	for relPath, expectedContent := range files {
		fullPath := filepath.Join(tempDir, relPath)

		content, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("failed to read file %s: %v", relPath, err)
			continue
		}

		if string(content) != expectedContent {
			t.Errorf("file %s: expected content %q, got %q", relPath, expectedContent, string(content))
		}
	}
}

// TestAssertFileExists tests file existence assertion
func TestAssertFileExists(t *testing.T) {
	tempDir := CreateTempDir(t, "testutil-exists-*")

	// Create a test file
	testFile := filepath.Join(tempDir, "exists.txt")
	CreateTestFile(t, testFile, "content")

	// This should pass (no error)
	AssertFileExists(t, testFile)
}

// TestAssertFileNotExists tests file non-existence assertion
func TestAssertFileNotExists(t *testing.T) {
	tempDir := CreateTempDir(t, "testutil-notexists-*")

	nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")

	// This should pass (no error)
	AssertFileNotExists(t, nonExistentFile)
}

// TestAssertFileContent tests file content assertion
func TestAssertFileContent(t *testing.T) {
	tempDir := CreateTempDir(t, "testutil-content-*")

	testFile := filepath.Join(tempDir, "test.txt")
	expectedContent := "expected content"
	CreateTestFile(t, testFile, expectedContent)

	// This should pass (no error)
	AssertFileContent(t, testFile, expectedContent)
}

// TestAssertDirectoryExists tests directory existence assertion
func TestAssertDirectoryExists(t *testing.T) {
	tempDir := CreateTempDir(t, "testutil-dir-*")

	// tempDir itself should exist
	AssertDirectoryExists(t, tempDir)

	// Create a subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	AssertDirectoryExists(t, subDir)
}

// TestAssertError tests error assertion
func TestAssertError(t *testing.T) {
	// Test with actual error - should pass
	AssertError(t, os.ErrNotExist)
}

// TestAssertNoError tests no-error assertion
func TestAssertNoError(t *testing.T) {
	// Test with nil error - should pass
	AssertNoError(t, nil)
}

// TestAssertEqual tests equality assertion
func TestAssertEqual(t *testing.T) {
	// Test with equal values - should pass
	AssertEqual(t, 42, 42)
	AssertEqual(t, "hello", "hello")
	AssertEqual(t, true, true)
}

// TestAssertNotEqual tests inequality assertion
func TestAssertNotEqual(t *testing.T) {
	// Test with unequal values - should pass
	AssertNotEqual(t, 42, 43)
	AssertNotEqual(t, "hello", "world")
}

// TestAssertTrue tests true assertion
func TestAssertTrue(t *testing.T) {
	// Test with true condition - should pass
	AssertTrue(t, true)
	AssertTrue(t, 1 == 1)
}

// TestAssertFalse tests false assertion
func TestAssertFalse(t *testing.T) {
	// Test with false condition - should pass
	AssertFalse(t, false)
	AssertFalse(t, 1 == 2)
}

// TestAssertContains tests contains assertion
func TestAssertContains(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	// Test with contained element - should pass
	AssertContains(t, slice, "banana")
}

// TestCreateGitRepo tests git repository creation
func TestCreateGitRepo(t *testing.T) {
	tempDir := CreateTempDir(t, "testutil-git-*")

	gitPath := CreateGitRepo(t, tempDir)

	// Verify .git directory exists
	gitDir := filepath.Join(gitPath, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		t.Fatalf("expected .git directory to exist: %v", err)
	}

	if !info.IsDir() {
		t.Error("expected .git to be a directory")
	}
}

// TestIntegrationCreateAndAssert tests a full workflow using multiple utilities
func TestIntegrationCreateAndAssert(t *testing.T) {
	// Create a test environment
	tempDir := CreateTempDir(t, "testutil-integration-*")

	// Create a directory structure
	files := map[string]string{
		"README.md":            "# Project",
		"src/main.go":          "package main",
		"src/util/helper.go":   "package util",
		"test/data/sample.txt": "test data",
	}

	CreateTestFiles(t, tempDir, files)

	// Assert all files exist
	for relPath, expectedContent := range files {
		fullPath := filepath.Join(tempDir, relPath)
		AssertFileExists(t, fullPath)
		AssertFileContent(t, fullPath, expectedContent)
	}

	// Assert directories exist
	AssertDirectoryExists(t, filepath.Join(tempDir, "src"))
	AssertDirectoryExists(t, filepath.Join(tempDir, "src", "util"))
	AssertDirectoryExists(t, filepath.Join(tempDir, "test", "data"))

	// Assert non-existent file doesn't exist
	AssertFileNotExists(t, filepath.Join(tempDir, "nonexistent.txt"))

	// Test other assertions
	AssertEqual(t, len(files), 4)
	AssertTrue(t, len(files) > 0)
	AssertFalse(t, len(files) == 0)
}

// TestHelperMessageFormatting tests that optional messages work correctly
func TestHelperMessageFormatting(t *testing.T) {
	// Test AssertEqual with custom message
	AssertEqual(t, 1, 1, "numbers should be equal")

	// Test AssertNoError with custom message
	AssertNoError(t, nil, "operation should succeed")

	// Test AssertTrue with custom message
	AssertTrue(t, true, "condition should be true")
}
