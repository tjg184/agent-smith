// Package main demonstrates the error message system in agent-smith
//
// This program showcases the various error messages available to help developers
// quickly identify and resolve issues.
//
// Run: go run examples/error-messages-demo.go
package main

import (
	"fmt"

	"github.com/tgaines/agent-smith/pkg/errors"
)

func main() {
	fmt.Println("=== Agent Smith Error Message System Demo ===\n")
	fmt.Println("This demo showcases the helpful error messages available in agent-smith.\n")

	// Example 1: Profile Not Found
	fmt.Println("Example 1: Profile Not Found")
	fmt.Println("Command: agent-smith profile activate work")
	fmt.Println(errors.NewProfileNotFoundError("work").Format())
	fmt.Println()

	// Example 2: Component Download Error
	fmt.Println("Example 2: Component Download Error (Authentication Failure)")
	fmt.Println("Command: agent-smith install skill owner/private-repo")
	downloadErr := errors.NewComponentDownloadError("skill", "https://github.com/owner/private-repo", fmt.Errorf("authentication failed"))
	fmt.Println(downloadErr.Format())
	fmt.Println()

	// Example 3: Component Download Error (Network Timeout)
	fmt.Println("Example 3: Component Download Error (Network Timeout)")
	fmt.Println("Command: agent-smith install agent company/internal-tools")
	networkErr := errors.NewComponentDownloadError("agent", "https://github.com/company/internal-tools", fmt.Errorf("network timeout"))
	fmt.Println(networkErr.Format())
	fmt.Println()

	// Example 4: Target Not Found
	fmt.Println("Example 4: Target Not Found")
	fmt.Println("Command: agent-smith link skill my-skill --target vscode")
	fmt.Println(errors.NewTargetNotFoundError("vscode").Format())
	fmt.Println()

	// Example 5: Target Directory Not Found
	fmt.Println("Example 5: Target Directory Not Found")
	fmt.Println("Command: agent-smith materialize skill my-skill --target opencode")
	fmt.Println(errors.NewTargetDirectoryNotFoundError("opencode").Format())
	fmt.Println()

	// Example 6: Invalid Target
	fmt.Println("Example 6: Invalid Target")
	fmt.Println("Command: agent-smith materialize skill my-skill --target invalid")
	fmt.Println(errors.NewInvalidTargetError("invalid").Format())
	fmt.Println()

	// Example 7: Component Not Installed
	fmt.Println("Example 7: Component Not Installed")
	fmt.Println("Command: agent-smith materialize skill nonexistent-skill --target opencode")
	fmt.Println(errors.NewComponentNotInstalledError("skill", "nonexistent-skill", "~/.agent-smith/skills/").Format())
	fmt.Println()

	// Example 8: Component Linker Error
	fmt.Println("Example 8: Component Linker Error (Permission Denied)")
	fmt.Println("Command: agent-smith link skill my-skill")
	linkerErr := errors.NewComponentLinkerError("skill", "OpenCode", fmt.Errorf("permission denied"))
	fmt.Println(linkerErr.Format())
	fmt.Println()

	// Example 9: Lock File Error
	fmt.Println("Example 9: Lock File Error (Corrupted)")
	fmt.Println("Command: agent-smith link all")
	lockErr := errors.NewLockFileError("read", "skills", fmt.Errorf("invalid YAML: unmarshal error"))
	fmt.Println(lockErr.Format())
	fmt.Println()

	// Example 10: Git Operation Error
	fmt.Println("Example 10: Git Operation Error (SSH Key Issue)")
	fmt.Println("Command: agent-smith install all git@github.com:org/repo.git")
	gitErr := errors.NewGitOperationError("clone", "git@github.com:org/repo.git", fmt.Errorf("ssh: handshake failed"))
	fmt.Println(gitErr.Format())
	fmt.Println()

	// Example 11: Materialization Error
	fmt.Println("Example 11: Materialization Error (Component Not Found)")
	fmt.Println("Command: agent-smith materialize skill my-skill --target opencode")
	materializeErr := errors.NewMaterializationError("skill", "my-skill", fmt.Errorf("component does not exist"))
	fmt.Println(materializeErr.Format())
	fmt.Println()

	// Example 12: Component Not Found in Project
	fmt.Println("Example 12: Component Not Found in Project (With Suggestions)")
	fmt.Println("Command: agent-smith info skill nonexistent-skill")
	availableComponents := []string{"skill-a", "skill-b", "skill-c"}
	componentErr := errors.NewComponentNotFoundInProjectError("skill", "nonexistent-skill", availableComponents)
	fmt.Println(componentErr.Format())
	fmt.Println()

	// Example 13: Invalid Flags
	fmt.Println("Example 13: Invalid Flags (Mutually Exclusive)")
	fmt.Println("Command: agent-smith install skill repo --profile work --target-dir ./local")
	fmt.Println(errors.NewInvalidFlagsError("--profile", "--target-dir").Format())
	fmt.Println()

	// Example 14: Missing Target Flag
	fmt.Println("Example 14: Missing Target Flag")
	fmt.Println("Command: agent-smith materialize skill my-skill")
	fmt.Println(errors.NewMissingTargetFlagError("materialize skill my-skill").Format())
	fmt.Println()

	// Example 15: Project Detection Error
	fmt.Println("Example 15: Project Detection Error")
	fmt.Println("Command: agent-smith materialize skill my-skill (outside project)")
	projectErr := errors.NewProjectDetectionError(fmt.Errorf("no .opencode or .claude directory found"))
	fmt.Println(projectErr.Format())
	fmt.Println()

	fmt.Println("=== End of Demo ===")
	fmt.Println("\nAll these error messages are:")
	fmt.Println("  ✓ Color-coded for easy identification")
	fmt.Println("  ✓ Include clear context about what went wrong")
	fmt.Println("  ✓ Provide actionable suggestions for resolution")
	fmt.Println("  ✓ Show concrete command examples")
	fmt.Println("  ✓ Help developers quickly resolve issues")
}
