package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/tgaines/agent-smith/cmd"
	"github.com/tgaines/agent-smith/internal/detector"
	"github.com/tgaines/agent-smith/internal/downloader"
	"github.com/tgaines/agent-smith/internal/fileutil"
	"github.com/tgaines/agent-smith/internal/formatter"
	"github.com/tgaines/agent-smith/internal/linker"
	metadataPkg "github.com/tgaines/agent-smith/internal/metadata"
	"github.com/tgaines/agent-smith/internal/models"
	"github.com/tgaines/agent-smith/internal/uninstaller"
	"github.com/tgaines/agent-smith/internal/updater"
	"github.com/tgaines/agent-smith/pkg/config"
	"github.com/tgaines/agent-smith/pkg/paths"
	"github.com/tgaines/agent-smith/pkg/profiles"
)

type BulkDownloader = downloader.BulkDownloader

// Re-export types for backward compatibility
type UpdateDetector = updater.UpdateDetector
type ComponentLockFile = metadataPkg.ComponentLockFile

// Cross-platform helper functions
func getCrossPlatformPermissions() os.FileMode {
	return fileutil.GetCrossPlatformPermissions()
}

func getCrossPlatformFilePermissions() os.FileMode {
	return fileutil.GetCrossPlatformFilePermissions()
}

func createDirectoryWithPermissions(path string) error {
	return fileutil.CreateDirectoryWithPermissions(path)
}

func createFileWithPermissions(path string, data []byte) error {
	return fileutil.CreateFileWithPermissions(path, data)
}

// parseFrontmatter extracts YAML frontmatter from a markdown file
// Frontmatter must be delimited by "---" at the start of the file
// Returns nil if no frontmatter is found (not an error)
func parseFrontmatter(filePath string) (*models.ComponentFrontmatter, error) {
	return fileutil.ParseFrontmatter(filePath)
}

// determineComponentName determines the component name using frontmatter or filename
// Priority: frontmatter.name > filename (without extension)
// Special files (README.md, index.md, main.md) are skipped
func determineComponentName(frontmatter *models.ComponentFrontmatter, fileName string) string {
	return fileutil.DetermineComponentName(frontmatter, fileName)
}

// determineDestinationFolderName determines the destination folder name using hierarchy heuristic
// Walks up from component file directory, skipping component-type names (agents/commands/skills)
// Returns first non-component-type directory name for preserving optional hierarchy
func determineDestinationFolderName(componentFilePath string) string {
	componentTypeNames := paths.GetComponentTypeNames()

	// Get directory containing the component file
	currentDir := filepath.Dir(componentFilePath)

	// Walk up the directory tree
	for {
		dirName := filepath.Base(currentDir)

		// Check if current directory name is a component type
		isComponentType := false
		for _, typeName := range componentTypeNames {
			if dirName == typeName {
				isComponentType = true
				break
			}
		}

		// If not a component type name, use it
		if !isComponentType && dirName != "." && dirName != "" {
			return dirName
		}

		// Go up one directory
		parentDir := filepath.Dir(currentDir)

		// Check if we've reached the root
		if parentDir == currentDir || parentDir == "." || parentDir == "/" || dirName == "" {
			// Reached root, fall back to "root"
			return "root"
		}

		currentDir = parentDir
	}
}

func NewUpdateDetector() *UpdateDetector {
	return updater.NewUpdateDetector()
}

func NewBulkDownloader() *BulkDownloader {
	return downloader.NewBulkDownloader()
}

// NewComponentLinker creates a new ComponentLinker with dependencies injected
func NewComponentLinker() (*linker.ComponentLinker, error) {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}

	// Check if a profile is active and use its path instead
	profileManager, err := profiles.NewProfileManager(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile manager: %w", err)
	}

	activeProfile, err := profileManager.GetActiveProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to get active profile: %w", err)
	}

	// If a profile is active, use the profile's base path instead
	if activeProfile != "" {
		profilesDir, err := paths.GetProfilesDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get profiles directory: %w", err)
		}
		agentsDir = filepath.Join(profilesDir, activeProfile)
		fmt.Printf("Using active profile: %s\n", activeProfile)
	}

	// Detect all available targets
	targets, err := config.DetectAllTargets()
	if err != nil {
		return nil, fmt.Errorf("failed to detect targets: %w", err)
	}

	det := detector.NewRepositoryDetector()

	return linker.NewComponentLinker(agentsDir, targets, det)
}

// NewComponentLinkerWithFilter creates a new ComponentLinker with filtered targets
// targetFilter can be: "opencode", "claudecode", "all", or "" (defaults to all)
func NewComponentLinkerWithFilter(targetFilter string) (*linker.ComponentLinker, error) {
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents directory: %w", err)
	}

	// Check if a profile is active and use its path instead
	profileManager, err := profiles.NewProfileManager(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile manager: %w", err)
	}

	activeProfile, err := profileManager.GetActiveProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to get active profile: %w", err)
	}

	// If a profile is active, use the profile's base path instead
	if activeProfile != "" {
		profilesDir, err := paths.GetProfilesDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get profiles directory: %w", err)
		}
		agentsDir = filepath.Join(profilesDir, activeProfile)
		fmt.Printf("Using active profile: %s\n", activeProfile)
	}

	var targets []config.Target

	// Detect all targets first
	allTargets, err := config.DetectAllTargets()
	if err != nil {
		return nil, fmt.Errorf("failed to detect targets: %w", err)
	}

	// Filter targets based on targetFilter parameter
	if targetFilter == "" || targetFilter == "all" {
		// No filter or "all" - use all detected targets
		targets = allTargets
	} else {
		// Filter for specific target
		for _, target := range allTargets {
			if target.GetName() == targetFilter {
				targets = append(targets, target)
				break
			}
		}
		// If no matching target found, return error
		if len(targets) == 0 {
			return nil, fmt.Errorf("target '%s' not found. Available targets: %v", targetFilter, getTargetNames(allTargets))
		}
	}

	det := detector.NewRepositoryDetector()

	return linker.NewComponentLinker(agentsDir, targets, det)
}

// getTargetNames returns a slice of target names for error reporting
func getTargetNames(targets []config.Target) []string {
	names := make([]string, len(targets))
	for i, target := range targets {
		names[i] = target.GetName()
	}
	return names
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strings []string, separator string) string {
	if len(strings) == 0 {
		return ""
	}
	result := strings[0]
	for i := 1; i < len(strings); i++ {
		result += separator + strings[i]
	}
	return result
}

func main() {
	// Set up handlers for Cobra commands
	cmd.SetHandlers(
		func(repoURL, name, profile, targetDir string) {
			if profile != "" && targetDir != "" {
				log.Fatal("Cannot specify both --profile and --target-dir flags")
			}

			if targetDir != "" {
				// Install to custom target directory (isolated testing)
				resolvedPath, err := paths.ResolveTargetDir(targetDir)
				if err != nil {
					log.Fatal("Failed to resolve target directory:", err)
				}

				if err := fileutil.CreateDirectoryWithPermissions(resolvedPath); err != nil {
					log.Fatal("Failed to create target directory:", err)
				}

				fmt.Printf("Installing to custom directory: %s\n", resolvedPath)
				dl := downloader.NewSkillDownloaderWithTargetDir(resolvedPath)
				if err := dl.DownloadSkill(repoURL, name); err != nil {
					log.Fatal("Failed to download skill:", err)
				}
			} else if profile != "" {
				// Install directly to profile
				pm, err := profiles.NewProfileManager(nil)
				if err != nil {
					log.Fatal("Failed to create profile manager:", err)
				}

				// Validate profile exists by scanning
				profilesList, err := pm.ScanProfiles()
				if err != nil {
					log.Fatal("Failed to scan profiles:", err)
				}

				profileExists := false
				for _, p := range profilesList {
					if p.Name == profile {
						profileExists = true
						break
					}
				}

				if !profileExists {
					log.Fatalf("Profile '%s' does not exist. Create it first with: agent-smith profiles create %s", profile, profile)
				}

				dl := downloader.NewSkillDownloaderForProfile(profile)
				if err := dl.DownloadSkill(repoURL, name); err != nil {
					log.Fatal("Failed to download skill:", err)
				}
			} else {
				// Standard installation to ~/.agents/
				dl := downloader.NewSkillDownloader()
				if err := dl.DownloadSkill(repoURL, name); err != nil {
					log.Fatal("Failed to download skill:", err)
				}
			}
		},
		func(repoURL, name, profile, targetDir string) {
			if profile != "" && targetDir != "" {
				log.Fatal("Cannot specify both --profile and --target-dir flags")
			}

			if targetDir != "" {
				// Install to custom target directory (isolated testing)
				resolvedPath, err := paths.ResolveTargetDir(targetDir)
				if err != nil {
					log.Fatal("Failed to resolve target directory:", err)
				}

				if err := fileutil.CreateDirectoryWithPermissions(resolvedPath); err != nil {
					log.Fatal("Failed to create target directory:", err)
				}

				fmt.Printf("Installing to custom directory: %s\n", resolvedPath)
				dl := downloader.NewAgentDownloaderWithTargetDir(resolvedPath)
				if err := dl.DownloadAgent(repoURL, name); err != nil {
					log.Fatal("Failed to download agent:", err)
				}
			} else if profile != "" {
				// Install directly to profile
				pm, err := profiles.NewProfileManager(nil)
				if err != nil {
					log.Fatal("Failed to create profile manager:", err)
				}

				// Validate profile exists by scanning
				profilesList, err := pm.ScanProfiles()
				if err != nil {
					log.Fatal("Failed to scan profiles:", err)
				}

				profileExists := false
				for _, p := range profilesList {
					if p.Name == profile {
						profileExists = true
						break
					}
				}

				if !profileExists {
					log.Fatalf("Profile '%s' does not exist. Create it first with: agent-smith profiles create %s", profile, profile)
				}

				dl := downloader.NewAgentDownloaderForProfile(profile)
				if err := dl.DownloadAgent(repoURL, name); err != nil {
					log.Fatal("Failed to download agent:", err)
				}
			} else {
				// Standard installation to ~/.agents/
				dl := downloader.NewAgentDownloader()
				if err := dl.DownloadAgent(repoURL, name); err != nil {
					log.Fatal("Failed to download agent:", err)
				}
			}
		},
		func(repoURL, name, profile, targetDir string) {
			if profile != "" && targetDir != "" {
				log.Fatal("Cannot specify both --profile and --target-dir flags")
			}

			if targetDir != "" {
				// Install to custom target directory (isolated testing)
				resolvedPath, err := paths.ResolveTargetDir(targetDir)
				if err != nil {
					log.Fatal("Failed to resolve target directory:", err)
				}

				if err := fileutil.CreateDirectoryWithPermissions(resolvedPath); err != nil {
					log.Fatal("Failed to create target directory:", err)
				}

				fmt.Printf("Installing to custom directory: %s\n", resolvedPath)
				dl := downloader.NewCommandDownloaderWithTargetDir(resolvedPath)
				if err := dl.DownloadCommand(repoURL, name); err != nil {
					log.Fatal("Failed to download command:", err)
				}
			} else if profile != "" {
				// Install directly to profile
				pm, err := profiles.NewProfileManager(nil)
				if err != nil {
					log.Fatal("Failed to create profile manager:", err)
				}

				// Validate profile exists by scanning
				profilesList, err := pm.ScanProfiles()
				if err != nil {
					log.Fatal("Failed to scan profiles:", err)
				}

				profileExists := false
				for _, p := range profilesList {
					if p.Name == profile {
						profileExists = true
						break
					}
				}

				if !profileExists {
					log.Fatalf("Profile '%s' does not exist. Create it first with: agent-smith profiles create %s", profile, profile)
				}

				dl := downloader.NewCommandDownloaderForProfile(profile)
				if err := dl.DownloadCommand(repoURL, name); err != nil {
					log.Fatal("Failed to download command:", err)
				}
			} else {
				// Standard installation to ~/.agents/
				dl := downloader.NewCommandDownloader()
				if err := dl.DownloadCommand(repoURL, name); err != nil {
					log.Fatal("Failed to download command:", err)
				}
			}
		},
		func(repoURL string, targetDir string) {
			var bulkDownloader *downloader.BulkDownloader

			if targetDir != "" {
				// Resolve the target directory path
				resolvedPath, err := paths.ResolveTargetDir(targetDir)
				if err != nil {
					log.Fatal("Failed to resolve target directory:", err)
				}

				// Create the target directory if it doesn't exist
				if err := fileutil.CreateDirectoryWithPermissions(resolvedPath); err != nil {
					log.Fatal("Failed to create target directory:", err)
				}

				fmt.Printf("Installing to custom directory: %s\n", resolvedPath)
				bulkDownloader = downloader.NewBulkDownloaderWithTargetDir(resolvedPath)
			} else {
				// Use default behavior (install to ~/.agents/)
				bulkDownloader = downloader.NewBulkDownloader()
			}

			if err := bulkDownloader.AddAll(repoURL); err != nil {
				log.Fatal("Failed to bulk download components:", err)
			}
		},
		func(componentType, componentName string) {
			detector := NewUpdateDetector()

			// Load metadata to get source URL
			metadata, err := detector.LoadMetadata(componentType, componentName)
			if err != nil {
				log.Fatal("Failed to load component metadata:", err)
			}

			if err := detector.UpdateComponent(componentType, componentName, metadata.Source); err != nil {
				log.Fatal("Failed to update component:", err)
			}
		},
		func() {
			detector := NewUpdateDetector()
			if err := detector.UpdateAll(); err != nil {
				log.Fatal("Failed to update components:", err)
			}
		},
		func(componentType, componentName, targetFilter string) {
			linker, err := NewComponentLinkerWithFilter(targetFilter)
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}
			if err := linker.LinkComponent(componentType, componentName); err != nil {
				log.Fatal("Failed to link component:", err)
			}
		},
		func(targetFilter string) {
			linker, err := NewComponentLinkerWithFilter(targetFilter)
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}
			if err := linker.LinkAllComponents(); err != nil {
				log.Fatal("Failed to link all components:", err)
			}
		},
		func(componentType, targetFilter string) {
			linker, err := NewComponentLinkerWithFilter(targetFilter)
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}
			if err := linker.LinkComponentsByType(componentType); err != nil {
				log.Fatal("Failed to link components:", err)
			}
		},
		func() {
			linker, err := NewComponentLinker()
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}
			if err := linker.DetectAndLinkLocalRepositories(); err != nil {
				log.Fatal("Failed to auto-link repositories:", err)
			}
		},
		func() {
			linker, err := NewComponentLinker()
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}
			if err := linker.ListLinkedComponents(); err != nil {
				log.Fatal("Failed to list linked components:", err)
			}
		},
		func() {
			linker, err := NewComponentLinker()
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}
			if err := linker.ShowLinkStatus(); err != nil {
				log.Fatal("Failed to show link status:", err)
			}
		},
		func(componentType, componentName, targetFilter string) {
			linker, err := NewComponentLinkerWithFilter(targetFilter)
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}
			if err := linker.UnlinkComponent(componentType, componentName, targetFilter); err != nil {
				log.Fatal("Failed to unlink component:", err)
			}
		},
		func(targetFilter string, force bool) {
			linker, err := NewComponentLinkerWithFilter(targetFilter)
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}
			if err := linker.UnlinkAllComponents(targetFilter, force); err != nil {
				log.Fatal("Failed to unlink all components:", err)
			}
		},
		func(componentType, targetFilter string, force bool) {
			linker, err := NewComponentLinkerWithFilter(targetFilter)
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}
			if err := linker.UnlinkComponentsByType(componentType, targetFilter, force); err != nil {
				log.Fatal("Failed to unlink components:", err)
			}
		},
		func(componentType, componentName, profile string) {
			// Determine base directory
			baseDir, err := paths.GetAgentsDir()
			if err != nil {
				log.Fatal("Failed to get agents directory:", err)
			}

			if profile != "" {
				// Use profile directory
				profilesDir, err := paths.GetProfilesDir()
				if err != nil {
					log.Fatal("Failed to get profiles directory:", err)
				}
				baseDir = filepath.Join(profilesDir, profile)

				// Validate profile exists
				if _, err := os.Stat(baseDir); os.IsNotExist(err) {
					log.Fatalf("Profile '%s' does not exist", profile)
				}
			}

			// Create linker for unlinking
			linker, err := NewComponentLinker()
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}

			// Create uninstaller
			uninstaller := uninstaller.NewUninstaller(baseDir, linker)

			// Uninstall component
			if err := uninstaller.UninstallComponent(componentType, componentName); err != nil {
				log.Fatal("Failed to uninstall component:", err)
			}
		},
		func(repoURL string, force bool) {
			// Get base directory (always ~/.agents/ for bulk uninstall)
			baseDir, err := paths.GetAgentsDir()
			if err != nil {
				log.Fatal("Failed to get base directory:", err)
			}

			// Create linker for unlinking
			linker, err := NewComponentLinker()
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}

			// Create uninstaller
			uninstaller := uninstaller.NewUninstaller(baseDir, linker)

			// Uninstall all components from source
			if err := uninstaller.UninstallAllFromSource(repoURL, force); err != nil {
				log.Fatal("Failed to uninstall components:", err)
			}
		},
		func() {
			pm, err := profiles.NewProfileManager(nil)
			if err != nil {
				log.Fatal("Failed to create profile manager:", err)
			}

			profilesList, err := pm.ScanProfiles()
			if err != nil {
				log.Fatal("Failed to scan profiles:", err)
			}

			// Get active profile
			activeProfile, err := pm.GetActiveProfile()
			if err != nil {
				log.Fatal("Failed to get active profile:", err)
			}

			// Display results
			if len(profilesList) == 0 {
				fmt.Println("No profiles found in ~/.agents/profiles/")
				fmt.Println("\nTo create a profile, run:")
				fmt.Println("  ./agent-smith profile create <profile-name>")
				return
			}

			fmt.Println("Available Profiles:")
			fmt.Println()

			for _, profile := range profilesList {
				// Count components
				agents, skills, commands := pm.CountComponents(profile)

				// Build component counts string
				var components []string
				if agents > 0 {
					components = append(components, fmt.Sprintf("%d agent(s)", agents))
				}
				if skills > 0 {
					components = append(components, fmt.Sprintf("%d skill(s)", skills))
				}
				if commands > 0 {
					components = append(components, fmt.Sprintf("%d command(s)", commands))
				}

				componentStr := ""
				if len(components) > 0 {
					componentStr = fmt.Sprintf(" (%s)", joinStrings(components, ", "))
				}

				// Check if this is the active profile
				activeIndicator := "  "
				activeLabel := ""
				if profile.Name == activeProfile {
					activeIndicator = fmt.Sprintf("%s ", formatter.SymbolSuccess)
					activeLabel = " [active]"
				}

				fmt.Printf("%s%-15s%s%s\n", activeIndicator, profile.Name, activeLabel, componentStr)
			}

			// Display legend
			fmt.Println("\nLegend:")
			fmt.Printf("  %s - Currently active profile\n", formatter.SymbolSuccess)

			// Display total count
			fmt.Printf("\nTotal: %d profile(s)\n", len(profilesList))
		},
		func(profileName string) {
			pm, err := profiles.NewProfileManager(nil)
			if err != nil {
				log.Fatal("Failed to create profile manager:", err)
			}

			// Load the profile
			profilesList, err := pm.ScanProfiles()
			if err != nil {
				log.Fatal("Failed to scan profiles:", err)
			}

			var targetProfile *profiles.Profile
			for _, p := range profilesList {
				if p.Name == profileName {
					targetProfile = p
					break
				}
			}

			if targetProfile == nil {
				log.Fatalf("Profile '%s' not found", profileName)
			}

			// Get active profile to show status
			activeProfile, err := pm.GetActiveProfile()
			if err != nil {
				log.Fatal("Failed to get active profile:", err)
			}

			// Display profile information
			fmt.Printf("Profile: %s", targetProfile.Name)
			if targetProfile.Name == activeProfile {
				fmt.Printf(" %s [active]", formatter.SymbolSuccess)
			}
			fmt.Println()
			fmt.Printf("Location: %s\n", targetProfile.BasePath)
			fmt.Println()

			// Get component names
			agents, skills, commands := pm.GetComponentNames(targetProfile)

			// Display agents
			if len(agents) > 0 {
				fmt.Printf("Agents (%d):\n", len(agents))
				for _, agent := range agents {
					fmt.Printf("  - %s\n", agent)
				}
				fmt.Println()
			}

			// Display skills
			if len(skills) > 0 {
				fmt.Printf("Skills (%d):\n", len(skills))
				for _, skill := range skills {
					fmt.Printf("  - %s\n", skill)
				}
				fmt.Println()
			}

			// Display commands
			if len(commands) > 0 {
				fmt.Printf("Commands (%d):\n", len(commands))
				for _, command := range commands {
					fmt.Printf("  - %s\n", command)
				}
				fmt.Println()
			}

			// Show empty state if no components
			if len(agents) == 0 && len(skills) == 0 && len(commands) == 0 {
				fmt.Println("This profile is empty.")
				fmt.Println("\nAdd components with:")
				fmt.Printf("  agent-smith profiles add <type> %s <component-name>\n", profileName)
			} else if targetProfile.Name != activeProfile {
				// Show activation hint if not active
				fmt.Println("To activate this profile:")
				fmt.Printf("  agent-smith profiles activate %s\n", profileName)
			}
		},
		func(profileName string) {
			pm, err := profiles.NewProfileManager(nil)
			if err != nil {
				log.Fatal("Failed to create profile manager:", err)
			}

			if err := pm.CreateProfile(profileName); err != nil {
				log.Fatal("Failed to create profile:", err)
			}
		},
		func(profileName string) {
			// Create component linker for defensive unlinking
			componentLinker, err := NewComponentLinker()
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}

			pm, err := profiles.NewProfileManager(componentLinker)
			if err != nil {
				log.Fatal("Failed to create profile manager:", err)
			}

			if err := pm.DeleteProfile(profileName); err != nil {
				log.Fatal("Failed to delete profile:", err)
			}
		},
		func(profileName string) {
			pm, err := profiles.NewProfileManager(nil)
			if err != nil {
				log.Fatal("Failed to create profile manager:", err)
			}

			if err := pm.ActivateProfile(profileName); err != nil {
				log.Fatal("Failed to activate profile:", err)
			}
		},
		func() {
			pm, err := profiles.NewProfileManager(nil)
			if err != nil {
				log.Fatal("Failed to create profile manager:", err)
			}

			if err := pm.DeactivateProfile(); err != nil {
				log.Fatal("Failed to deactivate profile:", err)
			}
		},
		func(componentType, profileName, componentName string) {
			pm, err := profiles.NewProfileManager(nil)
			if err != nil {
				log.Fatal("Failed to create profile manager:", err)
			}

			if err := pm.AddComponentToProfile(profileName, componentType, componentName); err != nil {
				log.Fatal("Failed to add component to profile:", err)
			}
		},
		func(componentType, profileName, componentName string) {
			// Create component linker to handle auto-unlinking
			componentLinker, err := NewComponentLinker()
			if err != nil {
				log.Fatal("Failed to create component linker:", err)
			}

			pm, err := profiles.NewProfileManager(componentLinker)
			if err != nil {
				log.Fatal("Failed to create profile manager:", err)
			}

			if err := pm.RemoveComponentFromProfile(profileName, componentType, componentName); err != nil {
				log.Fatal("Failed to remove component from profile:", err)
			}
		},
		func() {
			// Status handler - shows current system status
			pm, err := profiles.NewProfileManager(nil)
			if err != nil {
				log.Fatal("Failed to create profile manager:", err)
			}

			// Get active profile
			activeProfile, err := pm.GetActiveProfile()
			if err != nil {
				log.Fatal("Failed to get active profile:", err)
			}

			// Detect all available targets
			targets, err := config.DetectAllTargets()
			if err != nil {
				log.Fatal("Failed to detect targets:", err)
			}

			// Get agents directory
			agentsDir, err := paths.GetAgentsDir()
			if err != nil {
				log.Fatal("Failed to get agents directory:", err)
			}

			// Count components in ~/.agents/
			agentsPath := filepath.Join(agentsDir, "agents")
			skillsPath := filepath.Join(agentsDir, "skills")
			commandsPath := filepath.Join(agentsDir, "commands")

			agentsCount := 0
			skillsCount := 0
			commandsCount := 0

			if entries, err := os.ReadDir(agentsPath); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						agentsCount++
					}
				}
			}

			if entries, err := os.ReadDir(skillsPath); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						skillsCount++
					}
				}
			}

			if entries, err := os.ReadDir(commandsPath); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						commandsCount++
					}
				}
			}

			// Display status
			fmt.Println("Current Configuration:")
			fmt.Println()

			// Show active profile
			if activeProfile != "" {
				fmt.Printf("  Active Profile: %s %s\n", activeProfile, formatter.SymbolSuccess)
			} else {
				fmt.Println("  Active Profile: None")
			}

			// Show detected targets
			if len(targets) > 0 {
				var targetNames []string
				for _, target := range targets {
					targetNames = append(targetNames, target.GetName())
				}
				fmt.Printf("  Detected Targets: %s\n", joinStrings(targetNames, ", "))
			} else {
				fmt.Println("  Detected Targets: None")
			}

			fmt.Println()
			fmt.Println("Components in ~/.agents/:")
			fmt.Printf("  Agents: %d\n", agentsCount)
			fmt.Printf("  Skills: %d\n", skillsCount)
			fmt.Printf("  Commands: %d\n", commandsCount)

			// If there's an active profile, show its components
			if activeProfile != "" {
				profilesList, err := pm.ScanProfiles()
				if err == nil {
					for _, profile := range profilesList {
						if profile.Name == activeProfile {
							agents, skills, commands := pm.CountComponents(profile)
							fmt.Println()
							fmt.Printf("Active Profile (%s):\n", activeProfile)
							fmt.Printf("  Agents: %d\n", agents)
							fmt.Printf("  Skills: %d\n", skills)
							fmt.Printf("  Commands: %d\n", commands)
							break
						}
					}
				}
			}

			fmt.Println()
			fmt.Println("For more details:")
			fmt.Println("  - Run 'agent-smith link status' for link information")
			fmt.Println("  - Run 'agent-smith profiles list' to see all profiles")
		},
		func(name, path string) {
			// Load existing config
			cfg, err := config.LoadConfig()
			if err != nil {
				log.Fatal("Failed to load config:", err)
			}

			// Validate that target name doesn't already exist
			for _, target := range cfg.CustomTargets {
				if target.Name == name {
					log.Fatalf("Target '%s' already exists in config", name)
				}
			}

			// Create new custom target config
			newTarget := config.CustomTargetConfig{
				Name:        name,
				BaseDir:     path,
				SkillsDir:   "skills",
				AgentsDir:   "agents",
				CommandsDir: "commands",
			}

			// Add to config
			cfg.CustomTargets = append(cfg.CustomTargets, newTarget)

			// Save config
			if err := config.SaveConfig(cfg); err != nil {
				log.Fatal("Failed to save config:", err)
			}

			fmt.Printf("%s Successfully added custom target '%s'\n", formatter.SymbolSuccess, name)
			fmt.Printf("  Base directory: %s\n", path)
			fmt.Println("\nSubdirectories:")
			fmt.Printf("  Skills:   %s/skills\n", path)
			fmt.Printf("  Agents:   %s/agents\n", path)
			fmt.Printf("  Commands: %s/commands\n", path)
			fmt.Println("\nYou can now link components to this target:")
			fmt.Printf("  agent-smith link all --target %s\n", name)
		},
		func(name string) {
			// Load existing config
			cfg, err := config.LoadConfig()
			if err != nil {
				log.Fatal("Failed to load config:", err)
			}

			// Check if target exists and is a custom target
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
				log.Fatalf("Target '%s' not found in custom targets", name)
			}

			// Remove the target from the slice
			cfg.CustomTargets = append(cfg.CustomTargets[:targetIndex], cfg.CustomTargets[targetIndex+1:]...)

			// Save config
			if err := config.SaveConfig(cfg); err != nil {
				log.Fatal("Failed to save config:", err)
			}

			fmt.Printf("%s Successfully removed custom target '%s'\n", formatter.SymbolSuccess, name)
			fmt.Println("\nNote: This only removes the target from configuration.")
			fmt.Println("Components linked to this target are not automatically unlinked.")
		},
		func() {
			// Load config to distinguish between built-in and custom targets
			cfg, err := config.LoadConfig()
			if err != nil {
				log.Fatal("Failed to load config:", err)
			}

			// Get all built-in targets (even if not detected)
			builtInNames := []string{"opencode", "claudecode"}

			fmt.Println("Available Targets:")
			fmt.Println()

			// Display built-in targets
			fmt.Println("Built-in Targets:")
			for _, name := range builtInNames {
				var target config.Target
				var err error

				if name == "opencode" {
					target, err = config.NewOpencodeTarget()
				} else if name == "claudecode" {
					target, err = config.NewClaudeCodeTarget()
				}

				if err != nil {
					continue
				}

				baseDir, _ := target.GetBaseDir()
				exists := false
				if _, err := os.Stat(baseDir); err == nil {
					exists = true
				}

				symbol := formatter.SymbolNotLinked
				status := "[not found]"
				if exists {
					symbol = formatter.SymbolSuccess
					status = "[detected]"
				}

				fmt.Printf("  %s %-15s %-30s %s\n", symbol, name, baseDir, status)
			}

			// Display custom targets
			if len(cfg.CustomTargets) > 0 {
				fmt.Println()
				fmt.Println("Custom Targets:")
				for _, customTargetConfig := range cfg.CustomTargets {
					customTarget, err := config.NewCustomTarget(customTargetConfig)
					if err != nil {
						fmt.Printf("  %s %-15s <error loading target>\n", formatter.SymbolError, customTargetConfig.Name)
						continue
					}

					baseDir, _ := customTarget.GetBaseDir()
					exists := false
					if _, err := os.Stat(baseDir); err == nil {
						exists = true
					}

					symbol := formatter.SymbolNotLinked
					status := "[not found]"
					if exists {
						symbol = formatter.SymbolSuccess
						status = "[detected]"
					}

					fmt.Printf("  %s %-15s %-30s %s\n", symbol, customTargetConfig.Name, baseDir, status)
				}
			}

			// Display legend
			fmt.Println()
			fmt.Println("Legend:")
			fmt.Printf("  %s - Target directory exists\n", formatter.SymbolSuccess)
			fmt.Printf("  %s - Target directory not found\n", formatter.SymbolNotLinked)

			// Count available targets
			availableCount := 0
			totalCount := len(builtInNames) + len(cfg.CustomTargets)

			for _, name := range builtInNames {
				var target config.Target
				var err error

				if name == "opencode" {
					target, err = config.NewOpencodeTarget()
				} else if name == "claudecode" {
					target, err = config.NewClaudeCodeTarget()
				}

				if err == nil {
					baseDir, _ := target.GetBaseDir()
					if _, err := os.Stat(baseDir); err == nil {
						availableCount++
					}
				}
			}

			for _, customTargetConfig := range cfg.CustomTargets {
				customTarget, err := config.NewCustomTarget(customTargetConfig)
				if err == nil {
					baseDir, _ := customTarget.GetBaseDir()
					if _, err := os.Stat(baseDir); err == nil {
						availableCount++
					}
				}
			}

			fmt.Println()
			fmt.Printf("Total: %d target(s) (%d available)\n", totalCount, availableCount)
		},
	)

	// Execute Cobra command
	cmd.Execute()
}
