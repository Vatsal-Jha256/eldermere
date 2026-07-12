package httpapi

import (
	"strings"
	"testing"

	"github.com/Vatsal-Jha256/eldermere/apps/server/internal/game"
)

func TestPersistentStateChanged(t *testing.T) {
	before := game.PersistentState{
		RoomID: "lantern-yard",
		Party:  []string{"Oath Spirit"},
	}
	after := before

	if persistentStateChanged(before, after) {
		t.Fatal("expected identical states not to be treated as changed")
	}

	after.RoomID = "old-bridge"
	if !persistentStateChanged(before, after) {
		t.Fatal("expected room movement to be treated as changed")
	}
}

func TestPersistentStateChangedDetectsNestedState(t *testing.T) {
	before := game.PersistentState{
		RoomID:   "lantern-yard",
		Factions: map[string]int{"Round Table": 1},
	}
	after := game.PersistentState{
		RoomID:   "lantern-yard",
		Factions: map[string]int{"Round Table": 2},
	}

	if !persistentStateChanged(before, after) {
		t.Fatal("expected faction change to be treated as changed")
	}
}

func TestParseCommandTrimsRawAndJSONCommands(t *testing.T) {
	if got := parseCommand([]byte("  look  ")); got != "look" {
		t.Fatalf("raw command = %q, want look", got)
	}
	if got := parseCommand([]byte(`{"command":"  where  "}`)); got != "where" {
		t.Fatalf("json command = %q, want where", got)
	}
}

func TestCommandTooLong(t *testing.T) {
	if commandTooLong(strings.Repeat("a", maxCommandLength)) {
		t.Fatal("expected max-length command to be accepted")
	}
	if !commandTooLong(strings.Repeat("a", maxCommandLength+1)) {
		t.Fatal("expected oversized command to be rejected")
	}
}

func TestWebSocketAcceptOptionsWildcard(t *testing.T) {
	options := websocketAcceptOptions([]string{"*"})

	if options == nil || !options.InsecureSkipVerify {
		t.Fatalf("expected wildcard policy to allow insecure origin verification, got %#v", options)
	}
	if len(options.OriginPatterns) != 0 {
		t.Fatalf("OriginPatterns = %#v, want empty", options.OriginPatterns)
	}
}

func TestOriginPatternsFromAllowedOrigins(t *testing.T) {
	patterns := originPatterns([]string{"https://eldermere.pages.dev", "http://localhost:5173", "custom.example"})

	want := []string{"eldermere.pages.dev", "localhost:5173", "custom.example"}
	if len(patterns) != len(want) {
		t.Fatalf("patterns = %#v", patterns)
	}
	for i := range want {
		if patterns[i] != want[i] {
			t.Fatalf("patterns = %#v, want %#v", patterns, want)
		}
	}
}
