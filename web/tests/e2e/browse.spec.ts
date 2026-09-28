// Browsing (S02.3, FR-002): folders, the 50,000-item folder, views, sorting.
import { expect, test } from '@playwright/test';
import { largeFolderSize } from './setup';
import { cell, folderWith, openFolder, watchErrors } from './helpers';

test('opens folders, goes up, and shows breadcrumbs (B03: no page error)', async ({
  page,
  request
}) => {
  const errors = watchErrors(page);
  const folder = await folderWith(request, { 'sub/inner.txt': 'x', 'a.txt': 'a' });
  await openFolder(page, folder);
  await expect(page.getByRole('heading', { level: 1 })).toHaveText(folder.slice(1));
  await cell(page, 'sub').dblclick();
  await expect(page).toHaveURL(/\/sub$/);
  await expect(cell(page, 'inner.txt')).toBeVisible();
  await page
    .getByRole('navigation', { name: 'Folder' })
    .getByRole('link', { name: folder.slice(1) })
    .click();
  await expect(cell(page, 'a.txt')).toBeVisible();
  expect(errors).toEqual([]);
});

test('the 50,000-item folder shows its first page fast and scrolls to the end', async ({
  page
}) => {
  const errors = watchErrors(page);
  const start = Date.now();
  await page.goto('/files/large');
  await expect(cell(page, 'f00000.txt')).toBeVisible();
  const firstPage = Date.now() - start;
  await expect(page.getByTestId('file-view')).toHaveAttribute(
    'aria-rowcount',
    String(largeFolderSize)
  );
  await page.getByTestId('file-view').focus();
  const scrolled = Date.now();
  await page.keyboard.press('End');
  await expect(cell(page, 'f49999.txt')).toBeVisible();
  const toEnd = Date.now() - scrolled;
  test.info().annotations.push({
    type: 'timing',
    description: `first page ${firstPage} ms; to the end ${toEnd} ms`
  });
  expect(firstPage).toBeLessThan(5000);
  expect(toEnd).toBeLessThan(5000);
  // Only the rows on screen exist.
  expect(await page.getByRole('row').count()).toBeLessThan(80);
  expect(errors).toEqual([]);
});

test('switches between list and grid, and sorts', async ({ page, request }) => {
  const folder = await folderWith(request, { 'small.txt': 'x', 'large.txt': 'x'.repeat(5000) });
  await openFolder(page, folder);
  await page.getByRole('button', { name: 'Grid view' }).click();
  await expect(page.getByTestId('file-view')).not.toHaveAttribute('aria-colcount', '1');
  await page.getByRole('button', { name: 'List view' }).click();
  await expect(page.getByTestId('file-view')).toHaveAttribute('aria-colcount', '1');
  await page.getByRole('button', { name: 'Sort by size' }).click();
  await expect(page.getByRole('gridcell').first()).toHaveAccessibleName(/^small\.txt,/);
  await page.getByRole('button', { name: /Sort by size, now ascending/ }).click();
  await expect(page.getByRole('gridcell').first()).toHaveAccessibleName(/^large\.txt,/);
});

test('an empty folder and a missing one explain themselves', async ({ page, request }) => {
  const folder = await folderWith(request);
  await openFolder(page, folder);
  await expect(page.getByText('This folder is empty')).toBeVisible();
  await page.goto('/files/no-such-folder');
  await expect(page.getByText('It may have been moved, renamed, or deleted.')).toBeVisible();
  await expect(page.getByRole('button', { name: 'Try again' })).toBeVisible();
});
