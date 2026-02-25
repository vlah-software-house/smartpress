# Admin UX Improvements & Site Settings

**Date:** 2026-02-25
**Branch:** feat/admin-ux-settings

## Summary

Four admin UX improvements: collapsible sidebar, visit website link, view-on-site links for content, and a full site settings page with database persistence.

## Changes

### 1. Collapsible Sidebar (WordPress-style)
- `base.html`: Desktop sidebar toggles between w-64 (expanded) and w-16 (collapsed/icons-only).
- Alpine.js `collapsed` state persisted in `localStorage` ("sidebar-collapsed" key).
- Smooth CSS transitions on width change; labels use opacity/overflow for clean hide/show.
- Collapse/expand chevron button at the bottom of the sidebar.
- Native tooltips on collapsed icons (`:title` binding).

### 2. Visit Website Link
- Added "Visit Website" link with external-link icon at the bottom of sidebar nav.
- Opens in new tab (`target="_blank"`).
- Also added to mobile sidebar overlay.

### 3. View on Site Links
- `posts_list.html` / `pages_list.html`: "View" link in actions column for published content, opening `/{slug}` in a new tab.
- `content_form.html`: "View on site" link with icon below the slug field, shown only for existing published content.

### 4. Site Settings Page
- Migration `00013_create_site_settings.sql`: key-value `site_settings` table with sensible defaults (site_title, site_tagline, timezone, language, date_format, posts_per_page, site_url).
- `internal/models/site_setting.go`: `SiteSetting` struct and `SiteSettings` map type with `Get()` helper.
- `internal/store/site_setting.go`: `SiteSettingStore` with `All()`, `Get()`, `Set()`, `SetMany()` (transactional upsert).
- `internal/handlers/admin.go`: Updated `Admin` struct with `siteSettingStore`, updated `SettingsPage` to load settings, added `SettingsSave` handler.
- `internal/router/router.go`: Added `POST /admin/settings` route.
- `settings.html`: Full form with Site Title, Tagline, Site URL, Posts per Page, Timezone (30+ timezones), Language (15 languages), Date Format (6 formats), and Save button. AI Provider section retained below.
