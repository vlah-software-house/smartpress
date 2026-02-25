// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.
import { test, expect } from '@playwright/test';

test.describe('Settings & Users', () => {
  test('settings page shows AI provider configuration', async ({ page }) => {
    await page.goto('/admin/settings');
    await expect(page.locator('h1')).toContainText('Settings');

    // AI providers should be listed — use getByText with exact match
    // to avoid matching env var names like OPENAI_API_KEY.
    const main = page.locator('main');
    await expect(main.getByText('OpenAI', { exact: true })).toBeVisible();
    await expect(main.getByText('Google Gemini', { exact: true })).toBeVisible();
    await expect(main.getByText('Anthropic Claude', { exact: true })).toBeVisible();
    await expect(main.getByText('Mistral', { exact: true })).toBeVisible();
  });

  test('users page lists admin user', async ({ page }) => {
    await page.goto('/admin/users');
    await expect(page.locator('h1')).toContainText('Users');

    // The seeded admin should appear in the main content area.
    const main = page.locator('main');
    await expect(main.locator('text=admin@yaaicms.local')).toBeVisible();
  });
});
