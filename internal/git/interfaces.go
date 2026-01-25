package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Repository represents a git repository interface for testing
type Repository interface {
	// Head returns the reference where HEAD is pointing to
	Head() (*plumbing.Reference, error)
}

// Cloner provides an interface for cloning git repositories
type Cloner interface {
	// Clone clones a repository to the given path with specified options
	Clone(path string, isBare bool, options *git.CloneOptions) (Repository, error)

	// Open opens an existing repository at the given path
	Open(path string) (Repository, error)
}

// gitRepository wraps go-git Repository to implement our Repository interface
type gitRepository struct {
	repo *git.Repository
}

func (r *gitRepository) Head() (*plumbing.Reference, error) {
	return r.repo.Head()
}

// DefaultCloner is the default implementation using go-git
type DefaultCloner struct{}

func (c *DefaultCloner) Clone(path string, isBare bool, options *git.CloneOptions) (Repository, error) {
	repo, err := git.PlainClone(path, isBare, options)
	if err != nil {
		return nil, err
	}
	return &gitRepository{repo: repo}, nil
}

func (c *DefaultCloner) Open(path string) (Repository, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	return &gitRepository{repo: repo}, nil
}

// NewDefaultCloner creates a new default cloner
func NewDefaultCloner() Cloner {
	return &DefaultCloner{}
}
