package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tgaines/agent-smith/internal/models"
)

// TestNewRepositoryDetector tests creating a new detector with default config
func TestNewRepositoryDetector(t *testing.T) {
	rd := NewRepositoryDetector()

	if rd == nil {
		t.Fatal("expected non-nil RepositoryDetector")
	}

	if rd.patterns == nil {
		t.Error("expected patterns map to be initialized")
	}

	if rd.detectionConfig == nil {
		t.Error("expected detectionConfig to be initialized")
	}

	// Verify default patterns exist
	expectedPatterns := []string{"github", "gitlab", "bitbucket", "git"}
	for _, pattern := range expectedPatterns {
		if _, exists := rd.patterns[pattern]; !exists {
			t.Errorf("expected pattern %s to exist", pattern)
		}
	}
}

// TestNewRepositoryDetectorWithConfig tests creating detector with custom config
func TestNewRepositoryDetectorWithConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.json")

	// Use default config for now
	rd := NewRepositoryDetectorWithConfig(configPath)

	if rd == nil {
		t.Fatal("expected non-nil RepositoryDetector")
	}

	if rd.detectionConfig == nil {
		t.Error("expected detectionConfig to be initialized")
	}
}

// TestDetectProvider tests provider detection from various URLs
func TestDetectProvider(t *testing.T) {
	rd := NewRepositoryDetector()

	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "GitHub HTTPS",
			url:      "https://github.com/user/repo",
			expected: "github",
		},
		{
			name:     "GitHub with www",
			url:      "https://www.github.com/user/repo",
			expected: "github",
		},
		{
			name:     "GitHub SSH",
			url:      "git@github.com:user/repo.git",
			expected: "git", // Matches generic git pattern due to .git suffix
		},
		{
			name:     "GitLab HTTPS",
			url:      "https://gitlab.com/user/repo",
			expected: "gitlab",
		},
		{
			name:     "GitLab SSH",
			url:      "git@gitlab.com:user/repo.git",
			expected: "git", // Matches generic git pattern due to .git suffix
		},
		{
			name:     "Bitbucket HTTPS",
			url:      "https://bitbucket.org/user/repo",
			expected: "bitbucket",
		},
		{
			name:     "Bitbucket SSH",
			url:      "git@bitbucket.org:user/repo.git",
			expected: "git", // Matches generic git pattern due to .git suffix
		},
		{
			name:     "Generic git URL",
			url:      "https://git.example.com/repo.git",
			expected: "git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rd.DetectProvider(tt.url)
			if result != tt.expected {
				t.Errorf("expected provider %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestDetectProviderLocal tests local path detection
func TestDetectProviderLocal(t *testing.T) {
	rd := NewRepositoryDetector()

	// Create a temporary directory with .git folder
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	provider := rd.DetectProvider(tempDir)
	if provider != "local" {
		t.Errorf("expected provider 'local', got %s", provider)
	}
}

// TestIsLocalPath tests local path detection
func TestIsLocalPath(t *testing.T) {
	rd := NewRepositoryDetector()

	// Create a temporary directory
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Absolute path (existing)",
			path:     tempDir,
			expected: true,
		},
		{
			name:     "Non-existent absolute path",
			path:     "/non/existent/path",
			expected: false,
		},
		{
			name:     "HTTP URL",
			path:     "https://github.com/user/repo",
			expected: false,
		},
		{
			name:     "SSH URL",
			path:     "git@github.com:user/repo.git",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rd.IsLocalPath(tt.path)
			if result != tt.expected {
				t.Errorf("expected IsLocalPath(%s) = %v, got %v", tt.path, tt.expected, result)
			}
		})
	}
}

// TestNormalizeURL tests URL normalization
func TestNormalizeURL(t *testing.T) {
	rd := NewRepositoryDetector()

	tests := []struct {
		name      string
		url       string
		expected  string
		shouldErr bool
	}{
		{
			name:      "GitHub shorthand",
			url:       "user/repo",
			expected:  "https://github.com/user/repo",
			shouldErr: false,
		},
		{
			name:      "Full GitHub HTTPS URL",
			url:       "https://github.com/user/repo",
			expected:  "https://github.com/user/repo",
			shouldErr: false,
		},
		{
			name:      "GitHub SSH URL",
			url:       "git@github.com:user/repo.git",
			expected:  "git@github.com:user/repo.git",
			shouldErr: false,
		},
		{
			name:      "Invalid shorthand (no slash)",
			url:       "invalid",
			shouldErr: true,
		},
		{
			name:      "Invalid shorthand (empty parts)",
			url:       "/repo",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := rd.NormalizeURL(tt.url)
			if tt.shouldErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("expected normalized URL %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

// TestNormalizeURLLocal tests local path normalization
func TestNormalizeURLLocal(t *testing.T) {
	rd := NewRepositoryDetector()

	// Create a temporary directory with .git folder
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	result, err := rd.NormalizeURL(tempDir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should return absolute path
	absPath, _ := filepath.Abs(tempDir)
	if result != absPath {
		t.Errorf("expected absolute path %s, got %s", absPath, result)
	}
}

// TestValidateRepository tests repository validation
func TestValidateRepository(t *testing.T) {
	rd := NewRepositoryDetector()

	tests := []struct {
		name      string
		url       string
		shouldErr bool
	}{
		{
			name:      "Valid GitHub URL",
			url:       "https://github.com/user/repo",
			shouldErr: false,
		},
		{
			name:      "Valid GitLab URL",
			url:       "https://gitlab.com/user/repo",
			shouldErr: false,
		},
		{
			name:      "Valid SSH URL",
			url:       "git@github.com:user/repo.git",
			shouldErr: false,
		},
		{
			name:      "Valid git URL",
			url:       "https://git.example.com/repo.git",
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rd.ValidateRepository(tt.url)
			if tt.shouldErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestValidateRepositoryLocal tests local repository validation
func TestValidateRepositoryLocal(t *testing.T) {
	rd := NewRepositoryDetector()

	// Create a temporary directory with .git folder
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Valid local repository
	err := rd.ValidateRepository(tempDir)
	if err != nil {
		t.Errorf("expected valid local repository, got error: %v", err)
	}

	// Invalid local path (no .git folder)
	tempDirInvalid := t.TempDir()
	err = rd.ValidateRepository(tempDirInvalid)
	if err == nil {
		t.Error("expected error for local path without .git folder")
	}
}

// TestCreateDefaultDetectionConfig tests default config creation
func TestCreateDefaultDetectionConfig(t *testing.T) {
	config := createDefaultDetectionConfig()

	if config == nil {
		t.Fatal("expected non-nil DetectionConfig")
	}

	if config.Components == nil {
		t.Fatal("expected Components map to be initialized")
	}

	// Verify default components exist
	expectedComponents := []string{
		string(models.ComponentSkill),
		string(models.ComponentAgent),
		string(models.ComponentCommand),
	}

	for _, componentType := range expectedComponents {
		if _, exists := config.Components[componentType]; !exists {
			t.Errorf("expected component type %s to exist in default config", componentType)
		}
	}

	// Verify skill component has exact files
	skillPattern := config.Components[string(models.ComponentSkill)]
	if len(skillPattern.ExactFiles) == 0 {
		t.Error("expected skill component to have exact files configured")
	}

	// Verify agent component has path patterns
	agentPattern := config.Components[string(models.ComponentAgent)]
	if len(agentPattern.PathPatterns) == 0 {
		t.Error("expected agent component to have path patterns configured")
	}
}

// TestSaveDetectionConfig tests saving detection config
func TestSaveDetectionConfig(t *testing.T) {
	rd := NewRepositoryDetector()

	// Create a temporary directory for config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "detection-config.json")

	// Save config
	err := rd.SaveDetectionConfig(configPath)
	if err != nil {
		t.Errorf("unexpected error saving config: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("expected config file to be created")
	}
}

// TestLoadDetectionConfig tests loading detection config
func TestLoadDetectionConfig(t *testing.T) {
	rd := NewRepositoryDetector()

	// Test loading non-existent config (should use defaults)
	err := rd.loadDetectionConfig("/non/existent/path.json")
	if err != nil {
		t.Errorf("expected no error when loading non-existent config, got: %v", err)
	}

	if rd.detectionConfig == nil {
		t.Error("expected default config to be loaded")
	}
}
