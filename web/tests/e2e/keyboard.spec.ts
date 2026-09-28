// Keyboard-only use (S02.5-T04, S02.7-T02, NFR-015): every flow without a
// mouse, the focus after each step, and the focus fixes of bug S02-B10.
import { expect, test, type Page } from '@playwright/test';
import { cell, folderWith, names, openFolder } from './helpers';

const focused = (page: Page) =>
  page.evaluate(() => {
    const el = document.activeElement;
    return `${el?.getAttribute('role') ?? el?.tagName.toLowerCase()}:${el?.getAttribute('aria-label') ?? ''}`;
  });

async function tabToGrid(page: Page) {
  await page.locator('body').click({ position: { x: 1, y: 1 } });
  for (let i = 0; i < 30; i++) {
    await page.keyboard.press('Tab');
    if ((await focused(page)).startsWith('grid:')) return;
  }
  throw new Error('the file view is not reachable with Tab');
}

test('reaches the file view with Tab, with a visible focus ring', async ({ page, request }) => {
  const folder = await folderWith(request, { 'a.txt': 'a' });
  await openFolder(page, folder);
  await page.keyboard.press('Tab');
  await expect(page.getByRole('link', { name: 'Skip to content' })).toBeFocused();
  await tabToGrid(page);
  expect(await page.getByTestId('file-view').evaluate((el) => el.matches(':focus-visible'))).toBe(
    true
  );
});

test('opens folders, goes back up, and the view keeps the focus (B10)', async ({
  page,
  request
}) => {
  const folder = await folderWith(request, { 'sub/inner.txt': 'i', 'z.txt': 'z' });
  await openFolder(page, folder);
  await tabToGrid(page);
  await page.keyboard.press('Home');
  await page.keyboard.press('Enter'); // "sub" sorts first
  await expect(page).toHaveURL(/\/sub$/);
  await expect(cell(page, 'inner.txt')).toBeVisible();
  expect(await focused(page)).toMatch(/^grid:Items in sub/);
  await page.keyboard.press('Backspace');
  await expect(cell(page, 'z.txt')).toBeVisible();
  expect(await focused(page)).toMatch(/^grid:/);
});

test('creates, renames, and deletes with shortcuts only', async ({ page, request }) => {
  const folder = await folderWith(request, { 'old.txt': 'o' });
  await openFolder(page, folder);
  await tabToGrid(page);
  await page.keyboard.press('Shift+N');
  await expect(page.getByRole('dialog', { name: 'New folder' })).toBeVisible();
  await expect // the default name is selected a moment after the dialog opens
    .poll(() => page.evaluate(() => (document.activeElement as HTMLInputElement).selectionEnd))
    .toBe('New folder'.length);
  await page.keyboard.type('made by keys');
  await page.keyboard.press('Enter');
  await expect(cell(page, 'made by keys')).toBeVisible();
  expect(await focused(page)).toMatch(/^grid:/);

  await page.keyboard.press('End'); // old.txt
  await page.keyboard.press('F2');
  await expect(page.getByRole('dialog', { name: 'Rename' })).toBeVisible();
  await expect
    .poll(() => page.evaluate(() => (document.activeElement as HTMLInputElement).selectionEnd))
    .toBe(3);
  await page.keyboard.type('new');
  await page.keyboard.press('Enter');
  await expect(cell(page, 'new.txt')).toBeVisible();
  // B11: the renamed item stays selected and focused, so Delete acts on it.
  await expect(cell(page, 'new.txt')).toHaveAttribute('aria-selected', 'true');

  await page.keyboard.press('Delete');
  const dialog = page.getByRole('dialog', { name: 'Delete “new.txt”?' });
  await expect(dialog).toBeVisible();
  await page.keyboard.press('Enter'); // Cancel has the focus: Enter never deletes by accident
  await expect(dialog).toBeHidden();
  expect(await names(request, folder)).toContain('new.txt');
  await page.keyboard.press('Delete');
  await page.keyboard.press('Tab');
  await page.keyboard.press('Enter');
  await expect(cell(page, 'new.txt')).toBeHidden();
  expect(await names(request, folder)).toEqual(['made by keys']);
});

test('opens the menu with Shift+F10 and picks with the arrows', async ({ page, request }) => {
  const folder = await folderWith(request, { 'a.txt': 'a' });
  await openFolder(page, folder);
  await tabToGrid(page);
  await page.keyboard.press('Home');
  await page.keyboard.press('Shift+F10');
  await expect(page.getByRole('menu')).toBeVisible();
  await expect(page.getByRole('menuitem').first()).toBeFocused();
  await page.keyboard.press('Escape');
  await expect(page.getByRole('menu')).toBeHidden();
  expect(await focused(page)).toMatch(/^grid:/);
  await page.keyboard.press('?');
  await expect(page.getByRole('dialog', { name: 'Keyboard shortcuts' })).toBeVisible();
  await page.keyboard.press('Escape');
  expect(await focused(page)).toMatch(/^grid:/);
});

test('a click on the page background keeps the shortcuts working (B10)', async ({
  page,
  request
}) => {
  const folder = await folderWith(request, { 'a.txt': 'a' });
  await openFolder(page, folder);
  const main = page.locator('main');
  const box = (await main.boundingBox())!;
  await page.mouse.click(box.x + box.width - 5, box.y + box.height - 5); // empty space in the main area
  await page.keyboard.press('Shift+N');
  await expect(page.getByRole('dialog', { name: 'New folder' })).toBeVisible();
});
