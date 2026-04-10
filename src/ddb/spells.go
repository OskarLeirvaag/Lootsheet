package ddb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const spellsURL = "https://character-service.dndbeyond.com/character/v5/game-data/spells"

// SpellComponentNames maps component IDs to display strings.
var SpellComponentNames = map[int]string{1: "V", 2: "S", 3: "M"}

// ClassNames maps standard class IDs to display names.
var ClassNames = map[int]string{
	1: "Barbarian", 2: "Bard", 3: "Cleric", 4: "Druid", 5: "Fighter",
	6: "Monk", 7: "Paladin", 8: "Wizard", 9: "Sorcerer", 10: "Warlock",
	11: "Ranger", 12: "Artificer",
}

// FetchSpellsResult contains spells and their class associations.
type FetchSpellsResult struct {
	Spells       []RawSpellEntry
	SpellClasses map[int][]string // spell definition ID → class names
}

// FetchSpells retrieves all spells from DDB by iterating over class IDs.
// Requires authentication. Tracks which classes each spell belongs to.
func (c *Client) FetchSpells(ctx context.Context, classIDs []int) (*FetchSpellsResult, error) {
	if !c.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}

	seen := make(map[int]bool)
	spellClasses := make(map[int][]string)
	var all []RawSpellEntry

	for _, classID := range classIDs {
		className := ClassNames[classID]
		url := fmt.Sprintf("%s?classId=%d&classLevel=20&sharingSetting=2", spellsURL, classID)

		body, err := c.doGet(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("ddb spells (class=%d): %w", classID, err)
		}

		var resp SpellResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("ddb spells: decode: %w", err)
		}

		// Preserve raw JSON per spell.
		var rawResp struct {
			Data []json.RawMessage `json:"data"`
		}
		_ = json.Unmarshal(body, &rawResp)

		for i := range resp.Data {
			defID := resp.Data[i].Definition.ID
			if className != "" {
				spellClasses[defID] = append(spellClasses[defID], className)
			}
			if seen[defID] {
				continue // deduplicate across classes
			}
			seen[defID] = true
			if i < len(rawResp.Data) {
				resp.Data[i].Definition.RawJSON = rawResp.Data[i]
			}
			all = append(all, resp.Data[i])
		}
	}

	return &FetchSpellsResult{Spells: all, SpellClasses: spellClasses}, nil
}

// AllClassIDs returns the standard D&D class IDs for spell fetching.
func AllClassIDs() []int {
	return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
}

// FormatSpellComponents returns a display string like "V, S, M (a drop of mercury)".
func FormatSpellComponents(def *RawSpellDef) string {
	var parts []string
	for _, id := range def.Components {
		if name, ok := SpellComponentNames[id]; ok {
			parts = append(parts, name)
		}
	}
	result := strings.Join(parts, ", ")
	desc := strings.TrimSpace(def.ComponentsDescription)
	if desc != "" {
		result += " (" + desc + ")"
	}
	return result
}

// FormatSpellDuration returns a display string like "1 Hour" or "Concentration, 1 Minute".
func FormatSpellDuration(def *RawSpellDef) string {
	d := def.Duration
	if d.DurationType == "" {
		return "Instantaneous"
	}
	dur := ""
	if d.DurationInterval > 0 && d.DurationUnit != "" {
		dur = fmt.Sprintf("%d %s", d.DurationInterval, d.DurationUnit)
	} else if d.DurationType != "" {
		dur = d.DurationType
	}
	if def.Concentration {
		return "Concentration, " + dur
	}
	return dur
}

// FormatSpellRange returns a display string like "30 ft" or "Self".
func FormatSpellRange(def *RawSpellDef) string {
	r := def.Range
	if r.RangeValue > 0 {
		return fmt.Sprintf("%d ft", r.RangeValue)
	}
	if r.Origin != "" {
		return r.Origin
	}
	return ""
}
