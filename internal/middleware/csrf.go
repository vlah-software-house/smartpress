package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
)

const (
	// csrfTokenLength is the byte length of CSRF tokens (32 bytes = 64 hex chars).
	csrfTokenLength = 32

	// CSRFCookieName is the cookie that holds the CSRF token.
	CSRFCookieName = "sp_csrf"

	// CSRFHeaderName is the header HTMX sends the CSRF token in.
	// Configured via hx-headers in the admin layout.
	CSRFHeaderName = "X-CSRF-Token"

	// CSRFFormField is the hidden form field name for non-HTMX forms.
	CSRFFormField = "csrf_token"
)

// CSRF provides double-submit cookie CSRF protection. It generates a
// token stored in a cookie and validates that subsequent state-changing
// requests (POST, PUT, PATCH, DELETE) include the same token as a header
// or form field.
//
// This works well with HTMX: the admin layout sets hx-headers with the
// CSRF token so all HTMX requests include it automatically.
func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ensure a CSRF token cookie exists.
		cookie, err := r.Cookie(CSRFCookieName)
		if err != nil || cookie.Value == "" {
			token, err := generateCSRFToken()
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:     CSRFCookieName,
				Value:    token,
				Path:     "/",
				HttpOnly: false, // JS needs to read this for HTMX hx-headers
				Secure:   false, // Set to true behind TLS
				SameSite: http.SameSiteStrictMode,
			})
			cookie = &http.Cookie{Value: token}
		}

		// Safe methods don't need CSRF validation.
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// For state-changing methods, validate the token.
		// Check header first (HTMX), then form field.
		submitted := r.Header.Get(CSRFHeaderName)
		if submitted == "" {
			submitted = r.FormValue(CSRFFormField)
		}

		if subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(submitted)) != 1 {
			http.Error(w, "CSRF token mismatch", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetCSRFToken extracts the current CSRF token from the request cookie.
// Used in templates to populate hidden fields and HTMX headers.
func GetCSRFToken(r *http.Request) string {
	cookie, err := r.Cookie(CSRFCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// generateCSRFToken creates a cryptographically random token.
func generateCSRFToken() (string, error) {
	b := make([]byte, csrfTokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

