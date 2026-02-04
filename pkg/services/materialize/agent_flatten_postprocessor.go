package materialize

import (
	"fmt"
	"os"
	"path/filepath"
)

// AgentFlattenPostprocessor creates flattened symlinks for agents on the copilot target.
// GitHub Copilot expects agents to be flat files (.github/agents/my-agent.md) rather than
// nested folders (.github/agents/my-agent/my-agent.md). This postprocessor creates a
// symlink at the flat location pointing to the nested agent file, maintaining both
// structures for compatibility.
type AgentFlattenPostprocessor struct{}

// NewAgentFlattenPostprocessor creates a new agent flattening postprocessor
func NewAgentFlattenPostprocessor() *AgentFlattenPostprocessor {
	return &AgentFlattenPostprocessor{}
}

// Name returns the postprocessor name for logging
func (p *AgentFlattenPostprocessor) Name() string {
	return "AgentFlattenPostprocessor"
}

// ShouldProcess returns true only for agents being materialized to the copilot target
func (p *AgentFlattenPostprocessor) ShouldProcess(componentType, target string) bool {
	return componentType == "agents" && target == "copilot"
}

// Process creates a flattened symlink for the agent
// Structure: .github/agents/my-agent.md -> my-agent/my-agent.md (relative symlink)
func (p *AgentFlattenPostprocessor) Process(ctx PostprocessContext) error {
	// 1. Construct paths
	agentFile := filepath.Join(ctx.DestPath, ctx.ComponentName+".md")
	symlinkPath := filepath.Join(ctx.TargetDir, "agents", ctx.ComponentName+".md")
	relativeTarget := filepath.Join(ctx.ComponentName, ctx.ComponentName+".md")

	// 2. Verify agent file exists at the expected location
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		// Non-fatal: log warning and continue - agent may use different structure
		ctx.Formatter.WarningMsg("Agent file not found at expected location: %s", agentFile)
		ctx.Formatter.WarningMsg("Skipping flat symlink creation (agent will work in folder form)")
		return nil
	}

	// 3. Check if symlink already exists
	if info, err := os.Lstat(symlinkPath); err == nil {
		// Something exists at the symlink path
		if info.Mode()&os.ModeSymlink != 0 {
			// It's a symlink - check if it points to the correct location
			if target, err := os.Readlink(symlinkPath); err == nil && target == relativeTarget {
				// Correct symlink already exists - idempotent, nothing to do
				return nil
			}

			// Symlink exists but points to wrong location - remove and recreate
			if !ctx.DryRun {
				if err := os.Remove(symlinkPath); err != nil {
					ctx.Formatter.WarningMsg("Could not remove incorrect symlink: %v", err)
					return nil // Non-fatal
				}
			}
		} else {
			// Regular file exists where symlink should be - this is a conflict
			return fmt.Errorf("cannot create symlink: regular file exists at %s", symlinkPath)
		}
	}

	// 4. Create the symlink
	if ctx.DryRun {
		ctx.Formatter.Info("  Would create flat symlink: %s.md → %s/%s.md",
			ctx.ComponentName, ctx.ComponentName, ctx.ComponentName)
	} else {
		if err := os.Symlink(relativeTarget, symlinkPath); err != nil {
			// Non-fatal: log warning but don't fail - agent still works in folder form
			ctx.Formatter.WarningMsg("Could not create flattened symlink: %v", err)
			return nil
		}
		ctx.Formatter.Info("  Created flat symlink: %s.md → %s/%s.md",
			ctx.ComponentName, ctx.ComponentName, ctx.ComponentName)
	}

	return nil
}

// Cleanup removes the flattened symlink before re-materialization
// This is called when using --force to overwrite an existing component
func (p *AgentFlattenPostprocessor) Cleanup(ctx PostprocessContext) error {
	symlinkPath := filepath.Join(ctx.TargetDir, "agents", ctx.ComponentName+".md")

	// Check if symlink exists
	if info, err := os.Lstat(symlinkPath); err == nil {
		// Only remove if it's actually a symlink
		if info.Mode()&os.ModeSymlink != 0 {
			if !ctx.DryRun {
				if err := os.Remove(symlinkPath); err != nil {
					// Non-fatal: just log warning
					ctx.Formatter.WarningMsg("Could not remove symlink during cleanup: %v", err)
				}
			}
		}
	}

	// Always return nil - cleanup errors should never be fatal
	return nil
}
