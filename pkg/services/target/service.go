package target

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/pkg/config"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/services"
)

type Service struct {
	logger    *logger.Logger
	formatter *formatter.Formatter
}

func NewService(
	logger *logger.Logger,
	formatter *formatter.Formatter,
) services.TargetService {
	return &Service{
		logger:    logger,
		formatter: formatter,
	}
}

func (s *Service) AddCustomTarget(name, path string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	for _, target := range cfg.CustomTargets {
		if target.Name == name {
			return fmt.Errorf("target '%s' already exists in config", name)
		}
	}

	newTarget := config.CustomTargetConfig{
		Name:        name,
		BaseDir:     path,
		SkillsDir:   "skills",
		AgentsDir:   "agents",
		CommandsDir: "commands",
	}

	cfg.CustomTargets = append(cfg.CustomTargets, newTarget)

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	s.logger.Info("%s Successfully added custom target '%s'", formatter.SymbolSuccess, name)
	s.logger.Info("  Base directory: %s", path)
	s.logger.Info("\nSubdirectories:")
	s.logger.Info("  Skills:   %s/skills", path)
	s.logger.Info("  Agents:   %s/agents", path)
	s.logger.Info("  Commands: %s/commands", path)
	s.logger.Info("\nYou can now link components to this target:")
	s.logger.Info("  agent-smith link all --target %s", name)

	return nil
}

func (s *Service) RemoveCustomTarget(name string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	found := false
	targetIndex := -1
	for i, target := range cfg.CustomTargets {
		if target.Name == name {
			found = true
			targetIndex = i
			break
		}
	}

	if !found {
		return fmt.Errorf("target '%s' not found in custom targets", name)
	}

	cfg.CustomTargets = append(cfg.CustomTargets[:targetIndex], cfg.CustomTargets[targetIndex+1:]...)

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	s.logger.Info("%s Successfully removed custom target '%s'", formatter.SymbolSuccess, name)
	s.logger.Info("\nNote: This only removes the target from configuration.")
	s.logger.Info("Components linked to this target are not automatically unlinked.")

	return nil
}

func (s *Service) ListTargets() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	builtInNames := config.GetAllTargetTypes()

	f := formatter.New()
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	f.SectionHeader("Available Targets")

	type targetInfo struct {
		name     string
		baseDir  string
		exists   bool
		isCustom bool
		hasError bool
	}
	var allTargets []targetInfo

	for _, name := range builtInNames {
		target, err := config.NewTarget(name)

		if err != nil {
			allTargets = append(allTargets, targetInfo{
				name:     name,
				baseDir:  "error loading target",
				exists:   false,
				isCustom: false,
				hasError: true,
			})
			continue
		}

		baseDir, _ := target.GetGlobalBaseDir()
		exists := false
		if _, err := os.Stat(baseDir); err == nil {
			exists = true
		}

		allTargets = append(allTargets, targetInfo{
			name:     name,
			baseDir:  baseDir,
			exists:   exists,
			isCustom: false,
			hasError: false,
		})
	}

	for _, customTargetConfig := range cfg.CustomTargets {
		customTarget, err := config.NewCustomTarget(customTargetConfig)
		if err != nil {
			allTargets = append(allTargets, targetInfo{
				name:     customTargetConfig.Name,
				baseDir:  "error loading target",
				exists:   false,
				isCustom: true,
				hasError: true,
			})
			continue
		}

		baseDir, _ := customTarget.GetGlobalBaseDir()
		exists := false
		if _, err := os.Stat(baseDir); err == nil {
			exists = true
		}

		allTargets = append(allTargets, targetInfo{
			name:     customTargetConfig.Name,
			baseDir:  baseDir,
			exists:   exists,
			isCustom: true,
			hasError: false,
		})
	}

	table := formatter.NewBoxTable(os.Stdout, []string{"Status", "Target", "Type", "Location"})

	availableCount := 0
	for _, target := range allTargets {
		var statusSymbol string
		var targetType string

		if target.hasError {
			statusSymbol = red(formatter.SymbolError)
		} else if target.exists {
			statusSymbol = green(formatter.SymbolSuccess)
			availableCount++
		} else {
			statusSymbol = yellow(formatter.SymbolNotLinked)
		}

		if target.isCustom {
			targetType = "Custom"
		} else {
			targetType = "Built-in"
		}

		table.AddRow([]string{statusSymbol, target.name, targetType, target.baseDir})
	}

	table.Render()

	s.formatter.EmptyLine()
	totalCount := len(allTargets)
	if availableCount == totalCount {
		s.formatter.Info("%s All %d target(s) detected and available", green(formatter.SymbolSuccess), totalCount)
	} else if availableCount > 0 {
		s.formatter.Info("%s %d of %d target(s) available", yellow(formatter.SymbolWarning), availableCount, totalCount)
	} else {
		s.formatter.Info("%s No targets currently available", red(formatter.SymbolError))
	}

	s.formatter.EmptyLine()
	s.formatter.Info("Legend:")
	s.formatter.Info("  %s Available  %s Not found  %s Error",
		green(formatter.SymbolSuccess),
		yellow(formatter.SymbolNotLinked),
		red(formatter.SymbolError))

	return nil
}
