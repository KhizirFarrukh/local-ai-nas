import { describe, expect, it, vi } from 'vitest';
import { page, userEvent } from 'vitest/browser';
import { render } from 'vitest-browser-svelte';
// The app's styles, so the list is laid out and scrolls as in the app
// (only the rows on screen exist) and the blink's animation applies.
import '../../app.css';
import FileView from './FileView.svelte';
import { FolderListing } from './listing.svelte';
import { Selection } from './selection.svelte';
import { FolderSizes } from './sizes.svelte';
import type { FileItem } from './types';

function items(n: number): FileItem[] {
  return Array.from({ length: n }, (_, i) => ({
    path: `/f/item${String(i).padStart(3, '0')}${i % 3 === 0 ? '' : '.txt'}`,
    name: `item${String(i).padStart(3, '0')}${i % 3 === 0 ? '' : '.txt'}`,
    kind: i % 3 === 0 ? 'dir' : 'file',
    size: 1000 * i,
    mod_time: '2026-09-28T10:00:00Z',
    added_time: '2026-09-28T10:00:00Z'
  }));
}

/** A box of fixed size, as the main area is in the app. */
function box(width = 900): HTMLElement {
  const el = document.createElement('div');
  el.style.cssText = `width: ${width}px; height: 440px; display: flex; flex-direction: column`;
  document.body.append(el);
  return el;
}

async function setup(n = 30, extra: Record<string, unknown> = {}, width = 900) {
  const list = items(n);
  const listing = new FolderListing(
    async () => ({
      item: { path: '/f', name: 'f', kind: 'dir', size: 0, mod_time: '', added_time: '' },
      items: list,
      total: n
    }),
    '/f'
  );
  await listing.start();
  const selection = new Selection();
  const onopen = vi.fn();
  const onsort = vi.fn();
  const onmenu = vi.fn();
  const screen = await render(FileView, {
    target: box(width),
    props: {
      listing,
      selection,
      mode: 'list',
      sort: 'name',
      order: 'asc',
      onopen,
      onsort,
      onmenu,
      ...extra
    }
  });
  const grid = page.getByTestId('file-view');
  await expect.element(grid).toBeVisible();
  return { list, listing, selection, onopen, onsort, onmenu, grid, screen };
}

const selectedNames = (s: Selection) => [...s.items.values()].map((i) => i.name).sort();
const active = () => {
  const id = document.querySelector('[role=grid]')?.getAttribute('aria-activedescendant');
  return id ? document.getElementById(id)?.getAttribute('aria-label') : undefined;
};

describe('FileView as a list', () => {
  it('shows Size, Added, and Modified where they fit, else the size and date under the name', async () => {
    const header = () =>
      [...document.querySelectorAll('button[aria-label^="Sort by"]')].map((b) =>
        b.textContent?.trim()
      );
    const wide = await setup(30, {}, 850); // Type needs 900 px
    await vi.waitFor(() => expect(header()).toEqual(['Name', 'Size', 'Added', 'Modified']));
    await expect
      .poll(() => page.getByTestId('item-added').first().element().textContent)
      .toMatch(/2026/);
    await wide.screen.unmount();
    await setup(30, {}, 400);
    await vi.waitFor(() => expect(header()).toEqual(['Name']));
    await expect
      .poll(() => page.getByTestId('item-details').nth(1).element().textContent)
      .toMatch(/^1,000 bytes · Added .*2026/);
  });

  it('shows the size of a folder once it is added up, asking only for the folders on screen', async () => {
    const asked: string[] = [];
    const sizes = new FolderSizes(async (path) => {
      asked.push(path);
      return 4096;
    });
    await setup(300, { sizes });
    await expect
      .element(page.getByRole('gridcell', { name: /^item000, Folder, 4 KB, added/ }))
      .toBeVisible();
    await expect.element(page.getByTestId('item-size').first()).toHaveTextContent('4 KB');
    expect(asked.length).toBeGreaterThan(0);
    expect(asked.length).toBeLessThan(40); // not all 100 folders
    expect(asked.every((p) => !p.endsWith('.txt'))).toBe(true);
  });

  it('is one grid named by the folder, with one column and every item described', async () => {
    const { grid } = await setup();
    await expect.element(grid).toHaveAttribute('aria-label', 'Items in f');
    await expect.element(grid).toHaveAttribute('aria-colcount', '1');
    await expect.element(grid).toHaveAttribute('aria-multiselectable', 'true');
    await expect.element(grid).toHaveAttribute('aria-rowcount', '30');
    await expect
      .element(page.getByRole('gridcell', { name: /^item000, Folder, added .+, modified / }))
      .toBeVisible();
    await expect
      .element(
        page.getByRole('gridcell', {
          name: /^item001\.txt, TXT file, 1,000 bytes, added .+, modified /
        })
      )
      .toBeVisible();
  });

  it('moves and selects with the keys, and opens with Enter', async () => {
    const { grid, selection, onopen, list } = await setup();
    (grid.element() as HTMLElement).focus();
    await userEvent.keyboard('{ArrowDown}');
    expect(active()).toMatch(/^item001\.txt/);
    expect(selectedNames(selection)).toEqual(['item001.txt']);
    await userEvent.keyboard('{Shift>}{ArrowDown}{ArrowDown}{/Shift}');
    expect(selectedNames(selection)).toEqual(['item001.txt', 'item002.txt', 'item003']);
    await userEvent.keyboard('{Control>}{ArrowDown}{/Control}'); // Ctrl moves the focus only
    expect(active()).toMatch(/^item004/);
    expect(selection.count(30)).toBe(3);
    await userEvent.keyboard(' '); // Space adds the focused item
    expect(selection.count(30)).toBe(4);
    await userEvent.keyboard('{End}');
    expect(active()).toMatch(/^item029/);
    await userEvent.keyboard('{Home}');
    expect(active()).toMatch(/^item000/);
    await userEvent.keyboard('{PageDown}');
    expect(active()).not.toMatch(/^item000/);
    await userEvent.keyboard('{PageUp}');
    expect(active()).toMatch(/^item000/);
    await userEvent.keyboard('{ArrowUp}'); // stays at the top
    expect(active()).toMatch(/^item000/);
    await userEvent.keyboard('{Enter}');
    expect(onopen).toHaveBeenCalledWith(list[0], 0);
    await userEvent.keyboard('{Control>}a{/Control}');
    expect(selection.everything).toBe(true);
    await userEvent.keyboard('{Escape}');
    expect(selection.count(30)).toBe(0);
  });

  it('selects with the mouse as file managers do', async () => {
    const { selection, onopen, list } = await setup();
    const cell = (i: number) =>
      page.getByRole('gridcell', { name: new RegExp(`^${list[i].name.replace('.', '\\.')},`) });
    await cell(1).click();
    await cell(3).click({ modifiers: ['Shift'] });
    expect(selectedNames(selection)).toEqual(['item001.txt', 'item002.txt', 'item003']);
    await cell(2).click({ modifiers: ['Control'] });
    expect(selectedNames(selection)).toEqual(['item001.txt', 'item003']);
    await cell(5).click({ modifiers: ['Control', 'Shift'] }); // adds the range from the anchor
    expect(selection.count(30)).toBe(5);
    await cell(4).dblClick();
    expect(onopen).toHaveBeenCalledWith(list[4], 4);
    expect(selectedNames(selection)).toEqual(['item004.txt']);
    await expect.element(cell(4)).toHaveAttribute('aria-selected', 'true');
  });

  it('opens the menu of the selection by right-click, Shift+F10, and the menu key', async () => {
    const { grid, onmenu, selection, list } = await setup();
    await page.getByRole('gridcell', { name: /^item002\.txt,/ }).click({ button: 'right' });
    expect(onmenu).toHaveBeenLastCalledWith(list[2], expect.any(Number), expect.any(Number));
    expect(selectedNames(selection)).toEqual(['item002.txt']); // selected first, as in Explorer
    (grid.element() as HTMLElement).focus();
    await userEvent.keyboard('{Shift>}{F10}{/Shift}');
    expect(onmenu).toHaveBeenLastCalledWith(list[2], expect.any(Number), expect.any(Number));
    await userEvent.keyboard('{ContextMenu}');
    expect(onmenu).toHaveBeenCalledTimes(3);
  });

  it('sorts by the column buttons, which say the current order', async () => {
    const { onsort } = await setup();
    await expect
      .element(page.getByRole('button', { name: 'Sort by name, now ascending' }))
      .toBeVisible();
    await page.getByRole('button', { name: 'Sort by size' }).click();
    expect(onsort).toHaveBeenCalledWith('size');
  });

  it('clears the selection with a click on empty space, and shows cut items dimmed', async () => {
    const { selection, grid } = await setup(3, {
      dimmed: (p: string) => p.endsWith('item001.txt')
    });
    await page.getByRole('gridcell', { name: /^item001\.txt,/ }).click();
    await expect
      .element(page.getByRole('gridcell', { name: /cut, waiting to be pasted$/ }))
      .toBeVisible();
    const el = grid.element() as HTMLElement;
    const r = el.getBoundingClientRect();
    await userEvent.click(el, { position: { x: 20, y: r.height - 10 } });
    expect(selection.count(3)).toBe(0);
  });

  it('takes the focus when asked, as after opening a folder', async () => {
    await setup(5, { autofocus: true });
    expect(document.activeElement?.getAttribute('role')).toBe('grid');
  });
});

describe('FileView as a grid', () => {
  it('lays tiles in columns and moves in two directions', async () => {
    const { grid, selection } = await setup(30, { mode: 'grid' });
    const columns = Number((grid.element() as HTMLElement).getAttribute('aria-colcount'));
    expect(columns).toBeGreaterThan(1);
    await expect
      .element(page.getByRole('gridcell', { name: /^item001\.txt,/ }))
      .toHaveAttribute('aria-colindex', '2');
    (grid.element() as HTMLElement).focus();
    await userEvent.keyboard('{ArrowRight}');
    expect(active()).toMatch(/^item001/);
    await userEvent.keyboard('{ArrowDown}');
    expect(active()).toMatch(new RegExp(`^item${String(1 + columns).padStart(3, '0')}`));
    await userEvent.keyboard('{ArrowLeft}');
    expect(selection.count(30)).toBe(1);
  });
});

describe('FileView on a touch screen', () => {
  it('opens with a tap, selects and shows the menu with a long press, then taps toggle', async () => {
    const { onopen, onmenu, selection, list } = await setup(6);
    await expect.element(page.getByRole('gridcell', { name: /^item005/ })).toBeVisible();
    const cell = (i: number) =>
      page
        .getByRole('gridcell', { name: new RegExp(`^${list[i].name.replace('.', '\\.')},`) })
        .element() as HTMLElement;
    const touch = (el: HTMLElement, type: string) => {
      const r = el.getBoundingClientRect();
      el.dispatchEvent(
        new PointerEvent(type, {
          pointerType: 'touch',
          bubbles: true,
          clientX: r.left + 10,
          clientY: r.top + 10
        })
      );
    };
    touch(cell(1), 'pointerdown');
    touch(cell(1), 'pointerup');
    cell(1).click();
    expect(onopen).toHaveBeenCalledWith(list[1], 1);
    // A long press (500 ms) selects and opens the menu; the click after it is dropped.
    touch(cell(2), 'pointerdown');
    await new Promise((r) => setTimeout(r, 600));
    expect(onmenu).toHaveBeenCalledWith(list[2], expect.any(Number), expect.any(Number));
    expect(selectedNames(selection)).toEqual(['item002.txt']);
    touch(cell(2), 'pointerup');
    cell(2).click(); // dropped
    expect(selectedNames(selection)).toEqual(['item002.txt']);
    // While something is selected, taps add or remove.
    touch(cell(4), 'pointerdown');
    touch(cell(4), 'pointerup');
    cell(4).click();
    await vi.waitFor(() =>
      expect(selectedNames(selection)).toEqual(['item002.txt', 'item004.txt'])
    );
    // Moving the finger is scrolling, not a press.
    touch(cell(5), 'pointerdown');
    const r = cell(5).getBoundingClientRect();
    cell(5).dispatchEvent(
      new PointerEvent('pointermove', {
        pointerType: 'touch',
        bubbles: true,
        clientX: r.left + 60,
        clientY: r.top + 60
      })
    );
    await new Promise((r2) => setTimeout(r2, 600));
    expect(onmenu).toHaveBeenCalledTimes(1);
  });
});

describe('reveal (S02.4-T05)', () => {
  it('scrolls to an item and blinks it twice, then stops', async () => {
    const { screen, list } = await setup(300);
    const target = list[250];
    await (
      screen.component as unknown as { reveal: (i: number, p: string) => Promise<void> }
    ).reveal(250, target.path);
    const cell = page.getByRole('gridcell', {
      name: new RegExp(`^${target.name.replace('.', '\\.')},`)
    });
    await expect.element(cell).toBeVisible();
    await expect.element(cell).toHaveClass('flash');
    const style = getComputedStyle(cell.element());
    expect(style.animationName).toBe('flash');
    expect(style.animationIterationCount).toBe('2');
    await expect
      .poll(() => cell.element().classList.contains('flash'), { timeout: 3000 })
      .toBe(false);
  });
});
