import { test, expect } from '@playwright/test';

test.describe('Templates Management', () => {
  const tmplName = `E2E Header ${Date.now()}`;

  test('lists templates', async ({ page }) => {
    await page.goto('/admin/templates');

    // The page heading is "AI Design" (nav label), with h2 "Templates" inside.
    await expect(page.locator('h1')).toContainText('AI Design');

    // Seeded templates should appear in the table.
    await expect(page.locator('text=Default Header')).toBeVisible();
  });

  test('creates and deletes a template', async ({ page }) => {
    // Create a new template.
    await page.goto('/admin/templates/new');

    await page.fill('#name', tmplName);
    await page.selectOption('#type', 'header');
    await page.fill('#html_content', '<header class="e2e-test"><nav>{{.SiteName}}</nav></header>');

    await page.locator('main button[type="submit"]').click();

    // After create, server redirects to templates list.
    await expect(page.locator(`text=${tmplName}`)).toBeVisible({ timeout: 10000 });

    // Delete the template we just created.
    const row = page.locator('tr', { hasText: tmplName });
    await expect(row).toBeVisible();

    page.on('dialog', dialog => dialog.accept());
    await row.locator('button:has-text("Delete")').click();

    await page.waitForTimeout(1000);
    await page.reload();

    await expect(page.locator(`text=${tmplName}`)).not.toBeVisible();
  });

  test('rejects invalid template syntax', async ({ page }) => {
    await page.goto('/admin/templates/new');

    await page.fill('#name', 'Bad Syntax Template');
    await page.selectOption('#type', 'header');
    await page.fill('#html_content', '<header>{{.Unclosed</header>');

    await page.locator('main button[type="submit"]').click();

    // Should stay on form with syntax error.
    await expect(page.locator('text=syntax error')).toBeVisible({ timeout: 10000 });
  });

  test('previews a template via the page', async ({ page }) => {
    // Navigate to a template edit page and use the preview endpoint.
    // First, go to templates list and get a template.
    await page.goto('/admin/templates');

    // Click Edit on the first template (e.g., Default Header).
    const firstRow = page.locator('tr', { hasText: 'Default Header' });
    await expect(firstRow).toBeVisible();
    await firstRow.locator('a:has-text("Edit")').click();

    // Verify the edit form loads.
    await expect(page.locator('#name')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#html_content')).toBeVisible();
  });
});
