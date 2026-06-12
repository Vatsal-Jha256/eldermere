package game

import (
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
