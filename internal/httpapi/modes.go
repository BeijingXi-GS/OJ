package httpapi

import (
	"errors"
	"net/http"

	"oj.local/oj/internal/compilation"
	"oj.local/oj/internal/filejudge"
	"oj.local/oj/internal/problemstore"
)

func modeOf(r *http.Request) string {
	mode := r.Header.Get("X-OJ-Mode")
	if mode == "" {
		return "student"
	}
	return mode
}
func validModeHeader(r *http.Request) bool {
	if len(r.Header.Values("X-OJ-Mode")) > 1 {
		return false
	}
	switch modeOf(r) {
	case "student", "teacher", "engineer":
		return true
	}
	return false
}
func forbidden(w http.ResponseWriter) {
	compilation.WriteError(w, &compilation.Error{Status: 403, Code: "mode_forbidden", Message: "当前模式不可使用此功能。"})
}

// Construct the public result from an allowlist. A student response never contains
// the hidden case slice, even if the client asks for fields or ignores its UI.
func projectResult(result filejudge.Result, mode, problemType string) map[string]any {
	passed := 0
	for _, c := range result.Cases {
		if c.Verdict == "AC" {
			passed++
		}
	}
	message := result.Message
	if mode != "engineer" {
		switch result.Verdict {
		case "SYSTEM_ERROR":
			message = "判题服务暂时无法完成本次测试，请稍后重试。"
		case "TLE":
			message = "程序运行超过时间限制。"
		case "MLE":
			message = "程序使用的内存超过限制。"
		case "WA":
			message = "答案未通过测试，请检查解题逻辑和输出格式。"
		}
	}
	public := map[string]any{"mode": mode, "problem_type": problemType, "verdict": result.Verdict, "message": message, "duration_ms": result.DurationMS, "compiler": result.Compiler, "standard": result.Standard, "compile": result.Compile, "passed_cases": passed, "total_cases": len(result.Cases)}
	if problemType == "communication" {
		if mode == "engineer" {
			public["role_compiles"] = result.RoleCompiles
		} else {
			roles := make([]map[string]any, len(result.RoleCompiles))
			for i, record := range result.RoleCompiles {
				roles[i] = map[string]any{"role": record.Role, "compile": record.Compile}
			}
			public["role_compiles"] = roles
		}
	}
	if mode == "student" {
		return public
	}
	if mode == "engineer" {
		public["run_id"] = result.RunID
		public["source_sha256"] = result.SourceSHA256
		public["testdata_sha256"] = result.TestdataSHA256
		public["artifact_sha256"] = result.ArtifactSHA256
		public["runtime_image"] = result.RuntimeImage
		public["cases"] = result.Cases
		if result.InteractorCompile != nil {
			public["interactor_compile"] = result.InteractorCompile
			public["interactor_source_sha256"] = result.InteractorSourceSHA256
			public["interactor_artifact_sha256"] = result.InteractorArtifactSHA256
		}
		if result.CheckerSourceSHA256 != "" {
			public["checker_source_sha256"] = result.CheckerSourceSHA256
		}
		if result.CheckerCompile != nil {
			public["checker_compile"] = result.CheckerCompile
			if result.CheckerArtifactSHA256 != "" {
				public["checker_artifact_sha256"] = result.CheckerArtifactSHA256
			}
		}
		return public
	}
	cases := make([]map[string]any, len(result.Cases))
	for n, c := range result.Cases {
		row := map[string]any{"id": c.ID, "verdict": c.Verdict, "duration_ms": c.DurationMS, "input": c.Input, "expected": c.Expected, "input_truncated": c.InputTruncated, "expected_truncated": c.ExpectedTruncated, "stdout": c.Stdout, "stderr": c.Stderr, "output_truncated": c.OutputTruncated}
		if c.StartTimedOut {
			row["start_timed_out"] = true
		}
		if i := c.Interactor; i != nil {
			row["interactor"] = map[string]any{"stdout": i.Stdout, "stderr": i.Stderr, "output_truncated": i.OutputTruncated}
		}
		if c.CheckerFeedback != "" {
			row["checker_feedback"] = c.CheckerFeedback
		}
		if comm := c.Communication; comm != nil {
			row["communication"] = map[string]any{"bytes_observed": comm.BytesObserved, "bytes_transferred": comm.BytesTransferred, "byte_limit": comm.ByteLimit, "limit_exceeded": comm.LimitExceeded, "sender_stdout": comm.Sender.Stdout, "sender_stderr": comm.Sender.Stderr, "receiver_stderr": comm.Receiver.Stderr, "sender_output_truncated": comm.Sender.OutputTruncated, "receiver_output_truncated": comm.Receiver.OutputTruncated}
		}
		cases[n] = row
	}
	public["cases"] = cases
	return public
}

func uploadProblem(catalog *problemstore.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if modeOf(r) == "student" {
			forbidden(w)
			return
		}
		// Modes remain switchable. This local origin restriction is not authentication.
		origin := r.Header.Get("Origin")
		if len(r.Header.Values("Origin")) > 1 {
			forbidden(w)
			return
		}
		switch origin {
		case "", "http://127.0.0.1:5173", "http://localhost:5173", "http://127.0.0.1:8080", "http://localhost:8080":
		default:
			forbidden(w)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)
		reader, err := r.MultipartReader()
		if err != nil {
			compilation.WriteError(w, &compilation.Error{Status: 400, Code: "invalid_upload", Message: "请使用题目描述和输入输出文件上传题目。"})
			return
		}
		upload, err := problemstore.ParseUpload(reader)
		if err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				compilation.WriteError(w, &compilation.Error{Status: 413, Code: "upload_too_large", Message: "上传总大小不能超过 1 MiB。"})
				return
			}
			compilation.WriteError(w, &compilation.Error{Status: 400, Code: "invalid_upload", Message: "题目描述或数据文件不符合要求，请检查文件配对、大小和内容。"})
			return
		}
		definition, err := catalog.Create(upload)
		if err != nil {
			if errors.Is(err, problemstore.ErrInvalidUpload) {
				compilation.WriteError(w, &compilation.Error{Status: 400, Code: "invalid_upload", Message: "题目描述或数据文件不符合要求。"})
			} else {
				compilation.WriteError(w, storeError())
			}
			return
		}
		reply(w, 201, map[string]any{"id": definition.ID, "problem": publicProblem(definition)})
	}
}

func debugHandler(catalog *problemstore.Store, compiler compilation.Compiler, judge filejudge.Judge) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if modeOf(r) != "engineer" {
			forbidden(w)
			return
		}
		definitions, err := catalog.List()
		if err != nil {
			compilation.WriteError(w, storeError())
			return
		}
		meta := map[string]any{"available": false}
		if provider, ok := judge.(filejudge.MetadataProvider); ok {
			m, failure := provider.Metadata(r.Context(), "a-plus-b")
			if failure == nil {
				meta = map[string]any{"available": true, "problem_id": m.ProblemID, "case_count": m.CaseCount, "wall_time_ms": m.WallTimeMS, "memory_limit_bytes": m.MemoryLimitBytes, "checker": m.Checker, "problem_type": m.ProblemType}
			} else {
				meta["error_code"] = failure.Code
			}
		}
		reply(w, 200, map[string]any{"mode": "engineer", "local_demo": true, "compiler_configured": compiler != nil, "judge_configured": judge != nil, "problem_count": len(definitions), "judge_metadata": meta, "message": "配置状态不代表编译或运行已经通过；此检查仅读取判题元数据，不执行代码。"})
	}
}
