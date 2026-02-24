# Step 4: Admin UI Layout — Completed

**Date:** 2026-02-24
**Branch:** feat/project-scaffolding

## What was done

- Created `internal/render/` package — HTML template renderer
  - Supports full-page and HTMX partial rendering (auto-detects `HX-Request` header)
  - Embedded templates via `//go:embed` (no external files needed at runtime)
  - Standalone templates (login, 2fa_setup) render without sidebar
  - Layout templates render with base.html (sidebar + topbar + content)
  - Custom `activeClass` template function for sidebar highlighting

- Created `internal/handlers/admin.go` — Admin handler group
  - Dependency injection via struct (renderer, session store, DB)
  - Handlers: Dashboard, LoginPage, PostsList, PagesList, TemplatesList, UsersList, SettingsPage

- Created 9 HTML templates in `internal/render/templates/admin/`:
  - `base.html` — Full admin layout with dark sidebar, topbar with user menu, mobile responsive
  - `login.html` — Standalone centered login form
  - `2fa_setup.html` — Standalone 2FA QR code setup page
  - `dashboard.html` — Welcome card, stats grid (4 cards), quick actions
  - `posts_list.html` — Posts management table with "New Post" button
  - `pages_list.html` — Pages management table with "New Page" button
  - `templates_list.html` — AI Design empty state with "Generate Template" button
  - `users_list.html` — User management table with role badges and 2FA status
  - `settings.html` — AI provider config form (provider, model, API key)

- Fixed CSRF token availability on first request (stored in context, not just cookie)
- Updated router to use real handlers for all list/view pages
- Integrated TailwindCSS (CDN), HTMX 2.0.4, AlpineJS 3.14.8

## Design Choices
- WordPress-like layout: dark sidebar (gray-800), white content area
- Sidebar shows admin-only sections (Users, Settings) only for admin role
- Mobile responsive: hamburger menu triggers slide-out sidebar via AlpineJS
- User dropdown in topbar with avatar initial, name, email, sign out
- HTMX navigation: sidebar links use hx-get + hx-target="#main-content" + hx-push-url
- CSRF token auto-injected via hx-headers on body element

## Verification
- Login page renders with TailwindCSS styling and CSRF token
- Dashboard redirect works (303 to /admin/login without session)
- All templates compile without errors
