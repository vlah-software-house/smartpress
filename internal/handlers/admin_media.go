package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/gif" // register GIF decoder
	_ "image/png" // register PNG decoder
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	_ "golang.org/x/image/webp" // register WebP decoder
	"golang.org/x/image/draw"

	"smartpress/internal/middleware"
	"smartpress/internal/models"
	"smartpress/internal/render"
)

const (
	// maxUploadSize is the maximum allowed file upload size (50 MB).
	maxUploadSize = 50 << 20

	// thumbMaxWidth is the maximum thumbnail width in pixels.
	thumbMaxWidth = 400

	// thumbQuality is the JPEG quality for generated thumbnails.
	thumbQuality = 80

	// maxImagePixels caps the number of pixels to prevent memory bombs.
	// 10000x10000 = 100 million pixels, ~400 MB decoded in RGBA.
	maxImagePixels = 100_000_000

	// presignExpiry is how long a presigned URL for private files is valid.
	presignExpiry = 1 * time.Hour
)

// allowedMediaTypes defines MIME types accepted for upload.
var allowedMediaTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"image/webp":      true,
	"image/svg+xml":   true,
	"application/pdf": true,
}

// thumbableTypes are image types that support thumbnail generation.
// GIF is excluded to preserve animation; SVG is vector.
var thumbableTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

// MediaLibrary renders the media library admin page.
func (a *Admin) MediaLibrary(w http.ResponseWriter, r *http.Request) {
	if a.storageClient == nil {
		a.renderer.Page(w, r, "media_library", &render.PageData{
			Title:   "Media Library",
			Section: "media",
			Data:    map[string]any{"NoStorage": true},
		})
		return
	}

	page := 0
	items, _ := a.mediaStore.List(50, page*50)

	// Build URLs for each media item.
	type mediaView struct {
		models.Media
		URL      string
		ThumbURL string
	}
	var views []mediaView
	for _, m := range items {
		mv := mediaView{Media: m}
		if m.Bucket == a.storageClient.PublicBucket() {
			mv.URL = a.storageClient.FileURL(m.S3Key)
			if m.ThumbS3Key != nil {
				mv.ThumbURL = a.storageClient.FileURL(*m.ThumbS3Key)
			}
		}
		views = append(views, mv)
	}

	a.renderer.Page(w, r, "media_library", &render.PageData{
		Title:   "Media Library",
		Section: "media",
		Data: map[string]any{
			"Items":     views,
			"NoStorage": false,
		},
	})
}

// MediaUpload handles multipart file upload to S3.
func (a *Admin) MediaUpload(w http.ResponseWriter, r *http.Request) {
	if a.storageClient == nil {
		writeMediaError(w, "Object storage is not configured.", http.StatusServiceUnavailable)
		return
	}

	sess := middleware.SessionFromCtx(r.Context())

	// Limit request body to maxUploadSize + some overhead for form fields.
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize+1024)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeMediaError(w, "File too large. Maximum size is 50 MB.", http.StatusRequestEntityTooLarge)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeMediaError(w, "No file provided.", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file size.
	if header.Size > maxUploadSize {
		writeMediaError(w, "File too large. Maximum size is 50 MB.", http.StatusRequestEntityTooLarge)
		return
	}

	// Detect content type by sniffing the first 512 bytes.
	sniffBuf := make([]byte, 512)
	n, err := file.Read(sniffBuf)
	if err != nil && err != io.EOF {
		writeMediaError(w, "Failed to read file.", http.StatusInternalServerError)
		return
	}
	contentType := http.DetectContentType(sniffBuf[:n])

	// SVG detection: DetectContentType returns text/xml or application/xml for SVGs.
	if strings.HasSuffix(strings.ToLower(header.Filename), ".svg") &&
		(strings.Contains(contentType, "xml") || strings.Contains(contentType, "text/plain")) {
		contentType = "image/svg+xml"
	}

	if !allowedMediaTypes[contentType] {
		writeMediaError(w, fmt.Sprintf("File type %q is not allowed.", contentType), http.StatusBadRequest)
		return
	}

	// Seek back to start after sniffing.
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		writeMediaError(w, "Failed to process file.", http.StatusInternalServerError)
		return
	}

	// Determine target bucket.
	bucket := a.storageClient.PublicBucket()
	if r.FormValue("bucket") == "private" {
		bucket = a.storageClient.PrivateBucket()
	}

	// Generate a unique storage key.
	now := time.Now()
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = extensionFromType(contentType)
	}
	fileID := uuid.New().String()
	s3Key := fmt.Sprintf("media/%d/%02d/%s%s", now.Year(), now.Month(), fileID, ext)

	// Read the entire file into memory for upload and thumbnail generation.
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		writeMediaError(w, "Failed to read file.", http.StatusInternalServerError)
		return
	}

	// Upload original to S3.
	ctx := r.Context()
	if err := a.storageClient.Upload(ctx, bucket, s3Key, contentType, bytes.NewReader(fileBytes), int64(len(fileBytes))); err != nil {
		slog.Error("s3 upload failed", "error", err, "key", s3Key)
		writeMediaError(w, "Failed to upload file.", http.StatusInternalServerError)
		return
	}

	// Generate and upload thumbnail for supported image types.
	var thumbKey *string
	if thumbableTypes[contentType] {
		thumbData, err := generateThumbnail(bytes.NewReader(fileBytes), thumbMaxWidth)
		if err != nil {
			slog.Warn("thumbnail generation failed", "error", err, "key", s3Key)
		} else if thumbData != nil {
			tk := fmt.Sprintf("media/%d/%02d/%s_thumb.jpg", now.Year(), now.Month(), fileID)
			if err := a.storageClient.Upload(ctx, bucket, tk, "image/jpeg", bytes.NewReader(thumbData), int64(len(thumbData))); err != nil {
				slog.Warn("thumbnail upload failed", "error", err, "key", tk)
			} else {
				thumbKey = &tk
			}
		}
	}

	// Store metadata in PostgreSQL.
	altText := r.FormValue("alt_text")
	media := &models.Media{
		Filename:     fileID + ext,
		OriginalName: header.Filename,
		ContentType:  contentType,
		SizeBytes:    int64(len(fileBytes)),
		Bucket:       bucket,
		S3Key:        s3Key,
		ThumbS3Key:   thumbKey,
		UploaderID:   sess.UserID,
	}
	if altText != "" {
		media.AltText = &altText
	}

	created, err := a.mediaStore.Create(media)
	if err != nil {
		slog.Error("media db insert failed", "error", err, "key", s3Key)
		writeMediaError(w, "Failed to save file metadata.", http.StatusInternalServerError)
		return
	}

	// Build response URL.
	url := a.storageClient.FileURL(created.S3Key)
	var thumbURL string
	if created.ThumbS3Key != nil {
		thumbURL = a.storageClient.FileURL(*created.ThumbS3Key)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"id":        created.ID,
		"url":       url,
		"thumb_url": thumbURL,
		"filename":  created.OriginalName,
		"size":      created.HumanSize(),
		"type":      created.ContentType,
	})
}

// MediaDelete removes a media item from both S3 and the database.
func (a *Admin) MediaDelete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Delete from DB first (returns the row for S3 cleanup).
	deleted, err := a.mediaStore.Delete(id)
	if err != nil {
		slog.Error("media db delete failed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if deleted == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Clean up S3 objects (best-effort, don't fail the request).
	ctx := r.Context()
	if err := a.storageClient.Delete(ctx, deleted.Bucket, deleted.S3Key); err != nil {
		slog.Warn("s3 original delete failed", "error", err, "key", deleted.S3Key)
	}
	if deleted.ThumbS3Key != nil {
		if err := a.storageClient.Delete(ctx, deleted.Bucket, *deleted.ThumbS3Key); err != nil {
			slog.Warn("s3 thumbnail delete failed", "error", err, "key", *deleted.ThumbS3Key)
		}
	}

	// Return empty body for HTMX swap (removes the media card).
	w.WriteHeader(http.StatusOK)
}

// MediaServe provides the URL for a media item. Public files redirect to
// the direct S3 URL; private files get a time-limited presigned URL.
func (a *Admin) MediaServe(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	media, err := a.mediaStore.FindByID(id)
	if err != nil {
		slog.Error("media lookup failed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if media == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if media.Bucket == a.storageClient.PublicBucket() {
		http.Redirect(w, r, a.storageClient.FileURL(media.S3Key), http.StatusFound)
		return
	}

	// Private file — generate presigned URL.
	presigned, err := a.storageClient.PresignedURL(r.Context(), media.Bucket, media.S3Key, presignExpiry)
	if err != nil {
		slog.Error("presign failed", "error", err, "key", media.S3Key)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, presigned, http.StatusFound)
}

// generateThumbnail creates a JPEG thumbnail from an image, constrained
// to maxWidth while preserving aspect ratio. Returns nil if the image is
// already smaller than maxWidth.
func generateThumbnail(src io.Reader, maxWidth int) ([]byte, error) {
	// Decode config first to check dimensions without full decode.
	imgCfg, _, err := image.DecodeConfig(src)
	if err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	// Check for image bombs.
	if int64(imgCfg.Width)*int64(imgCfg.Height) > maxImagePixels {
		return nil, fmt.Errorf("image too large: %dx%d exceeds %d pixels", imgCfg.Width, imgCfg.Height, maxImagePixels)
	}

	// Skip thumbnail if image is already small enough.
	if imgCfg.Width <= maxWidth {
		return nil, nil
	}

	// Seek back to start for full decode.
	if seeker, ok := src.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return nil, fmt.Errorf("seek: %w", err)
		}
	} else {
		return nil, fmt.Errorf("source does not support seeking")
	}

	// Full decode.
	img, _, err := image.Decode(src)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	// Calculate thumbnail dimensions preserving aspect ratio.
	bounds := img.Bounds()
	ratio := float64(maxWidth) / float64(bounds.Dx())
	newWidth := maxWidth
	newHeight := int(float64(bounds.Dy()) * ratio)

	// Resize using CatmullRom (high quality).
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	// Encode to JPEG.
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: thumbQuality}); err != nil {
		return nil, fmt.Errorf("encode thumbnail: %w", err)
	}

	return buf.Bytes(), nil
}

// extensionFromType returns a file extension for known MIME types.
func extensionFromType(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/svg+xml":
		return ".svg"
	case "application/pdf":
		return ".pdf"
	default:
		return ""
	}
}

// writeMediaError writes a JSON error response for media operations.
func writeMediaError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
