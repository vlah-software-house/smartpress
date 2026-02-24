# Step 6: Dynamic Template Engine

**Date:** 2026-02-24
**Branch:** feat/dynamic-template-engine
**Status:** Complete

## What was built

### Template Store (`internal/store/template.go`)
- Full CRUD for the `templates` table (List, FindByID, FindActiveByType, Create, Update, Delete, Count)
- Transactional `Activate` method: deactivates all templates of the same type, then activates the target — ensures exactly one active template per type

### Template Engine (`internal/engine/engine.go`)
- Compiles AI-generated HTML+TailwindCSS from the database as Go `html/template`
- `RenderPage`: loads active header/footer/page templates, renders header and footer as fragments, injects them into the page template along with content data
- `RenderPostList`: renders the article_loop template with a list of published posts
- `ValidateTemplate`: syntax-checks a template string before saving to DB
- `ValidateAndRender`: used for admin live preview

### Public Handlers (`internal/handlers/public.go`)
- `Homepage`: tries article_loop with published posts → falls back to "home" page → falls back to a static default
- `Page`: finds content by slug, renders via engine, falls back to raw HTML if template engine fails

### Admin Template CRUD (updated `internal/handlers/admin.go`)
- TemplatesList, TemplateNew, TemplateCreate, TemplateEdit, TemplateUpdate, TemplateDelete, TemplateActivate, TemplatePreview
- Template validation before save
- Live preview via HTMX POST

### Templates UI
- `templates_list.html`: table with type badges, version, active/inactive status, activate/edit/delete actions
- `template_form.html`: create/edit form with name, type selector, HTML textarea editor, template variable help, live preview

### Router updates (`internal/router/router.go`)
- Added template CRUD routes under `/admin/templates`
- Added public routes: `/` (Homepage) and `/{slug}` (Page)
- Updated `New()` signature to accept `public *handlers.Public`

### Seeding (`internal/database/seed.go`)
- Refactored to idempotent per-table seeding (users, templates, content checked independently)
- Seeds 4 default templates: header, footer, page, article_loop — all active
- Seeds 2 sample content items: "home" page and "hello-world" post

### Wiring (`cmd/smartpress/main.go`)
- Added `templateStore`, `engine`, and `publicHandlers` initialization
- Updated `NewAdmin` and `router.New` calls with new dependencies

## Verified
- Clean build (`go build ./...`)
- Server starts, seeding runs on first launch
- Public homepage renders article_loop with header/footer
- Individual pages render via slug (`/hello-world`)
- 404 for missing slugs
- Health endpoint still works
