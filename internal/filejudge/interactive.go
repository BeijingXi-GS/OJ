package filejudge

import (
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

// This versioned, fixed protocol is part of the judge, never teacher input.
const guessNumberInteractor = `#include <fstream>
#include <iostream>
#include <sstream>
#include <string>
static int reject(const char* reason) { std::cerr << reason << '\n'; return 1; }
int main(int argc, char** argv) {
  if (argc != 2) return 125;
  std::ifstream data(argv[1]);
  int secret; std::string extra;
  if (!(data >> secret) || secret < 1 || secret > 1000 || (data >> extra)) return 125;
  std::cout << 1000 << std::endl;
  std::string line; int queries = 0;
  while (std::getline(std::cin, line)) {
    if (line.size() > 256) return reject("protocol line too long");
    std::istringstream row(line);
    std::string op; long long x;
    if (!(row >> op >> x) || (row >> extra) || (op != "?" && op != "!")) return reject("invalid protocol");
    if (x < 1 || x > 1000) return reject("query out of range");
    if (op == "!") return x == secret ? 0 : reject("wrong final answer");
    if (++queries > 10) return reject("too many queries");
    std::cout << (secret < x ? "<" : secret > x ? ">" : "=") << std::endl;
  }
  return reject("EOF before final answer");
}
`

const (
	participantSide = 1
	interactorSide  = 2
)

// There is one stop/retention line for the four protocol/diagnostic streams.
// The first write which crosses it records whose stream caused the stop.
type interactionBudget struct {
	mu     sync.Mutex
	total  atomic.Int64
	owner  atomic.Int32
	cancel context.CancelFunc
}

type interactionWriter struct {
	budget *interactionBudget
	side   int32
	store  *boundedStream
	peer   io.Writer
}

func (w *interactionWriter) Write(p []byte) (int, error) {
	n := len(p)
	// Reserving retained bytes and the first crossing owner is one event.
	// Neither I/O nor cancellation runs while this short lock is held.
	w.budget.mu.Lock()
	prior := w.budget.total.Load()
	total := prior + int64(n)
	w.budget.total.Store(total)
	keep := min(n, max(0, int(MaxRawOutput-prior)))
	first := total > MaxRawOutput && w.budget.owner.Load() == 0
	if first {
		w.budget.owner.Store(w.side)
	}
	w.budget.mu.Unlock()
	if keep > 0 {
		_, _ = w.store.Write(p[:keep])
	}
	if total > MaxRawOutput {
		if first {
			w.budget.cancel()
		}
		return n, nil
	}
	if w.peer != nil {
		_, err := w.peer.Write(p)
		// Closing a finished peer's stdin is normal protocol EOF. Its final
		// daemon state and attach completeness still decide the verdict.
		if err != nil && !errors.Is(err, io.ErrClosedPipe) {
			w.store.mu.Lock()
			w.store.err = err
			w.store.mu.Unlock()
			w.budget.cancel()
		}
	}
	return n, nil
}

type interactionResult struct {
	Participant, Interactor stageResult
	DurationMS              int64
	Failure, Cause          string
	ParticipantStopped      bool
	StartTimedOut           bool
	SupervisorKilled        [2]bool
}

func attachMatches(err error, state containerState) bool {
	if state.StartedAt == "" || strings.HasPrefix(state.StartedAt, "0001-") || state.Running || state.Error != "" {
		return false
	}
	if err == nil {
		return state.ExitCode == 0
	}
	var exit *exec.ExitError
	return errors.As(err, &exit) && exit.ExitCode() == state.ExitCode
}

func (e *Engine) cleanupInteraction(specs [2]stageSpec) ([2]containerState, [2]bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	var states [2]containerState
	var killed [2]bool
	var errs [2]error
	var group sync.WaitGroup
	for i := range specs {
		group.Add(1)
		go func(i int) {
			defer group.Done()
			states[i], killed[i], errs[i] = e.cleanupWithEvidence(ctx, specs[i])
		}(i)
	}
	group.Wait()
	if errs[0] != nil || errs[1] != nil {
		e.tainted.Store(true)
		return states, killed, errors.New("cleanup_failed")
	}
	return states, killed, nil
}

func (e *Engine) interaction(job context.Context, specs [2]stageSpec, stdout, stderr [2]*boundedStream) (result interactionResult) {
	started := time.Now()
	defer func() { result.DurationMS = time.Since(started).Milliseconds() }()
	stage, stopDeadline := context.WithTimeout(job, specs[0].Wall)
	defer stopDeadline()
	attach, stopAttach := context.WithCancel(stage)
	defer stopAttach()
	launched := false
	// Both identities are registered before any create, including uncertain
	// creates; cleanup queries each fixed name and run label even on failure.
	defer func() {
		states, killed, err := e.cleanupInteraction(specs)
		result.Participant.State, result.Interactor.State = states[0], states[1]
		result.SupervisorKilled = killed
		if err != nil {
			result.Failure = "cleanup_failed"
		} else if launched {
			for _, state := range states {
				if state.StartedAt == "" || strings.HasPrefix(state.StartedAt, "0001-") || state.Running || state.Error != "" {
					result.Failure = "container_execution_failed"
				}
			}
		} else if result.StartTimedOut {
			for _, state := range states {
				if (state.StartedAt != "" && !strings.HasPrefix(state.StartedAt, "0001-")) || state.OOMKilled || state.Running || state.Error != "" {
					result.Failure = "container_execution_failed"
				}
			}
		}
		if result.Failure != "cleanup_failed" && launched && !killed[1] && states[1].ExitCode != 0 && states[1].ExitCode != 1 {
			result.Failure = "interactor_execution_failed"
		}
		if result.Cause == "interactor_wa" {
			result.ParticipantStopped = killed[0]
			if !killed[0] && states[0].ExitCode != 0 {
				result.Cause = "participant_exit"
			}
		}
		if job.Err() != nil && result.Failure != "cleanup_failed" {
			result.Failure = jobFailure(job)
		}
		if result.Failure != "" || result.Cause != "interactor_wa" || states[0].OOMKilled || states[1].OOMKilled {
			// This flag certifies only the WA termination exception. A later
			// higher-priority failure must not claim that exception as valid.
			result.ParticipantStopped = false
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
		var controlFailure *controlError
		deadlineCause := err == nil || (errors.As(err, &controlFailure) && controlFailure.Kind == "timeout")
		if stage.Err() == context.DeadlineExceeded && deadlineCause {
			result.Cause, result.StartTimedOut = "wall_timeout", true
		} else {
			result.Failure = "container_create_failed"
		}
	}
	for _, spec := range specs {
		if stage.Err() != nil {
			startFailure(nil)
			return result
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
			return result
		}
	}
	if stage.Err() != nil {
		startFailure(nil)
		return result
	}
	toParticipant, sendParticipant := io.Pipe()
	toInteractor, sendInteractor := io.Pipe()
	var closeOnce sync.Once
	closePipes := func() {
		closeOnce.Do(func() {
			_ = toParticipant.Close()
			_ = toInteractor.Close()
			_ = sendParticipant.Close()
			_ = sendInteractor.Close()
		})
	}
	defer closePipes()
	// Cancellation must close blocked writes/reads before waiting for CLI
	// commands; otherwise a deadlocked protocol could outlive the deadline.
	watchDone := make(chan struct{})
	go func() {
		select {
		case <-attach.Done():
			closePipes()
		case <-watchDone:
		}
	}()
	defer close(watchDone)
	budget := &interactionBudget{cancel: stopAttach}
	readers := [2]io.ReadCloser{toParticipant, toInteractor}
	peers := [2]io.Writer{sendInteractor, sendParticipant}
	type completed struct {
		side int
		err  error
	}
	events := make(chan completed, 2)
	duplex := e.duplex
	if duplex == nil {
		duplex = dockerDuplexCommand
	}
	launched = true
	for i := range specs {
		go func(i int) {
			out := &interactionWriter{budget: budget, side: int32(i + 1), store: stdout[i], peer: peers[i]}
			errOut := &interactionWriter{budget: budget, side: int32(i + 1), store: stderr[i]}
			err := duplex(attach, []string{"start", "--attach", "--interactive", specs[i].Name}, readers[i], out, errOut)
			_ = readers[i].Close()
			if i == 0 {
				_ = sendInteractor.Close()
			} else {
				_ = sendParticipant.Close()
			}
			events <- completed{i, err}
		}(i)
	}
	var complete [2]bool
	for count := 0; count < 2; count++ {
		event := <-events
		i := event.side
		var inputFailure *duplexInputError
		if errors.As(event.err, &inputFailure) {
			result.Failure = "interaction_stdin_transfer_failed"
			stopAttach()
			continue
		}
		if attach.Err() != nil {
			var exit *exec.ExitError
			if event.err != nil && !errors.Is(event.err, context.Canceled) && !errors.Is(event.err, context.DeadlineExceeded) && !errors.As(event.err, &exit) {
				result.Failure = "interaction_attach_incomplete"
			}
			continue
		}
		c, err := e.inspect(stage, specs[i].Name)
		if err != nil || !attachMatches(event.err, c.State) {
			result.Failure = "interaction_attach_incomplete"
			stopAttach()
			continue
		}
		complete[i] = true
		if i == 0 && (c.State.OOMKilled || c.State.ExitCode != 0) {
			result.Cause = "participant_exit"
			stopAttach()
		}
		if i == 1 && (c.State.OOMKilled || c.State.ExitCode < 0 || c.State.ExitCode > 1) {
			result.Failure = "interactor_execution_failed"
			stopAttach()
		}
		if i == 1 && c.State.ExitCode == 1 && !c.State.OOMKilled {
			// Observe the participant before preempting it, so an already
			// completed natural crash is not disguised as supervised WA.
			participant, err := e.inspect(stage, specs[0].Name)
			if err != nil || participant.State.Error != "" || participant.State.StartedAt == "" || strings.HasPrefix(participant.State.StartedAt, "0001-") {
				result.Failure = "interaction_attach_incomplete"
			} else if participant.State.OOMKilled || (!participant.State.Running && participant.State.ExitCode != 0) {
				result.Cause = "participant_exit"
			} else {
				result.Cause = "interactor_wa"
			}
			stopAttach()
		}
	}
	closePipes()
	if job.Err() != nil {
		result.Failure = jobFailure(job)
	} else if result.Failure == "" && stage.Err() == context.DeadlineExceeded {
		result.Cause = "wall_timeout"
	} else if budget.owner.Load() != 0 {
		if budget.owner.Load() == participantSide {
			result.Cause = "participant_output_limit"
		} else {
			result.Failure = "interactor_output_limit"
		}
	}
	if result.Failure == "" && result.Cause == "" && (!complete[0] || !complete[1]) {
		result.Failure = "interaction_attach_incomplete"
	}
	for i := range stdout {
		if stdout[i].err != nil || stderr[i].err != nil {
			result.Failure = "interaction_io_failed"
		}
	}
	return result
}

func interactionVerdict(r interactionResult) string {
	if r.Failure != "" {
		return "SYSTEM_ERROR"
	}
	if r.Interactor.State.OOMKilled {
		return "SYSTEM_ERROR"
	}
	if r.Participant.State.OOMKilled {
		return "MLE"
	}
	if r.Cause == "wall_timeout" {
		return "TLE"
	}
	if r.Cause == "participant_output_limit" {
		return "OLE"
	}
	if r.Cause == "interactor_wa" {
		return "WA"
	}
	if r.Participant.State.ExitCode != 0 {
		return "RE"
	}
	if r.Interactor.State.ExitCode == 1 {
		return "WA"
	}
	if r.Interactor.State.ExitCode != 0 || r.Cause != "" || r.ParticipantStopped {
		return "SYSTEM_ERROR"
	}
	return "AC"
}

func (e *Engine) interactionSpecs(run, phase, artifact, interactor, input string, m Manifest) [2]stageSpec {
	base := stageSpec{RunID: run, Memory: m.MemoryLimitBytes, Pids: 16, Wall: time.Duration(m.WallTimeMS) * time.Millisecond, Interactive: true, Entry: "/bin/sh"}
	participant, judge := base, base
	participant.Stage, judge.Stage = phase+"-participant", phase+"-interactor"
	participant.Name, judge.Name = stageName(run, participant.Stage), stageName(run, judge.Stage)
	participant.Mounts = map[string]string{"/solution/program": artifact}
	judge.Mounts = map[string]string{"/solution/program": interactor, "/input.txt": input}
	participant.Args = []string{"-c", "exec /solution/program"}
	judge.Args = []string{"-c", "exec /solution/program /input.txt"}
	return [2]stageSpec{participant, judge}
}

func (e *Engine) compileInteractor(ctx context.Context, run, dir string) (CompileResult, string, string, error) {
	interactorDir := filepath.Join(dir, "interactor")
	if os.Mkdir(interactorDir, 0700) != nil {
		return CompileResult{Status: "system_error"}, "", "", errors.New("snapshot_write_failed")
	}
	return e.compileNamed(ctx, run, interactorDir, guessNumberInteractor, "interactor-compile")
}

func (e *Engine) judgeInteractive(ctx context.Context, run, dir, artifact string, snap Snapshot, result *Result) {
	if snap.Interactor != "guess-number-v1" {
		result.Message = failureMessage("interactor_protocol_invalid")
		return
	}
	result.InteractorSourceSHA256 = hash([]byte(guessNumberInteractor))
	compile, judge, artifactHash, err := e.compileInteractor(ctx, run, dir)
	result.InteractorCompile = &compile
	if err != nil || compile.Status != "success" {
		result.Message = failureMessage("interactor_compile_failed")
		return
	}
	result.InteractorArtifactSHA256 = artifactHash
	result.Verdict = "AC"
	for i, c := range snap.Cases {
		if ctx.Err() != nil {
			result.Verdict, result.Message = "SYSTEM_ERROR", failureMessage(jobFailure(ctx))
			return
		}
		input := filepath.Join(dir, "input-"+c.ID+".txt")
		if os.WriteFile(input, c.Input, 0644) != nil {
			result.Verdict, result.Message = "SYSTEM_ERROR", failureMessage("snapshot_write_failed")
			return
		}
		out := [2]*boundedStream{{limit: MaxRawOutput}, {limit: MaxRawOutput}}
		stderr := [2]*boundedStream{{limit: MaxRawOutput}, {limit: MaxRawOutput}}
		stage := e.interaction(ctx, e.interactionSpecs(run, "case-"+string(rune('a'+i)), artifact, judge, input, snap.Manifest), out, stderr)
		cr := CaseResult{ID: c.ID, Verdict: interactionVerdict(stage), DurationMS: stage.DurationMS, OOMKilled: stage.Participant.State.OOMKilled, StartTimedOut: stage.StartTimedOut}
		cr.Input, cr.InputTruncated = display(c.Input, MaxFile)
		if stage.Participant.State.StartedAt != "" && !strings.HasPrefix(stage.Participant.State.StartedAt, "0001-") {
			cr.ExitCode = &stage.Participant.State.ExitCode
		}
		var outTrunc, errTrunc bool
		cr.Stdout, outTrunc = display(out[0].data.Bytes(), MaxFile)
		cr.Stderr, errTrunc = display(stderr[0].data.Bytes(), 16384)
		cr.OutputTruncated = outTrunc || errTrunc || stage.Cause == "participant_output_limit"
		interactor := &InteractorResult{OOMKilled: stage.Interactor.State.OOMKilled, ParticipantStopped: stage.ParticipantStopped}
		if stage.Interactor.State.StartedAt != "" && !strings.HasPrefix(stage.Interactor.State.StartedAt, "0001-") {
			interactor.ExitCode = &stage.Interactor.State.ExitCode
		}
		interactor.Stdout, outTrunc = display(out[1].data.Bytes(), MaxFile)
		interactor.Stderr, errTrunc = display(stderr[1].data.Bytes(), 16384)
		interactor.OutputTruncated = outTrunc || errTrunc || stage.Failure == "interactor_output_limit"
		cr.Interactor = interactor
		result.Cases = append(result.Cases, cr)
		if cr.Verdict == "SYSTEM_ERROR" {
			result.Verdict, result.Message = "SYSTEM_ERROR", failureMessage(stage.Failure)
			return
		}
		if result.Verdict == "AC" && cr.Verdict != "AC" {
			result.Verdict = cr.Verdict
		}
	}
	result.Message = message(result.Verdict)
}

func (e *Engine) interactiveProbe(ctx context.Context, run, dir string) error {
	probeDir := filepath.Join(dir, "duplex-probe")
	if os.Mkdir(probeDir, 0700) != nil {
		return errors.New("interaction_probe_directory_failed")
	}
	cr, participant, _, err := e.compileNamed(ctx, run, probeDir, "#include <iostream>\nint main(){int bound;if(!(std::cin>>bound)||bound!=1000)return 2;std::cout<<\"! 617\"<<std::endl;}\n", "interaction-probe-compile")
	if err != nil || cr.Status != "success" {
		return errors.New("interaction_probe_compile_failed")
	}
	cr, interactor, _, err := e.compileInteractor(ctx, run, probeDir)
	if err != nil || cr.Status != "success" {
		return errors.New("interaction_probe_compile_failed")
	}
	input := filepath.Join(probeDir, "secret.txt")
	if os.WriteFile(input, []byte("617\n"), 0644) != nil {
		return errors.New("interaction_probe_input_failed")
	}
	out := [2]*boundedStream{{limit: MaxRawOutput}, {limit: MaxRawOutput}}
	stderr := [2]*boundedStream{{limit: MaxRawOutput}, {limit: MaxRawOutput}}
	r := e.interaction(ctx, e.interactionSpecs(run, "interaction-probe", participant, interactor, input, Manifest{WallTimeMS: 5000, MemoryLimitBytes: 64 * 1048576}), out, stderr)
	if interactionVerdict(r) != "AC" || out[0].data.String() != "! 617\n" || out[1].data.String() != "1000\n" || e.tainted.Load() {
		return errors.New("interaction_duplex_probe_failed")
	}
	return nil
}
