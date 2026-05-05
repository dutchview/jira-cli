package cmd

import (
	"net/http"
	"strings"
	"testing"
)

func TestIssuesCreate_PassesParent(t *testing.T) {
	rr := newRecorder()
	defer rr.Close()

	cmd := &IssuesCreateCmd{
		Project: "ED",
		Type:    "Story",
		Summary: "switch contracts offline",
		Parent:  "ED-7281",
	}
	_ = captureStdout(t, func() {
		if err := cmd.Run(rr.Client()); err != nil {
			t.Fatalf("Run: %v", err)
		}
	})

	var post *recordedRequest
	for i := range rr.Requests {
		if rr.Requests[i].Method == "POST" {
			post = &rr.Requests[i]
			break
		}
	}
	if post == nil {
		t.Fatalf("no POST recorded; requests=%+v", rr.Requests)
	}
	if post.Path != "/rest/api/3/issue" {
		t.Errorf("path = %q, want /rest/api/3/issue", post.Path)
	}
	fields := post.Body["fields"].(map[string]interface{})
	parent, ok := fields["parent"].(map[string]interface{})
	if !ok {
		t.Fatalf("parent missing from create body: %+v", fields)
	}
	if parent["key"] != "ED-7281" {
		t.Errorf("parent.key = %v, want ED-7281", parent["key"])
	}
}

func TestIssuesCreate_NoParent(t *testing.T) {
	rr := newRecorder()
	defer rr.Close()

	cmd := &IssuesCreateCmd{
		Project: "ED",
		Type:    "Story",
		Summary: "no parent here",
	}
	_ = captureStdout(t, func() {
		if err := cmd.Run(rr.Client()); err != nil {
			t.Fatalf("Run: %v", err)
		}
	})

	var post *recordedRequest
	for i := range rr.Requests {
		if rr.Requests[i].Method == "POST" {
			post = &rr.Requests[i]
		}
	}
	if post == nil {
		t.Fatalf("no POST recorded")
	}
	fields := post.Body["fields"].(map[string]interface{})
	if _, present := fields["parent"]; present {
		t.Errorf("parent must be omitted when --parent not set; fields=%+v", fields)
	}
}

func TestIssuesCreate_ParentHelpMentionsEpic(t *testing.T) {
	// Help text must crosswalk parent ↔ epic so users searching for "epic" find it.
	cmd := &IssuesCreateCmd{}
	helpText := findKongHelp(t, cmd, "Parent")
	lower := strings.ToLower(helpText)
	if !strings.Contains(lower, "epic") {
		t.Errorf("--parent help should mention 'epic'; got %q", helpText)
	}
}

func runSearchAndCaptureJQL(t *testing.T, cmd *IssuesSearchCmd) string {
	t.Helper()
	rr := newRecorder()
	defer rr.Close()
	rr.Responder = func(rr *requestRecorder, w http.ResponseWriter, r *http.Request, body map[string]interface{}) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issues":[],"total":0,"maxResults":50,"startAt":0}`))
	}
	_ = captureStdout(t, func() {
		if err := cmd.Run(rr.Client()); err != nil {
			t.Fatalf("Run: %v", err)
		}
	})
	for _, req := range rr.Requests {
		if req.Method == "POST" && req.Path == "/rest/api/3/search/jql" {
			jql, _ := req.Body["jql"].(string)
			return jql
		}
	}
	t.Fatalf("no search request recorded; requests=%+v", rr.Requests)
	return ""
}

func TestIssuesSearch_ParentFlagAppendsJQL(t *testing.T) {
	jql := runSearchAndCaptureJQL(t, &IssuesSearchCmd{Parent: "ED-7281", Max: 50})
	if !strings.Contains(jql, "parent = ED-7281") {
		t.Errorf("JQL missing parent clause; got %q", jql)
	}
}

func TestIssuesSearch_ParentCombinesWithStatus(t *testing.T) {
	jql := runSearchAndCaptureJQL(t, &IssuesSearchCmd{
		Parent: "ED-7281",
		Status: "In Progress",
		Max:    50,
	})
	if !strings.Contains(jql, "parent = ED-7281") {
		t.Errorf("JQL missing parent clause; got %q", jql)
	}
	if !strings.Contains(jql, `status = "In Progress"`) {
		t.Errorf("JQL missing status clause; got %q", jql)
	}
	if !strings.Contains(jql, " AND ") {
		t.Errorf("expected AND between clauses; got %q", jql)
	}
}

func TestIssuesSearch_NoParentNoChange(t *testing.T) {
	jql := runSearchAndCaptureJQL(t, &IssuesSearchCmd{Project: "ED", Max: 50})
	if strings.Contains(jql, "parent") {
		t.Errorf("JQL should not include 'parent' when --parent unset; got %q", jql)
	}
	if !strings.Contains(jql, "project = ED") {
		t.Errorf("JQL should still include project filter; got %q", jql)
	}
}

func TestIssuesGet_RendersParent(t *testing.T) {
	rr := newRecorder()
	defer rr.Close()
	rr.Responder = func(rr *requestRecorder, w http.ResponseWriter, r *http.Request, body map[string]interface{}) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"1","key":"ED-12105","self":"x",
			"fields":{
				"summary":"switch contracts offline",
				"parent":{"id":"99","key":"ED-7281","self":"y","fields":{"summary":"SecurED"}}
			}
		}`))
	}

	out := captureStdout(t, func() {
		cmd := &IssuesGetCmd{IssueKey: "ED-12105"}
		if err := cmd.Run(rr.Client()); err != nil {
			t.Fatalf("Run: %v", err)
		}
	})

	if !strings.Contains(out, "Parent: ED-7281") {
		t.Errorf("expected 'Parent: ED-7281' in output, got:\n%s", out)
	}
}

func TestIssuesGet_OmitsParentWhenAbsent(t *testing.T) {
	rr := newRecorder()
	defer rr.Close()
	rr.Responder = func(rr *requestRecorder, w http.ResponseWriter, r *http.Request, body map[string]interface{}) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"1","key":"ED-1","self":"x",
			"fields":{"summary":"no parent"}
		}`))
	}

	out := captureStdout(t, func() {
		cmd := &IssuesGetCmd{IssueKey: "ED-1"}
		if err := cmd.Run(rr.Client()); err != nil {
			t.Fatalf("Run: %v", err)
		}
	})

	if strings.Contains(out, "Parent:") {
		t.Errorf("expected no 'Parent:' line when issue has no parent; got:\n%s", out)
	}
}
