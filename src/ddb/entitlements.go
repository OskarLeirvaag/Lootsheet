package ddb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
)

const availableUserContentURL = "https://www.dndbeyond.com/mobile/api/v6/available-user-content"

// GetAvailableUserContent returns the set of source-book IDs the cobalt token
// owns. It calls /mobile/api/v6/available-user-content (a single request) and
// filters to the books license block (EntityTypeID 496802664), keeping only
// entities with isOwned=true.
//
// This is the same endpoint used by ddb-adventure-muncher to populate its book
// picker. It is the cheapest way to learn which sources the user actually owns
// without probing each source via the monster service.
func (c *Client) GetAvailableUserContent(ctx context.Context, cobalt string) ([]int, error) {
	if cobalt == "" {
		return nil, errors.New("ddb available-user-content: cobalt is required")
	}

	form := url.Values{}
	form.Set("token", cobalt)

	body, err := c.doFormPost(ctx, availableUserContentURL, form)
	if err != nil {
		return nil, fmt.Errorf("ddb available-user-content: %w", err)
	}

	var resp AvailableUserContent
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("ddb available-user-content: decode: %w", err)
	}
	if resp.Status != "success" {
		return nil, fmt.Errorf("ddb available-user-content: status %q", resp.Status)
	}

	// Diagnostic: log the shape we observed so we can tell whether the response
	// was wrapped in `data` or top-level, and how many books DDB returned.
	books := resp.Books()
	totalEntities := 0
	for _, b := range books {
		totalEntities += len(b.Entities)
	}
	slog.InfoContext(ctx, "ddb available-user-content parsed",
		slog.Bool("wrapped_in_data", len(resp.Licenses) == 0 && len(resp.Data.Licenses) > 0),
		slog.Int("license_blocks", len(books)),
		slog.Int("total_entities", totalEntities),
		slog.Int("body_bytes", len(body)),
		slog.String("body_preview", previewBody(body)),
	)

	return filterOwnedBookIDs(resp), nil
}

// previewBody returns a short prefix of the response body for diagnostic logs.
// Long enough to see the JSON shape, short enough to keep log lines readable.
func previewBody(body []byte) string {
	const maxPreview = 256
	if len(body) <= maxPreview {
		return string(body)
	}
	return string(body[:maxPreview]) + "…"
}

// filterOwnedBookIDs returns the source IDs the user owns from an
// available-user-content payload. Only the books license block (EntityTypeID
// 496802664) is considered; dice-set and other product types are dropped.
func filterOwnedBookIDs(resp AvailableUserContent) []int {
	var ids []int
	for _, block := range resp.Books() {
		if block.EntityTypeID != EntityTypeIDBooks {
			continue
		}
		for _, e := range block.Entities {
			if e.IsOwned {
				ids = append(ids, e.ID)
			}
		}
	}
	return ids
}
