package filejudge

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"oj.local/oj/internal/problems"
	"oj.local/oj/internal/problemstore"
)

func TestConstructionMultipleAnswersAndFailures(t *testing.T) {
	for n := 1; n <= 1000; n++ {
		var values []int
		if n == 2 || n == 3 {
			values = []int{-1}
		} else {
			for x := 2; x <= n; x += 2 {
				values = append(values, x)
			}
			for x := 1; x <= n; x += 2 {
				values = append(values, x)
			}
		}
		for _, reverse := range []bool{false, true} {
			var tokens []string
			for i := range values {
				j := i
				if reverse {
					j = len(values) - 1 - i
				}
				tokens = append(tokens, strconv.Itoa(values[j]))
			}
			if v, _, err := constructionVerdict([]byte(strconv.Itoa(n)), []byte(strings.Join(tokens, " "))); err != nil || v != "AC" {
				t.Fatalf("n=%d reverse=%v v=%s err=%v", n, reverse, v, err)
			}
		}
	}
	for _, pair := range [][2]string{{"4", "1 2 3 4"}, {"4", "2 4 1 1"}, {"4", "2 4 1 5"}, {"4", "2 4 1"}, {"4", "2 4 1 3 extra"}, {"4", "2.0 4 1 3"}, {"4", "-1"}, {"2", "1 2"}, {"3", "-1 extra"}, {"1", "2"}} {
		v, feedback, err := constructionVerdict([]byte(pair[0]), []byte(pair[1]))
		if err != nil || v != "WA" || feedback == "" {
			t.Fatal(pair, v, feedback, err)
		}
	}
	if v, _, err := constructionVerdict([]byte("bad"), []byte("-1")); v != "SYSTEM_ERROR" || err == nil {
		t.Fatal("bad configuration became participant WA")
	}
}

func cppFixture(t *testing.T) (*Engine, string, string) {
	t.Helper()
	u := problemstore.Upload{Descriptor: problemstore.Descriptor{Title: "构造题", Summary: "构造排列", Difficulty: "基础", Description: "任意合法排列", InputDescription: "整数", OutputDescription: "排列", Samples: []problems.Sample{{Input: "4\n", Output: "2 4 1 3\n"}}, StarterCode: "int main(){}", Type: "batch", Checker: "cpp-special-v1", CheckerSource: "checker.cpp", WallTimeMS: 500, MemoryLimitBytes: 64 * 1048576, Cases: []problemstore.CaseFile{{ID: "sample", Input: "sample.in"}}}, Files: map[string][]byte{"sample.in": []byte("4\n"), "checker.cpp": []byte("int main(){return 0;}\n")}}
	root := t.TempDir()
	store := &problemstore.Store{UploadRoot: filepath.Join(root, "uploads")}
	p, err := store.Create(u)
	if err != nil {
		t.Fatal(err)
	}
	jobs := filepath.Join(root, "jobs")
	if os.Mkdir(jobs, 0700) != nil {
		t.Fatal("jobs setup")
	}
	return &Engine{store: store, jobsRoot: jobs, image: "sha256:" + strings.Repeat("a", 64), busy: make(chan struct{}, 1)}, p.ID, filepath.Join(store.UploadRoot, p.ID)
}

// Fake only the Docker transport; real snapshot, artifact files, stage lifecycle,
// safe argv, compile budgets, verdict order, cleanup and taint all remain in use.
func fakeSpecialDocker(t *testing.T, e *Engine, scenario string, participantOutput []byte) *int {
	t.Helper()
	checkerRuns := 0
	containers := map[string]containerInfo{}
	mounts := map[string]map[string]string{}
	phases := map[string]string{}
	e.command = func(ctx context.Context, args []string, out, stderr io.Writer) error {
		switch args[0] {
		case "create":
			s := stageSpec{Name: args[2], Mounts: map[string]string{}}
			for i := 1; i < len(args)-1; i++ {
				switch args[i] {
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
				case "--entrypoint":
					s.Entry = args[i+1]
				}
			}
			if s.Stage == "checker-compile" && (len(s.Mounts) != 1 || s.Entry != "/usr/local/bin/judge-compile-helper" || s.Memory != 512*1048576 || s.Pids != 64) {
				t.Fatal("unsafe checker compilation", s)
			}
			if strings.HasPrefix(s.Stage, "checker-case-") {
				if len(s.Mounts) != 4 || s.Entry != "/checker/program" || s.Memory != 64*1048576 || s.Pids != 32 || s.Mounts["/solution/program"] != "" || strings.Contains(strings.Join(args, " "), "--interactive") {
					t.Fatal("unsafe checker execution", s)
				}
				if strings.Join(args[len(args)-3:], " ") != "/data/input.txt /data/answer.txt /data/output.txt" {
					t.Fatal("checker args not fixed")
				}
			}
			c := fakeInteractionInfo(e, s)
			c.Config.OpenStdin = false
			c.HostConfig.PidsLimit = int64(s.Pids)
			containers[s.Name], mounts[s.Name], phases[s.Name] = c, s.Mounts, s.Stage
		case "inspect":
			b, _ := json.Marshal([]containerInfo{containers[args[1]]})
			_, _ = out.Write(b)
		case "ps":
			name := strings.TrimSuffix(strings.TrimPrefix(args[len(args)-1], "name=^/"), "$")
			if _, ok := containers[name]; ok {
				_, _ = io.WriteString(out, name)
			}
		case "start":
			name := args[len(args)-1]
			c := containers[name]
			phase := phases[name]
			c.State.StartedAt = "2026-10-03T00:00:00Z"
			if phase == "compile" || phase == "checker-compile" {
				deadline, ok := ctx.Deadline()
				remaining := time.Until(deadline)
				if !ok || (phase == "checker-compile" && remaining > 10*time.Second) || (phase == "compile" && remaining > 20*time.Second) {
					t.Fatal("compile budget not applied", phase, remaining)
				}
				if (phase == "compile" && scenario == "participant_ce") || (phase == "checker-compile" && scenario == "checker_ce") {
					c.State.ExitCode = 1
					_, _ = io.WriteString(stderr, "compile failed")
				} else {
					_, _ = io.WriteString(out, "fake-elf-"+phase)
				}
			} else if strings.HasPrefix(phase, "checker-case-") {
				checkerRuns++
				b, err := os.ReadFile(mounts[name]["/data/output.txt"])
				if err != nil || string(b) != string(participantOutput) {
					t.Fatal("checker did not receive raw output", len(b), err)
				}
				b, err = os.ReadFile(mounts[name]["/data/answer.txt"])
				if err != nil || len(b) != 0 {
					t.Fatal("missing expected not materialized empty", err)
				}
				switch scenario {
				case "checker_wa":
					c.State.ExitCode = 1
					_, _ = io.WriteString(stderr, strings.Repeat("错", 1000))
				case "checker_exit":
					c.State.ExitCode = 2
				case "checker_oom":
					c.State.ExitCode = 137
					c.State.OOMKilled = true
				case "checker_ole":
					c.State.Running = true
					_, _ = out.Write(make([]byte, MaxRawOutput+1))
				case "checker_timeout":
					c.State.Running = true
					containers[name] = c
					<-ctx.Done()
					return ctx.Err()
				case "checker_attach":
					containers[name] = c
					return errors.New("broken attach")
				}
			} else {
				_, _ = out.Write(participantOutput)
				switch scenario {
				case "participant_re":
					c.State.ExitCode = 2
				case "participant_oom":
					c.State.ExitCode = 137
					c.State.OOMKilled = true
				case "participant_ole":
					c.State.Running = true
					_, _ = out.Write(make([]byte, MaxRawOutput+1))
				case "participant_timeout":
					c.State.Running = true
					containers[name] = c
					<-ctx.Done()
					return ctx.Err()
				}
			}
			containers[name] = c
		case "kill":
			c := containers[args[1]]
			c.State.Running = false
			c.State.ExitCode = 137
			containers[args[1]] = c
		case "rm":
			name := args[len(args)-1]
			if scenario == "checker_cleanup" && strings.HasPrefix(phases[name], "checker-case-") {
				return errors.New("failed cleanup")
			}
			delete(containers, name)
		default:
			t.Fatalf("unexpected Docker operation %v", args)
		}
		return nil
	}
	return &checkerRuns
}

func TestSpecialJudgeLifecycleAndPriority(t *testing.T) {
	wants := map[string]string{"normal": "AC", "checker_wa": "WA", "checker_ce": "SYSTEM_ERROR", "checker_exit": "SYSTEM_ERROR", "checker_oom": "SYSTEM_ERROR", "checker_ole": "SYSTEM_ERROR", "checker_timeout": "SYSTEM_ERROR", "checker_attach": "SYSTEM_ERROR", "checker_cleanup": "SYSTEM_ERROR", "participant_ce": "CE", "participant_re": "RE", "participant_oom": "MLE", "participant_ole": "OLE", "participant_timeout": "TLE"}
	for scenario, want := range wants {
		t.Run(scenario, func(t *testing.T) {
			e, id, _ := cppFixture(t)
			snapshot, snapshotErr := e.load(id)
			if snapshotErr != nil {
				t.Fatal(snapshotErr)
			}
			raw := []byte(strings.Repeat("x", 70000))
			runs := fakeSpecialDocker(t, e, scenario, raw)
			r, err := e.Judge(context.Background(), Request{ProblemID: id, Language: "cpp", Source: "int main(){}"})
			if err != nil || r.Verdict != want {
				t.Fatalf("want %s got %+v err=%v", want, r, err)
			}
			if !validResult(r) {
				t.Fatalf("engine generated a result rejected by its strict client: %+v", r)
			}
			if !MatchesSnapshot(r, snapshot) {
				t.Fatal("engine response failed trusted snapshot binding", r)
			}
			wire, marshalErr := json.Marshal(r)
			if marshalErr != nil || !completeWire(wire) {
				t.Fatal("engine result missing required wire fields", marshalErr)
			}
			if r.CheckerSourceSHA256 != hash([]byte("int main(){return 0;}\n")) {
				t.Fatal("source not bound before participant compile")
			}
			if strings.HasPrefix(scenario, "participant_") && *runs != 0 {
				t.Fatal("checker executed after participant failure")
			}
			if scenario == "participant_ce" && r.CheckerCompile != nil {
				t.Fatal("checker compiled before participant artifact")
			}
			if scenario == "normal" && (len(r.Cases) != 1 || !r.Cases[0].OutputTruncated || r.Cases[0].Checker == nil || r.Cases[0].Checker.ExitCode == nil || *r.Cases[0].Checker.ExitCode != 0 || r.CheckerCompile == nil || r.CheckerCompile.Status != "success") {
				t.Fatal("incomplete checker evidence")
			}
			if scenario == "checker_wa" && (len(r.Cases[0].CheckerFeedback) > 1024 || !strings.HasPrefix(r.Cases[0].CheckerFeedback, "错")) {
				t.Fatal("feedback not bounded")
			}
			entries, readErr := os.ReadDir(e.jobsRoot)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if scenario == "checker_cleanup" {
				if !e.CleanupFailed() || len(entries) != 1 {
					t.Fatal("cleanup failure not tainted/preserved")
				}
				// These are trusted fake artifacts in this test's TempDir. Keep
				// the failed-job evidence until assertions finish, then clear the
				// Windows readonly attribute so testing can remove its directory.
				for _, relative := range []string{"program", filepath.Join("checker-build", "program")} {
					artifact := filepath.Join(e.jobsRoot, entries[0].Name(), relative)
					t.Cleanup(func() {
						if err := os.Chmod(artifact, 0600); err != nil {
							t.Errorf("restore fake artifact for test cleanup: %v", err)
						}
					})
					st, err := os.Lstat(artifact)
					if err != nil || !st.Mode().IsRegular() || st.Mode().Perm()&0200 != 0 {
						t.Fatal("failed-job artifact not preserved readonly", err)
					}
				}
			} else if e.CleanupFailed() || len(entries) != 0 {
				t.Fatal("job not cleaned")
			}
		})
	}
}

func TestSpecialMultiCaseSnapshotBindingAndFailurePrefix(t *testing.T) {
	for _, scenario := range []string{"normal", "participant_re", "checker_exit"} {
		t.Run(scenario, func(t *testing.T) {
			e, id, dir := cppFixture(t)
			snap, err := e.load(id)
			if err != nil {
				t.Fatal(err)
			}
			m := snap.Manifest
			m.Cases = append(m.Cases, CaseFile{ID: "second", Input: "second.in"})
			writeManifest(t, dir, m)
			if err := os.WriteFile(filepath.Join(dir, "second.in"), []byte("4\n"), 0600); err != nil {
				t.Fatal(err)
			}
			snap, err = e.load(id)
			if err != nil {
				t.Fatal(err)
			}
			runs := fakeSpecialDocker(t, e, scenario, []byte("2 4 1 3\n"))
			r, failure := e.Judge(context.Background(), Request{ProblemID: id, Language: "cpp", Source: "int main(){}"})
			if failure != nil || !validResult(r) || !MatchesSnapshot(r, snap) {
				t.Fatal("multi-case snapshot not bound", r, failure)
			}
			if scenario == "checker_exit" {
				if r.Verdict != "SYSTEM_ERROR" || len(r.Cases) != 1 || *runs != 1 {
					t.Fatal("checker failure not bound to exact prefix", r)
				}
			} else {
				if len(r.Cases) != 2 || r.Cases[0].ID != "sample" || r.Cases[1].ID != "second" {
					t.Fatal("successful runtime did not traverse ordered cases", r)
				}
				want, runsWant := "AC", 2
				if scenario == "participant_re" {
					want, runsWant = "RE", 0
				}
				if r.Verdict != want || *runs != runsWant {
					t.Fatal("participant priority or case count invalid", r, *runs)
				}
			}
		})
	}
}

func TestSpecialSnapshotSourceDigestAndReload(t *testing.T) {
	e, id, dir := cppFixture(t)
	one, err := e.load(id)
	if err != nil {
		t.Fatal(err)
	}
	if one.CheckerSource != "int main(){return 0;}\n" || one.Manifest.CheckerSource != "checker.cpp" {
		t.Fatal("source missing from snapshot")
	}
	if err := os.WriteFile(filepath.Join(dir, "checker.cpp"), []byte("int main(){return 1;}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	two, err := LoadWithStore(&problemstore.Store{UploadRoot: e.store.UploadRoot}, id)
	if err != nil || one.SHA256 == two.SHA256 || one.CheckerSource == two.CheckerSource {
		t.Fatal("source mutation absent from digest", err)
	}
	if one.CheckerSource != "int main(){return 0;}\n" {
		t.Fatal("old snapshot mutated")
	}
	if err := os.WriteFile(filepath.Join(dir, "checker.cpp"), []byte{0}, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := e.load(id); err == nil {
		t.Fatal("invalid checker source loaded")
	}
}

func TestMalformedConstructionDataStopsBeforeCompile(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "construct-permutation")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, dir, Manifest{Type: "batch", Checker: "nonconsecutive-permutation-v1", WallTimeMS: 2000, MemoryLimitBytes: 64 * 1048576, Cases: []CaseFile{{ID: "broken", Input: "broken.in"}}})
	if err := os.WriteFile(filepath.Join(dir, "broken.in"), []byte("4 5\n"), 0600); err != nil {
		t.Fatal(err)
	}
	e := &Engine{dataRoot: root, busy: make(chan struct{}, 1), command: func(context.Context, []string, io.Writer, io.Writer) error {
		t.Fatal("invalid configuration reached Docker")
		return nil
	}}
	r, err := e.Judge(context.Background(), Request{ProblemID: "construct-permutation", Language: "cpp", Source: "int main(){}"})
	if err != nil || r.Verdict != "SYSTEM_ERROR" || r.Compile.Status != "system_error" || len(r.Cases) != 0 {
		t.Fatal("configuration error became student result", r, err)
	}
}

func TestSpecialCheckerResourceErrorsRemainSystemErrors(t *testing.T) {
	for _, s := range []stageResult{{Failure: "cleanup_failed"}, {State: containerState{OOMKilled: true, ExitCode: 137}}, {Cause: "wall_timeout"}, {Cause: "output_limit"}, {State: containerState{ExitCode: 137}}} {
		if v, _ := checkerVerdict(s); v != "SYSTEM_ERROR" {
			t.Fatal("checker fault attributed to participant", s, v)
		}
	}
	for _, code := range []int{0, 1} {
		want := "AC"
		if code == 1 {
			want = "WA"
		}
		if v, _ := checkerVerdict(stageResult{State: containerState{ExitCode: code}}); v != want {
			t.Fatal(code, v)
		}
	}
}
