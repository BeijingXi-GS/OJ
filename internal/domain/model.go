// Package domain contains draft contracts, independent of transport, storage and execution adapters.
// Contract status: proposed; see docs/architecture/DOMAIN_MODEL.md.
package domain

type ID string

type VersionSnapshot struct {
	Problem  ID `json:"problem_version_id"`
	Testdata ID `json:"testdata_version_id"`
	Checker  ID `json:"checker_version_id"`
}

type SourceUnit struct {
	ID                       ID     `json:"id"`
	Role                     string `json:"role"`
	Language                 string `json:"language"`
	SourceKey                string `json:"source_key"`
	SourceSHA256             string `json:"source_sha256"`
	RequestedCompilerVersion string `json:"requested_compiler_version"`
	RequestedRuntimeVersion  string `json:"requested_runtime_version"`
}

type Submission struct {
	ID               ID           `json:"id"`
	UserID           ID           `json:"user_id"`
	ProblemVersionID ID           `json:"problem_version_id"`
	Sources          []SourceUnit `json:"sources"`
}

// Artifact and compilation status are run-owned, never mutable fields of a source snapshot.
type CompiledArtifact struct {
	ID              ID     `json:"id"`
	RunID           ID     `json:"run_id"`
	SourceUnitID    ID     `json:"source_unit_id"`
	Key             string `json:"key"`
	SHA256          string `json:"sha256"`
	CompilerVersion string `json:"compiler_version"`
	RuntimeVersion  string `json:"runtime_version"`
}

type ResourceLimit struct {
	CPUTimeMS   int64 `json:"cpu_time_ms"`
	WallTimeMS  int64 `json:"wall_time_ms"`
	MemoryBytes int64 `json:"memory_bytes"`
	PIDs        int64 `json:"pids"`
	OutputBytes int64 `json:"output_bytes"`
	TempBytes   int64 `json:"temp_bytes"`
}

type ResourcePolicy struct {
	PerProcess        map[string]ResourceLimit `json:"per_process"`
	Aggregate         ResourceLimit            `json:"aggregate"`
	Compile           ResourceLimit            `json:"compile"`
	NodeOverheadBytes int64                    `json:"node_overhead_bytes"`
}

type JudgeJob struct {
	ID           ID              `json:"id"`
	RunID        ID              `json:"run_id"`
	SubmissionID ID              `json:"submission_id"`
	JudgeType    string          `json:"judge_type"`
	Versions     VersionSnapshot `json:"versions"`
	Sources      []SourceUnit    `json:"sources"`
	Resources    ResourcePolicy  `json:"resources"`
	Priority     int             `json:"priority"`
	Attempt      int             `json:"attempt"`
	Fence        int64           `json:"fence"`
}

type JudgeRun struct {
	ID           ID              `json:"id"`
	SubmissionID ID              `json:"submission_id"`
	Versions     VersionSnapshot `json:"versions"`
	State        RunState        `json:"state"`
	Verdict      Verdict         `json:"verdict,omitempty"`
}

type Failure struct {
	Category string `json:"category"`
	Code     string `json:"code"`
	Scope    string `json:"scope"`
	Role     string `json:"role,omitempty"`
	Observed int64  `json:"observed,omitempty"`
	Limit    int64  `json:"limit,omitempty"`
}

type JudgeResult struct {
	RunID                    ID               `json:"run_id"`
	Verdict                  Verdict          `json:"verdict"`
	Failure                  *Failure         `json:"failure,omitempty"`
	AggregatePeakMemoryBytes int64            `json:"aggregate_peak_memory_bytes"`
	PerRolePeakMemoryBytes   map[string]int64 `json:"per_role_peak_memory_bytes"`
}
