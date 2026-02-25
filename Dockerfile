# Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
# Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
# All rights reserved. See LICENSE for details.

# =============================================================================
# YaaiCMS — Production Multi-Stage Dockerfile
# =============================================================================
# Stage 1: Build TailwindCSS + vendor frontend JS
# Stage 2: Compile Go binary (with embedded assets + libvips for image processing)
# Stage 3: Minimal runtime image with libvips shared libraries
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

# Vendor HTMX, AlpineJS, and EasyMDE for offline serving.
RUN mkdir -p web/static/js web/static/css \
    && wget -q -O web/static/js/htmx.min.js \
       "https://unpkg.com/htmx.org@2.0.4/dist/htmx.min.js" \
    && wget -q -O web/static/js/alpine.min.js \
       "https://unpkg.com/alpinejs@3.14.8/dist/cdn.min.js" \
    && wget -q -O web/static/js/easymde.min.js \
       "https://unpkg.com/easymde@2.20.0/dist/easymde.min.js" \
    && wget -q -O web/static/css/easymde.min.css \
       "https://unpkg.com/easymde@2.20.0/dist/easymde.min.css"

# ---------------------------------------------------------------------------
# Stage 2: Go binary build (CGO enabled for libvips image processing)
# ---------------------------------------------------------------------------
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates build-base pkgconfig vips-dev

WORKDIR /app

# Download dependencies first (cached unless go.mod/go.sum change).
COPY go.mod go.sum ./
RUN go mod download

# Copy the full source tree.
COPY . .

# Overlay the compiled frontend assets so they get embedded via //go:embed.
COPY --from=frontend /build/web/static/ ./web/static/

# Build with CGO enabled for libvips bindings.
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /yaaicms \
    ./cmd/yaaicms

# ---------------------------------------------------------------------------
# Stage 3: Runtime
# ---------------------------------------------------------------------------
FROM alpine:3.21

# Install CA certificates, timezone data, and libvips runtime libraries.
RUN apk add --no-cache ca-certificates tzdata \
    vips libwebp libpng libjpeg-turbo tiff giflib \
    && adduser -D -H -s /sbin/nologin yaaicms

COPY --from=builder /yaaicms /usr/local/bin/yaaicms

# Limit glib malloc arenas to prevent memory fragmentation under load.
ENV MALLOC_ARENA_MAX=2

USER yaaicms

EXPOSE 8080

ENTRYPOINT ["yaaicms"]
