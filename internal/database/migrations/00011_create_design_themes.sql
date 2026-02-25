-- +goose Up

-- Design themes store the visual style brief that gets injected into all
-- template generation prompts, ensuring a cohesive look across header,
-- footer, page, and article_loop templates.
CREATE TABLE design_themes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(200) NOT NULL,
    style_prompt TEXT NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Only one theme can be active at a time.
CREATE UNIQUE INDEX idx_design_themes_active ON design_themes (is_active) WHERE is_active = TRUE;

-- +goose Down
DROP TABLE IF EXISTS design_themes;
