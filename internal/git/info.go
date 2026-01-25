package git

import (
	"fmt"
)

// GetCommitHash retrieves the current HEAD commit hash from a repository
func GetCommitHash(repo Repository) (string, error) {
	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}
	return ref.Hash().String(), nil
}

// GetCommitHashFromPath opens a repository and retrieves its HEAD commit hash
func GetCommitHashFromPath(cloner Cloner, path string) (string, error) {
	repo, err := OpenRepository(cloner, path)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}
	return GetCommitHash(repo)
}
