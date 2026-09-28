import { describe, expect, it, vi } from 'vitest';
import { page } from 'vitest/browser';
import { render } from 'vitest-browser-svelte';
import { ApiError } from '$lib/api/errors';
import { api, unwrap } from '$lib/api/client';
import DropZone from '$lib/uploads/DropZone.svelte';
import { focusFor, usingKeyboard } from '$lib/util/modality';
import { applyTheme, isThemeChoice, readTheme, Theme } from '$lib/util/theme.svelte';
import { announcer } from './announcer.svelte';
import ConnectionBanner from './ConnectionBanner.svelte';
import { connection } from './connection.svelte';
import LiveRegions from './LiveRegions.svelte';
import Nav from './Nav.svelte';
import ProgressPanel from './ProgressPanel.svelte';
import { tasks } from './tasks.svelte';
import Toaster from './Toaster.svelte';
import { toasts } from './toasts.svelte';

describe('LiveRegions', () => {
  it('shows what the announcer says, polite and urgent apart', async () => {
    await render(LiveRegions);
    announcer.say('3 selected');
    announcer.say('Upload failed', true);
    await expect.element(page.getByRole('status')).toHaveTextContent('3 selected');
    await expect.element(page.getByRole('alert')).toHaveTextContent('Upload failed');
  });
});

describe('Toaster', () => {
  it('stacks the notes in a named region, each dismissable', async () => {
    await render(Toaster);
    const id = toasts.push({ kind: 'error', message: 'It broke.' });
    await expect.element(page.getByRole('region', { name: 'Notifications' })).toBeVisible();
    await expect.element(page.getByText('It broke.')).toBeVisible();
    await page.getByRole('button', { name: 'Dismiss' }).click();
    await expect.element(page.getByText('It broke.')).not.toBeInTheDocument();
    expect(toasts.items.some((t) => t.id === id)).toBe(false);
  });
});

describe('ProgressPanel', () => {
  it('lists tasks with progress and cancel, folds, and says when all is finished', async () => {
    await render(ProgressPanel);
    const cancel = vi.fn();
    const task = tasks.start('Copying 3 items', { total: 3, cancel });
    task.update(1, 3, '1 of 3');
    const panel = page.getByRole('region', { name: 'In progress' });
    await expect.element(panel).toBeVisible();
    await expect
      .element(page.getByRole('button', { name: 'In progress (1)' }))
      .toHaveAttribute('aria-expanded', 'true');
    await expect
      .element(page.getByRole('progressbar', { name: 'Copying 3 items' }))
      .toHaveAttribute('aria-valuenow', '33');
    await expect.element(page.getByText('1 of 3')).toBeVisible();
    await page.getByRole('button', { name: 'Cancel Copying 3 items' }).click();
    expect(cancel).toHaveBeenCalledOnce();
    await page.getByRole('button', { name: 'In progress (1)' }).click();
    await expect
      .element(page.getByRole('button', { name: 'In progress (1)' }))
      .toHaveAttribute('aria-expanded', 'false');
    await expect.element(page.getByText('1 of 3')).not.toBeInTheDocument();
    task.finish('Copied.');
    await expect.element(panel).not.toBeInTheDocument();
  });
});

describe('ConnectionBanner', () => {
  it('appears when a request gets no answer, and goes when the server answers', async () => {
    const conn = connection();
    await render(ConnectionBanner);
    await expect.element(page.getByRole('alert')).not.toBeInTheDocument();
    // A request that gets no answer marks the connection lost.
    const failing = Promise.reject(new TypeError('Failed to fetch'));
    await expect(unwrap(failing)).rejects.toBeInstanceOf(ApiError);
    await expect.element(page.getByText('Can’t reach the NAS.')).toBeVisible();
    // "Try now" checks the health endpoint; the test server has none.
    const health = vi.spyOn(api, 'GET').mockResolvedValue({
      data: { status: 'ok', checks: [] },
      error: undefined,
      response: new Response(null, { status: 200 })
    } as never);
    await page.getByRole('button', { name: 'Try now' }).click();
    await expect.element(page.getByText('Can’t reach the NAS.')).not.toBeInTheDocument();
    expect(conn.online).toBe(true);
    expect(toasts.items.at(-1)?.message).toBe('Connected to the NAS again.');
    health.mockRestore();
  });
});

describe('Nav', () => {
  it('links the sections', async () => {
    await render(Nav, { layout: 'side' });
    await expect.element(page.getByRole('navigation', { name: 'Sections' })).toBeVisible();
    await expect
      .element(page.getByRole('link', { name: 'Files' }))
      .toHaveAttribute('href', '/files');
    await expect
      .element(page.getByRole('link', { name: 'Settings' }))
      .toHaveAttribute('href', '/settings');
    await render(Nav, { layout: 'bottom' });
    expect(page.getByRole('navigation', { name: 'Sections' }).elements()).toHaveLength(2);
  });
});

describe('DropZone', () => {
  function drag(type: string, files: File[] = []) {
    const data = new DataTransfer();
    for (const f of files) data.items.add(f);
    return new DragEvent(type, { dataTransfer: data, bubbles: true, cancelable: true });
  }

  it('shows where a drop goes while files are dragged, and hands the drop over', async () => {
    const ondrop = vi.fn();
    await render(DropZone, { props: { target: 'docs', ondrop } });
    const file = new File(['x'], 'a.txt');
    window.dispatchEvent(drag('dragenter', [file]));
    await expect.element(page.getByText('into docs')).toBeVisible();
    const over = drag('dragover', [file]);
    window.dispatchEvent(over);
    expect(over.defaultPrevented).toBe(true);
    window.dispatchEvent(drag('dragleave', [file]));
    await expect.element(page.getByText('into docs')).not.toBeInTheDocument();
    window.dispatchEvent(drag('dragenter', [file]));
    window.dispatchEvent(drag('drop', [file]));
    await vi.waitFor(() => expect(ondrop).toHaveBeenCalledOnce());
    expect(
      ondrop.mock.calls[0][0].files.map((f: { relativePath: string }) => f.relativePath)
    ).toEqual(['a.txt']);
    await expect.element(page.getByText('into docs')).not.toBeInTheDocument();
  });

  it('ignores drags without files', async () => {
    await render(DropZone, { props: { target: 'docs', ondrop: () => {} } });
    const text = new DataTransfer();
    text.setData('text/plain', 'hello');
    window.dispatchEvent(new DragEvent('dragenter', { dataTransfer: text, bubbles: true }));
    await new Promise((r) => setTimeout(r, 50));
    expect(page.getByText('into docs').elements()).toHaveLength(0);
  });
});

describe('theme', () => {
  function store(initial?: string) {
    const data = new Map<string, string>(initial ? [['local-ai-nas.theme', initial]] : []);
    return {
      getItem: (k: string) => data.get(k) ?? null,
      setItem: (k: string, v: string) => void data.set(k, v),
      data
    };
  }

  it('reads a remembered choice, anything else means the system theme', () => {
    expect(readTheme(store('dark'))).toBe('dark');
    expect(readTheme(store('neon'))).toBe('system');
    expect(readTheme(undefined)).toBe('system');
    expect(
      readTheme({
        getItem: () => {
          throw new Error('blocked');
        },
        setItem: () => {}
      })
    ).toBe('system');
    expect([isThemeChoice('light'), isThemeChoice(3)]).toEqual([true, false]);
  });

  it('applies and remembers a choice on an element', () => {
    const el = document.createElement('div');
    const s = store('light');
    const t = new Theme(s, el);
    expect(el.dataset.theme).toBe('light');
    t.set('system');
    expect(el.dataset.theme).toBeUndefined();
    expect(s.data.get('local-ai-nas.theme')).toBe('system');
    const blocked = new Theme(
      {
        getItem: () => null,
        setItem: () => {
          throw new Error('full');
        }
      },
      el
    );
    blocked.set('dark'); // not remembered, still applied
    expect(el.dataset.theme).toBe('dark');
    applyTheme(el, 'system');
    expect(el.dataset.theme).toBeUndefined();
  });
});

describe('focusFor', () => {
  it('focuses with a ring after keys, without one after the pointer', async () => {
    const a = document.createElement('button');
    const b = document.createElement('button');
    document.body.append(a, b);
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Shift' })); // modifiers alone do not count
    window.dispatchEvent(new PointerEvent('pointerdown'));
    expect(usingKeyboard()).toBe(false);
    focusFor(a);
    expect(document.activeElement).toBe(a);
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab' }));
    expect(usingKeyboard()).toBe(true);
    focusFor(b);
    expect(document.activeElement).toBe(b);
    expect(b.matches(':focus-visible')).toBe(true);
    focusFor(b); // already focused with a ring: nothing changes
    expect(document.activeElement).toBe(b);
    focusFor(null);
    focusFor(document.createElement('button')); // not in the document: ignored
    expect(document.activeElement).toBe(b);
    a.remove();
    b.remove();
  });
});
