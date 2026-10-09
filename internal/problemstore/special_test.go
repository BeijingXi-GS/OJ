package problemstore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func specialUpload() Upload {
	u := validUpload()
	u.Descriptor.Checker, u.Descriptor.CheckerSource = "cpp-special-v1", "checker.cpp"
	u.Descriptor.EngineeringNotes = "仅供调试的说明。"
	u.Files["checker.cpp"] = []byte("#include <fstream>\nint main(int argc,char** argv){return argc == 4 ? 0 : 2;}\n")
	return u
}

func TestSpecialUploadPersistenceAndFixedSource(t *testing.T) {
	for _, expected := range []bool{true, false} {
		u := specialUpload()
		if !expected {
			u.Descriptor.Cases[0].Expected = ""
			delete(u.Files, "sample.out")
		}
		parsed, err := parseParts(t, uploadParts(u))
		if err != nil {
			t.Fatal(err)
		}
		root := filepath.Join(t.TempDir(), "uploads")
		store := &Store{UploadRoot: root}
		p, err := store.Create(parsed)
		if err != nil {
			t.Fatal(err)
		}
		parsed.Files["checker.cpp"][0] = '!'
		reloaded := &Store{UploadRoot: root}
		found, ok, err := reloaded.Get(p.ID)
		if err != nil || !ok || found.EngineeringNotes != u.Descriptor.EngineeringNotes {
			t.Fatal("persisted definition lost", err)
		}
		b, err := os.ReadFile(filepath.Join(root, p.ID, "checker.cpp"))
		if err != nil || string(b) != string(u.Files["checker.cpp"]) {
			t.Fatal("source not immutable", err)
		}
		b, err = os.ReadFile(filepath.Join(root, p.ID, "manifest.json"))
		if err != nil {
			t.Fatal(err)
		}
		m, err := ParseManifest(b)
		if err != nil || m.CheckerSource != "checker.cpp" || (m.Cases[0].Expected != "") != expected {
			t.Fatal("manifest changed", err)
		}
	}
}

func TestSpecialUploadRejectsInvalidSourcesAndPolicy(t *testing.T) {
	for _, name := range []string{"missing", "extra_cpp", "path", "wrong_source", "empty", "whitespace", "nul", "utf8", "large", "five_cases", "wrong_pair", "interactive", "builtin_source", "ordinary_source"} {
		t.Run(name, func(t *testing.T) {
			u := specialUpload()
			switch name {
			case "missing":
				delete(u.Files, "checker.cpp")
			case "extra_cpp":
				u.Files["extra.cpp"] = []byte("int main(){}")
			case "path":
				u.Descriptor.CheckerSource = "../checker.cpp"
			case "wrong_source":
				u.Descriptor.CheckerSource = "other.cpp"
			case "empty":
				u.Files["checker.cpp"] = nil
			case "whitespace":
				u.Files["checker.cpp"] = []byte("  \n\t")
			case "nul":
				u.Files["checker.cpp"] = []byte("int main(){}\x00")
			case "utf8":
				u.Files["checker.cpp"] = []byte{255}
			case "large":
				u.Files["checker.cpp"] = []byte(strings.Repeat("a", MaxFile+1))
			case "five_cases":
				for i := 0; i < 4; i++ {
					u.Descriptor.Cases = append(u.Descriptor.Cases, CaseFile{ID: string(rune('a' + i)), Input: string(rune('a'+i)) + ".in"})
					u.Files[string(rune('a'+i))+".in"] = []byte("4\n")
				}
			case "wrong_pair":
				u.Descriptor.Cases[0].Expected = "different.out"
				u.Files["different.out"] = u.Files["sample.out"]
				delete(u.Files, "sample.out")
			case "interactive":
				u.Descriptor.Type = "interactive"
				u.Descriptor.Checker = "interaction-v1"
				u.Descriptor.Interactor = "guess-number-v1"
			case "builtin_source":
				u.Descriptor.Checker = "nonconsecutive-permutation-v1"
			case "ordinary_source":
				u.Descriptor.Checker = "tokens-v1"
			}
			if _, _, err := validateUpload(u); err == nil {
				t.Fatal("invalid upload accepted")
			}
			if _, err := parseParts(t, uploadParts(u)); err == nil {
				t.Fatal("invalid multipart accepted")
			}
		})
	}
}

func TestBuiltinConstructionInputValidation(t *testing.T) {
	for _, good := range []string{"1", " 1000\r\n", "\v\f4\t", "0004"} {
		if _, err := ParseConstructionSize([]byte(good)); err != nil {
			t.Fatal(good, err)
		}
	}
	for _, bad := range []string{"", "0", "1001", "-1", "+4", "4 5", "4.0", "1e2", "４", "\u00a04", "00004"} {
		if _, err := ParseConstructionSize([]byte(bad)); err == nil {
			t.Fatal("invalid n accepted", bad)
		}
	}
	u := validUpload()
	u.Descriptor.Checker = "nonconsecutive-permutation-v1"
	u.Descriptor.Cases[0].Expected = ""
	delete(u.Files, "sample.out")
	u.Files["sample.in"] = []byte("4\n")
	if _, _, err := validateUpload(u); err != nil {
		t.Fatal(err)
	}
	u.Files["sample.in"] = []byte("4 5\n")
	if _, _, err := validateUpload(u); err == nil {
		t.Fatal("invalid builtin test input published")
	}
}
