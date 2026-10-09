package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"oj.local/oj/internal/compilerworker"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	compiler, err := exec.LookPath("g++")
	if err != nil {
		log.Error("g++ not installed")
		os.Exit(1)
	}
	probe, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	version, err := exec.CommandContext(probe, compiler, "-dumpfullversion").Output()
	cancel()
	if err != nil {
		log.Error("g++ version check failed")
		os.Exit(1)
	}
	name := "GCC " + strings.TrimSpace(string(version))
	runner := compilerworker.New(compiler, name)
	addr := os.Getenv("COMPILER_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8081"
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	server := &http.Server{Addr: addr, Handler: runner.Handler(), BaseContext: nil, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 30 * time.Second}
	failures := make(chan error, 1)
	go func() {
		log.Info("compiler listening", "addr", addr, "compiler", name, "executes_programs", false)
		failures <- server.ListenAndServe()
	}()
	select {
	case err := <-failures:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("compiler stopped", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 18*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			_ = server.Close()
		}
	}
}
