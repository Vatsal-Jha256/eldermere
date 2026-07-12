package httpapi

import (
	"context"
	"errors"
	"testing"

	"github.com/Vatsal-Jha256/eldermere/apps/server/internal/game"
	"github.com/coder/websocket"
)

func TestRoomHubPresenceTracksJoinedAndLeftClients(t *testing.T) {
	restoreRoomWriter(t, func(context.Context, *websocket.Conn, []game.Event) error {
		return nil
	})

	hub := newRoomHub()
	arthur := &clientConn{displayName: "Arthur"}
	guinevere := &clientConn{displayName: "Guinevere"}

	hub.join(context.Background(), guinevere, "camelot")
	hub.join(context.Background(), arthur, "camelot")

	presence := hub.presence("camelot")
	if presence.Text != "Players here: Arthur, Guinevere." {
		t.Fatalf("presence text = %q", presence.Text)
	}

	hub.leave(context.Background(), arthur)

	presence = hub.presence("camelot")
	if presence.Text != "Players here: Guinevere." {
		t.Fatalf("presence after leave = %q", presence.Text)
	}
}

func TestRoomHubBroadcastPrunesFailedClients(t *testing.T) {
	restoreRoomWriter(t, func(context.Context, *websocket.Conn, []game.Event) error {
		return nil
	})

	hub := newRoomHub()
	arthur := &clientConn{displayName: "Arthur"}
	guinevere := &clientConn{displayName: "Guinevere"}
	hub.join(context.Background(), arthur, "camelot")
	hub.join(context.Background(), guinevere, "camelot")

	writeRoomEvents = func(context.Context, *websocket.Conn, []game.Event) error {
		return errors.New("connection dropped")
	}

	hub.broadcast(context.Background(), "camelot", game.Event{Type: "say", Text: "Test"}, nil)

	presence := hub.presence("camelot")
	if presence.Text != "Players here: none." {
		t.Fatalf("presence after failed broadcast = %q", presence.Text)
	}
}

func TestRoomHubPresenceAllGroupsPlayersByRoom(t *testing.T) {
	restoreRoomWriter(t, func(context.Context, *websocket.Conn, []game.Event) error {
		return nil
	})

	hub := newRoomHub()
	hub.join(context.Background(), &clientConn{displayName: "Guinevere"}, "camelot")
	hub.join(context.Background(), &clientConn{displayName: "Arthur"}, "camelot")
	hub.join(context.Background(), &clientConn{displayName: "Merlin"}, "counting-room")

	presence := hub.presenceAll()
	want := "Players online by room: camelot: Arthur, Guinevere; counting-room: Merlin."
	if presence.Text != want {
		t.Fatalf("presence all = %q, want %q", presence.Text, want)
	}
}

func restoreRoomWriter(t *testing.T, writer func(context.Context, *websocket.Conn, []game.Event) error) {
	t.Helper()

	previous := writeRoomEvents
	writeRoomEvents = writer
	t.Cleanup(func() {
		writeRoomEvents = previous
	})
}
