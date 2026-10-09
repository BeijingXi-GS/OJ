package filejudge

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"oj.local/oj/internal/problems"
	"oj.local/oj/internal/problemstore"
)

func TestEngineAcceptsRegisteredUploadBeforeCompile(t *testing.T) {
	for _, problemType := range []string{"batch", "interactive"} {
		t.Run(problemType, func(t *testing.T) {
			root := t.TempDir()
			store := &problemstore.Store{BuiltinRoot: filepath.Join(root, "missing-builtins"), UploadRoot: filepath.Join(root, "published")}
			upload := problemstore.Upload{Descriptor: problemstore.Descriptor{
				Title: "Uploaded local problem", Summary: "Test registered upload routing", Difficulty: "入门", Description: "A local test problem.",
				InputDescription: "Read the protocol or two integers.", OutputDescription: "Write the answer.", Notes: "Trusted file test.", StarterCode: "int main(){}",
				Samples: []problems.Sample{{Input: "1 2\n", Output: "3\n"}},
				Type:    problemType, Checker: "tokens-v1", WallTimeMS: 2000, MemoryLimitBytes: 64 * 1048576,
				Cases: []problemstore.CaseFile{{ID: "sample", Input: "sample.in", Expected: "sample.out"}},
			}, Files: map[string][]byte{"sample.in": []byte("1 2\n"), "sample.out": []byte("3\n")}}
			if problemType == "interactive" {
				upload.Descriptor.Checker, upload.Descriptor.Interactor = "interaction-v1", "guess-number-v1"
				upload.Descriptor.Cases[0].Expected = ""
				upload.Files = map[string][]byte{"sample.in": []byte("617\n")}
			}
			problem, err := store.Create(upload)
			if err != nil {
				t.Fatal("trusted upload setup failed", err)
			}
			jobs := filepath.Join(root, "jobs")
			if os.Mkdir(jobs, 0700) != nil {
				t.Fatal("trusted jobs setup failed")
			}
			creates := 0
			e := &Engine{store: store, jobsRoot: jobs, image: "sha256:" + strings.Repeat("a", 64), busy: make(chan struct{}, 1)}
			e.command = func(_ context.Context, args []string, _, _ io.Writer) error {
				switch args[0] {
				case "create":
					creates++
					// Stop at the observable compile boundary; no Docker or
					// submitted program is run by this trusted test.
					return errors.New("trusted fake refuses execution")
				case "ps":
					return nil // definite no container created by this fake.
				default:
					t.Fatalf("unexpected command after fake create failure: %v", args)
					return errors.New("unexpected fake control")
				}
			}
			result, failure := e.Judge(context.Background(), Request{ProblemID: problem.ID, Language: "cpp", Source: "int main(){}"})
			if failure != nil || creates != 1 || result.TestdataSHA256 == "" || result.Compile.Status != "system_error" || !strings.Contains(result.Message, "container_create_failed") || e.tainted.Load() {
				t.Fatalf("registered upload did not enter isolated compile: %+v %+v creates=%d", result, failure, creates)
			}
			if files, err := os.ReadDir(jobs); err != nil || len(files) != 0 {
				t.Fatal("fake rejected compile left a job directory", files, err)
			}
			_, failure = e.Judge(context.Background(), Request{ProblemID: "upload-" + strings.Repeat("b", 32), Language: "cpp", Source: "int main(){}"})
			if failure == nil || failure.Status != 404 || creates != 1 {
				t.Fatal("unregistered upload reached compile", failure, creates)
			}
		})
	}
}
