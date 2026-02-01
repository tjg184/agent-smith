package formatter

import (
	"strings"
)

// Default box width (80 characters including borders)
const DefaultBoxWidth = 80

// DrawBox creates a complete bordered box with a title and content
// The title is centered in the top border, and content is padded inside
func DrawBox(title, content string, width int) string {
	if width <= 0 {
		width = DefaultBoxWidth
	}

	var result strings.Builder

	// Draw top border with title
	result.WriteString(DrawHeader(title, width))
	result.WriteString("\n")

	// Draw content lines
	if content != "" {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			result.WriteString(formatContentLine(line, width))
			result.WriteString("\n")
		}
	}

	// Draw bottom border
	result.WriteString(DrawFooter(width))

	return result.String()
}

// DrawHeader creates a top border with an optional centered title
// If title is empty, creates a plain top border
func DrawHeader(title string, width int) string {
	if width <= 0 {
		width = DefaultBoxWidth
	}

	// Inner width excludes the two border characters
	innerWidth := width - 2

	if title == "" {
		// Plain top border: ┌──────┐
		return BoxTopLeft + strings.Repeat(BoxHorizontal, innerWidth) + BoxTopRight
	}

	// Calculate padding for centered title
	// Format: ┌─── Title ───┐
	titleWithSpaces := " " + title + " "
	titleLen := len(titleWithSpaces)

	if titleLen >= innerWidth {
		// Title too long, truncate it
		titleWithSpaces = " " + title[:innerWidth-5] + "... "
		titleLen = innerWidth
	}

	leftPadding := (innerWidth - titleLen) / 2
	rightPadding := innerWidth - titleLen - leftPadding

	return BoxTopLeft +
		strings.Repeat(BoxHorizontal, leftPadding) +
		titleWithSpaces +
		strings.Repeat(BoxHorizontal, rightPadding) +
		BoxTopRight
}

// DrawSeparator creates a horizontal separator line inside a box
// Format: ├───────┤
func DrawSeparator(width int) string {
	if width <= 0 {
		width = DefaultBoxWidth
	}

	innerWidth := width - 2
	return BoxTeeRight + strings.Repeat(BoxHorizontal, innerWidth) + BoxTeeLeft
}

// DrawFooter creates a bottom border
// Format: └───────┘
func DrawFooter(width int) string {
	if width <= 0 {
		width = DefaultBoxWidth
	}

	innerWidth := width - 2
	return BoxBottomLeft + strings.Repeat(BoxHorizontal, innerWidth) + BoxBottomRight
}

// formatContentLine formats a single line of content with borders and padding
// Format: │ content here                                  │
func formatContentLine(content string, width int) string {
	if width <= 0 {
		width = DefaultBoxWidth
	}

	// Inner width excludes the two border characters
	innerWidth := width - 2

	// Account for padding (1 space on each side)
	contentWidth := innerWidth - 2

	// Truncate if content is too long
	if len(content) > contentWidth {
		content = content[:contentWidth-3] + "..."
	}

	// Pad to full width
	padding := contentWidth - len(content)

	return BoxVertical + " " + content + strings.Repeat(" ", padding) + " " + BoxVertical
}

// DrawMultilineBox creates a box with multiple lines of content
// Each line in the content string will be formatted individually
func DrawMultilineBox(title string, lines []string, width int) string {
	if width <= 0 {
		width = DefaultBoxWidth
	}

	var result strings.Builder

	// Draw top border with title
	result.WriteString(DrawHeader(title, width))
	result.WriteString("\n")

	// Draw each content line
	for _, line := range lines {
		result.WriteString(formatContentLine(line, width))
		result.WriteString("\n")
	}

	// Draw bottom border
	result.WriteString(DrawFooter(width))

	return result.String()
}

// DrawBoxWithSections creates a box with multiple sections separated by horizontal lines
// sections is a slice of {title, content} pairs
func DrawBoxWithSections(boxTitle string, sections []Section, width int) string {
	if width <= 0 {
		width = DefaultBoxWidth
	}

	var result strings.Builder

	// Draw top border with title
	result.WriteString(DrawHeader(boxTitle, width))
	result.WriteString("\n")

	// Draw each section
	for i, section := range sections {
		// Add section title if present
		if section.Title != "" {
			result.WriteString(formatContentLine(section.Title, width))
			result.WriteString("\n")
		}

		// Add section content lines
		lines := strings.Split(section.Content, "\n")
		for _, line := range lines {
			result.WriteString(formatContentLine(line, width))
			result.WriteString("\n")
		}

		// Add separator between sections (but not after the last one)
		if i < len(sections)-1 {
			result.WriteString(DrawSeparator(width))
			result.WriteString("\n")
		}
	}

	// Draw bottom border
	result.WriteString(DrawFooter(width))

	return result.String()
}

// Section represents a section within a box
type Section struct {
	Title   string
	Content string
}
