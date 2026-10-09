package compilation

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientProtocol(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   int
	}{
		{"success", 200, `{"status":"success","diagnostics":"","duration_ms":1,"compiler":"GCC 14.2.0","standard":"c++20"}`, 0},
		{"busy", 429, "private secret", 429}, {"unavailable", 503, "private secret", 503}, {"failure", 500, "private secret", 502},
		{"bad status", 200, `{"status":"accepted"}`, 502},
		{"bad JSON", 200, "not json", 502}, {"oversized", 200, strings.Repeat("x", 8*MaxDiagnosticsBytes+1), 502},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body map[string]string
				_ = json.NewDecoder(r.Body).Decode(&body)
				if r.Method != "POST" || r.URL.Path != "/compile" || body["source"] != "original" {
					t.Error("request changed")
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer s.Close()
			c, err := NewClient(s.URL)
			if err != nil {
				t.Fatal(err)
			}
			_, failure := c.Compile(context.Background(), "original")
			if tc.want == 0 {
				if failure != nil {
					t.Fatal(failure)
				}
				return
			}
			if failure == nil || failure.Status != tc.want || strings.Contains(failure.Message, "private") {
				t.Fatalf("unexpected failure: %#v", failure)
			}
		})
	}
}

func TestClientCancellationAndRedirect(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/other", 307) }))
	defer s.Close()
	c, _ := NewClient(s.URL)
	if _, err := c.Compile(context.Background(), "x"); err == nil || err.Status != 502 {
		t.Fatal("followed redirect")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.Compile(ctx, "x"); err == nil || err.Status != 503 {
		t.Fatal("lost cancellation")
	}
}

func TestRejectUserLikeURLConfiguration(t *testing.T) {
	for _, input := range []string{"ftp://localhost", "http://user:secret@localhost", "http://localhost/path", "http://localhost?q=x", "invalid"} {
		if _, err := NewClient(input); err == nil {
			t.Fatalf("accepted %q", input)
		}
	}
}
