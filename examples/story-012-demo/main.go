// Package main demonstrates the new helpful error messages added in Story-012
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

	fmt.Println("=== Story-012: Helpful Error Messages Demo ===\n")

	// Example 1: Agents Directory Error
	fmt.Println("1. Agents Directory Error (permission denied):")
	errMsg1 := errors.NewAgentsDirectoryError(fmt.Errorf("permission denied"))
	log.ErrorMsg(errMsg1)

	// Example 2: Agents Directory Error (not found)
	fmt.Println("\n2. Agents Directory Error (not initialized):")
	errMsg2 := errors.NewAgentsDirectoryError(fmt.Errorf("no such file or directory"))
	log.ErrorMsg(errMsg2)

	// Example 3: Target Detection Error
	fmt.Println("\n3. Target Detection Error:")
	errMsg3 := errors.NewTargetDetectionError(fmt.Errorf("no targets found"))
	log.ErrorMsg(errMsg3)

	// Example 4: Active Profile Error
	fmt.Println("\n4. Active Profile Error:")
	errMsg4 := errors.NewActiveProfileError(fmt.Errorf("profile configuration corrupted"))
	log.ErrorMsg(errMsg4)

	// Example 5: Unknown Component Type Error
	fmt.Println("\n5. Unknown Component Type Error:")
	errMsg5 := errors.NewUnknownComponentTypeError("widget")
	log.ErrorMsg(errMsg5)

	// Example 6: Lock File Error (corrupted)
	fmt.Println("\n6. Lock File Error (corrupted):")
	errMsg6 := errors.NewLockFileError("read", "skills", fmt.Errorf("invalid YAML: unmarshal error"))
	log.ErrorMsg(errMsg6)

	// Example 7: Project Detection Error
	fmt.Println("\n7. Project Detection Error:")
	errMsg7 := errors.NewProjectDetectionError(fmt.Errorf("no .opencode or .claude directory found"))
	log.ErrorMsg(errMsg7)

	// Example 8: Materialization Error (not found)
	fmt.Println("\n8. Materialization Error (component not found):")
	errMsg8 := errors.NewMaterializationError("skill", "my-skill", fmt.Errorf("component does not exist"))
	log.ErrorMsg(errMsg8)

	// Example 9: Materialization Error (already exists)
	fmt.Println("\n9. Materialization Error (already exists):")
	errMsg9 := errors.NewMaterializationError("agent", "my-agent", fmt.Errorf("already exists"))
	log.ErrorMsg(errMsg9)

	// Example 10: Missing Arguments Error
	fmt.Println("\n10. Missing Arguments Error:")
	errMsg10 := errors.NewMissingArgumentsError("agent-smith install", "agent-smith install <type> <repo-url> <name>")
	log.ErrorMsg(errMsg10)

	fmt.Println("\n=== All error messages provide:")
	fmt.Println("  ✓ Clear description of what went wrong")
	fmt.Println("  ✓ Contextual information about the error")
	fmt.Println("  ✓ Actionable suggestions for resolution")
	fmt.Println("  ✓ Example commands when applicable")
	fmt.Println("\n=== End of Demo ===")
	os.Exit(0)
}
