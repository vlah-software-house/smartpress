import { test, expect } from '@playwright/test';

test.describe('Posts CRUD', () => {
  const uniqueSlug = `e2e-post-${Date.now()}`;

  test('lists existing posts', async ({ page }) => {
    await page.goto('/admin/posts');
    await expect(page.locator('h1')).toContainText('Posts');

    // The seeded "Hello World" post should appear.
    await expect(page.locator('text=Hello World')).toBeVisible();
  });

  test('creates a new post', async ({ page }) => {
    await page.goto('/admin/posts/new');

    // Fill in the form — field IDs: title, slug, body, status.
    await page.fill('#title', 'E2E Test Post');
    await page.fill('#slug', uniqueSlug);
    await page.fill('#body', '<p>This post was created by Playwright E2E tests.</p>');
    await page.selectOption('#status', 'draft');

    // Submit via the form's own button (not the Sign Out button).
    await page.locator('#content-form button[type="submit"]').click();

    // After create, server redirects to posts list (full navigation).
    await expect(page.locator('text=E2E Test Post')).toBeVisible({ timeout: 10000 });
  });

  test('edits an existing post', async ({ page }) => {
    await page.goto('/admin/posts');

    // Click on the E2E test post to edit (link in the table row).
    await page.locator('a:has-text("E2E Test Post")').click();

    // Verify the form is pre-filled.
    await expect(page.locator('#title')).toHaveValue('E2E Test Post');

    // Update the title.
    await page.fill('#title', 'E2E Test Post (Updated)');
    await page.selectOption('#status', 'published');

    // Submit the update (HTMX PUT swaps content).
    await page.locator('#content-form button[type="submit"]').click();

    // Verify the updated title appears in the list.
    await expect(page.locator('text=E2E Test Post (Updated)')).toBeVisible({ timeout: 10000 });
  });

  test('validates required fields', async ({ page }) => {
    await page.goto('/admin/posts/new');

    // Leave title empty, add body, and submit.
    await page.fill('#title', '');
    await page.fill('#body', 'Body without title');

    // The HTML5 required attribute on #title should prevent submission.
    // Verify we stay on the new post form.
    await page.locator('#content-form button[type="submit"]').click();
    await expect(page.locator('#title')).toBeVisible();
    await expect(page).toHaveURL(/\/admin\/posts\/new/);
  });

  test('deletes a post', async ({ page }) => {
    await page.goto('/admin/posts');

    // Find and delete the E2E test post.
    const postRow = page.locator('tr', { hasText: 'E2E Test Post' });
    await expect(postRow).toBeVisible();

    // HTMX delete uses hx-confirm — handle the confirmation dialog.
    page.on('dialog', dialog => dialog.accept());
    await postRow.locator('button:has-text("Delete")').click();

    // Wait for the HTMX content swap and verify removal.
    await page.waitForTimeout(1000);
    await page.reload();
    await expect(page.locator('text=E2E Test Post')).not.toBeVisible();
  });
});
