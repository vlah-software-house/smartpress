// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

package models

import (
	"time"

	"github.com/google/uuid"
)

// DesignTheme stores a visual style brief that is injected into all
// template generation prompts to ensure cohesive styling across header,
// footer, page, and article_loop templates.
type DesignTheme struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	StylePrompt string    `json:"style_prompt"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
