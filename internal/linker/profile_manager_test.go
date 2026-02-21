package linker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tjg184/agent-smith/internal/detector"
	"github.com/tjg184/agent-smith/pkg/config"
)

// mockProfileManager implements the ProfileManager interface for testing
type mockProfileManager struct {
	profiles      []*Profile
	activeProfile string
	scanErr       error
	activeErr     error
}

func (m *mockProfileManager) ScanProfiles() ([]*Profile, error) {
	if m.scanErr != nil {
		return nil, m.scanErr
	}
	return m.profiles, nil
}

func (m *mockProfileManager) GetActiveProfile() (string, error) {
	if m.activeErr != nil {
		return "", m.activeErr
	}
	return m.activeProfile, nil
}

// TestNewComponentLinker_WithProfileManager verifies that ComponentLinker accepts ProfileManager as a dependency
// This test ensures Story-003 is implemented: ComponentLinker can accept ProfileManager for multi-profile operations
func TestNewComponentLinker_WithProfileManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-pm-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create agents directory
	agentsDir := filepath.Join(tempDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents dir: %v", err)
	}

	// Create mock target
	targetDir := filepath.Join(tempDir, "target")
	targets := []config.Target{
		&mockTarget{name: "test-target", baseDir: targetDir},
	}

	// Create detector
	det := detector.NewRepositoryDetector()

	// Create mock ProfileManager
	pm := &mockProfileManager{
		profiles: []*Profile{
			{Name: "test-profile", BasePath: filepath.Join(tempDir, "profiles", "test-profile")},
		},
		activeProfile: "test-profile",
	}

	// Test 1: Create ComponentLinker with ProfileManager
	linker, err := NewComponentLinker(agentsDir, targets, det, pm)
	if err != nil {
		t.Fatalf("Failed to create ComponentLinker with ProfileManager: %v", err)
	}

	// Verify linker was created successfully
	if linker == nil {
		t.Fatal("Expected non-nil ComponentLinker")
	}

	// Verify profileManager is set
	if linker.profileManager == nil {
		t.Error("Expected profileManager to be set, but it's nil")
	}

	// Test 2: Create ComponentLinker without ProfileManager (backward compatibility)
	linker2, err := NewComponentLinker(agentsDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create ComponentLinker without ProfileManager: %v", err)
	}

	// Verify linker was created successfully
	if linker2 == nil {
		t.Fatal("Expected non-nil ComponentLinker")
	}

	// Verify profileManager is nil (backward compatibility)
	if linker2.profileManager != nil {
		t.Error("Expected profileManager to be nil for backward compatibility")
	}
}

// TestShowAllProfilesLinkStatus_WithoutProfileManager verifies that the method
// returns an appropriate error when ProfileManager is not available
func TestShowAllProfilesLinkStatus_WithoutProfileManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-pm-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create agents directory
	agentsDir := filepath.Join(tempDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents dir: %v", err)
	}

	// Create mock target
	targetDir := filepath.Join(tempDir, "target")
	targets := []config.Target{
		&mockTarget{name: "test-target", baseDir: targetDir},
	}

	// Create detector
	det := detector.NewRepositoryDetector()

	// Create ComponentLinker WITHOUT ProfileManager
	linker, err := NewComponentLinker(agentsDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create ComponentLinker: %v", err)
	}

	// Test: Calling ShowAllProfilesLinkStatus without ProfileManager should return error
	err = linker.ShowAllProfilesLinkStatus([]string{})
	if err == nil {
		t.Fatal("Expected error when calling ShowAllProfilesLinkStatus without ProfileManager")
	}

	// Verify error message
	expectedMsg := "profile manager not available"
	if err.Error() != "profile manager not available - this operation requires a profile manager" {
		t.Errorf("Expected error containing '%s', got: %v", expectedMsg, err)
	}
}

// TestShowAllProfilesLinkStatus_WithProfileManager verifies that the method
// works correctly when ProfileManager is available
func TestShowAllProfilesLinkStatus_WithProfileManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-pm-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create agents directory
	agentsDir := filepath.Join(tempDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents dir: %v", err)
	}

	// Create mock target
	targetDir := filepath.Join(tempDir, "target")
	targets := []config.Target{
		&mockTarget{name: "test-target", baseDir: targetDir},
	}

	// Create detector
	det := detector.NewRepositoryDetector()

	// Create mock ProfileManager with test data
	testProfileDir := filepath.Join(tempDir, "profiles", "test-profile")
	pm := &mockProfileManager{
		profiles: []*Profile{
			{
				Name:        "test-profile",
				BasePath:    testProfileDir,
				HasAgents:   true,
				HasSkills:   true,
				HasCommands: false,
			},
		},
		activeProfile: "test-profile",
	}

	// Create ComponentLinker WITH ProfileManager
	linker, err := NewComponentLinker(agentsDir, targets, det, pm)
	if err != nil {
		t.Fatalf("Failed to create ComponentLinker: %v", err)
	}

	// Test: Calling ShowAllProfilesLinkStatus with ProfileManager should not error
	// (Note: It might not find components, but it should run without the "profile manager not available" error)
	err = linker.ShowAllProfilesLinkStatus([]string{})
	if err != nil {
		// Check that it's not the "profile manager not available" error
		if err.Error() == "profile manager not available - this operation requires a profile manager" {
			t.Errorf("Expected method to work with ProfileManager, but got 'profile manager not available' error")
		}
		// Other errors (like "failed to get base directory") are acceptable in this test
		// since we're just testing dependency injection
	}
}

// TestProfileManagerInterface verifies that the ProfileManager interface
// is correctly defined and can be implemented
func TestProfileManagerInterface(t *testing.T) {
	// Test that mockProfileManager implements ProfileManager interface
	var _ ProfileManager = (*mockProfileManager)(nil)

	// Create an instance
	pm := &mockProfileManager{
		profiles: []*Profile{
			{Name: "test", BasePath: "/tmp/test"},
		},
		activeProfile: "test",
	}

	// Test ScanProfiles
	profiles, err := pm.ScanProfiles()
	if err != nil {
		t.Errorf("Unexpected error from ScanProfiles: %v", err)
	}
	if len(profiles) != 1 {
		t.Errorf("Expected 1 profile, got %d", len(profiles))
	}

	// Test GetActiveProfile
	active, err := pm.GetActiveProfile()
	if err != nil {
		t.Errorf("Unexpected error from GetActiveProfile: %v", err)
	}
	if active != "test" {
		t.Errorf("Expected active profile 'test', got '%s'", active)
	}
}
