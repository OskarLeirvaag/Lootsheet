package ddb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const monsterURL = "https://monster-service.dndbeyond.com/v1/Monster"

// FetchMonsters retrieves all monsters from DDB, paginated 100 at a time.
// Requires authentication. If sourceIDs is non-empty, only monsters from
// those sources are returned.
func (c *Client) FetchMonsters(ctx context.Context, sourceIDs []int) ([]RawMonster, error) {
	if !c.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}

	var all []RawMonster
	skip := 0
	take := 100

	for {
		var b strings.Builder
		// strings.Builder.Write never errors; discard Fprintf return values.
		_, _ = fmt.Fprintf(&b, "%s?skip=%d&take=%d", monsterURL, skip, take)
		for _, id := range sourceIDs {
			_, _ = fmt.Fprintf(&b, "&sources=%d", id)
		}
		url := b.String()

		body, err := c.doGet(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("ddb monsters (skip=%d): %w", skip, err)
		}

		var resp MonsterResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("ddb monsters: decode: %w", err)
		}

		// Preserve raw JSON per monster for detail_json storage.
		var rawResp struct {
			Data []json.RawMessage `json:"data"`
		}
		_ = json.Unmarshal(body, &rawResp)

		for i := range resp.Data {
			if i < len(rawResp.Data) {
				resp.Data[i].RawJSON = rawResp.Data[i]
			}
			all = append(all, resp.Data[i])
		}

		if len(resp.Data) < take {
			break // last page
		}
		skip += take
	}

	return all, nil
}

// FormatMonsterHP returns a display string like "7 (2d6)".
func FormatMonsterHP(m *RawMonster) string {
	hp := fmt.Sprintf("%d", m.AverageHitPoints)
	if m.HitPointDice.DiceString != "" {
		hp += " (" + m.HitPointDice.DiceString + ")"
	}
	return hp
}

// FormatMonsterAC returns a display string like "15 (leather armor, shield)".
func FormatMonsterAC(m *RawMonster) string {
	ac := fmt.Sprintf("%d", m.ArmorClass)
	desc := strings.TrimSpace(m.ArmorClassDescription)
	if desc != "" {
		ac += " " + desc
	}
	return ac
}
