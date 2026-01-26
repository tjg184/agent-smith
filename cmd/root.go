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

// isValidComponentType checks if a string is a valid component type
func isValidComponentType(componentType string) bool {
	return componentType == "skills" || componentType == "agents" || componentType == "commands"
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
		Use:   "update <type|all> [name]",
		Short: "Check and update a component or all components",
		Long: `Check and update a specific component or all downloaded components.

USAGE:
  agent-smith update <type> <name>  - Update a specific component
  agent-smith update all            - Update all downloaded components

EXAMPLES:
  # Update a specific skill
  agent-smith update skills mcp-builder

  # Update all components
  agent-smith update all`,
		Args: cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			if args[0] == "all" {
				handleUpdateAll()
			} else {
				if len(args) != 2 {
					cmd.PrintErrln("Error: update requires both type and name (or use 'update all')")
					os.Exit(1)
				}
				handleUpdate(args[0], args[1])
			}
		},
	})

	linkCmd := &cobra.Command{
		Use:   "link <type|all> [name]",
		Short: "Link a component or all components to detected targets",
		Long: `Link a specific component or all downloaded components to detected targets.

USAGE:
  agent-smith link <type> <name>  - Link a specific component
  agent-smith link <type>         - Link all components of a specific type
  agent-smith link all            - Link all downloaded components

FLAGS:
  --target, -t <target>  - Specify target to link to (opencode, claudecode, or all)
  --all-targets          - Explicitly link to all detected targets (default behavior)

EXAMPLES:
  # Link a specific skill to all detected targets (default)
  agent-smith link skills mcp-builder

  # Link all skills to OpenCode only
  agent-smith link skills --target opencode

  # Link all agents to Claude Code only
  agent-smith link agents --target claudecode

  # Link all commands to all targets explicitly
  agent-smith link commands --target all
  
  # Link all components to all detected targets
  agent-smith link all --all-targets`,
		Args: cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			allTargets, _ := cmd.Flags().GetBool("all-targets")

			// If --all-targets is specified, override targetFilter to "all"
			if allTargets {
				targetFilter = "all"
			}

			if args[0] == "all" {
				handleLinkAll(targetFilter)
			} else if len(args) == 1 && isValidComponentType(args[0]) {
				handleLinkType(args[0], targetFilter)
			} else if len(args) == 2 {
				handleLink(args[0], args[1], targetFilter)
			} else {
				cmd.PrintErrln("Error: link requires type and name, or just type, or 'all'")
				os.Exit(1)
			}
		},
	}
	linkCmd.Flags().StringP("target", "t", "", "Specify target to link to (opencode, claudecode, or all)")
	linkCmd.Flags().Bool("all-targets", false, "Link to all detected targets (default behavior)")
	rootCmd.AddCommand(linkCmd)

	rootCmd.AddCommand(&cobra.Command{
		Use:   "auto-link",
		Short: "Automatically detect and link components from current repository",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			handleAutoLink()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "list-links",
		Short: "List all components linked to opencode",
		Long: `List all components (skills, agents, and commands) currently linked to opencode.

This command shows the status of each linked component, including whether it's
a symlink or copied directory, and whether the link is valid or broken.

EXAMPLES:
  # List all linked components
  agent-smith list-links

The output shows:
  ✓ - Valid symlink
  ◆ - Copied directory
  ✗ - Broken link`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			handleListLinks()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "link-status",
		Short: "Show link status across all targets in a matrix view",
		Long: `Show the status of all components across all detected targets in a matrix format.

This command displays a table showing which components are linked to which targets,
making it easy to see what is installed where at a glance.

EXAMPLES:
  # Show link status matrix
  agent-smith link-status

The output shows:
  ✓ - Valid symlink
  ◆ - Copied directory
  ✗ - Broken link
  - - Not linked
  ? - Unknown status`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			handleLinkStatus()
		},
	})

	unlinkCmd := &cobra.Command{
		Use:   "unlink <type|all> [name]",
		Short: "Remove a linked component or all components from opencode",
		Long: `Remove a specific linked component or all linked components from opencode.

USAGE:
  agent-smith unlink <type> <name>  - Unlink a specific component
  agent-smith unlink <type>         - Unlink all components of a specific type
  agent-smith unlink all            - Unlink all components (with --force to skip confirmation)

COMPONENT TYPES:
  skills    - Remove linked skills
  agents    - Remove linked agents
  commands  - Remove linked commands

SAFETY:
  - Symlinks are removed immediately
  - Copied directories require confirmation before deletion
  - Source files in ~/.agents/ are never touched
  - 'unlink all' and 'unlink <type>' prompt for confirmation unless --force is used

EXAMPLES:
  # Unlink a specific skill
  agent-smith unlink skills mcp-builder

  # Unlink all skills
  agent-smith unlink skills

  # Unlink all agents with confirmation
  agent-smith unlink agents

  # Unlink all commands without confirmation
  agent-smith unlink commands --force

  # Unlink all components with confirmation
  agent-smith unlink all

  # Unlink all components without confirmation
  agent-smith unlink all --force`,
		Args: cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			force, _ := cmd.Flags().GetBool("force")
			if args[0] == "all" {
				handleUnlinkAll(force)
			} else if len(args) == 1 && isValidComponentType(args[0]) {
				handleUnlinkType(args[0], force)
			} else if len(args) == 2 {
				handleUnlink(args[0], args[1])
			} else {
				cmd.PrintErrln("Error: unlink requires type and name, or just type, or 'all'")
				os.Exit(1)
			}
		},
	}
	unlinkCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt (only for 'unlink all')")
	rootCmd.AddCommand(unlinkCmd)

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
	handleLink       func(componentType, componentName, targetFilter string)
	handleLinkAll    func(targetFilter string)
	handleLinkType   func(componentType, targetFilter string)
	handleAutoLink   func()
	handleListLinks  func()
	handleLinkStatus func()
	handleUnlink     func(componentType, componentName string)
	handleUnlinkAll  func(force bool)
	handleUnlinkType func(componentType string, force bool)
)

func SetHandlers(
	addSkill func(repoURL, name string),
	addAgent func(repoURL, name string),
	addCommand func(repoURL, name string),
	addAll func(repoURL string),
	run func(target string, args []string),
	update func(componentType, componentName string),
	updateAll func(),
	link func(componentType, componentName, targetFilter string),
	linkAll func(targetFilter string),
	linkType func(componentType, targetFilter string),
	autoLink func(),
	listLinks func(),
	linkStatus func(),
	unlink func(componentType, componentName string),
	unlinkAll func(force bool),
	unlinkType func(componentType string, force bool),
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
	handleLinkType = linkType
	handleAutoLink = autoLink
	handleListLinks = listLinks
	handleLinkStatus = linkStatus
	handleUnlink = unlink
	handleUnlinkAll = unlinkAll
	handleUnlinkType = unlinkType
}
