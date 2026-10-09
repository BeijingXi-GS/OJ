package filejudge

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"unicode/utf8"

	"oj.local/oj/internal/compilation"
)

const CommunicationContractVersion = "communication-mvp/0.1.0"
const MaxCommunicationBody = 1048576

type SourceUnit struct {
	Role     string `json:"role"`
	Language string `json:"language"`
	Source   string `json:"source"`
}
type RoleCompile struct {
	Role           string        `json:"role"`
	SourceSHA256   string        `json:"source_sha256"`
	ArtifactSHA256 string        `json:"artifact_sha256"`
	Compile        CompileResult `json:"compile"`
}
type RoleRun struct {
	ExitCode               *int   `json:"exit_code"`
	DurationMS             int64  `json:"duration_ms"`
	OOMKilled              bool   `json:"oom_killed"`
	Stdout                 string `json:"stdout"`
	Stderr                 string `json:"stderr"`
	OutputTruncated        bool   `json:"output_truncated"`
	TerminatedBySupervisor bool   `json:"terminated_by_supervisor"`
}
type CommunicationResult struct {
	BytesObserved    int64   `json:"bytes_observed"`
	BytesTransferred int64   `json:"bytes_transferred"`
	ByteLimit        int64   `json:"byte_limit"`
	LimitExceeded    bool    `json:"limit_exceeded"`
	Sender           RoleRun `json:"sender"`
	Receiver         RoleRun `json:"receiver"`
}
type CommunicationMetadata struct {
	Roles                     []string `json:"roles"`
	ByteLimit                 int64    `json:"byte_limit"`
	AggregateMemoryLimitBytes int64    `json:"aggregate_memory_limit_bytes"`
}

func invalidCommunication() *compilation.Error {
	return &compilation.Error{Status: 400, Code: "invalid_request", Message: "通信题需要发送端和接收端两份 C++ 源码。"}
}

// Validate the explicit role set before constructing the canonical binding.
func CanonicalSourceUnits(r Request) ([]SourceUnit, *compilation.Error) {
	if r.Language != "" || r.Source != "" || len(r.SourceUnits) != 2 {
		return nil, invalidCommunication()
	}
	units := make([]SourceUnit, 2)
	seen := [2]bool{}
	for _, u := range r.SourceUnits {
		i := -1
		if u.Role == "sender" {
			i = 0
		} else if u.Role == "receiver" {
			i = 1
		}
		if i < 0 || seen[i] || u.Language != "cpp" || strings.ContainsRune(u.Source, 0) {
			return nil, invalidCommunication()
		}
		if err := compilation.ValidateSource(u.Source); err != nil {
			return nil, err
		}
		seen[i], units[i] = true, u
	}
	return units, nil
}
func RequestSourceDigest(r Request) string {
	if r.SourceUnits == nil {
		return hash([]byte(r.Source))
	}
	u, err := CanonicalSourceUnits(r)
	if err != nil {
		return ""
	}
	b, _ := json.Marshal(u)
	return hash(b)
}

// Only the two successful, explicitly bound artifacts form a graph artifact.
func CommunicationArtifactDigest(records []RoleCompile) string {
	if len(records) != 2 {
		return ""
	}
	type binding struct {
		Role           string `json:"role"`
		ArtifactSHA256 string `json:"artifact_sha256"`
	}
	b := make([]binding, 2)
	for i, role := range []string{"sender", "receiver"} {
		c := records[i]
		if c.Role != role || c.Compile.Status != "success" || !digestRE.MatchString(c.ArtifactSHA256) {
			return ""
		}
		b[i] = binding{role, c.ArtifactSHA256}
	}
	data, _ := json.Marshal(b)
	return hash(data)
}

// Each endpoint uses its own size and shape boundary. Do not let the shared
// Request struct expand the old endpoint into an implicit communication route.
func DecodeRequest(w http.ResponseWriter, r *http.Request, dst *Request) *compilation.Error {
	return decodeJudgeRequest(w, r, dst, false)
}
func DecodeCommunicationRequest(w http.ResponseWriter, r *http.Request, dst *Request) *compilation.Error {
	return decodeJudgeRequest(w, r, dst, true)
}
func decodeJudgeRequest(w http.ResponseWriter, r *http.Request, dst *Request, communication bool) *compilation.Error {
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		return &compilation.Error{Status: 400, Code: "invalid_content_type", Message: "请使用 JSON 格式提交。"}
	}
	limit := int64(compilation.MaxBodyBytes)
	if communication {
		limit = MaxCommunicationBody
	}
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	data, err := io.ReadAll(r.Body)
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		return &compilation.Error{Status: 413, Code: "body_too_large", Message: "请求内容超过大小限制。"}
	}
	if err != nil || !uniqueJSON(data) {
		return invalidCommunication()
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	var value Request
	if d.Decode(&value) != nil {
		return invalidCommunication()
	}
	var keys map[string]json.RawMessage
	if json.Unmarshal(data, &keys) != nil || keys == nil {
		return invalidCommunication()
	}
	for _, raw := range keys {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return invalidCommunication()
		}
	}
	if communication {
		if _, present := keys["source_units"]; !present {
			return invalidCommunication()
		}
		var rawUnits []json.RawMessage
		if json.Unmarshal(keys["source_units"], &rawUnits) != nil {
			return invalidCommunication()
		}
		for _, raw := range rawUnits {
			if _, ok := requiredObject(raw, []string{"role", "language", "source"}, nil); !ok {
				return invalidCommunication()
			}
		}
		if _, e := CanonicalSourceUnits(value); e != nil {
			return e
		}
	} else if _, present := keys["source_units"]; present {
		return invalidCommunication()
	}
	*dst = value
	return nil
}

func validRoleRun(r RoleRun) bool {
	return r.DurationMS >= 0 && len(r.Stdout) <= MaxFile && len(r.Stderr) <= 16384 && utf8.ValidString(r.Stdout+r.Stderr) && (r.ExitCode == nil || (*r.ExitCode >= 0 && *r.ExitCode <= 255)) && (!r.TerminatedBySupervisor || (r.ExitCode != nil && *r.ExitCode == 137))
}

func validateTypedRequest(r Request, problemType string) *compilation.Error {
	if problemType == "communication" {
		_, err := CanonicalSourceUnits(r)
		return err
	}
	if r.SourceUnits != nil || r.Language != "cpp" {
		return invalidCommunication()
	}
	return compilation.ValidateSource(r.Source)
}

func validCommunicationMetadata(m *CommunicationMetadata, memory int64) bool {
	return m != nil && len(m.Roles) == 2 && m.Roles[0] == "sender" && m.Roles[1] == "receiver" && m.ByteLimit >= 1 && m.ByteLimit <= 65536 && m.AggregateMemoryLimitBytes == 2*memory
}
func completeCommunicationMetadata(data []byte) bool {
	var top map[string]json.RawMessage
	if json.Unmarshal(data, &top) != nil {
		return false
	}
	if raw, exists := top["communication"]; exists {
		if _, ok := requiredObject(raw, []string{"roles", "byte_limit", "aggregate_memory_limit_bytes"}, nil); !ok {
			return false
		}
	}
	return true
}
func completeCommunicationWire(data []byte) bool {
	var top map[string]json.RawMessage
	if json.Unmarshal(data, &top) != nil || top == nil {
		return false
	}
	if raw, exists := top["role_compiles"]; exists {
		var records []json.RawMessage
		if json.Unmarshal(raw, &records) != nil || records == nil {
			return false
		}
		for _, record := range records {
			f, ok := requiredObject(record, []string{"role", "source_sha256", "artifact_sha256", "compile"}, nil)
			if !ok {
				return false
			}
			if _, ok = requiredObject(f["compile"], []string{"status", "diagnostics", "diagnostics_truncated", "duration_ms"}, nil); !ok {
				return false
			}
		}
	}
	var cases []map[string]json.RawMessage
	if json.Unmarshal(top["cases"], &cases) != nil {
		return false
	}
	for _, row := range cases {
		if raw, exists := row["communication"]; exists {
			f, ok := requiredObject(raw, []string{"bytes_observed", "bytes_transferred", "byte_limit", "limit_exceeded", "sender", "receiver"}, nil)
			if !ok {
				return false
			}
			for _, role := range []string{"sender", "receiver"} {
				if _, ok = requiredObject(f[role], []string{"exit_code", "duration_ms", "oom_killed", "stdout", "stderr", "output_truncated", "terminated_by_supervisor"}, map[string]bool{"exit_code": true}); !ok {
					return false
				}
			}
		}
	}
	return true
}
func sameExit(a, b *int) bool {
	return (a == nil && b == nil) || (a != nil && b != nil && *a == *b)
}
func normalRole(r RoleRun) bool {
	return r.ExitCode != nil && *r.ExitCode == 0 && !r.OOMKilled && !r.TerminatedBySupervisor
}
func naturalFailure(r RoleRun) bool {
	return r.ExitCode != nil && *r.ExitCode != 0 && !r.OOMKilled && !r.TerminatedBySupervisor
}

func validCommunicationResult(r Result) bool {
	if !digestRE.MatchString(r.SourceSHA256) || !idRE.MatchString(r.RunID) || !verdictOK(r.Verdict) || r.DurationMS < 0 || r.Compiler != "GCC 14.2.0" || r.Standard != "c++20" || !strings.HasPrefix(r.RuntimeImage, "sha256:") || !digestRE.MatchString(strings.TrimPrefix(r.RuntimeImage, "sha256:")) || len(r.Message) > 1024 || !utf8.ValidString(r.Message) || r.Cases == nil || len(r.Cases) > 4 || r.Compile.DurationMS < 0 || len(r.Compile.Diagnostics) > MaxFile || !utf8.ValidString(r.Compile.Diagnostics) {
		return false
	}
	if r.InteractorCompile != nil || r.InteractorSourceSHA256 != "" || r.InteractorArtifactSHA256 != "" || r.CheckerCompile != nil || r.CheckerSourceSHA256 != "" || r.CheckerArtifactSHA256 != "" {
		return false
	}
	if r.TestdataSHA256 != "" && !digestRE.MatchString(r.TestdataSHA256) {
		return false
	}
	if len(r.RoleCompiles) > 2 {
		return false
	}
	if len(r.RoleCompiles) == 0 {
		return r.Verdict == "SYSTEM_ERROR" && r.Compile.Status == "system_error" && len(r.Cases) == 0 && r.ArtifactSHA256 == ""
	}
	for i, record := range r.RoleCompiles {
		if record.Role != []string{"sender", "receiver"}[i] || !digestRE.MatchString(record.SourceSHA256) || record.Compile.DurationMS < 0 || len(record.Compile.Diagnostics) > MaxFile || !utf8.ValidString(record.Compile.Diagnostics) {
			return false
		}
		if i > 0 && r.RoleCompiles[i-1].Compile.Status != "success" {
			return false
		}
		switch record.Compile.Status {
		case "success":
			if !digestRE.MatchString(record.ArtifactSHA256) {
				return false
			}
		case "compile_error", "system_error":
			if record.ArtifactSHA256 != "" {
				return false
			}
		default:
			return false
		}
	}
	aggregate := CommunicationArtifactDigest(r.RoleCompiles)
	if aggregate == "" {
		last := r.RoleCompiles[len(r.RoleCompiles)-1].Compile.Status
		if len(r.Cases) != 0 || r.ArtifactSHA256 != "" {
			return false
		}
		if last == "success" {
			return r.Verdict == "SYSTEM_ERROR" && r.Compile.Status == "system_error"
		}
		return r.Compile.Status == last && ((last == "compile_error" && (r.Verdict == "CE" || r.Verdict == "SYSTEM_ERROR")) || (last == "system_error" && r.Verdict == "SYSTEM_ERROR"))
	}
	if r.ArtifactSHA256 != aggregate || r.Compile.Status != "success" || r.Verdict == "CE" {
		return false
	}
	seen := map[string]bool{}
	overall := "AC"
	for _, c := range r.Cases {
		if overall == "SYSTEM_ERROR" || !idRE.MatchString(c.ID) || seen[c.ID] || !verdictOK(c.Verdict) || c.Verdict == "CE" || c.DurationMS < 0 || len(c.Stdout) > MaxFile || len(c.Stderr) > 16384 || len(c.Input) > MaxFile || len(c.Expected) > MaxFile || !utf8.ValidString(c.Stdout+c.Stderr+c.Input+c.Expected) || c.CPUTimeMS != nil || c.PeakMemoryBytes != nil || c.Interactor != nil || c.Checker != nil || c.CheckerFeedback != "" {
			return false
		}
		comm := c.Communication
		if comm == nil || !validRoleRun(comm.Sender) || !validRoleRun(comm.Receiver) || comm.ByteLimit < 1 || comm.ByteLimit > 65536 || comm.BytesObserved < 0 || comm.BytesTransferred < 0 || comm.BytesTransferred > comm.ByteLimit || comm.BytesObserved < comm.BytesTransferred || comm.LimitExceeded != (comm.BytesObserved > comm.ByteLimit) || !sameExit(c.ExitCode, comm.Receiver.ExitCode) || c.OOMKilled != (comm.Sender.OOMKilled || comm.Receiver.OOMKilled) || c.Stdout != comm.Receiver.Stdout {
			return false
		}
		diag, diagTruncated := display([]byte(comm.Sender.Stderr+comm.Receiver.Stderr), 16384)
		if c.Stderr != diag || c.OutputTruncated != (comm.Sender.OutputTruncated || comm.Receiver.OutputTruncated || diagTruncated) {
			return false
		}
		if c.StartTimedOut && c.Verdict != "TLE" && c.Verdict != "SYSTEM_ERROR" {
			return false
		}
		switch c.Verdict {
		case "AC":
			if !normalRole(comm.Sender) || !normalRole(comm.Receiver) || comm.LimitExceeded {
				return false
			}
		case "WA":
			if c.OOMKilled || (!comm.LimitExceeded && (!normalRole(comm.Sender) || !normalRole(comm.Receiver))) {
				return false
			}
		case "RE":
			if c.OOMKilled || comm.LimitExceeded || (!naturalFailure(comm.Sender) && !naturalFailure(comm.Receiver)) {
				return false
			}
		case "MLE":
			if !c.OOMKilled {
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
	return r.Verdict == "SYSTEM_ERROR" || (len(r.Cases) > 0 && r.TestdataSHA256 != "" && overall == r.Verdict)
}

// Bind the actual roles, not just the aggregate hash returned by a worker.
func MatchesRequest(r Result, request Request) bool {
	if r.SourceSHA256 != RequestSourceDigest(request) {
		return false
	}
	if request.SourceUnits == nil {
		if r.RoleCompiles != nil {
			return false
		}
		for _, c := range r.Cases {
			if c.Communication != nil {
				return false
			}
		}
		return true
	}
	units, err := CanonicalSourceUnits(request)
	if err != nil || !validCommunicationResult(r) {
		return false
	}
	for i, record := range r.RoleCompiles {
		if record.SourceSHA256 != hash([]byte(units[i].Source)) {
			return false
		}
	}
	return true
}

func matchesCommunicationSnapshot(r Result, snap Snapshot) bool {
	if snap.Manifest.Communication == nil || snap.Manifest.Checker != "communication-v1" || !validCommunicationResult(r) {
		return false
	}
	for i, c := range r.Cases {
		if i >= len(snap.Cases) || c.ID != snap.Cases[i].ID || c.Communication.ByteLimit != snap.Manifest.Communication.ByteLimit {
			return false
		}
		input, inputTruncated := display(snap.Cases[i].Input, MaxFile)
		expected, expectedTruncated := display(snap.Cases[i].Expected, MaxFile)
		if c.Input != input || c.InputTruncated != inputTruncated || c.Expected != expected || c.ExpectedTruncated != expectedTruncated {
			return false
		}
	}
	if r.TestdataSHA256 != snap.SHA256 {
		return r.Verdict == "SYSTEM_ERROR" && r.TestdataSHA256 == "" && len(r.Cases) == 0 && len(r.RoleCompiles) == 0
	}
	return r.Verdict == "SYSTEM_ERROR" || r.Compile.Status != "success" || len(r.Cases) == len(snap.Cases)
}

func CommunicationMetadataMatchesSnapshot(m Metadata, snap Snapshot, problemID string) bool {
	return validMetadata(m, problemID) && m.ProblemType == "communication" && snap.Manifest.Communication != nil && m.CaseCount == len(snap.Cases) && m.WallTimeMS == snap.Manifest.WallTimeMS && m.MemoryLimitBytes == snap.Manifest.MemoryLimitBytes && m.Communication.ByteLimit == snap.Manifest.Communication.ByteLimit
}
