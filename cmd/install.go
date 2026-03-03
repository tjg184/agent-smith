package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
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
		Long: `Download and install a skill from a git repository.

This command fetches a skill from any git repository (GitHub, GitLab, Bitbucket, or private)
and installs it to a repository-sourced profile. The skill will be automatically detected
if it contains a SKILL.md file.

AUTOMATIC PROFILE CREATION:
By default, this command creates a repository-sourced profile (📦) to namespace the skill
from the repository. The profile name is generated from the repository URL (e.g., "owner-repo").
If a profile already exists for the same repository, it will be reused and updated.

This behavior is consistent with 'install all' and makes it easy to:
  - Keep components from the same repo organized together
  - Update all components from the repo with 'update all'
  - Add more components to the same profile later

BEHAVIOR WITH MULTIPLE SKILLS:
When a repository contains multiple skills, the <skill-name> parameter is used to select
which skill to install. If the specified skill name is not found, the command will fail
and list all available skills in the repository.

REQUIRED PARAMETERS:
  <repository-url>  The URL or path to the git repository containing the skill
  <skill-name>      The name of the skill to install (also used as the local directory name)

EXAMPLES:
  # Download from GitHub using shorthand
  # Creates profile: openai-cookbook (📦)
  agent-smith install skill openai/cookbook gpt-skill

  # Download a specific skill from a repository with multiple skills
  # Installs to profile: example-skills-repo (📦)
  agent-smith install skill example/skills-repo my-specific-skill

  # Download using full URL
  agent-smith install skill https://github.com/example/repo my-skill

  # Download from local repository
  agent-smith install skill /path/to/local/skill local-skill

  # Force creation of a new profile with a custom name
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
	installSkillCmd.Flags().StringP("profile", "p", "", "Force creation of a new profile with a custom name")
	installSkillCmd.Flags().StringP("install-dir", "i", "", "Install to a custom directory (isolated from ~/.agent-smith/)")
	installCmd.AddCommand(installSkillCmd)

	installAgentCmd := &cobra.Command{
		Use:   "agent <repository-url> <agent-name>",
		Short: "Download an agent from a git repository",
		Long: `Download and install an AI agent from a git repository.

This command fetches an agent from any git repository (GitHub, GitLab, Bitbucket, or private)
and installs it to a repository-sourced profile. The agent will be automatically detected
based on path patterns and file extensions.

AUTOMATIC PROFILE CREATION:
By default, this command creates a repository-sourced profile (📦) to namespace the agent
from the repository. The profile name is generated from the repository URL (e.g., "owner-repo").
If a profile already exists for the same repository, it will be reused and updated.

This behavior is consistent with 'install all' and makes it easy to:
  - Keep components from the same repo organized together
  - Update all components from the repo with 'update all'
  - Add more components to the same profile later

BEHAVIOR WITH MULTIPLE AGENTS:
When a repository contains multiple agents, the <agent-name> parameter is used to select
which agent to install. If the specified agent name is not found, the command will fail
and list all available agents in the repository.

REQUIRED PARAMETERS:
  <repository-url>  The URL or path to the git repository containing the agent
  <agent-name>      The name of the agent to install (also used as the local directory name)

EXAMPLES:
  # Download from GitHub using shorthand
  # Creates profile: openai-assistant (📦)
  agent-smith install agent openai/assistant coding-agent

  # Download a specific agent from a repository with multiple agents
  # Installs to profile: example-agents-repo (📦)
  agent-smith install agent example/agents-repo my-specific-agent

  # Download using full URL
  agent-smith install agent https://github.com/example/repo my-agent

  # Download from local repository
  agent-smith install agent /path/to/local/agent local-agent

  # Force creation of a new profile with a custom name
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
	installAgentCmd.Flags().StringP("profile", "p", "", "Force creation of a new profile with a custom name")
	installAgentCmd.Flags().StringP("install-dir", "i", "", "Install to a custom directory (isolated from ~/.agent-smith/)")
	installCmd.AddCommand(installAgentCmd)

	installCommandCmd := &cobra.Command{
		Use:   "command <repository-url> <command-name>",
		Short: "Download a command from a git repository",
		Long: `Download and install a command-line tool from a git repository.

This command fetches a command from any git repository (GitHub, GitLab, Bitbucket, or private)
and installs it to a repository-sourced profile. The command will be automatically detected
based on path patterns and file extensions.

AUTOMATIC PROFILE CREATION:
By default, this command creates a repository-sourced profile (📦) to namespace the command
from the repository. The profile name is generated from the repository URL (e.g., "owner-repo").
If a profile already exists for the same repository, it will be reused and updated.

This behavior is consistent with 'install all' and makes it easy to:
  - Keep components from the same repo organized together
  - Update all components from the repo with 'update all'
  - Add more components to the same profile later

BEHAVIOR WITH MULTIPLE COMMANDS:
When a repository contains multiple commands, the <command-name> parameter is used to select
which command to install. If the specified command name is not found, the command will fail
and list all available commands in the repository.

REQUIRED PARAMETERS:
  <repository-url>  The URL or path to the git repository containing the command
  <command-name>    The name of the command to install (also used as the local directory name)

EXAMPLES:
  # Download from GitHub using shorthand
  # Creates profile: cli-tools-formatter (📦)
  agent-smith install command cli-tools/formatter json-formatter

  # Download a specific command from a repository with multiple commands
  # Installs to profile: example-commands-repo (📦)
  agent-smith install command example/commands-repo my-specific-command

  # Download using full URL
  agent-smith install command https://github.com/example/tool my-tool

  # Download from local repository
  agent-smith install command /path/to/local/command local-cmd

  # Force creation of a new profile with a custom name
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
	installCommandCmd.Flags().StringP("profile", "p", "", "Force creation of a new profile with a custom name")
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
}
