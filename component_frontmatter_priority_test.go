package main
import "github.com/tgaines/agent-smith/internal/models"

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFrontmatterNamePriority tests that frontmatter name takes priority over directory/filename
func TestFrontmatterNamePriority(t *testing.T) {
	tests := []struct {
		name          string
		componentType models.ComponentType
		filesToCreate map[string]string
		expectedNames []string
		description   string
	}{
		{
			name:          "skill-with-frontmatter-name",
			componentType: models.ComponentSkill,
			filesToCreate: map[string]string{
				"myskill/SKILL.md": `---
name: custom-skill-name
description: A custom skill
---

# My Skill`,
			},
			expectedNames: []string{"custom-skill-name"},
			description:   "Skill name from frontmatter takes priority over directory name",
		},
		{
			name:          "agent-with-frontmatter-name",
			componentType: models.ComponentAgent,
			filesToCreate: map[string]string{
				"agents/coding.md": `---
name: advanced-coder
description: An advanced coding agent
---

# Coding Agent`,
			},
			expectedNames: []string{"advanced-coder"},
			description:   "Agent name from frontmatter takes priority over filename",
		},
		{
			name:          "command-with-frontmatter-name",
			componentType: models.ComponentCommand,
			filesToCreate: map[string]string{
				"commands/deploy.md": `---
name: super-deploy
description: A deployment command
---

# Deploy Command`,
			},
			expectedNames: []string{"super-deploy"},
			description:   "Command name from frontmatter takes priority over filename",
		},
		{
			name:          "skill-without-frontmatter",
			componentType: models.ComponentSkill,
			filesToCreate: map[string]string{
				"python-dev/SKILL.md": "# Python Dev Skill\n\nNo frontmatter here.",
			},
			expectedNames: []string{"python-dev"},
			description:   "Skill without frontmatter uses directory name",
		},
		{
			name:          "agent-without-frontmatter",
			componentType: models.ComponentAgent,
			filesToCreate: map[string]string{
				"agents/helper.md": "# Helper Agent\n\nNo frontmatter here.",
			},
			expectedNames: []string{"helper"},
			description:   "Agent without frontmatter uses filename",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir, err := os.MkdirTemp("", "frontmatter-priority-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Create test files
			for filePath, content := range tt.filesToCreate {
				fullPath := filepath.Join(tempDir, filePath)
				dir := filepath.Dir(fullPath)

				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("Failed to create directory %s: %v", dir, err)
				}

				if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create file %s: %v", fullPath, err)
				}
			}

			// Create detector and find components
			detector := NewRepositoryDetector()
			components, err := detector.detectComponentsInRepo(tempDir)
			if err != nil {
				t.Fatalf("Failed to detect components: %v", err)
			}

			// Filter components by type
			var filteredComponents []models.DetectedComponent
			for _, comp := range components {
				if comp.Type == tt.componentType {
					filteredComponents = append(filteredComponents, comp)
				}
			}

			// Extract component names
			actualNames := make([]string, len(filteredComponents))
			for i, comp := range filteredComponents {
				actualNames[i] = comp.Name
			}

			// Compare with expected names
			if len(actualNames) != len(tt.expectedNames) {
				t.Errorf("Expected %d components, got %d", len(tt.expectedNames), len(actualNames))
				t.Logf("Expected: %v", tt.expectedNames)
				t.Logf("Actual:   %v", actualNames)
				return
			}

			// Create maps for comparison
			expectedMap := make(map[string]bool)
			for _, name := range tt.expectedNames {
				expectedMap[name] = true
			}

			actualMap := make(map[string]bool)
			for _, name := range actualNames {
				actualMap[name] = true
			}

			// Check for missing expected names
			for expected := range expectedMap {
				if !actualMap[expected] {
					t.Errorf("Missing expected component: %s", expected)
				}
			}

			// Check for unexpected names
			for actual := range actualMap {
				if !expectedMap[actual] {
					t.Errorf("Unexpected component detected: %s", actual)
				}
			}

			// Log detection details
			t.Logf("Test: %s", tt.description)
			t.Logf("Detected %d components of type %s", len(filteredComponents), tt.componentType)
			for _, comp := range filteredComponents {
				t.Logf("  - Name: %s, Path: %s, Source: %s", comp.Name, comp.Path, comp.SourceFile)
			}
		})
	}
}
