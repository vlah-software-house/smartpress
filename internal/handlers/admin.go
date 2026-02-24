// Package handlers contains the HTTP handlers for the SmartPress CMS.
// Handlers are grouped by concern (admin, public, auth) and receive
// their dependencies through the handler struct.
package handlers

import (
	"database/sql"
	"net/http"

	"smartpress/internal/render"
	"smartpress/internal/session"
)

// Admin groups all admin panel HTTP handlers and their dependencies.
type Admin struct {
	renderer *render.Renderer
	sessions *session.Store
	db       *sql.DB
}

// NewAdmin creates a new Admin handler group with the given dependencies.
func NewAdmin(renderer *render.Renderer, sessions *session.Store, db *sql.DB) *Admin {
	return &Admin{
		renderer: renderer,
		sessions: sessions,
		db:       db,
	}
}

// Dashboard renders the admin dashboard page.
func (a *Admin) Dashboard(w http.ResponseWriter, r *http.Request) {
	a.renderer.Page(w, r, "dashboard", &render.PageData{
		Title:   "Dashboard",
		Section: "dashboard",
	})
}

// LoginPage renders the login form.
func (a *Admin) LoginPage(w http.ResponseWriter, r *http.Request) {
	a.renderer.Page(w, r, "login", &render.PageData{
		Title: "Sign In",
	})
}

// PostsList renders the posts management page.
func (a *Admin) PostsList(w http.ResponseWriter, r *http.Request) {
	a.renderer.Page(w, r, "posts_list", &render.PageData{
		Title:   "Posts",
		Section: "posts",
	})
}

// PagesList renders the pages management page.
func (a *Admin) PagesList(w http.ResponseWriter, r *http.Request) {
	a.renderer.Page(w, r, "pages_list", &render.PageData{
		Title:   "Pages",
		Section: "pages",
	})
}

// TemplatesList renders the templates/AI design page.
func (a *Admin) TemplatesList(w http.ResponseWriter, r *http.Request) {
	a.renderer.Page(w, r, "templates_list", &render.PageData{
		Title:   "AI Design",
		Section: "templates",
	})
}

// UsersList renders the user management page (admin only).
func (a *Admin) UsersList(w http.ResponseWriter, r *http.Request) {
	a.renderer.Page(w, r, "users_list", &render.PageData{
		Title:   "Users",
		Section: "users",
	})
}

// SettingsPage renders the settings page.
func (a *Admin) SettingsPage(w http.ResponseWriter, r *http.Request) {
	a.renderer.Page(w, r, "settings", &render.PageData{
		Title:   "Settings",
		Section: "settings",
	})
}
