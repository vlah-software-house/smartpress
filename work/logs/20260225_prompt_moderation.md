# Prompt Safety Moderation

**Date:** 2026-02-25
**Branch:** feat/featured-image
**Commit:** b63c8fa

## Changes

### AI Moderation Layer (`internal/ai/moderation.go`)
- `Moderator` interface with `CheckSafety(ctx, text) (*ModerationResult, error)`
- `ModerationResult` struct: `Safe bool`, `Categories []string`
- OpenAI moderator using free `/v1/moderations` endpoint with `omni-moderation-latest` model (13 categories)
- Mistral moderator using `/v1/moderations` with `mistral-moderation-latest` model (9 categories, paid)
- Human-readable category formatting (e.g., "hate/threatening" -> "hate (threatening)")

### Registry Integration (`internal/ai/provider.go`)
- Added `moderator Moderator` field to Registry
- `NewRegistry` auto-configures moderator: OpenAI preferred (free), Mistral fallback
- Added `CheckPrompt(ctx, prompt)` method with graceful degradation (fail-open if no moderator configured)

### Handler Integration (`internal/handlers/admin_ai.go`)
- Added `checkPromptSafety` helper that runs moderation and writes error if flagged
- All 8 AI handlers now check prompts before generation:
  1. `AIGenerateContent` — checks `ai_content_prompt`
  2. `AIGenerateImage` — checks `ai_image_prompt`
  3. `AISuggestTitle` — checks `title + body`
  4. `AIGenerateExcerpt` — checks `body`
  5. `AISEOMetadata` — checks `title + body`
  6. `AIRewrite` — checks `body`
  7. `AIExtractTags` — checks `title + body`
  8. `AITemplateGenerate` — checks `prompt`

### Design Decisions
- OpenAI moderation is free and works regardless of which provider is active for generation
- Graceful degradation: if moderation API fails or no moderator configured, prompts pass through (providers have built-in safety filters)
- Flagged prompts return human-readable category names asking the user to reformulate
