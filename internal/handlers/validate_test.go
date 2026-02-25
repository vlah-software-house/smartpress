package handlers

import (
	"strings"
	"testing"
)

func TestValidateContent(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		slug      string
		body      string
		wantError bool
	}{
		{"valid", "My Title", "my-title", "Body text", false},
		{"empty title", "", "slug", "body", true},
		{"whitespace title", "   ", "slug", "body", true},
		{"title too long", strings.Repeat("a", 301), "slug", "body", true},
		{"slug too long", "title", strings.Repeat("a", 301), "body", true},
		{"body too long", "title", "slug", strings.Repeat("a", 100_001), true},
		{"empty body allowed", "title", "slug", "", false},
		{"empty slug allowed", "title", "", "body", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateContent(tt.title, tt.slug, tt.body)
			if tt.wantError && result == "" {
				t.Error("expected an error, got none")
			}
			if !tt.wantError && result != "" {
				t.Errorf("unexpected error: %s", result)
			}
		})
	}
}

func TestValidateMetadata(t *testing.T) {
	tests := []struct {
		name      string
		excerpt   string
		metaDesc  string
		metaKw    string
		wantError bool
	}{
		{"all empty", "", "", "", false},
		{"all valid", "excerpt", "description", "kw1, kw2", false},
		{"excerpt too long", strings.Repeat("a", 1001), "", "", true},
		{"meta desc too long", "", strings.Repeat("a", 501), "", true},
		{"meta kw too long", "", "", strings.Repeat("a", 501), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateMetadata(tt.excerpt, tt.metaDesc, tt.metaKw)
			if tt.wantError && result == "" {
				t.Error("expected an error, got none")
			}
			if !tt.wantError && result != "" {
				t.Errorf("unexpected error: %s", result)
			}
		})
	}
}

func TestValidateTemplate(t *testing.T) {
	tests := []struct {
		name        string
		tmplName    string
		htmlContent string
		wantError   bool
	}{
		{"valid", "My Template", "<div>Hello</div>", false},
		{"empty name", "", "<div>Hello</div>", true},
		{"whitespace name", "   ", "<div>Hello</div>", true},
		{"name too long", strings.Repeat("a", 201), "<div>Hello</div>", true},
		{"empty html", "Name", "", true},
		{"whitespace html", "Name", "   ", true},
		{"html too long", "Name", strings.Repeat("a", 500_001), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateTemplate(tt.tmplName, tt.htmlContent)
			if tt.wantError && result == "" {
				t.Error("expected an error, got none")
			}
			if !tt.wantError && result != "" {
				t.Errorf("unexpected error: %s", result)
			}
		})
	}
}
