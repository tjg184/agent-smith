//go:build integration
// +build integration

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// AgentSmithBinary is the path to the compiled agent-smith binary.
// It is built once at the start of the test suite and shared across all tests.
var AgentSmithBinary string

// TestMain runs before all tests and compiles the binary once.
// This significantly speeds up integration tests by avoiding 40+ recompilations.
func TestMain(m *testing.M) {
	// Create temp directory for binary
	tempDir, err := os.MkdirTemp("", "agent-smith-suite-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp directory: %v\n", err)
		os.Exit(1)
	}

	// Build binary once for all tests
	AgentSmithBinary = filepath.Join(tempDir, "agent-smith")
	repoRoot := filepath.Join("..", "..")

	fmt.Printf("Building agent-smith binary once for all integration tests...\n")
	cmd := exec.Command("go", "build", "-o", AgentSmithBinary, ".")
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build agent-smith: %v\nOutput: %s\n", err, string(output))
		os.RemoveAll(tempDir)
		os.Exit(1)
	}
	fmt.Printf("Binary compiled at: %s\n", AgentSmithBinary)

	// Run all tests
	code := m.Run()

	// Cleanup
	fmt.Printf("Cleaning up test binary...\n")
	os.RemoveAll(tempDir)
	os.Exit(code)
}
