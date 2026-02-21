package profiles

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tjg184/agent-smith/pkg/paths"
)

// Note: Profile activation/deactivation tests that require accessing the system
// paths (~/.agent-smith) are tested in integration tests. These package-level
// tests focus on validation logic and component counting/listing operations.

// TestProfileActivation_ValidationLogic tests profile name validation
func TestProfileActivation_ValidationLogic(t *testing.T) {
	tests := []struct {
		name          string
		profileName   string
		shouldError   bool
		errorContains string
	}{
		{
			name:          "empty profile name",
			profileName:   "",
			shouldError:   true,
			errorContains: "cannot be empty",
		},
		{
			name:          "hidden directory",
			profileName:   ".hidden",
			shouldError:   true,
			errorContains: "cannot start with '.'",
		},
		{
			name:          "path traversal",
			profileName:   "../etc",
			shouldError:   true,
			errorContains: "path traversal",
		},
		{
			name:          "path separator slash",
			profileName:   "my/profile",
			shouldError:   true,
			errorContains: "path separators",
		},
		{
			name:          "path separator backslash",
			profileName:   "my\\profile",
			shouldError:   true,
			errorContains: "path separators",
		},
		{
			name:        "valid simple name",
			profileName: "myprofile",
			shouldError: false,
		},
		{
			name:        "valid name with hyphen",
			profileName: "my-profile",
			shouldError: false,
		},
		{
			name:        "valid name with numbers",
			profileName: "profile123",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProfileName(tt.profileName)
			if tt.shouldError {
				if err == nil {
					t.Errorf("validateProfileName(%q) should return error", tt.profileName)
				}
			} else {
				if err != nil {
					t.Errorf("validateProfileName(%q) unexpected error: %v", tt.profileName, err)
				}
			}
		})
	}
}

// TestProfileActivation_NonExistentProfile tests that loading a non-existent
// profile returns an invalid profile
func TestProfileActivation_NonExistentProfile(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")

	pm := &ProfileManager{
		profilesDir: profilesDir,
		linker:      nil,
	}

	// Load a non-existent profile - should return invalid
	profile := pm.loadProfile("non-existent")
	if profile.IsValid() {
		t.Error("Non-existent profile should not be valid")
	}

	// Verify profile has correct name but no components
	if profile.Name != "non-existent" {
		t.Errorf("Profile name = %s, want non-existent", profile.Name)
	}
	if profile.HasAgents || profile.HasSkills || profile.HasCommands {
		t.Error("Non-existent profile should have no components")
	}
}

// TestProfileActivation_CountComponents tests that CountComponents correctly
// counts agents, skills, and commands in a profile
func TestProfileActivation_CountComponents(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")

	pm := &ProfileManager{
		profilesDir: profilesDir,
	}

	// Create a test profile
	profileName := "test-profile"
	profileDir := filepath.Join(profilesDir, profileName)

	// Create component directories with test components
	agentsDir := filepath.Join(profileDir, paths.AgentsSubDir)
	if err := os.MkdirAll(filepath.Join(agentsDir, "agent1"), 0755); err != nil {
		t.Fatalf("Failed to create agent1: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(agentsDir, "agent2"), 0755); err != nil {
		t.Fatalf("Failed to create agent2: %v", err)
	}

	skillsDir := filepath.Join(profileDir, paths.SkillsSubDir)
	if err := os.MkdirAll(filepath.Join(skillsDir, "skill1"), 0755); err != nil {
		t.Fatalf("Failed to create skill1: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(skillsDir, "skill2"), 0755); err != nil {
		t.Fatalf("Failed to create skill2: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(skillsDir, "skill3"), 0755); err != nil {
		t.Fatalf("Failed to create skill3: %v", err)
	}

	commandsDir := filepath.Join(profileDir, paths.CommandsSubDir)
	if err := os.MkdirAll(filepath.Join(commandsDir, "command1"), 0755); err != nil {
		t.Fatalf("Failed to create command1: %v", err)
	}

	// Create a hidden directory (should not be counted)
	if err := os.MkdirAll(filepath.Join(skillsDir, ".hidden"), 0755); err != nil {
		t.Fatalf("Failed to create .hidden: %v", err)
	}

	// Load the profile
	profile := pm.loadProfile(profileName)

	// Count components
	agents, skills, commands := pm.CountComponents(profile)

	// Verify counts
	expectedAgents := 2
	expectedSkills := 3
	expectedCommands := 1

	if agents != expectedAgents {
		t.Errorf("CountComponents agents = %d, want %d", agents, expectedAgents)
	}
	if skills != expectedSkills {
		t.Errorf("CountComponents skills = %d, want %d", skills, expectedSkills)
	}
	if commands != expectedCommands {
		t.Errorf("CountComponents commands = %d, want %d", commands, expectedCommands)
	}
}

// TestProfileActivation_CountComponentsEmptyProfile tests counting components
// in a profile with no components
func TestProfileActivation_CountComponentsEmptyProfile(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")

	pm := &ProfileManager{
		profilesDir: profilesDir,
	}

	// Create a test profile with empty component directories
	profileName := "empty-profile"
	profileDir := filepath.Join(profilesDir, profileName)

	// Create empty component directories
	if err := os.MkdirAll(filepath.Join(profileDir, paths.AgentsSubDir), 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(profileDir, paths.SkillsSubDir), 0755); err != nil {
		t.Fatalf("Failed to create skills directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(profileDir, paths.CommandsSubDir), 0755); err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	// Load the profile
	profile := pm.loadProfile(profileName)

	// Count components - should be zero
	agents, skills, commands := pm.CountComponents(profile)

	if agents != 0 {
		t.Errorf("CountComponents agents = %d, want 0", agents)
	}
	if skills != 0 {
		t.Errorf("CountComponents skills = %d, want 0", skills)
	}
	if commands != 0 {
		t.Errorf("CountComponents commands = %d, want 0", commands)
	}
}

// TestProfileActivation_GetComponentNames tests that GetComponentNames correctly
// retrieves component names from a profile
func TestProfileActivation_GetComponentNames(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")

	pm := &ProfileManager{
		profilesDir: profilesDir,
	}

	// Create a test profile
	profileName := "test-profile"
	profileDir := filepath.Join(profilesDir, profileName)

	// Create component directories with test components
	agentsDir := filepath.Join(profileDir, paths.AgentsSubDir)
	if err := os.MkdirAll(filepath.Join(agentsDir, "agent-a"), 0755); err != nil {
		t.Fatalf("Failed to create agent-a: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(agentsDir, "agent-b"), 0755); err != nil {
		t.Fatalf("Failed to create agent-b: %v", err)
	}

	skillsDir := filepath.Join(profileDir, paths.SkillsSubDir)
	if err := os.MkdirAll(filepath.Join(skillsDir, "skill-x"), 0755); err != nil {
		t.Fatalf("Failed to create skill-x: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(skillsDir, "skill-y"), 0755); err != nil {
		t.Fatalf("Failed to create skill-y: %v", err)
	}

	commandsDir := filepath.Join(profileDir, paths.CommandsSubDir)
	if err := os.MkdirAll(filepath.Join(commandsDir, "command-1"), 0755); err != nil {
		t.Fatalf("Failed to create command-1: %v", err)
	}

	// Create hidden directory (should not be included)
	if err := os.MkdirAll(filepath.Join(skillsDir, ".hidden"), 0755); err != nil {
		t.Fatalf("Failed to create .hidden: %v", err)
	}

	// Load the profile
	profile := pm.loadProfile(profileName)

	// Get component names
	agents, skills, commands := pm.GetComponentNames(profile)

	// Verify agent names
	expectedAgents := []string{"agent-a", "agent-b"}
	if len(agents) != len(expectedAgents) {
		t.Errorf("GetComponentNames agents count = %d, want %d", len(agents), len(expectedAgents))
	}
	for _, expected := range expectedAgents {
		found := false
		for _, actual := range agents {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected agent %s not found in %v", expected, agents)
		}
	}

	// Verify skill names
	expectedSkills := []string{"skill-x", "skill-y"}
	if len(skills) != len(expectedSkills) {
		t.Errorf("GetComponentNames skills count = %d, want %d", len(skills), len(expectedSkills))
	}
	for _, expected := range expectedSkills {
		found := false
		for _, actual := range skills {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected skill %s not found in %v", expected, skills)
		}
	}

	// Verify command names
	expectedCommands := []string{"command-1"}
	if len(commands) != len(expectedCommands) {
		t.Errorf("GetComponentNames commands count = %d, want %d", len(commands), len(expectedCommands))
	}
	for _, expected := range expectedCommands {
		found := false
		for _, actual := range commands {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected command %s not found in %v", expected, commands)
		}
	}

	// Verify hidden directory is not included
	for _, skill := range skills {
		if skill == ".hidden" {
			t.Error("Hidden directory should not be included in component names")
		}
	}
}

// TestProfileActivation_GetComponentNamesEmptyProfile tests that GetComponentNames
// returns empty slices for a profile with no components
func TestProfileActivation_GetComponentNamesEmptyProfile(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")

	pm := &ProfileManager{
		profilesDir: profilesDir,
	}

	// Create a test profile with empty component directories
	profileName := "empty-profile"
	profileDir := filepath.Join(profilesDir, profileName)

	// Create empty component directories
	if err := os.MkdirAll(filepath.Join(profileDir, paths.AgentsSubDir), 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(profileDir, paths.SkillsSubDir), 0755); err != nil {
		t.Fatalf("Failed to create skills directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(profileDir, paths.CommandsSubDir), 0755); err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	// Load the profile
	profile := pm.loadProfile(profileName)

	// Get component names - should be empty slices
	agents, skills, commands := pm.GetComponentNames(profile)

	if len(agents) != 0 {
		t.Errorf("GetComponentNames agents = %v, want empty slice", agents)
	}
	if len(skills) != 0 {
		t.Errorf("GetComponentNames skills = %v, want empty slice", skills)
	}
	if len(commands) != 0 {
		t.Errorf("GetComponentNames commands = %v, want empty slice", commands)
	}
}

// TestProfileActivation_ProfileDirectoryStructure tests that profiles have
// the correct directory structure after creation
func TestProfileActivation_ProfileDirectoryStructure(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")

	pm := &ProfileManager{
		profilesDir: profilesDir,
	}

	// Create a test profile
	profileName := "test-profile"
	if err := pm.CreateProfile(profileName); err != nil {
		t.Fatalf("CreateProfile failed: %v", err)
	}

	// Verify profile directory exists
	profileDir := filepath.Join(profilesDir, profileName)
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
	profile := pm.loadProfile(profileName)
	if !profile.IsValid() {
		t.Error("Created profile should be valid")
	}
	if !profile.HasAgents || !profile.HasSkills || !profile.HasCommands {
		t.Error("Created profile should have all component directories")
	}
}
