// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

// Package storage provides an S3-compatible object storage client for
// uploading, deleting, and serving media files. It wraps the AWS SDK v2
// and is configured for path-style access (required by CEPH/Hetzner).
package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Client wraps an S3 client for media operations on two buckets.
type Client struct {
	s3            *s3.Client
	presigner     *s3.PresignClient
	publicBucket  string
	privateBucket string
	endpoint      string
	publicURL     string // optional CDN/direct URL for public files
}

// New creates an S3 storage client configured for CEPH/Hetzner with
// path-style addressing. Returns (nil, nil) if endpoint or credentials
// are empty, allowing the app to start without storage.
func New(endpoint, region, accessKey, secretKey, publicBucket, privateBucket, publicURL string) (*Client, error) {
	if endpoint == "" || accessKey == "" || secretKey == "" {
		return nil, nil
	}

	// Strip trailing slash from endpoint for consistent URL building.
	endpoint = strings.TrimRight(endpoint, "/")

	// Build S3 client with static credentials and path-style access.
	s3Client := s3.New(s3.Options{
		Region: region,
		BaseEndpoint: aws.String(endpoint),
		Credentials: credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		UsePathStyle: true,
	})

	return &Client{
		s3:            s3Client,
		presigner:     s3.NewPresignClient(s3Client),
		publicBucket:  publicBucket,
		privateBucket: privateBucket,
		endpoint:      endpoint,
		publicURL:     strings.TrimRight(publicURL, "/"),
	}, nil
}

// Upload stores an object in the specified bucket. For the public bucket,
// objects are set to public-read ACL so they can be served directly.
func (c *Client) Upload(ctx context.Context, bucket, key, contentType string, body io.Reader, size int64) error {
	input := &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	}

	// Public bucket objects get public-read ACL.
	if bucket == c.publicBucket {
		input.ACL = s3types.ObjectCannedACLPublicRead
	}

	_, err := c.s3.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("s3 upload %s/%s: %w", bucket, key, err)
	}
	return nil
}

// Download retrieves an object from the specified bucket and returns its
// contents as a byte slice. Used for regenerating image variants from the
// original stored in S3.
func (c *Client) Download(ctx context.Context, bucket, key string) ([]byte, error) {
	output, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 download %s/%s: %w", bucket, key, err)
	}
	defer output.Body.Close()
	data, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("s3 read body %s/%s: %w", bucket, key, err)
	}
	return data, nil
}

// Delete removes an object from the specified bucket.
func (c *Client) Delete(ctx context.Context, bucket, key string) error {
	_, err := c.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("s3 delete %s/%s: %w", bucket, key, err)
	}
	return nil
}

// FileURL returns the public URL for a file in the public bucket.
// Uses the configured public URL if set, otherwise builds a path-style URL.
func (c *Client) FileURL(key string) string {
	if c.publicURL != "" {
		return c.publicURL + "/" + key
	}
	return c.endpoint + "/" + c.publicBucket + "/" + key
}

// PresignedURL generates a pre-signed GET URL for a private object.
// The URL is valid for the specified duration (max 7 days per S3 spec).
func (c *Client) PresignedURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	req, err := c.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", fmt.Errorf("s3 presign %s/%s: %w", bucket, key, err)
	}
	return req.URL, nil
}

// PublicBucket returns the name of the public bucket.
func (c *Client) PublicBucket() string {
	return c.publicBucket
}

// PrivateBucket returns the name of the private bucket.
func (c *Client) PrivateBucket() string {
	return c.privateBucket
}

// ExtractS3Key extracts the S3 object key from a public file URL.
// Returns the key and true if the URL matches the storage URL pattern,
// or ("", false) if it doesn't belong to this storage.
func (c *Client) ExtractS3Key(rawURL string) (string, bool) {
	// Try publicURL prefix first (CDN or custom domain).
	if c.publicURL != "" {
		prefix := c.publicURL + "/"
		if strings.HasPrefix(rawURL, prefix) {
			return rawURL[len(prefix):], true
		}
	}

	// Try endpoint/bucket prefix (path-style S3).
	prefix := c.endpoint + "/" + c.publicBucket + "/"
	if strings.HasPrefix(rawURL, prefix) {
		return rawURL[len(prefix):], true
	}

	return "", false
}
