# Step 14: Comprehensive Testing

**Date:** 2026-02-25
**Branch:** feat/phase4-polish

## Summary

Added comprehensive test coverage across all SmartPress packages, bringing total coverage from near-zero to 70.0% of statements. Tests cover unit, integration, and functional layers.

## Test Files Created/Modified

### New test files
- `internal/handlers/handler_test.go` — shared test infrastructure (testEnv, helpers, mocks)
- `internal/handlers/admin_crud_test.go` — 35 tests for admin Dashboard, Posts, Pages, Templates, Users, Settings
- `internal/handlers/auth_flow_test.go` — 17 tests for Login, 2FA Setup/Verify, Logout
- `internal/handlers/public_page_test.go` — 6 tests for Homepage, Page rendering, cache behavior
- `internal/handlers/admin_ai_handler_test.go` — 29 tests for AI assistant and template builder endpoints
- `internal/ai/provider_http_test.go` — 44 tests using httptest mocks for all 4 AI providers
- `internal/database/database_test.go` — 4 tests for Connect, Migrate, idempotency
- `internal/database/seed_test.go` — 1 test for seed idempotency
- `internal/router/router_test.go` — 2 tests for health endpoint

### Modified test files
- `internal/engine/engine_test.go` — added 9 integration tests for RenderPage/RenderPostList
- `internal/store/template_test.go` — made TestTemplateStoreList self-sufficient

### Production code changes
- `internal/ai/provider.go` — added `Register()` method to Registry for test provider injection

## Coverage by Package

| Package | Coverage |
|---------|---------|
| config | 100.0% |
| models | 100.0% |
| slug | 100.0% |
| ai | 93.7% |
| engine | 93.3% |
| middleware | 90.5% |
| render | 84.3% |
| store | 82.3% |
| cache | 81.0% |
| session | 81.4% |
| handlers | 69.7% |
| database | 43.7% |
| router | 3.9% |
| storage | 0.0% |
| **Total** | **70.0%** |

## Key Patterns

- Integration tests use real PostgreSQL and Valkey (DB 15) with `t.Skip` when unavailable
- AI providers tested via `net/http/httptest` mock servers
- Handler tests use `newTestEnv(t)` factory with full dependency wiring
- Chi URL params injected via `chi.NewRouteContext()`
- Session data injected via `middleware.SessionKey` context key
- Mock AI provider via `Registry.Register()` for handler-level tests
- All test data cleaned up via `t.Cleanup()`

## Playwright E2E Tests

Added 25 end-to-end tests using Playwright with Chromium, covering:

### Test Files
- `e2e/auth.setup.ts` — login + TOTP 2FA authentication setup (session reuse on re-runs)
- `e2e/dashboard.spec.ts` — dashboard stats, quick actions, sidebar navigation (2 tests)
- `e2e/posts.spec.ts` — full post CRUD lifecycle: list, create, edit, validate, delete (5 tests)
- `e2e/pages.spec.ts` — page CRUD: list, create, public visibility, delete (4 tests)
- `e2e/templates.spec.ts` — template management: list, create+delete, syntax validation, preview (4 tests)
- `e2e/login.spec.ts` — login flow: invalid credentials, empty fields, page structure (3 tests)
- `e2e/public.spec.ts` — public site: homepage, health, 404, auth redirect (4 tests)
- `e2e/settings.spec.ts` — settings page AI providers, users list (2 tests)

### Infrastructure
- `playwright.config.ts` — sequential execution, Chromium only, auth state persistence
- TOTP generation implemented via Node.js crypto (RFC 6238), no external dependencies
- Auth setup reads TOTP secret from PostgreSQL via `psql` (execFileSync)
- Session state cached in `e2e/.auth/admin.json` — avoids rate limiter on re-runs
- Rate limiter awareness: tests minimize auth endpoint hits to stay within 10 req/min limit

### Key Patterns
- Forms use `#content-form button[type="submit"]` or `main button[type="submit"]` to avoid matching the hidden Sign Out submit button in the admin header
- Assertions scoped to `page.locator('main')` to avoid strict-mode violations from sidebar duplicates
- HTMX navigation handled by using `page.goto()` instead of `waitForURL` where sidebar uses HTMX partial swaps
- `getByText('...', { exact: true })` for settings page to avoid matching env var names
