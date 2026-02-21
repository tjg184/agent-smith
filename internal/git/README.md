# Internal Git Package

This package provides isolated git operations with clean interfaces for testing.

## Overview

The `internal/git` package abstracts git operations behind mockable interfaces, making it easy to test code that depends on git without requiring actual repositories.

## Files

- **interfaces.go**: Core interfaces (`Repository`, `Cloner`) and default implementation
- **clone.go**: Repository cloning operations with helper functions
- **info.go**: Repository information extraction (commit hashes, etc.)
- **git_test.go**: Tests and mock implementations

## Key Interfaces

### Repository

Represents a git repository with minimal interface:

```go
type Repository interface {
    Head() (*plumbing.Reference, error)
}
```

### Cloner

Provides methods for cloning and opening repositories:

```go
type Cloner interface {
    Clone(path string, isBare bool, options *git.CloneOptions) (Repository, error)
    Open(path string) (Repository, error)
}
```

## Usage

### Basic Usage

```go
import gitpkg "github.com/tjg184/agent-smith/internal/git"

// Create a cloner
cloner := gitpkg.NewDefaultCloner()

// Clone a repository (shallow)
repo, err := gitpkg.CloneShallow(cloner, "/tmp/repo", "https://github.com/user/repo.git")
if err != nil {
    log.Fatal(err)
}

// Get commit hash
hash, err := gitpkg.GetCommitHash(repo)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Commit:", hash)
```

### Testing with Mocks

```go
import gitpkg "github.com/tjg184/agent-smith/internal/git"

// Create a mock cloner
mockCloner := &gitpkg.MockCloner{
    CloneFunc: func(path string, isBare bool, options *git.CloneOptions) (gitpkg.Repository, error) {
        return &gitpkg.MockRepository{
            HeadFunc: func() (*plumbing.Reference, error) {
                return plumbing.NewHashReference(plumbing.HEAD, 
                    plumbing.NewHash("abc123...")), nil
            },
        }, nil
    },
}

// Use mock in tests
repo, _ := gitpkg.CloneShallow(mockCloner, "/tmp/test", "https://test.com/repo.git")
hash, _ := gitpkg.GetCommitHash(repo)
```

## Helper Functions

### CloneShallow

Performs a shallow clone (depth 1) for quick repository access:

```go
repo, err := gitpkg.CloneShallow(cloner, "/tmp/repo", "https://github.com/user/repo.git")
```

### CloneBareShallow

Performs a shallow bare clone for metadata-only access:

```go
repo, err := gitpkg.CloneBareShallow(cloner, "/tmp/repo.git", "https://github.com/user/repo.git")
```

### OpenRepository

Opens an existing repository:

```go
repo, err := gitpkg.OpenRepository(cloner, "/path/to/repo")
```

### GetCommitHash

Extracts the current commit hash from a repository:

```go
hash, err := gitpkg.GetCommitHash(repo)
```

### GetCommitHashFromPath

Opens a repository and gets its commit hash in one call:

```go
hash, err := gitpkg.GetCommitHashFromPath(cloner, "/path/to/repo")
```

## Benefits

1. **Testability**: All git operations can be mocked for unit testing
2. **Isolation**: Git logic is separated from business logic
3. **Maintainability**: Centralized git operations make updates easier
4. **Consistency**: Helper functions ensure consistent clone options across codebase

## Migration Notes

When migrating existing code to use this package:

1. Replace `git.PlainClone()` with `gitpkg.CloneShallow()` or `gitpkg.CloneBareShallow()`
2. Replace `git.PlainOpen()` with `gitpkg.OpenRepository()`
3. Add a `cloner` field to structs that perform git operations
4. Use `gitpkg.GetCommitHash()` instead of calling `repo.Head()` directly
5. Inject mocks in tests for full test isolation
