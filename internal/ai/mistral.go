// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

package ai

import (
	"context"
	"net/http"
	"time"
)

// mistralProvider implements the Provider interface using Mistral's
// chat completions API, which is OpenAI-compatible.
type mistralProvider struct {
	inner *openAIProvider
}

// newMistral creates a new Mistral provider. Mistral uses an
// OpenAI-compatible API at a different base URL.
func newMistral(cfg ProviderConfig) *mistralProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.mistral.ai/v1"
	}
	return &mistralProvider{
		inner: &openAIProvider{
			config: cfg,
			client: &http.Client{Timeout: 60 * time.Second},
		},
	}
}

func (p *mistralProvider) Name() string { return "mistral" }

// Generate sends a chat completion request to Mistral's API.
func (p *mistralProvider) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	messages := []openAIMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	body := openAIRequest{
		Model:    p.inner.config.Model,
		Messages: messages,
	}

	return p.inner.doChat(ctx, body)
}
