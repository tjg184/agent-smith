package linker

import (
	"os"
	"path/filepath"
	"strings"
)

// linkFlatMdFiles walks srcDir recursively and creates a flat symlink in targetBaseDir
// for each .md file. Nested paths are flattened by joining segments with "-" so that
// all symlinks land directly in targetBaseDir regardless of source depth.
//
// e.g. srcDir=".../commands/commit", targetBaseDir="~/.config/opencode/commands"
//
//	commit.md              → commands/commit.md
//	ui-design/designer.md  → commands/ui-design-designer.md
func linkFlatMdFiles(srcDir, targetBaseDir string) ([]string, error) {
	var linked []string

	if resolved, err := filepath.EvalSymlinks(srcDir); err == nil {
		srcDir = resolved
	}

	if err := os.MkdirAll(targetBaseDir, 0755); err != nil {
		return nil, err
	}

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

		flatName := strings.ReplaceAll(rel, string(filepath.Separator), "-")
		dst := filepath.Join(targetBaseDir, flatName)

		if _, err := os.Lstat(dst); err == nil {
			if err := os.Remove(dst); err != nil {
				return err
			}
		}

		dstDir := targetBaseDir
		if realDir, err := filepath.EvalSymlinks(dstDir); err == nil {
			dstDir = realDir
		}

		relSymlink, err := filepath.Rel(dstDir, path)
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
