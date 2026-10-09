package filejudge

import (
	"encoding/json"
	"strings"
	"testing"
)

func customResult() Result {
	r := testResult()
	r.CheckerCompile = &CompileResult{Status: "success"}
	r.CheckerSourceSHA256 = hash([]byte("checker source"))
	r.CheckerArtifactSHA256 = strings.Repeat("e", 64)
	zero := 0
	r.Cases[0].Checker = &CheckerResult{ExitCode: &zero}
	return r
}
func TestCheckerWireStateValidation(t *testing.T) {
	for _, name := range []string{"valid", "unknown", "null_checker", "null_compile", "null_feedback", "missing_exit", "nonzero_ac", "oom_ac", "student_re_with_checker", "absent_run", "bad_compile", "oversize_feedback", "empty_source", "invalid_utf8", "duplicate_hash", "duplicate_exit"} {
		t.Run(name, func(t *testing.T) {
			r := customResult()
			switch name {
			case "nonzero_ac":
				one := 1
				r.Cases[0].Checker.ExitCode = &one
			case "oom_ac":
				r.Cases[0].Checker.OOMKilled = true
			case "student_re_with_checker":
				r.Verdict = "RE"
				r.Cases[0].Verdict = "RE"
			case "absent_run":
				r.Cases[0].Checker = nil
			case "bad_compile":
				r.CheckerCompile.Status = "compile_error"
			case "oversize_feedback":
				r.Cases[0].CheckerFeedback = strings.Repeat("a", 1025)
			}
			data, _ := json.Marshal(r)
			checkerJSON, _ := json.Marshal(r.Cases[0].Checker)
			compileJSON, _ := json.Marshal(r.CheckerCompile)
			s := string(data)
			switch name {
			case "unknown":
				s = strings.Replace(s, `"checker":{"exit_code":0`, `"checker":{"unexpected":0,"exit_code":0`, 1)
			case "null_checker":
				s = strings.Replace(s, `"checker":`+string(checkerJSON), `"checker":null`, 1)
			case "null_compile":
				s = strings.Replace(s, `"checker_compile":`+string(compileJSON), `"checker_compile":null`, 1)
			case "null_feedback":
				s = strings.Replace(s, `"checker":{`, `"checker_feedback":null,"checker":{`, 1)
			case "missing_exit":
				s = strings.Replace(s, `"checker":{"exit_code":0,`, `"checker":{`, 1)
			case "empty_source":
				s = strings.Replace(s, r.CheckerSourceSHA256, "", 1)
			case "invalid_utf8":
				s = strings.Replace(s, `"message":"`, "\"message\":\"\xff", 1)
			case "duplicate_hash":
				s = strings.Replace(s, `"checker_source_sha256":`, `"checker_source_sha256":"`+r.CheckerSourceSHA256+`","checker_source_sha256":`, 1)
			case "duplicate_exit":
				s = strings.Replace(s, `"checker":{"exit_code":0`, `"checker":{"exit_code":1,"exit_code":0`, 1)
			}
			var decoded Result
			d := json.NewDecoder(strings.NewReader(s))
			d.DisallowUnknownFields()
			got := d.Decode(&decoded) == nil && uniqueJSON([]byte(s)) && completeWire([]byte(s)) && validResult(decoded)
			if got != (name == "valid") {
				t.Fatalf("state %s accepted=%v", name, got)
			}
		})
	}
}
func TestCheckerSnapshotBinding(t *testing.T) {
	r := customResult()
	snap := Snapshot{SHA256: r.TestdataSHA256, Cases: []SnapshotCase{{ID: "sample"}}, CheckerSource: "checker source", Manifest: Manifest{Type: "batch", Checker: "cpp-special-v1"}}
	if !MatchesSnapshot(r, snap) {
		t.Fatal("matching custom rejected")
	}
	for _, checker := range []string{"tokens-v1", "exact-v1", "nonconsecutive-permutation-v1", "interaction-v1"} {
		s := snap
		s.Manifest.Checker = checker
		if checker == "interaction-v1" {
			s.Manifest.Type = "interactive"
		}
		if MatchesSnapshot(r, s) {
			t.Fatalf("checker %s accepted injected program", checker)
		}
	}
	snap.CheckerSource = "modified source"
	if MatchesSnapshot(r, snap) {
		t.Fatal("checker source change accepted")
	}
	snap.CheckerSource = "checker source"
	snap.SHA256 = strings.Repeat("f", 64)
	if MatchesSnapshot(r, snap) {
		t.Fatal("testdata change accepted")
	}
}
func TestSnapshotRequiresAllCasesOrSystemErrorPrefix(t *testing.T) {
	r := customResult()
	s := Snapshot{SHA256: r.TestdataSHA256, CheckerSource: "checker source", Manifest: Manifest{Type: "batch", Checker: "cpp-special-v1"}, Cases: []SnapshotCase{{ID: "sample"}, {ID: "second"}}}
	if MatchesSnapshot(r, s) {
		t.Fatal("partial AC accepted")
	}
	r.Verdict = "SYSTEM_ERROR"
	if !MatchesSnapshot(r, s) {
		t.Fatal("valid cancelled prefix rejected")
	}
	r.Cases[0].ID = "other"
	if MatchesSnapshot(r, s) {
		t.Fatal("wrong case accepted")
	}
	r = customResult()
	for i := 1; i < 5; i++ {
		c := r.Cases[0]
		c.ID = string(rune('a' + i))
		r.Cases = append(r.Cases, c)
	}
	if validResult(r) {
		t.Fatal("fifth custom case accepted")
	}
}
func TestMetadataNewCheckerLimitsAndDuplicateJSON(t *testing.T) {
	m := Metadata{ProblemID: "fixture", Enabled: true, ProblemType: "batch", Checker: "cpp-special-v1", CaseCount: 4, WallTimeMS: 2000, MemoryLimitBytes: 64 * 1048576}
	if !validMetadata(m, "fixture") {
		t.Fatal("custom rejected")
	}
	m.CaseCount = 5
	if validMetadata(m, "fixture") {
		t.Fatal("fifth custom case accepted")
	}
	if uniqueJSON([]byte(`{"checker":"tokens-v1","checker":"cpp-special-v1"}`)) {
		t.Fatal("metadata duplicate accepted")
	}
}
