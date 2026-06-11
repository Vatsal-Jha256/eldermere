package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/Vatsal-Jha256/eldermere/apps/server/internal/game"
	"github.com/coder/websocket"
)

type commandMessage struct {
	Command string `json:"command"`
}

func handleWebSocket(logger *slog.Logger, world game.World) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			logger.Warn("websocket accept failed", "error", err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "session ended")

		session := game.NewSession(world)
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
			err = writeEvents(ctx, conn, session.Handle(command))
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
