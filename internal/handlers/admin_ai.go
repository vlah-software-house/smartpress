package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"strings"

	"smartpress/internal/engine"
	"smartpress/internal/models"
	"smartpress/internal/render"
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

// --- AI Template Builder Endpoints ---
//
// These handlers power the "AI Design" template builder — a chat-based UI
// where users describe a template and the AI generates HTML+TailwindCSS
// with Go template variables.

// AITemplatePage renders the AI template builder chat interface.
func (a *Admin) AITemplatePage(w http.ResponseWriter, r *http.Request) {
	a.renderer.Page(w, r, "template_ai", &render.PageData{
		Title:   "AI Template Builder",
		Section: "templates",
	})
}

// templateGenResponse is the JSON response from the template generation endpoint.
type templateGenResponse struct {
	HTML            string `json:"html"`
	Message         string `json:"message"`
	Valid           bool   `json:"valid"`
	ValidationError string `json:"validation_error,omitempty"`
	Preview         string `json:"preview,omitempty"`
	Error           string `json:"error,omitempty"`
}

// templateSaveResponse is the JSON response from the template save endpoint.
type templateSaveResponse struct {
	ID    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

// AITemplateGenerate generates an HTML+TailwindCSS template from a user prompt.
// It accepts the template type, prompt, optional conversation history, and
// optional current HTML for iterative refinement. Returns JSON with the
// generated HTML, validation status, and a rendered preview.
func (a *Admin) AITemplateGenerate(w http.ResponseWriter, r *http.Request) {
	prompt := r.FormValue("prompt")
	tmplType := r.FormValue("template_type")
	chatHistory := r.FormValue("chat_history")
	currentHTML := r.FormValue("current_html")

	if prompt == "" {
		writeJSON(w, http.StatusBadRequest, templateGenResponse{Error: "Please enter a prompt."})
		return
	}

	// Build the system prompt with type-specific template variable documentation.
	systemPrompt := buildTemplateSystemPrompt(tmplType)

	// Build the user prompt with context.
	var userPrompt strings.Builder
	userPrompt.WriteString(fmt.Sprintf("Template type: %s\n\n", tmplType))

	if currentHTML != "" {
		userPrompt.WriteString("Current template HTML (modify based on my new request):\n```html\n")
		userPrompt.WriteString(truncate(currentHTML, 4000))
		userPrompt.WriteString("\n```\n\n")
	}

	if chatHistory != "" && currentHTML == "" {
		userPrompt.WriteString("Conversation so far:\n")
		userPrompt.WriteString(truncate(chatHistory, 2000))
		userPrompt.WriteString("\n\n")
	}

	userPrompt.WriteString("Request: ")
	userPrompt.WriteString(prompt)

	result, err := a.aiRegistry.Generate(r.Context(), systemPrompt, userPrompt.String())
	if err != nil {
		slog.Error("ai template generate failed", "error", err)
		writeJSON(w, http.StatusOK, templateGenResponse{
			Error: "AI request failed. Check your provider configuration.",
		})
		return
	}

	// Extract HTML from the response (the AI may wrap it in markdown code blocks).
	htmlContent := extractHTMLFromResponse(result)

	// Validate as a Go template.
	validationErr := a.engine.ValidateTemplate(htmlContent)
	valid := validationErr == nil
	validationErrStr := ""
	if validationErr != nil {
		validationErrStr = validationErr.Error()
	}

	// Generate a preview with dummy data.
	var previewHTML string
	if valid {
		previewData := buildPreviewData(tmplType)
		rendered, err := a.engine.ValidateAndRender(htmlContent, previewData)
		if err == nil {
			previewHTML = string(rendered)
		}
	}

	// Build a summary message for the chat.
	message := "Template generated successfully."
	if !valid {
		message = "Template generated but has a syntax error. I'll try to fix it — describe the issue or try again."
	}

	writeJSON(w, http.StatusOK, templateGenResponse{
		HTML:            htmlContent,
		Message:         message,
		Valid:           valid,
		ValidationError: validationErrStr,
		Preview:         previewHTML,
	})
}

// AITemplateSave saves a generated template to the database.
// Validates the template before saving and triggers cache invalidation.
func (a *Admin) AITemplateSave(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	tmplType := models.TemplateType(r.FormValue("type"))
	htmlContent := r.FormValue("html_content")

	if name == "" || htmlContent == "" {
		writeJSON(w, http.StatusBadRequest, templateSaveResponse{Error: "Name and HTML content are required."})
		return
	}

	// Validate the template syntax before saving.
	if err := a.engine.ValidateTemplate(htmlContent); err != nil {
		writeJSON(w, http.StatusOK, templateSaveResponse{Error: "Template syntax error: " + err.Error()})
		return
	}

	t := &models.Template{
		Name:        name,
		Type:        tmplType,
		HTMLContent: htmlContent,
	}

	created, err := a.templateStore.Create(t)
	if err != nil {
		slog.Error("ai save template failed", "error", err)
		writeJSON(w, http.StatusOK, templateSaveResponse{Error: "Failed to save template."})
		return
	}

	a.cacheLog.Log("template", created.ID, "create")
	writeJSON(w, http.StatusOK, templateSaveResponse{ID: created.ID.String()})
}

// buildTemplateSystemPrompt creates a system prompt that instructs the LLM
// to generate an HTML+TailwindCSS template with the correct Go template
// variables for the given template type.
func buildTemplateSystemPrompt(tmplType string) string {
	base := `You are an expert web designer who creates beautiful, modern HTML templates using TailwindCSS.
You generate complete, production-ready HTML+TailwindCSS templates for a CMS called SmartPress.

CRITICAL RULES:
1. Output ONLY the HTML template code. No explanations, no markdown code fences, no comments outside the HTML.
2. Use TailwindCSS utility classes for all styling. Do not use custom CSS.
3. Use Go template syntax for dynamic content: {{.VariableName}}
4. For raw HTML content (like Body, Header, Footer), the CMS handles escaping — just use {{.Body}} etc.
5. Templates should be responsive and look professional.
6. Use semantic HTML elements (header, nav, main, article, footer, etc.).
7. Do not include <html>, <head>, or <body> tags — these templates are rendered as fragments.
8. Include the TailwindCSS CDN script tag only in page-level templates that need it.`

	var vars string
	switch tmplType {
	case "header":
		vars = `
Available variables for HEADER templates:
- {{.SiteName}} — The site name (e.g., "SmartPress")
- {{.Year}} — Current year (e.g., 2026)

Header templates typically contain: site logo/name, navigation links, maybe a search bar.
The header is injected into page templates via {{.Header}}.`

	case "footer":
		vars = `
Available variables for FOOTER templates:
- {{.SiteName}} — The site name
- {{.Year}} — Current year

Footer templates typically contain: copyright notice, social links, secondary navigation.
The footer is injected into page templates via {{.Footer}}.`

	case "page":
		vars = `
Available variables for PAGE templates:
- {{.Title}} — Page/post title
- {{.Body}} — Content body (raw HTML from editor)
- {{.Header}} — Pre-rendered header HTML (include with {{.Header}})
- {{.Footer}} — Pre-rendered footer HTML (include with {{.Footer}})
- {{.Excerpt}} — Short summary text
- {{.MetaDescription}} — SEO meta description
- {{.MetaKeywords}} — SEO keywords
- {{.SiteName}} — The site name
- {{.Year}} — Current year
- {{.Slug}} — URL slug
- {{.PublishedAt}} — Publication date string

Page templates are FULL page layouts. Include {{.Header}} at the top and {{.Footer}} at the bottom.
Include the TailwindCSS CDN: <script src="https://cdn.tailwindcss.com"></script>
Wrap the page in a proper HTML structure with <html>, <head>, <body> tags.`

	case "article_loop":
		vars = `
Available variables for ARTICLE LOOP templates:
- {{.Title}} — Page title (e.g., "Blog")
- {{.Header}} — Pre-rendered header HTML
- {{.Footer}} — Pre-rendered footer HTML
- {{.SiteName}} — The site name
- {{.Year}} — Current year
- {{range .Posts}} ... {{end}} — Loop over posts, each with:
  - {{.Title}} — Post title
  - {{.Slug}} — Post URL slug (link as /{{.Slug}})
  - {{.Excerpt}} — Post excerpt/summary
  - {{.PublishedAt}} — Publication date

Article loop templates show a list/grid of blog posts. Include header and footer.
Include the TailwindCSS CDN: <script src="https://cdn.tailwindcss.com"></script>
Wrap the page in a proper HTML structure with <html>, <head>, <body> tags.`

	default:
		vars = "\nGenerate a generic HTML template using TailwindCSS."
	}

	return base + "\n" + vars
}

// buildPreviewData creates dummy data appropriate for the template type,
// used to render a preview of the generated template.
func buildPreviewData(tmplType string) any {
	switch tmplType {
	case "page":
		return engine.PageData{
			SiteName:        "SmartPress",
			Title:           "Preview Page Title",
			Body:            "<p>This is preview content. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.</p><p>Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.</p>",
			Excerpt:         "A brief preview excerpt for the page.",
			MetaDescription: "Preview meta description for search engines",
			Slug:            "preview-page",
			PublishedAt:     "February 25, 2026",
			Header:          "<header class='bg-gray-800 text-white p-4'><nav class='max-w-6xl mx-auto flex justify-between items-center'><span class='text-xl font-bold'>SmartPress</span><div class='space-x-4'><a href='/' class='hover:text-gray-300'>Home</a><a href='/blog' class='hover:text-gray-300'>Blog</a></div></nav></header>",
			Footer:          "<footer class='bg-gray-800 text-gray-400 p-6 text-center text-sm'>&copy; 2026 SmartPress. All rights reserved.</footer>",
			Year:            2026,
		}
	case "article_loop":
		return engine.ListData{
			SiteName: "SmartPress",
			Title:    "Blog",
			Posts: []engine.PostItem{
				{Title: "Getting Started with SmartPress", Slug: "getting-started", Excerpt: "Learn how to set up your SmartPress CMS and create your first blog post.", PublishedAt: "February 25, 2026"},
				{Title: "Building Modern Websites", Slug: "modern-websites", Excerpt: "Discover the latest techniques for building fast, responsive websites.", PublishedAt: "February 24, 2026"},
				{Title: "AI-Powered Content Creation", Slug: "ai-content", Excerpt: "How artificial intelligence is transforming the way we create web content.", PublishedAt: "February 23, 2026"},
			},
			Header: "<header class='bg-gray-800 text-white p-4'><nav class='max-w-6xl mx-auto flex justify-between items-center'><span class='text-xl font-bold'>SmartPress</span><div class='space-x-4'><a href='/' class='hover:text-gray-300'>Home</a><a href='/blog' class='hover:text-gray-300'>Blog</a></div></nav></header>",
			Footer: "<footer class='bg-gray-800 text-gray-400 p-6 text-center text-sm'>&copy; 2026 SmartPress. All rights reserved.</footer>",
			Year:   2026,
		}
	default:
		// Header and footer use nil data (they only access .SiteName and .Year).
		return struct {
			SiteName string
			Year     int
		}{SiteName: "SmartPress", Year: 2026}
	}
}

// extractHTMLFromResponse strips markdown code fences and other non-HTML
// content from the AI's response, returning clean HTML.
func extractHTMLFromResponse(response string) string {
	response = strings.TrimSpace(response)

	// Remove markdown code fences: ```html ... ``` or ``` ... ```
	if strings.HasPrefix(response, "```") {
		// Find the end of the opening fence line.
		firstNewline := strings.Index(response, "\n")
		if firstNewline != -1 {
			response = response[firstNewline+1:]
		}
		// Remove the closing fence.
		if idx := strings.LastIndex(response, "```"); idx != -1 {
			response = response[:idx]
		}
	}

	return strings.TrimSpace(response)
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
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
