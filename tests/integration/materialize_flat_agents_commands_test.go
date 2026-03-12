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

// TestMaterializeFlatAgentsAndCommands verifies that agents and commands are materialized
// as flat .md files directly inside the component type directory on opencode and claudecode
// targets, rather than wrapped in a subdirectory. This matches what `link` produces and
// what those editors actually load.
//
// Expected:  .opencode/agents/architect.md  (not .opencode/agents/architect/architect.md)
func TestMaterializeFlatAgentsAndCommands(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() { os.Chdir(originalDir) })

	tempDir := testutil.CreateTempDir(t, "agent-smith-flat-agents-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	binaryPath := AgentSmithBinary
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	lockFilePath := filepath.Join(agentSmithDir, ".component-lock.json")
	sourceURL := "git@github.com:example/repo"

	type component struct {
		compType       string
		name           string
		filesystemName string
		fileName       string
		content        string
	}

	components := []component{
		{
			compType:       "agents",
			name:           "architect",
			filesystemName: "architect",
			fileName:       "architect.md",
			content:        "---\ndescription: Software architect agent\n---\nYou are an architect.",
		},
		{
			compType:       "agents",
			name:           "code-reviewer",
			filesystemName: "code-reviewer",
			fileName:       "code-reviewer.md",
			content:        "---\ndescription: Code review agent\n---\nYou are a code reviewer.",
		},
		{
			compType:       "commands",
			name:           "review-diff",
			filesystemName: "review-diff",
			fileName:       "review-diff.md",
			content:        "# Review Diff\nReviews a git diff.",
		},
	}

	for _, c := range components {
		dir := filepath.Join(agentSmithDir, c.compType, c.filesystemName)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}
		if err := os.WriteFile(filepath.Join(dir, c.fileName), []byte(c.content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
		testutil.AddComponentToLockFile(t, lockFilePath, c.compType, c.name, sourceURL, map[string]interface{}{
			"source":         sourceURL,
			"sourceType":     "git",
			"sourceUrl":      sourceURL,
			"commitHash":     "abc123",
			"filesystemName": c.filesystemName,
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

	for _, c := range components {
		t.Run(c.compType+"/"+c.name, func(t *testing.T) {
			singularType := c.compType[:len(c.compType)-1]
			cmd := exec.Command(binaryPath, "materialize", singularType, c.name, "--target", "opencode")
			out, err := cmd.CombinedOutput()
			outStr := string(out)
			t.Logf("Output:\n%s", outStr)
			if err != nil {
				t.Fatalf("materialize failed: %v\nOutput: %s", err, outStr)
			}

			// Flat file must exist
			flatPath := filepath.Join(opencodeDir, c.compType, c.fileName)
			testutil.AssertFileExists(t, flatPath)

			content, err := os.ReadFile(flatPath)
			if err != nil {
				t.Fatalf("Failed to read materialized file: %v", err)
			}
			if string(content) != c.content {
				t.Errorf("Content mismatch: got %q, want %q", string(content), c.content)
			}

			// Wrapper subdirectory must NOT exist
			wrapperDir := filepath.Join(opencodeDir, c.compType, c.filesystemName)
			if info, err := os.Stat(wrapperDir); err == nil && info.IsDir() {
				t.Errorf("Wrapper directory %s should not exist; agent/command should be a flat file", wrapperDir)
			}
		})
	}
}

// TestMaterializeFlatAgentsIdempotent verifies that re-running materialize on already-materialized
// flat agents correctly skips them (files match at the flat path) without re-copying.
func TestMaterializeFlatAgentsIdempotent(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() { os.Chdir(originalDir) })

	tempDir := testutil.CreateTempDir(t, "agent-smith-flat-idempotent-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	binaryPath := AgentSmithBinary
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	lockFilePath := filepath.Join(agentSmithDir, ".component-lock.json")
	sourceURL := "git@github.com:example/repo"

	agentDir := filepath.Join(agentSmithDir, "agents", "architect")
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		t.Fatalf("Failed to create agent dir: %v", err)
	}
	content := "---\ndescription: Architect\n---\nYou are an architect."
	if err := os.WriteFile(filepath.Join(agentDir, "architect.md"), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write agent file: %v", err)
	}
	testutil.AddComponentToLockFile(t, lockFilePath, "agents", "architect", sourceURL, map[string]interface{}{
		"source":         sourceURL,
		"sourceType":     "git",
		"sourceUrl":      sourceURL,
		"commitHash":     "abc123",
		"filesystemName": "architect",
		"components":     1,
		"detection":      "single",
	})

	projectDir := filepath.Join(tempDir, "test-project")
	opencodeDir := filepath.Join(projectDir, ".opencode")
	if err := os.MkdirAll(opencodeDir, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	run := func(label string) string {
		cmd := exec.Command(binaryPath, "materialize", "agent", "architect", "--target", "opencode")
		out, err := cmd.CombinedOutput()
		outStr := string(out)
		t.Logf("%s output:\n%s", label, outStr)
		if err != nil {
			t.Fatalf("%s failed: %v\nOutput: %s", label, err, outStr)
		}
		return outStr
	}

	firstOut := run("First run")

	flatPath := filepath.Join(opencodeDir, "agents", "architect.md")
	testutil.AssertFileExists(t, flatPath)

	if strings.Contains(firstOut, "already exists and identical") {
		t.Errorf("First run should not skip: %s", firstOut)
	}

	secondOut := run("Second run")

	testutil.AssertFileExists(t, flatPath)

	if !strings.Contains(secondOut, "already exists and identical") {
		t.Errorf("Second run should skip (already exists and identical):\n%s", secondOut)
	}

	if strings.Contains(secondOut, "Materialized agents") {
		t.Errorf("Second run should not re-materialize:\n%s", secondOut)
	}
}
