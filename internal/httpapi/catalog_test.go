package httpapi

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"oj.local/oj/internal/compilation"
	"oj.local/oj/internal/filejudge"
	"oj.local/oj/internal/problems"
)

type catalogJudge struct {
	metadataIDs []string
	judgeIDs    []string
	wrongID     bool
}

func (j *catalogJudge) Judge(_ context.Context, r filejudge.Request) (filejudge.Result, *compilation.Error) {
	j.judgeIDs = append(j.judgeIDs, r.ProblemID)
	if r.SourceUnits != nil {
		return communicationFixtureResult(r), nil
	}
	snap, err := filejudge.Load("../../testdata/problems", r.ProblemID)
	if err != nil {
		panic(err)
	}
	return filejudge.Result{Verdict: "CE", TestdataSHA256: snap.SHA256, Compile: filejudge.CompileResult{Status: "compile_error"}, Cases: []filejudge.CaseResult{}}, nil
}
func (j *catalogJudge) Metadata(_ context.Context, id string) (filejudge.Metadata, *compilation.Error) {
	j.metadataIDs = append(j.metadataIDs, id)
	if strings.HasPrefix(id, "communication-") {
		return (&communicationJudge{}).Metadata(context.Background(), id)
	}
	responseID := id
	if j.wrongID {
		responseID = "a-plus-b"
	}
	count := 5
	if id == "a-plus-b" {
		count = 4
	}
	typeName, checker, wall := "batch", "tokens-v1", int64(2000)
	if id == "guess-number" {
		typeName, checker, count, wall = "interactive", "interaction-v1", 3, 5000
	}
	return filejudge.Metadata{ProblemID: responseID, Enabled: true, CaseCount: count, WallTimeMS: wall, MemoryLimitBytes: 64 * 1048576, Checker: checker, ProblemType: typeName}, nil
}

func TestAllCatalogProblemsUseMatchingMetadataAndCompileRoute(t *testing.T) {
	j := &catalogJudge{}
	compileCalls := 0
	ordinaryCount := 0
	h := newFixtureServices(nil, compilerFunc(func(context.Context, string) (compilation.Result, *compilation.Error) {
		compileCalls++
		return compilation.Result{Status: "success"}, nil
	}), j)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/problems", nil))
	var list struct {
		Problems []problemSummary `json:"problems"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &list) != nil || len(list.Problems) != len(problems.List()) {
		t.Fatalf("catalog list: %d %s", w.Code, w.Body.String())
	}
	for i, definition := range problems.List() {
		if definition.Type != "communication" {
			ordinaryCount++
		}
		if list.Problems[i].ID != definition.ID {
			t.Fatal("API list differs from trusted allowlist")
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/problems/"+definition.ID, nil))
		var p problem
		if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &p) != nil || p.ID != definition.ID || p.Judge.ProblemID != definition.ID || !p.Judge.Enabled || p.StarterCode != definition.StarterCode {
			t.Fatalf("problem/metadata mismatch: %d %s", w.Code, w.Body.String())
		}
		if len(j.metadataIDs) != i+1 || j.metadataIDs[i] != definition.ID {
			t.Fatal("metadata requested for a different problem")
		}
		data, _ := json.Marshal(filejudge.Request{ProblemID: definition.ID, Language: "cpp", Source: "int main(){}"})
		routes := []string{"compilations", "judgements"}
		if definition.Type == "communication" {
			request := communicationRequestFixture()
			request.ProblemID = definition.ID
			data, _ = json.Marshal(request)
			routes = []string{"communication-judgements"}
		}
		for _, route := range routes {
			r := httptest.NewRequest("POST", "/api/v1/"+route, strings.NewReader(string(data)))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != 200 {
				t.Fatalf("%s/%s rejected: %d %s", route, definition.ID, w.Code, w.Body.String())
			}
		}
		if j.judgeIDs[i] != definition.ID {
			t.Fatal("judge received wrong problem ID")
		}
	}
	if compileCalls != ordinaryCount {
		t.Fatalf("compile calls: %d", compileCalls)
	}
}

func TestUnknownCatalogIDNeverCallsServices(t *testing.T) {
	j := &catalogJudge{}
	compilerCalled := false
	h := newFixtureServices(nil, compilerFunc(func(context.Context, string) (compilation.Result, *compilation.Error) {
		compilerCalled = true
		return compilation.Result{}, nil
	}), j)
	for _, id := range []string{"missing", "../a-plus-b", `a-plus-b\sample.in`, "a-plus-b%2f.."} {
		for _, route := range []string{"compilations", "judgements"} {
			data, _ := json.Marshal(filejudge.Request{ProblemID: id, Language: "cpp", Source: "int main(){}"})
			r := httptest.NewRequest("POST", "/api/v1/"+route, strings.NewReader(string(data)))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != 404 {
				t.Fatalf("unknown ID: %d %s", w.Code, w.Body.String())
			}
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/problems/missing", nil))
	if w.Code != 404 || compilerCalled || len(j.judgeIDs) != 0 || len(j.metadataIDs) != 0 {
		t.Fatal("unknown problem reached a service")
	}
}

func TestProblemDetailRejectsCrossProblemMetadata(t *testing.T) {
	w := httptest.NewRecorder()
	newFixtureServices(nil, nil, &catalogJudge{wrongID: true}).ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/problems/palindrome", nil))
	if w.Code != 502 || !strings.Contains(w.Body.String(), "judge_protocol_error") {
		t.Fatalf("cross-problem metadata accepted: %d %s", w.Code, w.Body.String())
	}
}
