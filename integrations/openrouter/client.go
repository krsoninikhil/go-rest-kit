package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

const defaultBaseURL = "https://openrouter.ai"

// Config holds OpenRouter client configuration.
type Config struct {
	APIKey  string `json:"-"` // sensitive; omit from logs
	Model   string // e.g. "openai/gpt-4o-mini"
	BaseURL string // optional; default https://openrouter.ai
}

// Client calls the OpenRouter chat completions API.
type Client struct {
	config Config
	http   *http.Client
}

// NewClient creates an OpenRouter client. If BaseURL is empty, default is used.
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

// Chat sends system and user messages to OpenRouter and returns the assistant's content.
// It uses a single non-streaming request. ctx is used for cancellation.
func (c *Client) Chat(ctx context.Context, system, user string) (content string, err error) {
	if c.config.APIKey == "" {
		return "", errors.New("openrouter: API key is required")
	}
	if c.config.Model == "" {
		return "", errors.New("openrouter: model is required")
	}

	messages := []Message{
		{Role: "system", Content: system},
		{Role: "user", Content: user},
	}
	body := ChatRequest{Model: c.config.Model, Messages: messages}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return "", errors.Wrap(err, "openrouter: encode request")
	}

	url := c.config.BaseURL + "/api/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return "", errors.Wrap(err, "openrouter: create request")
	}
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "openrouter: request failed")
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", errors.Wrap(
			fmt.Errorf("openrouter: HTTP %d: %s", resp.StatusCode, string(respBody)),
			"openrouter",
		)
	}

	var out ChatResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", errors.Wrap(err, "openrouter: decode response")
	}
	if len(out.Choices) == 0 {
		return "", errors.New("openrouter: no choices in response")
	}
	content = out.Choices[0].Message.Content
	if content == "" {
		return "", errors.New("openrouter: empty content in first choice")
	}
	return content, nil
}
