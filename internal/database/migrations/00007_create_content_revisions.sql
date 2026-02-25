-- +goose Up
CREATE TABLE content_revisions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_id      UUID NOT NULL REFERENCES content(id) ON DELETE CASCADE,
    title           TEXT NOT NULL,
    slug            TEXT NOT NULL,
    body            TEXT NOT NULL DEFAULT '',
    excerpt         TEXT,
    status          TEXT NOT NULL DEFAULT 'draft',
    meta_description TEXT,
    meta_keywords   TEXT,
    featured_image_id UUID REFERENCES media(id) ON DELETE SET NULL,
    revision_title  TEXT NOT NULL DEFAULT '',
    revision_log    TEXT NOT NULL DEFAULT '',
    created_by      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_content_revisions_content_id ON content_revisions(content_id);
CREATE INDEX idx_content_revisions_created_at ON content_revisions(content_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS content_revisions;
