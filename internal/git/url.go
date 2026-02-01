package git

import (
	"fmt"
	"strings"
)

// NormalizeURL normalizes a Git repository URL to a canonical HTTPS format.
// It handles various URL formats including:
// - HTTPS URLs (https://github.com/owner/repo)
// - HTTP URLs (converted to HTTPS)
// - SSH URLs (git@github.com:owner/repo.git)
// - SSH protocol URLs (ssh://git@github.com/owner/repo)
// - Shorthand notation (owner/repo, defaults to GitHub)
//
// The function:
// - Removes trailing slashes
// - Removes .git suffixes
// - Converts SSH format to HTTPS
// - Normalizes domain names to lowercase
// - Upgrades HTTP to HTTPS
func NormalizeURL(repoURL string) (string, error) {
	repoURL = strings.TrimSpace(repoURL)

	if repoURL == "" {
		return "", fmt.Errorf("repository URL cannot be empty")
	}

	// Remove trailing slashes and .git suffix
	repoURL = strings.TrimRight(repoURL, "/")
	repoURL = strings.TrimSuffix(repoURL, ".git")

	// Convert SSH format to HTTPS (git@github.com:owner/repo -> https://github.com/owner/repo)
	if strings.HasPrefix(strings.ToLower(repoURL), "git@") {
		// Format: git@github.com:owner/repo
		repoURL = strings.TrimPrefix(repoURL, "git@")
		repoURL = strings.TrimPrefix(repoURL, "GIT@") // Handle uppercase
		repoURL = strings.Replace(repoURL, ":", "/", 1)
		repoURL = "https://" + repoURL
	}

	// Convert ssh:// format to HTTPS (ssh://git@github.com/owner/repo -> https://github.com/owner/repo)
	if strings.HasPrefix(strings.ToLower(repoURL), "ssh://") {
		repoURL = strings.TrimPrefix(repoURL, "ssh://")
		repoURL = strings.TrimPrefix(repoURL, "SSH://") // Handle uppercase
		repoURL = strings.TrimPrefix(repoURL, "git@")
		repoURL = strings.TrimPrefix(repoURL, "GIT@") // Handle uppercase
		repoURL = "https://" + repoURL
	}

	// Normalize protocol and domain to lowercase for case-insensitive comparison
	if strings.HasPrefix(strings.ToLower(repoURL), "https://") || strings.HasPrefix(strings.ToLower(repoURL), "http://") {
		// Parse URL to normalize the domain
		parts := strings.SplitN(repoURL, "://", 2)
		if len(parts) == 2 {
			protocol := strings.ToLower(parts[0])
			remainder := parts[1]
			// Split domain from path
			domainAndPath := strings.SplitN(remainder, "/", 2)
			domain := strings.ToLower(domainAndPath[0])
			path := ""
			if len(domainAndPath) > 1 {
				path = "/" + domainAndPath[1]
			}
			repoURL = protocol + "://" + domain + path
		}
	}

	// If it's already an HTTPS URL, validate and return
	if strings.HasPrefix(strings.ToLower(repoURL), "https://") {
		// Basic URL validation
		if !strings.Contains(repoURL, "://") {
			return "", fmt.Errorf("invalid URL format: %s", repoURL)
		}
		return repoURL, nil
	}

	// Convert HTTP to HTTPS
	if strings.HasPrefix(strings.ToLower(repoURL), "http://") {
		repoURL = strings.Replace(repoURL, "http://", "https://", 1)
		repoURL = strings.Replace(repoURL, "HTTP://", "https://", 1)
		return repoURL, nil
	}

	// Handle GitHub shorthand (owner/repo)
	if !strings.Contains(repoURL, "://") && !strings.HasPrefix(repoURL, "git@") {
		if !strings.Contains(repoURL, "/") {
			return "", fmt.Errorf("invalid repository format: %s", repoURL)
		}

		parts := strings.Split(repoURL, "/")
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid repository format: %s (expected owner/repo)", repoURL)
		}

		// Validate shorthand format
		if parts[0] == "" || parts[1] == "" {
			return "", fmt.Errorf("invalid repository format: %s (empty owner or repo)", repoURL)
		}

		// Default to GitHub for shorthand notation
		return fmt.Sprintf("https://github.com/%s", repoURL), nil
	}

	return "", fmt.Errorf("unrecognized repository URL format: %s", repoURL)
}
