package filejudge

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestInteractiveStdinHelper(t *testing.T) {
	if os.Getenv("OJ_TRUSTED_STDIN_HELPER") != "1" {
		return
	}
	if _, err := io.Copy(io.Discard, os.Stdin); err != nil {
		os.Exit(125)
	}
	os.Exit(0)
}

type faultyProtocolInput struct{ err error }

func (r *faultyProtocolInput) Read([]byte) (int, error) { return 0, r.err }
func (r *faultyProtocolInput) Close() error             { return nil }

func TestDuplexInputErrorsAreNotSuccessfulCompletion(t *testing.T) {
	for _, name := range []string{"normal_eof", "read_failure", "unexpected_closed_pipe", "supervised_close"} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			input := io.NopCloser(strings.NewReader("protocol bytes\n"))
			if name == "read_failure" {
				input = &faultyProtocolInput{err: errors.New("PRIVATE_HOST_READER_ERROR")}
			}
			if name == "unexpected_closed_pipe" {
				input = &faultyProtocolInput{err: io.ErrClosedPipe}
			}
			if name == "supervised_close" {
				reader, writer := io.Pipe()
				defer writer.Close()
				input = reader
				go func() { time.Sleep(25 * time.Millisecond); cancel() }()
			}
			c := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestInteractiveStdinHelper$")
			c.Env = append(os.Environ(), "OJ_TRUSTED_STDIN_HELPER=1")
			err := runDuplexCommand(ctx, c, input, io.Discard, io.Discard)
			var transfer *duplexInputError
			if name == "normal_eof" && err != nil {
				t.Fatal("normal protocol EOF failed", err)
			}
			if (name == "read_failure" || name == "unexpected_closed_pipe") && (!errors.As(err, &transfer) || strings.Contains(err.Error(), "PRIVATE")) {
				t.Fatal("unexpected stdin transfer error was ignored or exposed", err)
			}
			if name == "supervised_close" && errors.As(err, &transfer) {
				t.Fatal("known supervised pipe closure became an I/O failure", err)
			}
		})
	}
}

type heldBudgetWriter struct {
	entered chan struct{}
	release chan struct{}
}

func (w *heldBudgetWriter) Write(p []byte) (int, error) {
	close(w.entered)
	<-w.release
	return len(p), nil
}

func TestFirstBudgetOwnerIsRecordedBeforeRetentionOrCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	budget := &interactionBudget{cancel: cancel}
	budget.total.Store(MaxRawOutput - 1)
	held := &heldBudgetWriter{entered: make(chan struct{}), release: make(chan struct{})}
	participant := &interactionWriter{budget: budget, side: participantSide, store: &boundedStream{limit: MaxRawOutput, file: held}}
	interactor := &interactionWriter{budget: budget, side: interactorSide, store: &boundedStream{limit: MaxRawOutput}}
	done := make(chan struct{})
	go func() { _, _ = participant.Write([]byte("ab")); close(done) }()
	<-held.entered // first crossing has happened; its retention is suspended.
	_, _ = interactor.Write([]byte("x"))
	if budget.owner.Load() != participantSide || budget.total.Load() != MaxRawOutput+2 {
		t.Fatal("second crossing overtook the first owner")
	}
	close(held.release)
	<-done
	if ctx.Err() == nil || budget.owner.Load() != participantSide {
		t.Fatal("first crossing did not cancel or retain owner")
	}
}
