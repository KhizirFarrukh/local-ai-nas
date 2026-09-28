// Regression tests of S02 bugs that need the whole app (stage document,
// section 12): S02-B02 and S02-B05. (B03 is in browse.spec.ts, B06 in
// uploads.spec.ts, B10 in keyboard.spec.ts and previews.spec.ts.)
import { expect, test } from '@playwright/test';
import { folderWith, openFolder, note } from './helpers';

test('S02-B02: no page breaks the content security policy, and the route announcer stays hidden', async ({
  page,
  request
}) => {
  const violations: string[] = [];
  await page.addInitScript(() => {
    document.addEventListener('securitypolicyviolation', (e) => {
      (window as unknown as { cspViolations: string[] }).cspViolations ??= [];
      (window as unknown as { cspViolations: string[] }).cspViolations.push(
        `${e.violatedDirective} ${e.blockedURI}`
      );
    });
  });
  page.on('console', (m) => {
    if (/Content.Security.Policy/i.test(m.text())) violations.push(m.text());
  });
  const folder = await folderWith(request, { 'a.txt': 'a' });
  for (const path of ['/files', `/files${folder}`, '/files/media', '/settings', '/photos']) {
    await page.goto(path);
    await page.waitForLoadState('networkidle');
    const found = await page.evaluate(
      () => (window as unknown as { cspViolations?: string[] }).cspViolations ?? []
    );
    violations.push(...found);
  }
  // Navigating inside the app makes SvelteKit's announcer speak.
  await page.getByRole('link', { name: 'Files' }).first().click();
  const announcer = page.locator('#svelte-announcer');
  await expect(announcer).toHaveCount(1);
  const box = await announcer.boundingBox();
  expect(box === null || (box.width <= 1 && box.height <= 1)).toBe(true);
  expect(violations).toEqual([]);
});

test('S02-B05: with the server unreachable, moving between sections keeps the app and shows the banner', async ({
  page,
  context,
  request
}) => {
  const folder = await folderWith(request, { 'a.txt': 'a' });
  await openFolder(page, folder);
  await context.setOffline(true);
  for (const section of ['Settings', 'Photos', 'Files']) {
    await page
      .getByRole('navigation', { name: 'Sections' })
      .first()
      .getByRole('link', { name: section })
      .click();
    await expect(page.getByRole('heading', { level: 1 }).first()).toBeVisible();
  }
  await expect(page.getByText('Can’t reach the NAS.')).toBeVisible();
  // "Try now" while still offline checks and keeps the banner.
  await page.getByRole('button', { name: 'Try now' }).click();
  await expect(page.getByText('Can’t reach the NAS.')).toBeVisible();
  // Back online, the app checks by itself (the browser's online event).
  await context.setOffline(false);
  await expect(page.getByText('Can’t reach the NAS.')).toBeHidden();
  await expect(note(page, 'Connected to the NAS again.')).toBeVisible();
});
