package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
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

	profilesStatusCmd := &cobra.Command{
		Use:   "status [profile-name]",
		Short: "Show detailed information about a profile",
		Long: `Display detailed information about a specific profile.

If no profile name is provided, shows information about the currently active profile.

This command shows:
  - Profile name and active status
  - Profile location on disk
  - List of all agents in the profile
  - List of all skills in the profile
  - List of all commands in the profile

Use this before activating a profile to see exactly what components it contains.

EXAMPLES:
  # Show details of active profile
  agent-smith profile status
  
  # Show details of a specific profile
  agent-smith profile status my-profile
  
  # View contents before activating
  agent-smith profile status work-profile
  agent-smith profile activate work-profile`,
		Args: rangeArgsWithHelp(0, 1, "agent-smith profile status [profile-name]"),
		Run: func(cmd *cobra.Command, args []string) {
			profileName := ""
			if len(args) > 0 {
				profileName = args[0]
			}
			handleProfilesShow(profileName)
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

	// profiles share - Generate commands to recreate a profile
	profilesShareCmd := &cobra.Command{
		Use:   "share <profile-name>",
		Short: "Generate commands to recreate a profile",
		Long: `Generate a shareable text file containing the exact commands needed to recreate a profile.

The output includes all install commands required to recreate the profile from scratch.
You can share this with teammates, commit it to version control, or use it as backup.

Components installed from local paths are skipped as they cannot be recreated.

EXAMPLES:
  # Display commands to stdout
  agent-smith profile share work

  # Save to file
  agent-smith profile share work --output setup-work.txt

  # Copy to clipboard (macOS)
  agent-smith profile share work | pbcopy

  # Share base installation
  agent-smith profile share base --output my-setup.txt`,
		Args: exactArgsWithHelp(1, "agent-smith profile share <profile-name>"),
		Run: func(cmd *cobra.Command, args []string) {
			outputFile, _ := cmd.Flags().GetString("output")
			handleProfilesShare(args[0], outputFile)
		},
	}
	profilesShareCmd.Flags().StringP("output", "o", "", "Save commands to file instead of stdout")

	profilesRenameCmd := &cobra.Command{
		Use:   "rename <old-name> <new-name>",
		Short: "Rename a user-created profile",
		Long: `Rename a user-created profile (👤) to a new name.

The new name must follow the same rules as profile creation: only letters,
numbers, and hyphens are allowed.

If the profile is currently active, you will be prompted to confirm. Existing
symlinks are removed and automatically restored under the new name.

Only user-created profiles can be renamed. Repository-sourced profiles (📦)
cannot be renamed.

EXAMPLES:
  # Rename an inactive profile
  agent-smith profile rename old-name new-name

  # Rename the active profile (will prompt for confirmation)
  agent-smith profile rename my-profile my-new-profile`,
		Args: exactArgsWithHelp(2, "agent-smith profile rename <old-name> <new-name>"),
		Run: func(cmd *cobra.Command, args []string) {
			handleProfilesRename(args[0], args[1])
		},
	}

	profilesCmd.AddCommand(profilesListCmd)
	profilesCmd.AddCommand(profilesStatusCmd)
	profilesCmd.AddCommand(profilesCreateCmd)
	profilesCmd.AddCommand(profilesDeleteCmd)
	profilesCmd.AddCommand(profilesActivateCmd)
	profilesCmd.AddCommand(profilesDeactivateCmd)
	profilesCmd.AddCommand(profilesAddCmd)
	profilesCmd.AddCommand(profilesCopyCmd)
	profilesCmd.AddCommand(profilesRemoveCmd)
	profilesCmd.AddCommand(profilesCherryPickCmd)
	profilesCmd.AddCommand(profilesShareCmd)
	profilesCmd.AddCommand(profilesRenameCmd)
	rootCmd.AddCommand(profilesCmd)
}
