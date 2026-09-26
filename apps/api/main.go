package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"oj.local/oj/internal/httpapi"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	var ping httpapi.PingFunc
	if url := os.Getenv("DATABASE_URL"); url != "" {
		pool, err := pgxpool.New(ctx, url)
		if err != nil {
			log.Error("invalid database configuration")
			os.Exit(1)
		}
		defer pool.Close()
		ping = pool.Ping
	}
	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	server := &http.Server{Addr: addr, Handler: httpapi.New(ping), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	failures := make(chan error, 1)
	go func() {
		log.Info("api listening", "addr", addr, "judge_enabled", false)
		failures <- server.ListenAndServe()
	}()
	select {
	case err := <-failures:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("api stopped", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			log.Error("shutdown failed", "error", err)
			_ = server.Close()
		}
	}
}
