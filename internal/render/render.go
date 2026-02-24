// Package render provides HTML template rendering for the admin interface.
// It supports full-page and HTMX partial rendering, automatically detecting
// the request type via the HX-Request header.
package render

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"

	"smartpress/internal/middleware"
	"smartpress/internal/session"
)

//go:embed templates/admin/*.html
var adminFS embed.FS

// PageData holds all data passed to admin templates.
type PageData struct {
	Title     string            // Page title for <title> tag
	Section   string            // Active sidebar section (e.g., "dashboard", "posts")
	Session   *session.Data     // Current user session (nil if unauthenticated)
	CSRFToken string            // CSRF token for forms and HTMX headers
	Data      map[string]any    // Page-specific data
	Flashes   []Flash           // One-time notification messages
}

// Flash represents a one-time notification message displayed to the user.
type Flash struct {
	Type    string // "success", "error", "warning", "info"
	Message string
}

// Renderer handles template parsing and execution for admin pages.
type Renderer struct {
	templates map[string]*template.Template
	funcMap   template.FuncMap
}

// New creates a Renderer by parsing all admin templates from the embedded
// filesystem. Each page template is paired with the base layout.
func New() (*Renderer, error) {
	r := &Renderer{
		templates: make(map[string]*template.Template),
		funcMap: template.FuncMap{
			"activeClass": func(current, target string) string {
				if current == target {
					return "bg-gray-900 text-white"
				}
				return "text-gray-300 hover:bg-gray-700 hover:text-white"
			},
			// deref safely dereferences a string pointer for use in templates.
			"deref": func(s *string) string {
				if s == nil {
					return ""
				}
				return *s
			},
		},
	}

	// Find all page templates (everything except base.html).
	pages, err := filepath.Glob("internal/render/templates/admin/*.html")
	if err != nil {
		return nil, fmt.Errorf("glob templates: %w", err)
	}

	// If running from binary (embedded), list from embed.FS instead.
	if len(pages) == 0 {
		entries, err := adminFS.ReadDir("templates/admin")
		if err != nil {
			return nil, fmt.Errorf("read embedded templates: %w", err)
		}
		for _, e := range entries {
			if !e.IsDir() {
				pages = append(pages, e.Name())
			}
		}
	}

	// Parse each page template paired with the base layout.
	for _, page := range pages {
		name := filepath.Base(page)
		if name == "base.html" {
			continue
		}

		// Strip .html extension for the template name.
		tmplName := name[:len(name)-len(".html")]

		// Standalone templates (login, 2fa_setup) don't need the base layout.
		var tmpl *template.Template
		var parseErr error

		switch tmplName {
		case "login", "2fa_setup":
			tmpl, parseErr = template.New(name).Funcs(r.funcMap).ParseFS(
				adminFS, "templates/admin/"+name,
			)
		default:
			tmpl, parseErr = template.New("base.html").Funcs(r.funcMap).ParseFS(
				adminFS, "templates/admin/base.html", "templates/admin/"+name,
			)
		}

		if parseErr != nil {
			return nil, fmt.Errorf("parse template %s: %w", name, parseErr)
		}

		r.templates[tmplName] = tmpl
	}

	return r, nil
}

// Page renders a full admin page or an HTMX partial, depending on the
// request headers. For HTMX requests, only the "content" block is sent.
// For full page loads, the entire base layout is rendered.
func (rn *Renderer) Page(w http.ResponseWriter, r *http.Request, name string, data *PageData) {
	tmpl, ok := rn.templates[name]
	if !ok {
		http.Error(w, fmt.Sprintf("template %q not found", name), http.StatusInternalServerError)
		return
	}

	// Inject CSRF token from context (set by CSRF middleware).
	data.CSRFToken = middleware.CSRFTokenFromCtx(r.Context())

	// Inject session from context.
	if data.Session == nil {
		data.Session = middleware.SessionFromCtx(r.Context())
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// HTMX request: render only the content fragment.
	if isHTMX(r) {
		if err := executeTemplate(w, tmpl, "content", data); err != nil {
			http.Error(w, "template error", http.StatusInternalServerError)
		}
		return
	}

	// Full page request: render the complete layout.
	execName := "base.html"
	// Standalone pages (login, 2fa_setup) use their own root template.
	switch name {
	case "login", "2fa_setup":
		execName = name + ".html"
	}

	if err := executeTemplate(w, tmpl, execName, data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

// executeTemplate wraps template execution with error handling.
func executeTemplate(w io.Writer, tmpl *template.Template, name string, data any) error {
	return tmpl.ExecuteTemplate(w, name, data)
}

// isHTMX returns true if the request was made by HTMX (has HX-Request header).
func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

