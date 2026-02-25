// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

// Package handlers contains the HTTP handlers for the YaaiCMS CMS.
// Handlers are grouped by concern (admin, public, auth) and receive
// their dependencies through the handler struct.
package handlers

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/google/uuid"

	"yaaicms/internal/ai"
	"yaaicms/internal/cache"
	"yaaicms/internal/engine"
	"yaaicms/internal/middleware"
	"yaaicms/internal/models"
	"yaaicms/internal/render"
	"yaaicms/internal/session"
	"yaaicms/internal/slug"
	"yaaicms/internal/storage"
	"yaaicms/internal/store"
)

// AIProviderInfo holds display information about a configured AI provider.
// Used by the Settings page to show which providers are available.
type AIProviderInfo struct {
	Name      string // "openai", "gemini", "claude", "mistral"
	Label     string // Human-friendly label
	HasKey    bool   // Whether an API key is configured
	Active    bool   // Whether this is the currently active provider
	Model     string // Configured model name
	KeyEnvVar string // Environment variable name for the key
}

// AIConfig holds the AI provider configuration visible to admin handlers.
// Intentionally excludes actual API keys — only exposes what the UI needs.
type AIConfig struct {
	ActiveProvider string
	Providers      []AIProviderInfo
}

// Admin groups all admin panel HTTP handlers and their dependencies.
type Admin struct {
	renderer      *render.Renderer
	sessions      *session.Store
	contentStore  *store.ContentStore
	userStore     *store.UserStore
	templateStore *store.TemplateStore
	mediaStore    *store.MediaStore
	revisionStore *store.RevisionStore
	storageClient *storage.Client
	engine        *engine.Engine
	pageCache     *cache.PageCache
	cacheLog      *store.CacheLogStore
	aiRegistry    *ai.Registry
	aiConfig      *AIConfig
}

// NewAdmin creates a new Admin handler group with the given dependencies.
// storageClient and mediaStore may be nil if S3 is not configured.
func NewAdmin(renderer *render.Renderer, sessions *session.Store, contentStore *store.ContentStore, userStore *store.UserStore, templateStore *store.TemplateStore, mediaStore *store.MediaStore, revisionStore *store.RevisionStore, storageClient *storage.Client, eng *engine.Engine, pageCache *cache.PageCache, cacheLog *store.CacheLogStore, aiRegistry *ai.Registry, aiCfg *AIConfig) *Admin {
	return &Admin{
		renderer:      renderer,
		sessions:      sessions,
		contentStore:  contentStore,
		userStore:     userStore,
		templateStore: templateStore,
		mediaStore:    mediaStore,
		revisionStore: revisionStore,
		storageClient: storageClient,
		engine:        eng,
		pageCache:     pageCache,
		cacheLog:      cacheLog,
		aiRegistry:    aiRegistry,
		aiConfig:      aiCfg,
	}
}

// Dashboard renders the admin dashboard page with real stats.
func (a *Admin) Dashboard(w http.ResponseWriter, r *http.Request) {
	postCount, _ := a.contentStore.CountByType(models.ContentTypePost)
	pageCount, _ := a.contentStore.CountByType(models.ContentTypePage)
	users, _ := a.userStore.List()
	var mediaCount int
	if a.mediaStore != nil {
		mediaCount, _ = a.mediaStore.Count()
	}

	a.renderer.Page(w, r, "dashboard", &render.PageData{
		Title:   "Dashboard",
		Section: "dashboard",
		Data: map[string]any{
			"PostCount":  postCount,
			"PageCount":  pageCount,
			"UserCount":  len(users),
			"MediaCount": mediaCount,
		},
	})
}

// --- Posts CRUD ---

// PostsList renders the posts management page.
func (a *Admin) PostsList(w http.ResponseWriter, r *http.Request) {
	posts, err := a.contentStore.ListByType(models.ContentTypePost)
	if err != nil {
		slog.Error("list posts failed", "error", err)
	}

	a.renderer.Page(w, r, "posts_list", &render.PageData{
		Title:   "Posts",
		Section: "posts",
		Data:    map[string]any{"Items": posts},
	})
}

// PostNew renders the new post form.
func (a *Admin) PostNew(w http.ResponseWriter, r *http.Request) {
	a.renderer.Page(w, r, "content_form", &render.PageData{
		Title:   "New Post",
		Section: "posts",
		Data: map[string]any{
			"ContentType": "post",
			"IsNew":       true,
		},
	})
}

// PostCreate handles the new post form submission.
func (a *Admin) PostCreate(w http.ResponseWriter, r *http.Request) {
	sess := middleware.SessionFromCtx(r.Context())
	a.createContent(w, r, models.ContentTypePost, sess)
}

// PostEdit renders the edit post form.
func (a *Admin) PostEdit(w http.ResponseWriter, r *http.Request) {
	a.editContent(w, r, "posts")
}

// PostUpdate handles the edit post form submission.
func (a *Admin) PostUpdate(w http.ResponseWriter, r *http.Request) {
	a.updateContent(w, r, "posts")
}

// PostDelete handles post deletion.
func (a *Admin) PostDelete(w http.ResponseWriter, r *http.Request) {
	a.deleteContent(w, r, "posts")
}

// --- Pages CRUD ---

// PagesList renders the pages management page.
func (a *Admin) PagesList(w http.ResponseWriter, r *http.Request) {
	pages, err := a.contentStore.ListByType(models.ContentTypePage)
	if err != nil {
		slog.Error("list pages failed", "error", err)
	}

	a.renderer.Page(w, r, "pages_list", &render.PageData{
		Title:   "Pages",
		Section: "pages",
		Data:    map[string]any{"Items": pages},
	})
}

// PageNew renders the new page form.
func (a *Admin) PageNew(w http.ResponseWriter, r *http.Request) {
	a.renderer.Page(w, r, "content_form", &render.PageData{
		Title:   "New Page",
		Section: "pages",
		Data: map[string]any{
			"ContentType": "page",
			"IsNew":       true,
		},
	})
}

// PageCreate handles the new page form submission.
func (a *Admin) PageCreate(w http.ResponseWriter, r *http.Request) {
	sess := middleware.SessionFromCtx(r.Context())
	a.createContent(w, r, models.ContentTypePage, sess)
}

// PageEdit renders the edit page form.
func (a *Admin) PageEdit(w http.ResponseWriter, r *http.Request) {
	a.editContent(w, r, "pages")
}

// PageUpdate handles the edit page form submission.
func (a *Admin) PageUpdate(w http.ResponseWriter, r *http.Request) {
	a.updateContent(w, r, "pages")
}

// PageDelete handles page deletion.
func (a *Admin) PageDelete(w http.ResponseWriter, r *http.Request) {
	a.deleteContent(w, r, "pages")
}

// --- Shared content helpers ---

// createContent handles creating a new post or page from the form.
func (a *Admin) createContent(w http.ResponseWriter, r *http.Request, contentType models.ContentType, sess *session.Data) {
	title := r.FormValue("title")
	body := r.FormValue("body")
	status := models.ContentStatus(r.FormValue("status"))
	contentSlug := r.FormValue("slug")
	excerpt := r.FormValue("excerpt")
	metaDesc := r.FormValue("meta_description")
	metaKw := r.FormValue("meta_keywords")
	featuredImageIDStr := r.FormValue("featured_image_id")

	// Validate inputs.
	if errMsg := validateContent(title, contentSlug, body); errMsg != "" {
		section := "posts"
		if contentType == models.ContentTypePage {
			section = "pages"
		}
		a.renderer.Page(w, r, "content_form", &render.PageData{
			Title:   "New " + string(contentType),
			Section: section,
			Data: map[string]any{
				"ContentType": string(contentType),
				"IsNew":       true,
				"Error":       errMsg,
			},
		})
		return
	}
	if errMsg := validateMetadata(excerpt, metaDesc, metaKw); errMsg != "" {
		section := "posts"
		if contentType == models.ContentTypePage {
			section = "pages"
		}
		a.renderer.Page(w, r, "content_form", &render.PageData{
			Title:   "New " + string(contentType),
			Section: section,
			Data: map[string]any{
				"ContentType": string(contentType),
				"IsNew":       true,
				"Error":       errMsg,
			},
		})
		return
	}

	if contentSlug == "" {
		contentSlug = slug.Generate(title)
	}

	if status == "" {
		status = models.ContentStatusDraft
	}

	// Determine body format from the form (Markdown editor sets this).
	bodyFormat := models.BodyFormat(r.FormValue("body_format"))
	if bodyFormat != models.BodyFormatHTML {
		bodyFormat = models.BodyFormatMarkdown // default for new content
	}

	c := &models.Content{
		Type:       contentType,
		Title:      title,
		Slug:       contentSlug,
		Body:       body,
		BodyFormat: bodyFormat,
		Status:     status,
		AuthorID:   sess.UserID,
	}
	if excerpt != "" {
		c.Excerpt = &excerpt
	}
	if metaDesc != "" {
		c.MetaDescription = &metaDesc
	}
	if metaKw != "" {
		c.MetaKeywords = &metaKw
	}
	if fid, err := uuid.Parse(featuredImageIDStr); err == nil {
		c.FeaturedImageID = &fid
	}

	created, err := a.contentStore.Create(c)
	if err != nil {
		slog.Error("create content failed", "error", err, "type", contentType)
		section := "posts"
		if contentType == models.ContentTypePage {
			section = "pages"
		}
		a.renderer.Page(w, r, "content_form", &render.PageData{
			Title:   "New " + string(contentType),
			Section: section,
			Data: map[string]any{
				"ContentType": string(contentType),
				"IsNew":       true,
				"Error":       "Failed to create. The slug may already exist.",
				"Item":        c,
			},
		})
		return
	}

	// Invalidate cache for the new content (homepage may show it in listings).
	a.invalidateContentCache(r.Context(), created.ID, created.Slug, "create")

	if contentType == models.ContentTypePage {
		http.Redirect(w, r, "/admin/pages", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
	}
}

// editContent renders the edit form for a content item.
func (a *Admin) editContent(w http.ResponseWriter, r *http.Request, section string) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, err := a.contentStore.FindByID(id)
	if err != nil || item == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	contentType := "post"
	title := "Edit Post"
	if section == "pages" {
		contentType = "page"
		title = "Edit Page"
	}

	data := map[string]any{
		"ContentType": contentType,
		"IsNew":       false,
		"Item":        item,
	}

	// Resolve featured image URL for display in the form.
	if item.FeaturedImageID != nil && a.mediaStore != nil && a.storageClient != nil {
		if media, err := a.mediaStore.FindByID(*item.FeaturedImageID); err == nil && media != nil {
			if media.Bucket == a.storageClient.PublicBucket() {
				data["FeaturedImageURL"] = a.storageClient.FileURL(media.S3Key)
			}
			if media.ThumbS3Key != nil {
				data["FeaturedImageThumbURL"] = a.storageClient.FileURL(*media.ThumbS3Key)
			}
			data["FeaturedImageName"] = media.OriginalName
		}
	}

	// Load revisions for the history panel.
	revisions, err := a.revisionStore.ListByContentID(item.ID)
	if err != nil {
		slog.Error("failed to load revisions", "error", err)
	}
	data["Revisions"] = revisions

	a.renderer.Page(w, r, "content_form", &render.PageData{
		Title:   title,
		Section: section,
		Data:    data,
	})
}

// updateContent handles the edit form submission for a content item.
// Before applying changes, it snapshots the current state as a revision.
func (a *Admin) updateContent(w http.ResponseWriter, r *http.Request, section string) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, err := a.contentStore.FindByID(id)
	if err != nil || item == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Capture the old state BEFORE applying form values (for revision snapshot).
	oldTitle := item.Title
	oldBody := item.Body
	oldSlug := item.Slug
	oldExcerpt := item.Excerpt
	oldStatus := string(item.Status)
	oldMetaDesc := item.MetaDescription
	oldMetaKw := item.MetaKeywords
	oldFeaturedImageID := item.FeaturedImageID

	title := r.FormValue("title")
	body := r.FormValue("body")
	newSlug := r.FormValue("slug")
	excerpt := r.FormValue("excerpt")
	metaDesc := r.FormValue("meta_description")
	metaKw := r.FormValue("meta_keywords")
	featuredImageIDStr := r.FormValue("featured_image_id")
	revisionMessage := strings.TrimSpace(r.FormValue("revision_message"))

	// Validate inputs.
	if errMsg := validateContent(title, newSlug, body); errMsg != "" {
		a.renderer.Page(w, r, "content_form", &render.PageData{
			Title:   "Edit",
			Section: section,
			Data: map[string]any{
				"ContentType": string(item.Type),
				"IsNew":       false,
				"Item":        item,
				"Error":       errMsg,
			},
		})
		return
	}
	if errMsg := validateMetadata(excerpt, metaDesc, metaKw); errMsg != "" {
		a.renderer.Page(w, r, "content_form", &render.PageData{
			Title:   "Edit",
			Section: section,
			Data: map[string]any{
				"ContentType": string(item.Type),
				"IsNew":       false,
				"Item":        item,
				"Error":       errMsg,
			},
		})
		return
	}

	// Apply new values.
	item.Title = title
	item.Body = body
	item.Status = models.ContentStatus(r.FormValue("status"))
	item.Slug = newSlug

	// Update body format from the form.
	bodyFormat := models.BodyFormat(r.FormValue("body_format"))
	if bodyFormat != models.BodyFormatHTML {
		bodyFormat = models.BodyFormatMarkdown
	}
	oldBodyFormat := item.BodyFormat
	item.BodyFormat = bodyFormat

	if item.Slug == "" {
		item.Slug = slug.Generate(item.Title)
	}

	if excerpt != "" {
		item.Excerpt = &excerpt
	} else {
		item.Excerpt = nil
	}
	if metaDesc != "" {
		item.MetaDescription = &metaDesc
	} else {
		item.MetaDescription = nil
	}
	if metaKw != "" {
		item.MetaKeywords = &metaKw
	} else {
		item.MetaKeywords = nil
	}

	// Update featured image (posts only).
	if fid, err := uuid.Parse(featuredImageIDStr); err == nil {
		item.FeaturedImageID = &fid
	} else {
		item.FeaturedImageID = nil
	}

	// Create revision snapshot of the OLD state before persisting changes.
	sess := middleware.SessionFromCtx(r.Context())
	rev := &models.ContentRevision{
		ContentID:       item.ID,
		Title:           oldTitle,
		Slug:            oldSlug,
		Body:            oldBody,
		BodyFormat:      oldBodyFormat,
		Excerpt:         oldExcerpt,
		Status:          oldStatus,
		MetaDescription: oldMetaDesc,
		MetaKeywords:    oldMetaKw,
		FeaturedImageID: oldFeaturedImageID,
		RevisionTitle:   revisionMessage,
		CreatedBy:       sess.UserID,
	}

	created, revErr := a.revisionStore.Create(rev)
	if revErr != nil {
		slog.Error("failed to create revision", "content_id", item.ID, "error", revErr)
		// Non-fatal: proceed with the update even if revision creation fails.
	}

	if err := a.contentStore.Update(item); err != nil {
		slog.Error("update content failed", "error", err)
		a.renderer.Page(w, r, "content_form", &render.PageData{
			Title:   "Edit",
			Section: section,
			Data: map[string]any{
				"ContentType": string(item.Type),
				"IsNew":       false,
				"Item":        item,
				"Error":       "Failed to update. The slug may already exist.",
			},
		})
		return
	}

	// Generate AI revision title + changelog in the background.
	if created != nil {
		go a.generateRevisionMeta(created.ID, rev, item, revisionMessage)
	}

	a.invalidateContentCache(r.Context(), item.ID, item.Slug, "update")
	http.Redirect(w, r, "/admin/"+section, http.StatusSeeOther)
}

// generateRevisionMeta uses AI to create a short title and changelog for a
// revision, comparing the old state (rev) with the new state (updated item).
// Runs in a background goroutine — errors are logged but don't affect the user.
func (a *Admin) generateRevisionMeta(revID uuid.UUID, old *models.ContentRevision, updated *models.Content, userMessage string) {
	// Build a concise diff summary for the AI.
	var changes []string
	if old.Title != updated.Title {
		changes = append(changes, fmt.Sprintf("Title: %q -> %q", truncateStr(old.Title, 80), truncateStr(updated.Title, 80)))
	}
	if old.Body != updated.Body {
		oldLen := len(old.Body)
		newLen := len(updated.Body)
		changes = append(changes, fmt.Sprintf("Body: changed (%d -> %d chars)", oldLen, newLen))
	}
	if old.Slug != updated.Slug {
		changes = append(changes, fmt.Sprintf("Slug: %q -> %q", old.Slug, updated.Slug))
	}
	if old.Status != string(updated.Status) {
		changes = append(changes, fmt.Sprintf("Status: %s -> %s", old.Status, string(updated.Status)))
	}
	if ptrStr(old.Excerpt) != ptrStr(updated.Excerpt) {
		changes = append(changes, "Excerpt: updated")
	}
	if ptrStr(old.MetaDescription) != ptrStr(updated.MetaDescription) {
		changes = append(changes, "Meta description: updated")
	}
	if ptrStr(old.MetaKeywords) != ptrStr(updated.MetaKeywords) {
		changes = append(changes, "Meta keywords: updated")
	}
	if ptrUUID(old.FeaturedImageID) != ptrUUID(updated.FeaturedImageID) {
		changes = append(changes, "Featured image: changed")
	}

	if len(changes) == 0 {
		changes = append(changes, "No visible changes")
	}

	diffSummary := strings.Join(changes, "\n")

	// Generate revision title if the user didn't provide one.
	ctx := context.Background()
	revTitle := userMessage
	if revTitle == "" {
		prompt := fmt.Sprintf("Changes made:\n%s\n\nOld title: %q\nNew title: %q",
			diffSummary, truncateStr(old.Title, 100), truncateStr(updated.Title, 100))

		systemPrompt := `You are a version control assistant. Generate a very short revision title
(max 60 characters) that summarizes the changes made, like a git commit message.
Output ONLY the title text, nothing else. Use imperative mood (e.g. "Update title and body content").`

		result, err := a.aiRegistry.GenerateForTask(ctx, ai.TaskLight, systemPrompt, prompt)
		if err != nil {
			slog.Warn("ai revision title failed", "error", err)
			revTitle = "Content updated"
		} else {
			revTitle = strings.TrimSpace(result)
			revTitle = strings.Trim(revTitle, `"'`)
			if len(revTitle) > 80 {
				revTitle = revTitle[:77] + "..."
			}
		}
	}

	// Generate a changelog describing what changed.
	changelogPrompt := fmt.Sprintf("Changes:\n%s", diffSummary)

	changelogSystem := `You are a version control assistant. Generate a brief changelog (2-4 bullet points)
describing what changed in this content revision. Each bullet should start with "- ".
Be concise and factual. Output ONLY the bullet points, nothing else.`

	changelog, err := a.aiRegistry.GenerateForTask(ctx, ai.TaskLight, changelogSystem, changelogPrompt)
	if err != nil {
		slog.Warn("ai revision changelog failed", "error", err)
		changelog = diffSummary
	} else {
		changelog = strings.TrimSpace(changelog)
	}

	if err := a.revisionStore.UpdateMeta(revID, revTitle, changelog); err != nil {
		slog.Error("failed to update revision meta", "id", revID, "error", err)
	}
}

// RevisionRestore restores a content item to the state captured in a revision.
// Returns an HTML fragment that triggers a page redirect via HTMX.
func (a *Admin) RevisionRestore(w http.ResponseWriter, r *http.Request) {
	revIDStr := chi.URLParam(r, "revisionID")
	revID, err := uuid.Parse(revIDStr)
	if err != nil {
		http.Error(w, "Invalid revision ID", http.StatusBadRequest)
		return
	}

	rev, err := a.revisionStore.FindByID(revID)
	if err != nil || rev == nil {
		http.Error(w, "Revision not found", http.StatusNotFound)
		return
	}

	// Load the content item to check it exists and to create a "pre-restore" revision.
	item, err := a.contentStore.FindByID(rev.ContentID)
	if err != nil || item == nil {
		http.Error(w, "Content not found", http.StatusNotFound)
		return
	}

	// Create a revision of the current state before restoring.
	sess := middleware.SessionFromCtx(r.Context())
	preRestore := &models.ContentRevision{
		ContentID:       item.ID,
		Title:           item.Title,
		Slug:            item.Slug,
		Body:            item.Body,
		BodyFormat:      item.BodyFormat,
		Excerpt:         item.Excerpt,
		Status:          string(item.Status),
		MetaDescription: item.MetaDescription,
		MetaKeywords:    item.MetaKeywords,
		FeaturedImageID: item.FeaturedImageID,
		RevisionTitle:   "Before restore",
		RevisionLog:     fmt.Sprintf("- State before restoring to revision from %s", rev.CreatedAt.Format("Jan 2, 2006 15:04")),
		CreatedBy:       sess.UserID,
	}
	if _, err := a.revisionStore.Create(preRestore); err != nil {
		slog.Error("failed to create pre-restore revision", "error", err)
	}

	// Apply the revision data to the content item.
	item.Title = rev.Title
	item.Slug = rev.Slug
	item.Body = rev.Body
	item.BodyFormat = rev.BodyFormat
	item.Excerpt = rev.Excerpt
	item.Status = models.ContentStatus(rev.Status)
	item.MetaDescription = rev.MetaDescription
	item.MetaKeywords = rev.MetaKeywords
	item.FeaturedImageID = rev.FeaturedImageID

	if err := a.contentStore.Update(item); err != nil {
		slog.Error("restore revision failed", "error", err)
		http.Error(w, "Failed to restore revision", http.StatusInternalServerError)
		return
	}

	a.invalidateContentCache(r.Context(), item.ID, item.Slug, "restore")

	// Determine section for redirect.
	section := "posts"
	if item.Type == models.ContentTypePage {
		section = "pages"
	}
	redirectURL := fmt.Sprintf("/admin/%s/%s", section, item.ID)

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", redirectURL)
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// RevisionUpdateTitle updates a revision's user-provided title.
func (a *Admin) RevisionUpdateTitle(w http.ResponseWriter, r *http.Request) {
	revIDStr := chi.URLParam(r, "revisionID")
	revID, err := uuid.Parse(revIDStr)
	if err != nil {
		http.Error(w, "Invalid revision ID", http.StatusBadRequest)
		return
	}

	rev, err := a.revisionStore.FindByID(revID)
	if err != nil || rev == nil {
		http.Error(w, "Revision not found", http.StatusNotFound)
		return
	}

	newTitle := strings.TrimSpace(r.FormValue("revision_title"))
	if newTitle == "" {
		writeAIError(w, "Title cannot be empty.")
		return
	}

	if err := a.revisionStore.UpdateMeta(revID, newTitle, rev.RevisionLog); err != nil {
		slog.Error("failed to update revision title", "error", err)
		writeAIError(w, "Failed to update revision title.")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<span class="text-xs font-medium text-gray-900">%s</span>`, html.EscapeString(newTitle))
}

// ptrStr safely dereferences a *string pointer, returning "" if nil.
func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ptrUUID safely dereferences a *uuid.UUID pointer, returning "" if nil.
func ptrUUID(u *uuid.UUID) string {
	if u == nil {
		return ""
	}
	return u.String()
}

// truncateStr cuts a string to maxLen, appending "..." if truncated.
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// deleteContent handles content deletion.
func (a *Admin) deleteContent(w http.ResponseWriter, r *http.Request, section string) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Look up the slug before deleting so we can invalidate its cache entry.
	item, _ := a.contentStore.FindByID(id)

	if err := a.contentStore.Delete(id); err != nil {
		slog.Error("delete content failed", "error", err)
	} else if item != nil {
		a.invalidateContentCache(r.Context(), id, item.Slug, "delete")
	}

	http.Redirect(w, r, "/admin/"+section, http.StatusSeeOther)
}

// --- Template management ---

// TemplatesList renders the templates management page with real data.
func (a *Admin) TemplatesList(w http.ResponseWriter, r *http.Request) {
	templates, err := a.templateStore.List()
	if err != nil {
		slog.Error("list templates failed", "error", err)
	}

	a.renderer.Page(w, r, "templates_list", &render.PageData{
		Title:   "AI Design",
		Section: "templates",
		Data:    map[string]any{"Templates": templates},
	})
}

// TemplateNew renders the new template form.
func (a *Admin) TemplateNew(w http.ResponseWriter, r *http.Request) {
	a.renderer.Page(w, r, "template_form", &render.PageData{
		Title:   "New Template",
		Section: "templates",
		Data:    map[string]any{"IsNew": true},
	})
}

// TemplateCreate handles the new template form submission.
func (a *Admin) TemplateCreate(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	tmplType := models.TemplateType(r.FormValue("type"))
	htmlContent := r.FormValue("html_content")

	// Validate input lengths.
	if errMsg := validateTemplate(name, htmlContent); errMsg != "" {
		a.renderer.Page(w, r, "template_form", &render.PageData{
			Title:   "New Template",
			Section: "templates",
			Data: map[string]any{
				"IsNew": true,
				"Error": errMsg,
				"Item":  &models.Template{Name: name, Type: tmplType, HTMLContent: htmlContent},
			},
		})
		return
	}

	// Validate the template syntax before saving.
	if err := a.engine.ValidateTemplate(htmlContent); err != nil {
		a.renderer.Page(w, r, "template_form", &render.PageData{
			Title:   "New Template",
			Section: "templates",
			Data: map[string]any{
				"IsNew": true,
				"Error": "Template syntax error: " + err.Error(),
				"Item":  &models.Template{Name: name, Type: tmplType, HTMLContent: htmlContent},
			},
		})
		return
	}

	t := &models.Template{
		Name:        name,
		Type:        tmplType,
		HTMLContent: htmlContent,
	}

	created, err := a.templateStore.Create(t)
	if err != nil {
		slog.Error("create template failed", "error", err)
		a.renderer.Page(w, r, "template_form", &render.PageData{
			Title:   "New Template",
			Section: "templates",
			Data: map[string]any{
				"IsNew": true,
				"Error": "Failed to create template.",
				"Item":  t,
			},
		})
		return
	}

	// New templates aren't active yet, but log the event for auditing.
	a.cacheLog.Log("template", created.ID, "create")
	http.Redirect(w, r, "/admin/templates", http.StatusSeeOther)
}

// TemplateEdit renders the edit template form.
func (a *Admin) TemplateEdit(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, err := a.templateStore.FindByID(id)
	if err != nil || item == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	a.renderer.Page(w, r, "template_form", &render.PageData{
		Title:   "Edit Template",
		Section: "templates",
		Data: map[string]any{
			"IsNew": false,
			"Item":  item,
		},
	})
}

// TemplateUpdate handles the edit template form submission.
func (a *Admin) TemplateUpdate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, err := a.templateStore.FindByID(id)
	if err != nil || item == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	item.Name = r.FormValue("name")
	htmlContent := r.FormValue("html_content")

	// Validate syntax.
	if err := a.engine.ValidateTemplate(htmlContent); err != nil {
		a.renderer.Page(w, r, "template_form", &render.PageData{
			Title:   "Edit Template",
			Section: "templates",
			Data: map[string]any{
				"IsNew": false,
				"Error": "Template syntax error: " + err.Error(),
				"Item":  item,
			},
		})
		return
	}

	item.HTMLContent = htmlContent
	if err := a.templateStore.Update(item); err != nil {
		slog.Error("update template failed", "error", err)
	} else {
		// Template content changed — invalidate L1 (compiled) and L2 (rendered pages).
		a.invalidateTemplateCache(r.Context(), item.ID, "update")
	}

	http.Redirect(w, r, "/admin/templates", http.StatusSeeOther)
}

// TemplateActivate sets a template as active for its type.
func (a *Admin) TemplateActivate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := a.templateStore.Activate(id); err != nil {
		slog.Error("activate template failed", "error", err)
	} else {
		// Activation changes which template renders for a type — clear everything.
		a.invalidateAllTemplateCache(r.Context(), id, "update")
	}

	http.Redirect(w, r, "/admin/templates", http.StatusSeeOther)
}

// TemplateDelete handles template deletion.
func (a *Admin) TemplateDelete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := a.templateStore.Delete(id); err != nil {
		slog.Error("delete template failed", "error", err)
	} else {
		a.invalidateTemplateCache(r.Context(), id, "delete")
	}

	http.Redirect(w, r, "/admin/templates", http.StatusSeeOther)
}

// TemplatePreview renders a preview of a template with dummy data.
func (a *Admin) TemplatePreview(w http.ResponseWriter, r *http.Request) {
	htmlContent := r.FormValue("html_content")
	if htmlContent == "" {
		http.Error(w, "No template content", http.StatusBadRequest)
		return
	}

	data := engine.PageData{
		SiteName:        "YaaiCMS",
		Title:           "Preview Page Title",
		Body:            "<p>This is preview content. Lorem ipsum dolor sit amet, consectetur adipiscing elit.</p>",
		Excerpt:         "A brief preview excerpt.",
		MetaDescription: "Preview meta description",
		Slug:            "preview-page",
		PublishedAt:     "February 24, 2026",
		Header:          "<header><nav>Header Preview</nav></header>",
		Footer:          "<footer><p>Footer Preview</p></footer>",
		Year:            2026,
	}

	result, err := a.engine.ValidateAndRender(htmlContent, data)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		safeErr := html.EscapeString(err.Error())
		w.Write([]byte(`<div class="p-4 bg-red-50 border border-red-200 rounded text-red-800 text-sm">Template error: ` + safeErr + `</div>`))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(result)
}

// UsersList renders the user management page with real data.
func (a *Admin) UsersList(w http.ResponseWriter, r *http.Request) {
	users, err := a.userStore.List()
	if err != nil {
		slog.Error("list users failed", "error", err)
	}

	a.renderer.Page(w, r, "users_list", &render.PageData{
		Title:   "Users",
		Section: "users",
		Data:    map[string]any{"Users": users},
	})
}

// UserResetTwoFA resets another user's 2FA, forcing re-setup on next login.
func (a *Admin) UserResetTwoFA(w http.ResponseWriter, r *http.Request) {
	sess := middleware.SessionFromCtx(r.Context())

	idStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Cannot reset your own 2FA.
	if targetID == sess.UserID {
		http.Error(w, "Cannot reset your own 2FA", http.StatusForbidden)
		return
	}

	if err := a.userStore.ResetTOTP(targetID); err != nil {
		slog.Error("reset 2fa failed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	slog.Info("2fa reset by admin", "admin", sess.Email, "target_user", targetID)
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// UserNew renders the new user creation form.
func (a *Admin) UserNew(w http.ResponseWriter, r *http.Request) {
	a.renderer.Page(w, r, "user_form", &render.PageData{
		Title:   "New User",
		Section: "users",
		Data:    map[string]any{},
	})
}

// UserCreate handles the new user form submission.
func (a *Admin) UserCreate(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.FormValue("email"))
	displayName := strings.TrimSpace(r.FormValue("display_name"))
	password := r.FormValue("password")
	role := models.Role(r.FormValue("role"))

	// Validate inputs.
	var errMsg string
	switch {
	case email == "":
		errMsg = "Email is required."
	case displayName == "":
		errMsg = "Display name is required."
	case len(password) < 8:
		errMsg = "Password must be at least 8 characters."
	case role != models.RoleAdmin && role != models.RoleEditor && role != models.RoleAuthor:
		errMsg = "Invalid role."
	}

	if errMsg != "" {
		a.renderer.Page(w, r, "user_form", &render.PageData{
			Title:   "New User",
			Section: "users",
			Data: map[string]any{
				"Error":       errMsg,
				"Email":       email,
				"DisplayName": displayName,
				"Role":        string(role),
			},
		})
		return
	}

	// Check for duplicate email.
	existing, _ := a.userStore.FindByEmail(email)
	if existing != nil {
		a.renderer.Page(w, r, "user_form", &render.PageData{
			Title:   "New User",
			Section: "users",
			Data: map[string]any{
				"Error":       "A user with this email already exists.",
				"Email":       email,
				"DisplayName": displayName,
				"Role":        string(role),
			},
		})
		return
	}

	if _, err := a.userStore.Create(email, password, displayName, role); err != nil {
		slog.Error("create user failed", "error", err)
		a.renderer.Page(w, r, "user_form", &render.PageData{
			Title:   "New User",
			Section: "users",
			Data: map[string]any{
				"Error":       "Failed to create user.",
				"Email":       email,
				"DisplayName": displayName,
				"Role":        string(role),
			},
		})
		return
	}

	sess := middleware.SessionFromCtx(r.Context())
	slog.Info("user created", "admin", sess.Email, "new_user", email, "role", role)

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/admin/users")
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// --- Cache invalidation helpers ---

// invalidateContentCache purges the L2 page cache for a content item and
// logs the event. Always invalidates the homepage too since post listings
// or the "home" page might have changed.
func (a *Admin) invalidateContentCache(ctx context.Context, contentID uuid.UUID, contentSlug, action string) {
	a.pageCache.InvalidatePage(ctx, cache.SlugKey(contentSlug))
	a.pageCache.InvalidateHomepage(ctx)
	a.cacheLog.Log("content", contentID, action)
}

// invalidateTemplateCache purges both L1 (compiled template) and L2 (all
// rendered pages) caches, since any template change can affect any page.
func (a *Admin) invalidateTemplateCache(ctx context.Context, templateID uuid.UUID, action string) {
	a.engine.InvalidateTemplate(templateID.String())
	a.pageCache.InvalidateAll(ctx)
	a.cacheLog.Log("template", templateID, action)
}

// invalidateAllTemplateCache clears the entire L1 cache and all L2 pages.
// Used for template activation which changes the active template for a type.
func (a *Admin) invalidateAllTemplateCache(ctx context.Context, templateID uuid.UUID, action string) {
	a.engine.InvalidateAllTemplates()
	a.pageCache.InvalidateAll(ctx)
	a.cacheLog.Log("template", templateID, action)
}

// SettingsPage renders the settings page.
func (a *Admin) SettingsPage(w http.ResponseWriter, r *http.Request) {
	a.renderer.Page(w, r, "settings", &render.PageData{
		Title:   "Settings",
		Section: "settings",
		Data: map[string]any{
			"Providers": a.aiConfig.Providers,
		},
	})
}
