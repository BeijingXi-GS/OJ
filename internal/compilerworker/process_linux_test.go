//go:build linux

package compilerworker

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestTimeoutKillsCompilerChildren(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "child.pid")
	r := New("/bin/sh", "test process group")
	r.TempRoot = t.TempDir()
	r.Timeout = 300 * time.Millisecond
	r.command = func(ctx context.Context, _ string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, "/bin/sh", "-c", `sleep 30 & child=$!; echo "$child" > "$PID_FILE"; wait`)
		cmd.Env = append(os.Environ(), "PID_FILE="+pidFile)
		return cmd
	}
	result, failure := r.Compile(context.Background(), "x")
	if failure != nil || result.Status != "timeout" {
		t.Fatalf("%+v %v", result, failure)
	}
	data, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 1 {
		t.Fatal("invalid child PID")
	}
	for i := 0; i < 50; i++ {
		if errors.Is(syscall.Kill(pid, 0), syscall.ESRCH) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("compiler child remains after timeout; run container tests with --init")
}
