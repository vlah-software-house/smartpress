# AI Model Tiers & Gemini Image Generation

**Date:** 2026-02-25
**Branch:** feat/ai-model-tiers

## Summary

Implemented per-task AI model tiers so that lightweight tasks (title suggestions,
excerpt generation, SEO metadata, tag extraction) use cheaper/faster models while
content-heavy tasks (article generation, rewrites, template design) use pro models.
Also added Gemini image generation support via their OpenAI-compatible endpoint.

## Changes

### Model Tier System
- Added `TaskType` enum: `TaskContent`, `TaskTemplate`, `TaskLight`
- Extended `ProviderConfig` with `ModelLight`, `ModelContent`, `ModelTemplate`, `ModelImage` fields
- Added `ModelForTask()` resolution: task-specific override -> tier default -> base Model
- Added `GenerateWithModel()` to the `Provider` interface (all 4 providers implement it)
- Added `GenerateForTask()` to `Registry` for task-aware model selection
- New env vars per provider: `*_MODEL_LIGHT`, `*_MODEL_CONTENT`, `*_MODEL_TEMPLATE`, `*_MODEL_IMAGE`

### Gemini Image Generation
- Implemented `GenerateImage()` on `geminiProvider` using the OpenAI-compatible endpoint
  at `/v1beta/openai/images/generations`
- Uses `Authorization: Bearer` header (not `x-goog-api-key`) for compatibility
- Model configurable via `GEMINI_MODEL_IMAGE` env var (e.g., `gemini-2.5-flash-image`)
- Falls back gracefully — if `GEMINI_MODEL_IMAGE` not set, returns clear error

### Handler Updates
- All AI handler calls updated from `Generate()` to `GenerateForTask()` with appropriate task types
- Revision title/changelog generation uses `TaskLight`
- Content generation and rewrites use `TaskContent`
- Template generation uses `TaskTemplate`

## Task Routing

| Task | TaskType | Example Model (Gemini) |
|------|----------|----------------------|
| Suggest Title | TaskLight | gemini-2.5-flash-lite |
| Generate Excerpt | TaskLight | gemini-2.5-flash-lite |
| SEO Metadata | TaskLight | gemini-2.5-flash-lite |
| Extract Tags | TaskLight | gemini-2.5-flash-lite |
| Revision Notes | TaskLight | gemini-2.5-flash-lite |
| Generate Content | TaskContent | gemini-3.1-pro-preview |
| Rewrite Content | TaskContent | gemini-3.1-pro-preview |
| Generate Template | TaskTemplate | gemini-3.1-pro-preview |
| Generate Image | (ModelImage) | gemini-2.5-flash-image |

## Files Modified
- `internal/config/config.go` — new env var fields per provider
- `internal/ai/provider.go` — TaskType, ModelForTask, GenerateWithModel interface, GenerateForTask
- `internal/ai/openai.go` — GenerateWithModel, configurable image model
- `internal/ai/gemini.go` — GenerateWithModel, GenerateImage implementation
- `internal/ai/claude.go` — GenerateWithModel
- `internal/ai/mistral.go` — GenerateWithModel
- `internal/ai/image.go` — updated error message for Gemini
- `internal/handlers/admin_ai.go` — all Generate → GenerateForTask with task types
- `internal/handlers/admin.go` — revision notes → GenerateForTask TaskLight
- `cmd/yaaicms/main.go` — wired up all new config fields

## Testing
- Deployed to testing cluster
- Verified title suggestion works (TaskLight via Gemini)
- Verified tag extraction works (TaskLight via Gemini)
- Verified media picker with AI generation tab renders correctly
- All 4 AI providers initialized successfully
