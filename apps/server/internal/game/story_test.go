package game

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadStoryDocument(t *testing.T) {
	files := fstest.MapFS{
		"story_arcs.json": {
			Data: []byte(`{
				"arcs": [
					{
						"id": "sword-test",
						"title": "The Sword Test",
						"kind": "main",
						"lore_beats": ["Sword test and contested kingship"],
						"source_ids": ["malory-1251"],
						"summary": "Arthur's legitimacy is contested.",
						"original_hook": "The under-market sells false proof.",
						"steps": [
							{
								"id": "witness",
								"title": "Find a witness",
								"objective": "Find someone who saw the sword test.",
								"commands": ["quest", "go west"]
							}
						]
					}
				]
			}`),
		},
	}

	document, err := LoadStoryDocument(files, "story_arcs.json")
	if err != nil {
		t.Fatalf("load story document: %v", err)
	}
	if len(document.Arcs) != 1 || document.Arcs[0].ID != "sword-test" {
		t.Fatalf("unexpected story document: %#v", document)
	}
}

func TestValidateStoryDocumentRequiresLoreBeats(t *testing.T) {
	err := ValidateStoryDocument(StoryDocument{
		Arcs: []StoryArc{
			{
				ID:           "bad",
				Title:        "Bad",
				Kind:         "main",
				SourceIDs:    []string{"malory-1251"},
				Summary:      "Missing lore beats.",
				OriginalHook: "No grounding.",
				Steps: []StoryStep{
					{ID: "step", Title: "Step", Objective: "Do thing."},
				},
			},
		},
	})
	if err == nil {
		t.Fatal("expected missing lore beats to fail")
	}
}

func TestValidateStoryDocumentRejectsEmptyRequiredFaction(t *testing.T) {
	err := ValidateStoryDocument(StoryDocument{
		Arcs: []StoryArc{
			{
				ID:           "bad",
				Title:        "Bad",
				Kind:         "side",
				LoreBeats:    []string{"Faction gates should name a faction."},
				SourceIDs:    []string{"malory-1251"},
				Summary:      "Bad faction gate.",
				OriginalHook: "No faction name.",
				RequiredFactions: map[string]int{
					" ": 1,
				},
				Steps: []StoryStep{
					{ID: "step", Title: "Step", Objective: "Do thing."},
				},
			},
		},
	})
	if err == nil {
		t.Fatal("expected empty required faction to fail")
	}
}

func TestLoadStoryArcsFromContentPacks(t *testing.T) {
	root := t.TempDir()
	packDir := filepath.Join(root, "arthurian-core")
	if err := os.Mkdir(packDir, 0o755); err != nil {
		t.Fatalf("make pack dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packDir, "pack.json"), []byte(`{
		"id": "arthurian-core",
		"name": "Arthurian Core",
		"myth_region": "Arthurian",
		"tags": ["arthurian"],
		"rooms_file": "rooms.json",
		"entry_room": "stone-yard",
		"story_file": "story_arcs.json"
	}`), 0o644); err != nil {
		t.Fatalf("write pack manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packDir, "rooms.json"), []byte(`{
		"rooms": [
			{
				"id": "stone-yard",
				"name": "Stone Yard",
				"description": "A valid room.",
				"exits": {}
			}
		]
	}`), 0o644); err != nil {
		t.Fatalf("write rooms: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packDir, "story_arcs.json"), []byte(`{
		"arcs": [
			{
				"id": "sword-test",
				"title": "The Sword Test",
				"kind": "main",
				"lore_beats": ["Sword test and contested kingship"],
				"source_ids": ["malory-1251"],
				"summary": "Arthur's legitimacy is contested.",
				"original_hook": "The under-market sells false proof.",
				"steps": [
					{
						"id": "witness",
						"title": "Find a witness",
						"objective": "Find someone who saw the sword test."
					}
				]
			}
		]
	}`), 0o644); err != nil {
		t.Fatalf("write story arcs: %v", err)
	}

	arcs, err := LoadStoryArcsFromContentPacks(root)
	if err != nil {
		t.Fatalf("load story arcs: %v", err)
	}
	if len(arcs) != 1 || arcs[0].ID != "sword-test" {
		t.Fatalf("unexpected arcs: %#v", arcs)
	}

	content, err := LoadStoryContentFromContentPacks(root)
	if err != nil {
		t.Fatalf("load story content: %v", err)
	}
	if !storyContains(content.Tags, "arthurian") {
		t.Fatalf("expected pack tags to seed story content, got %#v", content.Tags)
	}

	runtimeContent, err := LoadPackRuntimeContentFromContentPacks(root)
	if err != nil {
		t.Fatalf("load runtime content: %v", err)
	}
	if runtimeContent.Entries["arthurian-core"] != "stone-yard" {
		t.Fatalf("expected arthurian-core entry room, got %#v", runtimeContent.Entries)
	}
}

func TestValidatePackRuntimeReferencesChecksStoryRoomHints(t *testing.T) {
	content := PackRuntimeContent{
		Stories: StoryContent{
			Arcs: []StoryArc{
				{
					ID:           "bad-room",
					Title:        "Bad Room",
					Kind:         "side",
					LoreBeats:    []string{"Story steps need valid rooms."},
					SourceIDs:    []string{"malory-1251"},
					Summary:      "Invalid room hint.",
					OriginalHook: "A builder typo points nowhere.",
					Steps: []StoryStep{
						{ID: "step", Title: "Step", RoomHint: "missing-room", Objective: "Go nowhere."},
					},
				},
			},
		},
	}

	err := ValidatePackRuntimeReferences(content, NewStarterWorld(), map[string]bool{"malory-1251": true})
	if err == nil || !strings.Contains(err.Error(), "unknown room_hint") {
		t.Fatalf("expected bad room hint error, got %v", err)
	}
}

func TestValidatePackRuntimeReferencesChecksStorySources(t *testing.T) {
	content := PackRuntimeContent{
		Stories: StoryContent{
			Arcs: []StoryArc{
				{
					ID:           "bad-source",
					Title:        "Bad Source",
					Kind:         "side",
					LoreBeats:    []string{"Story arcs need cited sources."},
					SourceIDs:    []string{"missing-source"},
					Summary:      "Invalid source id.",
					OriginalHook: "A builder typo cites nothing.",
					Steps: []StoryStep{
						{ID: "step", Title: "Step", RoomHint: "lantern-yard", Objective: "Use a known room."},
					},
				},
			},
		},
	}

	err := ValidatePackRuntimeReferences(content, NewStarterWorld(), map[string]bool{"malory-1251": true})
	if err == nil || !strings.Contains(err.Error(), "unknown source id") {
		t.Fatalf("expected bad source id error, got %v", err)
	}
}

func TestArthurianCoreMainPlotCanAdvanceInOrder(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file path")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../.."))
	content, err := LoadPackRuntimeContentFromContentPacks(filepath.Join(repoRoot, "content-packs"))
	if err != nil {
		t.Fatalf("load content packs: %v", err)
	}
	world, err := NewStarterWorld().WithPackRuntimeContent(content)
	if err != nil {
		t.Fatalf("attach runtime content: %v", err)
	}
	session := NewSessionWithRoller(world, func(sides int) int {
		return 1
	})

	mainArcIDs := []string{
		"sword-test",
		"merlins-ledger",
		"scabbard-problem",
		"round-table-fractures",
		"grail-sorting",
		"mordreds-claim",
		"avalon-ambiguity",
	}

	for _, arcID := range mainArcIDs {
		arc, ok := world.stories[arcID]
		if !ok {
			t.Fatalf("missing main arc %q", arcID)
		}
		events := session.Handle("story start " + arcID)
		if len(events) != 1 || !strings.Contains(events[0].Text, "Story started") {
			t.Fatalf("start %s: %#v", arcID, events)
		}
		for _, step := range arc.Steps {
			if step.RoomHint != "" {
				if _, ok := world.rooms[step.RoomHint]; !ok {
					t.Fatalf("arc %s step %s points to missing room %q", arcID, step.ID, step.RoomHint)
				}
				session.roomID = step.RoomHint
			}
			events = session.Handle("story next")
			if len(events) != 1 || strings.Contains(events[0].Text, "needs room") || strings.Contains(events[0].Text, "locked") {
				t.Fatalf("advance %s/%s from room %q: %#v", arcID, step.ID, session.roomID, events)
			}
		}
		if !storyContains(session.story.CompletedArcIDs, arcID) {
			t.Fatalf("expected arc %s to be complete, got %#v", arcID, session.story)
		}
	}

	for _, tag := range []string{"arthurian-v1-spine-complete", "avalon-ambiguity", "camlann-route-set", "grail-witness"} {
		if !storyContains(session.story.Tags, tag) {
			t.Fatalf("expected completed main plot tag %q in %#v", tag, session.story.Tags)
		}
	}
}
