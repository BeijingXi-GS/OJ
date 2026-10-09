package compilerworker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"oj.local/oj/internal/compilation"
)

// A separate OS process exercises output drainage, cancellation and cleanup.
func TestCompilerHelper(t *testing.T) {
	mode := os.Getenv("OJ_COMPILER_TEST_HELPER")
	if mode == "" {
		return
	}
	switch mode {
	case "success":
		_ = os.WriteFile("program", []byte("must never execute"), 0600)
	case "error":
		fmt.Fprint(os.Stderr, "solution.cpp: syntax error")
		os.Exit(1)
	case "large":
		fmt.Fprint(os.Stderr, strings.Repeat("中", compilation.MaxDiagnosticsBytes))
		os.Exit(1)
	case "slow":
		time.Sleep(10 * time.Second)
	}
	os.Exit(0)
}

func helperRunner(t *testing.T, mode string) *Runner {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	r := New(exe, "test compiler")
	r.TempRoot = t.TempDir()
	r.command = func(ctx context.Context, _ string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, exe, "-test.run=^TestCompilerHelper$")
		cmd.Env = append(os.Environ(), "OJ_COMPILER_TEST_HELPER="+mode)
		return cmd
	}
	return r
}

func TestCompileLifecycle(t *testing.T) {
	for _, tc := range []struct{ mode, status string }{{"success", "success"}, {"error", "compile_error"}, {"large", "compile_error"}, {"slow", "timeout"}} {
		t.Run(tc.mode, func(t *testing.T) {
			r := helperRunner(t, tc.mode)
			if tc.mode == "slow" {
				r.Timeout = 100 * time.Millisecond
			}
			result, failure := r.Compile(context.Background(), "int main(){}")
			if failure != nil || result.Status != tc.status {
				t.Fatalf("result=%+v error=%v", result, failure)
			}
			if tc.mode == "large" && (!result.DiagnosticsTruncated || len(result.Diagnostics) > compilation.MaxDiagnosticsBytes) {
				t.Fatal("output limit not enforced")
			}
			if tc.mode == "slow" && result.DurationMS > 3000 {
				t.Fatal("timeout not bounded")
			}
			entries, err := os.ReadDir(r.TempRoot)
			if err != nil || len(entries) != 0 {
				t.Fatal("temporary files remain")
			}
		})
	}
}

func TestBusyAndCompilerFailure(t *testing.T) {
	r := helperRunner(t, "success")
	r.busy <- struct{}{}
	if _, err := r.Compile(context.Background(), "x"); err == nil || err.Status != 429 {
		t.Fatal("did not reject concurrent compile")
	}
	<-r.busy
	r.command = nil
	r.CompilerPath = "missing-test-compiler"
	if _, err := r.Compile(context.Background(), "x"); err == nil || err.Status != 502 {
		t.Fatal("launch failure misreported as compile error")
	}
	entries, _ := os.ReadDir(r.TempRoot)
	if len(entries) != 0 {
		t.Fatal("failed launch left files")
	}
}
