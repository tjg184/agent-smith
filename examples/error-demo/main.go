// Package main demonstrates the colored error messages
package main

import (
	"fmt"
	"os"

	"github.com/tgaines/agent-smith/pkg/errors"
	"github.com/tgaines/agent-smith/pkg/logger"
)

func main() {
	// Create a logger
	log := logger.Default(false, false)

	fmt.Println("=== Demonstrating Colored Error Messages ===\n")

	// Example 1: Profile not found error
	fmt.Println("1. Profile Not Found Error:")
	errMsg1 := errors.NewProfileNotFoundError("my-project")
	log.ErrorMsg(errMsg1)

	// Example 2: Invalid flags error
	fmt.Println("\n2. Invalid Flags Error:")
	errMsg2 := errors.NewInvalidFlagsError("--profile", "--target-dir")
	log.ErrorMsg(errMsg2)

	// Example 3: Component download error
	fmt.Println("\n3. Component Download Error:")
	errMsg3 := errors.NewComponentDownloadError("skill", "https://github.com/user/repo",
		fmt.Errorf("repository not found: 404"))
	log.ErrorMsg(errMsg3)

	// Example 4: Component linker error
	fmt.Println("\n4. Component Linker Error:")
	errMsg4 := errors.NewComponentLinkerError("skill", "OpenCode",
		fmt.Errorf("permission denied: cannot create symlink"))
	log.ErrorMsg(errMsg4)

	// Example 5: No active profile error
	fmt.Println("\n5. No Active Profile Error:")
	errMsg5 := errors.NewNoActiveProfileError()
	log.ErrorMsg(errMsg5)

	// Example 6: Invalid component type error
	fmt.Println("\n6. Invalid Component Type Error:")
	errMsg6 := errors.NewInvalidComponentTypeError("widget", []string{"skills", "agents", "commands"})
	log.ErrorMsg(errMsg6)

	// Example 7: Simple error with color
	fmt.Println("\n7. Simple Error Messages:")
	log.Error("Failed to download skill: repository not found")

	// Example 8: Simple warning with color
	fmt.Println("\n8. Simple Warning:")
	log.Warn("Profile directory already exists, using existing profile")

	// Example 9: Custom structured error
	fmt.Println("\n9. Custom Structured Error:")
	customErr := errors.New("Failed to parse configuration file").
		WithContext("The configuration file contains invalid YAML syntax").
		WithDetails(
			"Line 15: unexpected character '{'",
			"Expected a valid YAML mapping or sequence",
		).
		WithSuggestion("Check your YAML syntax using a validator").
		WithExample("yamllint ~/.agent-smith/config.yaml")
	log.ErrorMsg(customErr)

	fmt.Println("\n=== End of Demonstration ===")
	os.Exit(0)
}
