package handlers

import (
	"strings"
	"unicode/utf8"
)

// Validation limits for content and template fields.
const (
	maxTitleLen       = 300
	maxSlugLen        = 300
	maxBodyLen        = 100_000
	maxExcerptLen     = 1_000
	maxMetaDescLen    = 500
	maxMetaKeywordLen = 500
	maxTemplateNameLen = 200
	maxTemplateHTMLLen = 500_000
)

// validateContent checks content form inputs and returns the first error found.
func validateContent(title, slug, body string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return "Title is required."
	}
	if utf8.RuneCountInString(title) > maxTitleLen {
		return "Title is too long (max 300 characters)."
	}
	if utf8.RuneCountInString(slug) > maxSlugLen {
		return "Slug is too long (max 300 characters)."
	}
	if utf8.RuneCountInString(body) > maxBodyLen {
		return "Body is too long (max 100,000 characters)."
	}
	return ""
}

// validateMetadata checks optional SEO metadata fields.
func validateMetadata(excerpt, metaDesc, metaKw string) string {
	if utf8.RuneCountInString(excerpt) > maxExcerptLen {
		return "Excerpt is too long (max 1,000 characters)."
	}
	if utf8.RuneCountInString(metaDesc) > maxMetaDescLen {
		return "Meta description is too long (max 500 characters)."
	}
	if utf8.RuneCountInString(metaKw) > maxMetaKeywordLen {
		return "Meta keywords are too long (max 500 characters)."
	}
	return ""
}

// validateTemplate checks template form inputs and returns the first error found.
func validateTemplate(name, htmlContent string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Template name is required."
	}
	if utf8.RuneCountInString(name) > maxTemplateNameLen {
		return "Template name is too long (max 200 characters)."
	}
	htmlContent = strings.TrimSpace(htmlContent)
	if htmlContent == "" {
		return "Template HTML content is required."
	}
	if utf8.RuneCountInString(htmlContent) > maxTemplateHTMLLen {
		return "Template HTML content is too long (max 500,000 characters)."
	}
	return ""
}
