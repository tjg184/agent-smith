package formatter

import (
	"bytes"
	"testing"
)

// TestEnhancedFormatterIntegration tests all enhanced formatter methods together
func TestEnhancedFormatterIntegration(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	// Test Section
	f.Section("Enhanced Formatter Demo")
	output := buf.String()
	if !contains(output, "Enhanced Formatter Demo") {
		t.Errorf("Expected section title in output")
	}
	if !contains(output, BoxHorizontal) {
		t.Errorf("Expected horizontal line in section")
	}

	// Test SuccessWithDetail
	buf.Reset()
	f.SuccessWithDetail("skill", "api-design", "installed to ~/.agent-smith/skills/api-design")
	output = buf.String()
	if !contains(output, SymbolSuccess) {
		t.Errorf("Expected success symbol")
	}
	if !contains(output, "api-design") {
		t.Errorf("Expected component name")
	}
	if !contains(output, "→") {
		t.Errorf("Expected arrow symbol for detail")
	}

	// Test ErrorWithContext
	buf.Reset()
	f.ErrorWithContext("Failed to download", nil, "Check your network")
	output = buf.String()
	if !contains(output, SymbolError) {
		t.Errorf("Expected error symbol")
	}
	if !contains(output, "Try:") {
		t.Errorf("Expected suggestion prefix")
	}

	// Test Divider
	buf.Reset()
	f.Divider()
	output = buf.String()
	if !contains(output, BoxHorizontal) {
		t.Errorf("Expected horizontal line in divider")
	}

	// Test KeyValue
	buf.Reset()
	f.KeyValue("Name", "test-component")
	f.KeyValue("Type", "skill")
	output = buf.String()
	if !contains(output, "Name:") || !contains(output, "test-component") {
		t.Errorf("Expected key-value pair")
	}
	if !contains(output, "Type:") || !contains(output, "skill") {
		t.Errorf("Expected second key-value pair")
	}

	// Test List
	buf.Reset()
	items := []string{"api-design", "event-sourcing"}
	f.List(items)
	output = buf.String()
	if !contains(output, "• api-design") {
		t.Errorf("Expected first list item")
	}
	if !contains(output, "• event-sourcing") {
		t.Errorf("Expected second list item")
	}

	// Test NextSteps
	buf.Reset()
	commands := map[string]string{
		"agent-smith link all": "Link components to targets",
		"agent-smith status":   "View current configuration",
	}
	f.NextSteps(commands)
	output = buf.String()
	if !contains(output, "Next steps:") {
		t.Errorf("Expected next steps header")
	}
	if !contains(output, "agent-smith link all") {
		t.Errorf("Expected command in next steps")
	}
}

// TestFormatterMethodsIntegrateWithColors verifies that formatter methods work with color system
func TestFormatterMethodsIntegrateWithColors(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	// Test all methods produce output (regardless of color status)
	methods := []func(){
		func() { f.Section("Test Section") },
		func() { f.SuccessWithDetail("skill", "test", "detail") },
		func() { f.ErrorWithContext("error", nil, "suggestion") },
		func() { f.Divider() },
		func() { f.KeyValue("key", "value") },
		func() { f.List([]string{"item"}) },
		func() { f.NextSteps(map[string]string{"cmd": "desc"}) },
	}

	for i, method := range methods {
		buf.Reset()
		method()
		output := buf.String()
		if output == "" {
			t.Errorf("Method %d produced no output", i)
		}
	}
}

// TestAllFormatterMethodsRespectTTY ensures methods don't crash with non-TTY output
func TestAllFormatterMethodsRespectTTY(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	// All methods should work regardless of TTY status
	f.Section("Configuration")
	f.SuccessWithDetail("agent", "test-agent", "/path/to/agent")
	f.ErrorWithContext("Failed operation", nil, "Try this")
	f.Divider()
	f.KeyValue("Setting", "Value")
	f.List([]string{"one", "two", "three"})
	f.NextSteps(map[string]string{
		"command1": "description1",
		"command2": "description2",
	})

	output := buf.String()

	// Verify basic structure is present
	if !contains(output, "Configuration") {
		t.Error("Expected section title")
	}
	if !contains(output, "test-agent") {
		t.Error("Expected component name")
	}
	if !contains(output, "Failed operation") {
		t.Error("Expected error message")
	}
	if !contains(output, "Setting") {
		t.Error("Expected key-value")
	}
	if !contains(output, "Next steps:") {
		t.Error("Expected next steps")
	}
}

// Helper function to check if output contains a string
func contains(output, substr string) bool {
	return len(output) > 0 && len(substr) > 0 &&
		len(output) >= len(substr) &&
		findSubstring(output, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
