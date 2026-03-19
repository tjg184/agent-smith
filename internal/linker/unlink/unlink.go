package linkerUnlink

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/internal/linker/profilepicker"
	"github.com/tjg184/agent-smith/pkg/config"
	"github.com/tjg184/agent-smith/pkg/paths"
)

// UnlinkComponent removes a linked component from configured targets.
// targetFilter: "" or "all" = all targets; specific name = only that target.
func UnlinkComponent(agentsDir string, targets []config.Target, f *formatter.Formatter, componentType, componentName, targetFilter string) error {
	if componentType != "skills" && componentType != "agents" && componentType != "commands" {
		return fmt.Errorf("invalid component type: %s (must be skills, agents, or commands)", componentType)
	}

	targetsToUnlink := filterTargets(targets, targetFilter)
	if len(targetsToUnlink) == 0 {
		if targetFilter != "" && targetFilter != "all" {
			availableTargets := make([]string, 0, len(targets))
			for _, target := range targets {
				availableTargets = append(availableTargets, target.GetName())
			}

			if len(availableTargets) == 0 {
				return fmt.Errorf("target '%s' does not exist and no targets are configured", targetFilter)
			}

			return fmt.Errorf("target '%s' does not exist\n\nAvailable targets:\n  - %s\n\nExample:\n  agent-smith unlink %s %s --target %s",
				targetFilter,
				strings.Join(availableTargets, "\n  - "),
				componentType,
				componentName,
				availableTargets[0])
		}
		return fmt.Errorf("no targets available")
	}

	successCount := 0
	failedCount := 0
	var errors []string
	var unlinkedTargets []string

	if targetFilter != "" && targetFilter != "all" {
		f.SectionHeader(fmt.Sprintf("Unlinking %s '%s' from: %s", componentType, componentName, targetFilter))
	} else {
		targetNames := make([]string, 0, len(targetsToUnlink))
		for _, target := range targetsToUnlink {
			targetNames = append(targetNames, target.GetName())
		}
		f.SectionHeader(fmt.Sprintf("Unlinking %s '%s' from: %s", componentType, componentName, strings.Join(targetNames, ", ")))
	}

	for _, target := range targetsToUnlink {
		componentDir, err := target.GetGlobalComponentDir(componentType)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to get target component directory for %s: %v", target.GetName(), err))
			failedCount++
			continue
		}

		targetName := target.GetName()

		if componentType == "commands" || componentType == "agents" {
			srcComponentTypeDir := filepath.Join(agentsDir, componentType)
			if !isFlatMdLinked(componentName, srcComponentTypeDir, componentDir) {
				continue
			}

			f.ProgressMsg(fmt.Sprintf("Unlinking from %s", targetName), componentName)

			if err := unlinkFlatMdFiles(componentName, srcComponentTypeDir, componentDir); err != nil {
				f.ProgressFailed()
				errors = append(errors, fmt.Sprintf("failed to unlink from %s: %v", targetName, err))
				failedCount++
				continue
			}

			f.ProgressComplete()
			unlinkedTargets = append(unlinkedTargets, targetName)
			successCount++
			continue
		}

		linkPath := filepath.Join(componentDir, componentName)

		if _, err := os.Lstat(linkPath); os.IsNotExist(err) {
			continue
		}

		linkType, targetPath, _ := analyzeLinkStatus(linkPath)

		if linkType == "copied" {
			f.WarningMsg("'%s' is a copied directory in %s, not a symlink", componentName, targetName)
			fmt.Printf("This will permanently delete: %s\n", linkPath)
			fmt.Print("Continue? [y/N]: ")

			var response string
			fmt.Scanln(&response)

			if strings.ToLower(strings.TrimSpace(response)) != "y" {
				f.InfoMsg("Unlink cancelled for %s", targetName)
				continue
			}
		}

		f.ProgressMsg(fmt.Sprintf("Unlinking from %s", targetName), componentName)

		if linkType == "copied" {
			if err := os.RemoveAll(linkPath); err != nil {
				f.ProgressFailed()
				errors = append(errors, fmt.Sprintf("failed to remove copied directory from %s: %v", targetName, err))
				failedCount++
				continue
			}
		} else {
			if err := os.Remove(linkPath); err != nil {
				f.ProgressFailed()
				errors = append(errors, fmt.Sprintf("failed to remove link from %s: %v", targetName, err))
				failedCount++
				continue
			}
		}

		f.ProgressComplete()
		unlinkedTargets = append(unlinkedTargets, targetName)

		if linkType == "symlink" && targetPath != "" {
			f.DetailItem("Source", targetPath)
		}
		successCount++
	}

	if len(errors) > 0 {
		for _, errMsg := range errors {
			f.WarningMsg("%s", errMsg)
		}
		if successCount == 0 {
			return fmt.Errorf("failed to unlink from any target")
		}
	}

	if successCount == 0 {
		return fmt.Errorf("component %s/%s is not linked to any target", componentType, componentName)
	}

	f.EmptyLine()
	f.CounterSummary(successCount+failedCount, successCount, failedCount, 0)

	return nil
}

// UnlinkComponentsByType removes all linked components of a specific type.
func UnlinkComponentsByType(agentsDir string, targets []config.Target, f *formatter.Formatter, componentType, targetFilter string, force bool) error {
	targetsToUnlink := filterTargets(targets, targetFilter)
	if len(targetsToUnlink) == 0 {
		if targetFilter != "" && targetFilter != "all" {
			availableTargets := make([]string, 0, len(targets))
			for _, target := range targets {
				availableTargets = append(availableTargets, target.GetName())
			}

			if len(availableTargets) == 0 {
				return fmt.Errorf("target '%s' does not exist and no targets are configured", targetFilter)
			}

			return fmt.Errorf("target '%s' does not exist\n\nAvailable targets:\n  - %s\n\nExample:\n  agent-smith unlink %s --target %s",
				targetFilter,
				strings.Join(availableTargets, "\n  - "),
				componentType,
				availableTargets[0])
		}
		return fmt.Errorf("no targets available")
	}

	totalLinks := 0
	copiedDirs := 0

	for _, target := range targetsToUnlink {
		componentDir, err := target.GetGlobalComponentDir(componentType)
		if err != nil {
			return fmt.Errorf("failed to get target component directory: %w", err)
		}

		if _, err := os.Stat(componentDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(componentDir)
		if err != nil {
			return fmt.Errorf("failed to read %s directory: %w", componentType, err)
		}

		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			fullPath := filepath.Join(componentDir, entry.Name())
			linkType, _, _ := analyzeLinkStatus(fullPath)

			if linkType == "copied" {
				copiedDirs++
				continue
			}
			totalLinks++
		}
	}

	if totalLinks == 0 && copiedDirs == 0 {
		f.InfoMsg("No linked %s found", componentType)
		return nil
	}

	if !force {
		if totalLinks > 0 {
			targetNames := make([]string, 0, len(targetsToUnlink))
			for _, target := range targetsToUnlink {
				targetNames = append(targetNames, target.GetName())
			}
			targetStr := strings.Join(targetNames, ", ")

			if targetFilter != "" && targetFilter != "all" {
				fmt.Printf("This will unlink %d %s from: %s", totalLinks, componentType, targetStr)
			} else {
				fmt.Printf("This will unlink %d %s from all targets", totalLinks, componentType)
			}
			fmt.Println()
		}
		if copiedDirs > 0 {
			f.InfoMsg("Note: %d copied directories will be skipped (not deleted)", copiedDirs)
		}
		if totalLinks == 0 {
			f.InfoMsg("No symlinked %s to unlink (only copied directories found)", componentType)
			return nil
		}
		fmt.Print("Continue? [y/N]: ")

		var response string
		fmt.Scanln(&response)

		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("Unlink cancelled.")
			return nil
		}
	}

	if targetFilter != "" && targetFilter != "all" {
		f.SectionHeader(fmt.Sprintf("Unlinking all %s from: %s", componentType, targetFilter))
	} else {
		targetNames := make([]string, 0, len(targetsToUnlink))
		for _, target := range targetsToUnlink {
			targetNames = append(targetNames, target.GetName())
		}
		f.SectionHeader(fmt.Sprintf("Unlinking all %s from: %s", componentType, strings.Join(targetNames, ", ")))
	}

	removedCount := 0
	skippedCount := 0
	errorCount := 0

	for _, target := range targetsToUnlink {
		componentDir, err := target.GetGlobalComponentDir(componentType)
		if err != nil {
			return fmt.Errorf("failed to get target component directory: %w", err)
		}

		if _, err := os.Stat(componentDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(componentDir)
		if err != nil {
			f.WarningMsg("Failed to read %s directory for %s: %v", componentType, target.GetName(), err)
			continue
		}

		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			fullPath := filepath.Join(componentDir, entry.Name())
			linkType, _, _ := analyzeLinkStatus(fullPath)

			if linkType == "copied" {
				skippedCount++
				continue
			}

			f.ProgressMsg(fmt.Sprintf("Unlinking %s from %s", componentType, target.GetName()), entry.Name())

			err := os.Remove(fullPath)

			if err != nil {
				f.ProgressFailed()
				f.WarningMsg("Failed to unlink %s/%s from %s: %v", componentType, entry.Name(), target.GetName(), err)
				errorCount++
			} else {
				f.ProgressComplete()
				removedCount++
			}
		}
	}

	f.EmptyLine()
	f.CounterSummary(removedCount+errorCount, removedCount, errorCount, skippedCount)

	return nil
}

// UnlinkAllComponents removes all linked components from configured targets.
// allProfiles: if true, unlinks from all profiles; if false, only from current profile (agentsDir).
func UnlinkAllComponents(agentsDir string, targets []config.Target, f *formatter.Formatter, targetFilter string, force bool, allProfiles bool) error {
	targetsToUnlink := filterTargets(targets, targetFilter)
	if len(targetsToUnlink) == 0 {
		if targetFilter != "" && targetFilter != "all" {
			availableTargets := make([]string, 0, len(targets))
			for _, target := range targets {
				availableTargets = append(availableTargets, target.GetName())
			}

			if len(availableTargets) == 0 {
				return fmt.Errorf("target '%s' does not exist and no targets are configured", targetFilter)
			}

			return fmt.Errorf("target '%s' does not exist\n\nAvailable targets:\n  - %s\n\nExample:\n  agent-smith unlink all --target %s",
				targetFilter,
				strings.Join(availableTargets, "\n  - "),
				availableTargets[0])
		}
		return fmt.Errorf("no targets available")
	}

	componentTypes := paths.GetComponentTypes()

	totalLinks := 0
	copiedDirs := 0
	skippedProfilesCount := 0
	skippedProfilesMap := make(map[string]int)

	for _, target := range targetsToUnlink {
		for _, componentType := range componentTypes {
			componentDir, err := target.GetGlobalComponentDir(componentType)
			if err != nil {
				return fmt.Errorf("failed to get target component directory: %w", err)
			}

			if _, err := os.Stat(componentDir); os.IsNotExist(err) {
				continue
			}

			entries, err := os.ReadDir(componentDir)
			if err != nil {
				return fmt.Errorf("failed to read %s directory: %w", componentType, err)
			}

			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), ".") {
					continue
				}

				fullPath := filepath.Join(componentDir, entry.Name())
				linkType, _, _ := analyzeLinkStatus(fullPath)

				if linkType == "copied" {
					copiedDirs++
					continue
				}

				if linkType == "symlink" || linkType == "broken" {
					fromAgentSmith, err := isSymlinkFromAgentSmith(agentsDir, fullPath)
					if err == nil && !fromAgentSmith {
						skippedProfilesCount++
						continue
					}
				}

				if !allProfiles && (linkType == "symlink" || linkType == "broken") {
					belongsToProfile, err := isSymlinkFromCurrentProfile(agentsDir, fullPath)
					if err == nil && !belongsToProfile {
						profileName := profilepicker.GetProfileNameFromSymlink(fullPath)
						if profileName != "" {
							skippedProfilesMap[profileName]++
						}
						skippedProfilesCount++
						continue
					}
				}

				totalLinks++
			}
		}
	}

	if totalLinks == 0 && copiedDirs == 0 && skippedProfilesCount == 0 {
		f.InfoMsg("No linked components found")
		return nil
	}

	currentProfileName := getProfileFromPath(agentsDir)
	profilesExist := anyProfilesExist()

	if !force {
		if totalLinks > 0 {
			targetStr := "all targets"
			if targetFilter != "" && targetFilter != "all" {
				targetStr = targetFilter
			} else if len(targetsToUnlink) > 0 {
				targetNames := make([]string, 0, len(targetsToUnlink))
				for _, target := range targetsToUnlink {
					targetNames = append(targetNames, target.GetName())
				}
				targetStr = strings.Join(targetNames, ", ")
			}

			profileMsg := ""
			if allProfiles {
				profileMsg = " from all profiles"
			} else if profilesExist {
				if currentProfileName == paths.BaseProfileName {
					profileMsg = " from base installation"
				} else {
					profileMsg = fmt.Sprintf(" from profile '%s'", currentProfileName)
				}
			}
			fmt.Printf("This will unlink %d symlinked components%s from: %s", totalLinks, profileMsg, targetStr)
			fmt.Println()
		}
		if copiedDirs > 0 {
			f.InfoMsg("Note: %d copied directories will be skipped (not deleted)", copiedDirs)
		}
		if skippedProfilesCount > 0 && !allProfiles {
			profileNames := make([]string, 0, len(skippedProfilesMap))
			for profileName := range skippedProfilesMap {
				profileNames = append(profileNames, profileName)
			}
			if len(profileNames) == 1 {
				f.InfoMsg("Note: %d components from profile '%s' will be skipped", skippedProfilesCount, profileNames[0])
			} else if len(profileNames) > 1 {
				f.InfoMsg("Note: %d components from other profiles will be skipped:", skippedProfilesCount)
				for _, profileName := range profileNames {
					count := skippedProfilesMap[profileName]
					fmt.Printf("  - %s: %d components\n", profileName, count)
				}
			}
		}
		if totalLinks == 0 {
			if skippedProfilesCount > 0 {
				if currentProfileName == paths.BaseProfileName {
					f.InfoMsg("No symlinked components from base installation to unlink (found %d from profiles)", skippedProfilesCount)
				} else {
					f.InfoMsg("No symlinked components from profile '%s' to unlink (found %d from other profiles)", currentProfileName, skippedProfilesCount)
				}
			} else {
				f.InfoMsg("No symlinked components to unlink (only copied directories found)")
			}
			return nil
		}
		fmt.Print("Continue? [y/N]: ")

		var response string
		fmt.Scanln(&response)

		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("Unlink cancelled.")
			return nil
		}
	}

	headerMsg := buildHeaderMsg(agentsDir, targetFilter, targetsToUnlink, allProfiles, profilesExist, currentProfileName)
	f.SectionHeader(headerMsg)

	removedCount := 0
	skippedCount := 0
	skippedByProfile := make(map[string][]string)
	errorCount := 0

	for _, target := range targetsToUnlink {
		for _, componentType := range componentTypes {
			componentDir, err := target.GetGlobalComponentDir(componentType)
			if err != nil {
				return fmt.Errorf("failed to get target component directory: %w", err)
			}

			if _, err := os.Stat(componentDir); os.IsNotExist(err) {
				continue
			}

			entries, err := os.ReadDir(componentDir)
			if err != nil {
				f.WarningMsg("Failed to read %s directory for %s: %v", componentType, target.GetName(), err)
				continue
			}

			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), ".") {
					continue
				}

				fullPath := filepath.Join(componentDir, entry.Name())
				linkType, _, _ := analyzeLinkStatus(fullPath)

				if linkType == "copied" {
					skippedCount++
					continue
				}

				if linkType == "symlink" || linkType == "broken" {
					fromAgentSmith, err := isSymlinkFromAgentSmith(agentsDir, fullPath)
					if err == nil && !fromAgentSmith {
						skippedCount++
						continue
					}
				}

				if !allProfiles && (linkType == "symlink" || linkType == "broken") {
					belongsToProfile, err := isSymlinkFromCurrentProfile(agentsDir, fullPath)
					if err == nil && !belongsToProfile {
						profileName := profilepicker.GetProfileNameFromSymlink(fullPath)
						if profileName != "" {
							skippedByProfile[profileName] = append(skippedByProfile[profileName], fmt.Sprintf("%s/%s", componentType, entry.Name()))
						}
						skippedCount++
						continue
					}
				}

				profileNote := ""
				if profilesExist && (linkType == "symlink" || linkType == "broken") {
					profileName := profilepicker.GetProfileNameFromSymlink(fullPath)
					if profileName != "" && profileName != paths.BaseProfileName {
						profileNote = fmt.Sprintf(" [%s]", profileName)
					}
				}

				f.ProgressMsg(fmt.Sprintf("Unlinking %s from %s%s", componentType, target.GetName(), profileNote), entry.Name())

				err := os.Remove(fullPath)

				if err != nil {
					f.ProgressFailed()
					f.WarningMsg("Failed to unlink %s/%s from %s: %v", componentType, entry.Name(), target.GetName(), err)
					errorCount++
				} else {
					f.ProgressComplete()
					removedCount++
				}
			}
		}
	}

	f.EmptyLine()
	f.CounterSummary(removedCount+errorCount, removedCount, errorCount, skippedCount)

	if len(skippedByProfile) > 0 && !allProfiles {
		f.EmptyLine()
		fmt.Println("Skipped components from other profiles:")
		for profileName, components := range skippedByProfile {
			fmt.Printf("  Profile '%s' (%d components):\n", profileName, len(components))
			for _, comp := range components {
				fmt.Printf("    - %s\n", comp)
			}
		}
		f.EmptyLine()
		if currentProfileName == paths.BaseProfileName {
			f.InfoMsg("Use --all-profiles flag to unlink components from all profiles")
		} else {
			f.InfoMsg("Use --all-profiles flag to unlink components from all profiles, or switch to the respective profile")
		}
	}

	return nil
}

func buildHeaderMsg(agentsDir, targetFilter string, targetsToUnlink []config.Target, allProfiles, profilesExist bool, currentProfileName string) string {
	if targetFilter != "" && targetFilter != "all" {
		if allProfiles {
			return fmt.Sprintf("Unlinking all components (all profiles) from: %s", targetFilter)
		} else if profilesExist {
			if currentProfileName == paths.BaseProfileName {
				return fmt.Sprintf("Unlinking components (base installation) from: %s", targetFilter)
			}
			return fmt.Sprintf("Unlinking components (profile '%s') from: %s", currentProfileName, targetFilter)
		}
		return fmt.Sprintf("Unlinking components from: %s", targetFilter)
	}

	targetNames := make([]string, 0, len(targetsToUnlink))
	for _, target := range targetsToUnlink {
		targetNames = append(targetNames, target.GetName())
	}
	targetList := strings.Join(targetNames, ", ")

	if allProfiles {
		return fmt.Sprintf("Unlinking all components (all profiles) from: %s", targetList)
	} else if profilesExist {
		if currentProfileName == paths.BaseProfileName {
			return fmt.Sprintf("Unlinking components (base installation) from: %s", targetList)
		}
		return fmt.Sprintf("Unlinking components (profile '%s') from: %s", currentProfileName, targetList)
	}
	return fmt.Sprintf("Unlinking components from: %s", targetList)
}

func filterTargets(targets []config.Target, targetFilter string) []config.Target {
	if targetFilter == "" || targetFilter == "all" {
		return targets
	}

	filtered := make([]config.Target, 0)
	for _, target := range targets {
		if target.GetName() == targetFilter {
			filtered = append(filtered, target)
		}
	}
	return filtered
}

// IsSymlinkFromCurrentProfile reports whether the symlink points into the same profile
// as agentsDir. Exported so the parent linker package can provide a wrapper for tests.
func IsSymlinkFromCurrentProfile(agentsDir, symlinkPath string) (bool, error) {
	return isSymlinkFromCurrentProfile(agentsDir, symlinkPath)
}

func isSymlinkFromCurrentProfile(agentsDir, symlinkPath string) (bool, error) {
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		return false, err
	}

	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(symlinkPath), target)
	}

	target = filepath.Clean(target)

	currentProfile := getProfileFromPath(agentsDir)
	targetProfile := getProfileFromPath(target)
	return currentProfile == targetProfile, nil
}

func isSymlinkFromAgentSmith(agentsDir, symlinkPath string) (bool, error) {
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		return false, err
	}

	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(symlinkPath), target)
	}

	target = filepath.Clean(target)

	cleanAgentsDir := filepath.Clean(agentsDir)
	if strings.HasPrefix(target, cleanAgentsDir) {
		return true, nil
	}

	baseAgentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return false, nil
	}

	if strings.HasPrefix(target, baseAgentsDir) {
		return true, nil
	}

	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		return false, nil
	}

	if strings.HasPrefix(target, profilesDir) {
		return true, nil
	}

	return false, nil
}

func anyProfilesExist() bool {
	profilesDir, err := paths.GetProfilesDir()
	if err != nil {
		return false
	}

	info, err := os.Stat(profilesDir)
	if err != nil || !info.IsDir() {
		return false
	}

	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			return true
		}
	}

	return false
}

func getProfileFromPath(path string) string {
	path = filepath.Clean(path)

	parent := filepath.Dir(path)
	if filepath.Base(parent) == "profiles" {
		return filepath.Base(path)
	}

	dir := parent
	for {
		grandparent := filepath.Dir(dir)
		if filepath.Base(grandparent) == "profiles" {
			return filepath.Base(dir)
		}
		if grandparent == dir || grandparent == "." || grandparent == "/" {
			return paths.BaseProfileName
		}
		dir = grandparent
	}
}

func analyzeLinkStatus(path string) (linkType string, target string, valid bool) {
	info, err := os.Lstat(path)
	if err != nil {
		return "missing", "", false
	}

	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return "broken", "", false
		}

		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(path), target)
		}

		if _, err := os.Stat(target); err == nil {
			return "symlink", target, true
		}
		return "broken", target, false
	}

	if info.IsDir() {
		return "copied", path, true
	}

	return "unknown", "", false
}

func isFlatMdLinked(componentName, componentTypeDir, targetBaseDir string) bool {
	componentRoot := filepath.Clean(filepath.Join(componentTypeDir, componentName))
	expectedPrefix := componentRoot + string(filepath.Separator)

	found := false
	_ = filepath.WalkDir(targetBaseDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || found {
			return nil
		}

		info, err := os.Lstat(path)
		if err != nil || info.Mode()&os.ModeSymlink == 0 {
			return nil
		}

		target, err := os.Readlink(path)
		if err != nil {
			return nil
		}

		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(path), target)
		}
		target = filepath.Clean(target)

		if target == componentRoot || strings.HasPrefix(target, expectedPrefix) {
			found = true
		}

		return nil
	})

	return found
}

func unlinkFlatMdFiles(componentName, componentTypeDir, targetBaseDir string) error {
	componentRoot := filepath.Clean(filepath.Join(componentTypeDir, componentName))
	expectedPrefix := componentRoot + string(filepath.Separator)

	return filepath.WalkDir(targetBaseDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}

		info, err := os.Lstat(path)
		if err != nil || info.Mode()&os.ModeSymlink == 0 {
			return nil
		}

		target, err := os.Readlink(path)
		if err != nil {
			return nil
		}

		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(path), target)
		}
		target = filepath.Clean(target)

		pointsIntoComponent := target == componentRoot ||
			strings.HasPrefix(target, expectedPrefix)

		if pointsIntoComponent {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
		}

		return nil
	})
}
