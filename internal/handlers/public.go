package handlers

import (
	"html"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"smartpress/internal/cache"
	"smartpress/internal/engine"
	"smartpress/internal/models"
	"smartpress/internal/store"
)

// Public groups handlers for the public-facing site rendered by the
// dynamic template engine. It checks the L2 Valkey page cache before
// invoking the template engine, and stores rendered results on miss.
type Public struct {
	engine       *engine.Engine
	contentStore *store.ContentStore
	pageCache    *cache.PageCache
}

// NewPublic creates a new Public handler group.
func NewPublic(eng *engine.Engine, contentStore *store.ContentStore, pageCache *cache.PageCache) *Public {
	return &Public{
		engine:       eng,
		contentStore: contentStore,
		pageCache:    pageCache,
	}
}

// Homepage renders the site homepage. If an article_loop template is active,
// it renders a blog-style post listing. Otherwise, it looks for a page with
// slug "home" or falls back to a simple default.
func (p *Public) Homepage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check L2 cache first.
	if cached, ok := p.pageCache.Get(ctx, cache.HomepageKey()); ok {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(cached)
		return
	}

	// Try to render a blog-style homepage with the article_loop template.
	posts, err := p.contentStore.ListPublishedByType(models.ContentTypePost)
	if err != nil {
		slog.Error("list published posts failed", "error", err)
	}

	if len(posts) > 0 {
		rendered, err := p.engine.RenderPostList(posts)
		if err == nil {
			p.pageCache.Set(ctx, cache.HomepageKey(), rendered)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(rendered)
			return
		}
		slog.Warn("article_loop render failed, trying homepage", "error", err)
	}

	// Fall back to a "home" page if it exists.
	home, err := p.contentStore.FindBySlug("home")
	if err == nil && home != nil {
		rendered, err := p.engine.RenderPage(home)
		if err == nil {
			p.pageCache.Set(ctx, cache.HomepageKey(), rendered)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(rendered)
			return
		}
		slog.Warn("homepage render failed", "error", err)
	}

	// Default fallback when no templates or content exist yet (not cached).
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html><head><title>SmartPress</title>
<script src="https://cdn.tailwindcss.com"></script></head>
<body class="bg-gray-100 flex items-center justify-center min-h-screen">
<div class="text-center">
<h1 class="text-4xl font-bold text-gray-900"><span class="text-indigo-600">Smart</span>Press</h1>
<p class="mt-2 text-gray-500">Your site is running. Set up templates in the admin panel.</p>
<a href="/admin/login" class="mt-4 inline-block text-indigo-600 hover:text-indigo-800 text-sm">Go to Admin Panel</a>
</div></body></html>`))
}

// Page renders a public page or post by its slug using the template engine.
func (p *Public) Page(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slugParam := chi.URLParam(r, "slug")

	// Check L2 cache first.
	if cached, ok := p.pageCache.Get(ctx, cache.SlugKey(slugParam)); ok {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(cached)
		return
	}

	content, err := p.contentStore.FindBySlug(slugParam)
	if err != nil {
		slog.Error("find content by slug failed", "error", err, "slug", slugParam)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if content == nil {
		http.NotFound(w, r)
		return
	}

	rendered, err := p.engine.RenderPage(content)
	if err != nil {
		slog.Error("render page failed", "error", err, "slug", slugParam)
		// Fall back to a safe error page when the template engine fails.
		// Never render raw user content — it bypasses html/template escaping.
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		safeTitle := html.EscapeString(content.Title)
		w.Write([]byte(`<!DOCTYPE html><html><head><title>` + safeTitle + `</title>
<script src="https://cdn.tailwindcss.com"></script></head>
<body class="bg-gray-100 flex items-center justify-center min-h-screen">
<div class="text-center">
<h1 class="text-3xl font-bold text-gray-900">` + safeTitle + `</h1>
<p class="mt-2 text-gray-500">This page could not be rendered. Please check your templates.</p>
<a href="/" class="mt-4 inline-block text-indigo-600 hover:text-indigo-800 text-sm">Go to Homepage</a>
</div></body></html>`))
		return
	}

	// Store in L2 cache.
	p.pageCache.Set(ctx, cache.SlugKey(slugParam), rendered)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(rendered)
}
