# Built-in Help Page

**Date:** 2025-02-25
**Branch:** feat/built-in-help

## Summary

Added a comprehensive built-in help page to the admin panel, accessible from the sidebar. The help is a single scrollable page with 12 sections covering every admin feature.

## Changes

- `internal/render/templates/admin/help.html` — New 618-line template with:
  - Table of contents (3-column responsive grid with anchor links)
  - 12 sections: Dashboard, Posts, Pages, Content Editor, AI Assistant, AI Design, Media Library, Categories, Users, Settings, Sidebar & Navigation, Security
  - Section numbering with indigo-100 circular badges and Heroicon SVGs
  - Blue callout boxes for tips, amber for warnings
  - Alpine.js back-to-top floating button
  - Tailwind CSS styling consistent with admin aesthetic
- `internal/handlers/admin.go` — Added `HelpPage` handler
- `internal/router/router.go` — Added `GET /admin/help` route
- `internal/render/templates/admin/base.html` — Added Help link (question mark icon) to desktop and mobile sidebars, placed before "Visit Website"

## Testing

- Deployed to testing environment and verified via Playwright snapshot
- All 12 sections render correctly with proper styling and navigation
