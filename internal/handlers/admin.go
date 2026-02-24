// Package handlers contains the HTTP handlers for the SmartPress CMS.
// Handlers are grouped by concern (admin, public, auth) and receive
// their dependencies through the handler struct.
package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"smartpress/internal/cache"
	"smartpress/internal/engine"
	"smartpress/internal/middleware"
	"smartpress/internal/models"
	"smartpress/internal/render"
	"smartpress/internal/session"
	"smartpress/internal/slug"
	"smartpress/internal/store"
)

// AIProviderInfo holds display information about a configured AI provider.
// Used by the Settings page to show which providers are available.
type AIProviderInfo struct {
	Name      string // "openai", "gemini", "claude", "mistral"
	Label     string // Human-friendly label
	HasKey    bool   // Whether an API key is configured
	Active    bool   // Whether this is the currently active provider
	Model     string // Configured model name
	KeyEnvVar string // Environment variable name for the key
}

// AIConfig holds the AI provider configuration visible to admin handlers.
// Intentionally excludes actual API keys — only exposes what the UI needs.
type AIConfig struct {
	ActiveProvider string
	Providers      []AIProviderInfo
}

// Admin groups all admin panel HTTP handlers and their dependencies.
type Admin struct {
	renderer      *render.Renderer
	sessions      *session.Store
	contentStore  *store.ContentStore
	userStore     *store.UserStore
	templateStore *store.TemplateStore
	engine        *engine.Engine
	pageCache     *cache.PageCache
	cacheLog      *store.CacheLogStore
	aiConfig      *AIConfig
}

// NewAdmin creates a new Admin handler group with the given dependencies.
func NewAdmin(renderer *render.Renderer, sessions *session.Store, contentStore *store.ContentStore, userStore *store.UserStore, templateStore *store.TemplateStore, eng *engine.Engine, pageCache *cache.PageCache, cacheLog *store.CacheLogStore, aiCfg *AIConfig) *Admin {
	return &Admin{
		renderer:      renderer,
		sessions:      sessions,
		contentStore:  contentStore,
		userStore:     userStore,
		templateStore: templateStore,
		engine:        eng,
		pageCache:     pageCache,
		cacheLog:      cacheLog,
		aiConfig:      aiCfg,
	}
}

// Dashboard renders the admin dashboard page with real stats.
func (a *Admin) Dashboard(w http.ResponseWriter, r *http.Request) {
	postCount, _ := a.contentStore.CountByType(models.ContentTypePost)
	pageCount, _ := a.contentStore.CountByType(models.ContentTypePage)
	users, _ := a.userStore.List()

	a.renderer.Page(w, r, "dashboard", &render.PageData{
		Title:   "Dashboard",
		Section: "dashboard",
		Data: map[string]any{
			"PostCount": postCount,
			"PageCount": pageCount,
			"UserCount": len(users),
		},
	})
}

// --- Posts CRUD ---

// PostsList renders the posts management page.
func (a *Admin) PostsList(w http.ResponseWriter, r *http.Request) {
	posts, err := a.contentStore.ListByType(models.ContentTypePost)
	if err != nil {
		slog.Error("list posts failed", "error", err)
	}

	a.renderer.Page(w, r, "posts_list", &render.PageData{
		Title:   "Posts",
		Section: "posts",
		Data:    map[string]any{"Items": posts},
	})
}

// PostNew renders the new post form.
func (a *Admin) PostNew(w http.ResponseWriter, r *http.Request) {
	a.renderer.Page(w, r, "content_form", &render.PageData{
		Title:   "New Post",
		Section: "posts",
		Data: map[string]any{
			"ContentType": "post",
			"IsNew":       true,
		},
	})
}

// PostCreate handles the new post form submission.
func (a *Admin) PostCreate(w http.ResponseWriter, r *http.Request) {
	sess := middleware.SessionFromCtx(r.Context())
	a.createContent(w, r, models.ContentTypePost, sess)
}

// PostEdit renders the edit post form.
func (a *Admin) PostEdit(w http.ResponseWriter, r *http.Request) {
	a.editContent(w, r, "posts")
}

// PostUpdate handles the edit post form submission.
func (a *Admin) PostUpdate(w http.ResponseWriter, r *http.Request) {
	a.updateContent(w, r, "posts")
}

// PostDelete handles post deletion.
func (a *Admin) PostDelete(w http.ResponseWriter, r *http.Request) {
	a.deleteContent(w, r, "posts")
}

// --- Pages CRUD ---

// PagesList renders the pages management page.
func (a *Admin) PagesList(w http.ResponseWriter, r *http.Request) {
	pages, err := a.contentStore.ListByType(models.ContentTypePage)
	if err != nil {
		slog.Error("list pages failed", "error", err)
	}

	a.renderer.Page(w, r, "pages_list", &render.PageData{
		Title:   "Pages",
		Section: "pages",
		Data:    map[string]any{"Items": pages},
	})
}

// PageNew renders the new page form.
func (a *Admin) PageNew(w http.ResponseWriter, r *http.Request) {
	a.renderer.Page(w, r, "content_form", &render.PageData{
		Title:   "New Page",
		Section: "pages",
		Data: map[string]any{
			"ContentType": "page",
			"IsNew":       true,
		},
	})
}

// PageCreate handles the new page form submission.
func (a *Admin) PageCreate(w http.ResponseWriter, r *http.Request) {
	sess := middleware.SessionFromCtx(r.Context())
	a.createContent(w, r, models.ContentTypePage, sess)
}

// PageEdit renders the edit page form.
func (a *Admin) PageEdit(w http.ResponseWriter, r *http.Request) {
	a.editContent(w, r, "pages")
}

// PageUpdate handles the edit page form submission.
func (a *Admin) PageUpdate(w http.ResponseWriter, r *http.Request) {
	a.updateContent(w, r, "pages")
}

// PageDelete handles page deletion.
func (a *Admin) PageDelete(w http.ResponseWriter, r *http.Request) {
	a.deleteContent(w, r, "pages")
}

// --- Shared content helpers ---

// createContent handles creating a new post or page from the form.
func (a *Admin) createContent(w http.ResponseWriter, r *http.Request, contentType models.ContentType, sess *session.Data) {
	title := r.FormValue("title")
	body := r.FormValue("body")
	status := models.ContentStatus(r.FormValue("status"))
	contentSlug := r.FormValue("slug")
	excerpt := r.FormValue("excerpt")
	metaDesc := r.FormValue("meta_description")
	metaKw := r.FormValue("meta_keywords")

	if contentSlug == "" {
		contentSlug = slug.Generate(title)
	}

	if status == "" {
		status = models.ContentStatusDraft
	}

	c := &models.Content{
		Type:     contentType,
		Title:    title,
		Slug:     contentSlug,
		Body:     body,
		Status:   status,
		AuthorID: sess.UserID,
	}
	if excerpt != "" {
		c.Excerpt = &excerpt
	}
	if metaDesc != "" {
		c.MetaDescription = &metaDesc
	}
	if metaKw != "" {
		c.MetaKeywords = &metaKw
	}

	created, err := a.contentStore.Create(c)
	if err != nil {
		slog.Error("create content failed", "error", err, "type", contentType)
		section := "posts"
		if contentType == models.ContentTypePage {
			section = "pages"
		}
		a.renderer.Page(w, r, "content_form", &render.PageData{
			Title:   "New " + string(contentType),
			Section: section,
			Data: map[string]any{
				"ContentType": string(contentType),
				"IsNew":       true,
				"Error":       "Failed to create. The slug may already exist.",
				"Item":        c,
			},
		})
		return
	}

	// Invalidate cache for the new content (homepage may show it in listings).
	a.invalidateContentCache(r.Context(), created.ID, created.Slug, "create")

	if contentType == models.ContentTypePage {
		http.Redirect(w, r, "/admin/pages", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
	}
}

// editContent renders the edit form for a content item.
func (a *Admin) editContent(w http.ResponseWriter, r *http.Request, section string) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, err := a.contentStore.FindByID(id)
	if err != nil || item == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	contentType := "post"
	title := "Edit Post"
	if section == "pages" {
		contentType = "page"
		title = "Edit Page"
	}

	a.renderer.Page(w, r, "content_form", &render.PageData{
		Title:   title,
		Section: section,
		Data: map[string]any{
			"ContentType": contentType,
			"IsNew":       false,
			"Item":        item,
		},
	})
}

// updateContent handles the edit form submission for a content item.
func (a *Admin) updateContent(w http.ResponseWriter, r *http.Request, section string) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, err := a.contentStore.FindByID(id)
	if err != nil || item == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	item.Title = r.FormValue("title")
	item.Body = r.FormValue("body")
	item.Status = models.ContentStatus(r.FormValue("status"))
	item.Slug = r.FormValue("slug")

	if item.Slug == "" {
		item.Slug = slug.Generate(item.Title)
	}

	excerpt := r.FormValue("excerpt")
	metaDesc := r.FormValue("meta_description")
	metaKw := r.FormValue("meta_keywords")

	if excerpt != "" {
		item.Excerpt = &excerpt
	} else {
		item.Excerpt = nil
	}
	if metaDesc != "" {
		item.MetaDescription = &metaDesc
	} else {
		item.MetaDescription = nil
	}
	if metaKw != "" {
		item.MetaKeywords = &metaKw
	} else {
		item.MetaKeywords = nil
	}

	if err := a.contentStore.Update(item); err != nil {
		slog.Error("update content failed", "error", err)
		a.renderer.Page(w, r, "content_form", &render.PageData{
			Title:   "Edit",
			Section: section,
			Data: map[string]any{
				"ContentType": string(item.Type),
				"IsNew":       false,
				"Item":        item,
				"Error":       "Failed to update. The slug may already exist.",
			},
		})
		return
	}

	a.invalidateContentCache(r.Context(), item.ID, item.Slug, "update")
	http.Redirect(w, r, "/admin/"+section, http.StatusSeeOther)
}

// deleteContent handles content deletion.
func (a *Admin) deleteContent(w http.ResponseWriter, r *http.Request, section string) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Look up the slug before deleting so we can invalidate its cache entry.
	item, _ := a.contentStore.FindByID(id)

	if err := a.contentStore.Delete(id); err != nil {
		slog.Error("delete content failed", "error", err)
	} else if item != nil {
		a.invalidateContentCache(r.Context(), id, item.Slug, "delete")
	}

	http.Redirect(w, r, "/admin/"+section, http.StatusSeeOther)
}

// --- Template management ---

// TemplatesList renders the templates management page with real data.
func (a *Admin) TemplatesList(w http.ResponseWriter, r *http.Request) {
	templates, err := a.templateStore.List()
	if err != nil {
		slog.Error("list templates failed", "error", err)
	}

	a.renderer.Page(w, r, "templates_list", &render.PageData{
		Title:   "AI Design",
		Section: "templates",
		Data:    map[string]any{"Templates": templates},
	})
}

// TemplateNew renders the new template form.
func (a *Admin) TemplateNew(w http.ResponseWriter, r *http.Request) {
	a.renderer.Page(w, r, "template_form", &render.PageData{
		Title:   "New Template",
		Section: "templates",
		Data:    map[string]any{"IsNew": true},
	})
}

// TemplateCreate handles the new template form submission.
func (a *Admin) TemplateCreate(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	tmplType := models.TemplateType(r.FormValue("type"))
	htmlContent := r.FormValue("html_content")

	// Validate the template syntax before saving.
	if err := a.engine.ValidateTemplate(htmlContent); err != nil {
		a.renderer.Page(w, r, "template_form", &render.PageData{
			Title:   "New Template",
			Section: "templates",
			Data: map[string]any{
				"IsNew": true,
				"Error": "Template syntax error: " + err.Error(),
				"Item":  &models.Template{Name: name, Type: tmplType, HTMLContent: htmlContent},
			},
		})
		return
	}

	t := &models.Template{
		Name:        name,
		Type:        tmplType,
		HTMLContent: htmlContent,
	}

	created, err := a.templateStore.Create(t)
	if err != nil {
		slog.Error("create template failed", "error", err)
		a.renderer.Page(w, r, "template_form", &render.PageData{
			Title:   "New Template",
			Section: "templates",
			Data: map[string]any{
				"IsNew": true,
				"Error": "Failed to create template.",
				"Item":  t,
			},
		})
		return
	}

	// New templates aren't active yet, but log the event for auditing.
	a.cacheLog.Log("template", created.ID, "create")
	http.Redirect(w, r, "/admin/templates", http.StatusSeeOther)
}

// TemplateEdit renders the edit template form.
func (a *Admin) TemplateEdit(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, err := a.templateStore.FindByID(id)
	if err != nil || item == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	a.renderer.Page(w, r, "template_form", &render.PageData{
		Title:   "Edit Template",
		Section: "templates",
		Data: map[string]any{
			"IsNew": false,
			"Item":  item,
		},
	})
}

// TemplateUpdate handles the edit template form submission.
func (a *Admin) TemplateUpdate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, err := a.templateStore.FindByID(id)
	if err != nil || item == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	item.Name = r.FormValue("name")
	htmlContent := r.FormValue("html_content")

	// Validate syntax.
	if err := a.engine.ValidateTemplate(htmlContent); err != nil {
		a.renderer.Page(w, r, "template_form", &render.PageData{
			Title:   "Edit Template",
			Section: "templates",
			Data: map[string]any{
				"IsNew": false,
				"Error": "Template syntax error: " + err.Error(),
				"Item":  item,
			},
		})
		return
	}

	item.HTMLContent = htmlContent
	if err := a.templateStore.Update(item); err != nil {
		slog.Error("update template failed", "error", err)
	} else {
		// Template content changed — invalidate L1 (compiled) and L2 (rendered pages).
		a.invalidateTemplateCache(r.Context(), item.ID, "update")
	}

	http.Redirect(w, r, "/admin/templates", http.StatusSeeOther)
}

// TemplateActivate sets a template as active for its type.
func (a *Admin) TemplateActivate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := a.templateStore.Activate(id); err != nil {
		slog.Error("activate template failed", "error", err)
	} else {
		// Activation changes which template renders for a type — clear everything.
		a.invalidateAllTemplateCache(r.Context(), id, "update")
	}

	http.Redirect(w, r, "/admin/templates", http.StatusSeeOther)
}

// TemplateDelete handles template deletion.
func (a *Admin) TemplateDelete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := a.templateStore.Delete(id); err != nil {
		slog.Error("delete template failed", "error", err)
	} else {
		a.invalidateTemplateCache(r.Context(), id, "delete")
	}

	http.Redirect(w, r, "/admin/templates", http.StatusSeeOther)
}

// TemplatePreview renders a preview of a template with dummy data.
func (a *Admin) TemplatePreview(w http.ResponseWriter, r *http.Request) {
	htmlContent := r.FormValue("html_content")
	if htmlContent == "" {
		http.Error(w, "No template content", http.StatusBadRequest)
		return
	}

	data := engine.PageData{
		SiteName:        "SmartPress",
		Title:           "Preview Page Title",
		Body:            "<p>This is preview content. Lorem ipsum dolor sit amet, consectetur adipiscing elit.</p>",
		Excerpt:         "A brief preview excerpt.",
		MetaDescription: "Preview meta description",
		Slug:            "preview-page",
		PublishedAt:     "February 24, 2026",
		Header:          "<header><nav>Header Preview</nav></header>",
		Footer:          "<footer><p>Footer Preview</p></footer>",
		Year:            2026,
	}

	result, err := a.engine.ValidateAndRender(htmlContent, data)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `<div class="p-4 bg-red-50 border border-red-200 rounded text-red-800 text-sm">Template error: %s</div>`, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(result)
}

// UsersList renders the user management page with real data.
func (a *Admin) UsersList(w http.ResponseWriter, r *http.Request) {
	users, err := a.userStore.List()
	if err != nil {
		slog.Error("list users failed", "error", err)
	}

	a.renderer.Page(w, r, "users_list", &render.PageData{
		Title:   "Users",
		Section: "users",
		Data:    map[string]any{"Users": users},
	})
}

// UserResetTwoFA resets another user's 2FA, forcing re-setup on next login.
func (a *Admin) UserResetTwoFA(w http.ResponseWriter, r *http.Request) {
	sess := middleware.SessionFromCtx(r.Context())

	idStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Cannot reset your own 2FA.
	if targetID == sess.UserID {
		http.Error(w, "Cannot reset your own 2FA", http.StatusForbidden)
		return
	}

	if err := a.userStore.ResetTOTP(targetID); err != nil {
		slog.Error("reset 2fa failed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	slog.Info("2fa reset by admin", "admin", sess.Email, "target_user", targetID)
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// --- Cache invalidation helpers ---

// invalidateContentCache purges the L2 page cache for a content item and
// logs the event. Always invalidates the homepage too since post listings
// or the "home" page might have changed.
func (a *Admin) invalidateContentCache(ctx context.Context, contentID uuid.UUID, contentSlug, action string) {
	a.pageCache.InvalidatePage(ctx, cache.SlugKey(contentSlug))
	a.pageCache.InvalidateHomepage(ctx)
	a.cacheLog.Log("content", contentID, action)
}

// invalidateTemplateCache purges both L1 (compiled template) and L2 (all
// rendered pages) caches, since any template change can affect any page.
func (a *Admin) invalidateTemplateCache(ctx context.Context, templateID uuid.UUID, action string) {
	a.engine.InvalidateTemplate(templateID.String())
	a.pageCache.InvalidateAll(ctx)
	a.cacheLog.Log("template", templateID, action)
}

// invalidateAllTemplateCache clears the entire L1 cache and all L2 pages.
// Used for template activation which changes the active template for a type.
func (a *Admin) invalidateAllTemplateCache(ctx context.Context, templateID uuid.UUID, action string) {
	a.engine.InvalidateAllTemplates()
	a.pageCache.InvalidateAll(ctx)
	a.cacheLog.Log("template", templateID, action)
}

// SettingsPage renders the settings page.
func (a *Admin) SettingsPage(w http.ResponseWriter, r *http.Request) {
	a.renderer.Page(w, r, "settings", &render.PageData{
		Title:   "Settings",
		Section: "settings",
		Data: map[string]any{
			"Providers": a.aiConfig.Providers,
		},
	})
}
