import { describe, expect, it, vi } from 'vitest';
import { page, userEvent } from 'vitest/browser';
import { render } from 'vitest-browser-svelte';
import { FolderListing } from '$lib/files/listing.svelte';
import type { FileItem } from '$lib/files/types';
import AudioView from './AudioView.svelte';
import FallbackCard from './FallbackCard.svelte';
import ImageView from './ImageView.svelte';
import PdfView from './PdfView.svelte';
import PreviewFrame from './PreviewFrame.svelte';
import TextView from './TextView.svelte';
import VideoView from './VideoView.svelte';

const file = (name: string, size = 10, mime?: string): FileItem => ({
  path: `/f/${name}`,
  name,
  kind: 'file',
  size,
  mod_time: '2026-09-28T00:00:00Z',
  mime
});

/** A small PNG made on a canvas. */
function pngUrl(): string {
  const c = document.createElement('canvas');
  c.width = 4;
  c.height = 3;
  c.getContext('2d')?.fillRect(0, 0, 4, 3);
  return c.toDataURL('image/png');
}

/** A tenth of a second of silence as a WAV file. */
function wavUrl(): string {
  const samples = 800;
  const b = new DataView(new ArrayBuffer(44 + samples * 2));
  const str = (o: number, s: string) =>
    [...s].forEach((ch, i) => b.setUint8(o + i, ch.charCodeAt(0)));
  str(0, 'RIFF');
  b.setUint32(4, 36 + samples * 2, true);
  str(8, 'WAVEfmt ');
  b.setUint32(16, 16, true);
  b.setUint16(20, 1, true);
  b.setUint16(22, 1, true);
  b.setUint32(24, 8000, true);
  b.setUint32(28, 16000, true);
  b.setUint16(32, 2, true);
  b.setUint16(34, 16, true);
  str(36, 'data');
  b.setUint32(40, samples * 2, true);
  return URL.createObjectURL(new Blob([b.buffer], { type: 'audio/wav' }));
}

/** A valid PDF with n empty pages. */
function pdfUrl(n: number): string {
  const objects = [
    '<< /Type /Catalog /Pages 2 0 R >>',
    `<< /Type /Pages /Kids [${Array.from({ length: n }, (_, i) => `${3 + i} 0 R`).join(' ')}] /Count ${n} >>`,
    ...Array.from({ length: n }, () => '<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 100] >>')
  ];
  let out = '%PDF-1.4\n';
  const offsets: number[] = [];
  objects.forEach((body, i) => {
    offsets.push(out.length);
    out += `${i + 1} 0 obj\n${body}\nendobj\n`;
  });
  const xref = out.length;
  out += `xref\n0 ${objects.length + 1}\n0000000000 65535 f \n`;
  out += offsets.map((o) => `${String(o).padStart(10, '0')} 00000 n \n`).join('');
  out += `trailer\n<< /Size ${objects.length + 1} /Root 1 0 R >>\nstartxref\n${xref}\n%%EOF\n`;
  return URL.createObjectURL(new Blob([out], { type: 'application/pdf' }));
}

const blobUrl = (text: string) => URL.createObjectURL(new Blob([text], { type: 'text/plain' }));

describe('ImageView', () => {
  it('shows an image that loads, and reports one that cannot be read', async () => {
    const onfail = vi.fn();
    await render(ImageView, { item: file('dot.png'), src: pngUrl(), onfail });
    await expect.element(page.getByRole('img', { name: 'dot.png' })).toBeVisible();
    expect(onfail).not.toHaveBeenCalled();
    await render(ImageView, { item: file('bad.png'), src: blobUrl('not an image'), onfail });
    await vi.waitFor(() =>
      expect(onfail).toHaveBeenCalledWith(expect.stringMatching(/cannot be shown/))
    );
  });
});

describe('AudioView and VideoView', () => {
  it('give the browser player a name, and report a format it cannot play', async () => {
    const onfail = vi.fn();
    await render(AudioView, { item: file('quiet.wav'), src: wavUrl(), onfail });
    await expect
      .element(page.getByTestId('preview-audio'))
      .toHaveAttribute('aria-label', 'quiet.wav');
    await expect.element(page.getByText('quiet.wav')).toBeVisible();
    await new Promise((r) => setTimeout(r, 200));
    expect(onfail).not.toHaveBeenCalled();

    await render(AudioView, { item: file('bad.mp3'), src: blobUrl('noise'), onfail });
    await vi.waitFor(() =>
      expect(onfail).toHaveBeenCalledWith(expect.stringMatching(/audio cannot be played/))
    );

    const onVideoFail = vi.fn();
    await render(VideoView, {
      item: file('clip.mp4'),
      src: blobUrl('not a video'),
      onfail: onVideoFail
    });
    const video = page.getByTestId('preview-video');
    await expect.element(video).toHaveAttribute('preload', 'metadata');
    await expect.element(video).toHaveAttribute('aria-label', 'clip.mp4');
    await vi.waitFor(() =>
      expect(onVideoFail).toHaveBeenCalledWith(expect.stringMatching(/video cannot be played/))
    );
  });
});

describe('TextView', () => {
  it('shows text as text, never as HTML', async () => {
    await render(TextView, {
      item: file('page.html', 30),
      src: blobUrl('<b id="x">bold</b>'),
      onfail: () => {}
    });
    const pre = page.getByTestId('preview-text');
    await expect.element(pre).toHaveTextContent('<b id="x">bold</b>');
    expect(document.getElementById('x')).toBeNull();
    await expect.element(pre).toHaveAttribute('tabindex', '0');
  });

  it('says when only the start is shown, and when the file is empty', async () => {
    const big = 'x'.repeat(300 * 1024);
    await render(TextView, {
      item: file('big.log', big.length),
      src: blobUrl(big),
      onfail: () => {}
    });
    await expect
      .element(page.getByTestId('preview-notice'))
      .toHaveTextContent('Showing the first 256 KB of 300 KB. Download the file to see all of it.');
    await render(TextView, { item: file('empty.txt', 0), src: blobUrl(''), onfail: () => {} });
    await expect.element(page.getByText('(This file is empty.)')).toBeVisible();
  });

  it('reports a file that cannot be read', async () => {
    const onfail = vi.fn();
    await render(TextView, {
      item: file('gone.txt'),
      src: 'http://127.0.0.1:1/unreachable',
      onfail
    });
    await vi.waitFor(() =>
      expect(onfail).toHaveBeenCalledWith(expect.stringMatching(/^This file could not be read/))
    );
  });
});

/** A box of fixed size, as the preview area is in the app. */
function box(): HTMLElement {
  const el = document.createElement('div');
  el.style.cssText = 'width: 400px; height: 300px; display: flex';
  document.body.append(el);
  return el;
}

describe('PdfView', () => {
  it('shows page 1 and turns pages with the keys', async () => {
    const onfail = vi.fn();
    await render(PdfView, {
      target: box(),
      props: { item: file('two.pdf'), src: pdfUrl(2), onfail }
    });
    const canvas = page.getByTestId('preview-pdf');
    await expect.element(canvas).toHaveAttribute('aria-label', 'Page 1 of two.pdf');
    await expect.element(page.getByText('Page 1 of 2')).toBeVisible();
    await userEvent.keyboard('{PageDown}');
    await expect.element(canvas).toHaveAttribute('aria-label', 'Page 2 of two.pdf');
    await userEvent.keyboard('{PageDown}'); // stays on the last page
    await expect.element(canvas).toHaveAttribute('aria-label', 'Page 2 of two.pdf');
    await userEvent.keyboard('{Home}');
    await expect.element(canvas).toHaveAttribute('aria-label', 'Page 1 of two.pdf');
    await page.getByRole('button', { name: 'Next page' }).click();
    await expect.element(canvas).toHaveAttribute('aria-label', 'Page 2 of two.pdf');
    expect(onfail).not.toHaveBeenCalled();
  });

  it('explains a file that is not a PDF', async () => {
    const onfail = vi.fn();
    await render(PdfView, {
      target: box(),
      props: { item: file('fake.pdf'), src: blobUrl('hello'), onfail }
    });
    await vi.waitFor(
      () => expect(onfail).toHaveBeenCalledWith('This file is not a valid PDF, or it is damaged.'),
      {
        timeout: 5000
      }
    );
  });
});

describe('FallbackCard', () => {
  it('describes the file and offers the download', async () => {
    await render(FallbackCard, { item: file('setup.exe', 2048) });
    await expect.element(page.getByText('EXE file · 2 KB')).toBeVisible();
    await expect
      .element(page.getByText('There is no preview for this type of file.'))
      .toBeVisible();
    await expect
      .element(page.getByRole('link', { name: 'Download' }))
      .toHaveAttribute('href', '/api/v1/files/content?path=%2Ff%2Fsetup.exe');
  });
});

describe('PreviewFrame', () => {
  const items: FileItem[] = [
    file('a.txt', 5, 'text/plain'),
    { ...file('folder'), kind: 'dir' },
    file('b.bin', 7),
    file('c.png', 9, 'image/png')
  ];

  async function listingOf(list: FileItem[]) {
    const listing = new FolderListing(
      async () => ({ item: { ...file('f'), kind: 'dir' }, items: list, total: list.length }),
      '/f'
    );
    await listing.start();
    return listing;
  }

  it('names the file, steps over folders with the arrows, and closes with Escape', async () => {
    const listing = await listingOf(items);
    const onstep = vi.fn();
    const onclose = vi.fn();
    await render(PreviewFrame, { listing, index: 2, onstep, onclose });
    const dialog = page.getByRole('dialog', { name: 'Preview of b.bin' });
    await expect.element(dialog).toBeVisible();
    await expect.element(page.getByTestId('preview-fallback')).toBeVisible();
    await userEvent.keyboard('{ArrowLeft}');
    expect(onstep).toHaveBeenLastCalledWith(0, items[0]); // the folder is skipped
    await userEvent.keyboard('{ArrowRight}');
    expect(onstep).toHaveBeenLastCalledWith(3, items[3]);
    await expect
      .element(page.getByRole('link', { name: 'Download' }).first())
      .toHaveAttribute('href', '/api/v1/files/content?path=%2Ff%2Fb.bin');
    await userEvent.keyboard('{Escape}');
    await vi.waitFor(() => expect(onclose).toHaveBeenCalledOnce());
  });

  it('disables stepping past the ends and shows a failed view as the fallback card', async () => {
    const listing = await listingOf([file('only.png', 3, 'image/png')]);
    await render(PreviewFrame, { listing, index: 0, onstep: () => {}, onclose: () => {} });
    await expect.element(page.getByRole('button', { name: 'Previous file' })).toBeDisabled();
    await expect.element(page.getByRole('button', { name: 'Next file' })).toBeDisabled();
    // The image address answers nothing an <img> can show: the card explains.
    await expect.element(page.getByTestId('preview-fallback'), { timeout: 5000 }).toBeVisible();
    await expect.element(page.getByText(/This image cannot be shown/)).toBeVisible();
  });
});
