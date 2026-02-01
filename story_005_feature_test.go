//go:build integration
// +build integration

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestStory005FeatureDemo demonstrates the complete Story-005 feature:
// Installing component versions to experimental profiles while keeping stable versions in production
func TestStory005FeatureDemo(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-story-005-demo-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Build agent-smith binary
	binaryPath := filepath.Join(tempDir, "agent-smith")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build agent-smith: %v\nOutput: %s", err, string(output))
	}

	// Set HOME to test directory to avoid affecting actual configuration
	oldHome := os.Getenv("HOME")
	testHome := tempDir
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", oldHome)

	// Create a mock git repository with test components
	mockRepoDir := filepath.Join(tempDir, "mock-repo")
	if err := os.MkdirAll(mockRepoDir, 0755); err != nil {
		t.Fatalf("Failed to create mock repo: %v", err)
	}

	// Initialize git repo
	exec.Command("git", "init", mockRepoDir).Run()
	exec.Command("git", "-C", mockRepoDir, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", mockRepoDir, "config", "user.name", "Test User").Run()

	// Create a test skill
	skillDir := filepath.Join(mockRepoDir, "skills", "test-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create skill directory: %v", err)
	}

	skillContent := `# Test Skill

This is a test skill for demonstrating Story-005 functionality.
`
	skillMdPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillMdPath, []byte(skillContent), 0644); err != nil {
		t.Fatalf("Failed to write SKILL.md: %v", err)
	}

	// Commit the skill
	exec.Command("git", "-C", mockRepoDir, "add", ".").Run()
	exec.Command("git", "-C", mockRepoDir, "commit", "-m", "Initial commit").Run()

	t.Run("Scenario: Developer wants to test experimental versions separately", func(t *testing.T) {
		// Step 1: Create production profile
		t.Log("Step 1: Creating production profile...")
		createProdCmd := exec.Command(binaryPath, "profile", "create", "production")
		if output, err := createProdCmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to create production profile: %v\nOutput: %s", err, string(output))
		}

		// Step 2: Install stable version to production profile
		t.Log("Step 2: Installing stable version to production profile...")
		installProdCmd := exec.Command(binaryPath, "install", "skill", mockRepoDir, "test-skill", "--profile", "production")
		if output, err := installProdCmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to install to production: %v\nOutput: %s", err, string(output))
		}

		// Verify production installation
		prodProfileDir := filepath.Join(testHome, ".agent-smith", "profiles", "production", "skills", "test-skill")
		if _, err := os.Stat(prodProfileDir); os.IsNotExist(err) {
			t.Errorf("Production skill not installed at expected location: %s", prodProfileDir)
		}
		t.Logf("✓ Verified: Stable version installed in production profile at %s", prodProfileDir)

		// Step 3: Create experimental profile
		t.Log("Step 3: Creating experimental profile...")
		createExpCmd := exec.Command(binaryPath, "profile", "create", "experimental")
		if output, err := createExpCmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to create experimental profile: %v\nOutput: %s", err, string(output))
		}

		// Step 4: Simulate updating the skill in the repo (for experimental testing)
		t.Log("Step 4: Updating skill in repository (simulating new version)...")
		updatedSkillContent := skillContent + "\n## New Feature\n\nThis is an experimental new feature!\n"
		if err := os.WriteFile(skillMdPath, []byte(updatedSkillContent), 0644); err != nil {
			t.Fatalf("Failed to update SKILL.md: %v", err)
		}
		exec.Command("git", "-C", mockRepoDir, "add", ".").Run()
		exec.Command("git", "-C", mockRepoDir, "commit", "-m", "Add experimental feature").Run()

		// Step 5: Install experimental version to experimental profile
		t.Log("Step 5: Installing experimental version to experimental profile...")
		installExpCmd := exec.Command(binaryPath, "install", "skill", mockRepoDir, "test-skill", "--profile", "experimental")
		if output, err := installExpCmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to install to experimental: %v\nOutput: %s", err, string(output))
		}

		// Verify experimental installation
		expProfileDir := filepath.Join(testHome, ".agent-smith", "profiles", "experimental", "skills", "test-skill")
		if _, err := os.Stat(expProfileDir); os.IsNotExist(err) {
			t.Errorf("Experimental skill not installed at expected location: %s", expProfileDir)
		}
		t.Logf("✓ Verified: Experimental version installed in experimental profile at %s", expProfileDir)

		// Step 6: Verify both versions exist independently
		t.Log("Step 6: Verifying both versions exist independently...")

		// Read production version
		prodContent, err := os.ReadFile(filepath.Join(prodProfileDir, "SKILL.md"))
		if err != nil {
			t.Fatalf("Failed to read production skill: %v", err)
		}

		// Read experimental version
		expContent, err := os.ReadFile(filepath.Join(expProfileDir, "SKILL.md"))
		if err != nil {
			t.Fatalf("Failed to read experimental skill: %v", err)
		}

		// Verify they are different
		if strings.Contains(string(prodContent), "New Feature") {
			t.Error("Production version incorrectly contains experimental feature")
		}
		if !strings.Contains(string(expContent), "New Feature") {
			t.Error("Experimental version missing new feature")
		}
		t.Log("✓ Verified: Production has stable version, experimental has new version")

		// Step 7: Test switching between profiles
		t.Log("Step 7: Testing profile switching...")

		// Check current status (production should be auto-activated)
		statusCmd := exec.Command(binaryPath, "status")
		output, err := statusCmd.CombinedOutput()
		if err != nil {
			t.Logf("Status command output: %s", string(output))
		}
		if !strings.Contains(string(output), "production") {
			t.Errorf("Status should show 'production' as active profile (auto-activated), got: %s", string(output))
		}
		t.Log("✓ Verified: Production profile was auto-activated on first install")

		// Activate experimental profile
		activateExpCmd := exec.Command(binaryPath, "profile", "activate", "experimental")
		if output, err := activateExpCmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to activate experimental profile: %v\nOutput: %s", err, string(output))
		}

		// Check status shows experimental as active
		statusCmd2 := exec.Command(binaryPath, "status")
		output2, err := statusCmd2.CombinedOutput()
		if err != nil {
			t.Logf("Status command output: %s", string(output2))
		}
		if !strings.Contains(string(output2), "experimental") {
			t.Errorf("Status should show 'experimental' as active profile, got: %s", string(output2))
		}
		t.Log("✓ Verified: Experimental profile activated successfully")

		// Step 8: List profiles to see both
		t.Log("Step 8: Listing all profiles...")
		listCmd := exec.Command(binaryPath, "profile", "list")
		listOutput, err := listCmd.CombinedOutput()
		if err != nil {
			t.Logf("List command output: %s", string(listOutput))
		}

		if !strings.Contains(string(listOutput), "production") {
			t.Error("Profile list should contain 'production'")
		}
		if !strings.Contains(string(listOutput), "experimental") {
			t.Error("Profile list should contain 'experimental'")
		}
		t.Log("✓ Verified: Both profiles listed successfully")

		// Step 9: Show detailed profile info
		t.Log("Step 9: Showing profile details...")
		showProdCmd := exec.Command(binaryPath, "profile", "show", "production")
		showProdOutput, err := showProdCmd.CombinedOutput()
		if err != nil {
			t.Logf("Show production output: %s", string(showProdOutput))
		}

		if !strings.Contains(string(showProdOutput), "test-skill") {
			t.Error("Production profile details should show test-skill")
		}
		t.Log("✓ Verified: Profile details display correctly")

		t.Log("\n=== Story-005 Feature Demo Complete ===")
		t.Log("✓ Developer can install stable versions to production profile")
		t.Log("✓ Developer can install experimental versions to experimental profile")
		t.Log("✓ Both versions exist independently")
		t.Log("✓ Developer can switch between profiles")
		t.Log("✓ Developer can list and inspect profiles")
	})
}
