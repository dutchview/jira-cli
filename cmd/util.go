package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mauricejumelet/jira-cli/internal/adf"
	"github.com/mauricejumelet/jira-cli/internal/api"
)

func printJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatTimestamp(ts string) string {
	if ts == "" {
		return "N/A"
	}
	// Jira timestamps: 2024-01-15T10:30:00.000+0000
	t, err := time.Parse("2006-01-02T15:04:05.000-0700", ts)
	if err != nil {
		// Try alternate format
		t, err = time.Parse("2006-01-02T15:04:05.000Z0700", ts)
		if err != nil {
			return ts
		}
	}
	return t.Format("2006-01-02 15:04")
}

func makeHyperlink(url, text string) string {
	// OSC 8 terminal hyperlink
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, text)
}

func issueURL(baseURL, issueKey string) string {
	return strings.TrimRight(baseURL, "/") + "/browse/" + issueKey
}

// newMentionResolver creates a MentionResolver that looks up JIRA users by display name.
func newMentionResolver(client *api.Client) adf.MentionResolver {
	// Cache resolved names to avoid repeated API calls in the same text.
	cache := map[string]*resolvedUser{}

	return func(name string) (string, string, bool) {
		nameLower := strings.ToLower(name)

		// Check cache first
		if cached, ok := cache[nameLower]; ok {
			if cached == nil {
				return "", "", false
			}
			return cached.accountID, cached.displayName, true
		}

		users, err := client.SearchUsers(name, 10)
		if err != nil {
			cache[nameLower] = nil
			return "", "", false
		}

		// Find exact display name match (case-insensitive)
		for _, u := range users {
			if strings.EqualFold(u.DisplayName, name) {
				cache[nameLower] = &resolvedUser{u.AccountID, u.DisplayName}
				return u.AccountID, u.DisplayName, true
			}
		}

		cache[nameLower] = nil
		return "", "", false
	}
}

type resolvedUser struct {
	accountID   string
	displayName string
}
