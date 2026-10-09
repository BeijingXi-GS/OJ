package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"oj.local/oj/internal/compilation"
	"oj.local/oj/internal/filejudge"
)

func communicationRequestFixture() filejudge.Request {
	return filejudge.Request{ProblemID: "communication-sum", SourceUnits: []filejudge.SourceUnit{{Role: "receiver", Language: "cpp", Source: "int main(){return 0;}"}, {Role: "sender", Language: "cpp", Source: "int main(){}"}}}
}
func fixtureDigest(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
func communicationFixtureResult(request filejudge.Request) filejudge.Result {
	snap, err := filejudge.Load("../../testdata/problems", request.ProblemID)
	if err != nil {
		panic(err)
	}
	units, failure := filejudge.CanonicalSourceUnits(request)
	if failure != nil {
		panic(failure)
	}
	r := filejudge.Result{RunID: "communication-fixture", Verdict: "AC", Message: "所有测试用例通过。", SourceSHA256: filejudge.RequestSourceDigest(request), TestdataSHA256: snap.SHA256, RuntimeImage: "sha256:" + strings.Repeat("a", 64), Compiler: "GCC 14.2.0", Standard: "c++20", Compile: filejudge.CompileResult{Status: "success"}, Cases: []filejudge.CaseResult{}}
	for _, u := range units {
		r.RoleCompiles = append(r.RoleCompiles, filejudge.RoleCompile{Role: u.Role, SourceSHA256: fixtureDigest(u.Source), ArtifactSHA256: fixtureDigest(u.Role), Compile: filejudge.CompileResult{Status: "success"}})
	}
	r.ArtifactSHA256 = filejudge.CommunicationArtifactDigest(r.RoleCompiles)
	for _, c := range snap.Cases {
		zero := 0
		out := string(c.Expected)
		payload := out
		if request.ProblemID == "communication-a-plus-b" {
			payload = string(c.Input)
		}
		comm := &filejudge.CommunicationResult{BytesObserved: int64(len(payload)), BytesTransferred: int64(len(payload)), ByteLimit: snap.Manifest.Communication.ByteLimit, Sender: filejudge.RoleRun{ExitCode: &zero, Stdout: payload}, Receiver: filejudge.RoleRun{ExitCode: &zero, Stdout: out}}
		r.Cases = append(r.Cases, filejudge.CaseResult{ID: c.ID, Verdict: "AC", ExitCode: &zero, Input: string(c.Input), Expected: out, Stdout: out, Communication: comm})
	}
	return r
}

type communicationJudge struct {
	calls         int
	mutate        func(*filejudge.Result)
	wrongMetadata bool
}

func (j *communicationJudge) Judge(_ context.Context, request filejudge.Request) (filejudge.Result, *compilation.Error) {
	j.calls++
	r := communicationFixtureResult(request)
	if j.mutate != nil {
		j.mutate(&r)
	}
	return r, nil
}
func (j *communicationJudge) Metadata(_ context.Context, id string) (filejudge.Metadata, *compilation.Error) {
	snap, err := filejudge.Load("../../testdata/problems", id)
	if err != nil {
		panic(err)
	}
	m := filejudge.Metadata{ProblemID: id, Enabled: true, CaseCount: len(snap.Cases), WallTimeMS: snap.Manifest.WallTimeMS, MemoryLimitBytes: snap.Manifest.MemoryLimitBytes, Checker: snap.Manifest.Checker, ProblemType: snap.Manifest.Type}
	if snap.Manifest.Communication != nil {
		m.Communication = &filejudge.CommunicationMetadata{Roles: []string{"sender", "receiver"}, ByteLimit: snap.Manifest.Communication.ByteLimit, AggregateMemoryLimitBytes: 2 * m.MemoryLimitBytes}
	}
	if j.wrongMetadata {
		m.Communication.ByteLimit++
	}
	return m, nil
}
func TestCommunicationAPIProjectionsAndRoleBinding(t *testing.T) {
	for _, mode := range []string{"student", "teacher", "engineer"} {
		t.Run(mode, func(t *testing.T) {
			j := &communicationJudge{}
			h := newFixtureServices(nil, nil, j)
			body, _ := json.Marshal(communicationRequestFixture())
			req := httptest.NewRequest("POST", "/api/v1/communication-judgements", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-OJ-Mode", mode)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			if w.Code != 200 || j.calls != 1 {
				t.Fatalf("%d %s", w.Code, w.Body.String())
			}
			var result map[string]json.RawMessage
			if json.Unmarshal(w.Body.Bytes(), &result) != nil {
				t.Fatal("invalid JSON")
			}
			var roles []map[string]json.RawMessage
			if json.Unmarshal(result["role_compiles"], &roles) != nil || len(roles) != 2 || string(roles[0]["role"]) != `"sender"` || string(roles[1]["role"]) != `"receiver"` {
				t.Fatal("role binding lost")
			}
			if mode == "student" {
				for _, key := range []string{"cases", "source_sha256", "runtime_image", "testdata_sha256", "artifact_sha256"} {
					if result[key] != nil {
						t.Fatal("student leakage", key)
					}
				}
			}
			for _, role := range roles {
				if mode != "engineer" && (len(role) != 2 || role["source_sha256"] != nil || role["artifact_sha256"] != nil) {
					t.Fatal("role hash leakage")
				}
			}
			if mode == "teacher" {
				var cases []map[string]json.RawMessage
				_ = json.Unmarshal(result["cases"], &cases)
				var comm map[string]json.RawMessage
				_ = json.Unmarshal(cases[0]["communication"], &comm)
				if len(comm) != 9 || comm["sender_stdout"] == nil || comm["bytes_transferred"] == nil || comm["sender"] != nil || comm["exit_code"] != nil {
					t.Fatal("teacher projection wrong", string(cases[0]["communication"]))
				}
			}
			if mode == "engineer" && !bytes.Contains(result["cases"], []byte("terminated_by_supervisor")) {
				t.Fatal("engineer evidence missing")
			}
		})
	}
}
func TestCommunicationAPIRejectsForgedWorkerBinding(t *testing.T) {
	for name, mutate := range map[string]func(*filejudge.Result){
		"role source":        func(r *filejudge.Result) { r.RoleCompiles[0].SourceSHA256 = fixtureDigest("wrong") },
		"aggregate source":   func(r *filejudge.Result) { r.SourceSHA256 = fixtureDigest("wrong") },
		"aggregate artifact": func(r *filejudge.Result) { r.ArtifactSHA256 = fixtureDigest("wrong") },
		"limit":              func(r *filejudge.Result) { r.Cases[0].Communication.ByteLimit++ },
		"input":              func(r *filejudge.Result) { r.Cases[0].Input = "forged" },
		"role order":         func(r *filejudge.Result) { r.RoleCompiles[0], r.RoleCompiles[1] = r.RoleCompiles[1], r.RoleCompiles[0] },
	} {
		t.Run(name, func(t *testing.T) {
			j := &communicationJudge{mutate: mutate}
			body, _ := json.Marshal(communicationRequestFixture())
			req := httptest.NewRequest("POST", "/api/v1/communication-judgements", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			newFixtureServices(nil, nil, j).ServeHTTP(w, req)
			if w.Code != 502 {
				t.Fatalf("forgery accepted %d %s", w.Code, w.Body.String())
			}
		})
	}
}
func TestCommunicationProblemMetadataAndStudentNotes(t *testing.T) {
	for _, mode := range []string{"student", "teacher", "engineer"} {
		j := &communicationJudge{}
		req := httptest.NewRequest("GET", "/api/v1/problems/communication-sum", nil)
		req.Header.Set("X-OJ-Mode", mode)
		w := httptest.NewRecorder()
		newFixtureServices(nil, nil, j).ServeHTTP(w, req)
		var p map[string]json.RawMessage
		_ = json.Unmarshal(w.Body.Bytes(), &p)
		if w.Code != 200 || p["starter_codes"] == nil || (mode != "engineer" && p["engineering_notes"] != nil) {
			t.Fatalf("problem metadata %d %s", w.Code, w.Body.String())
		}
	}
	w := httptest.NewRecorder()
	newFixtureServices(nil, nil, &communicationJudge{wrongMetadata: true}).ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/problems/communication-sum", nil))
	if w.Code != 502 {
		t.Fatal("mismatched byte limit accepted")
	}
}
func TestCommunicationRouteSeparation(t *testing.T) {
	j := &communicationJudge{}
	h := newFixtureServices(nil, nil, j)
	good, _ := json.Marshal(communicationRequestFixture())
	for _, route := range []string{"judgements", "compilations"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/"+route, bytes.NewReader(good))
		req.Header.Set("Content-Type", "application/json")
		h.ServeHTTP(w, req)
		if w.Code != 400 {
			t.Fatal("old route accepted multi-source", w.Code)
		}
	}
	ordinary := communicationRequestFixture()
	ordinary.ProblemID = "a-plus-b"
	body, _ := json.Marshal(ordinary)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/communication-judgements", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(w, req)
	if w.Code != 400 || j.calls != 0 {
		t.Fatal("route separation lost", w.Code, j.calls)
	}
}
