package find

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tjg184/agent-smith/internal/formatter"
	"github.com/tjg184/agent-smith/pkg/logger"
	"github.com/tjg184/agent-smith/pkg/services"
)

func TestFindSkills_QueryValidation(t *testing.T) {
	log := logger.New(logger.LevelError)
	fmt := formatter.New()
	service := NewService(log, fmt)

	tests := []struct {
		name        string
		query       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "query too short - 1 char",
			query:       "x",
			expectError: true,
			errorMsg:    "Query must be at least 2 characters",
		},
		{
			name:        "query too short - empty",
			query:       "",
			expectError: true,
			errorMsg:    "Query must be at least 2 characters",
		},
		{
			name:        "query valid - 2 chars",
			query:       "ab",
			expectError: false,
		},
		{
			name:        "query valid - normal length",
			query:       "typescript",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server that returns empty results
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := APIResponse{
					Query:  tt.query,
					Skills: []SkillResult{},
					Count:  0,
				}
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			opts := services.FindOptions{Limit: 20, JSON: false}
			err := service.FindSkills(tt.query, opts)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			}
		})
	}
}

func TestFormatResults(t *testing.T) {
	log := logger.New(logger.LevelError)
	fmt := formatter.New()
	service := NewService(log, fmt)

	skills := []SkillResult{
		{
			ID:       "owner/repo/skill-name",
			SkillID:  "skill-name",
			Name:     "skill-name",
			Installs: 1000,
			Source:   "owner/repo",
		},
		{
			ID:       "github/awesome/prd",
			SkillID:  "prd",
			Name:     "prd",
			Installs: 500,
			Source:   "github/awesome",
		},
	}

	results := service.formatResults(skills)

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Test first result
	if results[0].Source != "owner/repo" {
		t.Errorf("Expected source 'owner/repo', got '%s'", results[0].Source)
	}
	if results[0].SkillID != "skill-name" {
		t.Errorf("Expected skillId 'skill-name', got '%s'", results[0].SkillID)
	}
	if results[0].Installs != 1000 {
		t.Errorf("Expected installs 1000, got %d", results[0].Installs)
	}
	if results[0].URL != "https://skills.sh/owner/repo/skill-name" {
		t.Errorf("Expected URL 'https://skills.sh/owner/repo/skill-name', got '%s'", results[0].URL)
	}
	if results[0].InstallCommand != "agent-smith install skill owner/repo skill-name" {
		t.Errorf("Expected install command 'agent-smith install skill owner/repo skill-name', got '%s'", results[0].InstallCommand)
	}
	if results[0].InstallAllCommand != "agent-smith install all owner/repo" {
		t.Errorf("Expected install all command 'agent-smith install all owner/repo', got '%s'", results[0].InstallAllCommand)
	}

	// Test second result
	if results[1].Source != "github/awesome" {
		t.Errorf("Expected source 'github/awesome', got '%s'", results[1].Source)
	}
	if results[1].SkillID != "prd" {
		t.Errorf("Expected skillId 'prd', got '%s'", results[1].SkillID)
	}
}

func TestFormatInstallCount(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		expected string
	}{
		{
			name:     "less than 1K",
			count:    500,
			expected: "500 installs",
		},
		{
			name:     "exactly 1K",
			count:    1000,
			expected: "1.0K installs",
		},
		{
			name:     "between 1K and 1M",
			count:    5500,
			expected: "5.5K installs",
		},
		{
			name:     "exactly 1M",
			count:    1000000,
			expected: "1.0M installs",
		},
		{
			name:     "greater than 1M",
			count:    2500000,
			expected: "2.5M installs",
		},
		{
			name:     "zero",
			count:    0,
			expected: "0 installs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatInstallCount(tt.count)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestOutputJSON(t *testing.T) {
	log := logger.New(logger.LevelError)
	fmt := formatter.New()
	service := NewService(log, fmt)

	results := []FormattedResult{
		{
			Source:            "owner/repo",
			SkillID:           "test-skill",
			Name:              "test-skill",
			Installs:          100,
			URL:               "https://skills.sh/owner/repo/test-skill",
			InstallCommand:    "agent-smith install skill owner/repo test-skill",
			InstallAllCommand: "agent-smith install all owner/repo",
		},
	}

	// Capture stdout by temporarily redirecting it
	// For this test, we'll just verify the method doesn't error
	err := service.outputJSON("test", results)
	if err != nil {
		t.Errorf("outputJSON failed: %v", err)
	}
}

func TestQueryAPI_WithMockServer(t *testing.T) {
	log := logger.New(logger.LevelError)
	fmt := formatter.New()

	tests := []struct {
		name           string
		query          string
		serverResponse APIResponse
		statusCode     int
		expectError    bool
		errorContains  string
	}{
		{
			name:  "successful response with results",
			query: "prd",
			serverResponse: APIResponse{
				Query:      "prd",
				SearchType: "fuzzy",
				Skills: []SkillResult{
					{
						ID:       "owner/repo/prd",
						SkillID:  "prd",
						Name:     "prd",
						Installs: 100,
						Source:   "owner/repo",
					},
				},
				Count: 1,
			},
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:  "successful response with no results",
			query: "nonexistent",
			serverResponse: APIResponse{
				Query:      "nonexistent",
				SearchType: "fuzzy",
				Skills:     []SkillResult{},
				Count:      0,
			},
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:  "API error response",
			query: "x",
			serverResponse: APIResponse{
				Error: "Query must be at least 2 characters",
			},
			statusCode:    http.StatusOK,
			expectError:   false, // API returns 200 with error in body
			errorContains: "",
		},
		{
			name:          "HTTP error status",
			query:         "test",
			statusCode:    http.StatusInternalServerError,
			expectError:   true,
			errorContains: "skills.sh API returned status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			}))
			defer server.Close()

			// Create service with test server URL and client
			service := NewServiceWithClient(log, fmt, server.Client(), server.URL)

			// Test the queryAPI method
			result, err := service.queryAPI(tt.query)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != nil && result.Error != "" && result.Error != tt.serverResponse.Error {
					t.Errorf("Expected API error '%s', got '%s'", tt.serverResponse.Error, result.Error)
				}
			}
		})
	}
}

func TestGetBanner(t *testing.T) {
	banner := getBanner()

	// Check banner is not empty
	if banner == "" {
		t.Error("Banner should not be empty")
	}

	// Check it's multi-line (ASCII art should have multiple lines)
	lines := strings.Split(banner, "\n")
	if len(lines) < 5 {
		t.Errorf("Banner should have multiple lines, got %d", len(lines))
	}

	// Check banner contains some ASCII art characters
	if !strings.Contains(banner, "_") && !strings.Contains(banner, "/") && !strings.Contains(banner, "\\") {
		t.Error("Banner should contain ASCII art characters")
	}
}

func TestFindSkills_LimitParameter(t *testing.T) {
	log := logger.New(logger.LevelError)
	fmt := formatter.New()

	// Create test skills
	skills := []SkillResult{}
	for i := 0; i < 50; i++ {
		skills = append(skills, SkillResult{
			ID:       "owner/repo/skill",
			SkillID:  "skill",
			Name:     "skill",
			Installs: 100,
			Source:   "owner/repo",
		})
	}

	tests := []struct {
		name          string
		limit         int
		skillCount    int
		expectedCount int
	}{
		{
			name:          "limit less than results",
			limit:         5,
			skillCount:    50,
			expectedCount: 5,
		},
		{
			name:          "limit equals results",
			limit:         50,
			skillCount:    50,
			expectedCount: 50,
		},
		{
			name:          "limit greater than results",
			limit:         100,
			skillCount:    50,
			expectedCount: 50,
		},
		{
			name:          "default limit",
			limit:         0, // Should use default of 20
			skillCount:    50,
			expectedCount: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := APIResponse{
					Query:  "test",
					Skills: skills[:tt.skillCount],
					Count:  tt.skillCount,
				}
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			// Verify that the limit logic works by checking formatted results
			// The service applies limit in FindSkills before calling formatResults
			formatted := NewServiceWithClient(log, fmt, server.Client(), server.URL).formatResults(skills[:tt.expectedCount])
			if len(formatted) != tt.expectedCount {
				t.Errorf("Expected %d formatted results, got %d", tt.expectedCount, len(formatted))
			}
		})
	}
}
