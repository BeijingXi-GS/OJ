package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadiness(t *testing.T) {
	for _, tc := range []struct {
		name string
		ping PingFunc
		code int
		body string
	}{
		{"unconfigured", nil, 503, "database_not_configured"},
		{"unavailable", func(context.Context) error { return errors.New("secret connection details") }, 503, "database_unavailable"},
		{"ready", func(ctx context.Context) error {
			if _, ok := ctx.Deadline(); !ok {
				t.Error("missing deadline")
			}
			return nil
		}, 200, "ready"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			New(tc.ping).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/readyz", nil))
			if w.Code != tc.code || !strings.Contains(w.Body.String(), tc.body) {
				t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
			}
			if strings.Contains(w.Body.String(), "secret") {
				t.Fatal("internal details leaked")
			}
		})
	}
}

func TestScaffoldDoesNotAcceptSubmissions(t *testing.T) {
	w := httptest.NewRecorder()
	New(nil).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/submissions", strings.NewReader("{}")))
	if w.Code != http.StatusNotFound {
		t.Fatalf("unexpected status %d", w.Code)
	}
}
