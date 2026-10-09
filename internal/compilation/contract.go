// Package compilation defines the compile-only protocol shared by API and worker.
package compilation

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"unicode/utf8"
)

const (
	MaxSourceBytes      = 65536
	MaxBodyBytes        = 262144
	MaxDiagnosticsBytes = 65536
)

type Result struct {
	Status               string `json:"status"`
	Diagnostics          string `json:"diagnostics"`
	DiagnosticsTruncated bool   `json:"diagnostics_truncated"`
	DurationMS           int64  `json:"duration_ms"`
	Compiler             string `json:"compiler"`
	Standard             string `json:"standard"`
}

type Error struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string { return e.Code }

func WriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func WriteError(w http.ResponseWriter, err *Error) {
	WriteJSON(w, err.Status, map[string]any{"error": err})
}

func Decode(w http.ResponseWriter, r *http.Request, dst any) *Error {
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		return &Error{400, "invalid_content_type", "请使用 JSON 格式提交。"}
	}
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(dst)
	if err == nil {
		var extra any
		if err = decoder.Decode(&extra); err == io.EOF {
			return nil
		}
	}
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		return &Error{413, "body_too_large", "请求内容超过大小限制。"}
	}
	return &Error{400, "invalid_request", "请求格式或字段无效。"}
}

func ValidateSource(source string) *Error {
	if len(source) > MaxSourceBytes {
		return &Error{413, "source_too_large", "源码不能超过 64 KiB。"}
	}
	if !utf8.ValidString(source) || strings.TrimSpace(source) == "" {
		return &Error{400, "invalid_source", "请输入有效的 C++ 源码。"}
	}
	return nil
}
