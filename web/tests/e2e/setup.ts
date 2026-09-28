// The system tests' server (S02.8-T02): the real binary with the embedded
// interface (run `pnpm build` first), on a fresh storage root holding the
// fixtures below. Prepared once, when Playwright loads its config; the
// workers find it through the environment.
import { execFileSync } from 'node:child_process';
import { copyFileSync, mkdirSync, mkdtempSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { deflateSync } from 'node:zlib';

const here = dirname(fileURLToPath(import.meta.url));
const repo = join(here, '..', '..', '..');

/** How many items the large folder has (NFR-003's folder size). */
export const largeFolderSize = 50_000;

/** A valid PDF with n pages, each saying its number. */
export function pdf(n: number): Buffer {
  const pages = Array.from({ length: n }, (_, i) => i);
  const objects = [
    '<< /Type /Catalog /Pages 2 0 R >>',
    `<< /Type /Pages /Kids [${pages.map((i) => `${3 + i * 2} 0 R`).join(' ')}] /Count ${n} >>`
  ];
  for (const i of pages) {
    const text = `BT /F1 24 Tf 40 60 Td (Page ${i + 1}) Tj ET`;
    objects.push(
      `<< /Type /Page /Parent 2 0 R /MediaBox [0 0 300 150] /Contents ${4 + i * 2} 0 R /Resources << /Font << /F1 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >> >> >> >>`,
      `<< /Length ${text.length} >>\nstream\n${text}\nendstream`
    );
  }
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
  return Buffer.from(out, 'latin1');
}

/** A small PNG of one color. */
export function png(width: number, height: number): Buffer {
  const crcTable = Array.from({ length: 256 }, (_, n) => {
    let c = n;
    for (let k = 0; k < 8; k++) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
    return c >>> 0;
  });
  const crc = (b: Buffer) => {
    let c = 0xffffffff;
    for (const x of b) c = crcTable[(c ^ x) & 0xff] ^ (c >>> 8);
    return (c ^ 0xffffffff) >>> 0;
  };
  const chunk = (type: string, data: Buffer) => {
    const len = Buffer.alloc(4);
    len.writeUInt32BE(data.length);
    const body = Buffer.concat([Buffer.from(type, 'latin1'), data]);
    const sum = Buffer.alloc(4);
    sum.writeUInt32BE(crc(body));
    return Buffer.concat([len, body, sum]);
  };
  const header = Buffer.alloc(13);
  header.writeUInt32BE(width, 0);
  header.writeUInt32BE(height, 4);
  header[8] = 8; // bit depth
  header[9] = 2; // RGB
  const row = Buffer.concat([Buffer.from([0]), Buffer.alloc(width * 3, 0x3c)]);
  const raw = Buffer.concat(Array.from({ length: height }, () => row));
  return Buffer.concat([
    Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]),
    chunk('IHDR', header),
    chunk('IDAT', deflateSync(raw)),
    chunk('IEND', Buffer.alloc(0))
  ]);
}

export interface E2EServer {
  root: string;
  binary: string;
  port: number;
}

/** Builds the binary and prepares the storage root, once per test run. */
export function prepare(port: number): E2EServer {
  if (process.env.E2E_ROOT && process.env.E2E_BINARY) {
    return { root: process.env.E2E_ROOT, binary: process.env.E2E_BINARY, port };
  }
  const work = mkdtempSync(join(tmpdir(), 'lan-e2e-'));
  const binary = join(work, process.platform === 'win32' ? 'nas.exe' : 'nas');
  execFileSync('go', ['build', '-o', binary, './cmd/local-ai-nas'], {
    cwd: repo,
    stdio: 'inherit'
  });

  const root = join(work, 'root');
  const files = join(root, 'files', 'u0001');
  const media = join(files, 'media');
  mkdirSync(media, { recursive: true });
  writeFileSync(join(media, 'photo.png'), png(64, 48));
  writeFileSync(join(media, 'doc.pdf'), pdf(3));
  writeFileSync(join(media, 'notes.txt'), 'Hello from the NAS.\n');
  writeFileSync(join(media, 'big.log'), 'log line\n'.repeat(40_000)); // 352 KiB, over the 256 KiB preview
  writeFileSync(
    join(media, 'page.html'),
    '<h1 id="injected">Injected</h1><script>window.ran = true</script>'
  );
  writeFileSync(
    join(media, 'active.svg'),
    '<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" onload="parent.document.title=\'svg ran\'">' +
      '<script>parent.document.title = "svg ran"</script><rect width="40" height="40" fill="teal"/></svg>'
  );
  writeFileSync(join(media, 'setup.exe'), Buffer.alloc(2048, 7));
  copyFileSync(join(here, 'fixtures', 'clip.webm'), join(media, 'clip.webm'));
  copyFileSync(join(here, 'fixtures', 'tone.webm'), join(media, 'tone.webm'));

  const large = join(files, 'large');
  mkdirSync(large);
  for (let i = 0; i < largeFolderSize; i++) {
    writeFileSync(join(large, `f${String(i).padStart(5, '0')}.txt`), '');
  }
  process.env.E2E_ROOT = root;
  process.env.E2E_BINARY = binary;
  return { root, binary, port };
}
