package materializer

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyDirectory recursively copies a directory from src to dst
func CopyDirectory(src, dst string) error {
	// Get source directory info
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

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := CopyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Get source file info
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy file contents
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

	// Walk the directory tree
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, only hash files
		if info.IsDir() {
			return nil
		}

		// Get relative path for consistent hashing
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		// Write relative path to hash with null separator
		if _, err := hash.Write([]byte(relPath)); err != nil {
			return err
		}
		if _, err := hash.Write([]byte("\x00")); err != nil {
			return err
		}

		// Read and hash file contents
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := io.Copy(hash, file); err != nil {
			return err
		}

		// Add null separator between files
		if _, err := hash.Write([]byte("\x00")); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to calculate directory hash: %w", err)
	}

	// Return plain hex format (no sha256: prefix) to match metadata.ComputeLocalFolderHash
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
