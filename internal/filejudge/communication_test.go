package filejudge

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"oj.local/oj/internal/problemstore"
)

func TestCommunicationPipeRawBytesAndEOF(t *testing.T) {
	p := newCommunicationPipe()
	raw := []byte{'a', 0, 0xff, '\n', 'b'}
	if n, err := p.Write(raw); err != nil || n != len(raw) {
		t.Fatal(n, err)
	}
	p.finish()
	read, err := io.ReadAll(p)
	if err != nil || !bytes.Equal(read, raw) {
		t.Fatal("raw bytes or EOF changed", read, err)
	}
	if _, err := p.Write([]byte("late")); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal("write after EOF accepted", err)
	}
	q := newCommunicationPipe()
	done := make(chan error, 1)
	go func() { _, err := q.Read(make([]byte, 1)); done <- err }()
	_ = q.Close()
	select {
	case err := <-done:
		if !errors.Is(err, io.ErrClosedPipe) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancelled read remained blocked")
	}
}

func TestCommunicationRelayCountsBytesAndStopsWithoutReceiverRead(t *testing.T) {
	for _, chunks := range [][][]byte{{[]byte("abc"), []byte("de")}, {[]byte("abcde")}, {[]byte{0, 0xff, 'a', 'b', 'c'}}} {
		ctx, cancel := context.WithCancel(context.Background())
		pipe := newCommunicationPipe()
		relay := &communicationRelay{pipe: pipe, limit: 4, retained: &boundedStream{limit: MaxFile}, cancel: cancel}
		for _, chunk := range chunks {
			_, _ = relay.Write(chunk)
		}
		if ctx.Err() == nil || !relay.exceeded.Load() || relay.observed.Load() != 5 || relay.transferred.Load() != 4 || pipe.data.Len() != 4 {
			t.Fatal("limit depended on characters/chunks/read", relay.observed.Load(), relay.transferred.Load(), pipe.data.Len())
		}
		cancel()
		_ = pipe.Close()
	}
	pipe := newCommunicationPipe()
	_ = pipe.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	relay := &communicationRelay{pipe: pipe, limit: 4, retained: &boundedStream{limit: MaxFile}, cancel: cancel}
	_, _ = relay.Write([]byte("abc"))
	if relay.transferred.Load() != 0 || relay.observed.Load() != 3 || ctx.Err() != nil || !relay.closedPeer.Load() {
		t.Fatal("early close fabricated transferred bytes")
	}
}

func TestCommunicationIsolationAndVerdictPriority(t *testing.T) {
	e := &Engine{image: "sha256:" + strings.Repeat("a", 64)}
	specs := e.communicationSpecs("run", "case-a", [2]string{"sender-bin", "receiver-bin"}, "hidden-input", Manifest{WallTimeMS: 5000, MemoryLimitBytes: 64 * 1048576})
	if len(specs[0].Mounts) != 2 || specs[0].Mounts["/input.txt"] != "hidden-input" || len(specs[1].Mounts) != 1 || specs[1].Mounts["/solution/program"] != "receiver-bin" || specs[1].Mounts["/input.txt"] != "" {
		t.Fatal("role isolation lost", specs)
	}
	for _, spec := range specs {
		args := strings.Join(e.createArgs(spec), " ")
		if !strings.Contains(args, "--pids-limit 32") || !strings.Contains(args, "--interactive") || strings.Contains(args, "/expected") || strings.Contains(args, "--network host") {
			t.Fatal(args)
		}
		c := fakeInteractionInfo(e, spec)
		c.HostConfig.PidsLimit = 32
		if err := e.verifyContainer(c, spec); err != nil {
			t.Fatal(err)
		}
		c.Config.Tty = true
		if e.verifyContainer(c, spec) == nil {
			t.Fatal("TTY accepted")
		}
	}
	for _, tc := range []struct {
		r    communicationStage
		want string
	}{
		{communicationStage{Failure: "cleanup_failed", Cause: "wall_timeout", Roles: [2]stageResult{{State: containerState{OOMKilled: true}}}}, "SYSTEM_ERROR"},
		{communicationStage{Cause: "communication_limit", Roles: [2]stageResult{{State: containerState{OOMKilled: true}}}}, "MLE"},
		{communicationStage{Cause: "wall_timeout", LimitExceeded: true}, "TLE"},
		{communicationStage{Cause: "output_limit", LimitExceeded: true}, "OLE"},
		{communicationStage{Cause: "communication_limit", Roles: [2]stageResult{{State: containerState{ExitCode: 137}}}}, "WA"},
		{communicationStage{Cause: "role_exit", Roles: [2]stageResult{{State: containerState{ExitCode: 2}}, {State: containerState{ExitCode: 137}}}, Killed: [2]bool{false, true}}, "RE"},
	} {
		if got := communicationVerdict(tc.r, []byte("3"), []byte("3")); got != tc.want {
			t.Fatal(tc.want, got, tc.r)
		}
	}
}

func TestCommunicationLifecycle(t *testing.T) {
	two, killed := trustedExitError(t, 2), trustedExitError(t, 137)
	wants := map[string]string{"normal": "AC", "wrong_answer": "WA", "raw_bytes": "AC", "exact_limit": "AC", "early_receiver": "AC", "early_stdin_closed": "AC", "sender_re": "RE", "receiver_re": "RE", "natural137": "RE", "sender_oom": "MLE", "receiver_oom": "MLE", "timeout": "TLE", "partial_start_timeout": "TLE", "partial_start_missing": "SYSTEM_ERROR", "communication_limit": "WA", "sender_stderr_limit": "OLE", "receiver_stdout_limit": "OLE", "receiver_stderr_limit": "OLE", "stdin_unknown": "SYSTEM_ERROR", "attach_incomplete": "SYSTEM_ERROR", "create_uncertain": "SYSTEM_ERROR", "create_deadline": "TLE", "cleanup_failure": "SYSTEM_ERROR", "cancelled": "SYSTEM_ERROR", "job_deadline": "SYSTEM_ERROR", "kill_race": "RE"}
	for name, want := range wants {
		t.Run(name, func(t *testing.T) {
			job, stopJob := context.WithCancel(context.Background())
			defer stopJob()
			if name == "job_deadline" {
				var stop context.CancelFunc
				job, stop = context.WithTimeout(context.Background(), 35*time.Millisecond)
				defer stop()
			}
			e := &Engine{image: "sha256:" + strings.Repeat("a", 64)}
			specs := e.communicationSpecs(strings.Repeat("b", 32), "case-a", [2]string{"sender", "receiver"}, "input", Manifest{WallTimeMS: 100, MemoryLimitBytes: 64 * 1048576})
			var mu sync.Mutex
			containers := map[string]containerInfo{}
			setState := func(container string, running bool, code int, oom bool) {
				mu.Lock()
				defer mu.Unlock()
				c := containers[container]
				c.State = containerState{Running: running, ExitCode: code, OOMKilled: oom, StartedAt: "2026-10-05T00:00:00Z"}
				containers[container] = c
			}
			deadlineTriggered := false
			e.command = func(ctx context.Context, args []string, out, _ io.Writer) error {
				mu.Lock()
				defer mu.Unlock()
				switch args[0] {
				case "create":
					for _, spec := range specs {
						if spec.Name == args[2] {
							c := fakeInteractionInfo(e, spec)
							c.HostConfig.PidsLimit = 32
							containers[spec.Name] = c
						}
					}
					if name == "create_deadline" && !deadlineTriggered {
						deadlineTriggered = true
						<-ctx.Done()
						return ctx.Err()
					}
					if name == "create_uncertain" && args[2] == specs[1].Name {
						return errors.New("lost create reply")
					}
				case "inspect":
					c, ok := containers[args[1]]
					if !ok {
						return errors.New("missing fake container")
					}
					data, _ := json.Marshal([]containerInfo{c})
					_, _ = out.Write(data)
				case "ps":
					container := strings.TrimSuffix(strings.TrimPrefix(args[len(args)-1], "name=^/"), "$")
					if name == "partial_start_missing" && container == specs[1].Name {
						delete(containers, container)
					}
					if _, ok := containers[container]; ok {
						_, _ = io.WriteString(out, container)
					}
				case "kill":
					c := containers[args[1]]
					c.State.Running = false
					c.State.ExitCode = 137
					containers[args[1]] = c
					if name == "kill_race" && args[1] == specs[1].Name {
						return errors.New("natural exit raced kill")
					}
				case "rm":
					if name == "cleanup_failure" && args[len(args)-1] == specs[1].Name {
						return errors.New("cleanup unavailable")
					}
					delete(containers, args[len(args)-1])
				default:
					t.Errorf("unexpected control %v", args)
				}
				return nil
			}
			started := [2]chan struct{}{make(chan struct{}), make(chan struct{})}
			payload := []byte("3\n")
			limit := int64(32)
			if name == "raw_bytes" {
				payload = []byte{'a', 0, 0xff, '\n'}
			}
			if name == "exact_limit" {
				limit = int64(len(payload))
			}
			e.duplex = func(ctx context.Context, args []string, input io.ReadCloser, out, diag io.Writer) error {
				container := args[len(args)-1]
				side := 0
				if container == specs[1].Name {
					side = 1
				}
				if (name == "partial_start_timeout" || name == "partial_start_missing") && side == 1 {
					close(started[1])
					<-started[0]
					<-ctx.Done()
					return ctx.Err()
				}
				setState(container, true, 0, false)
				close(started[side])
				<-started[1-side]
				if name == "timeout" || name == "partial_start_timeout" || name == "partial_start_missing" || name == "cancelled" || name == "job_deadline" {
					if name == "cancelled" {
						stopJob()
					}
					<-ctx.Done()
					return ctx.Err()
				}
				if name == "kill_race" && side == 1 {
					<-ctx.Done()
					return ctx.Err()
				}
				if name == "sender_re" && side == 0 || name == "receiver_re" && side == 1 || name == "kill_race" && side == 0 {
					setState(container, false, 2, false)
					return two
				}
				if name == "natural137" && side == 0 {
					setState(container, false, 137, false)
					return killed
				}
				if name == "sender_oom" && side == 0 || name == "receiver_oom" && side == 1 {
					setState(container, false, 137, true)
					return killed
				}
				if name == "stdin_unknown" && side == 1 {
					setState(container, false, 0, false)
					return &duplexInputError{}
				}
				if name == "early_stdin_closed" && side == 1 {
					_, _ = io.WriteString(out, "3\n")
					setState(container, false, 0, false)
					return &duplexInputError{pipeClosed: true}
				}
				if name == "early_receiver" && side == 1 {
					_, _ = io.WriteString(out, "3\n")
					setState(container, false, 0, false)
					return nil
				}
				if name == "attach_incomplete" && side == 1 {
					setState(container, false, 0, false)
					return errors.New("broken transport")
				}
				if name == "communication_limit" && side == 0 {
					_, _ = out.Write(make([]byte, 33))
					<-ctx.Done()
					return ctx.Err()
				}
				if name == "sender_stderr_limit" && side == 0 || name == "receiver_stderr_limit" && side == 1 {
					_, _ = diag.Write(make([]byte, MaxRawOutput+1))
					<-ctx.Done()
					return ctx.Err()
				}
				if name == "receiver_stdout_limit" && side == 1 {
					_, _ = out.Write(make([]byte, MaxRawOutput+1))
					<-ctx.Done()
					return ctx.Err()
				}
				if side == 0 {
					_, _ = out.Write(payload)
				} else {
					raw, err := io.ReadAll(input)
					if ctx.Err() != nil {
						return ctx.Err()
					}
					if len(raw) == 0 && err == nil && (name == "sender_re" || name == "sender_oom" || name == "natural137") {
						setState(container, false, 2, false)
						return two
					}
					if err != nil || !bytes.Equal(raw, payload) {
						t.Errorf("raw channel or EOF changed: %v %v", raw, err)
						return errors.New("invalid trusted pipe")
					}
					answer := "3\n"
					if name == "wrong_answer" {
						answer = "4\n"
					}
					_, _ = io.WriteString(out, answer)
				}
				setState(container, false, 0, false)
				return nil
			}
			out := [2]*boundedStream{{limit: MaxFile}, {limit: MaxRawOutput}}
			diag := [2]*boundedStream{{limit: MaxRawOutput}, {limit: MaxRawOutput}}
			r := e.communication(job, specs, limit, out, diag)
			if got := communicationVerdict(r, out[1].data.Bytes(), []byte("3\n")); got != want {
				t.Fatalf("want %s got %s: %+v", want, got, r)
			}
			if name == "normal" || name == "raw_bytes" || name == "exact_limit" {
				if r.BytesObserved != int64(len(payload)) || r.BytesTransferred != int64(len(payload)) || r.LimitExceeded || r.Killed[0] || r.Killed[1] {
					t.Fatal("normal byte/kill evidence invalid", r)
				}
			}
			if name == "communication_limit" && (r.BytesObserved != 33 || r.BytesTransferred > limit || !r.LimitExceeded) {
				t.Fatal("overlimit evidence invalid", r)
			}
			if name == "create_deadline" && (!r.StartTimedOut || r.Roles[0].State.StartedAt != "" || r.Roles[1].State.StartedAt != "") {
				t.Fatal("startup timeout fabricated runtime", r)
			}
			if name == "partial_start_timeout" && (!r.StartTimedOut || r.Roles[0].State.StartedAt == "" || r.Roles[1].State.StartedAt != "" || r.Killed[1]) {
				t.Fatal("partial startup timeout fabricated peer execution", r)
			}
			if name == "kill_race" && r.Killed[1] {
				t.Fatal("failed kill claimed confirmed termination", r)
			}
			mu.Lock()
			remaining := len(containers)
			mu.Unlock()
			if name == "cleanup_failure" {
				if !e.tainted.Load() || remaining != 1 {
					t.Fatal("cleanup not tainted/preserved")
				}
			} else if e.tainted.Load() || remaining != 0 {
				t.Fatal("container remained", remaining)
			}
		})
	}
}

func fakeCommunicationCompile(t *testing.T, e *Engine, scenario string) (*[]string, *int) {
	t.Helper()
	phases := []string{}
	runs := 0
	containers := map[string]containerInfo{}
	specs := map[string]stageSpec{}
	e.command = func(ctx context.Context, args []string, out, diag io.Writer) error {
		switch args[0] {
		case "create":
			spec := stageSpec{Name: args[2], Mounts: map[string]string{}}
			for i := 1; i < len(args)-1; i++ {
				switch args[i] {
				case "--label":
					if strings.HasPrefix(args[i+1], "oj.judge.run=") {
						spec.RunID = strings.TrimPrefix(args[i+1], "oj.judge.run=")
					}
					if strings.HasPrefix(args[i+1], "oj.judge.stage=") {
						spec.Stage = strings.TrimPrefix(args[i+1], "oj.judge.stage=")
					}
				case "--memory":
					spec.Memory, _ = strconv.ParseInt(args[i+1], 10, 64)
				case "--pids-limit":
					spec.Pids, _ = strconv.Atoi(args[i+1])
				case "--mount":
					fields, err := csv.NewReader(strings.NewReader(args[i+1])).Read()
					if err != nil {
						t.Fatal(err)
					}
					spec.Mounts[strings.TrimPrefix(fields[2], "dst=")] = strings.TrimPrefix(fields[1], "src=")
				}
			}
			phases = append(phases, spec.Stage)
			c := fakeInteractionInfo(e, spec)
			c.Config.OpenStdin = false
			c.HostConfig.PidsLimit = int64(spec.Pids)
			containers[spec.Name], specs[spec.Name] = c, spec
		case "inspect":
			data, _ := json.Marshal([]containerInfo{containers[args[1]]})
			_, _ = out.Write(data)
		case "ps":
			name := strings.TrimSuffix(strings.TrimPrefix(args[len(args)-1], "name=^/"), "$")
			if _, ok := containers[name]; ok {
				_, _ = io.WriteString(out, name)
			}
		case "start":
			name := args[len(args)-1]
			c := containers[name]
			spec := specs[name]
			c.State.StartedAt = "2026-10-05T00:00:00Z"
			source, err := os.ReadFile(spec.Mounts["/source/solution.cpp"])
			if err != nil {
				t.Fatal(err)
			}
			role := strings.TrimSuffix(spec.Stage, "-compile")
			if string(source) != "trusted-"+role {
				t.Fatal("source bound by array order", role, string(source))
			}
			if spec.Memory != 512*1048576 || spec.Pids != 64 {
				t.Fatal("compile limits", spec)
			}
			if scenario == role+"_ce" {
				c.State.ExitCode = 1
				_, _ = io.WriteString(diag, role+" compile error")
			} else if scenario == role+"_sys" {
				c.State.ExitCode = 2
			} else {
				_, _ = io.WriteString(out, "fake-elf-"+role)
			}
			containers[name] = c
			if scenario == "receiver_build_failure" && role == "sender" {
				if err := os.Mkdir(filepath.Join(filepath.Dir(spec.Mounts["/source/solution.cpp"]), "..", "receiver-build"), 0700); err != nil {
					t.Fatal(err)
				}
			}
		case "rm":
			delete(containers, args[len(args)-1])
		default:
			t.Fatalf("runtime reached after failed compile: %v", args)
		}
		return nil
	}
	e.duplex = func(context.Context, []string, io.ReadCloser, io.Writer, io.Writer) error {
		runs++
		t.Fatal("failed compile launched role")
		return nil
	}
	return &phases, &runs
}

func TestCommunicationCompileFailureOrderAndSourceBinding(t *testing.T) {
	for _, scenario := range []string{"sender_ce", "receiver_ce", "sender_sys", "receiver_sys", "receiver_build_failure"} {
		t.Run(scenario, func(t *testing.T) {
			root := filepath.Join("..", "..", "testdata", "problems")
			e := &Engine{dataRoot: root, jobsRoot: t.TempDir(), image: "sha256:" + strings.Repeat("a", 64), busy: make(chan struct{}, 1)}
			phases, runs := fakeCommunicationCompile(t, e, scenario)
			req := Request{ProblemID: "communication-sum", SourceUnits: []SourceUnit{{Role: "receiver", Language: "cpp", Source: "trusted-receiver"}, {Role: "sender", Language: "cpp", Source: "trusted-sender"}}}
			r, failure := e.Judge(context.Background(), req)
			want := "SYSTEM_ERROR"
			if strings.HasSuffix(scenario, "_ce") {
				want = "CE"
			}
			if failure != nil || r.Verdict != want || len(r.Cases) != 0 || r.ArtifactSHA256 != "" || *runs != 0 {
				t.Fatal("compile failure order/aggregate", r, failure)
			}
			count := 1
			if strings.HasPrefix(scenario, "receiver_") && scenario != "receiver_build_failure" {
				count = 2
			}
			if len(r.RoleCompiles) != count || r.RoleCompiles[0].Role != "sender" || r.RoleCompiles[0].SourceSHA256 != hash([]byte("trusted-sender")) {
				t.Fatal("compile role evidence", r)
			}
			if count == 2 && (r.RoleCompiles[1].Role != "receiver" || r.RoleCompiles[1].SourceSHA256 != hash([]byte("trusted-receiver"))) {
				t.Fatal("receiver binding lost", r)
			}
			if len(*phases) != count {
				t.Fatal("extra compile launched after first failed role", *phases)
			}
			for i, phase := range *phases {
				if phase != r.RoleCompiles[i].Role+"-compile" {
					t.Fatal("compile command identity differs from recorded role", *phases)
				}
			}
			snap, err := e.load(req.ProblemID)
			if err != nil {
				t.Fatal(err)
			}
			wire, err := json.Marshal(r)
			if err != nil || !completeWire(wire) || !validResult(r) || !MatchesSnapshot(r, snap) || !MatchesRequest(r, req) {
				t.Fatal("compile result violates strict contract", r, err)
			}
			entries, err := os.ReadDir(e.jobsRoot)
			if err != nil || len(entries) != 0 || e.tainted.Load() {
				t.Fatal("failed compile not cleaned", err)
			}
		})
	}
}

func TestCommunicationSnapshotIncludesLimitsAndImmutableData(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "communication-sum")
	if os.Mkdir(dir, 0700) != nil {
		t.Fatal("fixture directory")
	}
	m := Manifest{Type: "communication", Checker: "communication-v1", WallTimeMS: 1000, MemoryLimitBytes: 64 * 1048576, Communication: &problemstore.CommunicationConfig{ByteLimit: 32}, Cases: []CaseFile{{ID: "sample", Input: "sample.in", Expected: "sample.out"}}}
	writeManifest(t, dir, m)
	if os.WriteFile(filepath.Join(dir, "sample.in"), []byte("2\n1 2\n"), 0600) != nil || os.WriteFile(filepath.Join(dir, "sample.out"), []byte("3\n"), 0600) != nil {
		t.Fatal("fixture data")
	}
	first, err := Load(root, "communication-sum")
	if err != nil {
		t.Fatal(err)
	}
	m.Communication.ByteLimit = 16
	writeManifest(t, dir, m)
	second, err := Load(root, "communication-sum")
	if err != nil || first.SHA256 == second.SHA256 || first.Manifest.Communication.ByteLimit != 32 {
		t.Fatal("limit absent or mutable snapshot", err)
	}
	if os.WriteFile(filepath.Join(dir, "sample.out"), []byte("4\n"), 0600) != nil {
		t.Fatal("data mutation")
	}
	third, err := Load(root, "communication-sum")
	if err != nil || third.SHA256 == second.SHA256 || string(first.Cases[0].Expected) != "3\n" {
		t.Fatal("answer absent or mutable snapshot", err)
	}
}

func TestCommunicationWorkerRoutesAndTrustedMetadata(t *testing.T) {
	e := &Engine{dataRoot: filepath.Join("..", "..", "testdata", "problems"), image: "sha256:" + strings.Repeat("a", 64), command: func(context.Context, []string, io.Writer, io.Writer) error {
		t.Fatal("invalid route reached Docker")
		return nil
	}}
	for _, tc := range []struct{ path, body string }{
		{"/judge", `{"problem_id":"communication-sum","language":"cpp","source":"int main(){}"}`},
		{"/judge", `{"problem_id":"a-plus-b","language":"cpp","source":"int main(){}","source_units":[]}`},
		{"/internal/communication-judge", `{"problem_id":"a-plus-b","source_units":[{"role":"sender","language":"cpp","source":"int main(){}"},{"role":"receiver","language":"cpp","source":"int main(){}"}]}`},
		{"/internal/communication-judge", `{"problem_id":"communication-sum","source_units":[{"role":"sender","language":"cpp","source":"int main(){}"},{"role":"sender","language":"cpp","source":"int main(){}"}]}`},
	} {
		request := httptest.NewRequest("POST", tc.path, strings.NewReader(tc.body))
		request.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		e.Handler().ServeHTTP(w, request)
		if w.Code != 400 {
			t.Fatal("endpoint mixed source shapes/types", tc.path, w.Code, w.Body)
		}
	}
	for _, id := range []string{"communication-sum", "communication-maximum", "communication-a-plus-b"} {
		w := httptest.NewRecorder()
		e.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/problem/"+id, nil))
		var meta Metadata
		if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &meta) != nil || meta.ProblemType != "communication" || meta.Checker != "communication-v1" || meta.CaseCount != 4 || meta.Communication == nil || len(meta.Communication.Roles) != 2 || meta.Communication.Roles[0] != "sender" || meta.Communication.Roles[1] != "receiver" || meta.Communication.AggregateMemoryLimitBytes != 2*meta.MemoryLimitBytes {
			t.Fatal("invalid trusted communication metadata", w.Code, w.Body)
		}
		limit := int64(32)
		if id == "communication-maximum" {
			limit = 16
		}
		if meta.Communication.ByteLimit != limit {
			t.Fatal("wrong byte limit")
		}
	}
}

// The fake daemon owns only trusted Go data. No submitted C++ or Docker CLI is
// executed; mount and state facts independently exercise the complete pipeline.
func TestCommunicationJudgeSuccessfulBindingsAndRuntimeFailures(t *testing.T) {
	two, killed := trustedExitError(t, 2), trustedExitError(t, 137)
	for _, scenario := range []string{"normal", "wrong_answer", "sender_re", "receiver_oom", "communication_limit", "receiver_output_limit", "shared_timeout", "combined_diagnostics", "pair_sum", "pair_wrong"} {
		t.Run(scenario, func(t *testing.T) {
			root := t.TempDir()
			problemID := "communication-sum"
			data := "2\n1 2\n"
			if strings.HasPrefix(scenario, "pair_") {
				problemID = "communication-a-plus-b"
				data = "1 2\n"
			}
			dir := filepath.Join(root, problemID)
			if err := os.Mkdir(dir, 0700); err != nil {
				t.Fatal(err)
			}
			writeManifest(t, dir, Manifest{Type: "communication", Checker: "communication-v1", WallTimeMS: 500, MemoryLimitBytes: 64 * 1048576, Communication: &problemstore.CommunicationConfig{ByteLimit: 32}, Cases: []CaseFile{{ID: "sample", Input: "sample.in", Expected: "sample.out"}}})
			if os.WriteFile(filepath.Join(dir, "sample.in"), []byte(data), 0600) != nil || os.WriteFile(filepath.Join(dir, "sample.out"), []byte("3\n"), 0600) != nil {
				t.Fatal("fixture write")
			}
			e := &Engine{dataRoot: root, jobsRoot: t.TempDir(), image: "sha256:" + strings.Repeat("a", 64), busy: make(chan struct{}, 1)}
			var mu sync.Mutex
			containers := map[string]containerInfo{}
			specs := map[string]stageSpec{}
			setState := func(name string, running bool, code int, oom bool) {
				mu.Lock()
				defer mu.Unlock()
				c := containers[name]
				c.State = containerState{Running: running, ExitCode: code, OOMKilled: oom, StartedAt: "2026-10-05T00:00:00Z"}
				containers[name] = c
			}
			e.command = func(_ context.Context, args []string, out, diag io.Writer) error {
				mu.Lock()
				defer mu.Unlock()
				switch args[0] {
				case "create":
					s := stageSpec{Name: args[2], Mounts: map[string]string{}}
					for i := 1; i < len(args)-1; i++ {
						switch args[i] {
						case "--interactive":
							s.Interactive = true
						case "--label":
							if strings.HasPrefix(args[i+1], "oj.judge.run=") {
								s.RunID = strings.TrimPrefix(args[i+1], "oj.judge.run=")
							}
							if strings.HasPrefix(args[i+1], "oj.judge.stage=") {
								s.Stage = strings.TrimPrefix(args[i+1], "oj.judge.stage=")
							}
						case "--memory":
							s.Memory, _ = strconv.ParseInt(args[i+1], 10, 64)
						case "--pids-limit":
							s.Pids, _ = strconv.Atoi(args[i+1])
						case "--mount":
							fields, err := csv.NewReader(strings.NewReader(args[i+1])).Read()
							if err != nil {
								t.Fatal(err)
							}
							s.Mounts[strings.TrimPrefix(fields[2], "dst=")] = strings.TrimPrefix(fields[1], "src=")
						}
					}
					c := fakeInteractionInfo(e, s)
					c.Config.OpenStdin = s.Interactive
					c.HostConfig.PidsLimit = int64(s.Pids)
					containers[s.Name], specs[s.Name] = c, s
				case "inspect":
					c, ok := containers[args[1]]
					if !ok {
						return errors.New("missing fake container")
					}
					raw, _ := json.Marshal([]containerInfo{c})
					_, _ = out.Write(raw)
				case "ps":
					name := strings.TrimSuffix(strings.TrimPrefix(args[len(args)-1], "name=^/"), "$")
					if _, ok := containers[name]; ok {
						_, _ = io.WriteString(out, name)
					}
				case "start":
					name := args[len(args)-1]
					spec := specs[name]
					if !strings.HasSuffix(spec.Stage, "-compile") {
						t.Fatal("communication must use duplex attach")
					}
					role := strings.TrimSuffix(spec.Stage, "-compile")
					data, err := os.ReadFile(spec.Mounts["/source/solution.cpp"])
					if err != nil || string(data) != "trusted-"+role {
						t.Fatal("role source lost", err)
					}
					_, _ = io.WriteString(out, "fake-elf-"+role)
					c := containers[name]
					c.State.StartedAt = "2026-10-05T00:00:00Z"
					containers[name] = c
				case "kill":
					c := containers[args[1]]
					c.State.Running, c.State.ExitCode = false, 137
					containers[args[1]] = c
				case "rm":
					delete(containers, args[len(args)-1])
				default:
					t.Fatal("unexpected control", args)
				}
				return nil
			}
			started := [2]chan struct{}{make(chan struct{}), make(chan struct{})}
			e.duplex = func(ctx context.Context, args []string, input io.ReadCloser, out, diag io.Writer) error {
				name := args[len(args)-1]
				mu.Lock()
				spec := specs[name]
				mu.Unlock()
				side := 0
				if strings.HasSuffix(spec.Stage, "-receiver") {
					side = 1
				}
				if side == 0 && (len(spec.Mounts) != 2 || spec.Mounts["/input.txt"] == "") || side == 1 && (len(spec.Mounts) != 1 || spec.Mounts["/input.txt"] != "") {
					t.Fatal("input or answer exposed", spec)
				}
				setState(name, true, 0, false)
				close(started[side])
				<-started[1-side]
				if scenario == "shared_timeout" {
					<-ctx.Done()
					return ctx.Err()
				}
				if scenario == "sender_re" && side == 0 {
					setState(name, false, 2, false)
					return two
				}
				if scenario == "receiver_oom" && side == 1 {
					setState(name, false, 137, true)
					return killed
				}
				if scenario == "communication_limit" && side == 0 {
					_, _ = out.Write(make([]byte, 33))
					<-ctx.Done()
					return ctx.Err()
				}
				if scenario == "receiver_output_limit" && side == 1 {
					_, _ = out.Write(make([]byte, MaxRawOutput+1))
					<-ctx.Done()
					return ctx.Err()
				}
				if scenario == "combined_diagnostics" {
					_, _ = io.WriteString(diag, strings.Repeat("d", 12000))
				}
				if side == 0 {
					raw, err := os.ReadFile(spec.Mounts["/input.txt"])
					if err != nil {
						t.Fatal(err)
					}
					values := bytes.NewReader(raw)
					if strings.HasPrefix(scenario, "pair_") {
						var a, b int64
						if _, err := fmt.Fscan(values, &a, &b); err != nil {
							t.Fatal(err)
						}
						_, _ = fmt.Fprintln(out, a, b)
						setState(name, false, 0, false)
						return nil
					}
					var n int
					if _, err := fmt.Fscan(values, &n); err != nil {
						t.Fatal(err)
					}
					var answer, x int64
					for i := 0; i < n; i++ {
						if _, err := fmt.Fscan(values, &x); err != nil {
							t.Fatal(err)
						}
						answer += x
					}
					_, _ = fmt.Fprintln(out, answer)
				} else {
					message, err := io.ReadAll(input)
					if ctx.Err() != nil {
						return ctx.Err()
					}
					if err != nil {
						return err
					}
					if scenario == "sender_re" && len(message) == 0 {
						setState(name, false, 2, false)
						return two
					}
					if scenario == "wrong_answer" {
						message = []byte("4\n")
					}
					if strings.HasPrefix(scenario, "pair_") {
						if string(message) != "1 2\n" {
							t.Fatal("sender precomputed answer or lost raw pair/EOF", string(message))
						}
						var a, b int64
						if _, err := fmt.Fscan(bytes.NewReader(message), &a, &b); err != nil {
							t.Fatal(err)
						}
						answer := a + b
						if scenario == "pair_wrong" {
							answer = 5
						}
						message = []byte(fmt.Sprintln(answer))
					}
					_, _ = out.Write(message)
				}
				setState(name, false, 0, false)
				return nil
			}
			request := Request{ProblemID: problemID, SourceUnits: []SourceUnit{{Role: "receiver", Language: "cpp", Source: "trusted-receiver"}, {Role: "sender", Language: "cpp", Source: "trusted-sender"}}}
			r, failure := e.Judge(context.Background(), request)
			want := map[string]string{"normal": "AC", "wrong_answer": "WA", "sender_re": "RE", "receiver_oom": "MLE", "communication_limit": "WA", "receiver_output_limit": "OLE", "shared_timeout": "TLE", "combined_diagnostics": "AC", "pair_sum": "AC", "pair_wrong": "WA"}[scenario]
			if failure != nil || r.Verdict != want {
				t.Fatal("runtime verdict", want, r, failure)
			}
			snap, err := e.load(request.ProblemID)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(r)
			if err != nil || !completeWire(raw) || !validResult(r) || !MatchesRequest(r, request) || !MatchesSnapshot(r, snap) {
				t.Fatal("pipeline facts violate wire/request/snapshot binding", r, err)
			}
			if len(r.RoleCompiles) != 2 || r.RoleCompiles[0].ArtifactSHA256 != hash([]byte("fake-elf-sender")) || r.RoleCompiles[1].ArtifactSHA256 != hash([]byte("fake-elf-receiver")) || r.ArtifactSHA256 != CommunicationArtifactDigest(r.RoleCompiles) {
				t.Fatal("artifact role binding lost", r)
			}
			if strings.HasPrefix(scenario, "pair_") {
				communication := r.Cases[0].Communication
				answer := "3\n"
				if scenario == "pair_wrong" {
					answer = "5\n"
				}
				if communication.Sender.Stdout != "1 2\n" || communication.Receiver.Stdout != answer || communication.BytesObserved != 4 || communication.BytesTransferred != 4 || communication.LimitExceeded || communication.ByteLimit != 32 {
					t.Fatal("simple pair communication evidence incorrect", r.Cases[0])
				}
			}
			entries, err := os.ReadDir(e.jobsRoot)
			if err != nil || len(entries) != 0 || e.tainted.Load() {
				t.Fatal("job cleanup", err)
			}
		})
	}
}
