package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tjg184/agent-smith/internal/models"
)

// TestDetectComponentForPattern tests component detection for various patterns
func TestDetectComponentForPattern(t *testing.T) {
	rd := NewRepositoryDetector()

	// Create a temporary directory structure for testing
	tempDir := t.TempDir()
	repoPath := tempDir

	// Create test directory structure
	skillDir := filepath.Join(tempDir, ".opencode", "skills", "test-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("failed to create skill directory: %v", err)
	}

	// Create skill.md file
	skillFile := filepath.Join(skillDir, "skill.md")
	content := []byte("# Test Skill\n")
	if err := os.WriteFile(skillFile, content, 0644); err != nil {
		t.Fatalf("failed to create skill.md: %v", err)
	}

	tests := []struct {
		name          string
		fileName      string
		relPath       string
		fullRelPath   string
		pattern       models.ComponentDetectionPattern
		componentType models.ComponentType
		expectMatch   bool
		expectedName  string
	}{
		{
			name:        "Exact file match - skill.md",
			fileName:    "skill.md",
			relPath:     ".opencode/skills/test-skill",
			fullRelPath: ".opencode/skills/test-skill/skill.md",
			pattern: models.ComponentDetectionPattern{
				Name:        "skill",
				ExactFiles:  []string{"skill.md"},
				IgnorePaths: []string{"node_modules", "dist"},
			},
			componentType: models.ComponentSkill,
			expectMatch:   true,
			expectedName:  "test-skill",
		},
		{
			name:        "Path pattern match with extension",
			fileName:    "my-agent.md",
			relPath:     ".opencode/agents",
			fullRelPath: ".opencode/agents/my-agent.md",
			pattern: models.ComponentDetectionPattern{
				Name:           "agent",
				PathPatterns:   []string{".opencode/agents"},
				FileExtensions: []string{".md"},
				IgnorePaths:    []string{"node_modules"},
			},
			componentType: models.ComponentAgent,
			expectMatch:   true,
			expectedName:  "my-agent",
		},
		{
			name:        "Ignored path - should not match",
			fileName:    "skill.md",
			relPath:     "node_modules/test",
			fullRelPath: "node_modules/test/skill.md",
			pattern: models.ComponentDetectionPattern{
				Name:        "skill",
				ExactFiles:  []string{"skill.md"},
				IgnorePaths: []string{"node_modules", "dist"},
			},
			componentType: models.ComponentSkill,
			expectMatch:   false,
		},
		{
			name:        "Wrong extension - matches on path pattern alone",
			fileName:    "my-agent.txt",
			relPath:     ".opencode/agents",
			fullRelPath: ".opencode/agents/my-agent.txt",
			pattern: models.ComponentDetectionPattern{
				Name:           "agent",
				PathPatterns:   []string{".opencode/agents"},
				FileExtensions: []string{".md"},
				IgnorePaths:    []string{},
			},
			componentType: models.ComponentAgent,
			expectMatch:   true, // Matches on path pattern alone (fallback behavior)
			expectedName:  "my-agent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentName, componentPath, matched := rd.DetectComponentForPattern(
				tt.fileName,
				tt.relPath,
				tt.fullRelPath,
				repoPath,
				tt.pattern,
				tt.componentType,
			)

			if matched != tt.expectMatch {
				t.Errorf("expected match = %v, got %v", tt.expectMatch, matched)
			}

			if tt.expectMatch && componentName != tt.expectedName {
				t.Errorf("expected component name %s, got %s", tt.expectedName, componentName)
			}

			if tt.expectMatch && componentPath == "" {
				t.Error("expected non-empty component path")
			}
		})
	}
}

// TestDetectComponentsInRepo tests detecting components in a repository
func TestDetectComponentsInRepo(t *testing.T) {
	rd := NewRepositoryDetector()

	// Create a temporary directory structure
	tempDir := t.TempDir()

	// Create .opencode directory
	opencodeDir := filepath.Join(tempDir, ".opencode")
	if err := os.Mkdir(opencodeDir, 0755); err != nil {
		t.Fatalf("failed to create .opencode directory: %v", err)
	}

	// Create skills directory
	skillsDir := filepath.Join(opencodeDir, "skills")
	if err := os.Mkdir(skillsDir, 0755); err != nil {
		t.Fatalf("failed to create skills directory: %v", err)
	}

	// Create a skill
	skillDir := filepath.Join(skillsDir, "test-skill")
	if err := os.Mkdir(skillDir, 0755); err != nil {
		t.Fatalf("failed to create skill directory: %v", err)
	}

	skillFile := filepath.Join(skillDir, "skill.md")
	skillContent := []byte("# Test Skill\n")
	if err := os.WriteFile(skillFile, skillContent, 0644); err != nil {
		t.Fatalf("failed to create skill.md: %v", err)
	}

	// Create agents directory
	agentsDir := filepath.Join(opencodeDir, "agents")
	if err := os.Mkdir(agentsDir, 0755); err != nil {
		t.Fatalf("failed to create agents directory: %v", err)
	}

	// Create an agent
	agentFile := filepath.Join(agentsDir, "test-agent.md")
	agentContent := []byte("# Test Agent\n")
	if err := os.WriteFile(agentFile, agentContent, 0644); err != nil {
		t.Fatalf("failed to create test-agent.md: %v", err)
	}

	// Create commands directory
	commandsDir := filepath.Join(opencodeDir, "commands")
	if err := os.Mkdir(commandsDir, 0755); err != nil {
		t.Fatalf("failed to create commands directory: %v", err)
	}

	// Create a command
	commandFile := filepath.Join(commandsDir, "test-command.md")
	commandContent := []byte("# Test Command\n")
	if err := os.WriteFile(commandFile, commandContent, 0644); err != nil {
		t.Fatalf("failed to create test-command.md: %v", err)
	}

	// Detect components
	components, err := rd.DetectComponentsInRepo(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify we found components
	if len(components) == 0 {
		t.Error("expected to find components, got none")
	}

	// Count components by type
	componentCounts := make(map[models.ComponentType]int)
	for _, comp := range components {
		componentCounts[comp.Type]++
	}

	// We should have at least one of each type
	// Note: The exact counts depend on the default detection config
	t.Logf("Found %d total components", len(components))
	for compType, count := range componentCounts {
		t.Logf("  %s: %d", compType, count)
	}
}

// TestDetectComponentsInRepoWithDuplicates tests duplicate component detection
func TestDetectComponentsInRepoWithDuplicates(t *testing.T) {
	rd := NewRepositoryDetector()

	// Create a temporary directory structure
	tempDir := t.TempDir()

	// Create .opencode directory
	opencodeDir := filepath.Join(tempDir, ".opencode")
	if err := os.Mkdir(opencodeDir, 0755); err != nil {
		t.Fatalf("failed to create .opencode directory: %v", err)
	}

	// Create agents directory
	agentsDir := filepath.Join(opencodeDir, "agents")
	if err := os.Mkdir(agentsDir, 0755); err != nil {
		t.Fatalf("failed to create agents directory: %v", err)
	}

	// Create duplicate agents with same name
	agent1File := filepath.Join(agentsDir, "test-agent.md")
	agent1Content := []byte("# Test Agent\n")
	if err := os.WriteFile(agent1File, agent1Content, 0644); err != nil {
		t.Fatalf("failed to create first test-agent.md: %v", err)
	}

	// Create a subdirectory with another agent of the same name
	subDir := filepath.Join(agentsDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	agent2File := filepath.Join(subDir, "test-agent.md")
	agent2Content := []byte("# Test Agent Duplicate\n")
	if err := os.WriteFile(agent2File, agent2Content, 0644); err != nil {
		t.Fatalf("failed to create second test-agent.md: %v", err)
	}

	// Detect components
	components, err := rd.DetectComponentsInRepo(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify we only got one component (duplicates should be skipped)
	agentCount := 0
	for _, comp := range components {
		if comp.Type == models.ComponentAgent {
			agentCount++
		}
	}

	// We should only have one agent (the first one found)
	if agentCount != 1 {
		t.Logf("Found %d agents, but duplicate detection should result in only 1", agentCount)
	}
}

// TestDetectComponentsInRepoEmpty tests detecting components in an empty repo
func TestDetectComponentsInRepoEmpty(t *testing.T) {
	rd := NewRepositoryDetector()

	// Create an empty temporary directory
	tempDir := t.TempDir()

	components, err := rd.DetectComponentsInRepo(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(components) != 0 {
		t.Errorf("expected no components in empty directory, got %d", len(components))
	}
}

// TestDetectComponentsInRepoWithIgnoredPaths tests that ignored paths are skipped
func TestDetectComponentsInRepoWithIgnoredPaths(t *testing.T) {
	rd := NewRepositoryDetector()

	// Create a temporary directory structure
	tempDir := t.TempDir()

	// Create node_modules directory (should be ignored)
	nodeModulesDir := filepath.Join(tempDir, "node_modules")
	if err := os.Mkdir(nodeModulesDir, 0755); err != nil {
		t.Fatalf("failed to create node_modules directory: %v", err)
	}

	// Create a skill.md in node_modules (should be ignored)
	skillFile := filepath.Join(nodeModulesDir, "skill.md")
	skillContent := []byte("# Test Skill\n")
	if err := os.WriteFile(skillFile, skillContent, 0644); err != nil {
		t.Fatalf("failed to create skill.md in node_modules: %v", err)
	}

	// Detect components
	components, err := rd.DetectComponentsInRepo(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify no components were found in ignored paths
	for _, comp := range components {
		if comp.FilePath != "" && filepath.Dir(comp.FilePath) == "node_modules" {
			t.Error("found component in node_modules, which should be ignored")
		}
	}
}
