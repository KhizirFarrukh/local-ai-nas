import Folder from '@lucide/svelte/icons/folder';
import { createRawSnippet, flushSync } from 'svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { page, userEvent } from 'vitest/browser';
import { render } from 'vitest-browser-svelte';
import { ApiError } from '$lib/api/errors';
import Breadcrumbs from './Breadcrumbs.svelte';
import Button from './Button.svelte';
import Checkbox from './Checkbox.svelte';
import Dialog from './Dialog.svelte';
import EmptyState from './EmptyState.svelte';
import ErrorPanel from './ErrorPanel.svelte';
import IconButton from './IconButton.svelte';
import Menu from './Menu.svelte';
import ProgressBar from './ProgressBar.svelte';
import Spinner from './Spinner.svelte';
import TextField from './TextField.svelte';
import ThemeSwitch from './ThemeSwitch.svelte';
import Toast from './Toast.svelte';
import Tooltip from './Tooltip.svelte';

const text = (html: string) => createRawSnippet(() => ({ render: () => html }));

afterEach(() => {
  document.documentElement.removeAttribute('data-theme');
});

describe('Breadcrumbs', () => {
  it('links every level but the current one', async () => {
    await render(Breadcrumbs, {
      crumbs: [
        { label: 'Files', href: '/files' },
        { label: 'docs', href: '/files/docs' }
      ],
      label: 'Folder'
    });
    await expect.element(page.getByRole('navigation', { name: 'Folder' })).toBeVisible();
    await expect
      .element(page.getByRole('link', { name: 'Files' }))
      .toHaveAttribute('href', '/files');
    await expect.element(page.getByText('docs')).toHaveAttribute('aria-current', 'page');
    expect(page.getByRole('link', { name: 'docs' }).elements()).toHaveLength(0);
  });
});

describe('Button and IconButton', () => {
  it('render as buttons with their text, variant, and events', async () => {
    const onclick = vi.fn();
    await render(Button, { variant: 'danger', onclick, children: text('<span>Delete</span>') });
    const button = page.getByRole('button', { name: 'Delete' });
    await expect.element(button).toHaveAttribute('type', 'button');
    await expect.element(button).toHaveClass(/bg-danger/);
    await button.click();
    expect(onclick).toHaveBeenCalledOnce();
  });

  it('names an icon button for screen readers and as a tooltip', async () => {
    await render(IconButton, { label: 'Close', size: 'sm', children: text('<svg></svg>') });
    const button = page.getByRole('button', { name: 'Close' });
    await expect.element(button).toHaveAttribute('title', 'Close');
    await expect.element(button).toHaveClass(/size-8/);
  });
});

describe('Checkbox', () => {
  it('toggles and clears "some selected" on click', async () => {
    const screen = await render(Checkbox, { label: 'Select all', indeterminate: true });
    const box = page.getByRole('checkbox', { name: 'Select all' });
    expect((box.element() as HTMLInputElement).indeterminate).toBe(true);
    await box.click();
    await expect.element(box).toBeChecked();
    expect((box.element() as HTMLInputElement).indeterminate).toBe(false);
    await screen.rerender({ label: undefined, 'aria-label': 'Row' });
    await expect.element(page.getByRole('checkbox', { name: 'Row' })).toBeVisible();
  });
});

describe('EmptyState and ErrorPanel', () => {
  it('shows a title, a description, and actions', async () => {
    await render(EmptyState, {
      icon: Folder,
      title: 'This folder is empty',
      description: 'Upload files.',
      actions: text('<button>Upload</button>')
    });
    await expect.element(page.getByRole('heading', { name: 'This folder is empty' })).toBeVisible();
    await expect.element(page.getByText('Upload files.')).toBeVisible();
    await expect.element(page.getByRole('button', { name: 'Upload' })).toBeVisible();
  });

  it('explains an error in plain words, with the request ID and a retry', async () => {
    const onretry = vi.fn();
    await render(ErrorPanel, {
      error: new ApiError({ status: 500, code: 'internal', message: 'x', correlationId: 'abc123' }),
      onretry
    });
    await expect.element(page.getByRole('alert')).toBeVisible();
    await expect.element(page.getByText('Something went wrong on the NAS')).toBeVisible();
    await expect.element(page.getByText('Request abc123')).toBeVisible();
    await page.getByRole('button', { name: 'Try again' }).click();
    expect(onretry).toHaveBeenCalledOnce();
  });
});

describe('ProgressBar and Spinner', () => {
  it('reports progress as a percentage, clamped', async () => {
    const screen = await render(ProgressBar, { value: 1, max: 4, label: 'Uploading a.bin' });
    const bar = page.getByRole('progressbar', { name: 'Uploading a.bin' });
    await expect.element(bar).toHaveAttribute('aria-valuenow', '25');
    await screen.rerender({ value: 9, max: 4 });
    await expect.element(bar).toHaveAttribute('aria-valuenow', '100');
    await screen.rerender({ value: undefined });
    await expect.element(bar).not.toHaveAttribute('aria-valuenow');
  });

  it('names what is loading', async () => {
    await render(Spinner, { label: 'Loading the folder' });
    await expect.element(page.getByRole('status')).toHaveTextContent('Loading the folder');
  });
});

describe('TextField', () => {
  it('labels the input, binds the value, and describes a hint or an error', async () => {
    const screen = await render(TextField, { label: 'Name', hint: 'Letters and digits.' });
    const input = page.getByRole('textbox', { name: 'Name' });
    await input.fill('report');
    await expect.element(input).toHaveValue('report');
    await expect.element(input).toHaveAccessibleDescription('Letters and digits.');
    await screen.rerender({ error: 'Name not allowed' });
    await expect.element(input).toHaveAttribute('aria-invalid', 'true');
    await expect.element(input).toHaveAccessibleDescription('Name not allowed');
  });
});

describe('ThemeSwitch', () => {
  it('sets the theme on <html> and remembers it', async () => {
    await render(ThemeSwitch);
    await page.getByRole('radio', { name: 'Dark' }).click();
    await expect
      .element(page.getByRole('radio', { name: 'Dark' }))
      .toHaveAttribute('aria-checked', 'true');
    expect(document.documentElement.dataset.theme).toBe('dark');
    expect(localStorage.getItem('local-ai-nas.theme')).toBe('dark');
    await page.getByRole('radio', { name: 'System' }).click();
    expect(document.documentElement.dataset.theme).toBeUndefined();
  });
});

describe('Toast and Tooltip', () => {
  it('shows a note with its detail and a dismiss button, and no live role of its own', async () => {
    const onclose = vi.fn();
    const screen = await render(Toast, {
      kind: 'error',
      message: 'Upload failed.',
      detail: 'Request 1',
      onclose
    });
    await expect.element(page.getByText('Upload failed.')).toBeVisible();
    await expect.element(page.getByText('Request 1')).toBeVisible();
    expect(screen.container.querySelector('[role="alert"], [role="status"]')).toBeNull();
    await page.getByRole('button', { name: 'Dismiss' }).click();
    expect(onclose).toHaveBeenCalledOnce();
  });

  it('describes its trigger', async () => {
    const trigger = createRawSnippet((attrs: () => { 'aria-describedby': string }) => ({
      render: () => `<button aria-describedby="${attrs()['aria-describedby']}">Info</button>`
    }));
    await render(Tooltip, { text: 'More about this', children: trigger });
    await expect
      .element(page.getByRole('button', { name: 'Info' }))
      .toHaveAccessibleDescription('More about this');
  });
});

describe('Dialog', () => {
  it('opens modally, names itself, closes on Escape, and gives the focus back', async () => {
    const opener = document.createElement('button');
    opener.textContent = 'Open';
    document.body.append(opener);
    opener.focus();
    const onclose = vi.fn();
    const screen = await render(Dialog, {
      open: true,
      title: 'Rename',
      description: 'Choose a new name.',
      children: text('<input aria-label="New name" />'),
      actions: text('<button>OK</button>'),
      onclose
    });
    const dialog = page.getByRole('dialog', { name: 'Rename' });
    await expect.element(dialog).toBeVisible();
    await expect.element(dialog).toHaveAccessibleDescription('Choose a new name.');
    await userEvent.keyboard('{Escape}');
    await expect.element(dialog).not.toBeInTheDocument();
    expect(onclose).toHaveBeenCalledOnce();
    expect(document.activeElement).toBe(opener);
    // Opened again from its owner, then closed by the owner.
    await screen.rerender({ open: true });
    await expect.element(dialog).toBeVisible();
    await screen.rerender({ open: false });
    await expect.element(dialog).not.toBeInTheDocument();
    opener.remove();
  });

  it('closes on a backdrop click unless persistent', async () => {
    const onclose = vi.fn();
    const screen = await render(Dialog, {
      open: true,
      title: 'Info',
      size: 'sm',
      phone: 'full',
      onclose
    });
    const el = screen.container.ownerDocument.querySelector('dialog') as HTMLDialogElement;
    el.dispatchEvent(new MouseEvent('click', { bubbles: true })); // the target is the dialog: the backdrop
    await expect.element(page.getByRole('dialog')).not.toBeInTheDocument();
    await vi.waitFor(() => expect(onclose).toHaveBeenCalledOnce()); // "close" fires a moment later

    await render(Dialog, { open: true, title: 'Typing', persistent: true });
    const kept = document.querySelectorAll('dialog')[1] as HTMLDialogElement;
    kept.dispatchEvent(new MouseEvent('click', { bubbles: true }));
    flushSync();
    expect(kept.open).toBe(true);
  });
});

describe('Menu', () => {
  const items = (fn: (label: string) => void) => [
    { label: 'Open', onselect: () => fn('Open') },
    'separator' as const,
    { label: 'Rename…', shortcut: 'F2', onselect: () => fn('Rename') },
    { label: 'Disabled', disabled: true, onselect: () => fn('Disabled') },
    { label: 'Delete', danger: true, onselect: () => fn('Delete') }
  ];

  it('focuses the first item, moves with the arrows, Home, and End, and picks with Enter', async () => {
    const picked: string[] = [];
    const opener = document.createElement('button');
    document.body.append(opener);
    opener.focus();
    await render(Menu, {
      open: true,
      x: 10,
      y: 10,
      label: 'Item actions',
      items: items((l) => picked.push(l))
    });
    await expect.element(page.getByRole('menu', { name: 'Item actions' })).toBeVisible();
    await expect.element(page.getByRole('menuitem', { name: 'Open' })).toHaveFocus();
    await userEvent.keyboard('{ArrowDown}');
    await expect.element(page.getByRole('menuitem', { name: /Rename/ })).toHaveFocus();
    await userEvent.keyboard('{ArrowDown}'); // the disabled item is skipped
    await expect.element(page.getByRole('menuitem', { name: 'Delete' })).toHaveFocus();
    await userEvent.keyboard('{ArrowDown}'); // wraps
    await expect.element(page.getByRole('menuitem', { name: 'Open' })).toHaveFocus();
    await userEvent.keyboard('{ArrowUp}');
    await expect.element(page.getByRole('menuitem', { name: 'Delete' })).toHaveFocus();
    await userEvent.keyboard('{Home}');
    await userEvent.keyboard('{End}');
    await userEvent.keyboard('{Enter}');
    expect(picked).toEqual(['Delete']);
    await expect.element(page.getByRole('menu')).not.toBeInTheDocument();
    expect(document.activeElement).toBe(opener);
    opener.remove();
  });

  it('closes on Escape, Tab, a click outside, and a resize', async () => {
    for (const close of [
      () => userEvent.keyboard('{Escape}'),
      () => userEvent.keyboard('{Tab}'),
      async () =>
        void document.body.dispatchEvent(new PointerEvent('pointerdown', { bubbles: true })),
      async () => void window.dispatchEvent(new Event('resize'))
    ]) {
      const onclose = vi.fn();
      const screen = await render(Menu, {
        open: true,
        x: 5,
        y: 5,
        label: 'M',
        items: items(() => {}),
        onclose
      });
      await expect.element(page.getByRole('menu')).toBeVisible();
      await close();
      await expect.element(page.getByRole('menu')).not.toBeInTheDocument();
      expect(onclose).toHaveBeenCalledOnce();
      await screen.unmount();
    }
  });

  it('stays inside the window', async () => {
    await render(Menu, { open: true, x: 5000, y: 5000, label: 'M', items: items(() => {}) });
    const box = page.getByRole('menu').element().getBoundingClientRect();
    expect(box.right).toBeLessThanOrEqual(window.innerWidth);
    expect(box.bottom).toBeLessThanOrEqual(window.innerHeight);
  });
});
