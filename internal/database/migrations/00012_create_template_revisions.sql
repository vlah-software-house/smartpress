-- Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
-- Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
-- All rights reserved. See LICENSE for details.

-- +goose Up

-- Template revisions store snapshots of a template's state before each update,
-- enabling undo/restore. Follows the same pattern as content_revisions.
CREATE TABLE template_revisions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id     UUID NOT NULL REFERENCES templates(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    html_content    TEXT NOT NULL DEFAULT '',
    revision_title  TEXT NOT NULL DEFAULT '',
    revision_log    TEXT NOT NULL DEFAULT '',
    created_by      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_template_revisions_template_id
    ON template_revisions(template_id);
CREATE INDEX idx_template_revisions_created_at
    ON template_revisions(template_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS template_revisions;
