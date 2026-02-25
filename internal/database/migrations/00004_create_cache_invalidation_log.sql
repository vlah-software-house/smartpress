-- Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
-- Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
-- All rights reserved. See LICENSE for details.

-- +goose Up
CREATE TABLE cache_invalidation_log (
    id              BIGSERIAL   PRIMARY KEY,
    entity_type     VARCHAR(50) NOT NULL,
    entity_id       UUID        NOT NULL,
    action          VARCHAR(20) NOT NULL,
    invalidated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT cache_log_entity_type_check CHECK (entity_type IN ('content', 'template')),
    CONSTRAINT cache_log_action_check      CHECK (action IN ('create', 'update', 'delete'))
);

CREATE INDEX idx_cache_log_entity     ON cache_invalidation_log (entity_type, entity_id);
CREATE INDEX idx_cache_log_invalidated ON cache_invalidation_log (invalidated_at DESC);

-- +goose Down
DROP TABLE IF EXISTS cache_invalidation_log;
