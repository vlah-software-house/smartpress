# Image Provider Selection & Gemini Native API Fix

**Date:** 2026-02-25
**Branch:** feat/image-provider-selection

## Summary

Removed hardcoded OpenAI preference for image generation. Users can now
select which image provider to use (OpenAI DALL-E or Google Gemini) from
both the Featured Image AI section and the Media Picker AI tab.

Also fixed Gemini image generation by switching from the non-functional
OpenAI-compatible endpoint to Gemini's native generateContent API.

## Changes

### User-selectable Image Provider
- `internal/ai/image.go` — `GenerateImage` now accepts a `provider` parameter;
  resolution: explicit → active provider → any fallback. Added `ImageProviders()`
  method that returns names of image-capable providers.
- `internal/handlers/admin_ai.go` — `AIGenerateImage` reads `image_provider` form
  value and passes it to the registry. New `AIImageProviders` endpoint returns JSON
  list of image-capable providers for frontend selectors.
- `internal/router/router.go` — Added `GET /admin/ai/image-providers` route.
- `content_form.html` — Added provider selector dropdowns to both the Featured
  Image AI section (Alpine `featuredImageAI` component) and the Media Picker AI
  tab (extended `mediaPicker` component). Selectors auto-hide when only one
  provider supports images.

### Gemini Native Image API
- `internal/ai/gemini.go` — Replaced OpenAI-compatible endpoint
  (`/v1beta/openai/images/generations`) with native generateContent API
  (`/v1beta/models/{model}:generateContent`). The OpenAI-compatible endpoint
  returns 404 for all Gemini image models despite being documented.
  Native API uses `responseModalities: ["IMAGE", "TEXT"]` and returns image
  data as `inlineData` parts with base64-encoded bytes.

## Testing
- Verified OpenAI and Gemini both appear in provider selector
- Generated image with Gemini (gemini-2.5-flash-image) — sushi plate — success
- Image saved to S3, thumbnail created, media record created
- Alt text auto-populated from prompt
- Descriptive filename generated from prompt
