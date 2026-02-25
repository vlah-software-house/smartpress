import { test, expect } from '@playwright/test';

test.describe('Dashboard', () => {
  test('shows dashboard with stats and quick actions', async ({ page }) => {
    await page.goto('/admin/dashboard');

    // Page heading in the top bar.
    await expect(page.locator('h1')).toContainText('Dashboard');

    // Stats cards in the main content area.
    const main = page.locator('main');
    await expect(main.locator('text=Total Posts')).toBeVisible();
    await expect(main.locator('text=Total Pages')).toBeVisible();
    await expect(main.locator('text=Users')).toBeVisible();

    // Quick action links.
    await expect(main.locator('a:has-text("New Post")')).toBeVisible();
    await expect(main.locator('a:has-text("New Page")')).toBeVisible();
  });

  test('sidebar navigation links work', async ({ page }) => {
    // Navigate via full page loads (sidebar uses HTMX partial swap).
    await page.goto('/admin/posts');
    await expect(page.locator('h1')).toContainText('Posts');

    await page.goto('/admin/pages');
    await expect(page.locator('h1')).toContainText('Pages');

    await page.goto('/admin/dashboard');
    await expect(page.locator('text=Welcome back')).toBeVisible();
  });
});
