-- Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
-- Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
-- All rights reserved. See LICENSE for details.

-- +goose Up
ALTER TABLE content ADD COLUMN featured_image_id UUID REFERENCES media(id) ON DELETE SET NULL;
CREATE INDEX idx_content_featured_image ON content (featured_image_id) WHERE featured_image_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_content_featured_image;
ALTER TABLE content DROP COLUMN IF EXISTS featured_image_id;
