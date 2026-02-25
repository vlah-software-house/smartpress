# AI Design System Improvements

**Date:** 2026-02-25
**Branch:** feat/ai-design-improvements

## Summary

Two improvements to the AI template generation system:
1. **Detailed system prompts** — Each template variable now includes purpose, content type, and usage patterns so the AI generates better-informed templates.
2. **Real content preview** — Users can select existing posts/pages for template preview instead of hardcoded dummy data.

## Changes

### Modified
- `internal/handlers/admin_ai.go` — Rewrote `buildTemplateSystemPrompt()` with detailed variable documentation, added `AIPreviewContentList`, `buildRealPreviewData`, `buildRealPagePreview`, `buildRealArticleLoopPreview`, `buildSrcsetForPreview` methods
- `internal/handlers/admin.go` — Modified `TemplatePreview` to accept `content_id` and `template_type` params for real content and type-aware preview
- `internal/engine/engine.go` — Added exported `RewriteBodyImages()` wrapper for use by admin handlers
- `internal/router/router.go` — Added `GET /admin/ai/preview-content` route
- `internal/render/templates/admin/template_ai.html` — Added content selector dropdown, wired content_id into generate and preview flows, improved template type help text
- `internal/render/templates/admin/template_form.html` — Added content selector to preview section, new `refreshFormPreview()` method
- `internal/handlers/admin_ai_test.go` — Updated system prompt test expectations

## How it works

### System Prompt Improvements
Each variable now documents:
- Data type (string, template.HTML, int)
- Whether it's always set or may be empty
- What the content looks like (e.g., Body is Markdown-converted HTML with headings, lists, inline images with srcset)
- Recommended HTML pattern for rendering (e.g., exact `<img srcset>` pattern with `{{if}}` guards)
- Design guidelines specific to each template type

### Real Content Preview
1. `AIPreviewContentList` returns JSON list of available posts/pages (published + drafts)
2. Frontend shows a `<select>` dropdown when template type is page or article_loop
3. When a content item is selected, its ID is sent in the `content_id` form param
4. Backend fetches real content, converts Markdown to HTML, resolves featured images with srcset variants
5. For article_loop, selecting "Real published posts" fetches all published posts with their images
6. Falls back to dummy data when no content is selected or the fetch fails
