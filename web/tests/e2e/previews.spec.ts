// Previews (S02.6, FR-014, FR-019): each kind, seeking by byte ranges,
// active content never running, and the keys (bug S02-B10).
import { expect, test, type Page } from '@playwright/test';
import { cell, openFolder, watchErrors } from './helpers';

async function preview(page: Page, name: string) {
  await openFolder(page, '/media');
  await cell(page, name).dblclick();
  const frame = page.getByRole('dialog', { name: `Preview of ${name}` });
  await expect(frame).toBeVisible();
  return frame;
}

test('shows an image, and steps between files with the arrows', async ({ page }) => {
  const errors = watchErrors(page);
  const frame = await preview(page, 'photo.png');
  await expect(frame.getByRole('img', { name: 'photo.png' })).toBeVisible();
  await expect(frame.getByText('PNG file ·')).toBeVisible();
  await page.keyboard.press('ArrowRight'); // the next file in name order
  await expect(page.getByRole('dialog', { name: 'Preview of setup.exe' })).toBeVisible();
  await expect(page.getByTestId('preview-fallback')).toContainText(
    'There is no preview for this type of file.'
  );
  await page.keyboard.press('Escape');
  await expect(page.getByTestId('preview')).toBeHidden();
  await expect(page).toHaveURL(/\/files\/media$/);
  expect(errors).toEqual([]);
});

test('shows the start of a large text file, and says so', async ({ page }) => {
  await preview(page, 'big.log');
  await expect(page.getByTestId('preview-notice')).toHaveText(
    /Showing the first 256 KB of 352 KB\./
  );
  await expect(page.getByTestId('preview-text')).toContainText('log line');
});

test('shows a PDF a page at a time', async ({ page }) => {
  await preview(page, 'doc.pdf');
  const canvas = page.getByTestId('preview-pdf');
  await expect(canvas).toHaveAttribute('aria-label', 'Page 1 of doc.pdf');
  await expect(page.getByText('Page 1 of 3')).toBeVisible();
  await page.keyboard.press('PageDown');
  await expect(canvas).toHaveAttribute('aria-label', 'Page 2 of doc.pdf');
  await page.keyboard.press('End');
  await expect(canvas).toHaveAttribute('aria-label', 'Page 3 of doc.pdf');
});

for (const [name, testId] of [
  ['clip.webm', 'preview-video'],
  ['tone.webm', 'preview-video']
] as const) {
  test(`plays ${name} and seeks by byte ranges (206)`, async ({ page }) => {
    const ranges: number[] = [];
    page.on('response', (r) => {
      if (
        r.url().includes(encodeURIComponent(`/media/${name}`)) &&
        r.request().headers()['range']
      ) {
        ranges.push(r.status());
      }
    });
    await preview(page, name);
    const media = page.getByTestId(testId).or(page.getByTestId('preview-audio'));
    await expect(media).toBeVisible();
    await expect
      .poll(() => media.evaluate((m: HTMLMediaElement) => m.readyState), { timeout: 15_000 })
      .toBeGreaterThanOrEqual(1);
    await media.evaluate((m: HTMLMediaElement) => {
      m.muted = true;
      m.currentTime = 9;
      return m.play().catch(() => {});
    });
    await expect
      .poll(() => media.evaluate((m: HTMLMediaElement) => m.currentTime), { timeout: 15_000 })
      .toBeGreaterThan(8.5);
    expect(ranges.length).toBeGreaterThan(0);
    expect(ranges.every((s) => s === 206)).toBe(true);
  });
}

test('never runs active content: HTML is shown as text, an SVG as an image', async ({ page }) => {
  await preview(page, 'page.html');
  await expect(page.getByTestId('preview-text')).toContainText('<h1 id="injected">Injected</h1>');
  expect(await page.locator('#injected').count()).toBe(0);
  expect(await page.evaluate(() => (window as unknown as { ran?: boolean }).ran)).toBeUndefined();
  await page.keyboard.press('Escape');
  // Closing goes back in history; navigating before that lands made
  // Chromium abort the next page.goto (net::ERR_ABORTED, seen on Linux).
  await expect(page.getByTestId('preview')).toBeHidden();

  const title = await page.title();
  await preview(page, 'active.svg');
  await expect(page.getByRole('img', { name: 'active.svg' })).toBeVisible();
  await page.waitForTimeout(300);
  expect(await page.title()).toBe(title);
  expect(await page.title()).not.toBe('svg ran');
});

test('the keys still work when the focused button turns disabled (B10)', async ({ page }) => {
  await preview(page, 'tone.webm'); // the last file in the folder
  const next = page.getByRole('button', { name: 'Next file' });
  await expect(next).toBeDisabled();
  await page.getByRole('button', { name: 'Previous file' }).click(); // to the file before
  await expect(page.getByRole('dialog', { name: 'Preview of setup.exe' })).toBeVisible();
  await page.keyboard.press('ArrowRight');
  await expect(page.getByRole('dialog', { name: 'Preview of tone.webm' })).toBeVisible();
  await expect(next).toBeDisabled(); // the button that had been used is now disabled
  await page.keyboard.press('ArrowLeft');
  await expect(page.getByRole('dialog', { name: 'Preview of setup.exe' })).toBeVisible();
  await page.keyboard.press('Escape');
  await expect(page.getByTestId('preview')).toBeHidden();
});

test('the download address never renders a file in the app', async ({ page, request }) => {
  const res = await request.get('/api/v1/files/content?path=%2Fmedia%2Factive.svg');
  expect(res.headers()['content-disposition']).toMatch(/^attachment/);
  expect(res.headers()['content-security-policy']).toContain('sandbox');
  expect(res.headers()['x-content-type-options']).toBe('nosniff');
  await page.goto('/files');
  expect(await page.title()).not.toBe('svg ran');
});
