package uninstaller

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tgaines/agent-smith/internal/linker"
	"github.com/tgaines/agent-smith/internal/metadata"
)

// Uninstaller handles component removal
type Uninstaller struct {
	baseDir string
	linker  *linker.ComponentLinker
}

// NewUninstaller creates a new Uninstaller instance
func NewUninstaller(baseDir string, componentLinker *linker.ComponentLinker) *Uninstaller {
	return &Uninstaller{
		baseDir: baseDir,
		linker:  componentLinker,
	}
}

// UninstallComponent removes a single component
func (u *Uninstaller) UninstallComponent(componentType, name string) error {
	// Validate component type
	if componentType != "skills" && componentType != "agents" && componentType != "commands" {
		return fmt.Errorf("invalid component type: %s (must be skills, agents, or commands)", componentType)
	}

	// Check if component exists in lock file
	_, err := metadata.LoadLockFileEntry(u.baseDir, componentType, name)
	if err != nil {
		return fmt.Errorf("component '%s' not installed", name)
	}

	// Auto-unlink component from all targets (silent if not linked)
	if u.linker != nil {
		// Try to unlink, but don't fail if it's not linked
		_ = u.linker.UnlinkComponent(componentType, name)
	}

	// Remove component directory from filesystem
	componentDir := filepath.Join(u.baseDir, componentType, name)
	if err := os.RemoveAll(componentDir); err != nil {
		return fmt.Errorf("failed to remove component directory: %w", err)
	}

	// Remove entry from lock file
	if err := metadata.RemoveLockFileEntry(u.baseDir, componentType, name); err != nil {
		// Log warning but continue - directory is already removed
		fmt.Printf("Warning: Could not update lock file: %v\n", err)
	}

	// Display success message
	fmt.Printf("✓ Removed %s: %s\n", componentType, name)

	return nil
}
