// Package handlers contains the HTTP handlers for the SmartPress CMS.
// Handlers are grouped by concern (admin, public, auth) and receive
// their dependencies through the handler struct.
package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"smartpress/internal/middleware"
	"smartpress/internal/models"
	"smartpress/internal/render"
	"smartpress/internal/session"
	"smartpress/internal/slug"
	"smartpress/internal/store"
)

// Admin groups all admin panel HTTP handlers and their dependencies.
type Admin struct {
	renderer     *render.Renderer
	sessions     *session.Store
	contentStore *store.ContentStore
	userStore    *store.UserStore
}

// NewAdmin creates a new Admin handler group with the given dependencies.
func NewAdmin(renderer *render.Renderer, sessions *session.Store, contentStore *store.ContentStore, userStore *store.UserStore) *Admin {
	return &Admin{
		renderer:     renderer,
		sessions:     sessions,
		contentStore: contentStore,
		userStore:    userStore,
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

	_, err := a.contentStore.Create(c)
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

	if err := a.contentStore.Delete(id); err != nil {
		slog.Error("delete content failed", "error", err)
	}

	http.Redirect(w, r, "/admin/"+section, http.StatusSeeOther)
}

// --- Templates & Users list pages ---

// TemplatesList renders the templates/AI design page.
func (a *Admin) TemplatesList(w http.ResponseWriter, r *http.Request) {
	a.renderer.Page(w, r, "templates_list", &render.PageData{
		Title:   "AI Design",
		Section: "templates",
	})
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

// SettingsPage renders the settings page.
func (a *Admin) SettingsPage(w http.ResponseWriter, r *http.Request) {
	a.renderer.Page(w, r, "settings", &render.PageData{
		Title:   "Settings",
		Section: "settings",
	})
}
