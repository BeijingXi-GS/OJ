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

func interactiveWireResult() Result {
	r := testResult()
	code := 0
	r.InteractorCompile = &CompileResult{Status: "success"}
	r.InteractorSourceSHA256 = InteractorSourceDigest()
	r.InteractorArtifactSHA256 = strings.Repeat("f", 64)
	r.Cases[0].Input = "617\n"
	r.Cases[0].Expected = ""
	r.Cases[0].Stdout = "! 617\n"
	r.Cases[0].Interactor = &InteractorResult{ExitCode: &code, Stdout: "1000\n"}
	return r
}

func TestInteractiveClientEvidenceContradictions(t *testing.T) {
	for _, name := range []string{"AC", "supervised_WA", "natural_WA", "natural_137_RE", "true_OOM_MLE", "TLE", "OLE", "interactor_system_error", "unmarked_137_WA", "AC_interactor_reject", "AC_supervised", "RE_supervised", "WA_interactor_OOM", "batch_interactor", "missing_stop", "null_stop", "null_interactor", "null_interactor_compile", "null_input", "null_truncation", "unknown_interactor_field", "bad_interactor_hash", "compile_system_error", "compile_error_false_AC"} {
		t.Run(name, func(t *testing.T) {
			r := interactiveWireResult()
			r.SourceSHA256 = hash([]byte("int main(){}"))
			want := true
			participant, interactor := 0, 0
			switch name {
			case "supervised_WA":
				r.Verdict = "WA"
				r.Cases[0].Verdict = "WA"
				participant, interactor = 137, 1
				r.Cases[0].Interactor.ParticipantStopped = true
			case "natural_WA":
				r.Verdict = "WA"
				r.Cases[0].Verdict = "WA"
				interactor = 1
			case "natural_137_RE":
				r.Verdict = "RE"
				r.Cases[0].Verdict = "RE"
				participant, interactor = 137, 137
			case "true_OOM_MLE":
				r.Verdict = "MLE"
				r.Cases[0].Verdict = "MLE"
				participant, interactor = 137, 137
				r.Cases[0].OOMKilled = true
			case "TLE", "OLE":
				r.Verdict = name
				r.Cases[0].Verdict = name
				participant, interactor = 137, 137
			case "interactor_system_error":
				r.Verdict = "SYSTEM_ERROR"
				r.Cases[0].Verdict = "SYSTEM_ERROR"
				interactor = 137
				r.Cases[0].Interactor.OOMKilled = true
			case "unmarked_137_WA":
				r.Verdict = "WA"
				r.Cases[0].Verdict = "WA"
				participant, interactor = 137, 1
				want = false
			case "AC_interactor_reject":
				interactor = 1
				want = false
			case "AC_supervised":
				r.Cases[0].Interactor.ParticipantStopped = true
				want = false
			case "RE_supervised":
				r.Verdict = "RE"
				r.Cases[0].Verdict = "RE"
				participant, interactor = 137, 1
				r.Cases[0].Interactor.ParticipantStopped = true
				want = false
			case "WA_interactor_OOM":
				r.Verdict = "WA"
				r.Cases[0].Verdict = "WA"
				interactor = 1
				r.Cases[0].Interactor.OOMKilled = true
				want = false
			case "batch_interactor":
				r.InteractorCompile = nil
				r.InteractorSourceSHA256 = ""
				r.InteractorArtifactSHA256 = ""
				want = false
			case "bad_interactor_hash":
				r.InteractorArtifactSHA256 = "bad"
				want = false
			case "compile_system_error":
				r.Verdict = "SYSTEM_ERROR"
				r.InteractorCompile.Status = "compile_error"
				r.InteractorArtifactSHA256 = ""
				r.Cases = []CaseResult{}
			case "compile_error_false_AC":
				r.InteractorCompile.Status = "compile_error"
				r.InteractorArtifactSHA256 = ""
				want = false
			}
			if len(r.Cases) > 0 {
				r.Cases[0].ExitCode = &participant
				r.Cases[0].Interactor.ExitCode = &interactor
			}
			data, _ := json.Marshal(r)
			switch name {
			case "missing_stop":
				data = bytes.Replace(data, []byte(`,"participant_stopped":false`), nil, 1)
				want = false
			case "null_stop":
				data = bytes.Replace(data, []byte(`"participant_stopped":false`), []byte(`"participant_stopped":null`), 1)
				want = false
			case "null_input":
				data = bytes.Replace(data, []byte(`"input":"617\n"`), []byte(`"input":null`), 1)
				want = false
			case "null_truncation":
				data = bytes.Replace(data, []byte(`"input_truncated":false`), []byte(`"input_truncated":null`), 1)
				want = false
			case "null_interactor_compile":
				var m map[string]json.RawMessage
				_ = json.Unmarshal(data, &m)
				m["interactor_compile"] = json.RawMessage(`null`)
				data, _ = json.Marshal(m)
				want = false
			case "null_interactor":
				data = bytes.Replace(data, []byte(`"interactor":{`), []byte(`"interactor":null,"discarded":{`), 1)
				want = false
			case "unknown_interactor_field":
				data = bytes.Replace(data, []byte(`"participant_stopped":false`), []byte(`"participant_stopped":false,"fake":true`), 1)
				want = false
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(data) }))
			defer server.Close()
			client, err := NewClient(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			_, failure := client.Judge(context.Background(), Request{ProblemID: "guess-number", Language: "cpp", Source: "int main(){}"})
			if want && failure != nil {
				t.Fatalf("valid %s rejected: %v", name, failure)
			}
			if !want && (failure == nil || failure.Status != 502) {
				t.Fatalf("invalid %s accepted: %v", name, failure)
			}
		})
	}
}

func TestUploadedInteractiveMetadataAndRequiredType(t *testing.T) {
	id := "upload-" + strings.Repeat("a", 32)
	for _, name := range []string{"valid", "missing_type", "null_type", "wrong_checker", "unknown_type"} {
		t.Run(name, func(t *testing.T) {
			meta := Metadata{ProblemID: id, Enabled: true, CaseCount: 3, WallTimeMS: 5000, MemoryLimitBytes: 64 * 1048576, Checker: "interaction-v1", ProblemType: "interactive"}
			if name == "wrong_checker" {
				meta.Checker = "tokens-v1"
			}
			if name == "unknown_type" {
				meta.ProblemType = "custom"
			}
			data, _ := json.Marshal(meta)
			if name == "missing_type" {
				data = bytes.Replace(data, []byte(`,"problem_type":"interactive"`), nil, 1)
			}
			if name == "null_type" {
				data = bytes.Replace(data, []byte(`"problem_type":"interactive"`), []byte(`"problem_type":null`), 1)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/problem/"+id {
					t.Error("wrong uploaded endpoint")
				}
				_, _ = w.Write(data)
			}))
			defer server.Close()
			client, _ := NewClient(server.URL)
			_, failure := client.Metadata(context.Background(), id)
			if name == "valid" {
				if failure != nil {
					t.Fatal(failure)
				}
			} else if failure == nil || failure.Status != 502 {
				t.Fatal("invalid metadata accepted")
			}
		})
	}
}
