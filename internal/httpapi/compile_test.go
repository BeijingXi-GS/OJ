package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"oj.local/oj/internal/compilation"
)

type compilerFunc func(context.Context, string) (compilation.Result, *compilation.Error)

func (f compilerFunc) Compile(ctx context.Context, source string) (compilation.Result, *compilation.Error) {
	return f(ctx, source)
}

func TestProblems(t *testing.T) {
	for _, tc := range []struct {
		path    string
		code    int
		content string
	}{
		{"/api/v1/problems", 200, "a-plus-b"}, {"/api/v1/problems/a-plus-b", 200, "starter_code"}, {"/api/v1/problems/missing", 404, "problem_not_found"},
	} {
		w := httptest.NewRecorder()
		New(nil).ServeHTTP(w, httptest.NewRequest("GET", tc.path, nil))
		if w.Code != tc.code || !strings.Contains(w.Body.String(), tc.content) {
			t.Fatalf("%s: %d %s", tc.path, w.Code, w.Body.String())
		}
	}
}

func TestCompilationValidation(t *testing.T) {
	longSource, _ := json.Marshal(map[string]string{"problem_id": "a-plus-b", "language": "cpp", "source": strings.Repeat("中", 22000)})
	for _, tc := range []struct {
		name, body, contentType string
		code                    int
	}{
		{"malformed", "{", "application/json", 400},
		{"extra JSON", `{} {}`, "application/json", 400},
		{"unknown field", `{"flags":"--help"}`, "application/json", 400},
		{"wrong language", `{"problem_id":"a-plus-b","language":"python","source":"x"}`, "application/json", 400},
		{"no problem", `{"language":"cpp","source":"x"}`, "application/json", 400},
		{"missing problem", `{"problem_id":"missing","language":"cpp","source":"x"}`, "application/json", 404},
		{"blank", `{"problem_id":"a-plus-b","language":"cpp","source":"   "}`, "application/json", 400},
		{"UTF8 byte limit", string(longSource), "application/json", 413},
		{"body limit", `{"source":"` + strings.Repeat("x", compilation.MaxBodyBytes) + `"}`, "application/json", 413},
		{"content type", `{}`, "text/plain", 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			h := NewWithCompiler(nil, compilerFunc(func(context.Context, string) (compilation.Result, *compilation.Error) {
				called = true
				return compilation.Result{}, nil
			}))
			r := httptest.NewRequest("POST", "/api/v1/compilations", strings.NewReader(tc.body))
			r.Header.Set("Content-Type", tc.contentType)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != tc.code || called {
				t.Fatalf("%d %s, compiler called=%v", w.Code, w.Body.String(), called)
			}
		})
	}
}

func TestCompilationForwardsResultsAndErrors(t *testing.T) {
	for _, status := range []string{"success", "compile_error", "timeout"} {
		t.Run(status, func(t *testing.T) {
			h := NewWithCompiler(nil, compilerFunc(func(_ context.Context, source string) (compilation.Result, *compilation.Error) {
				if source != "int main(){}" {
					t.Error("source changed")
				}
				return compilation.Result{Status: status, Diagnostics: "<script>compiler text</script>", Compiler: "GCC 14.2.0", Standard: "c++20"}, nil
			}))
			w := compileRequest(h)
			var result compilation.Result
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || w.Code != 200 || result.Status != status || result.Diagnostics != "<script>compiler text</script>" {
				t.Fatalf("%d %s", w.Code, w.Body.String())
			}
		})
	}
	for _, code := range []int{429, 502, 503} {
		h := NewWithCompiler(nil, compilerFunc(func(context.Context, string) (compilation.Result, *compilation.Error) {
			return compilation.Result{}, &compilation.Error{Status: code, Code: "worker_error", Message: "稍后重试"}
		}))
		if w := compileRequest(h); w.Code != code {
			t.Fatalf("got %d want %d", w.Code, code)
		}
	}
	if w := compileRequest(New(nil)); w.Code != 503 {
		t.Fatalf("nil compiler: %d", w.Code)
	}
}

func compileRequest(h http.Handler) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", "/api/v1/compilations", strings.NewReader(`{"problem_id":"a-plus-b","language":"cpp","source":"int main(){}"}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
