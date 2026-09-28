// Downloads (S02.4-T04, FR-005, FR-006): a file by its link, a folder and a
// selection as ZIP archives.
import { readFileSync } from 'node:fs';
import { expect, test } from '@playwright/test';
import { cell, folderWith, openFolder, note } from './helpers';

/** The names in a ZIP file, read from its central directory. */
function zipNames(data: Buffer): string[] {
  const end = data.lastIndexOf(Buffer.from([0x50, 0x4b, 0x05, 0x06]));
  expect(end, 'end of central directory').toBeGreaterThan(-1);
  const count = data.readUInt16LE(end + 10);
  let at = data.readUInt32LE(end + 16);
  const names: string[] = [];
  for (let i = 0; i < count; i++) {
    expect(data.readUInt32LE(at)).toBe(0x02014b50);
    const nameLength = data.readUInt16LE(at + 28);
    const extra = data.readUInt16LE(at + 30);
    const comment = data.readUInt16LE(at + 32);
    names.push(data.subarray(at + 46, at + 46 + nameLength).toString('utf8'));
    at += 46 + nameLength + extra + comment;
  }
  return names.sort();
}

test('downloads a file by its link, byte for byte', async ({ page, request }) => {
  const folder = await folderWith(request, { 'report é.txt': 'quarterly numbers\n' });
  await openFolder(page, folder);
  await cell(page, 'report é.txt').hover();
  const download = page.waitForEvent('download');
  await page.getByRole('link', { name: 'Download report é.txt' }).click();
  const d = await download;
  expect(d.suggestedFilename()).toBe('report é.txt');
  expect(readFileSync(await d.path()).toString()).toBe('quarterly numbers\n');
});

test('downloads the folder on screen as one ZIP', async ({ page, request }) => {
  const folder = await folderWith(request, { 'a.txt': 'a', 'sub/b.txt': 'b' });
  await openFolder(page, folder);
  const download = page.waitForEvent('download');
  await page.getByRole('button', { name: 'Download folder' }).click();
  const d = await download;
  expect(d.suggestedFilename()).toBe(`${folder.slice(1)}.zip`);
  const top = folder.slice(1);
  expect(zipNames(readFileSync(await d.path()))).toEqual([
    `${top}/`,
    `${top}/a.txt`,
    `${top}/sub/`,
    `${top}/sub/b.txt`
  ]);
  await expect(note(page, /^Downloading “.*\.zip” \(4 items/)).toBeVisible();
});

test('downloads a selection of several items as one ZIP', async ({ page, request }) => {
  const folder = await folderWith(request, { 'a.txt': 'a', 'b.txt': 'b', 'c.txt': 'c' });
  await openFolder(page, folder);
  await cell(page, 'a.txt').click();
  await cell(page, 'b.txt').click({ modifiers: ['Shift'] });
  const download = page.waitForEvent('download');
  await page
    .getByRole('toolbar', { name: 'Selected items' })
    .getByRole('button', { name: 'Download' })
    .click();
  const d = await download;
  expect(d.suggestedFilename()).toMatch(/^download-\d{4}-\d{2}-\d{2}\.zip$/);
  expect(zipNames(readFileSync(await d.path()))).toEqual(['a.txt', 'b.txt']);
});
