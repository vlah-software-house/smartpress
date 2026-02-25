// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

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

// TestCSRFTokenFromCtx verifies that the CSRF token is available in the
// request context after the middleware runs.
func TestCSRFTokenFromCtx(t *testing.T) {
	t.Run("token is set in context on GET", func(t *testing.T) {
		var ctxToken string
		csrf := NewCSRF(false)
		handler := csrf(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxToken = CSRFTokenFromCtx(r.Context())
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/admin/login", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if ctxToken == "" {
			t.Error("CSRFTokenFromCtx returned empty string, expected a token")
		}

		// Verify the context token matches the cookie token.
		var cookieToken string
		for _, c := range rr.Result().Cookies() {
			if c.Name == CSRFCookieName {
				cookieToken = c.Value
			}
		}
		if ctxToken != cookieToken {
			t.Errorf("context token %q != cookie token %q", ctxToken, cookieToken)
		}
	})

	t.Run("returns empty string when not in context", func(t *testing.T) {
		token := CSRFTokenFromCtx(httptest.NewRequest(http.MethodGet, "/", nil).Context())
		if token != "" {
			t.Errorf("expected empty string, got %q", token)
		}
	})

	t.Run("token reuses existing cookie", func(t *testing.T) {
		var ctxToken string
		csrf := NewCSRF(false)
		handler := csrf(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxToken = CSRFTokenFromCtx(r.Context())
			w.WriteHeader(http.StatusOK)
		}))

		// First request generates a token.
		getReq := httptest.NewRequest(http.MethodGet, "/admin/login", nil)
		getRR := httptest.NewRecorder()
		handler.ServeHTTP(getRR, getReq)

		var originalToken string
		for _, c := range getRR.Result().Cookies() {
			if c.Name == CSRFCookieName {
				originalToken = c.Value
			}
		}

		// Second request with existing cookie should reuse the token.
		getReq2 := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
		getReq2.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: originalToken})
		getRR2 := httptest.NewRecorder()
		handler.ServeHTTP(getRR2, getReq2)

		if ctxToken != originalToken {
			t.Errorf("context token %q != original token %q", ctxToken, originalToken)
		}
	})
}

// TestCSRFAcceptsFormFieldToken verifies that the CSRF middleware accepts
// the token submitted via a form field (not just the header).
func TestCSRFAcceptsFormFieldToken(t *testing.T) {
	csrf := NewCSRF(false)
	handler := csrf(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// GET to get a token.
	getReq := httptest.NewRequest(http.MethodGet, "/admin/content", nil)
	getRR := httptest.NewRecorder()
	handler.ServeHTTP(getRR, getReq)

	var token string
	for _, c := range getRR.Result().Cookies() {
		if c.Name == CSRFCookieName {
			token = c.Value
		}
	}

	// POST with valid token in form field should succeed.
	postReq := httptest.NewRequest(http.MethodPost, "/admin/content?"+CSRFFormField+"="+token, nil)
	for _, c := range getRR.Result().Cookies() {
		postReq.AddCookie(c)
	}
	postRR := httptest.NewRecorder()
	handler.ServeHTTP(postRR, postReq)

	if postRR.Code != http.StatusOK {
		t.Errorf("POST with form field token: got %d, want 200", postRR.Code)
	}
}

// TestCSRFSafeMethodsPassThrough verifies HEAD and OPTIONS also pass
// through without CSRF validation.
func TestCSRFSafeMethodsPassThrough(t *testing.T) {
	methods := []string{http.MethodGet, http.MethodHead, http.MethodOptions}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			var called bool
			csrf := NewCSRF(false)
			handler := csrf(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(method, "/admin/dashboard", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if !called {
				t.Error("handler should be called for safe method")
			}
			if rr.Code != http.StatusOK {
				t.Errorf("status: got %d, want 200", rr.Code)
			}
		})
	}
}

// TestCSRFUnsafeMethodsWithoutCookie verifies that PUT, PATCH, DELETE
// also require CSRF validation.
func TestCSRFUnsafeMethodsRequireToken(t *testing.T) {
	methods := []string{http.MethodPut, http.MethodPatch, http.MethodDelete}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			csrf := NewCSRF(false)
			handler := csrf(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			// GET to get a token/cookie.
			getReq := httptest.NewRequest(http.MethodGet, "/admin/content", nil)
			getRR := httptest.NewRecorder()
			handler.ServeHTTP(getRR, getReq)

			// Unsafe method without token should be rejected.
			req := httptest.NewRequest(method, "/admin/content/1", nil)
			for _, c := range getRR.Result().Cookies() {
				req.AddCookie(c)
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusForbidden {
				t.Errorf("%s without token: got %d, want 403", method, rr.Code)
			}
		})
	}
}
