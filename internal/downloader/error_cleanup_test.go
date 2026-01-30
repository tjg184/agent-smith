//go:build integration
// +build integration

package downloader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tgaines/agent-smith/internal/detector"
	"github.com/tgaines/agent-smith/internal/fileutil"
	gitpkg "github.com/tgaines/agent-smith/internal/git"
)

// TestAgentDownloadErrorCleanup tests that directories are cleaned up on errors
func TestAgentDownloadErrorCleanup(t *testing.T) {
	// Create temporary base directory
	tempDir, err := os.MkdirTemp("", "agent-smith-cleanup-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	baseDir := filepath.Join(tempDir, "agents")
	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		t.Fatalf("Failed to create base directory: %v", err)
	}

	// Create downloader with test base directory
	ad := &AgentDownloader{
		baseDir:  baseDir,
		detector: detector.NewRepositoryDetector(),
		cloner:   gitpkg.NewDefaultCloner(),
	}

	// Test case: Invalid repository URL should not leave empty directories
	t.Run("InvalidRepoURL", func(t *testing.T) {
		agentName := "invalid-agent"
		invalidRepoURL := "https://github.com/nonexistent/invalid-repo-xyz123"

		// Attempt download with invalid URL
		err := ad.DownloadAgent(invalidRepoURL, agentName)
		if err == nil {
			t.Fatal("Expected error for invalid repository, got nil")
		}

		// Verify directory was cleaned up
		agentDir := filepath.Join(baseDir, agentName)
		if _, err := os.Stat(agentDir); !os.IsNotExist(err) {
			t.Errorf("Expected agent directory to be cleaned up, but it exists: %s", agentDir)
		}
	})

	// Test case: Non-matching component name should clean up directory
	t.Run("NonMatchingComponentName", func(t *testing.T) {
		// Create a temporary mock repository
		mockRepoDir, err := os.MkdirTemp("", "mock-repo-*")
		if err != nil {
			t.Fatalf("Failed to create mock repo: %v", err)
		}
		defer os.RemoveAll(mockRepoDir)

		// Create agent files with specific names
		agentsDir := filepath.Join(mockRepoDir, "agents")
		if err := os.MkdirAll(agentsDir, 0755); err != nil {
			t.Fatalf("Failed to create agents directory: %v", err)
		}

		// Create agent1 and agent2 directories
		agent1Dir := filepath.Join(agentsDir, "agent1")
		agent2Dir := filepath.Join(agentsDir, "agent2")
		if err := os.MkdirAll(agent1Dir, 0755); err != nil {
			t.Fatalf("Failed to create agent1 directory: %v", err)
		}
		if err := os.MkdirAll(agent2Dir, 0755); err != nil {
			t.Fatalf("Failed to create agent2 directory: %v", err)
		}

		// Create AGENT.md files
		if err := os.WriteFile(filepath.Join(agent1Dir, "AGENT.md"), []byte("# Agent 1"), 0644); err != nil {
			t.Fatalf("Failed to create AGENT.md: %v", err)
		}
		if err := os.WriteFile(filepath.Join(agent2Dir, "AGENT.md"), []byte("# Agent 2"), 0644); err != nil {
			t.Fatalf("Failed to create AGENT.md: %v", err)
		}

		// Try to download a non-existent agent (should fail and clean up)
		nonExistentAgent := "agent3"
		err = ad.DownloadAgent(mockRepoDir, nonExistentAgent, mockRepoDir)

		if err == nil {
			t.Fatal("Expected error for non-existent agent, got nil")
		}

		// Verify directory was cleaned up
		agentDir := filepath.Join(baseDir, nonExistentAgent)
		if _, err := os.Stat(agentDir); !os.IsNotExist(err) {
			t.Errorf("Expected agent directory to be cleaned up, but it exists: %s", agentDir)
		}
	})
}

// TestCommandDownloadErrorCleanup tests that command directories are cleaned up on errors
func TestCommandDownloadErrorCleanup(t *testing.T) {
	// Create temporary base directory
	tempDir, err := os.MkdirTemp("", "agent-smith-cleanup-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	baseDir := filepath.Join(tempDir, "commands")
	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		t.Fatalf("Failed to create base directory: %v", err)
	}

	// Create downloader with test base directory
	cd := &CommandDownloader{
		baseDir:  baseDir,
		detector: detector.NewRepositoryDetector(),
		cloner:   gitpkg.NewDefaultCloner(),
	}

	// Test case: Invalid repository URL should not leave empty directories
	t.Run("InvalidRepoURL", func(t *testing.T) {
		commandName := "invalid-command"
		invalidRepoURL := "https://github.com/nonexistent/invalid-repo-xyz123"

		// Attempt download with invalid URL
		err := cd.DownloadCommand(invalidRepoURL, commandName)
		if err == nil {
			t.Fatal("Expected error for invalid repository, got nil")
		}

		// Verify directory was cleaned up
		commandDir := filepath.Join(baseDir, commandName)
		if _, err := os.Stat(commandDir); !os.IsNotExist(err) {
			t.Errorf("Expected command directory to be cleaned up, but it exists: %s", commandDir)
		}
	})
}

// TestSkillDownloadErrorCleanup tests that skill directories are cleaned up on errors
func TestSkillDownloadErrorCleanup(t *testing.T) {
	// Create temporary base directory
	tempDir, err := os.MkdirTemp("", "agent-smith-cleanup-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	baseDir := filepath.Join(tempDir, "skills")
	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		t.Fatalf("Failed to create base directory: %v", err)
	}

	// Create downloader with test base directory
	sd := &SkillDownloader{
		baseDir:  baseDir,
		detector: detector.NewRepositoryDetector(),
		cloner:   gitpkg.NewDefaultCloner(),
	}

	// Test case: Invalid repository URL should not leave empty directories
	t.Run("InvalidRepoURL", func(t *testing.T) {
		skillName := "invalid-skill"
		invalidRepoURL := "https://github.com/nonexistent/invalid-repo-xyz123"

		// Attempt download with invalid URL
		err := sd.DownloadSkill(invalidRepoURL, skillName)
		if err == nil {
			t.Fatal("Expected error for invalid repository, got nil")
		}

		// Verify directory was cleaned up
		skillDir := filepath.Join(baseDir, skillName)
		if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
			t.Errorf("Expected skill directory to be cleaned up, but it exists: %s", skillDir)
		}
	})
}

// TestPartialInstallCleanup tests that partial installations are cleaned up on errors
func TestPartialInstallCleanup(t *testing.T) {
	// Create temporary base directory
	tempDir, err := os.MkdirTemp("", "agent-smith-cleanup-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	baseDir := filepath.Join(tempDir, "agents")
	if err := fileutil.CreateDirectoryWithPermissions(baseDir); err != nil {
		t.Fatalf("Failed to create base directory: %v", err)
	}

	// Test that even if a directory is created, it's removed on error
	agentName := "test-agent"
	agentDir := filepath.Join(baseDir, agentName)

	// Manually create directory to simulate partial install
	if err := fileutil.CreateDirectoryWithPermissions(agentDir); err != nil {
		t.Fatalf("Failed to create agent directory: %v", err)
	}

	// Create a test file to simulate partial installation
	testFile := filepath.Join(agentDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Verify the directory exists with content
	if _, err := os.Stat(agentDir); os.IsNotExist(err) {
		t.Fatal("Expected agent directory to exist before cleanup")
	}
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("Expected test file to exist before cleanup")
	}

	// Now simulate cleanup
	os.RemoveAll(agentDir)

	// Verify directory was cleaned up completely
	if _, err := os.Stat(agentDir); !os.IsNotExist(err) {
		t.Errorf("Expected agent directory to be cleaned up, but it exists: %s", agentDir)
	}
}
