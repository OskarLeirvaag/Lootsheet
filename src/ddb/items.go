package ddb

import (
	"context"
	"encoding/json"
	"fmt"
)

const itemsURL = "https://character-service.dndbeyond.com/character/v5/game-data/items"

// FetchItems retrieves all items from DDB. Requires authentication.
//
// campaignID, when non-zero, is appended as `&campaignId=N` so DDB also
// returns items from books shared into that campaign (in addition to books
// the user owns). Pass 0 to omit and get owned-only content.
func (c *Client) FetchItems(ctx context.Context, campaignID int) ([]RawItem, error) {
	if !c.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}

	url := itemsURL + "?sharingSetting=2"
	if campaignID > 0 {
		url += fmt.Sprintf("&campaignId=%d", campaignID)
	}

	body, err := c.doGet(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("ddb items: %w", err)
	}

	var resp ItemResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("ddb items: decode: %w", err)
	}

	// Preserve raw JSON per item.
	var rawResp struct {
		Data []json.RawMessage `json:"data"`
	}
	_ = json.Unmarshal(body, &rawResp)

	for i := range resp.Data {
		if i < len(rawResp.Data) {
			resp.Data[i].RawJSON = rawResp.Data[i]
		}
	}

	return resp.Data, nil
}

// ItemTypeName returns the best type name for an item.
func ItemTypeName(item *RawItem) string {
	if item.FilterType != "" {
		return item.FilterType
	}
	return item.Type
}
