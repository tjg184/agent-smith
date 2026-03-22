package linkutil

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/tjg184/agent-smith/pkg/paths"
)

func ProfileFromPath(path string) string {
	path = filepath.Clean(path)

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
			return paths.BaseProfileName
		}
		dir = grandparent
	}
}

func AnalyzeLinkStatus(path string) (linkType string, target string, valid bool) {
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

func IsFlatMdLinked(componentName, componentTypeDir, targetBaseDir string) bool {
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
