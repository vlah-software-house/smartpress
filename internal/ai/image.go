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

// GenerateImage creates an image using the specified provider. If provider
// is empty, uses the active text provider when it supports images, otherwise
// falls back to any available image-capable provider.
func (r *Registry) GenerateImage(ctx context.Context, provider, prompt string) ([]byte, string, error) {
	ig, err := r.resolveImageGenerator(provider)
	if err != nil {
		return nil, "", err
	}
	return ig.GenerateImage(ctx, prompt)
}

// SupportsImageGeneration returns true if any registered provider can generate images.
func (r *Registry) SupportsImageGeneration() bool {
	return len(r.ImageProviders()) > 0
}

// ImageProviders returns the names of all registered providers that support
// image generation. Used by the UI to populate provider selectors.
func (r *Registry) ImageProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var names []string
	for name, p := range r.providers {
		if _, ok := p.(ImageGenerator); ok {
			names = append(names, name)
		}
	}
	return names
}

// resolveImageGenerator finds the ImageGenerator for the requested provider.
// If provider is empty, prefers the active text provider, then falls back
// to any available image-capable provider.
func (r *Registry) resolveImageGenerator(provider string) (ImageGenerator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Explicit provider requested.
	if provider != "" {
		p, ok := r.providers[provider]
		if !ok {
			return nil, fmt.Errorf("ai: provider %q is not available", provider)
		}
		ig, ok := p.(ImageGenerator)
		if !ok {
			return nil, fmt.Errorf("ai: provider %q does not support image generation", provider)
		}
		return ig, nil
	}

	// No provider specified — try the active text provider first.
	if p, ok := r.providers[r.active]; ok {
		if ig, ok := p.(ImageGenerator); ok {
			return ig, nil
		}
	}

	// Fallback: any provider that supports images.
	for _, p := range r.providers {
		if ig, ok := p.(ImageGenerator); ok {
			return ig, nil
		}
	}

	return nil, fmt.Errorf("ai: no provider supports image generation (requires OpenAI key for DALL-E or Gemini key with GEMINI_MODEL_IMAGE set)")
}
