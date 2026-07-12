package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/Vatsal-Jha256/eldermere/apps/server/internal/game"
	"github.com/Vatsal-Jha256/eldermere/apps/server/internal/storage"
	"github.com/coder/websocket"
)

type commandMessage struct {
	Command string `json:"command"`
}

const maxCommandLength = 512

func handleWebSocket(logger *slog.Logger, world game.World, store storage.Store, hub *roomHub, allowedOrigins []string) http.HandlerFunc {
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

		conn, err := websocket.Accept(w, r, websocketAcceptOptions(allowedOrigins))
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
		client := &clientConn{
			playerID:    playerID,
			displayName: displayName,
			conn:        conn,
		}
		hub.join(r.Context(), client, session.RoomID())
		defer hub.leave(context.Background(), client)

		if err := writeEvents(r.Context(), conn, session.Welcome()); err != nil {
			logger.Warn("websocket welcome failed", "error", err)
			return
		}
		if err := writeEvents(r.Context(), conn, hub.recent(session.RoomID())); err != nil {
			logger.Warn("websocket recent log failed", "error", err)
			return
		}

		for {
			_, payload, err := conn.Read(r.Context())
			if err != nil {
				logger.Info("websocket closed", "error", err)
				return
			}

			command := parseCommand(payload)
			if commandTooLong(command) {
				if err := writeEvents(r.Context(), conn, []game.Event{{Type: "error", Text: fmt.Sprintf("Command is too long. Keep commands under %d characters.", maxCommandLength)}}); err != nil {
					logger.Warn("websocket command limit write failed", "error", err)
					return
				}
				continue
			}
			if isGlobalPresenceCommand(command) {
				if err := writeEvents(r.Context(), conn, []game.Event{hub.presenceAll()}); err != nil {
					logger.Warn("websocket global presence write failed", "error", err)
					return
				}
				continue
			}
			if isPresenceCommand(command) {
				if err := writeEvents(r.Context(), conn, []game.Event{hub.presence(session.RoomID())}); err != nil {
					logger.Warn("websocket presence write failed", "error", err)
					return
				}
				continue
			}

			beforeRoomID := session.RoomID()
			beforeState := session.PersistentState()
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			events := session.Handle(command)
			afterState := session.PersistentState()
			if persistentStateChanged(beforeState, afterState) {
				if saveErr := store.SavePlayerState(ctx, playerID, afterState); saveErr != nil {
					cancel()
					logger.Warn("save player state failed", "player_id", playerID, "error", saveErr)
					conn.Close(websocket.StatusInternalError, "failed to save state")
					return
				}
			}
			if beforeRoomID != session.RoomID() {
				hub.leave(ctx, client)
				hub.join(ctx, client, session.RoomID())
			}
			broadcastRoomEvents(ctx, hub, client, session.RoomID(), displayName, events)
			err = writeEvents(ctx, conn, events)
			cancel()
			if err != nil {
				logger.Warn("websocket write failed", "error", err)
				return
			}
		}
	}
}

func websocketAcceptOptions(allowedOrigins []string) *websocket.AcceptOptions {
	if originPolicyAllowsAny(allowedOrigins) {
		return &websocket.AcceptOptions{InsecureSkipVerify: true}
	}
	return &websocket.AcceptOptions{OriginPatterns: originPatterns(allowedOrigins)}
}

func originPolicyAllowsAny(allowedOrigins []string) bool {
	if len(allowedOrigins) == 0 {
		return true
	}
	for _, allowed := range allowedOrigins {
		if strings.TrimSpace(allowed) == "*" {
			return true
		}
	}
	return false
}

func originPatterns(allowedOrigins []string) []string {
	patterns := make([]string, 0, len(allowedOrigins))
	for _, allowed := range allowedOrigins {
		allowed = strings.TrimSpace(allowed)
		if allowed == "" || allowed == "*" {
			continue
		}
		parsed, err := url.Parse(allowed)
		if err == nil && parsed.Host != "" {
			patterns = append(patterns, parsed.Host)
			continue
		}
		patterns = append(patterns, allowed)
	}
	return patterns
}

func persistentStateChanged(before game.PersistentState, after game.PersistentState) bool {
	return !reflect.DeepEqual(before, after)
}

func broadcastRoomEvents(ctx context.Context, hub *roomHub, client *clientConn, roomID string, displayName string, events []game.Event) {
	for _, event := range events {
		switch event.Type {
		case "say":
			hub.broadcast(ctx, roomID, game.Event{
				Type: "say",
				Text: fmt.Sprintf("%s says: %s", displayName, quotedSpeech(event.Text)),
			}, client)
		case "fight", "party":
			hub.broadcast(ctx, roomID, game.Event{
				Type: "room_event",
				Text: fmt.Sprintf("%s: %s", displayName, event.Text),
			}, client)
		}
	}
}

func quotedSpeech(text string) string {
	text = strings.TrimPrefix(text, "You say, ")
	return strings.Trim(text, `"`)
}

func parseCommand(payload []byte) string {
	var message commandMessage
	if err := json.Unmarshal(payload, &message); err == nil && message.Command != "" {
		return strings.TrimSpace(message.Command)
	}
	return strings.TrimSpace(string(payload))
}

func commandTooLong(command string) bool {
	return len([]rune(command)) > maxCommandLength
}

func isGlobalPresenceCommand(command string) bool {
	switch strings.ToLower(strings.TrimSpace(command)) {
	case "who all", "players all", "presence all":
		return true
	default:
		return false
	}
}

func isPresenceCommand(command string) bool {
	switch strings.ToLower(strings.TrimSpace(command)) {
	case "who", "players", "presence":
		return true
	default:
		return false
	}
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
