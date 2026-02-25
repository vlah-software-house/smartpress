# Step 8: LLM API Integration

**Date:** 2026-02-25
**Branch:** feat/ai-integration
**Status:** Complete

## What was built

### AI Provider Abstraction (`internal/ai/`)
- **`provider.go`** — `Provider` interface with `Generate(ctx, systemPrompt, userPrompt) (string, error)` and `Name()`. Registry pattern for managing multiple providers with runtime switching.
- **`openai.go`** — OpenAI chat completions via raw HTTP (`POST /v1/chat/completions`). Bearer token auth.
- **`claude.go`** — Anthropic Messages API via raw HTTP (`POST /v1/messages`). `x-api-key` header auth.
- **`gemini.go`** — Google Gemini REST API (`POST /v1beta/models/{model}:generateContent`). `x-goog-api-key` header auth. Supports `system_instruction`.
- **`mistral.go`** — Mistral API, reuses OpenAI-compatible format with different base URL (`https://api.mistral.ai/v1`).

### Registry (`ai.Registry`)
- Factory creates providers only for configs with non-empty API keys
- `SetActive(name)` for runtime provider switching
- `Available()` lists providers with valid keys
- `HasProvider(name)` checks availability
- Threaded through admin handlers for future AI features

### Wiring
- `main.go` creates `ai.Registry` with all four provider configs from `config.Config`
- `Admin` handler struct receives `aiRegistry` for use in Steps 9-10
- Logs available providers on startup

### Tests (`provider_test.go`)
- `TestRegistryBasics` — unit test: active selection, switching, key filtering
- `TestGeminiLive`, `TestClaudeLive`, `TestOpenAILive`, `TestMistralLive` — live API integration tests (skipped if key not set)

## Design decisions
- **Raw HTTP over SDKs**: each provider is ~80 lines of self-contained HTTP code, no external dependencies beyond `net/http` + `encoding/json`
- **Mistral reuses OpenAI internals**: both use the OpenAI chat completions format, so `mistralProvider` wraps `openAIProvider.doChat()` with a different base URL
- **60-second HTTP timeout**: LLM responses can take 10-30 seconds for complex prompts

## Verified
- Claude: Pass ("2+2 equals 4.")
- OpenAI: Pass with gpt-4o ("2+2 equals 4.") — gpt-5.2 not available for project
- Mistral: Pass ("2+2 equals 4.")
- Gemini: Quota exceeded on free tier — HTTP layer works correctly
- Registry unit tests: Pass
