package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"oj.local/oj/internal/filejudge"
	"oj.local/oj/internal/problems"
	"oj.local/oj/internal/problemstore"
)

func TestEngineeringNotesOnlyInEngineerProblemResponse(t *testing.T) {
	h := newFixtureServices(nil, nil, nil)
	for _, mode := range []string{"student", "teacher", "engineer"} {
		r := httptest.NewRequest("GET", "/api/v1/problems/sort-numbers", nil)
		r.Header.Set("X-OJ-Mode", mode)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		var p map[string]any
		if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &p) != nil {
			t.Fatal(w.Code, w.Body.String())
		}
		_, engineering := p["engineering_notes"]
		if engineering != (mode == "engineer") {
			t.Fatal("engineering projection", mode)
		}
		for _, technical := range []string{"token", "演示解法", "编译与链接", "墙钟", "容器", "启动开销"} {
			if strings.Contains(p["notes"].(string), technical) {
				t.Fatal("public note", technical)
			}
		}
	}
}
func TestSpecialProjectionAndTrustedBindingThroughHTTP(t *testing.T) {
	store := &problemstore.Store{UploadRoot: filepath.Join(t.TempDir(), "upload")}
	source := "int main(){return 0;}"
	p, err := store.Create(problemstore.Upload{Descriptor: problemstore.Descriptor{
		Title: "构造", Summary: "构造", Difficulty: "入门", Description: "排列", InputDescription: "n", OutputDescription: "合法排列", Notes: "可有多个合法答案。", EngineeringNotes: "private-checker-implementation",
		Samples: []problems.Sample{{Input: "4", Output: "2 4 1 3"}}, StarterCode: "int main(){}", Type: "batch", Checker: "cpp-special-v1", CheckerSource: "checker.cpp", WallTimeMS: 2000, MemoryLimitBytes: 64 * 1048576, Cases: []problemstore.CaseFile{{ID: "sample", Input: "sample.in"}},
	}, Files: map[string][]byte{"sample.in": []byte("4\n"), "checker.cpp": []byte(source)}})
	if err != nil {
		t.Fatal(err)
	}
	snap, err := filejudge.LoadWithStore(store, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(source))
	zero, one := 0, 1
	j := &evidenceJudge{result: filejudge.Result{Verdict: "WA", TestdataSHA256: snap.SHA256, Compile: filejudge.CompileResult{Status: "success"}, CheckerCompile: &filejudge.CompileResult{Status: "success", Diagnostics: "private-checker-compile"}, CheckerSourceSHA256: hex.EncodeToString(sum[:]), CheckerArtifactSHA256: strings.Repeat("c", 64), Cases: []filejudge.CaseResult{{ID: "sample", Verdict: "WA", ExitCode: &zero, Input: "secret-n", Stdout: "wrong-permutation", CheckerFeedback: "duplicate number", Checker: &filejudge.CheckerResult{ExitCode: &one, Stderr: "private-checker-log"}}}}}
	h := NewWithCatalog(nil, nil, j, store)
	post := func(mode string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "/api/v1/judgements", strings.NewReader(`{"problem_id":"`+p.ID+`","language":"cpp","source":"int main(){}"}`))
		r.Header.Set("X-OJ-Mode", mode)
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	for _, mode := range []string{"student", "teacher", "engineer"} {
		w := post(mode)
		if w.Code != 200 {
			t.Fatal(mode, w.Code, w.Body.String())
		}
		body := w.Body.String()
		if mode == "student" {
			for _, hidden := range []string{"secret-n", "duplicate number", "checker", "wrong-permutation"} {
				if strings.Contains(body, hidden) {
					t.Fatal("student leak", hidden)
				}
			}
		}
		if mode == "teacher" {
			if !strings.Contains(body, "duplicate number") {
				t.Fatal("feedback absent")
			}
			for _, hidden := range []string{"checker_compile", "checker_source_sha256", "private-checker-log", "exit_code", "oom_killed"} {
				if strings.Contains(body, hidden) {
					t.Fatal("teacher debug leak", hidden)
				}
			}
		}
		if mode == "engineer" && (!strings.Contains(body, "checker_source_sha256") || !strings.Contains(body, "private-checker-log")) {
			t.Fatal("engineer evidence absent")
		}
	}
	j.result.CheckerSourceSHA256 = strings.Repeat("f", 64)
	if w := post("student"); w.Code != 502 {
		t.Fatal("wrong checker source accepted", w.Body.String())
	}
	j.result.CheckerSourceSHA256 = hex.EncodeToString(sum[:])
	j.result.Cases[0].ID = "other"
	if w := post("teacher"); w.Code != 502 {
		t.Fatal("wrong case accepted", w.Body.String())
	}
}
