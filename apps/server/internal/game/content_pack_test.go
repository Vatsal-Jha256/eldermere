package game

import (
	"testing"
	"testing/fstest"
)

func TestLoadContentPack(t *testing.T) {
	files := fstest.MapFS{
		"pack.json": {
			Data: []byte(`{
				"id": "greek-crossing",
				"name": "Greek Crossing",
				"myth_region": "Greek",
				"tags": ["greek", "underworld-route"],
				"rooms_file": "rooms.json",
				"interactions": [
					{
						"id": "grail-disturbs-underworld",
						"when_tags": ["arthurian", "grail-curse"],
						"adds_tags": ["underworld-unrest"],
						"description": "A Grail curse disturbs the underworld ferry routes."
					}
				]
			}`),
		},
	}

	pack, err := LoadContentPack(files, "pack.json")
	if err != nil {
		t.Fatalf("load content pack: %v", err)
	}
	if pack.ID != "greek-crossing" {
		t.Fatalf("expected greek-crossing, got %q", pack.ID)
	}
	if len(pack.Interactions) != 1 {
		t.Fatalf("expected one interaction, got %#v", pack.Interactions)
	}
}

func TestValidateContentPackRequiresInteractionTags(t *testing.T) {
	err := ValidateContentPack(ContentPack{
		ID:         "bad-pack",
		Name:       "Bad Pack",
		MythRegion: "Greek",
		RoomsFile:  "rooms.json",
		Interactions: []PackInteraction{
			{
				ID:          "missing-tags",
				Description: "No tags.",
			},
		},
	})
	if err == nil {
		t.Fatal("expected missing interaction tags to fail")
	}
}
