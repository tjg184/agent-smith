package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDirectoryCopyingWithResources tests that copyDirectoryContents preserves all files
func TestDirectoryCopyingWithResources(t *testing.T) {
	// Create temporary source directory
	srcDir, err := os.MkdirTemp("", "src-test-*")
	if err != nil {
		t.Fatalf("Failed to create src temp directory: %v", err)
	}
	defer os.RemoveAll(srcDir)

	// Create temporary destination directory
	dstDir, err := os.MkdirTemp("", "dst-test-*")
	if err != nil {
		t.Fatalf("Failed to create dst temp directory: %v", err)
	}
	defer os.RemoveAll(dstDir)

	// Create a complex directory structure with various file types
	testFiles := map[string]string{
		"SKILL.md":                      "# My Skill\nThis is the main skill file",
		"README.md":                     "# Documentation\nHow to use this skill",
		"template.txt":                  "Template content",
		"config.json":                   `{"setting": "value"}`,
		"support/helper.md":             "# Helper\nSupport documentation",
		"support/example.txt":           "Example file",
		"resources/image.txt":           "image placeholder",
		"resources/data/sample.csv":     "col1,col2\nval1,val2",
		"nested/deep/structure/file.md": "Deeply nested file",
		".hidden":                       "Hidden file content",
	}

	// Create all test files
	for relPath, content := range testFiles {
		fullPath := filepath.Join(srcDir, relPath)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Create a SkillDownloader instance
	sd := &SkillDownloader{
		baseDir: filepath.Join(os.TempDir(), "skills"),
	}

	// Copy directory contents
	err = sd.copyDirectoryContents(srcDir, dstDir)
	if err != nil {
		t.Fatalf("copyDirectoryContents failed: %v", err)
	}

	// Verify all files were copied
	for relPath, expectedContent := range testFiles {
		dstPath := filepath.Join(dstDir, relPath)

		// Check file exists
		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			t.Errorf("File not copied: %s", relPath)
			continue
		}

		// Check file content
		actualContent, err := os.ReadFile(dstPath)
		if err != nil {
			t.Errorf("Failed to read copied file %s: %v", relPath, err)
			continue
		}

		if string(actualContent) != expectedContent {
			t.Errorf("File content mismatch for %s:\nExpected: %s\nActual: %s",
				relPath, expectedContent, string(actualContent))
		}
	}

	// Verify directory structure is preserved
	expectedDirs := []string{
		"support",
		"resources",
		"resources/data",
		"nested",
		"nested/deep",
		"nested/deep/structure",
	}

	for _, dir := range expectedDirs {
		dirPath := filepath.Join(dstDir, dir)
		info, err := os.Stat(dirPath)
		if err != nil {
			t.Errorf("Directory not created: %s (error: %v)", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("Path exists but is not a directory: %s", dir)
		}
	}

	t.Logf("SUCCESS: All %d files and %d directories copied correctly",
		len(testFiles), len(expectedDirs))
}

// TestComponentDownloadPreservesResources tests end-to-end component download with resources
func TestComponentDownloadPreservesResources(t *testing.T) {
	// Create temporary repository directory
	repoDir, err := os.MkdirTemp("", "repo-test-*")
	if err != nil {
		t.Fatalf("Failed to create repo temp directory: %v", err)
	}
	defer os.RemoveAll(repoDir)

	// Create a skill with support files
	skillDir := filepath.Join(repoDir, "skills", "my-skill")
	testFiles := map[string]string{
		filepath.Join(skillDir, "SKILL.md"):                 "# My Skill",
		filepath.Join(skillDir, "README.md"):                "# Documentation",
		filepath.Join(skillDir, "template.md"):              "# Template",
		filepath.Join(skillDir, "resources", "example.txt"): "Example resource",
		filepath.Join(skillDir, "support", "helper.md"):     "# Helper",
	}

	// Create all test files
	for fullPath, content := range testFiles {
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Detect components in the repository
	detector := NewRepositoryDetector()
	components, err := detector.detectComponentsInRepo(repoDir)
	if err != nil {
		t.Fatalf("Failed to detect components: %v", err)
	}

	// Filter for skills
	var skillComponents []DetectedComponent
	for _, comp := range components {
		if comp.Type == ComponentSkill {
			skillComponents = append(skillComponents, comp)
		}
	}

	if len(skillComponents) != 1 {
		t.Fatalf("Expected 1 skill component, got %d", len(skillComponents))
	}

	component := skillComponents[0]

	// Verify component path is correct
	expectedPath := filepath.Join("skills", "my-skill")
	if component.Path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, component.Path)
	}

	// Create temporary destination directory
	dstDir, err := os.MkdirTemp("", "dst-skill-test-*")
	if err != nil {
		t.Fatalf("Failed to create dst temp directory: %v", err)
	}
	defer os.RemoveAll(dstDir)

	// Create a SkillDownloader and copy the component
	sd := &SkillDownloader{
		baseDir:  dstDir,
		detector: detector,
	}

	componentSrcPath := filepath.Join(repoDir, component.Path)
	componentDstPath := filepath.Join(dstDir, component.Name)

	err = os.MkdirAll(componentDstPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create component dst directory: %v", err)
	}

	err = sd.copyDirectoryContents(componentSrcPath, componentDstPath)
	if err != nil {
		t.Fatalf("Failed to copy component: %v", err)
	}

	// Verify all files were copied
	expectedFiles := []string{
		"SKILL.md",
		"README.md",
		"template.md",
		filepath.Join("resources", "example.txt"),
		filepath.Join("support", "helper.md"),
	}

	for _, relPath := range expectedFiles {
		fullPath := filepath.Join(componentDstPath, relPath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected file not copied: %s", relPath)
		}
	}

	t.Logf("SUCCESS: Component downloaded with all %d support files preserved", len(expectedFiles))
}

// TestMultipleComponentsWithResources tests that multiple components can be downloaded independently
func TestMultipleComponentsWithResources(t *testing.T) {
	// Create temporary repository directory
	repoDir, err := os.MkdirTemp("", "multi-repo-test-*")
	if err != nil {
		t.Fatalf("Failed to create repo temp directory: %v", err)
	}
	defer os.RemoveAll(repoDir)

	// Create multiple skills, each with their own resources
	skills := map[string][]string{
		"skill-a": {"SKILL.md", "README.md", "resources/data.txt"},
		"skill-b": {"SKILL.md", "template.md", "support/helper.md"},
		"skill-c": {"SKILL.md", "config.json"},
	}

	for skillName, files := range skills {
		skillDir := filepath.Join(repoDir, "skills", skillName)
		for _, file := range files {
			fullPath := filepath.Join(skillDir, file)
			dir := filepath.Dir(fullPath)

			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("Failed to create directory %s: %v", dir, err)
			}

			content := "# Content for " + skillName + "/" + file
			if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to create file %s: %v", fullPath, err)
			}
		}
	}

	// Detect components
	detector := NewRepositoryDetector()
	components, err := detector.detectComponentsInRepo(repoDir)
	if err != nil {
		t.Fatalf("Failed to detect components: %v", err)
	}

	// Filter for skills
	var skillComponents []DetectedComponent
	for _, comp := range components {
		if comp.Type == ComponentSkill {
			skillComponents = append(skillComponents, comp)
		}
	}

	if len(skillComponents) != 3 {
		t.Fatalf("Expected 3 skill components, got %d", len(skillComponents))
	}

	// Create temporary destination directory
	dstDir, err := os.MkdirTemp("", "multi-dst-test-*")
	if err != nil {
		t.Fatalf("Failed to create dst temp directory: %v", err)
	}
	defer os.RemoveAll(dstDir)

	// Copy each component
	sd := &SkillDownloader{
		baseDir:  dstDir,
		detector: detector,
	}

	for _, component := range skillComponents {
		componentSrcPath := filepath.Join(repoDir, component.Path)
		componentDstPath := filepath.Join(dstDir, component.Name)

		err = os.MkdirAll(componentDstPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create component dst directory: %v", err)
		}

		err = sd.copyDirectoryContents(componentSrcPath, componentDstPath)
		if err != nil {
			t.Fatalf("Failed to copy component %s: %v", component.Name, err)
		}

		// Verify files for this component
		expectedFiles := skills[component.Name]
		for _, relPath := range expectedFiles {
			fullPath := filepath.Join(componentDstPath, relPath)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Errorf("Expected file not copied for %s: %s", component.Name, relPath)
			}
		}
	}

	t.Logf("SUCCESS: All 3 components downloaded independently with resources preserved")
}
