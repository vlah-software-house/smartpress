#!/usr/bin/env bash
# Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
# Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
# All rights reserved. See LICENSE for details.
# build-public-css.sh — Compile TailwindCSS from database-stored templates.
#
# This script extracts all template HTML from the YaaiCMS database,
# writes it to a temporary directory (.tailwind-content/), and runs
# the Tailwind CLI to produce a minified CSS file containing only the
# classes actually used in public-facing templates.
#
# Usage:
#   ./scripts/build-public-css.sh
#
# Environment variables (reads from .env/.secrets if present):
#   POSTGRES_HOST     (default: localhost)
#   POSTGRES_PORT     (default: 5432)
#   POSTGRES_USER     (default: yaaicms)
#   POSTGRES_PASSWORD (default: changeme)
#   POSTGRES_DB       (default: yaaicms)
#
# Output: web/static/css/public.css
#
# Requirements: psql, npx (with tailwindcss installed)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Load environment from .env or .secrets if available.
for envfile in "$PROJECT_DIR/.env" "$PROJECT_DIR/.secrets"; do
    if [[ -f "$envfile" ]]; then
        # shellcheck disable=SC2046
        export $(grep -v '^#' "$envfile" | grep -v '^\s*$' | xargs)
    fi
done

DB_HOST="${POSTGRES_HOST:-localhost}"
DB_PORT="${POSTGRES_PORT:-5432}"
DB_USER="${POSTGRES_USER:-yaaicms}"
DB_NAME="${POSTGRES_DB:-yaaicms}"
export PGPASSWORD="${POSTGRES_PASSWORD:-changeme}"

CONTENT_DIR="$PROJECT_DIR/.tailwind-content"
OUTPUT_FILE="$PROJECT_DIR/web/static/css/public.css"

echo "==> Extracting templates from database..."
mkdir -p "$CONTENT_DIR"
rm -f "$CONTENT_DIR"/*.html

# Extract each template's HTML content into a separate file.
TEMPLATE_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
    -t -A -c "SELECT count(*) FROM templates;")

if [[ "$TEMPLATE_COUNT" -eq 0 ]]; then
    echo "    No templates found in database. Generating minimal CSS."
    # Create an empty placeholder so Tailwind still runs.
    echo "<!-- no templates -->" > "$CONTENT_DIR/empty.html"
else
    echo "    Found $TEMPLATE_COUNT template(s)."
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -t -A -c "SELECT id || '.html', html_content FROM templates;" | \
    while IFS='|' read -r filename content; do
        echo "$content" > "$CONTENT_DIR/$filename"
    done
fi

echo "==> Compiling public TailwindCSS..."
cd "$PROJECT_DIR"
npx tailwindcss -i web/static/css/input.css -o "$OUTPUT_FILE" --minify 2>&1

echo "==> Done: $OUTPUT_FILE ($(wc -c < "$OUTPUT_FILE") bytes)"

# Clean up extracted templates.
rm -rf "$CONTENT_DIR"
