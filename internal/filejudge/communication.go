package filejudge

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// The entire accepted message is bounded by the trusted case byte limit. A
// nonblocking writer permits immediate over-limit termination even when the
// receiver does not read. Accepted bytes are not evidence of program consumption.
type communicationPipe struct {
	mu          sync.Mutex
	ready       *sync.Cond
	data        bytes.Buffer
	eof, closed bool
}

func newCommunicationPipe() *communicationPipe {
	p := &communicationPipe{}
	p.ready = sync.NewCond(&p.mu)
	return p
}
func (p *communicationPipe) Read(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for p.data.Len() == 0 && !p.eof && !p.closed {
		p.ready.Wait()
	}
	if p.closed {
		return 0, io.ErrClosedPipe
	}
	if p.data.Len() > 0 {
		return p.data.Read(b)
	}
	return 0, io.EOF
}
func (p *communicationPipe) Write(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed || p.eof {
		return 0, io.ErrClosedPipe
	}
	n, err := p.data.Write(b)
	p.ready.Broadcast()
	return n, err
}
func (p *communicationPipe) Close() error {
	p.mu.Lock()
	p.closed = true
	p.data.Reset()
	p.ready.Broadcast()
	p.mu.Unlock()
	return nil
}
func (p *communicationPipe) finish() { p.mu.Lock(); p.eof = true; p.ready.Broadcast(); p.mu.Unlock() }

type communicationRelay struct {
	pipe                  *communicationPipe
	retained              *boundedStream
	limit                 int64
	observed, transferred atomic.Int64
	exceeded, closedPeer  atomic.Bool
	cancel                context.CancelFunc
	mu                    sync.Mutex
	failure               error
}

func (w *communicationRelay) Write(p []byte) (int, error) {
	n := len(p)
	prior := w.observed.Add(int64(n)) - int64(n)
	allowed := min(n, max(0, int(w.limit-prior)))
	if allowed > 0 {
		_, _ = w.retained.Write(p[:allowed])
		if !w.closedPeer.Load() {
			transferred, err := w.pipe.Write(p[:allowed])
			w.transferred.Add(int64(transferred))
			if err != nil {
				if errors.Is(err, io.ErrClosedPipe) {
					w.closedPeer.Store(true)
				} else {
					w.mu.Lock()
					w.failure = err
					w.mu.Unlock()
					w.cancel()
				}
			}
		}
	}
	if prior+int64(n) > w.limit && w.exceeded.CompareAndSwap(false, true) {
		w.cancel()
	}
	return n, nil
}

type communicationStage struct {
	Roles                           [2]stageResult
	DurationMS                      int64
	Cause, Failure                  string
	StartTimedOut                   bool
	Killed                          [2]bool
	BytesObserved, BytesTransferred int64
	LimitExceeded                   bool
}

func (e *Engine) cleanupCommunication(specs [2]stageSpec) ([2]containerState, [2]bool, [2]bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	var states [2]containerState
	var killed, observed [2]bool
	var errs [2]error
	var group sync.WaitGroup
	for i := range specs {
		group.Add(1)
		go func(i int) {
			defer group.Done()
			states[i], killed[i], observed[i], errs[i] = e.cleanupWithObservation(ctx, specs[i])
		}(i)
	}
	group.Wait()
	if errs[0] != nil || errs[1] != nil {
		e.tainted.Store(true)
		return states, killed, observed, errors.New("cleanup_failed")
	}
	return states, killed, observed, nil
}

func (e *Engine) communicationSpecs(run, phase string, artifacts [2]string, input string, m Manifest) [2]stageSpec {
	base := stageSpec{RunID: run, Memory: m.MemoryLimitBytes, Pids: 32, Wall: time.Duration(m.WallTimeMS) * time.Millisecond, Interactive: true, Entry: "/bin/sh"}
	sender, receiver := base, base
	sender.Stage, receiver.Stage = phase+"-sender", phase+"-receiver"
	sender.Name, receiver.Name = stageName(run, sender.Stage), stageName(run, receiver.Stage)
	sender.Mounts = map[string]string{"/solution/program": artifacts[0], "/input.txt": input}
	receiver.Mounts = map[string]string{"/solution/program": artifacts[1]}
	sender.Args = []string{"-c", "exec /solution/program < /input.txt"}
	receiver.Args = []string{"-c", "exec /solution/program"}
	return [2]stageSpec{sender, receiver}
}

func (e *Engine) communication(job context.Context, specs [2]stageSpec, limit int64, stdout, stderr [2]*boundedStream) (result communicationStage) {
	started := time.Now()
	defer func() { result.DurationMS = time.Since(started).Milliseconds() }()
	stage, stopDeadline := context.WithTimeout(job, specs[0].Wall)
	defer stopDeadline()
	attach, stopAttach := context.WithCancel(stage)
	defer stopAttach()
	launched := false
	var complete [2]bool
	defer func() {
		states, killed, observed, err := e.cleanupCommunication(specs)
		result.Killed = killed
		for i := range states {
			result.Roles[i].State = states[i]
		}
		if err != nil {
			result.Failure = "cleanup_failed"
		} else if launched {
			for i, state := range states {
				if !observed[i] {
					result.Failure = "container_execution_failed"
					continue
				}
				started := state.StartedAt != "" && !strings.HasPrefix(state.StartedAt, "0001-")
				if !started && result.Cause == "wall_timeout" && !state.Running && !state.OOMKilled && state.Error == "" && state.ExitCode == 0 {
					// Shared startup time is part of the case wall budget. A
					// verified, still-unstarted peer has no exit evidence.
					result.StartTimedOut = true
					continue
				}
				if !started || state.Running || state.Error != "" {
					result.Failure = "container_execution_failed"
				}
			}
		} else if result.StartTimedOut {
			for _, state := range states {
				if state.StartedAt != "" && !strings.HasPrefix(state.StartedAt, "0001-") || state.Running || state.OOMKilled || state.Error != "" {
					result.Failure = "container_execution_failed"
				}
			}
		}
		if job.Err() != nil && result.Failure != "cleanup_failed" {
			result.Failure = jobFailure(job)
		}
		if result.Failure != "" {
			result.StartTimedOut = false
		}
	}()
	startFailure := func(err error) {
		if job.Err() != nil {
			result.Failure = jobFailure(job)
			return
		}
		var ce *controlError
		if stage.Err() == context.DeadlineExceeded && (err == nil || errors.As(err, &ce) && ce.Kind == "timeout") {
			result.Cause = "wall_timeout"
			result.StartTimedOut = true
		} else {
			result.Failure = "container_create_failed"
		}
	}
	for _, spec := range specs {
		if stage.Err() != nil {
			startFailure(nil)
			return
		}
		_, err := e.control(stage, e.createArgs(spec)...)
		if err == nil {
			var c containerInfo
			c, err = e.inspect(stage, spec.Name)
			if err == nil {
				err = e.verifyContainer(c, spec)
			}
		}
		if err != nil {
			startFailure(err)
			return
		}
	}
	if stage.Err() != nil {
		startFailure(nil)
		return
	}
	pipe := newCommunicationPipe()
	defer pipe.Close()
	watchDone := make(chan struct{})
	go func() {
		select {
		case <-attach.Done():
			_ = pipe.Close()
		case <-watchDone:
		}
	}()
	defer close(watchDone)
	relay := &communicationRelay{pipe: pipe, retained: stdout[0], limit: limit, cancel: stopAttach}
	budget := &outputBudget{limit: MaxRawOutput, cancel: stopAttach}
	stderr[0].budget, stdout[1].budget, stderr[1].budget = budget, budget, budget
	type completed struct {
		side       int
		err        error
		durationMS int64
	}
	events := make(chan completed, 2)
	duplex := e.duplex
	if duplex == nil {
		duplex = dockerDuplexCommand
	}
	inputs := [2]io.ReadCloser{io.NopCloser(strings.NewReader("")), pipe}
	outputs := [2]io.Writer{relay, stdout[1]}
	launched = true
	for i := range specs {
		go func(i int) {
			start := time.Now()
			err := duplex(attach, []string{"start", "--attach", "--interactive", specs[i].Name}, inputs[i], outputs[i], stderr[i])
			if i == 0 {
				pipe.finish()
			} else {
				_ = pipe.Close()
			}
			events <- completed{i, err, time.Since(start).Milliseconds()}
		}(i)
	}
	for count := 0; count < 2; count++ {
		event := <-events
		i := event.side
		result.Roles[i].DurationMS = event.durationMS
		var transfer *duplexInputError
		if errors.As(event.err, &transfer) {
			// A receiver which naturally finished may close its CLI stdin between a
			// channel read and write. Only its independent exit evidence exempts this.
			if i == 1 && transfer.pipeClosed && attach.Err() == nil {
				c, err := e.inspect(stage, specs[i].Name)
				if err == nil && attachMatches(transfer.attachErr, c.State) {
					complete[i] = true
					if c.State.OOMKilled || c.State.ExitCode != 0 {
						result.Cause = "role_exit"
						stopAttach()
					}
					continue
				}
			}
			result.Failure = "communication_stdin_transfer_failed"
			stopAttach()
			continue
		}
		if attach.Err() != nil {
			var exit *exec.ExitError
			if event.err != nil && !errors.Is(event.err, context.Canceled) && !errors.Is(event.err, context.DeadlineExceeded) && !errors.As(event.err, &exit) {
				result.Failure = "communication_attach_incomplete"
			}
			continue
		}
		c, err := e.inspect(stage, specs[i].Name)
		if err != nil || !attachMatches(event.err, c.State) {
			result.Failure = "communication_attach_incomplete"
			stopAttach()
			continue
		}
		complete[i] = true
		if c.State.OOMKilled || c.State.ExitCode != 0 {
			result.Cause = "role_exit"
			stopAttach()
		}
	}
	result.BytesObserved, result.BytesTransferred = relay.observed.Load(), relay.transferred.Load()
	result.LimitExceeded = relay.exceeded.Load()
	if job.Err() != nil {
		result.Failure = jobFailure(job)
	} else if stage.Err() == context.DeadlineExceeded {
		result.Cause = "wall_timeout"
	} else if budget.fired.Load() {
		result.Cause = "output_limit"
	} else if relay.exceeded.Load() {
		result.Cause = "communication_limit"
	}
	relay.mu.Lock()
	relayFailure := relay.failure
	relay.mu.Unlock()
	if relayFailure != nil {
		result.Failure = "communication_io_failed"
	}
	if result.Failure == "" && result.Cause == "" && (!complete[0] || !complete[1]) {
		result.Failure = "communication_attach_incomplete"
	}
	for i := range stdout {
		if stdout[i].err != nil || stderr[i].err != nil {
			result.Failure = "communication_io_failed"
		}
	}
	return
}

func communicationVerdict(r communicationStage, actual, expected []byte) string {
	if r.Failure != "" {
		return "SYSTEM_ERROR"
	}
	if r.Roles[0].State.OOMKilled || r.Roles[1].State.OOMKilled {
		return "MLE"
	}
	if r.Cause == "wall_timeout" {
		return "TLE"
	}
	if r.Cause == "output_limit" {
		return "OLE"
	}
	if r.Cause == "communication_limit" {
		return "WA"
	}
	for i := range r.Roles {
		if !r.Killed[i] && r.Roles[i].State.ExitCode != 0 {
			return "RE"
		}
	}
	if r.Cause != "" || r.Killed[0] || r.Killed[1] {
		return "SYSTEM_ERROR"
	}
	if !check("tokens-v1", actual, expected) {
		return "WA"
	}
	return "AC"
}

func (e *Engine) judgeCommunication(ctx context.Context, run, dir string, req Request, snap Snapshot, result *Result) {
	units, invalid := CanonicalSourceUnits(req)
	if invalid != nil || snap.Manifest.Communication == nil {
		result.Message = failureMessage("communication_configuration_invalid")
		return
	}
	var artifacts [2]string
	result.Compile = CompileResult{Status: "system_error"}
	for i, u := range units {
		build := filepath.Join(dir, u.Role+"-build")
		if os.Mkdir(build, 0700) != nil {
			result.Message = failureMessage("snapshot_write_failed")
			return
		}
		c, a, digest, err := e.compileNamed(ctx, run, build, u.Source, u.Role+"-compile")
		result.RoleCompiles = append(result.RoleCompiles, RoleCompile{Role: u.Role, SourceSHA256: hash([]byte(u.Source)), ArtifactSHA256: digest, Compile: c})
		result.Compile.DurationMS += c.DurationMS
		result.Compile.DiagnosticsTruncated = result.Compile.DiagnosticsTruncated || c.DiagnosticsTruncated
		if c.Diagnostics != "" {
			var truncated bool
			result.Compile.Diagnostics, truncated = display([]byte("["+u.Role+"]\n"+c.Diagnostics), MaxFile)
			result.Compile.DiagnosticsTruncated = result.Compile.DiagnosticsTruncated || truncated
		}
		if err != nil {
			result.Message = failureMessage(err.Error())
			return
		}
		if c.Status == "compile_error" {
			result.Compile.Status = "compile_error"
			result.Verdict = "CE"
			result.Message = "通信题的 " + u.Role + " 程序编译失败。"
			return
		}
		if c.Status != "success" {
			result.Message = failureMessage("compile_helper_failed")
			return
		}
		artifacts[i] = a
	}
	result.Compile.Status = "success"
	result.ArtifactSHA256 = CommunicationArtifactDigest(result.RoleCompiles)
	result.Verdict = "AC"
	for i, c := range snap.Cases {
		if ctx.Err() != nil {
			result.Verdict = "SYSTEM_ERROR"
			result.Message = failureMessage(jobFailure(ctx))
			return
		}
		input := filepath.Join(dir, "input-"+c.ID+".txt")
		if os.WriteFile(input, c.Input, 0644) != nil {
			result.Verdict = "SYSTEM_ERROR"
			result.Message = failureMessage("snapshot_write_failed")
			return
		}
		out := [2]*boundedStream{{limit: MaxFile}, {limit: MaxRawOutput}}
		stderr := [2]*boundedStream{{limit: MaxRawOutput}, {limit: MaxRawOutput}}
		stage := e.communication(ctx, e.communicationSpecs(run, "case-"+string(rune('a'+i)), artifacts, input, snap.Manifest), snap.Manifest.Communication.ByteLimit, out, stderr)
		cr := CaseResult{ID: c.ID, Verdict: communicationVerdict(stage, out[1].data.Bytes(), c.Expected), DurationMS: stage.DurationMS, OOMKilled: stage.Roles[0].State.OOMKilled || stage.Roles[1].State.OOMKilled, StartTimedOut: stage.StartTimedOut}
		cr.Input, cr.InputTruncated = display(c.Input, MaxFile)
		cr.Expected, cr.ExpectedTruncated = display(c.Expected, MaxFile)
		communication := &CommunicationResult{BytesObserved: stage.BytesObserved, BytesTransferred: stage.BytesTransferred, ByteLimit: snap.Manifest.Communication.ByteLimit, LimitExceeded: stage.LimitExceeded}
		roles := [2]*RoleRun{&communication.Sender, &communication.Receiver}
		for i, role := range roles {
			role.DurationMS, role.OOMKilled, role.TerminatedBySupervisor = stage.Roles[i].DurationMS, stage.Roles[i].State.OOMKilled, stage.Killed[i]
			if state := stage.Roles[i].State; state.StartedAt != "" && !strings.HasPrefix(state.StartedAt, "0001-") {
				code := state.ExitCode
				role.ExitCode = &code
			}
			var a, b bool
			role.Stdout, a = display(out[i].data.Bytes(), MaxFile)
			role.Stderr, b = display(stderr[i].data.Bytes(), 16384)
			role.OutputTruncated = a || b || stage.Cause == "output_limit" || i == 0 && stage.LimitExceeded
		}
		cr.Communication = communication
		cr.ExitCode = communication.Receiver.ExitCode
		cr.Stdout = communication.Receiver.Stdout
		combined := []byte(communication.Sender.Stderr + communication.Receiver.Stderr)
		var diagTruncated bool
		cr.Stderr, diagTruncated = display(combined, 16384)
		cr.OutputTruncated = communication.Sender.OutputTruncated || communication.Receiver.OutputTruncated || diagTruncated
		result.Cases = append(result.Cases, cr)
		if cr.Verdict == "SYSTEM_ERROR" {
			result.Verdict = "SYSTEM_ERROR"
			result.Message = failureMessage(stage.Failure)
			return
		}
		if result.Verdict == "AC" && cr.Verdict != "AC" {
			result.Verdict = cr.Verdict
		}
	}
	result.Message = message(result.Verdict)
}
