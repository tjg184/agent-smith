package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tjg184/agent-smith/internal/models"
)

// TestCopyFileErrorMessages tests that CopyFile provides clear, actionable error messages
func TestCopyFileErrorMessages(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("source file does not exist", func(t *testing.T) {
		srcFile := filepath.Join(tempDir, "nonexistent.txt")
		dstFile := filepath.Join(tempDir, "destination.txt")

		err := CopyFile(srcFile, dstFile)
		if err == nil {
			t.Fatal("expected error when copying non-existent file")
		}

		errMsg := err.Error()
		if !strings.Contains(errMsg, "source file does not exist") {
			t.Errorf("error message should mention source file does not exist, got: %s", errMsg)
		}
		if !strings.Contains(errMsg, srcFile) {
			t.Errorf("error message should include source file path, got: %s", errMsg)
		}
	})

	t.Run("destination directory does not exist", func(t *testing.T) {
		// Create a source file
		srcFile := filepath.Join(tempDir, "source.txt")
		if err := os.WriteFile(srcFile, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		// Try to copy to a non-existent directory
		nonExistentDir := filepath.Join(tempDir, "nonexistent", "subdir")
		dstFile := filepath.Join(nonExistentDir, "destination.txt")

		err := CopyFile(srcFile, dstFile)
		if err == nil {
			t.Fatal("expected error when destination directory does not exist")
		}

		errMsg := err.Error()
		if !strings.Contains(errMsg, "destination directory does not exist") {
			t.Errorf("error message should mention destination directory does not exist, got: %s", errMsg)
		}
	})
}

// TestCopyDirectoryContentsErrorMessages tests that CopyDirectoryContents provides clear error messages
func TestCopyDirectoryContentsErrorMessages(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("source directory does not exist", func(t *testing.T) {
		srcDir := filepath.Join(tempDir, "nonexistent")
		dstDir := filepath.Join(tempDir, "destination")

		err := CopyDirectoryContents(srcDir, dstDir)
		if err == nil {
			t.Fatal("expected error when copying non-existent directory")
		}

		errMsg := err.Error()
		if !strings.Contains(errMsg, "source directory does not exist") {
			t.Errorf("error message should mention source directory does not exist, got: %s", errMsg)
		}
		if !strings.Contains(errMsg, srcDir) {
			t.Errorf("error message should include source directory path, got: %s", errMsg)
		}
	})

	t.Run("source is not a directory", func(t *testing.T) {
		// Create a file (not a directory)
		srcFile := filepath.Join(tempDir, "file.txt")
		if err := os.WriteFile(srcFile, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		dstDir := filepath.Join(tempDir, "destination")

		err := CopyDirectoryContents(srcFile, dstDir)
		if err == nil {
			t.Fatal("expected error when source is not a directory")
		}

		errMsg := err.Error()
		if !strings.Contains(errMsg, "source is not a directory") {
			t.Errorf("error message should mention source is not a directory, got: %s", errMsg)
		}
	})
}

// TestCopyComponentFilesErrorMessages tests that CopyComponentFiles provides clear error messages
func TestCopyComponentFilesErrorMessages(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("component file does not exist", func(t *testing.T) {
		repoPath := filepath.Join(tempDir, "repo")
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			t.Fatalf("failed to create repo directory: %v", err)
		}

		component := models.DetectedComponent{
			Name:     "test-component",
			FilePath: "components/test.md",
			Type:     "skill",
		}

		dstDir := filepath.Join(tempDir, "destination")

		err := CopyComponentFiles(repoPath, component, dstDir)
		if err == nil {
			t.Fatal("expected error when component file does not exist")
		}

		errMsg := err.Error()
		if !strings.Contains(errMsg, "component file does not exist") {
			t.Errorf("error message should mention component file does not exist, got: %s", errMsg)
		}
		if !strings.Contains(errMsg, component.Name) {
			t.Errorf("error message should include component name, got: %s", errMsg)
		}
	})

	t.Run("component name included in error for missing component", func(t *testing.T) {
		repoPath := filepath.Join(tempDir, "repo2")
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			t.Fatalf("failed to create repo directory: %v", err)
		}

		// Create a component descriptor but don't create the actual file
		component := models.DetectedComponent{
			Name:     "my-custom-skill-name",
			FilePath: "my-skill/SKILL.md",
			Type:     "skill",
		}

		dstDir := filepath.Join(tempDir, "destination2")

		err := CopyComponentFiles(repoPath, component, dstDir)
		if err == nil {
			t.Fatal("expected error when component file does not exist")
		}

		errMsg := err.Error()
		// Should contain the component name from the DetectedComponent
		if !strings.Contains(errMsg, "my-custom-skill-name") {
			t.Errorf("error message should include component name 'my-custom-skill-name', got: %s", errMsg)
		}
		if !strings.Contains(errMsg, "component file does not exist") {
			t.Errorf("error message should mention component file does not exist, got: %s", errMsg)
		}
	})
}

// TestErrorMessagesProvideContext verifies that error messages help developers understand what went wrong
func TestErrorMessagesProvideContext(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name            string
		setupFunc       func() error
		expectedPhrases []string
		description     string
	}{
		{
			name: "CopyFile with missing source provides file path",
			setupFunc: func() error {
				return CopyFile(
					filepath.Join(tempDir, "missing.txt"),
					filepath.Join(tempDir, "dest.txt"),
				)
			},
			expectedPhrases: []string{"source file does not exist", "missing.txt"},
			description:     "Should clearly state the source file is missing and show the path",
		},
		{
			name: "CopyDirectoryContents with file instead of directory",
			setupFunc: func() error {
				srcFile := filepath.Join(tempDir, "notadir.txt")
				os.WriteFile(srcFile, []byte("test"), 0644)
				return CopyDirectoryContents(srcFile, filepath.Join(tempDir, "dest"))
			},
			expectedPhrases: []string{"source is not a directory", "notadir.txt"},
			description:     "Should clearly state the source is not a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setupFunc()
			if err == nil {
				t.Fatal("expected an error but got nil")
			}

			errMsg := err.Error()
			for _, phrase := range tt.expectedPhrases {
				if !strings.Contains(errMsg, phrase) {
					t.Errorf("%s\nExpected error to contain %q, but got:\n%s",
						tt.description, phrase, errMsg)
				}
			}
		})
	}
}
