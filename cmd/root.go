package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/tgaines/agent-smith/pkg/paths"
	"github.com/tgaines/agent-smith/pkg/profiles"
)

var rootCmd = &cobra.Command{
	Use:   "agent-smith",
	Short: "Agent Smith - A CLI tool for managing AI agents, skills, and commands",
	Long:  getBanner(),
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
	pm, _ := profiles.NewProfileManager(nil)
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

// isValidComponentType checks if a string is a valid component type
func isValidComponentType(componentType string) bool {
	return componentType == "skills" || componentType == "agents" || componentType == "commands"
}

// validateComponentType validates that a component type is valid and returns a helpful error if not
func validateComponentType(componentType string) error {
	if !isValidComponentType(componentType) {
		return fmt.Errorf("invalid component type '%s'\n\nValid component types:\n  - skills\n  - agents\n  - commands\n\nExample:\n  agent-smith update skills my-skill", componentType)
	}
	return nil
}

// exactArgsWithHelp returns a custom validator that provides helpful error messages
func exactArgsWithHelp(n int, usage string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < n {
			return fmt.Errorf("missing required arguments\n\nUsage: %s\n\nRun '%s --help' for more information", usage, cmd.CommandPath())
		}
		if len(args) > n {
			return fmt.Errorf("too many arguments provided\n\nUsage: %s\n\nRun '%s --help' for more information", usage, cmd.CommandPath())
		}
		return nil
	}
}

// noArgsWithHelp returns a custom validator for commands that accept no arguments
func noArgsWithHelp(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("this command does not accept arguments\n\nUsage: %s\n\nRun '%s --help' for more information", cmd.Use, cmd.CommandPath())
	}
	return nil
}

// rangeArgsWithHelp returns a custom validator for commands with a range of arguments
func rangeArgsWithHelp(min, max int, usage string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < min {
			return fmt.Errorf("missing required arguments\n\nUsage: %s\n\nRun '%s --help' for more information", usage, cmd.CommandPath())
		}
		if len(args) > max {
			return fmt.Errorf("too many arguments provided\n\nUsage: %s\n\nRun '%s --help' for more information", usage, cmd.CommandPath())
		}
		return nil
	}
}

// exactArgsWithComponentTypeValidation returns a validator that checks both argument count and component type
func exactArgsWithComponentTypeValidation(n int, componentTypeIndex int, usage string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		// First check argument count
		if len(args) < n {
			return fmt.Errorf("missing required arguments\n\nUsage: %s\n\nRun '%s --help' for more information", usage, cmd.CommandPath())
		}
		if len(args) > n {
			return fmt.Errorf("too many arguments provided\n\nUsage: %s\n\nRun '%s --help' for more information", usage, cmd.CommandPath())
		}

		// Then validate component type if specified
		if componentTypeIndex >= 0 && componentTypeIndex < len(args) {
			if err := validateComponentType(args[componentTypeIndex]); err != nil {
				return err
			}
		}

		return nil
	}
}

func init() {
	// Hide completion command from help output
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	// Create 'install' parent command with subcommands
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install components from git repositories",
		Long: `Install components (skills, agents, commands) from git repositories.

This command provides a modern interface for downloading and installing AI components
from any git repository (GitHub, GitLab, Bitbucket, or private repositories).

USAGE:
  You must specify a subcommand and provide the required parameters:
    agent-smith install skill <repository-url> <skill-name>
    agent-smith install agent <repository-url> <agent-name>
    agent-smith install command <repository-url> <command-name>
    agent-smith install all <repository-url>

  When installing a specific component (skill/agent/command), the name parameter is used
  to select which component to install from repositories containing multiple components.

REPOSITORY URL FORMATS:
  GitHub shorthand:     owner/repo
  Full GitHub URL:      https://github.com/owner/repo
  GitLab URL:           https://gitlab.com/owner/repo
  SSH URL:              git@github.com:owner/repo.git
  Local path:           /path/to/local/repo

EXAMPLES:
  # Install a specific skill from GitHub
  agent-smith install skill openai/cookbook gpt-skill

  # Install an agent from a full URL
  agent-smith install agent https://github.com/example/agent my-agent

  # Install all components from a repository
  agent-smith install all openai/cookbook`,
	}

	// Add subcommands to 'install' command
	installSkillCmd := &cobra.Command{
		Use:   "skill <repository-url> <skill-name>",
		Short: "Download a skill from a git repository",
		Long: `Download and install a skill from a git repository to your local agents directory.

This command fetches a skill from any git repository (GitHub, GitLab, Bitbucket, or private)
and installs it to ~/.agent-smith/skills/<skill-name>. The skill will be automatically detected
if it contains a SKILL.md file.

BEHAVIOR WITH MULTIPLE SKILLS:
When a repository contains multiple skills, the <skill-name> parameter is used to select
which skill to install. If the specified skill name is not found, the command will fail
and list all available skills in the repository.

REQUIRED PARAMETERS:
  <repository-url>  The URL or path to the git repository containing the skill
  <skill-name>      The name of the skill to install (also used as the local directory name)

EXAMPLES:
  # Download from GitHub using shorthand
  agent-smith install skill openai/cookbook gpt-skill

  # Download a specific skill from a repository with multiple skills
  agent-smith install skill example/skills-repo my-specific-skill

  # Download using full URL
  agent-smith install skill https://github.com/example/repo my-skill

  # Download from local repository
  agent-smith install skill /path/to/local/skill local-skill

  # Install directly to a profile
  agent-smith install skill openai/cookbook gpt-skill --profile work

  # Install to custom directory for testing (isolated from ~/.agent-smith/)
   agent-smith install skill ./my-skill test-skill --install-dir ./test-components`,
		Args: exactArgsWithHelp(2, "agent-smith install skill <repository-url> <skill-name>"),
		Run: func(cmd *cobra.Command, args []string) {
			profile, _ := cmd.Flags().GetString("profile")
			installDir, _ := cmd.Flags().GetString("install-dir")
			handleAddSkill(args[0], args[1], profile, installDir)
		},
	}
	installSkillCmd.Flags().StringP("profile", "p", "", "Install directly to a profile instead of ~/.agent-smith/")
	installSkillCmd.Flags().StringP("install-dir", "i", "", "Install to a custom directory (isolated from ~/.agent-smith/)")
	installCmd.AddCommand(installSkillCmd)

	installAgentCmd := &cobra.Command{
		Use:   "agent <repository-url> <agent-name>",
		Short: "Download an agent from a git repository",
		Long: `Download and install an AI agent from a git repository to your local agents directory.

This command fetches an agent from any git repository (GitHub, GitLab, Bitbucket, or private)
and installs it to ~/.agent-smith/agents/<agent-name>. The agent will be automatically detected
based on path patterns and file extensions.

REQUIRED PARAMETERS:
  <repository-url>  The URL or path to the git repository containing the agent
  <agent-name>      The name to use when installing the agent locally

EXAMPLES:
  # Download from GitHub using shorthand
  agent-smith install agent openai/assistant coding-agent

  # Download using full URL
  agent-smith install agent https://github.com/example/agent my-agent

  # Download from local repository
  agent-smith install agent /path/to/local/agent local-agent

  # Install directly to a profile
  agent-smith install agent openai/assistant coding-agent --profile work

  # Install to custom directory for testing (isolated from ~/.agent-smith/)
   agent-smith install agent ./my-agent test-agent --install-dir ./test-components`,
		Args: exactArgsWithHelp(2, "agent-smith install agent <repository-url> <agent-name>"),
		Run: func(cmd *cobra.Command, args []string) {
			profile, _ := cmd.Flags().GetString("profile")
			installDir, _ := cmd.Flags().GetString("install-dir")
			handleAddAgent(args[0], args[1], profile, installDir)
		},
	}
	installAgentCmd.Flags().StringP("profile", "p", "", "Install directly to a profile instead of ~/.agent-smith/")
	installAgentCmd.Flags().StringP("install-dir", "i", "", "Install to a custom directory (isolated from ~/.agent-smith/)")
	installCmd.AddCommand(installAgentCmd)

	installCommandCmd := &cobra.Command{
		Use:   "command <repository-url> <command-name>",
		Short: "Download a command from a git repository",
		Long: `Download and install a command-line tool from a git repository to your local agents directory.

This command fetches a command from any git repository (GitHub, GitLab, Bitbucket, or private)
and installs it to ~/.agent-smith/commands/<command-name>. The command will be automatically detected
based on path patterns and file extensions.

REQUIRED PARAMETERS:
  <repository-url>  The URL or path to the git repository containing the command
  <command-name>    The name to use when installing the command locally

EXAMPLES:
  # Download from GitHub using shorthand
  agent-smith install command cli-tools/formatter json-formatter

  # Download using full URL
  agent-smith install command https://github.com/example/tool my-tool

  # Download from local repository
  agent-smith install command /path/to/local/command local-cmd

  # Install directly to a profile
  agent-smith install command cli-tools/formatter json-formatter --profile work

  # Install to custom directory for testing (isolated from ~/.agent-smith/)
   agent-smith install command ./my-command test-command --install-dir ./test-components`,
		Args: exactArgsWithHelp(2, "agent-smith install command <repository-url> <command-name>"),
		Run: func(cmd *cobra.Command, args []string) {
			profile, _ := cmd.Flags().GetString("profile")
			installDir, _ := cmd.Flags().GetString("install-dir")
			handleAddCommand(args[0], args[1], profile, installDir)
		},
	}
	installCommandCmd.Flags().StringP("profile", "p", "", "Install directly to a profile instead of ~/.agent-smith/")
	installCommandCmd.Flags().StringP("install-dir", "i", "", "Install to a custom directory (isolated from ~/.agent-smith/)")
	installCmd.AddCommand(installCommandCmd)

	installAllCmd := &cobra.Command{
		Use:   "all <repository-url>",
		Short: "Download all components from a git repository",
		Long: `Download and install all components (skills, agents, and commands) from a git repository.

This command fetches a repository and automatically detects all AI components
within it, then downloads them to their respective directories. Components are
detected based on the presence of SKILL.md files or path patterns.

AUTOMATIC PROFILE CREATION:
By default, this command creates a repository-sourced profile (📦) to namespace
the components from the repository. The profile name is generated from the
repository URL (e.g., "owner-repo"). If a profile already exists for the same
repository, it will be reused and updated.

Repository-sourced profiles make it easy to:
  - Keep all components from a repo organized together
  - Update all components from the repo with 'update all'
  - Switch between different repositories

REQUIRED PARAMETERS:
  <repository-url>  The URL or path to the git repository containing components

EXAMPLES:
  # Download all components from GitHub using shorthand
  # Creates profile: openai-cookbook (📦)
  agent-smith install all openai/cookbook

  # Download using full URL
  agent-smith install all https://github.com/example/monorepo

  # Download from local repository
  agent-smith install all /path/to/local/repo

  # Install to a custom directory (project-local, no profile)
   agent-smith install all openai/cookbook --install-dir ./tools

   # Force creation of a new profile with a custom name
   agent-smith install all openai/cookbook --profile my-custom-profile`,
		Args: exactArgsWithHelp(1, "agent-smith install all <repository-url>"),
		Run: func(cmd *cobra.Command, args []string) {
			profile, _ := cmd.Flags().GetString("profile")
			installDir, _ := cmd.Flags().GetString("install-dir")
			handleAddAll(args[0], profile, installDir)
		},
	}
	installAllCmd.Flags().StringP("profile", "p", "", "Force creation of a new profile with a custom name")
	installAllCmd.Flags().StringP("install-dir", "i", "", "Install to a custom directory instead of ~/.agent-smith/")
	installCmd.AddCommand(installAllCmd)

	rootCmd.AddCommand(installCmd)

	updateCmd := &cobra.Command{
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
  agent-smith update all

  # Update components in a specific profile (bypasses active profile)
  agent-smith update all --profile work
  agent-smith update skills my-skill --profile work`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("missing required arguments\n\nUsage: agent-smith update <type|all> [name]\n\nRun '%s --help' for more information", cmd.CommandPath())
			}
			if len(args) > 2 {
				return fmt.Errorf("too many arguments provided\n\nUsage: agent-smith update <type|all> [name]\n\nRun '%s --help' for more information", cmd.CommandPath())
			}
			// If first arg is not "all", validate it's a valid component type and require name
			if args[0] != "all" {
				if err := validateComponentType(args[0]); err != nil {
					return err
				}
				if len(args) != 2 {
					return fmt.Errorf("missing component name\n\nUsage: agent-smith update <type> <name>\n\nExample:\n  agent-smith update skills my-skill\n\nRun '%s --help' for more information", cmd.CommandPath())
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			profile, _ := cmd.Flags().GetString("profile")
			if args[0] == "all" {
				handleUpdateAll(profile)
			} else {
				handleUpdate(args[0], args[1], profile)
			}
		},
	}
	updateCmd.Flags().StringP("profile", "p", "", "Update components in a specific profile instead of the active profile")
	rootCmd.AddCommand(updateCmd)

	// Create 'link' parent command with subcommands
	linkCmd := &cobra.Command{
		Use:   "link",
		Short: "Link components to AI editor targets",
		Long: `Link installed components (skills, agents, commands) to your AI editors.

QUICK START:
  agent-smith link all                    # Link everything to all editors
  agent-smith link skill my-skill         # Link one skill to all editors
  agent-smith link skills --to opencode   # Link all skills to OpenCode only

COMMAND GROUPS:
  Link specific components:
    link skill <name>     Link one skill
    link agent <name>     Link one agent
    link command <name>   Link one command

  Link all components by type:
    link skills           Link all skills
    link agents           Link all agents
    link commands         Link all commands

  Link everything:
    link all              Link all components (skills + agents + commands)

  Inspection commands:
    link status           Show matrix view: which components → which editors
    link list             Simple list of all linked components

FLAGS:
  --to, -t <target>     Target editor (opencode, claudecode, copilot, or all)
                        Default: all detected editors
  --profile <name>      Link FROM specific profile (bypasses active profile)

PROFILE AWARENESS:
Link commands automatically use components from your active profile.
Activate a profile first, then run link commands to apply it to your editors.

  Active profile:    Sources from ~/.agent-smith/profiles/<active-profile>/
  No profile:        Sources from ~/.agent-smith/ (base installation)

See 'agent-smith profile --help' for profile management.`,
	}

	// Add subcommands to 'link' command
	// Singular commands - operate on ONE component
	linkSkillCmd := &cobra.Command{
		Use:   "skill <name>",
		Short: "Link one skill to editors",
		Long: `Link a specific skill to AI editor targets.

EXAMPLES:
  # Link a skill to all editors (default)
  agent-smith link skill mcp-builder

  # Link a skill to OpenCode only
  agent-smith link skill mcp-builder --to opencode

  # Link a skill from a specific profile
  agent-smith link skill mcp-builder --profile work`,
		Args: exactArgsWithHelp(1, "agent-smith link skill <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("to")
			profile, _ := cmd.Flags().GetString("profile")

			// Link specific skill
			handleLink("skills", args[0], targetFilter, profile)
		},
	}
	linkSkillCmd.Flags().StringP("to", "t", "", "Target editor (opencode, claudecode, copilot, or all)")
	linkSkillCmd.Flags().String("profile", "", "Link from specific profile (bypasses active profile)")
	linkCmd.AddCommand(linkSkillCmd)

	// Plural command - operate on ALL skills
	linkSkillsCmd := &cobra.Command{
		Use:   "skills",
		Short: "Link all skills to editors",
		Long: `Link all skills to AI editor targets.

EXAMPLES:
  # Link all skills to all editors
  agent-smith link skills

  # Link all skills to Claude Code only
  agent-smith link skills --to claudecode

  # Link all skills from a specific profile
  agent-smith link skills --profile work`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("to")
			profile, _ := cmd.Flags().GetString("profile")

			// Link all skills
			handleLinkType("skills", targetFilter, profile)
		},
	}
	linkSkillsCmd.Flags().StringP("to", "t", "", "Target editor (opencode, claudecode, copilot, or all)")
	linkSkillsCmd.Flags().String("profile", "", "Link from specific profile (bypasses active profile)")
	linkCmd.AddCommand(linkSkillsCmd)

	linkAgentCmd := &cobra.Command{
		Use:   "agent <name>",
		Short: "Link one agent to editors",
		Long: `Link a specific agent to AI editor targets.

EXAMPLES:
  # Link an agent to all editors (default)
  agent-smith link agent coding-assistant

  # Link an agent to OpenCode only
  agent-smith link agent coding-assistant --to opencode

  # Link an agent from a specific profile
  agent-smith link agent coding-assistant --profile work`,
		Args: exactArgsWithHelp(1, "agent-smith link agent <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("to")
			profile, _ := cmd.Flags().GetString("profile")

			// Link specific agent
			handleLink("agents", args[0], targetFilter, profile)
		},
	}
	linkAgentCmd.Flags().StringP("to", "t", "", "Target editor (opencode, claudecode, copilot, or all)")
	linkAgentCmd.Flags().String("profile", "", "Link from specific profile (bypasses active profile)")
	linkCmd.AddCommand(linkAgentCmd)

	// Plural command - operate on ALL agents
	linkAgentsCmd := &cobra.Command{
		Use:   "agents",
		Short: "Link all agents to editors",
		Long: `Link all agents to AI editor targets.

EXAMPLES:
  # Link all agents to all editors
  agent-smith link agents

  # Link all agents to Claude Code only
  agent-smith link agents --to claudecode

  # Link all agents from a specific profile
  agent-smith link agents --profile work`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("to")
			profile, _ := cmd.Flags().GetString("profile")

			// Link all agents
			handleLinkType("agents", targetFilter, profile)
		},
	}
	linkAgentsCmd.Flags().StringP("to", "t", "", "Target editor (opencode, claudecode, copilot, or all)")
	linkAgentsCmd.Flags().String("profile", "", "Link from specific profile (bypasses active profile)")
	linkCmd.AddCommand(linkAgentsCmd)

	linkCommandCmd := &cobra.Command{
		Use:   "command <name>",
		Short: "Link one command to editors",
		Long: `Link a specific command to AI editor targets.

EXAMPLES:
  # Link a command to all editors (default)
  agent-smith link command json-formatter

  # Link a command to OpenCode only
  agent-smith link command json-formatter --to opencode

  # Link a command from a specific profile
  agent-smith link command json-formatter --profile work`,
		Args: exactArgsWithHelp(1, "agent-smith link command <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("to")
			profile, _ := cmd.Flags().GetString("profile")

			// Link specific command
			handleLink("commands", args[0], targetFilter, profile)
		},
	}
	linkCommandCmd.Flags().StringP("to", "t", "", "Target editor (opencode, claudecode, copilot, or all)")
	linkCommandCmd.Flags().String("profile", "", "Link from specific profile (bypasses active profile)")
	linkCmd.AddCommand(linkCommandCmd)

	// Plural command - operate on ALL commands
	linkCommandsCmd := &cobra.Command{
		Use:   "commands",
		Short: "Link all commands to editors",
		Long: `Link all commands to AI editor targets.

EXAMPLES:
  # Link all commands to all editors
  agent-smith link commands

  # Link all commands to Claude Code only
  agent-smith link commands --to claudecode

  # Link all commands from a specific profile
  agent-smith link commands --profile work`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("to")
			profile, _ := cmd.Flags().GetString("profile")

			// Link all commands
			handleLinkType("commands", targetFilter, profile)
		},
	}
	linkCommandsCmd.Flags().StringP("to", "t", "", "Target editor (opencode, claudecode, copilot, or all)")
	linkCommandsCmd.Flags().String("profile", "", "Link from specific profile (bypasses active profile)")
	linkCmd.AddCommand(linkCommandsCmd)

	linkAllCmd := &cobra.Command{
		Use:   "all",
		Short: "Link all components to editors",
		Long: `Link all components (skills, agents, commands) to AI editor targets.

This is the most common command - it links everything you've installed to your editors.

EXAMPLES:
  # Link all components to all editors (default)
  agent-smith link all

  # Link all components to OpenCode only
  agent-smith link all --to opencode

  # Link all components from a specific profile
  agent-smith link all --profile work

  # Link all components from all profiles
  agent-smith link all --all-profiles`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("to")
			profile, _ := cmd.Flags().GetString("profile")
			allProfiles, _ := cmd.Flags().GetBool("all-profiles")

			handleLinkAll(targetFilter, profile, allProfiles)
		},
	}
	linkAllCmd.Flags().StringP("to", "t", "", "Target editor (opencode, claudecode, copilot, or all)")
	linkAllCmd.Flags().String("profile", "", "Link from specific profile (bypasses active profile)")
	linkAllCmd.Flags().Bool("all-profiles", false, "Link components from all profiles simultaneously")
	linkCmd.AddCommand(linkAllCmd)

	linkAutoCmd := &cobra.Command{
		Use:   "auto",
		Short: "Auto-detect and link components in current directory",
		Long: `Auto-detect and link components from the current repository.

Scans the current directory for AI components and automatically links them to
your editors. Useful when developing components locally.

DETECTION PATTERNS:
  - Skills: Files named SKILL.md
  - Agents: Files in /agents/ directories
  - Commands: Files in /commands/ directories

EXAMPLES:
  # Auto-detect and link all components in current repo
  cd /path/to/my-components
  agent-smith link auto`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			handleAutoLink()
		},
	}
	linkCmd.AddCommand(linkAutoCmd)

	linkListCmd := &cobra.Command{
		Use:   "list",
		Short: "Simple list of linked components",
		Long: `List all components currently linked to AI editors.

Use 'link list' for a quick overview of what is linked. For a detailed matrix
view showing which components are linked to which editors, use 'link status'.

EXAMPLES:
  # Quick list of all linked components
  agent-smith link list`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			handleListLinks()
		},
	}
	linkCmd.AddCommand(linkListCmd)

	linkStatusCmd := &cobra.Command{
		Use:   "status",
		Short: "Matrix view: components vs editors",
		Long: `Show a detailed matrix of which components are linked to which editors.

This is more detailed than 'link list' - it shows a table with components as rows
and editors as columns, making it easy to see exactly what is linked where.

EXAMPLES:
  # Show status for current profile/base only
  agent-smith link status

  # Show status for all profiles
  agent-smith link status --all-profiles

  # Show status for specific profiles only
  agent-smith link status --all-profiles --profile=work,personal

LEGEND:
  ✓ - Valid symlink (linked and working)
  ◆ - Copied directory (linked but copied, not symlinked)
  ✗ - Broken link (link exists but target is missing)
  - - Not linked
  ? - Unknown status`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			allProfiles, _ := cmd.Flags().GetBool("all-profiles")
			profileFilter, _ := cmd.Flags().GetStringSlice("profile")
			handleLinkStatus(allProfiles, profileFilter)
		},
	}
	linkStatusCmd.Flags().Bool("all-profiles", false, "Show link status for all profiles")
	linkStatusCmd.Flags().StringSlice("profile", []string{}, "Filter to specific profiles (requires --all-profiles)")
	linkCmd.AddCommand(linkStatusCmd)

	rootCmd.AddCommand(linkCmd)

	// Create 'unlink' parent command with subcommands
	unlinkCmd := &cobra.Command{
		Use:   "unlink",
		Short: "Remove linked components from targets",
		Long: `Remove linked components (skills, agents, commands) from detected targets.

This command provides a modern interface for unlinking downloaded AI components
from supported targets (OpenCode, Claude Code, etc.).

SAFETY:
  - Symlinks are removed immediately
  - Copied directories require confirmation before deletion
  - Source files in ~/.agent-smith/ are never touched
  - Bulk operations (skills, agents, commands, all) prompt for confirmation unless --force is used`,
	}

	// Singular commands - operate on ONE component
	unlinkSkillCmd := &cobra.Command{
		Use:   "skill <name>",
		Short: "Unlink a specific skill from targets",
		Long: `Unlink a specific skill from detected targets.

This command removes the linked skill from OpenCode, Claude Code, or other
supported targets. Source files in ~/.agent-smith/skills/ are never touched.

EXAMPLES:
  # Unlink a specific skill from all targets
  agent-smith unlink skill mcp-builder

  # Unlink a specific skill from OpenCode only
  agent-smith unlink skill mcp-builder --target opencode

  # Unlink a skill from a specific profile
  agent-smith unlink skill mcp-builder --profile work`,
		Args: exactArgsWithHelp(1, "agent-smith unlink skill <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			profile, _ := cmd.Flags().GetString("profile")
			handleUnlinkWithProfile("skills", args[0], targetFilter, profile)
		},
	}
	unlinkSkillCmd.Flags().StringP("target", "t", "", "Target to unlink from (opencode, claudecode, copilot, or all). Default: unlink from all detected targets")
	unlinkSkillCmd.Flags().StringP("profile", "p", "", "Unlink from a specific profile without switching to it")
	unlinkCmd.AddCommand(unlinkSkillCmd)

	unlinkAgentCmd := &cobra.Command{
		Use:   "agent <name>",
		Short: "Unlink a specific agent from targets",
		Long: `Unlink a specific agent from detected targets.

This command removes the linked agent from OpenCode, Claude Code, or other
supported targets. Source files in ~/.agent-smith/agents/ are never touched.

EXAMPLES:
  # Unlink a specific agent from all targets
  agent-smith unlink agent coding-assistant

  # Unlink a specific agent from OpenCode only
  agent-smith unlink agent coding-assistant --target opencode

  # Unlink an agent from a specific profile
  agent-smith unlink agent coding-assistant --profile work`,
		Args: exactArgsWithHelp(1, "agent-smith unlink agent <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			profile, _ := cmd.Flags().GetString("profile")
			handleUnlinkWithProfile("agents", args[0], targetFilter, profile)
		},
	}
	unlinkAgentCmd.Flags().StringP("target", "t", "", "Target to unlink from (opencode, claudecode, copilot, or all). Default: unlink from all detected targets")
	unlinkAgentCmd.Flags().StringP("profile", "p", "", "Unlink from a specific profile without switching to it")
	unlinkCmd.AddCommand(unlinkAgentCmd)

	unlinkCommandCmd := &cobra.Command{
		Use:   "command <name>",
		Short: "Unlink a specific command from targets",
		Long: `Unlink a specific command from detected targets.

This command removes the linked command from OpenCode, Claude Code, or other
supported targets. Source files in ~/.agent-smith/commands/ are never touched.

EXAMPLES:
  # Unlink a specific command from all targets
  agent-smith unlink command json-formatter

  # Unlink a specific command from OpenCode only
  agent-smith unlink command json-formatter --target opencode

  # Unlink a command from a specific profile
  agent-smith unlink command json-formatter --profile work`,
		Args: exactArgsWithHelp(1, "agent-smith unlink command <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			profile, _ := cmd.Flags().GetString("profile")
			handleUnlinkWithProfile("commands", args[0], targetFilter, profile)
		},
	}
	unlinkCommandCmd.Flags().StringP("target", "t", "", "Target to unlink from (opencode, claudecode, copilot, or all). Default: unlink from all detected targets")
	unlinkCommandCmd.Flags().StringP("profile", "p", "", "Unlink from a specific profile without switching to it")
	unlinkCmd.AddCommand(unlinkCommandCmd)

	// Plural commands - operate on ALL components of a type
	unlinkSkillsCmd := &cobra.Command{
		Use:   "skills [name]",
		Short: "Unlink all skills from targets, or a specific skill if name provided",
		Long: `Unlink all skills from detected targets, or a specific skill if name is provided.

This command removes all linked skills from OpenCode, Claude Code, or other
supported targets. Source files in ~/.agent-smith/skills/ are never touched.

For backward compatibility, you can also provide a skill name to unlink just
that specific skill (equivalent to 'unlink skill <name>').

EXAMPLES:
  # Unlink all skills with confirmation
  agent-smith unlink skills

  # Unlink all skills without confirmation
  agent-smith unlink skills --force

  # Unlink all skills from OpenCode only
  agent-smith unlink skills --target opencode

  # Unlink all skills from a specific profile
  agent-smith unlink skills --profile work

  # Unlink a specific skill (backward compatibility)
  agent-smith unlink skills mcp-builder`,
		Args: rangeArgsWithHelp(0, 1, "agent-smith unlink skills [name]"),
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			profile, _ := cmd.Flags().GetString("profile")
			// Backward compatibility: if a name is provided, unlink that specific skill
			if len(args) == 1 {
				handleUnlinkWithProfile("skills", args[0], targetFilter, profile)
			} else {
				force, _ := cmd.Flags().GetBool("force")
				handleUnlinkTypeWithProfile("skills", targetFilter, force, profile)
			}
		},
	}
	unlinkSkillsCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	unlinkSkillsCmd.Flags().StringP("target", "t", "", "Target to unlink from (opencode, claudecode, copilot, or all). Default: unlink from all detected targets")
	unlinkSkillsCmd.Flags().StringP("profile", "p", "", "Unlink from a specific profile without switching to it")
	unlinkCmd.AddCommand(unlinkSkillsCmd)

	unlinkAgentsCmd := &cobra.Command{
		Use:   "agents [name]",
		Short: "Unlink all agents from targets, or a specific agent if name provided",
		Long: `Unlink all agents from detected targets, or a specific agent if name is provided.

This command removes all linked agents from OpenCode, Claude Code, or other
supported targets. Source files in ~/.agent-smith/agents/ are never touched.

For backward compatibility, you can also provide an agent name to unlink just
that specific agent (equivalent to 'unlink agent <name>').

EXAMPLES:
  # Unlink all agents with confirmation
  agent-smith unlink agents

  # Unlink all agents without confirmation
  agent-smith unlink agents --force

  # Unlink all agents from OpenCode only
  agent-smith unlink agents --target opencode

  # Unlink all agents from a specific profile
  agent-smith unlink agents --profile work

  # Unlink a specific agent (backward compatibility)
  agent-smith unlink agents coding-assistant`,
		Args: rangeArgsWithHelp(0, 1, "agent-smith unlink agents [name]"),
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			profile, _ := cmd.Flags().GetString("profile")
			// Backward compatibility: if a name is provided, unlink that specific agent
			if len(args) == 1 {
				handleUnlinkWithProfile("agents", args[0], targetFilter, profile)
			} else {
				force, _ := cmd.Flags().GetBool("force")
				handleUnlinkTypeWithProfile("agents", targetFilter, force, profile)
			}
		},
	}
	unlinkAgentsCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	unlinkAgentsCmd.Flags().StringP("target", "t", "", "Target to unlink from (opencode, claudecode, copilot, or all). Default: unlink from all detected targets")
	unlinkAgentsCmd.Flags().StringP("profile", "p", "", "Unlink from a specific profile without switching to it")
	unlinkCmd.AddCommand(unlinkAgentsCmd)

	unlinkCommandsCmd := &cobra.Command{
		Use:   "commands [name]",
		Short: "Unlink all commands from targets, or a specific command if name provided",
		Long: `Unlink all commands from detected targets, or a specific command if name is provided.

This command removes all linked commands from OpenCode, Claude Code, or other
supported targets. Source files in ~/.agent-smith/commands/ are never touched.

For backward compatibility, you can also provide a command name to unlink just
that specific command (equivalent to 'unlink command <name>').

EXAMPLES:
  # Unlink all commands with confirmation
  agent-smith unlink commands

  # Unlink all commands without confirmation
  agent-smith unlink commands --force

  # Unlink all commands from OpenCode only
  agent-smith unlink commands --target opencode

  # Unlink all commands from a specific profile
  agent-smith unlink commands --profile work

  # Unlink a specific command (backward compatibility)
  agent-smith unlink commands json-formatter`,
		Args: rangeArgsWithHelp(0, 1, "agent-smith unlink commands [name]"),
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			profile, _ := cmd.Flags().GetString("profile")
			// Backward compatibility: if a name is provided, unlink that specific command
			if len(args) == 1 {
				handleUnlinkWithProfile("commands", args[0], targetFilter, profile)
			} else {
				force, _ := cmd.Flags().GetBool("force")
				handleUnlinkTypeWithProfile("commands", targetFilter, force, profile)
			}
		},
	}
	unlinkCommandsCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	unlinkCommandsCmd.Flags().StringP("target", "t", "", "Target to unlink from (opencode, claudecode, copilot, or all). Default: unlink from all detected targets")
	unlinkCommandsCmd.Flags().StringP("profile", "p", "", "Unlink from a specific profile without switching to it")
	unlinkCmd.AddCommand(unlinkCommandsCmd)

	unlinkAllCmd := &cobra.Command{
		Use:   "all",
		Short: "Unlink all components from targets",
		Long: `Unlink all components (skills, agents, and commands) from detected targets.

By default, only components from the currently active profile are unlinked.
Use --all-profiles to unlink components from all profiles.
Use --profile to unlink from a specific profile without switching to it.

This command removes all linked components from OpenCode, Claude Code, or other
supported targets. Source files in ~/.agent-smith/ are never touched.

EXAMPLES:
  # Unlink all components from current profile with confirmation
  agent-smith unlink all

  # Unlink all components from current profile without confirmation
  agent-smith unlink all --force

  # Unlink all components from all profiles
  agent-smith unlink all --all-profiles

  # Unlink all components from a specific profile
  agent-smith unlink all --profile work

  # Unlink all components from OpenCode only
  agent-smith unlink all --target opencode`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("target")
			force, _ := cmd.Flags().GetBool("force")
			allProfiles, _ := cmd.Flags().GetBool("all-profiles")
			profile, _ := cmd.Flags().GetString("profile")
			handleUnlinkAllWithProfile(targetFilter, force, allProfiles, profile)
		},
	}
	unlinkAllCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	unlinkAllCmd.Flags().StringP("target", "t", "", "Target to unlink from (opencode, claudecode, copilot, or all). Default: unlink from all detected targets")
	unlinkAllCmd.Flags().Bool("all-profiles", false, "Unlink components from all profiles (default: current profile only)")
	unlinkAllCmd.Flags().StringP("profile", "p", "", "Unlink from a specific profile without switching to it")
	unlinkCmd.AddCommand(unlinkAllCmd)

	rootCmd.AddCommand(unlinkCmd)

	// Create 'uninstall' parent command with subcommands
	uninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove installed components from the system",
		Long: `Remove installed components (skills, agents, commands) from ~/.agent-smith/.

This command removes components from the system by:
  1. Automatically unlinking from all detected targets
  2. Removing the component directory from filesystem
  3. Removing the entry from lock files

SAFETY:
  - Components are automatically unlinked before removal
  - Source directories in ~/.agent-smith/ are permanently deleted
  - Lock file entries are removed to maintain consistency`,
	}

	// Individual component uninstall commands
	uninstallSkillCmd := &cobra.Command{
		Use:   "skill <name>",
		Short: "Remove a specific skill",
		Long: `Remove a specific skill from ~/.agent-smith/skills/.

This command removes the skill from the system by:
  1. Automatically unlinking from all detected targets
  2. Removing the skill directory from filesystem
  3. Removing the entry from .component-lock.json

EXAMPLES:
  # Remove a specific skill
  agent-smith uninstall skill mcp-builder

  # Remove from a profile
  agent-smith uninstall skill mcp-builder --profile work`,
		Args: exactArgsWithHelp(1, "agent-smith uninstall skill <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			profile, _ := cmd.Flags().GetString("profile")
			handleUninstall("skills", args[0], profile)
		},
	}
	uninstallSkillCmd.Flags().StringP("profile", "p", "", "Remove from a specific profile instead of ~/.agent-smith/")
	uninstallCmd.AddCommand(uninstallSkillCmd)

	uninstallAgentCmd := &cobra.Command{
		Use:   "agent <name>",
		Short: "Remove a specific agent",
		Long: `Remove a specific agent from ~/.agent-smith/agents/.

This command removes the agent from the system by:
  1. Automatically unlinking from all detected targets
  2. Removing the agent directory from filesystem
  3. Removing the entry from .component-lock.json

EXAMPLES:
  # Remove a specific agent
  agent-smith uninstall agent coding-assistant

  # Remove from a profile
  agent-smith uninstall agent coding-assistant --profile work`,
		Args: exactArgsWithHelp(1, "agent-smith uninstall agent <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			profile, _ := cmd.Flags().GetString("profile")
			handleUninstall("agents", args[0], profile)
		},
	}
	uninstallAgentCmd.Flags().StringP("profile", "p", "", "Remove from a specific profile instead of ~/.agent-smith/")
	uninstallCmd.AddCommand(uninstallAgentCmd)

	uninstallCommandCmd := &cobra.Command{
		Use:   "command <name>",
		Short: "Remove a specific command",
		Long: `Remove a specific command from ~/.agent-smith/commands/.

This command removes the command from the system by:
  1. Automatically unlinking from all detected targets
  2. Removing the command directory from filesystem
  3. Removing the entry from .component-lock.json

EXAMPLES:
  # Remove a specific command
  agent-smith uninstall command json-formatter

  # Remove from a profile
  agent-smith uninstall command json-formatter --profile work`,
		Args: exactArgsWithHelp(1, "agent-smith uninstall command <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			profile, _ := cmd.Flags().GetString("profile")
			handleUninstall("commands", args[0], profile)
		},
	}
	uninstallCommandCmd.Flags().StringP("profile", "p", "", "Remove from a specific profile instead of ~/.agent-smith/")
	uninstallCmd.AddCommand(uninstallCommandCmd)

	// Bulk uninstall from repository
	uninstallAllCmd := &cobra.Command{
		Use:   "all <repository-url>",
		Short: "Remove all components from a repository",
		Long: `Remove all components installed from a specific repository.

This command finds all components (skills, agents, commands) that were installed
from the specified repository and removes them from the system.

The repository URL can be in any of these formats:
  - GitHub shorthand: owner/repo
  - Full HTTPS URL: https://github.com/owner/repo
  - SSH URL: git@github.com:owner/repo.git

EXAMPLES:
  # Remove all components from a repository
  agent-smith uninstall all anthropics/skills

  # Remove without confirmation prompt
  agent-smith uninstall all https://github.com/anthropics/skills --force

SAFETY:
  - Shows a list of components before removal
  - Prompts for confirmation (unless --force flag is used)
  - Automatically unlinks components from all targets
  - Continues with remaining components if some fail`,
		Args: exactArgsWithHelp(1, "agent-smith uninstall all <repository-url>"),
		Run: func(cmd *cobra.Command, args []string) {
			force, _ := cmd.Flags().GetBool("force")
			handleUninstallAll(args[0], force)
		},
	}
	uninstallAllCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	uninstallCmd.AddCommand(uninstallAllCmd)

	rootCmd.AddCommand(uninstallCmd)

	// Create 'profiles' parent command with subcommands
	profilesCmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage profiles for context switching",
		Long: `Manage profiles to switch between different sets of agents, skills, and commands.
		
Profiles serve two purposes in agent-smith:

1. Repository Namespaces (📦):
   - Automatically created when you run 'install all <repo-url>'
   - Tied to the source repository for easy updates
   - Used to namespace components from a specific repo

2. User Collections (👤):
   - Manually created via 'profiles create'
   - Used for organizing and cherry-picking components across repos
   - Fully customizable for different projects or workflows

Both profile types can be activated, linked, and managed identically.
Use 'profile list --type repo' or '--type user' to filter by type.`,
	}

	profilesListCmd := &cobra.Command{
		Use:   "list",
		Short: "List available profiles",
		Long: `List all available profiles found in ~/.agent-smith/profiles/.

This command shows all valid profiles (those containing at least one component
directory), indicates which profile is currently active, and displays component
counts for each profile.

Profiles are marked with visual indicators:
  📦 - Repository-sourced profiles (created from install all)
  👤 - User-created profiles (created via profiles create)

Filtering options:
  --profile: Filter to show only specific profiles (can be specified multiple times)
  --active-only: Show only the currently active profile
  --type: Filter by profile type (repo or user)`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			profileFilter, _ := cmd.Flags().GetStringSlice("profile")
			activeOnly, _ := cmd.Flags().GetBool("active-only")
			typeFilter, _ := cmd.Flags().GetString("type")
			handleProfilesList(profileFilter, activeOnly, typeFilter)
		},
	}

	profilesListCmd.Flags().StringSlice("profile", []string{}, "Filter to specific profiles")
	profilesListCmd.Flags().Bool("active-only", false, "Show only the active profile")
	profilesListCmd.Flags().String("type", "", "Filter by profile type (repo or user)")

	profilesShowCmd := &cobra.Command{
		Use:   "show <profile-name>",
		Short: "Show detailed information about a profile",
		Long: `Display detailed information about a specific profile.

This command shows:
  - Profile name and active status
  - Profile location on disk
  - List of all agents in the profile
  - List of all skills in the profile
  - List of all commands in the profile

Use this before activating a profile to see exactly what components it contains.

EXAMPLES:
  # Show details of a profile
  agent-smith profile show my-profile
  
  # View contents before activating
  agent-smith profile show work-profile
  agent-smith profile activate work-profile`,
		Args: exactArgsWithHelp(1, "agent-smith profile show <profile-name>"),
		Run: func(cmd *cobra.Command, args []string) {
			handleProfilesShow(args[0])
		},
	}

	profilesCreateCmd := &cobra.Command{
		Use:   "create <profile-name>",
		Short: "Create a new user-created profile (👤)",
		Long: `Create a new user-created profile with empty component directories.

This command creates a profile marked as type="user" (👤), intended for 
organizing and cherry-picking components across different repositories.

The profile directory structure is created at ~/.agent-smith/profiles/<profile-name>/
with the following subdirectories:
  - agents/
  - skills/
  - commands/

User-created profiles are ideal for:
  - Organizing components from multiple repositories
  - Creating custom collections for specific workflows
  - Building project-specific component sets

Note: Repository-sourced profiles (📦) are automatically created when you
run 'install all <repo-url>' and are tied to their source repository.

After creation, you can add components to the profile and activate it with:
  agent-smith profile activate <profile-name>`,
		Args: exactArgsWithHelp(1, "agent-smith profile create <profile-name>"),
		Run: func(cmd *cobra.Command, args []string) {
			handleProfilesCreate(args[0])
		},
	}

	profilesDeleteCmd := &cobra.Command{
		Use:   "delete <profile-name>",
		Short: "Delete a profile",
		Long: `Delete a profile and all its contents.

This command permanently removes a profile directory and all components within it.
The profile must be deactivated before it can be deleted.

WARNING: This operation cannot be undone. All components in the profile will be permanently deleted.

EXAMPLES:
  # Delete a profile
  agent-smith profile delete my-profile

  # If the profile is active, deactivate it first
  agent-smith profile deactivate
  agent-smith profile delete my-profile`,
		Args: exactArgsWithHelp(1, "agent-smith profile delete <profile-name>"),
		Run: func(cmd *cobra.Command, args []string) {
			handleProfilesDelete(args[0])
		},
	}

	profilesActivateCmd := &cobra.Command{
		Use:   "activate <profile-name>",
		Short: "Activate a specific profile",
		Long: `Activate a profile without immediately affecting your editor.

This command will:
1. Update the active profile state
2. Deactivate any currently active profile

This does NOT immediately modify your editor configuration. To apply this profile
to your editor, run:
  agent-smith link all

Only one profile can be active at a time. The active profile persists across sessions.`,
		Args: exactArgsWithHelp(1, "agent-smith profile activate <profile-name>"),
		Run: func(cmd *cobra.Command, args []string) {
			handleProfilesActivate(args[0])
		},
	}

	profilesDeactivateCmd := &cobra.Command{
		Use:   "deactivate",
		Short: "Deactivate the current profile",
		Long: `Deactivate the currently active profile without immediately affecting your editor.

This command will:
1. Clear the active profile state
2. Return to base state (no profile active)

This does NOT immediately modify your editor configuration. To apply this change
to your editor, run:
  agent-smith link all

This allows you to control when changes are applied to your editor.`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			handleProfilesDeactivate()
		},
	}

	profilesAddCmd := &cobra.Command{
		Use:   "add <type> <profile-name> <component-name>",
		Short: "Add an existing component to a profile",
		Long: `Add an existing component from ~/.agent-smith/ to a profile.

This command copies a component (skill, agent, or command) from your base
~/.agent-smith/ directory to a specific profile. The component must already exist
in ~/.agent-smith/ before it can be added to a profile.

COMPONENT TYPES:
  skills   - Copy a skill to the profile
  agents   - Copy an agent to the profile
  commands - Copy a command to the profile

EXAMPLES:
  # Add a skill to a profile
  agent-smith profile add skills my-profile gpt-skill

  # Add an agent to a profile
  agent-smith profile add agents work-profile coding-agent

  # Add a command to a profile
  agent-smith profile add commands dev-profile test-runner`,
		Args: exactArgsWithComponentTypeValidation(3, 0, "agent-smith profile add <type> <profile-name> <component-name>"),
		Run: func(cmd *cobra.Command, args []string) {
			handleProfilesAdd(args[0], args[1], args[2])
		},
	}

	// profiles copy - Copy a component from one profile to another
	profilesCopyCmd := &cobra.Command{
		Use:   "copy <type> <source-profile> <target-profile> <component-name>",
		Short: "Copy a component from one profile to another",
		Long: `Copy a component (skill, agent, or command) from one profile to another.

This command copies both the component files and the lock file entry, allowing
the component to be updated independently in both profiles. This is useful for
creating specialized profiles or testing new versions.

COMPONENT TYPES:
  skills   - Copy a skill between profiles
  agents   - Copy an agent between profiles
  commands - Copy a command between profiles

EXAMPLES:
  # Copy a skill from work profile to personal profile
  agent-smith profile copy skills work-profile personal-profile api-design

  # Copy an agent from team profile to solo profile
  agent-smith profile copy agents team-profile solo-profile code-reviewer

  # Copy a command from dev profile to prod profile
  agent-smith profile copy commands dev-profile prod-profile test-runner`,
		Args: exactArgsWithComponentTypeValidation(4, 0, "agent-smith profile copy <type> <source-profile> <target-profile> <component-name>"),
		Run: func(cmd *cobra.Command, args []string) {
			handleProfilesCopy(args[0], args[1], args[2], args[3])
		},
	}

	// profiles remove - Remove a component from a profile
	profilesRemoveCmd := &cobra.Command{
		Use:   "remove [component-type] [profile] [name]",
		Short: "Remove a component from a profile",
		Long: `Remove a component from a profile by deleting it from the profile directory.

COMPONENT TYPES:
  skills   - Remove a skill from the profile
  agents   - Remove an agent from the profile
  commands - Remove a command from the profile

EXAMPLES:
  # Remove a skill from a profile
  agent-smith profile remove skills my-profile gpt-skill

  # Remove an agent from a profile
  agent-smith profile remove agents work-profile coding-agent

  # Remove a command from a profile
  agent-smith profile remove commands dev-profile test-runner`,
		Args: exactArgsWithComponentTypeValidation(3, 0, "agent-smith profile remove <type> <profile-name> <component-name>"),
		Run: func(cmd *cobra.Command, args []string) {
			handleProfilesRemove(args[0], args[1], args[2])
		},
	}

	// profiles cherry-pick - Create a profile by selecting components from existing profiles
	var cherryPickSources []string
	profilesCherryPickCmd := &cobra.Command{
		Use:   "cherry-pick <target-profile>",
		Short: "Create or enhance a profile by selecting components from existing profiles",
		Long: `Create or enhance a profile by cherry-picking components from existing profiles.

This command provides an interactive interface to select specific components
(agents, skills, commands) from one or more source profiles and copy them to
a target profile. This is useful for creating specialized toolsets or project-specific
configurations.

By default, components from all profiles are shown. Use --source flags to limit
the selection to specific source profiles.

FLAGS:
  --source <profile>  Limit selection to components from this profile (repeatable)

EXAMPLES:
  # Cherry-pick from all profiles
  agent-smith profile cherry-pick my-new-profile

  # Cherry-pick only from work-profile
  agent-smith profile cherry-pick project-x --source work-profile

  # Cherry-pick from multiple specific profiles
  agent-smith profile cherry-pick custom --source work --source personal`,
		Args: exactArgsWithHelp(1, "agent-smith profile cherry-pick <target-profile>"),
		Run: func(cmd *cobra.Command, args []string) {
			handleProfilesCherryPick(args[0], cherryPickSources)
		},
	}
	profilesCherryPickCmd.Flags().StringSliceVarP(&cherryPickSources, "source", "s", []string{}, "Source profile(s) to cherry-pick from (repeatable)")

	profilesCmd.AddCommand(profilesListCmd)
	profilesCmd.AddCommand(profilesShowCmd)
	profilesCmd.AddCommand(profilesCreateCmd)
	profilesCmd.AddCommand(profilesDeleteCmd)
	profilesCmd.AddCommand(profilesActivateCmd)
	profilesCmd.AddCommand(profilesDeactivateCmd)
	profilesCmd.AddCommand(profilesAddCmd)
	profilesCmd.AddCommand(profilesCopyCmd)
	profilesCmd.AddCommand(profilesRemoveCmd)
	profilesCmd.AddCommand(profilesCherryPickCmd)
	rootCmd.AddCommand(profilesCmd)

	// Add status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show current status and active profile",
		Long: `Display the current configuration status including:
  - Active profile (if any)
  - Detected targets (OpenCode, Claude Code, etc.)
  - Component counts in ~/.agent-smith/
  - Quick summary of system state

This provides a dashboard view of your agent-smith installation.`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			handleStatus()
		},
	}
	rootCmd.AddCommand(statusCmd)

	// Create 'target' parent command with subcommands
	targetCmd := &cobra.Command{
		Use:   "target",
		Short: "Manage custom targets",
		Long: `Manage custom targets for linking components.

Custom targets allow you to link components to additional editors or tools
beyond the built-in OpenCode and Claude Code targets.`,
	}

	targetAddCmd := &cobra.Command{
		Use:   "add <name> <path>",
		Short: "Register a new custom target",
		Long: `Register a new custom target for linking components.

This command adds a custom target to your configuration file, allowing you
to link components to additional editors or tools beyond the built-in targets.

The target will use the following subdirectories (relative to the path):
  - skills/   - For skills
  - agents/   - For agents
  - commands/ - For commands

EXAMPLES:
  # Add a custom target for Cursor
  agent-smith target add cursor ~/.cursor

  # Add a custom target for VS Code
  agent-smith target add vscode ~/.vscode/agent-smith

After adding a target, you can link components to it using:
  agent-smith link all --target <name>`,
		Args: exactArgsWithHelp(2, "agent-smith target add <name> <path>"),
		Run: func(cmd *cobra.Command, args []string) {
			handleTargetAdd(args[0], args[1])
		},
	}
	targetCmd.AddCommand(targetAddCmd)

	targetRemoveCmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Unregister a custom target",
		Long: `Unregister a custom target from your configuration.

This command removes a custom target from your configuration file. Only custom
targets can be removed - built-in targets (opencode, claudecode) cannot be removed.

EXAMPLES:
  # Remove a custom target
  agent-smith target remove cursor

  # Remove a custom target for VS Code
  agent-smith target remove vscode

Note: This only removes the target from the configuration. It does not unlink
any components that are currently linked to this target.`,
		Args: exactArgsWithHelp(1, "agent-smith target remove <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			handleTargetRemove(args[0])
		},
	}
	targetCmd.AddCommand(targetRemoveCmd)

	targetListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all available targets",
		Long: `List all available targets for linking components.

This command displays all targets that are configured in your system, including:
  - Built-in targets (OpenCode, Claude Code)
  - Custom targets from your configuration

For each target, it shows:
  - Target name
  - Base directory path
  - Whether the directory currently exists

EXAMPLES:
  # List all available targets
  agent-smith target list`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			handleTargetList()
		},
	}
	targetCmd.AddCommand(targetListCmd)

	rootCmd.AddCommand(targetCmd)

	// Create 'materialize' parent command with subcommands
	materializeCmd := &cobra.Command{
		Use:   "materialize",
		Short: "Materialize components to project directories",
		Long: `Materialize components (skills, agents, commands) to project directories for version control.

This command copies components from ~/.agent-smith/ to project-local directories
(.opencode/ or .claude/) so they can be committed to version control and shared
with your team.

USAGE:
  agent-smith materialize skill <name> --target <opencode|claudecode|copilot|universal|all>
  agent-smith materialize agent <name> --target <opencode|claudecode|copilot|universal|all>
  agent-smith materialize command <name> --target <opencode|claudecode|copilot|universal|all>

FLAGS:
  --target, -t <target>  - Target to materialize to (opencode, claudecode, copilot, universal, or all)
                           universal = target-agnostic storage in .agents/ directory
                           Can also be set via AGENT_SMITH_TARGET environment variable
  --project-dir <path>   - Optional, override project directory detection
  --force, -f            - Overwrite existing component if it differs
  --dry-run              - Preview what will be materialized without making changes

EXAMPLES:
  # Materialize a skill to OpenCode
  agent-smith materialize skill my-skill --target opencode

  # Materialize to both targets
  agent-smith materialize skill my-skill --target all

  # Materialize using environment variable
  export AGENT_SMITH_TARGET=opencode
  agent-smith materialize skill my-skill

  # Materialize from specific directory
  agent-smith materialize skill my-skill --target opencode --project-dir ./my-project
  
  # Force overwrite existing component
  agent-smith materialize skill my-skill --target opencode --force

  # Preview materialization without making changes
  agent-smith materialize skill my-skill --target opencode --dry-run`,
	}

	materializeSkillCmd := &cobra.Command{
		Use:   "skill <name>",
		Short: "Materialize a skill to project directories",
		Long: `Materialize a skill from ~/.agent-smith/skills/ to project directories.

This command copies the entire skill directory to .opencode/skills/ or .claude/skills/
with full provenance tracking in .component-lock.json.

The target can be specified with --target flag or AGENT_SMITH_TARGET environment variable.

EXAMPLES:
  # Materialize a skill to OpenCode
  agent-smith materialize skill my-skill --target opencode

  # Materialize to both targets
  agent-smith materialize skill my-skill --target all

  # Use environment variable for default target
  export AGENT_SMITH_TARGET=opencode
  agent-smith materialize skill my-skill

  # Preview without making changes
  agent-smith materialize skill my-skill --target opencode --dry-run`,
		Args: exactArgsWithHelp(1, "agent-smith materialize skill <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			target, _ := cmd.Flags().GetString("target")
			projectDir, _ := cmd.Flags().GetString("project-dir")
			force, _ := cmd.Flags().GetBool("force")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			profile, _ := cmd.Flags().GetString("profile")
			source, _ := cmd.Flags().GetString("source")
			handleMaterializeComponent("skills", args[0], target, projectDir, force, dryRun, profile, source)
		},
	}
	materializeSkillCmd.Flags().StringP("target", "t", "", "Target to materialize to (opencode, claudecode, copilot, universal, or all). Can also use AGENT_SMITH_TARGET environment variable")
	materializeSkillCmd.Flags().String("project-dir", "", "Override project directory detection")
	materializeSkillCmd.Flags().BoolP("force", "f", false, "Overwrite existing component if it differs")
	materializeSkillCmd.Flags().Bool("dry-run", false, "Preview what will be materialized without making changes")
	materializeSkillCmd.Flags().StringP("profile", "p", "", "Materialize from specific profile (use 'base' for ~/.agent-smith/)")
	materializeSkillCmd.Flags().StringP("source", "s", "", "Source URL to disambiguate when component exists in multiple sources")
	materializeCmd.AddCommand(materializeSkillCmd)

	materializeAgentCmd := &cobra.Command{
		Use:   "agent <name>",
		Short: "Materialize an agent to project directories",
		Long: `Materialize an agent from ~/.agent-smith/agents/ to project directories.

This command copies the entire agent directory to .opencode/agents/ or .claude/agents/
with full provenance tracking in .component-lock.json.

The target can be specified with --target flag or AGENT_SMITH_TARGET environment variable.

EXAMPLES:
  # Materialize an agent to OpenCode
  agent-smith materialize agent my-agent --target opencode

  # Materialize to both targets
  agent-smith materialize agent my-agent --target all

  # Use environment variable for default target
  export AGENT_SMITH_TARGET=claudecode
  agent-smith materialize agent my-agent

  # Preview without making changes
  agent-smith materialize agent my-agent --target opencode --dry-run`,
		Args: exactArgsWithHelp(1, "agent-smith materialize agent <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			target, _ := cmd.Flags().GetString("target")
			projectDir, _ := cmd.Flags().GetString("project-dir")
			force, _ := cmd.Flags().GetBool("force")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			profile, _ := cmd.Flags().GetString("profile")
			source, _ := cmd.Flags().GetString("source")
			handleMaterializeComponent("agents", args[0], target, projectDir, force, dryRun, profile, source)
		},
	}
	materializeAgentCmd.Flags().StringP("target", "t", "", "Target to materialize to (opencode, claudecode, copilot, universal, or all). Can also use AGENT_SMITH_TARGET environment variable")
	materializeAgentCmd.Flags().String("project-dir", "", "Override project directory detection")
	materializeAgentCmd.Flags().BoolP("force", "f", false, "Overwrite existing component if it differs")
	materializeAgentCmd.Flags().Bool("dry-run", false, "Preview what will be materialized without making changes")
	materializeAgentCmd.Flags().StringP("profile", "p", "", "Materialize from specific profile (use 'base' for ~/.agent-smith/)")
	materializeAgentCmd.Flags().StringP("source", "s", "", "Source URL to disambiguate when component exists in multiple sources")
	materializeCmd.AddCommand(materializeAgentCmd)

	materializeCommandCmd := &cobra.Command{
		Use:   "command <name>",
		Short: "Materialize a command to project directories",
		Long: `Materialize a command from ~/.agent-smith/commands/ to project directories.

This command copies the entire command directory to .opencode/commands/ or .claude/commands/
with full provenance tracking in .component-lock.json.

The target can be specified with --target flag or AGENT_SMITH_TARGET environment variable.

EXAMPLES:
  # Materialize a command to OpenCode
  agent-smith materialize command my-command --target opencode

  # Materialize to both targets
  agent-smith materialize command my-command --target all

  # Use environment variable for default target
  export AGENT_SMITH_TARGET=all
  agent-smith materialize command my-command

  # Preview without making changes
  agent-smith materialize command my-command --target opencode --dry-run`,
		Args: exactArgsWithHelp(1, "agent-smith materialize command <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			target, _ := cmd.Flags().GetString("target")
			projectDir, _ := cmd.Flags().GetString("project-dir")
			force, _ := cmd.Flags().GetBool("force")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			profile, _ := cmd.Flags().GetString("profile")
			source, _ := cmd.Flags().GetString("source")
			handleMaterializeComponent("commands", args[0], target, projectDir, force, dryRun, profile, source)
		},
	}
	materializeCommandCmd.Flags().StringP("target", "t", "", "Target to materialize to (opencode, claudecode, copilot, universal, or all). Can also use AGENT_SMITH_TARGET environment variable")
	materializeCommandCmd.Flags().String("project-dir", "", "Override project directory detection")
	materializeCommandCmd.Flags().BoolP("force", "f", false, "Overwrite existing component if it differs")
	materializeCommandCmd.Flags().Bool("dry-run", false, "Preview what will be materialized without making changes")
	materializeCommandCmd.Flags().StringP("profile", "p", "", "Materialize from specific profile (use 'base' for ~/.agent-smith/)")
	materializeCommandCmd.Flags().StringP("source", "s", "", "Source URL to disambiguate when component exists in multiple sources")
	materializeCmd.AddCommand(materializeCommandCmd)

	// Plural command - operate on ALL skills
	materializeSkillsCmd := &cobra.Command{
		Use:   "skills",
		Short: "Materialize all skills to project directories",
		Long: `Materialize all skills to project directories.

This command copies all skills from ~/.agent-smith/skills/ to .opencode/skills/
or .claude/skills/ with full provenance tracking. It continues on error with
individual components.

The target can be specified with --target flag or AGENT_SMITH_TARGET environment variable.

EXAMPLES:
  # Materialize all skills to OpenCode
  agent-smith materialize skills --target opencode

  # Materialize all skills to Claude Code
  agent-smith materialize skills --target claudecode

  # Materialize all skills from a specific profile
  agent-smith materialize skills --target opencode --profile work

  # Preview without making changes
  agent-smith materialize skills --target opencode --dry-run`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			target, _ := cmd.Flags().GetString("target")
			projectDir, _ := cmd.Flags().GetString("project-dir")
			force, _ := cmd.Flags().GetBool("force")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			profile, _ := cmd.Flags().GetString("profile")
			handleMaterializeType("skills", target, projectDir, force, dryRun, profile)
		},
	}
	materializeSkillsCmd.Flags().StringP("target", "t", "", "Target to materialize to (opencode, claudecode, copilot, universal, or all). Can also use AGENT_SMITH_TARGET environment variable")
	materializeSkillsCmd.Flags().String("project-dir", "", "Override project directory detection")
	materializeSkillsCmd.Flags().BoolP("force", "f", false, "Overwrite existing components if they differ")
	materializeSkillsCmd.Flags().Bool("dry-run", false, "Preview what will be materialized without making changes")
	materializeSkillsCmd.Flags().StringP("profile", "p", "", "Materialize from specific profile (use 'base' for ~/.agent-smith/)")
	materializeCmd.AddCommand(materializeSkillsCmd)

	// Plural command - operate on ALL agents
	materializeAgentsCmd := &cobra.Command{
		Use:   "agents",
		Short: "Materialize all agents to project directories",
		Long: `Materialize all agents to project directories.

This command copies all agents from ~/.agent-smith/agents/ to .opencode/agents/
or .claude/agents/ with full provenance tracking. It continues on error with
individual components.

The target can be specified with --target flag or AGENT_SMITH_TARGET environment variable.

EXAMPLES:
  # Materialize all agents to OpenCode
  agent-smith materialize agents --target opencode

  # Materialize all agents to Claude Code
  agent-smith materialize agents --target claudecode

  # Materialize all agents from a specific profile
  agent-smith materialize agents --target opencode --profile work

  # Preview without making changes
  agent-smith materialize agents --target opencode --dry-run`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			target, _ := cmd.Flags().GetString("target")
			projectDir, _ := cmd.Flags().GetString("project-dir")
			force, _ := cmd.Flags().GetBool("force")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			profile, _ := cmd.Flags().GetString("profile")
			handleMaterializeType("agents", target, projectDir, force, dryRun, profile)
		},
	}
	materializeAgentsCmd.Flags().StringP("target", "t", "", "Target to materialize to (opencode, claudecode, copilot, universal, or all). Can also use AGENT_SMITH_TARGET environment variable")
	materializeAgentsCmd.Flags().String("project-dir", "", "Override project directory detection")
	materializeAgentsCmd.Flags().BoolP("force", "f", false, "Overwrite existing components if they differ")
	materializeAgentsCmd.Flags().Bool("dry-run", false, "Preview what will be materialized without making changes")
	materializeAgentsCmd.Flags().StringP("profile", "p", "", "Materialize from specific profile (use 'base' for ~/.agent-smith/)")
	materializeCmd.AddCommand(materializeAgentsCmd)

	// Plural command - operate on ALL commands
	materializeCommandsCmd := &cobra.Command{
		Use:   "commands",
		Short: "Materialize all commands to project directories",
		Long: `Materialize all commands to project directories.

This command copies all commands from ~/.agent-smith/commands/ to .opencode/commands/
or .claude/commands/ with full provenance tracking. It continues on error with
individual components.

The target can be specified with --target flag or AGENT_SMITH_TARGET environment variable.

EXAMPLES:
  # Materialize all commands to OpenCode
  agent-smith materialize commands --target opencode

  # Materialize all commands to Claude Code
  agent-smith materialize commands --target claudecode

  # Materialize all commands from a specific profile
  agent-smith materialize commands --target opencode --profile work

  # Preview without making changes
  agent-smith materialize commands --target opencode --dry-run`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			target, _ := cmd.Flags().GetString("target")
			projectDir, _ := cmd.Flags().GetString("project-dir")
			force, _ := cmd.Flags().GetBool("force")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			profile, _ := cmd.Flags().GetString("profile")
			handleMaterializeType("commands", target, projectDir, force, dryRun, profile)
		},
	}
	materializeCommandsCmd.Flags().StringP("target", "t", "", "Target to materialize to (opencode, claudecode, copilot, universal, or all). Can also use AGENT_SMITH_TARGET environment variable")
	materializeCommandsCmd.Flags().String("project-dir", "", "Override project directory detection")
	materializeCommandsCmd.Flags().BoolP("force", "f", false, "Overwrite existing components if they differ")
	materializeCommandsCmd.Flags().Bool("dry-run", false, "Preview what will be materialized without making changes")
	materializeCommandsCmd.Flags().StringP("profile", "p", "", "Materialize from specific profile (use 'base' for ~/.agent-smith/)")
	materializeCmd.AddCommand(materializeCommandsCmd)

	materializeAllCmd := &cobra.Command{
		Use:   "all",
		Short: "Materialize all installed components to project directories",
		Long: `Materialize all installed components (skills, agents, commands) to project directories.

This command copies all components from ~/.agent-smith/ to .opencode/ or .claude/
with full provenance tracking. It continues on error with individual components.

The target can be specified with --target flag or AGENT_SMITH_TARGET environment variable.

EXAMPLES:
  # Materialize all components to OpenCode
  agent-smith materialize all --target opencode

  # Materialize all to both targets
  agent-smith materialize all --target all

  # Use environment variable for default target
  export AGENT_SMITH_TARGET=claudecode
  agent-smith materialize all

  # Materialize from specific profile
  agent-smith materialize all --target opencode --profile work

  # Preview without making changes
  agent-smith materialize all --target opencode --dry-run`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			target, _ := cmd.Flags().GetString("target")
			projectDir, _ := cmd.Flags().GetString("project-dir")
			force, _ := cmd.Flags().GetBool("force")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			profile, _ := cmd.Flags().GetString("profile")
			handleMaterializeAll(target, projectDir, force, dryRun, profile)
		},
	}
	materializeAllCmd.Flags().StringP("target", "t", "", "Target to materialize to (opencode, claudecode, copilot, universal, or all). Can also use AGENT_SMITH_TARGET environment variable")
	materializeAllCmd.Flags().String("project-dir", "", "Override project directory detection")
	materializeAllCmd.Flags().BoolP("force", "f", false, "Overwrite existing components if they differ")
	materializeAllCmd.Flags().Bool("dry-run", false, "Preview what will be materialized without making changes")
	materializeAllCmd.Flags().StringP("profile", "p", "", "Materialize from specific profile (use 'base' for ~/.agent-smith/)")
	materializeCmd.AddCommand(materializeAllCmd)

	materializeListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all materialized components in the current project",
		Long: `List all materialized components (skills, agents, commands) in the current project.

This command shows all components that have been materialized to .opencode/ or .claude/
directories. The output is grouped by target and component type, showing the component
name and source repository for each.

The command auto-detects the project root by walking up from the current directory
looking for .opencode/ or .claude/ directories.

EXAMPLES:
  # List all materialized components in the current project
  agent-smith materialize list

  # List from a specific project directory
  agent-smith materialize list --project-dir ~/my-project`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			projectDir, _ := cmd.Flags().GetString("project-dir")
			handleMaterializeList(projectDir)
		},
	}
	materializeListCmd.Flags().String("project-dir", "", "Override project directory detection")
	materializeCmd.AddCommand(materializeListCmd)

	materializeInfoCmd := &cobra.Command{
		Use:   "info <type> <name>",
		Short: "Show provenance information for a materialized component",
		Long: `Show detailed provenance information for a specific materialized component.

This command displays the origin and metadata for a component that has been
materialized to the current project, including source repository, commit hash,
materialization time, and sync status.

The component type must be one of: skills, agents, commands

EXAMPLES:
  # Show info for a materialized skill
  agent-smith materialize info skills my-skill

  # Show info for a materialized agent
  agent-smith materialize info agents my-agent

  # Show info for a specific target
  agent-smith materialize info skills my-skill --target opencode

  # Show info from a specific project directory
  agent-smith materialize info skills my-skill --project-dir ~/my-project`,
		Args: exactArgsWithComponentTypeValidation(2, 0, "agent-smith materialize info <type> <name>"),
		Run: func(cmd *cobra.Command, args []string) {
			target, _ := cmd.Flags().GetString("target")
			projectDir, _ := cmd.Flags().GetString("project-dir")
			source, _ := cmd.Flags().GetString("source")
			handleMaterializeInfo(args[0], args[1], target, projectDir, source)
		},
	}
	materializeInfoCmd.Flags().StringP("target", "t", "", "Optional target to check (opencode or claudecode). If not specified, shows info for all targets")
	materializeInfoCmd.Flags().String("project-dir", "", "Override project directory detection")
	materializeInfoCmd.Flags().StringP("source", "s", "", "Source URL to disambiguate when component exists in multiple sources")
	materializeCmd.AddCommand(materializeInfoCmd)

	materializeStatusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show sync status of materialized components",
		Long: `Show which materialized components are in sync or out of sync with their sources.

This command checks all materialized components in the current project and compares
them with their source components in ~/.agent-smith/ or profiles. Components are
marked as:
  ✓ in sync - materialized copy matches source
  ⚠ out of sync - source has been updated
  ✗ source missing - source component no longer installed

The command auto-detects the project root by walking up from the current directory
looking for .opencode/ or .claude/ directories.

EXAMPLES:
  # Check sync status of all materialized components
  agent-smith materialize status

  # Check status for specific target only
  agent-smith materialize status --target opencode

  # Check status in a specific project directory
  agent-smith materialize status --project-dir ~/my-project`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			target, _ := cmd.Flags().GetString("target")
			projectDir, _ := cmd.Flags().GetString("project-dir")
			handleMaterializeStatus(target, projectDir)
		},
	}
	materializeStatusCmd.Flags().StringP("target", "t", "", "Check specific target only (opencode or claudecode)")
	materializeStatusCmd.Flags().String("project-dir", "", "Override project directory detection")
	materializeCmd.AddCommand(materializeStatusCmd)

	materializeUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update out-of-sync materialized components",
		Long: `Re-materialize components where the source has been updated.

This command intelligently updates only those materialized components where the
source component has changed (smart mode by default). It:
  - Compares source hash with stored metadata
  - Re-materializes only components that are out of sync
  - Skips components that are already in sync
  - Warns and skips components with missing sources

The command auto-detects the project root and processes all target directories
(.opencode/ and .claude/) unless --target is specified.

EXAMPLES:
  # Update only out-of-sync components (smart mode)
  agent-smith materialize update

  # Force re-materialize all components
  agent-smith materialize update --force

  # Preview what would be updated
  agent-smith materialize update --dry-run

  # Update specific target only
  agent-smith materialize update --target opencode

  # Update in a specific project directory
  agent-smith materialize update --project-dir ~/my-project`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			target, _ := cmd.Flags().GetString("target")
			projectDir, _ := cmd.Flags().GetString("project-dir")
			force, _ := cmd.Flags().GetBool("force")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			source, _ := cmd.Flags().GetString("source")
			handleMaterializeUpdate(target, projectDir, source, force, dryRun)
		},
	}
	materializeUpdateCmd.Flags().StringP("target", "t", "", "Update specific target only (opencode or claudecode)")
	materializeUpdateCmd.Flags().String("project-dir", "", "Override project directory detection")
	materializeUpdateCmd.Flags().BoolP("force", "f", false, "Re-materialize all components (ignore sync status)")
	materializeUpdateCmd.Flags().Bool("dry-run", false, "Preview what would be updated without making changes")
	materializeUpdateCmd.Flags().StringP("source", "s", "", "Source URL to disambiguate when component exists in multiple sources")
	materializeCmd.AddCommand(materializeUpdateCmd)

	rootCmd.AddCommand(materializeCmd)

	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
	rootCmd.PersistentFlags().Bool("verbose", false, "Show informational output (default: show only errors)")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable verbose debug output for troubleshooting")
}

// These functions will be implemented in main.go to keep existing logic
var (
	handleAddSkill              func(repoURL, name, profile, targetDir string)
	handleAddAgent              func(repoURL, name, profile, targetDir string)
	handleAddCommand            func(repoURL, name, profile, targetDir string)
	handleAddAll                func(repoURL, profile, targetDir string)
	handleUpdate                func(componentType, componentName, profile string)
	handleUpdateAll             func(profile string)
	handleLink                  func(componentType, componentName, targetFilter, profile string)
	handleLinkAll               func(targetFilter, profile string, allProfiles bool)
	handleLinkType              func(componentType, targetFilter, profile string)
	handleAutoLink              func()
	handleListLinks             func()
	handleLinkStatus            func(allProfiles bool, profileFilter []string)
	handleUnlink                func(componentType, componentName, targetFilter string)
	handleUnlinkWithProfile     func(componentType, componentName, targetFilter, profile string)
	handleUnlinkAll             func(targetFilter string, force bool, allProfiles bool)
	handleUnlinkAllWithProfile  func(targetFilter string, force bool, allProfiles bool, profile string)
	handleUnlinkType            func(componentType, targetFilter string, force bool)
	handleUnlinkTypeWithProfile func(componentType, targetFilter string, force bool, profile string)
	handleUninstall             func(componentType, componentName, profile string)
	handleUninstallAll          func(repoURL string, force bool)
	handleProfilesList          func(profileFilter []string, activeOnly bool, typeFilter string)
	handleProfilesShow          func(profileName string)
	handleProfilesCreate        func(profileName string)
	handleProfilesDelete        func(profileName string)
	handleProfilesActivate      func(profileName string)
	handleProfilesDeactivate    func()
	handleProfilesAdd           func(componentType, profileName, componentName string)
	handleProfilesCopy          func(componentType, sourceProfile, targetProfile, componentName string)
	handleProfilesRemove        func(componentType, profileName, componentName string)
	handleProfilesCherryPick    func(targetProfile string, sourceProfiles []string)
	handleStatus                func()
	handleTargetAdd             func(name, path string)
	handleTargetRemove          func(name string)
	handleTargetList            func()
	handleMaterializeComponent  func(componentType, componentName, target, projectDir string, force, dryRun bool, fromProfile, source string)
	handleMaterializeType       func(componentType, target, projectDir string, force, dryRun bool, fromProfile string)
	handleMaterializeAll        func(target, projectDir string, force, dryRun bool, fromProfile string)
	handleMaterializeList       func(projectDir string)
	handleMaterializeInfo       func(componentType, componentName, target, projectDir, source string)
	handleMaterializeStatus     func(target, projectDir string)
	handleMaterializeUpdate     func(target, projectDir, source string, force, dryRun bool)
)

func SetHandlers(
	addSkill func(repoURL, name, profile, targetDir string),
	addAgent func(repoURL, name, profile, targetDir string),
	addCommand func(repoURL, name, profile, targetDir string),
	addAll func(repoURL, profile, targetDir string),
	update func(componentType, componentName, profile string),
	updateAll func(profile string),
	link func(componentType, componentName, targetFilter, profile string),
	linkAll func(targetFilter, profile string, allProfiles bool),
	linkType func(componentType, targetFilter, profile string),
	autoLink func(),
	listLinks func(),
	linkStatus func(allProfiles bool, profileFilter []string),
	unlink func(componentType, componentName, targetFilter string),
	unlinkWithProfile func(componentType, componentName, targetFilter, profile string),
	unlinkAll func(targetFilter string, force bool, allProfiles bool),
	unlinkAllWithProfile func(targetFilter string, force bool, allProfiles bool, profile string),
	unlinkType func(componentType, targetFilter string, force bool),
	unlinkTypeWithProfile func(componentType, targetFilter string, force bool, profile string),
	uninstall func(componentType, componentName, profile string),
	uninstallAll func(repoURL string, force bool),
	profilesList func(profileFilter []string, activeOnly bool, typeFilter string),
	profilesShow func(profileName string),
	profilesCreate func(profileName string),
	profilesDelete func(profileName string),
	profilesActivate func(profileName string),
	profilesDeactivate func(),
	profilesAdd func(componentType, profileName, componentName string),
	profilesCopy func(componentType, sourceProfile, targetProfile, componentName string),
	profilesRemove func(componentType, profileName, componentName string),
	profilesCherryPick func(targetProfile string, sourceProfiles []string),
	status func(),
	targetAdd func(name, path string),
	targetRemove func(name string),
	targetList func(),
	materializeComponent func(componentType, componentName, target, projectDir string, force, dryRun bool, fromProfile, source string),
	materializeType func(componentType, target, projectDir string, force, dryRun bool, fromProfile string),
	materializeAll func(target, projectDir string, force, dryRun bool, fromProfile string),
	materializeList func(projectDir string),
	materializeInfo func(componentType, componentName, target, projectDir, source string),
	materializeStatus func(target, projectDir string),
	materializeUpdate func(target, projectDir, source string, force, dryRun bool),
) {
	handleAddSkill = addSkill
	handleAddAgent = addAgent
	handleAddCommand = addCommand
	handleAddAll = addAll
	handleUpdate = update
	handleUpdateAll = updateAll
	handleLink = link
	handleLinkAll = linkAll
	handleLinkType = linkType
	handleAutoLink = autoLink
	handleListLinks = listLinks
	handleLinkStatus = linkStatus
	handleUnlink = unlink
	handleUnlinkWithProfile = unlinkWithProfile
	handleUnlinkAll = unlinkAll
	handleUnlinkAllWithProfile = unlinkAllWithProfile
	handleUnlinkType = unlinkType
	handleUnlinkTypeWithProfile = unlinkTypeWithProfile
	handleUninstall = uninstall
	handleUninstallAll = uninstallAll
	handleProfilesList = profilesList
	handleProfilesShow = profilesShow
	handleProfilesCreate = profilesCreate
	handleProfilesDelete = profilesDelete
	handleProfilesActivate = profilesActivate
	handleProfilesDeactivate = profilesDeactivate
	handleProfilesAdd = profilesAdd
	handleProfilesCopy = profilesCopy
	handleProfilesRemove = profilesRemove
	handleProfilesCherryPick = profilesCherryPick
	handleStatus = status
	handleTargetAdd = targetAdd
	handleTargetRemove = targetRemove
	handleTargetList = targetList
	handleMaterializeComponent = materializeComponent
	handleMaterializeType = materializeType
	handleMaterializeAll = materializeAll
	handleMaterializeList = materializeList
	handleMaterializeInfo = materializeInfo
	handleMaterializeStatus = materializeStatus
	handleMaterializeUpdate = materializeUpdate
}
