package git

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// GetAuthMethod returns an appropriate auth method for the given URL
// It tries multiple authentication methods in order:
// 1. SSH agent (for git@... URLs)
// 2. GitHub token from environment (for HTTPS URLs)
// 3. No auth (for public repos)
func GetAuthMethod(url string) (transport.AuthMethod, error) {
	// For SSH URLs, use SSH agent
	if isSSHURL(url) {
		return getSSHAuth()
	}

	// For HTTPS URLs, try token-based auth
	if isHTTPSURL(url) {
		return getHTTPSAuth()
	}

	// No auth needed for public repos or local paths
	return nil, nil
}

// isSSHURL checks if the URL is an SSH URL
func isSSHURL(url string) bool {
	return strings.HasPrefix(url, "git@") ||
		strings.HasPrefix(url, "ssh://")
}

// isHTTPSURL checks if the URL is an HTTPS URL
func isHTTPSURL(url string) bool {
	return strings.HasPrefix(url, "https://")
}

// getSSHAuth attempts to create SSH authentication using ssh-agent
func getSSHAuth() (transport.AuthMethod, error) {
	// Try to use ssh-agent
	auth, err := ssh.NewSSHAgentAuth("git")
	if err != nil {
		// If ssh-agent is not available, try to use default SSH keys
		return getDefaultSSHKeys()
	}
	return auth, nil
}

// getDefaultSSHKeys attempts to use default SSH keys from ~/.ssh
func getDefaultSSHKeys() (transport.AuthMethod, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Try common SSH key locations in order
	keyPaths := []string{
		homeDir + "/.ssh/id_rsa",
		homeDir + "/.ssh/id_ed25519",
		homeDir + "/.ssh/id_ecdsa",
		homeDir + "/.ssh/id_dsa",
	}

	var lastErr error
	for _, keyPath := range keyPaths {
		// Check if key file exists
		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			continue
		}

		// Try to use the key (will prompt for passphrase if needed)
		auth, err := ssh.NewPublicKeysFromFile("git", keyPath, "")
		if err == nil {
			return auth, nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to load SSH keys: %w", lastErr)
	}

	return nil, fmt.Errorf("no SSH keys found in ~/.ssh/")
}

// getHTTPSAuth attempts to create HTTPS authentication using environment variables or git credential helper
func getHTTPSAuth() (transport.AuthMethod, error) {
	// Try environment variables first
	// Support common token environment variables
	tokenVars := []string{
		"GITHUB_TOKEN",
		"GH_TOKEN",
		"GITLAB_TOKEN",
		"GIT_TOKEN",
	}

	for _, varName := range tokenVars {
		if token := os.Getenv(varName); token != "" {
			return &http.BasicAuth{
				Username: "git", // Can be anything except an empty string
				Password: token,
			}, nil
		}
	}

	// If no token found, return nil (will attempt without auth for public repos)
	return nil, nil
}

// isSSHAgentAvailable checks if ssh-agent is running and accessible
func isSSHAgentAvailable() bool {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		return false
	}

	// Try to connect to the socket
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
