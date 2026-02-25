import { test, expect } from '@playwright/test';

test.describe('Public Site', () => {
  // Public pages don't need authentication — use a fresh context.
  test.use({ storageState: { cookies: [], origins: [] } });

  test('homepage loads successfully', async ({ page }) => {
    await page.goto('/');
    expect(await page.title()).toContain('SmartPress');
    await expect(page.locator('body')).toBeVisible();
  });

  test('health endpoint returns ok', async ({ request }) => {
    const response = await request.get('/health');
    expect(response.status()).toBe(200);
    const body = await response.json();
    expect(body.status).toBe('ok');
  });

  test('nonexistent page returns 404', async ({ page }) => {
    const response = await page.goto('/nonexistent-slug-xyz-12345');
    expect(response?.status()).toBe(404);
  });

  test('admin dashboard redirects to login when not authenticated', async ({ page }) => {
    await page.goto('/admin/dashboard');
    // Should redirect to login.
    await expect(page).toHaveURL(/\/admin\/login/);
  });
});
