package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/tjg184/agent-smith/pkg/help"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/paths"
	"github.com/tjg184/agent-smith/pkg/profiles"
	locksvc "github.com/tjg184/agent-smith/pkg/services/lock"
)

// Version is the current version of agent-smith.
// This is set by GoReleaser during build using ldflags.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "agent-smith",
	Short:   "Agent Smith - A CLI tool for managing AI agents, skills, and commands",
	Long:    getBanner(),
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		// Show enhanced welcome screen when no subcommand is provided
		showWelcomeScreen()
	},
}

// getBanner returns the ASCII art banner for Agent Smith
func getBanner() string {
	return `
  ___                   _     _____           _ _   _     
 / _ \                 | |   /  ___|         (_) | | |    
/ /_\ \ __ _  ___ _ __ | |_  \ ` + "`" + `--. _ __ ___  _| |_| |__  
|  _  |/ _` + "`" + ` |/ _ \ '_ \| __|  ` + "`" + `--. \ '_ ` + "`" + ` _ \| | __| '_ \ 
| | | | (_| |  __/ | | | |_  /\__/ / | | | | | | |_| | | |
\_| |_/\__, |\___|_| |_|\__| \____/|_| |_| |_|_|\__|_| |_|
        __/ |                                             
       |___/                                              
A CLI tool for managing AI agents, skills, and commands from git repositories.
`
}

// Execute runs the root command
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// showWelcomeScreen displays an enhanced welcome screen with status and quick actions
func showWelcomeScreen() {
	// Color definitions
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()
	highlight := color.New(color.FgHiWhite, color.Bold).SprintFunc()
	gray := color.New(color.FgHiBlack).SprintFunc()

	// Banner
	fmt.Print(getBanner())
	fmt.Println()

	// Status Section
	showSystemStatus(bold, cyan, gray)
	fmt.Println()

	// Quick Start Section
	fmt.Println(bold("QUICK START"))
	fmt.Printf("  %s %s\n", highlight("→"), cyan("agent-smith install all owner/repo"))
	fmt.Printf("    %s\n", yellow("Install components from a repository"))
	fmt.Println()
	fmt.Printf("  %s %s\n", highlight("→"), cyan("agent-smith link all"))
	fmt.Printf("    %s\n", yellow("Link everything to your AI editors"))
	fmt.Println()
	fmt.Printf("  %s %s\n", highlight("→"), cyan("agent-smith status"))
	fmt.Printf("    %s\n", yellow("Check your current setup"))
	fmt.Println()

	// Core Commands Section
	fmt.Println(bold("CORE COMMANDS"))
	fmt.Printf("  %s Install components from git repositories\n", green("install "))
	fmt.Printf("  %s Link components to AI editor targets\n", green("link    "))
	fmt.Printf("  %s Update installed components\n", green("update  "))
	fmt.Printf("  %s Manage profiles for context switching\n", green("profile "))
	fmt.Println()

	// Help hint
	fmt.Printf("Run %s for all commands or %s for details.\n",
		cyan("agent-smith --help"),
		cyan("agent-smith <command> --help"))
}

// showSystemStatus displays a brief system status in the welcome screen
func showSystemStatus(bold func(...interface{}) string, cyan func(...interface{}) string, gray func(...interface{}) string) {
	// Initialize profile manager to check status
	pm, _ := profiles.NewProfileManager(nil, locksvc.NewService(logger.New(logger.LevelError)))
	activeProfile, _ := pm.GetActiveProfile()

	// Count components
	skillsDir, _ := paths.GetSkillsDir()
	agentsDir, _ := paths.GetAgentsDir()
	commandsDir, _ := paths.GetCommandsDir()

	skillsCount := countComponents(skillsDir)
	agentsCount := countComponents(agentsDir)
	commandsCount := countComponents(commandsDir)

	fmt.Println(bold("SYSTEM STATUS"))

	// Profile status
	if activeProfile != "" {
		fmt.Printf("  Profile: %s\n", cyan(activeProfile))
	} else {
		fmt.Printf("  Profile: %s\n", gray("none (using base installation)"))
	}

	// Component counts
	total := skillsCount + agentsCount + commandsCount
	if total > 0 {
		parts := []string{}
		if skillsCount > 0 {
			parts = append(parts, fmt.Sprintf("%d skills", skillsCount))
		}
		if agentsCount > 0 {
			parts = append(parts, fmt.Sprintf("%d agents", agentsCount))
		}
		if commandsCount > 0 {
			parts = append(parts, fmt.Sprintf("%d commands", commandsCount))
		}
		fmt.Printf("  Components: %d installed (%s)\n", total, strings.Join(parts, ", "))
	} else {
		fmt.Printf("  Components: %s\n", gray("none installed yet"))
	}
}

// countComponents counts the number of components in a directory
func countComponents(dir string) int {
	if dir == "" {
		return 0
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			count++
		}
	}
	return count
}

func init() {
	// Setup colored help templates for all commands
	help.SetupCustomTemplates(rootCmd)

	// Hide completion command from help output
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")

	// Add persistent flags for all commands
	rootCmd.PersistentFlags().Bool("verbose", false, "Show informational output (default: show only errors)")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable verbose debug output for troubleshooting")
}
