// Package filejudge implements the local, file-backed Docker OCI judge increment.
package filejudge

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"oj.local/oj/internal/compilation"
	"oj.local/oj/internal/problems"
	"oj.local/oj/internal/problemstore"
)

const (
	MaxCases     = 8
	MaxFile      = 65536
	MaxRawOutput = 1048576
	MaxArtifact  = 32 * 1048576
	// Eight cases can expand to 14.25 MiB after JSON control-byte escaping.
	// Field, file and raw-output limits remain unchanged.
	MaxResponse = 16 * 1048576
)

type Request struct {
	ProblemID   string       `json:"problem_id"`
	Language    string       `json:"language"`
	Source      string       `json:"source"`
	SourceUnits []SourceUnit `json:"source_units,omitempty"`
}
type CompileResult struct {
	Status               string `json:"status"`
	Diagnostics          string `json:"diagnostics"`
	DiagnosticsTruncated bool   `json:"diagnostics_truncated"`
	DurationMS           int64  `json:"duration_ms"`
}
type CaseResult struct {
	ID                string               `json:"id"`
	Verdict           string               `json:"verdict"`
	DurationMS        int64                `json:"duration_ms"`
	CPUTimeMS         *int64               `json:"cpu_time_ms"`
	PeakMemoryBytes   *int64               `json:"peak_memory_bytes"`
	ExitCode          *int                 `json:"exit_code"`
	OOMKilled         bool                 `json:"oom_killed"`
	Stdout            string               `json:"stdout"`
	Stderr            string               `json:"stderr"`
	OutputTruncated   bool                 `json:"output_truncated"`
	Input             string               `json:"input"`
	Expected          string               `json:"expected"`
	InputTruncated    bool                 `json:"input_truncated"`
	ExpectedTruncated bool                 `json:"expected_truncated"`
	StartTimedOut     bool                 `json:"start_timed_out,omitempty"`
	Interactor        *InteractorResult    `json:"interactor,omitempty"`
	Checker           *CheckerResult       `json:"checker,omitempty"`
	CheckerFeedback   string               `json:"checker_feedback,omitempty"`
	Communication     *CommunicationResult `json:"communication,omitempty"`
}
type CheckerResult struct {
	ExitCode        *int   `json:"exit_code"`
	DurationMS      int64  `json:"duration_ms"`
	OOMKilled       bool   `json:"oom_killed"`
	Stdout          string `json:"stdout"`
	Stderr          string `json:"stderr"`
	OutputTruncated bool   `json:"output_truncated"`
}
type InteractorResult struct {
	ExitCode           *int   `json:"exit_code"`
	OOMKilled          bool   `json:"oom_killed"`
	Stdout             string `json:"stdout"`
	Stderr             string `json:"stderr"`
	OutputTruncated    bool   `json:"output_truncated"`
	ParticipantStopped bool   `json:"participant_stopped"`
}
type Result struct {
	RunID                    string         `json:"run_id"`
	Verdict                  string         `json:"verdict"`
	Message                  string         `json:"message"`
	DurationMS               int64          `json:"duration_ms"`
	SourceSHA256             string         `json:"source_sha256"`
	TestdataSHA256           string         `json:"testdata_sha256"`
	ArtifactSHA256           string         `json:"artifact_sha256"`
	RuntimeImage             string         `json:"runtime_image"`
	Compiler                 string         `json:"compiler"`
	Standard                 string         `json:"standard"`
	Compile                  CompileResult  `json:"compile"`
	Cases                    []CaseResult   `json:"cases"`
	InteractorCompile        *CompileResult `json:"interactor_compile,omitempty"`
	InteractorSourceSHA256   string         `json:"interactor_source_sha256,omitempty"`
	InteractorArtifactSHA256 string         `json:"interactor_artifact_sha256,omitempty"`
	CheckerCompile           *CompileResult `json:"checker_compile,omitempty"`
	CheckerSourceSHA256      string         `json:"checker_source_sha256,omitempty"`
	CheckerArtifactSHA256    string         `json:"checker_artifact_sha256,omitempty"`
	RoleCompiles             []RoleCompile  `json:"role_compiles,omitempty"`
}
type Judge interface {
	Judge(context.Context, Request) (Result, *compilation.Error)
}

// InteractorSourceDigest binds the only supported protocol to its exact source.
func InteractorSourceDigest() string { return hash([]byte(guessNumberInteractor)) }

// MatchesProblemType is checked with the trusted catalog before public projection.
// Wire validation alone cannot infer the requested type from optional fields.
func MatchesProblemType(r Result, problemType string) bool {
	if problemType == "communication" {
		return validCommunicationResult(r)
	}
	if r.RoleCompiles != nil {
		return false
	}
	for _, c := range r.Cases {
		if c.Communication != nil {
			return false
		}
	}
	if problemType == "batch" {
		if r.InteractorCompile != nil || r.InteractorSourceSHA256 != "" || r.InteractorArtifactSHA256 != "" {
			return false
		}
		for _, c := range r.Cases {
			if c.Interactor != nil || c.StartTimedOut {
				return false
			}
		}
		return true
	}
	if problemType != "interactive" {
		return false
	}
	if r.InteractorCompile == nil {
		return len(r.Cases) == 0 && r.InteractorSourceSHA256 == "" && r.InteractorArtifactSHA256 == "" && (r.Verdict == "CE" || r.Verdict == "SYSTEM_ERROR")
	}
	if r.InteractorSourceSHA256 != InteractorSourceDigest() {
		return false
	}
	if r.Compile.Status != "success" {
		return false
	}
	for _, c := range r.Cases {
		if c.Interactor == nil {
			return false
		}
	}
	return (r.InteractorCompile.Status == "success" && digestRE.MatchString(r.InteractorArtifactSHA256)) || (r.Verdict == "SYSTEM_ERROR" && len(r.Cases) == 0)
}

type Metadata struct {
	ProblemID        string                 `json:"problem_id"`
	Enabled          bool                   `json:"enabled"`
	CaseCount        int                    `json:"case_count"`
	WallTimeMS       int64                  `json:"wall_time_ms"`
	MemoryLimitBytes int64                  `json:"memory_limit_bytes"`
	Checker          string                 `json:"checker"`
	ProblemType      string                 `json:"problem_type"`
	Communication    *CommunicationMetadata `json:"communication,omitempty"`
}
type MetadataProvider interface {
	Metadata(context.Context, string) (Metadata, *compilation.Error)
}

func (c *Client) Metadata(parent context.Context, problemID string) (Metadata, *compilation.Error) {
	var result Metadata
	if !idRE.MatchString(problemID) {
		return result, &compilation.Error{Status: 404, Code: "problem_not_found", Message: "题目不存在。"}
	}
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", strings.TrimSuffix(c.endpoint, "/judge")+"/problem/"+url.PathEscape(problemID), nil)
	if err != nil {
		return result, unavailable()
	}
	res, err := c.http.Do(req)
	if err != nil {
		return result, unavailable()
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return result, unavailable()
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, 4097))
	if err != nil || len(data) > 4096 {
		return result, protocolError()
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	var extra any
	_, complete := requiredObject(data, []string{"problem_id", "enabled", "case_count", "wall_time_ms", "memory_limit_bytes", "checker", "problem_type"}, nil)
	if !complete || !uniqueJSON(data) || !completeCommunicationMetadata(data) || d.Decode(&result) != nil || d.Decode(&extra) != io.EOF || !validMetadata(result, problemID) {
		return result, protocolError()
	}
	return result, nil
}

func validMetadata(m Metadata, id string) bool {
	batch := m.ProblemType == "batch" && (m.Checker == "tokens-v1" || m.Checker == "exact-v1" || m.Checker == "nonconsecutive-permutation-v1" || m.Checker == "cpp-special-v1")
	interactive := m.ProblemType == "interactive" && m.Checker == "interaction-v1"
	communication := m.ProblemType == "communication" && m.Checker == "communication-v1" && m.CaseCount <= 4 && validCommunicationMetadata(m.Communication, m.MemoryLimitBytes)
	return m.ProblemID == id && m.Enabled && m.CaseCount >= 1 && m.CaseCount <= MaxCases && (m.Checker != "cpp-special-v1" || m.CaseCount <= 4) && m.WallTimeMS >= 500 && m.WallTimeMS <= 5000 && m.MemoryLimitBytes >= 32*1048576 && m.MemoryLimitBytes <= 256*1048576 && (communication || ((batch || interactive) && m.Communication == nil))
}

var digestRE = regexp.MustCompile(`^[a-f0-9]{64}$`)
var idRE = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)

func ValidateRequest(r Request) *compilation.Error {
	p, exists := problems.Get(r.ProblemID)
	if !exists {
		return &compilation.Error{Status: 404, Code: "problem_not_found", Message: "题目不存在。"}
	}
	return validateTypedRequest(r, p.Type)
}
func ValidateRequestWithStore(r Request, store *problemstore.Store) *compilation.Error {
	if r.ProblemID == "" {
		return invalidCommunication()
	}
	if !idRE.MatchString(r.ProblemID) {
		return &compilation.Error{Status: 404, Code: "problem_not_found", Message: "题目不存在。"}
	}
	p, exists, err := store.Get(r.ProblemID)
	if err != nil {
		return &compilation.Error{Status: 503, Code: "problem_store_unavailable", Message: "题库暂时不可用。"}
	}
	if !exists {
		return &compilation.Error{Status: 404, Code: "problem_not_found", Message: "题目不存在。"}
	}
	return validateTypedRequest(r, p.Type)
}
func verdictOK(v string) bool {
	switch v {
	case "AC", "WA", "TLE", "MLE", "CE", "RE", "OLE", "SYSTEM_ERROR":
		return true
	}
	return false
}
func display(data []byte, limit int) (string, bool) {
	s := strings.ToValidUTF8(string(data), "�")
	truncated := len(s) > limit
	if truncated {
		s = s[:limit]
		for !utf8.ValidString(s) {
			s = s[:len(s)-1]
		}
	}
	return s, truncated
}
func check(checker string, actual, expected []byte) bool {
	if checker == "exact-v1" {
		return bytes.Equal(actual, expected)
	}
	if checker != "tokens-v1" {
		return false
	}
	a, b := bytes.Fields(actual), bytes.Fields(expected)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !bytes.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}
func message(v string) string {
	switch v {
	case "AC":
		return "所有测试用例通过。"
	case "WA":
		return "答案不满足题目要求。"
	case "TLE":
		return "运行超过墙钟时间限制（含启动开销）。"
	case "MLE":
		return "容器内存超限，已取得 OOM 证据。"
	case "CE":
		return "C++ 编译或链接失败。"
	case "RE":
		return "程序异常退出。"
	case "OLE":
		return "程序输出超过 1 MiB 限制。"
	}
	return "判题基础设施处理失败。"
}

type Client struct {
	endpoint string
	http     *http.Client
}

func NewClient(base string) (*Client, error) {
	u, err := url.Parse(base)
	if err != nil || u == nil || u.Scheme != "http" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return nil, unavailable()
	}
	return &Client{strings.TrimRight(base, "/") + "/judge", &http.Client{Timeout: 110 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}
func unavailable() *compilation.Error {
	return &compilation.Error{Status: 503, Code: "judge_unavailable", Message: "判题服务暂时不可用，请稍后重试。"}
}
func protocolError() *compilation.Error {
	return &compilation.Error{Status: 502, Code: "judge_protocol_error", Message: "判题服务返回无效结果。"}
}
func (c *Client) Judge(ctx context.Context, request Request) (Result, *compilation.Error) {
	endpoint := c.endpoint
	if request.SourceUnits != nil {
		if _, failure := CanonicalSourceUnits(request); failure != nil {
			return Result{}, failure
		}
		endpoint = strings.TrimSuffix(c.endpoint, "/judge") + "/internal/communication-judge"
	}
	data, _ := json.Marshal(request)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(data))
	if err != nil {
		return Result{}, unavailable()
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return Result{}, unavailable()
	}
	defer res.Body.Close()
	if res.StatusCode == 429 {
		return Result{}, &compilation.Error{Status: 429, Code: "judge_busy", Message: "判题服务正在处理其他请求，请稍后重试。"}
	}
	if res.StatusCode == 503 {
		return Result{}, unavailable()
	}
	if res.StatusCode != 200 {
		return Result{}, protocolError()
	}
	data, err = io.ReadAll(io.LimitReader(res.Body, MaxResponse+1))
	if err != nil || len(data) > MaxResponse {
		return Result{}, protocolError()
	}
	var r Result
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if d.Decode(&r) != nil {
		return Result{}, protocolError()
	}
	var extra any
	if d.Decode(&extra) != io.EOF || !uniqueJSON(data) || !completeWire(data) || !validResult(r) || !MatchesRequest(r, request) {
		return Result{}, protocolError()
	}
	return r, nil
}

func requiredObject(raw []byte, keys []string, nullable map[string]bool) (map[string]json.RawMessage, bool) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return nil, false
	}
	for _, k := range keys {
		v, ok := fields[k]
		if !ok || (!nullable[k] && bytes.Equal(bytes.TrimSpace(v), []byte("null"))) {
			return nil, false
		}
	}
	return fields, true
}
func completeWire(data []byte) bool {
	if !completeCommunicationWire(data) {
		return false
	}
	r, ok := requiredObject(data, []string{"run_id", "verdict", "message", "duration_ms", "source_sha256", "testdata_sha256", "artifact_sha256", "runtime_image", "compiler", "standard", "compile", "cases"}, nil)
	if !ok {
		return false
	}
	if _, ok = requiredObject(r["compile"], []string{"status", "diagnostics", "diagnostics_truncated", "duration_ms"}, nil); !ok {
		return false
	}
	for _, k := range []string{"interactor_source_sha256", "interactor_artifact_sha256", "checker_source_sha256", "checker_artifact_sha256"} {
		if raw, exists := r[k]; exists && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return false
		}
	}
	for _, k := range []string{"checker_source_sha256", "checker_artifact_sha256"} {
		if raw, exists := r[k]; exists {
			var value string
			if json.Unmarshal(raw, &value) != nil || !digestRE.MatchString(value) {
				return false
			}
		}
	}
	if raw, exists := r["interactor_compile"]; exists {
		if _, ok = requiredObject(raw, []string{"status", "diagnostics", "diagnostics_truncated", "duration_ms"}, nil); !ok {
			return false
		}
	}
	if raw, exists := r["checker_compile"]; exists {
		if _, ok = requiredObject(raw, []string{"status", "diagnostics", "diagnostics_truncated", "duration_ms"}, nil); !ok {
			return false
		}
	}
	var cases []json.RawMessage
	if json.Unmarshal(r["cases"], &cases) != nil || cases == nil {
		return false
	}
	for _, c := range cases {
		fields, good := requiredObject(c, []string{"id", "verdict", "duration_ms", "cpu_time_ms", "peak_memory_bytes", "exit_code", "oom_killed", "stdout", "stderr", "output_truncated", "input", "expected", "input_truncated", "expected_truncated"}, map[string]bool{"cpu_time_ms": true, "peak_memory_bytes": true, "exit_code": true})
		if !good {
			return false
		}
		for _, k := range []string{"input", "expected", "input_truncated", "expected_truncated", "start_timed_out", "checker_feedback"} {
			if raw, exists := fields[k]; exists && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				return false
			}
		}
		if raw, exists := fields["interactor"]; exists {
			if _, ok = requiredObject(raw, []string{"exit_code", "oom_killed", "stdout", "stderr", "output_truncated", "participant_stopped"}, map[string]bool{"exit_code": true}); !ok {
				return false
			}
		}
		if raw, exists := fields["checker"]; exists {
			if _, ok = requiredObject(raw, []string{"exit_code", "duration_ms", "oom_killed", "stdout", "stderr", "output_truncated"}, map[string]bool{"exit_code": true}); !ok {
				return false
			}
		}
	}
	return true
}
func validResult(r Result) bool {
	if !digestRE.MatchString(r.SourceSHA256) || !idRE.MatchString(r.RunID) || !verdictOK(r.Verdict) || r.DurationMS < 0 || r.Compiler != "GCC 14.2.0" || r.Standard != "c++20" || !strings.HasPrefix(r.RuntimeImage, "sha256:") || !digestRE.MatchString(strings.TrimPrefix(r.RuntimeImage, "sha256:")) || len(r.Message) > 1024 || r.Cases == nil || len(r.Cases) > MaxCases || len(r.Compile.Diagnostics) > MaxFile || !utf8.ValidString(r.Compile.Diagnostics) || r.Compile.DurationMS < 0 {
		return false
	}
	if r.TestdataSHA256 != "" && !digestRE.MatchString(r.TestdataSHA256) {
		return false
	}
	if r.ArtifactSHA256 != "" && !digestRE.MatchString(r.ArtifactSHA256) {
		return false
	}
	if r.RoleCompiles != nil {
		return validCommunicationResult(r)
	}
	for _, c := range r.Cases {
		if c.Communication != nil {
			return false
		}
	}
	interactive := r.InteractorCompile != nil
	if !validCheckerResult(r) {
		return false
	}
	if interactive {
		if r.Compile.Status != "success" {
			return false
		}
		ic := r.InteractorCompile
		if r.InteractorSourceSHA256 != InteractorSourceDigest() || len(ic.Diagnostics) > MaxFile || !utf8.ValidString(ic.Diagnostics) || ic.DurationMS < 0 {
			return false
		}
		switch ic.Status {
		case "success":
			if !digestRE.MatchString(r.InteractorArtifactSHA256) {
				return false
			}
		case "compile_error", "system_error":
			if r.Verdict != "SYSTEM_ERROR" || len(r.Cases) != 0 || r.InteractorArtifactSHA256 != "" {
				return false
			}
		default:
			return false
		}
	} else if r.InteractorSourceSHA256 != "" || r.InteractorArtifactSHA256 != "" {
		return false
	}
	switch r.Compile.Status {
	case "success":
		if !digestRE.MatchString(r.ArtifactSHA256) || r.Verdict == "CE" {
			return false
		}
	case "compile_error":
		if (r.Verdict != "CE" && r.Verdict != "SYSTEM_ERROR") || len(r.Cases) != 0 || r.ArtifactSHA256 != "" {
			return false
		}
	case "system_error":
		if r.Verdict != "SYSTEM_ERROR" || len(r.Cases) != 0 {
			return false
		}
	default:
		return false
	}
	seen := map[string]bool{}
	overall := "AC"
	for _, c := range r.Cases {
		if overall == "SYSTEM_ERROR" {
			return false
		}
		if !idRE.MatchString(c.ID) || seen[c.ID] || !verdictOK(c.Verdict) || c.Verdict == "CE" || c.DurationMS < 0 || len(c.Stdout) > MaxFile || len(c.Stderr) > 16384 || len(c.Input) > MaxFile || len(c.Expected) > MaxFile || !utf8.ValidString(c.Stdout+c.Stderr+c.Input+c.Expected) || c.CPUTimeMS != nil || c.PeakMemoryBytes != nil || (c.ExitCode != nil && (*c.ExitCode < 0 || *c.ExitCode > 255)) || (c.Verdict == "MLE" && !c.OOMKilled) {
			return false
		}
		if c.Verdict != "SYSTEM_ERROR" && c.ExitCode == nil && !c.StartTimedOut {
			return false
		}
		if interactive != (c.Interactor != nil) {
			return false
		}
		if c.StartTimedOut && (!interactive || r.Compile.Status != "success" || r.InteractorCompile.Status != "success" || c.Verdict != "TLE" || c.ExitCode != nil || c.OOMKilled || c.Interactor.ExitCode != nil || c.Interactor.OOMKilled || c.Interactor.ParticipantStopped) {
			return false
		}
		stopped := false
		if i := c.Interactor; i != nil {
			stopped = i.ParticipantStopped
			if len(i.Stdout) > MaxFile || len(i.Stderr) > 16384 || !utf8.ValidString(i.Stdout+i.Stderr) || (i.ExitCode != nil && (*i.ExitCode < 0 || *i.ExitCode > 255)) || (stopped && (c.Verdict != "WA" || i.ExitCode == nil || *i.ExitCode != 1 || i.OOMKilled || c.OOMKilled)) {
				return false
			}
			if c.Verdict != "SYSTEM_ERROR" && ((i.ExitCode == nil && !c.StartTimedOut) || i.OOMKilled) {
				return false
			}
			if c.Verdict == "AC" && (*i.ExitCode != 0 || stopped) {
				return false
			}
			if c.Verdict == "WA" && *i.ExitCode != 1 {
				return false
			}
		}
		switch c.Verdict {
		case "AC":
			if *c.ExitCode != 0 || c.OOMKilled || stopped {
				return false
			}
		case "WA":
			if (*c.ExitCode != 0 && !stopped) || c.OOMKilled {
				return false
			}
		case "RE":
			if *c.ExitCode == 0 || c.OOMKilled {
				return false
			}
		case "TLE", "OLE":
			if c.OOMKilled {
				return false
			}
		}
		seen[c.ID] = true
		if c.Verdict == "SYSTEM_ERROR" {
			overall = "SYSTEM_ERROR"
		} else if overall == "AC" && c.Verdict != "AC" {
			overall = c.Verdict
		}
	}
	if r.Compile.Status == "success" && r.Verdict != "SYSTEM_ERROR" && (len(r.Cases) == 0 || r.TestdataSHA256 == "" || overall != r.Verdict) {
		return false
	}
	return true
}
