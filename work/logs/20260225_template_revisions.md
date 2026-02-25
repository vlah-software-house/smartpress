# Template Revisions

**Date:** 2026-02-25
**Branch:** feat/template-revisions

## Summary

Added revision history for templates, following the same pattern as the existing content revision system. Every template update now snapshots the old state, enabling undo/restore.

## Changes

### Database
- Migration `00012_create_template_revisions.sql`: new `template_revisions` table with CASCADE FK to templates, indexed by template_id and created_at DESC.

### Backend
- `internal/models/template.go`: Added `TemplateRevision` struct.
- `internal/store/template_revision.go`: New store with Create, ListByTemplateID, FindByID, UpdateMeta.
- `internal/handlers/admin.go`:
  - `TemplateEdit` now loads and passes revisions to the form.
  - `TemplateUpdate` captures old state and creates revision snapshot before saving; redirects back to edit form instead of list.
  - `TemplateRevisionRestore`: creates "Before restore" backup, then applies revision data.
  - `TemplateRevisionUpdateTitle`: updates revision title via HTMX.
  - `generateTemplateRevisionMeta`: background goroutine for AI-generated titles and changelogs.
- `internal/router/router.go`: Routes for `/admin/template-revisions/{revisionID}/restore` and `.../title`.
- `cmd/yaaicms/main.go`: Wires `TemplateRevisionStore` into Admin handler.

### Frontend
- `template_form.html`: Added revision note input field (optional, before actions), and revision history accordion panel after the form (matching content_form.html pattern).
