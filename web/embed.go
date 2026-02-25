// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

// Package web provides embedded static assets (CSS, JS) for the admin interface.
// In development, templates load assets from CDN; in production, the compiled
// and vendored files are embedded here and served at /static/.
package web

import "embed"

// StaticFS embeds the web/static/ directory tree. In Docker builds, this
// includes the compiled TailwindCSS and vendored HTMX/AlpineJS files.
// In local development it may only contain the input.css source file.
//
//go:embed all:static
var StaticFS embed.FS
