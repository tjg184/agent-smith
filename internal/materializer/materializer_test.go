package materializer

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("writeFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("readFile %s: %v", path, err)
	}
	return string(data)
}

// CopyDirectory

func TestCopyDirectory_CopiesFilesAndSubdirs(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, "file.txt"), "hello")
	writeFile(t, filepath.Join(src, "sub", "nested.txt"), "world")

	if err := CopyDirectory(src, dst); err != nil {
		t.Fatalf("CopyDirectory error: %v", err)
	}

	if readFile(t, filepath.Join(dst, "file.txt")) != "hello" {
		t.Error("expected file.txt content 'hello'")
	}
	if readFile(t, filepath.Join(dst, "sub", "nested.txt")) != "world" {
		t.Error("expected sub/nested.txt content 'world'")
	}
}

func TestCopyDirectory_SourceNotExist(t *testing.T) {
	dst := t.TempDir()
	err := CopyDirectory("/nonexistent/path/xyz", dst)
	if err == nil {
		t.Fatal("expected error for non-existent source, got nil")
	}
}

func TestCopyDirectory_EmptyDirectory(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	if err := CopyDirectory(src, dst); err != nil {
		t.Fatalf("CopyDirectory on empty dir returned error: %v", err)
	}
}

// CopyFlatMdFiles

func TestCopyFlatMdFiles_CopiesNonIgnoredMdFiles(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, "my-skill.md"), "skill content")
	writeFile(t, filepath.Join(src, "README.md"), "readme")
	writeFile(t, filepath.Join(src, "LICENSE.md"), "license")
	writeFile(t, filepath.Join(src, "CHANGELOG.md"), "changelog")
	writeFile(t, filepath.Join(src, "DOCS.md"), "docs")
	writeFile(t, filepath.Join(src, "not-md.txt"), "ignored")

	if err := CopyFlatMdFiles(src, dst); err != nil {
		t.Fatalf("CopyFlatMdFiles error: %v", err)
	}

	if readFile(t, filepath.Join(dst, "my-skill.md")) != "skill content" {
		t.Error("expected my-skill.md to be copied with correct content")
	}

	for _, ignored := range []string{"README.md", "LICENSE.md", "CHANGELOG.md", "DOCS.md", "not-md.txt"} {
		if _, err := os.Stat(filepath.Join(dst, ignored)); !os.IsNotExist(err) {
			t.Errorf("expected %s to NOT be copied, but it exists in dest", ignored)
		}
	}
}

func TestCopyFlatMdFiles_NoMdFiles(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	writeFile(t, filepath.Join(src, "script.sh"), "#!/bin/bash")

	if err := CopyFlatMdFiles(src, dst); err != nil {
		t.Fatalf("CopyFlatMdFiles error: %v", err)
	}
}

func TestCopyFlatMdFiles_NestedMdFilesFlattened(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, "ui-design", "ui-designer.md"), "designer content")
	writeFile(t, filepath.Join(src, "top-level.md"), "top content")

	if err := CopyFlatMdFiles(src, dst); err != nil {
		t.Fatalf("CopyFlatMdFiles error: %v", err)
	}

	if readFile(t, filepath.Join(dst, "ui-design-ui-designer.md")) != "designer content" {
		t.Error("expected nested file to be flattened to ui-design-ui-designer.md")
	}
	if readFile(t, filepath.Join(dst, "top-level.md")) != "top content" {
		t.Error("expected top-level.md to be copied unchanged")
	}

	if _, err := os.Stat(filepath.Join(dst, "ui-design")); !os.IsNotExist(err) {
		t.Error("subdirectory ui-design should not exist in flat dest")
	}
}

// FlatMdFilesMatch

func TestFlatMdFilesMatch_IdenticalContent(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, "skill.md"), "content")
	writeFile(t, filepath.Join(dst, "skill.md"), "content")

	match, err := FlatMdFilesMatch(src, dst)
	if err != nil {
		t.Fatalf("FlatMdFilesMatch error: %v", err)
	}
	if !match {
		t.Error("expected true for identical content, got false")
	}
}

func TestFlatMdFilesMatch_DifferentContent(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, "skill.md"), "original")
	writeFile(t, filepath.Join(dst, "skill.md"), "modified")

	match, err := FlatMdFilesMatch(src, dst)
	if err != nil {
		t.Fatalf("FlatMdFilesMatch error: %v", err)
	}
	if match {
		t.Error("expected false for different content, got true")
	}
}

func TestFlatMdFilesMatch_MissingDestFile(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, "skill.md"), "content")

	match, err := FlatMdFilesMatch(src, dst)
	if err != nil {
		t.Fatalf("FlatMdFilesMatch returned unexpected error for missing dest: %v", err)
	}
	if match {
		t.Error("expected false when dest file is missing, got true")
	}
}

func TestFlatMdFilesMatch_NoMdFilesInSrc(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	match, err := FlatMdFilesMatch(src, dst)
	if err != nil {
		t.Fatalf("FlatMdFilesMatch error: %v", err)
	}
	if match {
		t.Error("expected false when src has no .md files, got true")
	}
}

func TestFlatMdFilesMatch_NestedSourceMatchesFlatDest(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, "ui-design", "ui-designer.md"), "designer content")
	writeFile(t, filepath.Join(dst, "ui-design-ui-designer.md"), "designer content")

	match, err := FlatMdFilesMatch(src, dst)
	if err != nil {
		t.Fatalf("FlatMdFilesMatch error: %v", err)
	}
	if !match {
		t.Error("expected true when nested src matches flat dest, got false")
	}
}

func TestFlatMdFilesMatch_NestedSourceMismatchesFlatDest(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, "ui-design", "ui-designer.md"), "original")
	writeFile(t, filepath.Join(dst, "ui-design-ui-designer.md"), "modified")

	match, err := FlatMdFilesMatch(src, dst)
	if err != nil {
		t.Fatalf("FlatMdFilesMatch error: %v", err)
	}
	if match {
		t.Error("expected false for differing content, got true")
	}
}

// FlatMdFilesAreRegular

func TestFlatMdFilesAreRegular_RegularFiles(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, "tdd-red.md"), "content")
	writeFile(t, filepath.Join(dst, "tdd-red.md"), "content")

	if !FlatMdFilesAreRegular(src, dst) {
		t.Error("expected true when dest files are regular files")
	}
}

func TestFlatMdFilesAreRegular_Symlinks(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, "tdd-red.md"), "content")
	// Place a symlink at the flat dest path instead of a real file
	if err := os.Symlink(filepath.Join(src, "tdd-red.md"), filepath.Join(dst, "tdd-red.md")); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	if FlatMdFilesAreRegular(src, dst) {
		t.Error("expected false when dest file is a symlink")
	}
}

func TestFlatMdFilesAreRegular_MissingFile(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, "tdd-red.md"), "content")
	// dst has no corresponding file

	if FlatMdFilesAreRegular(src, dst) {
		t.Error("expected false when dest file is missing")
	}
}

// RemoveFlatMdFiles

func TestRemoveFlatMdFiles_RemovesExpectedFiles(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, "skill.md"), "content")
	writeFile(t, filepath.Join(dst, "skill.md"), "content")

	if err := RemoveFlatMdFiles(src, dst); err != nil {
		t.Fatalf("RemoveFlatMdFiles error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dst, "skill.md")); !os.IsNotExist(err) {
		t.Error("expected skill.md to be removed from dst")
	}
}

func TestRemoveFlatMdFiles_NestedSourceRemovesFlatName(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, "ui-design", "ui-designer.md"), "content")
	writeFile(t, filepath.Join(dst, "ui-design-ui-designer.md"), "content")

	if err := RemoveFlatMdFiles(src, dst); err != nil {
		t.Fatalf("RemoveFlatMdFiles error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dst, "ui-design-ui-designer.md")); !os.IsNotExist(err) {
		t.Error("expected ui-design-ui-designer.md to be removed from dst")
	}
}

func TestRemoveFlatMdFiles_MissingDestFileIsNoop(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, "skill.md"), "content")

	if err := RemoveFlatMdFiles(src, dst); err != nil {
		t.Fatalf("RemoveFlatMdFiles should not error on missing dest file, got: %v", err)
	}
}

// DirectoriesMatch

func TestDirectoriesMatch_IdenticalDirectories(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	writeFile(t, filepath.Join(dir1, "a.txt"), "same")
	writeFile(t, filepath.Join(dir2, "a.txt"), "same")

	match, err := DirectoriesMatch(dir1, dir2)
	if err != nil {
		t.Fatalf("DirectoriesMatch error: %v", err)
	}
	if !match {
		t.Error("expected true for identical directories, got false")
	}
}

func TestDirectoriesMatch_DifferentContent(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	writeFile(t, filepath.Join(dir1, "a.txt"), "version1")
	writeFile(t, filepath.Join(dir2, "a.txt"), "version2")

	match, err := DirectoriesMatch(dir1, dir2)
	if err != nil {
		t.Fatalf("DirectoriesMatch error: %v", err)
	}
	if match {
		t.Error("expected false for different content, got true")
	}
}

func TestDirectoriesMatch_ExtraFileInDir2(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	writeFile(t, filepath.Join(dir1, "a.txt"), "same")
	writeFile(t, filepath.Join(dir2, "a.txt"), "same")
	writeFile(t, filepath.Join(dir2, "extra.txt"), "extra")

	match, err := DirectoriesMatch(dir1, dir2)
	if err != nil {
		t.Fatalf("DirectoriesMatch error: %v", err)
	}
	if match {
		t.Error("expected false when dir2 has extra file, got true")
	}
}

// CalculateDirectoryHash

func TestCalculateDirectoryHash_ConsistentForSameContent(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	writeFile(t, filepath.Join(dir1, "file.txt"), "consistent")
	writeFile(t, filepath.Join(dir2, "file.txt"), "consistent")

	hash1, err := CalculateDirectoryHash(dir1)
	if err != nil {
		t.Fatalf("hash dir1 error: %v", err)
	}
	hash2, err := CalculateDirectoryHash(dir2)
	if err != nil {
		t.Fatalf("hash dir2 error: %v", err)
	}
	if hash1 != hash2 {
		t.Errorf("expected identical hashes for same content, got %s vs %s", hash1, hash2)
	}
}

func TestCalculateDirectoryHash_DifferentForDifferentContent(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	writeFile(t, filepath.Join(dir1, "file.txt"), "version-a")
	writeFile(t, filepath.Join(dir2, "file.txt"), "version-b")

	hash1, err := CalculateDirectoryHash(dir1)
	if err != nil {
		t.Fatalf("hash dir1 error: %v", err)
	}
	hash2, err := CalculateDirectoryHash(dir2)
	if err != nil {
		t.Fatalf("hash dir2 error: %v", err)
	}
	if hash1 == hash2 {
		t.Error("expected different hashes for different content, got identical hashes")
	}
}
