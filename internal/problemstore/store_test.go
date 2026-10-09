package problemstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"oj.local/oj/internal/problems"
)

func validUpload() Upload {
	return Upload{Descriptor: Descriptor{
		Title: "教师整数和", Summary: "读取整数并求和。", Difficulty: "入门", Description: "输出 a+b。",
		InputDescription: "两个整数。", OutputDescription: "一个整数。", Notes: "本机演示题。",
		Samples: []problems.Sample{{Input: "1 2\n", Output: "3\n"}}, StarterCode: "int main(){}",
		Type: "batch", Checker: "tokens-v1", WallTimeMS: 2000, MemoryLimitBytes: 64 * 1048576,
		Cases: []CaseFile{{"sample", "sample.in", "sample.out"}},
	}, Files: map[string][]byte{"sample.in": []byte("1 2\n"), "sample.out": []byte("3\n")}}
}

type partData struct {
	name, filename string
	content        []byte
	file           bool
}

func parseParts(t *testing.T, parts []partData) (Upload, error) {
	t.Helper()
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	for _, p := range parts {
		var err error
		if p.file {
			w, e := writer.CreateFormFile(p.name, p.filename)
			err = e
			if e == nil {
				_, err = w.Write(p.content)
			}
		} else {
			w, e := writer.CreateFormField(p.name)
			err = e
			if e == nil {
				_, err = w.Write(p.content)
			}
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return ParseUpload(multipart.NewReader(&b, writer.Boundary()))
}
func uploadParts(u Upload) []partData {
	b, _ := json.Marshal(u.Descriptor)
	parts := []partData{{name: "descriptor", content: b}}
	for name, content := range u.Files {
		parts = append(parts, partData{name: "files", filename: name, content: content, file: true})
	}
	return parts
}

func TestMultipartOrderAndGeneratedPersistentRegistration(t *testing.T) {
	u := validUpload()
	parts := uploadParts(u)
	// Descriptor need not be the first MIME part; references are checked after reading all.
	parts = append(parts[1:], parts[0])
	parsed, err := parseParts(t, parts)
	if err != nil {
		t.Fatal(err)
	}
	uploadRoot := filepath.Join(t.TempDir(), "published")
	s := &Store{BuiltinRoot: "unused-builtins", UploadRoot: uploadRoot}
	p, err := s.Create(parsed)
	if err != nil {
		t.Fatal(err)
	}
	if !uploadIDRE.MatchString(p.ID) || problems.Exists(p.ID) || p.Type != "batch" {
		t.Fatal("uncontrolled or builtin ID", p.ID)
	}
	parsed.Files["sample.in"][0] = '9'
	parsed.Descriptor.Samples[0].Output = "mutated"
	p.Samples[0].Output = "mutated result"
	restarted := &Store{BuiltinRoot: "unused-builtins", UploadRoot: uploadRoot}
	loaded, exists, err := restarted.Get(p.ID)
	if err != nil || !exists || loaded.Samples[0].Output != "3\n" {
		t.Fatal("persistent registration lost or exposed mutable samples", err)
	}
	b, err := os.ReadFile(filepath.Join(uploadRoot, p.ID, "sample.in"))
	if err != nil || string(b) != "1 2\n" {
		t.Fatal("uploaded bytes mutated", err)
	}
	list, err := restarted.List()
	if err != nil || len(list) != len(problems.List())+1 {
		t.Fatal("persistent registration list", err, len(list))
	}
	location, exists, err := restarted.Resolve(p.ID)
	if err != nil || !exists || location.Root != uploadRoot || location.Definition.ID != p.ID {
		t.Fatal("uncontrolled resolve", err)
	}
	entries, err := os.ReadDir(uploadRoot)
	if err != nil || len(entries) != 1 || entries[0].Name() != p.ID {
		t.Fatal("partial publication remains", err)
	}
	second, err := s.Create(validUpload())
	if err != nil || second.ID == p.ID {
		t.Fatal("create overwrote previous registration", err)
	}
}

func TestMultipartRejectsUnsafeAndAmbiguousParts(t *testing.T) {
	for _, name := range []string{"parent_path", "absolute_path", "windows_path", "duplicate_file", "extra_file", "missing_file", "descriptor_twice", "unknown_part", "descriptor_file", "archive", "large_file", "large_descriptor", "unknown_json", "trailing_json", "duplicate_json", "invalid_utf8", "too_many_parts", "total_limit"} {
		t.Run(name, func(t *testing.T) {
			parts := uploadParts(validUpload())
			switch name {
			case "parent_path":
				parts[1].filename = "../sample.in"
			case "absolute_path":
				parts[1].filename = "/tmp/sample.in"
			case "windows_path":
				parts[1].filename = `C:\secrets\sample.in`
			case "duplicate_file":
				parts = append(parts, parts[1])
			case "extra_file":
				parts = append(parts, partData{name: "files", filename: "extra.in", content: []byte("1"), file: true})
			case "missing_file":
				parts = parts[:2]
			case "descriptor_twice":
				parts = append(parts, parts[0])
			case "unknown_part":
				parts[0].name = "command"
			case "descriptor_file":
				parts[0].file = true
				parts[0].filename = "descriptor.json"
			case "archive":
				parts[1].filename = "cases.zip"
			case "large_file":
				parts[1].content = bytes.Repeat([]byte("x"), MaxFile+1)
			case "large_descriptor":
				parts[0].content = bytes.Repeat([]byte(" "), MaxFile+1)
			case "unknown_json":
				parts[0].content = []byte(`{"command":"sh"}`)
			case "trailing_json":
				parts[0].content = append(parts[0].content, []byte(` {}`)...)
			case "duplicate_json":
				parts[0].content = []byte(`{"title":"first","title":"second"}`)
			case "invalid_utf8":
				parts[0].content = []byte{'"', 0xff, '"'}
			case "too_many_parts":
				for i := 0; i < 18; i++ {
					parts = append(parts, partData{name: "files", filename: "extra-" + strings.Repeat("a", i) + ".in", content: []byte("1"), file: true})
				}
			case "total_limit":
				parts = parts[:1]
				for i := 0; i < 16; i++ {
					parts = append(parts, partData{name: "files", filename: strings.Repeat("a", i+1) + ".in", content: bytes.Repeat([]byte("x"), MaxFile), file: true})
				}
			}
			if _, err := parseParts(t, parts); !errors.Is(err, ErrInvalidUpload) {
				t.Fatalf("unsafe parts accepted or uncontrolled error: %v", err)
			}
		})
	}
}

func TestInvalidUploadsNeverCreatePublicOrStagingDirectories(t *testing.T) {
	for _, name := range []string{"type", "checker", "limits", "empty_title", "empty_starter", "too_many_samples", "too_many_cases", "duplicate_id", "duplicate_name", "reserved", "extra_source", "interactive_source", "interactive_expected", "interactive_range", "interactive_extra_token", "interactive_unicode_space"} {
		t.Run(name, func(t *testing.T) {
			u := validUpload()
			if strings.HasPrefix(name, "interactive_") {
				u.Descriptor.Type = "interactive"
				u.Descriptor.Checker = "interaction-v1"
				u.Descriptor.Interactor = "guess-number-v1"
				u.Descriptor.Cases[0].Expected = ""
				delete(u.Files, "sample.out")
				u.Files["sample.in"] = []byte("7\n")
			}
			switch name {
			case "type":
				u.Descriptor.Type = "shell"
			case "checker":
				u.Descriptor.Checker = "custom"
			case "limits":
				u.Descriptor.MemoryLimitBytes = 1024
			case "empty_title":
				u.Descriptor.Title = " "
			case "empty_starter":
				u.Descriptor.StarterCode = ""
			case "too_many_samples":
				u.Descriptor.Samples = make([]problems.Sample, 5)
			case "too_many_cases":
				u.Descriptor.Cases = make([]CaseFile, 9)
			case "duplicate_id":
				u.Descriptor.Cases = append(u.Descriptor.Cases, u.Descriptor.Cases[0])
			case "duplicate_name":
				u.Descriptor.Cases = append(u.Descriptor.Cases, CaseFile{"other", "sample.in", "sample.out"})
			case "reserved":
				u.Descriptor.Cases[0].Input = "con.in"
			case "extra_source":
				u.Files["solution.cpp"] = []byte("int main(){}")
			case "interactive_source":
				u.Descriptor.Interactor = "interactor.cpp"
			case "interactive_expected":
				u.Descriptor.Cases[0].Expected = "sample.out"
				u.Files["sample.out"] = []byte("7")
			case "interactive_range":
				u.Files["sample.in"] = []byte("1001\n")
			case "interactive_extra_token":
				u.Files["sample.in"] = []byte("7 8\n")
			case "interactive_unicode_space":
				u.Files["sample.in"] = []byte("7\u00a0")
			}
			root := filepath.Join(t.TempDir(), "not-created")
			if _, err := (&Store{UploadRoot: root}).Create(u); !errors.Is(err, ErrInvalidUpload) {
				t.Fatalf("invalid upload accepted: %v", err)
			}
			if _, err := os.Lstat(root); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("validation touched publication root")
			}
		})
	}
}

func TestOnlyCompleteExplicitRegistrationIsDiscoverable(t *testing.T) {
	root := t.TempDir()
	s := &Store{UploadRoot: root}
	for _, id := range []string{"arbitrary-folder", ".staging-upload-" + strings.Repeat("0", 32), "upload-" + strings.Repeat("0", 32)} {
		if err := os.Mkdir(filepath.Join(root, id), 0700); err != nil {
			t.Fatal(err)
		}
	}
	list, err := s.List()
	if err != nil || len(list) != len(problems.List()) {
		t.Fatal("unregistered folder became a problem", err, len(list))
	}
	if _, exists, err := s.Get("upload-" + strings.Repeat("0", 32)); err != nil || exists {
		t.Fatal("missing registration accepted", err)
	}
	p, err := s.Create(validUpload())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, p.ID, "problem.json"), []byte(`{"version":1}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, exists, err := s.Get(p.ID); err == nil || exists || strings.Contains(err.Error(), root) {
		t.Fatal("corrupt registration ignored or path leaked", err)
	}
	if _, err := s.List(); err == nil {
		t.Fatal("corrupt registered problem hidden from list")
	}
}

func TestStoreRejectsSymlinkRootsAndRegistration(t *testing.T) {
	target := t.TempDir()
	parent := t.TempDir()
	link := filepath.Join(parent, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Skip("OS does not permit symlink creation")
	}
	s := &Store{UploadRoot: link}
	if _, err := s.Create(validUpload()); err == nil {
		t.Fatal("symlink upload root accepted")
	}
	if _, err := s.List(); err == nil {
		t.Fatal("symlink upload root listed")
	}
	s = &Store{UploadRoot: target}
	p, err := s.Create(validUpload())
	if err != nil {
		t.Fatal(err)
	}
	reg := filepath.Join(target, p.ID, "problem.json")
	copyName := filepath.Join(target, p.ID, "copy.json")
	if err := os.Rename(reg, copyName); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("copy.json", reg); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.Get(p.ID); err == nil {
		t.Fatal("root-internal registration symlink accepted")
	}
}

func TestInteractiveSecretAndImmutablePublishedType(t *testing.T) {
	u := validUpload()
	u.Descriptor.Type = "interactive"
	u.Descriptor.Checker = "interaction-v1"
	u.Descriptor.Interactor = "guess-number-v1"
	u.Descriptor.Cases[0].Expected = ""
	delete(u.Files, "sample.out")
	u.Files["sample.in"] = []byte(" \t1000\r\n")
	s := &Store{UploadRoot: t.TempDir()}
	p, err := s.Create(u)
	if err != nil || p.Type != "interactive" {
		t.Fatal("interactive upload failed", err)
	}
	loaded, exists, err := s.Get(p.ID)
	if err != nil || !exists || loaded.Type != "interactive" {
		t.Fatal("interactive reload failed", err)
	}
	for _, data := range []string{"0", "1001", "1 2", "1x", "-1", "", "١"} {
		if ValidateInput("interactive", []byte(data)) == nil {
			t.Fatalf("invalid secret accepted: %q", data)
		}
	}
	for _, data := range []string{"1", "500\n", "1000", " \t7\r\n"} {
		if err := ValidateInput("interactive", []byte(data)); err != nil {
			t.Fatalf("valid secret rejected: %q %v", data, err)
		}
	}
}

type failedReader struct{ err error }

func (r failedReader) Read([]byte) (int, error) { return 0, r.err }
func TestMultipartTransportErrorCauseIsAvailableButTextIsSanitized(t *testing.T) {
	original := &http.MaxBytesError{Limit: MaxUpload}
	_, err := ParseUpload(multipart.NewReader(failedReader{err: original}, "valid-boundary"))
	var maxError *http.MaxBytesError
	if !errors.Is(err, ErrInvalidUpload) || !errors.As(err, &maxError) || maxError.Limit != MaxUpload || err.Error() != "invalid_upload: multipart_read_failed" {
		t.Fatal("transport cause lost or text leaked", err)
	}
}
