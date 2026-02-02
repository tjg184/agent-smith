package git

import (
	"fmt"
	"strings"
)

// NormalizeURL normalizes a Git repository URL to a canonical format.
// It handles various URL formats including:
// - HTTPS URLs (https://github.com/owner/repo)
// - HTTP URLs (converted to HTTPS)
// - SSH URLs (git@github.com:owner/repo.git) - preserved as SSH
// - SSH protocol URLs (ssh://git@github.com/owner/repo) - preserved as SSH
// - Shorthand notation (owner/repo, defaults to GitHub HTTPS)
//
// The function:
// - Removes trailing slashes
// - Removes .git suffixes
// - Preserves SSH URLs (no longer converts to HTTPS)
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

	// Preserve SSH URLs (git@github.com:owner/repo format)
	if strings.HasPrefix(strings.ToLower(repoURL), "git@") {
		// SSH URLs are already in canonical format, just return them
		return repoURL, nil
	}

	// Preserve SSH protocol URLs (ssh://git@github.com/owner/repo format)
	if strings.HasPrefix(strings.ToLower(repoURL), "ssh://") {
		// SSH URLs are already in canonical format, just return them
		return repoURL, nil
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
