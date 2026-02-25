// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"yaaicms/internal/engine"
	"yaaicms/internal/middleware"
	"yaaicms/internal/models"
	"yaaicms/internal/render"
	"yaaicms/internal/slug"
)

// --- AI Assistant Endpoints ---
//
// These handlers power the content editor's AI assistant panel.
// Each endpoint accepts form values (title, body, tone) from HTMX requests,
// calls the active AI provider, and returns HTML fragments that get swapped
// into the assistant panel's result areas.

// AIGenerateContent generates a full article from a user-provided topic prompt.
// Returns an HTML fragment with the generated content and an "Apply" button
// that fills the body textarea.
func (a *Admin) AIGenerateContent(w http.ResponseWriter, r *http.Request) {
	prompt := strings.TrimSpace(r.FormValue("ai_content_prompt"))
	contentType := r.FormValue("content_type")

	if prompt == "" {
		writeAIError(w, "Please describe what you'd like to write about.")
		return
	}

	if contentType == "" {
		contentType = "article"
	}

	if !a.checkPromptSafety(w, r, prompt) {
		return
	}

	systemPrompt := fmt.Sprintf(`You are an expert content writer for a CMS. Write a complete %s based on the user's description.

Rules:
- Output ONLY the article body as clean Markdown.
- Use ## and ### for subheadings (not # — the CMS adds the title separately).
- Use standard Markdown syntax: **bold**, *italic*, > blockquotes, - lists, 1. numbered lists, [links](url), etc.
- Do NOT wrap the output in code fences.
- Write 3-6 well-structured paragraphs with subheadings where appropriate.
- Make the content informative, engaging, and ready to publish.`, contentType)

	result, err := a.aiRegistry.Generate(r.Context(), systemPrompt, prompt)
	if err != nil {
		slog.Error("ai generate content failed", "error", err)
		writeAIError(w, "AI request failed. Check your provider configuration.")
		return
	}

	result = extractHTMLFromResponse(result)

	fragment := fmt.Sprintf(
		`<div class="space-y-3">
			<div class="text-xs text-gray-700 bg-gray-50 rounded p-3 max-h-48 overflow-y-auto prose prose-sm">%s</div>
			<button type="button"
				onclick="if(window._markdownEditor){window._markdownEditor.value(%s)}else{document.getElementById('body').value=%s}; document.getElementById('body').dispatchEvent(new Event('input'))"
				class="w-full rounded-md bg-indigo-600 px-3 py-2 text-sm font-medium text-white hover:bg-indigo-500 transition-colors">
				Apply to Content
			</button>
		</div>`,
		result,
		quoteJSString(result),
		quoteJSString(result),
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fragment))
}

// AIGenerateImage generates an image using the active AI provider's image
// generation capability (e.g., DALL-E 3 for OpenAI). The generated image is
// uploaded to S3 and stored as a media record. Returns an HTML fragment with
// a preview and a "Use as Featured Image" button.
func (a *Admin) AIGenerateImage(w http.ResponseWriter, r *http.Request) {
	prompt := strings.TrimSpace(r.FormValue("ai_image_prompt"))
	if prompt == "" {
		writeAIError(w, "Please describe the image you'd like to generate.")
		return
	}

	if a.storageClient == nil || a.mediaStore == nil {
		writeAIError(w, "Object storage is not configured. Cannot save generated images.")
		return
	}

	if !a.aiRegistry.SupportsImageGeneration() {
		writeAIError(w, "Image generation requires an OpenAI API key for DALL-E. Configure OPENAI_API_KEY to enable this feature.")
		return
	}

	if !a.checkPromptSafety(w, r, prompt) {
		return
	}

	sess := middleware.SessionFromCtx(r.Context())

	// Generate the image.
	imgBytes, contentType, err := a.aiRegistry.GenerateImage(r.Context(), prompt)
	if err != nil {
		slog.Error("ai generate image failed", "error", err)
		writeAIError(w, "Image generation failed. Check your provider configuration and API limits.")
		return
	}

	// Upload to S3 as a media item (same pipeline as manual uploads).
	now := time.Now()
	fileID := uuid.New().String()
	ext := ".png"
	if contentType == "image/jpeg" {
		ext = ".jpg"
	} else if contentType == "image/webp" {
		ext = ".webp"
	}
	s3Key := fmt.Sprintf("media/%d/%02d/%s%s", now.Year(), now.Month(), fileID, ext)
	bucket := a.storageClient.PublicBucket()

	ctx := r.Context()
	if err := a.storageClient.Upload(ctx, bucket, s3Key, contentType, bytes.NewReader(imgBytes), int64(len(imgBytes))); err != nil {
		slog.Error("ai image s3 upload failed", "error", err, "key", s3Key)
		writeAIError(w, "Failed to upload generated image.")
		return
	}

	// Generate thumbnail.
	var thumbKey *string
	thumbData, err := generateThumbnail(bytes.NewReader(imgBytes), thumbMaxWidth)
	if err != nil {
		slog.Warn("ai image thumbnail failed", "error", err)
	} else if thumbData != nil {
		tk := fmt.Sprintf("media/%d/%02d/%s_thumb.jpg", now.Year(), now.Month(), fileID)
		if err := a.storageClient.Upload(ctx, bucket, tk, "image/jpeg", bytes.NewReader(thumbData), int64(len(thumbData))); err != nil {
			slog.Warn("ai image thumbnail upload failed", "error", err)
		} else {
			thumbKey = &tk
		}
	}

	// Create media record. Derive a descriptive filename from the prompt.
	altText := truncate(prompt, 500)
	safeName := slug.Generate(prompt)
	if len(safeName) > 80 {
		safeName = safeName[:80]
		// Trim at the last hyphen to avoid cutting a word in half.
		if i := strings.LastIndex(safeName, "-"); i > 20 {
			safeName = safeName[:i]
		}
	}
	if safeName == "" {
		safeName = "ai-generated"
	}
	media := &models.Media{
		Filename:     fileID + ext,
		OriginalName: safeName + ext,
		ContentType:  contentType,
		SizeBytes:    int64(len(imgBytes)),
		Bucket:       bucket,
		S3Key:        s3Key,
		ThumbS3Key:   thumbKey,
		AltText:      &altText,
		UploaderID:   sess.UserID,
	}

	created, err := a.mediaStore.Create(media)
	if err != nil {
		slog.Error("ai image media insert failed", "error", err)
		writeAIError(w, "Failed to save image metadata.")
		return
	}

	// Build image URLs for the response.
	imgURL := a.storageClient.FileURL(created.S3Key)
	var thumbURL string
	if created.ThumbS3Key != nil {
		thumbURL = a.storageClient.FileURL(*created.ThumbS3Key)
	} else {
		thumbURL = imgURL
	}

	// Return JSON when the client requests it (used by the media picker modal),
	// otherwise return an HTML fragment (used by the featured image HTMX flow).
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"id":        created.ID.String(),
			"url":       imgURL,
			"thumb_url": thumbURL,
			"filename":  created.OriginalName,
			"alt_text":  altText,
		})
		return
	}

	fragment := fmt.Sprintf(
		`<div class="space-y-3">
			<img src="%s" alt="%s" class="w-full rounded-lg shadow-sm border border-gray-200">
			<button type="button"
				onclick="document.getElementById('featured_image_id').value = '%s';
				         document.getElementById('featured-image-preview').src = '%s';
				         document.getElementById('featured-image-container').classList.remove('hidden');
				         document.getElementById('featured-image-empty').classList.add('hidden')"
				class="w-full rounded-md bg-indigo-600 px-3 py-2 text-sm font-medium text-white hover:bg-indigo-500 transition-colors">
				Use as Featured Image
			</button>
		</div>`,
		html.EscapeString(thumbURL),
		html.EscapeString(altText),
		created.ID.String(),
		html.EscapeString(imgURL),
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fragment))
}

// AISuggestTitle generates title suggestions based on the content body.
// Returns an HTML fragment with clickable title options.
func (a *Admin) AISuggestTitle(w http.ResponseWriter, r *http.Request) {
	body := r.FormValue("body")
	title := r.FormValue("title")

	if body == "" && title == "" {
		writeAIError(w, "Please write some content or a working title first.")
		return
	}

	promptText := title + " " + body
	if !a.checkPromptSafety(w, r, truncate(promptText, 2000)) {
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

	if !a.checkPromptSafety(w, r, truncate(body, 2000)) {
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

	promptText := title + " " + body
	if !a.checkPromptSafety(w, r, truncate(promptText, 2000)) {
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

	if !a.checkPromptSafety(w, r, truncate(body, 3000)) {
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
The content uses Markdown formatting — preserve all Markdown syntax. Output ONLY the rewritten content, nothing else.`, toneDesc)

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
			<button type="button" onclick="if(window._markdownEditor){window._markdownEditor.value(%s)}else{document.getElementById('body').value=%s}"
				class="w-full rounded-md bg-indigo-50 border border-indigo-200 px-2 py-1 text-xs font-medium text-indigo-700 hover:bg-indigo-100 transition-colors">
				Apply to Content
			</button>
		</div>`,
		html.EscapeString(result),
		quoteJSString(result),
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

	promptText := title + " " + body
	if !a.checkPromptSafety(w, r, truncate(promptText, 2000)) {
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

// --- Provider Switching ---

// AISetProvider switches the active AI provider at runtime.
// Accepts a "provider" form value and returns an HTML fragment that replaces
// the provider selector UI to reflect the new active provider.
func (a *Admin) AISetProvider(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("provider"))
	if name == "" {
		writeAIError(w, "No provider specified.")
		return
	}

	if err := a.aiRegistry.SetActive(name); err != nil {
		slog.Warn("failed to switch AI provider", "provider", name, "error", err)
		writeAIError(w, fmt.Sprintf("Cannot switch to %q: provider not available (no API key configured).", name))
		return
	}

	// Update the cached AIConfig so the UI reflects the change.
	a.refreshAIConfig(name)

	slog.Info("ai provider switched", "provider", name)

	// If this is an HTMX request from the AI assistant panel, return the
	// updated provider selector dropdown.
	source := r.Header.Get("HX-Target")
	if r.Header.Get("HX-Request") == "true" && source == "ai-provider-select" {
		a.writeProviderSelector(w, name)
		return
	}

	// HTMX request from settings page or non-HTMX: redirect to settings.
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/admin/settings")
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/admin/settings", http.StatusSeeOther)
}

// AIProviderStatus returns the current provider selector fragment.
// Used by the content form to load the initial state of the provider dropdown.
func (a *Admin) AIProviderStatus(w http.ResponseWriter, r *http.Request) {
	a.writeProviderSelector(w, a.aiRegistry.ActiveName())
}

// refreshAIConfig updates the cached AIConfig after a provider switch.
func (a *Admin) refreshAIConfig(activeName string) {
	a.aiConfig.ActiveProvider = activeName
	for i := range a.aiConfig.Providers {
		a.aiConfig.Providers[i].Active = a.aiConfig.Providers[i].Name == activeName
	}
}

// writeProviderSelector writes an HTML fragment containing the provider
// dropdown selector with the current active provider highlighted.
func (a *Admin) writeProviderSelector(w http.ResponseWriter, active string) {
	var sb strings.Builder
	sb.WriteString(`<select name="provider" `)
	sb.WriteString(`hx-post="/admin/ai/set-provider" `)
	sb.WriteString(`hx-target="#ai-provider-select" `)
	sb.WriteString(`hx-swap="innerHTML" `)
	sb.WriteString(`hx-include="[name='csrf_token']" `)
	sb.WriteString(`class="block w-full rounded-md border border-gray-300 px-2 py-1 text-xs shadow-sm `)
	sb.WriteString(`focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none">`)

	for _, p := range a.aiConfig.Providers {
		if !p.HasKey {
			continue
		}
		selected := ""
		if p.Name == active {
			selected = " selected"
		}
		sb.WriteString(fmt.Sprintf(
			`<option value="%s"%s>%s (%s)</option>`,
			html.EscapeString(p.Name),
			selected,
			html.EscapeString(p.Label),
			html.EscapeString(p.Model),
		))
	}
	sb.WriteString(`</select>`)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(sb.String()))
}

// --- Helper functions ---

// checkPromptSafety runs the user prompt through the moderation API.
// Returns true if the prompt is safe (or if no moderator is available).
// If the prompt is flagged, writes an error response and returns false.
func (a *Admin) checkPromptSafety(w http.ResponseWriter, r *http.Request, prompt string) bool {
	result, err := a.aiRegistry.CheckPrompt(r.Context(), prompt)
	if err != nil {
		slog.Warn("moderation check failed, allowing prompt", "error", err)
		return true // fail open — providers have their own safety filters
	}

	if result.Safe {
		return true
	}

	categories := strings.Join(result.Categories, ", ")
	slog.Warn("prompt flagged by moderation", "categories", categories)

	msg := fmt.Sprintf(
		"Your prompt was flagged for: %s. Please reformulate your request and try again.",
		categories,
	)
	writeAIError(w, msg)
	return false
}

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

	// Check prompt safety before generating.
	modResult, err := a.aiRegistry.CheckPrompt(r.Context(), prompt)
	if err != nil {
		slog.Warn("moderation check failed for template prompt", "error", err)
	} else if !modResult.Safe {
		categories := strings.Join(modResult.Categories, ", ")
		slog.Warn("template prompt flagged by moderation", "categories", categories)
		writeJSON(w, http.StatusOK, templateGenResponse{
			Error: fmt.Sprintf("Your prompt was flagged for: %s. Please reformulate your request and try again.", categories),
		})
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
You generate complete, production-ready HTML+TailwindCSS templates for a CMS called YaaiCMS.

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
- {{.SiteName}} — The site name (e.g., "YaaiCMS")
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
- {{.FeaturedImageURL}} — Public URL of the featured image (empty string if none)
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
  - {{.FeaturedImageURL}} — Public URL of the featured image (empty string if none)
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
			SiteName:         "YaaiCMS",
			Title:            "Preview Page Title",
			Body:             "<p>This is preview content. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.</p><p>Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.</p>",
			Excerpt:          "A brief preview excerpt for the page.",
			MetaDescription:  "Preview meta description for search engines",
			FeaturedImageURL: "https://placehold.co/1200x630/0f172a/e2e8f0?text=Featured+Image",
			Slug:             "preview-page",
			PublishedAt:      "February 25, 2026",
			Header:           "<header class='bg-gray-800 text-white p-4'><nav class='max-w-6xl mx-auto flex justify-between items-center'><span class='text-xl font-bold'>YaaiCMS</span><div class='space-x-4'><a href='/' class='hover:text-gray-300'>Home</a><a href='/blog' class='hover:text-gray-300'>Blog</a></div></nav></header>",
			Footer:           "<footer class='bg-gray-800 text-gray-400 p-6 text-center text-sm'>&copy; 2026 YaaiCMS. All rights reserved.</footer>",
			Year:             2026,
		}
	case "article_loop":
		return engine.ListData{
			SiteName: "YaaiCMS",
			Title:    "Blog",
			Posts: []engine.PostItem{
				{Title: "Getting Started with YaaiCMS", Slug: "getting-started", Excerpt: "Learn how to set up your YaaiCMS CMS and create your first blog post.", FeaturedImageURL: "https://placehold.co/800x450/0f172a/e2e8f0?text=Post+1", PublishedAt: "February 25, 2026"},
				{Title: "Building Modern Websites", Slug: "modern-websites", Excerpt: "Discover the latest techniques for building fast, responsive websites.", FeaturedImageURL: "https://placehold.co/800x450/1e3a5f/e2e8f0?text=Post+2", PublishedAt: "February 24, 2026"},
				{Title: "AI-Powered Content Creation", Slug: "ai-content", Excerpt: "How artificial intelligence is transforming the way we create web content.", FeaturedImageURL: "https://placehold.co/800x450/3b0764/e2e8f0?text=Post+3", PublishedAt: "February 23, 2026"},
			},
			Header: "<header class='bg-gray-800 text-white p-4'><nav class='max-w-6xl mx-auto flex justify-between items-center'><span class='text-xl font-bold'>YaaiCMS</span><div class='space-x-4'><a href='/' class='hover:text-gray-300'>Home</a><a href='/blog' class='hover:text-gray-300'>Blog</a></div></nav></header>",
			Footer: "<footer class='bg-gray-800 text-gray-400 p-6 text-center text-sm'>&copy; 2026 YaaiCMS. All rights reserved.</footer>",
			Year:   2026,
		}
	default:
		// Header and footer use nil data (they only access .SiteName and .Year).
		return struct {
			SiteName string
			Year     int
		}{SiteName: "YaaiCMS", Year: 2026}
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
