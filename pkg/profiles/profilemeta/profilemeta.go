package profilemeta

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tjg184/agent-smith/internal/detector"
)

// ProfileMetadata stores metadata about a profile's source.
type ProfileMetadata struct {
	Type      string `json:"type"`       // "repo" or "user"
	SourceURL string `json:"source_url"` // Only populated for type="repo"
}

// Save saves metadata about a profile's source URL to the given profileDir.
// The source URL is normalized before being saved for consistent duplicate detection.
func Save(profileDir, sourceURL string) error {
	metadataPath := filepath.Join(profileDir, ".profile-metadata")

	rd := detector.NewRepositoryDetector()
	normalizedURL, err := rd.NormalizeURL(sourceURL)
	if err != nil {
		normalizedURL = sourceURL
	}

	meta := ProfileMetadata{
		Type:      "repo",
		SourceURL: normalizedURL,
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// SaveUser saves metadata for a user-created profile (type="user", no source URL).
func SaveUser(profileDir string) error {
	metadataPath := filepath.Join(profileDir, ".profile-metadata")

	meta := ProfileMetadata{Type: "user"}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// Load loads metadata for the given profileDir.
// Returns nil (no error) if the metadata file does not exist.
func Load(profileDir string) (*ProfileMetadata, error) {
	metadataPath := filepath.Join(profileDir, ".profile-metadata")

	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var meta ProfileMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &meta, nil
}

// GetProfileType returns the type of a profile by reading its metadata file.
// Returns "unknown" for profiles without metadata (legacy profiles).
func GetProfileType(profileDir string) (string, error) {
	meta, err := Load(profileDir)
	if err != nil {
		return "", fmt.Errorf("failed to load metadata: %w", err)
	}

	if meta == nil || meta.Type == "" {
		return "unknown", nil
	}

	return meta.Type, nil
}

// FindBySourceURL scans profilesDir for a profile whose saved source URL matches repoURL.
// The input URL is normalized before comparison. Returns "" if not found.
func FindBySourceURL(profilesDir, repoURL string) (string, error) {
	rd := detector.NewRepositoryDetector()
	normalizedURL, err := rd.NormalizeURL(repoURL)
	if err != nil {
		normalizedURL = repoURL
	}

	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read profiles directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		profileDir := filepath.Join(profilesDir, entry.Name())
		meta, err := Load(profileDir)
		if err != nil {
			continue
		}
		if meta != nil && meta.SourceURL == normalizedURL {
			return entry.Name(), nil
		}
	}

	return "", nil
}

// GenerateNameFromRepo generates a unique profile name from a repository URL.
// If the name already exists in existingProfiles, appends a short hash suffix.
func GenerateNameFromRepo(repoURL string, existingProfiles []string) string {
	repoURL = strings.TrimRight(repoURL, "/")
	repoURL = strings.TrimSuffix(repoURL, ".git")

	var baseName string

	slashCount := strings.Count(repoURL, "/")
	isShorthand := slashCount == 1 &&
		!strings.Contains(repoURL, "://") &&
		!strings.Contains(repoURL, "@") &&
		!strings.Contains(repoURL, ".") &&
		!strings.HasPrefix(repoURL, "./") &&
		!strings.HasPrefix(repoURL, "../")

	if isShorthand {
		baseName = SanitizeForProfileName(strings.ReplaceAll(repoURL, "/", "-"))
	} else if strings.Contains(repoURL, "github.com") || strings.Contains(repoURL, "gitlab.com") || strings.Contains(repoURL, "bitbucket.org") {
		parts := strings.Split(repoURL, "/")
		if len(parts) >= 2 {
			owner := parts[len(parts)-2]
			repo := parts[len(parts)-1]
			if strings.Contains(owner, ":") {
				owner = strings.Split(owner, ":")[1]
			}
			baseName = fmt.Sprintf("%s-%s", SanitizeForProfileName(owner), SanitizeForProfileName(repo))
		}
	} else if filepath.IsAbs(repoURL) || strings.HasPrefix(repoURL, "./") || strings.HasPrefix(repoURL, "../") {
		baseName = SanitizeForProfileName(filepath.Base(repoURL))
	} else {
		parts := strings.Split(repoURL, "/")
		if len(parts) > 0 {
			baseName = SanitizeForProfileName(parts[len(parts)-1])
		} else {
			baseName = "repo"
		}
	}

	if baseName == "" {
		baseName = "repo"
	}

	profileName := baseName
	existsMap := make(map[string]bool, len(existingProfiles))
	for _, p := range existingProfiles {
		existsMap[p] = true
	}

	if !existsMap[profileName] {
		return profileName
	}

	hash := sha256.Sum256([]byte(repoURL))
	shortHash := hex.EncodeToString(hash[:])[:6]
	profileName = fmt.Sprintf("%s-%s", baseName, shortHash)

	counter := 2
	base := profileName
	for existsMap[profileName] {
		profileName = fmt.Sprintf("%s-%d", base, counter)
		counter++
	}

	return profileName
}

// ValidateProfileName checks that the profile name meets naming requirements.
func ValidateProfileName(name string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	if strings.Contains(name, "..") || strings.Contains(name, "./") {
		return fmt.Errorf("profile name cannot contain path traversal patterns (.. or ./)")
	}

	if strings.HasPrefix(name, ".") {
		return fmt.Errorf("profile name cannot start with '.' (hidden directories not allowed)")
	}

	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("profile name cannot contain path separators (/ or \\)")
	}

	validPattern := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	if !validPattern.MatchString(name) {
		return fmt.Errorf("profile name must contain only letters, numbers, and hyphens (got '%s')", name)
	}

	return nil
}

func SanitizeForProfileName(input string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9-]+`)
	sanitized := reg.ReplaceAllString(input, "-")
	sanitized = strings.Trim(sanitized, "-")
	if sanitized == "" {
		return "repo"
	}
	return sanitized
}
