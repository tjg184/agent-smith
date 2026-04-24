package cmd

import (
	"github.com/spf13/cobra"
)

func makeLinkRun(componentType string) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		targetFilter, _ := cmd.Flags().GetString("to")
		profile, _ := cmd.Flags().GetString("profile")
		handleLink(componentType, args[0], targetFilter, profile)
	}
}

func makeLinkTypeRun(componentType string) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		targetFilter, _ := cmd.Flags().GetString("to")
		profile, _ := cmd.Flags().GetString("profile")
		handleLinkType(componentType, targetFilter, profile)
	}
}

func init() {
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
  --to, -t <target>     Target editor (opencode, claudecode, copilot, universal, or all)
                        Default: all detected editors
  --profile <name>      Link FROM specific profile (advanced)

By default, links components from your active repo.
Use --profile <name> for advanced profile switching.`,
	}

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
		Run:  makeLinkRun("skills"),
	}
	addLinkTargetFlags(linkSkillCmd)
	linkCmd.AddCommand(linkSkillCmd)

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
		Run:  makeLinkTypeRun("skills"),
	}
	addLinkTargetFlags(linkSkillsCmd)
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
		Run:  makeLinkRun("agents"),
	}
	addLinkTargetFlags(linkAgentCmd)
	linkCmd.AddCommand(linkAgentCmd)

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
		Run:  makeLinkTypeRun("agents"),
	}
	addLinkTargetFlags(linkAgentsCmd)
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
		Run:  makeLinkRun("commands"),
	}
	addLinkTargetFlags(linkCommandCmd)
	linkCmd.AddCommand(linkCommandCmd)

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
		Run:  makeLinkTypeRun("commands"),
	}
	addLinkTargetFlags(linkCommandsCmd)
	linkCmd.AddCommand(linkCommandsCmd)

	linkAllCmd := &cobra.Command{
		Use:   "all [repository-url]",
		Short: "Link all components to editors",
		Long: `Link all components (skills, agents, commands) to AI editor targets.

This is the most common command - it links everything you've installed to your editors.

Optionally provide a repository URL to link only components from that specific repo.

EXAMPLES:
  # Link everything to all editors
  agent-smith link all

  # Link components from a specific repository only
  agent-smith link all owner/repo

  # Link to a specific editor only
  agent-smith link all --to opencode

  # Link to the universal target (~/.agents/)
  agent-smith link all --to universal

ADVANCED:
  # Link from a named profile
  agent-smith link all --profile work

  # Link from all profiles simultaneously
  agent-smith link all --all-profiles`,
		Args: rangeArgsWithHelp(0, 1, "agent-smith link all [repository-url]"),
		Run: func(cmd *cobra.Command, args []string) {
			targetFilter, _ := cmd.Flags().GetString("to")
			profile, _ := cmd.Flags().GetString("profile")
			allProfiles, _ := cmd.Flags().GetBool("all-profiles")

			var repoURL string
			if len(args) == 1 {
				repoURL = args[0]
			}

			handleLinkAll(targetFilter, profile, repoURL, allProfiles)
		},
	}
	addLinkTargetFlags(linkAllCmd)
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

By default shows all installed repos. Use --profile to scope to a specific one.

EXAMPLES:
  # Show status for all installed repos
  agent-smith link status

  # Show status for a specific repo's components
  agent-smith link status --profile owner-repo

  # Show only linked components (hide unlinked)
  agent-smith link status --linked-only

LEGEND:
  ✓ - Valid symlink (linked and working)
  ◆ - Copied directory (linked but copied, not symlinked)
  ✗ - Broken link (link exists but target is missing)
  - - Not linked
  ? - Unknown status`,
		Args: noArgsWithHelp,
		Run: func(cmd *cobra.Command, args []string) {
			profileFilter, _ := cmd.Flags().GetStringSlice("profile")
			linkedOnly, _ := cmd.Flags().GetBool("linked-only")
			handleLinkStatus(true, profileFilter, linkedOnly)
		},
	}
	linkStatusCmd.Flags().StringSlice("profile", []string{}, "Scope to a specific profile")
	linkStatusCmd.Flags().Bool("linked-only", false, "Show only components that have at least one link")
	linkCmd.AddCommand(linkStatusCmd)

	rootCmd.AddCommand(linkCmd)
}
