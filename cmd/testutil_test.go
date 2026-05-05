package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/dutchview/jira-cli/internal/api"
	"github.com/dutchview/jira-cli/internal/config"
)

// findKongHelp returns the `help:"..."` struct tag for the named field on v.
// Used to assert that user-facing help text says what we expect.
func findKongHelp(t *testing.T, v interface{}, fieldName string) string {
	t.Helper()
	rt := reflect.TypeOf(v)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	f, ok := rt.FieldByName(fieldName)
	if !ok {
		t.Fatalf("no field named %q on %s", fieldName, rt.Name())
	}
	return f.Tag.Get("help")
}

// recordedRequest captures one inbound request for assertions.
type recordedRequest struct {
	Method string
	Path   string
	Body   map[string]interface{}
}

// requestRecorder spins up an httptest.Server that records every request and
// answers via the given responder (or a default 200 with stub JSON).
type requestRecorder struct {
	mu        sync.Mutex
	Requests  []recordedRequest
	Server    *httptest.Server
	Responder func(rr *requestRecorder, w http.ResponseWriter, r *http.Request, body map[string]interface{})
}

func newRecorder() *requestRecorder {
	rr := &requestRecorder{}
	rr.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf, _ := io.ReadAll(r.Body)
		var body map[string]interface{}
		if len(buf) > 0 {
			_ = json.Unmarshal(buf, &body)
		}
		rr.mu.Lock()
		rr.Requests = append(rr.Requests, recordedRequest{
			Method: r.Method,
			Path:   r.URL.Path,
			Body:   body,
		})
		rr.mu.Unlock()
		if rr.Responder != nil {
			rr.Responder(rr, w, r, body)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"1","key":"ED-1","self":"x","fields":{"summary":"s"}}`)
	}))
	return rr
}

func (rr *requestRecorder) Close() { rr.Server.Close() }

func (rr *requestRecorder) Client() *api.Client {
	return api.NewClient(&config.Config{
		BaseURL:  rr.Server.URL,
		Email:    "t@e.com",
		APIToken: "tok",
	})
}

// captureStdout swaps os.Stdout for the duration of fn and returns what was written.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	done := make(chan []byte, 1)
	go func() {
		buf, _ := io.ReadAll(r)
		done <- buf
	}()

	fn()
	_ = w.Close()
	os.Stdout = orig
	return string(<-done)
}
