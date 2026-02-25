package slug

import "testing"

// TestGenerate exercises the slug generator with a broad range of inputs
// covering typical titles, special characters, unicode, edge cases, and
// boundary conditions.
func TestGenerate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// --- Normal titles ---
		{
			name:  "simple two words",
			input: "Hello World",
			want:  "hello-world",
		},
		{
			name:  "title with year",
			input: "Hello World 2026",
			want:  "hello-world-2026",
		},
		{
			name:  "already lowercase",
			input: "already lowercase",
			want:  "already-lowercase",
		},
		{
			name:  "single word",
			input: "GoLang",
			want:  "golang",
		},
		{
			name:  "mixed case sentence",
			input: "The Quick Brown Fox Jumps Over the Lazy Dog",
			want:  "the-quick-brown-fox-jumps-over-the-lazy-dog",
		},

		// --- Special characters ---
		{
			name:  "punctuation marks",
			input: "Hello, World! How's it going?",
			want:  "hello-world-hows-it-going",
		},
		{
			name:  "ampersand and at sign",
			input: "Rock & Roll @ the Arena",
			want:  "rock-roll-the-arena",
		},
		{
			name:  "parentheses and brackets",
			input: "Version (2.0) [Beta]",
			want:  "version-20-beta",
		},
		{
			name:  "slashes and pipes",
			input: "Frontend/Backend | Full Stack",
			want:  "frontendbackend-full-stack",
		},
		{
			name:  "hash and dollar",
			input: "Issue #42 costs $100",
			want:  "issue-42-costs-100",
		},
		{
			name:  "plus and equals",
			input: "1 + 1 = 2",
			want:  "1-1-2",
		},

		// --- Unicode and accented characters ---
		{
			name:  "accented latin characters",
			input: "Cafe Resume Noel",
			want:  "cafe-resume-noel",
		},
		{
			name:  "french accents stripped",
			input: "Les Miserables a la carte",
			want:  "les-miserables-a-la-carte",
		},
		{
			name:  "german umlauts stripped",
			input: "Uber die Brucke",
			want:  "uber-die-brucke",
		},
		{
			name:  "emoji stripped",
			input: "Hello World",
			want:  "hello-world",
		},
		{
			name:  "chinese characters stripped",
			input: "Hello World",
			want:  "hello-world",
		},
		{
			name:  "only unicode chars",
			input: "Cliches",
			want:  "cliches",
		},

		// --- Whitespace handling ---
		{
			name:  "leading spaces",
			input: "   hello world",
			want:  "hello-world",
		},
		{
			name:  "trailing spaces",
			input: "hello world   ",
			want:  "hello-world",
		},
		{
			name:  "leading and trailing spaces",
			input: "  hello world  ",
			want:  "hello-world",
		},
		{
			name:  "multiple consecutive spaces collapsed",
			input: "hello    world",
			want:  "hello-world",
		},
		{
			name:  "tabs preserved as whitespace",
			input: "hello\tworld",
			want:  "hello\tworld",
		},
		{
			name:  "newlines preserved as whitespace",
			input: "hello\nworld",
			want:  "hello\nworld",
		},

		// --- Hyphen handling ---
		{
			name:  "leading hyphens",
			input: "---hello world",
			want:  "hello-world",
		},
		{
			name:  "trailing hyphens",
			input: "hello world---",
			want:  "hello-world",
		},
		{
			name:  "multiple hyphens between words",
			input: "hello---world",
			want:  "hello-world",
		},
		{
			name:  "single hyphen preserved",
			input: "well-known fact",
			want:  "well-known-fact",
		},
		{
			name:  "hyphens and spaces mixed",
			input: "  --hello -- world--  ",
			want:  "hello-world",
		},

		// --- Edge cases ---
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only spaces",
			input: "     ",
			want:  "",
		},
		{
			name:  "only hyphens",
			input: "-----",
			want:  "",
		},
		{
			name:  "only special characters",
			input: "!@#$%^&*()",
			want:  "",
		},
		{
			name:  "single character",
			input: "A",
			want:  "a",
		},
		{
			name:  "single number",
			input: "5",
			want:  "5",
		},
		{
			name:  "single hyphen",
			input: "-",
			want:  "",
		},
		{
			name:  "single space",
			input: " ",
			want:  "",
		},

		// --- Numbers ---
		{
			name:  "all numbers",
			input: "123456",
			want:  "123456",
		},
		{
			name:  "numbers with spaces",
			input: "12 34 56",
			want:  "12-34-56",
		},
		{
			name:  "version number",
			input: "Version 2.0.1",
			want:  "version-201",
		},
		{
			name:  "date-like string",
			input: "2026-02-25",
			want:  "2026-02-25",
		},
		{
			name:  "mixed words and numbers",
			input: "Chapter 3 Section 14",
			want:  "chapter-3-section-14",
		},

		// --- Long strings ---
		{
			name:  "very long title",
			input: "This is a very long title that goes on and on and on and on and might be used as a blog post title by someone who really likes long titles and does not care about brevity at all",
			want:  "this-is-a-very-long-title-that-goes-on-and-on-and-on-and-on-and-might-be-used-as-a-blog-post-title-by-someone-who-really-likes-long-titles-and-does-not-care-about-brevity-at-all",
		},

		// --- Realistic blog titles ---
		{
			name:  "tech blog title",
			input: "How to Deploy Go Apps on Kubernetes (2026 Edition)",
			want:  "how-to-deploy-go-apps-on-kubernetes-2026-edition",
		},
		{
			name:  "question title",
			input: "What is HTMX? A Complete Guide",
			want:  "what-is-htmx-a-complete-guide",
		},
		{
			name:  "colon separated title",
			input: "Go: The Complete Developer Guide",
			want:  "go-the-complete-developer-guide",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Generate(tt.input)
			if got != tt.want {
				t.Errorf("Generate(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestGenerate_Idempotent verifies that generating a slug from an already
// valid slug produces the same result.
func TestGenerate_Idempotent(t *testing.T) {
	slugs := []string{
		"hello-world",
		"my-blog-post-2026",
		"a",
		"123",
	}

	for _, s := range slugs {
		t.Run(s, func(t *testing.T) {
			got := Generate(s)
			if got != s {
				t.Errorf("Generate(%q) = %q, want idempotent result %q", s, got, s)
			}
		})
	}
}

// TestGenerate_ConsistentCase verifies that slugs are always lowercase
// regardless of input casing.
func TestGenerate_ConsistentCase(t *testing.T) {
	inputs := []string{
		"HELLO WORLD",
		"Hello World",
		"hElLo WoRlD",
		"hello world",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			got := Generate(input)
			if got != "hello-world" {
				t.Errorf("Generate(%q) = %q, want %q", input, got, "hello-world")
			}
		})
	}
}
