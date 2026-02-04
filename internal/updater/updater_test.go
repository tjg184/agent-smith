package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestNewUpdateDetectorWithProfile_ExplicitProfile tests that when an explicit profile
// is provided, it uses that profile's directory
func TestNewUpdateDetectorWithProfile_ExplicitProfile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-updater-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set HOME to test directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Create test profile directory
	testProfileName := "test-profile"
	testProfileDir := filepath.Join(tempDir, ".agent-smith", "profiles", testProfileName)
	if err := os.MkdirAll(testProfileDir, 0755); err != nil {
		t.Fatalf("Failed to create test profile directory: %v", err)
	}

	// When explicit profile is provided, it should use that profile
	detector := NewUpdateDetectorWithProfile(testProfileName)

	expectedBaseDir := filepath.Join(tempDir, ".agent-smith", "profiles", testProfileName)
	if detector.baseDir != expectedBaseDir {
		t.Errorf("Expected baseDir %s, got %s", expectedBaseDir, detector.baseDir)
	}

	if detector.profileName != testProfileName {
		t.Errorf("Expected profileName %s, got %s", testProfileName, detector.profileName)
	}
}

// TestNewUpdateDetectorWithProfile_NoProfileNoActive tests that when no profile is
// provided and no active profile exists, it uses the base directory
func TestNewUpdateDetectorWithProfile_NoProfileNoActive(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-updater-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set HOME to test directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Create base directory
	baseDir := filepath.Join(tempDir, ".agent-smith")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatalf("Failed to create base directory: %v", err)
	}

	// When no profile is provided and no active profile, should use base directory
	detector := NewUpdateDetectorWithProfile("")

	expectedBaseDir := filepath.Join(tempDir, ".agent-smith")
	if detector.baseDir != expectedBaseDir {
		t.Errorf("Expected baseDir %s, got %s", expectedBaseDir, detector.baseDir)
	}

	if detector.profileName != "" {
		t.Errorf("Expected empty profileName, got %s", detector.profileName)
	}
}

// TestNewUpdateDetectorWithProfile_NoProfileWithActive tests that when no profile is
// provided but an active profile exists, it uses the active profile
func TestNewUpdateDetectorWithProfile_NoProfileWithActive(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-updater-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set HOME to test directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Create test profile directory
	testProfileName := "test-profile"
	testProfileDir := filepath.Join(tempDir, ".agent-smith", "profiles", testProfileName)
	if err := os.MkdirAll(testProfileDir, 0755); err != nil {
		t.Fatalf("Failed to create test profile directory: %v", err)
	}

	// Activate a profile (active profile file is in ~/.agent-smith/, not profiles/)
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	activeProfileFile := filepath.Join(agentSmithDir, ".active-profile")
	if err := os.MkdirAll(agentSmithDir, 0755); err != nil {
		t.Fatalf("Failed to create agent-smith directory: %v", err)
	}
	if err := os.WriteFile(activeProfileFile, []byte(testProfileName), 0644); err != nil {
		t.Fatalf("Failed to write active profile file: %v", err)
	}

	// When no profile is provided but active profile exists, should use active profile
	detector := NewUpdateDetectorWithProfile("")

	expectedBaseDir := filepath.Join(tempDir, ".agent-smith", "profiles", testProfileName)
	if detector.baseDir != expectedBaseDir {
		t.Errorf("Expected baseDir %s, got %s", expectedBaseDir, detector.baseDir)
	}

	if detector.profileName != testProfileName {
		t.Errorf("Expected profileName %s, got %s", testProfileName, detector.profileName)
	}
}

// TestNewUpdateDetectorWithProfile_ExplicitProfileOverridesActive tests that when both
// an explicit profile and an active profile exist, the explicit profile takes precedence
func TestNewUpdateDetectorWithProfile_ExplicitProfileOverridesActive(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-updater-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set HOME to test directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Create two test profile directories
	activeProfileName := "active-profile"
	explicitProfileName := "explicit-profile"

	activeProfileDir := filepath.Join(tempDir, ".agent-smith", "profiles", activeProfileName)
	if err := os.MkdirAll(activeProfileDir, 0755); err != nil {
		t.Fatalf("Failed to create active profile directory: %v", err)
	}

	explicitProfileDir := filepath.Join(tempDir, ".agent-smith", "profiles", explicitProfileName)
	if err := os.MkdirAll(explicitProfileDir, 0755); err != nil {
		t.Fatalf("Failed to create explicit profile directory: %v", err)
	}

	// Set active profile (active profile file is in ~/.agent-smith/, not profiles/)
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	activeProfileFile := filepath.Join(agentSmithDir, ".active-profile")
	if err := os.MkdirAll(agentSmithDir, 0755); err != nil {
		t.Fatalf("Failed to create agent-smith directory: %v", err)
	}
	if err := os.WriteFile(activeProfileFile, []byte(activeProfileName), 0644); err != nil {
		t.Fatalf("Failed to write active profile file: %v", err)
	}

	// Create detector with explicit profile (should override active profile)
	detector := NewUpdateDetectorWithProfile(explicitProfileName)

	expectedBaseDir := filepath.Join(tempDir, ".agent-smith", "profiles", explicitProfileName)
	if detector.baseDir != expectedBaseDir {
		t.Errorf("Expected baseDir %s, got %s", expectedBaseDir, detector.baseDir)
	}

	if detector.profileName != explicitProfileName {
		t.Errorf("Expected profileName %s, got %s", explicitProfileName, detector.profileName)
	}
}

// TestNewUpdateDetectorWithBaseDir tests that the new constructor correctly uses
// an explicit base directory without any profile detection logic
func TestNewUpdateDetectorWithBaseDir(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-updater-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a custom base directory
	customBaseDir := filepath.Join(tempDir, "custom-base-directory")
	if err := os.MkdirAll(customBaseDir, 0755); err != nil {
		t.Fatalf("Failed to create custom base directory: %v", err)
	}

	// Create detector with explicit base directory
	detector := NewUpdateDetectorWithBaseDir(customBaseDir)

	// Verify that the detector uses the exact base directory provided
	if detector.baseDir != customBaseDir {
		t.Errorf("Expected baseDir %s, got %s", customBaseDir, detector.baseDir)
	}

	// Verify that profileName is empty (no profile management)
	if detector.profileName != "" {
		t.Errorf("Expected empty profileName, got %s", detector.profileName)
	}

	// Verify that detector is not nil (basic sanity check)
	if detector.detector == nil {
		t.Error("Expected detector to be initialized, got nil")
	}
}

// TestNewUpdateDetectorWithBaseDir_ProfileDirectory tests that when given a profile
// directory path, it works correctly without trying to detect or manage profiles
func TestNewUpdateDetectorWithBaseDir_ProfileDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-updater-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set HOME to test directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Create a profile directory structure
	profileName := "test-profile"
	profileDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName)
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		t.Fatalf("Failed to create profile directory: %v", err)
	}

	// Also create and activate a DIFFERENT profile to ensure it's ignored
	activeProfileName := "active-profile"
	activeProfileDir := filepath.Join(tempDir, ".agent-smith", "profiles", activeProfileName)
	if err := os.MkdirAll(activeProfileDir, 0755); err != nil {
		t.Fatalf("Failed to create active profile directory: %v", err)
	}

	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	activeProfileFile := filepath.Join(agentSmithDir, ".active-profile")
	if err := os.WriteFile(activeProfileFile, []byte(activeProfileName), 0644); err != nil {
		t.Fatalf("Failed to write active profile file: %v", err)
	}

	// Create detector with explicit profile directory (should NOT use active profile)
	detector := NewUpdateDetectorWithBaseDir(profileDir)

	// Verify that the detector uses the exact profile directory provided
	if detector.baseDir != profileDir {
		t.Errorf("Expected baseDir %s, got %s", profileDir, detector.baseDir)
	}

	// Verify that profileName is empty (caller manages the directory)
	if detector.profileName != "" {
		t.Errorf("Expected empty profileName, got %s", detector.profileName)
	}
}

// TestNewUpdateDetectorWithBaseDir_BaseDirectory tests that when given the base
// agent-smith directory, it works correctly
func TestNewUpdateDetectorWithBaseDir_BaseDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-updater-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create base agent-smith directory
	baseDir := filepath.Join(tempDir, ".agent-smith")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatalf("Failed to create base directory: %v", err)
	}

	// Create detector with base directory
	detector := NewUpdateDetectorWithBaseDir(baseDir)

	// Verify that the detector uses the base directory
	if detector.baseDir != baseDir {
		t.Errorf("Expected baseDir %s, got %s", baseDir, detector.baseDir)
	}

	// Verify that profileName is empty
	if detector.profileName != "" {
		t.Errorf("Expected empty profileName, got %s", detector.profileName)
	}
}

// TestLoadMetadata_Success tests successfully loading metadata from a lock file
func TestLoadMetadata_Success(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-updater-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a sample lock file in the base directory (not in subdirectory)
	lockFilePath := filepath.Join(tempDir, ".skill-lock.json")
	lockFileContent := `{
  "version": 4,
  "skills": {
    "https://github.com/user/repo": {
      "test-skill": {
        "source": "github.com/user/repo",
        "sourceType": "git",
        "sourceUrl": "https://github.com/user/repo",
        "commitHash": "abc123",
        "installedAt": "2024-01-01T00:00:00Z",
        "updatedAt": "2024-01-01T00:00:00Z",
        "version": 3
      }
    }
  }
}`
	if err := os.WriteFile(lockFilePath, []byte(lockFileContent), 0644); err != nil {
		t.Fatalf("Failed to write lock file: %v", err)
	}

	// Create detector
	detector := NewUpdateDetectorWithBaseDir(tempDir)

	// Load metadata
	metadata, err := detector.LoadMetadata("skills", "test-skill")
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	// Verify metadata
	if metadata.Name != "test-skill" {
		t.Errorf("Expected name 'test-skill', got '%s'", metadata.Name)
	}
	if metadata.Source != "https://github.com/user/repo" {
		t.Errorf("Expected source 'https://github.com/user/repo', got '%s'", metadata.Source)
	}
	if metadata.Commit != "abc123" {
		t.Errorf("Expected commit 'abc123', got '%s'", metadata.Commit)
	}
}

// TestLoadMetadata_NotFound tests loading metadata when component is not found
func TestLoadMetadata_NotFound(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-updater-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create an empty lock file in the base directory
	lockFilePath := filepath.Join(tempDir, ".skill-lock.json")
	lockFileContent := `{
  "version": 4,
  "skills": {}
}`
	if err := os.WriteFile(lockFilePath, []byte(lockFileContent), 0644); err != nil {
		t.Fatalf("Failed to write lock file: %v", err)
	}

	// Create detector
	detector := NewUpdateDetectorWithBaseDir(tempDir)

	// Attempt to load non-existent metadata
	_, err = detector.LoadMetadata("skills", "nonexistent-skill")
	if err == nil {
		t.Fatal("Expected error when loading non-existent component, got nil")
	}
}

// TestLoadMetadata_MissingLockFile tests loading metadata when lock file doesn't exist
func TestLoadMetadata_MissingLockFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-updater-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create detector (no lock file exists)
	detector := NewUpdateDetectorWithBaseDir(tempDir)

	// Attempt to load metadata
	_, err = detector.LoadMetadata("skills", "test-skill")
	if err == nil {
		t.Fatal("Expected error when lock file doesn't exist, got nil")
	}
}

// TestLoadMetadata_AllComponentTypes tests loading metadata for different component types
func TestLoadMetadata_AllComponentTypes(t *testing.T) {
	testCases := []struct {
		componentType string
		lockFileName  string
	}{
		{"skills", ".skill-lock.json"},
		{"agents", ".agent-lock.json"},
		{"commands", ".command-lock.json"},
	}

	for _, tc := range testCases {
		t.Run(tc.componentType, func(t *testing.T) {
			// Create a temporary directory for testing
			tempDir, err := os.MkdirTemp("", "agent-smith-updater-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Create a sample lock file in the base directory
			lockFilePath := filepath.Join(tempDir, tc.lockFileName)
			lockFileContent := fmt.Sprintf(`{
  "version": 4,
  "%s": {
    "https://github.com/user/repo": {
      "test-component": {
        "source": "github.com/user/repo",
        "sourceType": "git",
        "sourceUrl": "https://github.com/user/repo",
        "commitHash": "def456",
        "installedAt": "2024-01-01T00:00:00Z",
        "updatedAt": "2024-01-01T00:00:00Z",
        "version": 3
      }
    }
  }
}`, tc.componentType)
			if err := os.WriteFile(lockFilePath, []byte(lockFileContent), 0644); err != nil {
				t.Fatalf("Failed to write lock file: %v", err)
			}

			// Create detector
			detector := NewUpdateDetectorWithBaseDir(tempDir)

			// Load metadata
			metadata, err := detector.LoadMetadata(tc.componentType, "test-component")
			if err != nil {
				t.Fatalf("Failed to load metadata for %s: %v", tc.componentType, err)
			}

			// Verify metadata
			if metadata.Name != "test-component" {
				t.Errorf("Expected name 'test-component', got '%s'", metadata.Name)
			}
			if metadata.Commit != "def456" {
				t.Errorf("Expected commit 'def456', got '%s'", metadata.Commit)
			}
		})
	}
}

// TestLoadMetadata_NoCommitHash tests behavior when commit hash is missing
func TestLoadMetadata_NoCommitHash(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-updater-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a lock file without commit hash (old format) in the base directory
	lockFilePath := filepath.Join(tempDir, ".skill-lock.json")
	lockFileContent := `{
  "version": 4,
  "skills": {
    "https://github.com/user/repo": {
      "test-skill": {
        "source": "github.com/user/repo",
        "sourceType": "git",
        "sourceUrl": "https://github.com/user/repo",
        "installedAt": "2024-01-01T00:00:00Z",
        "updatedAt": "2024-01-01T00:00:00Z",
        "version": 3
      }
    }
  }
}`
	if err := os.WriteFile(lockFilePath, []byte(lockFileContent), 0644); err != nil {
		t.Fatalf("Failed to write lock file: %v", err)
	}

	// Create detector
	detector := NewUpdateDetectorWithBaseDir(tempDir)

	// Load metadata
	metadata, err := detector.LoadMetadata("skills", "test-skill")
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	// Verify that commit hash is empty
	if metadata.Commit != "" {
		t.Errorf("Expected empty commit hash, got '%s'", metadata.Commit)
	}
}

// TestHasUpdates_NoCommitHash tests that HasUpdates returns true when no commit hash is stored
func TestHasUpdates_NoCommitHash(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-updater-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a lock file without commit hash (old format)
	lockFilePath := filepath.Join(tempDir, ".skill-lock.json")
	lockFileContent := `{
  "version": 4,
  "skills": {
    "https://github.com/user/repo": {
      "test-skill": {
        "source": "github.com/user/repo",
        "sourceType": "git",
        "sourceUrl": "https://github.com/user/repo",
        "installedAt": "2024-01-01T00:00:00Z",
        "updatedAt": "2024-01-01T00:00:00Z",
        "version": 3
      }
    }
  }
}`
	if err := os.WriteFile(lockFilePath, []byte(lockFileContent), 0644); err != nil {
		t.Fatalf("Failed to write lock file: %v", err)
	}

	// Create detector
	detector := NewUpdateDetectorWithBaseDir(tempDir)

	// Check for updates (should return true since no commit hash is stored)
	hasUpdates, err := detector.HasUpdates("skills", "test-skill", "https://github.com/user/repo")
	if err != nil {
		t.Fatalf("Failed to check for updates: %v", err)
	}

	if !hasUpdates {
		t.Error("Expected HasUpdates to return true when no commit hash is stored")
	}
}

// TestHasUpdates_MissingMetadata tests that HasUpdates returns error when metadata is missing
func TestHasUpdates_MissingMetadata(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-updater-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create an empty lock file
	lockFilePath := filepath.Join(tempDir, ".skill-lock.json")
	lockFileContent := `{
  "version": 3,
  "skills": {}
}`
	if err := os.WriteFile(lockFilePath, []byte(lockFileContent), 0644); err != nil {
		t.Fatalf("Failed to write lock file: %v", err)
	}

	// Create detector
	detector := NewUpdateDetectorWithBaseDir(tempDir)

	// Check for updates (should return error since component doesn't exist)
	_, err = detector.HasUpdates("skills", "nonexistent-skill", "https://github.com/user/repo")
	if err == nil {
		t.Fatal("Expected error when checking updates for non-existent component, got nil")
	}
}

// TestUpdateAll_NoComponents tests UpdateAll behavior when no components are installed
func TestUpdateAll_NoComponents(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-updater-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create detector
	detector := NewUpdateDetectorWithBaseDir(tempDir)

	// Update all (should complete without errors even with no components)
	err = detector.UpdateAll()
	if err != nil {
		t.Fatalf("UpdateAll failed: %v", err)
	}
}

// TestUpdateAll_EmptyDirectories tests UpdateAll behavior with empty component directories
func TestUpdateAll_EmptyDirectories(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-updater-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create empty component directories
	for _, componentType := range []string{"skills", "agents", "commands"} {
		typeDir := filepath.Join(tempDir, componentType)
		if err := os.MkdirAll(typeDir, 0755); err != nil {
			t.Fatalf("Failed to create %s directory: %v", componentType, err)
		}
	}

	// Create detector
	detector := NewUpdateDetectorWithBaseDir(tempDir)

	// Update all (should complete without errors)
	err = detector.UpdateAll()
	if err != nil {
		t.Fatalf("UpdateAll failed: %v", err)
	}
}
