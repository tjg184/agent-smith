package formatter

import (
	"fmt"
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
	fmt.Fprintln(f.writer, "\nInstallation Summary")

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

	// Display each type section
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

	// Display summary
	fmt.Fprintln(f.writer)
	summaryMsg := fmt.Sprintf("Successfully installed: %d/%d components", successCount, len(results))
	if failureCount > 0 {
		summaryMsg += fmt.Sprintf(" | Failed: %d components", failureCount)
	}
	fmt.Fprintln(f.writer, summaryMsg)
}

// displayTypeSection displays a section of results for a specific component type
func (f *Formatter) displayTypeSection(typeName string, results []InstallResult) {
	fmt.Fprintf(f.writer, "\n%s:\n", typeName)

	// Create table with headers
	table := NewBoxTable(f.writer, []string{"Status", "Component", "Result"})

	for _, result := range results {
		status := SymbolSuccess
		statusText := "Success"
		if !result.Success {
			status = SymbolError
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
