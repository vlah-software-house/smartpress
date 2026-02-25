// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

package database

import (
	"database/sql"
	"fmt"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
)

// Seed populates the database with initial development data.
// It creates a default admin user if none exists. The admin will be
// prompted to set up 2FA on first login (totp_enabled = false).
func Seed(db *sql.DB) error {
	// Each seed function is idempotent — checks its own table before inserting.
	if err := seedAdminUser(db); err != nil {
		return err
	}
	if err := seedTemplates(db); err != nil {
		return fmt.Errorf("seed templates: %w", err)
	}
	if err := seedContent(db); err != nil {
		return fmt.Errorf("seed content: %w", err)
	}
	return nil
}

// seedAdminUser creates a default admin if no users exist. The admin must
// set up 2FA on first login (totp_enabled = false).
func seedAdminUser(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return fmt.Errorf("seed check users: %w", err)
	}
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("seed bcrypt: %w", err)
	}

	_, err = db.Exec(`
		INSERT INTO users (email, password_hash, display_name, role, totp_enabled)
		VALUES ($1, $2, $3, $4, $5)
	`, "admin@yaaicms.local", string(hash), "Admin", "admin", false)
	if err != nil {
		return fmt.Errorf("seed insert admin: %w", err)
	}

	slog.Info("database seeded with default admin user",
		"email", "admin@yaaicms.local",
		"password", "admin",
	)
	return nil
}

// seedTemplates creates a minimal set of active templates (header, footer,
// page, article_loop) so the public site works immediately after setup.
func seedTemplates(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM templates").Scan(&count); err != nil {
		return fmt.Errorf("check templates: %w", err)
	}
	if count > 0 {
		return nil
	}

	templates := []struct {
		name, tmplType, html string
	}{
		{
			name:     "Default Header",
			tmplType: "header",
			html: `<header class="bg-white border-b border-gray-200">
  <div class="max-w-5xl mx-auto px-4 py-4 flex items-center justify-between">
    <a href="/" class="text-xl font-bold text-indigo-600">YaaiCMS</a>
    <nav class="space-x-4 text-sm text-gray-600">
      <a href="/" class="hover:text-gray-900">Home</a>
    </nav>
  </div>
</header>`,
		},
		{
			name:     "Default Footer",
			tmplType: "footer",
			html: `<footer class="bg-gray-50 border-t border-gray-200 mt-12">
  <div class="max-w-5xl mx-auto px-4 py-6 text-center text-sm text-gray-500">
    &copy; {{ .Year }} YaaiCMS. All rights reserved.
  </div>
</footer>`,
		},
		{
			name:     "Default Page",
			tmplType: "page",
			html: `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{ .Title }} — {{ .SiteName }}</title>
  {{ if .MetaDescription }}<meta name="description" content="{{ .MetaDescription }}">{{ end }}
  {{ if .MetaKeywords }}<meta name="keywords" content="{{ .MetaKeywords }}">{{ end }}
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-white text-gray-900 min-h-screen flex flex-col">
  {{ .Header }}
  <main class="flex-1 max-w-3xl mx-auto px-4 py-8 w-full">
    <h1 class="text-3xl font-bold mb-6">{{ .Title }}</h1>
    <article class="prose max-w-none">{{ .Body }}</article>
  </main>
  {{ .Footer }}
</body>
</html>`,
		},
		{
			name:     "Default Article Loop",
			tmplType: "article_loop",
			html: `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{ .Title }} — {{ .SiteName }}</title>
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-white text-gray-900 min-h-screen flex flex-col">
  {{ .Header }}
  <main class="flex-1 max-w-3xl mx-auto px-4 py-8 w-full">
    <h1 class="text-3xl font-bold mb-8">{{ .Title }}</h1>
    {{ range .Posts }}
    <article class="mb-8 pb-8 border-b border-gray-200 last:border-0">
      <h2 class="text-xl font-semibold">
        <a href="/{{ .Slug }}" class="text-indigo-600 hover:text-indigo-800">{{ .Title }}</a>
      </h2>
      {{ if .PublishedAt }}<time class="text-sm text-gray-500">{{ .PublishedAt }}</time>{{ end }}
      {{ if .Excerpt }}<p class="mt-2 text-gray-600">{{ .Excerpt }}</p>{{ end }}
    </article>
    {{ end }}
  </main>
  {{ .Footer }}
</body>
</html>`,
		},
	}

	for _, t := range templates {
		_, err := db.Exec(`
			INSERT INTO templates (name, type, html_content, version, is_active)
			VALUES ($1, $2, $3, 1, true)
		`, t.name, t.tmplType, t.html)
		if err != nil {
			return fmt.Errorf("insert template %q: %w", t.name, err)
		}
	}

	slog.Info("seeded default templates", "count", len(templates))
	return nil
}

// seedContent creates a sample homepage and blog post so the public site
// renders meaningful content right after setup.
func seedContent(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM content").Scan(&count); err != nil {
		return fmt.Errorf("check content: %w", err)
	}
	if count > 0 {
		return nil
	}

	// Look up the admin user to assign as author.
	var authorID string
	if err := db.QueryRow("SELECT id FROM users LIMIT 1").Scan(&authorID); err != nil {
		return fmt.Errorf("find author: %w", err)
	}

	pages := []struct {
		contentType, title, slug, body, excerpt, status string
		published                                       bool
	}{
		{
			contentType: "page",
			title:       "Welcome to YaaiCMS",
			slug:        "home",
			body:        `<p>This is your new YaaiCMS site. Edit this page from the <a href="/admin">admin panel</a>, or create AI-powered templates to customize the look and feel.</p>`,
			status:      "published",
			published:   true,
		},
		{
			contentType: "post",
			title:       "Hello World",
			slug:        "hello-world",
			body:        `<p>This is a sample blog post created during setup. You can edit or delete it from the admin panel.</p>`,
			excerpt:     "A sample blog post to get you started with YaaiCMS.",
			status:      "published",
			published:   true,
		},
	}

	for _, p := range pages {
		publishedAt := "NULL"
		if p.published {
			publishedAt = "NOW()"
		}

		var excerpt *string
		if p.excerpt != "" {
			excerpt = &p.excerpt
		}

		_, err := db.Exec(fmt.Sprintf(`
			INSERT INTO content (type, title, slug, body, excerpt, status, author_id, published_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, %s)
		`, publishedAt), p.contentType, p.title, p.slug, p.body, excerpt, p.status, authorID)
		if err != nil {
			return fmt.Errorf("insert content %q: %w", p.slug, err)
		}
	}

	slog.Info("seeded sample content", "count", len(pages))
	return nil
}
