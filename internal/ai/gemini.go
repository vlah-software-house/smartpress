// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// geminiProvider implements the Provider interface using the Google
// Gemini REST API (POST /v1beta/models/{model}:generateContent).
type geminiProvider struct {
	config ProviderConfig
	client *http.Client
}

// newGemini creates a new Google Gemini provider.
func newGemini(cfg ProviderConfig) *geminiProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://generativelanguage.googleapis.com"
	}
	return &geminiProvider{
		config: cfg,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (p *geminiProvider) Name() string { return "gemini" }

// Generate sends a generateContent request to the Gemini API using the default model.
func (p *geminiProvider) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return p.GenerateWithModel(ctx, "", systemPrompt, userPrompt)
}

// GenerateWithModel sends a generateContent request using a specific model.
// If model is empty, the provider's default model is used.
func (p *geminiProvider) GenerateWithModel(ctx context.Context, model, systemPrompt, userPrompt string) (string, error) {
	if model == "" {
		model = p.config.Model
	}

	body := geminiRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{{Text: systemPrompt}},
		},
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: userPrompt}}},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("gemini marshal: %w", err)
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent",
		p.config.BaseURL, model)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("gemini request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini http: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("gemini read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result geminiResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("gemini unmarshal: %w", err)
	}

	if len(result.Candidates) == 0 {
		return "", fmt.Errorf("gemini: no candidates returned")
	}

	// Extract text from the first candidate's parts.
	for _, part := range result.Candidates[0].Content.Parts {
		if part.Text != "" {
			return part.Text, nil
		}
	}

	return "", fmt.Errorf("gemini: no text in response")
}

// GenerateImage creates an image using Gemini's OpenAI-compatible image
// generation endpoint. Uses ModelImage from config (e.g., "gemini-2.5-flash-image").
// Returns PNG image bytes and the content type.
func (p *geminiProvider) GenerateImage(ctx context.Context, prompt string) ([]byte, string, error) {
	model := p.config.ModelImage
	if model == "" {
		return nil, "", fmt.Errorf("gemini: image generation requires GEMINI_MODEL_IMAGE to be set")
	}

	body := openAIImageRequest{
		Model:          model,
		Prompt:         prompt,
		N:              1,
		ResponseFormat: "b64_json",
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, "", fmt.Errorf("gemini image marshal: %w", err)
	}

	url := p.config.BaseURL + "/v1beta/openai/images/generations"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, "", fmt.Errorf("gemini image request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	imgClient := &http.Client{Timeout: 120 * time.Second}
	resp, err := imgClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("gemini image http: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("gemini image read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("gemini image API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result openAIImageResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, "", fmt.Errorf("gemini image unmarshal: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, "", fmt.Errorf("gemini image: no images returned")
	}

	imgBytes, err := base64.StdEncoding.DecodeString(result.Data[0].B64JSON)
	if err != nil {
		return nil, "", fmt.Errorf("gemini image decode base64: %w", err)
	}

	return imgBytes, "image/png", nil
}

// --- Gemini API types ---

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	SystemInstruction *geminiContent  `json:"system_instruction,omitempty"`
	Contents          []geminiContent `json:"contents"`
}

type geminiCandidate struct {
	Content geminiContent `json:"content"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}
