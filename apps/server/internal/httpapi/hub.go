package httpapi

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Vatsal-Jha256/eldermere/apps/server/internal/game"
	"github.com/coder/websocket"
)

const roomLogLimit = 30

var writeRoomEvents = writeEvents

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

func (h *roomHub) presence(roomID string) game.Event {
	h.mu.RLock()
	defer h.mu.RUnlock()

	names := make([]string, 0, len(h.rooms[roomID]))
	for client := range h.rooms[roomID] {
		names = append(names, client.displayName)
	}
	if len(names) == 0 {
		return game.Event{Type: "presence", Text: "Players here: none."}
	}
	sortStrings(names)
	return game.Event{Type: "presence", Text: fmt.Sprintf("Players here: %s.", strings.Join(names, ", "))}
}

func (h *roomHub) presenceAll() game.Event {
	h.mu.RLock()
	defer h.mu.RUnlock()

	roomIDs := make([]string, 0, len(h.rooms))
	for roomID, clients := range h.rooms {
		if len(clients) > 0 {
			roomIDs = append(roomIDs, roomID)
		}
	}
	if len(roomIDs) == 0 {
		return game.Event{Type: "presence", Text: "Players online: none."}
	}
	sortStrings(roomIDs)

	parts := make([]string, 0, len(roomIDs))
	for _, roomID := range roomIDs {
		names := make([]string, 0, len(h.rooms[roomID]))
		for client := range h.rooms[roomID] {
			names = append(names, client.displayName)
		}
		sortStrings(names)
		parts = append(parts, fmt.Sprintf("%s: %s", roomID, strings.Join(names, ", ")))
	}
	return game.Event{Type: "presence", Text: fmt.Sprintf("Players online by room: %s.", strings.Join(parts, "; "))}
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
	if !shouldRetainHistory(event) {
		return
	}
	events := append(h.history[roomID], event)
	if len(events) > roomLogLimit {
		events = events[len(events)-roomLogLimit:]
	}
	h.history[roomID] = events
}

func shouldRetainHistory(event game.Event) bool {
	return event.Type != "presence"
}

func (h *roomHub) broadcastLocked(ctx context.Context, roomID string, event game.Event, except *clientConn) {
	failed := make([]*clientConn, 0)
	for client := range h.rooms[roomID] {
		if client == except {
			continue
		}
		if err := writeRoomEvents(ctx, client.conn, []game.Event{event}); err != nil {
			if client.conn != nil {
				client.conn.Close(websocket.StatusInternalError, "broadcast failed")
			}
			failed = append(failed, client)
		}
	}
	for _, client := range failed {
		h.removeLocked(client)
	}
}

func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		value := values[i]
		j := i - 1
		for j >= 0 && values[j] > value {
			values[j+1] = values[j]
			j--
		}
		values[j+1] = value
	}
}
