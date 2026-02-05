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
