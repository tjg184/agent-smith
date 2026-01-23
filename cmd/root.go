package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "agent-smith",
	Short: "Agent Smith - A CLI tool for managing AI agents, skills, and commands",
	Long: `Agent Smith is a powerful CLI tool for downloading, managing, and executing
AI agents, skills, and commands from git repositories.

It provides npm-like functionality for AI components, allowing you to:
- Download and install agents, skills, and commands
- Execute components without installation (npx-like)
- Update and manage installed components
- Link components to opencode configuration`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Add legacy commands as Cobra commands for better organization
	rootCmd.AddCommand(&cobra.Command{
		Use:   "add-skill <repository-url> <skill-name>",
		Short: "Download a skill from a git repository",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Call the existing main function logic
			handleAddSkill(args[0], args[1])
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "add-agent <repository-url> <agent-name>",
		Short: "Download an agent from a git repository",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			handleAddAgent(args[0], args[1])
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "add-command <repository-url> <command-name>",
		Short: "Download a command from a git repository",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			handleAddCommand(args[0], args[1])
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "add-all <repository-url>",
		Short: "Download all components from a git repository",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			handleAddAll(args[0])
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "npx <repository-or-package> [args...]",
		Short: "Execute a component without installing (npx-like)",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			handleRun(args[0], args[1:])
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "run <repository-or-package> [args...]",
		Short: "Execute a component without installing",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			handleRun(args[0], args[1:])
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "update <type> <name>",
		Short: "Check and update a specific component",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			handleUpdate(args[0], args[1])
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "update-all",
		Short: "Check and update all downloaded components",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			handleUpdateAll()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "link <type> <name>",
		Short: "Link a downloaded component to opencode",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			handleLink(args[0], args[1])
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "link-all",
		Short: "Link all downloaded components to opencode",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			handleLinkAll()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "auto-link",
		Short: "Automatically detect and link components from current repository",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			handleAutoLink()
		},
	})

	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
}

// These functions will be implemented in main.go to keep existing logic
var (
	handleAddSkill   func(repoURL, name string)
	handleAddAgent   func(repoURL, name string)
	handleAddCommand func(repoURL, name string)
	handleAddAll     func(repoURL string)
	handleRun        func(target string, args []string)
	handleUpdate     func(componentType, componentName string)
	handleUpdateAll  func()
	handleLink       func(componentType, componentName string)
	handleLinkAll    func()
	handleAutoLink   func()
)

func SetHandlers(
	addSkill func(repoURL, name string),
	addAgent func(repoURL, name string),
	addCommand func(repoURL, name string),
	addAll func(repoURL string),
	run func(target string, args []string),
	update func(componentType, componentName string),
	updateAll func(),
	link func(componentType, componentName string),
	linkAll func(),
	autoLink func(),
) {
	handleAddSkill = addSkill
	handleAddAgent = addAgent
	handleAddCommand = addCommand
	handleAddAll = addAll
	handleRun = run
	handleUpdate = update
	handleUpdateAll = updateAll
	handleLink = link
	handleLinkAll = linkAll
	handleAutoLink = autoLink
}
