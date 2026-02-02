package detector

import (
	"testing"
)

// TestShouldIgnorePath tests path ignore logic
func TestShouldIgnorePath(t *testing.T) {
	rd := NewRepositoryDetector()

	ignorePaths := []string{"node_modules", "dist", "build", ".git"}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Exact match",
			path:     "node_modules",
			expected: true,
		},
		{
			name:     "Path starts with ignored",
			path:     "node_modules/package",
			expected: true,
		},
		{
			name:     "Path contains ignored in middle",
			path:     "src/node_modules/package",
			expected: true,
		},
		{
			name:     "Path ends with ignored",
			path:     "src/build",
			expected: true,
		},
		{
			name:     "Path not ignored",
			path:     "src/components",
			expected: false,
		},
		{
			name:     "Partial match (should not ignore)",
			path:     "node_modules_backup",
			expected: false,
		},
		{
			name:     "Partial match in path (should not ignore)",
			path:     "src/node_modules_backup/file.js",
			expected: false,
		},
		{
			name:     "Hidden git directory",
			path:     ".git",
			expected: true,
		},
		{
			name:     "Git subdirectory",
			path:     ".git/config",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rd.ShouldIgnorePath(tt.path, ignorePaths)
			if result != tt.expected {
				t.Errorf("ShouldIgnorePath(%s) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

// TestMatchesExactFile tests exact file matching
func TestMatchesExactFile(t *testing.T) {
	rd := NewRepositoryDetector()

	exactFiles := []string{"skill.md", "README.md", "package.json"}

	tests := []struct {
		name     string
		fileName string
		expected bool
	}{
		{
			name:     "Match skill.md",
			fileName: "skill.md",
			expected: true,
		},
		{
			name:     "Match README.md",
			fileName: "README.md",
			expected: true,
		},
		{
			name:     "Match package.json",
			fileName: "package.json",
			expected: true,
		},
		{
			name:     "No match - different file",
			fileName: "index.js",
			expected: false,
		},
		{
			name:     "No match - case sensitive",
			fileName: "SKILL.MD",
			expected: false,
		},
		{
			name:     "No match - partial name",
			fileName: "skill.txt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rd.MatchesExactFile(tt.fileName, exactFiles)
			if result != tt.expected {
				t.Errorf("MatchesExactFile(%s) = %v, expected %v", tt.fileName, result, tt.expected)
			}
		})
	}
}

// TestMatchesPathPattern tests path pattern matching
func TestMatchesPathPattern(t *testing.T) {
	rd := NewRepositoryDetector()

	pathPatterns := []string{".opencode/agents", ".opencode/commands", "agents/"}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Match .opencode/agents",
			path:     ".opencode/agents/my-agent.md",
			expected: true,
		},
		{
			name:     "Match .opencode/commands",
			path:     ".opencode/commands/my-command.md",
			expected: true,
		},
		{
			name:     "Match agents/ suffix",
			path:     "src/agents/",
			expected: true,
		},
		{
			name:     "Match agents/ in path",
			path:     "src/agents/my-agent.md",
			expected: true,
		},
		{
			name:     "No match",
			path:     "src/components/button.js",
			expected: false,
		},
		{
			name:     "No match - similar but different",
			path:     ".opencode/skills/my-skill.md",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rd.MatchesPathPattern(tt.path, pathPatterns)
			if result != tt.expected {
				t.Errorf("MatchesPathPattern(%s) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

// TestMatchesFileExtension tests file extension matching
func TestMatchesFileExtension(t *testing.T) {
	rd := NewRepositoryDetector()

	extensions := []string{".md", ".js", ".ts", ".go"}

	tests := []struct {
		name     string
		fileName string
		expected bool
	}{
		{
			name:     "Match .md",
			fileName: "README.md",
			expected: true,
		},
		{
			name:     "Match .js",
			fileName: "index.js",
			expected: true,
		},
		{
			name:     "Match .ts",
			fileName: "app.ts",
			expected: true,
		},
		{
			name:     "Match .go",
			fileName: "main.go",
			expected: true,
		},
		{
			name:     "No match .txt",
			fileName: "notes.txt",
			expected: false,
		},
		{
			name:     "No match .py",
			fileName: "script.py",
			expected: false,
		},
		{
			name:     "No match - no extension",
			fileName: "justfile",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rd.MatchesFileExtension(tt.fileName, extensions)
			if result != tt.expected {
				t.Errorf("MatchesFileExtension(%s) = %v, expected %v", tt.fileName, result, tt.expected)
			}
		})
	}
}

// TestMatchesFileExtensionEdgeCases tests edge cases
func TestMatchesFileExtensionEdgeCases(t *testing.T) {
	rd := NewRepositoryDetector()

	tests := []struct {
		name       string
		fileName   string
		extensions []string
		expected   bool
	}{
		{
			name:       "Empty extensions list",
			fileName:   "file.txt",
			extensions: []string{},
			expected:   false,
		},
		{
			name:       "Empty filename",
			fileName:   "",
			extensions: []string{".txt"},
			expected:   false,
		},
		{
			name:       "Multiple dots in filename",
			fileName:   "file.test.js",
			extensions: []string{".js"},
			expected:   true,
		},
		{
			name:       "Dot at start of filename",
			fileName:   ".gitignore",
			extensions: []string{".gitignore"},
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rd.MatchesFileExtension(tt.fileName, tt.extensions)
			if result != tt.expected {
				t.Errorf("MatchesFileExtension(%s, %v) = %v, expected %v", tt.fileName, tt.extensions, result, tt.expected)
			}
		})
	}
}
