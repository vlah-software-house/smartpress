-- +goose Up
CREATE TABLE templates (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         VARCHAR(255) NOT NULL,
    type         VARCHAR(50)  NOT NULL,
    html_content TEXT         NOT NULL DEFAULT '',
    version      INTEGER      NOT NULL DEFAULT 1,
    is_active    BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT templates_type_check CHECK (type IN ('header', 'footer', 'page', 'article_loop'))
);

CREATE INDEX idx_templates_type      ON templates (type);
CREATE INDEX idx_templates_is_active ON templates (is_active) WHERE is_active = TRUE;

-- +goose Down
DROP TABLE IF EXISTS templates;
