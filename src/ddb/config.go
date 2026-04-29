package ddb

import (
	"context"
	"encoding/json"
	"fmt"
)

// FetchConfig retrieves the game configuration (sources, conditions, rules,
// lookup tables). No authentication required.
func (c *Client) FetchConfig(ctx context.Context) (*Config, error) {
	body, err := c.doGet(ctx, configURL)
	if err != nil {
		return nil, fmt.Errorf("ddb config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(body, &cfg); err != nil {
		return nil, fmt.Errorf("ddb config: decode: %w", err)
	}

	return &cfg, nil
}

// ChallengeRatingLabel returns a display string (e.g. "1/4") for a CR ID.
func (cfg *Config) ChallengeRatingLabel(id int) string {
	for _, cr := range cfg.ChallengeRatings {
		if cr.ID == id {
			return formatCR(cr.Value)
		}
	}
	return "?"
}

// CreatureSizeName returns the size name for a size ID.
func (cfg *Config) CreatureSizeName(id int) string {
	for _, s := range cfg.CreatureSizes {
		if s.ID == id {
			return s.Name
		}
	}
	return ""
}

// MonsterTypeName returns the type name for a type ID.
func (cfg *Config) MonsterTypeName(id int) string {
	for _, t := range cfg.MonsterTypes {
		if t.ID == id {
			return t.Name
		}
	}
	return ""
}

// SourceName returns the source description for a source ID.
func (cfg *Config) SourceName(id int) string {
	for _, s := range cfg.Sources {
		if s.ID == id {
			return s.Description
		}
	}
	return ""
}

// CR fractional values used by DDB.
const (
	crOneEighth = 0.125
	crOneQuarter = 0.25
	crOneHalf   = 0.5
)

func formatCR(v float64) string {
	switch v {
	case crOneEighth:
		return "1/8"
	case crOneQuarter:
		return "1/4"
	case crOneHalf:
		return "1/2"
	default:
		if v == float64(int(v)) {
			return fmt.Sprintf("%d", int(v))
		}
		return fmt.Sprintf("%.1f", v)
	}
}
