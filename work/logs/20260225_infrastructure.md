# Step 13: Infrastructure

**Date:** 2026-02-25
**Branch:** feat/phase4-polish
**Status:** Complete

## What was built

### TailwindCSS Build Pipeline

**Admin CSS (compile-time):**
- `package.json` with tailwindcss + @tailwindcss/forms
- `tailwind.config.js` scanning `internal/render/templates/admin/*.html` and `.tailwind-content/` (for DB templates)
- `web/static/css/input.css` — Tailwind directives (@tailwind base/components/utilities)
- `web/embed.go` — `//go:embed all:static` embeds compiled CSS + vendored JS into the binary
- Admin templates conditionally load CDN (dev) or local compiled files (production) via `isDev` template function

**Public CSS (DB templates, on-demand):**
- `scripts/build-public-css.sh` — Extracts all template HTML from PostgreSQL, writes to temp directory, runs Tailwind CLI, outputs minified `web/static/css/public.css`
- Reads credentials from `.env` / `.secrets`
- Can be run as a CI/CD step or manually after template changes

**Static file serving:**
- `web/embed.go` provides `web.StaticFS` embedded filesystem
- Router serves `/static/*` from the embedded FS
- In development, templates use CDN instead (no build step needed)

### Production Dockerfile (Multi-Stage Build)

Three stages:
1. **Frontend (node:22-alpine):** Installs tailwindcss, compiles admin CSS (minified, tree-shaken), downloads vendored HTMX 2.0.4 and AlpineJS 3.14.8
2. **Builder (golang:1.25-alpine):** Downloads Go modules, overlays compiled static assets, builds fully static binary with `CGO_ENABLED=0 -ldflags="-s -w"`
3. **Runtime (alpine:3.21):** Minimal image with ca-certificates and tzdata, runs as non-root `smartpress` user

### Kubernetes Manifests (Kustomize Overlays)

**Base (`deploy/base/`):**
- `deployment.yaml` — Single replica, health checks (liveness + readiness on /health), resource limits, secrets via envFrom
- `service.yaml` — ClusterIP on port 80 → container 8080
- `ingress.yaml` — nginx ingress class, 50MB body size limit, TLS
- `kustomization.yaml` — Standard labels with `includeSelectors`

**Testing overlay (`deploy/overlays/testing/`):**
- Namespace: `smartpress-testing`
- Ingress: `smartpress.test.vlah.sh` with cert-manager TLS
- Deployment patch: `APP_ENV=testing`, remote DB at 10.0.0.6:5432, Valkey at 10.0.0.6:6379 (DB 101), bumped resource limits (1 CPU, 512Mi)
- Secrets: Generated from `.secrets` via `kubectl create secret generic --from-env-file`

## Bug fix

- **2fa_verify standalone template:** Added `2fa_verify` to the `standaloneTemplates` map in render.go. Previously it was paired with `base.html` but defines its own full HTML page, causing the 2FA verify screen to render with an empty admin layout on full page loads.

## Files created
- `package.json` — Tailwind build scripts
- `tailwind.config.js` — Content paths for admin + DB templates
- `web/static/css/input.css` — Tailwind directives
- `web/embed.go` — Static asset embedding
- `scripts/build-public-css.sh` — DB template CSS compilation pipeline
- `Dockerfile` — Production multi-stage build
- `.dockerignore` — Exclude dev artifacts from Docker context
- `deploy/base/kustomization.yaml` — Kustomize base
- `deploy/base/deployment.yaml` — App deployment
- `deploy/base/service.yaml` — ClusterIP service
- `deploy/base/ingress.yaml` — Ingress with TLS
- `deploy/overlays/testing/kustomization.yaml` — Testing overlay
- `deploy/overlays/testing/ingress-patch.yaml` — Testing domain + cert-manager
- `deploy/overlays/testing/deployment-patch.yaml` — Testing env vars + resources

## Files modified
- `internal/render/render.go` — Added `devMode` param, `isDev` template func, `standaloneTemplates` map
- `internal/render/templates/admin/base.html` — Conditional CDN/local asset loading
- `internal/render/templates/admin/login.html` — Conditional CDN/local asset loading
- `internal/render/templates/admin/2fa_setup.html` — Conditional CDN/local asset loading
- `internal/render/templates/admin/2fa_verify.html` — Conditional CDN/local asset loading
- `internal/router/router.go` — Added static file serving from embedded FS
- `cmd/smartpress/main.go` — Pass `cfg.IsDev()` to renderer
- `.gitignore` — Added built assets, tailwind temp, K8s secrets

## Tests
- All 63 existing tests pass
- Kustomize base and testing overlay validate successfully

## Verified
- `go build ./...` — clean compilation
- `go test ./...` — 63 tests pass
- `kubectl kustomize deploy/base/` — valid output
- `kubectl kustomize deploy/overlays/testing/` — valid output with correct namespace, domain, DB/Valkey hosts
