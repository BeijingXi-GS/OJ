package domain

type RunState string

const (
	Pending   RunState = "PENDING"
	Queued    RunState = "QUEUED"
	Compiling RunState = "COMPILING"
	Judging   RunState = "JUDGING"
	Finished  RunState = "FINISHED"
	Cancelled RunState = "CANCELLED"
)

type Verdict string

const (
	Accepted              Verdict = "ACCEPTED"
	WrongAnswer           Verdict = "WRONG_ANSWER"
	CompileError          Verdict = "COMPILE_ERROR"
	RuntimeError          Verdict = "RUNTIME_ERROR"
	TimeLimitExceeded     Verdict = "TIME_LIMIT_EXCEEDED"
	MemoryLimitExceeded   Verdict = "MEMORY_LIMIT_EXCEEDED"
	OutputLimitExceeded   Verdict = "OUTPUT_LIMIT_EXCEEDED"
	IdlenessLimitExceeded Verdict = "IDLENESS_LIMIT_EXCEEDED"
	PartialAccepted       Verdict = "PARTIAL_ACCEPTED"
	SystemError           Verdict = "SYSTEM_ERROR"
	VerdictCancelled      Verdict = "CANCELLED"
)

// CanTransition is draft policy only. A repository must still apply a fenced compare-and-swap.
// Queue redelivery keeps the logical run state; rejudge creates a different run.
func CanTransition(from, to RunState) bool {
	switch from {
	case Pending:
		return to == Queued || to == Finished || to == Cancelled
	case Queued:
		return to == Compiling || to == Finished || to == Cancelled
	case Compiling:
		return to == Judging || to == Finished || to == Cancelled
	case Judging:
		return to == Finished || to == Cancelled
	default:
		return false
	}
}
