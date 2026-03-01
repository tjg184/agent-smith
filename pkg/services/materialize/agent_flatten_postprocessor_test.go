package materialize

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tjg184/agent-smith/internal/formatter"
)

func TestAgentFlattenPostprocessor_ShouldProcess(t *testing.T) {
	tests := []struct {
		name          string
		componentType string
		target        string
		want          bool
	}{
		{
			name:          "should process agents for copilot",
			componentType: "agents",
			target:        "copilot",
			want:          true,
		},
		{
			name:          "should not process agents for opencode",
			componentType: "agents",
			target:        "opencode",
			want:          false,
		},
		{
			name:          "should not process agents for claudecode",
			componentType: "agents",
			target:        "claudecode",
			want:          false,
		},
		{
			name:          "should not process skills for copilot",
			componentType: "skills",
			target:        "copilot",
			want:          false,
		},
		{
			name:          "should not process commands for copilot",
			componentType: "commands",
			target:        "copilot",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewAgentFlattenPostprocessor()
			got := p.ShouldProcess(tt.componentType, tt.target)
			if got != tt.want {
				t.Errorf("ShouldProcess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgentFlattenPostprocessor_Name(t *testing.T) {
	p := NewAgentFlattenPostprocessor()
	if p.Name() != "AgentFlattenPostprocessor" {
		t.Errorf("Name() = %v, want AgentFlattenPostprocessor", p.Name())
	}
}

func TestAgentFlattenPostprocessor_Process_Success(t *testing.T) {
	// Create temp directory structure
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, ".github")
	agentsDir := filepath.Join(targetDir, "agents")
	agentFolder := filepath.Join(agentsDir, "my-agent")

	// Create directory structure
	if err := os.MkdirAll(agentFolder, 0755); err != nil {
		t.Fatalf("Failed to create agent folder: %v", err)
	}

	// Create agent file
	agentFile := filepath.Join(agentFolder, "my-agent.md")
	if err := os.WriteFile(agentFile, []byte("# My Agent"), 0644); err != nil {
		t.Fatalf("Failed to create agent file: %v", err)
	}

	// Run postprocessor
	p := NewAgentFlattenPostprocessor()
	ctx := PostprocessContext{
		ComponentType:  "agents",
		ComponentName:  "my-agent",
		FilesystemName: "my-agent",
		Target:         "copilot",
		TargetDir:      targetDir,
		DestPath:       agentFolder,
		DryRun:         false,
		Formatter:      formatter.New(),
	}

	if err := p.Process(ctx); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	// Verify symlink was created
	symlinkPath := filepath.Join(agentsDir, "my-agent.md")
	info, err := os.Lstat(symlinkPath)
	if err != nil {
		t.Fatalf("Symlink not created: %v", err)
	}

	// Verify it's actually a symlink
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("Created file is not a symlink")
	}

	// Verify symlink target is relative
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}

	expectedTarget := "my-agent/my-agent.md"
	if target != expectedTarget {
		t.Errorf("Symlink target = %v, want %v", target, expectedTarget)
	}

	// Verify symlink is readable and points to correct content
	content, err := os.ReadFile(symlinkPath)
	if err != nil {
		t.Fatalf("Failed to read through symlink: %v", err)
	}

	if string(content) != "# My Agent" {
		t.Errorf("Symlink content = %v, want '# My Agent'", string(content))
	}
}

func TestAgentFlattenPostprocessor_Process_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, ".github")
	agentsDir := filepath.Join(targetDir, "agents")
	agentFolder := filepath.Join(agentsDir, "my-agent")

	if err := os.MkdirAll(agentFolder, 0755); err != nil {
		t.Fatalf("Failed to create agent folder: %v", err)
	}

	agentFile := filepath.Join(agentFolder, "my-agent.md")
	if err := os.WriteFile(agentFile, []byte("# My Agent"), 0644); err != nil {
		t.Fatalf("Failed to create agent file: %v", err)
	}

	p := NewAgentFlattenPostprocessor()
	ctx := PostprocessContext{
		ComponentType:  "agents",
		ComponentName:  "my-agent",
		FilesystemName: "my-agent",
		Target:         "copilot",
		TargetDir:      targetDir,
		DestPath:       agentFolder,
		DryRun:         true,
		Formatter:      formatter.New(),
	}

	if err := p.Process(ctx); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	// Verify symlink was NOT created in dry-run mode
	symlinkPath := filepath.Join(agentsDir, "my-agent.md")
	if _, err := os.Lstat(symlinkPath); err == nil {
		t.Error("Symlink should not be created in dry-run mode")
	}
}

func TestAgentFlattenPostprocessor_Process_MissingAgentFile(t *testing.T) {
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, ".github")
	agentsDir := filepath.Join(targetDir, "agents")
	agentFolder := filepath.Join(agentsDir, "my-agent")

	if err := os.MkdirAll(agentFolder, 0755); err != nil {
		t.Fatalf("Failed to create agent folder: %v", err)
	}

	// Don't create the agent file - should handle gracefully

	p := NewAgentFlattenPostprocessor()
	ctx := PostprocessContext{
		ComponentType:  "agents",
		ComponentName:  "my-agent",
		FilesystemName: "my-agent",
		Target:         "copilot",
		TargetDir:      targetDir,
		DestPath:       agentFolder,
		DryRun:         false,
		Formatter:      formatter.New(),
	}

	// Should not error - just log warning
	if err := p.Process(ctx); err != nil {
		t.Errorf("Process() should not error on missing file, got: %v", err)
	}

	// Symlink should not be created
	symlinkPath := filepath.Join(agentsDir, "my-agent.md")
	if _, err := os.Lstat(symlinkPath); err == nil {
		t.Error("Symlink should not be created when agent file is missing")
	}
}

func TestAgentFlattenPostprocessor_Process_SymlinkAlreadyExists(t *testing.T) {
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, ".github")
	agentsDir := filepath.Join(targetDir, "agents")
	agentFolder := filepath.Join(agentsDir, "my-agent")

	if err := os.MkdirAll(agentFolder, 0755); err != nil {
		t.Fatalf("Failed to create agent folder: %v", err)
	}

	agentFile := filepath.Join(agentFolder, "my-agent.md")
	if err := os.WriteFile(agentFile, []byte("# My Agent"), 0644); err != nil {
		t.Fatalf("Failed to create agent file: %v", err)
	}

	// Create symlink manually
	symlinkPath := filepath.Join(agentsDir, "my-agent.md")
	if err := os.Symlink("my-agent/my-agent.md", symlinkPath); err != nil {
		t.Fatalf("Failed to create existing symlink: %v", err)
	}

	p := NewAgentFlattenPostprocessor()
	ctx := PostprocessContext{
		ComponentType:  "agents",
		ComponentName:  "my-agent",
		FilesystemName: "my-agent",
		Target:         "copilot",
		TargetDir:      targetDir,
		DestPath:       agentFolder,
		DryRun:         false,
		Formatter:      formatter.New(),
	}

	// Should not error - idempotent
	if err := p.Process(ctx); err != nil {
		t.Errorf("Process() should be idempotent, got error: %v", err)
	}

	// Symlink should still exist and be correct
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}

	if target != "my-agent/my-agent.md" {
		t.Errorf("Symlink target changed, got %v", target)
	}
}

func TestAgentFlattenPostprocessor_Process_RegularFileConflict(t *testing.T) {
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, ".github")
	agentsDir := filepath.Join(targetDir, "agents")
	agentFolder := filepath.Join(agentsDir, "my-agent")

	if err := os.MkdirAll(agentFolder, 0755); err != nil {
		t.Fatalf("Failed to create agent folder: %v", err)
	}

	agentFile := filepath.Join(agentFolder, "my-agent.md")
	if err := os.WriteFile(agentFile, []byte("# My Agent"), 0644); err != nil {
		t.Fatalf("Failed to create agent file: %v", err)
	}

	// Create a regular file where symlink should go (conflict)
	conflictFile := filepath.Join(agentsDir, "my-agent.md")
	if err := os.WriteFile(conflictFile, []byte("conflict"), 0644); err != nil {
		t.Fatalf("Failed to create conflict file: %v", err)
	}

	p := NewAgentFlattenPostprocessor()
	ctx := PostprocessContext{
		ComponentType:  "agents",
		ComponentName:  "my-agent",
		FilesystemName: "my-agent",
		Target:         "copilot",
		TargetDir:      targetDir,
		DestPath:       agentFolder,
		DryRun:         false,
		Formatter:      formatter.New(),
	}

	// Should return fatal error for file conflict
	err := p.Process(ctx)
	if err == nil {
		t.Error("Process() should error when regular file conflicts with symlink location")
	}
}

func TestAgentFlattenPostprocessor_Cleanup(t *testing.T) {
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, ".github")
	agentsDir := filepath.Join(targetDir, "agents")
	agentFolder := filepath.Join(agentsDir, "my-agent")

	if err := os.MkdirAll(agentFolder, 0755); err != nil {
		t.Fatalf("Failed to create agent folder: %v", err)
	}

	// Create symlink
	symlinkPath := filepath.Join(agentsDir, "my-agent.md")
	if err := os.Symlink("my-agent/my-agent.md", symlinkPath); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	p := NewAgentFlattenPostprocessor()
	ctx := PostprocessContext{
		ComponentType:  "agents",
		ComponentName:  "my-agent",
		FilesystemName: "my-agent",
		Target:         "copilot",
		TargetDir:      targetDir,
		DestPath:       agentFolder,
		DryRun:         false,
		Formatter:      formatter.New(),
	}

	// Run cleanup
	if err := p.Cleanup(ctx); err != nil {
		t.Errorf("Cleanup() error = %v", err)
	}

	// Verify symlink was removed
	if _, err := os.Lstat(symlinkPath); err == nil {
		t.Error("Cleanup() should remove symlink")
	}
}

func TestAgentFlattenPostprocessor_Cleanup_SymlinkMissing(t *testing.T) {
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, ".github")
	agentsDir := filepath.Join(targetDir, "agents")
	agentFolder := filepath.Join(agentsDir, "my-agent")

	if err := os.MkdirAll(agentFolder, 0755); err != nil {
		t.Fatalf("Failed to create agent folder: %v", err)
	}

	// Don't create symlink

	p := NewAgentFlattenPostprocessor()
	ctx := PostprocessContext{
		ComponentType:  "agents",
		ComponentName:  "my-agent",
		FilesystemName: "my-agent",
		Target:         "copilot",
		TargetDir:      targetDir,
		DestPath:       agentFolder,
		DryRun:         false,
		Formatter:      formatter.New(),
	}

	// Should not error when symlink doesn't exist
	if err := p.Cleanup(ctx); err != nil {
		t.Errorf("Cleanup() should handle missing symlink gracefully, got: %v", err)
	}
}

func TestAgentFlattenPostprocessor_Process_MultipleFiles(t *testing.T) {
	// Create temp directory structure
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, ".github")
	agentsDir := filepath.Join(targetDir, "agents")
	agentFolder := filepath.Join(agentsDir, "backend-development")

	// Create directory structure
	if err := os.MkdirAll(agentFolder, 0755); err != nil {
		t.Fatalf("Failed to create agent folder: %v", err)
	}

	// Create multiple agent files
	agentFiles := []string{
		"tdd-orchestrator.md",
		"temporal-python-pro.md",
		"event-sourcing-architect.md",
	}

	for _, filename := range agentFiles {
		content := "# Agent: " + filename
		if err := os.WriteFile(filepath.Join(agentFolder, filename), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create agent file %s: %v", filename, err)
		}
	}

	// Create postprocessor and context
	p := NewAgentFlattenPostprocessor()
	registry := make(map[string]string)
	ctx := PostprocessContext{
		ComponentType:   "agents",
		ComponentName:   "backend-development",
		FilesystemName:  "backend-development",
		Target:          "copilot",
		TargetDir:       targetDir,
		DestPath:        agentFolder,
		DryRun:          false,
		Formatter:       formatter.New(),
		SymlinkRegistry: registry,
	}

	// Run process
	if err := p.Process(ctx); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	// Verify all 3 symlinks created
	for _, filename := range agentFiles {
		symlinkPath := filepath.Join(agentsDir, filename)

		// Check symlink exists
		info, err := os.Lstat(symlinkPath)
		if err != nil {
			t.Errorf("Symlink %s not created: %v", filename, err)
			continue
		}

		// Verify it's a symlink
		if info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("%s is not a symlink", filename)
		}

		// Verify target
		target, err := os.Readlink(symlinkPath)
		expectedTarget := filepath.Join("backend-development", filename)
		if err != nil {
			t.Errorf("Cannot read symlink %s: %v", filename, err)
		} else if target != expectedTarget {
			t.Errorf("Symlink %s target = %s, want %s", filename, target, expectedTarget)
		}

		// Verify symlink is readable
		content, err := os.ReadFile(symlinkPath)
		if err != nil {
			t.Errorf("Cannot read through symlink %s: %v", filename, err)
		} else if len(content) == 0 {
			t.Errorf("Symlink %s points to empty file", filename)
		}
	}

	// Verify all files registered
	if len(registry) != 3 {
		t.Errorf("Registry should have 3 entries, got %d", len(registry))
	}
}

func TestAgentFlattenPostprocessor_Process_MixedFiles(t *testing.T) {
	// Create temp directory structure
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, ".github")
	agentsDir := filepath.Join(targetDir, "agents")
	agentFolder := filepath.Join(agentsDir, "my-agent")

	if err := os.MkdirAll(agentFolder, 0755); err != nil {
		t.Fatalf("Failed to create agent folder: %v", err)
	}

	// Create agent file and non-agent files
	if err := os.WriteFile(filepath.Join(agentFolder, "agent.md"), []byte("# Agent"), 0644); err != nil {
		t.Fatalf("Failed to create agent.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentFolder, "README.md"), []byte("# README"), 0644); err != nil {
		t.Fatalf("Failed to create README.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentFolder, "notes.txt"), []byte("notes"), 0644); err != nil {
		t.Fatalf("Failed to create notes.txt: %v", err)
	}

	p := NewAgentFlattenPostprocessor()
	ctx := PostprocessContext{
		ComponentType:   "agents",
		ComponentName:   "my-agent",
		FilesystemName:  "my-agent",
		Target:          "copilot",
		TargetDir:       targetDir,
		DestPath:        agentFolder,
		DryRun:          false,
		Formatter:       formatter.New(),
		SymlinkRegistry: make(map[string]string),
	}

	if err := p.Process(ctx); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	// Verify only agent.md symlink created
	agentSymlink := filepath.Join(agentsDir, "agent.md")
	if _, err := os.Lstat(agentSymlink); err != nil {
		t.Error("agent.md symlink should be created")
	}

	// Verify README.md symlink NOT created
	readmeSymlink := filepath.Join(agentsDir, "README.md")
	if _, err := os.Lstat(readmeSymlink); err == nil {
		t.Error("README.md symlink should not be created")
	}

	// Verify notes.txt symlink NOT created
	notesSymlink := filepath.Join(agentsDir, "notes.txt")
	if _, err := os.Lstat(notesSymlink); err == nil {
		t.Error("notes.txt symlink should not be created")
	}
}

func TestAgentFlattenPostprocessor_Process_IgnoredFiles(t *testing.T) {
	// Create temp directory structure
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, ".github")
	agentsDir := filepath.Join(targetDir, "agents")
	agentFolder := filepath.Join(agentsDir, "my-agent")

	if err := os.MkdirAll(agentFolder, 0755); err != nil {
		t.Fatalf("Failed to create agent folder: %v", err)
	}

	// Create ignored files (various cases)
	ignoredFiles := []string{
		"README.md",
		"readme.md", // lowercase
		"ReadMe.MD", // mixed case
		"LICENSE.md",
		"DOCS.md",
		"CHANGELOG.md",
	}

	for _, filename := range ignoredFiles {
		if err := os.WriteFile(filepath.Join(agentFolder, filename), []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", filename, err)
		}
	}

	p := NewAgentFlattenPostprocessor()
	ctx := PostprocessContext{
		ComponentType:   "agents",
		ComponentName:   "my-agent",
		FilesystemName:  "my-agent",
		Target:          "copilot",
		TargetDir:       targetDir,
		DestPath:        agentFolder,
		DryRun:          false,
		Formatter:       formatter.New(),
		SymlinkRegistry: make(map[string]string),
	}

	if err := p.Process(ctx); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	// Verify no symlinks created for any ignored files
	for _, filename := range ignoredFiles {
		symlinkPath := filepath.Join(agentsDir, filename)
		if _, err := os.Lstat(symlinkPath); err == nil {
			t.Errorf("Symlink for %s should not be created (ignored file)", filename)
		}
	}
}

func TestAgentFlattenPostprocessor_Process_NameConflict(t *testing.T) {
	// Create temp directory structure
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, ".github")
	agentsDir := filepath.Join(targetDir, "agents")

	// Create two agent folders with same filename
	agent1Folder := filepath.Join(agentsDir, "backend-dev")
	agent2Folder := filepath.Join(agentsDir, "api-scaffold")

	if err := os.MkdirAll(agent1Folder, 0755); err != nil {
		t.Fatalf("Failed to create agent1 folder: %v", err)
	}
	if err := os.MkdirAll(agent2Folder, 0755); err != nil {
		t.Fatalf("Failed to create agent2 folder: %v", err)
	}

	// Both have an "api.md" file
	if err := os.WriteFile(filepath.Join(agent1Folder, "api.md"), []byte("# Backend API"), 0644); err != nil {
		t.Fatalf("Failed to create api.md in agent1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agent2Folder, "api.md"), []byte("# API Scaffold"), 0644); err != nil {
		t.Fatalf("Failed to create api.md in agent2: %v", err)
	}

	p := NewAgentFlattenPostprocessor()
	registry := make(map[string]string)

	// Process first agent
	ctx1 := PostprocessContext{
		ComponentType:   "agents",
		ComponentName:   "backend-dev",
		FilesystemName:  "backend-dev",
		Target:          "copilot",
		TargetDir:       targetDir,
		DestPath:        agent1Folder,
		DryRun:          false,
		Formatter:       formatter.New(),
		SymlinkRegistry: registry,
	}

	if err := p.Process(ctx1); err != nil {
		t.Fatalf("Process() for agent1 error = %v", err)
	}

	// Verify first symlink created
	symlinkPath := filepath.Join(agentsDir, "api.md")
	if _, err := os.Lstat(symlinkPath); err != nil {
		t.Error("First api.md symlink should be created")
	}

	// Verify registry has entry
	if comp, exists := registry["api.md"]; !exists || comp != "backend-dev" {
		t.Errorf("Registry should have api.md -> backend-dev, got: %v", comp)
	}

	// Process second agent (should skip due to conflict)
	ctx2 := PostprocessContext{
		ComponentType:   "agents",
		ComponentName:   "api-scaffold",
		FilesystemName:  "api-scaffold",
		Target:          "copilot",
		TargetDir:       targetDir,
		DestPath:        agent2Folder,
		DryRun:          false,
		Formatter:       formatter.New(),
		SymlinkRegistry: registry,
	}

	if err := p.Process(ctx2); err != nil {
		t.Fatalf("Process() for agent2 should not return error (non-fatal conflict), got: %v", err)
	}

	// Verify symlink still points to first agent
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Fatalf("Cannot read symlink: %v", err)
	}
	expectedTarget := filepath.Join("backend-dev", "api.md")
	if target != expectedTarget {
		t.Errorf("Symlink should still point to backend-dev/api.md, got: %s", target)
	}

	// Registry should still have backend-dev
	if comp := registry["api.md"]; comp != "backend-dev" {
		t.Errorf("Registry should still have api.md -> backend-dev, got: %v", comp)
	}
}

func TestAgentFlattenPostprocessor_Process_NoMarkdownFiles(t *testing.T) {
	// Create temp directory structure with no .md files
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, ".github")
	agentsDir := filepath.Join(targetDir, "agents")
	agentFolder := filepath.Join(agentsDir, "my-agent")

	if err := os.MkdirAll(agentFolder, 0755); err != nil {
		t.Fatalf("Failed to create agent folder: %v", err)
	}

	// Create non-markdown file
	if err := os.WriteFile(filepath.Join(agentFolder, "notes.txt"), []byte("notes"), 0644); err != nil {
		t.Fatalf("Failed to create notes.txt: %v", err)
	}

	p := NewAgentFlattenPostprocessor()
	ctx := PostprocessContext{
		ComponentType:   "agents",
		ComponentName:   "my-agent",
		FilesystemName:  "my-agent",
		Target:          "copilot",
		TargetDir:       targetDir,
		DestPath:        agentFolder,
		DryRun:          false,
		Formatter:       formatter.New(),
		SymlinkRegistry: make(map[string]string),
	}

	// Should not error
	if err := p.Process(ctx); err != nil {
		t.Errorf("Process() should handle no markdown files gracefully, got: %v", err)
	}

	// Verify no symlinks created
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		t.Fatalf("Cannot read agents dir: %v", err)
	}

	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 {
			t.Errorf("No symlinks should be created, found: %s", entry.Name())
		}
	}
}

func TestAgentFlattenPostprocessor_Cleanup_MultipleSymlinks(t *testing.T) {
	// Create temp directory structure
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, ".github")
	agentsDir := filepath.Join(targetDir, "agents")
	agentFolder := filepath.Join(agentsDir, "backend-dev")

	if err := os.MkdirAll(agentFolder, 0755); err != nil {
		t.Fatalf("Failed to create agent folder: %v", err)
	}

	// Create agent files
	agentFiles := []string{"api.md", "tdd.md", "temporal.md"}
	for _, filename := range agentFiles {
		if err := os.WriteFile(filepath.Join(agentFolder, filename), []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", filename, err)
		}
	}

	// Create symlinks manually
	for _, filename := range agentFiles {
		symlinkPath := filepath.Join(agentsDir, filename)
		relativeTarget := filepath.Join("backend-dev", filename)
		if err := os.Symlink(relativeTarget, symlinkPath); err != nil {
			t.Fatalf("Failed to create symlink %s: %v", filename, err)
		}
	}

	// Verify symlinks exist
	for _, filename := range agentFiles {
		symlinkPath := filepath.Join(agentsDir, filename)
		if _, err := os.Lstat(symlinkPath); err != nil {
			t.Fatalf("Symlink %s should exist before cleanup", filename)
		}
	}

	// Run cleanup
	p := NewAgentFlattenPostprocessor()
	ctx := PostprocessContext{
		ComponentType:  "agents",
		ComponentName:  "backend-dev",
		FilesystemName: "backend-dev",
		Target:         "copilot",
		TargetDir:      targetDir,
		DestPath:       agentFolder,
		DryRun:         false,
		Formatter:      formatter.New(),
	}

	if err := p.Cleanup(ctx); err != nil {
		t.Errorf("Cleanup() error = %v", err)
	}

	// Verify all symlinks removed
	for _, filename := range agentFiles {
		symlinkPath := filepath.Join(agentsDir, filename)
		if _, err := os.Lstat(symlinkPath); err == nil {
			t.Errorf("Symlink %s should be removed by cleanup", filename)
		}
	}
}
