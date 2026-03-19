package copy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjg184/agent-smith/pkg/services"
)

// CopyComponentBetweenProfiles copies a component from sourceBaseDir to targetBaseDir,
// including its lock file entry.
func CopyComponentBetweenProfiles(
	sourceBaseDir, targetBaseDir, componentType, componentName string,
	lockService services.ComponentLockService,
) error {
	if err := copyComponentWithMetadata(sourceBaseDir, targetBaseDir, componentType, componentName, lockService); err != nil {
		return err
	}

	var sourceURL string
	sourceLockPath := filepath.Join(sourceBaseDir, fmt.Sprintf(".%s-lock.json", componentType[:len(componentType)-1]))
	if lockData, err := os.ReadFile(sourceLockPath); err == nil {
		var lockFile struct {
			Skills map[string]struct {
				SourceUrl string `json:"sourceUrl"`
			} `json:"skills,omitempty"`
			Agents map[string]struct {
				SourceUrl string `json:"sourceUrl"`
			} `json:"agents,omitempty"`
			Commands map[string]struct {
				SourceUrl string `json:"sourceUrl"`
			} `json:"commands,omitempty"`
		}
		if json.Unmarshal(lockData, &lockFile) == nil {
			switch componentType {
			case "skills":
				if entry, ok := lockFile.Skills[componentName]; ok {
					sourceURL = entry.SourceUrl
				}
			case "agents":
				if entry, ok := lockFile.Agents[componentName]; ok {
					sourceURL = entry.SourceUrl
				}
			case "commands":
				if entry, ok := lockFile.Commands[componentName]; ok {
					sourceURL = entry.SourceUrl
				}
			}
		}
	}

	componentSingular := strings.TrimSuffix(componentType, "s")
	fmt.Printf("\n✓ Successfully copied %s '%s'\n", componentSingular, componentName)
	if sourceURL != "" {
		fmt.Printf("  Source: %s\n", sourceURL)
	}
	fmt.Printf("  Location: %s\n", filepath.Join(targetBaseDir, componentType, componentName))
	fmt.Printf("\nBoth profiles can now update this component independently.\n")

	return nil
}

// AddComponentToProfile copies a component from agentsDir to profileBaseDir.
func AddComponentToProfile(agentsDir, profileBaseDir, componentType, componentName string, lockService services.ComponentLockService) error {
	srcDir := filepath.Join(agentsDir, componentType, componentName)

	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("component '%s' not found in ~/.agent-smith/%s/", componentName, componentType)
	}

	info, err := os.Lstat(srcDir)
	if err != nil {
		return fmt.Errorf("failed to stat component: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("cannot add component '%s': it is a symlink from an active profile. Deactivate the profile first.", componentName)
	}

	return copyComponentWithMetadata(agentsDir, profileBaseDir, componentType, componentName, lockService)
}

// RemoveComponentFromProfile removes a component directory from profileBaseDir.
// unlinkFn is called before removal when the profile is active (may be nil).
func RemoveComponentFromProfile(
	profileBaseDir, componentType, componentName string,
	unlinkFn func() error,
) error {
	componentPath := filepath.Join(profileBaseDir, componentType, componentName)

	if _, err := os.Stat(componentPath); os.IsNotExist(err) {
		return fmt.Errorf("component '%s' not found in profile", componentName)
	}

	if unlinkFn != nil {
		if err := unlinkFn(); err != nil {
			fmt.Printf("Warning: failed to unlink before removal: %v\n", err)
		}
	}

	if err := os.RemoveAll(componentPath); err != nil {
		return fmt.Errorf("failed to remove component: %w", err)
	}

	return nil
}

// CopyDirectory recursively copies a source directory to dst.
func CopyDirectory(src, dst string) error {
	return copyDirectory(src, dst)
}

func copyComponentWithMetadata(
	sourceBaseDir, targetBaseDir, componentType, componentName string,
	lockService services.ComponentLockService,
) error {
	srcDir := filepath.Join(sourceBaseDir, componentType, componentName)
	dstDir := filepath.Join(targetBaseDir, componentType, componentName)

	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("component '%s' does not exist in source profile (expected at: %s)", componentName, srcDir)
	} else if err != nil {
		return fmt.Errorf("failed to access source component '%s' at %s: %w", componentName, srcDir, err)
	}

	if _, err := os.Stat(dstDir); err == nil {
		return fmt.Errorf("component '%s' already exists in target profile at: %s\n\nTo overwrite, first remove the existing component:\n  agent-smith remove %s %s", componentName, dstDir, componentType, componentName)
	}

	fmt.Printf("Copying component files...\n")

	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if entry.IsDir() {
			if err := copyDirectory(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy directory %s: %w", entry.Name(), err)
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", entry.Name(), err)
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", entry.Name(), err)
			}
		}
	}

	fmt.Printf("Copying metadata...\n")

	entry, err := lockService.LoadEntry(sourceBaseDir, componentType, componentName)
	if err != nil {
		fmt.Printf("Note: No lock file entry found in source (manual component)\n")
		return nil
	}

	fmt.Printf("Found metadata for component from source: %s\n", entry.SourceUrl)

	if err := lockService.SaveEntry(targetBaseDir, componentType, componentName, entry); err != nil {
		fmt.Printf("Warning: Failed to save metadata to target: %v\n", err)
		return nil
	}

	fmt.Printf("✓ Metadata copied successfully\n")
	return nil
}

func copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("failed to determine relative path for %s: %w", path, err)
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dstPath, err)
			}
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("permission denied reading file %s: %w", path, err)
			}
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("permission denied writing file %s: %w", dstPath, err)
			}
			return fmt.Errorf("failed to write file %s: %w", dstPath, err)
		}

		return nil
	})
}
