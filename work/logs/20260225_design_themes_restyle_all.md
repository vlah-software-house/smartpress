# Design Themes & Restyle All

**Date:** 2026-02-25
**Branch:** feat/design-theme-restyle

## Summary

Added a design theme system and "Restyle All" workflow to the AI Template Builder, enabling users to maintain visual consistency across all 4 template types (header, footer, page, article loop).

## Changes

### Database
- Migration `00011_create_design_themes.sql`: new `design_themes` table with partial unique index ensuring at most one active theme at the DB level.

### Backend
- `internal/models/design_theme.go`: DesignTheme model (id, name, style_prompt, is_active, timestamps).
- `internal/store/design_theme.go`: Full CRUD store with transactional Activate (deactivates all others atomically).
- `internal/handlers/admin_ai.go`:
  - Theme CRUD handlers: List, Create, Update, Activate, Deactivate, Delete, ActiveTheme.
  - `getActiveDesignBrief()` helper fetches the active theme's style prompt.
  - `buildTemplateSystemPrompt()` now accepts variadic `designBrief` parameter.
  - `AITemplateGenerate` injects active design brief into prompts.
  - `AIRestylePreview` renders all 4 templates together (header/footer output injected into page/article_loop preview data).
- `internal/router/router.go`: Routes for theme endpoints and restyle-preview.
- `cmd/yaaicms/main.go`: Wires DesignThemeStore into Admin handler.

### Frontend
- `template_ai.html`: Tabbed UI with "Single Template" (unchanged) and "Restyle All" modes.
  - Design Brief panel: theme CRUD (save/load/activate/deactivate/delete), style prompt textarea.
  - Restyle prompt + trigger button; generates header -> footer -> page -> article_loop sequentially.
  - Progress indicator with per-step status (pending/running/done/error).
  - Tabbed code viewer for each generated template.
  - Combined preview with Page View / Blog View tabs using AIRestylePreview endpoint.
  - Save All button: saves all 4 templates with a name prefix.

## Architecture Decisions
- Sequential generation: each template sees previously generated ones as context for visual consistency.
- Partial unique index: enforces single active theme at DB level without application-level races.
- Design brief injection is variadic and optional — backward-compatible with single template mode.
