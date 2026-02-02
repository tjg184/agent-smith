package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureTargetStructure_CreatesNewStructure(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "agent-smith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	targetDir := filepath.Join(tempDir, ".opencode")

	// Call EnsureTargetStructure - should create everything
	created, err := EnsureTargetStructure(targetDir)
	if err != nil {
		t.Fatalf("EnsureTargetStructure failed: %v", err)
	}

	// Should return true since directories were created
	if !created {
		t.Errorf("Expected created=true for new structure, got false")
	}

	// Verify target directory exists
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		t.Errorf("Target directory was not created")
	}

	// Verify subdirectories exist
	subdirs := []string{"skills", "agents", "commands"}
	for _, subdir := range subdirs {
		subdirPath := filepath.Join(targetDir, subdir)
		if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
			t.Errorf("Subdirectory %s was not created", subdir)
		}
	}
}

func TestEnsureTargetStructure_ExistingStructure(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "agent-smith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	targetDir := filepath.Join(tempDir, ".claude")

	// Pre-create the structure
	subdirs := []string{"skills", "agents", "commands"}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("Failed to pre-create target directory: %v", err)
	}
	for _, subdir := range subdirs {
		subdirPath := filepath.Join(targetDir, subdir)
		if err := os.MkdirAll(subdirPath, 0755); err != nil {
			t.Fatalf("Failed to pre-create subdirectory %s: %v", subdir, err)
		}
	}

	// Call EnsureTargetStructure - structure already exists
	created, err := EnsureTargetStructure(targetDir)
	if err != nil {
		t.Fatalf("EnsureTargetStructure failed: %v", err)
	}

	// Should return false since directories already existed
	if created {
		t.Errorf("Expected created=false for existing structure, got true")
	}
}

func TestEnsureTargetStructure_PartialStructure(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "agent-smith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	targetDir := filepath.Join(tempDir, ".opencode")

	// Pre-create only the target directory, but not subdirectories
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("Failed to pre-create target directory: %v", err)
	}

	// Pre-create only some subdirectories
	if err := os.MkdirAll(filepath.Join(targetDir, "skills"), 0755); err != nil {
		t.Fatalf("Failed to pre-create skills directory: %v", err)
	}

	// Call EnsureTargetStructure - missing subdirectories should be created
	created, err := EnsureTargetStructure(targetDir)
	if err != nil {
		t.Fatalf("EnsureTargetStructure failed: %v", err)
	}

	// Should return true since some directories needed to be created
	if !created {
		t.Errorf("Expected created=true for partial structure, got false")
	}

	// Verify all subdirectories now exist
	subdirs := []string{"skills", "agents", "commands"}
	for _, subdir := range subdirs {
		subdirPath := filepath.Join(targetDir, subdir)
		if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
			t.Errorf("Subdirectory %s was not created", subdir)
		}
	}
}

func TestFindProjectRootFromDir_StopsAtGitDirectory(t *testing.T) {
	// Create temporary directory structure:
	// tempDir/
	//   .git/
	//   src/
	//     components/
	tempDir, err := os.MkdirTemp("", "agent-smith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .git directory to mark project root
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Create nested directory
	nestedDir := filepath.Join(tempDir, "src", "components")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	// Try to find project root from nested directory
	// Should find the .git directory and return it as project root
	projectRoot, err := FindProjectRootFromDir(nestedDir)
	if err != nil {
		t.Fatalf("Expected to find project root at .git/, got error: %v", err)
	}

	// Verify project root is the temp directory (where .git is)
	if projectRoot != tempDir {
		t.Errorf("Expected project root to be %q, got %q", tempDir, projectRoot)
	}
}

func TestFindProjectRootFromDir_FindsProjectWithinGitRepo(t *testing.T) {
	// Create temporary directory structure:
	// tempDir/
	//   .git/
	//   .opencode/
	//   src/
	//     components/
	tempDir, err := os.MkdirTemp("", "agent-smith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .git directory to mark project root
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Create .opencode directory at project root
	opencodeDir := filepath.Join(tempDir, ".opencode")
	if err := os.MkdirAll(opencodeDir, 0755); err != nil {
		t.Fatalf("Failed to create .opencode directory: %v", err)
	}

	// Create nested directory
	nestedDir := filepath.Join(tempDir, "src", "components")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	// Try to find project root from nested directory
	projectRoot, err := FindProjectRootFromDir(nestedDir)
	if err != nil {
		t.Fatalf("Expected to find project root, got error: %v", err)
	}

	// Verify project root is the temp directory (where .opencode and .git are)
	if projectRoot != tempDir {
		t.Errorf("Expected project root to be %q, got %q", tempDir, projectRoot)
	}
}

func TestFindProjectRootFromDir_DoesNotEscapeGitRepo(t *testing.T) {
	// Create temporary directory structure:
	// tempDir/
	//   .opencode/  (this should NOT be found)
	//   my-project/
	//     .git/
	//     src/
	tempDir, err := os.MkdirTemp("", "agent-smith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .opencode in parent directory
	parentOpencode := filepath.Join(tempDir, ".opencode")
	if err := os.MkdirAll(parentOpencode, 0755); err != nil {
		t.Fatalf("Failed to create parent .opencode directory: %v", err)
	}

	// Create project directory with .git
	projectDir := filepath.Join(tempDir, "my-project")
	gitDir := filepath.Join(projectDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Create nested directory in project
	srcDir := filepath.Join(projectDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create src directory: %v", err)
	}

	// Try to find project root from src directory
	// Should find the .git directory in my-project, NOT the parent .opencode
	projectRoot, err := FindProjectRootFromDir(srcDir)
	if err != nil {
		t.Fatalf("Expected to find project root at .git/, got error: %v", err)
	}

	// Verify project root is the my-project directory (where .git is)
	if projectRoot != projectDir {
		t.Errorf("Expected project root to be %q (my-project/.git), got %q", projectDir, projectRoot)
	}
}

func TestFindProjectRootFromDir_ErrorWhenNoProjectFound(t *testing.T) {
	// Create temporary directory with no project markers
	tempDir, err := os.MkdirTemp("", "agent-smith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a nested directory without any project markers
	testDir := filepath.Join(tempDir, "no-project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Try to find project root - should return error
	_, err = FindProjectRootFromDir(testDir)
	if err == nil {
		t.Fatal("Expected error when no project markers found, got nil")
	}

	// Verify error message contains required information
	errMsg := err.Error()

	// Check for key phrases
	requiredPhrases := []string{
		"no project boundary detected",
		"Supported project markers:",
		".opencode/",
		".claude/",
		".git/",
		"go.mod",
		"package.json",
		"pyproject.toml",
		"Cargo.toml",
		"composer.json",
		"pom.xml",
		"build.gradle",
		"Gemfile",
		"mix.exs",
		"To fix this:",
		"mkdir -p .opencode/",
		"--project-dir",
		"git init",
	}

	for _, phrase := range requiredPhrases {
		if !contains(errMsg, phrase) {
			t.Errorf("Error message missing required phrase: %q\nFull error: %s", phrase, errMsg)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAtIndex(s, substr))
}

func containsAtIndex(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
