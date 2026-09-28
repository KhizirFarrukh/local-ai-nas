// Shared helpers of the system tests (S02.8-T02).
import { expect, type APIRequestContext, type Page } from '@playwright/test';

let counter = 0;

/** A folder of the test's own, created through the API, with files in it. */
export async function folderWith(
  request: APIRequestContext,
  files: Record<string, string | Buffer> = {},
  prefix = 'case'
): Promise<string> {
  const folder = `/${prefix}-${Date.now().toString(36)}-${++counter}`;
  await mkdir(request, folder);
  for (const [name, content] of Object.entries(files)) {
    const parts = name.split('/');
    for (let i = 1; i < parts.length; i++) {
      await mkdir(request, `${folder}/${parts.slice(0, i).join('/')}`, true);
    }
    const res = await request.put(
      `/api/v1/files/content?path=${encodeURIComponent(`${folder}/${name}`)}`,
      {
        data: typeof content === 'string' ? Buffer.from(content) : content
      }
    );
    expect(res.status(), `upload ${name}`).toBe(201);
  }
  return folder;
}

export async function mkdir(request: APIRequestContext, path: string, existingOk = false) {
  const res = await request.post('/api/v1/files/folders', {
    data: { path, parents: true, on_conflict: existingOk ? 'overwrite' : 'fail' }
  });
  expect([200, 201]).toContain(res.status());
}

/** The names in a folder, from the API. */
export async function names(request: APIRequestContext, folder: string): Promise<string[]> {
  const res = await request.get(
    `/api/v1/files/items?path=${encodeURIComponent(folder)}&limit=1000`
  );
  expect(res.ok()).toBe(true);
  const body = (await res.json()) as { items?: { name: string }[] };
  return (body.items ?? []).map((i) => i.name).sort();
}

/** The content of a file, from the API. */
export async function content(request: APIRequestContext, path: string): Promise<string> {
  const res = await request.get(`/api/v1/files/content?path=${encodeURIComponent(path)}`);
  expect(res.ok()).toBe(true);
  return res.text();
}

/** The app URL of a folder, each name encoded. */
export function href(folder: string): string {
  return '/files' + folder.split('/').map(encodeURIComponent).join('/');
}

/** Opens a folder and waits for its view. */
export async function openFolder(page: Page, folder: string) {
  await page.goto(href(folder));
  await expect(
    page.getByTestId('file-view').or(page.getByText('This folder is empty'))
  ).toBeVisible();
}

/** The cell of an item, by its name. */
export function cell(page: Page, name: string) {
  return page.getByRole('gridcell', {
    name: new RegExp(`^${name.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')},`)
  });
}

/** Collects page errors and console errors, for "no error" checks. */
export function watchErrors(page: Page): string[] {
  const errors: string[] = [];
  page.on('pageerror', (e) => errors.push(`pageerror: ${e.message}`));
  page.on('console', (m) => {
    if (m.type() === 'error') errors.push(`console: ${m.text()}`);
  });
  return errors;
}

/**
 * A notification's text. The same text is also in the live region for
 * screen readers (S02.7-T02), so notifications are found in their region.
 */
export function note(page: Page, text: string | RegExp) {
  return page.getByRole('region', { name: 'Notifications' }).getByText(text);
}
