package formatter

import (
	"fmt"

	"github.com/fatih/color"
)

// InstallResult represents the result of a single component installation
type InstallResult struct {
	Name    string
	Type    string
	Success bool
	Error   string
}

// DisplaySummaryTable shows a formatted table of installation results
func (f *Formatter) DisplaySummaryTable(results []InstallResult, skillCount, agentCount, commandCount int) {
	f.EmptyLine()
	f.SectionHeader("Installation Summary")

	// Group results by type
	skillResults := []InstallResult{}
	agentResults := []InstallResult{}
	commandResults := []InstallResult{}

	for _, result := range results {
		switch result.Type {
		case "skill":
			skillResults = append(skillResults, result)
		case "agent":
			agentResults = append(agentResults, result)
		case "command":
			commandResults = append(commandResults, result)
		}
	}

	// Display each type section with box tables
	if len(skillResults) > 0 {
		f.displayTypeSection("Skills", skillResults)
	}
	if len(agentResults) > 0 {
		f.displayTypeSection("Agents", agentResults)
	}
	if len(commandResults) > 0 {
		f.displayTypeSection("Commands", commandResults)
	}

	// Calculate summary statistics
	successCount := 0
	failureCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			failureCount++
		}
	}

	// Display summary with colored symbols
	f.EmptyLine()
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	if failureCount > 0 {
		fmt.Fprintf(f.writer, "%s Successfully installed: %d/%d components | %s Failed: %d\n",
			green(SymbolSuccess), successCount, len(results), red(SymbolError), failureCount)
	} else {
		fmt.Fprintf(f.writer, "%s Successfully installed: %d/%d components\n",
			green(SymbolSuccess), successCount, len(results))
	}

	// Add "Next steps" section
	f.displayNextSteps()
}

// displayTypeSection displays a section of results for a specific component type
func (f *Formatter) displayTypeSection(typeName string, results []InstallResult) {
	fmt.Fprintf(f.writer, "\n%s:\n", typeName)

	// Create table with headers using box-drawing characters
	table := NewBoxTable(f.writer, []string{"Status", "Component", "Result"})

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	for _, result := range results {
		var status string
		statusText := "Success"
		if result.Success {
			status = green(SymbolSuccess)
		} else {
			status = red(SymbolError)
			statusText = "Failed"
		}

		table.AddRow([]string{status, result.Name, statusText})

		// Add error details as a separate row if needed
		if !result.Success && result.Error != "" {
			table.AddRow([]string{"", "└─ Error", result.Error})
		}
	}

	table.Render()
}

// displayNextSteps displays common follow-up commands after installation
func (f *Formatter) displayNextSteps() {
	f.EmptyLine()
	cyan := color.New(color.FgCyan).SprintFunc()

	fmt.Fprintln(f.writer, "Next steps:")
	fmt.Fprintf(f.writer, "  • Link components: %s\n", cyan("agent-smith link all"))
	fmt.Fprintf(f.writer, "  • View status: %s\n", cyan("agent-smith status"))
	fmt.Fprintf(f.writer, "  • List components: %s\n", cyan("agent-smith list"))
}
