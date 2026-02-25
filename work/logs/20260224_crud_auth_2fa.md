# Step 5: CRUD for Pages & Posts + Auth + 2FA — Completed

**Date:** 2026-02-24
**Branch:** feat/project-scaffolding

## What was done

### Store Layer (`internal/store/`)
- `user.go` — FindByEmail, FindByID, List, Create, SetTOTPSecret, EnableTOTP, ResetTOTP, Delete, CheckPassword
- `content.go` — ListByType, FindByID, FindBySlug, Create, Update, Delete, CountByType

### Authentication (`internal/handlers/auth.go`)
- **Login flow**: email/password validation with bcrypt → session creation → redirect to 2FA
- **2FA setup**: TOTP secret generation (pquerna/otp) → QR code (skip2/go-qrcode) → verify code → enable
- **2FA verify**: existing users enter authenticator code on login
- **Logout**: session destruction + cookie clear

### Content CRUD (`internal/handlers/admin.go`)
- Posts: list, new, create, edit, update, delete
- Pages: list, new, create, edit, update, delete
- Shared helpers for create/edit/update/delete (DRY across posts and pages)
- User management: list users with real data, reset 2FA for other users
- Dashboard: real stats from database (post count, page count, user count)

### Slug Generation (`internal/slug/slug.go`)
- Converts titles to URL-friendly slugs (lowercase, hyphens, no special chars)
- Auto-generates from title if slug field left empty
- JavaScript auto-slug in the form (real-time preview as you type the title)

### Templates
- `content_form.html` — Unified form for creating/editing posts and pages
  - Quill rich text editor (CDN) with visual/HTML toggle (AlpineJS tabs)
  - Slug auto-generation from title (AlpineJS x-on:input)
  - SEO accordion (excerpt, meta description, meta keywords)
  - Draft/Published status selector
- `2fa_verify.html` — Code entry page for returning users
- Updated all list templates with dynamic data (range loops, conditional badges)
- Updated login/2fa_setup with error message display
- Added `deref` template function for nullable string fields

## Auth Flow Summary
1. GET /admin/login → login form
2. POST /admin/login → validate credentials → create session (TwoFADone=false)
3. If !totp_enabled → redirect to GET /admin/2fa/setup (QR code + secret)
4. If totp_enabled → redirect to GET /admin/2fa/verify (code entry)
5. POST /admin/2fa/verify → validate TOTP code → enable if first time → TwoFADone=true
6. Redirect to /admin/dashboard

## Verification
- Login with admin@yaaicms.local / admin → redirects to 2FA setup
- 2FA setup shows real QR code PNG and manual secret
- CSRF protection working (token from cookie validates on POST)
- Dashboard shows real counts from database
