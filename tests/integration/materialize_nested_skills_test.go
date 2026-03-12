//go:build integration
// +build integration

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tjg184/agent-smith/internal/testutil"
)

// TestMaterializeSkillsWithNestedFilesystemNames verifies that `materialize skills`
// correctly handles components whose filesystemName contains path separators
// (e.g. "kotlin/convert-groovy-kotlin"). This was a bug where the filesystem walk
// only saw the top-level directory "kotlin/" and failed to match any lock file entry,
// producing spurious "Skipping untracked directory" warnings and silently skipping
// all nested skills.
func TestMaterializeSkillsWithNestedFilesystemNames(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() { os.Chdir(originalDir) })

	tempDir := testutil.CreateTempDir(t, "agent-smith-nested-skills-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	binaryPath := AgentSmithBinary

	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	lockFilePath := filepath.Join(agentSmithDir, ".component-lock.json")

	// Skills whose filesystemName contains a category prefix — the exact shape
	// produced when installing from a repo with structure skills/<category>/<skill>/SKILL.md
	nestedSkills := []struct {
		name           string
		filesystemName string
		content        string
	}{
		{
			name:           "convert-groovy-kotlin",
			filesystemName: "kotlin/convert-groovy-kotlin",
			content:        "# Convert Groovy to Kotlin",
		},
		{
			name:           "design-architecture",
			filesystemName: "architecture/design-architecture",
			content:        "# Design Architecture",
		},
		{
			name:           "audit-security",
			filesystemName: "security/audit-security",
			content:        "# Audit Security",
		},
	}

	// Also include a flat skill to confirm it still works alongside nested ones
	flatSkills := []struct {
		name           string
		filesystemName string
		content        string
	}{
		{
			name:           "onboarding",
			filesystemName: "onboarding",
			content:        "# Onboarding",
		},
	}

	sourceURL := "git@github.com:example/repo"

	for _, skill := range nestedSkills {
		skillDir := filepath.Join(agentSmithDir, "skills", skill.filesystemName)
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			t.Fatalf("Failed to create skill dir %s: %v", skillDir, err)
		}
		skillFile := filepath.Join(skillDir, "SKILL.md")
		if err := os.WriteFile(skillFile, []byte(skill.content), 0644); err != nil {
			t.Fatalf("Failed to write skill file %s: %v", skillFile, err)
		}
		testutil.AddComponentToLockFile(t, lockFilePath, "skills", skill.name, sourceURL, map[string]interface{}{
			"source":         sourceURL,
			"sourceType":     "git",
			"sourceUrl":      sourceURL,
			"commitHash":     "abc123",
			"filesystemName": skill.filesystemName,
			"components":     1,
			"detection":      "single",
		})
	}

	for _, skill := range flatSkills {
		skillDir := filepath.Join(agentSmithDir, "skills", skill.filesystemName)
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			t.Fatalf("Failed to create skill dir %s: %v", skillDir, err)
		}
		skillFile := filepath.Join(skillDir, "SKILL.md")
		if err := os.WriteFile(skillFile, []byte(skill.content), 0644); err != nil {
			t.Fatalf("Failed to write skill file %s: %v", skillFile, err)
		}
		testutil.AddComponentToLockFile(t, lockFilePath, "skills", skill.name, sourceURL, map[string]interface{}{
			"source":         sourceURL,
			"sourceType":     "git",
			"sourceUrl":      sourceURL,
			"commitHash":     "abc123",
			"filesystemName": skill.filesystemName,
			"components":     1,
			"detection":      "single",
		})
	}

	projectDir := filepath.Join(tempDir, "test-project")
	opencodeDir := filepath.Join(projectDir, ".opencode")
	if err := os.MkdirAll(opencodeDir, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	cmd := exec.Command(binaryPath, "materialize", "skills", "--target", "opencode")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	t.Logf("Output:\n%s", outputStr)

	if err != nil {
		t.Fatalf("materialize skills failed: %v\nOutput: %s", err, outputStr)
	}

	if strings.Contains(outputStr, "Skipping untracked directory") {
		t.Errorf("Got unexpected 'Skipping untracked directory' warnings — nested skills are being ignored:\n%s", outputStr)
	}

	// All nested skills should be materialized to their flat destination names
	for _, skill := range nestedSkills {
		destPath := filepath.Join(opencodeDir, "skills", skill.filesystemName, "SKILL.md")
		testutil.AssertFileExists(t, destPath)

		content, err := os.ReadFile(destPath)
		if err != nil {
			t.Errorf("Failed to read materialized file for %s: %v", skill.name, err)
			continue
		}
		if string(content) != skill.content {
			t.Errorf("Content mismatch for %s: got %q, want %q", skill.name, string(content), skill.content)
		}
	}

	for _, skill := range flatSkills {
		destPath := filepath.Join(opencodeDir, "skills", skill.filesystemName, "SKILL.md")
		testutil.AssertFileExists(t, destPath)
	}
}
