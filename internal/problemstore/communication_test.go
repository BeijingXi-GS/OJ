package problemstore

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func communicationUpload() Upload {
	u := validUpload()
	u.Descriptor.Type, u.Descriptor.Checker = "communication", "communication-v1"
	u.Descriptor.StarterCodes = map[string]string{"sender": u.Descriptor.StarterCode, "receiver": "#include <iostream>\nint main(){int n;std::cin>>n;std::cout<<n;}"}
	u.Descriptor.Communication = &CommunicationConfig{ByteLimit: 32}
	return u
}

func TestCommunicationUploadPersistsOwnedTemplatesAndConfig(t *testing.T) {
	u := communicationUpload()
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
	original := u.Descriptor.StarterCodes["receiver"]
	parsed.Descriptor.StarterCodes["receiver"] = "mutated"
	parsed.Descriptor.Communication.ByteLimit = 1
	p.StarterCodes["sender"] = "mutated result"
	fresh := &Store{UploadRoot: root}
	p, ok, err := fresh.Get(p.ID)
	if err != nil || !ok || p.Type != "communication" || p.StarterCodes["receiver"] != original || p.StarterCode != p.StarterCodes["sender"] {
		t.Fatal("templates lost or aliased", err)
	}
	raw, err := os.ReadFile(filepath.Join(root, p.ID, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	m, err := ParseManifest(raw)
	if err != nil || m.Communication == nil || m.Communication.ByteLimit != 32 || m.Checker != "communication-v1" {
		t.Fatal("config not frozen", err)
	}
	config := m.Communication
	normalized, err := NormalizeManifest(m)
	if err != nil {
		t.Fatal(err)
	}
	config.ByteLimit = 1
	if normalized.Communication.ByteLimit != 32 {
		t.Fatal("normalized config aliased")
	}
	p.StarterCodes["receiver"] = "changed after reload"
	p, ok, err = fresh.Get(p.ID)
	if err != nil || !ok || p.StarterCodes["receiver"] != original {
		t.Fatal("reload map alias", err)
	}
}

func TestCommunicationUploadRejectsAmbiguousConfiguration(t *testing.T) {
	for _, name := range []string{"missing_config", "zero", "too_large", "missing_role", "extra_role", "wrong_sender", "empty_receiver", "nul", "bad_utf8", "large_role", "missing_expected", "wrong_pair", "interactor", "checker_source", "batch_config", "batch_templates", "five_cases"} {
		t.Run(name, func(t *testing.T) {
			u := communicationUpload()
			switch name {
			case "missing_config":
				u.Descriptor.Communication = nil
			case "zero":
				u.Descriptor.Communication.ByteLimit = 0
			case "too_large":
				u.Descriptor.Communication.ByteLimit = MaxFile + 1
			case "missing_role":
				delete(u.Descriptor.StarterCodes, "receiver")
			case "extra_role":
				u.Descriptor.StarterCodes["other"] = "int main(){}"
			case "wrong_sender":
				u.Descriptor.StarterCode = "different"
			case "empty_receiver":
				u.Descriptor.StarterCodes["receiver"] = " "
			case "nul":
				u.Descriptor.StarterCodes["receiver"] = "int main(){}\x00"
			case "bad_utf8":
				u.Descriptor.StarterCodes["receiver"] = string([]byte{0xff})
			case "large_role":
				u.Descriptor.StarterCodes["receiver"] = strings.Repeat("x", MaxFile+1)
			case "missing_expected":
				u.Descriptor.Cases[0].Expected = ""
				delete(u.Files, "sample.out")
			case "wrong_pair":
				u.Descriptor.Cases[0].Expected = "answer.out"
				u.Files["answer.out"] = u.Files["sample.out"]
				delete(u.Files, "sample.out")
			case "interactor":
				u.Descriptor.Interactor = "guess-number-v1"
			case "checker_source":
				u.Descriptor.CheckerSource = "checker.cpp"
				u.Files["checker.cpp"] = []byte("int main(){}")
			case "batch_config":
				u.Descriptor.Type = "batch"
				u.Descriptor.Checker = "tokens-v1"
				u.Descriptor.StarterCodes = nil
			case "batch_templates":
				u.Descriptor.Type = "batch"
				u.Descriptor.Checker = "tokens-v1"
				u.Descriptor.Communication = nil
			case "five_cases":
				for i := 0; i < 4; i++ {
					id := strings.Repeat("a", i+1)
					u.Descriptor.Cases = append(u.Descriptor.Cases, CaseFile{id, id + ".in", id + ".out"})
					u.Files[id+".in"] = []byte("1 2\n")
					u.Files[id+".out"] = []byte("3\n")
				}
			}
			root := filepath.Join(t.TempDir(), "untouched")
			if _, err := (&Store{UploadRoot: root}).Create(u); !errors.Is(err, ErrInvalidUpload) {
				t.Fatal("invalid communication accepted", err)
			}
			if _, err := os.Lstat(root); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("invalid upload touched store")
			}
		})
	}
	for _, raw := range []string{`{"byte_limit":1,"byte_limit":2}`, `{"byte_limit":32,"path":"/secret"}`} {
		u := communicationUpload()
		data, _ := json.Marshal(u.Descriptor)
		data = []byte(strings.Replace(string(data), `{"byte_limit":32}`, raw, 1))
		parts := uploadParts(u)
		parts[0].content = data
		if _, err := parseParts(t, parts); !errors.Is(err, ErrInvalidUpload) {
			t.Fatal("ambiguous config accepted", err)
		}
	}
}

func TestCommunicationUploadRejectsCaseAliasesAndExplicitNull(t *testing.T) {
	u := communicationUpload()
	descriptor, err := json.Marshal(u.Descriptor)
	if err != nil {
		t.Fatal(err)
	}
	starterCodes, err := json.Marshal(u.Descriptor.StarterCodes)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, before, after string }{
		{"byte_alias", `"byte_limit":32`, `"byte_limit":32,"Byte_Limit":1`},
		{"config_alias", `"communication":{"byte_limit":32}`, `"communication":null,"Communication":{"byte_limit":32}`},
		{"starter_alias", `"starter_codes":`, `"Starter_Codes":`},
		{"role_alias", `"receiver":`, `"Receiver":`},
		{"unicode_fold_alias", `"starter_codes":`, `"ſtarter_codes":`},
		{"config_null", `"communication":{"byte_limit":32}`, `"communication":null`},
		{"starter_null", `"starter_codes":` + string(starterCodes), `"starter_codes":null`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			changed := strings.Replace(string(descriptor), tc.before, tc.after, 1)
			if changed == string(descriptor) {
				t.Fatal("fixture did not inject alias")
			}
			parts := uploadParts(u)
			parts[0].content = []byte(changed)
			if _, err := parseParts(t, parts); !errors.Is(err, ErrInvalidUpload) {
				t.Fatal("ambiguous descriptor accepted", err)
			}
		})
	}
	for _, problemType := range []string{"batch", "interactive"} {
		u := validUpload()
		if problemType == "interactive" {
			u.Descriptor.Type, u.Descriptor.Checker, u.Descriptor.Interactor = "interactive", "interaction-v1", "guess-number-v1"
			u.Descriptor.Cases[0].Expected = ""
			delete(u.Files, "sample.out")
			u.Files["sample.in"] = []byte("617\n")
		}
		for _, field := range []string{"communication", "starter_codes"} {
			parts := uploadParts(u)
			parts[0].content = append(parts[0].content[:len(parts[0].content)-1], []byte(`,"`+field+`":null}`)...)
			if _, err := parseParts(t, parts); !errors.Is(err, ErrInvalidUpload) {
				t.Fatal("legacy type accepted explicit null configuration", problemType, field, err)
			}
		}
	}
}

func TestCommunicationManifestReloadRejectsAliasAndNull(t *testing.T) {
	u := communicationUpload()
	root := filepath.Join(t.TempDir(), "uploads")
	store := &Store{UploadRoot: root}
	p, err := store.Create(u)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, p.ID, "manifest.json")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, before, after string }{
		{"nested_alias", `"byte_limit":32`, `"byte_limit":32,"Byte_Limit":1`},
		{"top_alias", `"communication":{"byte_limit":32}`, `"communication":null,"Communication":{"byte_limit":32}`},
		{"explicit_null", `"communication":{"byte_limit":32}`, `"communication":null`},
		{"type_alias", `"type":"communication"`, `"type":"communication","Type":"batch"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			changed := strings.Replace(string(original), tc.before, tc.after, 1)
			if changed == string(original) {
				t.Fatal("fixture did not inject alias")
			}
			if _, err := ParseManifest([]byte(changed)); !errors.Is(err, ErrInvalidUpload) {
				t.Fatal("manifest parser accepted alias/null", err)
			}
			if err := os.WriteFile(path, []byte(changed), 0600); err != nil {
				t.Fatal(err)
			}
			if _, ok, err := (&Store{UploadRoot: root}).Get(p.ID); err == nil || ok {
				t.Fatal("reload accepted alias/null", err)
			}
		})
	}
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := store.Get(p.ID); err != nil || !ok {
		t.Fatal("canonical manifest failed after restoration", err)
	}
}
