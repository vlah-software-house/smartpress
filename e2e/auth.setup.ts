// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.
import { test as setup, expect } from '@playwright/test';
import * as crypto from 'crypto';
import * as fs from 'fs';
import { execFileSync } from 'child_process';

const ADMIN_EMAIL = 'admin@yaaicms.local';
const ADMIN_PASSWORD = 'admin';
const AUTH_FILE = 'e2e/.auth/admin.json';

// Database connection for reading the TOTP secret.
const DB_HOST = process.env.DB_HOST || 'localhost';
const DB_PORT = process.env.DB_PORT || '5432';
const DB_USER = process.env.DB_USER || 'yaaicms';
const DB_PASS = process.env.DB_PASSWORD || 'changeme';
const DB_NAME = process.env.DB_NAME || 'yaaicms';

/**
 * Read the admin user's TOTP secret from the database.
 * Uses execFileSync (no shell) to avoid command injection.
 */
function getTOTPSecret(): string {
  const result = execFileSync(
    'psql',
    [
      '-h', DB_HOST,
      '-p', DB_PORT,
      '-U', DB_USER,
      '-d', DB_NAME,
      '-t', '-A',
      '-c', `SELECT totp_secret FROM users WHERE email='${ADMIN_EMAIL}'`,
    ],
    { encoding: 'utf-8', env: { ...process.env, PGPASSWORD: DB_PASS } },
  ).trim();
  if (!result) {
    throw new Error('TOTP secret not found in database for admin user');
  }
  return result;
}

/**
 * Generate a 6-digit TOTP code from a base32-encoded secret.
 * Implements RFC 6238 (TOTP) with HMAC-SHA1, 30-second time step, 6 digits.
 */
function generateTOTP(base32Secret: string): string {
  // Decode base32 to raw bytes.
  const base32chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567';
  let bits = '';
  for (const c of base32Secret.toUpperCase().replace(/[= ]+/g, '')) {
    const val = base32chars.indexOf(c);
    if (val === -1) continue;
    bits += val.toString(2).padStart(5, '0');
  }
  const keyBytes = Buffer.alloc(Math.floor(bits.length / 8));
  for (let i = 0; i < keyBytes.length; i++) {
    keyBytes[i] = parseInt(bits.substring(i * 8, i * 8 + 8), 2);
  }

  // Calculate time counter (30-second step).
  const timeStep = Math.floor(Date.now() / 1000 / 30);
  const timeBuffer = Buffer.alloc(8);
  timeBuffer.writeBigInt64BE(BigInt(timeStep));

  // HMAC-SHA1.
  const hmac = crypto.createHmac('sha1', keyBytes);
  hmac.update(timeBuffer);
  const hash = hmac.digest();

  // Dynamic truncation (RFC 4226 §5.4).
  const offset = hash[hash.length - 1] & 0x0f;
  const code =
    (((hash[offset] & 0x7f) << 24) |
      ((hash[offset + 1] & 0xff) << 16) |
      ((hash[offset + 2] & 0xff) << 8) |
      (hash[offset + 3] & 0xff)) %
    1_000_000;

  return code.toString().padStart(6, '0');
}

setup('authenticate as admin', async ({ page, context }) => {
  // If a valid auth state exists, check if the session is still active.
  // This avoids hitting the rate limiter on repeated runs.
  if (fs.existsSync(AUTH_FILE)) {
    const stored = JSON.parse(fs.readFileSync(AUTH_FILE, 'utf-8'));
    await context.addCookies(stored.cookies || []);

    const response = await page.goto('/admin/dashboard');
    if (response && response.url().includes('/admin/dashboard')) {
      // Session is still valid — reuse it.
      await context.storageState({ path: AUTH_FILE });
      return;
    }
  }

  // Session expired or doesn't exist — perform full login + 2FA flow.
  await page.goto('/admin/login');
  await expect(page.locator('#email')).toBeVisible();

  await page.fill('#email', ADMIN_EMAIL);
  await page.fill('#password', ADMIN_PASSWORD);
  await page.click('button[type="submit"]');

  // After login, we'll land on either:
  // 1. /admin/2fa/setup (first time — no TOTP secret yet)
  // 2. /admin/2fa/verify (returning user with TOTP enabled)
  // 3. /admin/dashboard (shouldn't happen — 2FA is mandatory)
  await page.waitForURL(/\/(admin\/2fa|admin\/dashboard)/);

  const url = page.url();

  if (url.includes('/2fa/setup') || url.includes('/2fa/verify')) {
    // The server has stored the TOTP secret in the DB by this point.
    // Read it directly and generate a valid code.
    const secret = getTOTPSecret();
    const code = generateTOTP(secret);

    await page.fill('#code', code);
    await page.click('button[type="submit"]');
  }

  // Wait for dashboard content to appear (more reliable than waitForURL
  // which depends on the page "load" event fully firing).
  await expect(page.locator('text=Welcome back')).toBeVisible({ timeout: 15000 });

  // Save authentication state for other tests.
  await page.context().storageState({ path: AUTH_FILE });
});
