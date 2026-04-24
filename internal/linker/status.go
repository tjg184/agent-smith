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
	Profile    string // profile name, or "" if not under a profile
}

// getSourceDescription returns a human-readable description of the source directory
func (cl *ComponentLinker) getSourceDescription() string {
	// Check if this is a profile directory
	if filepath.Base(filepath.Dir(cl.agentsDir)) == "profiles" {
		profileName := filepath.Base(cl.agentsDir)
		return fmt.Sprintf("Source: %s (profile '%s')", cl.agentsDir, profileName)
	}
	return fmt.Sprintf("Source: %s", cl.agentsDir)
}

func getProfileFromPath(path string) string {
	path = filepath.Clean(path)

	// Check if the path itself is a profile directory (e.g., ~/.agent-smith/profiles/work)
	// In this case, the parent directory is "profiles"
	parent := filepath.Dir(path)
	if filepath.Base(parent) == "profiles" {
		return filepath.Base(path)
	}

	dir := parent
	for {
		grandparent := filepath.Dir(dir)
		if filepath.Base(grandparent) == "profiles" {
			return filepath.Base(dir)
		}
		if grandparent == dir || grandparent == "." || grandparent == "/" {
			return ""
		}
		dir = grandparent
	}
}

func GetProfileNameFromSymlink(symlinkPath string) string {
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		return "" // Broken or not a symlink
	}

	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(symlinkPath), target)
	}

	return getProfileFromPath(target)
}

// analyzeLinkStatus analyzes the status of a link/directory
func (cl *ComponentLinker) analyzeLinkStatus(path string) (linkType string, target string, valid bool) {
	info, err := os.Lstat(path)
	if err != nil {
		return "missing", "", false
	}

	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return "broken", "", false
		}

		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(path), target)
		}

		if _, err := os.Stat(target); err == nil {
			return "symlink", target, true
		}
		return "broken", target, false
	}

	if info.IsDir() {
		return "copied", path, true
	}

	return "unknown", "", false
}
