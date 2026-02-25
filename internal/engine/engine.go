// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

// Package engine provides the dynamic template rendering engine for public
// pages. It loads AI-generated templates from the database, compiles them
// as Go html/templates, and renders public pages by injecting content data.
package engine

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"time"

	"yaaicms/internal/markdown"
	"yaaicms/internal/models"
	"yaaicms/internal/store"
)

// PageData holds all variables available to a page template when rendering
// a public page. Template authors (or AI) can use these as {{.Title}}, etc.
type PageData struct {
	SiteName         string
	Title            string
	Body             template.HTML // Content body — raw HTML from editor
	Excerpt          string
	MetaDescription  string
	MetaKeywords     string
	FeaturedImageURL string        // Public URL of the featured image (empty if none)
	Slug             string
	PublishedAt      string
	Header           template.HTML // Pre-rendered header fragment
	Footer           template.HTML // Pre-rendered footer fragment
	Year             int
}

// PostItem represents a single post in a listing (used by article_loop template).
type PostItem struct {
	Title            string
	Slug             string
	Excerpt          string
	FeaturedImageURL string // Public URL of the featured image (empty if none)
	PublishedAt      string
}

// ListData holds variables available to the article_loop template.
type ListData struct {
	SiteName string
	Title    string
	Posts    []PostItem
	Header   template.HTML
	Footer   template.HTML
	Year     int
}

// Engine compiles and renders templates from the database. It maintains
// an in-memory cache (L1) of compiled Go templates keyed by ID+version,
// so repeated renders skip the expensive template.Parse step.
type Engine struct {
	templateStore *store.TemplateStore
	cache         *templateCache
}

// New creates a new template rendering engine with an empty L1 cache.
func New(templateStore *store.TemplateStore) *Engine {
	return &Engine{
		templateStore: templateStore,
		cache:         newTemplateCache(),
	}
}

// InvalidateTemplate removes a specific template from the L1 cache.
// Called by admin handlers after template update or delete.
func (e *Engine) InvalidateTemplate(id string) {
	e.cache.invalidate(id)
}

// InvalidateAllTemplates clears the entire L1 cache. Called after
// template activation since it changes which template serves each type.
func (e *Engine) InvalidateAllTemplates() {
	e.cache.invalidateAll()
}

// RenderPage renders a content item using the active page template,
// header, and footer. featuredImageURL is the public URL for the
// featured image (pass "" if none). Returns the complete HTML as a byte slice.
func (e *Engine) RenderPage(content *models.Content, featuredImageURL string) ([]byte, error) {
	// Load active templates for each component.
	header, err := e.renderFragment(models.TemplateTypeHeader, nil)
	if err != nil {
		slog.Warn("header template not found or failed", "error", err)
		header = ""
	}

	footer, err := e.renderFragment(models.TemplateTypeFooter, nil)
	if err != nil {
		slog.Warn("footer template not found or failed", "error", err)
		footer = ""
	}

	// Load the active page template.
	pageTmpl, err := e.templateStore.FindActiveByType(models.TemplateTypePage)
	if err != nil || pageTmpl == nil {
		return nil, fmt.Errorf("no active page template found")
	}

	// Build the data for the page template.
	publishedAt := ""
	if content.PublishedAt != nil {
		publishedAt = content.PublishedAt.Format("January 2, 2006")
	}

	// Convert Markdown body to HTML if needed; raw HTML is passed through unchanged.
	bodyHTML := content.Body
	if content.BodyFormat == models.BodyFormatMarkdown {
		rendered, err := markdown.ToHTML(content.Body)
		if err != nil {
			slog.Warn("markdown conversion failed, using raw body", "error", err)
		} else {
			bodyHTML = rendered
		}
	}

	data := PageData{
		SiteName:         "YaaiCMS",
		Title:            content.Title,
		Body:             template.HTML(bodyHTML),
		FeaturedImageURL: featuredImageURL,
		Slug:             content.Slug,
		PublishedAt:      publishedAt,
		Header:           template.HTML(header),
		Footer:           template.HTML(footer),
		Year:             time.Now().Year(),
	}

	if content.Excerpt != nil {
		data.Excerpt = *content.Excerpt
	}
	if content.MetaDescription != nil {
		data.MetaDescription = *content.MetaDescription
	}
	if content.MetaKeywords != nil {
		data.MetaKeywords = *content.MetaKeywords
	}

	// Compile and execute the page template (L1 cached by ID+version).
	return e.compileAndRender(pageTmpl.ID.String(), pageTmpl.Version, pageTmpl.HTMLContent, data)
}

// RenderPostList renders the article_loop template with a list of posts.
// featuredImages maps content ID strings to their public image URLs.
func (e *Engine) RenderPostList(posts []models.Content, featuredImages map[string]string) ([]byte, error) {
	header, err := e.renderFragment(models.TemplateTypeHeader, nil)
	if err != nil {
		slog.Warn("header template not found or failed", "error", err)
		header = ""
	}

	footer, err := e.renderFragment(models.TemplateTypeFooter, nil)
	if err != nil {
		slog.Warn("footer template not found or failed", "error", err)
		footer = ""
	}

	loopTmpl, err := e.templateStore.FindActiveByType(models.TemplateTypeArticleLoop)
	if err != nil || loopTmpl == nil {
		return nil, fmt.Errorf("no active article_loop template found")
	}

	var postItems []PostItem
	for _, p := range posts {
		item := PostItem{
			Title: p.Title,
			Slug:  p.Slug,
		}
		if p.Excerpt != nil {
			item.Excerpt = *p.Excerpt
		}
		if p.PublishedAt != nil {
			item.PublishedAt = p.PublishedAt.Format("January 2, 2006")
		}
		if featuredImages != nil {
			item.FeaturedImageURL = featuredImages[p.ID.String()]
		}
		postItems = append(postItems, item)
	}

	data := ListData{
		SiteName: "YaaiCMS",
		Title:    "Blog",
		Posts:    postItems,
		Header:   template.HTML(header),
		Footer:   template.HTML(footer),
		Year:     time.Now().Year(),
	}

	return e.compileAndRender(loopTmpl.ID.String(), loopTmpl.Version, loopTmpl.HTMLContent, data)
}

// ValidateTemplate attempts to compile a template string and returns an
// error if the Go template syntax is invalid. Used before saving to DB.
func (e *Engine) ValidateTemplate(htmlContent string) error {
	_, err := template.New("validate").Parse(htmlContent)
	if err != nil {
		return fmt.Errorf("invalid template syntax: %w", err)
	}
	return nil
}

// ValidateAndRender compiles a template string and renders it with the
// given data. Used for live preview in the admin panel. Not cached since
// preview content is ephemeral.
func (e *Engine) ValidateAndRender(htmlContent string, data any) ([]byte, error) {
	return e.compileAndRender("", 0, htmlContent, data)
}

// renderFragment loads and renders a template fragment (header or footer).
func (e *Engine) renderFragment(tmplType models.TemplateType, data any) (string, error) {
	tmpl, err := e.templateStore.FindActiveByType(tmplType)
	if err != nil || tmpl == nil {
		return "", fmt.Errorf("no active %s template", tmplType)
	}

	result, err := e.compileAndRender(tmpl.ID.String(), tmpl.Version, tmpl.HTMLContent, data)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

// compileAndRender compiles a template string and executes it with the
// given data. If id and version are provided (non-empty id), the compiled
// template is cached in L1 to avoid re-parsing on subsequent requests.
func (e *Engine) compileAndRender(id string, version int, tmplContent string, data any) ([]byte, error) {
	var compiled *template.Template

	// Try L1 cache first (skip for ad-hoc renders like preview).
	if id != "" {
		compiled = e.cache.get(id, version)
	}

	if compiled == nil {
		var err error
		compiled, err = template.New("page").Parse(tmplContent)
		if err != nil {
			return nil, fmt.Errorf("compile template: %w", err)
		}
		// Store in L1 cache for next time.
		if id != "" {
			e.cache.put(id, version, compiled)
		}
	}

	var buf bytes.Buffer
	if err := compiled.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return buf.Bytes(), nil
}
