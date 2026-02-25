# Responsive Image Optimization with libvips

**Date:** 2025-02-25
**Branch:** feat/responsive-image-optimization

## Summary

Replaced the stdlib-based JPEG thumbnail pipeline with a libvips-powered responsive image variant system. All uploaded images now automatically get WebP variants at four breakpoints (320, 640, 1024, 1920px), and public templates can use `<img srcset>` for mobile-first delivery.

## Changes

### New packages/files
- `internal/imaging/imaging.go` — libvips wrapper with `GenerateVariants()`, `Startup()`/`Shutdown()`
- `internal/store/media_variant.go` — `VariantStore` with batch CRUD, missing-variant finder
- `internal/database/migrations/00010_create_media_variants.sql` — `media_variants` table

### Modified
- `Dockerfile` — CGO_ENABLED=1, Alpine vips-dev (build) and vips runtime packages
- `cmd/yaaicms/main.go` — imaging lifecycle, VariantStore wiring
- `internal/handlers/admin.go` — Added `variantStore` to Admin struct
- `internal/handlers/admin_media.go` — Replaced `generateThumbnail()` with `generateAndUploadVariants()` + `saveVariants()` helpers; added `MediaRegenerateVariants` (single) and `MediaRegenerateBulk` (batch) endpoints
- `internal/handlers/admin_ai.go` — AI image generation now creates variants
- `internal/handlers/public.go` — Added `variantStore` dep, `resolveFeaturedImage()` returns srcset data, batch variant lookup for post listings
- `internal/engine/engine.go` — Added `FeaturedImage` type; `PageData`/`PostItem` now include `FeaturedImageSrcset` and `FeaturedImageAlt`
- `internal/storage/s3.go` — Added `Download()` method for fetching originals during regeneration
- `internal/store/media.go` — Added `UpdateThumbKey()` for regeneration
- `internal/router/router.go` — Added `/{id}/regenerate` and `/regenerate-all` routes
- `internal/render/templates/admin/media_library.html` — "Regenerate All" button, per-image regenerate action
- `internal/models/media.go` — Added `MediaVariant` struct

### Variant breakpoints
| Name  | Width  | Quality | Purpose |
|-------|--------|---------|---------|
| thumb | 320px  | 75      | Admin thumbnails |
| sm    | 640px  | 80      | Mobile |
| md    | 1024px | 80      | Tablet |
| lg    | 1920px | 80      | Desktop |

### How srcset works
Templates receive `FeaturedImageSrcset` as a pre-built string:
```
/media/2026/02/uuid_sm.webp 640w, /media/2026/02/uuid_md.webp 1024w, /media/2026/02/uuid_lg.webp 1920w
```
Authors use: `<img src="{{.FeaturedImageURL}}" srcset="{{.FeaturedImageSrcset}}" sizes="..." alt="{{.FeaturedImageAlt}}">`
