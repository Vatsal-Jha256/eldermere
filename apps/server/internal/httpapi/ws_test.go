package httpapi

import (
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
