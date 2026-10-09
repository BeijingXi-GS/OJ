package filejudge

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// A trusted test helper supplies real exec.ExitError values without running a
// submitted program, calling Docker, or relying on OS-specific shell commands.
func TestInteractiveExitHelper(t *testing.T) {
	if value := os.Getenv("OJ_TRUSTED_TEST_EXIT"); value != "" {
		code, err := strconv.Atoi(value)
		if err != nil {
			os.Exit(125)
		}
		os.Exit(code)
	}
}

func trustedExitError(t *testing.T, code int) error {
	t.Helper()
	c := exec.Command(os.Args[0], "-test.run=^TestInteractiveExitHelper$")
	c.Env = append(os.Environ(), "OJ_TRUSTED_TEST_EXIT="+strconv.Itoa(code))
	err := c.Run()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != code {
		t.Fatal("trusted helper failed", err)
	}
	return err
}

func fakeInteractionInfo(e *Engine, s stageSpec) containerInfo {
	var c containerInfo
	c.ID, c.Image = s.Name, e.image
	c.Config.User = "10002:10002"
	c.Config.OpenStdin = true
	c.Config.Labels = map[string]string{"oj.judge.owner": "file-mvp-v1", "oj.judge.run": s.RunID}
	h := &c.HostConfig
	h.ReadonlyRootfs = true
	h.NetworkMode, h.IpcMode, h.CgroupnsMode = "none", "private", "private"
	h.Memory, h.MemorySwap, h.NanoCpus, h.PidsLimit = s.Memory, s.Memory, 1000000000, 16
	h.CapDrop, h.SecurityOpt = []string{"ALL"}, []string{"no-new-privileges"}
	init := true
	h.Init = &init
	h.LogConfig.Type = "none"
	for dst, src := range s.Mounts {
		c.Mounts = append(c.Mounts, struct {
			Type, Source, Destination string
			RW                        bool
		}{Type: "bind", Source: src, Destination: dst})
	}
	return c
}

func TestInteractionIsolationArgs(t *testing.T) {
	e := &Engine{image: "sha256:" + strings.Repeat("a", 64)}
	specs := e.interactionSpecs("run", "case-a", "participant-binary", "interactor-binary", "secret-input", Manifest{MemoryLimitBytes: 64 * 1048576, WallTimeMS: 5000})
	if len(specs[0].Mounts) != 1 || specs[0].Mounts["/input.txt"] != "" || specs[0].Mounts["/solution/program"] != "participant-binary" || specs[1].Mounts["/input.txt"] != "secret-input" {
		t.Fatal("secret is not confined to interactor")
	}
	for _, s := range specs {
		args := strings.Join(e.createArgs(s), " ")
		if !strings.Contains(args, "--interactive") || strings.Contains(args, "--privileged") || strings.Contains(args, "--network host") || !strings.Contains(args, "--read-only") {
			t.Fatal("invalid interaction create", args)
		}
		c := fakeInteractionInfo(e, s)
		if err := e.verifyContainer(c, s); err != nil {
			t.Fatal(err)
		}
		c.Config.Tty = true
		if e.verifyContainer(c, s) == nil {
			t.Fatal("TTY could merge/corrupt protocol streams")
		}
	}
}

func TestInteractionAggregateBudget(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	budget := &interactionBudget{cancel: cancel}
	a, b := &boundedStream{limit: MaxRawOutput}, &boundedStream{limit: MaxRawOutput}
	participant := &interactionWriter{budget: budget, side: participantSide, store: a}
	interactor := &interactionWriter{budget: budget, side: interactorSide, store: b}
	_, _ = participant.Write(make([]byte, 700000))
	_, _ = interactor.Write(make([]byte, 700000))
	if ctx.Err() == nil || budget.owner.Load() != interactorSide || a.data.Len()+b.data.Len() != MaxRawOutput {
		t.Fatal("aggregate byte line or first owner lost")
	}
	_, _ = participant.Write(make([]byte, 10))
	if budget.owner.Load() != interactorSide || a.data.Len()+b.data.Len() != MaxRawOutput {
		t.Fatal("first owner or retained bytes changed after stop")
	}
}

func TestInteractionLifecycle(t *testing.T) {
	one, two, killed := trustedExitError(t, 1), trustedExitError(t, 2), trustedExitError(t, 137)
	for _, name := range []string{"normal", "wa_waiting", "wa_natural137", "wa_kill_fail", "participant_re", "participant_137", "participant_oom", "timeout", "participant_ole", "interactor_ole", "interactor_bad_exit", "interactor_oom", "attach_incomplete", "stdin_transfer_error", "create_uncertain", "create_deadline", "inspect_deadline", "create_job_deadline", "cleanup_failure", "cancelled", "job_deadline", "participant_re_late_interactor_bad", "timeout_late_interactor_bad", "participant_ole_late_interactor_bad"} {
		t.Run(name, func(t *testing.T) {
			job, stopJob := context.WithCancel(context.Background())
			defer stopJob()
			if name == "job_deadline" || name == "create_job_deadline" {
				var stop context.CancelFunc
				job, stop = context.WithTimeout(context.Background(), 35*time.Millisecond)
				defer stop()
			}
			e := &Engine{image: "sha256:" + strings.Repeat("a", 64)}
			specs := e.interactionSpecs(strings.Repeat("b", 32), "case-a", "participant", "interactor", "secret", Manifest{MemoryLimitBytes: 64 * 1048576, WallTimeMS: 200})
			var mu sync.Mutex
			deadlineTriggered := false
			containers := map[string]containerInfo{}
			setState := func(container string, running bool, code int, oom bool) {
				mu.Lock()
				defer mu.Unlock()
				c := containers[container]
				c.State = containerState{Running: running, ExitCode: code, OOMKilled: oom, StartedAt: "2026-10-02T00:00:00Z"}
				containers[container] = c
			}
			e.command = func(ctx context.Context, args []string, out, _ io.Writer) error {
				mu.Lock()
				defer mu.Unlock()
				switch args[0] {
				case "create":
					container := args[2]
					for _, spec := range specs {
						if container == spec.Name {
							containers[container] = fakeInteractionInfo(e, spec)
						}
					}
					if (name == "create_deadline" || name == "create_job_deadline") && !deadlineTriggered {
						deadlineTriggered = true
						<-ctx.Done()
						return ctx.Err()
					}
					if name == "create_uncertain" && container == specs[1].Name {
						return errors.New("lost create reply")
					}
				case "inspect":
					c, found := containers[args[1]]
					if !found {
						return errors.New("missing fake container")
					}
					if name == "inspect_deadline" && !deadlineTriggered {
						deadlineTriggered = true
						<-ctx.Done()
						return ctx.Err()
					}
					b, _ := json.Marshal([]containerInfo{c})
					_, _ = out.Write(b)
				case "ps":
					container := strings.TrimSuffix(strings.TrimPrefix(args[len(args)-1], "name=^/"), "$")
					if _, found := containers[container]; found {
						_, _ = io.WriteString(out, container)
					}
				case "kill":
					c := containers[args[1]]
					c.State.Running, c.State.ExitCode = false, 137
					containers[args[1]] = c
					if name == "wa_kill_fail" && args[1] == specs[0].Name {
						return errors.New("already naturally exited")
					}
				case "rm":
					if name == "cleanup_failure" && args[len(args)-1] == specs[1].Name {
						return errors.New("daemon unavailable")
					}
					delete(containers, args[len(args)-1])
				default:
					t.Errorf("unexpected fake control %v", args)
				}
				return nil
			}
			participantStarted, interactorStarted := make(chan struct{}), make(chan struct{})
			e.duplex = func(ctx context.Context, args []string, input io.ReadCloser, out, _ io.Writer) error {
				container := args[len(args)-1]
				participant := container == specs[0].Name
				setState(container, true, 0, false)
				if participant {
					close(participantStarted)
					<-interactorStarted
				} else {
					close(interactorStarted)
					<-participantStarted
				}
				if !participant && strings.HasSuffix(name, "_late_interactor_bad") {
					<-ctx.Done()
					setState(container, false, 2, false)
					return two
				}
				if name == "cancelled" || name == "job_deadline" || strings.HasPrefix(name, "timeout") || (strings.HasPrefix(name, "wa_") && participant) {
					if name == "cancelled" {
						stopJob()
					}
					<-ctx.Done()
					if name == "wa_natural137" {
						setState(container, false, 137, false)
					}
					return ctx.Err()
				}
				if participant {
					switch name {
					case "participant_re", "participant_re_late_interactor_bad":
						setState(container, false, 2, false)
						return two
					case "participant_137", "participant_oom":
						setState(container, false, 137, name == "participant_oom")
						return killed
					case "participant_ole", "participant_ole_late_interactor_bad":
						_, _ = out.Write(make([]byte, MaxRawOutput+1))
						<-ctx.Done()
						return ctx.Err()
					case "stdin_transfer_error":
						setState(container, false, 0, false)
						return &duplexInputError{}
					}
					if name == "interactor_ole" || name == "interactor_bad_exit" || name == "interactor_oom" || name == "attach_incomplete" {
						<-ctx.Done()
						return ctx.Err()
					}
					reader := bufio.NewReader(input)
					line, err := reader.ReadString('\n')
					if err != nil || line != "1000\n" {
						return errors.New("protocol bound missing")
					}
					_, _ = io.WriteString(out, "! 617\n")
					setState(container, false, 0, false)
					return nil
				}
				switch name {
				case "participant_re", "participant_137", "participant_oom", "participant_ole", "stdin_transfer_error":
					<-ctx.Done()
					return ctx.Err()
				case "wa_waiting", "wa_natural137", "wa_kill_fail":
					setState(container, false, 1, false)
					return one
				case "interactor_bad_exit":
					setState(container, false, 2, false)
					return two
				case "interactor_oom":
					setState(container, false, 137, true)
					return killed
				case "interactor_ole":
					_, _ = out.Write(make([]byte, MaxRawOutput+1))
					<-ctx.Done()
					return ctx.Err()
				case "attach_incomplete":
					setState(container, false, 0, false)
					return errors.New("lost output stream")
				}
				_, _ = io.WriteString(out, "1000\n")
				reader := bufio.NewReader(input)
				line, err := reader.ReadString('\n')
				if err != nil || line != "! 617\n" {
					return errors.New("final answer missing")
				}
				setState(container, false, 0, false)
				return nil
			}
			out := [2]*boundedStream{{limit: MaxRawOutput}, {limit: MaxRawOutput}}
			stderr := [2]*boundedStream{{limit: MaxRawOutput}, {limit: MaxRawOutput}}
			r := e.interaction(job, specs, out, stderr)
			want := "SYSTEM_ERROR"
			switch name {
			case "normal":
				want = "AC"
			case "wa_waiting":
				want = "WA"
			case "participant_re", "participant_137", "wa_natural137", "wa_kill_fail":
				want = "RE"
			case "participant_oom":
				want = "MLE"
			case "timeout", "create_deadline", "inspect_deadline":
				want = "TLE"
			case "participant_ole":
				want = "OLE"
			}
			if got := interactionVerdict(r); got != want {
				t.Fatalf("want %s got %s: %+v", want, got, r)
			}
			// Preserve actual terminal facts and verify they can cross the
			// strict internal wire, especially supervised WA with exit 137.
			cr := CaseResult{ID: "sample", Verdict: want, OOMKilled: r.Participant.State.OOMKilled, StartTimedOut: r.StartTimedOut, Interactor: &InteractorResult{OOMKilled: r.Interactor.State.OOMKilled, ParticipantStopped: r.ParticipantStopped}}
			if r.Participant.State.StartedAt != "" {
				cr.ExitCode = &r.Participant.State.ExitCode
			}
			if r.Interactor.State.StartedAt != "" {
				cr.Interactor.ExitCode = &r.Interactor.State.ExitCode
			}
			wire := Result{RunID: specs[0].RunID, Verdict: want, SourceSHA256: strings.Repeat("c", 64), TestdataSHA256: strings.Repeat("d", 64), ArtifactSHA256: strings.Repeat("e", 64), RuntimeImage: e.image, Compiler: "GCC 14.2.0", Standard: "c++20", Compile: CompileResult{Status: "success"}, InteractorCompile: &CompileResult{Status: "success"}, InteractorSourceSHA256: InteractorSourceDigest(), InteractorArtifactSHA256: strings.Repeat("a", 64), Cases: []CaseResult{cr}}
			data, marshalErr := json.Marshal(wire)
			if marshalErr != nil || !completeWire(data) || !validResult(wire) {
				t.Fatal("actual interaction facts violate internal wire", string(data), marshalErr)
			}
			if name == "wa_waiting" && (!r.ParticipantStopped || r.Participant.State.ExitCode != 137 || r.Interactor.State.ExitCode != 1) {
				t.Fatal("supervised WA lost real participant/interactor evidence", r)
			}
			if (name == "wa_natural137" || name == "wa_kill_fail") && (r.ParticipantStopped || r.SupervisorKilled[0]) {
				t.Fatal("natural exit 137 claimed supervised termination", r)
			}
			if (name == "create_deadline" || name == "inspect_deadline") && (!r.StartTimedOut || cr.ExitCode != nil || cr.Interactor.ExitCode != nil) {
				t.Fatal("startup wall timeout fabricated exit evidence", r)
			}
			if name == "normal" && (r.ParticipantStopped || out[0].data.String() != "! 617\n" || out[1].data.String() != "1000\n") {
				t.Fatal("normal protocol did not finish intact", r)
			}
			if name == "cleanup_failure" {
				if !e.tainted.Load() || len(containers) != 1 {
					t.Fatal("failed cleanup did not taint/preserve evidence")
				}
			} else if e.tainted.Load() || len(containers) != 0 {
				t.Fatal("interaction left a container or unexpected taint")
			}
		})
	}
}
