// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

package models

import (
	"time"

	"github.com/google/uuid"
)

// ContentType distinguishes between posts and pages in the unified content table.
type ContentType string

const (
	ContentTypePost ContentType = "post"
	ContentTypePage ContentType = "page"
)

// ContentStatus represents the publishing state of a content item.
type ContentStatus string

const (
	ContentStatusDraft     ContentStatus = "draft"
	ContentStatusPublished ContentStatus = "published"
)

// Content represents a post or page in the CMS. Posts and pages share the
// same table, differentiated by the Type field.
type Content struct {
	ID              uuid.UUID     `json:"id"`
	Type            ContentType   `json:"type"`
	Title           string        `json:"title"`
	Slug            string        `json:"slug"`
	Body            string        `json:"body"`
	Excerpt         *string       `json:"excerpt,omitempty"`
	Status          ContentStatus `json:"status"`
	MetaDescription *string       `json:"meta_description,omitempty"`
	MetaKeywords    *string       `json:"meta_keywords,omitempty"`
	FeaturedImageID *uuid.UUID    `json:"featured_image_id,omitempty"`
	AuthorID        uuid.UUID     `json:"author_id"`
	PublishedAt     *time.Time    `json:"published_at,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// IsPublished returns true if the content item is in published status.
func (c *Content) IsPublished() bool {
	return c.Status == ContentStatusPublished
}

// ContentRevision stores a snapshot of a content item's state before an edit.
// Created automatically on every save, it enables reverting to previous versions.
type ContentRevision struct {
	ID              uuid.UUID  `json:"id"`
	ContentID       uuid.UUID  `json:"content_id"`
	Title           string     `json:"title"`
	Slug            string     `json:"slug"`
	Body            string     `json:"body"`
	Excerpt         *string    `json:"excerpt,omitempty"`
	Status          string     `json:"status"`
	MetaDescription *string    `json:"meta_description,omitempty"`
	MetaKeywords    *string    `json:"meta_keywords,omitempty"`
	FeaturedImageID *uuid.UUID `json:"featured_image_id,omitempty"`
	RevisionTitle   string     `json:"revision_title"`
	RevisionLog     string     `json:"revision_log"`
	CreatedBy       uuid.UUID  `json:"created_by"`
	CreatedAt       time.Time  `json:"created_at"`
}
