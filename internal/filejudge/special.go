package filejudge

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"oj.local/oj/internal/problemstore"
)

func constructionVerdict(input, actual []byte) (string, string, error) {
	n, err := problemstore.ParseConstructionSize(input)
	if err != nil {
		return "SYSTEM_ERROR", "", err
	}
	tokens := bytes.Fields(actual)
	if n == 2 || n == 3 {
		if len(tokens) == 1 && string(tokens[0]) == "-1" {
			return "AC", "无解情况回答正确。", nil
		}
		return "WA", "n 为 2 或 3 时不存在合法排列，必须只输出 -1。", nil
	}
	if len(tokens) != n {
		return "WA", "输出整数个数应恰好等于 n。", nil
	}
	seen := make([]bool, n+1)
	previous := 0
	for i, token := range tokens {
		// No decimal punctuation, scientific notation or embedded signs.
		for _, digit := range token {
			if digit < '0' || digit > '9' {
				return "WA", "排列必须由十进制整数构成。", nil
			}
		}
		x, err := strconv.Atoi(string(token))
		if err != nil || x < 1 || x > n {
			return "WA", "排列中的整数必须在 1 到 n 之间。", nil
		}
		if seen[x] {
			return "WA", "排列中出现了重复整数。", nil
		}
		if i > 0 && (previous-x == 1 || x-previous == 1) {
			return "WA", "存在相邻数字之差的绝对值为 1。", nil
		}
		seen[x], previous = true, x
	}
	return "AC", "排列满足全部构造条件。", nil
}

func checkerRuntimeSpec(run, stage, artifact, input, answer, output string) stageSpec {
	return stageSpec{RunID: run, Name: stageName(run, stage), Stage: stage, Memory: 64 * 1048576, Pids: 32, Wall: 2 * time.Second,
		Mounts: map[string]string{"/checker/program": artifact, "/data/input.txt": input, "/data/answer.txt": answer, "/data/output.txt": output},
		Entry:  "/checker/program", Args: []string{"/data/input.txt", "/data/answer.txt", "/data/output.txt"}}
}

func checkerVerdict(s stageResult) (string, string) {
	if s.Failure != "" {
		return "SYSTEM_ERROR", s.Failure
	}
	if s.State.OOMKilled {
		return "SYSTEM_ERROR", "checker_memory_limit"
	}
	if s.Cause != "" {
		return "SYSTEM_ERROR", "checker_" + s.Cause
	}
	if s.State.ExitCode == 0 {
		return "AC", ""
	}
	if s.State.ExitCode == 1 {
		return "WA", ""
	}
	return "SYSTEM_ERROR", "checker_invalid_exit"
}

func (e *Engine) runChecker(ctx context.Context, run, dir string, index int, artifact string, input, answer, rawOutput []byte) (string, *CheckerResult, string, string) {
	caseDir := filepath.Join(dir, "checker-case-"+strconv.Itoa(index))
	if err := os.Mkdir(caseDir, 0700); err != nil {
		return "SYSTEM_ERROR", nil, "", "snapshot_write_failed"
	}
	paths := []string{filepath.Join(caseDir, "input.txt"), filepath.Join(caseDir, "answer.txt"), filepath.Join(caseDir, "output.txt")}
	for i, data := range [][]byte{input, answer, rawOutput} {
		if os.WriteFile(paths[i], data, 0644) != nil {
			return "SYSTEM_ERROR", nil, "", "snapshot_write_failed"
		}
	}
	spec := checkerRuntimeSpec(run, "checker-case-"+strconv.Itoa(index), artifact, paths[0], paths[1], paths[2])
	out, stderr := &boundedStream{limit: MaxRawOutput}, &boundedStream{limit: MaxRawOutput}
	stage := e.stage(ctx, spec, out, stderr)
	r := &CheckerResult{DurationMS: stage.DurationMS, OOMKilled: stage.State.OOMKilled}
	if stage.State.StartedAt != "" && !strings.HasPrefix(stage.State.StartedAt, "0001-") {
		r.ExitCode = &stage.State.ExitCode
	}
	var outTrunc, errTrunc bool
	r.Stdout, outTrunc = display(out.data.Bytes(), MaxFile)
	r.Stderr, errTrunc = display(stderr.data.Bytes(), 16384)
	r.OutputTruncated = outTrunc || errTrunc || stage.Cause == "output_limit"
	verdict, failure := checkerVerdict(stage)
	feedback, _ := display(stderr.data.Bytes(), 1024)
	return verdict, r, feedback, failure
}
