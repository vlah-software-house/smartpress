package models

import "testing"

// TestMediaIsImage verifies that IsImage correctly identifies image content
// types by checking for the "image/" prefix.
func TestMediaIsImage(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		// Image types
		{name: "jpeg", contentType: "image/jpeg", want: true},
		{name: "png", contentType: "image/png", want: true},
		{name: "gif", contentType: "image/gif", want: true},
		{name: "webp", contentType: "image/webp", want: true},
		{name: "svg+xml", contentType: "image/svg+xml", want: true},
		{name: "bmp", contentType: "image/bmp", want: true},
		{name: "tiff", contentType: "image/tiff", want: true},
		{name: "avif", contentType: "image/avif", want: true},

		// Non-image types
		{name: "pdf", contentType: "application/pdf", want: false},
		{name: "json", contentType: "application/json", want: false},
		{name: "html", contentType: "text/html", want: false},
		{name: "plain text", contentType: "text/plain", want: false},
		{name: "mp4 video", contentType: "video/mp4", want: false},
		{name: "mp3 audio", contentType: "audio/mpeg", want: false},
		{name: "zip archive", contentType: "application/zip", want: false},
		{name: "octet-stream", contentType: "application/octet-stream", want: false},

		// Edge cases
		{name: "empty content type", contentType: "", want: false},
		{name: "only image prefix no slash", contentType: "image", want: false},
		{name: "IMAGE uppercase", contentType: "IMAGE/PNG", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Media{ContentType: tt.contentType}
			got := m.IsImage()
			if got != tt.want {
				t.Errorf("Media{ContentType: %q}.IsImage() = %v, want %v",
					tt.contentType, got, tt.want)
			}
		})
	}
}

// TestMediaHumanSize verifies the human-readable file size formatting
// across byte, kilobyte, and megabyte ranges.
func TestMediaHumanSize(t *testing.T) {
	tests := []struct {
		name      string
		sizeBytes int64
		want      string
	}{
		// Byte range (< 1024)
		{name: "zero bytes", sizeBytes: 0, want: "0 B"},
		{name: "one byte", sizeBytes: 1, want: "1 B"},
		{name: "500 bytes", sizeBytes: 500, want: "500 B"},
		{name: "1023 bytes", sizeBytes: 1023, want: "1023 B"},

		// Kilobyte range (1024 <= x < 1048576)
		{name: "exactly 1 KB", sizeBytes: 1024, want: "1 KB"},
		{name: "1.5 KB", sizeBytes: 1536, want: "2 KB"},
		{name: "10 KB", sizeBytes: 10240, want: "10 KB"},
		{name: "100 KB", sizeBytes: 102400, want: "100 KB"},
		{name: "512 KB", sizeBytes: 524288, want: "512 KB"},
		{name: "just under 1 MB", sizeBytes: 1048575, want: "1024 KB"},

		// Megabyte range (>= 1048576)
		{name: "exactly 1 MB", sizeBytes: 1048576, want: "1.0 MB"},
		{name: "1.5 MB", sizeBytes: 1572864, want: "1.5 MB"},
		{name: "10 MB", sizeBytes: 10485760, want: "10.0 MB"},
		{name: "100 MB", sizeBytes: 104857600, want: "100.0 MB"},
		{name: "2.3 MB", sizeBytes: 2411724, want: "2.3 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Media{SizeBytes: tt.sizeBytes}
			got := m.HumanSize()
			if got != tt.want {
				t.Errorf("Media{SizeBytes: %d}.HumanSize() = %q, want %q",
					tt.sizeBytes, got, tt.want)
			}
		})
	}
}
