package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ContentPack struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	MythRegion   string            `json:"myth_region"`
	Tags         []string          `json:"tags"`
	RoomsFile    string            `json:"rooms_file"`
	EntryRoom    string            `json:"entry_room,omitempty"`
	StoryFile    string            `json:"story_file,omitempty"`
	Interactions []PackInteraction `json:"interactions"`
}

type PackInteraction struct {
	ID          string   `json:"id"`
	WhenTags    []string `json:"when_tags"`
	AddsTags    []string `json:"adds_tags"`
	Description string   `json:"description"`
}

func LoadContentPack(files fs.FS, path string) (ContentPack, error) {
	payload, err := fs.ReadFile(files, path)
	if err != nil {
		return ContentPack{}, err
	}

	var pack ContentPack
	if err := json.Unmarshal(payload, &pack); err != nil {
		return ContentPack{}, err
	}
	if err := ValidateContentPack(pack); err != nil {
		return ContentPack{}, err
	}
	return pack, nil
}

func ValidateContentPack(pack ContentPack) error {
	if strings.TrimSpace(pack.ID) == "" {
		return fmt.Errorf("pack id is required")
	}
	if strings.TrimSpace(pack.Name) == "" {
		return fmt.Errorf("pack %q name is required", pack.ID)
	}
	if strings.TrimSpace(pack.MythRegion) == "" {
		return fmt.Errorf("pack %q myth_region is required", pack.ID)
	}
	if strings.TrimSpace(pack.RoomsFile) == "" {
		return fmt.Errorf("pack %q rooms_file is required", pack.ID)
	}

	seen := map[string]bool{}
	for _, tag := range pack.Tags {
		if strings.TrimSpace(tag) == "" {
			return fmt.Errorf("pack %q has empty tag", pack.ID)
		}
		seen[tag] = true
	}

	for _, interaction := range pack.Interactions {
		if strings.TrimSpace(interaction.ID) == "" {
			return fmt.Errorf("pack %q has interaction with empty id", pack.ID)
		}
		if strings.TrimSpace(interaction.Description) == "" {
			return fmt.Errorf("pack %q interaction %q description is required", pack.ID, interaction.ID)
		}
		if len(interaction.WhenTags) == 0 {
			return fmt.Errorf("pack %q interaction %q requires at least one when_tag", pack.ID, interaction.ID)
		}
		for _, tag := range interaction.WhenTags {
			if strings.TrimSpace(tag) == "" {
				return fmt.Errorf("pack %q interaction %q has empty when_tag", pack.ID, interaction.ID)
			}
		}
		for _, tag := range interaction.AddsTags {
			if strings.TrimSpace(tag) == "" {
				return fmt.Errorf("pack %q interaction %q has empty adds_tag", pack.ID, interaction.ID)
			}
		}
	}

	return nil
}

type StoryDocument struct {
	Arcs []StoryArc `json:"arcs"`
}

type StoryContent struct {
	Arcs []StoryArc
	Tags []string
}

type PackRuntimeContent struct {
	Rooms   []Room
	Stories StoryContent
	Entries map[string]string
}

type StoryArc struct {
	ID               string         `json:"id"`
	Title            string         `json:"title"`
	Kind             string         `json:"kind"`
	LoreBeats        []string       `json:"lore_beats"`
	SourceIDs        []string       `json:"source_ids"`
	Summary          string         `json:"summary"`
	OriginalHook     string         `json:"original_hook"`
	RequiredTags     []string       `json:"required_tags,omitempty"`
	RequiredFactions map[string]int `json:"required_factions,omitempty"`
	AddsTags         []string       `json:"adds_tags,omitempty"`
	Steps            []StoryStep    `json:"steps"`
	VariationTags    []string       `json:"variation_tags,omitempty"`
}

type StoryStep struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	RoomHint       string         `json:"room_hint,omitempty"`
	Objective      string         `json:"objective"`
	Commands       []string       `json:"commands,omitempty"`
	OutcomeTags    []string       `json:"outcome_tags,omitempty"`
	FactionEffects map[string]int `json:"faction_effects,omitempty"`
}

func LoadStoryDocument(files fs.FS, path string) (StoryDocument, error) {
	payload, err := fs.ReadFile(files, path)
	if err != nil {
		return StoryDocument{}, err
	}

	var document StoryDocument
	if err := json.Unmarshal(payload, &document); err != nil {
		return StoryDocument{}, err
	}
	if err := ValidateStoryDocument(document); err != nil {
		return StoryDocument{}, err
	}
	return document, nil
}

func LoadStoryArcsFromContentPacks(root string) ([]StoryArc, error) {
	content, err := LoadPackRuntimeContentFromContentPacks(root)
	if err != nil {
		return nil, err
	}
	return content.Stories.Arcs, nil
}

func LoadStoryContentFromContentPacks(root string) (StoryContent, error) {
	content, err := LoadPackRuntimeContentFromContentPacks(root)
	if err != nil {
		return StoryContent{}, err
	}
	return content.Stories, nil
}

func LoadPackRuntimeContentFromContentPacks(root string) (PackRuntimeContent, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return PackRuntimeContent{}, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	content := PackRuntimeContent{
		Entries: map[string]string{},
	}
	for _, name := range names {
		packPath := filepath.Join(root, name)
		pack, err := LoadContentPack(os.DirFS(packPath), "pack.json")
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return PackRuntimeContent{}, fmt.Errorf("load pack %s: %w", name, err)
		}
		rooms, err := LoadRooms(os.DirFS(packPath), pack.RoomsFile)
		if err != nil {
			return PackRuntimeContent{}, fmt.Errorf("load rooms for pack %s: %w", name, err)
		}
		if len(rooms) == 0 {
			return PackRuntimeContent{}, fmt.Errorf("pack %s has no rooms", name)
		}
		entryRoom := pack.EntryRoom
		if entryRoom == "" {
			entryRoom = rooms[0].ID
		}
		if !roomIDExists(rooms, entryRoom) {
			return PackRuntimeContent{}, fmt.Errorf("pack %s entry_room %q does not exist", name, entryRoom)
		}
		content.Rooms = append(content.Rooms, rooms...)
		content.Entries[pack.ID] = entryRoom
		content.Stories.Tags = append(content.Stories.Tags, pack.Tags...)
		if pack.StoryFile == "" {
			continue
		}
		document, err := LoadStoryDocument(os.DirFS(packPath), pack.StoryFile)
		if err != nil {
			return PackRuntimeContent{}, fmt.Errorf("load story file for pack %s: %w", name, err)
		}
		content.Stories.Arcs = append(content.Stories.Arcs, document.Arcs...)
	}

	content.Stories.Tags = appendStoryTags(nil, content.Stories.Tags...)
	return content, nil
}

func roomIDExists(rooms []Room, id string) bool {
	for _, room := range rooms {
		if room.ID == id {
			return true
		}
	}
	return false
}

func ValidateStoryDocument(document StoryDocument) error {
	if len(document.Arcs) == 0 {
		return fmt.Errorf("story document must include at least one arc")
	}

	seen := map[string]bool{}
	for _, arc := range document.Arcs {
		if strings.TrimSpace(arc.ID) == "" {
			return fmt.Errorf("story arc id is required")
		}
		if seen[arc.ID] {
			return fmt.Errorf("duplicate story arc id %q", arc.ID)
		}
		seen[arc.ID] = true
		if strings.TrimSpace(arc.Title) == "" {
			return fmt.Errorf("story arc %q title is required", arc.ID)
		}
		if arc.Kind != "main" && arc.Kind != "side" {
			return fmt.Errorf("story arc %q kind must be main or side", arc.ID)
		}
		if len(arc.LoreBeats) == 0 {
			return fmt.Errorf("story arc %q must cite at least one lore beat", arc.ID)
		}
		if len(arc.SourceIDs) == 0 {
			return fmt.Errorf("story arc %q must cite at least one source id", arc.ID)
		}
		if strings.TrimSpace(arc.Summary) == "" {
			return fmt.Errorf("story arc %q summary is required", arc.ID)
		}
		if strings.TrimSpace(arc.OriginalHook) == "" {
			return fmt.Errorf("story arc %q original_hook is required", arc.ID)
		}
		if len(arc.Steps) == 0 {
			return fmt.Errorf("story arc %q must include at least one step", arc.ID)
		}
		for faction := range arc.RequiredFactions {
			if strings.TrimSpace(faction) == "" {
				return fmt.Errorf("story arc %q has empty required faction", arc.ID)
			}
		}
		for _, step := range arc.Steps {
			if strings.TrimSpace(step.ID) == "" {
				return fmt.Errorf("story arc %q has step with empty id", arc.ID)
			}
			if strings.TrimSpace(step.Title) == "" {
				return fmt.Errorf("story arc %q step %q title is required", arc.ID, step.ID)
			}
			if strings.TrimSpace(step.Objective) == "" {
				return fmt.Errorf("story arc %q step %q objective is required", arc.ID, step.ID)
			}
		}
	}

	return nil
}
