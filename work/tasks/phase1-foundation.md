# Phase 1: Foundation & "The WordPress Feel" (MVP 1)

## Step 1: Project Scaffolding
- [x] Initialize git repo and Go module
- [x] Create project directory structure (standard Go layout)
- [x] Docker Compose for PostgreSQL + Valkey
- [x] Create `.secrets.example` template and `.gitignore`
- [x] Set up `work/tasks/` and `work/logs/` directories
- [x] Create entry point (`cmd/smartpress/main.go`) with graceful shutdown
- [x] Create config loader (`internal/config/config.go`)

## Step 2: Database Schema & Migrations
- [x] Choose and integrate migration tool (goose v3)
- [x] Design schema: `users` table (roles, auth, 2FA with TOTP)
- [x] Design schema: `content` table (unified posts+pages via type field)
- [x] Design schema: `templates` table (html_content, version, is_active, type)
- [x] Design schema: `cache_invalidation_log` table
- [x] Write database seeding scripts (default admin user)
- [x] Create Go models (User, Content, Template)
- [x] Wire database connect → migrate → seed into main.go

## Step 3: Go Server & Router
- [x] Integrate Chi router (v5.2.5)
- [x] Middleware: structured logging (slog)
- [x] Middleware: panic recovery with stack trace
- [x] Middleware: CSRF protection (double-submit cookie, HTMX-compatible)
- [x] Middleware: session/auth (LoadSession, RequireAuth, Require2FA, RequireAdmin)
- [x] Session store (Valkey-backed, 24h TTL, secure cookies)
- [x] Admin route group (`/admin/*`) with auth + 2FA middleware
- [x] Public route group (`/*`)
- [x] Valkey client connection (`internal/cache/valkey.go`)

## Step 4: Admin UI Layout
- [x] Admin HTML shell (sidebar, top bar, main content area)
- [x] TailwindCSS integration (CDN for dev)
- [x] HTMX-powered navigation (hx-get, hx-target="#main-content", hx-push-url)
- [x] AlpineJS for dropdowns, modals, mobile sidebar
- [x] Template renderer with full-page / HTMX-partial auto-detection
- [x] 9 admin templates (base, login, 2fa, dashboard, posts, pages, templates, users, settings)
- [x] Handler group with dependency injection (renderer, sessions, DB)
- [x] CSRF token context injection (works on first visit)

## Step 5: CRUD for Pages & Posts
- [x] User authentication (login + logout with bcrypt)
- [x] 2FA with TOTP (QR code setup, mandatory, no disable, admin reset)
- [x] Store layer (UserStore, ContentStore) for DB access
- [x] Pages: list, create, edit, delete
- [x] Posts: list, create, edit, delete
- [x] Text editor integration (Quill CDN with visual/HTML toggle)
- [x] Slug auto-generation (Go + JavaScript real-time)
- [x] Draft/publish status toggle
- [x] SEO metadata fields (excerpt, meta description, meta keywords)
- [x] Dashboard with real database stats
- [x] Users list with 2FA status + admin reset
