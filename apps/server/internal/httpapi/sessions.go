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
