// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

package models

import (
	"time"

	"github.com/google/uuid"
)

// Category represents a hierarchical content category.
// Posts can have at most one category assigned.
type Category struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description string     `json:"description"`
	ParentID    *uuid.UUID `json:"parent_id"`
	SortOrder   int        `json:"sort_order"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Virtual fields populated by store methods.
	Children  []Category `json:"children,omitempty"`
	Depth     int        `json:"depth"`
	PostCount int        `json:"post_count"`
}
