package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
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
}
