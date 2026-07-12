package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
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
	world := loadWorld(logger, cfg)
	hub := newRoomHub()

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
	mux.HandleFunc("GET /ws", handleWebSocket(logger, world, store, hub, cfg.AllowedOrigins))

	return withRequestLogging(logger, withCORS(mux, cfg.AllowedOrigins))
}

func loadWorld(logger *slog.Logger, cfg config.Config) game.World {
	world := game.NewStarterWorld()
	candidates := []string{}
	if cfg.ContentPacksDir != "" {
		candidates = append(candidates, cfg.ContentPacksDir)
	}
	candidates = append(candidates, "content-packs", "../../content-packs", "/app/content-packs")

	for _, candidate := range candidates {
		content, err := game.LoadPackRuntimeContentFromContentPacks(candidate)
		if err != nil {
			logger.Debug("content packs not loaded", "path", candidate, "error", err)
			continue
		}
		withPacks, err := world.WithPackRuntimeContent(content)
		if err != nil {
			logger.Warn("content packs failed runtime validation", "path", candidate, "error", err)
			return world
		}
		logger.Info("content packs loaded", "path", candidate, "rooms", len(content.Rooms), "arcs", len(content.Stories.Arcs), "seed_tags", len(content.Stories.Tags))
		return withPacks
	}

	logger.Warn("no content pack directory found")
	return world
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func withCORS(next http.Handler, allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := allowedCORSOrigin(r.Header.Get("Origin"), allowedOrigins); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func allowedCORSOrigin(requestOrigin string, allowedOrigins []string) string {
	if len(allowedOrigins) == 0 {
		return "*"
	}
	for _, allowed := range allowedOrigins {
		allowed = strings.TrimSpace(allowed)
		if allowed == "*" {
			return "*"
		}
		if requestOrigin != "" && strings.EqualFold(allowed, requestOrigin) {
			return requestOrigin
		}
	}
	return ""
}

func withRequestLogging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("request handled", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}
