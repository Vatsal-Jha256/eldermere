package game

import (
	"testing"
	"testing/fstest"
)

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

func TestLoadWorldValidatesExitTargets(t *testing.T) {
	files := fstest.MapFS{
		"rooms.json": {
			Data: []byte(`{
				"rooms": [
					{
						"id": "start",
						"name": "Start",
						"description": "A valid room.",
						"exits": { "north": "missing" }
					}
				]
			}`),
		},
	}

	_, err := LoadWorld(files, "rooms.json")
	if err == nil {
		t.Fatal("expected invalid exit target to fail")
	}
}

func TestLoadWorldFromContentFile(t *testing.T) {
	world := NewStarterWorld()
	session := NewSession(world)

	events := session.Handle("go east")
	if len(events) != 2 {
		t.Fatalf("expected move and room events, got %#v", events)
	}
	if events[1].Room == nil || events[1].Room.ID != "market-under" {
		t.Fatalf("expected market-under room, got %#v", events)
	}
}

func TestFightUsesRoomEncounter(t *testing.T) {
	session := NewSessionWithRoller(NewStarterWorld(), func(sides int) int {
		return 10
	})
	session.Handle("go north")

	events := session.Handle("fight")
	if len(events) != 1 {
		t.Fatalf("expected one fight event, got %#v", events)
	}
	if events[0].Type != "fight" {
		t.Fatalf("expected fight event, got %q", events[0].Type)
	}
}

func TestRecruitAddsCompanionToParty(t *testing.T) {
	session := NewSessionWithRoller(NewStarterWorld(), func(sides int) int {
		return 20
	})
	session.Handle("go east")

	events := session.Handle("recruit")
	if len(events) != 1 || events[0].Type != "party" {
		t.Fatalf("expected party event, got %#v", events)
	}

	events = session.Handle("party")
	if len(events) != 1 {
		t.Fatalf("expected one party status event, got %#v", events)
	}
	if events[0].Text != "Party: Oath Spirit." {
		t.Fatalf("expected Oath Spirit in party, got %q", events[0].Text)
	}
}

func TestQuestCanStartCollectItemAndComplete(t *testing.T) {
	session := NewSession(NewStarterWorld())

	events := session.Handle("quest")
	if len(events) != 1 || events[0].Type != "quest" {
		t.Fatalf("expected quest start event, got %#v", events)
	}

	session.Handle("go east")
	session.Handle("go down")

	events = session.Handle("take")
	if len(events) != 1 || events[0].Type != "inventory" {
		t.Fatalf("expected inventory event after take, got %#v", events)
	}

	session.Handle("go up")
	session.Handle("go west")

	events = session.Handle("quest")
	if len(events) != 1 {
		t.Fatalf("expected quest completion event, got %#v", events)
	}
	if events[0].Type != "quest" || !session.quest.Completed {
		t.Fatalf("expected completed quest, got event %#v state %#v", events[0], session.quest)
	}
}

func TestSessionCanResumePersistentState(t *testing.T) {
	session := NewSession(NewStarterWorld())
	session.Handle("quest")
	session.Handle("go east")
	session.Handle("go down")
	session.Handle("take")

	resumed := NewSessionFromState(NewStarterWorld(), session.PersistentState())
	events := resumed.Handle("inventory")
	if len(events) != 1 {
		t.Fatalf("expected inventory event, got %#v", events)
	}
	if events[0].Text != "Inventory: Excalibur Fragment." {
		t.Fatalf("expected restored inventory, got %q", events[0].Text)
	}

	events = resumed.Handle("look")
	if events[0].Room == nil || events[0].Room.ID != "smuggler-vault" {
		t.Fatalf("expected resumed room smuggler-vault, got %#v", events)
	}
}

func TestGatedExitRequiresMapItem(t *testing.T) {
	session := NewSession(NewStarterWorld())

	events := session.Handle("go under")
	if len(events) != 1 || events[0].Type != "error" {
		t.Fatalf("expected locked route error, got %#v", events)
	}

	session.Handle("go west")
	session.Handle("take")
	session.Handle("go east")

	events = session.Handle("go under")
	if len(events) != 2 {
		t.Fatalf("expected move and room after map unlock, got %#v", events)
	}
	if events[1].Room == nil || events[1].Room.ID != "smuggler-vault" {
		t.Fatalf("expected smuggler-vault after gated move, got %#v", events)
	}
}

func TestUnlockedGatedExitAppearsInRoomView(t *testing.T) {
	session := NewSession(NewStarterWorld())
	session.Handle("go west")
	session.Handle("take")
	session.Handle("go east")

	events := session.Handle("look")
	if len(events) != 1 || events[0].Room == nil {
		t.Fatalf("expected room event, got %#v", events)
	}
	if events[0].Room.Exits["under"] != "smuggler-vault" {
		t.Fatalf("expected unlocked under exit in room view, got %#v", events[0].Room.Exits)
	}
}

func TestFightCanChangeFactionReputation(t *testing.T) {
	session := NewSessionWithRoller(NewStarterWorld(), func(sides int) int {
		return 20
	})
	session.Handle("go north")

	events := session.Handle("fight")
	if len(events) != 1 || events[0].Type != "fight" {
		t.Fatalf("expected fight event, got %#v", events)
	}

	events = session.Handle("factions")
	if len(events) != 1 {
		t.Fatalf("expected factions event, got %#v", events)
	}
	if events[0].Text != "Factions: Camelot Underbelly +1, Mordred's Brokers -1." {
		t.Fatalf("unexpected faction text: %q", events[0].Text)
	}
}

func TestQuestStoresVariant(t *testing.T) {
	session := NewSessionWithRoller(NewStarterWorld(), func(sides int) int {
		return 1
	})

	session.Handle("quest")
	if session.quest.Variant == "" {
		t.Fatal("expected quest variant to be selected")
	}

	events := session.Handle("quest")
	if len(events) != 1 || events[0].Text == "Quest active: find the stolen Excalibur fragment in the under-market route." {
		t.Fatalf("expected variant quest text, got %#v", events)
	}
}
