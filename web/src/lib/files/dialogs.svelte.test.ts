import { describe, expect, it, vi } from 'vitest';
import { page, userEvent } from 'vitest/browser';
import { render } from 'vitest-browser-svelte';
import { ApiError } from '$lib/api/errors';
import { fakeApi, problem } from '$lib/testing/fake-api';
import ConflictDialog from './ConflictDialog.svelte';
import { conflicts } from './conflicts.svelte';
import DeleteDialog from './DeleteDialog.svelte';
import DownloadAction from './DownloadAction.svelte';
import FolderPicker from './FolderPicker.svelte';
import NameDialog from './NameDialog.svelte';
import ShortcutsDialog from './ShortcutsDialog.svelte';
import type { FileItem } from './types';

const item = (path: string, kind: FileItem['kind'] = 'file'): FileItem => ({
  path,
  name: path.split('/').at(-1) ?? '',
  kind,
  size: 1,
  mod_time: '2026-09-28T00:00:00Z'
});

describe('NameDialog', () => {
  it('starts with the name, the stem selected, and submits it', async () => {
    const onsubmit = vi.fn(async () => {});
    await render(NameDialog, {
      open: true,
      title: 'Rename',
      label: 'New name',
      submitLabel: 'Rename',
      initial: 'report.final.pdf',
      selectStem: true,
      onsubmit
    });
    const field = page.getByRole('textbox', { name: 'New name' });
    await expect.element(field).toHaveValue('report.final.pdf');
    await expect.element(field).toHaveFocus();
    // The dialog focuses the field at once; the stem is selected a tick later.
    const input = field.element() as HTMLInputElement;
    await vi.waitFor(() =>
      expect([input.selectionStart, input.selectionEnd]).toEqual([0, 'report.final'.length])
    );
    await userEvent.keyboard('summary');
    await userEvent.keyboard('{Enter}');
    expect(onsubmit).toHaveBeenCalledWith('summary.pdf');
    await expect.element(page.getByRole('dialog')).not.toBeInTheDocument();
  });

  it('shows the refused name rule and stays open; an empty name cannot be sent', async () => {
    const onsubmit = vi.fn(async () => {
      throw new ApiError({
        status: 400,
        code: 'invalid_name',
        rule: 'forbidden_character',
        message: 'x'
      });
    });
    await render(NameDialog, {
      open: true,
      title: 'New folder',
      label: 'Name',
      submitLabel: 'Create',
      onsubmit
    });
    const create = page.getByRole('button', { name: 'Create' });
    await expect.element(create).toBeDisabled();
    await page.getByRole('textbox', { name: 'Name' }).fill('a:b');
    await create.click();
    await expect
      .element(page.getByRole('textbox', { name: 'Name' }))
      .toHaveAccessibleDescription(/Names cannot contain/);
    await expect.element(page.getByRole('dialog')).toBeVisible();
    // Typing again clears the error.
    await userEvent.keyboard('c');
    await expect.element(page.getByText(/Names cannot contain/)).not.toBeInTheDocument();
  });
});

describe('DeleteDialog', () => {
  it('says it is permanent, lists up to five items, and confirms', async () => {
    const onconfirm = vi.fn();
    const items = [item('/d', 'dir'), ...Array.from({ length: 6 }, (_, i) => item(`/f${i}`))];
    await render(DeleteDialog, { open: true, items, onconfirm });
    await expect.element(page.getByRole('dialog', { name: 'Delete 7 items?' })).toBeVisible();
    await expect.element(page.getByText(/This is permanent/)).toBeVisible();
    await expect
      .element(page.getByText(/Folders are deleted with everything in them/))
      .toBeVisible();
    await expect.element(page.getByText('and 2 more')).toBeVisible();
    await page.getByRole('button', { name: 'Delete' }).click();
    expect(onconfirm).toHaveBeenCalledOnce();
  });

  it('names a single item and cancels', async () => {
    const onconfirm = vi.fn();
    await render(DeleteDialog, { open: true, items: [item('/a.txt')], onconfirm });
    await expect.element(page.getByRole('dialog', { name: 'Delete “a.txt”?' })).toBeVisible();
    await page.getByRole('button', { name: 'Cancel' }).click();
    expect(onconfirm).not.toHaveBeenCalled();
  });
});

describe('ShortcutsDialog', () => {
  it('lists the shortcuts', async () => {
    await render(ShortcutsDialog, { open: true });
    await expect.element(page.getByRole('dialog', { name: 'Keyboard shortcuts' })).toBeVisible();
    await expect.element(page.getByText('Shift+N')).toBeVisible();
    await expect.element(page.getByText('New folder')).toBeVisible();
  });
});

describe('ConflictDialog', () => {
  it('offers replace, keep both, and skip for a file, with "apply to all"', async () => {
    await render(ConflictDialog);
    const answer = conflicts.ask({
      name: 'a.txt',
      folder: false,
      target: '/docs/a.txt',
      many: true,
      stoppable: true
    });
    const dialog = page.getByRole('dialog', { name: '“a.txt” is already there' });
    await expect.element(dialog).toBeVisible();
    await expect
      .element(dialog)
      .toHaveAccessibleDescription(/An item with this name is already in “docs”/);
    await expect.element(page.getByText('such as “a (1).txt”', { exact: false })).toBeVisible();
    await page.getByRole('checkbox').click();
    await page.getByRole('button', { name: /^Replace/ }).click();
    expect(await answer).toEqual({ choice: 'overwrite', all: true });
  });

  it('offers no replace for a folder, and Escape stops a stoppable run', async () => {
    await render(ConflictDialog);
    const answer = conflicts.ask({
      name: 'photos',
      folder: true,
      target: '/photos',
      many: true,
      stoppable: true
    });
    await expect
      .element(page.getByRole('dialog'))
      .toHaveAccessibleDescription(/A folder with this name is already in Files/);
    expect(page.getByRole('button', { name: /^Replace/ }).elements()).toHaveLength(0);
    await expect.element(page.getByText('such as “photos (1)”', { exact: false })).toBeVisible();
    await expect.element(page.getByRole('button', { name: 'Stop the operation' })).toBeVisible();
    await userEvent.keyboard('{Escape}');
    expect(await answer).toEqual({ choice: 'cancel', all: false });

    const skip = conflicts.ask({
      name: 'b',
      folder: false,
      target: '/b',
      many: false,
      stoppable: false
    });
    await expect.element(page.getByRole('dialog')).toBeVisible();
    await userEvent.keyboard('{Escape}');
    expect(await skip).toEqual({ choice: 'skip', all: false });
  });
});

describe('FolderPicker', () => {
  const tree: Record<string, FileItem[]> = {
    '/': [item('/docs', 'dir'), item('/photos', 'dir'), item('/a.txt')],
    '/docs': [item('/docs/old', 'dir')],
    '/docs/old': []
  };

  it('browses folders only and picks one', async () => {
    const { api, calls } = fakeApi((req) => ({
      body: {
        item: item(req.query.get('path') ?? '/', 'dir'),
        items: tree[req.query.get('path') ?? '/'] ?? []
      }
    }));
    const onpick = vi.fn();
    await render(FolderPicker, {
      open: true,
      title: 'Move 1 item',
      submitLabel: 'Move here',
      start: '/',
      items: [item('/a.txt')],
      move: true,
      onpick,
      api
    });
    await expect.element(page.getByRole('button', { name: 'docs' })).toBeVisible();
    expect(page.getByRole('button', { name: 'a.txt' }).elements()).toHaveLength(0);
    expect(calls[0].query.get('sort')).toBe('kind');
    // Moving to its own folder is refused.
    await expect.element(page.getByText('It is already in this folder.')).toBeVisible();
    await expect.element(page.getByRole('button', { name: 'Move here' })).toBeDisabled();
    await page.getByRole('button', { name: 'docs' }).click();
    await expect.element(page.getByText('Into “docs”.')).toBeVisible();
    await page.getByRole('button', { name: 'old' }).click();
    await expect.element(page.getByText('No folders in old.')).toBeVisible();
    await page.getByRole('button', { name: 'docs' }).click(); // the trail back up
    await page.getByRole('button', { name: 'Move here' }).click();
    expect(onpick).toHaveBeenCalledWith('/docs');
  });

  it('refuses a folder going into itself, and shows a load error', async () => {
    const { api } = fakeApi((req) =>
      req.query.get('path') === '/broken'
        ? problem(500, 'internal')
        : {
            body: { item: item('/', 'dir'), items: [item('/docs', 'dir'), item('/broken', 'dir')] }
          }
    );
    await render(FolderPicker, {
      open: true,
      title: 'Copy',
      submitLabel: 'Copy here',
      start: '/',
      items: [item('/docs', 'dir')],
      onpick: () => {},
      api
    });
    await page.getByRole('button', { name: 'docs' }).click();
    await expect.element(page.getByText('“docs” cannot go into itself.')).toBeVisible();
    await page.getByRole('button', { name: 'Files' }).click();
    await page.getByRole('button', { name: 'broken' }).click();
    await expect.element(page.getByText(/Something went wrong on the NAS/)).toBeVisible();
  });
});

describe('DownloadAction', () => {
  it('links a file for download and offers a folder as a ZIP', async () => {
    await render(DownloadAction, { item: item('/docs/a b.txt') });
    const link = page.getByRole('link', { name: 'Download a b.txt' });
    await expect
      .element(link)
      .toHaveAttribute('href', '/api/v1/files/content?path=%2Fdocs%2Fa%20b.txt');
    await expect.element(link).toHaveAttribute('tabindex', '-1');
    await render(DownloadAction, { item: item('/docs', 'dir') });
    await expect
      .element(page.getByRole('button', { name: 'Download docs as a ZIP file' }))
      .toBeVisible();
  });

  it('offers nothing for links and special files', async () => {
    const screen = await render(DownloadAction, { item: item('/l', 'symlink') });
    expect(screen.container.querySelector('a, button')).toBeNull();
  });
});
