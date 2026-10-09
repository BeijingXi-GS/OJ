package compilation

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Compiler interface {
	Compile(context.Context, string) (Result, *Error)
}

type Client struct {
	endpoint string
	http     *http.Client
}

func NewClient(base string) (*Client, error) {
	u, err := url.Parse(base)
	if err != nil || u == nil || u.Scheme != "http" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return nil, &Error{503, "invalid_compiler_configuration", "编译服务配置无效。"}
	}
	return &Client{endpoint: strings.TrimRight(base, "/") + "/compile", http: &http.Client{
		Timeout:       20 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}, nil
}

func (c *Client) Compile(ctx context.Context, source string) (Result, *Error) {
	data, _ := json.Marshal(map[string]string{"source": source})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(data))
	if err != nil {
		return Result{}, unavailable()
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return Result{}, unavailable()
	}
	defer res.Body.Close()
	if res.StatusCode == 429 {
		return Result{}, &Error{429, "compiler_busy", "编译服务正在处理其他请求，请稍后重试。"}
	}
	if res.StatusCode == 503 {
		return Result{}, unavailable()
	}
	if res.StatusCode != 200 {
		return Result{}, invalidResponse()
	}
	data, err = io.ReadAll(io.LimitReader(res.Body, 8*MaxDiagnosticsBytes+1))
	if err != nil || len(data) > 8*MaxDiagnosticsBytes {
		return Result{}, invalidResponse()
	}
	var result Result
	if err := json.Unmarshal(data, &result); err != nil {
		return Result{}, invalidResponse()
	}
	if (result.Status != "success" && result.Status != "compile_error" && result.Status != "timeout") || len(result.Diagnostics) > MaxDiagnosticsBytes || result.DurationMS < 0 || result.Standard != "c++20" || result.Compiler == "" || len(result.Compiler) > 128 {
		return Result{}, invalidResponse()
	}
	return result, nil
}

func unavailable() *Error {
	return &Error{503, "compiler_unavailable", "编译服务暂时不可用，请稍后重试。"}
}
func invalidResponse() *Error {
	return &Error{502, "compiler_failure", "编译服务处理失败，请稍后重试。"}
}
