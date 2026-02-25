// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

package engine

import (
	"testing"
)

// TestImgSrcRegex verifies the regex matches <img> tags correctly.
func TestImgSrcRegex(t *testing.T) {
	tests := []struct {
		name       string
		html       string
		wantCount  int
		wantSrcURL string // expected src from first match
	}{
		{
			name:       "simple img tag",
			html:       `<img src="https://example.com/image.png">`,
			wantCount:  1,
			wantSrcURL: "https://example.com/image.png",
		},
		{
			name:       "img with attributes before src",
			html:       `<img class="w-full" src="https://example.com/image.png" alt="test">`,
			wantCount:  1,
			wantSrcURL: "https://example.com/image.png",
		},
		{
			name:       "img with single quotes",
			html:       `<img src='https://example.com/image.png'>`,
			wantCount:  1,
			wantSrcURL: "https://example.com/image.png",
		},
		{
			name:       "self-closing img",
			html:       `<img src="https://example.com/image.png" />`,
			wantCount:  1,
			wantSrcURL: "https://example.com/image.png",
		},
		{
			name:       "multiple images",
			html:       `<p><img src="https://a.com/1.png"></p><p><img src="https://b.com/2.png"></p>`,
			wantCount:  2,
			wantSrcURL: "https://a.com/1.png",
		},
		{
			name:      "no images",
			html:      `<p>No images here</p>`,
			wantCount: 0,
		},
		{
			name:       "img with loading and class",
			html:       `<img src="https://s3.example.com/bucket/media/2026/02/abc.png" alt="Photo" class="w-full h-full object-cover" loading="lazy">`,
			wantCount:  1,
			wantSrcURL: "https://s3.example.com/bucket/media/2026/02/abc.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := imgSrcRe.FindAllStringSubmatch(tt.html, -1)
			if len(matches) != tt.wantCount {
				t.Errorf("got %d matches, want %d", len(matches), tt.wantCount)
				return
			}
			if tt.wantCount > 0 && matches[0][2] != tt.wantSrcURL {
				t.Errorf("src URL = %q, want %q", matches[0][2], tt.wantSrcURL)
			}
		})
	}
}

// TestRewriteBodyImages_NoDeps verifies the rewriter is a no-op
// when media dependencies are not configured.
func TestRewriteBodyImages_NoDeps(t *testing.T) {
	eng := &Engine{cache: newTemplateCache()}

	input := `<p><img src="https://s3.example.com/bucket/media/img.png" alt="test"></p>`
	got := eng.rewriteBodyImages(input)
	if got != input {
		t.Errorf("expected passthrough when deps are nil, got:\n%s", got)
	}
}

// TestRewriteBodyImages_SkipsSrcset verifies tags that already have
// srcset are not modified.
func TestRewriteBodyImages_SkipsSrcset(t *testing.T) {
	eng := &Engine{cache: newTemplateCache()}

	input := `<img src="img.png" srcset="img_sm.webp 640w" alt="test">`
	got := eng.rewriteBodyImages(input)
	if got != input {
		t.Errorf("should skip tags with existing srcset, got:\n%s", got)
	}
}
