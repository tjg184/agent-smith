package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjg184/agent-smith/internal/fileutil"
	"github.com/tjg184/agent-smith/internal/metadata"
	"github.com/tjg184/agent-smith/internal/models"
)

// DetectComponentForPattern checks if a file matches a component detection pattern
func (rd *RepositoryDetector) DetectComponentForPattern(fileName, relPath, fullRelPath, repoPath string, pattern models.ComponentDetectionPattern, componentType models.ComponentType) (string, string, bool) {
	// Debug logging for component detection process
	if rd.logger != nil {
		rd.logger.Debug("Processing file: %s, relPath: %s, fileName: %s", fullRelPath, relPath, fileName)
		rd.logger.Debug("Component pattern: %s, exactFiles: %v", pattern.Name, pattern.ExactFiles)
	}

	// Check if path should be ignored
	if rd.ShouldIgnorePath(relPath, pattern.IgnorePaths) {
		if rd.logger != nil {
			rd.logger.Debug("Path ignored: %s", relPath)
		}
		return "", "", false
	}

	var frontmatter *models.ComponentFrontmatter
	if strings.HasSuffix(fileName, ".md") {
		fullFilePath := filepath.Join(repoPath, fullRelPath)
		parsedFrontmatter, err := fileutil.ParseFrontmatter(fullFilePath)
		if err != nil {
			if rd.logger != nil {
				rd.logger.Debug("Failed to parse frontmatter from %s: %v", fullFilePath, err)
			}
		} else if parsedFrontmatter != nil {
			frontmatter = parsedFrontmatter
			if rd.logger != nil {
				rd.logger.Debug("Parsed frontmatter from %s: name=%s", fullFilePath, frontmatter.Name)
			}
		}
	}

	if rd.MatchesExactFile(fileName, pattern.ExactFiles) {
		// Use fullRelPath to get the correct directory containing the component file
		componentDir := filepath.Dir(fullRelPath)
		if rd.logger != nil {
			rd.logger.Debug("Exact file match, componentDir: %s", componentDir)
		}

		if componentDir == "." {
			componentName := "root-" + pattern.Name
			if rd.logger != nil {
				rd.logger.Debug("Root component, name: %s", componentName)
			}
			return componentName, componentDir, true
		}

		// For exact file matches, use frontmatter name if available, otherwise use directory name
		var componentName string
		if frontmatter != nil && strings.TrimSpace(frontmatter.Name) != "" {
			componentName = strings.TrimSpace(frontmatter.Name)
		} else {
			componentName = filepath.Base(componentDir)
		}

		if rd.logger != nil {
			rd.logger.Debug("Extracted component name: %s from directory: %s (frontmatter: %v)", componentName, componentDir, frontmatter != nil)
			rd.logger.Debug("Component name: '%s', componentKey: '%s-%s'", componentName, pattern.Name, componentName)
		}
		return componentName, componentDir, true
	}

	if len(pattern.PathPatterns) > 0 && len(pattern.FileExtensions) > 0 {
		if rd.MatchesPathPattern(relPath, pattern.PathPatterns) && rd.MatchesFileExtension(fileName, pattern.FileExtensions) {
			// Use determineComponentName with frontmatter priority
			componentName := fileutil.DetermineComponentName(frontmatter, fileName)

			// Skip if determineComponentName returns empty (special files like README.md)
			if componentName == "" {
				if rd.logger != nil {
					rd.logger.Debug("Path pattern + extension match, but component name is empty (special file), skipping")
				}
				return "", "", false
			}

			if rd.logger != nil {
				rd.logger.Debug("Path pattern + extension match, name: %s (frontmatter: %v)", componentName, frontmatter != nil)
			}
			return componentName, fullRelPath, true
		}
		if rd.logger != nil {
			rd.logger.Debug("Path pattern + extension check failed")
		}
	}

	if len(pattern.PathPatterns) > 0 && rd.MatchesPathPattern(relPath, pattern.PathPatterns) {
		// Use determineComponentName with frontmatter priority
		componentName := fileutil.DetermineComponentName(frontmatter, fileName)

		// Skip if determineComponentName returns empty (special files like README.md)
		if componentName == "" {
			if rd.logger != nil {
				rd.logger.Debug("Path pattern match, but component name is empty (special file), skipping")
			}
			return "", "", false
		}

		if rd.logger != nil {
			rd.logger.Debug("Path pattern match, name: %s (frontmatter: %v)", componentName, frontmatter != nil)
		}
		return componentName, fullRelPath, true
	}
	if rd.logger != nil {
		rd.logger.Debug("Path pattern check failed")
	}

	if rd.logger != nil {
		rd.logger.Debug("No pattern matched for file: %s", fileName)
	}
	return "", "", false
}

// DetectComponentsInRepo detects all components in a repository
func (rd *RepositoryDetector) DetectComponentsInRepo(repoPath string) ([]models.DetectedComponent, error) {
	var components []models.DetectedComponent

	// Track all component occurrences for duplicate detection
	type ComponentOccurrence struct {
		component models.DetectedComponent
		path      string
		hash      string // Content hash for detecting identical duplicates
	}
	seenComponents := make(map[string][]ComponentOccurrence)
	hasConflicts := false

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fileName := filepath.Base(path)
		parentDir := filepath.Dir(path)
		relPath, err := filepath.Rel(repoPath, parentDir)
		if err != nil {
			return err
		}

		// Full relative path including filename for path-based detection
		fullRelPath, err := filepath.Rel(repoPath, path)
		if err != nil {
			return err
		}

		// Check each component type using its detection pattern
		for componentTypeStr, pattern := range rd.detectionConfig.Components {
			componentType := models.ComponentType(componentTypeStr)

			if componentName, componentPath, matched := rd.DetectComponentForPattern(fileName, relPath, fullRelPath, repoPath, pattern, componentType); matched {
				if rd.logger != nil {
					rd.logger.Debug("Match result: true for componentType: %s", componentTypeStr)
				}

				// Handle default component names
				if componentName == "" || componentName == "." {
					componentName = fmt.Sprintf("root-%s", pattern.Name)
					if rd.logger != nil {
						rd.logger.Debug("Applied default component name: %s", componentName)
					}
				}

				componentKey := fmt.Sprintf("%s-%s", pattern.Name, componentName)
				if rd.logger != nil {
					rd.logger.Debug("Component key: %s", componentKey)
				}

				if existing, exists := seenComponents[componentKey]; exists {
					// Duplicate detected - compute hashes to compare content

					// Compute hash for the new duplicate
					duplicateHash, err := metadata.ComputeComponentHash(repoPath, componentPath)
					if err != nil {
						if rd.logger != nil {
							rd.logger.Debug("Failed to compute hash for duplicate component %s: %v", componentPath, err)
						}
						duplicateHash = ""
					}

					// Get hash of first occurrence (compute if not already cached)
					if existing[0].hash == "" {
						firstHash, err := metadata.ComputeComponentHash(repoPath, existing[0].component.Path)
						if err != nil {
							if rd.logger != nil {
								rd.logger.Debug("Failed to compute hash for first occurrence %s: %v", existing[0].component.Path, err)
							}
							firstHash = ""
						}
						existing[0].hash = firstHash
						seenComponents[componentKey][0] = existing[0]
					}

					// Compare hashes to determine if this is a real conflict
					// Only treat as different if both hashes computed successfully and don't match
					contentDiffers := false
					if duplicateHash != "" && existing[0].hash != "" {
						// Both hashes computed successfully - compare them
						contentDiffers = (duplicateHash != existing[0].hash)
					} else {
						// Hash computation failed for at least one - treat conservatively as different
						contentDiffers = true
					}

					if contentDiffers {
						hasConflicts = true
						// Only log warning for actual conflicts (different content)
						if rd.logger != nil {
							rd.logger.Warn("⚠️  WARNING: Duplicate component name detected!")
							rd.logger.Warn("    Component: %s (%s)", componentName, pattern.Name)
							rd.logger.Warn("    First occurrence: %s", existing[0].path)
							rd.logger.Warn("    Duplicate at: %s (WILL BE SKIPPED)", fullRelPath)
							rd.logger.Warn("    Content differs - this may cause conflicts")
						}
					} else {
						// Identical content - suppress warning (log only in debug mode)
						if rd.logger != nil {
							rd.logger.Debug("Duplicate component '%s' has identical content, suppressing warning", componentName)
							rd.logger.Debug("    First occurrence: %s", existing[0].path)
							rd.logger.Debug("    Duplicate at: %s (WILL BE SKIPPED)", fullRelPath)
						}
					}

					// Track this duplicate occurrence
					seenComponents[componentKey] = append(seenComponents[componentKey], ComponentOccurrence{
						component: models.DetectedComponent{
							Type:       componentType,
							Name:       componentName,
							Path:       componentPath,
							SourceFile: fileName,
							FilePath:   fullRelPath, // Track full path from repo root
						},
						path: fullRelPath,
						hash: duplicateHash,
					})
				} else {
					// First occurrence - add to components list
					component := models.DetectedComponent{
						Type:       componentType,
						Name:       componentName,
						Path:       componentPath,
						SourceFile: fileName,
						FilePath:   fullRelPath, // Track full path from repo root
					}
					components = append(components, component)
					seenComponents[componentKey] = []ComponentOccurrence{{
						component: component,
						path:      fullRelPath,
						hash:      "", // Hash computed lazily when needed
					}}
					if rd.logger != nil {
						rd.logger.Debug("Added component: %s (key: %s)", componentName, componentKey)
					}
				}
			}
		}

		return nil
	})

	if rd.logger != nil {
		rd.logger.Debug("Total components detected: %d", len(components))
	}

	// Count components by type for debugging
	skillCount := 0
	agentCount := 0
	commandCount := 0
	for _, comp := range components {
		switch comp.Type {
		case models.ComponentSkill:
			skillCount++
		case models.ComponentAgent:
			agentCount++
		case models.ComponentCommand:
			commandCount++
		}
	}
	if rd.logger != nil {
		rd.logger.Debug("Component breakdown - Skills: %d, Agents: %d, Commands: %d", skillCount, agentCount, commandCount)
	}

	// Display duplicate warnings summary only if there are actual conflicts
	// (duplicates with different content) and warnings are not suppressed
	if hasConflicts && !rd.suppressDuplicateWarning {
		fmt.Printf("\n")
		fmt.Printf("╔════════════════════════════════════════════════════════════════════╗\n")
		fmt.Printf("║  ⚠️  WARNING: Duplicate Component Names Detected                  ║\n")
		fmt.Printf("╚════════════════════════════════════════════════════════════════════╝\n\n")

		duplicateCount := 0
		for _, occurrences := range seenComponents {
			if len(occurrences) > 1 {
				// Check if any occurrence has different content
				hasConflict := false
				firstHash := occurrences[0].hash
				if firstHash == "" {
					// Compute first hash if not already done
					var err error
					firstHash, err = metadata.ComputeComponentHash(repoPath, occurrences[0].component.Path)
					if err != nil {
						hasConflict = true // Treat hash failure as potential conflict
					}
				}

				for i := 1; i < len(occurrences); i++ {
					hash := occurrences[i].hash
					if hash == "" {
						// Compute hash if not already done
						var err error
						hash, err = metadata.ComputeComponentHash(repoPath, occurrences[i].component.Path)
						if err != nil {
							hasConflict = true
							break
						}
					}
					if hash != firstHash {
						hasConflict = true
						break
					}
				}

				// Only show duplicates with different content
				if hasConflict {
					duplicateCount++
					// Parse component type from key
					componentType := "component"
					if len(occurrences) > 0 {
						componentType = string(occurrences[0].component.Type)
					}

					fmt.Printf("  [%d] %s '%s' found in %d locations:\n", duplicateCount, componentType, occurrences[0].component.Name, len(occurrences))
					for i, occ := range occurrences {
						if i == 0 {
							fmt.Printf("      ✓ %s (USED - first occurrence)\n", occ.path)
						} else {
							fmt.Printf("      ✗ %s (SKIPPED - duplicate #%d)\n", occ.path, i)
						}
					}
					fmt.Printf("\n")
				}
			}
		}

		if duplicateCount > 0 {
			fmt.Printf("  Resolution Required:\n")
			fmt.Printf("  • Only the FIRST occurrence of each component will be used\n")
			fmt.Printf("  • Subsequent duplicates have been SKIPPED\n")
			fmt.Printf("  • To resolve: Rename or remove duplicate components\n")
			fmt.Printf("\n")
			fmt.Printf("  Total conflicts found: %d\n", duplicateCount)
			fmt.Printf("════════════════════════════════════════════════════════════════════\n\n")
		}
	}

	return components, err
}
