// Sizes, dates, and finished uploads (S02.3-T05, S02.4-T05; FR-214–FR-216,
// the user's walkthrough in S007): folder sizes, the added and modified
// dates, sorting by the added date, an uploaded folder appearing without a
// reload (bug S02-B12), and a single upload scrolled into view with two
// blinks, also in a folder larger than one page.
import { mkdirSync, mkdtempSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { expect, test, type APIRequestContext, type Page } from '@playwright/test';
import { cell, folderWith, mkdir, openFolder } from './helpers';

const year = String(new Date().getFullYear());

/** The item names in the order the view shows them. */
async function shown(page: Page): Promise<string[]> {
  const labels = await page
    .getByRole('gridcell')
    .evaluateAll((cells) => cells.map((c) => c.getAttribute('aria-label') ?? ''));
  return labels.map((l) => l.split(',')[0]);
}

async function put(request: APIRequestContext, path: string, body: string) {
  const res = await request.put(`/api/v1/files/content?path=${encodeURIComponent(path)}`, {
    data: Buffer.from(body)
  });
  expect(res.status()).toBe(201);
}

test('a folder shows its size, and every item its added and modified dates', async ({
  page,
  request
}) => {
  const folder = await folderWith(request, {
    'sub/a.txt': 'abc',
    'sub/deeper/b.txt': '1234',
    'c.txt': 'x'
  });
  await openFolder(page, folder);
  await expect(page.getByRole('button', { name: /^Sort by added/ })).toBeVisible();
  await expect(page.getByRole('button', { name: /^Sort by modified/ })).toBeVisible();
  await expect(cell(page, 'sub').getByTestId('item-size')).toHaveText('7 bytes');
  await expect(cell(page, 'sub')).toHaveAttribute(
    'aria-label',
    /^sub, Folder, 7 bytes, added .+, modified /
  );
  await expect(cell(page, 'c.txt').getByTestId('item-added')).toContainText(year);
  // The server's answer for the folder, as the view asked for it.
  const usage = await (
    await request.get(`/api/v1/files/usage?path=${encodeURIComponent(`${folder}/sub`)}`)
  ).json();
  expect(usage).toMatchObject({ size: 7, files: 2, folders: 1 });
});

test.describe('on a phone', () => {
  test.use({ viewport: { width: 360, height: 740 }, hasTouch: true });

  test('the size and the added date are under the name', async ({ page, request }) => {
    const folder = await folderWith(request, { 'c.txt': 'x' });
    await openFolder(page, folder);
    await expect(cell(page, 'c.txt').getByTestId('item-details')).toHaveText(
      new RegExp(`^1 byte · Added .*${year}`)
    );
    await expect(page.getByRole('button', { name: /^Sort by size/ })).toHaveCount(0);
  });
});

test('sorts by the date items were added', async ({ page, request }) => {
  const folder = await folderWith(request);
  for (const name of ['c.txt', 'a.txt', 'b.txt']) {
    await put(request, `${folder}/${name}`, name);
    await page.waitForTimeout(50); // distinct times on coarse clocks
  }
  await openFolder(page, folder);
  await expect.poll(() => shown(page)).toEqual(['a.txt', 'b.txt', 'c.txt']);
  await page.getByRole('button', { name: /^Sort by added/ }).click();
  await expect.poll(() => shown(page)).toEqual(['c.txt', 'a.txt', 'b.txt']);
  await page.getByRole('button', { name: /^Sort by added/ }).click();
  await expect.poll(() => shown(page)).toEqual(['b.txt', 'a.txt', 'c.txt']);
});

test('S02-B12: an uploaded folder appears without a reload, and blinks twice', async ({
  page,
  request,
  browserName
}) => {
  test.skip(
    browserName !== 'chromium',
    'Playwright sets folder inputs in Chromium-based browsers only'
  );
  const folder = await folderWith(request, { 'existing.txt': 'e' });
  const dir = mkdtempSync(join(tmpdir(), 'lan-b12-'));
  mkdirSync(join(dir, 'trip', 'day 2'), { recursive: true });
  writeFileSync(join(dir, 'trip', 'day 2', 'b.txt'), 'b');
  await openFolder(page, folder);
  await page.getByTestId('folder-input').setInputFiles(join(dir, 'trip'));
  const trip = cell(page, 'trip');
  await expect(trip).toBeVisible(); // no reload
  await expect(trip).toHaveClass(/\bflash\b/);
  await expect(trip).not.toHaveClass(/\bflash\b/, { timeout: 5_000 });
  await expect(trip.getByTestId('item-size')).toHaveText('1 byte');
});

test('a single upload into a large folder is scrolled into view and blinks twice', async ({
  page,
  request
}) => {
  const folder = await folderWith(request);
  // More than one page (500), so the new item's page is not loaded yet.
  const names = Array.from({ length: 600 }, (_, i) => `file-${String(i).padStart(3, '0')}.txt`);
  for (let i = 0; i < names.length; i += 50) {
    await Promise.all(names.slice(i, i + 50).map((n) => put(request, `${folder}/${n}`, 'x')));
  }
  await openFolder(page, folder);
  const dir = mkdtempSync(join(tmpdir(), 'lan-reveal-'));
  const file = join(dir, 'zzz-new.txt');
  writeFileSync(file, 'new');
  await page.getByTestId('file-input').setInputFiles(file);
  const added = cell(page, 'zzz-new.txt');
  await expect(added).toBeInViewport({ timeout: 15_000 });
  await expect(added).toHaveClass(/\bflash\b/);
  const style = await added.evaluate((el) => {
    const s = getComputedStyle(el);
    return { name: s.animationName, count: s.animationIterationCount };
  });
  expect(style).toEqual({ name: 'flash', count: '2' });
  await expect(added).not.toHaveClass(/\bflash\b/, { timeout: 5_000 });
});

test('an upload of several items only refreshes the list', async ({ page, request }) => {
  const folder = await folderWith(request);
  await mkdir(request, `${folder}/keep`);
  const dir = mkdtempSync(join(tmpdir(), 'lan-two-'));
  writeFileSync(join(dir, 'one.txt'), '1');
  writeFileSync(join(dir, 'two.txt'), '2');
  await openFolder(page, folder);
  await page.getByTestId('file-input').setInputFiles([join(dir, 'one.txt'), join(dir, 'two.txt')]);
  await expect(cell(page, 'one.txt')).toBeVisible();
  await expect(cell(page, 'two.txt')).toBeVisible();
  await page.waitForTimeout(300);
  await expect(page.locator('.flash')).toHaveCount(0);
});
