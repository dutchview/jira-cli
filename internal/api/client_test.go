package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dutchview/jira-cli/internal/config"
)

// newTestClient builds a Client pointed at the given httptest.Server.
func newTestClient(srv *httptest.Server) *Client {
	return NewClient(&config.Config{
		BaseURL:  srv.URL,
		Email:    "test@example.com",
		APIToken: "tok",
	})
}

func TestGetIssue_ParsesParent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || !strings.HasPrefix(r.URL.Path, "/rest/api/3/issue/ED-12105") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"id":   "1",
			"key":  "ED-12105",
			"self": "https://example/rest/api/3/issue/1",
			"fields": {
				"summary": "switch contracts offline",
				"parent": {
					"id":   "999",
					"key":  "ED-7281",
					"fields": { "summary": "SecurED" }
				}
			}
		}`)
	}))
	defer srv.Close()

	client := newTestClient(srv)
	issue, err := client.GetIssue("ED-12105", nil, nil)
	if err != nil {
		t.Fatalf("GetIssue error: %v", err)
	}
	if issue.Fields.Parent == nil {
		t.Fatalf("expected Parent to be populated, got nil")
	}
	if issue.Fields.Parent.Key != "ED-7281" {
		t.Errorf("Parent.Key = %q, want ED-7281", issue.Fields.Parent.Key)
	}
}

func TestGetIssue_NoParent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"id": "2", "key": "ED-1", "self": "x",
			"fields": { "summary": "no parent" }
		}`)
	}))
	defer srv.Close()

	client := newTestClient(srv)
	issue, err := client.GetIssue("ED-1", nil, nil)
	if err != nil {
		t.Fatalf("GetIssue error: %v", err)
	}
	if issue.Fields.Parent != nil {
		t.Errorf("expected nil Parent, got %+v", issue.Fields.Parent)
	}
}

func TestUpdateIssue_SetParent(t *testing.T) {
	srv, captured, path := captureRequest(t, nil)
	defer srv.Close()

	client := newTestClient(srv)
	err := client.UpdateIssue("ED-12105", map[string]interface{}{
		"parent": map[string]string{"key": "ED-7281"},
	})
	if err != nil {
		t.Fatalf("UpdateIssue error: %v", err)
	}
	if got, want := *path, "/rest/api/3/issue/ED-12105"; got != want {
		t.Errorf("path = %q, want %q", got, want)
	}
	fields, ok := (*captured)["fields"].(map[string]interface{})
	if !ok {
		t.Fatalf("fields missing or wrong type: %+v", *captured)
	}
	parent, ok := fields["parent"].(map[string]interface{})
	if !ok {
		t.Fatalf("parent missing or wrong type: %+v", fields)
	}
	if parent["key"] != "ED-7281" {
		t.Errorf("parent.key = %v, want ED-7281", parent["key"])
	}
}

func TestUpdateIssue_ClearParent(t *testing.T) {
	srv, captured, _ := captureRequest(t, nil)
	defer srv.Close()

	client := newTestClient(srv)
	err := client.UpdateIssue("ED-12105", map[string]interface{}{
		"parent": nil,
	})
	if err != nil {
		t.Fatalf("UpdateIssue error: %v", err)
	}
	fields, ok := (*captured)["fields"].(map[string]interface{})
	if !ok {
		t.Fatalf("fields missing: %+v", *captured)
	}
	// After JSON round-trip, an explicit null becomes the JSON null, which
	// decodes back to a nil interface{}. The "parent" key MUST be present.
	if _, present := fields["parent"]; !present {
		t.Fatalf("parent key missing from request body; got fields=%+v", fields)
	}
	if fields["parent"] != nil {
		t.Errorf("parent should be JSON null, got %#v", fields["parent"])
	}
}

func TestCreateIssue_WithParent(t *testing.T) {
	srv, captured, path := captureRequest(t, nil)
	defer srv.Close()

	client := newTestClient(srv)
	_, err := client.CreateIssue("ED", "summary", "Story", nil, "", "", nil, "", "ED-7281")
	if err != nil {
		t.Fatalf("CreateIssue error: %v", err)
	}
	if *path != "/rest/api/3/issue" {
		t.Errorf("path = %q, want /rest/api/3/issue", *path)
	}
	fields := (*captured)["fields"].(map[string]interface{})
	parent, ok := fields["parent"].(map[string]interface{})
	if !ok {
		t.Fatalf("parent missing: %+v", fields)
	}
	if parent["key"] != "ED-7281" {
		t.Errorf("parent.key = %v, want ED-7281", parent["key"])
	}
}

func TestCreateIssue_NoParentOmitted(t *testing.T) {
	srv, captured, _ := captureRequest(t, nil)
	defer srv.Close()

	client := newTestClient(srv)
	_, err := client.CreateIssue("ED", "summary", "Story", nil, "", "", nil, "", "")
	if err != nil {
		t.Fatalf("CreateIssue error: %v", err)
	}
	fields := (*captured)["fields"].(map[string]interface{})
	if _, present := fields["parent"]; present {
		t.Errorf("parent must be omitted when empty; got fields=%+v", fields)
	}
}

func TestCreateIssueWiki_WithParent(t *testing.T) {
	srv, captured, path := captureRequest(t, nil)
	defer srv.Close()

	client := newTestClient(srv)
	_, err := client.CreateIssueWiki("ED", "summary", "Story", "", "", "", nil, "", "ED-7281")
	if err != nil {
		t.Fatalf("CreateIssueWiki error: %v", err)
	}
	if *path != "/rest/api/2/issue" {
		t.Errorf("path = %q, want /rest/api/2/issue", *path)
	}
	fields := (*captured)["fields"].(map[string]interface{})
	parent, ok := fields["parent"].(map[string]interface{})
	if !ok {
		t.Fatalf("parent missing: %+v", fields)
	}
	if parent["key"] != "ED-7281" {
		t.Errorf("parent.key = %v, want ED-7281", parent["key"])
	}
}

// captureRequest decodes the JSON body of a single intercepted request.
func captureRequest(t *testing.T, h func(body map[string]interface{}, r *http.Request)) (*httptest.Server, *map[string]interface{}, *string) {
	t.Helper()
	var captured map[string]interface{}
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		buf, _ := io.ReadAll(r.Body)
		if len(buf) > 0 {
			if err := json.Unmarshal(buf, &captured); err != nil {
				t.Errorf("decode body: %v (raw=%s)", err, string(buf))
			}
		}
		if h != nil {
			h(captured, r)
		}
		w.Header().Set("Content-Type", "application/json")
		// Default response: a stub issue payload.
		_, _ = io.WriteString(w, `{"id":"1","key":"ED-1","self":"x","fields":{"summary":"s"}}`)
	}))
	return srv, &captured, &path
}
