// Package router sets up all HTTP routes and middleware chains for the
// SmartPress CMS. It organizes routes into public and admin groups with
// appropriate middleware stacks.
package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"smartpress/internal/middleware"
	"smartpress/internal/session"
)

// New creates and returns the configured Chi router with all middleware
// and route groups wired up.
func New(sessionStore *session.Store) chi.Router {
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
		r.Get("/login", placeholderHandler("Login Page"))
		r.Post("/login", placeholderHandler("Login Submit"))
		r.Post("/logout", placeholderHandler("Logout"))

		// 2FA setup — requires auth but NOT completed 2FA (chicken-and-egg).
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth)
			r.Get("/2fa/setup", placeholderHandler("2FA Setup Page"))
			r.Post("/2fa/setup", placeholderHandler("2FA Setup Submit"))
			r.Post("/2fa/verify", placeholderHandler("2FA Verify"))
		})

		// Authenticated + 2FA-verified admin area.
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth)
			r.Use(middleware.Require2FA)

			// Dashboard
			r.Get("/", placeholderHandler("Dashboard"))
			r.Get("/dashboard", placeholderHandler("Dashboard"))

			// Content management (posts + pages)
			r.Route("/posts", func(r chi.Router) {
				r.Get("/", placeholderHandler("List Posts"))
				r.Get("/new", placeholderHandler("New Post"))
				r.Post("/", placeholderHandler("Create Post"))
				r.Get("/{id}", placeholderHandler("Edit Post"))
				r.Put("/{id}", placeholderHandler("Update Post"))
				r.Delete("/{id}", placeholderHandler("Delete Post"))
			})

			r.Route("/pages", func(r chi.Router) {
				r.Get("/", placeholderHandler("List Pages"))
				r.Get("/new", placeholderHandler("New Page"))
				r.Post("/", placeholderHandler("Create Page"))
				r.Get("/{id}", placeholderHandler("Edit Page"))
				r.Put("/{id}", placeholderHandler("Update Page"))
				r.Delete("/{id}", placeholderHandler("Delete Page"))
			})

			// Template management (AI Design)
			r.Route("/templates", func(r chi.Router) {
				r.Get("/", placeholderHandler("List Templates"))
				r.Get("/new", placeholderHandler("New Template"))
				r.Post("/", placeholderHandler("Create Template"))
				r.Get("/{id}", placeholderHandler("Edit Template"))
				r.Put("/{id}", placeholderHandler("Update Template"))
				r.Delete("/{id}", placeholderHandler("Delete Template"))
			})

			// User management — admin only
			r.Route("/users", func(r chi.Router) {
				r.Use(middleware.RequireAdmin)
				r.Get("/", placeholderHandler("List Users"))
				r.Get("/new", placeholderHandler("New User"))
				r.Post("/", placeholderHandler("Create User"))
				r.Get("/{id}", placeholderHandler("Edit User"))
				r.Put("/{id}", placeholderHandler("Update User"))
				r.Delete("/{id}", placeholderHandler("Delete User"))
				r.Post("/{id}/reset-2fa", placeholderHandler("Reset User 2FA"))
			})

			// Settings
			r.Get("/settings", placeholderHandler("Settings"))
			r.Put("/settings", placeholderHandler("Update Settings"))
		})
	})

	// Public routes — served by the dynamic template engine.
	r.Group(func(r chi.Router) {
		r.Get("/", placeholderHandler("Homepage"))
		r.Get("/{slug}", placeholderHandler("Public Page"))
	})

	return r
}

// healthHandler returns a simple JSON health check response.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// placeholderHandler returns a handler that displays a placeholder page
// name. These will be replaced with real handlers in subsequent steps.
func placeholderHandler(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<h1>SmartPress — " + name + "</h1><p>Coming soon.</p>"))
	}
}
