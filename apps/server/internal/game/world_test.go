package game

import "testing"

func TestSessionLookAndMove(t *testing.T) {
	session := NewSession(NewStarterWorld())

	events := session.Handle("look")
	if len(events) != 1 || events[0].Room == nil {
		t.Fatalf("expected one room event, got %#v", events)
	}
	if events[0].Room.Name != "Lantern Yard" {
		t.Fatalf("expected Lantern Yard, got %q", events[0].Room.Name)
	}

	events = session.Handle("go north")
	if len(events) != 2 {
		t.Fatalf("expected move and room events, got %#v", events)
	}
	if events[1].Room == nil || events[1].Room.Name != "Old Bridge" {
		t.Fatalf("expected Old Bridge after moving north, got %#v", events)
	}
}

func TestSessionUnknownCommand(t *testing.T) {
	session := NewSession(NewStarterWorld())

	events := session.Handle("dance")
	if len(events) != 1 {
		t.Fatalf("expected one event, got %#v", events)
	}
	if events[0].Type != "error" {
		t.Fatalf("expected error event, got %q", events[0].Type)
	}
}
