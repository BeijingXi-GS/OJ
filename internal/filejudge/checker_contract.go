package filejudge

import (
	"bytes"
	"encoding/json"
	"io"
	"unicode/utf8"
)

// Decode with a token walk before the typed decoder: encoding/json otherwise
// accepts the last of multiple values for the same key, including nested keys.
func uniqueJSON(data []byte) bool {
	if !utf8.Valid(data) {
		return false
	}
	d := json.NewDecoder(bytes.NewReader(data))
	var value func() bool
	value = func() bool {
		token, err := d.Token()
		if err != nil {
			return false
		}
		if delim, ok := token.(json.Delim); ok {
			switch delim {
			case '{':
				seen := map[string]bool{}
				for d.More() {
					key, err := d.Token()
					s, ok := key.(string)
					if err != nil || !ok || !canonicalJSONKey(s) || seen[s] || !value() {
						return false
					}
					seen[s] = true
				}
				end, err := d.Token()
				return err == nil && end == json.Delim('}')
			case '[':
				for d.More() {
					if !value() {
						return false
					}
				}
				end, err := d.Token()
				return err == nil && end == json.Delim(']')
			default:
				return false
			}
		}
		return true
	}
	if !value() {
		return false
	}
	_, err := d.Token()
	return err == io.EOF
}

// Fixed protocol keys are ASCII. encoding/json otherwise also accepts Unicode
// case-fold aliases (for example long-s in source/sender).
func canonicalJSONKey(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_') {
			return false
		}
	}
	return true
}

func validCheckerResult(r Result) bool {
	custom := r.CheckerCompile != nil
	if custom && len(r.Cases) > 4 {
		return false
	}
	if r.CheckerSourceSHA256 != "" && !digestRE.MatchString(r.CheckerSourceSHA256) {
		return false
	}
	if custom {
		cc := r.CheckerCompile
		if r.InteractorCompile != nil || r.Compile.Status != "success" || !digestRE.MatchString(r.CheckerSourceSHA256) || cc.DurationMS < 0 || len(cc.Diagnostics) > MaxFile || !utf8.ValidString(cc.Diagnostics) {
			return false
		}
		switch cc.Status {
		case "success":
			if !digestRE.MatchString(r.CheckerArtifactSHA256) {
				return false
			}
		case "compile_error", "system_error":
			if r.Verdict != "SYSTEM_ERROR" || len(r.Cases) != 0 || r.CheckerArtifactSHA256 != "" {
				return false
			}
		default:
			return false
		}
	} else if r.CheckerArtifactSHA256 != "" || (r.CheckerSourceSHA256 != "" && (len(r.Cases) != 0 || (r.Verdict != "CE" && r.Verdict != "SYSTEM_ERROR"))) {
		return false
	}
	for _, c := range r.Cases {
		if len(c.CheckerFeedback) > 1024 || !utf8.ValidString(c.CheckerFeedback) {
			return false
		}
		ch := c.Checker
		if ch == nil {
			if custom && (c.Verdict == "AC" || c.Verdict == "WA") {
				return false
			}
			continue
		}
		if !custom || r.CheckerCompile.Status != "success" || c.Interactor != nil || ch.DurationMS < 0 || len(ch.Stdout) > MaxFile || len(ch.Stderr) > 16384 || !utf8.ValidString(ch.Stdout+ch.Stderr) || (ch.ExitCode != nil && (*ch.ExitCode < 0 || *ch.ExitCode > 255)) {
			return false
		}
		if c.ExitCode == nil || *c.ExitCode != 0 || c.OOMKilled {
			return false
		}
		switch c.Verdict {
		case "AC":
			if ch.ExitCode == nil || *ch.ExitCode != 0 || ch.OOMKilled {
				return false
			}
		case "WA":
			if ch.ExitCode == nil || *ch.ExitCode != 1 || ch.OOMKilled {
				return false
			}
		case "SYSTEM_ERROR":
		default:
			return false
		}
	}
	return true
}

// Bind optional fields to the trusted per-request snapshot, not just "batch".
// A changed server-side manifest/source while a request runs fails closed.
func MatchesSnapshot(r Result, snap Snapshot) bool {
	if snap.Manifest.Type == "communication" {
		return matchesCommunicationSnapshot(r, snap)
	}
	if !MatchesProblemType(r, snap.Manifest.Type) {
		return false
	}
	if r.TestdataSHA256 != snap.SHA256 {
		return r.Verdict == "SYSTEM_ERROR" && r.TestdataSHA256 == "" && len(r.Cases) == 0 && r.CheckerCompile == nil && r.CheckerSourceSHA256 == "" && r.CheckerArtifactSHA256 == ""
	}
	if len(r.Cases) > len(snap.Cases) {
		return false
	}
	for i, c := range r.Cases {
		if c.ID != snap.Cases[i].ID {
			return false
		}
	}
	if r.Compile.Status != "success" && len(r.Cases) != 0 {
		return false
	}
	if r.Compile.Status == "success" && r.Verdict != "SYSTEM_ERROR" && len(r.Cases) != len(snap.Cases) {
		return false
	}
	if snap.Manifest.Checker == "cpp-special-v1" {
		if r.CheckerSourceSHA256 != hash([]byte(snap.CheckerSource)) {
			return false
		}
		if r.Compile.Status == "success" && r.Verdict != "SYSTEM_ERROR" && (r.CheckerCompile == nil || r.CheckerCompile.Status != "success") {
			return false
		}
	} else {
		if r.CheckerCompile != nil || r.CheckerSourceSHA256 != "" || r.CheckerArtifactSHA256 != "" {
			return false
		}
		for _, c := range r.Cases {
			if c.Checker != nil || (c.CheckerFeedback != "" && snap.Manifest.Checker != "nonconsecutive-permutation-v1") {
				return false
			}
		}
	}
	return true
}
