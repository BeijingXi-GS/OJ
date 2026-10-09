package filejudge

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func contractCommunicationRequest() Request {
	return Request{ProblemID: "communication-sum", SourceUnits: []SourceUnit{{Role: "receiver", Language: "cpp", Source: "int main(){return 0;}"}, {Role: "sender", Language: "cpp", Source: "int main(){}"}}}
}
func contractCommunicationResult(t *testing.T) (Result, Snapshot, Request) {
	t.Helper()
	snap, err := Load("../../testdata/problems", "communication-sum")
	if err != nil {
		t.Fatal(err)
	}
	request := contractCommunicationRequest()
	units, failure := CanonicalSourceUnits(request)
	if failure != nil {
		t.Fatal(failure)
	}
	r := Result{RunID: "contract-communication", Verdict: "AC", SourceSHA256: RequestSourceDigest(request), TestdataSHA256: snap.SHA256, RuntimeImage: "sha256:" + strings.Repeat("a", 64), Compiler: "GCC 14.2.0", Standard: "c++20", Compile: CompileResult{Status: "success"}, Cases: []CaseResult{}}
	for _, u := range units {
		r.RoleCompiles = append(r.RoleCompiles, RoleCompile{Role: u.Role, SourceSHA256: hash([]byte(u.Source)), ArtifactSHA256: hash([]byte(u.Role)), Compile: CompileResult{Status: "success"}})
	}
	r.ArtifactSHA256 = CommunicationArtifactDigest(r.RoleCompiles)
	for _, c := range snap.Cases {
		zero := 0
		out := string(c.Expected)
		comm := &CommunicationResult{ByteLimit: snap.Manifest.Communication.ByteLimit, BytesObserved: int64(len(out)), BytesTransferred: int64(len(out)), Sender: RoleRun{ExitCode: &zero, Stdout: out}, Receiver: RoleRun{ExitCode: &zero, Stdout: out}}
		r.Cases = append(r.Cases, CaseResult{ID: c.ID, Verdict: "AC", ExitCode: &zero, Input: string(c.Input), Expected: out, Stdout: out, Communication: comm})
	}
	return r, snap, request
}
func TestCommunicationRequestWireBoundaries(t *testing.T) {
	good, _ := json.Marshal(contractCommunicationRequest())
	for name, body := range map[string]string{
		"duplicate top":      strings.Replace(string(good), `"problem_id":`, `"problem_id":"wrong","problem_id":`, 1),
		"duplicate nested":   strings.Replace(string(good), `"role":"receiver"`, `"role":"sender","role":"receiver"`, 1),
		"case alias":         strings.Replace(string(good), `"source_units":`, `"Source_Units":`, 1),
		"Unicode fold alias": strings.Replace(string(good), `"source":"int main(){return 0;}"`, `"source":"int main(){return 0;}","ſource":"replaced"`, 1),
		"null units":         `{"problem_id":"communication-sum","source_units":null}`,
		"null role":          strings.Replace(string(good), `"role":"receiver"`, `"role":null`, 1),
		"unknown nested":     strings.Replace(string(good), `"role":"receiver"`, `"role":"receiver","argv":[]`, 1),
		"duplicate role":     strings.Replace(string(good), `"role":"receiver"`, `"role":"sender"`, 1),
		"top source":         strings.Replace(string(good), `"source":""`, `"source":"int main(){}"`, 1),
		"nul source":         strings.Replace(string(good), `int main(){return 0;}`, `\u0000`, 1),
		"missing receiver":   `{"problem_id":"communication-sum","source_units":[{"role":"sender","language":"cpp","source":"int main(){}"}]}`,
		"extra JSON":         string(good) + ` {}`,
	} {
		t.Run(name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/", strings.NewReader(body))
			r.Header.Set("Content-Type", "application/json")
			var request Request
			if DecodeCommunicationRequest(httptest.NewRecorder(), r, &request) == nil {
				t.Fatal("invalid request accepted")
			}
		})
	}
	max := contractCommunicationRequest()
	for i := range max.SourceUnits {
		max.SourceUnits[i].Source = strings.Repeat("\x01", MaxFile)
	}
	body, _ := json.Marshal(max)
	if len(body) <= 262144 || len(body) > MaxCommunicationBody {
		t.Fatal("escape fixture wrong size", len(body))
	}
	r := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	var parsed Request
	if failure := DecodeCommunicationRequest(httptest.NewRecorder(), r, &parsed); failure != nil {
		t.Fatal("valid maximum escaped source rejected", failure)
	}
	r = httptest.NewRequest("POST", "/", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	if failure := DecodeRequest(httptest.NewRecorder(), r, &parsed); failure == nil || failure.Status != 413 {
		t.Fatal("old boundary changed", failure)
	}
	r = httptest.NewRequest("POST", "/", strings.NewReader(strings.Repeat(" ", MaxCommunicationBody+1)))
	r.Header.Set("Content-Type", "application/json")
	if failure := DecodeCommunicationRequest(httptest.NewRecorder(), r, &parsed); failure == nil || failure.Status != 413 {
		t.Fatal("new bound missing", failure)
	}
	max.SourceUnits[0].Source = strings.Repeat("x", MaxFile+1)
	if _, failure := CanonicalSourceUnits(max); failure == nil || failure.Status != 413 {
		t.Fatal("per-role bound changed", failure)
	}
}
func TestCommunicationCanonicalBindingAndTerminationEvidence(t *testing.T) {
	r, snap, request := contractCommunicationResult(t)
	if !validResult(r) || !MatchesSnapshot(r, snap) || !MatchesRequest(r, request) {
		t.Fatal("good fixture rejected")
	}
	reversed := request
	reversed.SourceUnits = []SourceUnit{request.SourceUnits[1], request.SourceUnits[0]}
	if RequestSourceDigest(reversed) != RequestSourceDigest(request) {
		t.Fatal("array order changed source binding")
	}
	for name, mutate := range map[string]func(*Result){
		"wrong source":          func(r *Result) { r.RoleCompiles[0].SourceSHA256 = hash([]byte("different")) },
		"wrong artifact":        func(r *Result) { r.ArtifactSHA256 = hash([]byte("different")) },
		"role order":            func(r *Result) { r.RoleCompiles[0], r.RoleCompiles[1] = r.RoleCompiles[1], r.RoleCompiles[0] },
		"limit":                 func(r *Result) { r.Cases[0].Communication.ByteLimit++ },
		"over-transfer":         func(r *Result) { r.Cases[0].Communication.BytesTransferred = 65537 },
		"fake termination zero": func(r *Result) { r.Cases[0].Communication.Sender.TerminatedBySupervisor = true },
		"fake termination two": func(r *Result) {
			code := 2
			r.Cases[0].Communication.Sender.ExitCode = &code
			r.Cases[0].Communication.Sender.TerminatedBySupervisor = true
		},
		"fake termination nil": func(r *Result) {
			r.Cases[0].Communication.Sender.ExitCode = nil
			r.Cases[0].Communication.Sender.TerminatedBySupervisor = true
		},
		"re without failure": func(r *Result) { r.Verdict = "RE"; r.Cases[0].Verdict = "RE" },
		"mle without oom":    func(r *Result) { r.Verdict = "MLE"; r.Cases[0].Verdict = "MLE" },
	} {
		t.Run(name, func(t *testing.T) {
			clone, _, _ := contractCommunicationResult(t)
			mutate(&clone)
			if validResult(clone) && MatchesRequest(clone, request) && MatchesSnapshot(clone, snap) {
				t.Fatal("forged evidence accepted")
			}
		})
	}
	for _, supervised := range []bool{false, true} {
		clone, _, _ := contractCommunicationResult(t)
		code := 137
		clone.Cases[0].Communication.Sender.ExitCode = &code
		clone.Cases[0].Communication.Sender.TerminatedBySupervisor = supervised
		if supervised {
			clone.Verdict = "TLE"
			clone.Cases[0].Verdict = "TLE"
		} else {
			clone.Verdict = "RE"
			clone.Cases[0].Verdict = "RE"
		}
		if !validCommunicationResult(clone) {
			t.Fatal("legitimate 137 misclassified", supervised)
		}
	}
	early := r
	early.RoleCompiles = nil
	early.Cases = []CaseResult{}
	early.Compile = CompileResult{Status: "system_error"}
	early.ArtifactSHA256 = ""
	early.TestdataSHA256 = ""
	early.Verdict = "SYSTEM_ERROR"
	if !MatchesRequest(early, request) || !MatchesSnapshot(early, snap) {
		t.Fatal("early SYS rejected")
	}
	early.Verdict = "CE"
	early.Compile.Status = "compile_error"
	if validCommunicationResult(early) {
		t.Fatal("zero record CE accepted")
	}
}
func TestCommunicationClientStrictNestedWire(t *testing.T) {
	good, _, request := contractCommunicationResult(t)
	body, _ := json.Marshal(good)
	for name, wire := range map[string][]byte{
		"good":                body,
		"null roles":          bytes.Replace(body, []byte(`"role_compiles":[`), []byte(`"role_compiles":null,"discard":[`), 1),
		"missing role exit":   bytes.Replace(body, []byte(`"exit_code":0,"duration_ms":0`), []byte(`"duration_ms":0`), 1),
		"null boolean":        bytes.Replace(body, []byte(`"limit_exceeded":false`), []byte(`"limit_exceeded":null`), 1),
		"duplicate nested":    bytes.Replace(body, []byte(`"bytes_observed":`), []byte(`"bytes_observed":0,"bytes_observed":`), 1),
		"unknown nested":      bytes.Replace(body, []byte(`"sender":{`), []byte(`"sender":{"command":"ignored",`), 1),
		"case alias nested":   bytes.Replace(body, []byte(`"sender":{`), []byte(`"sender":{"Exit_Code":0,`), 1),
		"Unicode fold nested": bytes.Replace(body, []byte(`"sender":{`), []byte(`"ſender":null,"sender":{`), 1),
	} {
		t.Run(name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/internal/communication-judge" {
					t.Error("wrong worker endpoint", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(wire)
			}))
			defer s.Close()
			c, err := NewClient(s.URL)
			if err != nil {
				t.Fatal(err)
			}
			_, failure := c.Judge(context.Background(), request)
			if (failure == nil) != (name == "good") {
				t.Fatal("strict wire boundary", name, failure)
			}
		})
	}
}
