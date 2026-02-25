// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

// Package slug provides URL-friendly slug generation from arbitrary strings.
package slug

import (
	"regexp"
	"strings"
)

var (
	// nonAlphanumeric matches anything that isn't a letter, digit, or space.
	nonAlphanumeric = regexp.MustCompile(`[^a-z0-9\s-]`)
	// multipleHyphens collapses consecutive hyphens into one.
	multipleHyphens = regexp.MustCompile(`-{2,}`)
)

// Generate creates a URL-friendly slug from the given string.
// Example: "Hello, World! 2026" → "hello-world-2026"
func Generate(s string) string {
	result := strings.ToLower(strings.TrimSpace(s))
	result = nonAlphanumeric.ReplaceAllString(result, "")
	result = strings.ReplaceAll(result, " ", "-")
	result = multipleHyphens.ReplaceAllString(result, "-")
	result = strings.Trim(result, "-")
	return result
}
