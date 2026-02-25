// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

// Package router tests verify the HTTP routing configuration, middleware
// chains, and the health endpoint.
package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/health", nil)

	healthHandler(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want 200", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("content-type: got %q, want %q", ct, "application/json")
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status field: got %q, want %q", body["status"], "ok")
	}
}

func TestHealthHandlerMethods(t *testing.T) {
	// Health endpoint only accepts GET.
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/health", nil)

	healthHandler(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("GET /health: got %d, want 200", w.Code)
	}
}
