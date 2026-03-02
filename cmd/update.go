package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
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
}
