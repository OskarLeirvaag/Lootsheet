package ddb

import (
	"context"
	"encoding/json"
	"fmt"
)

const itemsURL = "https://character-service.dndbeyond.com/character/v5/game-data/items"

// FetchItems retrieves all items from DDB. Requires authentication.
func (c *Client) FetchItems(ctx context.Context) ([]RawItem, error) {
	if !c.IsAuthenticated() {
		return nil, fmt.Errorf("ddb items: not authenticated")
	}

	url := itemsURL + "?sharingSetting=2"

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
