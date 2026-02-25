# Markdown Editor with Media Picker

**Date:** 2026-02-25
**Branch:** feat/markdown-editor

## Summary

Replaced the plain HTML textarea in the Posts and Pages content editor with a full-featured Markdown editor (EasyMDE) including side-by-side preview, formatting toolbar, and an integrated media picker modal for inserting images from the media library.

## Changes

### Backend

- **Migration 00009:** Added `body_format` column (`html`|`markdown`) to both `content` and `content_revisions` tables, defaulting to `html` for backward compatibility
- **Models:** Added `BodyFormat` type with `BodyFormatHTML` and `BodyFormatMarkdown` constants
- **Store (content/revision):** Updated all queries to include `body_format` in SELECT, INSERT, UPDATE
- **Markdown package:** Created `internal/markdown/` with goldmark wrapper (GFM, Typographer, syntax highlighting with Monokai, unsafe HTML pass-through)
- **Engine:** Updated `RenderPage` to convert Markdown to HTML via goldmark when `body_format=markdown`
- **Handlers:** Updated `createContent`, `updateContent`, and `RevisionRestore` to handle `body_format`
- **Media JSON endpoint:** Added `GET /admin/media/json` returning JSON array of public images for the picker modal
- **AI prompts:** Updated `AIGenerateContent` and `AIRewrite` to output/preserve Markdown syntax

### Frontend

- **EasyMDE integration:** Initialized on the body textarea with custom toolbar (Bold, Italic, Heading, Quote, Lists, Link, Image from Media Library, Table, HR, Code, Preview, Side-by-Side, Fullscreen, Guide)
- **Media picker modal:** Alpine.js component with Library/Upload tabs, image grid, alt text input, and Markdown `![alt](url)` insertion into EasyMDE
- **CSS overrides:** Styled EasyMDE to match the admin theme (indigo focus colors, code blocks, blockquotes)
- **AI integration:** Updated "Apply to Content" buttons to use `window._markdownEditor.value()` API

### Infrastructure

- **Dockerfile:** Added EasyMDE JS and CSS vendoring for production builds
- **base.html:** Added EasyMDE CSS/JS includes (CDN for dev, /static/ for prod)

## Testing

- EasyMDE loads correctly with full toolbar
- Side-by-side preview renders Markdown in real-time
- Media picker opens, shows library images, allows selection with alt text
- Insert button writes `![alt](url)` into editor at cursor position
- Form submission saves `body_format=markdown` and Markdown body to DB
- Revision history correctly captures pre-update body_format
- Public page renders Markdown via goldmark (headings, bold, lists, blockquotes)
- Zero console errors/warnings
