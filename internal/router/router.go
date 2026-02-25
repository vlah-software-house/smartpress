// Package router sets up all HTTP routes and middleware chains for the
// SmartPress CMS. It organizes routes into public and admin groups with
// appropriate middleware stacks.
package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"smartpress/internal/handlers"
	"smartpress/internal/middleware"
	"smartpress/internal/session"
)

// New creates and returns the configured Chi router with all middleware
// and route groups wired up.
func New(sessionStore *session.Store, admin *handlers.Admin, auth *handlers.Auth, public *handlers.Public) chi.Router {
	r := chi.NewRouter()

	// Global middleware — applied to every request.
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.LoadSession(sessionStore))

	// Health check — no auth, no CSRF.
	r.Get("/health", healthHandler)

	// Admin routes — require authentication and CSRF protection.
	r.Route("/admin", func(r chi.Router) {
		r.Use(middleware.CSRF)

		// Auth pages — accessible without a session.
		r.Get("/login", auth.LoginPage)
		r.Post("/login", auth.LoginSubmit)
		r.Post("/logout", auth.Logout)

		// 2FA — requires auth but NOT completed 2FA.
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth)
			r.Get("/2fa/setup", auth.TwoFASetupPage)
			r.Get("/2fa/verify", auth.TwoFAVerifyPage)
			r.Post("/2fa/verify", auth.TwoFAVerifySubmit)
		})

		// Authenticated + 2FA-verified admin area.
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth)
			r.Use(middleware.Require2FA)

			// Dashboard
			r.Get("/", admin.Dashboard)
			r.Get("/dashboard", admin.Dashboard)

			// Posts
			r.Route("/posts", func(r chi.Router) {
				r.Get("/", admin.PostsList)
				r.Get("/new", admin.PostNew)
				r.Post("/", admin.PostCreate)
				r.Get("/{id}", admin.PostEdit)
				r.Put("/{id}", admin.PostUpdate)
				r.Delete("/{id}", admin.PostDelete)
			})

			// Pages
			r.Route("/pages", func(r chi.Router) {
				r.Get("/", admin.PagesList)
				r.Get("/new", admin.PageNew)
				r.Post("/", admin.PageCreate)
				r.Get("/{id}", admin.PageEdit)
				r.Put("/{id}", admin.PageUpdate)
				r.Delete("/{id}", admin.PageDelete)
			})

			// Templates (AI Design)
			r.Route("/templates", func(r chi.Router) {
				r.Get("/", admin.TemplatesList)
				r.Get("/new", admin.TemplateNew)
				r.Post("/", admin.TemplateCreate)
				r.Post("/preview", admin.TemplatePreview)
				r.Get("/{id}", admin.TemplateEdit)
				r.Put("/{id}", admin.TemplateUpdate)
				r.Delete("/{id}", admin.TemplateDelete)
				r.Post("/{id}/activate", admin.TemplateActivate)
			})

			// User management — admin only
			r.Route("/users", func(r chi.Router) {
				r.Use(middleware.RequireAdmin)
				r.Get("/", admin.UsersList)
				r.Post("/{id}/reset-2fa", admin.UserResetTwoFA)
			})

			// AI Assistant (content editor helpers)
			r.Route("/ai", func(r chi.Router) {
				r.Post("/suggest-title", admin.AISuggestTitle)
				r.Post("/generate-excerpt", admin.AIGenerateExcerpt)
				r.Post("/seo-metadata", admin.AISEOMetadata)
				r.Post("/rewrite", admin.AIRewrite)
				r.Post("/extract-tags", admin.AIExtractTags)
			})

			// Settings
			r.Get("/settings", admin.SettingsPage)
		})
	})

	// Public routes — served by the dynamic template engine.
	r.Get("/", public.Homepage)
	r.Get("/{slug}", public.Page)

	return r
}

// healthHandler returns a simple JSON health check response.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
