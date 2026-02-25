// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Media represents a file uploaded to S3-compatible object storage.
// Metadata is stored in PostgreSQL; the file itself lives in the bucket.
type Media struct {
	ID           uuid.UUID `json:"id"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"original_name"`
	ContentType  string    `json:"content_type"`
	SizeBytes    int64     `json:"size_bytes"`
	Bucket       string    `json:"bucket"`
	S3Key        string    `json:"s3_key"`
	ThumbS3Key   *string   `json:"thumb_s3_key,omitempty"`
	AltText      *string   `json:"alt_text,omitempty"`
	UploaderID   uuid.UUID `json:"uploader_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// IsImage returns true if the media item is an image type.
func (m *Media) IsImage() bool {
	return strings.HasPrefix(m.ContentType, "image/")
}

// HumanSize returns a human-readable file size string.
func (m *Media) HumanSize() string {
	const (
		kb = 1024
		mb = 1024 * kb
	)
	switch {
	case m.SizeBytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(m.SizeBytes)/float64(mb))
	case m.SizeBytes >= kb:
		return fmt.Sprintf("%.0f KB", float64(m.SizeBytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", m.SizeBytes)
	}
}
