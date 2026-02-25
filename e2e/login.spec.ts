import { test, expect } from '@playwright/test';

test.describe('Login Flow', () => {
  // Login tests use a fresh browser context (no stored auth).
  test.use({ storageState: { cookies: [], origins: [] } });

  test('shows error for invalid credentials', async ({ page }) => {
    await page.goto('/admin/login');

    await page.fill('#email', 'wrong@example.com');
    await page.fill('#password', 'wrongpassword');
    await page.click('button[type="submit"]');

    // Should show error and stay on login page.
    await expect(page.locator('text=Invalid')).toBeVisible();
    await expect(page).toHaveURL(/\/admin\/login/);
  });

  test('shows error for empty fields', async ({ page }) => {
    await page.goto('/admin/login');

    // Submit with empty fields — HTML5 validation should prevent submission.
    await page.fill('#email', '');
    await page.click('button[type="submit"]');

    // Should stay on login page.
    await expect(page).toHaveURL(/\/admin\/login/);
  });

  test('login page has proper structure', async ({ page }) => {
    await page.goto('/admin/login');

    // Check page elements — h1 contains "SmartPress".
    await expect(page.locator('h1')).toContainText('SmartPress');
    await expect(page.locator('#email')).toBeVisible();
    await expect(page.locator('#password')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();

    // CSRF token should be present.
    await expect(page.locator('input[name="csrf_token"]')).toBeAttached();
  });
});
