# Step 2: Database Schema & Migrations — Completed

**Date:** 2026-02-24
**Branch:** feat/project-scaffolding

## What was done

- Integrated goose (v3.27.0) for SQL migrations with `//go:embed`
- Integrated pgx (v5.8.0) as PostgreSQL driver via `database/sql` stdlib interface
- Created 4 migration files:
  - `00001_create_users.sql` — users table with 2FA fields (totp_secret, totp_enabled)
  - `00002_create_content.sql` — unified content table (posts + pages via type field)
  - `00003_create_templates.sql` — AI-generated template storage with versioning
  - `00004_create_cache_invalidation_log.sql` — tracks cache purge events
- Created `internal/database/` package:
  - `database.go` — Connect (pooled), Migrate (embedded SQL)
  - `seed.go` — creates default admin user (totp_enabled=false, forces 2FA setup)
- Created `internal/models/` package:
  - `user.go` — User struct with Role type, 2FA helpers
  - `content.go` — Content struct with ContentType and ContentStatus types
  - `template.go` — Template struct with TemplateType
- Wired database into `cmd/yaaicms/main.go` (connect → migrate → seed)

## 2FA Design Decisions
- `totp_enabled = false` on new users → forces 2FA setup on first login
- No UI to disable 2FA (field can only go true → false via admin reset)
- Admin reset: sets totp_secret = NULL, totp_enabled = false → re-setup on next login

## Verification
- All 4 migrations applied successfully against PostgreSQL 17
- Admin user seeded: admin@yaaicms.local / admin
- Tables confirmed: users, content, templates, cache_invalidation_log, goose_db_version
