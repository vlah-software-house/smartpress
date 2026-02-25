# Prompt Safety Moderation

**Date:** 2026-02-25
**Branch:** feat/featured-image
**Commits:** b63c8fa, d5a37e8

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

### Fallback Moderator (`internal/ai/moderation.go` — d5a37e8)
- `fallbackModerator` wraps primary (OpenAI) + secondary (Mistral) moderators
- Uses `atomic.Bool` for lock-free primary/secondary switching
- On 401/403 from primary: permanently switches to secondary for all future calls
- On transient errors: tries secondary once without permanent switch
- Fixes issue where OpenAI `sk-proj-*` (project-scoped) keys return 403 on moderation endpoint

### Design Decisions
- OpenAI moderation is free and works regardless of which provider is active for generation
- Fallback moderator ensures moderation works even with restricted API keys
- Graceful degradation: if both moderators fail or none configured, prompts pass through (providers have built-in safety filters)
- Flagged prompts return human-readable category names asking the user to reformulate
