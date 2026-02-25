# Step 12: Media Management

**Date:** 2026-02-25
**Branch:** feat/phase4-polish
**Status:** Complete

## What was built

### S3 Storage Client (`internal/storage/s3.go`)
- AWS SDK v2 wrapper configured for CEPH/Hetzner with path-style addressing
- `New()` returns `(nil, nil)` if credentials are empty — app starts without storage
- `Upload()` sets `public-read` ACL for the public bucket, standard ACL for private
- `Delete()` removes objects from either bucket
- `FileURL()` constructs path-style public URLs, with optional CDN/public URL override
- `PresignedURL()` generates time-limited signed URLs for private files via `s3.PresignClient`
- `PublicBucket()` / `PrivateBucket()` getters for bucket names

### Database Migration (`00005_create_media.sql`)
- `media` table: id (UUID), filename, original_name, content_type, size_bytes, bucket, s3_key (UNIQUE), thumb_s3_key, alt_text, uploader_id (FK to users), created_at
- Indexes on uploader_id, bucket, created_at DESC, content_type

### Media Model (`internal/models/media.go`)
- `Media` struct with JSON tags for API responses
- `IsImage()` checks `strings.HasPrefix(contentType, "image/")`
- `HumanSize()` formats bytes as "B", "KB", or "MB"

### Media Store (`internal/store/media.go`)
- `Create()` INSERT RETURNING full row
- `FindByID()` by UUID
- `List(limit, offset)` paginated, ordered by created_at DESC
- `Delete()` returns the deleted row for S3 cleanup
- `Count()` for dashboard stats

### Upload Handler (`internal/handlers/admin_media.go`)
- **File validation:** Content type detected via `http.DetectContentType()` (sniffs first 512 bytes, doesn't trust headers). Allowlist: JPEG, PNG, GIF, WebP, SVG, PDF.
- **SVG detection:** Special handling since `DetectContentType` returns `text/xml` for SVGs
- **Max size:** 50 MB enforced via `http.MaxBytesReader` + `ParseMultipartForm`
- **Storage key:** `media/{year}/{month}/{uuid}.{ext}` — collision-free
- **Bucket selection:** Defaults to public, `bucket=private` form field switches to private
- **Thumbnail generation:** For JPEG/PNG/WebP images wider than 400px:
  - `image.DecodeConfig` first to check dimensions (prevents memory bombs — max 100M pixels)
  - Full decode + `draw.CatmullRom.Scale` for high-quality resize
  - Encoded as JPEG quality 80
  - Uploaded to same bucket with `_thumb.jpg` suffix
- **Response:** JSON with id, url, thumb_url, filename, size, type

### Media Library Page Handler (`MediaLibrary`)
- Loads media items with pagination (50 per page)
- Builds view structs with URLs (public bucket) and thumbnail URLs
- Graceful handling when S3 is not configured (shows config message)

### Media Delete Handler (`MediaDelete`)
- Deletes from PostgreSQL first (returns row for cleanup)
- Best-effort S3 cleanup (original + thumbnail)
- Returns empty body for HTMX swap (removes card from grid)

### Media Serve Handler (`MediaServe`)
- Public files: HTTP redirect to direct S3 URL
- Private files: generates 1-hour presigned URL, redirects

### Admin UI (`media_library.html`)
- **Upload modal**: Drag-and-drop zone with file picker fallback
- **Alt text input**: For image accessibility
- **Bucket selector**: Public (direct URL) or Private (signed URL)
- **Upload progress**: XHR with `upload.onprogress` shows real-time percentage bar
- **Media grid**: Responsive 2-5 column grid with aspect-square thumbnails
- **Hover actions**: Copy URL button, Delete button (with confirmation)
- **Empty state**: Helpful message when no files uploaded
- **No-storage state**: Configuration instructions when S3 not set up

### Navigation + Dashboard Updates
- Added "Media" nav item with photo icon to desktop and mobile sidebars
- Dashboard stats: replaced static "Templates: 0" with dynamic "Media Files" count

### Configuration (`config.go`)
- 7 new fields: S3Endpoint, S3Region, S3AccessKey, S3SecretKey, S3BucketPublic, S3BucketPrivate, S3PublicURL
- Region defaults to `fsn1` (Hetzner Frankfurt)
- Bucket names default to `yaaicms-public` and `yaaicms-private`

### Wiring (`main.go`)
- Conditional S3 connection (skips with warning if not configured)
- MediaStore always initialized (migration creates table regardless)
- storageClient and mediaStore passed to Admin handler group

### Routes (`router.go`)
- `GET /admin/media` — Media library page
- `POST /admin/media` — File upload
- `DELETE /admin/media/{id}` — Delete media
- `GET /admin/media/{id}/url` — Get/redirect to file URL

## Tests (63 total, all passing)

### New tests (12)
- `TestGenerateThumbnail` — 3 subtests (JPEG resize, PNG resize, skip small image)
- `TestExtensionFromType` — 7 subtests (all MIME types + unknown)
- `TestMediaModelMethods` — 2 subtests (IsImage, HumanSize with 5 cases)

### Existing tests (51)
- All passing unchanged

## Dependencies added
- `github.com/aws/aws-sdk-go-v2` (core + config + credentials + service/s3)
- `github.com/aws/smithy-go`
- `golang.org/x/image` (draw for CatmullRom scaling, webp decoder)

## Verified
- Server starts clean, migration 00005 runs successfully
- Media library page loads (behind auth — correctly redirects to 2FA)
- Graceful degradation when S3 not configured (warning log + UI message)
- Dashboard shows Media Files count
- Navigation shows Media link in both desktop and mobile views
