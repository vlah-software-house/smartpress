# Step 7: Caching Layer

**Date:** 2026-02-24
**Branch:** feat/dynamic-template-engine
**Status:** Complete

## What was built

### L1: In-memory compiled template cache (`internal/engine/cache.go`)
- `sync.RWMutex`-protected map keyed by template ID + version
- Compiled `*template.Template` stored in memory — avoids re-parsing HTML strings on every request
- Auto-invalidated on version bump (ID+version key means any update is a cache miss)
- `InvalidateTemplate(id)` — purge a specific template
- `InvalidateAllTemplates()` — full clear (used on template activation)

### L2: Valkey full-page HTML cache (`internal/cache/page.go`)
- Key format: `page:<slug>` and `page:_homepage`
- 5-minute TTL (configurable)
- `Get`/`Set` for cache read/write
- `InvalidatePage(slug)` — single page purge
- `InvalidateHomepage()` — homepage purge
- `InvalidateAll()` — SCAN + DEL all `page:*` keys

### Cache invalidation audit log (`internal/store/cache_log.go`)
- Logs events to `cache_invalidation_log` table (entity_type, entity_id, action)
- Best-effort — failures are logged but don't break the flow

### Invalidation hooks in admin handlers
- Content create/update/delete → purge L2 for that slug + homepage
- Template update/delete → purge L1 for that template + all L2 pages
- Template activate → clear entire L1 + all L2 pages (type-wide change)

### Wiring
- Engine now holds L1 cache internally, compiles once per template version
- Public handlers check L2 before calling engine, store results on miss
- Admin handlers call invalidation helpers after successful mutations
- `main.go` creates `PageCache` and `CacheLogStore`, passes to handlers

## Performance verified
- First homepage request: 5.5ms (DB queries + template compilation + render)
- Second homepage request: 0.16ms (L2 Valkey hit) — **35x faster**
- Valkey keys confirmed: `page:_homepage`, `page:hello-world`
- TTL confirmed: 300s (5 minutes)
- Content hash identical across cached/uncached responses
