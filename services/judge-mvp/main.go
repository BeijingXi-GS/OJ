package main

import (
	"context"
	"crypto/subtle"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"oj.local/oj/internal/filejudge"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancelLifecycle := context.WithCancel(ctx)
	defer cancelLifecycle()
	go watchOwnedShutdown(os.Stdin, cancelLifecycle, os.Getenv("JUDGE_CONTROL_TOKEN") != "")
	e, err := filejudge.NewEngine(".local/judge-profile.json", "testdata/problems", ".local/judge-jobs")
	if err != nil {
		log.Error("judge refused startup", "reason", err.Error())
		os.Exit(1)
	}
	preflight, cancel := context.WithTimeout(ctx, 70*time.Second)
	log.Info("judge preflight started")
	err = e.Preflight(preflight)
	cancel()
	if err != nil {
		if ctx.Err() != nil {
			if e.CleanupFailed() {
				log.Error("judge startup cancellation cleanup incomplete")
				os.Exit(1)
			}
			log.Info("judge startup cancelled; supervisor must confirm cleanup")
			return
		}
		log.Error("judge preflight failed", "reason", err.Error())
		os.Exit(1)
	}
	if ctx.Err() != nil {
		return
	}
	shutdownRequested := make(chan struct{}, 1)
	mux := http.NewServeMux()
	mux.Handle("/", e.Handler())
	mux.HandleFunc("POST /shutdown", func(w http.ResponseWriter, r *http.Request) {
		token := os.Getenv("JUDGE_CONTROL_TOKEN")
		if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(r.Header.Get("X-Judge-Control"))) != 1 {
			http.Error(w, "forbidden", 403)
			return
		}
		w.WriteHeader(202)
		select {
		case shutdownRequested <- struct{}{}:
		default:
		}
	})
	server := &http.Server{Addr: "127.0.0.1:8082", Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 120 * time.Second, IdleTimeout: 60 * time.Second}
	failed := make(chan error, 1)
	go func() {
		log.Info("judge ready", "addr", server.Addr, "profile", "judge-mvp/0.1.0")
		failed <- server.ListenAndServe()
	}()
	select {
	case err := <-failed:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("judge stopped")
			os.Exit(1)
		}
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		if server.Shutdown(shutdown) != nil {
			log.Error("judge shutdown incomplete")
			_ = server.Close()
			os.Exit(1)
		}
	case <-shutdownRequested:
		shutdown, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		if server.Shutdown(shutdown) != nil {
			log.Error("judge shutdown incomplete")
			_ = server.Close()
			os.Exit(1)
		}
	}
}
