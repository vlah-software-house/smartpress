-- +goose Up
-- Add body_format column to distinguish Markdown from legacy HTML content.
-- Existing rows default to 'html' to preserve backward compatibility.
ALTER TABLE content
    ADD COLUMN body_format VARCHAR(10) NOT NULL DEFAULT 'html'
    CONSTRAINT content_body_format_check CHECK (body_format IN ('html', 'markdown'));

-- Also add to revisions so restored content preserves its format.
ALTER TABLE content_revisions
    ADD COLUMN body_format VARCHAR(10) NOT NULL DEFAULT 'html'
    CONSTRAINT revision_body_format_check CHECK (body_format IN ('html', 'markdown'));

-- +goose Down
ALTER TABLE content_revisions DROP COLUMN body_format;
ALTER TABLE content DROP COLUMN body_format;
