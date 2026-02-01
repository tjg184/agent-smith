package formatter

import (
	"fmt"
	"io"
	"strings"
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
		columnSizes[i] = len(header)
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
			cellLen := len(cell)
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
		fmt.Fprintf(bt.writer, " %-*s %s", bt.columnSizes[i], header, BoxVertical)
	}
	fmt.Fprintln(bt.writer)

	// Print header separator
	bt.printBorder(BoxTeeRight, BoxCross, BoxTeeLeft)

	// Print rows
	for _, row := range bt.rows {
		fmt.Fprint(bt.writer, BoxVertical)
		for i, cell := range row {
			if i < len(bt.columnSizes) {
				fmt.Fprintf(bt.writer, " %-*s %s", bt.columnSizes[i], cell, BoxVertical)
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
