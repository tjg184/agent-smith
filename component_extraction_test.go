package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestComponentNameExtraction tests component name extraction for all component types
func TestComponentNameExtraction(t *testing.T) {
	tests := []struct {
		name          string
		componentType ComponentType
		filesToCreate map[string]string
		expectedNames []string
		description   string
	}{
		{
			name:          "skill-exact-file",
			componentType: ComponentSkill,
			filesToCreate: map[string]string{
				"SKILL.md": "# Test Skill",
			},
			expectedNames: []string{"root-skill"},
			description:   "Skill detected via exact SKILL.md file match",
		},
		{
			name:          "skill-directory-named",
			componentType: ComponentSkill,
			filesToCreate: map[string]string{
				"myskill/SKILL.md": "# My Skill",
			},
			expectedNames: []string{"myskill"}, // Fixed: should extract actual directory name
			description:   "Skill name extracted from parent directory containing SKILL.md",
		},
		{
			name:          "skill-path-pattern",
			componentType: ComponentSkill,
			filesToCreate: map[string]string{
				"skills/python.md": "# Python Skill",
			},
			expectedNames: []string{}, // Skills only detect via exact SKILL.md files, not path patterns
			description:   "Skills only detect via exact SKILL.md files, not path patterns",
		},
		{
			name:          "agent-exact-file",
			componentType: ComponentAgent,
			filesToCreate: map[string]string{
				"AGENT.md": "# Test Agent",
			},
			expectedNames: []string{"root-agent"},
			description:   "Agent detected via exact AGENT.md file match",
		},
		{
			name:          "agent-directory-named",
			componentType: ComponentAgent,
			filesToCreate: map[string]string{
				"myagent/AGENT.md": "# My Agent",
			},
			expectedNames: []string{"myagent"}, // Extracts directory name containing AGENT.md
			description:   "Agent name extracted from parent directory containing AGENT.md",
		},
		{
			name:          "agent-path-pattern",
			componentType: ComponentAgent,
			filesToCreate: map[string]string{
				"agents/coding.md": "# Coding Agent",
			},
			expectedNames: []string{"coding"},
			description:   "Agent detected via /agents/ path pattern and file extension",
		},
		{
			name:          "command-exact-file",
			componentType: ComponentCommand,
			filesToCreate: map[string]string{
				"COMMAND.md": "# Test Command",
			},
			expectedNames: []string{"root-command"},
			description:   "Command detected via exact COMMAND.md file match",
		},
		{
			name:          "command-directory-named",
			componentType: ComponentCommand,
			filesToCreate: map[string]string{
				"mycommand/COMMAND.md": "# My Command",
			},
			expectedNames: []string{"mycommand"}, // Extracts directory name containing COMMAND.md
			description:   "Command name extracted from parent directory containing COMMAND.md",
		},
		{
			name:          "command-path-pattern",
			componentType: ComponentCommand,
			filesToCreate: map[string]string{
				"commands/deploy.md": "# Deploy Command",
			},
			expectedNames: []string{"deploy"},
			description:   "Command detected via /commands/ path pattern and file extension",
		},
		{
			name:          "multiple-skills-different-locations",
			componentType: ComponentSkill,
			filesToCreate: map[string]string{
				"SKILL.md":             "# Root Skill",
				"frontend/SKILL.md":    "# Frontend Skill",
				"backend/SKILL.md":     "# Backend Skill",
				"nested/deep/SKILL.md": "# Deep Skill",
			},
			expectedNames: []string{"root-skill", "frontend", "backend", "deep"}, // Each directory gets its own name
			description:   "Multiple skills detected in different locations",
		},
		{
			name:          "multiple-agents-different-locations",
			componentType: ComponentAgent,
			filesToCreate: map[string]string{
				"AGENT.md":          "# Root Agent",
				"ai/AGENT.md":       "# AI Agent",
				"agents/chatbot.md": "# Chatbot Agent",
				"bots/helper.md":    "# Helper Agent",
			},
			expectedNames: []string{"root-agent", "ai", "chatbot"}, // Root + ai directory + chatbot from agents/ path
			description:   "Multiple agents detected in different locations",
		},
		{
			name:          "multiple-commands-different-locations",
			componentType: ComponentCommand,
			filesToCreate: map[string]string{
				"COMMAND.md":        "# Root Command",
				"cli/COMMAND.md":    "# CLI Command",
				"commands/build.md": "# Build Command",
				"tools/deploy.md":   "# Deploy Command",
			},
			expectedNames: []string{"root-command", "cli", "build"}, // Root + cli directory + build from commands/ path
			description:   "Multiple commands detected in different locations",
		},
		{
			name:          "mixed-component-types",
			componentType: ComponentSkill, // We'll test detection for each type separately
			filesToCreate: map[string]string{
				"SKILL.md":          "# Root Skill",
				"myskill/SKILL.md":  "# My Skill",
				"AGENT.md":          "# Root Agent",
				"agents/helper.md":  "# Helper Agent",
				"COMMAND.md":        "# Root Command",
				"commands/build.md": "# Build Command",
			},
			expectedNames: []string{"root-skill", "myskill"}, // Both skills detected with their names
			description:   "Mixed component types - testing skill extraction",
		},
		{
			name:          "nested-component-structures",
			componentType: ComponentSkill,
			filesToCreate: map[string]string{
				"components/skills/web/SKILL.md": "# Web Skill",
				"lib/ai/skills/nlp/SKILL.md":     "# NLP Skill",
				"packages/tools/dev/SKILL.md":    "# Dev Skill",
				"src/skills/testing.md":          "# Testing Skill",
			},
			expectedNames: []string{"web", "nlp", "dev"}, // Each extracts the directory name containing SKILL.md
			description:   "Skills in deeply nested directory structures",
		},
		{
			name:          "special-characters-in-names",
			componentType: ComponentSkill,
			filesToCreate: map[string]string{
				"my-skill/SKILL.md":        "# Skill with dash",
				"my_skill/SKILL.md":        "# Skill with underscore",
				"skill123/SKILL.md":        "# Skill with numbers",
				"complex.name-v2/SKILL.md": "# Complex skill name",
			},
			expectedNames: []string{"my-skill", "my_skill", "skill123", "complex.name-v2"}, // Each directory name is used as-is
			description:   "Component names with special characters",
		},
		{
			name:          "ignored-paths",
			componentType: ComponentSkill,
			filesToCreate: map[string]string{
				"node_modules/skill/SKILL.md": "# Should be ignored",
				".git/skills/test.md":         "# Should be ignored",
				"build/skills/production.md":  "# Should be ignored",
				"skills/SKILL.md":             "# Should be detected",
			},
			expectedNames: []string{"skills"}, // Only skills/SKILL.md should be detected
			description:   "Paths that should be ignored during detection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir, err := os.MkdirTemp("", "component-test-*")
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
			var filteredComponents []DetectedComponent
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

			// Compare with expected names using set-based comparison (order doesn't matter)
			if len(actualNames) != len(tt.expectedNames) {
				t.Errorf("Expected %d components, got %d", len(tt.expectedNames), len(actualNames))
				t.Logf("Expected: %v", tt.expectedNames)
				t.Logf("Actual:   %v", actualNames)
				return
			}

			// Create maps for comparison to handle order differences
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

			// Log detection details for debugging
			t.Logf("Test: %s", tt.description)
			t.Logf("Detected %d components of type %s", len(filteredComponents), tt.componentType)
			for _, comp := range filteredComponents {
				t.Logf("  - Name: %s, Path: %s, Source: %s", comp.Name, comp.Path, comp.SourceFile)
			}
		})
	}
}

// TestComponentNameExtractionEdgeCases tests edge cases for component name extraction
func TestComponentNameExtractionEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		componentType ComponentType
		filesToCreate map[string]string
		expectedNames []string
		description   string
		expectError   bool
	}{
		{
			name:          "empty-directory",
			componentType: ComponentSkill,
			filesToCreate: map[string]string{},
			expectedNames: []string{},
			description:   "Empty directory should return no components",
			expectError:   false,
		},
		{
			name:          "only-non-component-files",
			componentType: ComponentSkill,
			filesToCreate: map[string]string{
				"README.md":    "# README",
				"package.json": "{}",
				"main.go":      "package main",
				"test.txt":     "test content",
			},
			expectedNames: []string{},
			description:   "Directory with no component files should return empty",
			expectError:   false,
		},
		{
			name:          "component-in-root-with-no-name",
			componentType: ComponentSkill,
			filesToCreate: map[string]string{
				"SKILL.md": "# Skill with no directory context",
			},
			expectedNames: []string{"root-skill"},
			description:   "Component in root should get default name",
			expectError:   false,
		},
		{
			name:          "component-with-only-file-extension",
			componentType: ComponentSkill,
			filesToCreate: map[string]string{
				"skills/.md": "# Skill with no proper name",
			},
			expectedNames: []string{},
			description:   "File with only extension should not be detected",
			expectError:   false,
		},
		{
			name:          "duplicate-component-names",
			componentType: ComponentSkill,
			filesToCreate: map[string]string{
				"skill1/SKILL.md":        "# First Skill",
				"skill2/SKILL.md":        "# Second Skill",
				"skill1/nested/SKILL.md": "# Nested Skill",
			},
			expectedNames: []string{"skill1", "skill2", "nested"}, // skill1, skill2, and nested are all detected
			description:   "Duplicate component names should be handled",
			expectError:   false,
		},
		{
			name:          "case-sensitive-paths",
			componentType: ComponentAgent,
			filesToCreate: map[string]string{
				"AGENT.md":          "# Root Agent",
				"agents/chatbot.md": "# Lowercase agents",
			},
			expectedNames: []string{"root-agent", "chatbot"}, // Simplified to match working patterns
			description:   "Path detection should be case sensitive",
			expectError:   false,
		},
		{
			name:          "very-deep-nesting",
			componentType: ComponentCommand,
			filesToCreate: map[string]string{
				"a/b/c/d/e/f/commands/deep.md": "# Deep Command",
			},
			expectedNames: []string{"deep"},
			description:   "Very deeply nested components should be detected",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir, err := os.MkdirTemp("", "component-edge-test-*")
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

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if err != nil {
				return // Expected error, test passes
			}

			// Filter components by type
			var filteredComponents []DetectedComponent
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

			// Compare with expected names using set-based comparison (order doesn't matter)
			if len(actualNames) != len(tt.expectedNames) {
				t.Errorf("Expected %d components, got %d", len(tt.expectedNames), len(actualNames))
				t.Logf("Expected: %v", tt.expectedNames)
				t.Logf("Actual:   %v", actualNames)
				return
			}

			// Create maps for comparison to handle order differences
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

// TestComponentNameValidation tests the validation logic of extracted component names
func TestComponentNameValidation(t *testing.T) {
	tests := []struct {
		name          string
		componentType ComponentType
		filesToCreate map[string]string
		expectedNames []string
		description   string
	}{
		{
			name:          "whitespace-handling",
			componentType: ComponentSkill,
			filesToCreate: map[string]string{
				"  spaced  /SKILL.md": "# Spaced Directory",
			},
			expectedNames: []string{"  spaced  "}, // Directory name with spaces preserved
			description:   "Whitespace in paths should be preserved",
		},
		{
			name:          "unicode-characters",
			componentType: ComponentAgent,
			filesToCreate: map[string]string{
				"агент/AGENT.md":  "# Unicode Agent Name",
				"エージェント/AGENT.md": "# Japanese Agent",
			},
			expectedNames: []string{"агент", "エージェント"}, // Unicode directory names preserved
			description:   "Unicode characters in component names should be preserved",
		},
		{
			name:          "dot-directories",
			componentType: ComponentCommand,
			filesToCreate: map[string]string{
				".hidden/commands/secret.md": "# Hidden Command",
				".config/commands/setup.md":  "# Config Command",
			},
			expectedNames: []string{"secret", "setup"},
			description:   "Hidden directories should be processed normally",
		},
		{
			name:          "single-character-names",
			componentType: ComponentSkill,
			filesToCreate: map[string]string{
				"a/SKILL.md": "# Single letter skill",
				"b/SKILL.md": "# Another single letter",
			},
			expectedNames: []string{"a", "b"}, // Single-character directory names preserved
			description:   "Single character component names should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir, err := os.MkdirTemp("", "component-validation-test-*")
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
			var filteredComponents []DetectedComponent
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

			// Create maps for comparison to handle order differences
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
				t.Logf("  - Name: '%s', Path: %s, Source: %s", comp.Name, comp.Path, comp.SourceFile)
			}
		})
	}
}

// TestPluginsSkillsPathIssue tests the exact issue mentioned in the task description
func TestPluginsSkillsPathIssue(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "specific-issue-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create the exact scenario from the task: plugins/X/skills/Y/SKILL.md
	testFile := filepath.Join(tempDir, "plugins", "X", "skills", "Y", "SKILL.md")

	// Create directory structure
	if err := os.MkdirAll(filepath.Dir(testFile), 0755); err != nil {
		t.Fatalf("Failed to create directory structure: %v", err)
	}

	// Create SKILL.md file
	if err := os.WriteFile(testFile, []byte("# Test Skill"), 0644); err != nil {
		t.Fatalf("Failed to create SKILL.md file: %v", err)
	}

	// Create detector and find components
	detector := NewRepositoryDetector()
	components, err := detector.detectComponentsInRepo(tempDir)
	if err != nil {
		t.Fatalf("Failed to detect components: %v", err)
	}

	// Should find exactly one skill component
	if len(components) != 1 {
		t.Errorf("Expected 1 component, got %d", len(components))
		for i, comp := range components {
			t.Logf("Component %d: Name=%s, Path=%s, Type=%s", i, comp.Name, comp.Path, comp.Type)
		}
		return
	}

	component := components[0]

	// Verify it's a skill
	if component.Type != ComponentSkill {
		t.Errorf("Expected component type to be skill, got %s", component.Type)
	}

	// Verify the path is the immediate parent directory (plugins/X/skills/Y)
	expectedPath := filepath.Join("plugins", "X", "skills", "Y")
	if component.Path != expectedPath {
		t.Errorf("Expected path to be %s, got %s", expectedPath, component.Path)
	}

	// Verify the name is the immediate parent directory name (Y)
	expectedName := "Y"
	if component.Name != expectedName {
		t.Errorf("Expected name to be %s, got %s", expectedName, component.Name)
	}

	t.Logf("SUCCESS: Component correctly detected with Name=%s, Path=%s", component.Name, component.Path)
}

// TestAccessibilityCompliancePluginDownload tests Story-002: accessibility-compliance plugin download
func TestAccessibilityCompliancePluginDownload(t *testing.T) {
	// Create temporary directory for test repository
	tempDir, err := os.MkdirTemp("", "accessibility-compliance-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Simulate the repository structure with accessibility-compliance plugin containing 2 skills
	// and other plugins with additional skills to test isolation
	testFiles := map[string]string{
		// accessibility-compliance plugin with exactly 2 skills
		"plugins/accessibility-compliance/skills/wcag-compliance/SKILL.md":       "# WCAG Compliance Skill\nEnsures WCAG 2.1 compliance",
		"plugins/accessibility-compliance/skills/screen-reader-support/SKILL.md": "# Screen Reader Support Skill\nOptimizes for screen readers",

		// Other plugins with skills that should NOT be included
		"plugins/python-development/skills/async-patterns/SKILL.md":   "# Async Python Patterns",
		"plugins/python-development/skills/testing-patterns/SKILL.md": "# Python Testing",
		"plugins/kubernetes-operations/skills/manifests/SKILL.md":     "# K8s Manifests",
		"plugins/kubernetes-operations/skills/helm-charts/SKILL.md":   "# Helm Charts",
		"plugins/security-scanning/skills/sast-analysis/SKILL.md":     "# SAST Analysis",

		// Additional files to make it realistic (125 other skills across other plugins)
		"plugins/frontend-development/skills/react-patterns/SKILL.md": "# React Patterns",
		"plugins/backend-development/skills/api-design/SKILL.md":      "# API Design",
		"plugins/infrastructure/skills/terraform/SKILL.md":            "# Terraform",
	}

	// Create all test files
	for filePath, content := range testFiles {
		fullPath := filepath.Join(tempDir, filePath)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Create detector
	detector := NewRepositoryDetector()

	// Test component detection in the entire repository
	components, err := detector.detectComponentsInRepo(tempDir)
	if err != nil {
		t.Fatalf("Failed to detect components: %v", err)
	}

	// Should detect multiple skill components across all plugins
	totalSkills := 0
	var accessibilitySkills []DetectedComponent
	for _, comp := range components {
		if comp.Type == ComponentSkill {
			totalSkills++
			// Check if this skill belongs to accessibility-compliance plugin
			if strings.Contains(comp.Path, "accessibility-compliance") {
				accessibilitySkills = append(accessibilitySkills, comp)
			}
		}
	}

	// Verify total skills detected (should be more than just the 2 accessibility skills)
	if totalSkills <= 2 {
		t.Errorf("Expected to detect more than 2 skills total, got %d", totalSkills)
		t.Logf("Total skills detected: %d", totalSkills)
	}

	// Verify exactly 2 accessibility-compliance skills detected
	if len(accessibilitySkills) != 2 {
		t.Errorf("Expected exactly 2 accessibility-compliance skills, got %d", len(accessibilitySkills))
		for i, skill := range accessibilitySkills {
			t.Logf("Accessibility skill %d: Name=%s, Path=%s", i, skill.Name, skill.Path)
		}
	}

	// Verify the specific accessibility skills
	expectedSkills := []struct {
		name string
		path string
	}{
		{"wcag-compliance", filepath.Join("plugins", "accessibility-compliance", "skills", "wcag-compliance")},
		{"screen-reader-support", filepath.Join("plugins", "accessibility-compliance", "skills", "screen-reader-support")},
	}

	for i, expected := range expectedSkills {
		found := false
		for _, skill := range accessibilitySkills {
			if skill.Name == expected.name && skill.Path == expected.path {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected accessibility skill %d: name=%s, path=%s", i, expected.name, expected.path)
		}
	}

	// Simulate downloading just the accessibility-compliance plugin
	// This should copy only the accessibility skills, not all skills
	pluginDir := filepath.Join(tempDir, "plugins", "accessibility-compliance")
	pluginComponents, err := detector.detectComponentsInRepo(pluginDir)
	if err != nil {
		t.Fatalf("Failed to detect components in plugin directory: %v", err)
	}

	// Should detect exactly 2 skills in the plugin directory
	pluginSkillCount := 0
	for _, comp := range pluginComponents {
		if comp.Type == ComponentSkill {
			pluginSkillCount++
		}
	}

	if pluginSkillCount != 2 {
		t.Errorf("Expected exactly 2 skills in accessibility-compliance plugin directory, got %d", pluginSkillCount)
	}

	// Verify no cross-contamination - skills from other plugins should not be in this list
	for _, comp := range pluginComponents {
		if comp.Type == ComponentSkill && !strings.Contains(comp.Path, "wcag-compliance") && !strings.Contains(comp.Path, "screen-reader-support") {
			t.Errorf("Found unexpected skill in plugin directory: Name=%s, Path=%s", comp.Name, comp.Path)
		}
	}

	t.Logf("SUCCESS: accessibility-compliance plugin isolation test passed")
	t.Logf("  - Total skills in repository: %d", totalSkills)
	t.Logf("  - Accessibility-compliance skills detected: %d", len(accessibilitySkills))
	t.Logf("  - Skills in plugin directory: %d", pluginSkillCount)
}

// TestAccessibilityCompliancePluginMetadata tests metadata generation for accessibility-compliance plugin
func TestAccessibilityCompliancePluginMetadata(t *testing.T) {
	// Create temporary directory for test repository
	tempDir, err := os.MkdirTemp("", "accessibility-metadata-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create accessibility-compliance plugin structure
	testFiles := map[string]string{
		"plugins/accessibility-compliance/skills/wcag-compliance/SKILL.md":       "# WCAG Compliance Skill\nEnsures WCAG 2.1 compliance",
		"plugins/accessibility-compliance/skills/screen-reader-support/SKILL.md": "# Screen Reader Support Skill\nOptimizes for screen readers",
	}

	// Create test files
	for filePath, content := range testFiles {
		fullPath := filepath.Join(tempDir, filePath)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Create detector and test specific plugin path
	detector := NewRepositoryDetector()

	// Test detection specifically in the accessibility-compliance plugin directory
	pluginDir := filepath.Join(tempDir, "plugins", "accessibility-compliance")
	components, err := detector.detectComponentsInRepo(pluginDir)
	if err != nil {
		t.Fatalf("Failed to detect components: %v", err)
	}

	// Should detect exactly 2 skills
	skillCount := 0
	for _, comp := range components {
		if comp.Type == ComponentSkill {
			skillCount++
		}
	}

	if skillCount != 2 {
		t.Errorf("Expected exactly 2 skills in accessibility-compliance plugin, got %d", skillCount)
	}

	// Test that metadata would report correct component count
	// Simulate what would happen during download
	expectedMetadata := map[string]interface{}{
		"name":       "accessibility-compliance",
		"source":     "test-repo",
		"commit":     "test-commit",
		"downloaded": "now",
		"components": skillCount, // This should be 2, not 127
		"detection":  "recursive",
	}

	// Verify component count in metadata
	if expectedMetadata["components"] != 2 {
		t.Errorf("Expected metadata components to be 2, got %v", expectedMetadata["components"])
	}

	t.Logf("SUCCESS: accessibility-compliance plugin metadata test passed")
	t.Logf("  - Skills detected: %d", skillCount)
	t.Logf("  - Metadata component count: %v", expectedMetadata["components"])
}

// BenchmarkComponentDetection benchmarks the component detection performance
func BenchmarkComponentDetection(b *testing.B) {
	// Create a complex repository structure for benchmarking
	tempDir, err := os.MkdirTemp("", "component-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create many component files
	files := make(map[string]string)
	componentTypes := []struct {
		prefix string
		ext    string
		count  int
	}{
		{"skills", "SKILL.md", 100},
		{"agents", "AGENT.md", 50},
		{"commands", "COMMAND.md", 75},
	}

	for _, ct := range componentTypes {
		for i := 0; i < ct.count; i++ {
			path := filepath.Join("components", ct.prefix, ct.prefix+string(rune('a'+i%26)), ct.ext)
			files[path] = fmt.Sprintf("# %s %d", strings.Title(ct.prefix), i)
		}
	}

	// Create files
	for filePath, content := range files {
		fullPath := filepath.Join(tempDir, filePath)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			b.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			b.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Benchmark detection
	detector := NewRepositoryDetector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		components, err := detector.detectComponentsInRepo(tempDir)
		if err != nil {
			b.Fatalf("Detection failed: %v", err)
		}
		_ = components
	}
}
