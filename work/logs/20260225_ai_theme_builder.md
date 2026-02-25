# Step 10: AI Theme Builder

**Date:** 2026-02-25
**Branch:** feat/ai-integration
**Status:** Complete

## What was built

### AI Template Builder Chat UI (`template_ai.html`)
- **Two-column layout**: Left column has chat + generated code viewer, right column has live preview + save form
- **Template type selector**: Buttons for header, footer, page, article_loop — each shows available Go template variables
- **Chat interface**: AlpineJS-powered conversation with user messages (right-aligned, indigo) and AI responses (left-aligned, gray)
- **Generated code viewer**: Shows the HTML with copy button and validation status badge (green checkmark or red error)
- **Live preview**: Iframe rendering the template with dummy data, refreshable
- **Save form**: Name input + save button, creates template in DB directly from the builder
- **Loading states**: Spinner during AI generation, disabled inputs while loading

### AI Template Generation Handler (`AITemplateGenerate`)
- Accepts: prompt, template_type, chat_history (for context), current_html (for refinement)
- Builds type-specific system prompt with available Go template variables documented
- Sends to active AI provider via `aiRegistry.Generate()`
- Strips markdown code fences from AI response (`extractHTMLFromResponse`)
- Validates as Go template via `engine.ValidateTemplate()`
- Renders preview with type-appropriate dummy data (`buildPreviewData`)
- Returns JSON: `{html, message, valid, validation_error, preview}`

### AI Template Save Handler (`AITemplateSave`)
- Validates template syntax before saving
- Creates template in DB via `templateStore.Create()`
- Logs cache invalidation event
- Returns JSON: `{id}` or `{error}`

### System Prompts (`buildTemplateSystemPrompt`)
Type-specific prompts that document:
- **Header**: `{{.SiteName}}`, `{{.Year}}` — navigation, logo
- **Footer**: `{{.SiteName}}`, `{{.Year}}` — copyright, links
- **Page**: Full set — `{{.Title}}`, `{{.Body}}`, `{{.Header}}`, `{{.Footer}}`, `{{.Excerpt}}`, `{{.MetaDescription}}`, `{{.MetaKeywords}}`, `{{.SiteName}}`, `{{.Year}}`, `{{.Slug}}`, `{{.PublishedAt}}`
- **Article Loop**: `{{range .Posts}}` loop with `{{.Title}}`, `{{.Slug}}`, `{{.Excerpt}}`, `{{.PublishedAt}}`

### Preview Data (`buildPreviewData`)
Returns type-appropriate dummy data:
- **Page**: `engine.PageData` with sample header, footer, title, body, SEO fields
- **Article Loop**: `engine.ListData` with 3 sample posts
- **Header/Footer**: Struct with SiteName and Year

### Templates List Update
- Added "Generate with AI" primary button (indigo) alongside existing "New Template" (now secondary, white)
- Links to `/admin/templates/ai`

### Server Timeout Fix
- Increased `WriteTimeout` from 10s to 90s to accommodate AI endpoint response times (15-30s typical, up to 60s)

## Design decisions
- **JSON API for template builder** — Unlike the content AI assistant (HTML fragments), the template builder uses JSON responses because the frontend needs structured data (HTML, validation status, preview) that AlpineJS processes client-side
- **Conversational refinement** — Chat history and current HTML are passed with each request, allowing iterative design: "make the header sticky" or "add a dark mode toggle"
- **Code fence stripping** — LLMs commonly wrap code in markdown fences; `extractHTMLFromResponse` handles this transparently
- **Iframe preview** — Uses `srcdoc` attribute to safely render the generated template in an isolated context
- **Type-specific system prompts** — Each template type gets a tailored prompt listing exactly which Go template variables are available, reducing hallucinated variables

## Tests (30 total, all passing)
- `TestExtractHTMLFromResponse` — 5 subtests (plain HTML, html fence, generic fence, whitespace, trailing content)
- `TestBuildTemplateSystemPrompt` — 4 subtests (header, footer, page, article_loop variable inclusion)
- `TestBuildPreviewData` — 3 checks (page returns PageData, article_loop returns ListData with posts, header returns struct)
- Plus existing 18 tests from Step 9

## Verified (live testing)
- AI builder page loads at `/admin/templates/ai`
- "Generate with AI" button appears on templates list
- Header template: AI generates valid `{{.SiteName}}` navbar (558 bytes, valid Go template, preview renders)
- Page template: AI generates full HTML with `{{.Title}}`, `{{.Body}}`, `{{.Header}}`, `{{.Footer}}`, SEO meta (2582 bytes, valid, preview renders)
- Save: Template saved to DB with UUID, appears in templates list as "Inactive"
- Cache log: `create` action logged for saved templates
