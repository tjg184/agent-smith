package help

import (
	"regexp"
	"strings"

	"github.com/tjg184/agent-smith/pkg/colors"
)

// ColorizeText parses and colorizes help text line by line
func ColorizeText(text string) string {
	if !colors.IsEnabled() {
		return text
	}

	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		result = append(result, colorizeLine(line))
	}

	return strings.Join(result, "\n")
}

// colorizeLine applies colorization to a single line based on patterns
func colorizeLine(line string) string {
	// Priority order matters!

	// 1. Section headers (highest priority - whole line)
	if isSectionHeader(line) {
		return colorizeSection(line)
	}

	// 2. Multi-pattern lines (comments + commands + parameters)
	if isComment(line) && isCommandExample(line) {
		return colorizeMultiPattern(line)
	}

	// 3. Individual patterns
	if isComment(line) {
		return colorizeComment(line)
	}

	if isCommandExample(line) {
		result := colorizeCommand(line)
		if hasURL(result) {
			result = colorizeURLs(result)
		}
		return result
	}

	if hasURL(line) {
		return colorizeURLs(line)
	}

	return line
}

// Pattern Detection Functions

// isSectionHeader detects section headers like "USAGE:", "EXAMPLES:", "FLAGS:"
// Must be at the start of the line (no leading whitespace) to be considered a header
func isSectionHeader(line string) bool {
	// Match lines that start with uppercase letters/spaces followed by a colon
	// Do NOT use TrimSpace - headers must be at column 0
	match, _ := regexp.MatchString(`^[A-Z][A-Z ]+:$`, line)
	return match
}

// isComment detects lines starting with #
func isComment(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "#")
}

// isCommandExample detects lines containing "agent-smith"
func isCommandExample(line string) bool {
	return strings.Contains(line, "agent-smith")
}

// hasURL detects URLs in the line
func hasURL(line string) bool {
	patterns := []string{
		`https?://`,
		`git@`,
		`[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+`, // GitHub shorthand
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			return true
		}
	}
	return false
}

// Colorization Functions

// colorizeSection colorizes section headers with cyan bold
func colorizeSection(line string) string {
	return colors.InfoBold(line)
}

// colorizeComment colorizes comment lines with gray/muted
func colorizeComment(line string) string {
	return colors.Muted(line)
}

// colorizeCommand colorizes command examples with green for commands and yellow for parameters
func colorizeCommand(line string) string {
	// Preserve indentation
	indent := getIndentation(line)
	trimmed := strings.TrimSpace(line)

	// Step 1: Colorize agent-smith commands (including subcommands)
	// Match "agent-smith" followed by one or more words
	re := regexp.MustCompile(`(agent-smith(?:\s+[a-z-]+)*)`)
	trimmed = re.ReplaceAllStringFunc(trimmed, func(match string) string {
		return colors.Success(match)
	})

	// Step 2: Colorize parameters <...>
	re = regexp.MustCompile(`<[^>]+>`)
	trimmed = re.ReplaceAllStringFunc(trimmed, func(match string) string {
		return colors.Warning(match)
	})

	// Step 3: Colorize [flags] and [options]
	re = regexp.MustCompile(`\[[^\]]+\]`)
	trimmed = re.ReplaceAllStringFunc(trimmed, func(match string) string {
		// Only colorize if it looks like a flag/option, not a description
		if strings.Contains(match, "flag") || strings.Contains(match, "option") || strings.Contains(match, "command") {
			return colors.Warning(match)
		}
		return match
	})

	return indent + trimmed
}

// colorizeURLs colorizes URLs with cyan
func colorizeURLs(line string) string {
	// Preserve indentation
	indent := getIndentation(line)
	trimmed := strings.TrimSpace(line)

	// Pattern 1: HTTP/HTTPS URLs
	re := regexp.MustCompile(`(https?://[^\s]+)`)
	trimmed = re.ReplaceAllStringFunc(trimmed, func(match string) string {
		return colors.Info(match)
	})

	// Pattern 2: Git SSH URLs
	re = regexp.MustCompile(`(git@[^\s]+)`)
	trimmed = re.ReplaceAllStringFunc(trimmed, func(match string) string {
		return colors.Info(match)
	})

	// Pattern 3: GitHub shorthand (owner/repo) - but be careful not to match file paths
	// Only match if preceded by whitespace or start of line and followed by whitespace or end
	re = regexp.MustCompile(`(^|\s)([a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+)(\s|$)`)
	trimmed = re.ReplaceAllStringFunc(trimmed, func(match string) string {
		// Extract the owner/repo part
		parts := strings.Fields(match)
		if len(parts) > 0 && strings.Contains(parts[0], "/") {
			colorized := colors.Info(parts[0])
			return strings.Replace(match, parts[0], colorized, 1)
		}
		return match
	})

	return indent + trimmed
}

// colorizeMultiPattern handles lines with multiple patterns (e.g., comments + commands)
func colorizeMultiPattern(line string) string {
	indent := getIndentation(line)
	trimmed := strings.TrimSpace(line)

	// Find where the comment starts
	commentIndex := strings.Index(trimmed, "#")
	if commentIndex == -1 {
		return colorizeCommand(line)
	}

	// Split into before and after comment
	beforeComment := trimmed[:commentIndex]
	afterComment := trimmed[commentIndex:]

	// Colorize command part (if any)
	if isCommandExample(beforeComment) {
		beforeComment = strings.TrimSpace(colorizeCommand("  " + beforeComment))
	}

	// Colorize comment part
	afterComment = colors.Muted(afterComment)

	result := beforeComment
	if beforeComment != "" {
		result += " "
	}
	result += afterComment

	return indent + result
}

// Helper Functions

// getIndentation returns the leading whitespace of a line
func getIndentation(line string) string {
	trimmed := strings.TrimLeft(line, " \t")
	return line[:len(line)-len(trimmed)]
}
