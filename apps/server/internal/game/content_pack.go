package game

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
)

type ContentPack struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	MythRegion   string            `json:"myth_region"`
	Tags         []string          `json:"tags"`
	RoomsFile    string            `json:"rooms_file"`
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

type StoryArc struct {
	ID            string      `json:"id"`
	Title         string      `json:"title"`
	Kind          string      `json:"kind"`
	LoreBeats     []string    `json:"lore_beats"`
	SourceIDs     []string    `json:"source_ids"`
	Summary       string      `json:"summary"`
	OriginalHook  string      `json:"original_hook"`
	RequiredTags  []string    `json:"required_tags,omitempty"`
	AddsTags      []string    `json:"adds_tags,omitempty"`
	Steps         []StoryStep `json:"steps"`
	VariationTags []string    `json:"variation_tags,omitempty"`
}

type StoryStep struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	RoomHint    string   `json:"room_hint,omitempty"`
	Objective   string   `json:"objective"`
	Commands    []string `json:"commands,omitempty"`
	OutcomeTags []string `json:"outcome_tags,omitempty"`
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
