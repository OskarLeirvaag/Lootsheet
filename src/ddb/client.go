package ddb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	authURL   = "https://auth-service.dndbeyond.com/v1/cobalt-token"
	configURL = "https://www.dndbeyond.com/api/config/json"

	httpTimeout     = 30 * time.Second
	maxAuthBody     = 1 << 20   // 1 MB
	maxResponseBody = 100 << 20 // 100 MB
	truncateLen     = 200
)

// ErrNotAuthenticated is returned when an API call requires a bearer token.
var ErrNotAuthenticated = errors.New("ddb: not authenticated")

// Client is an HTTP client for the D&D Beyond API.
type Client struct {
	httpClient  *http.Client
	bearerToken string
}

// NewClient creates a new DDB API client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: httpTimeout},
	}
}

// Authenticate exchanges a cobalt cookie for a bearer token.
// The bearer token is cached in memory for subsequent requests.
func (c *Client) Authenticate(ctx context.Context, cobalt string) error {
	startedAt := time.Now()
	slog.InfoContext(ctx, "ddb auth started")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authURL, bytes.NewReader([]byte("{}")))
	if err != nil {
		return fmt.Errorf("ddb auth: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "CobaltSession="+cobalt)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "ddb auth failed", slog.String("error", err.Error()), slog.Duration("duration", time.Since(startedAt)))
		return fmt.Errorf("ddb auth: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxAuthBody))
		slog.ErrorContext(ctx, "ddb auth failed",
			slog.Int("status", resp.StatusCode),
			slog.Duration("duration", time.Since(startedAt)),
		)
		return fmt.Errorf("ddb auth: status %d: %s", resp.StatusCode, truncate(string(body), truncateLen))
	}

	var auth AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&auth); err != nil {
		return fmt.Errorf("ddb auth: decode: %w", err)
	}
	if auth.Token == "" {
		return errors.New("ddb auth: empty token (invalid cobalt cookie?)")
	}

	c.bearerToken = auth.Token
	slog.InfoContext(ctx, "ddb auth completed", slog.Duration("duration", time.Since(startedAt)))
	return nil
}

// IsAuthenticated reports whether the client has a bearer token.
func (c *Client) IsAuthenticated() bool {
	return c.bearerToken != ""
}

// doGet performs an authenticated GET request.
func (c *Client) doGet(ctx context.Context, endpoint string) ([]byte, error) {
	startedAt := time.Now()
	slog.InfoContext(ctx, "ddb get started", slog.String("url", endpoint))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if c.bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "ddb get failed",
			slog.String("url", endpoint),
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(startedAt)),
		)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		slog.ErrorContext(ctx, "ddb get failed",
			slog.String("url", endpoint),
			slog.Int("status", resp.StatusCode),
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(startedAt)),
		)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		slog.ErrorContext(ctx, "ddb get failed",
			slog.String("url", endpoint),
			slog.Int("status", resp.StatusCode),
			slog.Int("bytes", len(body)),
			slog.Duration("duration", time.Since(startedAt)),
		)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(body), truncateLen))
	}

	slog.InfoContext(ctx, "ddb get completed",
		slog.String("url", endpoint),
		slog.Int("status", resp.StatusCode),
		slog.Int("bytes", len(body)),
		slog.Duration("duration", time.Since(startedAt)),
	)
	return body, nil
}

// doFormPost performs a POST with application/x-www-form-urlencoded body.
// Used by the mobile/api/v6 endpoints (user-data, available-user-content) which
// accept the cobalt token directly in the body rather than as a Cookie.
func (c *Client) doFormPost(ctx context.Context, endpoint string, form url.Values) ([]byte, error) {
	startedAt := time.Now()
	slog.InfoContext(ctx, "ddb post started", slog.String("url", endpoint))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "ddb post failed",
			slog.String("url", endpoint),
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(startedAt)),
		)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		slog.ErrorContext(ctx, "ddb post failed",
			slog.String("url", endpoint),
			slog.Int("status", resp.StatusCode),
			slog.Int("bytes", len(body)),
			slog.Duration("duration", time.Since(startedAt)),
		)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(body), truncateLen))
	}

	slog.InfoContext(ctx, "ddb post completed",
		slog.String("url", endpoint),
		slog.Int("status", resp.StatusCode),
		slog.Int("bytes", len(body)),
		slog.Duration("duration", time.Since(startedAt)),
	)
	return body, nil
}

func truncate(s string, n int) string { //nolint:unparam // n is always truncateLen today; keep param for clarity at call sites
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
