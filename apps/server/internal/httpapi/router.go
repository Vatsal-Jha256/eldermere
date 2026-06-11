package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/Vatsal-Jha256/eldermere/apps/server/internal/config"
	"github.com/Vatsal-Jha256/eldermere/apps/server/internal/game"
	"github.com/Vatsal-Jha256/eldermere/apps/server/internal/storage"
)

type healthResponse struct {
	Status string `json:"status"`
	App    string `json:"app"`
	Env    string `json:"env"`
	Time   string `json:"time"`
}

func NewRouter(cfg config.Config, logger *slog.Logger, store storage.Store) http.Handler {
	mux := http.NewServeMux()
	world := game.NewStarterWorld()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{
			Status: "ok",
			App:    "eldermere",
			Env:    cfg.AppEnv,
			Time:   time.Now().UTC().Format(time.RFC3339),
		})
	})

	mux.HandleFunc("GET /api/v1/status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{
			Status: "ok",
			App:    "eldermere",
			Env:    cfg.AppEnv,
			Time:   time.Now().UTC().Format(time.RFC3339),
		})
	})

	mux.HandleFunc("POST /api/v1/sessions", handleCreateSession(logger, store))
	mux.HandleFunc("GET /ws", handleWebSocket(logger, world, store))

	return withRequestLogging(logger, withCORS(mux))
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withRequestLogging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("request handled", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}
