package handlers

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/pquerna/otp/totp"
	qrcode "github.com/skip2/go-qrcode"

	"smartpress/internal/middleware"
	"smartpress/internal/render"
	"smartpress/internal/session"
	"smartpress/internal/store"
)

// Auth groups all authentication-related HTTP handlers.
type Auth struct {
	renderer  *render.Renderer
	sessions  *session.Store
	userStore *store.UserStore
}

// NewAuth creates a new Auth handler group.
func NewAuth(renderer *render.Renderer, sessions *session.Store, userStore *store.UserStore) *Auth {
	return &Auth{
		renderer:  renderer,
		sessions:  sessions,
		userStore: userStore,
	}
}

// LoginPage renders the login form.
func (a *Auth) LoginPage(w http.ResponseWriter, r *http.Request) {
	// If already logged in with 2FA complete, redirect to dashboard.
	sess := middleware.SessionFromCtx(r.Context())
	if sess != nil && sess.TwoFADone {
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	a.renderer.Page(w, r, "login", &render.PageData{
		Title: "Sign In",
	})
}

// LoginSubmit processes the login form.
func (a *Auth) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Find the user by email.
	user, err := a.userStore.FindByEmail(email)
	if err != nil {
		slog.Error("login lookup failed", "error", err)
		a.renderer.Page(w, r, "login", &render.PageData{
			Title: "Sign In",
			Data:  map[string]any{"Error": "An unexpected error occurred."},
		})
		return
	}

	// Validate credentials.
	if user == nil || !a.userStore.CheckPassword(user, password) {
		a.renderer.Page(w, r, "login", &render.PageData{
			Title: "Sign In",
			Data:  map[string]any{"Error": "Invalid email or password."},
		})
		return
	}

	// Create a session. TwoFADone starts as false — user must complete 2FA.
	_, err = a.sessions.Create(r.Context(), w, &session.Data{
		UserID:      user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        string(user.Role),
		TwoFADone:   false,
	})
	if err != nil {
		slog.Error("session create failed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Route based on 2FA status:
	// - Not set up yet → go to setup page
	// - Already set up → go to verification page
	if user.Needs2FASetup() {
		http.Redirect(w, r, "/admin/2fa/setup", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/admin/2fa/verify", http.StatusSeeOther)
	}
}

// TwoFASetupPage generates a TOTP secret and displays the QR code.
func (a *Auth) TwoFASetupPage(w http.ResponseWriter, r *http.Request) {
	sess := middleware.SessionFromCtx(r.Context())
	if sess == nil {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	// Generate a new TOTP key.
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "SmartPress",
		AccountName: sess.Email,
	})
	if err != nil {
		slog.Error("totp generate failed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Save the secret to the database.
	if err := a.userStore.SetTOTPSecret(sess.UserID, key.Secret()); err != nil {
		slog.Error("save totp secret failed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Generate QR code as base64-encoded PNG.
	qrPNG, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		slog.Error("qr code generation failed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	qrBase64 := base64.StdEncoding.EncodeToString(qrPNG)

	a.renderer.Page(w, r, "2fa_setup", &render.PageData{
		Title: "Set Up Two-Factor Authentication",
		Data: map[string]any{
			"QRCode": qrBase64,
			"Secret": key.Secret(),
		},
	})
}

// TwoFAVerifyPage renders the 2FA code entry form (for users who already have 2FA set up).
func (a *Auth) TwoFAVerifyPage(w http.ResponseWriter, r *http.Request) {
	sess := middleware.SessionFromCtx(r.Context())
	if sess == nil {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	a.renderer.Page(w, r, "2fa_verify", &render.PageData{
		Title: "Two-Factor Authentication",
	})
}

// TwoFAVerifySubmit validates the TOTP code and completes authentication.
func (a *Auth) TwoFAVerifySubmit(w http.ResponseWriter, r *http.Request) {
	sess := middleware.SessionFromCtx(r.Context())
	if sess == nil {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	code := r.FormValue("code")

	// Look up the user's TOTP secret.
	user, err := a.userStore.FindByID(sess.UserID)
	if err != nil || user == nil {
		slog.Error("user lookup for 2fa failed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if user.TOTPSecret == nil {
		http.Redirect(w, r, "/admin/2fa/setup", http.StatusSeeOther)
		return
	}

	// Validate the TOTP code.
	valid := totp.Validate(code, *user.TOTPSecret)
	if !valid {
		// Determine which page to show the error on.
		templateName := "2fa_verify"
		if !user.TOTPEnabled {
			templateName = "2fa_setup"

			// Re-generate QR code for the setup page.
			qrPNG, _ := qrcode.Encode(
				fmt.Sprintf("otpauth://totp/SmartPress:%s?secret=%s&issuer=SmartPress", user.Email, *user.TOTPSecret),
				qrcode.Medium, 256,
			)
			qrBase64 := base64.StdEncoding.EncodeToString(qrPNG)

			a.renderer.Page(w, r, templateName, &render.PageData{
				Title: "Set Up Two-Factor Authentication",
				Data: map[string]any{
					"Error":  "Invalid code. Please try again.",
					"QRCode": qrBase64,
					"Secret": *user.TOTPSecret,
				},
			})
			return
		}

		a.renderer.Page(w, r, templateName, &render.PageData{
			Title: "Two-Factor Authentication",
			Data:  map[string]any{"Error": "Invalid code. Please try again."},
		})
		return
	}

	// If this is the first-time setup, enable TOTP in the database.
	if !user.TOTPEnabled {
		if err := a.userStore.EnableTOTP(user.ID); err != nil {
			slog.Error("enable totp failed", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	// Mark 2FA as complete in the session.
	sess.TwoFADone = true
	if err := a.sessions.Update(r.Context(), r, sess); err != nil {
		slog.Error("session update failed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

// Logout destroys the session and redirects to the login page.
func (a *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	a.sessions.Destroy(r.Context(), w, r)
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}
