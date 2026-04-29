package ddb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const userCampaignsURL = "https://www.dndbeyond.com/api/campaign/stt/user-campaigns"

// GetUserCampaigns returns the cobalt user's DDB campaigns. Each campaign's
// `ID` can be passed as `campaignId=` to spell/item fetches to surface
// content shared into that campaign.
//
// Auth: requires the bearer token from Authenticate plus the cobalt cookie —
// matching the auth header set used by ddb-proxy/campaign.js.
func (c *Client) GetUserCampaigns(ctx context.Context, cobalt string) ([]UserCampaign, error) {
	if !c.IsAuthenticated() {
		return nil, errors.New("ddb user-campaigns: not authenticated; call Authenticate first")
	}
	if cobalt == "" {
		return nil, errors.New("ddb user-campaigns: cobalt is required")
	}

	startedAt := time.Now()
	slog.InfoContext(ctx, "ddb get started", slog.String("url", userCampaignsURL))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userCampaignsURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cookie", "CobaltSession="+cobalt)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "ddb get failed", slog.String("url", userCampaignsURL), slog.String("error", err.Error()), slog.Duration("duration", time.Since(startedAt)))
		return nil, fmt.Errorf("ddb user-campaigns: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return nil, fmt.Errorf("ddb user-campaigns: read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ddb user-campaigns: HTTP %d: %s", resp.StatusCode, truncate(string(body), truncateLen))
	}

	slog.InfoContext(ctx, "ddb get completed",
		slog.String("url", userCampaignsURL),
		slog.Int("status", resp.StatusCode),
		slog.Int("bytes", len(body)),
		slog.Duration("duration", time.Since(startedAt)),
	)

	var resp2 userCampaignsResponse
	if err := json.Unmarshal(body, &resp2); err != nil {
		return nil, fmt.Errorf("ddb user-campaigns: decode: %w", err)
	}
	if resp2.Status != "success" {
		return nil, fmt.Errorf("ddb user-campaigns: status %q", resp2.Status)
	}
	return resp2.Data, nil
}
