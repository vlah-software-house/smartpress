# AI Improve Panel for Template Editor

**Date:** 2026-02-25
**Branch:** main (direct commit)

## Summary

Added an AI Improve panel to the template edit form (`/admin/templates/:id`), allowing users to describe template improvements in natural language. The AI processes the current template HTML and the user's request, then updates the template in-place.

## Changes

### `internal/render/templates/admin/template_form.html`
- Added `x-data="templateEditor()"` Alpine.js component wrapper
- Added `x-ref="editor"` to textarea for programmatic access
- Added `x-model="templateType"` to type select (new templates)
- Added collapsible "AI Improve" panel between editor and preview:
  - Toggle button with sparkle icon and chevron
  - Chat-style message history (user messages right-aligned indigo, AI responses left-aligned gray)
  - Loading spinner during AI generation
  - Text input with Enter key support
  - "Improve" / "Working..." button state
- Updated help text to include missing template variables:
  - Page: `FeaturedImageURL`, `Excerpt`, `Slug`, `PublishedAt`
  - Article loop: `FeaturedImageURL`
  - New section: Header/Footer variables (`SiteName`, `Year`)
- Added `<script>` with `templateEditor()` function:
  - Reads current HTML from textarea
  - POSTs to `/admin/ai/generate-template` with `current_html`, `prompt`, `template_type`
  - Updates textarea and dispatches input event on success
  - Auto-refreshes live preview via `htmx.ajax()`
  - Shows validation warnings if template has issues

## Testing

- Verified panel expands/collapses correctly
- Tested AI improvement request: "Make the header sticky with a subtle shadow, add a mobile hamburger menu"
- Confirmed textarea updated with improved HTML
- Confirmed live preview auto-refreshed with new content
- Verified chat history displays user and AI messages correctly
