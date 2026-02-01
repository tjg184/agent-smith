package formatter

import (
	"fmt"

	"github.com/tgaines/agent-smith/pkg/colors"
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

	if failureCount > 0 {
		fmt.Fprintf(f.writer, "%s Successfully installed: %d/%d components | %s Failed: %d\n",
			colors.Success(SymbolSuccess), successCount, len(results), colors.Error(SymbolError), failureCount)
	} else {
		fmt.Fprintf(f.writer, "%s Successfully installed: %d/%d components\n",
			colors.Success(SymbolSuccess), successCount, len(results))
	}

	// Add "Next steps" section
	f.displayNextSteps()
}

// displayTypeSection displays a section of results for a specific component type
func (f *Formatter) displayTypeSection(typeName string, results []InstallResult) {
	fmt.Fprintf(f.writer, "\n%s:\n", typeName)

	// Create table with headers using box-drawing characters
	table := NewBoxTable(f.writer, []string{"Status", "Component", "Result"})

	for _, result := range results {
		var status string
		statusText := "Success"
		if result.Success {
			status = colors.Success(SymbolSuccess)
		} else {
			status = colors.Error(SymbolError)
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

	fmt.Fprintln(f.writer, "Next steps:")
	fmt.Fprintf(f.writer, "  • Link components: %s\n", colors.Info("agent-smith link all"))
	fmt.Fprintf(f.writer, "  • View status: %s\n", colors.Info("agent-smith status"))
	fmt.Fprintf(f.writer, "  • List components: %s\n", colors.Info("agent-smith list"))
}
