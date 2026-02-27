package anthropic

// MessagesRequest is the request body for Anthropic Messages API.
type MessagesRequest struct {
	Model      string    `json:"model"`
	MaxTokens  int       `json:"max_tokens"`
	System     string    `json:"system,omitempty"`
	Messages   []Message `json:"messages"`
}

// Message represents a single message (user or assistant).
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// MessagesResponse is the response from Anthropic Messages API.
type MessagesResponse struct {
	ID         string        `json:"id,omitempty"`
	Type       string        `json:"type,omitempty"`
	Role       string        `json:"role,omitempty"`
	Content    []ContentBlock `json:"content,omitempty"`
	StopReason string        `json:"stop_reason,omitempty"`
	Usage      *Usage        `json:"usage,omitempty"`
}

// ContentBlock is a single content block (e.g. text).
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Usage holds token usage stats.
type Usage struct {
	InputTokens  int `json:"input_tokens,omitempty"`
	OutputTokens int `json:"output_tokens,omitempty"`
}
