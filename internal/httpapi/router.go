// Package httpapi is the control-plane HTTP adapter. It must never execute user code.
package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type PingFunc func(context.Context) error

func New(ping PingFunc) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recoverer)
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		reply(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if ping == nil {
			reply(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "reason": "database_not_configured"})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := ping(ctx); err != nil {
			reply(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "reason": "database_unavailable"})
			return
		}
		reply(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	r.Get("/api/v1/system", func(w http.ResponseWriter, _ *http.Request) {
		reply(w, http.StatusOK, map[string]any{"name": "OJ", "stage": "foundation", "judge_enabled": false, "contract_version": "0.1.0-draft"})
	})
	return r
}

func reply(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
