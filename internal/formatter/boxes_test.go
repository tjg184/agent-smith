package formatter

import (
	"strings"
	"testing"
)

func TestDrawHeader(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		width    int
		contains []string
	}{
		{
			name:  "header with title",
			title: "Test Title",
			width: 40,
			contains: []string{
				BoxTopLeft,
				BoxTopRight,
				"Test Title",
				BoxHorizontal,
			},
		},
		{
			name:  "header without title",
			title: "",
			width: 40,
			contains: []string{
				BoxTopLeft,
				BoxTopRight,
				BoxHorizontal,
			},
		},
		{
			name:  "header with default width",
			title: "Default",
			width: 0,
			contains: []string{
				BoxTopLeft,
				BoxTopRight,
				"Default",
			},
		},
		{
			name:  "header with very long title",
			title: "This is a very long title that should be truncated to fit within the box width limits",
			width: 30,
			contains: []string{
				BoxTopLeft,
				BoxTopRight,
				"...",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DrawHeader(tt.title, tt.width)

			// Check that all required characters are present
			for _, str := range tt.contains {
				if !strings.Contains(result, str) {
					t.Errorf("DrawHeader() missing expected string %q\nGot: %s", str, result)
				}
			}

			// Verify width is correct (count runes, not bytes)
			expectedWidth := tt.width
			if expectedWidth <= 0 {
				expectedWidth = DefaultBoxWidth
			}
			actualWidth := len([]rune(result))
			if actualWidth != expectedWidth {
				t.Errorf("DrawHeader() width = %d, want %d\nGot: %s", actualWidth, expectedWidth, result)
			}

			// Verify starts with top-left and ends with top-right
			if !strings.HasPrefix(result, BoxTopLeft) {
				t.Errorf("DrawHeader() should start with BoxTopLeft")
			}
			if !strings.HasSuffix(result, BoxTopRight) {
				t.Errorf("DrawHeader() should end with BoxTopRight")
			}
		})
	}
}

func TestDrawFooter(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{
			name:  "footer with width 40",
			width: 40,
		},
		{
			name:  "footer with default width",
			width: 0,
		},
		{
			name:  "footer with width 20",
			width: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DrawFooter(tt.width)

			// Verify contains box-drawing characters
			if !strings.Contains(result, BoxBottomLeft) {
				t.Error("DrawFooter() missing BoxBottomLeft")
			}
			if !strings.Contains(result, BoxBottomRight) {
				t.Error("DrawFooter() missing BoxBottomRight")
			}
			if !strings.Contains(result, BoxHorizontal) {
				t.Error("DrawFooter() missing BoxHorizontal")
			}

			// Verify width (count runes, not bytes)
			expectedWidth := tt.width
			if expectedWidth <= 0 {
				expectedWidth = DefaultBoxWidth
			}
			actualWidth := len([]rune(result))
			if actualWidth != expectedWidth {
				t.Errorf("DrawFooter() width = %d, want %d", actualWidth, expectedWidth)
			}

			// Verify starts with bottom-left and ends with bottom-right
			if !strings.HasPrefix(result, BoxBottomLeft) {
				t.Error("DrawFooter() should start with BoxBottomLeft")
			}
			if !strings.HasSuffix(result, BoxBottomRight) {
				t.Error("DrawFooter() should end with BoxBottomRight")
			}
		})
	}
}

func TestDrawSeparator(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{
			name:  "separator with width 40",
			width: 40,
		},
		{
			name:  "separator with default width",
			width: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DrawSeparator(tt.width)

			// Verify contains required characters
			if !strings.Contains(result, BoxTeeRight) {
				t.Error("DrawSeparator() missing BoxTeeRight")
			}
			if !strings.Contains(result, BoxTeeLeft) {
				t.Error("DrawSeparator() missing BoxTeeLeft")
			}
			if !strings.Contains(result, BoxHorizontal) {
				t.Error("DrawSeparator() missing BoxHorizontal")
			}

			// Verify width (count runes, not bytes)
			expectedWidth := tt.width
			if expectedWidth <= 0 {
				expectedWidth = DefaultBoxWidth
			}
			actualWidth := len([]rune(result))
			if actualWidth != expectedWidth {
				t.Errorf("DrawSeparator() width = %d, want %d", actualWidth, expectedWidth)
			}

			// Verify starts with tee-right and ends with tee-left
			if !strings.HasPrefix(result, BoxTeeRight) {
				t.Error("DrawSeparator() should start with BoxTeeRight")
			}
			if !strings.HasSuffix(result, BoxTeeLeft) {
				t.Error("DrawSeparator() should end with BoxTeeLeft")
			}
		})
	}
}

func TestFormatContentLine(t *testing.T) {
	tests := []struct {
		name    string
		content string
		width   int
	}{
		{
			name:    "normal content",
			content: "Hello World",
			width:   40,
		},
		{
			name:    "empty content",
			content: "",
			width:   40,
		},
		{
			name:    "very long content",
			content: "This is a very long line that should be truncated to fit within the specified width",
			width:   30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatContentLine(tt.content, tt.width)

			// Verify width (count runes, not bytes)
			expectedWidth := tt.width
			if expectedWidth <= 0 {
				expectedWidth = DefaultBoxWidth
			}
			actualWidth := len([]rune(result))
			if actualWidth != expectedWidth {
				t.Errorf("formatContentLine() width = %d, want %d\nGot: %s", actualWidth, expectedWidth, result)
			}

			// Verify starts and ends with vertical bars
			if !strings.HasPrefix(result, BoxVertical) {
				t.Error("formatContentLine() should start with BoxVertical")
			}
			if !strings.HasSuffix(result, BoxVertical) {
				t.Error("formatContentLine() should end with BoxVertical")
			}

			// Verify content is present (or truncated marker)
			if tt.content != "" {
				contentWidth := tt.width - 4 // 2 for borders, 2 for padding
				if len(tt.content) > contentWidth {
					if !strings.Contains(result, "...") {
						t.Error("formatContentLine() should contain truncation marker for long content")
					}
				} else {
					if !strings.Contains(result, tt.content) {
						t.Errorf("formatContentLine() should contain content %q\nGot: %s", tt.content, result)
					}
				}
			}
		})
	}
}

func TestDrawBox(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		content string
		width   int
	}{
		{
			name:    "box with title and content",
			title:   "Test Box",
			content: "Line 1\nLine 2\nLine 3",
			width:   50,
		},
		{
			name:    "box with title only",
			title:   "Empty Box",
			content: "",
			width:   40,
		},
		{
			name:    "box without title",
			title:   "",
			content: "Just content",
			width:   40,
		},
		{
			name:    "box with default width",
			title:   "Default",
			content: "Content",
			width:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DrawBox(tt.title, tt.content, tt.width)

			// Verify box structure
			if !strings.Contains(result, BoxTopLeft) {
				t.Error("DrawBox() should contain BoxTopLeft")
			}
			if !strings.Contains(result, BoxTopRight) {
				t.Error("DrawBox() should contain BoxTopRight")
			}
			if !strings.Contains(result, BoxBottomLeft) {
				t.Error("DrawBox() should contain BoxBottomLeft")
			}
			if !strings.Contains(result, BoxBottomRight) {
				t.Error("DrawBox() should contain BoxBottomRight")
			}

			// Verify title if present
			if tt.title != "" && !strings.Contains(result, tt.title) {
				t.Errorf("DrawBox() should contain title %q\nGot: %s", tt.title, result)
			}

			// Verify content if present
			if tt.content != "" {
				lines := strings.Split(tt.content, "\n")
				for _, line := range lines {
					if line != "" && !strings.Contains(result, line) {
						t.Errorf("DrawBox() should contain content line %q\nGot: %s", line, result)
					}
				}
			}

			// Verify each line has correct width (count runes, not bytes)
			expectedWidth := tt.width
			if expectedWidth <= 0 {
				expectedWidth = DefaultBoxWidth
			}
			resultLines := strings.Split(result, "\n")
			for i, line := range resultLines {
				if line != "" {
					actualWidth := len([]rune(line))
					if actualWidth != expectedWidth {
						t.Errorf("DrawBox() line %d width = %d, want %d\nLine: %s", i, actualWidth, expectedWidth, line)
					}
				}
			}
		})
	}
}

func TestDrawMultilineBox(t *testing.T) {
	tests := []struct {
		name  string
		title string
		lines []string
		width int
	}{
		{
			name:  "multiline box with multiple lines",
			title: "List",
			lines: []string{"Item 1", "Item 2", "Item 3"},
			width: 40,
		},
		{
			name:  "multiline box with empty lines",
			title: "Mixed",
			lines: []string{"First", "", "Third"},
			width: 40,
		},
		{
			name:  "multiline box with no lines",
			title: "Empty",
			lines: []string{},
			width: 40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DrawMultilineBox(tt.title, tt.lines, tt.width)

			// Verify box structure
			if !strings.Contains(result, BoxTopLeft) {
				t.Error("DrawMultilineBox() should contain BoxTopLeft")
			}
			if !strings.Contains(result, BoxBottomRight) {
				t.Error("DrawMultilineBox() should contain BoxBottomRight")
			}

			// Verify all non-empty lines are present
			for _, line := range tt.lines {
				if line != "" && !strings.Contains(result, line) {
					t.Errorf("DrawMultilineBox() should contain line %q\nGot: %s", line, result)
				}
			}
		})
	}
}

func TestDrawBoxWithSections(t *testing.T) {
	tests := []struct {
		name     string
		boxTitle string
		sections []Section
		width    int
	}{
		{
			name:     "box with multiple sections",
			boxTitle: "Configuration",
			sections: []Section{
				{Title: "Section 1", Content: "Content 1"},
				{Title: "Section 2", Content: "Content 2"},
			},
			width: 50,
		},
		{
			name:     "box with sections without titles",
			boxTitle: "Data",
			sections: []Section{
				{Title: "", Content: "Line 1\nLine 2"},
				{Title: "", Content: "Line 3"},
			},
			width: 40,
		},
		{
			name:     "box with single section",
			boxTitle: "Single",
			sections: []Section{
				{Title: "Only Section", Content: "Only Content"},
			},
			width: 40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DrawBoxWithSections(tt.boxTitle, tt.sections, tt.width)

			// Verify box structure
			if !strings.Contains(result, BoxTopLeft) {
				t.Error("DrawBoxWithSections() should contain BoxTopLeft")
			}
			if !strings.Contains(result, BoxBottomRight) {
				t.Error("DrawBoxWithSections() should contain BoxBottomRight")
			}

			// Verify box title
			if !strings.Contains(result, tt.boxTitle) {
				t.Errorf("DrawBoxWithSections() should contain box title %q", tt.boxTitle)
			}

			// Verify sections
			for _, section := range tt.sections {
				if section.Title != "" && !strings.Contains(result, section.Title) {
					t.Errorf("DrawBoxWithSections() should contain section title %q", section.Title)
				}
				if section.Content != "" {
					lines := strings.Split(section.Content, "\n")
					for _, line := range lines {
						if line != "" && !strings.Contains(result, line) {
							t.Errorf("DrawBoxWithSections() should contain content line %q", line)
						}
					}
				}
			}

			// Verify separators between sections (but not after last)
			if len(tt.sections) > 1 {
				separatorCount := strings.Count(result, BoxTeeRight)
				expectedSeparators := len(tt.sections) - 1
				if separatorCount != expectedSeparators {
					t.Errorf("DrawBoxWithSections() should have %d separators, got %d", expectedSeparators, separatorCount)
				}
			}
		})
	}
}

func TestDefaultBoxWidth(t *testing.T) {
	if DefaultBoxWidth != 80 {
		t.Errorf("DefaultBoxWidth = %d, want 80", DefaultBoxWidth)
	}
}

func TestBoxCharacterConstants(t *testing.T) {
	// Verify that box-drawing constants are defined
	// (they should be from box_table.go)
	constants := map[string]string{
		"BoxTopLeft":     BoxTopLeft,
		"BoxTopRight":    BoxTopRight,
		"BoxBottomLeft":  BoxBottomLeft,
		"BoxBottomRight": BoxBottomRight,
		"BoxHorizontal":  BoxHorizontal,
		"BoxVertical":    BoxVertical,
		"BoxTeeRight":    BoxTeeRight,
		"BoxTeeLeft":     BoxTeeLeft,
	}

	for name, value := range constants {
		if value == "" {
			t.Errorf("%s should not be empty", name)
		}
	}
}
