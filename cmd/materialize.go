package cmd

import (
	"github.com/spf13/cobra"
)

func makeMaterializeComponentRun(componentType string) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		projectDir, _ := cmd.Flags().GetString("project-dir")
		force, _ := cmd.Flags().GetBool("force")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		profile, _ := cmd.Flags().GetString("profile")
		source, _ := cmd.Flags().GetString("source")
		handleMaterializeComponent(componentType, args[0], target, projectDir, force, dryRun, profile, source)
	}
}

func makeMaterializeTypeRun(componentType string) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		projectDir, _ := cmd.Flags().GetString("project-dir")
		force, _ := cmd.Flags().GetBool("force")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		profile, _ := cmd.Flags().GetString("profile")
		handleMaterializeType(componentType, target, projectDir, force, dryRun, profile)
	}
}

func init() {
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
		Run:  makeMaterializeComponentRun("skills"),
	}
	addMaterializeFlags(materializeSkillCmd)
	addSourceFlag(materializeSkillCmd)
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
		Run:  makeMaterializeComponentRun("agents"),
	}
	addMaterializeFlags(materializeAgentCmd)
	addSourceFlag(materializeAgentCmd)
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
		Run:  makeMaterializeComponentRun("commands"),
	}
	addMaterializeFlags(materializeCommandCmd)
	addSourceFlag(materializeCommandCmd)
	materializeCmd.AddCommand(materializeCommandCmd)

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
		Run:  makeMaterializeTypeRun("skills"),
	}
	addMaterializeFlags(materializeSkillsCmd)
	materializeSkillsCmd.Flags().Lookup("force").Usage = "Overwrite existing components if they differ"
	materializeCmd.AddCommand(materializeSkillsCmd)

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
		Run:  makeMaterializeTypeRun("agents"),
	}
	addMaterializeFlags(materializeAgentsCmd)
	materializeAgentsCmd.Flags().Lookup("force").Usage = "Overwrite existing components if they differ"
	materializeCmd.AddCommand(materializeAgentsCmd)

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
		Run:  makeMaterializeTypeRun("commands"),
	}
	addMaterializeFlags(materializeCommandsCmd)
	materializeCommandsCmd.Flags().Lookup("force").Usage = "Overwrite existing components if they differ"
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
	addMaterializeFlags(materializeAllCmd)
	materializeAllCmd.Flags().Lookup("force").Usage = "Overwrite existing components if they differ"
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
	addSourceFlag(materializeInfoCmd)
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
	addSourceFlag(materializeUpdateCmd)
	materializeCmd.AddCommand(materializeUpdateCmd)

	rootCmd.AddCommand(materializeCmd)
}
