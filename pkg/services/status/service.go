package status

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/tgaines/agent-smith/internal/formatter"
	"github.com/tgaines/agent-smith/pkg/config"
	"github.com/tgaines/agent-smith/pkg/logger"
	"github.com/tgaines/agent-smith/pkg/paths"
	"github.com/tgaines/agent-smith/pkg/profiles"
	"github.com/tgaines/agent-smith/pkg/services"
)

// Service implements the StatusService interface
type Service struct {
	profileManager *profiles.ProfileManager
	logger         *logger.Logger
	formatter      *formatter.Formatter
}

// NewService creates a new StatusService with dependencies injected
func NewService(
	profileManager *profiles.ProfileManager,
	logger *logger.Logger,
	formatter *formatter.Formatter,
) services.StatusService {
	return &Service{
		profileManager: profileManager,
		logger:         logger,
		formatter:      formatter,
	}
}

// ShowSystemStatus displays the current system status including active profile,
// detected targets, and component counts
func (s *Service) ShowSystemStatus() error {
	s.logger.Debug("[DEBUG] ShowSystemStatus called")

	// Get active profile
	activeProfile, err := s.profileManager.GetActiveProfile()
	if err != nil {
		return fmt.Errorf("failed to get active profile: %w", err)
	}
	s.logger.Debug("[DEBUG] Active profile: %s", activeProfile)

	// Detect all available targets
	s.logger.Debug("[DEBUG] Detecting targets")
	targets, err := config.DetectAllTargets()
	if err != nil {
		return fmt.Errorf("failed to detect targets: %w", err)
	}
	s.logger.Debug("[DEBUG] Detected %d target(s)", len(targets))

	// Get agents directory
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return fmt.Errorf("failed to get agents directory: %w", err)
	}
	s.logger.Debug("[DEBUG] Agents directory: %s", agentsDir)

	// Count components in ~/.agent-smith/
	agentsCount, skillsCount, commandsCount := s.countBaseComponents(agentsDir)

	// Display status with modern formatting
	f := formatter.New()
	f.SectionHeader("Agent Smith Status")

	// Show active profile
	if activeProfile != "" {
		green := color.New(color.FgGreen).SprintFunc()
		s.formatter.Info("  Active Profile:     %s %s", green(activeProfile), formatter.ColoredSuccess())
	} else {
		gray := color.New(color.FgHiBlack).SprintFunc()
		s.formatter.Info("  Active Profile:     %s", gray("None"))
	}

	// Show detected targets
	if len(targets) > 0 {
		var targetNames []string
		for _, target := range targets {
			targetNames = append(targetNames, target.GetName())
		}
		cyan := color.New(color.FgCyan).SprintFunc()
		s.formatter.Info("  Detected Targets:   %s", cyan(s.joinStrings(targetNames, ", ")))
	} else {
		gray := color.New(color.FgHiBlack).SprintFunc()
		s.formatter.Info("  Detected Targets:   %s", gray("None"))
	}

	// Show base components count
	s.formatter.EmptyLine()
	bold := color.New(color.Bold).SprintFunc()
	s.formatter.Info("%s", bold("Base Components (~/.agent-smith/)"))
	s.formatter.Info("  • Agents:           %d", agentsCount)
	s.formatter.Info("  • Skills:           %d", skillsCount)
	s.formatter.Info("  • Commands:         %d", commandsCount)

	// If there's an active profile, show its components
	if activeProfile != "" {
		profilesList, err := s.profileManager.ScanProfiles()
		if err == nil {
			for _, profile := range profilesList {
				if profile.Name == activeProfile {
					agents, skills, commands := s.profileManager.CountComponents(profile)
					s.formatter.EmptyLine()
					green := color.New(color.FgGreen, color.Bold).SprintFunc()
					s.formatter.Info("%s", green("Active Profile Components"))
					s.formatter.Info("  • Agents:           %d", agents)
					s.formatter.Info("  • Skills:           %d", skills)
					s.formatter.Info("  • Commands:         %d", commands)
					break
				}
			}
		}
	}

	// Show helpful commands
	s.formatter.EmptyLine()
	dim := color.New(color.Faint).SprintFunc()
	s.formatter.Info("%s", dim("Quick Actions:"))
	s.formatter.Info("  %s agent-smith link status     %s", dim("•"), dim("View component link status"))
	s.formatter.Info("  %s agent-smith profile list    %s", dim("•"), dim("List all profiles"))
	s.formatter.EmptyLine()

	return nil
}

// countBaseComponents counts the number of components in the base directory
func (s *Service) countBaseComponents(baseDir string) (agents, skills, commands int) {
	agentsPath := filepath.Join(baseDir, "agents")
	skillsPath := filepath.Join(baseDir, "skills")
	commandsPath := filepath.Join(baseDir, "commands")

	if entries, err := os.ReadDir(agentsPath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				agents++
			}
		}
	}

	if entries, err := os.ReadDir(skillsPath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				skills++
			}
		}
	}

	if entries, err := os.ReadDir(commandsPath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				commands++
			}
		}
	}

	return agents, skills, commands
}

// joinStrings joins a slice of strings with a separator
func (s *Service) joinStrings(strings []string, separator string) string {
	if len(strings) == 0 {
		return ""
	}
	result := strings[0]
	for i := 1; i < len(strings); i++ {
		result += separator + strings[i]
	}
	return result
}
