package filejudge

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCheckerAndDisplay(t *testing.T) {
	for _, tc := range []struct {
		checker, a, b string
		want          bool
	}{
		{"tokens-v1", "  3\n", "3", true}, {"exact-v1", "3\n", "3", false}, {"tokens-v1", "1\u20032", "1 2", true}, {"tokens-v1", "03", "3", false}, {"tokens-v1", "3 4", "3", false}, {"exact-v1", string([]byte{255}), string([]byte{255}), true},
	} {
		if got := check(tc.checker, []byte(tc.a), []byte(tc.b)); got != tc.want {
			t.Fatalf("%+v: %v", tc, got)
		}
	}
	s, truncated := display([]byte("你你"), 4)
	if s != "你" || !truncated {
		t.Fatal(s, truncated)
	}
	s, truncated = display(bytes.Repeat([]byte{255, 'a'}, 65536), 65536)
	if len(s) > 65536 || !truncated {
		t.Fatal("invalid UTF8 display unbounded")
	}
	actual := append(bytes.Repeat([]byte(" "), 70000), []byte("3\n")...)
	s, truncated = display(actual, 65536)
	if !truncated || len(s) != 65536 || !check("tokens-v1", actual, []byte("3")) {
		t.Fatal("checker used truncated display")
	}
}
func testResult() Result {
	code := 0
	return Result{RunID: strings.Repeat("a", 32), Verdict: "AC", Message: message("AC"), SourceSHA256: strings.Repeat("a", 64), TestdataSHA256: strings.Repeat("b", 64), ArtifactSHA256: strings.Repeat("c", 64), RuntimeImage: "sha256:" + strings.Repeat("d", 64), Compiler: "GCC 14.2.0", Standard: "c++20", Compile: CompileResult{Status: "success"}, Cases: []CaseResult{{ID: "sample", Verdict: "AC", ExitCode: &code, Stdout: "3\n"}}}
}
func TestClientProtocol(t *testing.T) {
	for _, name := range []string{"valid", "wrong_source", "bad_enum", "missing_required", "null_required", "bad_peak", "null_AC_exit", "nonzero_AC_exit", "AC_oom", "success_CE", "cleanup_overrides_CE", "oversize", "busy", "unavailable", "cancelled"} {
		t.Run(name, func(t *testing.T) {
			r := testResult()
			r.SourceSHA256 = hash([]byte("int main(){}"))
			if name == "wrong_source" {
				r.SourceSHA256 = strings.Repeat("f", 64)
			}
			if name == "null_AC_exit" {
				r.Cases[0].ExitCode = nil
			}
			if name == "nonzero_AC_exit" {
				code := 1
				r.Cases[0].ExitCode = &code
			}
			if name == "AC_oom" {
				r.Cases[0].OOMKilled = true
			}
			if name == "success_CE" {
				r.Verdict = "CE"
			}
			if name == "cleanup_overrides_CE" {
				r.Verdict = "SYSTEM_ERROR"
				r.Compile.Status = "compile_error"
				r.ArtifactSHA256 = ""
				r.Cases = []CaseResult{}
			}
			if name == "bad_enum" {
				r.Cases[0].Verdict = "UNKNOWN"
			}
			if name == "bad_peak" {
				n := int64(1)
				r.Cases[0].PeakMemoryBytes = &n
			}
			data, _ := json.Marshal(r)
			if name == "missing_required" {
				data = bytes.Replace(data, []byte(`"oom_killed":false,`), nil, 1)
			}
			if name == "null_required" {
				data = bytes.Replace(data, []byte(`"duration_ms":0`), []byte(`"duration_ms":null`), 1)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				switch name {
				case "busy":
					w.WriteHeader(429)
				case "unavailable":
					w.WriteHeader(503)
				case "oversize":
					_, _ = w.Write(bytes.Repeat([]byte{'x'}, MaxResponse+1))
				default:
					_, _ = w.Write(data)
				}
			}))
			defer server.Close()
			client, err := NewClient(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			if name == "cancelled" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			_, failure := client.Judge(ctx, Request{ProblemID: "a-plus-b", Language: "cpp", Source: "int main(){}"})
			want := 502
			switch name {
			case "valid", "cleanup_overrides_CE":
				want = 0
			case "busy":
				want = 429
			case "unavailable", "cancelled":
				want = 503
			}
			if want == 0 && failure != nil {
				t.Fatal(failure)
			}
			if want != 0 && (failure == nil || failure.Status != want) {
				t.Fatalf("want %d got %v", want, failure)
			}
		})
	}
}
func TestVerdictEvidence(t *testing.T) {
	for _, tc := range []struct {
		s    stageResult
		want string
	}{
		{stageResult{State: containerState{ExitCode: 137}}, "RE"},
		{stageResult{State: containerState{ExitCode: 137, OOMKilled: true}, Cause: "wall_timeout"}, "MLE"},
		{stageResult{State: containerState{ExitCode: 137}, Cause: "wall_timeout"}, "TLE"},
		{stageResult{Cause: "output_limit"}, "OLE"},
		{stageResult{Failure: "cleanup_failed", State: containerState{OOMKilled: true}}, "SYSTEM_ERROR"},
	} {
		if got := caseVerdict(tc.s, "tokens-v1", []byte("3"), []byte("3")); got != tc.want {
			t.Fatal(tc.want, got)
		}
	}
}

func TestClientMaximumEscapedSnapshots(t *testing.T) {
	for _, kind := range []string{"batch", "interactive"} {
		t.Run(kind, func(t *testing.T) {
			request := Request{ProblemID: "a-plus-b", Language: "cpp", Source: "int main(){}"}
			r := testResult()
			r.SourceSHA256 = hash([]byte(request.Source))
			r.Compile.Diagnostics = strings.Repeat("\x00", MaxFile)
			if kind == "interactive" {
				request.ProblemID = "guess-number"
				ic := r.Compile
				r.InteractorCompile = &ic
				r.InteractorSourceSHA256 = InteractorSourceDigest()
				r.InteractorArtifactSHA256 = strings.Repeat("e", 64)
			}
			zero := 0
			r.Cases = nil
			for _, id := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
				row := CaseResult{ID: id, Verdict: "AC", ExitCode: &zero, Input: strings.Repeat("\x00", MaxFile), Expected: strings.Repeat("\x00", MaxFile), Stdout: strings.Repeat("\x00", MaxFile), Stderr: strings.Repeat("\x00", 16384)}
				if kind == "interactive" {
					row.Interactor = &InteractorResult{ExitCode: &zero, Stdout: strings.Repeat("\x00", MaxFile), Stderr: strings.Repeat("\x00", 16384)}
				}
				r.Cases = append(r.Cases, row)
			}
			data, err := json.Marshal(r)
			if err != nil || len(data) <= 6*1048576 || len(data) > MaxResponse {
				t.Fatalf("escaped bound: bytes=%d error=%v", len(data), err)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(data) }))
			defer server.Close()
			client, err := NewClient(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			result, failure := client.Judge(context.Background(), request)
			if failure != nil || len(result.Cases) != MaxCases || result.Cases[7].Expected != r.Cases[7].Expected {
				t.Fatalf("valid escaped snapshots rejected: %v", failure)
			}
		})
	}
}
