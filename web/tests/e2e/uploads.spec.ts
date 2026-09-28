// Uploads (S02.4, FR-003, FR-080): files and folders, pause and resume, a
// dropped network, a reload, drag and drop, and a taken name. The server
// runs with a 1 MiB chunk limit (playwright.config.ts), and the tests slow
// each chunk down, so there is time to act during an upload.
import { createHash } from 'node:crypto';
import { mkdirSync, mkdtempSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { expect, test, type Page } from '@playwright/test';
import { content, folderWith, names, openFolder, note } from './helpers';

/** A file on disk of n bytes, with its SHA-256. */
function bigFile(n: number, name = 'big.bin') {
  const dir = mkdtempSync(join(tmpdir(), 'lan-up-'));
  const data = Buffer.alloc(n);
  for (let i = 0; i < n; i += 4096) data.writeUInt32LE(i, i); // not all zeros
  const path = join(dir, name);
  writeFileSync(path, data);
  return { path, sha: 'sha256:' + createHash('sha256').update(data).digest('hex') };
}

/** Slows every chunk (PATCH) down, and counts the offsets they start at. */
async function slowChunks(page: Page, ms: number) {
  const offsets: number[] = [];
  await page.route('**/api/v1/files/uploads/**', async (route) => {
    const req = route.request();
    if (req.method() === 'PATCH') {
      offsets.push(Number(req.headers()['upload-offset']));
      await new Promise((r) => setTimeout(r, ms));
    }
    await route.continue();
  });
  return offsets;
}

async function details(page: Page, path: string) {
  const res = await page.request.get(`/api/v1/files/items?path=${encodeURIComponent(path)}`);
  return (await res.json()).item as { size: number; content_hash?: string };
}

test('choosing files queues and uploads them, with their hashes (B06)', async ({
  page,
  request
}) => {
  const folder = await folderWith(request);
  await openFolder(page, folder);
  const chooser = page.waitForEvent('filechooser');
  await page.getByRole('button', { name: 'Upload', exact: true }).click();
  await (
    await chooser
  ).setFiles([
    { name: 'one.txt', mimeType: 'text/plain', buffer: Buffer.from('first') },
    { name: 'two.txt', mimeType: 'text/plain', buffer: Buffer.from('second') }
  ]);
  await expect(page.getByTestId('upload')).toHaveCount(2);
  await expect(note(page, 'Uploaded 2 files.')).toBeVisible();
  expect(await names(request, folder)).toEqual(['one.txt', 'two.txt']);
  expect(await content(request, `${folder}/two.txt`)).toBe('second');
  const hash = 'sha256:' + createHash('sha256').update('second').digest('hex');
  expect((await details(page, `${folder}/two.txt`)).content_hash).toBe(hash);
});

test('a large upload pauses and resumes from where it stopped', async ({ page, request }) => {
  const folder = await folderWith(request);
  const file = bigFile(12 << 20);
  await openFolder(page, folder);
  const offsets = await slowChunks(page, 150);
  await page.getByTestId('file-input').setInputFiles(file.path);
  const pause = page.getByRole('button', { name: 'Pause big.bin' });
  await expect(pause).toBeVisible();
  await expect.poll(() => offsets.length).toBeGreaterThan(2);
  await pause.click();
  await expect(page.getByTestId('upload-status')).toContainText('Paused at');
  const pausedAt = offsets.length;
  await page.waitForTimeout(800);
  expect(offsets.length).toBeLessThanOrEqual(pausedAt + 1); // nothing more is sent while paused
  await page.getByRole('button', { name: 'Resume big.bin' }).click();
  await expect(page.getByTestId('upload-status')).toHaveText('Done', { timeout: 30_000 });
  const d = await details(page, `${folder}/big.bin`);
  expect(d.size).toBe(12 << 20);
  expect(d.content_hash).toBe(file.sha);
  // The chunks never started over.
  expect([...offsets].sort((a, b) => a - b)).toEqual(offsets);
});

test('an upload continues by itself after the network comes back', async ({
  page,
  request,
  context
}) => {
  const folder = await folderWith(request);
  const file = bigFile(8 << 20, 'net.bin');
  await openFolder(page, folder);
  const offsets = await slowChunks(page, 150);
  await page.getByTestId('file-input').setInputFiles(file.path);
  await expect.poll(() => offsets.length).toBeGreaterThan(2);
  await context.setOffline(true);
  await page.waitForTimeout(1500);
  await context.setOffline(false);
  await expect(page.getByTestId('upload-status')).toHaveText('Done', { timeout: 60_000 });
  expect((await details(page, `${folder}/net.bin`)).content_hash).toBe(file.sha);
});

test('after a reload, adding the same file again continues the upload', async ({
  page,
  request
}) => {
  const folder = await folderWith(request);
  const file = bigFile(10 << 20, 'again.bin');
  await openFolder(page, folder);
  let offsets = await slowChunks(page, 150);
  await page.getByTestId('file-input').setInputFiles(file.path);
  await expect.poll(() => offsets.length).toBeGreaterThan(3);
  await page.getByRole('button', { name: 'Pause again.bin' }).click();
  await expect(page.getByTestId('upload-status')).toContainText('Paused at');
  await page.reload();
  await expect(
    page.getByTestId('file-view').or(page.getByText('This folder is empty'))
  ).toBeVisible();
  await page.unrouteAll({ behavior: 'wait' });
  offsets = await slowChunks(page, 0);
  await page.getByTestId('file-input').setInputFiles(file.path);
  await expect(page.getByTestId('upload-status')).toHaveText('Done', { timeout: 30_000 });
  expect(offsets[0]).toBeGreaterThan(0); // continued, not started over
  expect((await details(page, `${folder}/again.bin`)).content_hash).toBe(file.sha);
});

test('uploads a folder with its subfolders', async ({ page, request, browserName }) => {
  test.skip(
    browserName !== 'chromium',
    'Playwright sets folder inputs in Chromium-based browsers only'
  );
  const folder = await folderWith(request);
  const dir = mkdtempSync(join(tmpdir(), 'lan-dir-'));
  mkdirSync(join(dir, 'trip', 'day 2'), { recursive: true });
  writeFileSync(join(dir, 'trip', 'a.txt'), 'a');
  writeFileSync(join(dir, 'trip', 'day 2', 'b.txt'), 'b');
  await openFolder(page, folder);
  await page.getByTestId('folder-input').setInputFiles(join(dir, 'trip'));
  await expect(note(page, 'Uploaded 2 files.')).toBeVisible();
  expect(await names(request, `${folder}/trip`)).toEqual(['a.txt', 'day 2']);
  expect(await content(request, `${folder}/trip/day 2/b.txt`)).toBe('b');
});

test('files dropped on the page upload into the folder on screen', async ({ page, request }) => {
  const folder = await folderWith(request);
  await openFolder(page, folder);
  await page.evaluate(() => {
    const dt = new DataTransfer();
    dt.items.add(new File(['dropped'], 'dropped.txt', { type: 'text/plain' }));
    for (const type of ['dragenter', 'dragover', 'drop']) {
      window.dispatchEvent(
        new DragEvent(type, { dataTransfer: dt, bubbles: true, cancelable: true })
      );
    }
  });
  await expect(note(page, 'Uploaded 1 file.')).toBeVisible();
  expect(await content(request, `${folder}/dropped.txt`)).toBe('dropped');
});

test('a taken name asks, and "Keep both" gives the upload a numbered name', async ({
  page,
  request
}) => {
  const folder = await folderWith(request, { 'report.txt': 'old' });
  await openFolder(page, folder);
  await page
    .getByTestId('file-input')
    .setInputFiles({ name: 'report.txt', mimeType: 'text/plain', buffer: Buffer.from('new') });
  const dialog = page.getByRole('dialog', { name: '“report.txt” is already there' });
  await expect(dialog).toBeVisible();
  await dialog.getByRole('button', { name: /^Keep both/ }).click();
  await expect(page.getByTestId('upload-status')).toContainText('Done, saved as');
  expect(await names(request, folder)).toEqual(['report (1).txt', 'report.txt']);
  expect(await content(request, `${folder}/report.txt`)).toBe('old');
});
