package detector

import (
	"path/filepath"
	"strings"
)

// ShouldIgnorePath checks if a path should be ignored during detection
func (rd *RepositoryDetector) ShouldIgnorePath(relPath string, ignorePaths []string) bool {
	// Normalize path to use forward slashes for consistent matching
	normalizedPath := filepath.ToSlash(relPath)

	for _, ignorePath := range ignorePaths {
		// Check if ignore pattern matches as a whole path component
		// Pattern matches if it appears as:
		// 1. Exact match: "build"
		// 2. At the start: "build/..."
		// 3. After a separator: ".../build/..."
		// 4. At the end: ".../build"

		if normalizedPath == ignorePath {
			return true // Exact match
		}

		if strings.HasPrefix(normalizedPath, ignorePath+"/") {
			return true // Pattern at start: "build/..."
		}

		if strings.Contains(normalizedPath, "/"+ignorePath+"/") {
			return true // Pattern in middle: ".../build/..."
		}

		if strings.HasSuffix(normalizedPath, "/"+ignorePath) {
			return true // Pattern at end: ".../build"
		}
	}
	return false
}

// MatchesExactFile checks if the filename matches any exact file patterns
func (rd *RepositoryDetector) MatchesExactFile(fileName string, exactFiles []string) bool {
	for _, exactFile := range exactFiles {
		if fileName == exactFile {
			return true
		}
	}
	return false
}

// MatchesPathPattern checks if the relative path matches any path patterns.
// For patterns with surrounding slashes (e.g., "/agents/"), matches directory boundaries:
// - containing the directory (e.g., "plugins/agents/test")
// - starting with the directory (e.g., "agents/test")
// - ending with the directory (e.g., ".opencode/agents")
// - equaling the directory exactly (e.g., "agents")
// This prevents false matches on names ending with the directory name (e.g., "dispatching-parallel-agents").
func (rd *RepositoryDetector) MatchesPathPattern(relPath string, pathPatterns []string) bool {
	for _, pattern := range pathPatterns {
		if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
			dirName := strings.Trim(pattern, "/")

			if strings.Contains(relPath, pattern) {
				return true
			}

			if strings.HasPrefix(relPath, dirName+"/") {
				return true
			}

			if strings.HasSuffix(relPath, "/"+dirName) {
				return true
			}

			if relPath == dirName {
				return true
			}
		} else {
			if strings.Contains(relPath, pattern) || strings.HasSuffix(relPath, pattern) {
				return true
			}
		}
	}
	return false
}

// MatchesFileExtension checks if the file has any of the specified extensions
func (rd *RepositoryDetector) MatchesFileExtension(fileName string, fileExtensions []string) bool {
	for _, ext := range fileExtensions {
		if strings.HasSuffix(fileName, ext) {
			return true
		}
	}
	return false
}
