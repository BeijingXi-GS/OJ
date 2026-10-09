// Package httpapi is the control-plane HTTP adapter. It must never execute user code.
package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"oj.local/oj/internal/compilation"
	"oj.local/oj/internal/filejudge"
	"oj.local/oj/internal/problemstore"
)

type PingFunc func(context.Context) error

func New(ping PingFunc) http.Handler { return NewWithCompiler(ping, nil) }
func NewWithCompiler(ping PingFunc, compiler compilation.Compiler) http.Handler {
	return NewWithServices(ping, compiler, nil)
}
func NewWithServices(ping PingFunc, compiler compilation.Compiler, judge filejudge.Judge) http.Handler {
	return NewWithCatalog(ping, compiler, judge, &problemstore.Store{BuiltinRoot: "testdata/problems"})
}

func NewWithCatalog(ping PingFunc, compiler compilation.Compiler, judge filejudge.Judge, catalog *problemstore.Store) http.Handler {
	if catalog == nil {
		catalog = &problemstore.Store{}
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recoverer)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") && !validModeHeader(r) {
				compilation.WriteError(w, &compilation.Error{Status: 400, Code: "invalid_mode", Message: "模式仅支持学生、教师和工程师。"})
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { reply(w, 200, map[string]string{"status": "ok"}) })
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if ping == nil {
			reply(w, 503, map[string]string{"status": "not_ready", "reason": "database_not_configured"})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if ping(ctx) != nil {
			reply(w, 503, map[string]string{"status": "not_ready", "reason": "database_unavailable"})
			return
		}
		reply(w, 200, map[string]string{"status": "ready"})
	})
	r.Get("/api/v1/system", func(w http.ResponseWriter, _ *http.Request) {
		stage := "compile_mvp"
		if judge != nil {
			stage = "judge_mvp"
		}
		reply(w, 200, map[string]any{"name": "OJ", "stage": stage, "judge_enabled": judge != nil, "compile_enabled": compiler != nil, "contract_version": "0.1.0-draft", "compile_contract_version": "compile-mvp/0.1.0", "judge_contract_version": "judge-mvp/0.1.0", "modes_contract_version": "modes-interactive/0.1.0", "communication_contract_version": filejudge.CommunicationContractVersion, "local_demo": true})
	})
	r.Get("/api/v1/problems", func(w http.ResponseWriter, _ *http.Request) {
		definitions, err := catalog.List()
		if err != nil {
			compilation.WriteError(w, storeError())
			return
		}
		list := make([]problemSummary, len(definitions))
		for i, p := range definitions {
			list[i] = problemSummary{ID: p.ID, Title: p.Title, Difficulty: p.Difficulty, Summary: p.Summary, Type: p.Type}
		}
		reply(w, 200, map[string]any{"problems": list})
	})
	r.Get("/api/v1/problems/{id}", func(w http.ResponseWriter, r *http.Request) {
		definition, exists, err := catalog.Get(chi.URLParam(r, "id"))
		if err != nil {
			compilation.WriteError(w, storeError())
			return
		}
		if !exists {
			compilation.WriteError(w, notFound())
			return
		}
		p := publicProblem(definition)
		p.Judge = filejudge.Metadata{ProblemID: p.ID, Enabled: false, ProblemType: p.Type}
		if provider, ok := judge.(filejudge.MetadataProvider); ok {
			meta, err := provider.Metadata(r.Context(), p.ID)
			if err != nil {
				compilation.WriteError(w, err)
				return
			}
			if meta.ProblemType == "" {
				meta.ProblemType = "batch"
			}
			if meta.ProblemID != p.ID || meta.ProblemType != p.Type {
				compilation.WriteError(w, &compilation.Error{Status: 502, Code: "judge_protocol_error", Message: "判题服务返回无效结果。"})
				return
			}
			p.Judge = meta
			if p.Type == "communication" {
				snapshot, failure := filejudge.LoadWithStore(catalog, p.ID)
				if failure != nil || !filejudge.CommunicationMetadataMatchesSnapshot(meta, snapshot, p.ID) {
					compilation.WriteError(w, &compilation.Error{Status: 502, Code: "judge_protocol_error", Message: "判题服务返回无效结果。"})
					return
				}
			}
		} else if judge != nil {
			compilation.WriteError(w, &compilation.Error{Status: 503, Code: "judge_unavailable", Message: "判题服务暂时不可用，请稍后重试。"})
			return
		}
		if modeOf(r) == "engineer" {
			p.EngineeringNotes = definition.EngineeringNotes
			if p.Type == "communication" {
				p.EngineeringNotes += " 双角色分别编译后运行；发送端 stdout 原始字节单向传到接收端 stdin，接收端 stdout 按 token 比较。通信量是代理成功写入受控管道的字节，不代表程序实际消费。每例共享墙钟预算，含容器启动开销；聚合内存上限是两个角色限额之和，未采集实际峰值。"
			} else if judge != nil {
				p.EngineeringNotes += " 编译按钮仅检查编译与链接；运行测试会在独立受限容器中执行。时间为墙钟时间，含容器启动开销。"
			} else {
				p.EngineeringNotes += " 当前未配置判题服务，仅检查编译与链接。"
			}
		}
		reply(w, 200, p)
	})
	r.Post("/api/v1/problems", uploadProblem(catalog))
	r.Get("/api/v1/debug", debugHandler(catalog, compiler, judge))
	r.Post("/api/v1/compilations", func(w http.ResponseWriter, r *http.Request) {
		var request filejudge.Request
		if err := filejudge.DecodeRequest(w, r, &request); err != nil {
			compilation.WriteError(w, err)
			return
		}
		if err := filejudge.ValidateRequestWithStore(request, catalog); err != nil {
			compilation.WriteError(w, err)
			return
		}
		if compiler == nil {
			compilation.WriteError(w, &compilation.Error{Status: 503, Code: "compiler_unavailable", Message: "编译服务尚未启动。"})
			return
		}
		result, err := compiler.Compile(r.Context(), request.Source)
		if err != nil {
			compilation.WriteError(w, err)
			return
		}
		reply(w, 200, result)
	})
	judgeHandler := func(communication bool) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var request filejudge.Request
			decode := filejudge.DecodeRequest
			if communication {
				decode = filejudge.DecodeCommunicationRequest
			}
			if err := decode(w, r, &request); err != nil {
				compilation.WriteError(w, err)
				return
			}
			if err := filejudge.ValidateRequestWithStore(request, catalog); err != nil {
				compilation.WriteError(w, err)
				return
			}
			definition, exists, storeErr := catalog.Get(request.ProblemID)
			if storeErr != nil {
				compilation.WriteError(w, storeError())
				return
			}
			if !exists {
				compilation.WriteError(w, notFound())
				return
			}
			if communication != (definition.Type == "communication") {
				compilation.WriteError(w, &compilation.Error{Status: 400, Code: "invalid_request", Message: "请选择与题型一致的提交方式。"})
				return
			}
			if judge == nil {
				compilation.WriteError(w, &compilation.Error{Status: 503, Code: "judge_unavailable", Message: "判题服务尚未启动。"})
				return
			}
			snapshot, snapshotErr := filejudge.LoadWithStore(catalog, request.ProblemID)
			if snapshotErr != nil {
				compilation.WriteError(w, &compilation.Error{Status: 503, Code: "problem_data_unavailable", Message: "题目数据暂时不可用，请联系教师。"})
				return
			}
			result, err := judge.Judge(r.Context(), request)
			if err != nil {
				compilation.WriteError(w, err)
				return
			}
			if !filejudge.MatchesSnapshot(result, snapshot) || (communication && !filejudge.MatchesRequest(result, request)) {
				compilation.WriteError(w, &compilation.Error{Status: 502, Code: "judge_protocol_error", Message: "判题服务返回无效结果。"})
				return
			}
			reply(w, 200, projectResult(result, modeOf(r), definition.Type))
		}
	}
	r.Post("/api/v1/judgements", judgeHandler(false))
	r.Post("/api/v1/communication-judgements", judgeHandler(true))
	return r
}

func notFound() *compilation.Error {
	return &compilation.Error{Status: 404, Code: "problem_not_found", Message: "题目不存在。"}
}
func storeError() *compilation.Error {
	return &compilation.Error{Status: 503, Code: "problem_store_unavailable", Message: "题库暂时不可用。"}
}
func reply(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
