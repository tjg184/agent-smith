package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

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
