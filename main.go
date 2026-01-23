package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

const (
	agentsDir    = "~/.agents"
	skillsDir    = agentsDir + "/skills"
	agentsSubDir = agentsDir + "/agents"
	commandsDir  = agentsDir + "/commands"
	opencodeDir  = "~/.config/opencode"
)

type RepositoryDetector struct {
	patterns map[string]string
}

type ComponentType string

const (
	ComponentSkill   ComponentType = "skill"
	ComponentAgent   ComponentType = "agent"
	ComponentCommand ComponentType = "command"
)

type DetectedComponent struct {
	Type       ComponentType
	Name       string
	Path       string
	SourceFile string
}

type SkillDownloader struct {
	baseDir  string
	detector *RepositoryDetector
}

type AgentDownloader struct {
	baseDir  string
	detector *RepositoryDetector
}

type CommandDownloader struct {
	baseDir  string
	detector *RepositoryDetector
}

type ComponentLinker struct {
	agentsDir   string
	opencodeDir string
}

type UpdateDetector struct {
	baseDir  string
	detector *RepositoryDetector
}

type ComponentMetadata struct {
	Name       string `json:"name"`
	Source     string `json:"source"`
	Commit     string `json:"commit"`
	Downloaded string `json:"downloaded"`
	Components int    `json:"components,omitempty"`
	Detection  string `json:"detection,omitempty"`
}

func NewRepositoryDetector() *RepositoryDetector {
	return &RepositoryDetector{
		patterns: map[string]string{
			"github":    `^https?://(?:www\.)?github\.com/[^/]+/[^/]+$`,
			"gitlab":    `^https?://(?:www\.)?gitlab\.com/[^/]+/[^/]+$`,
			"bitbucket": `^https?://(?:www\.)?bitbucket\.org/[^/]+/[^/]+$`,
			"git":       `^(https?://|git@|ssh://).+\.git$`,
		},
	}
}

func (rd *RepositoryDetector) detectProvider(repoURL string) string {
	repoURL = strings.TrimSpace(repoURL)

	for provider, pattern := range rd.patterns {
		if matched, _ := regexp.MatchString(pattern, repoURL); matched {
			return provider
		}
	}

	return "generic"
}

func (rd *RepositoryDetector) normalizeURL(repoURL string) (string, error) {
	repoURL = strings.TrimSpace(repoURL)

	// If it's already a full URL, return as-is
	if strings.HasPrefix(repoURL, "http://") || strings.HasPrefix(repoURL, "https://") ||
		strings.HasPrefix(repoURL, "git@") || strings.HasPrefix(repoURL, "ssh://") {
		return repoURL, nil
	}

	// Handle GitHub shorthand (owner/repo)
	if !strings.Contains(repoURL, "/") {
		return "", fmt.Errorf("invalid repository format: %s", repoURL)
	}

	parts := strings.Split(repoURL, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid repository format: %s", repoURL)
	}

	// Default to GitHub for shorthand notation
	return fmt.Sprintf("https://github.com/%s", repoURL), nil
}

func (rd *RepositoryDetector) detectComponentsInRepo(repoPath string) ([]DetectedComponent, error) {
	var components []DetectedComponent

	// Walk the repository to detect components
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

		// Detect skill components
		if fileName == "SKILL.md" {
			componentName := filepath.Base(parentDir)
			if componentName == "" || componentName == "." {
				componentName = "root-skill"
			}
			components = append(components, DetectedComponent{
				Type:       ComponentSkill,
				Name:       componentName,
				Path:       relPath,
				SourceFile: fileName,
			})
		}

		// Detect agent components
		if fileName == "AGENT.md" {
			componentName := filepath.Base(parentDir)
			if componentName == "" || componentName == "." {
				componentName = "root-agent"
			}
			components = append(components, DetectedComponent{
				Type:       ComponentAgent,
				Name:       componentName,
				Path:       relPath,
				SourceFile: fileName,
			})
		}

		// Detect command components
		if fileName == "COMMAND.md" {
			componentName := filepath.Base(parentDir)
			if componentName == "" || componentName == "." {
				componentName = "root-command"
			}
			components = append(components, DetectedComponent{
				Type:       ComponentCommand,
				Name:       componentName,
				Path:       relPath,
				SourceFile: fileName,
			})
		}

		return nil
	})

	return components, err
}

func NewSkillDownloader() *SkillDownloader {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get user home directory:", err)
	}

	baseDir := filepath.Join(home, ".agents", "skills")

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		log.Fatal("Failed to create skills directory:", err)
	}

	return &SkillDownloader{
		baseDir:  baseDir,
		detector: NewRepositoryDetector(),
	}
}

func NewAgentDownloader() *AgentDownloader {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get user home directory:", err)
	}

	baseDir := filepath.Join(home, ".agents", "agents")

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		log.Fatal("Failed to create agents directory:", err)
	}

	return &AgentDownloader{
		baseDir:  baseDir,
		detector: NewRepositoryDetector(),
	}
}

func NewCommandDownloader() *CommandDownloader {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get user home directory:", err)
	}

	baseDir := filepath.Join(home, ".agents", "commands")

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		log.Fatal("Failed to create commands directory:", err)
	}

	return &CommandDownloader{
		baseDir:  baseDir,
		detector: NewRepositoryDetector(),
	}
}

func NewComponentLinker() *ComponentLinker {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get user home directory:", err)
	}

	agentsDir := filepath.Join(home, ".agents")
	opencodeDir := filepath.Join(home, ".config", "opencode")

	// Create opencode directory if it doesn't exist
	if err := os.MkdirAll(opencodeDir, 0755); err != nil {
		log.Fatal("Failed to create opencode directory:", err)
	}

	return &ComponentLinker{
		agentsDir:   agentsDir,
		opencodeDir: opencodeDir,
	}
}

func NewUpdateDetector() *UpdateDetector {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get user home directory:", err)
	}

	baseDir := filepath.Join(home, ".agents")

	return &UpdateDetector{
		baseDir:  baseDir,
		detector: NewRepositoryDetector(),
	}
}

func (sd *SkillDownloader) parseRepoURL(repoURL string) (string, error) {
	return sd.detector.normalizeURL(repoURL)
}

func (sd *SkillDownloader) downloadSkill(repoURL, skillName string) error {
	fullURL, err := sd.parseRepoURL(repoURL)
	if err != nil {
		return err
	}

	// Create temporary directory for repository detection
	tempDir, err := os.MkdirTemp("", "agent-smith-detect-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Clone repository to temporary location for detection
	_, err = git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:           fullURL,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository for detection: %w", err)
	}

	// Detect components in the repository
	components, err := sd.detector.detectComponentsInRepo(tempDir)
	if err != nil {
		return fmt.Errorf("failed to detect components: %w", err)
	}

	// Filter for skill components
	var skillComponents []DetectedComponent
	for _, comp := range components {
		if comp.Type == ComponentSkill {
			skillComponents = append(skillComponents, comp)
		}
	}

	if len(skillComponents) == 0 {
		// No skill components detected, fall back to original behavior
		return sd.downloadSkillDirect(fullURL, skillName)
	}

	// Create skill directory
	skillDir := filepath.Join(sd.baseDir, skillName)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	// If only one skill component found, copy its contents
	if len(skillComponents) == 1 {
		component := skillComponents[0]
		componentPath := filepath.Join(tempDir, component.Path)

		// Copy component contents to skill directory
		err = sd.copyDirectoryContents(componentPath, skillDir)
		if err != nil {
			os.RemoveAll(skillDir)
			return fmt.Errorf("failed to copy skill contents: %w", err)
		}
	} else {
		// Multiple skills found, create a monorepo structure
		for _, component := range skillComponents {
			componentDir := filepath.Join(skillDir, component.Name)
			componentPath := filepath.Join(tempDir, component.Path)

			err = os.MkdirAll(componentDir, 0755)
			if err != nil {
				continue
			}

			err = sd.copyDirectoryContents(componentPath, componentDir)
			if err != nil {
				log.Printf("Warning: failed to copy skill %s: %v", component.Name, err)
			}
		}
	}

	// Clone the repository again to get proper git history for metadata
	repo, err := git.PlainClone(skillDir+".git", true, &git.CloneOptions{
		URL:           fullURL,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	})
	if err != nil {
		// Non-fatal, continue without git metadata
		log.Printf("Warning: failed to clone repository for metadata: %v", err)
	}

	var commitHash string
	if repo != nil {
		if ref, err := repo.Head(); err == nil {
			commitHash = ref.Hash().String()
		}
	}

	// Create metadata
	metadata := map[string]interface{}{
		"name":       skillName,
		"source":     fullURL,
		"commit":     commitHash,
		"downloaded": "now",
		"components": len(skillComponents),
		"detection":  "recursive",
	}

	// Save metadata file
	metadataFile := filepath.Join(skillDir, ".skill-metadata.json")
	if err := sd.saveMetadata(metadataFile, metadata); err != nil {
		log.Printf("Warning: failed to save metadata: %v", err)
	}

	// Clean up git clone if it exists
	if _, err := os.Stat(skillDir + ".git"); err == nil {
		os.RemoveAll(skillDir + ".git")
	}

	fmt.Printf("Successfully downloaded skill '%s' from %s\n", skillName, fullURL)
	fmt.Printf("Skill stored in: %s\n", skillDir)
	fmt.Printf("Components detected: %d\n", len(skillComponents))

	return nil
}

func (sd *SkillDownloader) downloadSkillDirect(fullURL, skillName string) error {
	// Create skill directory
	skillDir := filepath.Join(sd.baseDir, skillName)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	// Clone repository directly
	repo, err := git.PlainClone(skillDir, false, &git.CloneOptions{
		URL:           fullURL,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	})
	if err != nil {
		os.RemoveAll(skillDir)
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Get repository info for metadata
	var commitHash string
	if ref, err := repo.Head(); err == nil {
		commitHash = ref.Hash().String()
	}

	// Create metadata
	metadata := map[string]interface{}{
		"name":       skillName,
		"source":     fullURL,
		"commit":     commitHash,
		"downloaded": "now",
		"detection":  "direct",
	}

	// Save metadata file
	metadataFile := filepath.Join(skillDir, ".skill-metadata.json")
	if err := sd.saveMetadata(metadataFile, metadata); err != nil {
		log.Printf("Warning: failed to save metadata: %v", err)
	}

	// Create SKILL.md if it doesn't exist
	skillFile := filepath.Join(skillDir, "SKILL.md")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		if err := sd.createSkillFile(skillFile, skillName, fullURL); err != nil {
			log.Printf("Warning: failed to create SKILL.md: %v", err)
		}
	}

	return nil
}

func (sd *SkillDownloader) copyDirectoryContents(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return sd.copyFile(path, dstPath)
	})
}

func (sd *SkillDownloader) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}

func (sd *SkillDownloader) saveMetadata(filePath string, metadata map[string]interface{}) error {
	metadata["downloaded"] = time.Now().Format(time.RFC3339)

	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return os.WriteFile(filePath, jsonData, 0644)
}

func (sd *SkillDownloader) createSkillFile(filePath, skillName, source string) error {
	content := fmt.Sprintf(`# %s

Downloaded from: %s

## Description

This skill was automatically downloaded by Agent Smith.

## Usage

Add usage instructions here.

---
*Auto-generated by Agent Smith*
`, skillName, source)

	return os.WriteFile(filePath, []byte(content), 0644)
}

func (cd *CommandDownloader) parseRepoURL(repoURL string) (string, error) {
	return cd.detector.normalizeURL(repoURL)
}

func (cd *CommandDownloader) downloadCommand(repoURL, commandName string) error {
	fullURL, err := cd.parseRepoURL(repoURL)
	if err != nil {
		return err
	}

	// Create temporary directory for repository detection
	tempDir, err := os.MkdirTemp("", "agent-smith-detect-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Clone repository to temporary location for detection
	_, err = git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:           fullURL,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository for detection: %w", err)
	}

	// Detect components in the repository
	components, err := cd.detector.detectComponentsInRepo(tempDir)
	if err != nil {
		return fmt.Errorf("failed to detect components: %w", err)
	}

	// Filter for command components
	var commandComponents []DetectedComponent
	for _, comp := range components {
		if comp.Type == ComponentCommand {
			commandComponents = append(commandComponents, comp)
		}
	}

	if len(commandComponents) == 0 {
		// No command components detected, fall back to original behavior
		return cd.downloadCommandDirect(fullURL, commandName)
	}

	// Create command directory
	commandDir := filepath.Join(cd.baseDir, commandName)
	if err := os.MkdirAll(commandDir, 0755); err != nil {
		return fmt.Errorf("failed to create command directory: %w", err)
	}

	// If only one command component found, copy its contents
	if len(commandComponents) == 1 {
		component := commandComponents[0]
		componentPath := filepath.Join(tempDir, component.Path)

		// Copy component contents to command directory
		err = cd.copyDirectoryContents(componentPath, commandDir)
		if err != nil {
			os.RemoveAll(commandDir)
			return fmt.Errorf("failed to copy command contents: %w", err)
		}
	} else {
		// Multiple commands found, create a monorepo structure
		for _, component := range commandComponents {
			componentDir := filepath.Join(commandDir, component.Name)
			componentPath := filepath.Join(tempDir, component.Path)

			err = os.MkdirAll(componentDir, 0755)
			if err != nil {
				continue
			}

			err = cd.copyDirectoryContents(componentPath, componentDir)
			if err != nil {
				log.Printf("Warning: failed to copy command %s: %v", component.Name, err)
			}
		}
	}

	// Clone the repository again to get proper git history for metadata
	repo, err := git.PlainClone(commandDir+".git", true, &git.CloneOptions{
		URL:           fullURL,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	})
	if err != nil {
		// Non-fatal, continue without git metadata
		log.Printf("Warning: failed to clone repository for metadata: %v", err)
	}

	var commitHash string
	if repo != nil {
		if ref, err := repo.Head(); err == nil {
			commitHash = ref.Hash().String()
		}
	}

	// Create metadata
	metadata := map[string]interface{}{
		"name":       commandName,
		"source":     fullURL,
		"commit":     commitHash,
		"downloaded": "now",
		"components": len(commandComponents),
		"detection":  "recursive",
	}

	// Save metadata file
	metadataFile := filepath.Join(commandDir, ".command-lock.json")
	if err := cd.saveMetadata(metadataFile, metadata); err != nil {
		log.Printf("Warning: failed to save metadata: %v", err)
	}

	// Clean up git clone if it exists
	if _, err := os.Stat(commandDir + ".git"); err == nil {
		os.RemoveAll(commandDir + ".git")
	}

	fmt.Printf("Successfully downloaded command '%s' from %s\n", commandName, fullURL)
	fmt.Printf("Command stored in: %s\n", commandDir)
	fmt.Printf("Components detected: %d\n", len(commandComponents))

	return nil
}

func (cd *CommandDownloader) downloadCommandDirect(fullURL, commandName string) error {
	// Create command directory
	commandDir := filepath.Join(cd.baseDir, commandName)
	if err := os.MkdirAll(commandDir, 0755); err != nil {
		return fmt.Errorf("failed to create command directory: %w", err)
	}

	// Clone repository directly
	repo, err := git.PlainClone(commandDir, false, &git.CloneOptions{
		URL:           fullURL,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	})
	if err != nil {
		os.RemoveAll(commandDir)
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Get repository info for metadata
	var commitHash string
	if ref, err := repo.Head(); err == nil {
		commitHash = ref.Hash().String()
	}

	// Create metadata
	metadata := map[string]interface{}{
		"name":       commandName,
		"source":     fullURL,
		"commit":     commitHash,
		"downloaded": "now",
		"detection":  "direct",
	}

	// Save metadata file
	metadataFile := filepath.Join(commandDir, ".command-lock.json")
	if err := cd.saveMetadata(metadataFile, metadata); err != nil {
		log.Printf("Warning: failed to save metadata: %v", err)
	}

	// Create COMMAND.md if it doesn't exist
	commandFile := filepath.Join(commandDir, "COMMAND.md")
	if _, err := os.Stat(commandFile); os.IsNotExist(err) {
		if err := cd.createCommandFile(commandFile, commandName, fullURL); err != nil {
			log.Printf("Warning: failed to create COMMAND.md: %v", err)
		}
	}

	return nil
}

func (cd *CommandDownloader) copyDirectoryContents(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return cd.copyFile(path, dstPath)
	})
}

func (cd *CommandDownloader) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}

func (cd *CommandDownloader) saveMetadata(filePath string, metadata map[string]interface{}) error {
	metadata["downloaded"] = time.Now().Format(time.RFC3339)

	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return os.WriteFile(filePath, jsonData, 0644)
}

func (cd *CommandDownloader) createCommandFile(filePath, commandName, source string) error {
	content := fmt.Sprintf(`# %s

Downloaded from: %s

## Description

This command was automatically downloaded by Agent Smith.

## Usage

Add usage instructions here.

---
*Auto-generated by Agent Smith*
`, commandName, source)

	return os.WriteFile(filePath, []byte(content), 0644)
}

func (ad *AgentDownloader) parseRepoURL(repoURL string) (string, error) {
	return ad.detector.normalizeURL(repoURL)
}

func (ad *AgentDownloader) downloadAgent(repoURL, agentName string) error {
	fullURL, err := ad.parseRepoURL(repoURL)
	if err != nil {
		return err
	}

	// Create temporary directory for repository detection
	tempDir, err := os.MkdirTemp("", "agent-smith-detect-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Clone repository to temporary location for detection
	_, err = git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:           fullURL,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository for detection: %w", err)
	}

	// Detect components in the repository
	components, err := ad.detector.detectComponentsInRepo(tempDir)
	if err != nil {
		return fmt.Errorf("failed to detect components: %w", err)
	}

	// Filter for agent components
	var agentComponents []DetectedComponent
	for _, comp := range components {
		if comp.Type == ComponentAgent {
			agentComponents = append(agentComponents, comp)
		}
	}

	if len(agentComponents) == 0 {
		// No agent components detected, fall back to original behavior
		return ad.downloadAgentDirect(fullURL, agentName)
	}

	// Create agent directory
	agentDir := filepath.Join(ad.baseDir, agentName)
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		return fmt.Errorf("failed to create agent directory: %w", err)
	}

	// If only one agent component found, copy its contents
	if len(agentComponents) == 1 {
		component := agentComponents[0]
		componentPath := filepath.Join(tempDir, component.Path)

		// Copy component contents to agent directory
		err = ad.copyDirectoryContents(componentPath, agentDir)
		if err != nil {
			os.RemoveAll(agentDir)
			return fmt.Errorf("failed to copy agent contents: %w", err)
		}
	} else {
		// Multiple agents found, create a monorepo structure
		for _, component := range agentComponents {
			componentDir := filepath.Join(agentDir, component.Name)
			componentPath := filepath.Join(tempDir, component.Path)

			err = os.MkdirAll(componentDir, 0755)
			if err != nil {
				continue
			}

			err = ad.copyDirectoryContents(componentPath, componentDir)
			if err != nil {
				log.Printf("Warning: failed to copy agent %s: %v", component.Name, err)
			}
		}
	}

	// Clone the repository again to get proper git history for metadata
	repo, err := git.PlainClone(agentDir+".git", true, &git.CloneOptions{
		URL:           fullURL,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	})
	if err != nil {
		// Non-fatal, continue without git metadata
		log.Printf("Warning: failed to clone repository for metadata: %v", err)
	}

	var commitHash string
	if repo != nil {
		if ref, err := repo.Head(); err == nil {
			commitHash = ref.Hash().String()
		}
	}

	// Create metadata
	metadata := map[string]interface{}{
		"name":       agentName,
		"source":     fullURL,
		"commit":     commitHash,
		"downloaded": "now",
		"components": len(agentComponents),
		"detection":  "recursive",
	}

	// Save metadata file
	metadataFile := filepath.Join(agentDir, ".agent-lock.json")
	if err := ad.saveMetadata(metadataFile, metadata); err != nil {
		log.Printf("Warning: failed to save metadata: %v", err)
	}

	// Clean up git clone if it exists
	if _, err := os.Stat(agentDir + ".git"); err == nil {
		os.RemoveAll(agentDir + ".git")
	}

	fmt.Printf("Successfully downloaded agent '%s' from %s\n", agentName, fullURL)
	fmt.Printf("Agent stored in: %s\n", agentDir)
	fmt.Printf("Components detected: %d\n", len(agentComponents))

	return nil
}

func (ad *AgentDownloader) downloadAgentDirect(fullURL, agentName string) error {
	// Create agent directory
	agentDir := filepath.Join(ad.baseDir, agentName)
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		return fmt.Errorf("failed to create agent directory: %w", err)
	}

	// Clone repository directly
	repo, err := git.PlainClone(agentDir, false, &git.CloneOptions{
		URL:           fullURL,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	})
	if err != nil {
		os.RemoveAll(agentDir)
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Get repository info for metadata
	var commitHash string
	if ref, err := repo.Head(); err == nil {
		commitHash = ref.Hash().String()
	}

	// Create metadata
	metadata := map[string]interface{}{
		"name":       agentName,
		"source":     fullURL,
		"commit":     commitHash,
		"downloaded": "now",
		"detection":  "direct",
	}

	// Save metadata file
	metadataFile := filepath.Join(agentDir, ".agent-lock.json")
	if err := ad.saveMetadata(metadataFile, metadata); err != nil {
		log.Printf("Warning: failed to save metadata: %v", err)
	}

	// Create AGENT.md if it doesn't exist
	agentFile := filepath.Join(agentDir, "AGENT.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		if err := ad.createAgentFile(agentFile, agentName, fullURL); err != nil {
			log.Printf("Warning: failed to create AGENT.md: %v", err)
		}
	}

	return nil
}

func (ad *AgentDownloader) copyDirectoryContents(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return ad.copyFile(path, dstPath)
	})
}

func (ad *AgentDownloader) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}

func (ad *AgentDownloader) saveMetadata(filePath string, metadata map[string]interface{}) error {
	metadata["downloaded"] = time.Now().Format(time.RFC3339)

	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return os.WriteFile(filePath, jsonData, 0644)
}

func (ad *AgentDownloader) createAgentFile(filePath, agentName, source string) error {
	content := fmt.Sprintf(`# %s

Downloaded from: %s

## Description

This agent was automatically downloaded by Agent Smith.

## Usage

Add usage instructions here.

---
*Auto-generated by Agent Smith*
`, agentName, source)

	return os.WriteFile(filePath, []byte(content), 0644)
}

func (cl *ComponentLinker) createSymlink(src, dst string) error {
	// Remove existing destination if it exists
	if _, err := os.Lstat(dst); err == nil {
		os.Remove(dst)
	}

	// Create relative path for the symlink
	relPath, err := filepath.Rel(filepath.Dir(dst), src)
	if err != nil {
		return fmt.Errorf("failed to create relative path: %w", err)
	}

	// Create the symbolic link
	if err := os.Symlink(relPath, dst); err != nil {
		// Try fallback to junction on Windows
		if runtime.GOOS == "windows" {
			return cl.createJunction(src, dst)
		}
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

func (cl *ComponentLinker) createJunction(src, dst string) error {
	// For Windows, we would need to use Windows API calls
	// For now, fall back to copying the directory
	return cl.copyDirectory(src, dst)
}

func (cl *ComponentLinker) copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return cl.copyFile(path, dstPath)
	})
}

func (cl *ComponentLinker) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}

func (cl *ComponentLinker) linkComponent(componentType, componentName string) error {
	srcDir := filepath.Join(cl.agentsDir, componentType, componentName)
	dstDir := filepath.Join(cl.opencodeDir, componentType, componentName)

	// Check if source component exists
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("component %s/%s does not exist in %s", componentType, componentName, cl.agentsDir)
	}

	// Create destination directory
	if err := os.MkdirAll(filepath.Dir(dstDir), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Create symlink or copy
	if err := cl.createSymlink(srcDir, dstDir); err != nil {
		return fmt.Errorf("failed to link component: %w", err)
	}

	fmt.Printf("Successfully linked %s '%s' to opencode\n", componentType, componentName)
	fmt.Printf("Source: %s\n", srcDir)
	fmt.Printf("Target: %s\n", dstDir)

	return nil
}

func (cl *ComponentLinker) linkAllComponents() error {
	componentTypes := []string{"skills", "agents", "commands"}

	for _, componentType := range componentTypes {
		typeDir := filepath.Join(cl.agentsDir, componentType)
		if _, err := os.Stat(typeDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(typeDir)
		if err != nil {
			fmt.Printf("Warning: failed to read %s directory: %v\n", componentType, err)
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				if err := cl.linkComponent(componentType, entry.Name()); err != nil {
					fmt.Printf("Warning: failed to link %s/%s: %v\n", componentType, entry.Name(), err)
				}
			}
		}
	}

	return nil
}

func (ud *UpdateDetector) loadMetadata(componentType, componentName string) (*ComponentMetadata, error) {
	var metadataFile string
	switch componentType {
	case "skills":
		metadataFile = filepath.Join(ud.baseDir, "skills", componentName, ".skill-metadata.json")
	case "agents":
		metadataFile = filepath.Join(ud.baseDir, "agents", componentName, ".agent-lock.json")
	case "commands":
		metadataFile = filepath.Join(ud.baseDir, "commands", componentName, ".command-lock.json")
	default:
		return nil, fmt.Errorf("unknown component type: %s", componentType)
	}

	data, err := os.ReadFile(metadataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata ComponentMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

func (ud *UpdateDetector) getCurrentRepoSHA(repoURL string) (string, error) {
	fullURL, err := ud.detector.normalizeURL(repoURL)
	if err != nil {
		return "", fmt.Errorf("failed to normalize URL: %w", err)
	}

	// Create temporary directory for checking current state
	tempDir, err := os.MkdirTemp("", "agent-smith-check-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Clone repository to get current HEAD
	repo, err := git.PlainClone(tempDir, true, &git.CloneOptions{
		URL:           fullURL,
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	// Get HEAD commit hash
	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	return ref.Hash().String(), nil
}

func (ud *UpdateDetector) HasUpdates(componentType, componentName, repoURL string) (bool, error) {
	// Load existing metadata
	metadata, err := ud.loadMetadata(componentType, componentName)
	if err != nil {
		return false, fmt.Errorf("failed to load metadata: %w", err)
	}

	// Get current repository SHA
	currentSHA, err := ud.getCurrentRepoSHA(repoURL)
	if err != nil {
		return false, fmt.Errorf("failed to get current repository SHA: %w", err)
	}

	// Compare stored SHA with current SHA
	return metadata.Commit != currentSHA, nil
}

func (ud *UpdateDetector) UpdateComponent(componentType, componentName, repoURL string) error {
	hasUpdates, err := ud.HasUpdates(componentType, componentName, repoURL)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !hasUpdates {
		fmt.Printf("Component %s/%s is already up to date\n", componentType, componentName)
		return nil
	}

	fmt.Printf("Updates detected for %s/%s, downloading new version...\n", componentType, componentName)

	// Re-download the component with the latest changes
	switch componentType {
	case "skills":
		downloader := NewSkillDownloader()
		return downloader.downloadSkill(repoURL, componentName)
	case "agents":
		downloader := NewAgentDownloader()
		return downloader.downloadAgent(repoURL, componentName)
	case "commands":
		downloader := NewCommandDownloader()
		return downloader.downloadCommand(repoURL, componentName)
	default:
		return fmt.Errorf("unknown component type: %s", componentType)
	}
}

func (ud *UpdateDetector) UpdateAll() error {
	componentTypes := []string{"skills", "agents", "commands"}

	for _, componentType := range componentTypes {
		typeDir := filepath.Join(ud.baseDir, componentType)
		if _, err := os.Stat(typeDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(typeDir)
		if err != nil {
			fmt.Printf("Warning: failed to read %s directory: %v\n", componentType, err)
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				componentName := entry.Name()

				// Load metadata to get source URL
				metadata, err := ud.loadMetadata(componentType, componentName)
				if err != nil {
					fmt.Printf("Warning: failed to load metadata for %s/%s: %v\n", componentType, componentName, err)
					continue
				}

				if err := ud.UpdateComponent(componentType, componentName, metadata.Source); err != nil {
					fmt.Printf("Warning: failed to update %s/%s: %v\n", componentType, componentName, err)
				}
			}
		}
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: agent-smith <command> [args...]")
		fmt.Println("Commands:")
		fmt.Println("  add-skill   <repository-url> <skill-name>   Download a skill from a git repository")
		fmt.Println("  add-agent   <repository-url> <agent-name>   Download an agent from a git repository")
		fmt.Println("  add-command <repository-url> <command-name> Download a command from a git repository")
		fmt.Println("  update      <type> <name>                  Check and update a specific component")
		fmt.Println("  update-all                                  Check and update all downloaded components")
		fmt.Println("  link        <type> <name>                   Link a downloaded component to opencode")
		fmt.Println("  link-all                                     Link all downloaded components to opencode")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  agent-smith add-skill owner/repo my-skill")
		fmt.Println("  agent-smith add-skill https://github.com/owner/repo my-skill")
		fmt.Println("  agent-smith add-agent owner/repo my-agent")
		fmt.Println("  agent-smith add-agent https://github.com/owner/repo my-agent")
		fmt.Println("  agent-smith add-command owner/repo my-command")
		fmt.Println("  agent-smith add-command https://github.com/owner/repo my-command")
		fmt.Println("  agent-smith update skills my-skill")
		fmt.Println("  agent-smith update agents my-agent")
		fmt.Println("  agent-smith update commands my-command")
		fmt.Println("  agent-smith update-all")
		fmt.Println("  agent-smith link skills my-skill")
		fmt.Println("  agent-smith link agents my-agent")
		fmt.Println("  agent-smith link commands my-command")
		fmt.Println("  agent-smith link-all")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "add-skill":
		if len(os.Args) < 4 {
			fmt.Println("Usage: agent-smith add-skill <repository-url> <skill-name>")
			os.Exit(1)
		}
		repoURL := os.Args[2]
		name := os.Args[3]
		downloader := NewSkillDownloader()
		if err := downloader.downloadSkill(repoURL, name); err != nil {
			log.Fatal("Failed to download skill:", err)
		}
	case "add-agent":
		if len(os.Args) < 4 {
			fmt.Println("Usage: agent-smith add-agent <repository-url> <agent-name>")
			os.Exit(1)
		}
		repoURL := os.Args[2]
		name := os.Args[3]
		downloader := NewAgentDownloader()
		if err := downloader.downloadAgent(repoURL, name); err != nil {
			log.Fatal("Failed to download agent:", err)
		}
	case "add-command":
		if len(os.Args) < 4 {
			fmt.Println("Usage: agent-smith add-command <repository-url> <command-name>")
			os.Exit(1)
		}
		repoURL := os.Args[2]
		name := os.Args[3]
		downloader := NewCommandDownloader()
		if err := downloader.downloadCommand(repoURL, name); err != nil {
			log.Fatal("Failed to download command:", err)
		}
	case "link":
		if len(os.Args) < 4 {
			fmt.Println("Usage: agent-smith link <type> <name>")
			fmt.Println("Types: skills, agents, commands")
			os.Exit(1)
		}
		componentType := os.Args[2]
		componentName := os.Args[3]
		linker := NewComponentLinker()
		if err := linker.linkComponent(componentType, componentName); err != nil {
			log.Fatal("Failed to link component:", err)
		}
	case "update":
		if len(os.Args) < 4 {
			fmt.Println("Usage: agent-smith update <type> <name>")
			fmt.Println("Types: skills, agents, commands")
			os.Exit(1)
		}
		componentType := os.Args[2]
		componentName := os.Args[3]

		// Validate component type
		if componentType != "skills" && componentType != "agents" && componentType != "commands" {
			fmt.Println("Invalid component type. Use: skills, agents, or commands")
			os.Exit(1)
		}

		detector := NewUpdateDetector()

		// Load metadata to get source URL
		metadata, err := detector.loadMetadata(componentType, componentName)
		if err != nil {
			log.Fatal("Failed to load component metadata:", err)
		}

		if err := detector.UpdateComponent(componentType, componentName, metadata.Source); err != nil {
			log.Fatal("Failed to update component:", err)
		}
	case "update-all":
		detector := NewUpdateDetector()
		if err := detector.UpdateAll(); err != nil {
			log.Fatal("Failed to update components:", err)
		}
	case "link-all":
		linker := NewComponentLinker()
		if err := linker.linkAllComponents(); err != nil {
			log.Fatal("Failed to link all components:", err)
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Supported commands: add-skill, add-agent, add-command, update, update-all, link, link-all")
		os.Exit(1)
	}
}
