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

// GenerateImage calls the active provider's image generation if supported.
// Returns an error if the active provider does not implement ImageGenerator.
func (r *Registry) GenerateImage(ctx context.Context, prompt string) ([]byte, string, error) {
	p, err := r.Active()
	if err != nil {
		return nil, "", err
	}

	ig, ok := p.(ImageGenerator)
	if !ok {
		return nil, "", fmt.Errorf("ai: provider %q does not support image generation", p.Name())
	}

	return ig.GenerateImage(ctx, prompt)
}

// SupportsImageGeneration returns true if the active provider can generate images.
func (r *Registry) SupportsImageGeneration() bool {
	p, err := r.Active()
	if err != nil {
		return false
	}
	_, ok := p.(ImageGenerator)
	return ok
}
