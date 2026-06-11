package httpapi

import (
	"context"
	"fmt"
	"sync"

	"github.com/Vatsal-Jha256/eldermere/apps/server/internal/game"
	"github.com/coder/websocket"
)

const roomLogLimit = 30

type roomHub struct {
	mu      sync.RWMutex
	rooms   map[string]map[*clientConn]bool
	history map[string][]game.Event
}

type clientConn struct {
	playerID    string
	displayName string
	conn        *websocket.Conn
	roomID      string
}

func newRoomHub() *roomHub {
	return &roomHub{
		rooms:   map[string]map[*clientConn]bool{},
		history: map[string][]game.Event{},
	}
}

func (h *roomHub) join(ctx context.Context, client *clientConn, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client.roomID != "" {
		h.removeLocked(client)
	}
	client.roomID = roomID
	if h.rooms[roomID] == nil {
		h.rooms[roomID] = map[*clientConn]bool{}
	}
	h.rooms[roomID][client] = true

	joined := game.Event{Type: "presence", Text: fmt.Sprintf("%s enters the room.", client.displayName)}
	h.appendHistoryLocked(roomID, joined)
	h.broadcastLocked(ctx, roomID, joined, client)
}

func (h *roomHub) leave(ctx context.Context, client *clientConn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client.roomID == "" {
		return
	}

	roomID := client.roomID
	h.removeLocked(client)
	left := game.Event{Type: "presence", Text: fmt.Sprintf("%s leaves the room.", client.displayName)}
	h.appendHistoryLocked(roomID, left)
	h.broadcastLocked(ctx, roomID, left, client)
}

func (h *roomHub) recent(roomID string) []game.Event {
	h.mu.RLock()
	defer h.mu.RUnlock()

	events := h.history[roomID]
	copied := make([]game.Event, len(events))
	copy(copied, events)
	return copied
}

func (h *roomHub) broadcast(ctx context.Context, roomID string, event game.Event, except *clientConn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.appendHistoryLocked(roomID, event)
	h.broadcastLocked(ctx, roomID, event, except)
}

func (h *roomHub) removeLocked(client *clientConn) {
	clients := h.rooms[client.roomID]
	if clients != nil {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.rooms, client.roomID)
		}
	}
	client.roomID = ""
}

func (h *roomHub) appendHistoryLocked(roomID string, event game.Event) {
	events := append(h.history[roomID], event)
	if len(events) > roomLogLimit {
		events = events[len(events)-roomLogLimit:]
	}
	h.history[roomID] = events
}

func (h *roomHub) broadcastLocked(ctx context.Context, roomID string, event game.Event, except *clientConn) {
	for client := range h.rooms[roomID] {
		if client == except {
			continue
		}
		if err := writeEvents(ctx, client.conn, []game.Event{event}); err != nil {
			client.conn.Close(websocket.StatusInternalError, "broadcast failed")
		}
	}
}
