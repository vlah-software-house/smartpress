-- +goose Up
-- Hierarchical categories for posts. Each post can have at most one category.
CREATE TABLE categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    parent_id   UUID REFERENCES categories(id) ON DELETE SET NULL,
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_categories_parent_id ON categories(parent_id);
CREATE INDEX idx_categories_sort_order ON categories(sort_order);

-- Add optional category reference to content (posts only by convention).
ALTER TABLE content ADD COLUMN category_id UUID REFERENCES categories(id) ON DELETE SET NULL;
CREATE INDEX idx_content_category_id ON content(category_id);

-- Also track category in content revisions for accurate restore.
ALTER TABLE content_revisions ADD COLUMN category_id UUID REFERENCES categories(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE content_revisions DROP COLUMN IF EXISTS category_id;
ALTER TABLE content DROP COLUMN IF EXISTS category_id;
DROP TABLE IF EXISTS categories;
