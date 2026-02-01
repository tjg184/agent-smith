package git

import (
	"testing"
)

// TestNormalizeURL tests URL normalization for various Git URL formats
func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		expected  string
		shouldErr bool
	}{
		// GitHub shorthand notation
		{
			name:      "GitHub shorthand",
			url:       "owner/repo",
			expected:  "https://github.com/owner/repo",
			shouldErr: false,
		},
		{
			name:      "GitHub shorthand with special chars",
			url:       "my-org/my-repo",
			expected:  "https://github.com/my-org/my-repo",
			shouldErr: false,
		},

		// HTTPS URLs
		{
			name:      "Full GitHub HTTPS URL",
			url:       "https://github.com/owner/repo",
			expected:  "https://github.com/owner/repo",
			shouldErr: false,
		},
		{
			name:      "GitLab HTTPS URL",
			url:       "https://gitlab.com/owner/repo",
			expected:  "https://gitlab.com/owner/repo",
			shouldErr: false,
		},
		{
			name:      "Bitbucket HTTPS URL",
			url:       "https://bitbucket.org/owner/repo",
			expected:  "https://bitbucket.org/owner/repo",
			shouldErr: false,
		},

		// HTTPS URLs with .git suffix
		{
			name:      "HTTPS URL with .git suffix",
			url:       "https://github.com/owner/repo.git",
			expected:  "https://github.com/owner/repo",
			shouldErr: false,
		},

		// HTTPS URLs with trailing slash
		{
			name:      "HTTPS URL with trailing slash",
			url:       "https://github.com/owner/repo/",
			expected:  "https://github.com/owner/repo",
			shouldErr: false,
		},
		{
			name:      "HTTPS URL with .git and trailing slash",
			url:       "https://github.com/owner/repo.git/",
			expected:  "https://github.com/owner/repo",
			shouldErr: false,
		},

		// HTTP URLs (should be upgraded to HTTPS)
		{
			name:      "HTTP converts to HTTPS",
			url:       "http://github.com/owner/repo",
			expected:  "https://github.com/owner/repo",
			shouldErr: false,
		},
		{
			name:      "HTTP with .git converts to HTTPS",
			url:       "http://github.com/owner/repo.git",
			expected:  "https://github.com/owner/repo",
			shouldErr: false,
		},

		// SSH URLs (git@host:path format)
		{
			name:      "GitHub SSH URL",
			url:       "git@github.com:owner/repo.git",
			expected:  "https://github.com/owner/repo",
			shouldErr: false,
		},
		{
			name:      "GitHub SSH URL without .git",
			url:       "git@github.com:owner/repo",
			expected:  "https://github.com/owner/repo",
			shouldErr: false,
		},
		{
			name:      "GitLab SSH URL",
			url:       "git@gitlab.com:owner/repo.git",
			expected:  "https://gitlab.com/owner/repo",
			shouldErr: false,
		},
		{
			name:      "Bitbucket SSH URL",
			url:       "git@bitbucket.org:owner/repo.git",
			expected:  "https://bitbucket.org/owner/repo",
			shouldErr: false,
		},

		// SSH URLs (ssh://git@host/path format)
		{
			name:      "GitHub SSH protocol URL",
			url:       "ssh://git@github.com/owner/repo.git",
			expected:  "https://github.com/owner/repo",
			shouldErr: false,
		},
		{
			name:      "GitHub SSH protocol URL without .git",
			url:       "ssh://git@github.com/owner/repo",
			expected:  "https://github.com/owner/repo",
			shouldErr: false,
		},
		{
			name:      "GitLab SSH protocol URL",
			url:       "ssh://git@gitlab.com/owner/repo.git",
			expected:  "https://gitlab.com/owner/repo",
			shouldErr: false,
		},

		// Case insensitive handling
		{
			name:      "Uppercase protocol and domain",
			url:       "HTTPS://GITHUB.COM/owner/repo",
			expected:  "https://github.com/owner/repo",
			shouldErr: false,
		},
		{
			name:      "Mixed case protocol and domain",
			url:       "https://GitHub.Com/owner/repo",
			expected:  "https://github.com/owner/repo",
			shouldErr: false,
		},
		{
			name:      "Uppercase HTTP to HTTPS",
			url:       "HTTP://GITHUB.COM/owner/repo",
			expected:  "https://github.com/owner/repo",
			shouldErr: false,
		},
		{
			name:      "Uppercase SSH format",
			url:       "GIT@github.com:owner/repo.git",
			expected:  "https://github.com/owner/repo",
			shouldErr: false,
		},
		{
			name:      "Uppercase SSH protocol",
			url:       "SSH://GIT@github.com/owner/repo.git",
			expected:  "https://github.com/owner/repo",
			shouldErr: false,
		},

		// Whitespace handling
		{
			name:      "URL with leading whitespace",
			url:       "  https://github.com/owner/repo",
			expected:  "https://github.com/owner/repo",
			shouldErr: false,
		},
		{
			name:      "URL with trailing whitespace",
			url:       "https://github.com/owner/repo  ",
			expected:  "https://github.com/owner/repo",
			shouldErr: false,
		},
		{
			name:      "URL with both leading and trailing whitespace",
			url:       "  https://github.com/owner/repo  ",
			expected:  "https://github.com/owner/repo",
			shouldErr: false,
		},

		// Error cases
		{
			name:      "Empty URL",
			url:       "",
			shouldErr: true,
		},
		{
			name:      "Only whitespace",
			url:       "   ",
			shouldErr: true,
		},
		{
			name:      "Invalid shorthand (no slash)",
			url:       "invalid",
			shouldErr: true,
		},
		{
			name:      "Invalid shorthand (empty owner)",
			url:       "/repo",
			shouldErr: true,
		},
		{
			name:      "Invalid shorthand (empty repo)",
			url:       "owner/",
			shouldErr: true,
		},
		{
			name:      "Invalid shorthand (too many parts)",
			url:       "owner/repo/extra",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeURL(tt.url)
			if tt.shouldErr {
				if err == nil {
					t.Errorf("expected error for URL %q, got nil", tt.url)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for URL %q: %v", tt.url, err)
				}
				if result != tt.expected {
					t.Errorf("NormalizeURL(%q) = %q, want %q", tt.url, result, tt.expected)
				}
			}
		})
	}
}

// TestNormalizeURLEquivalence tests that different URL formats for the same repository
// normalize to the same canonical form. This is critical for ensuring consistent
// handling of repository URLs across different formats.
func TestNormalizeURLEquivalence(t *testing.T) {
	equivalentGroups := []struct {
		name string
		urls []string
	}{
		{
			name: "GitHub repository variations",
			urls: []string{
				"https://github.com/owner/repo",
				"https://github.com/owner/repo/",
				"https://github.com/owner/repo.git",
				"http://github.com/owner/repo",
				"git@github.com:owner/repo",
				"git@github.com:owner/repo.git",
				"ssh://git@github.com/owner/repo",
				"ssh://git@github.com/owner/repo.git",
				"owner/repo",
				"HTTPS://GITHUB.COM/owner/repo",
				"https://GitHub.Com/owner/repo",
			},
		},
		{
			name: "GitLab repository variations",
			urls: []string{
				"https://gitlab.com/owner/repo",
				"https://gitlab.com/owner/repo/",
				"https://gitlab.com/owner/repo.git",
				"http://gitlab.com/owner/repo",
				"git@gitlab.com:owner/repo",
				"git@gitlab.com:owner/repo.git",
				"ssh://git@gitlab.com/owner/repo",
				"ssh://git@gitlab.com/owner/repo.git",
			},
		},
		{
			name: "Bitbucket repository variations",
			urls: []string{
				"https://bitbucket.org/owner/repo",
				"https://bitbucket.org/owner/repo/",
				"https://bitbucket.org/owner/repo.git",
				"http://bitbucket.org/owner/repo",
				"git@bitbucket.org:owner/repo",
				"git@bitbucket.org:owner/repo.git",
				"ssh://git@bitbucket.org/owner/repo",
			},
		},
	}

	for _, group := range equivalentGroups {
		t.Run(group.name, func(t *testing.T) {
			var normalizedURLs []string
			for _, url := range group.urls {
				normalized, err := NormalizeURL(url)
				if err != nil {
					t.Errorf("failed to normalize %q: %v", url, err)
					continue
				}
				normalizedURLs = append(normalizedURLs, normalized)
			}

			// All normalized URLs in a group should be identical
			if len(normalizedURLs) > 0 {
				expected := normalizedURLs[0]
				for i, normalized := range normalizedURLs {
					if normalized != expected {
						t.Errorf("URL %q normalized to %q, expected %q", group.urls[i], normalized, expected)
					}
				}
			}
		})
	}
}

// TestNormalizeURLRealWorldExamples tests URL normalization with real-world repository examples
func TestNormalizeURLRealWorldExamples(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Anthropic skills repo HTTPS",
			input:    "https://github.com/anthropics/skills",
			expected: "https://github.com/anthropics/skills",
		},
		{
			name:     "Anthropic skills repo SSH",
			input:    "git@github.com:anthropics/skills.git",
			expected: "https://github.com/anthropics/skills",
		},
		{
			name:     "Anthropic skills repo shorthand",
			input:    "anthropics/skills",
			expected: "https://github.com/anthropics/skills",
		},
		{
			name:     "Go-git library",
			input:    "https://github.com/go-git/go-git",
			expected: "https://github.com/go-git/go-git",
		},
		{
			name:     "Kubernetes main repo",
			input:    "https://github.com/kubernetes/kubernetes",
			expected: "https://github.com/kubernetes/kubernetes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeURL(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("NormalizeURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
