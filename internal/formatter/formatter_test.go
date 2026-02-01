package formatter

import (
	"bytes"
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
	if !strings.Contains(output, "Total: 10") {
		t.Errorf("Expected total count, got: %s", output)
	}
	if !strings.Contains(output, "Successful: 7") {
		t.Errorf("Expected success count, got: %s", output)
	}
	if !strings.Contains(output, "Failed: 2") {
		t.Errorf("Expected failed count, got: %s", output)
	}
	if !strings.Contains(output, "Skipped: 1") {
		t.Errorf("Expected skipped count, got: %s", output)
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
