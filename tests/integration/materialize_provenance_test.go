//go:build integration
// +build integration

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tgaines/agent-smith/internal/testutil"
	"github.com/tgaines/agent-smith/pkg/project"
)

// TestMaterializeProvenance verifies that provenance tracking captures
// all required metadata for materialized components.
// This test covers Story-004 acceptance criteria:
// - .materializations.json file created in .opencode/ or .claude/
// - Metadata includes: source repo URL, source type, commit hash, original path, materialization timestamp
// - Metadata includes sourceHash and currentHash for sync detection
// - Metadata loaded from lock files
// - JSON formatted with indentation for readability
func TestMaterializeProvenance(t *testing.T) {
	// Save current directory and restore it after test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(originalDir)
	})

	// Create temporary directories
	tempDir := testutil.CreateTempDir(t, "agent-smith-provenance-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	// Build the binary
	binaryPath := filepath.Join(tempDir, "agent-smith")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = filepath.Join(originalDir, "../..")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, string(output))
	}

	// Set up test environment
	baseDir := filepath.Join(tempDir, ".agent-smith")
	skillsDir := filepath.Join(baseDir, "skills")
	testSkillDir := filepath.Join(skillsDir, "provenance-test")

	// Create directories
	mkdirErr := os.MkdirAll(testSkillDir, 0755)
	testutil.AssertNoError(t, mkdirErr, "Failed to create test skill directory")

	// Create test skill with multiple files in nested structure
	skillContent := "# Provenance Test Skill\n\nThis is a test skill for provenance tracking."
	err = os.WriteFile(filepath.Join(testSkillDir, "SKILL.md"), []byte(skillContent), 0644)
	testutil.AssertNoError(t, err, "Failed to write SKILL.md")

	// Add a subdirectory with a file
	subDir := filepath.Join(testSkillDir, "lib")
	err = os.MkdirAll(subDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create lib subdirectory")
	err = os.WriteFile(filepath.Join(subDir, "helper.md"), []byte("# Helper"), 0644)
	testutil.AssertNoError(t, err, "Failed to write helper.md")

	t.Logf("Created test skill at: %s", testSkillDir)

	// Create lock file with comprehensive metadata
	lockFilePath := filepath.Join(baseDir, ".skill-lock.json")
	lockData := map[string]interface{}{
		"version": 3,
		"skills": map[string]interface{}{
			"provenance-test": map[string]interface{}{
				"source":       "test-provenance-repo",
				"sourceType":   "github",
				"sourceUrl":    "https://github.com/example/provenance-test",
				"commitHash":   "fedcba9876543210",
				"originalPath": "skills/provenance-test/SKILL.md",
				"installedAt":  "2024-01-15T10:30:00Z",
				"updatedAt":    "2024-01-16T14:20:00Z",
				"version":      3,
			},
		},
	}
	lockJSON, err := json.MarshalIndent(lockData, "", "  ")
	testutil.AssertNoError(t, err, "Failed to marshal lock data")
	err = os.WriteFile(lockFilePath, lockJSON, 0644)
	testutil.AssertNoError(t, err, "Failed to write lock file")

	t.Logf("Created lock file at: %s", lockFilePath)

	// Create project directory
	projectDir := filepath.Join(tempDir, "test-project")
	opencodeDir := filepath.Join(projectDir, ".opencode")
	err = os.MkdirAll(opencodeDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create project directory")

	t.Logf("Created project directory at: %s", projectDir)

	// Change to project directory for auto-detection
	err = os.Chdir(projectDir)
	testutil.AssertNoError(t, err, "Failed to change to project directory")

	// Capture start time (before materialization)
	startTime := time.Now()

	// Run materialize command
	cmd = exec.Command(binaryPath, "materialize", "skill", "provenance-test", "--target", "opencode", "--verbose")
	matOutput, matErr := cmd.CombinedOutput()
	outputStr := string(matOutput)
	t.Logf("Materialize output:\n%s", outputStr)

	if matErr != nil {
		t.Fatalf("Materialize failed: %v\nOutput: %s", matErr, outputStr)
	}

	// Capture end time (after materialization)
	endTime := time.Now()

	// Verify metadata file exists
	metadataPath := filepath.Join(opencodeDir, ".materializations.json")
	testutil.AssertFileExists(t, metadataPath)

	// Load and parse metadata
	metadataBytes, err := os.ReadFile(metadataPath)
	testutil.AssertNoError(t, err, "Failed to read metadata file")

	var metadata project.MaterializationMetadata
	err = json.Unmarshal(metadataBytes, &metadata)
	testutil.AssertNoError(t, err, "Failed to parse metadata")

	// Verify metadata structure
	testutil.AssertEqual(t, 1, metadata.Version, "Incorrect metadata version")

	// Verify skill entry exists
	skillMeta, exists := metadata.Skills["provenance-test"]
	testutil.AssertTrue(t, exists, "provenance-test not found in metadata")

	// Story-004 Acceptance Criteria Verification:

	// 1. Source repo URL
	testutil.AssertEqual(t, "https://github.com/example/provenance-test", skillMeta.Source,
		"Source URL not correctly captured from lock file")

	// 2. Source type
	testutil.AssertEqual(t, "github", skillMeta.SourceType,
		"Source type not correctly captured from lock file")

	// 3. Commit hash
	testutil.AssertEqual(t, "fedcba9876543210", skillMeta.CommitHash,
		"Commit hash not correctly captured from lock file")

	// 4. Original path
	testutil.AssertEqual(t, "skills/provenance-test/SKILL.md", skillMeta.OriginalPath,
		"Original path not correctly captured from lock file")

	// 5. Materialization timestamp
	testutil.AssertTrue(t, skillMeta.MaterializedAt != "", "MaterializedAt timestamp is empty")

	// Verify timestamp is in RFC3339 format and within expected time range
	materializedTime, err := time.Parse(time.RFC3339, skillMeta.MaterializedAt)
	testutil.AssertNoError(t, err, "MaterializedAt timestamp is not in RFC3339 format")

	// Check timestamp is reasonable (between start and end of test)
	testutil.AssertTrue(t, materializedTime.After(startTime.Add(-time.Second)),
		"MaterializedAt timestamp is before test started")
	testutil.AssertTrue(t, materializedTime.Before(endTime.Add(time.Second)),
		"MaterializedAt timestamp is after test ended")

	// 6. Source hash for sync detection
	testutil.AssertTrue(t, skillMeta.SourceHash != "", "SourceHash is empty")
	testutil.AssertTrue(t, strings.HasPrefix(skillMeta.SourceHash, "sha256:"),
		"SourceHash does not have sha256: prefix")
	testutil.AssertTrue(t, len(skillMeta.SourceHash) > 10, "SourceHash is too short")

	// 7. Current hash for sync detection
	testutil.AssertTrue(t, skillMeta.CurrentHash != "", "CurrentHash is empty")
	testutil.AssertTrue(t, strings.HasPrefix(skillMeta.CurrentHash, "sha256:"),
		"CurrentHash does not have sha256: prefix")
	testutil.AssertTrue(t, len(skillMeta.CurrentHash) > 10, "CurrentHash is too short")

	// Verify that sourceHash and currentHash match (since we just materialized)
	testutil.AssertEqual(t, skillMeta.SourceHash, skillMeta.CurrentHash,
		"SourceHash and CurrentHash should be identical immediately after materialization")

	// 8. Verify JSON is formatted with indentation for readability
	metadataStr := string(metadataBytes)
	testutil.AssertTrue(t, strings.Contains(metadataStr, "\n"),
		"Metadata JSON is not formatted with newlines")
	testutil.AssertTrue(t, strings.Contains(metadataStr, "  "),
		"Metadata JSON is not formatted with indentation")

	// Verify proper JSON structure (should be parseable with proper indentation)
	var prettyJSON map[string]interface{}
	err = json.Unmarshal(metadataBytes, &prettyJSON)
	testutil.AssertNoError(t, err, "Metadata JSON is not valid")

	// Verify the materialized component files exist and match source
	destSkillPath := filepath.Join(opencodeDir, "skills", "provenance-test", "SKILL.md")
	testutil.AssertFileExists(t, destSkillPath)

	destContent, err := os.ReadFile(destSkillPath)
	testutil.AssertNoError(t, err, "Failed to read materialized SKILL.md")
	testutil.AssertEqual(t, skillContent, string(destContent), "Materialized content does not match source")

	// Verify nested file structure is preserved
	destHelperPath := filepath.Join(opencodeDir, "skills", "provenance-test", "lib", "helper.md")
	testutil.AssertFileExists(t, destHelperPath)

	t.Logf("✓ All Story-004 provenance tracking acceptance criteria verified successfully")
	t.Logf("  - Source repo URL: %s", skillMeta.Source)
	t.Logf("  - Source type: %s", skillMeta.SourceType)
	t.Logf("  - Commit hash: %s", skillMeta.CommitHash)
	t.Logf("  - Original path: %s", skillMeta.OriginalPath)
	t.Logf("  - Materialized at: %s", skillMeta.MaterializedAt)
	t.Logf("  - Source hash: %s", skillMeta.SourceHash[:20]+"...")
	t.Logf("  - Current hash: %s", skillMeta.CurrentHash[:20]+"...")
	t.Logf("  - JSON formatted with indentation: YES")
}

// TestMaterializeProvenanceMultipleComponents verifies that metadata
// correctly tracks multiple components across different types
func TestMaterializeProvenanceMultipleComponents(t *testing.T) {
	// Save current directory and restore it after test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(originalDir)
	})

	// Create temporary directories
	tempDir := testutil.CreateTempDir(t, "agent-smith-provenance-multi-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	// Build the binary
	binaryPath := filepath.Join(tempDir, "agent-smith")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = filepath.Join(originalDir, "../..")
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, string(output))
	}

	// Set up test environment
	baseDir := filepath.Join(tempDir, ".agent-smith")

	// Create multiple component types
	components := []struct {
		componentType string
		componentName string
		fileName      string
		content       string
		sourceUrl     string
		commitHash    string
	}{
		{
			componentType: "skills",
			componentName: "skill-one",
			fileName:      "SKILL.md",
			content:       "# Skill One",
			sourceUrl:     "https://github.com/example/skill-one",
			commitHash:    "skill1hash123",
		},
		{
			componentType: "skills",
			componentName: "skill-two",
			fileName:      "SKILL.md",
			content:       "# Skill Two",
			sourceUrl:     "https://github.com/example/skill-two",
			commitHash:    "skill2hash456",
		},
		{
			componentType: "agents",
			componentName: "agent-one",
			fileName:      "AGENT.md",
			content:       "# Agent One",
			sourceUrl:     "https://github.com/example/agent-one",
			commitHash:    "agent1hash789",
		},
	}

	// Create components and lock files
	for _, comp := range components {
		compDir := filepath.Join(baseDir, comp.componentType, comp.componentName)
		err := os.MkdirAll(compDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create component directory")

		err = os.WriteFile(filepath.Join(compDir, comp.fileName), []byte(comp.content), 0644)
		testutil.AssertNoError(t, err, "Failed to write component file")

		// Create lock file for this component type
		lockFilePath := filepath.Join(baseDir, "."+comp.componentType[:len(comp.componentType)-1]+"-lock.json")

		// Read existing lock file or create new one
		var lockFile map[string]interface{}
		lockData, err := os.ReadFile(lockFilePath)
		if err != nil {
			lockFile = map[string]interface{}{
				"version":          3,
				comp.componentType: make(map[string]interface{}),
			}
		} else {
			json.Unmarshal(lockData, &lockFile)
		}

		// Add component to lock file
		componentMap := lockFile[comp.componentType].(map[string]interface{})
		componentMap[comp.componentName] = map[string]interface{}{
			"source":       "test-repo",
			"sourceType":   "github",
			"sourceUrl":    comp.sourceUrl,
			"commitHash":   comp.commitHash,
			"originalPath": comp.componentType + "/" + comp.componentName + "/" + comp.fileName,
			"installedAt":  "2024-01-01T00:00:00Z",
			"updatedAt":    "2024-01-01T00:00:00Z",
			"version":      3,
		}

		lockJSON, err := json.MarshalIndent(lockFile, "", "  ")
		testutil.AssertNoError(t, err, "Failed to marshal lock data")
		err = os.WriteFile(lockFilePath, lockJSON, 0644)
		testutil.AssertNoError(t, err, "Failed to write lock file")
	}

	// Create project directory
	projectDir := filepath.Join(tempDir, "test-project")
	opencodeDir := filepath.Join(projectDir, ".opencode")
	projErr := os.MkdirAll(opencodeDir, 0755)
	testutil.AssertNoError(t, projErr, "Failed to create project directory")

	// Change to project directory
	chdirErr := os.Chdir(projectDir)
	testutil.AssertNoError(t, chdirErr, "Failed to change to project directory")

	// Materialize all components
	for _, comp := range components {
		cmd := exec.Command(binaryPath, "materialize",
			comp.componentType[:len(comp.componentType)-1], // Remove trailing 's'
			comp.componentName,
			"--target", "opencode")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to materialize %s %s: %v\nOutput: %s",
				comp.componentType, comp.componentName, err, string(output))
		}
		t.Logf("Materialized %s %s", comp.componentType, comp.componentName)
	}

	// Load metadata
	metadataPath := filepath.Join(opencodeDir, ".materializations.json")
	metadataBytes, err := os.ReadFile(metadataPath)
	testutil.AssertNoError(t, err, "Failed to read metadata file")

	var metadata project.MaterializationMetadata
	err = json.Unmarshal(metadataBytes, &metadata)
	testutil.AssertNoError(t, err, "Failed to parse metadata")

	// Verify all components are tracked
	testutil.AssertEqual(t, 2, len(metadata.Skills), "Expected 2 skills in metadata")
	testutil.AssertEqual(t, 1, len(metadata.Agents), "Expected 1 agent in metadata")

	// Verify each component has unique provenance
	skill1, exists := metadata.Skills["skill-one"]
	testutil.AssertTrue(t, exists, "skill-one not found in metadata")
	testutil.AssertEqual(t, "https://github.com/example/skill-one", skill1.Source,
		"skill-one has incorrect source")
	testutil.AssertEqual(t, "skill1hash123", skill1.CommitHash,
		"skill-one has incorrect commit hash")

	skill2, exists := metadata.Skills["skill-two"]
	testutil.AssertTrue(t, exists, "skill-two not found in metadata")
	testutil.AssertEqual(t, "https://github.com/example/skill-two", skill2.Source,
		"skill-two has incorrect source")
	testutil.AssertEqual(t, "skill2hash456", skill2.CommitHash,
		"skill-two has incorrect commit hash")

	agent1, exists := metadata.Agents["agent-one"]
	testutil.AssertTrue(t, exists, "agent-one not found in metadata")
	testutil.AssertEqual(t, "https://github.com/example/agent-one", agent1.Source,
		"agent-one has incorrect source")
	testutil.AssertEqual(t, "agent1hash789", agent1.CommitHash,
		"agent-one has incorrect commit hash")

	// Verify each component has different hashes (different content)
	testutil.AssertTrue(t, skill1.SourceHash != skill2.SourceHash,
		"Different components should have different hashes")
	testutil.AssertTrue(t, skill1.SourceHash != agent1.SourceHash,
		"Different component types should have different hashes")

	t.Logf("✓ Multiple component provenance tracking verified successfully")
}
