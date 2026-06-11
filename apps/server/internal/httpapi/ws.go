package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Vatsal-Jha256/eldermere/apps/server/internal/game"
	"github.com/Vatsal-Jha256/eldermere/apps/server/internal/storage"
	"github.com/coder/websocket"
)

type commandMessage struct {
	Command string `json:"command"`
}

func handleWebSocket(logger *slog.Logger, world game.World, store storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		playerID := strings.TrimSpace(r.URL.Query().Get("player_id"))
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		if playerID == "" {
			http.Error(w, "player_id is required", http.StatusBadRequest)
			return
		}
		if token == "" {
			http.Error(w, "token is required", http.StatusBadRequest)
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

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			logger.Warn("websocket accept failed", "error", err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "session ended")

		session := game.NewSession(world)
		state, ok, err := store.LoadPlayerState(r.Context(), playerID)
		if err != nil {
			logger.Warn("load player state failed", "player_id", playerID, "error", err)
			conn.Close(websocket.StatusInternalError, "failed to load state")
			return
		}
		if ok {
			session = game.NewSessionFromState(world, state)
		}

		if err := writeEvents(r.Context(), conn, session.Welcome()); err != nil {
			logger.Warn("websocket welcome failed", "error", err)
			return
		}

		for {
			_, payload, err := conn.Read(r.Context())
			if err != nil {
				logger.Info("websocket closed", "error", err)
				return
			}

			command := parseCommand(payload)
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			events := session.Handle(command)
			if saveErr := store.SavePlayerState(ctx, playerID, session.PersistentState()); saveErr != nil {
				cancel()
				logger.Warn("save player state failed", "player_id", playerID, "error", saveErr)
				conn.Close(websocket.StatusInternalError, "failed to save state")
				return
			}
			err = writeEvents(ctx, conn, events)
			cancel()
			if err != nil {
				logger.Warn("websocket write failed", "error", err)
				return
			}
		}
	}
}

func parseCommand(payload []byte) string {
	var message commandMessage
	if err := json.Unmarshal(payload, &message); err == nil && message.Command != "" {
		return message.Command
	}
	return string(payload)
}

func writeEvents(ctx context.Context, conn *websocket.Conn, events []game.Event) error {
	for _, event := range events {
		payload, err := json.Marshal(event)
		if err != nil {
			return err
		}
		if err := conn.Write(ctx, websocket.MessageText, payload); err != nil {
			return err
		}
	}
	return nil
}
