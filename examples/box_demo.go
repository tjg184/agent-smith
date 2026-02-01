package main

import (
	"fmt"

	"github.com/tgaines/agent-smith/internal/formatter"
)

func main() {
	fmt.Println("=== Box-Drawing Utilities Demo ===\n")

	// Example 1: Simple box with title
	fmt.Println("1. Simple box with title:")
	box1 := formatter.DrawBox("Configuration", "Setting: value\nOption: enabled", 60)
	fmt.Println(box1)
	fmt.Println()

	// Example 2: Header only
	fmt.Println("2. Header:")
	header := formatter.DrawHeader("Status Report", 50)
	fmt.Println(header)
	fmt.Println()

	// Example 3: Separator
	fmt.Println("3. Separator:")
	separator := formatter.DrawSeparator(50)
	fmt.Println(separator)
	fmt.Println()

	// Example 4: Footer
	fmt.Println("4. Footer:")
	footer := formatter.DrawFooter(50)
	fmt.Println(footer)
	fmt.Println()

	// Example 5: Multiline box
	fmt.Println("5. Multiline box with multiple lines:")
	lines := []string{
		"✓ Successfully installed api-design",
		"✓ Successfully installed code-review",
		"✗ Failed to install debugging-strategies",
	}
	box2 := formatter.DrawMultilineBox("Installation Summary", lines, 70)
	fmt.Println(box2)
	fmt.Println()

	// Example 6: Box with sections
	fmt.Println("6. Box with sections:")
	sections := []formatter.Section{
		{Title: "Skills", Content: "• api-design\n• code-review"},
		{Title: "Agents", Content: "• backend-security-coder\n• sql-pro"},
		{Title: "Commands", Content: "• install\n• link"},
	}
	box3 := formatter.DrawBoxWithSections("Components", sections, 60)
	fmt.Println(box3)
	fmt.Println()

	// Example 7: Default width (80 characters)
	fmt.Println("7. Box with default width (80 characters):")
	box4 := formatter.DrawBox("Default Width Box", "This box uses the default width of 80 characters", 0)
	fmt.Println(box4)
}
