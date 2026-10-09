package main

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

func TestOwnedPipeCancelsOnCommandEOFAndOversize(t *testing.T) {
	for _, input := range []string{"shutdown\n", "", strings.Repeat("x", 1024)} {
		ctx, cancel := context.WithCancel(context.Background())
		watchOwnedShutdown(strings.NewReader(input), cancel, true)
		if ctx.Err() == nil {
			t.Fatal("owned pipe did not cancel")
		}
		cancel()
	}
}
func TestUnownedPipeDoesNotReadOrCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	watchOwnedShutdown(failReader{t}, cancel, false)
	if ctx.Err() != nil {
		t.Fatal("unowned process cancelled")
	}
}

type failReader struct{ t *testing.T }

func (r failReader) Read([]byte) (int, error) { r.t.Fatal("unowned stdin read"); return 0, io.EOF }
func TestUnknownCommandDoesNotCancelBeforePipeCloses(t *testing.T) {
	r, w := io.Pipe()
	defer r.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { watchOwnedShutdown(r, cancel, true); close(done) }()
	if _, err := w.Write([]byte("noop\n")); err != nil {
		t.Fatal(err)
	}
	select {
	case <-ctx.Done():
		t.Fatal("unknown command cancelled")
	default:
	}
	if _, err := w.Write([]byte("shutdown\n")); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("shutdown command ignored")
	}
	_ = w.Close()
}
