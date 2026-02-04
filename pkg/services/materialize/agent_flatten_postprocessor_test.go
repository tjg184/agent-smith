package materialize

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tgaines/agent-smith/internal/formatter"
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
		ComponentType: "agents",
		ComponentName: "my-agent",
		Target:        "copilot",
		TargetDir:     targetDir,
		DestPath:      agentFolder,
		DryRun:        false,
		Formatter:     formatter.New(),
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
		ComponentType: "agents",
		ComponentName: "my-agent",
		Target:        "copilot",
		TargetDir:     targetDir,
		DestPath:      agentFolder,
		DryRun:        true,
		Formatter:     formatter.New(),
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
		ComponentType: "agents",
		ComponentName: "my-agent",
		Target:        "copilot",
		TargetDir:     targetDir,
		DestPath:      agentFolder,
		DryRun:        false,
		Formatter:     formatter.New(),
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
		ComponentType: "agents",
		ComponentName: "my-agent",
		Target:        "copilot",
		TargetDir:     targetDir,
		DestPath:      agentFolder,
		DryRun:        false,
		Formatter:     formatter.New(),
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
		ComponentType: "agents",
		ComponentName: "my-agent",
		Target:        "copilot",
		TargetDir:     targetDir,
		DestPath:      agentFolder,
		DryRun:        false,
		Formatter:     formatter.New(),
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
		ComponentType: "agents",
		ComponentName: "my-agent",
		Target:        "copilot",
		TargetDir:     targetDir,
		DestPath:      agentFolder,
		DryRun:        false,
		Formatter:     formatter.New(),
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
		ComponentType: "agents",
		ComponentName: "my-agent",
		Target:        "copilot",
		TargetDir:     targetDir,
		DestPath:      agentFolder,
		DryRun:        false,
		Formatter:     formatter.New(),
	}

	// Should not error when symlink doesn't exist
	if err := p.Cleanup(ctx); err != nil {
		t.Errorf("Cleanup() should handle missing symlink gracefully, got: %v", err)
	}
}
