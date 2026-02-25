// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

// limiterEntry tracks request timestamps for a single client.
type limiterEntry struct {
	mu        sync.Mutex
	timestamps []time.Time
}

// RateLimiter provides per-IP rate limiting using a sliding window.
type RateLimiter struct {
	mu       sync.RWMutex
	clients  map[string]*limiterEntry
	limit    int           // max requests per window
	window   time.Duration // sliding window duration
	stopCh   chan struct{}
}

// NewRateLimiter creates a rate limiter that allows limit requests per window.
// It starts a background goroutine to clean up expired entries.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*limiterEntry),
		limit:   limit,
		window:  window,
		stopCh:  make(chan struct{}),
	}

	// Periodic cleanup of expired entries every 5 minutes.
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				rl.cleanup()
			case <-rl.stopCh:
				return
			}
		}
	}()

	return rl
}

// Stop terminates the background cleanup goroutine.
func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

// allow checks whether the given key is within the rate limit.
func (rl *RateLimiter) allow(key string) bool {
	rl.mu.RLock()
	entry, exists := rl.clients[key]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		// Double-check after acquiring write lock.
		entry, exists = rl.clients[key]
		if !exists {
			entry = &limiterEntry{}
			rl.clients[key] = entry
		}
		rl.mu.Unlock()
	}

	now := time.Now()
	cutoff := now.Add(-rl.window)

	entry.mu.Lock()
	defer entry.mu.Unlock()

	// Remove expired timestamps.
	valid := entry.timestamps[:0]
	for _, ts := range entry.timestamps {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	entry.timestamps = valid

	if len(entry.timestamps) >= rl.limit {
		return false
	}

	entry.timestamps = append(entry.timestamps, now)
	return true
}

// cleanup removes entries with no recent activity.
func (rl *RateLimiter) cleanup() {
	cutoff := time.Now().Add(-rl.window)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	for key, entry := range rl.clients {
		entry.mu.Lock()
		hasRecent := false
		for _, ts := range entry.timestamps {
			if ts.After(cutoff) {
				hasRecent = true
				break
			}
		}
		entry.mu.Unlock()

		if !hasRecent {
			delete(rl.clients, key)
		}
	}
}

// Middleware returns an HTTP middleware that rate-limits by client IP.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !rl.allow(ip) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP extracts the client's IP address, checking X-Forwarded-For
// and X-Real-IP headers for proxied requests.
func clientIP(r *http.Request) string {
	// Check X-Forwarded-For first (may contain multiple IPs).
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first (leftmost) IP — the original client.
		if idx := strings.IndexByte(xff, ','); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP.
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr (strip port).
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}
