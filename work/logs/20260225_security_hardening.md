# Step 11: Security & Hardening

**Date:** 2026-02-25
**Branch:** feat/phase4-polish
**Status:** Complete

## What was fixed

### 1. Critical XSS — Public page fallback (`public.go`)
- **Before:** `fmt.Fprintf(w, "...%s...", content.Title, content.Body)` — no HTML escaping, raw content rendered as HTML
- **After:** Fallback page uses `html.EscapeString(content.Title)` and no longer renders the body at all — shows a safe "could not be rendered" error page instead
- Renamed local `html` variables to `rendered`/`cached` to avoid shadowing the `html` package import

### 2. XSS — Template preview error (`admin.go`)
- **Before:** `fmt.Fprintf(w, "...%s...", err.Error())` — error messages could contain HTML from crafted template input
- **After:** Uses `html.EscapeString(err.Error())` with string concatenation instead of `fmt.Fprintf`

### 3. Environment-aware secure cookies (`session.go`, `csrf.go`)
- **Session cookie:** `NewStore` now accepts a `secure bool` parameter; cookie `Secure` flag set based on environment
- **CSRF cookie:** New `NewCSRF(secure bool)` factory function; old `CSRF` variable kept as backward-compatible alias
- **Wiring:** `main.go` computes `secureCookies := !cfg.IsDev()` and passes to both session store and router
- In development (`APP_ENV=development`): `Secure: false` (works over HTTP on localhost)
- In production/testing: `Secure: true` (cookies only sent over HTTPS)

### 4. Security headers middleware (`security.go`)
- New `SecureHeaders` middleware added to global middleware chain (before Logger)
- Headers set on every response:
  - `X-Content-Type-Options: nosniff` — prevents MIME-sniffing
  - `X-Frame-Options: SAMEORIGIN` — clickjacking protection
  - `X-XSS-Protection: 0` — disables legacy XSS filter (CSP preferred)
  - `Referrer-Policy: strict-origin-when-cross-origin` — controls referer leakage
  - `Permissions-Policy: interest-cohort=()` — opts out of FLoC

### 5. Rate limiting (`ratelimit.go`)
- New `RateLimiter` with per-IP sliding window algorithm
- Background goroutine cleans up expired entries every 5 minutes
- **Auth endpoints** (login, logout, 2FA): 10 requests/minute per IP
- **AI endpoints**: 30 requests/minute per IP
- Returns HTTP 429 Too Many Requests when limit exceeded
- `clientIP()` helper checks X-Forwarded-For, X-Real-IP, then RemoteAddr

### 6. AI Registry concurrency safety (`provider.go`)
- Added `sync.RWMutex` to `Registry` struct
- All public methods (`Active`, `SetActive`, `ActiveName`, `Available`, `HasProvider`) now acquire appropriate read/write locks
- Safe for concurrent access from multiple HTTP handler goroutines

### 7. DB connection pool limits (`database.go`)
- `SetMaxOpenConns(25)` — prevents exhausting PostgreSQL connection limit
- `SetMaxIdleConns(5)` — keeps a small warm pool
- `SetConnMaxLifetime(5 * time.Minute)` — recycles stale connections

### 8. Input validation (`validate.go`)
- Content forms: title required, max 300 chars; slug max 300; body max 100K; excerpt max 1K; meta fields max 500
- Template forms: name required, max 200 chars; HTML required, max 500K chars
- Validation runs before any DB operations in `createContent`, `updateContent`, `TemplateCreate`
- Uses `utf8.RuneCountInString` for correct Unicode character counting

### 9. 2FA setup GET side effect fix (`auth.go`)
- **Before:** `TwoFASetupPage` generated a new TOTP secret on every GET, saving to DB — refreshing the page silently rotated the secret, breaking already-scanned QR codes
- **After:** Checks if user already has a pending (unverified) secret and reuses it; only generates new secret on first visit; redirects away if 2FA is already enabled

## Tests (51 total, all passing)

### New middleware tests (14)
- `TestSecureHeaders` — 5 subtests (one per header)
- `TestRateLimiterAllow` — basic allow/deny
- `TestRateLimiterWindowExpiry` — window sliding behavior
- `TestRateLimiterMiddleware` — HTTP integration test
- `TestClientIP` — 5 subtests (XFF single, XFF multiple, XRI, RemoteAddr, no port)
- `TestRateLimiterCleanup` — expired entry removal
- `TestNewCSRFSecureFlag` — 2 subtests (secure true/false cookie flag)
- `TestCSRFRejectsStateMutationWithoutToken` — POST without token returns 403
- `TestCSRFAcceptsValidToken` — POST with valid header token returns 200

### New handler tests (20)
- `TestValidateContent` — 8 subtests (valid, empty title, whitespace title, title/slug/body too long, empty body/slug ok)
- `TestValidateMetadata` — 5 subtests (all empty, valid, excerpt/desc/kw too long)
- `TestValidateTemplate` — 7 subtests (valid, empty/whitespace name, name too long, empty/whitespace/too long HTML)

### Existing tests (30 handlers + 5 AI = 35)
- All passing unchanged

## Files created
- `internal/middleware/security.go` — SecureHeaders middleware
- `internal/middleware/security_test.go` — tests
- `internal/middleware/ratelimit.go` — RateLimiter with sliding window
- `internal/middleware/ratelimit_test.go` — tests
- `internal/middleware/csrf_test.go` — tests for NewCSRF
- `internal/handlers/validate.go` — input validation helpers
- `internal/handlers/validate_test.go` — tests

## Files modified
- `internal/handlers/public.go` — XSS fix, variable renaming
- `internal/handlers/admin.go` — preview XSS fix, input validation, `fmt` → `html` import
- `internal/handlers/auth.go` — 2FA setup GET side effect fix
- `internal/session/session.go` — environment-aware Secure cookie flag
- `internal/middleware/csrf.go` — NewCSRF factory with Secure flag parameter
- `internal/ai/provider.go` — mutex for Registry concurrency safety
- `internal/database/database.go` — connection pool limits
- `internal/router/router.go` — SecureHeaders, NewCSRF, rate limiting wiring
- `cmd/smartpress/main.go` — secureCookies flag passed to session store and router
