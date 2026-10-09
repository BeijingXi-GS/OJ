package filejudge

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"oj.local/oj/internal/compilation"
	"oj.local/oj/internal/problems"
	"oj.local/oj/internal/problemstore"
)

type Profile struct {
	Version         string `json:"version"`
	Image           string `json:"image"`
	ReferenceCommit string `json:"reference_commit"`
}
type Engine struct {
	image, dataRoot, jobsRoot string
	command                   commandFunc
	duplex                    duplexCommandFunc
	store                     *problemstore.Store
	tainted                   atomic.Bool
	busy                      chan struct{}
}

// CleanupFailed reports a retained cleanup failure to the service lifecycle.
// A false value is not a substitute for the supervisor's actual residue checks.
func (e *Engine) CleanupFailed() bool { return e.tainted.Load() }

func NewEngine(profilePath, dataRoot, jobsRoot string) (*Engine, error) {
	file, err := os.Open(profilePath)
	if err != nil {
		return nil, errors.New("profile_missing")
	}
	defer file.Close()
	d := json.NewDecoder(io.LimitReader(file, 8193))
	d.DisallowUnknownFields()
	var p Profile
	var extra any
	if d.Decode(&p) != nil || d.Decode(&extra) != io.EOF || p.Version != "judge-mvp/0.1.0" || p.ReferenceCommit != "4502d2a798f3130d3c540c5c31d45251ca103623" || !strings.HasPrefix(p.Image, "sha256:") || !digestRE.MatchString(strings.TrimPrefix(p.Image, "sha256:")) {
		return nil, errors.New("profile_invalid")
	}
	dataRoot, err = filepath.Abs(dataRoot)
	if err != nil {
		return nil, errors.New("data_root_invalid")
	}
	jobsRoot, err = filepath.Abs(jobsRoot)
	if err != nil {
		return nil, errors.New("jobs_root_invalid")
	}
	if err = os.MkdirAll(jobsRoot, 0700); err != nil {
		return nil, errors.New("jobs_root_create_failed")
	}
	info, err := os.Lstat(jobsRoot)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("jobs_root_invalid")
	}
	uploadRoot := os.Getenv("PROBLEM_UPLOAD_ROOT")
	if uploadRoot == "" {
		uploadRoot = ".local/teacher-problems"
	}
	return &Engine{image: p.Image, dataRoot: dataRoot, jobsRoot: jobsRoot, command: dockerCommand, duplex: dockerDuplexCommand, store: &problemstore.Store{BuiltinRoot: dataRoot, UploadRoot: uploadRoot}, busy: make(chan struct{}, 1)}, nil
}
func newID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	return hex.EncodeToString(b), err
}
func (e *Engine) removeJob(dir string) error {
	if e.tainted.Load() {
		return errors.New("cleanup_tainted")
	}
	rel, err := filepath.Rel(e.jobsRoot, dir)
	if err != nil || filepath.Dir(rel) != "." || !strings.HasPrefix(rel, "run-") || filepath.Clean(filepath.Dir(dir)) != filepath.Clean(e.jobsRoot) {
		return errors.New("cleanup_path_invalid")
	}
	return removeJobFiles(e.jobsRoot, rel)
}
func (e *Engine) compile(ctx context.Context, run, dir, source string) (CompileResult, string, string, error) {
	return e.compileNamed(ctx, run, dir, source, "compile")
}
func (e *Engine) compileNamed(ctx context.Context, run, dir, source, phase string) (CompileResult, string, string, error) {
	return e.compileNamedBudget(ctx, run, dir, source, phase, 20*time.Second)
}
func (e *Engine) compileNamedBudget(ctx context.Context, run, dir, source, phase string, wall time.Duration) (CompileResult, string, string, error) {
	r := CompileResult{Status: "system_error"}
	sourcePath := filepath.Join(dir, "source.cpp")
	artifactPath := filepath.Join(dir, "program")
	if os.WriteFile(sourcePath, []byte(source), 0644) != nil {
		return r, "", "", errors.New("snapshot_write_failed")
	}
	artifact, err := os.OpenFile(artifactPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return r, "", "", errors.New("artifact_write_failed")
	}
	out := &boundedStream{limit: MaxArtifact, file: artifact}
	diagnostics := &boundedStream{limit: MaxRawOutput}
	spec := stageSpec{RunID: run, Name: stageName(run, phase), Stage: phase, Memory: 512 * 1048576, Pids: 64, Wall: wall, Compile: true, Mounts: map[string]string{"/source/solution.cpp": sourcePath}, Entry: "/usr/local/bin/judge-compile-helper"}
	stage := e.stage(ctx, spec, out, diagnostics)
	closeErr := artifact.Close()
	r.DurationMS = stage.DurationMS
	r.Diagnostics, r.DiagnosticsTruncated = display(diagnostics.data.Bytes(), MaxFile)
	if stage.Failure != "" {
		return r, "", "", errors.New(stage.Failure)
	}
	if closeErr != nil || stage.Cause != "" || stage.State.OOMKilled {
		return r, "", "", errors.New("compile_resource_or_io_limit")
	}
	if stage.State.ExitCode == 1 {
		r.Status = "compile_error"
		return r, "", "", nil
	}
	if stage.State.ExitCode != 0 {
		return r, "", "", errors.New("compile_helper_failed")
	}
	st, err := os.Lstat(artifactPath)
	if err != nil || !st.Mode().IsRegular() || st.Size() == 0 || st.Size() > MaxArtifact || out.count != st.Size() {
		return r, "", "", errors.New("artifact_invalid")
	}
	data, err := os.ReadFile(artifactPath)
	if err != nil || os.Chmod(artifactPath, 0555) != nil {
		return r, "", "", errors.New("artifact_finalize_failed")
	}
	r.Status = "success"
	return r, artifactPath, hash(data), nil
}
func failureMessage(reason string) string {
	switch reason {
	case "job_timeout":
		return "判题任务超过总体处理预算，已执行清理。"
	case "request_cancelled":
		return "判题请求已取消，已执行清理。"
	case "cleanup_failed":
		return "隔离环境清理未确认，服务已停止接收请求，请检查问题记录。"
	case "testdata_load_failed":
		return "题目测试文件或 manifest 无效，请检查服务器配置。"
	case "compile_resource_or_io_limit":
		return "编译阶段超过资源限制或产物读写失败。"
	}
	return "判题基础设施处理失败（" + reason + "）。"
}
func (e *Engine) Judge(parent context.Context, req Request) (result Result, httpErr *compilation.Error) {
	if e.store != nil {
		if err := ValidateRequestWithStore(req, e.store); err != nil {
			return result, err
		}
	} else if err := ValidateRequest(req); err != nil {
		return result, err
	}
	if req.SourceUnits != nil {
		units, err := CanonicalSourceUnits(req)
		if err != nil {
			return result, err
		}
		req.SourceUnits = units
	}
	registered, err := e.registered(req.ProblemID)
	if err != nil {
		return result, unavailable()
	}
	if !registered {
		return result, &compilation.Error{Status: 404, Code: "problem_not_found", Message: "题目不存在。"}
	}
	if e.tainted.Load() {
		return result, unavailable()
	}
	select {
	case e.busy <- struct{}{}:
		defer func() { <-e.busy }()
	default:
		return result, &compilation.Error{Status: 429, Code: "judge_busy", Message: "判题服务正在处理其他请求，请稍后重试。"}
	}
	if e.tainted.Load() {
		return result, unavailable()
	}
	started := time.Now()
	ctx, cancel := context.WithTimeout(parent, 70*time.Second)
	defer cancel()
	run, err := newID()
	if err != nil {
		return result, unavailable()
	}
	result = Result{RunID: run, Verdict: "SYSTEM_ERROR", Message: message("SYSTEM_ERROR"), SourceSHA256: RequestSourceDigest(req), RuntimeImage: e.image, Compiler: "GCC 14.2.0", Standard: "c++20", Compile: CompileResult{Status: "system_error"}, Cases: []CaseResult{}}
	defer func() {
		result.DurationMS = time.Since(started).Milliseconds()
		if ctx.Err() != nil && !e.tainted.Load() {
			result.Verdict = "SYSTEM_ERROR"
			result.Message = failureMessage(jobFailure(ctx))
		}
	}()
	snap, err := e.load(req.ProblemID)
	if err != nil {
		result.Message = failureMessage("testdata_load_failed")
		return result, nil
	}
	result.TestdataSHA256 = snap.SHA256
	if snap.Manifest.Checker == "cpp-special-v1" {
		result.CheckerSourceSHA256 = hash([]byte(snap.CheckerSource))
	}
	dir, err := os.MkdirTemp(e.jobsRoot, "run-"+run+"-")
	if err != nil {
		result.Message = failureMessage("snapshot_write_failed")
		return result, nil
	}
	defer func() {
		if e.tainted.Load() {
			return
		}
		if e.removeJob(dir) != nil {
			e.tainted.Store(true)
			result.Verdict = "SYSTEM_ERROR"
			result.Message = failureMessage("cleanup_failed")
		}
	}()
	if snap.Manifest.Type == "communication" {
		e.judgeCommunication(ctx, run, dir, req, snap, &result)
		return result, nil
	}
	compile, artifact, artifactHash, err := e.compile(ctx, run, dir, req.Source)
	result.Compile = compile
	if err != nil {
		result.Message = failureMessage(err.Error())
		return result, nil
	}
	if compile.Status == "compile_error" {
		result.Verdict = "CE"
		result.Message = message("CE")
		return result, nil
	}
	result.ArtifactSHA256 = artifactHash
	if snap.Manifest.Type == "interactive" {
		e.judgeInteractive(ctx, run, dir, artifact, snap, &result)
		return result, nil
	}
	var checkerArtifact string
	if snap.Manifest.Checker == "cpp-special-v1" {
		checkerDir := filepath.Join(dir, "checker-build")
		if os.Mkdir(checkerDir, 0700) != nil {
			result.Message = failureMessage("snapshot_write_failed")
			return result, nil
		}
		cc, ca, digest, compileErr := e.compileNamedBudget(ctx, run, checkerDir, snap.CheckerSource, "checker-compile", 10*time.Second)
		result.CheckerCompile = &cc
		if compileErr != nil {
			result.Message = failureMessage(compileErr.Error())
			return result, nil
		}
		if cc.Status != "success" {
			result.Message = "题目特殊判题程序编译失败，请教师检查特判程序。"
			return result, nil
		}
		checkerArtifact, result.CheckerArtifactSHA256 = ca, digest
	}
	result.Verdict = "AC"
	for i, c := range snap.Cases {
		if ctx.Err() != nil {
			result.Verdict = "SYSTEM_ERROR"
			result.Message = failureMessage(jobFailure(ctx))
			return result, nil
		}
		input := filepath.Join(dir, "input-"+c.ID+".txt")
		if os.WriteFile(input, c.Input, 0644) != nil {
			result.Verdict = "SYSTEM_ERROR"
			result.Message = failureMessage("snapshot_write_failed")
			return result, nil
		}
		spec := e.runtimeSpec(run, "case-"+string(rune('a'+i)), artifact, input, snap.Manifest)
		out, stderr := &boundedStream{limit: MaxRawOutput}, &boundedStream{limit: MaxRawOutput}
		stage := e.stage(ctx, spec, out, stderr)
		verdict := executionVerdict(stage)
		var checkerRun *CheckerResult
		feedback, checkerFailure := "", ""
		if verdict == "AC" {
			switch snap.Manifest.Checker {
			case "nonconsecutive-permutation-v1":
				verdict, feedback, err = constructionVerdict(c.Input, out.data.Bytes())
				if err != nil {
					verdict, checkerFailure = "SYSTEM_ERROR", "testdata_load_failed"
				}
			case "cpp-special-v1":
				verdict, checkerRun, feedback, checkerFailure = e.runChecker(ctx, run, dir, i, checkerArtifact, c.Input, c.Expected, out.data.Bytes())
			default:
				if !check(snap.Manifest.Checker, out.data.Bytes(), c.Expected) {
					verdict = "WA"
				}
			}
		}
		cr := CaseResult{ID: c.ID, Verdict: verdict, DurationMS: stage.DurationMS, OOMKilled: stage.State.OOMKilled, Checker: checkerRun, CheckerFeedback: feedback}
		cr.Input, cr.InputTruncated = display(c.Input, MaxFile)
		cr.Expected, cr.ExpectedTruncated = display(c.Expected, MaxFile)
		if stage.State.StartedAt != "" && !strings.HasPrefix(stage.State.StartedAt, "0001-") {
			cr.ExitCode = &stage.State.ExitCode
		}
		var outTrunc, errTrunc bool
		cr.Stdout, outTrunc = display(out.data.Bytes(), MaxFile)
		cr.Stderr, errTrunc = display(stderr.data.Bytes(), 16384)
		cr.OutputTruncated = outTrunc || errTrunc || stage.Cause == "output_limit"
		result.Cases = append(result.Cases, cr)
		if cr.Verdict == "SYSTEM_ERROR" {
			result.Verdict = cr.Verdict
			result.Message = failureMessage(stage.Failure)
			if checkerFailure != "" {
				result.Message = failureMessage(checkerFailure)
			}
			return result, nil
		}
		if result.Verdict == "AC" && cr.Verdict != "AC" {
			result.Verdict = cr.Verdict
		}
	}
	result.Message = message(result.Verdict)
	return result, nil
}

// Preflight uses the actual compile/export/bind/exec path as well as kernel facts.
// It cannot grant readiness based only on advertised Docker settings.
func (e *Engine) Preflight(ctx context.Context) (resultErr error) {
	b, err := e.control(ctx, "version", "--format", "{{.Server.Os}} {{.Server.Arch}}")
	if err != nil {
		return fmt.Errorf("docker_version_query_failed: %w", err)
	}
	if strings.TrimSpace(string(b)) != "linux amd64" {
		return errors.New("docker_linux_amd64_required")
	}
	b, err = e.control(ctx, "info", "--format", "{{.CgroupVersion}}")
	if err != nil {
		return fmt.Errorf("docker_cgroup_query_failed: %w", err)
	}
	if strings.TrimSpace(string(b)) != "2" {
		return errors.New("cgroup_v2_required")
	}
	b, err = e.control(ctx, "image", "inspect", e.image, "--format", "{{json .Config.Volumes}}")
	if err != nil || (strings.TrimSpace(string(b)) != "null" && strings.TrimSpace(string(b)) != "{}") {
		return errors.New("image_volumes_forbidden")
	}
	b, err = e.control(ctx, "ps", "-aq", "--filter", "label="+ownerLabel)
	if err != nil || len(strings.Fields(string(b))) != 0 {
		return errors.New("judge_container_leftovers")
	}
	entries, err := os.ReadDir(e.jobsRoot)
	if err != nil || len(entries) != 0 {
		return errors.New("judge_job_leftovers")
	}
	registered := problems.List()
	if e.store != nil {
		registered, err = e.store.List()
		if err != nil {
			return errors.New("problem_registry_invalid")
		}
	}
	for _, problem := range registered {
		if _, err = e.load(problem.ID); err != nil {
			return fmt.Errorf("testdata_preflight_failed_%s", problem.ID)
		}
	}
	run, err := newID()
	if err != nil {
		return errors.New("probe_identity_failed")
	}
	dir, err := os.MkdirTemp(e.jobsRoot, "run-"+run+"-probe,")
	if err != nil {
		return errors.New("probe_directory_failed")
	}
	defer func() {
		if !e.tainted.Load() {
			if e.removeJob(dir) != nil {
				e.tainted.Store(true)
				resultErr = errors.New("preflight_cleanup_failed")
			}
		}
	}()
	cr, artifact, _, err := e.compile(ctx, run, dir, "#include <cstdio>\nstatic_assert(__GNUC__ == 14 && __GNUC_MINOR__ == 2 && __GNUC_PATCHLEVEL__ == 0);\nint main(){puts(\"probe-ok\");}\n")
	if err != nil || cr.Status != "success" {
		return errors.New("compile_snapshot_probe_failed")
	}
	input := filepath.Join(dir, "input.txt")
	if os.WriteFile(input, nil, 0644) != nil {
		return errors.New("probe_input_failed")
	}
	s := e.runtimeSpec(run, "probe", artifact, input, Manifest{WallTimeMS: 5000, MemoryLimitBytes: 64 * 1048576})
	s.Args = []string{"-c", `set -eu; cat /sys/fs/cgroup/memory.max /sys/fs/cgroup/memory.swap.max /sys/fs/cgroup/pids.max /sys/fs/cgroup/cpu.max; grep -E '^(CapEff|NoNewPrivs|Seccomp):' /proc/self/status; test "$(id -u)" = 10002; test ! -e /var/run/docker.sock; test ! -e /expected; test ! -e /testdata; test ! -s /source/solution.cpp; test "$(ls /sys/class/net)" = lo; if touch /etc/oj-probe 2>/dev/null; then exit 91; fi; exec /solution/program < /input.txt`}
	out, stderr := &boundedStream{limit: MaxRawOutput}, &boundedStream{limit: MaxRawOutput}
	r := e.stage(ctx, s, out, stderr)
	actual := out.data.String()
	if r.Failure != "" || r.Cause != "" || r.State.ExitCode != 0 || !strings.HasPrefix(actual, probeExpected()) || !strings.Contains(actual, "CapEff:\t0000000000000000") || !strings.Contains(actual, "NoNewPrivs:\t1") || !strings.Contains(actual, "Seccomp:\t2") || !strings.HasSuffix(actual, "probe-ok\n") {
		return errors.New("runtime_kernel_probe_failed")
	}
	if e.tainted.Load() {
		return errors.New("preflight_cleanup_failed")
	}
	if err = e.interactiveProbe(ctx, run, dir); err != nil {
		return err
	}
	return nil
}

func (e *Engine) load(id string) (Snapshot, error) {
	if e.store != nil {
		return LoadWithStore(e.store, id)
	}
	return Load(e.dataRoot, id)
}

func (e *Engine) registered(id string) (bool, error) {
	if e.store == nil {
		return problems.Exists(id), nil
	}
	_, found, err := e.store.Get(id)
	return found, err
}
func (e *Engine) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		if e.tainted.Load() {
			compilation.WriteError(w, unavailable())
			return
		}
		compilation.WriteJSON(w, 200, map[string]string{"status": "ok", "profile": "judge-mvp/0.1.0", "image": e.image, "modes_contract_version": "modes-interactive/0.1.0", "special_judge_contract_version": "special-judge/0.1.0", "communication_contract_version": CommunicationContractVersion})
	})
	mux.HandleFunc("POST /judge", func(w http.ResponseWriter, r *http.Request) {
		var req Request
		if err := DecodeRequest(w, r, &req); err != nil {
			compilation.WriteError(w, err)
			return
		}
		result, err := e.Judge(r.Context(), req)
		if err != nil {
			compilation.WriteError(w, err)
			return
		}
		compilation.WriteJSON(w, 200, result)
	})
	mux.HandleFunc("POST /internal/communication-judge", func(w http.ResponseWriter, r *http.Request) {
		var req Request
		if err := DecodeCommunicationRequest(w, r, &req); err != nil {
			compilation.WriteError(w, err)
			return
		}
		result, err := e.Judge(r.Context(), req)
		if err != nil {
			compilation.WriteError(w, err)
			return
		}
		compilation.WriteJSON(w, 200, result)
	})
	metadata := func(w http.ResponseWriter, id string) {
		registered, err := e.registered(id)
		if err != nil {
			compilation.WriteError(w, unavailable())
			return
		}
		if !registered {
			compilation.WriteError(w, &compilation.Error{Status: 404, Code: "problem_not_found", Message: "题目不存在。"})
			return
		}
		if e.tainted.Load() {
			compilation.WriteError(w, unavailable())
			return
		}
		snap, err := e.load(id)
		if err != nil {
			compilation.WriteError(w, &compilation.Error{Status: 503, Code: "testdata_unavailable", Message: "服务器测试文件配置无效。"})
			return
		}
		problemType := snap.Manifest.Type
		if problemType == "" {
			problemType = "batch"
		}
		meta := Metadata{ProblemID: id, ProblemType: problemType, Enabled: true, CaseCount: len(snap.Cases), WallTimeMS: snap.Manifest.WallTimeMS, MemoryLimitBytes: snap.Manifest.MemoryLimitBytes, Checker: snap.Manifest.Checker}
		if snap.Manifest.Communication != nil {
			meta.Communication = &CommunicationMetadata{Roles: []string{"sender", "receiver"}, ByteLimit: snap.Manifest.Communication.ByteLimit, AggregateMemoryLimitBytes: 2 * snap.Manifest.MemoryLimitBytes}
		}
		compilation.WriteJSON(w, 200, meta)
	}
	mux.HandleFunc("GET /problem", func(w http.ResponseWriter, _ *http.Request) { metadata(w, "a-plus-b") })
	mux.HandleFunc("GET /problem/{id}", func(w http.ResponseWriter, r *http.Request) { metadata(w, r.PathValue("id")) })
	return mux
}
