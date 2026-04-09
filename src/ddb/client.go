package ddb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	authURL   = "https://auth-service.dndbeyond.com/v1/cobalt-token"
	configURL = "https://www.dndbeyond.com/api/config/json"
)

// Client is an HTTP client for the D&D Beyond API.
type Client struct {
	httpClient  *http.Client
	bearerToken string
}

// NewClient creates a new DDB API client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Authenticate exchanges a cobalt cookie for a bearer token.
// The bearer token is cached in memory for subsequent requests.
func (c *Client) Authenticate(ctx context.Context, cobalt string) error {
	req, err := http.NewRequestWithContext(ctx, "POST", authURL, bytes.NewReader([]byte("{}")))
	if err != nil {
		return fmt.Errorf("ddb auth: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "CobaltSession="+cobalt)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ddb auth: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ddb auth: status %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	var auth AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&auth); err != nil {
		return fmt.Errorf("ddb auth: decode: %w", err)
	}
	if auth.Token == "" {
		return fmt.Errorf("ddb auth: empty token (invalid cobalt cookie?)")
	}

	c.bearerToken = auth.Token
	return nil
}

// IsAuthenticated reports whether the client has a bearer token.
func (c *Client) IsAuthenticated() bool {
	return c.bearerToken != ""
}

// doGet performs an authenticated GET request.
func (c *Client) doGet(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	if c.bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	return body, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
