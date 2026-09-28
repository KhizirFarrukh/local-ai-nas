// Phones and tablets (S02.7-T01, NFR-015): layouts at 360, 768, and 1024 px,
// 44 px touch targets, the bottom bar, and toolbars that stay on screen
// (bug S02-B10: the selection bar ran off a phone's screen).
import { expect, test, type Page } from '@playwright/test';
import { cell, folderWith, openFolder } from './helpers';

const noSideScroll = (page: Page) =>
  page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth);

for (const width of [360, 768, 1024]) {
  test.describe(`${width} px wide`, () => {
    test.use({ viewport: { width, height: width < 700 ? 740 : 900 }, hasTouch: width < 700 });

    test('the main flows fit the screen', async ({ page, request }) => {
      const folder = await folderWith(request, { 'a.txt': 'a', 'b.txt': 'b' });
      await openFolder(page, folder);
      expect(await noSideScroll(page)).toBe(true);
      const nav = page.getByRole('navigation', { name: 'Sections' });
      if (width < 768) {
        // The bottom bar on phones.
        await expect(nav.last()).toBeVisible();
        const link = (await nav.last().getByRole('link', { name: 'Files' }).boundingBox())!;
        expect(link.height).toBeGreaterThanOrEqual(44);
        expect(link.y).toBeGreaterThan(600);
      } else {
        await expect(nav.first()).toBeVisible();
      }
      // Selecting shows the selection bar, whose buttons stay on screen.
      await cell(page, 'a.txt').click({ modifiers: width < 700 ? [] : ['Control'] });
      if (width < 700) {
        await cell(page, 'a.txt').click(); // a tap opens a file; the selection starts with a long press
        await page.keyboard.press('Escape');
      }
      await page.getByTestId('file-view').focus();
      await page.keyboard.press('Control+a');
      const bar = page.getByTestId('selection-bar');
      await expect(bar).toBeVisible();
      for (const button of await bar.getByRole('button').all()) {
        if (await button.isVisible()) {
          const b = (await button.boundingBox())!;
          expect(b.x + b.width).toBeLessThanOrEqual(width + 0.5);
          if (width < 700) expect(b.height).toBeGreaterThanOrEqual(44);
        }
      }
      expect(await noSideScroll(page)).toBe(true);
    });
  });
}

test.describe('on a phone', () => {
  test.use({ viewport: { width: 360, height: 740 }, hasTouch: true });

  test('dialogs take the whole screen when they hold a list', async ({ page, request }) => {
    const folder = await folderWith(request, { 'a.txt': 'a' });
    await openFolder(page, folder);
    await page.getByTestId('file-view').focus();
    await page.keyboard.press('Home');
    await page.keyboard.press('Shift+F10');
    await page.getByRole('menuitem', { name: 'Move to…' }).click();
    const dialog = page.getByRole('dialog');
    const box = (await dialog.boundingBox())!;
    expect(box.width).toBeGreaterThanOrEqual(359);
    expect(box.height).toBeGreaterThanOrEqual(700);
  });
});
