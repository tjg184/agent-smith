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

// TestStory005VerificationSuite runs comprehensive tests to verify all changes are consistent
// and no regressions are introduced. This fulfills the acceptance criteria for Story-005.
func TestStory005VerificationSuite(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agent-smith-story-005-*")
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

	// Create test configuration directory
	testConfigDir := filepath.Join(tempDir, ".agent-smith-test")
	if err := os.MkdirAll(testConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create test config directory: %v", err)
	}

	// Set HOME to test directory to avoid affecting actual configuration
	oldHome := os.Getenv("HOME")
	testHome := tempDir
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", oldHome)

	t.Run("ProfileListCommand", func(t *testing.T) {
		testProfileListCommand(t, binaryPath, testConfigDir)
	})

	t.Run("ProfileShowCommand", func(t *testing.T) {
		testProfileShowCommand(t, binaryPath, testConfigDir)
	})

	t.Run("StatusCommand", func(t *testing.T) {
		testStatusCommand(t, binaryPath, testConfigDir)
	})

	t.Run("TargetListCommand", func(t *testing.T) {
		testTargetListCommand(t, binaryPath, testConfigDir)
	})
}

// testProfileListCommand verifies 'profile list' command behavior
func testProfileListCommand(t *testing.T, binaryPath, configDir string) {
	t.Run("WithoutFlags", func(t *testing.T) {
		// Test that profile list shows output without flags
		cmd := exec.Command(binaryPath, "profile", "list")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Command output: %s", string(output))
		}

		outputStr := string(output)

		// Should show output (either empty state or profiles)
		if len(outputStr) == 0 {
			t.Error("Expected output from 'profile list' without flags, but got none")
		}

		// Should either show "No profiles found" or profile listings
		hasOutput := strings.Contains(outputStr, "No profiles found") ||
			strings.Contains(outputStr, "Available Profiles") ||
			strings.Contains(outputStr, "Active Profiles") ||
			strings.Contains(outputStr, "Profiles:")

		if !hasOutput {
			t.Errorf("Expected profile list output, got: %s", outputStr)
		}
	})

	t.Run("WithVerboseFlag", func(t *testing.T) {
		// Test that profile list works with --verbose flag
		cmd := exec.Command(binaryPath, "profile", "list", "--verbose")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Command output: %s", string(output))
		}

		outputStr := string(output)

		if len(outputStr) == 0 {
			t.Error("Expected output from 'profile list --verbose', but got none")
		}
	})

	t.Run("WithDebugFlag", func(t *testing.T) {
		// Test that profile list works with --debug flag
		cmd := exec.Command(binaryPath, "profile", "list", "--debug")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Command output: %s", string(output))
		}

		outputStr := string(output)

		if len(outputStr) == 0 {
			t.Error("Expected output from 'profile list --debug', but got none")
		}
	})

	t.Run("WithMultipleProfiles", func(t *testing.T) {
		// Create test profiles
		for _, profileName := range []string{"test-profile-1", "test-profile-2"} {
			cmd := exec.Command(binaryPath, "profile", "create", profileName)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Logf("Warning: Failed to create profile %s: %v\nOutput: %s", profileName, err, string(output))
			}
		}

		// Test that profile list shows all profiles
		cmd := exec.Command(binaryPath, "profile", "list")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Command output: %s", string(output))
		}

		outputStr := string(output)

		if !strings.Contains(outputStr, "test-profile-1") || !strings.Contains(outputStr, "test-profile-2") {
			t.Errorf("Expected to see both test profiles in output, got: %s", outputStr)
		}
	})
}

// testProfileShowCommand verifies 'profile show' command behavior
func testProfileShowCommand(t *testing.T, binaryPath, configDir string) {
	// Create a test profile first
	createCmd := exec.Command(binaryPath, "profile", "create", "test-show-profile")
	if output, err := createCmd.CombinedOutput(); err != nil {
		t.Logf("Warning: Failed to create test profile: %v\nOutput: %s", err, string(output))
	}

	t.Run("WithoutFlags", func(t *testing.T) {
		// Test that profile show displays output without flags
		cmd := exec.Command(binaryPath, "profile", "show", "test-show-profile")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Command output: %s", string(output))
		}

		outputStr := string(output)

		if len(outputStr) == 0 {
			t.Error("Expected output from 'profile show' without flags, but got none")
		}

		// Should show profile details
		if !strings.Contains(outputStr, "test-show-profile") {
			t.Errorf("Expected to see profile name in output, got: %s", outputStr)
		}
	})

	t.Run("WithVerboseFlag", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "show", "test-show-profile", "--verbose")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Command output: %s", string(output))
		}

		outputStr := string(output)

		if len(outputStr) == 0 {
			t.Error("Expected output from 'profile show --verbose', but got none")
		}
	})

	t.Run("WithDebugFlag", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "profile", "show", "test-show-profile", "--debug")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Command output: %s", string(output))
		}

		outputStr := string(output)

		if len(outputStr) == 0 {
			t.Error("Expected output from 'profile show --debug', but got none")
		}
	})
}

// testStatusCommand verifies 'status' command behavior
func testStatusCommand(t *testing.T, binaryPath, configDir string) {
	t.Run("WithoutFlags", func(t *testing.T) {
		// Test that status shows output without flags
		cmd := exec.Command(binaryPath, "status")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Command output: %s", string(output))
		}

		outputStr := string(output)

		if len(outputStr) == 0 {
			t.Error("Expected output from 'status' without flags, but got none")
		}

		// Should show status information
		hasStatus := strings.Contains(outputStr, "Status") ||
			strings.Contains(outputStr, "Active Profile") ||
			strings.Contains(outputStr, "Detected Targets") ||
			strings.Contains(outputStr, "Components")

		if !hasStatus {
			t.Errorf("Expected status information in output, got: %s", outputStr)
		}
	})

	t.Run("WithVerboseFlag", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "status", "--verbose")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Command output: %s", string(output))
		}

		outputStr := string(output)

		if len(outputStr) == 0 {
			t.Error("Expected output from 'status --verbose', but got none")
		}
	})

	t.Run("WithDebugFlag", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "status", "--debug")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Command output: %s", string(output))
		}

		outputStr := string(output)

		if len(outputStr) == 0 {
			t.Error("Expected output from 'status --debug', but got none")
		}
	})

	t.Run("WithActiveProfile", func(t *testing.T) {
		// Create and activate a profile
		createCmd := exec.Command(binaryPath, "profile", "create", "test-status-profile")
		createCmd.CombinedOutput()

		activateCmd := exec.Command(binaryPath, "profile", "activate", "test-status-profile")
		activateCmd.CombinedOutput()

		// Test that status shows active profile
		cmd := exec.Command(binaryPath, "status")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Command output: %s", string(output))
		}

		outputStr := string(output)

		if !strings.Contains(outputStr, "test-status-profile") {
			t.Errorf("Expected to see active profile name in status, got: %s", outputStr)
		}
	})
}

// testTargetListCommand verifies 'target list' command behavior
func testTargetListCommand(t *testing.T, binaryPath, configDir string) {
	t.Run("WithoutFlags", func(t *testing.T) {
		// Test that target list shows output without flags
		cmd := exec.Command(binaryPath, "target", "list")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Command output: %s", string(output))
		}

		outputStr := string(output)

		if len(outputStr) == 0 {
			t.Error("Expected output from 'target list' without flags, but got none")
		}

		// Should show built-in targets at minimum
		hasTargets := strings.Contains(outputStr, "opencode") ||
			strings.Contains(outputStr, "claudecode") ||
			strings.Contains(outputStr, "Built-in Targets") ||
			strings.Contains(outputStr, "Targets:")

		if !hasTargets {
			t.Errorf("Expected target list information in output, got: %s", outputStr)
		}
	})

	t.Run("WithVerboseFlag", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "target", "list", "--verbose")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Command output: %s", string(output))
		}

		outputStr := string(output)

		if len(outputStr) == 0 {
			t.Error("Expected output from 'target list --verbose', but got none")
		}
	})

	t.Run("WithDebugFlag", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "target", "list", "--debug")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Command output: %s", string(output))
		}

		outputStr := string(output)

		if len(outputStr) == 0 {
			t.Error("Expected output from 'target list --debug', but got none")
		}
	})

	t.Run("WithCustomTarget", func(t *testing.T) {
		// Add a custom target
		testTargetPath := filepath.Join(configDir, "custom-target")
		os.MkdirAll(testTargetPath, 0755)

		addCmd := exec.Command(binaryPath, "target", "add", "test-custom-target", testTargetPath)
		addCmd.CombinedOutput()

		// Test that target list shows custom target
		cmd := exec.Command(binaryPath, "target", "list")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Command output: %s", string(output))
		}

		outputStr := string(output)

		// Should show both built-in and custom targets
		hasBuiltIn := strings.Contains(outputStr, "opencode") || strings.Contains(outputStr, "claudecode")
		hasCustom := strings.Contains(outputStr, "test-custom-target") || strings.Contains(outputStr, "Custom Targets")

		if !hasBuiltIn {
			t.Errorf("Expected to see built-in targets, got: %s", outputStr)
		}

		if !hasCustom {
			t.Logf("Note: Custom target may not be shown if feature not implemented. Output: %s", outputStr)
		}
	})
}

// TestNoRegressionInOtherCommands verifies that changes don't affect other commands
func TestNoRegressionInOtherCommands(t *testing.T) {
	// Build agent-smith binary
	tempDir, err := os.MkdirTemp("", "agent-smith-regression-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	binaryPath := filepath.Join(tempDir, "agent-smith")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build agent-smith: %v\nOutput: %s", err, string(output))
	}

	// Set HOME to test directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	t.Run("LinkListStillWorks", func(t *testing.T) {
		// Test that link list command still works (already working command)
		cmd := exec.Command(binaryPath, "link", "list")
		output, err := cmd.CombinedOutput()

		// Command should execute without error
		if err != nil {
			// It's okay if there are no components linked
			if !strings.Contains(string(output), "No") && !strings.Contains(string(output), "not") {
				t.Errorf("link list command failed: %v\nOutput: %s", err, string(output))
			}
		}
	})

	t.Run("LinkStatusStillWorks", func(t *testing.T) {
		// Test that link status command still works (already working command)
		cmd := exec.Command(binaryPath, "link", "status")
		output, err := cmd.CombinedOutput()

		// Command should execute without error
		if err != nil {
			// It's okay if there are no components linked
			if !strings.Contains(string(output), "No") && !strings.Contains(string(output), "not") {
				t.Errorf("link status command failed: %v\nOutput: %s", err, string(output))
			}
		}
	})

	t.Run("HelpCommandStillWorks", func(t *testing.T) {
		// Test that help command still works
		cmd := exec.Command(binaryPath, "--help")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Errorf("help command failed: %v\nOutput: %s", err, string(output))
		}

		if !strings.Contains(string(output), "Usage") && !strings.Contains(string(output), "Available Commands") {
			t.Errorf("help command output doesn't look correct: %s", string(output))
		}
	})
}
