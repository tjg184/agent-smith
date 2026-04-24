package status

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/pkg/config"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/profiles"
	"github.com/tjg184/agent-smith/pkg/services"
)

type Service struct {
	profileManager *profiles.ProfileManager
	logger         *logger.Logger
	formatter      *formatter.Formatter
}

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

func (s *Service) ShowSystemStatus() error {
	s.logger.Debug("[DEBUG] ShowSystemStatus called")

	activeProfile, err := s.profileManager.GetActiveProfile()
	if err != nil {
		return fmt.Errorf("failed to get active profile: %w", err)
	}
	s.logger.Debug("[DEBUG] Active profile: %s", activeProfile)

	s.logger.Debug("[DEBUG] Detecting targets")
	targets, err := config.DetectAllTargets()
	if err != nil {
		return fmt.Errorf("failed to detect targets: %w", err)
	}
	s.logger.Debug("[DEBUG] Detected %d target(s)", len(targets))

	f := formatter.New()
	f.SectionHeader("Agent Smith Status")

	if activeProfile != "" {
		green := color.New(color.FgGreen).SprintFunc()
		s.formatter.Info("  Active Profile:     %s %s", green(activeProfile), formatter.ColoredSuccess())
	} else {
		gray := color.New(color.FgHiBlack).SprintFunc()
		s.formatter.Info("  Active Profile:     %s", gray("None"))
	}

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

	profilesList, err := s.profileManager.ScanProfiles()
	if err == nil && len(profilesList) > 0 {
		s.formatter.EmptyLine()
		bold := color.New(color.Bold).SprintFunc()
		s.formatter.Info("%s", bold("Installed Components"))
		for _, profile := range profilesList {
			agents, skills, commands := s.profileManager.CountComponents(profile)
			total := agents + skills + commands
			if activeProfile == profile.Name {
				green := color.New(color.FgGreen).SprintFunc()
				s.formatter.Info("  • %s %s  (%d components)", green(profile.Name), formatter.ColoredSuccess(), total)
			} else {
				s.formatter.Info("  • %s  (%d components)", profile.Name, total)
			}
		}
	}

	s.formatter.EmptyLine()
	dim := color.New(color.Faint).SprintFunc()
	s.formatter.Info("%s", dim("Quick Actions:"))
	s.formatter.Info("  %s agent-smith link status     %s", dim("•"), dim("View component link status"))
	s.formatter.Info("  %s agent-smith profile list    %s", dim("•"), dim("List all profiles"))
	s.formatter.EmptyLine()

	return nil
}

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
