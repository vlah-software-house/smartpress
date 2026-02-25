package handlers

import (
	"testing"
)

func TestParseNumberedList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "dot numbered",
			input:    "1. First Title\n2. Second Title\n3. Third Title",
			expected: []string{"First Title", "Second Title", "Third Title"},
		},
		{
			name:     "paren numbered",
			input:    "1) First Title\n2) Second Title",
			expected: []string{"First Title", "Second Title"},
		},
		{
			name:     "dash bullets",
			input:    "- First Title\n- Second Title",
			expected: []string{"First Title", "Second Title"},
		},
		{
			name:     "with quotes",
			input:    `1. "First Title"` + "\n" + `2. 'Second Title'`,
			expected: []string{"First Title", "Second Title"},
		},
		{
			name:     "with empty lines",
			input:    "\n1. First\n\n2. Second\n\n",
			expected: []string{"First", "Second"},
		},
		{
			name:     "no prefix",
			input:    "Just a single line",
			expected: []string{"Just a single line"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseNumberedList(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("got %d items, want %d: %v", len(result), len(tt.expected), result)
			}
			for i, item := range result {
				if item != tt.expected[i] {
					t.Errorf("item %d: got %q, want %q", i, item, tt.expected[i])
				}
			}
		})
	}
}

func TestParseSEOResult(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantDesc     string
		wantKeywords string
	}{
		{
			name:         "standard format",
			input:        "DESCRIPTION: A great article about Go.\nKEYWORDS: go, programming, google",
			wantDesc:     "A great article about Go.",
			wantKeywords: "go, programming, google",
		},
		{
			name:         "lowercase prefixes",
			input:        "description: Some description.\nkeywords: key1, key2",
			wantDesc:     "Some description.",
			wantKeywords: "key1, key2",
		},
		{
			name:         "meta prefixes",
			input:        "Meta Description: SEO optimized.\nMeta Keywords: seo, web",
			wantDesc:     "SEO optimized.",
			wantKeywords: "seo, web",
		},
		{
			name:         "extra whitespace",
			input:        "\n\nDESCRIPTION:   Spaced out.  \nKEYWORDS:   a, b, c  \n",
			wantDesc:     "Spaced out.",
			wantKeywords: "a, b, c",
		},
		{
			name:         "no match",
			input:        "Some random text without structure",
			wantDesc:     "",
			wantKeywords: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc, kw := parseSEOResult(tt.input)
			if desc != tt.wantDesc {
				t.Errorf("description: got %q, want %q", desc, tt.wantDesc)
			}
			if kw != tt.wantKeywords {
				t.Errorf("keywords: got %q, want %q", kw, tt.wantKeywords)
			}
		})
	}
}

func TestParseTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "comma separated",
			input:    "go, programming, web development, api",
			expected: []string{"go", "programming", "web development", "api"},
		},
		{
			name:     "with quotes",
			input:    `"go", "programming", "web"`,
			expected: []string{"go", "programming", "web"},
		},
		{
			name:     "with dashes and bullets",
			input:    "- go, - programming, * web",
			expected: []string{"go", "programming", "web"},
		},
		{
			name:     "extra whitespace",
			input:    "  go ,  programming  , web  ",
			expected: []string{"go", "programming", "web"},
		},
		{
			name:     "empty items filtered",
			input:    "go,,, programming",
			expected: []string{"go", "programming"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTags(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("got %d tags, want %d: %v", len(result), len(tt.expected), result)
			}
			for i, tag := range result {
				if tag != tt.expected[i] {
					t.Errorf("tag %d: got %q, want %q", i, tag, tt.expected[i])
				}
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("short", 100); got != "short" {
		t.Errorf("short string: got %q", got)
	}
	if got := truncate("hello world", 5); got != "hello..." {
		t.Errorf("truncated: got %q, want %q", got, "hello...")
	}
}

func TestQuoteJSString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "hello", "'hello'"},
		{"single quote", "it's", `'it\'s'`},
		{"double quote", `he said "hi"`, `'he said \x22hi\x22'`},
		{"html tags", "<script>alert(1)</script>", `'\x3cscript\x3ealert(1)\x3c/script\x3e'`},
		{"newline", "line1\nline2", `'line1\nline2'`},
		{"backslash", `back\slash`, `'back\\slash'`},
		{"ampersand", "a&b", `'a\x26b'`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := quoteJSString(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
