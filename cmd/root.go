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
		Long: `Download and install a skill from a git repository to your local agents directory.

This command fetches a skill from any git repository (GitHub, GitLab, Bitbucket, or private)
and installs it to ~/.agents/skills/<skill-name>. The skill can include multiple components
and will be automatically detected if it contains a SKILL.md file.

REPOSITORY URL FORMATS:
  GitHub shorthand:     owner/repo
  Full GitHub URL:      https://github.com/owner/repo
  GitLab URL:           https://gitlab.com/owner/repo
  SSH URL:              git@github.com:owner/repo.git
  Local path:           /path/to/local/repo

EXAMPLES:
  # Download from GitHub using shorthand
  agent-smith add-skill openai/cookbook gpt-skill

  # Download using full URL
  agent-smith add-skill https://github.com/example/repo my-skill

  # Download from local repository
  agent-smith add-skill /path/to/local/skill local-skill

The skill will be available for execution with 'agent-smith npx' or 'agent-smith run'.`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Call the existing main function logic
			handleAddSkill(args[0], args[1])
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "add-agent <repository-url> <agent-name>",
		Short: "Download an agent from a git repository",
		Long: `Download and install an AI agent from a git repository to your local agents directory.

This command fetches an agent from any git repository (GitHub, GitLab, Bitbucket, or private)
and installs it to ~/.agents/agents/<agent-name>. The agent will be automatically detected
based on path patterns and file extensions.

REPOSITORY URL FORMATS:
  GitHub shorthand:     owner/repo
  Full GitHub URL:      https://github.com/owner/repo
  GitLab URL:           https://gitlab.com/owner/repo
  SSH URL:              git@github.com:owner/repo.git
  Local path:           /path/to/local/repo

EXAMPLES:
  # Download from GitHub using shorthand
  agent-smith add-agent openai/assistant coding-agent

  # Download using full URL
  agent-smith add-agent https://github.com/example/agent my-agent

  # Download from local repository
  agent-smith add-agent /path/to/local/agent local-agent

The agent will be available for execution with 'agent-smith npx' or 'agent-smith run'.`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			handleAddAgent(args[0], args[1])
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "add-command <repository-url> <command-name>",
		Short: "Download a command from a git repository",
		Long: `Download and install a command-line tool from a git repository to your local agents directory.

This command fetches a command from any git repository (GitHub, GitLab, Bitbucket, or private)
and installs it to ~/.agents/commands/<command-name>. The command will be automatically detected
based on path patterns and file extensions.

REPOSITORY URL FORMATS:
  GitHub shorthand:     owner/repo
  Full GitHub URL:      https://github.com/owner/repo
  GitLab URL:           https://gitlab.com/owner/repo
  SSH URL:              git@github.com:owner/repo.git
  Local path:           /path/to/local/repo

EXAMPLES:
  # Download from GitHub using shorthand
  agent-smith add-command cli-tools/formatter json-formatter

  # Download using full URL
  agent-smith add-command https://github.com/example/tool my-tool

  # Download from local repository
  agent-smith add-command /path/to/local/command local-cmd

The command will be available for execution with 'agent-smith npx' or 'agent-smith run'.`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			handleAddCommand(args[0], args[1])
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "add-all <repository-url>",
		Short: "Download all components from a git repository",
		Long: `Download and install all components (skills, agents, and commands) from a git repository.

This command fetches a repository and automatically detects all AI components
within it, then downloads them to their respective directories. Components are
detected based on the presence of SKILL.md files or path patterns.

REPOSITORY URL FORMATS:
  GitHub shorthand:     owner/repo
  Full GitHub URL:      https://github.com/owner/repo
  GitLab URL:           https://gitlab.com/owner/repo
  SSH URL:              git@github.com:owner/repo.git
  Local path:           /path/to/local/repo

EXAMPLES:
  # Download all components from GitHub using shorthand
  agent-smith add-all openai/cookbook

  # Download using full URL
  agent-smith add-all https://github.com/example/monorepo

  # Download from local repository
  agent-smith add-all /path/to/local/repo

All detected components will be downloaded to ~/.agents/{skills,agents,commands}/`,
		Args: cobra.ExactArgs(1),
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
