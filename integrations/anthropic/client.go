package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

const (
	defaultBaseURL          = "https://api.anthropic.com"
	defaultAnthropicVersion = "2023-06-01"
	defaultMaxTokens        = 4096
)

// Config holds Anthropic client configuration.
type Config struct {
	APIKey  string `json:"-"` // sensitive
	Model   string // e.g. claude-3-5-haiku-20241022
	BaseURL string // optional; default https://api.anthropic.com
}

// Client calls the Anthropic Messages API.
type Client struct {
	config Config
	http   *http.Client
}

// NewClient creates an Anthropic client. If BaseURL is empty, default is used.
func NewClient(config Config) *Client {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{
		config: Config{
			APIKey:  config.APIKey,
			Model:   config.Model,
			BaseURL: baseURL,
		},
		http: &http.Client{},
	}
}

// Chat sends system and user messages to Anthropic and returns the assistant's text content.
func (c *Client) Chat(ctx context.Context, system, user string) (content string, err error) {
	if c.config.APIKey == "" {
		return "", errors.New("anthropic: API key is required")
	}
	if c.config.Model == "" {
		return "", errors.New("anthropic: model is required")
	}

	body := MessagesRequest{
		Model:     c.config.Model,
		MaxTokens: defaultMaxTokens,
		System:    system,
		Messages:  []Message{{Role: "user", Content: user}},
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return "", errors.Wrap(err, "anthropic: encode request")
	}

	url := c.config.BaseURL + "/v1/messages"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return "", errors.Wrap(err, "anthropic: create request")
	}
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("anthropic-version", defaultAnthropicVersion)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "anthropic: request failed")
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", errors.Wrap(
			fmt.Errorf("anthropic: HTTP %d: %s", resp.StatusCode, string(respBody)),
			"anthropic",
		)
	}

	var out MessagesResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", errors.Wrap(err, "anthropic: decode response")
	}
	for _, block := range out.Content {
		if block.Type == "text" && block.Text != "" {
			return block.Text, nil
		}
	}
	return "", errors.New("anthropic: no text content in response")
}
