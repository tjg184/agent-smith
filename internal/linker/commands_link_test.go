package linker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLinkFlatMdFiles_SingleFile(t *testing.T) {
	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "commands", "commit")
	targetDir := filepath.Join(tmp, "opencode", "commands")
	must(t, os.MkdirAll(srcDir, 0755))
	must(t, os.WriteFile(filepath.Join(srcDir, "commit.md"), []byte("# commit"), 0644))

	linked, err := linkFlatMdFiles(srcDir, targetDir)
	if err != nil {
		t.Fatalf("linkFlatMdFiles: %v", err)
	}
	if len(linked) != 1 {
		t.Fatalf("expected 1 linked, got %d", len(linked))
	}

	dst := filepath.Join(targetDir, "commit.md")
	assertSymlinkPointsInto(t, dst, srcDir)
}

func TestLinkFlatMdFiles_NestedFile(t *testing.T) {
	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "commands", "commit")
	targetDir := filepath.Join(tmp, "opencode", "commands")
	must(t, os.MkdirAll(filepath.Join(srcDir, "sub"), 0755))
	must(t, os.WriteFile(filepath.Join(srcDir, "sub", "other.md"), []byte("# other"), 0644))

	linked, err := linkFlatMdFiles(srcDir, targetDir)
	if err != nil {
		t.Fatalf("linkFlatMdFiles: %v", err)
	}
	if len(linked) != 1 {
		t.Fatalf("expected 1 linked, got %d", len(linked))
	}

	// Nested sub/other.md must be flattened to sub-other.md at the target root
	dst := filepath.Join(targetDir, "sub-other.md")
	assertSymlinkPointsInto(t, dst, srcDir)
}

func TestLinkFlatMdFiles_MultipleFiles(t *testing.T) {
	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "agents", "ui-design")
	targetDir := filepath.Join(tmp, "opencode", "agents")
	must(t, os.MkdirAll(srcDir, 0755))
	must(t, os.WriteFile(filepath.Join(srcDir, "ui-designer.md"), []byte("# ui"), 0644))
	must(t, os.WriteFile(filepath.Join(srcDir, "accessibility.md"), []byte("# a11y"), 0644))
	must(t, os.WriteFile(filepath.Join(srcDir, "README.txt"), []byte("ignored"), 0644))

	linked, err := linkFlatMdFiles(srcDir, targetDir)
	if err != nil {
		t.Fatalf("linkFlatMdFiles: %v", err)
	}
	if len(linked) != 2 {
		t.Fatalf("expected 2 linked, got %d", len(linked))
	}
}

func TestUnlinkFlatMdFiles_RemovesLinks(t *testing.T) {
	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "commands", "commit")
	targetDir := filepath.Join(tmp, "opencode", "commands")
	must(t, os.MkdirAll(srcDir, 0755))
	must(t, os.WriteFile(filepath.Join(srcDir, "commit.md"), []byte("# commit"), 0644))

	if _, err := linkFlatMdFiles(srcDir, targetDir); err != nil {
		t.Fatalf("setup link: %v", err)
	}

	if err := unlinkFlatMdFiles("commit", filepath.Join(tmp, "commands"), targetDir); err != nil {
		t.Fatalf("unlinkFlatMdFiles: %v", err)
	}

	dst := filepath.Join(targetDir, "commit.md")
	if _, err := os.Lstat(dst); !os.IsNotExist(err) {
		t.Errorf("expected symlink to be removed, but it still exists")
	}
}

func TestUnlinkFlatMdFiles_PrunesEmptyDirs(t *testing.T) {
	tmp := t.TempDir()
	// With flat layout there are no subdirs to prune, but the function must still
	// remove the flat symlink and leave targetDir intact.
	srcDir := filepath.Join(tmp, "commands", "commit")
	targetDir := filepath.Join(tmp, "opencode", "commands")
	must(t, os.MkdirAll(filepath.Join(srcDir, "sub"), 0755))
	must(t, os.WriteFile(filepath.Join(srcDir, "sub", "other.md"), []byte("# other"), 0644))

	if _, err := linkFlatMdFiles(srcDir, targetDir); err != nil {
		t.Fatalf("setup link: %v", err)
	}

	// Flat symlink is at targetDir/sub-other.md
	flatSymlink := filepath.Join(targetDir, "sub-other.md")
	if _, err := os.Lstat(flatSymlink); err != nil {
		t.Fatalf("expected flat symlink to exist before unlink: %v", err)
	}

	if err := unlinkFlatMdFiles("commit", filepath.Join(tmp, "commands"), targetDir); err != nil {
		t.Fatalf("unlinkFlatMdFiles: %v", err)
	}

	if _, err := os.Lstat(flatSymlink); !os.IsNotExist(err) {
		t.Errorf("expected flat symlink to be removed")
	}
}

func TestUnlinkFlatMdFiles_LeavesNonEmptyDirs(t *testing.T) {
	tmp := t.TempDir()
	commitSrc := filepath.Join(tmp, "commands", "commit")
	otherSrc := filepath.Join(tmp, "commands", "other")
	targetDir := filepath.Join(tmp, "opencode", "commands")
	must(t, os.MkdirAll(commitSrc, 0755))
	must(t, os.MkdirAll(otherSrc, 0755))
	must(t, os.WriteFile(filepath.Join(commitSrc, "commit.md"), []byte("# commit"), 0644))
	must(t, os.WriteFile(filepath.Join(otherSrc, "other.md"), []byte("# other"), 0644))

	if _, err := linkFlatMdFiles(commitSrc, targetDir); err != nil {
		t.Fatalf("setup commit link: %v", err)
	}
	if _, err := linkFlatMdFiles(otherSrc, targetDir); err != nil {
		t.Fatalf("setup other link: %v", err)
	}

	if err := unlinkFlatMdFiles("commit", filepath.Join(tmp, "commands"), targetDir); err != nil {
		t.Fatalf("unlinkFlatMdFiles: %v", err)
	}

	// other.md should still exist
	otherDst := filepath.Join(targetDir, "other.md")
	if _, err := os.Lstat(otherDst); err != nil {
		t.Errorf("other.md should still exist: %v", err)
	}
	// targetDir itself should still exist
	if _, err := os.Lstat(targetDir); err != nil {
		t.Errorf("targetDir should still exist: %v", err)
	}
}

func TestIsFlatMdLinked_True(t *testing.T) {
	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "commands", "commit")
	targetDir := filepath.Join(tmp, "opencode", "commands")
	must(t, os.MkdirAll(srcDir, 0755))
	must(t, os.WriteFile(filepath.Join(srcDir, "commit.md"), []byte("# commit"), 0644))

	if _, err := linkFlatMdFiles(srcDir, targetDir); err != nil {
		t.Fatalf("setup link: %v", err)
	}

	if !isFlatMdLinked("commit", filepath.Join(tmp, "commands"), targetDir) {
		t.Error("expected isFlatMdLinked to return true")
	}
}

func TestIsFlatMdLinked_False(t *testing.T) {
	tmp := t.TempDir()
	targetDir := filepath.Join(tmp, "opencode", "commands")
	must(t, os.MkdirAll(targetDir, 0755))

	if isFlatMdLinked("commit", filepath.Join(tmp, "commands"), targetDir) {
		t.Error("expected isFlatMdLinked to return false")
	}
}

func TestIsFlatMdLinked_OtherComponentNotLinked(t *testing.T) {
	tmp := t.TempDir()
	commitSrc := filepath.Join(tmp, "commands", "commit")
	targetDir := filepath.Join(tmp, "opencode", "commands")
	must(t, os.MkdirAll(commitSrc, 0755))
	must(t, os.WriteFile(filepath.Join(commitSrc, "commit.md"), []byte("# commit"), 0644))

	if _, err := linkFlatMdFiles(commitSrc, targetDir); err != nil {
		t.Fatalf("setup link: %v", err)
	}

	if isFlatMdLinked("other", filepath.Join(tmp, "commands"), targetDir) {
		t.Error("expected isFlatMdLinked to return false for unlinked component")
	}
}

// assertSymlinkPointsInto checks that path is a symlink whose target is inside dir.
func assertSymlinkPointsInto(t *testing.T, path, dir string) {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("expected symlink at %s: %v", path, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%s is not a symlink", path)
	}
	target, err := os.Readlink(path)
	if err != nil {
		t.Fatalf("readlink %s: %v", path, err)
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(path), target)
	}
	target = filepath.Clean(target)
	dir = filepath.Clean(dir)
	if target != dir && len(target) <= len(dir) {
		t.Errorf("symlink %s points to %s, expected inside %s", path, target, dir)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
