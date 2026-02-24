# Phase 2: Dynamic Template Engine

## Step 6: Templates DB & Renderer
- [ ] Templates table with versioning
- [ ] Go template compiler (reads from DB, compiles in memory)
- [ ] Public page serving: route → query → inject content → render

## Step 7: Caching Layer
- [ ] Level 1: In-memory compiled template cache (sync.RWMutex map)
- [ ] Level 2: Full-page HTML cache in Valkey
- [ ] Cache invalidation on admin "Save" actions
