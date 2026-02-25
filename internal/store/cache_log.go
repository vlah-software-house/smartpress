// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

// cache_log.go records cache invalidation events in the database for
// audit and debugging purposes. Each entry captures what was invalidated,
// when, and why (create/update/delete).
package store

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

// CacheLogStore handles cache invalidation log operations.
type CacheLogStore struct {
	db *sql.DB
}

// NewCacheLogStore creates a new CacheLogStore.
func NewCacheLogStore(db *sql.DB) *CacheLogStore {
	return &CacheLogStore{db: db}
}

// Log records a cache invalidation event.
func (s *CacheLogStore) Log(entityType string, entityID uuid.UUID, action string) {
	_, err := s.db.Exec(`
		INSERT INTO cache_invalidation_log (entity_type, entity_id, action)
		VALUES ($1, $2, $3)
	`, entityType, entityID, action)
	if err != nil {
		// Log but don't fail — cache logging is best-effort.
		slog.Warn("failed to log cache invalidation",
			"entity_type", entityType,
			"entity_id", entityID,
			"action", action,
			"error", err,
		)
		return
	}
	slog.Debug("cache invalidation logged",
		"entity_type", entityType,
		"entity_id", entityID,
		"action", action,
	)
}

// RecentEntries returns the most recent cache invalidation events for
// debugging. Limited to the specified count.
func (s *CacheLogStore) RecentEntries(limit int) ([]CacheLogEntry, error) {
	rows, err := s.db.Query(`
		SELECT id, entity_type, entity_id, action, invalidated_at
		FROM cache_invalidation_log
		ORDER BY invalidated_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("query cache log: %w", err)
	}
	defer rows.Close()

	var entries []CacheLogEntry
	for rows.Next() {
		var e CacheLogEntry
		if err := rows.Scan(&e.ID, &e.EntityType, &e.EntityID, &e.Action, &e.InvalidatedAt); err != nil {
			return nil, fmt.Errorf("scan cache log: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// CacheLogEntry represents a single cache invalidation event.
type CacheLogEntry struct {
	ID            int64
	EntityType    string
	EntityID      uuid.UUID
	Action        string
	InvalidatedAt string
}
