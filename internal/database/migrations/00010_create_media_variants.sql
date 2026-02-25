-- +goose Up
-- Responsive image variants: each media item can have multiple
-- WebP-optimised versions at different breakpoints (thumb, sm, md, lg).
CREATE TABLE media_variants (
    id         UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    media_id   UUID          NOT NULL REFERENCES media(id) ON DELETE CASCADE,
    name       VARCHAR(20)   NOT NULL,  -- variant name: thumb, sm, md, lg
    width      INT           NOT NULL,
    height     INT           NOT NULL,
    s3_key     VARCHAR(1000) NOT NULL UNIQUE,
    content_type VARCHAR(50) NOT NULL DEFAULT 'image/webp',
    size_bytes BIGINT        NOT NULL,
    created_at TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- Fast lookup of variants for a given media item.
CREATE INDEX idx_media_variants_media_id ON media_variants (media_id);

-- Unique constraint: one variant name per media item.
CREATE UNIQUE INDEX idx_media_variants_media_name ON media_variants (media_id, name);

-- +goose Down
DROP TABLE IF EXISTS media_variants;
