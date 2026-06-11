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
