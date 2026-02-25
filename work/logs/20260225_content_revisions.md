# Content Revision System

**Date:** 2026-02-25
**Branch:** feat/featured-image
**Commit:** 18ec30c

## Summary

Implemented a full content revision system for Pages and Posts. Every save
creates a snapshot of the previous state, enabling users to browse revision
history and restore to any past version.

## Changes

### Database
- **Migration 00007**: `content_revisions` table with snapshot fields (title,
  slug, body, excerpt, status, meta, featured image) plus revision metadata
  (revision_title, revision_log, created_by, created_at). FK to `content(id)`
  with `ON DELETE CASCADE`.
- **Migration 00008**: Adds `'restore'` to `cache_log_action_check` constraint
  so cache invalidation logs work when restoring a revision.

### Backend
- **`models/content.go`**: Added `ContentRevision` struct.
- **`store/revision.go`**: New `RevisionStore` with `Create`, `ListByContentID`,
  `FindByID`, `UpdateMeta`, `Count`.
- **`handlers/admin.go`**:
  - `updateContent` — captures old state before applying form values, creates
    revision snapshot, fires background goroutine for AI metadata.
  - `generateRevisionMeta` — builds diff summary comparing old vs new fields,
    uses AI to generate a concise title (if user didn't provide one) and a
    2-4 bullet changelog. Falls back gracefully on AI errors.
  - `RevisionRestore` — creates "Before restore" revision of current state,
    applies revision data back to content, invalidates cache.
  - `RevisionUpdateTitle` — HTMX endpoint for renaming revisions.
  - `editContent` — loads revision list and passes to template.
- **`router/router.go`**: Routes for restore (`POST /admin/revisions/{revisionID}/restore`)
  and title update (`PUT /admin/revisions/{revisionID}/title`).
- **`cmd/yaaicms/main.go`**: Creates `RevisionStore`, passes to `NewAdmin`.

### Frontend
- **`content_form.html`**:
  - "Revision note" text input in the Status & Actions card (edit mode only).
  - "Revision History" expandable sidebar section showing all revisions with
    AI-generated title, timestamp, and status badge.
  - Accordion behaviour via Alpine.js `expandedRev` state.
  - Expanded view shows changelog, old title/slug, and "Restore this revision"
    button with confirmation dialog.

## Testing

Verified on https://yaaicms.test.vlah.sh:
1. **User-provided title**: Saved post with revision message → appears correctly.
2. **AI auto-generated title**: Saved with no message → AI generated "Update excerpt".
3. **AI changelog**: Bullet-point diff summary generated and rendered.
4. **Restore**: Reverts content, creates pre-restore snapshot, cache invalidated.
5. **Cache constraint**: Migration 00008 resolved `cache_log_action_check` violation
   on restore action.
