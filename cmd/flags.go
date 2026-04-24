package cmd

import "github.com/spf13/cobra"

func addMaterializeFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("target", "t", "", "Target to materialize to (opencode, claudecode, copilot, universal, or all). Can also use AGENT_SMITH_TARGET environment variable")
	cmd.Flags().String("project-dir", "", "Override project directory detection")
	cmd.Flags().BoolP("force", "f", false, "Overwrite existing component if it differs")
	cmd.Flags().Bool("dry-run", false, "Preview what will be materialized without making changes")
	cmd.Flags().StringP("profile", "p", "", "Materialize from specific profile (use 'base' for ~/.agent-smith/)")
}

func addSourceFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("source", "s", "", "Source URL to disambiguate when component exists in multiple sources")
}

func addLinkTargetFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("to", "t", "", "Target editor (opencode, claudecode, copilot, universal, or all)")
	cmd.Flags().String("profile", "", "Link from specific profile (bypasses active profile)")
}

func addUnlinkFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("target", "t", "", "Target to unlink from (opencode, claudecode, copilot, or all). Default: unlink from all detected targets")
	cmd.Flags().StringP("profile", "p", "", "Unlink from a specific profile without switching to it")
}

func addForceFlag(cmd *cobra.Command) {
	cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}

func addInstallFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("profile", "p", "", "Force creation of a new profile with a custom name")
	cmd.Flags().StringP("install-dir", "i", "", "Install to a custom directory (isolated from ~/.agent-smith/)")
}

func addUninstallComponentFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("profile", "p", "", "Remove from a specific profile instead of ~/.agent-smith/")
	cmd.Flags().StringP("source", "s", "", "Source repository URL to disambiguate when the same name exists in multiple sources")
}
