package formatter

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/mattn/go-runewidth"
)

// Box-drawing character constants for table formatting
const (
	BoxTopLeft     = "┌"
	BoxTopRight    = "┐"
	BoxBottomLeft  = "└"
	BoxBottomRight = "┘"
	BoxHorizontal  = "─"
	BoxVertical    = "│"
	BoxTeeDown     = "┬"
	BoxTeeUp       = "┴"
	BoxTeeRight    = "├"
	BoxTeeLeft     = "┤"
	BoxCross       = "┼"
)

// ansiRegex matches ANSI color codes
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// VisibleLength returns the visible length of a string, excluding ANSI color codes
// and accounting for multi-byte Unicode characters
func VisibleLength(s string) int {
	// Remove ANSI color codes first
	stripped := ansiRegex.ReplaceAllString(s, "")

	// Calculate the display width, accounting for emoji variation selectors
	// Many terminals render emoji with variation selector (U+FE0F) as 2 characters wide
	width := 0
	runes := []rune(stripped)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		// Check if this is followed by a variation selector
		hasVariationSelector := false
		if i+1 < len(runes) && runes[i+1] == '\uFE0F' {
			hasVariationSelector = true
		}

		// If the rune is an emoji or has a variation selector, it likely renders as 2 chars
		// Common emoji ranges and symbols that render wide:
		// U+2600-U+26FF (Miscellaneous Symbols) - includes ⚠ (U+26A0)
		// U+1F300-U+1F9FF (Emoji)
		if hasVariationSelector || (r >= 0x1F300 && r <= 0x1F9FF) {
			width += 2
			// Skip the variation selector if present
			if i+1 < len(runes) && runes[i+1] == '\uFE0F' {
				i++
			}
		} else {
			width += runewidth.RuneWidth(r)
		}
	}

	return width
}

// BoxTable represents a table with box-drawing characters
type BoxTable struct {
	writer      io.Writer
	headers     []string
	rows        [][]string
	columnSizes []int
}

// NewBoxTable creates a new box table
func NewBoxTable(writer io.Writer, headers []string) *BoxTable {
	columnSizes := make([]int, len(headers))
	for i, header := range headers {
		columnSizes[i] = VisibleLength(header)
	}
	return &BoxTable{
		writer:      writer,
		headers:     headers,
		rows:        [][]string{},
		columnSizes: columnSizes,
	}
}

// AddRow adds a row to the table
func (bt *BoxTable) AddRow(cells []string) {
	// Update column sizes if needed
	for i, cell := range cells {
		if i < len(bt.columnSizes) {
			cellLen := VisibleLength(cell)
			if cellLen > bt.columnSizes[i] {
				bt.columnSizes[i] = cellLen
			}
		}
	}
	bt.rows = append(bt.rows, cells)
}

// Render renders the table with box-drawing characters
func (bt *BoxTable) Render() {
	// Print top border
	bt.printBorder(BoxTopLeft, BoxTeeDown, BoxTopRight)

	// Print header
	fmt.Fprint(bt.writer, BoxVertical)
	for i, header := range bt.headers {
		// Calculate padding needed to account for ANSI color codes
		visLen := VisibleLength(header)
		padding := bt.columnSizes[i] - visLen
		fmt.Fprintf(bt.writer, " %s%s %s", header, strings.Repeat(" ", padding), BoxVertical)
	}
	fmt.Fprintln(bt.writer)

	// Print header separator
	bt.printBorder(BoxTeeRight, BoxCross, BoxTeeLeft)

	// Print rows
	for _, row := range bt.rows {
		fmt.Fprint(bt.writer, BoxVertical)
		for i, cell := range row {
			if i < len(bt.columnSizes) {
				// Calculate padding needed to account for ANSI color codes
				visLen := VisibleLength(cell)
				padding := bt.columnSizes[i] - visLen
				fmt.Fprintf(bt.writer, " %s%s %s", cell, strings.Repeat(" ", padding), BoxVertical)
			}
		}
		fmt.Fprintln(bt.writer)
	}

	// Print bottom border
	bt.printBorder(BoxBottomLeft, BoxTeeUp, BoxBottomRight)
}

// printBorder prints a horizontal border line
func (bt *BoxTable) printBorder(left, middle, right string) {
	fmt.Fprint(bt.writer, left)
	for i, size := range bt.columnSizes {
		fmt.Fprint(bt.writer, strings.Repeat(BoxHorizontal, size+2))
		if i < len(bt.columnSizes)-1 {
			fmt.Fprint(bt.writer, middle)
		}
	}
	fmt.Fprintln(bt.writer, right)
}

// SimpleBoxTable is a helper for quickly rendering a simple table
func SimpleBoxTable(writer io.Writer, headers []string, rows [][]string) {
	table := NewBoxTable(writer, headers)
	for _, row := range rows {
		table.AddRow(row)
	}
	table.Render()
}
