// Package sandbox defines the isolation boundary. No insecure fallback is provided.
package sandbox

import (
	"context"
	"oj.local/oj/internal/domain"
	"oj.local/oj/internal/execution"
)

type ProcessResult struct {
	ProcessID         string
	ExitCode          int
	Signal            string
	CPUTimeMS         int64
	PeakMemoryBytes   int64
	StdoutArtifactKey string
	StderrArtifactKey string
}

type ExecutionResult struct {
	Processes                []ProcessResult
	AggregatePeakMemoryBytes int64
	Failure                  *domain.Failure
}

// Runtime owns graph-wide cleanup and aggregate accounting. Implementations must
// enforce compile isolation separately and return only after all child processes exit.
type Runtime interface {
	Run(context.Context, execution.Graph) (ExecutionResult, error)
}
