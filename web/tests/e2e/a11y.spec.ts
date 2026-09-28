// Accessibility checks with axe (S02.7, NFR-015) on the main screens, in
// both themes: no violation that axe rates serious or critical.
import AxeBuilder from '@axe-core/playwright';
import { expect, test, type Page } from '@playwright/test';
import { cell, folderWith, openFolder } from './helpers';

async function axe(page: Page, label: string) {
  const result = await new AxeBuilder({ page }).analyze();
  const serious = result.violations.filter(
    (v) => v.impact === 'serious' || v.impact === 'critical'
  );
  const all = result.violations.map(
    (v) => `${v.impact} ${v.id}: ${v.nodes.map((n) => n.target.join(' ')).join(', ')}`
  );
  test.info().annotations.push({
    type: 'axe',
    description: `${label}: ${all.length ? all.join('; ') : 'no violations'}`
  });
  expect(serious, `${label}: ${all.join('\n')}`).toEqual([]);
  expect(result.violations, `${label} (any impact): ${all.join('\n')}`).toEqual([]);
}

for (const scheme of ['light', 'dark'] as const) {
  test.describe(`${scheme} theme`, () => {
    test.use({ colorScheme: scheme });

    test('files, selection, menu, dialogs, preview, settings, photos', async ({
      page,
      request
    }) => {
      const folder = await folderWith(request, { 'a.txt': 'a', 'b.txt': 'b', 'sub/c.txt': 'c' });
      await openFolder(page, folder);
      await axe(page, 'files');
      await cell(page, 'a.txt').click();
      await axe(page, 'a selection');
      await cell(page, 'a.txt').click({ button: 'right' });
      await expect(page.getByRole('menu')).toBeVisible();
      await axe(page, 'the menu');
      await page.keyboard.press('Escape');
      await page.keyboard.press('F2');
      await expect(page.getByRole('dialog')).toBeVisible();
      await axe(page, 'rename');
      await page.keyboard.press('Escape');
      await page.keyboard.press('Delete');
      await axe(page, 'delete');
      await page.keyboard.press('Escape');
      await page.getByRole('button', { name: 'Clear the selection' }).click();
      await page.getByRole('button', { name: 'Grid view' }).click();
      await axe(page, 'grid');
      await page.getByRole('button', { name: 'List view' }).click();
      await openFolder(page, '/media');
      await cell(page, 'notes.txt').dblclick();
      await expect(page.getByTestId('preview')).toBeVisible();
      await axe(page, 'preview');
      await page.goto('/settings');
      await axe(page, 'settings');
      await page.goto('/photos');
      await axe(page, 'photos');
    });
  });
}
