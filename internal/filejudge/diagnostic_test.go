package filejudge

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStartupIOErrorDoesNotExposeHostPath(t *testing.T) {
	dir := t.TempDir()
	profilePath := filepath.Join(dir, "profile.json")
	profile, err := json.Marshal(Profile{Version: "judge-mvp/0.1.0", Image: "sha256:" + strings.Repeat("a", 64), ReferenceCommit: "4502d2a798f3130d3c540c5c31d45251ca103623"})
	if err != nil || os.WriteFile(profilePath, profile, 0600) != nil {
		t.Fatal("test profile setup failed")
	}
	privatePath := filepath.Join(dir, "PRIVATE_PATH")
	if err := os.WriteFile(privatePath, []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, jobs := range []string{privatePath, filepath.Join(privatePath, "jobs")} {
		_, err := NewEngine(profilePath, dir, jobs)
		if err == nil || err.Error() != "jobs_root_create_failed" || strings.Contains(err.Error(), dir) {
			t.Fatalf("unsafe startup diagnostic: %v", err)
		}
	}
}

func TestControlErrorClassifiedAndSanitized(t *testing.T) {
	for _, name := range []string{"cancelled", "deadline", "deadline_success_nil", "stdout_limit", "stderr_limit", "missing_cli", "io_error"} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if name == "deadline" || name == "deadline_success_nil" {
				var done context.CancelFunc
				ctx, done = context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
				defer done()
			}
			e := &Engine{command: func(_ context.Context, _ []string, out, stderr io.Writer) error {
				switch name {
				case "cancelled":
					cancel()
					return context.Canceled
				case "deadline":
					return context.DeadlineExceeded
				case "deadline_success_nil":
					_, _ = io.WriteString(out, "2")
					return nil
				case "stdout_limit":
					_, _ = io.WriteString(out, strings.Repeat("x", 65537))
				case "stderr_limit":
					_, _ = io.WriteString(stderr, strings.Repeat("credential=SECRET_PRIVATE_PATH", 1000))
				case "missing_cli":
					return &exec.Error{Name: "SECRET_PRIVATE_PATH", Err: errors.New("credential=SECRET")}
				default:
					return errors.New("SECRET_PRIVATE_PATH")
				}
				return nil
			}}
			_, err := e.control(ctx, "info", "--format", "{{.CgroupVersion}}")
			if err == nil || strings.Contains(err.Error(), "SECRET") || strings.Contains(err.Error(), "credential") || strings.Contains(err.Error(), "PRIVATE_PATH") {
				t.Fatal("unsafe or absent diagnostic", err)
			}
			kind := "io_error"
			switch name {
			case "cancelled":
				kind = "cancelled"
			case "deadline", "deadline_success_nil":
				kind = "timeout"
			case "stdout_limit", "stderr_limit":
				kind = "output_limit"
			case "missing_cli":
				kind = "start_failed"
			}
			if err.Error() != "docker_info_"+kind {
				t.Fatal(err)
			}
		})
	}
}
func TestCgroupQueryFailureDiffersFromUnsupportedVersion(t *testing.T) {
	for _, queryFails := range []bool{true, false} {
		e := &Engine{command: func(_ context.Context, args []string, out, _ io.Writer) error {
			if args[0] == "version" {
				_, _ = io.WriteString(out, "linux amd64")
				return nil
			}
			if queryFails {
				return errors.New("private daemon error")
			}
			_, _ = io.WriteString(out, "1")
			return nil
		}}
		err := e.Preflight(context.Background())
		if err == nil {
			t.Fatal("preflight accepted missing isolation")
		}
		if queryFails && !strings.HasPrefix(err.Error(), "docker_cgroup_query_failed: docker_info_io_error") {
			t.Fatal(err)
		}
		if !queryFails && err.Error() != "cgroup_v2_required" {
			t.Fatal(err)
		}
	}
}
