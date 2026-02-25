# Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
# Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
# All rights reserved. See LICENSE for details.

# =============================================================================
# YaaiCMS — Production Multi-Stage Dockerfile
# =============================================================================
# Stage 1: Build TailwindCSS + vendor frontend JS
# Stage 2: Compile Go binary (with embedded assets)
# Stage 3: Minimal runtime image
# =============================================================================

# ---------------------------------------------------------------------------
# Stage 1: Frontend assets
# ---------------------------------------------------------------------------
FROM node:22-alpine AS frontend

WORKDIR /build

# Install Tailwind dependencies.
COPY package.json ./
RUN npm install --ignore-scripts

# Copy Tailwind config and admin templates to scan for CSS classes.
COPY tailwind.config.js ./
COPY web/static/css/input.css ./web/static/css/
COPY internal/render/templates/ ./internal/render/templates/

# Compile admin CSS (minified, tree-shaken to only used classes).
RUN npx tailwindcss -i ./web/static/css/input.css \
    -o ./web/static/css/admin.css --minify

# Vendor HTMX and AlpineJS for offline serving.
RUN mkdir -p web/static/js \
    && wget -q -O web/static/js/htmx.min.js \
       "https://unpkg.com/htmx.org@2.0.4/dist/htmx.min.js" \
    && wget -q -O web/static/js/alpine.min.js \
       "https://unpkg.com/alpinejs@3.14.8/dist/cdn.min.js"

# ---------------------------------------------------------------------------
# Stage 2: Go binary build
# ---------------------------------------------------------------------------
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Download dependencies first (cached unless go.mod/go.sum change).
COPY go.mod go.sum ./
RUN go mod download

# Copy the full source tree.
COPY . .

# Overlay the compiled frontend assets so they get embedded via //go:embed.
COPY --from=frontend /build/web/static/ ./web/static/

# Build a fully static binary with stripped debug symbols.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /yaaicms \
    ./cmd/yaaicms

# ---------------------------------------------------------------------------
# Stage 3: Runtime
# ---------------------------------------------------------------------------
FROM alpine:3.21

# Install CA certificates for HTTPS calls to AI providers and S3.
RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -H -s /sbin/nologin yaaicms

COPY --from=builder /yaaicms /usr/local/bin/yaaicms

USER yaaicms

EXPOSE 8080

ENTRYPOINT ["yaaicms"]
