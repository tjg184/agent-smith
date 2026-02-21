// Package help provides custom Cobra templates with colored output
package help

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/tjg184/agent-smith/pkg/colors"
)

func init() {
	// Register custom template functions for colorization
	cobra.AddTemplateFunc("colorizeText", ColorizeText)
	cobra.AddTemplateFunc("colorizeSection", colorizeHeaderText)
	cobra.AddTemplateFunc("colorizeCommand", colorizeCommandText)
	cobra.AddTemplateFunc("colorizeFlagUsages", colorizeFlagUsages)
	cobra.AddTemplateFunc("colorizeHint", colorizeHintText)
}

// SetupCustomTemplates configures colored help templates for all commands
func SetupCustomTemplates(rootCmd *cobra.Command) {
	rootCmd.SetHelpTemplate(getHelpTemplate())
	rootCmd.SetUsageTemplate(getUsageTemplate())
}

// getHelpTemplate returns the custom help template with colorization
func getHelpTemplate() string {
	return `{{if .Long}}{{colorizeText .Long}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`
}

// getUsageTemplate returns the custom usage template with colored sections
func getUsageTemplate() string {
	return `{{colorizeSection "Usage:"}}{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

{{colorizeSection "Aliases:"}}
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

{{colorizeSection "Examples:"}}
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

{{colorizeSection "Available Commands:"}}{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{colorizeCommand .Name}} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{colorizeCommand .Name}} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{colorizeCommand .Name}} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{colorizeSection "Flags:"}}
{{colorizeFlagUsages .LocalFlags}}{{end}}{{if .HasAvailableInheritedFlags}}

{{colorizeSection "Global Flags:"}}
{{colorizeFlagUsages .InheritedFlags}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

{{colorizeHint (printf "Use \"%s [command] --help\" for more information about a command." .CommandPath)}}{{end}}`
}

// colorizeHeaderText colorizes section headers (like "Usage:", "Flags:")
func colorizeHeaderText(text string) string {
	if !colors.IsEnabled() {
		return text
	}
	return colors.InfoBold(text)
}

// colorizeCommandText colorizes command names in the command list
func colorizeCommandText(text string) string {
	if !colors.IsEnabled() {
		return text
	}
	return colors.Success(text)
}

// colorizeHintText colorizes hint text (muted)
func colorizeHintText(text string) string {
	if !colors.IsEnabled() {
		return text
	}
	return colors.Muted(text)
}

// colorizeFlagUsages colorizes flag usage strings
func colorizeFlagUsages(flagSet interface{}) string {
	// Get the flag usages string
	var usages string
	if fs, ok := flagSet.(interface{ FlagUsages() string }); ok {
		usages = fs.FlagUsages()
	} else {
		return ""
	}

	if !colors.IsEnabled() {
		return usages
	}

	// Colorize flag names (lines that start with spaces and have flags)
	lines := strings.Split(usages, "\n")
	var result []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			result = append(result, line)
			continue
		}

		// Colorize flag names (e.g., -h, --help)
		colorized := colorizeFlagLine(line)
		result = append(result, colorized)
	}

	return strings.Join(result, "\n")
}

// colorizeFlagLine colorizes a single flag line
func colorizeFlagLine(line string) string {
	// Preserve indentation
	indent := getIndentation(line)
	trimmed := strings.TrimSpace(line)

	// Find and colorize flag names (starts with - or --)
	// Format is typically: -f, --flag string   description
	parts := strings.Fields(trimmed)
	if len(parts) == 0 {
		return line
	}

	var colorizedParts []string
	for i, part := range parts {
		if strings.HasPrefix(part, "-") {
			// Colorize flag name
			colorizedParts = append(colorizedParts, colors.Info(part))
		} else if i < 3 && (part == "string" || part == "int" || part == "bool" || strings.HasPrefix(part, "[")) {
			// Colorize type annotation
			colorizedParts = append(colorizedParts, colors.Warning(part))
		} else {
			// Regular text
			colorizedParts = append(colorizedParts, part)
		}
	}

	return indent + strings.Join(colorizedParts, " ")
}
