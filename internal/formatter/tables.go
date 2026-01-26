package formatter

import (
	"fmt"
	"strings"
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
	fmt.Fprintln(f.writer, "\n"+strings.Repeat("=", 80))
	fmt.Fprintln(f.writer, "Installation Summary")
	fmt.Fprintln(f.writer, strings.Repeat("=", 80))

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
	fmt.Fprintln(f.writer, "\n"+strings.Repeat("-", 80))
	fmt.Fprintf(f.writer, "Successfully installed: %d/%d components\n", successCount, len(results))
	if failureCount > 0 {
		fmt.Fprintf(f.writer, "Failed: %d components\n", failureCount)
	}
	fmt.Fprintln(f.writer, strings.Repeat("=", 80))
}

// displayTypeSection displays a section of results for a specific component type
func (f *Formatter) displayTypeSection(typeName string, results []InstallResult) {
	fmt.Fprintf(f.writer, "\n%s:\n", typeName)
	fmt.Fprintln(f.writer, strings.Repeat("-", 80))

	for _, result := range results {
		status := SymbolSuccess
		statusText := "Success"
		if !result.Success {
			status = SymbolError
			statusText = "Failed"
		}

		fmt.Fprintf(f.writer, "  %s  %-40s  %s\n", status, result.Name, statusText)
		if !result.Success && result.Error != "" {
			fmt.Fprintf(f.writer, "      Error: %s\n", result.Error)
		}
	}
}
