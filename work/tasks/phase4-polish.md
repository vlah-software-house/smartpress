# Phase 4: Polish & Production

## Step 11: Security & Hardening
- [x] XSS sanitization on AI-generated HTML (public fallback + template preview)
- [x] CSRF secure cookies (environment-aware), rate limiting (auth + AI), input validation
- [x] html/template safe rendering audit (removed raw fmt.Fprintf, added html.EscapeString)
- [x] Security headers middleware (X-Content-Type-Options, X-Frame-Options, Referrer-Policy)
- [x] AI Registry concurrency safety (sync.RWMutex)
- [x] DB connection pool limits (25 open, 5 idle, 5min lifetime)
- [x] 2FA setup GET side effect fix (reuse pending secret on refresh)

## Step 12: Media Management
- [ ] Image upload, storage, serving
- [ ] Thumbnail generation
- [ ] Media library admin UI

## Step 13: Infrastructure
- [ ] Production Dockerfile (multi-stage build)
- [ ] Kubernetes manifests (Kustomize overlays)
- [ ] TailwindCSS build pipeline (compile from DB classes)

## Step 14: Testing
- [ ] Unit tests (near 100% coverage)
- [ ] Functional tests for handlers
- [ ] Playwright E2E tests for admin UI
