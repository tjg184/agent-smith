package formatter

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestSectionHeader(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.SectionHeader("Test Section")

	output := buf.String()
	if !strings.Contains(output, "=== Test Section ===") {
		t.Errorf("Expected section header to contain '=== Test Section ===', got: %s", output)
	}
}

func TestSubsectionHeader(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.SubsectionHeader("Test Subsection")

	output := buf.String()
	if !strings.Contains(output, "--- Test Subsection ---") {
		t.Errorf("Expected subsection header to contain '--- Test Subsection ---', got: %s", output)
	}
}

func TestProgressMessages(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.ProgressMsg("Installing", "test-component")
	f.ProgressComplete()

	output := buf.String()
	if !strings.Contains(output, "Installing: test-component...") {
		t.Errorf("Expected progress message, got: %s", output)
	}
	if !strings.Contains(output, "Done") {
		t.Errorf("Expected completion message, got: %s", output)
	}
}

func TestSuccessMsg(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.SuccessMsg("Operation completed successfully")

	output := buf.String()
	if !strings.Contains(output, SymbolSuccess) {
		t.Errorf("Expected success symbol, got: %s", output)
	}
	if !strings.Contains(output, "Operation completed successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}
}

func TestErrorMsg(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.ErrorMsg("Operation failed")

	output := buf.String()
	if !strings.Contains(output, SymbolError) {
		t.Errorf("Expected error symbol, got: %s", output)
	}
	if !strings.Contains(output, "Operation failed") {
		t.Errorf("Expected error message, got: %s", output)
	}
}

func TestWarningMsg(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.WarningMsg("This is a warning")

	output := buf.String()
	if !strings.Contains(output, SymbolWarning) {
		t.Errorf("Expected warning symbol, got: %s", output)
	}
	if !strings.Contains(output, "This is a warning") {
		t.Errorf("Expected warning message, got: %s", output)
	}
}

func TestInfoMsg(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.InfoMsg("This is info")

	output := buf.String()
	if !strings.Contains(output, "• This is info") {
		t.Errorf("Expected info message with bullet, got: %s", output)
	}
}

func TestListItem(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.ListItem("Item 1")
	f.ListItem("Item 2")

	output := buf.String()
	if !strings.Contains(output, "  • Item 1") {
		t.Errorf("Expected list item 1, got: %s", output)
	}
	if !strings.Contains(output, "  • Item 2") {
		t.Errorf("Expected list item 2, got: %s", output)
	}
}

func TestDetailItem(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.DetailItem("Name", "test-component")

	output := buf.String()
	if !strings.Contains(output, "Name") {
		t.Errorf("Expected detail key, got: %s", output)
	}
	if !strings.Contains(output, "test-component") {
		t.Errorf("Expected detail value, got: %s", output)
	}
}

func TestCounterSummary(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.CounterSummary(10, 7, 2, 1)

	output := buf.String()
	// Check for box-drawing characters
	if !strings.Contains(output, "┌") || !strings.Contains(output, "└") {
		t.Errorf("Expected box-drawing characters in output, got: %s", output)
	}
	// Check for total count (in table format)
	if !strings.Contains(output, "Total") && !strings.Contains(output, "10") {
		t.Errorf("Expected total count in table, got: %s", output)
	}
	// Check for successful count (with symbol)
	if !strings.Contains(output, "Successful") && !strings.Contains(output, "7") {
		t.Errorf("Expected success count in table, got: %s", output)
	}
	// Check for failed count (with symbol)
	if !strings.Contains(output, "Failed") && !strings.Contains(output, "2") {
		t.Errorf("Expected failed count in table, got: %s", output)
	}
	// Check for skipped count (with symbol)
	if !strings.Contains(output, "Skipped") && !strings.Contains(output, "1") {
		t.Errorf("Expected skipped count in table, got: %s", output)
	}
}

func TestColoredWarning(t *testing.T) {
	result := ColoredWarning()
	if !strings.Contains(result, SymbolWarning) {
		t.Errorf("Expected warning symbol in colored warning, got: %s", result)
	}
}

func TestInlineSuccess(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.InlineSuccess("Linking", "test-component")

	output := buf.String()
	if !strings.Contains(output, "Linking: test-component...") {
		t.Errorf("Expected inline success message, got: %s", output)
	}
	if !strings.Contains(output, "Done") {
		t.Errorf("Expected 'Done' in output, got: %s", output)
	}
}

func TestInlineSuccessWithNote(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.InlineSuccessWithNote("Linking", "test-component", "from profile: dev")

	output := buf.String()
	if !strings.Contains(output, "Linking: test-component...") {
		t.Errorf("Expected inline success message, got: %s", output)
	}
	if !strings.Contains(output, "Done") {
		t.Errorf("Expected 'Done' in output, got: %s", output)
	}
	if !strings.Contains(output, "from profile: dev") {
		t.Errorf("Expected note in output, got: %s", output)
	}
}

func TestInlineFailed(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.InlineFailed("Linking", "test-component")

	output := buf.String()
	if !strings.Contains(output, "Linking: test-component...") {
		t.Errorf("Expected inline failed message, got: %s", output)
	}
	if !strings.Contains(output, "Failed") {
		t.Errorf("Expected 'Failed' in output, got: %s", output)
	}
}

func TestStatusSuccess(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.StatusSuccess("Successfully activated profile '%s'", "dev")

	output := buf.String()
	if !strings.Contains(output, SymbolSuccess) {
		t.Errorf("Expected success symbol, got: %s", output)
	}
	if !strings.Contains(output, "Successfully activated profile 'dev'") {
		t.Errorf("Expected formatted message, got: %s", output)
	}
}

func TestStatusError(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.StatusError("Error: %v", "something went wrong")

	output := buf.String()
	if !strings.Contains(output, SymbolError) {
		t.Errorf("Expected error symbol, got: %s", output)
	}
	if !strings.Contains(output, "Error: something went wrong") {
		t.Errorf("Expected formatted message, got: %s", output)
	}
}

func TestStatusUpToDate(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.StatusUpToDate()

	output := buf.String()
	if !strings.Contains(output, SymbolSuccess) {
		t.Errorf("Expected success symbol, got: %s", output)
	}
	if !strings.Contains(output, "Up to date") {
		t.Errorf("Expected 'Up to date' message, got: %s", output)
	}
}

func TestStatusUpdating(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.StatusUpdating()

	output := buf.String()
	if !strings.Contains(output, SymbolUpdating) {
		t.Errorf("Expected updating symbol, got: %s", output)
	}
	if !strings.Contains(output, "Updating") {
		t.Errorf("Expected 'Updating' message, got: %s", output)
	}
}

func TestIndentedDetail(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.IndentedDetail("name", "test-value")

	output := buf.String()
	if !strings.Contains(output, "→") {
		t.Errorf("Expected arrow symbol, got: %s", output)
	}
	if !strings.Contains(output, "name: test-value") {
		t.Errorf("Expected detail message, got: %s", output)
	}
}

func TestIndentedError(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.IndentedError("Failed to process %s", "test-item")

	output := buf.String()
	if !strings.Contains(output, SymbolError) {
		t.Errorf("Expected error symbol, got: %s", output)
	}
	if !strings.Contains(output, "Failed to process test-item") {
		t.Errorf("Expected formatted message, got: %s", output)
	}
	// Check indentation (2 spaces)
	if !strings.HasPrefix(strings.TrimLeft(output, "\n"), "  ") {
		t.Errorf("Expected message to be indented, got: %s", output)
	}
}

func TestIndentedSuccess(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.IndentedSuccess("Updated successfully")

	output := buf.String()
	if !strings.Contains(output, SymbolSuccess) {
		t.Errorf("Expected success symbol, got: %s", output)
	}
	if !strings.Contains(output, "Updated successfully") {
		t.Errorf("Expected message, got: %s", output)
	}
	// Check indentation (2 spaces)
	if !strings.HasPrefix(strings.TrimLeft(output, "\n"), "  ") {
		t.Errorf("Expected message to be indented, got: %s", output)
	}
}

func TestPlainWarning(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.PlainWarning("%s has no commit hash stored", "component/name")

	output := buf.String()
	if !strings.Contains(output, "Warning:") {
		t.Errorf("Expected 'Warning:' prefix, got: %s", output)
	}
	if !strings.Contains(output, "component/name has no commit hash stored") {
		t.Errorf("Expected formatted message, got: %s", output)
	}
}

func TestSuccessWithDetail(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.SuccessWithDetail("skill", "api-design", "/path/to/skill")

	output := buf.String()
	if !strings.Contains(output, SymbolSuccess) {
		t.Errorf("Expected success symbol, got: %s", output)
	}
	if !strings.Contains(output, "Successfully skill: api-design") {
		t.Errorf("Expected success message with component type and name, got: %s", output)
	}
	if !strings.Contains(output, "→") {
		t.Errorf("Expected arrow symbol for detail, got: %s", output)
	}
	if !strings.Contains(output, "/path/to/skill") {
		t.Errorf("Expected detail information, got: %s", output)
	}
}

func TestSuccessWithDetail_NoDetail(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.SuccessWithDetail("agent", "test-agent", "")

	output := buf.String()
	if !strings.Contains(output, SymbolSuccess) {
		t.Errorf("Expected success symbol, got: %s", output)
	}
	if !strings.Contains(output, "Successfully agent: test-agent") {
		t.Errorf("Expected success message, got: %s", output)
	}
	// Should not have arrow symbol when no detail
	lines := strings.Split(output, "\n")
	if len(lines) > 2 {
		t.Errorf("Expected only one line (plus newline) when no detail provided, got: %d lines", len(lines))
	}
}

func TestErrorWithContext(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.ErrorWithContext("Failed to download component", nil, "Check your network connection")

	output := buf.String()
	if !strings.Contains(output, SymbolError) {
		t.Errorf("Expected error symbol, got: %s", output)
	}
	if !strings.Contains(output, "Failed to download component") {
		t.Errorf("Expected error message, got: %s", output)
	}
	if !strings.Contains(output, "Try: Check your network connection") {
		t.Errorf("Expected suggestion, got: %s", output)
	}

	// Test with actual error
	buf.Reset()
	testErr := errors.New("test error message")
	f.ErrorWithContext("Failed to read file", testErr, "")
	output = buf.String()
	if !strings.Contains(output, "└─") {
		t.Errorf("Expected error detail symbol, got: %s", output)
	}
	if !strings.Contains(output, "test error message") {
		t.Errorf("Expected error message in detail, got: %s", output)
	}
}

func TestSection(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.Section("Configuration")

	output := buf.String()
	if !strings.Contains(output, "Configuration") {
		t.Errorf("Expected section title, got: %s", output)
	}
	// Should have underline with horizontal line characters
	if !strings.Contains(output, BoxHorizontal) {
		t.Errorf("Expected horizontal line for underline, got: %s", output)
	}
	// Check that it starts with a newline
	if !strings.HasPrefix(output, "\n") {
		t.Errorf("Expected section to start with newline, got: %s", output)
	}
}

func TestDivider(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.Divider()

	output := buf.String()
	if !strings.Contains(output, BoxHorizontal) {
		t.Errorf("Expected horizontal line character, got: %s", output)
	}
	// Should be a full line of horizontal characters
	lineCount := strings.Count(strings.TrimSpace(output), BoxHorizontal)
	if lineCount != 40 {
		t.Errorf("Expected 40 horizontal characters, got: %d", lineCount)
	}
}

func TestKeyValue(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.KeyValue("Name", "test-component")
	f.KeyValue("Type", "skill")

	output := buf.String()
	if !strings.Contains(output, "Name:") {
		t.Errorf("Expected 'Name:' key, got: %s", output)
	}
	if !strings.Contains(output, "test-component") {
		t.Errorf("Expected value, got: %s", output)
	}
	if !strings.Contains(output, "Type:") {
		t.Errorf("Expected 'Type:' key, got: %s", output)
	}
	if !strings.Contains(output, "skill") {
		t.Errorf("Expected value, got: %s", output)
	}
}

func TestList(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	items := []string{"First item", "Second item", "Third item"}
	f.List(items)

	output := buf.String()
	if !strings.Contains(output, "• First item") {
		t.Errorf("Expected first bullet item, got: %s", output)
	}
	if !strings.Contains(output, "• Second item") {
		t.Errorf("Expected second bullet item, got: %s", output)
	}
	if !strings.Contains(output, "• Third item") {
		t.Errorf("Expected third bullet item, got: %s", output)
	}
	// Check that all three items are present
	bulletCount := strings.Count(output, "•")
	if bulletCount != 3 {
		t.Errorf("Expected 3 bullet points, got: %d", bulletCount)
	}
}

func TestList_Empty(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.List([]string{})

	output := buf.String()
	if output != "" {
		t.Errorf("Expected empty output for empty list, got: %s", output)
	}
}

func TestNextSteps(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	commands := map[string]string{
		"agent-smith link all": "Link components to targets",
		"agent-smith status":   "View current configuration",
	}
	f.NextSteps(commands)

	output := buf.String()
	if !strings.Contains(output, "Next steps:") {
		t.Errorf("Expected 'Next steps:' header, got: %s", output)
	}
	if !strings.Contains(output, "agent-smith link all") {
		t.Errorf("Expected first command, got: %s", output)
	}
	if !strings.Contains(output, "Link components to targets") {
		t.Errorf("Expected first description, got: %s", output)
	}
	if !strings.Contains(output, "agent-smith status") {
		t.Errorf("Expected second command, got: %s", output)
	}
	if !strings.Contains(output, "View current configuration") {
		t.Errorf("Expected second description, got: %s", output)
	}
	// Should have bullet points for each command
	bulletCount := strings.Count(output, "•")
	if bulletCount != 2 {
		t.Errorf("Expected 2 bullet points, got: %d", bulletCount)
	}
}

func TestNextSteps_Empty(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWithWriter(buf)

	f.NextSteps(map[string]string{})

	output := buf.String()
	if !strings.Contains(output, "Next steps:") {
		t.Errorf("Expected 'Next steps:' header even for empty map, got: %s", output)
	}
}
