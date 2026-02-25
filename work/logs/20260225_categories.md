# Categories Feature

**Date:** 2026-02-25
**Branch:** feat/categories
**Status:** Complete

## Summary

Implemented a hierarchical category system for posts with parent-child support, drag-and-drop tree reordering via SortableJS, and single-category-per-post assignment.

## Changes

### Database
- `00014_create_categories.sql`: Created `categories` table with self-referential `parent_id` FK, `sort_order`, and standard timestamps. Added `category_id` FK on `content` and `content_revisions` tables.

### Models
- `models/category.go`: Category struct with ID, Name, Slug, Description, ParentID, SortOrder, Children (tree), Depth, PostCount.
- `models/content.go`: Added `CategoryID *uuid.UUID` to Content and ContentRevision.

### Store
- `store/category.go`: Full CRUD (List with post counts, Tree, FlatTree for dropdowns, FindByID, Create, Update, Delete), transactional Reorder, NextSortOrder helper. Tree uses recursive `buildTree` with depth tracking.
- `store/content.go`: Added `category_id` to columns, scan, Create, Update queries.
- `store/revision.go`: Added `category_id` to columns, scan, Create queries for revision snapshots.

### Handlers
- `handlers/admin.go`: Added `categoryStore` field; wired category loading in PostNew and editContent (FlatTree for select dropdown); category_id parsing in createContent and updateContent; category tracking in revision snapshots and restore; CategoriesList, CategoryCreate, CategoryUpdate, CategoryDelete, CategoryReorder handlers.

### Templates
- `categories.html`: Category manager with SortableJS nested drag-and-drop tree, Alpine.js CRUD forms (add, edit modal, delete with confirmation), Save Order button, recursive `categoryNode` template.
- `content_form.html`: Category `<select>` dropdown for posts with hierarchical indentation using `catIndent` helper.
- `base.html`: Added Categories link to both desktop and mobile sidebars with folder icon.

### Render
- `render.go`: Added `catIndent` (depth-based `&nbsp;` indentation) and `uuidEq` (*uuid.UUID comparison) template helpers.

### Router
- `router.go`: Added `/admin/categories` route group (GET, POST, PUT, DELETE, POST /reorder).

### Wiring
- `main.go`: Created and wired `CategoryStore` into `NewAdmin`.
- `handler_test.go`: Updated `NewAdmin` call with `categoryStore` parameter.
