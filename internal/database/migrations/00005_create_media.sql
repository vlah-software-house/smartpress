-- Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
-- Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
-- All rights reserved. See LICENSE for details.

-- +goose Up
CREATE TABLE media (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    filename      VARCHAR(500)  NOT NULL,
    original_name VARCHAR(500)  NOT NULL,
    content_type  VARCHAR(100)  NOT NULL,
    size_bytes    BIGINT        NOT NULL,
    bucket        VARCHAR(100)  NOT NULL,
    s3_key        VARCHAR(1000) NOT NULL UNIQUE,
    thumb_s3_key  VARCHAR(1000),
    alt_text      VARCHAR(500),
    uploader_id   UUID          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_media_uploader     ON media (uploader_id);
CREATE INDEX idx_media_bucket       ON media (bucket);
CREATE INDEX idx_media_created      ON media (created_at DESC);
CREATE INDEX idx_media_content_type ON media (content_type);

-- +goose Down
DROP TABLE IF EXISTS media;
