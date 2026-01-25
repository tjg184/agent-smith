package main

import (
	"path/filepath"
	"testing"
)

// TestExtractPluginPath tests the extractPluginPath function with various path formats
func TestExtractPluginPath(t *testing.T) {
	tests := []struct {
		name          string
		componentPath string
		expectedPath  string
		description   string
	}{
		{
			name:          "plugin-ui-design",
			componentPath: "plugins/ui-design/agents/accessibility-expert.md",
			expectedPath:  "plugins/ui-design",
			description:   "Extract plugin path from nested agent path",
		},
		{
			name:          "plugin-python-development",
			componentPath: "plugins/python-development/skills/async-patterns/SKILL.md",
			expectedPath:  "plugins/python-development",
			description:   "Extract plugin path from deeply nested skill path",
		},
		{
			name:          "plugin-accessibility-compliance",
			componentPath: "plugins/accessibility-compliance/skills/wcag-compliance",
			expectedPath:  "plugins/accessibility-compliance",
			description:   "Extract plugin path from directory path without file",
		},
		{
			name:          "no-plugin-path",
			componentPath: "agents/chatbot.md",
			expectedPath:  "",
			description:   "Return empty string for non-plugin path",
		},
		{
			name:          "no-plugin-skill",
			componentPath: "skills/python/SKILL.md",
			expectedPath:  "",
			description:   "Return empty string for non-plugin skill path",
		},
		{
			name:          "root-level-file",
			componentPath: "SKILL.md",
			expectedPath:  "",
			description:   "Return empty string for root-level file",
		},
		{
			name:          "empty-path",
			componentPath: "",
			expectedPath:  "",
			description:   "Return empty string for empty path",
		},
		{
			name:          "plugins-only",
			componentPath: "plugins",
			expectedPath:  "",
			description:   "Return empty string if path is just 'plugins' directory",
		},
		{
			name:          "plugins-slash-only",
			componentPath: "plugins/",
			expectedPath:  "",
			description:   "Return empty string if path is 'plugins/' with no plugin name",
		},
		{
			name:          "windows-style-path",
			componentPath: "plugins\\ui-design\\agents\\accessibility-expert.md",
			expectedPath:  filepath.Join("plugins", "ui-design"),
			description:   "Handle Windows-style path separators",
		},
		{
			name:          "mixed-separators",
			componentPath: "plugins/ui-design\\agents/accessibility-expert.md",
			expectedPath:  filepath.Join("plugins", "ui-design"),
			description:   "Handle mixed path separators",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPluginPath(tt.componentPath)
			if result != tt.expectedPath {
				t.Errorf("extractPluginPath(%q) = %q; expected %q (%s)",
					tt.componentPath, result, tt.expectedPath, tt.description)
			}
		})
	}
}

// TestDetectCommonPluginPath tests the detectCommonPluginPath function
func TestDetectCommonPluginPath(t *testing.T) {
	tests := []struct {
		name         string
		components   []DetectedComponent
		expectedPath string
		description  string
	}{
		{
			name: "same-plugin-multiple-components",
			components: []DetectedComponent{
				{Type: ComponentAgent, Name: "accessibility-expert", Path: "plugins/ui-design/agents/accessibility-expert.md"},
				{Type: ComponentAgent, Name: "design-system-architect", Path: "plugins/ui-design/agents/design-system-architect.md"},
				{Type: ComponentAgent, Name: "ux-researcher", Path: "plugins/ui-design/agents/ux-researcher.md"},
			},
			expectedPath: "plugins/ui-design",
			description:  "All components from same plugin should return plugin path",
		},
		{
			name: "same-plugin-mixed-types",
			components: []DetectedComponent{
				{Type: ComponentAgent, Name: "accessibility-expert", Path: "plugins/ui-design/agents/accessibility-expert.md"},
				{Type: ComponentSkill, Name: "wcag-compliance", Path: "plugins/ui-design/skills/wcag-compliance/SKILL.md"},
				{Type: ComponentCommand, Name: "contrast-check", Path: "plugins/ui-design/commands/contrast-check.md"},
			},
			expectedPath: "plugins/ui-design",
			description:  "Mixed component types from same plugin should return plugin path",
		},
		{
			name: "different-plugins",
			components: []DetectedComponent{
				{Type: ComponentAgent, Name: "accessibility-expert", Path: "plugins/ui-design/agents/accessibility-expert.md"},
				{Type: ComponentSkill, Name: "async-patterns", Path: "plugins/python-development/skills/async-patterns/SKILL.md"},
			},
			expectedPath: "",
			description:  "Components from different plugins should return empty string",
		},
		{
			name: "no-plugin-structure",
			components: []DetectedComponent{
				{Type: ComponentAgent, Name: "chatbot", Path: "agents/chatbot.md"},
				{Type: ComponentAgent, Name: "helper", Path: "agents/helper.md"},
			},
			expectedPath: "",
			description:  "Components not in plugin structure should return empty string",
		},
		{
			name: "mixed-plugin-and-non-plugin",
			components: []DetectedComponent{
				{Type: ComponentAgent, Name: "accessibility-expert", Path: "plugins/ui-design/agents/accessibility-expert.md"},
				{Type: ComponentAgent, Name: "chatbot", Path: "agents/chatbot.md"},
			},
			expectedPath: "",
			description:  "Mixed plugin and non-plugin components should return empty string",
		},
		{
			name:         "empty-component-list",
			components:   []DetectedComponent{},
			expectedPath: "",
			description:  "Empty component list should return empty string",
		},
		{
			name: "single-plugin-component",
			components: []DetectedComponent{
				{Type: ComponentAgent, Name: "accessibility-expert", Path: "plugins/ui-design/agents/accessibility-expert.md"},
			},
			expectedPath: "plugins/ui-design",
			description:  "Single component in plugin should return plugin path",
		},
		{
			name: "single-non-plugin-component",
			components: []DetectedComponent{
				{Type: ComponentAgent, Name: "chatbot", Path: "agents/chatbot.md"},
			},
			expectedPath: "",
			description:  "Single component not in plugin should return empty string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectCommonPluginPath(tt.components)
			if result != tt.expectedPath {
				t.Errorf("detectCommonPluginPath() = %q; expected %q (%s)",
					result, tt.expectedPath, tt.description)
			}
		})
	}
}

// TestExtractPluginPathCrossPlatform tests cross-platform path handling
func TestExtractPluginPathCrossPlatform(t *testing.T) {
	tests := []struct {
		name          string
		componentPath string
		description   string
	}{
		{
			name:          "unix-separators",
			componentPath: "plugins/ui-design/agents/accessibility-expert.md",
			description:   "Unix-style forward slashes",
		},
		{
			name:          "windows-separators",
			componentPath: "plugins\\ui-design\\agents\\accessibility-expert.md",
			description:   "Windows-style backslashes",
		},
	}

	expectedPluginPath := filepath.Join("plugins", "ui-design")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPluginPath(tt.componentPath)
			if result != expectedPluginPath {
				t.Errorf("extractPluginPath(%q) = %q; expected %q (%s)",
					tt.componentPath, result, expectedPluginPath, tt.description)
			}
		})
	}
}
