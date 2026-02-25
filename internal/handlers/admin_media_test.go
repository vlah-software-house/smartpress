// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

package handlers

import (
	"testing"

	"yaaicms/internal/models"
)

func TestExtensionFromType(t *testing.T) {
	tests := []struct {
		contentType string
		want        string
	}{
		{"image/jpeg", ".jpg"},
		{"image/png", ".png"},
		{"image/gif", ".gif"},
		{"image/webp", ".webp"},
		{"image/svg+xml", ".svg"},
		{"application/pdf", ".pdf"},
		{"application/octet-stream", ""},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			got := extensionFromType(tt.contentType)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMediaModelMethods(t *testing.T) {
	t.Run("IsImage", func(t *testing.T) {
		m := &models.Media{ContentType: "image/jpeg"}
		if !m.IsImage() {
			t.Error("expected IsImage=true for image/jpeg")
		}
		m.ContentType = "application/pdf"
		if m.IsImage() {
			t.Error("expected IsImage=false for application/pdf")
		}
	})

	t.Run("HumanSize", func(t *testing.T) {
		tests := []struct {
			size int64
			want string
		}{
			{500, "500 B"},
			{1024, "1 KB"},
			{1536, "2 KB"},
			{1048576, "1.0 MB"},
			{5242880, "5.0 MB"},
		}
		for _, tt := range tests {
			m := &models.Media{SizeBytes: tt.size}
			got := m.HumanSize()
			if got != tt.want {
				t.Errorf("HumanSize(%d): got %q, want %q", tt.size, got, tt.want)
			}
		}
	})
}
