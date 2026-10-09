package problemstore

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"strings"
	"unicode/utf8"

	"oj.local/oj/internal/problems"
)

type Descriptor struct {
	Title             string               `json:"title"`
	Summary           string               `json:"summary"`
	Difficulty        string               `json:"difficulty"`
	Description       string               `json:"description"`
	InputDescription  string               `json:"input_description"`
	OutputDescription string               `json:"output_description"`
	Notes             string               `json:"notes"`
	EngineeringNotes  string               `json:"engineering_notes,omitempty"`
	Samples           []problems.Sample    `json:"samples"`
	StarterCode       string               `json:"starter_code"`
	StarterCodes      map[string]string    `json:"starter_codes,omitempty"`
	Type              string               `json:"type"`
	Checker           string               `json:"checker"`
	CheckerSource     string               `json:"checker_source,omitempty"`
	WallTimeMS        int64                `json:"wall_time_ms"`
	MemoryLimitBytes  int64                `json:"memory_limit_bytes"`
	Cases             []CaseFile           `json:"cases"`
	Interactor        string               `json:"interactor,omitempty"`
	Communication     *CommunicationConfig `json:"communication,omitempty"`
}

type Upload struct {
	Descriptor Descriptor
	Files      map[string][]byte
}

// Keep a transport cause available to the HTTP adapter (for its 413 response)
// while the error's text remains fixed and never exposes request content.
type uploadReadError struct{ cause error }

func (e *uploadReadError) Error() string        { return "invalid_upload: multipart_read_failed" }
func (e *uploadReadError) Unwrap() error        { return e.cause }
func (e *uploadReadError) Is(target error) bool { return target == ErrInvalidUpload }

func (d Descriptor) definition(id string) problems.Definition {
	return problems.Definition{ID: id, Title: d.Title, Summary: d.Summary, Difficulty: d.Difficulty, Description: d.Description, InputDescription: d.InputDescription, OutputDescription: d.OutputDescription, Notes: d.Notes, EngineeringNotes: d.EngineeringNotes, Samples: append([]problems.Sample(nil), d.Samples...), StarterCode: d.StarterCode, StarterCodes: problems.CloneStarterCodes(d.StarterCodes), Type: d.Type}
}

func (d Descriptor) manifest() Manifest {
	return Manifest{Type: d.Type, Checker: d.Checker, CheckerSource: d.CheckerSource, Interactor: d.Interactor, WallTimeMS: d.WallTimeMS, MemoryLimitBytes: d.MemoryLimitBytes, Cases: d.Cases, Communication: d.Communication}
}

// strictJSON rejects duplicate keys before the typed decoder rejects unknown fields.
func strictJSON(data []byte, target any) error {
	if len(data) > MaxFile || !utf8.Valid(data) {
		return invalid("descriptor_limit_or_encoding")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	var value func(string) error
	value = func(field string) error {
		token, err := d.Token()
		if err != nil {
			return err
		}
		if token == nil && (field == "communication" || field == "starter_codes") {
			return invalid("json_null_configuration")
		}
		if delim, ok := token.(json.Delim); ok {
			switch delim {
			case '{':
				seen := map[string]bool{}
				for d.More() {
					key, err := d.Token()
					if err != nil {
						return err
					}
					s, ok := key.(string)
					if !ok || seen[s] {
						return invalid("json_duplicate_key")
					}
					// Every store schema/map key is lowercase ASCII. The typed
					// decoder otherwise accepts case-insensitive aliases, including
					// some Unicode folds, as later writes to the same field.
					for _, c := range s {
						if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '_' {
							return invalid("json_key_not_canonical")
						}
					}
					seen[s] = true
					if err := value(s); err != nil {
						return err
					}
				}
			case '[':
				for d.More() {
					if err := value(""); err != nil {
						return err
					}
				}
			default:
				return invalid("json_invalid")
			}
			if _, err := d.Token(); err != nil {
				return err
			}
		}
		return nil
	}
	if value("") != nil {
		return invalid("json_invalid")
	}
	if _, err := d.Token(); err != io.EOF {
		return invalid("json_trailing")
	}
	d = json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if d.Decode(target) != nil {
		return invalid("json_unknown_or_invalid")
	}
	return nil
}

func validateDefinition(p problems.Definition) error {
	fields := []struct {
		value    string
		limit    int
		required bool
	}{
		{p.Title, 256, true}, {p.Summary, 1024, true}, {p.Difficulty, 64, true},
		{p.Description, 16384, true}, {p.InputDescription, 8192, true}, {p.OutputDescription, 8192, true},
		{p.Notes, 8192, false}, {p.EngineeringNotes, 8192, false}, {p.StarterCode, MaxFile, true},
	}
	for _, field := range fields {
		if !utf8.ValidString(field.value) || len(field.value) > field.limit || strings.IndexByte(field.value, 0) >= 0 || (field.required && strings.TrimSpace(field.value) == "") {
			return invalid("problem_text_invalid")
		}
	}
	if p.Type != "batch" && p.Type != "interactive" && p.Type != "communication" {
		return invalid("problem_type_invalid")
	}
	if p.Type == "communication" {
		if len(p.StarterCodes) != 2 || p.StarterCode != p.StarterCodes["sender"] {
			return invalid("communication_starter_invalid")
		}
		for _, role := range []string{"sender", "receiver"} {
			if ValidateCheckerSource([]byte(p.StarterCodes[role])) != nil {
				return invalid("communication_starter_invalid")
			}
		}
	} else if p.StarterCodes != nil {
		return invalid("communication_starter_forbidden")
	}
	if len(p.Samples) == 0 || len(p.Samples) > 4 {
		return invalid("public_samples_invalid")
	}
	for _, sample := range p.Samples {
		if !utf8.ValidString(sample.Input) || !utf8.ValidString(sample.Output) || len(sample.Input) > 4096 || len(sample.Output) > 4096 || strings.IndexByte(sample.Input, 0) >= 0 || strings.IndexByte(sample.Output, 0) >= 0 {
			return invalid("public_samples_invalid")
		}
	}
	return nil
}

func validateUpload(u Upload) (problems.Definition, Manifest, error) {
	m, err := NormalizeManifest(u.Descriptor.manifest())
	if err != nil {
		return problems.Definition{}, m, err
	}
	p := u.Descriptor.definition("")
	p.Type = m.Type
	if err := validateDefinition(p); err != nil {
		return p, m, err
	}
	descriptorBytes, err := json.Marshal(u.Descriptor)
	if err != nil || len(descriptorBytes) > MaxFile {
		return p, m, invalid("descriptor_limit")
	}
	needed := map[string]bool{}
	if m.CheckerSource != "" {
		needed[m.CheckerSource] = true
	}
	for _, c := range m.Cases {
		needed[c.Input] = true
		if c.Expected != "" {
			needed[c.Expected] = true
		}
	}
	if len(u.Files) > 16 || len(u.Files) != len(needed) {
		return p, m, invalid("files_missing_or_extra")
	}
	total := len(descriptorBytes)
	for name, content := range u.Files {
		if !needed[name] {
			return p, m, invalid("files_missing_or_extra")
		}
		if len(content) > MaxFile {
			return p, m, invalid("file_limit")
		}
		total += len(content)
	}
	if total > MaxUpload {
		return p, m, invalid("upload_limit")
	}
	for _, c := range m.Cases {
		if err := ValidateCaseInput(m, u.Files[c.Input]); err != nil {
			return p, m, err
		}
	}
	if m.CheckerSource != "" {
		if err := ValidateCheckerSource(u.Files[m.CheckerSource]); err != nil {
			return p, m, err
		}
	}
	return p, m, nil
}

// ParseUpload reads every part with a fixed cap, even when the caller has not yet
// applied its HTTP whole-request limit. The router also caps encoded body bytes.
func ParseUpload(reader *multipart.Reader) (Upload, error) {
	u := Upload{Files: map[string][]byte{}}
	descriptorSeen, total, parts := false, 0, 0
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return u, &uploadReadError{cause: err}
		}
		parts++
		if parts > 17 {
			_ = part.Close()
			return u, invalid("parts_limit")
		}
		kind, params, err := mime.ParseMediaType(part.Header.Get("Content-Disposition"))
		if err != nil || kind != "form-data" {
			_ = part.Close()
			return u, invalid("part_invalid")
		}
		name, filename := params["name"], params["filename"]
		_, hasFilename := params["filename"]
		data, err := io.ReadAll(io.LimitReader(part, MaxFile+1))
		if err != nil {
			return u, &uploadReadError{cause: err}
		}
		if len(data) > MaxFile {
			return u, invalid("file_limit")
		}
		_ = part.Close()
		total += len(data)
		if total > MaxUpload {
			return u, invalid("upload_limit")
		}
		if name == "descriptor" && !hasFilename && !descriptorSeen {
			if err := strictJSON(data, &u.Descriptor); err != nil {
				return u, err
			}
			descriptorSeen = true
		} else if name == "files" && hasFilename && (SafeFilename(filename, ".in") || SafeFilename(filename, ".out") || filename == "checker.cpp") {
			if _, exists := u.Files[filename]; exists {
				return u, invalid("duplicate_file")
			}
			u.Files[filename] = data
		} else {
			return u, invalid("part_invalid")
		}
	}
	if !descriptorSeen {
		return u, invalid("descriptor_missing")
	}
	if _, _, err := validateUpload(u); err != nil {
		return u, err
	}
	return u, nil
}
