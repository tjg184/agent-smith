package git

import (
	"errors"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// MockRepository is a mock implementation of Repository interface for testing
type MockRepository struct {
	HeadFunc func() (*plumbing.Reference, error)
}

func (m *MockRepository) Head() (*plumbing.Reference, error) {
	if m.HeadFunc != nil {
		return m.HeadFunc()
	}
	return nil, errors.New("not implemented")
}

// MockCloner is a mock implementation of Cloner interface for testing
type MockCloner struct {
	CloneFunc func(path string, isBare bool, options *git.CloneOptions) (Repository, error)
	OpenFunc  func(path string) (Repository, error)
}

func (m *MockCloner) Clone(path string, isBare bool, options *git.CloneOptions) (Repository, error) {
	if m.CloneFunc != nil {
		return m.CloneFunc(path, isBare, options)
	}
	return nil, errors.New("not implemented")
}

func (m *MockCloner) Open(path string) (Repository, error) {
	if m.OpenFunc != nil {
		return m.OpenFunc(path)
	}
	return nil, errors.New("not implemented")
}

// TestGetCommitHash tests the GetCommitHash function with a mock repository
func TestGetCommitHash(t *testing.T) {
	// Use a valid 40-character hash
	expectedHash := "abc123def4560000000000000000000000000000"
	mockRepo := &MockRepository{
		HeadFunc: func() (*plumbing.Reference, error) {
			return plumbing.NewHashReference(plumbing.HEAD, plumbing.NewHash(expectedHash)), nil
		},
	}

	hash, err := GetCommitHash(mockRepo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash != expectedHash {
		t.Errorf("expected hash %s, got %s", expectedHash, hash)
	}
}

// TestGetCommitHashError tests error handling in GetCommitHash
func TestGetCommitHashError(t *testing.T) {
	expectedError := errors.New("HEAD not found")
	mockRepo := &MockRepository{
		HeadFunc: func() (*plumbing.Reference, error) {
			return nil, expectedError
		},
	}

	_, err := GetCommitHash(mockRepo)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestCloneShallow tests the CloneShallow function with a mock cloner
func TestCloneShallow(t *testing.T) {
	expectedPath := "/tmp/test-repo"
	expectedURL := "https://github.com/test/repo.git"

	mockCloner := &MockCloner{
		CloneFunc: func(path string, isBare bool, options *git.CloneOptions) (Repository, error) {
			if path != expectedPath {
				t.Errorf("expected path %s, got %s", expectedPath, path)
			}
			if isBare {
				t.Error("expected isBare to be false")
			}
			if options.URL != expectedURL {
				t.Errorf("expected URL %s, got %s", expectedURL, options.URL)
			}
			if options.Depth != 1 {
				t.Errorf("expected depth 1, got %d", options.Depth)
			}
			if !options.SingleBranch {
				t.Error("expected SingleBranch to be true")
			}
			return &MockRepository{}, nil
		},
	}

	_, err := CloneShallow(mockCloner, expectedPath, expectedURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCloneBareShallow tests the CloneBareShallow function
func TestCloneBareShallow(t *testing.T) {
	expectedPath := "/tmp/test-repo"
	expectedURL := "https://github.com/test/repo.git"

	mockCloner := &MockCloner{
		CloneFunc: func(path string, isBare bool, options *git.CloneOptions) (Repository, error) {
			if !isBare {
				t.Error("expected isBare to be true")
			}
			if path != expectedPath {
				t.Errorf("expected path %s, got %s", expectedPath, path)
			}
			return &MockRepository{}, nil
		},
	}

	_, err := CloneBareShallow(mockCloner, expectedPath, expectedURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestOpenRepository tests the OpenRepository function
func TestOpenRepository(t *testing.T) {
	expectedPath := "/tmp/existing-repo"

	mockCloner := &MockCloner{
		OpenFunc: func(path string) (Repository, error) {
			if path != expectedPath {
				t.Errorf("expected path %s, got %s", expectedPath, path)
			}
			return &MockRepository{}, nil
		},
	}

	_, err := OpenRepository(mockCloner, expectedPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
