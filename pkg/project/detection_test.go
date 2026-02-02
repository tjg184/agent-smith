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

// TestFindProjectRootFromDir_GoProject tests detection of Go projects
func TestFindProjectRootFromDir_GoProject(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agent-smith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create go.mod file to mark Go project
	goModPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test\n"), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create nested directory
	nestedDir := filepath.Join(tempDir, "internal", "pkg")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	// Should find project root from nested directory
	projectRoot, err := FindProjectRootFromDir(nestedDir)
	if err != nil {
		t.Fatalf("Expected to find project root, got error: %v", err)
	}

	if projectRoot != tempDir {
		t.Errorf("Expected project root to be %q, got %q", tempDir, projectRoot)
	}
}

// TestFindProjectRootFromDir_NodeProject tests detection of Node.js projects
func TestFindProjectRootFromDir_NodeProject(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agent-smith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create package.json file to mark Node.js project
	packageJsonPath := filepath.Join(tempDir, "package.json")
	if err := os.WriteFile(packageJsonPath, []byte("{\"name\": \"test\"}\n"), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Create nested directory
	nestedDir := filepath.Join(tempDir, "src", "components")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	// Should find project root from nested directory
	projectRoot, err := FindProjectRootFromDir(nestedDir)
	if err != nil {
		t.Fatalf("Expected to find project root, got error: %v", err)
	}

	if projectRoot != tempDir {
		t.Errorf("Expected project root to be %q, got %q", tempDir, projectRoot)
	}
}

// TestFindProjectRootFromDir_PythonProject tests detection of Python projects
func TestFindProjectRootFromDir_PythonProject(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agent-smith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create pyproject.toml file to mark Python project
	pyprojectPath := filepath.Join(tempDir, "pyproject.toml")
	if err := os.WriteFile(pyprojectPath, []byte("[tool.poetry]\nname = \"test\"\n"), 0644); err != nil {
		t.Fatalf("Failed to create pyproject.toml: %v", err)
	}

	// Create nested directory
	nestedDir := filepath.Join(tempDir, "src", "app")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	// Should find project root from nested directory
	projectRoot, err := FindProjectRootFromDir(nestedDir)
	if err != nil {
		t.Fatalf("Expected to find project root, got error: %v", err)
	}

	if projectRoot != tempDir {
		t.Errorf("Expected project root to be %q, got %q", tempDir, projectRoot)
	}
}

// TestFindProjectRootFromDir_RustProject tests detection of Rust projects
func TestFindProjectRootFromDir_RustProject(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agent-smith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create Cargo.toml file to mark Rust project
	cargoPath := filepath.Join(tempDir, "Cargo.toml")
	if err := os.WriteFile(cargoPath, []byte("[package]\nname = \"test\"\n"), 0644); err != nil {
		t.Fatalf("Failed to create Cargo.toml: %v", err)
	}

	// Create nested directory
	nestedDir := filepath.Join(tempDir, "src", "lib")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	// Should find project root from nested directory
	projectRoot, err := FindProjectRootFromDir(nestedDir)
	if err != nil {
		t.Fatalf("Expected to find project root, got error: %v", err)
	}

	if projectRoot != tempDir {
		t.Errorf("Expected project root to be %q, got %q", tempDir, projectRoot)
	}
}

// TestFindProjectRootFromDir_PHPProject tests detection of PHP projects
func TestFindProjectRootFromDir_PHPProject(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agent-smith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create composer.json file to mark PHP project
	composerPath := filepath.Join(tempDir, "composer.json")
	if err := os.WriteFile(composerPath, []byte("{\"name\": \"test/test\"}\n"), 0644); err != nil {
		t.Fatalf("Failed to create composer.json: %v", err)
	}

	// Create nested directory
	nestedDir := filepath.Join(tempDir, "src", "Controller")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	// Should find project root from nested directory
	projectRoot, err := FindProjectRootFromDir(nestedDir)
	if err != nil {
		t.Fatalf("Expected to find project root, got error: %v", err)
	}

	if projectRoot != tempDir {
		t.Errorf("Expected project root to be %q, got %q", tempDir, projectRoot)
	}
}

// TestFindProjectRootFromDir_JavaMavenProject tests detection of Java Maven projects
func TestFindProjectRootFromDir_JavaMavenProject(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agent-smith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create pom.xml file to mark Java Maven project
	pomPath := filepath.Join(tempDir, "pom.xml")
	if err := os.WriteFile(pomPath, []byte("<project></project>\n"), 0644); err != nil {
		t.Fatalf("Failed to create pom.xml: %v", err)
	}

	// Create nested directory
	nestedDir := filepath.Join(tempDir, "src", "main", "java")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	// Should find project root from nested directory
	projectRoot, err := FindProjectRootFromDir(nestedDir)
	if err != nil {
		t.Fatalf("Expected to find project root, got error: %v", err)
	}

	if projectRoot != tempDir {
		t.Errorf("Expected project root to be %q, got %q", tempDir, projectRoot)
	}
}

// TestFindProjectRootFromDir_JavaGradleProject tests detection of Java Gradle projects
func TestFindProjectRootFromDir_JavaGradleProject(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agent-smith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create build.gradle file to mark Java Gradle project
	gradlePath := filepath.Join(tempDir, "build.gradle")
	if err := os.WriteFile(gradlePath, []byte("plugins {}\n"), 0644); err != nil {
		t.Fatalf("Failed to create build.gradle: %v", err)
	}

	// Create nested directory
	nestedDir := filepath.Join(tempDir, "src", "main", "kotlin")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	// Should find project root from nested directory
	projectRoot, err := FindProjectRootFromDir(nestedDir)
	if err != nil {
		t.Fatalf("Expected to find project root, got error: %v", err)
	}

	if projectRoot != tempDir {
		t.Errorf("Expected project root to be %q, got %q", tempDir, projectRoot)
	}
}

// TestFindProjectRootFromDir_RubyProject tests detection of Ruby projects
func TestFindProjectRootFromDir_RubyProject(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agent-smith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create Gemfile to mark Ruby project
	gemfilePath := filepath.Join(tempDir, "Gemfile")
	if err := os.WriteFile(gemfilePath, []byte("source 'https://rubygems.org'\n"), 0644); err != nil {
		t.Fatalf("Failed to create Gemfile: %v", err)
	}

	// Create nested directory
	nestedDir := filepath.Join(tempDir, "lib", "app")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	// Should find project root from nested directory
	projectRoot, err := FindProjectRootFromDir(nestedDir)
	if err != nil {
		t.Fatalf("Expected to find project root, got error: %v", err)
	}

	if projectRoot != tempDir {
		t.Errorf("Expected project root to be %q, got %q", tempDir, projectRoot)
	}
}

// TestFindProjectRootFromDir_ElixirProject tests detection of Elixir projects
func TestFindProjectRootFromDir_ElixirProject(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agent-smith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create mix.exs file to mark Elixir project
	mixPath := filepath.Join(tempDir, "mix.exs")
	if err := os.WriteFile(mixPath, []byte("defmodule Test.MixProject do\nend\n"), 0644); err != nil {
		t.Fatalf("Failed to create mix.exs: %v", err)
	}

	// Create nested directory
	nestedDir := filepath.Join(tempDir, "lib", "test")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	// Should find project root from nested directory
	projectRoot, err := FindProjectRootFromDir(nestedDir)
	if err != nil {
		t.Fatalf("Expected to find project root, got error: %v", err)
	}

	if projectRoot != tempDir {
		t.Errorf("Expected project root to be %q, got %q", tempDir, projectRoot)
	}
}

// TestFindProjectRootFromDir_MultipleMarkersPreferClosest tests that when multiple markers exist,
// the closest one to the starting directory is used
func TestFindProjectRootFromDir_MultipleMarkersPreferClosest(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agent-smith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create outer project with go.mod
	outerGoMod := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(outerGoMod, []byte("module outer\n"), 0644); err != nil {
		t.Fatalf("Failed to create outer go.mod: %v", err)
	}

	// Create nested project directory with package.json
	innerDir := filepath.Join(tempDir, "frontend")
	if err := os.MkdirAll(innerDir, 0755); err != nil {
		t.Fatalf("Failed to create inner directory: %v", err)
	}

	innerPackageJson := filepath.Join(innerDir, "package.json")
	if err := os.WriteFile(innerPackageJson, []byte("{\"name\": \"inner\"}\n"), 0644); err != nil {
		t.Fatalf("Failed to create inner package.json: %v", err)
	}

	// Create deeply nested directory in inner project
	deepDir := filepath.Join(innerDir, "src", "components")
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		t.Fatalf("Failed to create deep directory: %v", err)
	}

	// Should find the inner project root (frontend with package.json), not outer (with go.mod)
	projectRoot, err := FindProjectRootFromDir(deepDir)
	if err != nil {
		t.Fatalf("Expected to find project root, got error: %v", err)
	}

	if projectRoot != innerDir {
		t.Errorf("Expected project root to be %q (inner), got %q", innerDir, projectRoot)
	}
}

// TestFindProjectRootFromDir_PreferProjectMarkersOverBoundaries tests that
// .opencode and .claude directories are preferred over project boundary markers
func TestFindProjectRootFromDir_PreferProjectMarkersOverBoundaries(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agent-smith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .git directory (boundary marker)
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Create go.mod (boundary marker)
	goModPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test\n"), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create .opencode directory (preferred marker) in subdirectory
	subDir := filepath.Join(tempDir, "subproject")
	opencodeDir := filepath.Join(subDir, ".opencode")
	if err := os.MkdirAll(opencodeDir, 0755); err != nil {
		t.Fatalf("Failed to create .opencode directory: %v", err)
	}

	// Create nested directory in subproject
	nestedDir := filepath.Join(subDir, "src")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	// Should find .opencode in subproject, not stop at .git in parent
	projectRoot, err := FindProjectRootFromDir(nestedDir)
	if err != nil {
		t.Fatalf("Expected to find project root, got error: %v", err)
	}

	if projectRoot != subDir {
		t.Errorf("Expected project root to be %q (subproject with .opencode), got %q", subDir, projectRoot)
	}
}
