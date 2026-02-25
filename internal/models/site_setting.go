// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

package models

import "time"

// SiteSetting represents a single configuration key-value pair.
type SiteSetting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SiteSettings is a convenience map for accessing settings by key.
type SiteSettings map[string]string

// Get returns the value for a key, or the fallback if the key doesn't exist.
func (s SiteSettings) Get(key, fallback string) string {
	if v, ok := s[key]; ok && v != "" {
		return v
	}
	return fallback
}
