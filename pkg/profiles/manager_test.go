package profiles

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tgaines/agent-smith/pkg/paths"
)

func TestProfileManager_ScanProfiles(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	// Create profiles directory structure
	profilesDir := filepath.Join(tempDir, "profiles")

	// Create valid profile with agents directory
	codingProfile := filepath.Join(profilesDir, "coding")
	if err := os.MkdirAll(filepath.Join(codingProfile, paths.AgentsSubDir), 0755); err != nil {
		t.Fatalf("Failed to create coding profile: %v", err)
	}

	// Create valid profile with skills directory
	workProfile := filepath.Join(profilesDir, "work")
	if err := os.MkdirAll(filepath.Join(workProfile, paths.SkillsSubDir), 0755); err != nil {
		t.Fatalf("Failed to create work profile: %v", err)
	}

	// Create valid profile with all component directories
	fullProfile := filepath.Join(profilesDir, "full")
	if err := os.MkdirAll(filepath.Join(fullProfile, paths.AgentsSubDir), 0755); err != nil {
		t.Fatalf("Failed to create full profile agents: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(fullProfile, paths.SkillsSubDir), 0755); err != nil {
		t.Fatalf("Failed to create full profile skills: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(fullProfile, paths.CommandsSubDir), 0755); err != nil {
		t.Fatalf("Failed to create full profile commands: %v", err)
	}

	// Create invalid profile (no component directories)
	emptyProfile := filepath.Join(profilesDir, "empty")
	if err := os.MkdirAll(emptyProfile, 0755); err != nil {
		t.Fatalf("Failed to create empty profile: %v", err)
	}

	// Create hidden directory (should be ignored)
	hiddenProfile := filepath.Join(profilesDir, ".hidden")
	if err := os.MkdirAll(filepath.Join(hiddenProfile, paths.AgentsSubDir), 0755); err != nil {
		t.Fatalf("Failed to create hidden profile: %v", err)
	}

	// Create a file in profiles directory (should be ignored)
	testFile := filepath.Join(profilesDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create ProfileManager with custom profiles directory
	pm := &ProfileManager{profilesDir: profilesDir}

	// Scan profiles
	profiles, err := pm.ScanProfiles()
	if err != nil {
		t.Fatalf("ScanProfiles() error = %v", err)
	}

	// Should find exactly 3 valid profiles (coding, work, full)
	if len(profiles) != 3 {
		t.Errorf("Expected 3 valid profiles, got %d", len(profiles))
	}

	// Verify profiles are correctly loaded
	profileMap := make(map[string]*Profile)
	for _, p := range profiles {
		profileMap[p.Name] = p
	}

	// Check coding profile
	if coding, ok := profileMap["coding"]; !ok {
		t.Error("coding profile not found")
	} else {
		if !coding.HasAgents {
			t.Error("coding profile should have agents")
		}
		if coding.HasSkills || coding.HasCommands {
			t.Error("coding profile should not have skills or commands")
		}
	}

	// Check work profile
	if work, ok := profileMap["work"]; !ok {
		t.Error("work profile not found")
	} else {
		if !work.HasSkills {
			t.Error("work profile should have skills")
		}
		if work.HasAgents || work.HasCommands {
			t.Error("work profile should not have agents or commands")
		}
	}

	// Check full profile
	if full, ok := profileMap["full"]; !ok {
		t.Error("full profile not found")
	} else {
		if !full.HasAgents || !full.HasSkills || !full.HasCommands {
			t.Error("full profile should have all component directories")
		}
	}

	// Verify empty, hidden, and file are not included
	if _, ok := profileMap["empty"]; ok {
		t.Error("empty profile should not be included")
	}
	if _, ok := profileMap[".hidden"]; ok {
		t.Error("hidden profile should not be included")
	}
	if _, ok := profileMap["test.txt"]; ok {
		t.Error("test.txt file should not be included")
	}
}

func TestProfileManager_ScanProfiles_NoProfilesDirectory(t *testing.T) {
	// Create temporary directory but don't create profiles subdirectory
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")

	// Create ProfileManager with non-existent profiles directory
	pm := &ProfileManager{profilesDir: profilesDir}

	// Scan profiles should return empty list without error
	profiles, err := pm.ScanProfiles()
	if err != nil {
		t.Fatalf("ScanProfiles() error = %v", err)
	}

	if len(profiles) != 0 {
		t.Errorf("Expected 0 profiles, got %d", len(profiles))
	}
}

func TestProfileManager_loadProfile(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")

	// Create test profile with only skills
	testProfile := filepath.Join(profilesDir, "test")
	if err := os.MkdirAll(filepath.Join(testProfile, paths.SkillsSubDir), 0755); err != nil {
		t.Fatalf("Failed to create test profile: %v", err)
	}

	pm := &ProfileManager{profilesDir: profilesDir}

	// Load profile
	profile := pm.loadProfile("test")

	if profile.Name != "test" {
		t.Errorf("Expected profile name 'test', got '%s'", profile.Name)
	}

	if profile.BasePath != testProfile {
		t.Errorf("Expected base path '%s', got '%s'", testProfile, profile.BasePath)
	}

	if profile.HasAgents || profile.HasCommands {
		t.Error("Profile should not have agents or commands")
	}

	if !profile.HasSkills {
		t.Error("Profile should have skills")
	}

	if !profile.IsValid() {
		t.Error("Profile should be valid")
	}
}

func TestProfileManager_CreateProfile(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")

	// Create ProfileManager with custom profiles directory
	pm := &ProfileManager{profilesDir: profilesDir}

	// Create a new profile
	err := pm.CreateProfile("test-profile")
	if err != nil {
		t.Fatalf("CreateProfile() error = %v", err)
	}

	// Verify profile directory exists
	profileDir := filepath.Join(profilesDir, "test-profile")
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		t.Error("Profile directory was not created")
	}

	// Verify all component directories exist
	agentsDir := filepath.Join(profileDir, paths.AgentsSubDir)
	skillsDir := filepath.Join(profileDir, paths.SkillsSubDir)
	commandsDir := filepath.Join(profileDir, paths.CommandsSubDir)

	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		t.Error("Agents directory was not created")
	}
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		t.Error("Skills directory was not created")
	}
	if _, err := os.Stat(commandsDir); os.IsNotExist(err) {
		t.Error("Commands directory was not created")
	}

	// Verify the profile is valid and can be loaded
	profile := pm.loadProfile("test-profile")
	if !profile.IsValid() {
		t.Error("Created profile should be valid")
	}
	if !profile.HasAgents || !profile.HasSkills || !profile.HasCommands {
		t.Error("Created profile should have all component directories")
	}
}

func TestProfileManager_CreateProfile_AlreadyExists(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")

	// Create ProfileManager with custom profiles directory
	pm := &ProfileManager{profilesDir: profilesDir}

	// Create a profile
	err := pm.CreateProfile("existing-profile")
	if err != nil {
		t.Fatalf("CreateProfile() error = %v", err)
	}

	// Try to create the same profile again
	err = pm.CreateProfile("existing-profile")
	if err == nil {
		t.Error("CreateProfile() should return error for existing profile")
	}
}

func TestProfileManager_CreateProfile_EmptyName(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")

	// Create ProfileManager with custom profiles directory
	pm := &ProfileManager{profilesDir: profilesDir}

	// Try to create a profile with empty name
	err := pm.CreateProfile("")
	if err == nil {
		t.Error("CreateProfile() should return error for empty profile name")
	}
}

func TestProfileManager_DeleteProfile(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")

	// Create ProfileManager with custom profiles directory
	pm := &ProfileManager{profilesDir: profilesDir}

	// Create a new profile
	err := pm.CreateProfile("test-profile")
	if err != nil {
		t.Fatalf("CreateProfile() error = %v", err)
	}

	// Verify profile exists
	profileDir := filepath.Join(profilesDir, "test-profile")
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		t.Error("Profile directory was not created")
	}

	// Delete the profile
	err = pm.DeleteProfile("test-profile")
	if err != nil {
		t.Fatalf("DeleteProfile() error = %v", err)
	}

	// Verify profile no longer exists
	if _, err := os.Stat(profileDir); !os.IsNotExist(err) {
		t.Error("Profile directory should have been deleted")
	}

	// Verify it's not in the list of profiles
	profiles, err := pm.ScanProfiles()
	if err != nil {
		t.Fatalf("ScanProfiles() error = %v", err)
	}

	for _, p := range profiles {
		if p.Name == "test-profile" {
			t.Error("Deleted profile should not appear in scan results")
		}
	}
}

func TestProfileManager_DeleteProfile_NonExistent(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")

	// Create ProfileManager with custom profiles directory
	pm := &ProfileManager{profilesDir: profilesDir}

	// Try to delete a non-existent profile
	err := pm.DeleteProfile("non-existent")
	if err == nil {
		t.Error("DeleteProfile() should return error for non-existent profile")
	}
}

func TestProfileManager_DeleteProfile_EmptyName(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")

	// Create ProfileManager with custom profiles directory
	pm := &ProfileManager{profilesDir: profilesDir}

	// Try to delete a profile with empty name
	err := pm.DeleteProfile("")
	if err == nil {
		t.Error("DeleteProfile() should return error for empty profile name")
	}
}

func TestValidateProfileName(t *testing.T) {
	tests := []struct {
		name      string
		inputName string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid simple name",
			inputName: "myprofile",
			wantError: false,
		},
		{
			name:      "valid name with hyphen",
			inputName: "my-profile",
			wantError: false,
		},
		{
			name:      "valid name with numbers",
			inputName: "profile123",
			wantError: false,
		},
		{
			name:      "valid name with mixed case",
			inputName: "MyProfile",
			wantError: false,
		},
		{
			name:      "valid name complex",
			inputName: "My-Profile-123",
			wantError: false,
		},
		{
			name:      "empty name",
			inputName: "",
			wantError: true,
			errorMsg:  "profile name cannot be empty",
		},
		{
			name:      "hidden directory",
			inputName: ".hidden",
			wantError: true,
			errorMsg:  "profile name cannot start with '.'",
		},
		{
			name:      "name with forward slash",
			inputName: "my/profile",
			wantError: true,
			errorMsg:  "profile name cannot contain path separators",
		},
		{
			name:      "name with backslash",
			inputName: "my\\profile",
			wantError: true,
			errorMsg:  "profile name cannot contain path separators",
		},
		{
			name:      "path traversal with double dots",
			inputName: "../profile",
			wantError: true,
			errorMsg:  "profile name cannot contain path traversal patterns",
		},
		{
			name:      "path traversal with dot slash",
			inputName: "./profile",
			wantError: true,
			errorMsg:  "profile name cannot contain path traversal patterns",
		},
		{
			name:      "name with spaces",
			inputName: "my profile",
			wantError: true,
			errorMsg:  "profile name must contain only letters, numbers, and hyphens",
		},
		{
			name:      "name with underscore",
			inputName: "my_profile",
			wantError: true,
			errorMsg:  "profile name must contain only letters, numbers, and hyphens",
		},
		{
			name:      "name with special characters",
			inputName: "my@profile",
			wantError: true,
			errorMsg:  "profile name must contain only letters, numbers, and hyphens",
		},
		{
			name:      "name with dollar sign",
			inputName: "my$profile",
			wantError: true,
			errorMsg:  "profile name must contain only letters, numbers, and hyphens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProfileName(tt.inputName)
			if tt.wantError {
				if err == nil {
					t.Errorf("validateProfileName(%q) expected error containing %q, got nil", tt.inputName, tt.errorMsg)
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("validateProfileName(%q) error = %q, want error containing %q", tt.inputName, err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateProfileName(%q) unexpected error = %v", tt.inputName, err)
				}
			}
		})
	}
}

func TestProfileManager_CreateProfile_InvalidNames(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")

	// Create ProfileManager with custom profiles directory
	pm := &ProfileManager{profilesDir: profilesDir}

	invalidNames := []string{
		".hidden",
		"../etc",
		"./profile",
		"my/profile",
		"my\\profile",
		"my profile",
		"my_profile",
		"profile@123",
	}

	for _, name := range invalidNames {
		t.Run(name, func(t *testing.T) {
			err := pm.CreateProfile(name)
			if err == nil {
				t.Errorf("CreateProfile(%q) should return error for invalid name", name)
			}
		})
	}
}

func TestProfileManager_DeleteProfile_InvalidNames(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")

	// Create ProfileManager with custom profiles directory
	pm := &ProfileManager{profilesDir: profilesDir}

	invalidNames := []string{
		".hidden",
		"../etc",
		"./profile",
		"my/profile",
		"my\\profile",
		"my profile",
	}

	for _, name := range invalidNames {
		t.Run(name, func(t *testing.T) {
			err := pm.DeleteProfile(name)
			if err == nil {
				t.Errorf("DeleteProfile(%q) should return error for invalid name", name)
			}
		})
	}
}

func TestProfileManager_ActivateProfile_InvalidNames(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")

	// Create ProfileManager with custom profiles directory
	pm := &ProfileManager{profilesDir: profilesDir}

	invalidNames := []string{
		".hidden",
		"../etc",
		"./profile",
		"my/profile",
		"my\\profile",
		"my profile",
	}

	for _, name := range invalidNames {
		t.Run(name, func(t *testing.T) {
			err := pm.ActivateProfile(name)
			if err == nil {
				t.Errorf("ActivateProfile(%q) should return error for invalid name", name)
			}
		})
	}
}

func TestGenerateProfileNameFromRepo(t *testing.T) {
	tests := []struct {
		name             string
		repoURL          string
		existingProfiles []string
		expected         string
	}{
		{
			name:             "GitHub URL",
			repoURL:          "https://github.com/owner/repo",
			existingProfiles: []string{},
			expected:         "owner-repo",
		},
		{
			name:             "GitHub URL with .git suffix",
			repoURL:          "https://github.com/owner/repo.git",
			existingProfiles: []string{},
			expected:         "owner-repo",
		},
		{
			name:             "GitHub URL with trailing slash",
			repoURL:          "https://github.com/owner/repo/",
			existingProfiles: []string{},
			expected:         "owner-repo",
		},
		{
			name:             "GitLab URL",
			repoURL:          "https://gitlab.com/owner/repo",
			existingProfiles: []string{},
			expected:         "owner-repo",
		},
		{
			name:             "Bitbucket URL",
			repoURL:          "https://bitbucket.org/owner/repo",
			existingProfiles: []string{},
			expected:         "owner-repo",
		},
		{
			name:             "SSH GitHub URL",
			repoURL:          "git@github.com:owner/repo.git",
			existingProfiles: []string{},
			expected:         "owner-repo",
		},
		{
			name:             "Local absolute path",
			repoURL:          "/home/user/repos/my-repo",
			existingProfiles: []string{},
			expected:         "my-repo",
		},
		{
			name:             "Local relative path",
			repoURL:          "./my-repo",
			existingProfiles: []string{},
			expected:         "my-repo",
		},
		{
			name:             "Local relative path with parent",
			repoURL:          "../my-repo",
			existingProfiles: []string{},
			expected:         "my-repo",
		},
		{
			name:             "Profile name collision - adds hash",
			repoURL:          "https://github.com/owner/repo",
			existingProfiles: []string{"owner-repo"},
			expected:         "owner-repo-", // Will have hash suffix
		},
		{
			name:             "Profile name with special characters sanitized",
			repoURL:          "https://github.com/owner/repo_with_underscores",
			existingProfiles: []string{},
			expected:         "owner-repo-with-underscores",
		},
		{
			name:             "Multiple collisions",
			repoURL:          "https://github.com/owner/repo",
			existingProfiles: []string{"owner-repo", "owner-repo-abc123"},
			expected:         "owner-repo-", // Will have different hash
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateProfileNameFromRepo(tt.repoURL, tt.existingProfiles)

			// For collision tests, we check prefix instead of exact match
			if len(tt.existingProfiles) > 0 {
				if !hasPrefix(result, tt.expected) {
					t.Errorf("GenerateProfileNameFromRepo() = %v, want prefix %v", result, tt.expected)
				}
				// Verify hash suffix was added (6 characters)
				if len(result) != len(tt.expected)+6 && !hasPrefix(result, tt.expected) {
					t.Errorf("GenerateProfileNameFromRepo() with collision should add 6-char hash, got %v", result)
				}
			} else {
				if result != tt.expected {
					t.Errorf("GenerateProfileNameFromRepo() = %v, want %v", result, tt.expected)
				}
			}

			// Verify result is a valid profile name
			if err := validateProfileName(result); err != nil {
				t.Errorf("GenerateProfileNameFromRepo() produced invalid profile name: %v", err)
			}
		})
	}
}

func TestGenerateProfileNameFromRepo_Uniqueness(t *testing.T) {
	// Test that collision resolution produces unique names
	repoURL := "https://github.com/owner/repo"
	existingProfiles := []string{}

	// Generate first profile name
	profile1 := GenerateProfileNameFromRepo(repoURL, existingProfiles)
	if profile1 != "owner-repo" {
		t.Errorf("First profile name should be 'owner-repo', got %v", profile1)
	}

	// Generate second profile name (with collision)
	existingProfiles = append(existingProfiles, profile1)
	profile2 := GenerateProfileNameFromRepo(repoURL, existingProfiles)
	if profile2 == profile1 {
		t.Errorf("Second profile name should be different from first, both are %v", profile1)
	}

	// Generate third profile name (with two collisions)
	existingProfiles = append(existingProfiles, profile2)
	profile3 := GenerateProfileNameFromRepo(repoURL, existingProfiles)
	if profile3 == profile1 || profile3 == profile2 {
		t.Errorf("Third profile name should be unique, got %v", profile3)
	}
}

func TestSanitizeForProfileName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean name",
			input:    "myrepo",
			expected: "myrepo",
		},
		{
			name:     "name with underscores",
			input:    "my_repo",
			expected: "my-repo",
		},
		{
			name:     "name with spaces",
			input:    "my repo",
			expected: "my-repo",
		},
		{
			name:     "name with special characters",
			input:    "my@repo!",
			expected: "my-repo",
		},
		{
			name:     "name with multiple consecutive special chars",
			input:    "my___repo",
			expected: "my-repo",
		},
		{
			name:     "name with leading/trailing hyphens",
			input:    "-myrepo-",
			expected: "myrepo",
		},
		{
			name:     "empty after sanitization",
			input:    "!!!",
			expected: "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeForProfileName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeForProfileName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper function to check if a string has a prefix
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
