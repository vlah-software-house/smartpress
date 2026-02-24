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
- [ ] Integrate Chi router
- [ ] Middleware: structured logging
- [ ] Middleware: panic recovery
- [ ] Middleware: CSRF protection
- [ ] Middleware: session/auth
- [ ] Admin route group (`/admin/*`)
- [ ] Public route group (`/*`)

## Step 4: Admin UI Layout
- [ ] Admin HTML shell (sidebar, top bar, main content area)
- [ ] TailwindCSS integration (CDN for dev)
- [ ] HTMX-powered navigation (hx-get, hx-target)
- [ ] AlpineJS for dropdowns, modals, ephemeral state

## Step 5: CRUD for Pages & Posts
- [ ] User authentication (login page, session management)
- [ ] Pages: list, create, edit, delete
- [ ] Posts: list, create, edit, delete
- [ ] Text editor integration (Quill/Trix)
- [ ] Slug auto-generation
- [ ] Draft/publish status toggle
