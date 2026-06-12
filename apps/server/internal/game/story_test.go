package game

import (
	"os"
	"path/filepath"
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
