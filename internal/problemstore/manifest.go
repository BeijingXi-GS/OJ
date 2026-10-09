// Package problemstore provides the local MVP's explicitly registered problem catalog.
// Uploaded data never chooses a host path, executable, image or command.
package problemstore

import (
	"errors"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	MaxCases              = 8
	MaxFile               = 65536
	MaxUpload             = 1048576
	MaxSpecialCases       = 4
	MaxCommunicationCases = 4
)

var ErrInvalidUpload = errors.New("invalid_upload")
var idRE = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)
var uploadIDRE = regexp.MustCompile(`^upload-[a-f0-9]{32}$`)
var secretRE = regexp.MustCompile(`^[ \t\r\n]*[0-9]{1,4}[ \t\r\n]*$`)

func invalid(code string) error { return errors.Join(ErrInvalidUpload, errors.New(code)) }

type CaseFile struct {
	ID       string `json:"id"`
	Input    string `json:"input"`
	Expected string `json:"expected,omitempty"`
}

type Manifest struct {
	Type             string               `json:"type,omitempty"`
	Checker          string               `json:"checker"`
	CheckerSource    string               `json:"checker_source,omitempty"`
	Interactor       string               `json:"interactor,omitempty"`
	WallTimeMS       int64                `json:"wall_time_ms"`
	MemoryLimitBytes int64                `json:"memory_limit_bytes"`
	Cases            []CaseFile           `json:"cases"`
	Communication    *CommunicationConfig `json:"communication,omitempty"`
}

type CommunicationConfig struct {
	ByteLimit int64 `json:"byte_limit"`
}

func SafeFilename(name, suffix string) bool {
	if filepath.Base(name) != name || strings.ContainsAny(name, `/\:,`) || !strings.HasSuffix(name, suffix) {
		return false
	}
	stem := strings.TrimSuffix(name, suffix)
	if !idRE.MatchString(stem) {
		return false
	}
	upper := strings.ToUpper(stem)
	return upper != "CON" && upper != "PRN" && upper != "AUX" && upper != "NUL" && !strings.HasPrefix(upper, "COM") && !strings.HasPrefix(upper, "LPT")
}

// NormalizeManifest preserves old batch manifests while enforcing the same policy
// during upload, registration reload and worker snapshot loading.
func NormalizeManifest(m Manifest) (Manifest, error) {
	if m.Type == "" {
		m.Type = "batch"
	}
	if m.WallTimeMS < 500 || m.WallTimeMS > 5000 || m.MemoryLimitBytes < 32*1048576 || m.MemoryLimitBytes > 256*1048576 || len(m.Cases) == 0 || len(m.Cases) > MaxCases {
		return m, invalid("manifest_policy_invalid")
	}
	switch m.Type {
	case "batch":
		if !IsBatchChecker(m.Checker) || m.Interactor != "" || m.Communication != nil {
			return m, invalid("manifest_checker_invalid")
		}
		if m.Checker == "cpp-special-v1" {
			if m.CheckerSource != "checker.cpp" || len(m.Cases) > MaxSpecialCases {
				return m, invalid("manifest_checker_source_invalid")
			}
		} else if m.CheckerSource != "" {
			return m, invalid("manifest_checker_source_forbidden")
		}
	case "interactive":
		if m.Checker != "interaction-v1" || m.Interactor != "guess-number-v1" || m.CheckerSource != "" || m.Communication != nil {
			return m, invalid("manifest_interactor_invalid")
		}
	case "communication":
		if m.Checker != "communication-v1" || m.Interactor != "" || m.CheckerSource != "" || m.Communication == nil || m.Communication.ByteLimit < 1 || m.Communication.ByteLimit > MaxFile || len(m.Cases) > MaxCommunicationCases {
			return m, invalid("manifest_communication_invalid")
		}
	default:
		return m, invalid("manifest_type_invalid")
	}
	ids, files := map[string]bool{}, map[string]bool{}
	for _, c := range m.Cases {
		if !idRE.MatchString(c.ID) || ids[c.ID] || !SafeFilename(c.Input, ".in") || files[c.Input] {
			return m, invalid("case_invalid")
		}
		ids[c.ID], files[c.Input] = true, true
		if m.Type == "interactive" || m.Checker == "nonconsecutive-permutation-v1" {
			if c.Expected != "" {
				return m, invalid("expected_forbidden")
			}
		} else {
			if m.Type == "communication" && c.Expected != strings.TrimSuffix(c.Input, ".in")+".out" {
				return m, invalid("communication_expected_pair_invalid")
			}
			if m.Checker == "cpp-special-v1" && c.Expected == "" {
				continue
			}
			if m.Checker == "cpp-special-v1" && c.Expected != strings.TrimSuffix(c.Input, ".in")+".out" {
				return m, invalid("checker_expected_pair_invalid")
			}
			if !SafeFilename(c.Expected, ".out") || files[c.Expected] {
				return m, invalid("case_invalid")
			}
			files[c.Expected] = true
		}
	}
	m.Cases = append([]CaseFile(nil), m.Cases...)
	if m.Communication != nil {
		config := *m.Communication
		m.Communication = &config
	}
	return m, nil
}

func IsBatchChecker(checker string) bool {
	return checker == "tokens-v1" || checker == "exact-v1" || checker == "nonconsecutive-permutation-v1" || checker == "cpp-special-v1"
}

// ParseConstructionSize accepts only the advertised decimal integer and ASCII whitespace.
func ParseConstructionSize(input []byte) (int, error) {
	if len(input) > MaxFile {
		return 0, invalid("file_limit")
	}
	s := strings.Trim(string(input), " \t\r\n\v\f")
	if len(s) == 0 || len(s) > 4 {
		return 0, invalid("construction_input_invalid")
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, invalid("construction_input_invalid")
		}
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > 1000 {
		return 0, invalid("construction_input_invalid")
	}
	return n, nil
}

func ValidateCaseInput(m Manifest, input []byte) error {
	if err := ValidateInput(m.Type, input); err != nil {
		return err
	}
	if m.Checker == "nonconsecutive-permutation-v1" {
		_, err := ParseConstructionSize(input)
		return err
	}
	return nil
}

func ValidateCheckerSource(source []byte) error {
	if len(source) == 0 || len(source) > MaxFile || !utf8.Valid(source) || strings.TrimSpace(string(source)) == "" || strings.IndexByte(string(source), 0) >= 0 {
		return invalid("checker_source_invalid")
	}
	return nil
}

func ValidateInput(problemType string, input []byte) error {
	if len(input) > MaxFile {
		return invalid("file_limit")
	}
	if problemType != "interactive" {
		return nil
	}
	if !secretRE.Match(input) {
		return invalid("interactive_secret_invalid")
	}
	n, err := strconv.Atoi(strings.Trim(string(input), " \t\r\n"))
	if err != nil || n < 1 || n > 1000 {
		return invalid("interactive_secret_invalid")
	}
	return nil
}

func ParseManifest(raw []byte) (Manifest, error) {
	var m Manifest
	if err := strictJSON(raw, &m); err != nil {
		return m, err
	}
	return NormalizeManifest(m)
}
