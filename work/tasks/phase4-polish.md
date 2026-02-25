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
- [x] Image upload to S3-compatible storage (Hetzner CEPH, dual buckets)
- [x] Thumbnail generation (CatmullRom resize, JPEG output)
- [x] Media library admin UI (grid view, upload modal, drag-and-drop, delete)
- [x] Private file serving via presigned URLs
- [x] Content type validation (sniff-based, not header-trusted)

## Step 13: Infrastructure
- [ ] Production Dockerfile (multi-stage build)
- [ ] Kubernetes manifests (Kustomize overlays)
- [ ] TailwindCSS build pipeline (compile from DB classes)

## Step 14: Testing
- [ ] Unit tests (near 100% coverage)
- [ ] Functional tests for handlers
- [ ] Playwright E2E tests for admin UI
