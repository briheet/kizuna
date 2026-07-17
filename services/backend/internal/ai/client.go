package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/briheet/kizuna/backend/internal/config"
)

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		apiKey:     cfg.AI.APIKey,
		baseURL:    strings.TrimRight(cfg.AI.BaseURL, "/"),
		httpClient: &http.Client{Timeout: 45 * time.Second},
	}
}

func (c *Client) PostJSON(ctx context.Context, path string, input, output any) error {
	body, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("encode AI request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create AI request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("AI request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("AI request failed: %s: %s", resp.Status, strings.TrimSpace(string(detail)))
	}

	if err := json.NewDecoder(resp.Body).Decode(output); err != nil {
		return fmt.Errorf("decode AI response: %w", err)
	}

	return nil
}
