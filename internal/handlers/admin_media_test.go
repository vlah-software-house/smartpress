// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

package handlers

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"yaaicms/internal/models"
)

func TestGenerateThumbnail(t *testing.T) {
	t.Run("jpeg thumbnail", func(t *testing.T) {
		// Create a 800x600 test image.
		img := image.NewRGBA(image.Rect(0, 0, 800, 600))
		for y := 0; y < 600; y++ {
			for x := 0; x < 800; x++ {
				img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
			}
		}
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
			t.Fatal(err)
		}

		thumb, err := generateThumbnail(bytes.NewReader(buf.Bytes()), 400)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if thumb == nil {
			t.Fatal("expected thumbnail, got nil")
		}

		// Decode the thumbnail and verify dimensions.
		thumbImg, err := jpeg.Decode(bytes.NewReader(thumb))
		if err != nil {
			t.Fatalf("failed to decode thumbnail: %v", err)
		}
		bounds := thumbImg.Bounds()
		if bounds.Dx() != 400 {
			t.Errorf("width: got %d, want 400", bounds.Dx())
		}
		// Height should be proportional: 600 * (400/800) = 300
		if bounds.Dy() != 300 {
			t.Errorf("height: got %d, want 300", bounds.Dy())
		}
	})

	t.Run("png thumbnail", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 1200, 900))
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			t.Fatal(err)
		}

		thumb, err := generateThumbnail(bytes.NewReader(buf.Bytes()), 400)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if thumb == nil {
			t.Fatal("expected thumbnail, got nil")
		}
	})

	t.Run("skip small image", func(t *testing.T) {
		// Image smaller than maxWidth should return nil.
		img := image.NewRGBA(image.Rect(0, 0, 200, 150))
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, nil); err != nil {
			t.Fatal(err)
		}

		thumb, err := generateThumbnail(bytes.NewReader(buf.Bytes()), 400)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if thumb != nil {
			t.Error("expected nil for small image, got thumbnail data")
		}
	})
}

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
