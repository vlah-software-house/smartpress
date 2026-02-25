# Inline Content Image srcset Rewriting

**Date:** 2026-02-25
**Branch:** feat/body-image-srcset (merged to main)

## Summary

Added automatic responsive srcset rewriting for inline images in content bodies. The engine post-processes HTML after Markdown→HTML conversion, matching `<img src>` URLs against S3 storage, batch-fetching WebP variants, and injecting `srcset` + `sizes` attributes. Authors write standard `![alt](url)` Markdown — no changes needed.

## Changes

### Modified
- `internal/engine/engine.go` — Added `SetMediaDeps()`, `rewriteBodyImages()`, `buildSrcsetFromVariants()`, regex for img tag matching
- `internal/storage/s3.go` — Added `ExtractS3Key()` to extract S3 object key from public URL
- `internal/store/media.go` — Added `FindByS3Keys()` for batch media lookup by S3 key
- `cmd/yaaicms/main.go` — Wired media deps to engine via `SetMediaDeps()`

### New
- `internal/engine/rewrite_test.go` — Unit tests for img regex and rewriter no-op cases

## How it works

1. After Markdown→HTML conversion, `rewriteBodyImages()` scans for `<img>` tags
2. For each `src` URL, `ExtractS3Key()` checks if it matches our S3 storage pattern
3. All matched S3 keys are batch-fetched from `media` table via `FindByS3Keys()`
4. All media IDs are batch-fetched from `media_variants` via `FindByMediaIDs()`
5. Tags are rewritten back-to-front (to preserve indices) with `srcset` and `sizes`
6. Tags that already have `srcset` are left untouched

## Verified on testing cluster

Input Markdown: `![A fluffy bunny](https://fsn1.../00be99de...png)`

Output HTML:
```html
<img src="...00be99de...png"
     srcset="..._sm.webp 640w, ..._md.webp 1024w, ..._lg.webp 1792w"
     sizes="(max-width: 640px) 640px, (max-width: 1024px) 1024px, 1920px"
     alt="A fluffy bunny">
```
