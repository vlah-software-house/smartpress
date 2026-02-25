-- Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
-- Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
-- All rights reserved. See LICENSE for details.

-- +goose Up
-- Add 'restore' as a valid cache invalidation action (used when reverting content
-- to a previous revision).
ALTER TABLE cache_invalidation_log
    DROP CONSTRAINT cache_log_action_check,
    ADD  CONSTRAINT cache_log_action_check CHECK (action IN ('create', 'update', 'delete', 'restore'));

-- +goose Down
ALTER TABLE cache_invalidation_log
    DROP CONSTRAINT cache_log_action_check,
    ADD  CONSTRAINT cache_log_action_check CHECK (action IN ('create', 'update', 'delete'));
