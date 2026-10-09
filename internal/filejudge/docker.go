package filejudge

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const ownerLabel = "oj.judge.owner=file-mvp-v1"

type commandFunc func(context.Context, []string, io.Writer, io.Writer) error
type duplexCommandFunc func(context.Context, []string, io.ReadCloser, io.Writer, io.Writer) error

func dockerCommand(ctx context.Context, args []string, out, errOut io.Writer) error {
	c := exec.CommandContext(ctx, "docker", args...)
	c.Stdout = out
	c.Stderr = errOut
	c.WaitDelay = 2 * time.Second
	return c.Run()
}

// Own the stdin copy so a child which has exited cannot leave Cmd.Wait blocked
// on a still-open protocol pipe. The caller also closes all pipes on cancellation.
func dockerDuplexCommand(ctx context.Context, args []string, input io.ReadCloser, out, errOut io.Writer) error {
	c := exec.CommandContext(ctx, "docker", args...)
	return runDuplexCommand(ctx, c, input, out, errOut)
}

type duplexInputError struct {
	pipeClosed bool
	attachErr  error
}

func (*duplexInputError) Error() string { return "interactive_stdin_transfer_failed" }

func expectedPipeClosure(err error) bool {
	return errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrClosed) || errors.Is(err, syscall.EPIPE) || (runtime.GOOS == "windows" && (errors.Is(err, syscall.Errno(109)) || errors.Is(err, syscall.Errno(232))))
}

func runDuplexCommand(ctx context.Context, c *exec.Cmd, input io.ReadCloser, out, errOut io.Writer) error {
	c.Stdout, c.Stderr = out, errOut
	c.WaitDelay = 2 * time.Second
	stdin, err := c.StdinPipe()
	if err != nil {
		_ = input.Close()
		return err
	}
	copied := make(chan error, 1)
	var closing atomic.Bool
	go func() {
		_, copyErr := io.Copy(stdin, input)
		// Capture this fact at the failure, before the main goroutine begins
		// its final close. An unrelated read/write error is never exempted.
		if copyErr != nil && !(expectedPipeClosure(copyErr) && (closing.Load() || ctx.Err() != nil)) {
			copyErr = &duplexInputError{pipeClosed: expectedPipeClosure(copyErr)}
		} else {
			copyErr = nil
		}
		_ = stdin.Close()
		copied <- copyErr
	}()
	err = c.Run()
	closing.Store(true)
	_ = input.Close()
	_ = stdin.Close()
	if copyErr := <-copied; copyErr != nil {
		var transfer *duplexInputError
		if errors.As(copyErr, &transfer) {
			transfer.attachErr = err
		}
		return copyErr
	}
	return err
}

type cappedBuffer struct {
	data     bytes.Buffer
	limit    int
	overflow bool
}

func (b *cappedBuffer) Len() int      { return b.data.Len() }
func (b *cappedBuffer) Bytes() []byte { return b.data.Bytes() }

func (b *cappedBuffer) Write(p []byte) (int, error) {
	n := len(p)
	keep := min(n, max(0, b.limit-b.Len()))
	_, _ = b.data.Write(p[:keep])
	if keep < n {
		b.overflow = true
	}
	return n, nil
}

type containerState struct {
	Running   bool
	OOMKilled bool
	ExitCode  int
	Error     string
	StartedAt string
}
type containerInfo struct {
	ID     string
	Image  string
	State  containerState
	Config struct {
		User           string
		Labels         map[string]string
		OpenStdin, Tty bool
	}
	HostConfig struct {
		Privileged, ReadonlyRootfs                  bool
		NetworkMode, PidMode, IpcMode, CgroupnsMode string
		Memory, MemorySwap, NanoCpus                int64
		PidsLimit                                   int64
		CapAdd, CapDrop, SecurityOpt                []string
		Init                                        *bool
		LogConfig                                   struct{ Type string }
		Tmpfs                                       map[string]string
	}
	Mounts []struct {
		Type, Source, Destination string
		RW                        bool
	}
}
type stageSpec struct {
	RunID, Name, Stage string
	Memory             int64
	Pids               int
	Wall               time.Duration
	Compile            bool
	Interactive        bool
	Mounts             map[string]string // container destination -> generated host snapshot
	Entry              string
	Args               []string
}
type stageResult struct {
	State      containerState
	DurationMS int64
	Cause      string
	Failure    string
}

// Control errors carry only fixed operation/classification, never raw daemon
// output, argv (which can contain host paths), or inherited environment values.
type controlError struct {
	Operation string
	Kind      string
	ExitCode  int
}

func (e *controlError) Error() string {
	if e.Kind == "exit" {
		return fmt.Sprintf("docker_%s_exit_%d", e.Operation, e.ExitCode)
	}
	return "docker_" + e.Operation + "_" + e.Kind
}

func (e *Engine) control(parent context.Context, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	out, errOut := &cappedBuffer{limit: 65536}, &cappedBuffer{limit: 16384}
	err := e.command(ctx, args, out, errOut)
	if err != nil || ctx.Err() != nil || out.overflow || errOut.overflow {
		op := "unknown"
		if len(args) > 0 {
			switch args[0] {
			case "version", "info", "image", "ps", "inspect", "kill", "rm", "create":
				op = args[0]
			}
		}
		failure := &controlError{Operation: op, Kind: "io_error", ExitCode: -1}
		var exit *exec.ExitError
		var start *exec.Error
		switch {
		case ctx.Err() == context.DeadlineExceeded:
			failure.Kind = "timeout"
		case ctx.Err() == context.Canceled:
			failure.Kind = "cancelled"
		case out.overflow || errOut.overflow:
			failure.Kind = "output_limit"
		case errors.As(err, &exit):
			failure.Kind = "exit"
			failure.ExitCode = exit.ExitCode()
		case errors.As(err, &start):
			failure.Kind = "start_failed"
		case errors.Is(err, exec.ErrWaitDelay):
			failure.Kind = "pipe_timeout"
		}
		return nil, failure
	}
	return out.Bytes(), nil
}
func mountArg(source, dest string) string {
	var b strings.Builder
	w := csv.NewWriter(&b)
	_ = w.Write([]string{"type=bind", "src=" + source, "dst=" + dest, "readonly"})
	w.Flush()
	return strings.TrimSuffix(b.String(), "\n")
}
func (e *Engine) createArgs(s stageSpec) []string {
	a := []string{"create", "--name", s.Name, "--label", ownerLabel, "--label", "oj.judge.run=" + s.RunID, "--label", "oj.judge.stage=" + s.Stage, "--init", "--network", "none", "--ipc", "private", "--cgroupns", "private", "--cap-drop", "ALL", "--security-opt", "no-new-privileges", "--read-only", "--user", "10002:10002", "--log-driver", "none", "--memory", strconv.FormatInt(s.Memory, 10), "--memory-swap", strconv.FormatInt(s.Memory, 10), "--pids-limit", strconv.Itoa(s.Pids), "--cpus", "1", "--shm-size", "8m", "--ulimit", "fsize=16777216:16777216", "--ulimit", "nofile=64:64", "--ulimit", "core=0:0"}
	tmp := "/tmp:rw,nosuid,nodev,noexec,size=16777216"
	if s.Compile {
		for i := range a {
			if a[i] == "fsize=16777216:16777216" {
				a[i] = "fsize=33554432:33554432"
			}
		}
		tmp = "/tmp:rw,nosuid,nodev,size=16777216"
		a = append(a, "--tmpfs", "/work:rw,nosuid,nodev,exec,size=134217728")
	}
	a = append(a, "--tmpfs", tmp)
	if s.Interactive {
		a = append(a, "--interactive")
	}
	// There are only fixed destination names. Each value is CSV-encoded as one argv.
	for _, dest := range []string{"/source/solution.cpp", "/solution/program", "/input.txt", "/checker/program", "/data/input.txt", "/data/answer.txt", "/data/output.txt"} {
		if src, ok := s.Mounts[dest]; ok {
			a = append(a, "--mount", mountArg(src, dest))
		}
	}
	a = append(a, "--entrypoint", s.Entry, e.image)
	return append(a, s.Args...)
}
func (e *Engine) inspect(ctx context.Context, name string) (containerInfo, error) {
	var c containerInfo
	b, err := e.control(ctx, "inspect", name)
	if err != nil {
		return c, err
	}
	var items []containerInfo
	if json.Unmarshal(b, &items) != nil || len(items) != 1 {
		return c, errors.New("inspect_protocol")
	}
	return items[0], nil
}
func (e *Engine) verifyContainer(c containerInfo, s stageSpec) error {
	h := c.HostConfig
	if c.Image != e.image || c.Config.User != "10002:10002" || c.Config.Labels["oj.judge.run"] != s.RunID || c.Config.Labels["oj.judge.owner"] != "file-mvp-v1" || h.Privileged || !h.ReadonlyRootfs || h.NetworkMode != "none" || h.PidMode != "" || h.IpcMode != "private" || h.CgroupnsMode != "private" || h.Memory != s.Memory || h.MemorySwap != s.Memory || h.PidsLimit != int64(s.Pids) || h.NanoCpus != 1000000000 || len(h.CapAdd) != 0 || len(h.CapDrop) != 1 || strings.ToLower(h.CapDrop[0]) != "all" || h.Init == nil || !*h.Init || h.LogConfig.Type != "none" {
		return errors.New("container_profile_mismatch")
	}
	nnp := false
	for _, v := range h.SecurityOpt {
		if v == "no-new-privileges" || v == "no-new-privileges=true" {
			nnp = true
		} else {
			return errors.New("security_option_unexpected")
		}
	}
	if !nnp {
		return errors.New("no_new_privileges_missing")
	}
	if c.Config.Tty || c.Config.OpenStdin != s.Interactive {
		return errors.New("stdin_profile_mismatch")
	}
	seen := map[string]bool{}
	for _, m := range c.Mounts {
		if m.Type == "bind" {
			if _, ok := s.Mounts[m.Destination]; !ok || m.RW || seen[m.Destination] {
				return errors.New("mount_unexpected")
			}
			seen[m.Destination] = true
		} else if m.Type != "tmpfs" {
			return errors.New("mount_unexpected")
		}
	}
	if len(seen) != len(s.Mounts) {
		return errors.New("mount_missing")
	}
	return nil
}
func (e *Engine) matching(ctx context.Context, s stageSpec) ([]string, error) {
	b, err := e.control(ctx, "ps", "-aq", "--filter", "label="+ownerLabel, "--filter", "label=oj.judge.run="+s.RunID, "--filter", "name=^/"+s.Name+"$")
	if err != nil {
		return nil, err
	}
	return strings.Fields(string(b)), nil
}

// Cleanup is independent of the caller/job deadline, but has its own total budget.
func (e *Engine) cleanup(s stageSpec) (containerState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return e.cleanupWithin(ctx, s)
}
func (e *Engine) cleanupWithin(ctx context.Context, s stageSpec) (containerState, error) {
	state, _, err := e.cleanupWithEvidence(ctx, s)
	return state, err
}

func (e *Engine) cleanupWithEvidence(ctx context.Context, s stageSpec) (containerState, bool, error) {
	state, killed, _, err := e.cleanupWithObservation(ctx, s)
	return state, killed, err
}

// Absence proves no matching residue; it does not prove an unstarted runtime.
// Communication startup classification needs to distinguish these facts.
func (e *Engine) cleanupWithObservation(ctx context.Context, s stageSpec) (containerState, bool, bool, error) {
	var state containerState
	ids, err := e.matching(ctx, s)
	if err != nil {
		return state, false, false, err
	}
	if len(ids) == 0 {
		return state, false, false, nil
	}
	if len(ids) != 1 {
		return state, false, false, errors.New("cleanup_identity")
	}
	c, err := e.inspect(ctx, ids[0])
	if err != nil {
		return state, false, false, err
	}
	if c.Config.Labels["oj.judge.run"] != s.RunID || c.Config.Labels["oj.judge.owner"] != "file-mvp-v1" {
		return state, false, false, errors.New("cleanup_identity")
	}
	killConfirmed := false
	if c.State.Running {
		// A program may finish between inspect and kill; only the subsequent state decides.
		_, killErr := e.control(ctx, "kill", ids[0])
		c, err = e.inspect(ctx, ids[0])
		if err != nil {
			return state, false, false, err
		}
		killConfirmed = killErr == nil && !c.State.Running && c.State.ExitCode == 137
	}
	state = c.State
	if state.Running {
		return state, false, true, errors.New("cleanup_still_running")
	}
	if _, err = e.control(ctx, "rm", "--force", "--volumes", ids[0]); err != nil {
		return state, false, true, err
	}
	ids, err = e.matching(ctx, s)
	if err != nil || len(ids) != 0 {
		return state, false, true, errors.New("cleanup_not_confirmed")
	}
	return state, killConfirmed, true, nil
}

// The shared budget counts both streams. Full raw stdout survives up to the stop
// line; display truncation happens later and never changes checker input.
type outputBudget struct {
	total  atomic.Int64
	fired  atomic.Bool
	cancel context.CancelFunc
	limit  int64
}
type boundedStream struct {
	mu     sync.Mutex
	budget *outputBudget
	limit  int64
	count  int64
	data   bytes.Buffer
	file   io.Writer
	err    error
}

func (w *boundedStream) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n := len(p)
	w.count += int64(n)
	exceeded := w.count > w.limit
	keep := min(n, max(0, int(w.limit)-(int(w.count)-n)))
	if w.budget != nil {
		total := w.budget.total.Add(int64(n))
		exceeded = exceeded || total > w.budget.limit
		keep = min(keep, max(0, int(w.budget.limit-(total-int64(n)))))
	}
	if keep > 0 {
		if w.file != nil {
			_, err := w.file.Write(p[:keep])
			if err != nil {
				w.err = err
				exceeded = true
			}
		} else {
			_, _ = w.data.Write(p[:keep])
		}
	}
	if exceeded && w.budget != nil && w.budget.fired.CompareAndSwap(false, true) {
		w.budget.cancel()
	}
	return n, nil
}
func (e *Engine) stage(job context.Context, s stageSpec, out, errOut *boundedStream) stageResult {
	var r stageResult
	if job.Err() != nil {
		r.Failure = jobFailure(job)
		return r
	}
	_, err := e.control(job, e.createArgs(s)...)
	if err == nil {
		c, inspectErr := e.inspect(job, s.Name)
		if inspectErr != nil {
			err = inspectErr
		} else {
			err = e.verifyContainer(c, s)
		}
	}
	if err != nil {
		_, cleanupErr := e.cleanup(s)
		if cleanupErr != nil {
			e.tainted.Store(true)
			r.Failure = "cleanup_failed"
		} else if job.Err() != nil {
			r.Failure = jobFailure(job)
		} else {
			r.Failure = "container_create_failed"
		}
		return r
	}
	stage, deadlineCancel := context.WithTimeout(job, s.Wall)
	defer deadlineCancel()
	attach, cancel := context.WithCancel(stage)
	defer cancel()
	budget := &outputBudget{cancel: cancel, limit: MaxRawOutput}
	if s.Compile {
		errOut.budget = budget
		out.budget = &outputBudget{cancel: cancel, limit: MaxArtifact}
	} else {
		out.budget = budget
		errOut.budget = budget
	}
	started := time.Now()
	attachErr := e.command(attach, []string{"start", "--attach", s.Name}, out, errOut)
	r.DurationMS = time.Since(started).Milliseconds()
	if job.Err() != nil {
		r.Failure = jobFailure(job)
	} else if stage.Err() == context.DeadlineExceeded {
		r.Cause = "wall_timeout"
	} else if budget.fired.Load() || out.budget.fired.Load() {
		r.Cause = "output_limit"
	}
	state, cleanupErr := e.cleanup(s)
	r.State = state
	if cleanupErr != nil {
		e.tainted.Store(true)
		r.Failure = "cleanup_failed"
		return r
	}
	if job.Err() != nil {
		r.Failure = jobFailure(job)
	}
	if r.Failure != "" {
		return r
	}
	if state.StartedAt == "" || strings.HasPrefix(state.StartedAt, "0001-") || state.Running || state.Error != "" || out.err != nil || errOut.err != nil {
		r.Failure = "container_execution_failed"
		return r
	}
	if r.Cause == "" && attachErr != nil {
		var exit *exec.ExitError
		if !errors.As(attachErr, &exit) || exit.ExitCode() != state.ExitCode {
			r.Failure = "attach_incomplete"
		}
	}
	return r
}
func jobFailure(ctx context.Context) string {
	if ctx.Err() == context.DeadlineExceeded {
		return "job_timeout"
	}
	return "request_cancelled"
}
func caseVerdict(s stageResult, checker string, actual, expected []byte) string {
	if verdict := executionVerdict(s); verdict != "AC" {
		return verdict
	}
	if check(checker, actual, expected) {
		return "AC"
	}
	return "WA"
}

// Runtime failure is determined before any checker is allowed to inspect output.
func executionVerdict(s stageResult) string {
	if s.Failure != "" {
		return "SYSTEM_ERROR"
	}
	if s.State.OOMKilled {
		return "MLE"
	}
	if s.Cause == "wall_timeout" {
		return "TLE"
	}
	if s.Cause == "output_limit" {
		return "OLE"
	}
	if s.State.ExitCode != 0 {
		return "RE"
	}
	return "AC"
}
func stageName(run, stage string) string { return "oj-judge-" + run + "-" + stage }
func (e *Engine) runtimeSpec(run, stage, artifact, input string, m Manifest) stageSpec {
	return stageSpec{RunID: run, Name: stageName(run, stage), Stage: stage, Memory: m.MemoryLimitBytes, Pids: 16, Wall: time.Duration(m.WallTimeMS) * time.Millisecond, Mounts: map[string]string{"/solution/program": artifact, "/input.txt": input}, Entry: "/bin/sh", Args: []string{"-c", "exec /solution/program < /input.txt"}}
}
func probeExpected() string { return fmt.Sprintf("%d\n0\n16\n100000 100000\n", 64*1048576) }
