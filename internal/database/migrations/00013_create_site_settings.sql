-- +goose Up
-- Site-wide configuration (key-value). Seeded with defaults on first run.
CREATE TABLE site_settings (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed sensible defaults.
INSERT INTO site_settings (key, value) VALUES
    ('site_title',       'My Website'),
    ('site_tagline',     'Just another YaaiCMS site'),
    ('timezone',         'UTC'),
    ('language',         'en'),
    ('date_format',      '2006-01-02'),
    ('posts_per_page',   '10'),
    ('site_url',         '');

-- +goose Down
DROP TABLE IF EXISTS site_settings;
