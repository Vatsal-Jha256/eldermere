package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Vatsal-Jha256/eldermere/apps/server/internal/storage"
)

type createSessionRequest struct {
	DisplayName string `json:"display_name"`
}

type verifySessionRequest struct {
	PlayerID string `json:"player_id"`
	Token    string `json:"token"`
}

func handleCreateSession(logger *slog.Logger, store storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request createSessionRequest
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "invalid JSON body", http.StatusBadRequest)
				return
			}
		}

		session, err := store.CreatePlayerSession(r.Context(), strings.TrimSpace(request.DisplayName))
		if err != nil {
			logger.Error("create player session failed", "error", err)
			http.Error(w, "failed to create session", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusCreated, session)
	}
}

func handleVerifySession(logger *slog.Logger, store storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request verifySessionRequest
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "invalid JSON body", http.StatusBadRequest)
				return
			}
		}

		playerID := strings.TrimSpace(request.PlayerID)
		token := strings.TrimSpace(request.Token)
		if playerID == "" || token == "" {
			http.Error(w, "player_id and token are required", http.StatusBadRequest)
			return
		}

		verified, err := store.VerifyPlayerSession(r.Context(), playerID, token)
		if err != nil {
			logger.Warn("verify player session failed", "player_id", playerID, "error", err)
			http.Error(w, "failed to verify session", http.StatusInternalServerError)
			return
		}
		if !verified {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}

		displayName, ok, err := store.PlayerDisplayName(r.Context(), playerID)
		if err != nil {
			logger.Warn("load player display name failed", "player_id", playerID, "error", err)
			http.Error(w, "failed to load session", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}

		writeJSON(w, http.StatusOK, storage.PlayerSession{
			PlayerID:    playerID,
			DisplayName: displayName,
		})
	}
}
