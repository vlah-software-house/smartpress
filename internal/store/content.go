// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"yaaicms/internal/models"
)

// ContentStore handles all content-related database operations.
// It serves both posts and pages through the unified content table.
type ContentStore struct {
	db *sql.DB
}

// NewContentStore creates a new ContentStore with the given database connection.
func NewContentStore(db *sql.DB) *ContentStore {
	return &ContentStore{db: db}
}

// ListByType returns all content items of the given type, ordered by creation date descending.
func (s *ContentStore) ListByType(contentType models.ContentType) ([]models.Content, error) {
	rows, err := s.db.Query(`
		SELECT id, type, title, slug, body, excerpt, status,
		       meta_description, meta_keywords, author_id,
		       published_at, created_at, updated_at
		FROM content
		WHERE type = $1
		ORDER BY created_at DESC
	`, contentType)
	if err != nil {
		return nil, fmt.Errorf("list content by type: %w", err)
	}
	defer rows.Close()

	var items []models.Content
	for rows.Next() {
		var c models.Content
		if err := rows.Scan(
			&c.ID, &c.Type, &c.Title, &c.Slug, &c.Body, &c.Excerpt,
			&c.Status, &c.MetaDescription, &c.MetaKeywords, &c.AuthorID,
			&c.PublishedAt, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan content: %w", err)
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

// FindByID retrieves a content item by its UUID. Returns nil if not found.
func (s *ContentStore) FindByID(id uuid.UUID) (*models.Content, error) {
	c := &models.Content{}
	err := s.db.QueryRow(`
		SELECT id, type, title, slug, body, excerpt, status,
		       meta_description, meta_keywords, author_id,
		       published_at, created_at, updated_at
		FROM content WHERE id = $1
	`, id).Scan(
		&c.ID, &c.Type, &c.Title, &c.Slug, &c.Body, &c.Excerpt,
		&c.Status, &c.MetaDescription, &c.MetaKeywords, &c.AuthorID,
		&c.PublishedAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find content by id: %w", err)
	}
	return c, nil
}

// FindBySlug retrieves a published content item by its slug. Used for public page rendering.
func (s *ContentStore) FindBySlug(slug string) (*models.Content, error) {
	c := &models.Content{}
	err := s.db.QueryRow(`
		SELECT id, type, title, slug, body, excerpt, status,
		       meta_description, meta_keywords, author_id,
		       published_at, created_at, updated_at
		FROM content WHERE slug = $1 AND status = 'published'
	`, slug).Scan(
		&c.ID, &c.Type, &c.Title, &c.Slug, &c.Body, &c.Excerpt,
		&c.Status, &c.MetaDescription, &c.MetaKeywords, &c.AuthorID,
		&c.PublishedAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find content by slug: %w", err)
	}
	return c, nil
}

// Create inserts a new content item and returns it with the generated ID.
func (s *ContentStore) Create(c *models.Content) (*models.Content, error) {
	// If publishing, set the published_at timestamp.
	if c.Status == models.ContentStatusPublished && c.PublishedAt == nil {
		now := time.Now()
		c.PublishedAt = &now
	}

	result := &models.Content{}
	err := s.db.QueryRow(`
		INSERT INTO content (type, title, slug, body, excerpt, status,
		                     meta_description, meta_keywords, author_id, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, type, title, slug, body, excerpt, status,
		          meta_description, meta_keywords, author_id,
		          published_at, created_at, updated_at
	`, c.Type, c.Title, c.Slug, c.Body, c.Excerpt, c.Status,
		c.MetaDescription, c.MetaKeywords, c.AuthorID, c.PublishedAt,
	).Scan(
		&result.ID, &result.Type, &result.Title, &result.Slug, &result.Body,
		&result.Excerpt, &result.Status, &result.MetaDescription, &result.MetaKeywords,
		&result.AuthorID, &result.PublishedAt, &result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create content: %w", err)
	}
	return result, nil
}

// Update modifies an existing content item.
func (s *ContentStore) Update(c *models.Content) error {
	// If transitioning to published and no published_at set, set it now.
	if c.Status == models.ContentStatusPublished && c.PublishedAt == nil {
		now := time.Now()
		c.PublishedAt = &now
	}

	_, err := s.db.Exec(`
		UPDATE content SET
			title = $1, slug = $2, body = $3, excerpt = $4, status = $5,
			meta_description = $6, meta_keywords = $7, published_at = $8,
			updated_at = NOW()
		WHERE id = $9
	`, c.Title, c.Slug, c.Body, c.Excerpt, c.Status,
		c.MetaDescription, c.MetaKeywords, c.PublishedAt, c.ID,
	)
	if err != nil {
		return fmt.Errorf("update content: %w", err)
	}
	return nil
}

// Delete removes a content item by ID.
func (s *ContentStore) Delete(id uuid.UUID) error {
	_, err := s.db.Exec(`DELETE FROM content WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete content: %w", err)
	}
	return nil
}

// ListPublishedByType returns all published content of the given type,
// ordered by published date descending. Used for public page rendering.
func (s *ContentStore) ListPublishedByType(contentType models.ContentType) ([]models.Content, error) {
	rows, err := s.db.Query(`
		SELECT id, type, title, slug, body, excerpt, status,
		       meta_description, meta_keywords, author_id,
		       published_at, created_at, updated_at
		FROM content
		WHERE type = $1 AND status = 'published'
		ORDER BY published_at DESC NULLS LAST
	`, contentType)
	if err != nil {
		return nil, fmt.Errorf("list published content: %w", err)
	}
	defer rows.Close()

	var items []models.Content
	for rows.Next() {
		var c models.Content
		if err := rows.Scan(
			&c.ID, &c.Type, &c.Title, &c.Slug, &c.Body, &c.Excerpt,
			&c.Status, &c.MetaDescription, &c.MetaKeywords, &c.AuthorID,
			&c.PublishedAt, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan content: %w", err)
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

// CountByType returns the number of content items of the given type.
func (s *ContentStore) CountByType(contentType models.ContentType) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM content WHERE type = $1`, contentType).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count content: %w", err)
	}
	return count, nil
}
