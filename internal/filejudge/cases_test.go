package filejudge

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T) (string, Manifest) {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "a-plus-b")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	m := Manifest{Checker: "tokens-v1", WallTimeMS: 2000, MemoryLimitBytes: 64 * 1048576, Cases: []CaseFile{{ID: "sample", Input: "sample.in", Expected: "sample.out"}}}
	writeManifest(t, dir, m)
	for name, b := range map[string]string{"sample.in": "1 2\n", "sample.out": "3\n"} {
		if os.WriteFile(filepath.Join(dir, name), []byte(b), 0600) != nil {
			t.Fatal("write")
		}
	}
	return root, m
}
func writeManifest(t *testing.T, dir string, m Manifest) {
	t.Helper()
	data, _ := json.Marshal(m)
	if os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0600) != nil {
		t.Fatal("manifest write")
	}
}
func TestCaseSnapshot(t *testing.T) {
	root, _ := fixture(t)
	first, err := Load(root, "a-plus-b")
	if err != nil {
		t.Fatal(err)
	}
	if os.WriteFile(filepath.Join(root, "a-plus-b", "sample.in"), []byte("9 9\n"), 0600) != nil {
		t.Fatal("write")
	}
	second, err := Load(root, "a-plus-b")
	if err != nil {
		t.Fatal(err)
	}
	if string(first.Cases[0].Input) != "1 2\n" || first.SHA256 == second.SHA256 {
		t.Fatal("snapshot mutable or hash unchanged")
	}
}
func TestRejectedCaseFiles(t *testing.T) {
	for _, name := range []string{"missing", "large", "duplicate", "path", "directory", "unknown_json", "trailing_json", "policy"} {
		t.Run(name, func(t *testing.T) {
			root, m := fixture(t)
			dir := filepath.Join(root, "a-plus-b")
			switch name {
			case "missing":
				_ = os.Remove(filepath.Join(dir, "sample.in"))
			case "large":
				_ = os.WriteFile(filepath.Join(dir, "sample.in"), []byte(strings.Repeat("x", MaxFile+1)), 0600)
			case "duplicate":
				m.Cases = append(m.Cases, m.Cases[0])
				writeManifest(t, dir, m)
			case "path":
				m.Cases[0].Input = "../sample.in"
				writeManifest(t, dir, m)
			case "directory":
				_ = os.Remove(filepath.Join(dir, "sample.in"))
				_ = os.Mkdir(filepath.Join(dir, "sample.in"), 0700)
			case "unknown_json":
				_ = os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(`{"unknown":true}`), 0600)
			case "trailing_json":
				f, _ := os.OpenFile(filepath.Join(dir, "manifest.json"), os.O_APPEND|os.O_WRONLY, 0600)
				_, _ = f.WriteString(" {}")
				_ = f.Close()
			case "policy":
				m.WallTimeMS = 1
				writeManifest(t, dir, m)
			}
			if _, err := Load(root, "a-plus-b"); err == nil {
				t.Fatal("unsafe case accepted")
			}
		})
	}
}
func TestCaseSymlinkRejected(t *testing.T) {
	root, _ := fixture(t)
	dir := filepath.Join(root, "a-plus-b")
	_ = os.Remove(filepath.Join(dir, "sample.in"))
	if err := os.Symlink("sample.out", filepath.Join(dir, "sample.in")); err != nil {
		t.Skip("OS does not permit symlink creation")
	}
	if _, err := Load(root, "a-plus-b"); err == nil {
		t.Fatal("root-internal symlink accepted")
	}
}
