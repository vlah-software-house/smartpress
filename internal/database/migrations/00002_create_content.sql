-- +goose Up
CREATE TABLE content (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type             VARCHAR(20)  NOT NULL DEFAULT 'post',
    title            VARCHAR(500) NOT NULL,
    slug             VARCHAR(500) NOT NULL UNIQUE,
    body             TEXT         NOT NULL DEFAULT '',
    excerpt          TEXT,
    status           VARCHAR(20)  NOT NULL DEFAULT 'draft',
    meta_description VARCHAR(500),
    meta_keywords    VARCHAR(500),
    author_id        UUID         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    published_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT content_type_check   CHECK (type IN ('post', 'page')),
    CONSTRAINT content_status_check CHECK (status IN ('draft', 'published'))
);

CREATE INDEX idx_content_type       ON content (type);
CREATE INDEX idx_content_slug       ON content (slug);
CREATE INDEX idx_content_status     ON content (status);
CREATE INDEX idx_content_author_id  ON content (author_id);
CREATE INDEX idx_content_published  ON content (published_at DESC NULLS LAST);

-- +goose Down
DROP TABLE IF EXISTS content;
