package game

import (
	"strings"
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

func TestSessionWhereReportsRoomID(t *testing.T) {
	session := NewSession(NewStarterWorld())

	events := session.Handle("where")
	if len(events) != 1 || events[0].Room == nil {
		t.Fatalf("expected one room-aware event, got %#v", events)
	}
	if !strings.Contains(events[0].Text, "Lantern Yard") || !strings.Contains(events[0].Text, "`lantern-yard`") {
		t.Fatalf("expected room name and id, got %q", events[0].Text)
	}
	if events[0].Room.ID != "lantern-yard" {
		t.Fatalf("expected lantern-yard room view, got %#v", events[0].Room)
	}
}

func TestSessionLoreReportsRoomAtmosphere(t *testing.T) {
	session := NewSession(NewStarterWorld())

	events := session.Handle("lore")
	if len(events) != 1 || events[0].Type != "lore" {
		t.Fatalf("expected lore event, got %#v", events)
	}
	if !strings.Contains(events[0].Text, "arthurian court") || !strings.Contains(events[0].Text, "iron rain") || !strings.Contains(events[0].Text, "lanterns") {
		t.Fatalf("expected room atmosphere details, got %q", events[0].Text)
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
	if !strings.Contains(events[0].Text, "Try `help`") {
		t.Fatalf("expected unknown command to suggest help, got %q", events[0].Text)
	}
}

func TestHelpCommandTopics(t *testing.T) {
	session := NewSession(NewStarterWorld())

	events := session.Handle("help")
	if len(events) != 1 || events[0].Type != "help" {
		t.Fatalf("expected help event, got %#v", events)
	}
	if !strings.Contains(events[0].Text, "Help topics") || !strings.Contains(events[0].Text, "story") {
		t.Fatalf("expected help index, got %q", events[0].Text)
	}

	events = session.Handle("help story")
	if len(events) != 1 || !strings.Contains(events[0].Text, "story start <id>") {
		t.Fatalf("expected story help, got %#v", events)
	}
}

func TestTalkAliasUsesSayEvent(t *testing.T) {
	session := NewSession(NewStarterWorld())

	events := session.Handle("talk hello")
	if len(events) != 1 || events[0].Type != "say" {
		t.Fatalf("expected say event from talk alias, got %#v", events)
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

func TestOddsCommandShowsCurrentRoomChecks(t *testing.T) {
	session := NewSession(NewStarterWorld())
	session.Handle("go north")

	events := session.Handle("odds")
	if len(events) != 1 || events[0].Type != "odds" {
		t.Fatalf("expected odds event, got %#v", events)
	}
	if !strings.Contains(events[0].Text, "Fight odds:") || !strings.Contains(events[0].Text, "Bridge Debt Collector") {
		t.Fatalf("expected fight odds for current room, got %q", events[0].Text)
	}

	session.Handle("go south")
	session.Handle("go east")
	events = session.Handle("odds recruit")
	if len(events) != 1 || !strings.Contains(events[0].Text, "Recruit odds:") || !strings.Contains(events[0].Text, "Oath Spirit") {
		t.Fatalf("expected recruit odds for current room, got %#v", events)
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

func TestFightUsesCriticalOutcomeText(t *testing.T) {
	world, err := NewWorld([]Room{
		{
			ID:          "arena",
			Name:        "Arena",
			Description: "A valid test room.",
			Exits:       map[string]string{},
			Encounter: &Encounter{
				Name:     "Duelist",
				DC:       30,
				Win:      "You win.",
				Lose:     "You lose.",
				CritWin:  "You turn the duel into legend.",
				CritLose: "You drop the blade before the bell.",
			},
		},
	})
	if err != nil {
		t.Fatalf("build world: %v", err)
	}
	session := NewSessionWithRoller(world, sequenceRoller(20))
	session.roomID = "arena"

	events := session.Handle("fight")
	if len(events) != 1 || !strings.Contains(events[0].Text, "You turn the duel into legend.") {
		t.Fatalf("expected critical win text, got %#v", events)
	}
}

func TestRecruitUsesAdvantageRollMode(t *testing.T) {
	world, err := NewWorld([]Room{
		{
			ID:          "market",
			Name:        "Market",
			Description: "A valid test room.",
			Exits:       map[string]string{},
			Recruitable: &Recruitable{
				Name:     "Clever Page",
				DC:       15,
				RollMode: rollAdvantage,
				Success:  "The page joins.",
				Failure:  "The page waits.",
			},
		},
	})
	if err != nil {
		t.Fatalf("build world: %v", err)
	}
	session := NewSessionWithRoller(world, sequenceRoller(2, 14))
	session.roomID = "market"

	events := session.Handle("recruit")
	if len(events) != 1 || !strings.Contains(events[0].Text, "2/14 advantage") || !strings.Contains(events[0].Text, "The page joins.") {
		t.Fatalf("expected advantage recruit success, got %#v", events)
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

func TestStoryCommandListsAndShowsLoadedArcs(t *testing.T) {
	world, err := NewStarterWorld().WithStoryArcs([]StoryArc{
		{
			ID:           "sword-test",
			Title:        "The Sword Test",
			Kind:         "main",
			LoreBeats:    []string{"Sword test and contested kingship"},
			SourceIDs:    []string{"malory-1251"},
			Summary:      "Arthur's legitimacy is contested.",
			OriginalHook: "The under-market sells false proof.",
			Steps: []StoryStep{
				{ID: "witness", Title: "Find a witness", Objective: "Find someone who saw the sword test."},
			},
		},
	})
	if err != nil {
		t.Fatalf("attach stories: %v", err)
	}
	session := NewSession(world)

	events := session.Handle("story")
	if len(events) != 1 || events[0].Type != "story" {
		t.Fatalf("expected story list event, got %#v", events)
	}
	if !strings.Contains(events[0].Text, "sword-test") {
		t.Fatalf("expected story list to include sword-test, got %q", events[0].Text)
	}

	events = session.Handle("story sword-test")
	if len(events) != 1 || !strings.Contains(events[0].Text, "malory-1251") {
		t.Fatalf("expected story detail with source id, got %#v", events)
	}
}

func TestStoryCommandStartsAdvancesCompletesAndPersists(t *testing.T) {
	world, err := NewStarterWorld().WithStoryArcs([]StoryArc{
		{
			ID:            "sword-test",
			Title:         "The Sword Test",
			Kind:          "main",
			LoreBeats:     []string{"Sword test and contested kingship"},
			SourceIDs:     []string{"malory-1251"},
			Summary:       "Arthur's legitimacy is contested.",
			OriginalHook:  "The under-market sells false proof.",
			AddsTags:      []string{"sword-test", "contested-kingship"},
			VariationTags: []string{"bribed-witness"},
			Steps: []StoryStep{
				{
					ID:          "witness",
					Title:       "Find a witness",
					Objective:   "Find someone who saw the sword test.",
					OutcomeTags: []string{"witness-contradiction"},
				},
				{
					ID:          "publish",
					Title:       "Publish a version",
					Objective:   "Choose the public account.",
					OutcomeTags: []string{"arthur-accepted"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("attach stories: %v", err)
	}
	session := NewSessionWithRoller(world, func(sides int) int {
		return 1
	})

	events := session.Handle("story start sword-test")
	if len(events) != 1 || !strings.Contains(events[0].Text, "Variant: bribed-witness") {
		t.Fatalf("expected story start with variant, got %#v", events)
	}
	if session.story.ActiveArcID != "sword-test" {
		t.Fatalf("expected active sword-test story, got %#v", session.story)
	}

	resumed := NewSessionFromState(world, session.PersistentState())
	events = resumed.Handle("story next")
	if len(events) != 1 || !strings.Contains(events[0].Text, "Step 2/2") {
		t.Fatalf("expected story to advance to second step, got %#v", events)
	}

	events = resumed.Handle("story next")
	if len(events) != 1 || !strings.Contains(events[0].Text, "Story complete") {
		t.Fatalf("expected story completion, got %#v", events)
	}
	if resumed.story.ActiveArcID != "" || !storyContains(resumed.story.CompletedArcIDs, "sword-test") {
		t.Fatalf("expected completed story state, got %#v", resumed.story)
	}
	for _, tag := range []string{"arthur-accepted", "bribed-witness", "contested-kingship", "sword-test", "witness-contradiction"} {
		if !storyContains(resumed.story.Tags, tag) {
			t.Fatalf("expected story tag %q in %#v", tag, resumed.story.Tags)
		}
	}
}

func TestStoryRequiredTagsGateArcStart(t *testing.T) {
	baseWorld, err := NewStarterWorld().WithStoryContent(StoryContent{
		Tags: []string{"arthurian"},
		Arcs: []StoryArc{
			{
				ID:           "sword-test",
				Title:        "The Sword Test",
				Kind:         "main",
				LoreBeats:    []string{"Sword test and contested kingship"},
				SourceIDs:    []string{"malory-1251"},
				Summary:      "Arthur's legitimacy is contested.",
				OriginalHook: "The under-market sells false proof.",
				RequiredTags: []string{"arthurian"},
				AddsTags:     []string{"arthur-accepted"},
				Steps: []StoryStep{
					{ID: "witness", Title: "Find a witness", Objective: "Find someone who saw the sword test."},
				},
			},
			{
				ID:           "round-table-fractures",
				Title:        "The Table Makes Rivals Sit Still",
				Kind:         "main",
				LoreBeats:    []string{"The Round Table is both ideal and pressure cooker."},
				SourceIDs:    []string{"malory-1251"},
				Summary:      "Seat politics strain fellowship.",
				OriginalHook: "Every seat has a shadow claimant.",
				RequiredTags: []string{"arthur-accepted"},
				Steps: []StoryStep{
					{ID: "seat", Title: "Answer a seat claim", Objective: "Settle a claim."},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("attach stories: %v", err)
	}
	session := NewSession(baseWorld)

	events := session.Handle("story start round-table-fractures")
	if len(events) != 1 || !strings.Contains(events[0].Text, "Missing tags: arthur-accepted") {
		t.Fatalf("expected locked story arc, got %#v", events)
	}

	events = session.Handle("story start sword-test")
	if len(events) != 1 || !strings.Contains(events[0].Text, "Story started") {
		t.Fatalf("expected seed arthurian tag to unlock sword-test, got %#v", events)
	}
	events = session.Handle("story next")
	if len(events) != 1 || !strings.Contains(events[0].Text, "Story complete") {
		t.Fatalf("expected sword-test completion, got %#v", events)
	}
	events = session.Handle("story start round-table-fractures")
	if len(events) != 1 || !strings.Contains(events[0].Text, "Story started") {
		t.Fatalf("expected earned tag to unlock round-table-fractures, got %#v", events)
	}
}

func TestStoryRequiredFactionsGateArcStartAndEligibility(t *testing.T) {
	world, err := NewStarterWorld().WithStoryArcs([]StoryArc{
		{
			ID:           "broker-audit",
			Title:        "Broker Audit",
			Kind:         "side",
			LoreBeats:    []string{"Faction trust can unlock side business."},
			SourceIDs:    []string{"malory-1251"},
			Summary:      "A broker only speaks after the under-market trusts the player.",
			OriginalHook: "A locked ledger turns faction reputation into a story key.",
			RequiredFactions: map[string]int{
				"Camelot Underbelly": 1,
			},
			Steps: []StoryStep{
				{ID: "audit", Title: "Audit The Broker", Objective: "Read the broker's hidden ledger."},
			},
		},
	})
	if err != nil {
		t.Fatalf("attach stories: %v", err)
	}
	session := NewSession(world)

	events := session.Handle("story eligible")
	if len(events) != 1 || !strings.Contains(events[0].Text, "Eligible story arcs: none.") {
		t.Fatalf("expected no eligible arcs, got %#v", events)
	}

	events = session.Handle("story locked")
	if len(events) != 1 || !strings.Contains(events[0].Text, "Missing factions: Camelot Underbelly +1 (current +0)") {
		t.Fatalf("expected locked faction reason, got %#v", events)
	}

	events = session.Handle("story start broker-audit")
	if len(events) != 1 || !strings.Contains(events[0].Text, "Missing factions: Camelot Underbelly +1 (current +0)") {
		t.Fatalf("expected faction lock on start, got %#v", events)
	}

	session.factions["Camelot Underbelly"] = 1
	events = session.Handle("story eligible")
	if len(events) != 1 || !strings.Contains(events[0].Text, "broker-audit") {
		t.Fatalf("expected broker-audit to become eligible, got %#v", events)
	}

	events = session.Handle("story start broker-audit")
	if len(events) != 1 || !strings.Contains(events[0].Text, "Story started") {
		t.Fatalf("expected faction-gated story to start, got %#v", events)
	}
}

func TestTravelCommandMovesToContentPackEntryRoom(t *testing.T) {
	world, err := NewStarterWorld().WithPackRuntimeContent(PackRuntimeContent{
		Rooms: []Room{
			{
				ID:          "stone-yard",
				Name:        "Stone Yard",
				Description: "A valid content-pack room.",
				Exits:       map[string]string{},
			},
		},
		Entries: map[string]string{
			"arthurian-core": "stone-yard",
		},
	})
	if err != nil {
		t.Fatalf("attach pack runtime content: %v", err)
	}
	session := NewSession(world)

	events := session.Handle("travel arthurian-core")
	if len(events) != 2 {
		t.Fatalf("expected move and room events, got %#v", events)
	}
	if session.RoomID() != "stone-yard" {
		t.Fatalf("expected to travel to stone-yard, got %q", session.RoomID())
	}
	if events[1].Room == nil || events[1].Room.ID != "stone-yard" {
		t.Fatalf("expected stone-yard room event, got %#v", events)
	}
}

func TestStoryStepRequiresRoomAndCanChangeFaction(t *testing.T) {
	world, err := NewStarterWorld().WithPackRuntimeContent(PackRuntimeContent{
		Rooms: []Room{
			{
				ID:          "stone-yard",
				Name:        "Stone Yard",
				Description: "A valid content-pack room.",
				Exits:       map[string]string{},
			},
		},
		Stories: StoryContent{
			Tags: []string{"arthurian"},
			Arcs: []StoryArc{
				{
					ID:           "sword-test",
					Title:        "The Sword Test",
					Kind:         "main",
					LoreBeats:    []string{"Sword test and contested kingship"},
					SourceIDs:    []string{"malory-1251"},
					Summary:      "Arthur's legitimacy is contested.",
					OriginalHook: "The under-market sells false proof.",
					RequiredTags: []string{"arthurian"},
					Steps: []StoryStep{
						{
							ID:          "witness",
							Title:       "Find a witness",
							RoomHint:    "stone-yard",
							Objective:   "Find someone who saw the sword test.",
							OutcomeTags: []string{"witness-contradiction"},
							FactionEffects: map[string]int{
								"Round Table": 1,
							},
						},
					},
				},
			},
		},
		Entries: map[string]string{
			"arthurian-core": "stone-yard",
		},
	})
	if err != nil {
		t.Fatalf("attach pack runtime content: %v", err)
	}
	session := NewSession(world)

	session.Handle("story start sword-test")
	events := session.Handle("story status")
	if len(events) != 1 || !strings.Contains(events[0].Text, "Room: stone-yard") || !strings.Contains(events[0].Text, "Find someone who saw the sword test.") {
		t.Fatalf("expected story status to include room and objective guidance, got %#v", events)
	}
	events = session.Handle("story next")
	if len(events) != 1 || !strings.Contains(events[0].Text, "needs room `stone-yard`") {
		t.Fatalf("expected room-gated story step, got %#v", events)
	}

	session.Handle("travel arthurian-core")
	events = session.Handle("story next")
	if len(events) != 1 || !strings.Contains(events[0].Text, "Story complete") {
		t.Fatalf("expected story completion after reaching room, got %#v", events)
	}
	events = session.Handle("factions")
	if len(events) != 1 || !strings.Contains(events[0].Text, "Round Table +1") {
		t.Fatalf("expected story faction effect, got %#v", events)
	}
}
