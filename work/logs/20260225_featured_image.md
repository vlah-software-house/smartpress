# Featured Image Support for Posts

**Date:** 2026-02-25
**Branch:** feat/featured-image
**Commit:** 4d477bb

## Changes

### Database
- Migration 00006: added `featured_image_id UUID REFERENCES media(id) ON DELETE SET NULL` to content table with partial index

### Backend
- Updated Content model with `FeaturedImageID *uuid.UUID` field
- Refactored ContentStore to use shared `scanContent` helper and `contentColumns` constant (eliminates column list duplication across 5 queries)
- Added `ImageGenerator` interface in `ai/image.go` — optional interface for providers that support image generation
- Implemented DALL-E 3 image generation in OpenAI provider (`GenerateImage` method)
- Registry gains `GenerateImage()` and `SupportsImageGeneration()` methods
- New handler `AIGenerateImage` — generates image via AI, uploads to S3, creates media record, returns HTML fragment
- Updated `createContent` and `updateContent` handlers to read `featured_image_id` from form data
- Updated `editContent` to resolve featured image URL for template display
- Updated public handler to resolve featured image URLs for page and post list rendering
- Added `FeaturedImageURL` field to engine's `PageData` and `PostItem` structs

### Frontend
- Added "Featured Image" card in content form (posts only) with:
  - Image preview with hover-to-remove button
  - Drag-and-drop style upload area
  - AI image generation panel with prompt textarea
  - Hidden `featured_image_id` input that gets submitted with the form
- Upload uses fetch API to POST to existing `/admin/media` endpoint

### Template System
- Updated template variable documentation to include `{{.FeaturedImageURL}}` for page and article_loop templates
