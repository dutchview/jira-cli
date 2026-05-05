package cmd

import (
	"net/http"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
)

func TestParentSet_BulkSendsParentKey(t *testing.T) {
	rr := newRecorder()
	defer rr.Close()

	cmd := &IssuesParentSetCmd{
		Children: []string{"ED-12104", "ED-12105"},
		To:       "ED-7281",
	}
	out := captureStdout(t, func() {
		if err := cmd.Run(rr.Client()); err != nil {
			t.Fatalf("Run: %v", err)
		}
	})

	// Expect one PUT per child.
	puts := []recordedRequest{}
	for _, req := range rr.Requests {
		if req.Method == "PUT" {
			puts = append(puts, req)
		}
	}
	if len(puts) != 2 {
		t.Fatalf("expected 2 PUTs, got %d (requests=%+v)", len(puts), rr.Requests)
	}
	for i, key := range []string{"ED-12104", "ED-12105"} {
		wantPath := "/rest/api/3/issue/" + key
		if puts[i].Path != wantPath {
			t.Errorf("PUT[%d] path = %q, want %q", i, puts[i].Path, wantPath)
		}
		fields, ok := puts[i].Body["fields"].(map[string]interface{})
		if !ok {
			t.Fatalf("PUT[%d] missing fields: %+v", i, puts[i].Body)
		}
		parent, ok := fields["parent"].(map[string]interface{})
		if !ok {
			t.Fatalf("PUT[%d] missing parent: %+v", i, fields)
		}
		if parent["key"] != "ED-7281" {
			t.Errorf("PUT[%d] parent.key = %v, want ED-7281", i, parent["key"])
		}
	}

	// User-facing output should mention each key.
	for _, key := range []string{"ED-12104", "ED-12105", "ED-7281"} {
		if !strings.Contains(out, key) {
			t.Errorf("output missing %q: %s", key, out)
		}
	}
}

func TestParentSet_ContinueOnError(t *testing.T) {
	rr := newRecorder()
	defer rr.Close()

	rr.Responder = func(rr *requestRecorder, w http.ResponseWriter, r *http.Request, body map[string]interface{}) {
		if strings.HasSuffix(r.URL.Path, "ED-12105") && r.Method == "PUT" {
			http.Error(w, `{"errorMessages":["bad child"]}`, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"1","key":"ED-1","self":"x","fields":{"summary":"s"}}`))
	}

	cmd := &IssuesParentSetCmd{
		Children: []string{"ED-12104", "ED-12105", "ED-12106"},
		To:       "ED-7281",
	}
	var runErr error
	out := captureStdout(t, func() {
		runErr = cmd.Run(rr.Client())
	})

	// All three children should have been attempted.
	puts := 0
	for _, req := range rr.Requests {
		if req.Method == "PUT" {
			puts++
		}
	}
	if puts != 3 {
		t.Errorf("expected 3 PUTs (continue on error), got %d", puts)
	}

	// Error must be returned (so process exits non-zero).
	if runErr == nil {
		t.Errorf("expected non-nil error after partial failure")
	}

	// Output must mention success for the first and third, failure for the second.
	if !strings.Contains(out, "ED-12104") || !strings.Contains(out, "ED-12106") {
		t.Errorf("output missing successful keys: %s", out)
	}
	if !strings.Contains(out, "ED-12105") {
		t.Errorf("output missing failed key: %s", out)
	}
	lower := strings.ToLower(out)
	if !strings.Contains(lower, "fail") && !strings.Contains(lower, "error") {
		t.Errorf("output should mark the failure: %s", out)
	}
}

func TestParentClear_BulkSendsNullParent(t *testing.T) {
	rr := newRecorder()
	defer rr.Close()

	cmd := &IssuesParentClearCmd{
		Children: []string{"ED-12104", "ED-12105"},
	}
	_ = captureStdout(t, func() {
		if err := cmd.Run(rr.Client()); err != nil {
			t.Fatalf("Run: %v", err)
		}
	})

	puts := []recordedRequest{}
	for _, req := range rr.Requests {
		if req.Method == "PUT" {
			puts = append(puts, req)
		}
	}
	if len(puts) != 2 {
		t.Fatalf("expected 2 PUTs, got %d", len(puts))
	}
	for i, p := range puts {
		fields, ok := p.Body["fields"].(map[string]interface{})
		if !ok {
			t.Fatalf("PUT[%d] missing fields: %+v", i, p.Body)
		}
		if _, present := fields["parent"]; !present {
			t.Fatalf("PUT[%d] parent key missing from request body", i)
		}
		if fields["parent"] != nil {
			t.Errorf("PUT[%d] parent = %#v, want JSON null", i, fields["parent"])
		}
	}
}

func TestParentSet_RequiresTo(t *testing.T) {
	var cli struct {
		Issues IssuesCmd `cmd:""`
	}
	parser, err := kong.New(&cli)
	if err != nil {
		t.Fatalf("kong.New: %v", err)
	}
	if _, err := parser.Parse([]string{"issues", "parent", "set", "ED-1"}); err == nil {
		t.Errorf("expected error when --to is missing on `issues parent set`")
	}
	if _, err := parser.Parse([]string{"issues", "parent", "set", "ED-1", "--to", "ED-7281"}); err != nil {
		t.Errorf("expected --to to satisfy required flag, got %v", err)
	}
}

func TestParentClear_RejectsTo(t *testing.T) {
	var cli struct {
		Issues IssuesCmd `cmd:""`
	}
	parser, err := kong.New(&cli)
	if err != nil {
		t.Fatalf("kong.New: %v", err)
	}
	if _, err := parser.Parse([]string{"issues", "parent", "clear", "ED-1", "--to", "X"}); err == nil {
		t.Errorf("expected error: --to is not valid on `issues parent clear`")
	}
}
