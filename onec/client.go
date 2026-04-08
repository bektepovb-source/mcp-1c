package onec

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is an HTTP client for communicating with 1C:Enterprise.
type Client struct {
	BaseURL    string
	User       string
	Password   string
	HTTPClient *http.Client
}

// NewClient creates a client for 1C HTTP service.
// When user is non-empty, basic auth is added to every request.
func NewClient(baseURL, user, password string) *Client {
	return &Client{
		BaseURL:  baseURL,
		User:     user,
		Password: password,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Get performs a GET request to a 1C endpoint and decodes the JSON response.
func (c *Client) Get(ctx context.Context, endpoint string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+endpoint, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	return c.do(req, result)
}

// Post performs a POST request to a 1C endpoint with a JSON body and decodes the JSON response.
func (c *Client) Post(ctx context.Context, endpoint string, body any, result any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, result)
}

// do executes the request, checks the status, and decodes the JSON response.
func (c *Client) do(req *http.Request, result any) error {
	if c.User != "" {
		req.SetBasicAuth(c.User, c.Password)
	}
	req.Close = true // close connection after each request (avoids 1C session limit)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request to 1C: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("1C returned status %d: %s", resp.StatusCode, string(body))
	}

	// Limit response body to 50 MB to prevent OOM on unexpectedly large payloads.
	const maxResponseSize = 50 << 20 // 50 MB
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseSize)).Decode(result); err != nil {
		return fmt.Errorf("decoding 1C response: %w", err)
	}
	return nil
}
