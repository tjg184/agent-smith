package materializer

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CopyDirectory recursively copies a directory from src to dst
func CopyDirectory(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source directory: %w", err)
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := CopyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

var mdIgnoreList = []string{"readme.md", "license.md", "docs.md", "changelog.md"}

func isMdIgnored(name string) bool {
	lower := strings.ToLower(name)
	if !strings.HasSuffix(lower, ".md") {
		return true
	}
	for _, ignored := range mdIgnoreList {
		if lower == ignored {
			return true
		}
	}
	return false
}

// collectMdFiles returns the top-level .md files from srcDir, excluding ignored names.
func collectMdFiles(srcDir string) ([]string, error) {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read source directory: %w", err)
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !isMdIgnored(name) {
			files = append(files, name)
		}
	}
	return files, nil
}

// CopyFlatMdFiles copies top-level .md files from srcDir directly into destDir (no wrapper subdir).
// Mirrors what linkFlatMdFiles does for symlinks. Ignored files: README.md, LICENSE.md, DOCS.md, CHANGELOG.md.
func CopyFlatMdFiles(srcDir, destDir string) error {
	names, err := collectMdFiles(srcDir)
	if err != nil {
		return err
	}
	for _, name := range names {
		if err := copyFile(filepath.Join(srcDir, name), filepath.Join(destDir, name)); err != nil {
			return fmt.Errorf("failed to copy %s: %w", name, err)
		}
	}
	return nil
}

// FlatMdFilesMatch returns true when every top-level .md file in srcDir exists in destDir with identical content.
func FlatMdFilesMatch(srcDir, destDir string) (bool, error) {
	names, err := collectMdFiles(srcDir)
	if err != nil {
		return false, err
	}
	for _, name := range names {
		srcHash, err := fileHash(filepath.Join(srcDir, name))
		if err != nil {
			return false, err
		}
		dstHash, err := fileHash(filepath.Join(destDir, name))
		if err != nil {
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, err
		}
		if srcHash != dstHash {
			return false, nil
		}
	}
	return len(names) > 0, nil
}

func fileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// RemoveFlatMdFiles removes top-level .md files from destDir that correspond to files in srcDir.
func RemoveFlatMdFiles(srcDir, destDir string) error {
	names, err := collectMdFiles(srcDir)
	if err != nil {
		return err
	}
	for _, name := range names {
		target := filepath.Join(destDir, name)
		if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove %s: %w", name, err)
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	return nil
}

// CalculateDirectoryHash calculates a SHA-256 hash of all files in a directory
// Files are sorted by path to ensure consistent hashing
// This uses the same algorithm as metadata.ComputeLocalFolderHash for consistency
func CalculateDirectoryHash(dirPath string) (string, error) {
	hash := sha256.New()

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		if _, err := hash.Write([]byte(relPath)); err != nil {
			return err
		}
		if _, err := hash.Write([]byte("\x00")); err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := io.Copy(hash, file); err != nil {
			return err
		}

		if _, err := hash.Write([]byte("\x00")); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to calculate directory hash: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// DirectoriesMatch checks if two directories have identical content by comparing their hashes
func DirectoriesMatch(dir1, dir2 string) (bool, error) {
	hash1, err := CalculateDirectoryHash(dir1)
	if err != nil {
		return false, err
	}

	hash2, err := CalculateDirectoryHash(dir2)
	if err != nil {
		return false, err
	}

	return hash1 == hash2, nil
}
