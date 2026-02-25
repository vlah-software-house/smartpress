# Phase 3: The AI "Brain" (MVP 2)

## Step 8: LLM API Integration
- [x] Abstraction layer supporting multiple AI providers (OpenAI, Gemini, Claude, Mistral)
- [x] Configuration via .secrets (per-provider keys, models, optional base URLs)
- [x] Registry with runtime switching
- [x] Live API tests for all 4 providers

## Step 9: Content AI Assistant
- [x] AI Assistant sidebar/modal in editor
- [x] Rewrite tone, generate titles, SEO metadata, tag extraction
- [x] HTMX-driven responses (HTML fragment swaps)

## Step 10: AI Theme Builder
- [ ] Chat UI in admin "AI Design" section
- [ ] Prompt → LLM generates HTML+TailwindCSS with Go template vars
- [ ] Validation: parse HTML, compile as Go template, catch errors
- [ ] Live preview with dummy data
- [ ] Save to DB + cache invalidation
