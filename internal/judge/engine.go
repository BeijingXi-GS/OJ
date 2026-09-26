package judge

import (
	"context"
	"oj.local/oj/internal/domain"
)

type Engine interface {
	Judge(context.Context, domain.JudgeJob) (domain.JudgeResult, error)
}
