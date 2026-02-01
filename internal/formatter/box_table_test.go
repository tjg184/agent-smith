package formatter

import (
	"bytes"
	"strings"
	"testing"
)

func TestBoxTable(t *testing.T) {
	var buf bytes.Buffer

	table := NewBoxTable(&buf, []string{"Name", "Status", "Type"})
	table.AddRow([]string{"component-1", "✓", "skill"})
	table.AddRow([]string{"component-2", "✗", "agent"})
	table.AddRow([]string{"component-3", "◆", "command"})

	table.Render()

	output := buf.String()

	// Verify box-drawing characters are present
	if !strings.Contains(output, BoxTopLeft) {
		t.Error("Expected top-left corner character")
	}
	if !strings.Contains(output, BoxTopRight) {
		t.Error("Expected top-right corner character")
	}
	if !strings.Contains(output, BoxBottomLeft) {
		t.Error("Expected bottom-left corner character")
	}
	if !strings.Contains(output, BoxBottomRight) {
		t.Error("Expected bottom-right corner character")
	}
	if !strings.Contains(output, BoxHorizontal) {
		t.Error("Expected horizontal line character")
	}
	if !strings.Contains(output, BoxVertical) {
		t.Error("Expected vertical line character")
	}
	if !strings.Contains(output, BoxTeeDown) {
		t.Error("Expected tee-down character")
	}
	if !strings.Contains(output, BoxCross) {
		t.Error("Expected cross character")
	}

	// Verify content is present
	if !strings.Contains(output, "component-1") {
		t.Error("Expected to find component-1 in output")
	}
	if !strings.Contains(output, "component-2") {
		t.Error("Expected to find component-2 in output")
	}
	if !strings.Contains(output, "component-3") {
		t.Error("Expected to find component-3 in output")
	}

	// Verify headers are present
	if !strings.Contains(output, "Name") {
		t.Error("Expected to find Name header in output")
	}
	if !strings.Contains(output, "Status") {
		t.Error("Expected to find Status header in output")
	}
	if !strings.Contains(output, "Type") {
		t.Error("Expected to find Type header in output")
	}
}

func TestSimpleBoxTable(t *testing.T) {
	var buf bytes.Buffer

	headers := []string{"Col1", "Col2"}
	rows := [][]string{
		{"Value1", "Value2"},
		{"Value3", "Value4"},
	}

	SimpleBoxTable(&buf, headers, rows)

	output := buf.String()

	// Verify it produces output
	if len(output) == 0 {
		t.Error("Expected non-empty output")
	}

	// Verify content
	if !strings.Contains(output, "Value1") {
		t.Error("Expected to find Value1 in output")
	}
	if !strings.Contains(output, "Value2") {
		t.Error("Expected to find Value2 in output")
	}
}
