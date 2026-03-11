package linker

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/tjg184/agent-smith/internal/fileutil"
)

// linkFlatMdFiles walks srcDir recursively and creates a symlink in targetBaseDir
// for each .md file, preserving relative path within srcDir.
//
// e.g. srcDir=".../commands/commit", targetBaseDir="~/.config/opencode/commands"
//
//	commit.md     → commands/commit.md
//	sub/other.md  → commands/sub/other.md
func linkFlatMdFiles(srcDir, targetBaseDir string) ([]string, error) {
	var linked []string

	err := filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		dst := filepath.Join(targetBaseDir, rel)

		if err := fileutil.CreateDirectoryWithPermissions(filepath.Dir(dst)); err != nil {
			return err
		}

		if _, err := os.Lstat(dst); err == nil {
			if err := os.Remove(dst); err != nil {
				return err
			}
		}

		relSymlink, err := filepath.Rel(filepath.Dir(dst), path)
		if err != nil {
			return err
		}

		if err := os.Symlink(relSymlink, dst); err != nil {
			return err
		}

		linked = append(linked, dst)
		return nil
	})

	return linked, err
}

// unlinkFlatMdFiles removes symlinks inside targetBaseDir pointing into
// componentTypeDir/<componentName>/, then prunes newly-empty subdirectories.
func unlinkFlatMdFiles(componentName, componentTypeDir, targetBaseDir string) error {
	componentRoot := filepath.Clean(filepath.Join(componentTypeDir, componentName))
	expectedPrefix := componentRoot + string(filepath.Separator)

	return filepath.WalkDir(targetBaseDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}

		info, err := os.Lstat(path)
		if err != nil || info.Mode()&os.ModeSymlink == 0 {
			return nil
		}

		target, err := os.Readlink(path)
		if err != nil {
			return nil
		}

		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(path), target)
		}
		target = filepath.Clean(target)

		pointsIntoComponent := target == componentRoot ||
			strings.HasPrefix(target, expectedPrefix)

		if pointsIntoComponent {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
			pruneEmptyDirs(filepath.Dir(path), targetBaseDir)
		}

		return nil
	})
}

func pruneEmptyDirs(dir, stopAt string) {
	stopAt = filepath.Clean(stopAt)
	for {
		dir = filepath.Clean(dir)
		if dir == stopAt || !strings.HasPrefix(dir, stopAt) {
			return
		}
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			return
		}
		if err := os.Remove(dir); err != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}

func isFlatMdLinked(componentName, componentTypeDir, targetBaseDir string) bool {
	componentRoot := filepath.Clean(filepath.Join(componentTypeDir, componentName))
	expectedPrefix := componentRoot + string(filepath.Separator)

	found := false
	_ = filepath.WalkDir(targetBaseDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || found {
			return nil
		}

		info, err := os.Lstat(path)
		if err != nil || info.Mode()&os.ModeSymlink == 0 {
			return nil
		}

		target, err := os.Readlink(path)
		if err != nil {
			return nil
		}

		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(path), target)
		}
		target = filepath.Clean(target)

		if target == componentRoot || strings.HasPrefix(target, expectedPrefix) {
			found = true
		}

		return nil
	})

	return found
}
