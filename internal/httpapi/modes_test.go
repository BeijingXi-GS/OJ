package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"oj.local/oj/internal/compilation"
	"oj.local/oj/internal/filejudge"
	"oj.local/oj/internal/problemstore"
)

type evidenceJudge struct {
	result filejudge.Result
	calls  int
}

func (j *evidenceJudge) Judge(context.Context, filejudge.Request) (filejudge.Result, *compilation.Error) {
	j.calls++
	return j.result, nil
}
func TestModeProjectionThroughHTTP(t *testing.T) {
	zero, one := 0, 1
	j := &evidenceJudge{result: filejudge.Result{RunID: "debug-run", Verdict: "WA", Message: "wrong answer", Compiler: "GCC 14.2.0", Standard: "c++20", SourceSHA256: "source-debug", TestdataSHA256: "testdata-debug", ArtifactSHA256: "artifact-debug", RuntimeImage: "image-debug", Compile: filejudge.CompileResult{Status: "success", Diagnostics: "my compiler diagnostic"}, InteractorCompile: &filejudge.CompileResult{Status: "success", Diagnostics: "private interactor diagnostic"}, InteractorSourceSHA256: "interactor-hash", InteractorArtifactSHA256: "interactor-artifact", Cases: []filejudge.CaseResult{{ID: "secret-case", Verdict: "WA", ExitCode: &zero, Input: "secret-input", Expected: "secret-answer", Stdout: "hidden-output", Stderr: "hidden-error", Interactor: &filejudge.InteractorResult{ExitCode: &one, Stdout: "private-conversation", Stderr: "private-rejection"}}}}}
	h := newFixtureServices(nil, nil, j)
	snap, err := filejudge.Load("../../testdata/problems", "guess-number")
	if err != nil {
		t.Fatal(err)
	}
	j.result.TestdataSHA256 = snap.SHA256
	row := j.result.Cases[0]
	j.result.Cases = nil
	for _, c := range snap.Cases {
		row.ID = c.ID
		j.result.Cases = append(j.result.Cases, row)
	}
	j.result.InteractorSourceSHA256 = filejudge.InteractorSourceDigest()
	j.result.InteractorArtifactSHA256 = strings.Repeat("f", 64)
	for _, mode := range []string{"", "student", "teacher", "engineer"} {
		r := httptest.NewRequest("POST", "/api/v1/judgements", strings.NewReader(`{"problem_id":"guess-number","language":"cpp","source":"int main(){}"}`))
		r.Header.Set("Content-Type", "application/json")
		if mode != "" {
			r.Header.Set("X-OJ-Mode", mode)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		var body map[string]any
		if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &body) != nil {
			t.Fatalf("%s: %d %s", mode, w.Code, w.Body)
		}
		if body["passed_cases"] != float64(0) || body["total_cases"] != float64(len(snap.Cases)) {
			t.Fatal("aggregate lost")
		}
		if mode == "" || mode == "student" {
			if body["mode"] != "student" {
				t.Fatal("default is not student")
			}
			for _, secret := range []string{"secret-case", "secret-input", "secret-answer", "hidden-output", "hidden-error", "private-conversation", "private-rejection", "debug-run", "interactor-hash", "image-debug"} {
				if strings.Contains(w.Body.String(), secret) {
					t.Fatalf("student leak: %s", secret)
				}
			}
			for _, key := range []string{"cases", "run_id", "source_sha256", "testdata_sha256", "artifact_sha256", "runtime_image", "interactor_compile", "interactor_source_sha256", "interactor_artifact_sha256"} {
				if _, found := body[key]; found {
					t.Fatalf("student field leak: %s", key)
				}
			}
		} else if mode == "teacher" {
			if !strings.Contains(w.Body.String(), "secret-input") || !strings.Contains(w.Body.String(), "private-conversation") {
				t.Fatal("teacher evidence missing")
			}
			for _, key := range []string{"run_id", "runtime_image", "interactor_compile", "exit_code", "oom_killed", "participant_stopped"} {
				if strings.Contains(w.Body.String(), `"`+key+`"`) {
					t.Fatalf("teacher debug leak: %s", key)
				}
			}
		} else if body["run_id"] != "debug-run" || !strings.Contains(w.Body.String(), "participant_stopped") {
			t.Fatal("engineer evidence missing")
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("result cacheable")
		}
	}
	j.result.Verdict = "SYSTEM_ERROR"
	j.result.Message = "internal private failure"
	r := httptest.NewRequest("POST", "/api/v1/judgements", strings.NewReader(`{"problem_id":"guess-number","language":"cpp","source":"int main(){}"}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if strings.Contains(w.Body.String(), "internal private failure") {
		t.Fatal("student system failure leak")
	}
}

func TestInvalidModeNeverReachesWorker(t *testing.T) {
	j := &evidenceJudge{}
	for _, values := range [][]string{{"admin"}, {"teacher,student"}, {"Teacher"}, {"student", "teacher"}} {
		r := httptest.NewRequest("POST", "/api/v1/judgements", strings.NewReader(`{}`))
		for _, v := range values {
			r.Header.Add("X-OJ-Mode", v)
		}
		w := httptest.NewRecorder()
		newFixtureServices(nil, nil, j).ServeHTTP(w, r)
		if w.Code != 400 || j.calls != 0 {
			t.Fatal("invalid mode reached worker")
		}
	}
}

func uploadBody(t *testing.T, extra int) ([]byte, string) {
	t.Helper()
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	// A literal descriptor exercises the same multipart route used by browsers.
	part, err := writer.CreateFormField("descriptor")
	if err != nil {
		t.Fatal(err)
	}
	_, err = part.Write([]byte(`{"title":"测试题","summary":"求和","difficulty":"入门","description":"求和","input_description":"两个整数","output_description":"和","notes":"","samples":[{"input":"1 2","output":"3"}],"starter_code":"int main(){}","type":"batch","checker":"tokens-v1","wall_time_ms":2000,"memory_limit_bytes":67108864,"cases":[{"id":"sample","input":"sample.in","expected":"sample.out"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	for _, data := range []struct{ name, content string }{{"sample.in", "1 2\n"}, {"sample.out", "3\n"}} {
		p, err := writer.CreateFormFile("files", data.name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = p.Write([]byte(data.content))
	}
	if extra > 0 {
		p, err := writer.CreateFormField("padding")
		if err != nil {
			t.Fatal(err)
		}
		_, _ = p.Write(bytes.Repeat([]byte("x"), extra))
	}
	if writer.Close() != nil {
		t.Fatal("multipart close")
	}
	return b.Bytes(), writer.FormDataContentType()
}

func TestUploadModeOriginAndPersistentCatalog(t *testing.T) {
	store := &problemstore.Store{UploadRoot: filepath.Join(t.TempDir(), "uploaded")}
	h := NewWithCatalog(nil, nil, nil, store)
	data, contentType := uploadBody(t, 0)
	for _, tc := range []struct {
		mode, origin string
		want         int
	}{{"", "", 403}, {"student", "", 403}, {"teacher", "https://untrusted.example", 403}, {"teacher", "null", 403}, {"teacher", "http://localhost:5173", 201}, {"engineer", "", 201}} {
		r := httptest.NewRequest("POST", "/api/v1/problems", bytes.NewReader(data))
		r.Header.Set("Content-Type", contentType)
		r.Header.Set("X-OJ-Mode", tc.mode)
		if tc.origin != "" {
			r.Header.Set("Origin", tc.origin)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Fatalf("%+v: %d %s", tc, w.Code, w.Body)
		}
		if w.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Fatal("unexpected CORS")
		}
		if tc.want == 201 {
			var body struct {
				ID string `json:"id"`
			}
			if json.Unmarshal(w.Body.Bytes(), &body) != nil || !strings.HasPrefix(body.ID, "upload-") {
				t.Fatal("invalid generated ID")
			}
			fresh := NewWithCatalog(nil, nil, nil, &problemstore.Store{UploadRoot: store.UploadRoot})
			detail := httptest.NewRecorder()
			fresh.ServeHTTP(detail, httptest.NewRequest("GET", "/api/v1/problems/"+body.ID, nil))
			if detail.Code != 200 {
				t.Fatal("upload not persisted")
			}
			run := httptest.NewRequest("POST", "/api/v1/compilations", strings.NewReader(`{"problem_id":"`+body.ID+`","language":"cpp","source":"int main(){}"}`))
			run.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			fresh.ServeHTTP(w, run)
			if w.Code != 503 || !strings.Contains(w.Body.String(), "compiler_unavailable") {
				t.Fatal("registered upload rejected before compiler")
			}
		}
	}
	entries, err := os.ReadDir(store.UploadRoot)
	if err != nil || len(entries) != 2 {
		t.Fatal("forbidden upload changed catalog")
	}
	bad := httptest.NewRequest("POST", "/api/v1/problems", strings.NewReader(`{}`))
	bad.Header.Set("X-OJ-Mode", "teacher")
	bad.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, bad)
	if w.Code != 400 {
		t.Fatal("non-multipart accepted")
	}
}

func TestEngineerDebugIsReadOnlyAndModeLimited(t *testing.T) {
	j := &catalogJudge{}
	h := newFixtureServices(nil, nil, j)
	for _, mode := range []string{"", "student", "teacher", "engineer"} {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/debug", nil)
		r.Header.Set("X-OJ-Mode", mode)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		want := 403
		if mode == "engineer" {
			want = 200
		}
		if w.Code != want {
			t.Fatal("debug mode boundary")
		}
	}
	if len(j.judgeIDs) != 0 || len(j.metadataIDs) != 1 {
		t.Fatal("debug executed code or called metadata for disallowed modes")
	}
}

func TestRequestedProblemTypeCannotAcceptWrongWorkerProtocol(t *testing.T) {
	zero := 0
	for _, tc := range []struct {
		id          string
		interactive bool
	}{{"guess-number", false}, {"a-plus-b", true}} {
		r := filejudge.Result{Verdict: "AC", Compile: filejudge.CompileResult{Status: "success"}, Cases: []filejudge.CaseResult{{ID: "sample", Verdict: "AC", ExitCode: &zero}}}
		if tc.interactive {
			r.InteractorCompile = &filejudge.CompileResult{Status: "success"}
			r.InteractorSourceSHA256 = filejudge.InteractorSourceDigest()
			r.InteractorArtifactSHA256 = strings.Repeat("f", 64)
			r.Cases[0].Interactor = &filejudge.InteractorResult{ExitCode: &zero}
		}
		j := &evidenceJudge{result: r}
		body := `{"problem_id":"` + tc.id + `","language":"cpp","source":"int main(){}"}`
		request := httptest.NewRequest("POST", "/api/v1/judgements", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		newFixtureServices(nil, nil, j).ServeHTTP(w, request)
		if w.Code != 502 || !strings.Contains(w.Body.String(), "judge_protocol_error") {
			t.Fatal("wrong protocol projected as requested type")
		}
	}
}

func TestEncodedUploadLimitIs413WithoutPublication(t *testing.T) {
	store := &problemstore.Store{UploadRoot: filepath.Join(t.TempDir(), "uploaded")}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	descriptor := map[string]any{"title": "限制测试", "summary": "限制测试", "difficulty": "入门", "description": "说明", "input_description": "输入", "output_description": "输出", "notes": "", "samples": []map[string]string{{"input": "", "output": ""}}, "starter_code": "int main(){}", "type": "batch", "checker": "exact-v1", "wall_time_ms": 2000, "memory_limit_bytes": 67108864}
	cases := make([]problemstore.CaseFile, 8)
	for i := range cases {
		id := string(rune('a' + i))
		cases[i] = problemstore.CaseFile{ID: id, Input: id + ".in", Expected: id + ".out"}
	}
	descriptor["cases"] = cases
	raw, _ := json.Marshal(descriptor)
	part, err := writer.CreateFormField("descriptor")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write(raw)
	for _, c := range cases {
		for _, name := range []string{c.Input, c.Expected} {
			part, err := writer.CreateFormFile("files", name)
			if err != nil {
				t.Fatal(err)
			}
			_, _ = part.Write(bytes.Repeat([]byte("x"), 65536))
		}
	}
	_ = writer.Close()
	request := httptest.NewRequest("POST", "/api/v1/problems", bytes.NewReader(body.Bytes()))
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("X-OJ-Mode", "teacher")
	w := httptest.NewRecorder()
	NewWithCatalog(nil, nil, nil, store).ServeHTTP(w, request)
	if w.Code != 413 || !strings.Contains(w.Body.String(), "upload_too_large") {
		t.Fatalf("encoded limit: %d %s", w.Code, w.Body)
	}
	if _, err := os.Stat(store.UploadRoot); !os.IsNotExist(err) {
		t.Fatal("oversized upload published files")
	}
}
