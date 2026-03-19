package linker

import (
	"fmt"
	"path/filepath"

	"github.com/tjg184/agent-smith/internal/detector"
	"github.com/tjg184/agent-smith/pkg/config"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/paths"
)

// BuildOptions controls how a ComponentLinker is constructed.
//
// Profile resolution order:
//  1. If ExplicitProfile is set, use that profile directory directly.
//  2. If ActiveProfile is set, use that profile directory.
//  3. Otherwise use the base agents directory (~/.agent-smith/agents).
//
// TargetFilter selects a subset of detected targets. Empty string or "all"
// means all detected targets are used.
//
// ProfileManager, when non-nil, is passed through to the ComponentLinker so it
// can resolve profile information during linking operations (e.g. auto-link).
type BuildOptions struct {
	// ExplicitProfile is a pre-validated profile name whose directory is already
	// known to exist. The caller is responsible for validation before setting this.
	ExplicitProfile string

	// ActiveProfile is the currently active profile name, looked up from state.
	// Ignored when ExplicitProfile is set.
	ActiveProfile string

	// TargetFilter restricts which targets are included. Empty or "all" = all targets.
	TargetFilter string

	// Targets is the pre-detected list of available targets. If nil, Build detects them.
	Targets []config.Target

	// ProfileManager is forwarded to NewComponentLinker unchanged (may be nil).
	ProfileManager ProfileManager
}

// Build constructs a ComponentLinker from opts. The logger is attached to the
// repository detector; passing nil is safe (no debug output).
func Build(opts BuildOptions, log *logger.Logger) (*ComponentLinker, error) {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}

	if opts.ExplicitProfile != "" {
		profilesDir, err := paths.GetProfilesDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get profiles directory: %w", err)
		}
		agentsDir = filepath.Join(profilesDir, opts.ExplicitProfile)
	} else if opts.ActiveProfile != "" {
		profilesDir, err := paths.GetProfilesDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get profiles directory: %w", err)
		}
		agentsDir = filepath.Join(profilesDir, opts.ActiveProfile)
	}

	targets := opts.Targets
	if targets == nil {
		targets, err = config.DetectAllTargets()
		if err != nil {
			return nil, fmt.Errorf("failed to detect targets: %w", err)
		}
	}

	if opts.TargetFilter != "" && opts.TargetFilter != "all" {
		targets = filterByName(targets, opts.TargetFilter)
		if len(targets) == 0 {
			allNames := targetNames(opts.Targets)
			if allNames == nil {
				all, _ := config.DetectAllTargets()
				allNames = targetNames(all)
			}
			return nil, fmt.Errorf("target '%s' not found. Available targets: %v", opts.TargetFilter, allNames)
		}
	}

	det := detector.NewRepositoryDetector()
	if log != nil {
		det.SetLogger(log)
	}

	return NewComponentLinker(agentsDir, targets, det, opts.ProfileManager)
}

func filterByName(targets []config.Target, name string) []config.Target {
	for _, t := range targets {
		if t.GetName() == name {
			return []config.Target{t}
		}
	}
	return nil
}

func targetNames(targets []config.Target) []string {
	names := make([]string, len(targets))
	for i, t := range targets {
		names[i] = t.GetName()
	}
	return names
}
