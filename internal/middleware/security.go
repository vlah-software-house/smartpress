// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

package middleware

import "net/http"

// SecureHeaders adds security-related HTTP headers to every response.
// These headers protect against common web vulnerabilities like clickjacking,
// MIME-sniffing, and information leakage.
func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()

		// Prevent the browser from MIME-sniffing the Content-Type.
		h.Set("X-Content-Type-Options", "nosniff")

		// Prevent embedding in iframes from other origins (clickjacking).
		h.Set("X-Frame-Options", "SAMEORIGIN")

		// Disable the legacy XSS filter (can cause issues; CSP is preferred).
		h.Set("X-XSS-Protection", "0")

		// Control what information is sent in the Referer header.
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Prevent the site from being used in FLoC cohort calculations.
		h.Set("Permissions-Policy", "interest-cohort=()")

		next.ServeHTTP(w, r)
	})
}
