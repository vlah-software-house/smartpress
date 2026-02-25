// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

package ai

import (
	"context"
	"fmt"
)

// ImageGenerator is an optional interface that AI providers can implement
// to support image generation. Not all providers have this capability
// (e.g., Claude and Mistral are text-only).
type ImageGenerator interface {
	// GenerateImage creates an image from a text prompt. Returns the raw
	// image bytes and the MIME content type (e.g., "image/png").
	GenerateImage(ctx context.Context, prompt string) ([]byte, string, error)
}

// GenerateImage finds the best available image generator and creates an image.
// It prefers OpenAI (DALL-E) regardless of the active text provider, then
// falls back to any other provider that implements ImageGenerator.
func (r *Registry) GenerateImage(ctx context.Context, prompt string) ([]byte, string, error) {
	ig := r.findImageGenerator()
	if ig == nil {
		return nil, "", fmt.Errorf("ai: no provider supports image generation (OpenAI key required for DALL-E)")
	}
	return ig.GenerateImage(ctx, prompt)
}

// SupportsImageGeneration returns true if any registered provider can generate images.
func (r *Registry) SupportsImageGeneration() bool {
	return r.findImageGenerator() != nil
}

// findImageGenerator returns the best available ImageGenerator, preferring
// OpenAI (DALL-E). Falls back to any other provider that implements the interface.
func (r *Registry) findImageGenerator() ImageGenerator {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Prefer OpenAI for image generation (DALL-E).
	if p, ok := r.providers["openai"]; ok {
		if ig, ok := p.(ImageGenerator); ok {
			return ig
		}
	}

	// Fallback: check all providers.
	for _, p := range r.providers {
		if ig, ok := p.(ImageGenerator); ok {
			return ig
		}
	}

	return nil
}
