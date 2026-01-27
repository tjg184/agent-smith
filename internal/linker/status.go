package linker

import (
	"fmt"
	"os"
	"path/filepath"
)

// LinkStatus represents the status of a linked component
type LinkStatus struct {
	Name       string
	Type       string
	LinkType   string // "symlink", "copied", "broken", "missing"
	Target     string
	Valid      bool
	TargetPath string
}

// getSourceDescription returns a human-readable description of the source directory
func (cl *ComponentLinker) getSourceDescription() string {
	// Check if this is a profile directory
	if filepath.Base(filepath.Dir(cl.agentsDir)) == "profiles" {
		profileName := filepath.Base(cl.agentsDir)
		return fmt.Sprintf("Source: %s (profile '%s')", cl.agentsDir, profileName)
	}
	return fmt.Sprintf("Source: %s (base installation)", cl.agentsDir)
}

// analyzeLinkStatus analyzes the status of a link/directory
func (cl *ComponentLinker) analyzeLinkStatus(path string) (linkType string, target string, valid bool) {
	info, err := os.Lstat(path)
	if err != nil {
		return "missing", "", false
	}

	// Check if it's a symlink
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return "broken", "", false
		}

		// Resolve relative paths
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(path), target)
		}

		// Check if target exists
		if _, err := os.Stat(target); err == nil {
			return "symlink", target, true
		}
		return "broken", target, false
	}

	// If it's a directory, it's a copied component
	if info.IsDir() {
		return "copied", path, true
	}

	return "unknown", "", false
}
