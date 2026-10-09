package filejudge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"oj.local/oj/internal/problems"
)

func TestCatalogRequestValidation(t *testing.T) {
	for _, p := range problems.List() {
		req := Request{ProblemID: p.ID, Language: "cpp", Source: "int main(){}"}
		if p.Type == "communication" {
			req = Request{ProblemID: p.ID, SourceUnits: []SourceUnit{{Role: "receiver", Language: "cpp", Source: "int main(){}"}, {Role: "sender", Language: "cpp", Source: "int main(){}"}}}
		}
		if err := ValidateRequest(req); err != nil {
			t.Fatalf("%s rejected: %v", p.ID, err)
		}
	}
	for _, id := range []string{"../a-plus-b", "a-plus-b/sample.in", `a-plus-b\sample.in`, "a-plus-b%2f..", "unregistered"} {
		if err := ValidateRequest(Request{ProblemID: id, Language: "cpp", Source: "int main(){}"}); err == nil || err.Status != 404 {
			t.Fatalf("unsafe ID accepted: %q %v", id, err)
		}
	}
}

func TestUnregisteredDirectoryRejectedBeforeFilesystem(t *testing.T) {
	for _, root := range []string{filepath.Join(t.TempDir(), "missing-root"), t.TempDir()} {
		if _, err := Load(root, "unregistered"); err == nil || err.Error() != "unknown_problem" {
			t.Fatalf("unknown ID reached filesystem: %v", err)
		}
	}
	root, _ := fixture(t)
	if err := os.Rename(filepath.Join(root, "a-plus-b"), filepath.Join(root, "unregistered")); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(root, "unregistered"); err == nil || err.Error() != "unknown_problem" {
		t.Fatalf("directory alone enabled a problem: %v", err)
	}
}

// Calculate answers independently from the C++ demonstration templates and the
// HTTP smoke solutions, including constraint endpoints present in the fixtures.
func fixtureAnswer(t *testing.T, id string, input []byte) string {
	t.Helper()
	r := bytes.NewReader(input)
	if id == "a-plus-b" || id == "communication-a-plus-b" {
		var a, b int64
		limit := int64(1e9)
		if id == "communication-a-plus-b" {
			limit = 1e12
		}
		if _, err := fmt.Fscan(r, &a, &b); err != nil || a < -limit || a > limit || b < -limit || b > limit {
			t.Fatal("invalid A+B fixture")
		}
		var extra string
		if _, err := fmt.Fscan(r, &extra); err != io.EOF {
			t.Fatal("extra A+B fixture data")
		}
		return fmt.Sprint(a + b)
	}
	if id == "palindrome" {
		var s string
		if _, err := fmt.Fscan(r, &s); err != nil || len(s) < 1 || len(s) > 1000 {
			t.Fatal("invalid palindrome fixture")
		}
		for i := range s {
			if s[i] < 'a' || s[i] > 'z' {
				t.Fatal("non-lowercase palindrome fixture")
			}
		}
		for i := 0; i < len(s)/2; i++ {
			if s[i] != s[len(s)-1-i] {
				return "NO"
			}
		}
		return "YES"
	}
	var n int
	if _, err := fmt.Fscan(r, &n); err != nil || n < 1 || n > 1000 {
		t.Fatal("invalid array length")
	}
	v := make([]int64, n)
	for i := range v {
		if _, err := fmt.Fscan(r, &v[i]); err != nil || v[i] < -1e9 || v[i] > 1e9 {
			t.Fatal("invalid array element")
		}
	}
	var extra string
	if _, err := fmt.Fscan(r, &extra); err == nil {
		t.Fatal("extra fixture array element")
	}
	switch id {
	case "array-sum", "communication-sum":
		var sum int64
		for _, x := range v {
			sum += x
		}
		return fmt.Sprint(sum)
	case "maximum-element", "communication-maximum":
		sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
		return fmt.Sprint(v[len(v)-1])
	case "sort-numbers":
		sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
		tokens := make([]string, len(v))
		for i, x := range v {
			tokens[i] = fmt.Sprint(x)
		}
		return strings.Join(tokens, " ")
	default:
		t.Fatalf("no answer oracle for %s", id)
	}
	return ""
}

func TestRegisteredFixtureAnswers(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "problems")
	for _, p := range problems.List() {
		t.Run(p.ID, func(t *testing.T) {
			snapshot, err := Load(root, p.ID)
			if err != nil {
				t.Fatal(err)
			}
			if p.Type == "interactive" {
				if p.ID != "guess-number" || len(snapshot.Cases) != 3 || snapshot.Manifest.Type != "interactive" || snapshot.Manifest.Checker != "interaction-v1" || snapshot.Interactor != "guess-number-v1" || snapshot.Manifest.WallTimeMS != 5000 || snapshot.Manifest.MemoryLimitBytes != 64*1048576 {
					t.Fatal("unexpected interactive problem policy")
				}
				secrets := map[int]bool{}
				for _, c := range snapshot.Cases {
					var secret int
					r := bytes.NewReader(c.Input)
					if _, err := fmt.Fscan(r, &secret); err != nil || secret < 1 || secret > 1000 || len(c.Expected) != 0 || secrets[secret] {
						t.Fatal("invalid or duplicate interactive secret")
					}
					var extra string
					if _, err := fmt.Fscan(r, &extra); err != io.EOF {
						t.Fatal("extra interactive secret data")
					}
					secrets[secret] = true
				}
				if !secrets[1] || !secrets[500] || !secrets[1000] {
					t.Fatal("interactive endpoint fixtures absent")
				}
				return
			}
			if p.ID == "construct-permutation" {
				if len(snapshot.Cases) != 6 || snapshot.Manifest.Checker != "nonconsecutive-permutation-v1" {
					t.Fatal("construction policy incorrect")
				}
				for _, c := range snapshot.Cases {
					var n int
					if _, err := fmt.Fscan(bytes.NewReader(c.Input), &n); err != nil || n < 1 || n > 1000 || len(c.Expected) != 0 {
						t.Fatal("invalid construction fixture")
					}
				}
				for _, s := range p.Samples {
					if verdict, _, err := constructionVerdict([]byte(s.Input), []byte(s.Output)); err != nil || verdict != "AC" {
						t.Fatal("invalid construction public sample")
					}
				}
				return
			}
			if p.Type == "communication" {
				limit := int64(32)
				if p.ID == "communication-maximum" {
					limit = 16
				}
				if len(snapshot.Cases) != 4 || snapshot.Manifest.Checker != "communication-v1" || snapshot.Manifest.Communication == nil || snapshot.Manifest.Communication.ByteLimit != limit || len(p.StarterCodes) != 2 || p.StarterCode != p.StarterCodes["sender"] {
					t.Fatal("unexpected communication policy")
				}
				for _, c := range snapshot.Cases {
					if answer := fixtureAnswer(t, p.ID, c.Input); !check("tokens-v1", []byte(answer), c.Expected) {
						t.Fatal("incorrect communication standard answer", c.ID)
					}
				}
				for _, s := range p.Samples {
					if !check("tokens-v1", []byte(fixtureAnswer(t, p.ID, []byte(s.Input))), []byte(s.Output)) {
						t.Fatal("incorrect communication sample")
					}
				}
				return
			}
			if len(snapshot.Cases) < 4 || snapshot.Manifest.Checker != "tokens-v1" || snapshot.Manifest.WallTimeMS != 2000 || snapshot.Manifest.MemoryLimitBytes != 64*1048576 {
				t.Fatal("unexpected initial problem policy")
			}
			for _, c := range snapshot.Cases {
				if answer := fixtureAnswer(t, p.ID, c.Input); !check("tokens-v1", []byte(answer), c.Expected) {
					t.Fatalf("incorrect standard answer: %s/%s", p.ID, c.ID)
				}
			}
			for _, s := range p.Samples {
				if !check("tokens-v1", []byte(fixtureAnswer(t, p.ID, []byte(s.Input))), []byte(s.Output)) {
					t.Fatalf("incorrect public sample: %s", p.ID)
				}
			}
		})
	}
}

func TestMetadataClientProblemBinding(t *testing.T) {
	for _, name := range []string{"valid", "wrong_problem", "missing_problem", "bad_policy", "extra_field", "trailing", "unavailable"} {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" || r.URL.Path != "/problem/sort-numbers" {
					t.Errorf("wrong metadata endpoint: %s %s", r.Method, r.URL.Path)
				}
				if name == "unavailable" {
					w.WriteHeader(503)
					return
				}
				meta := Metadata{ProblemID: "sort-numbers", Enabled: true, CaseCount: 5, WallTimeMS: 2000, MemoryLimitBytes: 64 * 1048576, Checker: "tokens-v1", ProblemType: "batch"}
				if name == "wrong_problem" {
					meta.ProblemID = "a-plus-b"
				}
				if name == "bad_policy" {
					meta.CaseCount = 9
				}
				data, _ := json.Marshal(meta)
				if name == "missing_problem" {
					data = bytes.Replace(data, []byte(`"problem_id":"sort-numbers",`), nil, 1)
				}
				if name == "extra_field" {
					data = append(data[:len(data)-1], []byte(`,"unexpected":true}`)...)
				}
				if name == "trailing" {
					data = append(data, []byte(` {}`)...)
				}
				_, _ = w.Write(data)
			}))
			defer server.Close()
			client, err := NewClient(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			meta, failure := client.Metadata(context.Background(), "sort-numbers")
			if name == "valid" {
				if failure != nil || meta.ProblemID != "sort-numbers" || meta.CaseCount != 5 {
					t.Fatalf("valid metadata rejected: %+v %v", meta, failure)
				}
			} else {
				want := 502
				if name == "unavailable" {
					want = 503
				}
				if failure == nil || failure.Status != want {
					t.Fatalf("invalid metadata accepted: %s %v", name, failure)
				}
			}
		})
	}
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }))
	defer server.Close()
	client, _ := NewClient(server.URL)
	if _, failure := client.Metadata(context.Background(), "../a-plus-b"); failure == nil || failure.Status != 404 || called {
		t.Fatalf("unknown ID reached metadata service: %v", failure)
	}
}
