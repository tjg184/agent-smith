package find

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/fatih/color"
	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/services"
)

const (
	skillsAPIURL = "https://skills.sh/api/search"
	defaultLimit = 20
)

// Service handles finding components in remote registries
type Service struct {
	logger    *logger.Logger
	formatter *formatter.Formatter
	client    *http.Client
	apiURL    string
}

// NewService creates a new find service
func NewService(log *logger.Logger, fmt *formatter.Formatter) *Service {
	return &Service{
		logger:    log,
		formatter: fmt,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiURL: skillsAPIURL,
	}
}

// NewServiceWithClient creates a new find service with custom HTTP client and API URL (for testing)
func NewServiceWithClient(log *logger.Logger, fmt *formatter.Formatter, client *http.Client, apiURL string) *Service {
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	if apiURL == "" {
		apiURL = skillsAPIURL
	}
	return &Service{
		logger:    log,
		formatter: fmt,
		client:    client,
		apiURL:    apiURL,
	}
}

// SkillResult represents a single skill search result from skills.sh API
type SkillResult struct {
	ID       string `json:"id"`
	SkillID  string `json:"skillId"`
	Name     string `json:"name"`
	Installs int    `json:"installs"`
	Source   string `json:"source"`
}

// APIResponse represents the response from skills.sh API
type APIResponse struct {
	Query      string        `json:"query"`
	SearchType string        `json:"searchType"`
	Skills     []SkillResult `json:"skills"`
	Count      int           `json:"count"`
	Error      string        `json:"error,omitempty"`
}

// FormattedResult represents a result formatted for output
type FormattedResult struct {
	Source            string `json:"source"`
	SkillID           string `json:"skillId"`
	Name              string `json:"name"`
	Installs          int    `json:"installs"`
	URL               string `json:"url"`
	InstallCommand    string `json:"installCommand"`
	InstallAllCommand string `json:"installAllCommand"`
}

// FindSkills searches for skills in the skills.sh registry
func (s *Service) FindSkills(query string, opts services.FindOptions) error {
	// Validate query length
	if len(query) < 2 {
		return fmt.Errorf("Query must be at least 2 characters")
	}

	if opts.Limit == 0 {
		opts.Limit = defaultLimit
	}

	results, err := s.queryAPI(query)
	if err != nil {
		return err
	}

	if results.Error != "" {
		return fmt.Errorf("skills.sh API returned an error\nPlease try again later or visit https://skills.sh")
	}

	if len(results.Skills) == 0 {
		if opts.JSON {
			return s.outputJSON(query, []FormattedResult{})
		}
		fmt.Printf("No skills found matching '%s'\n\n", query)
		fmt.Println("Try different keywords or visit https://skills.sh to browse all skills.")
		return nil
	}

	// Apply limit
	if len(results.Skills) > opts.Limit {
		results.Skills = results.Skills[:opts.Limit]
	}

	// Format results
	formatted := s.formatResults(results.Skills)

	// Output
	if opts.JSON {
		return s.outputJSON(query, formatted)
	}

	return s.outputTerminal(formatted)
}

// queryAPI makes the HTTP request to skills.sh API
func (s *Service) queryAPI(query string) (*APIResponse, error) {
	url := fmt.Sprintf("%s?q=%s", s.apiURL, query)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to skills.sh registry\nCheck your internet connection and try again.")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("skills.sh API returned status %d\nPlease try again later or visit https://skills.sh", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read API response: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("Failed to parse API response: %w", err)
	}

	return &apiResp, nil
}

// formatResults converts API results to formatted results
func (s *Service) formatResults(skills []SkillResult) []FormattedResult {
	results := make([]FormattedResult, len(skills))
	for i, skill := range skills {
		results[i] = FormattedResult{
			Source:            skill.Source,
			SkillID:           skill.SkillID,
			Name:              skill.Name,
			Installs:          skill.Installs,
			URL:               fmt.Sprintf("https://skills.sh/%s/%s", skill.Source, skill.SkillID),
			InstallCommand:    fmt.Sprintf("agent-smith install skill %s %s", skill.Source, skill.SkillID),
			InstallAllCommand: fmt.Sprintf("agent-smith install all %s", skill.Source),
		}
	}
	return results
}

// outputJSON outputs results as JSON
func (s *Service) outputJSON(query string, results []FormattedResult) error {
	output := map[string]interface{}{
		"query":   query,
		"count":   len(results),
		"results": results,
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("Failed to generate JSON output: %w", err)
	}

	fmt.Println(string(jsonBytes))
	return nil
}

// outputTerminal outputs results with colored terminal formatting
func (s *Service) outputTerminal(results []FormattedResult) error {
	// Color definitions
	cyan := color.New(color.FgCyan).SprintFunc()
	gray := color.New(color.FgHiBlack).SprintFunc()
	brightWhite := color.New(color.FgHiWhite).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	// Print Agent Smith banner
	fmt.Print(getBanner())
	fmt.Println()

	// Print installation instructions header
	fmt.Printf("%s  agent-smith install skill <owner/repo> <skill-name>\n", gray("Install with:"))
	fmt.Printf("%s  agent-smith install all <owner/repo>\n", gray("          or:"))
	fmt.Println()

	// Print each result
	for _, result := range results {
		// Line 1: source@skillId and install count
		identifier := fmt.Sprintf("%s@%s", result.Source, result.SkillID)
		installCount := formatInstallCount(result.Installs)
		fmt.Printf("%s %s\n", brightWhite(identifier), cyan(installCount))

		// Line 2: skills.sh URL
		fmt.Printf("%s %s\n", gray("└"), gray(result.URL))

		// Line 3: install command (dimmed)
		fmt.Printf("  %s\n", yellow(result.InstallCommand))
		fmt.Println()
	}

	return nil
}

// getBanner returns the Agent Smith ASCII art banner
func getBanner() string {
	return `  ___                   _     _____           _ _   _     
 / _ \                 | |   /  ___|         (_) | | |    
/ /_\ \ __ _  ___ _ __ | |_  \ ` + "`" + `--. _ __ ___  _| |_| |__  
|  _  |/ _` + "`" + ` |/ _ \ '_ \| __|  ` + "`" + `--. \ '_ ` + "`" + ` _ \| | __| '_ \ 
| | | | (_| |  __/ | | | |_  /\__/ / | | | | | | |_| | | |
\_| |_/\__, |\___|_| |_|\__| \____/|_| |_| |_|_|\__|_| |_|
        __/ |                                             
       |___/                                              `
}

// formatInstallCount formats install count with K/M suffixes
func formatInstallCount(count int) string {
	if count >= 1000000 {
		return fmt.Sprintf("%.1fM installs", float64(count)/1000000)
	}
	if count >= 1000 {
		return fmt.Sprintf("%.1fK installs", float64(count)/1000)
	}
	return fmt.Sprintf("%d installs", count)
}
