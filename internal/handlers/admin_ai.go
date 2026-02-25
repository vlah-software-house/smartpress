package handlers

import (
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"strings"
)

// --- AI Assistant Endpoints ---
//
// These handlers power the content editor's AI assistant panel.
// Each endpoint accepts form values (title, body, tone) from HTMX requests,
// calls the active AI provider, and returns HTML fragments that get swapped
// into the assistant panel's result areas.

// AISuggestTitle generates title suggestions based on the content body.
// Returns an HTML fragment with clickable title options.
func (a *Admin) AISuggestTitle(w http.ResponseWriter, r *http.Request) {
	body := r.FormValue("body")
	title := r.FormValue("title")

	if body == "" && title == "" {
		writeAIError(w, "Please write some content or a working title first.")
		return
	}

	prompt := fmt.Sprintf("Title: %s\n\nContent:\n%s", title, truncate(body, 2000))

	systemPrompt := `You are a headline writing expert for a CMS. Generate exactly 5 compelling,
SEO-friendly title suggestions for the given content. Each title should be on its own line,
numbered 1-5. Keep titles under 70 characters. Do not include any other text or explanation.`

	result, err := a.aiRegistry.Generate(r.Context(), systemPrompt, prompt)
	if err != nil {
		slog.Error("ai suggest title failed", "error", err)
		writeAIError(w, "AI request failed. Check your provider configuration.")
		return
	}

	// Parse the numbered titles and render as clickable items.
	titles := parseNumberedList(result)
	if len(titles) == 0 {
		writeAIResult(w, result)
		return
	}

	var sb strings.Builder
	sb.WriteString(`<div class="space-y-1.5">`)
	for _, t := range titles {
		escaped := html.EscapeString(t)
		sb.WriteString(fmt.Sprintf(
			`<button type="button" onclick="document.getElementById('title').value = this.textContent.trim()"
				class="block w-full text-left text-xs px-2 py-1.5 rounded bg-indigo-50 text-indigo-800 hover:bg-indigo-100 transition-colors truncate"
				title="%s">%s</button>`,
			escaped, escaped,
		))
	}
	sb.WriteString(`</div>`)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(sb.String()))
}

// AIGenerateExcerpt creates a concise excerpt from the content body.
// Returns an HTML fragment with the excerpt and an "Apply" button.
func (a *Admin) AIGenerateExcerpt(w http.ResponseWriter, r *http.Request) {
	body := r.FormValue("body")
	title := r.FormValue("title")

	if body == "" {
		writeAIError(w, "Please write some content first so AI can generate an excerpt.")
		return
	}

	prompt := fmt.Sprintf("Title: %s\n\nContent:\n%s", title, truncate(body, 2000))

	systemPrompt := `You are a content summarization expert. Generate a compelling excerpt/summary
of the given content in 1-2 sentences (max 160 characters). The excerpt should capture the essence
of the content and entice readers to click. Output ONLY the excerpt text, nothing else.`

	result, err := a.aiRegistry.Generate(r.Context(), systemPrompt, prompt)
	if err != nil {
		slog.Error("ai generate excerpt failed", "error", err)
		writeAIError(w, "AI request failed. Check your provider configuration.")
		return
	}

	result = strings.TrimSpace(result)
	escaped := html.EscapeString(result)

	fragment := fmt.Sprintf(
		`<div class="space-y-2">
			<p class="text-xs text-gray-700 bg-gray-50 rounded p-2">%s</p>
			<button type="button" onclick="document.getElementById('excerpt').value = %s"
				class="w-full rounded-md bg-indigo-50 border border-indigo-200 px-2 py-1 text-xs font-medium text-indigo-700 hover:bg-indigo-100 transition-colors">
				Apply to Excerpt
			</button>
		</div>`,
		escaped,
		quoteJSString(result),
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fragment))
}

// AISEOMetadata generates SEO meta description and keywords from the content.
// Returns an HTML fragment with both fields and "Apply" buttons.
func (a *Admin) AISEOMetadata(w http.ResponseWriter, r *http.Request) {
	body := r.FormValue("body")
	title := r.FormValue("title")

	if body == "" && title == "" {
		writeAIError(w, "Please write some content or a title first.")
		return
	}

	prompt := fmt.Sprintf("Title: %s\n\nContent:\n%s", title, truncate(body, 2000))

	systemPrompt := `You are an SEO expert. For the given content, generate:
1. A meta description (max 160 characters, compelling for search results)
2. A comma-separated list of 5-8 relevant keywords

Output your response in EXACTLY this format (two lines only):
DESCRIPTION: <your meta description>
KEYWORDS: <keyword1, keyword2, keyword3, ...>

Do not include any other text.`

	result, err := a.aiRegistry.Generate(r.Context(), systemPrompt, prompt)
	if err != nil {
		slog.Error("ai seo metadata failed", "error", err)
		writeAIError(w, "AI request failed. Check your provider configuration.")
		return
	}

	desc, keywords := parseSEOResult(result)

	var sb strings.Builder
	sb.WriteString(`<div class="space-y-3">`)

	if desc != "" {
		escapedDesc := html.EscapeString(desc)
		sb.WriteString(fmt.Sprintf(
			`<div>
				<p class="text-xs font-medium text-gray-600 mb-1">Meta Description:</p>
				<p class="text-xs text-gray-700 bg-gray-50 rounded p-2">%s</p>
				<button type="button" onclick="document.getElementById('meta_description').value = %s"
					class="mt-1 w-full rounded-md bg-indigo-50 border border-indigo-200 px-2 py-1 text-xs font-medium text-indigo-700 hover:bg-indigo-100 transition-colors">
					Apply Description
				</button>
			</div>`,
			escapedDesc,
			quoteJSString(desc),
		))
	}

	if keywords != "" {
		escapedKw := html.EscapeString(keywords)
		sb.WriteString(fmt.Sprintf(
			`<div>
				<p class="text-xs font-medium text-gray-600 mb-1">Keywords:</p>
				<p class="text-xs text-gray-700 bg-gray-50 rounded p-2">%s</p>
				<button type="button" onclick="document.getElementById('meta_keywords').value = %s"
					class="mt-1 w-full rounded-md bg-indigo-50 border border-indigo-200 px-2 py-1 text-xs font-medium text-indigo-700 hover:bg-indigo-100 transition-colors">
					Apply Keywords
				</button>
			</div>`,
			escapedKw,
			quoteJSString(keywords),
		))
	}

	// Fallback if parsing failed: show raw result.
	if desc == "" && keywords == "" {
		writeAIResult(w, result)
		return
	}

	sb.WriteString(`</div>`)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(sb.String()))
}

// AIRewrite rewrites the content body in a specified tone.
// Returns an HTML fragment with the rewritten content and an "Apply" button.
func (a *Admin) AIRewrite(w http.ResponseWriter, r *http.Request) {
	body := r.FormValue("body")
	title := r.FormValue("title")
	tone := r.FormValue("tone")

	if body == "" {
		writeAIError(w, "Please write some content first so AI can rewrite it.")
		return
	}

	if tone == "" {
		tone = "professional"
	}

	toneDescriptions := map[string]string{
		"professional": "professional, clear, and authoritative",
		"casual":       "casual, friendly, and conversational",
		"formal":       "formal, academic, and precise",
		"persuasive":   "persuasive, compelling, and action-oriented",
		"concise":      "concise, direct, and to-the-point",
	}

	toneDesc, ok := toneDescriptions[tone]
	if !ok {
		toneDesc = "professional, clear, and authoritative"
	}

	prompt := fmt.Sprintf("Title: %s\n\nContent to rewrite:\n%s", title, truncate(body, 3000))

	systemPrompt := fmt.Sprintf(`You are a professional content editor. Rewrite the given content
in a %s tone. Preserve the key information and structure but adjust the language and style.
If the content contains HTML tags, preserve them. Output ONLY the rewritten content, nothing else.`, toneDesc)

	result, err := a.aiRegistry.Generate(r.Context(), systemPrompt, prompt)
	if err != nil {
		slog.Error("ai rewrite failed", "error", err)
		writeAIError(w, "AI request failed. Check your provider configuration.")
		return
	}

	result = strings.TrimSpace(result)

	fragment := fmt.Sprintf(
		`<div class="space-y-2">
			<div class="text-xs text-gray-700 bg-gray-50 rounded p-2 max-h-48 overflow-y-auto whitespace-pre-wrap">%s</div>
			<button type="button" onclick="document.getElementById('body').value = %s"
				class="w-full rounded-md bg-indigo-50 border border-indigo-200 px-2 py-1 text-xs font-medium text-indigo-700 hover:bg-indigo-100 transition-colors">
				Apply to Content
			</button>
		</div>`,
		html.EscapeString(result),
		quoteJSString(result),
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fragment))
}

// AIExtractTags extracts relevant tags/categories from the content.
// Returns an HTML fragment with clickable tag pills.
func (a *Admin) AIExtractTags(w http.ResponseWriter, r *http.Request) {
	body := r.FormValue("body")
	title := r.FormValue("title")

	if body == "" && title == "" {
		writeAIError(w, "Please write some content or a title first.")
		return
	}

	prompt := fmt.Sprintf("Title: %s\n\nContent:\n%s", title, truncate(body, 2000))

	systemPrompt := `You are a content categorization expert. Extract 5-10 relevant tags from
the given content. Tags should be short (1-3 words), lowercase, and relevant for blog categorization.
Output ONLY the tags as a comma-separated list on a single line. No other text.`

	result, err := a.aiRegistry.Generate(r.Context(), systemPrompt, prompt)
	if err != nil {
		slog.Error("ai extract tags failed", "error", err)
		writeAIError(w, "AI request failed. Check your provider configuration.")
		return
	}

	tags := parseTags(result)
	if len(tags) == 0 {
		writeAIResult(w, result)
		return
	}

	var sb strings.Builder
	sb.WriteString(`<div class="flex flex-wrap gap-1.5">`)
	for _, tag := range tags {
		escaped := html.EscapeString(tag)
		// Clicking a tag appends it to the meta_keywords field.
		sb.WriteString(fmt.Sprintf(
			`<button type="button"
				onclick="var f=document.getElementById('meta_keywords'); f.value = f.value ? f.value + ', %s' : '%s'"
				class="inline-flex items-center rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-700 hover:bg-indigo-100 hover:text-indigo-700 transition-colors cursor-pointer">
				%s
			</button>`,
			escaped, escaped, escaped,
		))
	}
	sb.WriteString(`</div>`)
	sb.WriteString(`<p class="text-xs text-gray-400 mt-1.5">Click tags to add them to keywords.</p>`)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(sb.String()))
}

// --- Helper functions ---

// writeAIError writes an error message HTML fragment.
func writeAIError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<p class="text-xs text-red-600 bg-red-50 rounded p-2">%s</p>`, html.EscapeString(msg))
}

// writeAIResult writes a plain text result as an HTML fragment.
func writeAIResult(w http.ResponseWriter, result string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<p class="text-xs text-gray-700 bg-gray-50 rounded p-2 whitespace-pre-wrap">%s</p>`,
		html.EscapeString(strings.TrimSpace(result)))
}

// truncate cuts a string to maxLen characters, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// parseNumberedList extracts items from a numbered list (e.g., "1. Title Here").
func parseNumberedList(text string) []string {
	var items []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Strip common numbered list prefixes: "1. ", "1) ", "- "
		for _, prefix := range []string{"1. ", "2. ", "3. ", "4. ", "5. ", "6. ", "7. ", "8. ", "9. ", "10. ",
			"1) ", "2) ", "3) ", "4) ", "5) ", "6) ", "7) ", "8) ", "9) ", "10) ",
			"- ", "* "} {
			if strings.HasPrefix(line, prefix) {
				line = strings.TrimPrefix(line, prefix)
				break
			}
		}
		line = strings.TrimSpace(line)
		// Remove surrounding quotes if present.
		line = strings.Trim(line, `"'`)
		if line != "" {
			items = append(items, line)
		}
	}
	return items
}

// parseSEOResult extracts meta description and keywords from the structured
// AI response (expects "DESCRIPTION: ..." and "KEYWORDS: ..." lines).
func parseSEOResult(text string) (description, keywords string) {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		upper := strings.ToUpper(line)
		if strings.HasPrefix(upper, "DESCRIPTION:") {
			description = strings.TrimSpace(line[len("DESCRIPTION:"):])
		} else if strings.HasPrefix(upper, "KEYWORDS:") {
			keywords = strings.TrimSpace(line[len("KEYWORDS:"):])
		} else if strings.HasPrefix(upper, "META DESCRIPTION:") {
			description = strings.TrimSpace(line[len("META DESCRIPTION:"):])
		} else if strings.HasPrefix(upper, "META KEYWORDS:") {
			keywords = strings.TrimSpace(line[len("META KEYWORDS:"):])
		}
	}
	return description, keywords
}

// parseTags splits a comma-separated tag string into individual trimmed tags.
func parseTags(text string) []string {
	var tags []string
	for _, tag := range strings.Split(text, ",") {
		tag = strings.TrimSpace(tag)
		// Remove surrounding quotes, dashes, bullets.
		tag = strings.Trim(tag, `"'-*`)
		tag = strings.TrimSpace(tag)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

// quoteJSString produces a JavaScript string literal safe for embedding in
// HTML attributes (e.g., onclick). Escapes backslashes, quotes, newlines,
// and HTML-significant characters using JS hex escapes to prevent XSS.
func quoteJSString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	s = strings.ReplaceAll(s, `"`, `\x22`)
	s = strings.ReplaceAll(s, "<", `\x3c`)
	s = strings.ReplaceAll(s, ">", `\x3e`)
	s = strings.ReplaceAll(s, "&", `\x26`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", ``)
	return `'` + s + `'`
}
