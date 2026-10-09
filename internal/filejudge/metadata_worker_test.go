package filejudge

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"oj.local/oj/internal/problems"
	"oj.local/oj/internal/problemstore"
)

func TestWorkerMetadataBoundToEachProblem(t *testing.T) {
	root := t.TempDir()
	for i, p := range problems.List() {
		dir := filepath.Join(root, p.ID)
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
		m := Manifest{Type: p.Type, Checker: "tokens-v1", WallTimeMS: 1000 + int64(i)*300, MemoryLimitBytes: 64 * 1048576, Cases: []CaseFile{{ID: "sample", Input: "sample.in", Expected: "sample.out"}}}
		files := map[string]string{"sample.in": "1 2\n", "sample.out": "3\n"}
		if p.Type == "interactive" {
			m.Checker, m.Interactor, m.Cases[0].Expected = "interaction-v1", "guess-number-v1", ""
			files = map[string]string{"sample.in": "617\n"}
		}
		if p.Type == "communication" {
			m.Checker = "communication-v1"
			m.Communication = &problemstore.CommunicationConfig{ByteLimit: 32}
		}
		writeManifest(t, dir, m)
		for name, contents := range files {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	e := &Engine{dataRoot: root}
	for i, p := range problems.List() {
		w := httptest.NewRecorder()
		e.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/problem/"+p.ID, nil))
		var meta Metadata
		if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &meta) != nil || meta.ProblemID != p.ID || meta.ProblemType != p.Type || meta.CaseCount != 1 || meta.WallTimeMS != 1000+int64(i)*300 {
			t.Fatalf("wrong metadata for %s: %d %s", p.ID, w.Code, w.Body)
		}
	}
	w := httptest.NewRecorder()
	e.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/problem", nil))
	var legacy Metadata
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &legacy) != nil || legacy.ProblemID != "a-plus-b" {
		t.Fatal("legacy metadata broken", w.Code, w.Body)
	}
	e.dataRoot = filepath.Join(root, "does-not-exist")
	w = httptest.NewRecorder()
	e.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/problem/unregistered", nil))
	if w.Code != 404 {
		t.Fatal("unknown id must be rejected before file I/O", w.Code)
	}
}

func TestWorkerRejectsInvalidUploadedRegistration(t *testing.T) {
	root := t.TempDir()
	id := "upload-" + strings.Repeat("a", 32)
	dir := filepath.Join(root, id)
	if os.Mkdir(dir, 0700) != nil || os.WriteFile(filepath.Join(dir, "problem.json"), []byte(`{}`), 0600) != nil {
		t.Fatal("registry test setup failed")
	}
	e := &Engine{store: &problemstore.Store{BuiltinRoot: filepath.Join(root, "missing"), UploadRoot: root}}
	for _, endpoint := range []string{"/problem/" + id, "/problem/upload-" + strings.Repeat("b", 32)} {
		w := httptest.NewRecorder()
		e.Handler().ServeHTTP(w, httptest.NewRequest("GET", endpoint, nil))
		want := 503
		if strings.HasSuffix(endpoint, strings.Repeat("b", 32)) {
			want = 404
		}
		if w.Code != want {
			t.Fatal("invalid registry or absent ID reached file judge", endpoint, w.Code, w.Body)
		}
	}
}

func TestWorkerReadyAdvertisesModesContractVersion(t *testing.T) {
	e := &Engine{image: "sha256:" + strings.Repeat("a", 64)}
	w := httptest.NewRecorder()
	e.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/healthz", nil))
	var health map[string]string
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &health) != nil || health["status"] != "ok" || health["modes_contract_version"] != "modes-interactive/0.1.0" || health["communication_contract_version"] != CommunicationContractVersion {
		t.Fatal("readiness could be mistaken for an old worker", w.Code, w.Body)
	}
	e.tainted.Store(true)
	w = httptest.NewRecorder()
	e.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/healthz", nil))
	if w.Code != 503 || strings.Contains(w.Body.String(), "modes_contract_version") {
		t.Fatal("tainted worker advertised readiness", w.Code, w.Body)
	}
}

func TestPreflightChecksNewRegisteredProblemData(t *testing.T) {
	root, _ := fixture(t)
	jobs := t.TempDir()
	e := &Engine{dataRoot: root, jobsRoot: jobs, image: "sha256:" + strings.Repeat("a", 64), command: func(_ context.Context, args []string, out, _ io.Writer) error {
		var value string
		switch args[0] {
		case "version":
			value = "linux amd64"
		case "info":
			value = "2"
		case "image":
			value = "null"
		case "ps":
			value = ""
		default:
			t.Fatalf("preflight executed before all registered data was checked: %v", args)
		}
		_, _ = io.WriteString(out, value)
		return nil
	}}
	err := e.Preflight(context.Background())
	if err == nil || !strings.HasPrefix(err.Error(), "testdata_preflight_failed_") || err.Error() == "testdata_preflight_failed_a-plus-b" {
		t.Fatal("missing new problem was not rejected", err)
	}
}
