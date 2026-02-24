# Phase 2: Dynamic Template Engine

## Step 6: Templates DB & Renderer
- [x] Templates table with versioning
- [x] Go template compiler (reads from DB, compiles in memory)
- [x] Public page serving: route → query → inject content → render
- [x] Admin template CRUD + live preview
- [x] Default template seeding (header, footer, page, article_loop)
- [x] Sample content seeding (home page, hello-world post)

## Step 7: Caching Layer
- [x] Level 1: In-memory compiled template cache (sync.RWMutex map, keyed by ID+version)
- [x] Level 2: Full-page HTML cache in Valkey (5min TTL, `page:*` keys)
- [x] Cache invalidation on admin "Save" actions (content + template mutations)
- [x] Audit log to cache_invalidation_log table
