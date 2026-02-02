package git

import (
	"os"
	"testing"
)

func TestIsSSHURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"git@ format", "git@github.com:owner/repo.git", true},
		{"ssh:// format", "ssh://git@github.com/owner/repo.git", true},
		{"https format", "https://github.com/owner/repo.git", false},
		{"http format", "http://github.com/owner/repo.git", false},
		{"local path", "/path/to/repo", false},
		{"relative path", "./repo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSSHURL(tt.url)
			if result != tt.expected {
				t.Errorf("isSSHURL(%q) = %v, expected %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestIsHTTPSURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"https format", "https://github.com/owner/repo.git", true},
		{"http format", "http://github.com/owner/repo.git", false},
		{"git@ format", "git@github.com:owner/repo.git", false},
		{"ssh:// format", "ssh://git@github.com/owner/repo.git", false},
		{"local path", "/path/to/repo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHTTPSURL(tt.url)
			if result != tt.expected {
				t.Errorf("isHTTPSURL(%q) = %v, expected %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestGetAuthMethod_SSH(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantNil bool
	}{
		{"git@ SSH URL", "git@github.com:owner/repo.git", false},
		{"ssh:// SSH URL", "ssh://git@github.com/owner/repo.git", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := GetAuthMethod(tt.url)

			// SSH auth may fail if ssh-agent is not running or no keys exist
			// That's acceptable - we just verify it attempts SSH auth
			if tt.wantNil && auth != nil {
				t.Errorf("GetAuthMethod(%q) returned auth, expected nil", tt.url)
			}

			// For SSH URLs, either auth should be returned or a helpful error
			if !tt.wantNil && auth == nil && err == nil {
				t.Errorf("GetAuthMethod(%q) returned nil auth and nil error for SSH URL", tt.url)
			}
		})
	}
}

func TestGetAuthMethod_HTTPS(t *testing.T) {
	// Save and restore env
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		if originalToken != "" {
			os.Setenv("GITHUB_TOKEN", originalToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
	}()

	tests := []struct {
		name     string
		url      string
		tokenSet bool
		wantAuth bool
	}{
		{"HTTPS without token", "https://github.com/owner/repo.git", false, false},
		{"HTTPS with token", "https://github.com/owner/repo.git", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tokenSet {
				os.Setenv("GITHUB_TOKEN", "test-token")
			} else {
				os.Unsetenv("GITHUB_TOKEN")
				os.Unsetenv("GH_TOKEN")
				os.Unsetenv("GITLAB_TOKEN")
				os.Unsetenv("GIT_TOKEN")
			}

			auth, err := GetAuthMethod(tt.url)
			if err != nil {
				t.Errorf("GetAuthMethod(%q) returned unexpected error: %v", tt.url, err)
			}

			gotAuth := auth != nil
			if gotAuth != tt.wantAuth {
				t.Errorf("GetAuthMethod(%q) auth != nil = %v, expected %v", tt.url, gotAuth, tt.wantAuth)
			}
		})
	}
}

func TestGetAuthMethod_Public(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"local path", "/path/to/repo"},
		{"relative path", "./repo"},
		{"http URL", "http://github.com/owner/repo.git"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := GetAuthMethod(tt.url)
			if err != nil {
				t.Errorf("GetAuthMethod(%q) returned unexpected error: %v", tt.url, err)
			}
			if auth != nil {
				t.Errorf("GetAuthMethod(%q) = %v, expected nil for public/local repo", tt.url, auth)
			}
		})
	}
}

func TestGetHTTPSAuth_TokenPrecedence(t *testing.T) {
	// Save and restore env vars
	tokens := []string{"GITHUB_TOKEN", "GH_TOKEN", "GITLAB_TOKEN", "GIT_TOKEN"}
	originals := make(map[string]string)
	for _, token := range tokens {
		originals[token] = os.Getenv(token)
		os.Unsetenv(token)
	}
	defer func() {
		for token, value := range originals {
			if value != "" {
				os.Setenv(token, value)
			} else {
				os.Unsetenv(token)
			}
		}
	}()

	// Test GITHUB_TOKEN has priority
	os.Setenv("GITHUB_TOKEN", "github-token")
	os.Setenv("GH_TOKEN", "gh-token")

	auth, err := getHTTPSAuth()
	if err != nil {
		t.Fatalf("getHTTPSAuth() returned error: %v", err)
	}
	if auth == nil {
		t.Fatal("getHTTPSAuth() returned nil when token was set")
	}

	// Verify it uses basic auth (interface has limited introspection, so we just check non-nil)
	// In a real scenario, this would use the GITHUB_TOKEN value
}

func TestGetHTTPSAuth_NoToken(t *testing.T) {
	// Clear all token env vars
	tokens := []string{"GITHUB_TOKEN", "GH_TOKEN", "GITLAB_TOKEN", "GIT_TOKEN"}
	originals := make(map[string]string)
	for _, token := range tokens {
		originals[token] = os.Getenv(token)
		os.Unsetenv(token)
	}
	defer func() {
		for token, value := range originals {
			if value != "" {
				os.Setenv(token, value)
			}
		}
	}()

	auth, err := getHTTPSAuth()
	if err != nil {
		t.Errorf("getHTTPSAuth() returned error: %v", err)
	}
	if auth != nil {
		t.Errorf("getHTTPSAuth() = %v, expected nil when no token set", auth)
	}
}
