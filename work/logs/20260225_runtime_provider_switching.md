# Runtime AI Provider Switching

**Date:** 2026-02-25
**Branch:** feat/featured-image
**Commit:** 0cb84cb

## Changes

### New Endpoints
- `POST /admin/ai/set-provider` — switches the active AI provider at runtime
- `GET /admin/ai/provider-status` — returns the provider selector dropdown HTML fragment

### Settings Page (`settings.html`)
- Removed Phase 3 placeholder text
- Added "Activate" buttons on provider cards with configured API keys
- Active provider shows checkmark icon and "Currently active" label
- Providers without keys show "Set ENV_VAR to enable" helper text
- HTMX-powered switching: click Activate -> HX-Redirect reloads the settings page

### AI Assistant Panel (`content_form.html`)
- Added Provider dropdown selector at the top of the AI Assistant panel
- Dropdown loads via `hx-get="/admin/ai/provider-status"` on panel open
- Changing the dropdown fires `hx-post="/admin/ai/set-provider"` with CSRF token
- Shows provider label + model name for each option (e.g., "OpenAI (gpt-5.2-pro)")

### Image Generation (`ai/image.go`)
- Image generation now uses `findImageGenerator()` which prefers OpenAI (DALL-E) regardless of the active text provider
- Falls back to any other provider implementing `ImageGenerator` if OpenAI is unavailable
- `SupportsImageGeneration()` checks all providers, not just the active one

### Handler (`admin_ai.go`)
- `AISetProvider` handler: validates provider exists, calls `aiRegistry.SetActive()`, refreshes `aiConfig` cache
- `AIProviderStatus` handler: returns current provider selector HTML fragment
- `writeProviderSelector` helper: generates `<select>` with HTMX attributes
- `refreshAIConfig` helper: updates cached `AIConfig` after provider switch
- Response routing: HTMX from assistant panel gets selector fragment; HTMX from settings gets HX-Redirect; non-HTMX gets 303 redirect
