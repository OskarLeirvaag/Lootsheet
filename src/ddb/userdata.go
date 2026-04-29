package ddb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

const userDataURL = "https://www.dndbeyond.com/mobile/api/v6/user-data"

// GetUserData verifies the cobalt cookie by fetching the user's identity.
// Used during compendium initialization to confirm authentication and surface
// the user's display name in the status message.
func (c *Client) GetUserData(ctx context.Context, cobalt string) (*UserData, error) {
	if cobalt == "" {
		return nil, errors.New("ddb user-data: cobalt is required")
	}

	form := url.Values{}
	form.Set("token", cobalt)

	body, err := c.doFormPost(ctx, userDataURL, form)
	if err != nil {
		return nil, fmt.Errorf("ddb user-data: %w", err)
	}

	var data UserData
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("ddb user-data: decode: %w", err)
	}
	if data.Status != "success" {
		return nil, fmt.Errorf("ddb user-data: status %q", data.Status)
	}
	return &data, nil
}
