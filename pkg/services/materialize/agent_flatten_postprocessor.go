package materialize

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AgentFlattenPostprocessor creates flattened symlinks for agents on the copilot target.
// GitHub Copilot expects agents to be flat files (.github/agents/my-agent.md) rather than
// nested folders (.github/agents/my-agent/my-agent.md). This postprocessor scans agent
// folders for all .md files and creates individual symlinks for each, supporting both
// single-file and multi-file agent structures.
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

// findAgentMarkdownFiles scans a directory for agent .md files
// Returns absolute paths to all valid agent files
// Excludes: README.md, LICENSE.md, DOCS.md, CHANGELOG.md (case-insensitive)
// Only processes top-level files (does not recurse into subdirectories)
func findAgentMarkdownFiles(dir string) ([]string, error) {
	var mdFiles []string
	ignorePatterns := []string{"README.md", "LICENSE.md", "DOCS.md", "CHANGELOG.md"}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue // Don't recurse into subdirectories
		}

		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".md") {
			continue // Skip non-markdown files
		}

		// Check if file should be ignored (case-insensitive)
		shouldIgnore := false
		for _, pattern := range ignorePatterns {
			if strings.EqualFold(name, pattern) {
				shouldIgnore = true
				break
			}
		}

		if !shouldIgnore {
			mdFiles = append(mdFiles, filepath.Join(dir, name))
		}
	}

	return mdFiles, nil
}

// Process creates flattened symlinks for all agent files in the folder
// Supports both single-file (my-agent/my-agent.md) and multi-file (backend-dev/*.md) patterns
func (p *AgentFlattenPostprocessor) Process(ctx PostprocessContext) error {
	// 1. Find all agent markdown files in the folder
	mdFiles, err := findAgentMarkdownFiles(ctx.DestPath)
	if err != nil {
		ctx.Formatter.WarningMsg("Could not scan agent folder: %v", err)
		return nil // Non-fatal
	}

	// 2. If no files found, log info and skip
	if len(mdFiles) == 0 {
		ctx.Formatter.Info("  No agent markdown files found in %s", ctx.ComponentName)
		return nil
	}

	// 3. Create symlink for each agent file
	agentsDir := filepath.Join(ctx.TargetDir, "agents")
	createdCount := 0
	skippedCount := 0

	for _, mdFile := range mdFiles {
		filename := filepath.Base(mdFile)
		symlinkPath := filepath.Join(agentsDir, filename)

		// Use FilesystemName (actual disk name) for the symlink target, not ComponentName
		// This handles cases where auto-suffixing occurred (e.g., accessibility-compliance-2)
		targetDir := ctx.FilesystemName
		if targetDir == "" {
			targetDir = ctx.ComponentName // Fallback for safety
		}
		relativeTarget := filepath.Join(targetDir, filename)

		// Check for name conflicts across components
		if ctx.SymlinkRegistry != nil {
			if existingComponent, exists := ctx.SymlinkRegistry[filename]; exists && existingComponent != ctx.ComponentName {
				ctx.Formatter.WarningMsg("⚠️  Name conflict: %s (from %s) conflicts with existing %s (from %s)",
					filename, ctx.ComponentName, filename, existingComponent)
				ctx.Formatter.WarningMsg("   Skipping symlink for %s → %s", filename, relativeTarget)
				skippedCount++
				continue
			}
		}

		// Check if symlink already exists with correct target (idempotent)
		if info, err := os.Lstat(symlinkPath); err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				// It's a symlink - check if it points to the correct location
				if target, err := os.Readlink(symlinkPath); err == nil && target == relativeTarget {
					// Correct symlink already exists - register and continue
					if ctx.SymlinkRegistry != nil {
						ctx.SymlinkRegistry[filename] = ctx.ComponentName
					}
					continue
				}

				// Symlink exists but points to wrong location - remove and recreate
				if !ctx.DryRun {
					if err := os.Remove(symlinkPath); err != nil {
						ctx.Formatter.WarningMsg("Could not remove incorrect symlink %s: %v", filename, err)
						skippedCount++
						continue
					}
				}
			} else {
				// Regular file exists where symlink should be - fatal conflict
				return fmt.Errorf("cannot create symlink: regular file exists at %s", symlinkPath)
			}
		}

		// Create the symlink
		if ctx.DryRun {
			ctx.Formatter.Info("  Would create flat symlink: %s → %s",
				filename, relativeTarget)
			createdCount++
		} else {
			if err := os.Symlink(relativeTarget, symlinkPath); err != nil {
				// Non-fatal: log warning but don't fail
				ctx.Formatter.WarningMsg("Could not create symlink %s: %v", filename, err)
				skippedCount++
				continue
			}
			ctx.Formatter.Info("  Created flat symlink: %s → %s",
				filename, relativeTarget)
			createdCount++
		}

		// Register symlink in registry to detect future conflicts
		if ctx.SymlinkRegistry != nil {
			ctx.SymlinkRegistry[filename] = ctx.ComponentName
		}
	}

	// Log summary if multiple files were processed
	if len(mdFiles) > 1 {
		if ctx.DryRun {
			ctx.Formatter.Info("  Would create %d symlink(s) for %s", createdCount, ctx.ComponentName)
		} else {
			ctx.Formatter.Info("  Created %d symlink(s) for %s", createdCount, ctx.ComponentName)
		}
		if skippedCount > 0 {
			ctx.Formatter.WarningMsg("  Skipped %d symlink(s) due to conflicts", skippedCount)
		}
	}

	return nil
}

// Cleanup removes all flattened symlinks for this component before re-materialization
// Scans the agents directory for symlinks pointing to files within the component folder
func (p *AgentFlattenPostprocessor) Cleanup(ctx PostprocessContext) error {
	agentsDir := filepath.Join(ctx.TargetDir, "agents")

	// Read all entries in agents directory
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil // Non-fatal: directory might not exist
	}

	removedCount := 0

	for _, entry := range entries {
		// Skip non-symlinks
		if entry.Type()&os.ModeSymlink == 0 {
			continue
		}

		symlinkPath := filepath.Join(agentsDir, entry.Name())

		// Read symlink target
		target, err := os.Readlink(symlinkPath)
		if err != nil {
			continue // Skip if we can't read it
		}

		// Check if symlink points to a file in our component folder
		// Target format: "filesystemName/filename.md" (use FilesystemName to handle auto-suffixing)
		targetDir := ctx.FilesystemName
		if targetDir == "" {
			targetDir = ctx.ComponentName // Fallback for safety
		}
		if strings.HasPrefix(target, targetDir+"/") {
			if !ctx.DryRun {
				if err := os.Remove(symlinkPath); err != nil {
					ctx.Formatter.WarningMsg("Could not remove symlink %s during cleanup: %v", entry.Name(), err)
				} else {
					removedCount++
				}
			}
		}
	}

	if removedCount > 0 && !ctx.DryRun {
		ctx.Formatter.Info("  Cleaned up %d symlink(s) for %s", removedCount, ctx.ComponentName)
	}

	return nil // Always non-fatal
}
