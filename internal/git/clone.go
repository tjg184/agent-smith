package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// CloneOptions provides options for cloning repositories
type CloneOptions struct {
	URL           string
	Depth         int
	ReferenceName plumbing.ReferenceName
	SingleBranch  bool
}

// ToGoGitCloneOptions converts our CloneOptions to go-git CloneOptions
// It automatically determines the appropriate authentication method based on the URL
func (opts *CloneOptions) ToGoGitCloneOptions() *git.CloneOptions {
	cloneOpts := &git.CloneOptions{
		URL:           opts.URL,
		Depth:         opts.Depth,
		ReferenceName: opts.ReferenceName,
		SingleBranch:  opts.SingleBranch,
	}

	// Attempt to get auth method for the URL
	// Errors are ignored as some repos are public and don't need auth
	if auth, _ := GetAuthMethod(opts.URL); auth != nil {
		cloneOpts.Auth = auth
	}

	return cloneOpts
}

// CloneRepository clones a repository to the specified path
func CloneRepository(cloner Cloner, path string, isBare bool, opts *CloneOptions) (Repository, error) {
	return cloner.Clone(path, isBare, opts.ToGoGitCloneOptions())
}

// CloneShallow performs a shallow clone (depth 1) of a repository
func CloneShallow(cloner Cloner, path string, url string) (Repository, error) {
	opts := &CloneOptions{
		URL:           url,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	}
	return CloneRepository(cloner, path, false, opts)
}

// CloneBareShallow performs a shallow bare clone (depth 1) of a repository
func CloneBareShallow(cloner Cloner, path string, url string) (Repository, error) {
	opts := &CloneOptions{
		URL:           url,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	}
	return CloneRepository(cloner, path, true, opts)
}

// OpenRepository opens an existing repository at the given path
func OpenRepository(cloner Cloner, path string) (Repository, error) {
	return cloner.Open(path)
}
