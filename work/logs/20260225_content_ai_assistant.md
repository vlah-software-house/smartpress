# Step 9: Content AI Assistant

**Date:** 2026-02-25
**Branch:** feat/ai-integration
**Status:** Complete

## What was built

### Content Editor Form (`content_form.html`)
- **Missing template created** — Referenced by 6 handlers but never existed. Now a full content editor with:
  - Title, slug (auto-generated if empty), body textarea, excerpt
  - SEO metadata fields (meta description, meta keywords)
  - Status selector (draft/published)
  - HTMX-powered form submission (PUT for edit, POST for create)
  - Responsive two-column layout with collapsible AI assistant panel

### AI Assistant Panel (right sidebar)
- **Toggle button** — Fixed-position "AI Assistant" button when panel is closed
- **AlpineJS transitions** — Smooth slide-in/out animation
- **5 AI actions**, each with HTMX-driven request/response:
  1. **Suggest Titles** — Generates 5 clickable title options (click to fill title field)
  2. **Generate Excerpt** — Creates 1-2 sentence summary with "Apply to Excerpt" button
  3. **SEO Metadata** — Generates meta description + keywords, each with "Apply" button
  4. **Rewrite Content** — Rewrites body in selected tone (professional, casual, formal, persuasive, concise)
  5. **Extract Tags** — Returns clickable tag pills that append to meta keywords field
- **Loading indicators** — Spinner per action using `htmx-indicator`
- **Error handling** — Red error messages for missing content or AI failures

### AI Handler Endpoints (`admin_ai.go`)
- `POST /admin/ai/suggest-title` — Title suggestions from content body
- `POST /admin/ai/generate-excerpt` — Excerpt generation
- `POST /admin/ai/seo-metadata` — SEO description + keywords (structured parsing)
- `POST /admin/ai/rewrite` — Tone-based content rewriting (5 tone options)
- `POST /admin/ai/extract-tags` — Tag extraction as clickable pills

### Response Parsing Helpers
- `parseNumberedList()` — Extracts items from "1. ", "1) ", "- ", "* " formats, strips quotes
- `parseSEOResult()` — Extracts "DESCRIPTION:" and "KEYWORDS:" from structured AI output
- `parseTags()` — Splits comma-separated tags, strips bullets/quotes/dashes
- `quoteJSString()` — XSS-safe JS string literal for HTML attributes (escapes `"`, `<`, `>`, `&`)
- `truncate()` — Content length limiter for prompt context
- `writeAIError()` / `writeAIResult()` — Error and plain-text HTML fragment helpers

### Tests (`admin_ai_test.go`)
- `TestParseNumberedList` — 6 subtests (dot, paren, dash, quotes, empty lines, no prefix)
- `TestParseSEOResult` — 5 subtests (standard, lowercase, meta prefixes, whitespace, no match)
- `TestParseTags` — 5 subtests (comma, quotes, dashes, whitespace, empty items)
- `TestTruncate` — Short and truncated cases
- `TestQuoteJSString` — 7 subtests (simple, single quote, double quote, HTML tags, newline, backslash, ampersand)

## Design decisions
- **HTML fragments over JSON** — Each AI endpoint returns ready-to-swap HTML, keeping the frontend logic minimal. HTMX replaces the result div directly.
- **"Apply" button pattern** — AI suggestions are non-destructive. Users see the result first, then explicitly click to populate form fields. No auto-fill.
- **System prompts per action** — Each endpoint crafts a task-specific system prompt that constrains the LLM output format for reliable parsing.
- **XSS prevention** — `quoteJSString` escapes `"`, `<`, `>`, `&` using JS hex escapes (`\x22`, `\x3c`, etc.) since the output lives in HTML onclick attributes. The display text uses `html.EscapeString`.
- **Prompt truncation** — Body content is truncated to 2000-3000 chars before sending to the LLM to control token costs while preserving enough context.

## Verified
- Content form renders in new post, new page, and edit modes
- All 5 AI endpoints return correct HTML fragments
- Title suggestions render as clickable buttons
- Excerpt generates with "Apply" button
- SEO metadata correctly parses description and keywords
- Rewrite produces tone-adjusted content
- Tag extraction returns clickable pill elements
- 18/18 unit tests pass
- Tested with Gemini provider (active)
