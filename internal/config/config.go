package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	BaseURL  string
	Email    string
	APIToken string
}

// ConfigLocations returns config file locations in order of increasing priority
// (lowest priority first). All existing files are loaded and merged, with
// later entries overriding earlier ones.
func ConfigLocations() []string {
	var locations []string

	homeDir, err := os.UserHomeDir()
	if err == nil {
		locations = append(locations, filepath.Join(homeDir, ".config", "jira", ".env"))
	}

	locations = append(locations, ".env") // Current directory (highest priority)

	return locations
}

// Load loads configuration from environment variables and optional .env files.
// The configFile parameter allows specifying a custom config file path.
// If empty, all default locations are loaded and merged in priority order:
//  1. ~/.config/jira/.env  (base config)
//  2. .env in current directory (overrides base config if present)
//
// OS environment variables always take precedence over file values.
func Load(configFile string) (*Config, error) {
	if configFile != "" {
		if err := godotenv.Load(configFile); err != nil {
			return nil, fmt.Errorf("failed to load config file %s: %w", configFile, err)
		}
	} else {
		// Read all files into a merged map (later files override earlier ones)
		merged := make(map[string]string)
		for _, loc := range ConfigLocations() {
			if fileEnv, err := godotenv.Read(loc); err == nil {
				for k, v := range fileEnv {
					merged[k] = v
				}
			}
		}
		// Apply merged values — OS environment variables take precedence
		for k, v := range merged {
			if os.Getenv(k) == "" {
				os.Setenv(k, v)
			}
		}
	}

	baseURL := os.Getenv("JIRA_BASE_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("JIRA_BASE_URL not set.\n\n%s", configHelp())
	}
	baseURL = strings.TrimRight(baseURL, "/")

	email := os.Getenv("JIRA_EMAIL")
	if email == "" {
		return nil, fmt.Errorf("JIRA_EMAIL not set.\n\n%s", configHelp())
	}

	apiToken := os.Getenv("JIRA_API_TOKEN")
	if apiToken == "" {
		return nil, fmt.Errorf("JIRA_API_TOKEN not set.\n\n%s", configHelp())
	}

	return &Config{
		BaseURL:  baseURL,
		Email:    email,
		APIToken: apiToken,
	}, nil
}

func configHelp() string {
	locations := ConfigLocations()
	var sb strings.Builder

	sb.WriteString("Configuration can be provided via:\n")
	sb.WriteString("  1. Environment variables (JIRA_BASE_URL, JIRA_EMAIL, JIRA_API_TOKEN)\n")
	sb.WriteString("  2. A .env file in one of these locations:\n")
	for _, loc := range locations {
		sb.WriteString(fmt.Sprintf("     - %s\n", loc))
	}
	sb.WriteString("  3. A custom config file via --config flag\n")
	sb.WriteString("\nExample .env file:\n")
	sb.WriteString("  JIRA_BASE_URL=https://yourcompany.atlassian.net\n")
	sb.WriteString("  JIRA_EMAIL=you@example.com\n")
	sb.WriteString("  JIRA_API_TOKEN=your_api_token\n")
	sb.WriteString("\nGet your API token at: https://id.atlassian.com/manage-profile/security/api-tokens")

	return sb.String()
}

// PrintConfigHelp prints the configuration help message.
func PrintConfigHelp() {
	fmt.Println("Jira CLI Configuration")
	fmt.Println("======================")
	fmt.Println()
	fmt.Println(configHelp())
}
