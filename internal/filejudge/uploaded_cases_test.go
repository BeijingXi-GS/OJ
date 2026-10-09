package filejudge

import (
	"os"
	"path/filepath"
	"testing"

	"oj.local/oj/internal/problems"
	"oj.local/oj/internal/problemstore"
)

func TestUploadedSnapshotUsesRegisteredStoreAndOwnedBytes(t *testing.T) {
	s := &problemstore.Store{BuiltinRoot: filepath.Join("..", "..", "testdata", "problems"), UploadRoot: t.TempDir()}
	u := problemstore.Upload{Descriptor: problemstore.Descriptor{Title: "教师题", Summary: "求和", Difficulty: "入门", Description: "求和。", InputDescription: "两个整数", OutputDescription: "整数和", Samples: []problems.Sample{{Input: "1 2", Output: "3"}}, StarterCode: "int main(){}", Type: "batch", Checker: "tokens-v1", WallTimeMS: 2000, MemoryLimitBytes: 64 * 1048576, Cases: []problemstore.CaseFile{{ID: "first", Input: "first.in", Expected: "first.out"}}}, Files: map[string][]byte{"first.in": []byte("1 2\n"), "first.out": []byte("3\n")}}
	p, err := s.Create(u)
	if err != nil {
		t.Fatal(err)
	}
	first, err := LoadWithStore(s, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if first.Manifest.Type != "batch" || first.Interactor != "" || len(first.Cases) != 1 {
		t.Fatal("wrong uploaded snapshot")
	}
	if _, err := Load(s.BuiltinRoot, p.ID); err == nil {
		t.Fatal("builtin-only loader enabled arbitrary uploaded id")
	}
	if err := os.WriteFile(filepath.Join(s.UploadRoot, p.ID, "first.in"), []byte("9 9\n"), 0600); err != nil {
		t.Fatal(err)
	}
	second, err := LoadWithStore(s, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if first.SHA256 == second.SHA256 || string(first.Cases[0].Input) != "1 2\n" {
		t.Fatal("run snapshot mutated or digest not changed")
	}
}

func TestBuiltinInteractiveSnapshotIsSecretOnly(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "problems")
	snapshot, err := Load(root, "guess-number")
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Manifest.Type != "interactive" || snapshot.Manifest.Checker != "interaction-v1" || snapshot.Interactor != "guess-number-v1" || len(snapshot.Cases) != 3 {
		t.Fatal("incomplete interactive snapshot")
	}
	for i, secret := range []string{"1\n", "500\n", "1000\n"} {
		if string(snapshot.Cases[i].Input) != secret || len(snapshot.Cases[i].Expected) != 0 {
			t.Fatal("incorrect or duplicated answer input", i)
		}
	}
}
