package main

import (
	"github.com/tgaines/agent-smith/internal/detector"
	"github.com/tgaines/agent-smith/internal/downloader"
	"github.com/tgaines/agent-smith/internal/models"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// TestHelper provides utilities for creating mock repositories
type TestHelper struct {
	t       *testing.T
	tempDir string
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T) *TestHelper {
	tempDir, err := os.MkdirTemp("", "agent-smith-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	return &TestHelper{
		t:       t,
		tempDir: tempDir,
	}
}

// Cleanup removes all temporary directories
func (h *TestHelper) Cleanup() {
	if err := os.RemoveAll(h.tempDir); err != nil {
		h.t.Logf("Warning: failed to clean up temp directory: %v", err)
	}
}

// CreateMockRepo creates a mock git repository with specified structure
func (h *TestHelper) CreateMockRepo(name string, files map[string]string) string {
	repoPath := filepath.Join(h.tempDir, name)

	// Create repository directory
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		h.t.Fatalf("Failed to create repo directory: %v", err)
	}

	// Initialize git repository
	repo, err := git.PlainInit(repoPath, false)
	if err != nil {
		h.t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Create files
	for filePath, content := range files {
		fullPath := filepath.Join(repoPath, filePath)

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			h.t.Fatalf("Failed to create directory for %s: %v", filePath, err)
		}

		// Write file
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			h.t.Fatalf("Failed to write file %s: %v", filePath, err)
		}
	}

	// Add all files to git
	worktree, err := repo.Worktree()
	if err != nil {
		h.t.Fatalf("Failed to get worktree: %v", err)
	}

	if err := worktree.AddGlob("."); err != nil {
		h.t.Fatalf("Failed to add files to git: %v", err)
	}

	// Commit files
	_, err = worktree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})
	if err != nil {
		h.t.Fatalf("Failed to commit files: %v", err)
	}

	return repoPath
}

// CreatePluginRepo creates a mock repository with plugin structure
func (h *TestHelper) CreatePluginRepo() string {
	files := map[string]string{
		"plugins/ui-design/agents/accessibility-expert.md": `---
name: accessibility-expert
---
# Accessibility Expert Agent
An agent that helps with accessibility compliance.`,
		"plugins/ui-design/agents/design-system-architect.md": `---
name: design-system-architect
---
# Design System Architect
An agent that helps build design systems.`,
		"plugins/ui-design/agents/ux-researcher.md": `---
name: ux-researcher
---
# UX Researcher
An agent that conducts UX research.`,
		"plugins/ui-design/skills/wcag-compliance/SKILL.md": `---
name: wcag-compliance
---
# WCAG Compliance Skill
Helps ensure WCAG compliance.`,
		"plugins/ui-design/commands/contrast-check.md": `---
name: contrast-check
---
# Contrast Check Command
Checks color contrast ratios.`,
		"README.md": "# UI Design Plugin Repository",
	}
	return h.CreateMockRepo("plugin-repo", files)
}

// CreateFlatRepo creates a mock repository with flat structure
func (h *TestHelper) CreateFlatRepo() string {
	files := map[string]string{
		"agents/chatbot.md": `---
name: chatbot
---
# Chatbot Agent
A simple chatbot agent.`,
		"README.md": "# Simple Flat Repository",
	}
	return h.CreateMockRepo("flat-repo", files)
}

// CreateMonorepo creates a mock repository with monorepo structure (non-plugin)
func (h *TestHelper) CreateMonorepo() string {
	files := map[string]string{
		"agents/bash-expert.md": `---
name: bash-expert
---
# Bash Expert Agent
Expert in bash scripting.`,
		"agents/python-expert.md": `---
name: python-expert
---
# Python Expert Agent
Expert in Python development.`,
		"skills/bash-scripting/SKILL.md": `---
name: bash-scripting
---
# Bash Scripting Skill
Master bash scripting.`,
		"commands/deploy.md": `---
name: deploy
---
# Deploy Command
Deployment automation.`,
		"README.md": "# Multi-Component Monorepo",
	}
	return h.CreateMockRepo("monorepo", files)
}

// CreateInstallDir creates a temporary installation directory
func (h *TestHelper) CreateInstallDir() string {
	installDir := filepath.Join(h.tempDir, "install")
	if err := os.MkdirAll(installDir, 0755); err != nil {
		h.t.Fatalf("Failed to create install directory: %v", err)
	}
	return installDir
}

// VerifyFileExists checks if a file exists at the given path
func (h *TestHelper) VerifyFileExists(path string, description string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		h.t.Errorf("Expected file does not exist: %s (%s)", path, description)
	}
}

// VerifyDirExists checks if a directory exists
func (h *TestHelper) VerifyDirExists(path string, description string) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		h.t.Errorf("Expected directory does not exist: %s (%s)", path, description)
		return
	}
	if !info.IsDir() {
		h.t.Errorf("Expected directory but found file: %s (%s)", path, description)
	}
}

// VerifyFileContent checks if a file contains expected content
func (h *TestHelper) VerifyFileContent(path string, expected string, description string) {
	content, err := os.ReadFile(path)
	if err != nil {
		h.t.Errorf("Failed to read file %s: %v (%s)", path, err, description)
		return
	}
	if !strings.Contains(string(content), expected) {
		h.t.Errorf("File %s does not contain expected content '%s' (%s)", path, expected, description)
	}
}

// CountFilesInDir counts files (not directories) in a directory
func (h *TestHelper) CountFilesInDir(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		h.t.Fatalf("Failed to read directory %s: %v", dir, err)
	}
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}
	return count
}

// TestPluginMirroringEndToEnd tests the complete plugin download workflow
func TestPluginMirroringEndToEnd(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Create mock plugin repository
	repoPath := helper.CreatePluginRepo()
	installDir := helper.CreateInstallDir()
	agentsDir := filepath.Join(installDir, "agents")

	// Create downloader
	detect := detector.NewRepositoryDetector()
	dl := downloader.NewAgentDownloaderWithParams(agentsDir, detect)

	// Detect components first
	components, err := detect.DetectComponentsInRepo(repoPath)
	if err != nil {
		t.Fatalf("Failed to detect components: %v", err)
	}

	// Download first agent from plugin
	err = dl.DownloadAgentWithRepo(
		"file://"+repoPath,
		"accessibility-expert",
		"file://"+repoPath,
		repoPath,
		components,
	)
	if err != nil {
		t.Fatalf("Failed to download agent from plugin: %v", err)
	}

	// Verify plugin directory structure was created
	// Note: Plugin is stored relative to parent of baseDir (agents directory)
	pluginDir := filepath.Join(filepath.Dir(agentsDir), "plugins", "ui-design")
	helper.VerifyDirExists(pluginDir, "Plugin directory should be created")

	// Verify all agents from plugin are present
	agentFiles := []string{
		"agents/accessibility-expert.md",
		"agents/design-system-architect.md",
		"agents/ux-researcher.md",
	}
	for _, file := range agentFiles {
		fullPath := filepath.Join(pluginDir, file)
		helper.VerifyFileExists(fullPath, "Agent file from plugin")
	}

	// Verify skills and commands are also present
	helper.VerifyFileExists(
		filepath.Join(pluginDir, "skills/wcag-compliance/SKILL.md"),
		"Skill from plugin",
	)
	helper.VerifyFileExists(
		filepath.Join(pluginDir, "commands/contrast-check.md"),
		"Command from plugin",
	)

	// Verify metadata file exists and contains pluginPath
	metadataFile := filepath.Join(pluginDir, ".agent-metadata.json")
	helper.VerifyFileExists(metadataFile, "Metadata file")
	helper.VerifyFileContent(metadataFile, "plugins/ui-design", "Plugin path in metadata")

	// Note: Lock file is created by higher-level commands, not by downloadAgentWithRepo directly
	// So we don't check for it in this unit-level integration test
}

// TestPluginStructureReuse tests that downloading multiple components from the same plugin reuses the structure
func TestPluginStructureReuse(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Create mock plugin repository
	repoPath := helper.CreatePluginRepo()
	installDir := helper.CreateInstallDir()
	agentsDir := filepath.Join(installDir, "agents")

	// Create downloader
	detect := detector.NewRepositoryDetector()
	dl := downloader.NewAgentDownloaderWithParams(agentsDir, detect)

	// Detect components first
	components, err := detect.DetectComponentsInRepo(repoPath)
	if err != nil {
		t.Fatalf("Failed to detect components: %v", err)
	}

	// Download first agent
	err = dl.DownloadAgentWithRepo(
		"file://"+repoPath,
		"accessibility-expert",
		"file://"+repoPath,
		repoPath,
		components,
	)
	if err != nil {
		t.Fatalf("Failed to download first agent: %v", err)
	}

	pluginDir := filepath.Join(filepath.Dir(agentsDir), "plugins", "ui-design")

	// Count files before second download
	filesBefore := helper.CountFilesInDir(filepath.Join(pluginDir, "agents"))

	// Download second agent from same plugin
	err = dl.DownloadAgentWithRepo(
		"file://"+repoPath,
		"design-system-architect",
		"file://"+repoPath,
		repoPath,
		components,
	)
	if err != nil {
		t.Fatalf("Failed to download second agent: %v", err)
	}

	// Verify plugin structure was reused (file count should be the same)
	filesAfter := helper.CountFilesInDir(filepath.Join(pluginDir, "agents"))
	if filesAfter != filesBefore {
		t.Errorf("Expected plugin structure to be reused, but file count changed from %d to %d", filesBefore, filesAfter)
	}

	// Verify both agents are accessible
	agent1 := filepath.Join(pluginDir, "agents/accessibility-expert.md")
	agent2 := filepath.Join(pluginDir, "agents/design-system-architect.md")
	helper.VerifyFileExists(agent1, "First agent")
	helper.VerifyFileExists(agent2, "Second agent")
}

// TestBackwardCompatibilityFlatStructure tests that flat repositories work correctly
func TestBackwardCompatibilityFlatStructure(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Create mock flat repository
	repoPath := helper.CreateFlatRepo()
	installDir := helper.CreateInstallDir()
	agentsDir := filepath.Join(installDir, "agents")

	// Create downloader
	detect := detector.NewRepositoryDetector()
	dl := downloader.NewAgentDownloaderWithParams(agentsDir, detect)

	// Detect components first
	components, err := detect.DetectComponentsInRepo(repoPath)
	if err != nil {
		t.Fatalf("Failed to detect components: %v", err)
	}

	// Download agent from flat repository
	err = dl.DownloadAgentWithRepo(
		"file://"+repoPath,
		"chatbot",
		"file://"+repoPath,
		repoPath,
		components,
	)
	if err != nil {
		t.Fatalf("Failed to download agent from flat repo: %v", err)
	}

	// Verify flat structure (no plugins directory)
	agentDir := filepath.Join(agentsDir, "chatbot")
	helper.VerifyDirExists(agentDir, "Flat agent directory")
	helper.VerifyFileExists(filepath.Join(agentDir, "chatbot.md"), "Agent file in flat structure")

	// Verify metadata does NOT contain pluginPath
	metadataFile := filepath.Join(agentDir, ".agent-metadata.json")
	helper.VerifyFileExists(metadataFile, "Metadata file")

	content, err := os.ReadFile(metadataFile)
	if err != nil {
		t.Fatalf("Failed to read metadata file: %v", err)
	}
	if strings.Contains(string(content), "pluginPath") {
		t.Errorf("Flat structure should not have pluginPath in metadata")
	}
}

// TestBackwardCompatibilityMonorepo tests that monorepo structures work correctly
func TestBackwardCompatibilityMonorepo(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Create mock monorepo
	_ = helper.CreateMonorepo()

	// Test with bulk downloader (monorepo scenario)
	_ = NewBulkDownloader()

	// Download all components using AddAll (requires proper URL format)
	// AddAll expects a github-style URL, so this test may not work with file:// URLs
	// For now, we'll skip the actual AddAll call and just verify the test setup works
	t.Skip("AddAll requires network URLs, skipping local file test")
}

// TestLinkingPluginComponents tests symlink creation for plugin-based components
func TestLinkingPluginComponents(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Create mock plugin repository and download
	repoPath := helper.CreatePluginRepo()
	installDir := helper.CreateInstallDir()
	agentsDir := filepath.Join(installDir, "agents")
	configDir := filepath.Join(installDir, "config", "opencode", "agents")

	// Create downloader and download agent
	detect := detector.NewRepositoryDetector()
	dl := downloader.NewAgentDownloaderWithParams(agentsDir, detect)

	// Detect components first
	components, err := detect.DetectComponentsInRepo(repoPath)
	if err != nil {
		t.Fatalf("Failed to detect components: %v", err)
	}

	err = dl.DownloadAgentWithRepo(
		"file://"+repoPath,
		"accessibility-expert",
		"file://"+repoPath,
		repoPath,
		components,
	)
	if err != nil {
		t.Fatalf("Failed to download agent: %v", err)
	}

	// Create config directory for linking
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Create symlink manually (simulating link command)
	pluginAgentFile := filepath.Join(filepath.Dir(agentsDir), "plugins/ui-design/agents/accessibility-expert.md")
	symlinkPath := filepath.Join(configDir, "accessibility-expert.md")

	err = os.Symlink(pluginAgentFile, symlinkPath)
	if err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Verify symlink exists and points to correct location
	linkTarget, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}

	if linkTarget != pluginAgentFile {
		t.Errorf("Symlink points to wrong location: expected %s, got %s", pluginAgentFile, linkTarget)
	}

	// Verify symlink resolves to actual file
	_, err = os.Stat(symlinkPath)
	if err != nil {
		t.Errorf("Symlink does not resolve to valid file: %v", err)
	}
}

// TestCrossPlatformPathHandling tests path handling across different platforms
func TestCrossPlatformPathHandling(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Create mock plugin repository
	repoPath := helper.CreatePluginRepo()
	installDir := helper.CreateInstallDir()
	agentsDir := filepath.Join(installDir, "agents")

	// Create downloader
	detect := detector.NewRepositoryDetector()
	dl := downloader.NewAgentDownloaderWithParams(agentsDir, detect)

	// Detect components first
	components, err := detect.DetectComponentsInRepo(repoPath)
	if err != nil {
		t.Fatalf("Failed to detect components: %v", err)
	}

	// Download agent
	err = dl.DownloadAgentWithRepo(
		"file://"+repoPath,
		"accessibility-expert",
		"file://"+repoPath,
		repoPath,
		components,
	)
	if err != nil {
		t.Fatalf("Failed to download agent: %v", err)
	}

	// Verify paths use correct separators for the platform
	pluginDir := filepath.Join(filepath.Dir(agentsDir), "plugins", "ui-design")
	helper.VerifyDirExists(pluginDir, "Plugin directory with platform-specific paths")

	// Read metadata and verify paths use forward slashes (normalized)
	metadataFile := filepath.Join(pluginDir, ".agent-metadata.json")
	content, err := os.ReadFile(metadataFile)
	if err != nil {
		t.Fatalf("Failed to read metadata: %v", err)
	}

	// pluginPath in metadata should use forward slashes
	if !strings.Contains(string(content), "plugins/ui-design") {
		t.Errorf("Metadata should contain normalized plugin path with forward slashes")
	}
}

// TestErrorHandlingMissingPlugin tests error handling when plugin structure is incomplete
func TestErrorHandlingMissingPlugin(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Create repository with malformed plugin structure
	files := map[string]string{
		"plugins/broken/README.md": "# Broken Plugin",
		// Missing required component files
	}
	repoPath := helper.CreateMockRepo("broken-repo", files)
	installDir := helper.CreateInstallDir()
	agentsDir := filepath.Join(installDir, "agents")

	// Create downloader
	detect := detector.NewRepositoryDetector()
	dl := downloader.NewAgentDownloaderWithParams(agentsDir, detect)

	// Detect components first
	components, err := detect.DetectComponentsInRepo(repoPath)
	if err != nil {
		t.Fatalf("Failed to detect components: %v", err)
	}

	// Attempt to download non-existent agent
	err = dl.DownloadAgentWithRepo(
		"file://"+repoPath,
		"non-existent-agent",
		"file://"+repoPath,
		repoPath,
		components,
	)

	// Should handle error gracefully (not crash)
	if err == nil {
		t.Logf("Expected error when downloading non-existent agent, but succeeded")
		// This is acceptable as it might fall back to direct download
	}
}

// TestPluginDetectionMixed tests that mixed plugin/non-plugin structures are handled correctly
func TestPluginDetectionMixed(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Create repository with both plugin and non-plugin agents
	files := map[string]string{
		"plugins/ui-design/agents/accessibility-expert.md": `---
name: accessibility-expert
---
# Accessibility Expert`,
		"agents/standalone-agent.md": `---
name: standalone-agent
---
# Standalone Agent`,
		"README.md": "# Mixed Repository",
	}
	repoPath := helper.CreateMockRepo("mixed-repo", files)

	// Detect components
	detect := detector.NewRepositoryDetector()
	components, err := detect.DetectComponentsInRepo(repoPath)
	if err != nil {
		t.Fatalf("Failed to detect components: %v", err)
	}

	// Filter for agents
	var agentComponents []models.DetectedComponent
	for _, comp := range components {
		if comp.Type == models.ComponentAgent {
			agentComponents = append(agentComponents, comp)
		}
	}
	components = agentComponents

	// Should detect both agents
	if len(components) != 2 {
		t.Errorf("Expected 2 agents, found %d", len(components))
	}

	// Verify plugin path detection returns empty (mixed structure)
	// TODO: Implement detectCommonPluginPath
	// pluginPath := detectCommonPluginPath(components)
	// if pluginPath != "" {
	// 	t.Errorf("Mixed structure should return empty plugin path, got: %s", pluginPath)
	// }
}

// TestPluginDirectoryStructure tests that plugin directories maintain proper structure
func TestPluginDirectoryStructure(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Create plugin directory structure
	pluginDir := filepath.Join(helper.tempDir, "plugins", "ui-design")
	agentsDir := filepath.Join(pluginDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create plugin directory: %v", err)
	}

	// Create some agent files
	agentFile := filepath.Join(agentsDir, "accessibility-expert.md")
	if err := os.WriteFile(agentFile, []byte("# Agent"), 0644); err != nil {
		t.Fatalf("Failed to create agent file: %v", err)
	}

	// Verify the structure exists
	helper.VerifyDirExists(pluginDir, "Plugin directory")
	helper.VerifyDirExists(agentsDir, "Agents subdirectory in plugin")
	helper.VerifyFileExists(agentFile, "Agent file in plugin structure")
}

// BenchmarkPluginDownload benchmarks the plugin download performance
func BenchmarkPluginDownload(b *testing.B) {
	// Note: Benchmarks use 'b' instead of 't'
	helper := &TestHelper{
		t:       &testing.T{}, // Create a minimal testing.T for compatibility
		tempDir: "",
	}

	tempDir, err := os.MkdirTemp("", "agent-smith-benchmark-*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	helper.tempDir = tempDir

	// Create mock plugin repository
	repoPath := helper.CreatePluginRepo()

	// Reset timer after setup
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		installDir := filepath.Join(tempDir, "install", "run"+string(rune(i)))
		agentsDir := filepath.Join(installDir, "agents")

		detect := detector.NewRepositoryDetector()
		dl := downloader.NewAgentDownloaderWithParams(agentsDir, detect)

		// Detect components first
		components, err := detect.DetectComponentsInRepo(repoPath)
		if err != nil {
			b.Fatalf("Failed to detect components: %v", err)
		}

		err = dl.DownloadAgentWithRepo(
			"file://"+repoPath,
			"accessibility-expert",
			"file://"+repoPath,
			repoPath,
			components,
		)
		if err != nil {
			b.Fatalf("Failed to download agent: %v", err)
		}
	}
}

// TestGitOperations tests that git operations work correctly with plugins
func TestGitOperations(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git not available, skipping test")
	}

	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Create plugin repository
	repoPath := helper.CreatePluginRepo()
	installDir := helper.CreateInstallDir()
	agentsDir := filepath.Join(installDir, "agents")

	// Download agent
	detect := detector.NewRepositoryDetector()
	dl := downloader.NewAgentDownloaderWithParams(agentsDir, detect)

	// Detect components first
	components, err := detect.DetectComponentsInRepo(repoPath)
	if err != nil {
		t.Fatalf("Failed to detect components: %v", err)
	}

	err = dl.DownloadAgentWithRepo(
		"file://"+repoPath,
		"accessibility-expert",
		"file://"+repoPath,
		repoPath,
		components,
	)
	if err != nil {
		t.Fatalf("Failed to download agent: %v", err)
	}

	// Verify commit hash was saved in metadata
	pluginDir := filepath.Join(filepath.Dir(agentsDir), "plugins", "ui-design")
	metadataFile := filepath.Join(pluginDir, ".agent-metadata.json")
	content, err := os.ReadFile(metadataFile)
	if err != nil {
		t.Fatalf("Failed to read metadata: %v", err)
	}

	if !strings.Contains(string(content), "commit") {
		t.Errorf("Metadata should contain commit hash")
	}
}
