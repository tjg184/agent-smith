package linker

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tjg184/agent-smith/pkg/paths"
)

// LinkStatus represents the status of a linked component
type LinkStatus struct {
	Name       string
	Type       string
	LinkType   string // "symlink", "copied", "broken", "missing"
	Target     string
	Valid      bool
	TargetPath string
	Profile    string // paths.BaseProfileName or profile name
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

// getProfileFromPath extracts the profile name from a component path
// Returns paths.BaseProfileName if the component is in the base installation, or the profile name
func getProfileFromPath(path string) string {
	// Clean the path first
	path = filepath.Clean(path)

	// Check if the path itself is a profile directory (e.g., ~/.agent-smith/profiles/work)
	// In this case, the parent directory is "profiles"
	parent := filepath.Dir(path)
	if filepath.Base(parent) == "profiles" {
		return filepath.Base(path)
	}

	// Walk up the directory tree to find "profiles" directory
	dir := parent
	for {
		grandparent := filepath.Dir(dir)
		if filepath.Base(grandparent) == "profiles" {
			// This is a profile directory
			return filepath.Base(dir)
		}
		if grandparent == dir || grandparent == "." || grandparent == "/" {
			// Reached root without finding "profiles"
			return paths.BaseProfileName
		}
		dir = grandparent
	}
}

// GetProfileNameFromSymlink extracts the profile name from a symlink's target path
// Returns the profile name if the symlink points to a profile, or paths.BaseProfileName if it points to base installation
// Returns empty string if the symlink is broken or invalid
func GetProfileNameFromSymlink(symlinkPath string) string {
	// Read the symlink target
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		return "" // Broken or not a symlink
	}

	// Resolve relative paths
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(symlinkPath), target)
	}

	// Use getProfileFromPath to extract profile name
	return getProfileFromPath(target)
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
