package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// openAIProvider implements the Provider interface using the OpenAI
// chat completions API (POST /v1/chat/completions).
type openAIProvider struct {
	config ProviderConfig
	client *http.Client
}

// newOpenAI creates a new OpenAI provider.
func newOpenAI(cfg ProviderConfig) *openAIProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	return &openAIProvider{
		config: cfg,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (p *openAIProvider) Name() string { return "openai" }

// Generate sends a chat completion request to OpenAI and returns the
// assistant's response text.
func (p *openAIProvider) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	messages := []openAIMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	body := openAIRequest{
		Model:    p.config.Model,
		Messages: messages,
	}

	return p.doChat(ctx, body)
}

// doChat performs the HTTP call to the chat completions endpoint.
// Shared between OpenAI and Mistral (same API format).
func (p *openAIProvider) doChat(ctx context.Context, body openAIRequest) (string, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("openai marshal: %w", err)
	}

	url := p.config.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("openai request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai http: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("openai read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result openAIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("openai unmarshal: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("openai: no choices returned")
	}

	return result.Choices[0].Message.Content, nil
}

// --- OpenAI-compatible request/response types ---
// Used by both OpenAI and Mistral providers.

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
}

type openAIChoice struct {
	Message openAIMessage `json:"message"`
}
