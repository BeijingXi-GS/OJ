package filejudge

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMountCSV(t *testing.T) {
	arg := mountArg(`D:/test, with space/program`, "/solution/program")
	fields, err := csv.NewReader(strings.NewReader(arg)).Read()
	if err != nil || len(fields) != 4 || fields[1] != `src=D:/test, with space/program` || fields[3] != "readonly" {
		t.Fatal(fields, err)
	}
}
func TestSharedOutputBudget(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	budget := &outputBudget{cancel: cancel, limit: MaxRawOutput}
	a, b := &boundedStream{budget: budget, limit: MaxRawOutput}, &boundedStream{budget: budget, limit: MaxRawOutput}
	_, _ = a.Write(make([]byte, 700000))
	_, _ = b.Write(make([]byte, 700000))
	if !budget.fired.Load() || ctx.Err() == nil || a.data.Len()+b.data.Len() > MaxRawOutput {
		t.Fatal("shared stop/retention budget ineffective")
	}
}
func TestStageLifecycle(t *testing.T) {
	for _, name := range []string{"normal", "create_uncertain", "attach_incomplete", "cancel_in_cleanup", "deadline_in_cleanup", "cleanup_failure", "timeout", "output_limit"} {
		t.Run(name, func(t *testing.T) {
			job, cancel := context.WithCancel(context.Background())
			defer cancel()
			if name == "deadline_in_cleanup" {
				var stop context.CancelFunc
				job, stop = context.WithTimeout(context.Background(), 30*time.Millisecond)
				defer stop()
			}
			e := &Engine{image: "sha256:" + strings.Repeat("a", 64), busy: make(chan struct{}, 1)}
			s := stageSpec{RunID: strings.Repeat("b", 32), Name: stageName(strings.Repeat("b", 32), "case-a"), Stage: "case-a", Memory: 64 * 1048576, Pids: 16, Wall: 5 * time.Millisecond, Mounts: map[string]string{}, Entry: "/bin/sh"}
			var c containerInfo
			c.ID = "fake-container"
			c.Image = e.image
			c.Config.User = "10002:10002"
			c.Config.Labels = map[string]string{"oj.judge.owner": "file-mvp-v1", "oj.judge.run": s.RunID}
			h := &c.HostConfig
			h.ReadonlyRootfs = true
			h.NetworkMode = "none"
			h.IpcMode = "private"
			h.CgroupnsMode = "private"
			h.Memory = s.Memory
			h.MemorySwap = s.Memory
			h.NanoCpus = 1000000000
			h.PidsLimit = 16
			h.CapDrop = []string{"ALL"}
			h.SecurityOpt = []string{"no-new-privileges"}
			init := true
			h.Init = &init
			h.LogConfig.Type = "none"
			present, attached := false, false
			e.command = func(ctx context.Context, args []string, out, stderr io.Writer) error {
				switch args[0] {
				case "create":
					present = true
					if name == "create_uncertain" {
						return errors.New("response lost")
					}
					_, _ = io.WriteString(out, c.ID)
				case "inspect":
					data, _ := json.Marshal([]containerInfo{c})
					_, _ = out.Write(data)
				case "ps":
					if attached && name == "cancel_in_cleanup" {
						cancel()
					}
					if attached && name == "deadline_in_cleanup" {
						time.Sleep(40 * time.Millisecond)
					}
					if present {
						_, _ = io.WriteString(out, c.ID)
					}
				case "start":
					attached = true
					c.State.StartedAt = "2026-10-02T00:00:00Z"
					c.State.Running = false
					if name == "timeout" {
						c.State.Running = true
						<-ctx.Done()
						return ctx.Err()
					}
					if name == "output_limit" {
						_, _ = out.Write(make([]byte, MaxRawOutput+1))
						c.State.Running = true
						return ctx.Err()
					}
					_, _ = io.WriteString(out, "3\n")
					if name == "attach_incomplete" {
						return errors.New("broken transport")
					}
				case "kill":
					c.State.Running = false
					c.State.ExitCode = 137
				case "rm":
					if name == "cleanup_failure" {
						return errors.New("daemon unavailable")
					}
					present = false
				default:
					t.Fatalf("unexpected command %v", args)
				}
				return nil
			}
			out, stderr := &boundedStream{limit: MaxRawOutput}, &boundedStream{limit: MaxRawOutput}
			r := e.stage(job, s, out, stderr)
			want := ""
			switch name {
			case "create_uncertain":
				want = "container_create_failed"
			case "attach_incomplete":
				want = "attach_incomplete"
			case "cancel_in_cleanup":
				want = "request_cancelled"
			case "deadline_in_cleanup":
				want = "job_timeout"
			case "cleanup_failure":
				want = "cleanup_failed"
			}
			if r.Failure != want {
				t.Fatalf("want %q got %+v", want, r)
			}
			if name == "cleanup_failure" {
				if !e.tainted.Load() || !present {
					t.Fatal("cleanup failure must taint and preserve evidence")
				}
				_, err := e.Judge(context.Background(), Request{ProblemID: "a-plus-b", Language: "cpp", Source: "int main(){}"})
				if err == nil || err.Status != 503 {
					t.Fatal("tainted worker accepted work")
				}
			} else if present {
				t.Fatal("container remained")
			}
			if name == "timeout" && r.Cause != "wall_timeout" {
				t.Fatal(r)
			}
			if name == "output_limit" && r.Cause != "output_limit" {
				t.Fatal(r)
			}
		})
	}
}
func TestStartupRefusesInvalidProfile(t *testing.T) {
	dir := t.TempDir()
	profile := dir + "/profile.json"
	_ = os.WriteFile(profile, []byte(`{"version":"judge-mvp/0.1.0","image":"mutable-tag","reference_commit":"4502d2a798f3130d3c540c5c31d45251ca103623"}`), 0600)
	if _, err := NewEngine(profile, dir, dir+"/jobs"); err == nil {
		t.Fatal("mutable profile accepted")
	}
}
