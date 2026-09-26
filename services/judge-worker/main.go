// This entrypoint is deliberately idle until Phase 1/3 implements judging and queue consumption.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	log.Info("worker scaffold started", "mode", "idle", "accepts_jobs", false)
	<-ctx.Done()
	log.Info("worker scaffold stopped")
}
