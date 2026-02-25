// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.
import { test, expect } from '@playwright/test';

test.describe('Pages CRUD', () => {
  const uniqueSlug = `e2e-page-${Date.now()}`;

  test('lists existing pages', async ({ page }) => {
    await page.goto('/admin/pages');
    await expect(page.locator('h1')).toContainText('Pages');
  });

  test('creates a new page', async ({ page }) => {
    await page.goto('/admin/pages/new');

    await page.fill('#title', 'E2E Test Page');
    await page.fill('#slug', uniqueSlug);
    await page.fill('#body', '<p>Page created by E2E test.</p>');
    await page.selectOption('#status', 'published');

    // Submit via the form's own button (not the Sign Out button).
    await page.locator('#content-form button[type="submit"]').click();

    // After create, server redirects to pages list.
    await expect(page.locator('text=E2E Test Page')).toBeVisible({ timeout: 10000 });
  });

  test('published page is visible on public site', async ({ page }) => {
    // Visit the public page by slug.
    const response = await page.goto(`/${uniqueSlug}`);

    // Should not be 404.
    expect(response?.status()).not.toBe(404);
    await expect(page.locator('body')).toContainText('E2E Test Page');
  });

  test('deletes the test page', async ({ page }) => {
    await page.goto('/admin/pages');

    const pageRow = page.locator('tr', { hasText: 'E2E Test Page' });
    await expect(pageRow).toBeVisible();

    page.on('dialog', dialog => dialog.accept());
    await pageRow.locator('button:has-text("Delete")').click();

    await page.waitForTimeout(1000);
    await page.reload();

    await expect(page.locator('text=E2E Test Page')).not.toBeVisible();
  });
});
