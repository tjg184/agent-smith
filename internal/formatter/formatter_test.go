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
