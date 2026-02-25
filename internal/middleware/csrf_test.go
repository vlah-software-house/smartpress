package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewCSRFSecureFlag(t *testing.T) {
	tests := []struct {
		name   string
		secure bool
	}{
		{"secure true", true},
		{"secure false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csrf := NewCSRF(tt.secure)
			handler := csrf(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/admin/login", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Check that a CSRF cookie was set.
			cookies := rr.Result().Cookies()
			found := false
			for _, c := range cookies {
				if c.Name == CSRFCookieName {
					found = true
					if c.Secure != tt.secure {
						t.Errorf("cookie Secure: got %v, want %v", c.Secure, tt.secure)
					}
					if c.SameSite != http.SameSiteStrictMode {
						t.Errorf("cookie SameSite: got %v, want StrictMode", c.SameSite)
					}
					if c.Value == "" {
						t.Error("cookie Value should not be empty")
					}
				}
			}
			if !found {
				t.Error("CSRF cookie not set")
			}
		})
	}
}

func TestCSRFRejectsStateMutationWithoutToken(t *testing.T) {
	csrf := NewCSRF(false)
	handler := csrf(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First GET to get a token.
	getReq := httptest.NewRequest(http.MethodGet, "/admin/login", nil)
	getRR := httptest.NewRecorder()
	handler.ServeHTTP(getRR, getReq)

	// POST without token should be rejected.
	postReq := httptest.NewRequest(http.MethodPost, "/admin/login", nil)
	// Include the CSRF cookie from the GET response.
	for _, c := range getRR.Result().Cookies() {
		postReq.AddCookie(c)
	}
	postRR := httptest.NewRecorder()
	handler.ServeHTTP(postRR, postReq)

	if postRR.Code != http.StatusForbidden {
		t.Errorf("POST without token: got %d, want 403", postRR.Code)
	}
}

func TestCSRFAcceptsValidToken(t *testing.T) {
	csrf := NewCSRF(false)
	handler := csrf(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// GET to get a token.
	getReq := httptest.NewRequest(http.MethodGet, "/admin/login", nil)
	getRR := httptest.NewRecorder()
	handler.ServeHTTP(getRR, getReq)

	var token string
	for _, c := range getRR.Result().Cookies() {
		if c.Name == CSRFCookieName {
			token = c.Value
		}
	}

	// POST with valid token in header should succeed.
	postReq := httptest.NewRequest(http.MethodPost, "/admin/login", nil)
	for _, c := range getRR.Result().Cookies() {
		postReq.AddCookie(c)
	}
	postReq.Header.Set(CSRFHeaderName, token)
	postRR := httptest.NewRecorder()
	handler.ServeHTTP(postRR, postReq)

	if postRR.Code != http.StatusOK {
		t.Errorf("POST with valid token: got %d, want 200", postRR.Code)
	}
}
