// File operations (S02.5, FR-007, FR-081): new folder, rename, move, copy,
// delete, the app's clipboard, and every choice of the conflict dialog.
import { expect, test, type Page } from '@playwright/test';
import { cell, content, folderWith, names, openFolder, note } from './helpers';

async function menuOf(page: Page, name: string, entry: string | RegExp) {
  await cell(page, name).click({ button: 'right' });
  await page.getByRole('menu').getByRole('menuitem', { name: entry }).click();
}

async function pick(page: Page, path: string[], submit: string) {
  const dialog = page.getByRole('dialog');
  await dialog.getByRole('button', { name: 'Files', exact: true }).click();
  for (const name of path) {
    await dialog.getByTestId('folder-picker').getByRole('button', { name, exact: true }).click();
  }
  await dialog.getByRole('button', { name: submit }).click();
}

test('makes a new folder, and refuses a name the server does not allow', async ({
  page,
  request
}) => {
  const folder = await folderWith(request);
  await openFolder(page, folder);
  await page.getByRole('button', { name: 'New folder' }).first().click();
  const dialog = page.getByRole('dialog', { name: 'New folder' });
  await dialog.getByRole('textbox', { name: 'Folder name' }).fill('a:b');
  await dialog.getByRole('button', { name: 'Create' }).click();
  await expect(dialog.getByText(/Names cannot contain/)).toBeVisible();
  await dialog.getByRole('textbox', { name: 'Folder name' }).fill('Reports 2026');
  await dialog.getByRole('button', { name: 'Create' }).click();
  await expect(dialog).toBeHidden();
  await expect(cell(page, 'Reports 2026')).toBeVisible();
});

test('renames from the menu, keeping the extension selected out', async ({ page, request }) => {
  const folder = await folderWith(request, { 'draft.txt': 'text' });
  await openFolder(page, folder);
  await menuOf(page, 'draft.txt', 'Rename…');
  // The stem ("draft") is selected a moment after the dialog opens.
  await expect
    .poll(() =>
      page.evaluate(() => {
        const input = document.activeElement as HTMLInputElement;
        return `${input.selectionStart}-${input.selectionEnd}`;
      })
    )
    .toBe('0-5');
  await page.keyboard.type('final');
  await page.keyboard.press('Enter');
  await expect(cell(page, 'final.txt')).toBeVisible();
  expect(await names(request, folder)).toEqual(['final.txt']);
});

test('moves into a folder where the name is taken, and replaces it', async ({ page, request }) => {
  const folder = await folderWith(request, { 'a.txt': 'new', 'to/a.txt': 'old' });
  await openFolder(page, folder);
  await menuOf(page, 'a.txt', 'Move to…');
  await pick(page, [folder.slice(1), 'to'], 'Move here');
  const conflict = page.getByRole('dialog', { name: '“a.txt” is already there' });
  await conflict.getByRole('button', { name: /^Replace/ }).click();
  await expect(note(page, 'Moved “a.txt”.')).toBeVisible();
  expect(await names(request, folder)).toEqual(['to']);
  expect(await content(request, `${folder}/to/a.txt`)).toBe('new');
});

test('copies with "Keep both", and skips the rest with "apply to all"', async ({
  page,
  request
}) => {
  const folder = await folderWith(request, {
    'a.txt': 'A',
    'b.txt': 'B',
    'c.txt': 'C',
    'to/a.txt': 'x',
    'to/b.txt': 'x',
    'to/c.txt': 'x'
  });
  await openFolder(page, folder);
  await menuOf(page, 'a.txt', 'Copy to…');
  await pick(page, [folder.slice(1), 'to'], 'Copy here');
  await page
    .getByRole('dialog')
    .getByRole('button', { name: /^Keep both/ })
    .click();
  await expect(note(page, 'Copied “a.txt”.')).toBeVisible();

  await cell(page, 'b.txt').click();
  await cell(page, 'c.txt').click({ modifiers: ['Shift'] });
  await page
    .getByRole('toolbar', { name: 'Selected items' })
    .getByRole('button', { name: 'Copy' })
    .click();
  await pick(page, [folder.slice(1), 'to'], 'Copy here');
  const conflict = page.getByRole('dialog', { name: '“b.txt” is already there' });
  await conflict.getByRole('checkbox', { name: /Do the same for the other files/ }).check();
  await conflict.getByRole('button', { name: /^Skip/ }).click();
  await expect(note(page, /Skipped 2 items: nothing was copied\./)).toBeVisible();
  expect(await names(request, `${folder}/to`)).toEqual(['a (1).txt', 'a.txt', 'b.txt', 'c.txt']);
  expect(await content(request, `${folder}/to/a (1).txt`)).toBe('A');
});

test('stops a bulk run from the conflict dialog', async ({ page, request }) => {
  const folder = await folderWith(request, { 'a.txt': '1', 'b.txt': '2', 'to/a.txt': 'x' });
  await openFolder(page, folder);
  await cell(page, 'a.txt').click();
  await cell(page, 'b.txt').click({ modifiers: ['Shift'] });
  await page
    .getByRole('toolbar', { name: 'Selected items' })
    .getByRole('button', { name: 'Move' })
    .click();
  await pick(page, [folder.slice(1), 'to'], 'Move here');
  await page.getByRole('dialog').getByRole('button', { name: 'Stop the operation' }).click();
  await expect(note(page, /Stopped before the rest\./)).toBeVisible();
  expect(await names(request, folder)).toEqual(['a.txt', 'b.txt', 'to']);
});

test('deletes after asking, and says it is permanent', async ({ page, request }) => {
  const folder = await folderWith(request, { 'x.txt': 'x', 'dir/y.txt': 'y' });
  await openFolder(page, folder);
  await cell(page, 'dir').click();
  await cell(page, 'x.txt').click({ modifiers: ['Control'] });
  await page
    .getByRole('toolbar', { name: 'Selected items' })
    .getByRole('button', { name: 'Delete' })
    .click();
  const dialog = page.getByRole('dialog', { name: 'Delete 2 items?' });
  await expect(dialog.getByText(/This is permanent/)).toBeVisible();
  await dialog.getByRole('button', { name: 'Delete' }).click();
  await expect(page.getByText('This folder is empty')).toBeVisible();
});

test('cuts here and pastes in another folder with the keyboard', async ({ page, request }) => {
  const folder = await folderWith(request, { 'note.txt': 'n', 'dest/.keep': '' });
  await openFolder(page, folder);
  await cell(page, 'note.txt').click();
  await page.keyboard.press('Control+x');
  await expect(cell(page, 'note.txt')).toHaveAccessibleName(/cut, waiting to be pasted$/);
  await cell(page, 'dest').dblclick();
  await expect(page).toHaveURL(/\/dest$/);
  await page.getByTestId('file-view').focus();
  await page.keyboard.press('Control+v');
  await expect(cell(page, 'note.txt')).toBeVisible();
  expect(await names(request, folder)).toEqual(['dest']);
});
