package metadata

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// ComputeLocalFolderHash computes a SHA256 hash of a folder's contents.
// This is used for both sourceHash (at install time) and currentHash (for drift detection).
// The hash is stable across git operations and file system changes that only affect timestamps.
func ComputeLocalFolderHash(folderPath string) (string, error) {
	hasher := sha256.New()

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(folderPath, path)
		if err != nil {
			return err
		}

		// Write relative path to hash (for directory structure)
		hasher.Write([]byte(relPath))
		hasher.Write([]byte("\x00")) // null separator

		// Write file content
		// Note: We deliberately exclude ModTime to make hash stable across git operations
		// and file system operations that only change timestamps
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		hasher.Write(data)
		hasher.Write([]byte("\x00")) // null separator between files
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to compute folder hash: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// ComputeComponentHash computes a SHA256 hash for a component (file or directory).
// For directories (like skills), it hashes all files recursively.
// For single files (like agents/commands), it hashes just that file.
// This is used for detecting identical duplicates vs. actual conflicts.
func ComputeComponentHash(repoPath, componentPath string) (string, error) {
	fullPath := filepath.Join(repoPath, componentPath)

	info, err := os.Stat(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to stat component path %s: %w", fullPath, err)
	}

	// For directories, use the folder hash
	if info.IsDir() {
		return ComputeLocalFolderHash(fullPath)
	}

	// For single files, compute hash of the file content
	hasher := sha256.New()
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", fullPath, err)
	}

	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
