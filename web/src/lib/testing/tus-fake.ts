// Test helper (S02.8-T01): a small in-memory tus server for the browser
// tests of the upload manager, served by Vite at the app's upload path
// while Vitest runs. It follows the protocol as the NAS uses it (core,
// creation, termination; the chunk limit in OPTIONS) and acts out the
// server's refusals by the target name:
//   "taken"   → 409 conflict while on_conflict is fail;
//   "CON."    → 400 invalid_name (reserved_name);
//   "nospace" → 507 insufficient_storage;
//   "flaky"   → the first PATCH answers 503 (retried by the client).
// GET /__tus/uploads lists what arrived, and GET /__tus/log the requests
// (method, upload, body size), for the tests to check. POST
// /api/v1/files/folders creates folders ("fail" in the path refuses), for
// folder uploads.
import type { IncomingMessage, ServerResponse } from 'node:http';
import type { Plugin } from 'vite';

const base = '/api/v1/files/uploads/';
export const fakeChunkBytes = 64 * 1024;

interface Upload {
  id: string;
  length: number;
  offset: number;
  target: string;
  policy: string;
  itemPath?: string;
  flaked?: boolean;
}

function problem(res: ServerResponse, status: number, code: string, detail: string, rule?: string) {
  res.writeHead(status, { 'Content-Type': 'application/problem+json', 'Tus-Resumable': '1.0.0' });
  res.end(JSON.stringify({ status, code, title: code, detail, ...(rule ? { rule } : {}) }));
}

function metadata(header: string | undefined): Record<string, string> {
  const out: Record<string, string> = {};
  for (const pair of (header ?? '').split(',')) {
    const [key, value] = pair.trim().split(' ');
    if (key) {
      out[key] = value ? Buffer.from(value, 'base64').toString('utf8') : '';
    }
  }
  return out;
}

function body(req: IncomingMessage): Promise<Buffer> {
  return new Promise((resolve, reject) => {
    const parts: Buffer[] = [];
    req.on('data', (c: Buffer) => parts.push(c));
    req.on('end', () => resolve(Buffer.concat(parts)));
    req.on('error', reject);
  });
}

/** "a.txt" → "a (1).txt", as the NAS names a kept-both upload. */
function numbered(path: string): string {
  const slash = path.lastIndexOf('/');
  const name = path.slice(slash + 1);
  const dot = name.lastIndexOf('.');
  const renamed = dot > 0 ? `${name.slice(0, dot)} (1)${name.slice(dot)}` : `${name} (1)`;
  return path.slice(0, slash + 1) + renamed;
}

export function tusFake(): Plugin {
  const uploads = new Map<string, Upload>();
  let next = 1;
  const folders: string[] = [];
  const log: { method: string; id: string; bytes: number }[] = [];
  return {
    name: 'local-ai-nas:tus-fake',
    configureServer(server) {
      server.middlewares.use(async (req, res, go) => {
        const url = new URL(req.url ?? '/', 'http://test');
        if (url.pathname === '/api/v1/files/folders' && req.method === 'POST') {
          const { path } = JSON.parse((await body(req)).toString('utf8')) as { path: string };
          folders.push(path);
          if (path.includes('fail')) {
            problem(res, 500, 'internal', 'the disk is on fire');
            return;
          }
          res.writeHead(201, { 'Content-Type': 'application/json' });
          res.end(
            JSON.stringify({
              path,
              name: path.split('/').at(-1),
              kind: 'dir',
              size: 0,
              mod_time: '2026-09-28T00:00:00Z'
            })
          );
          return;
        }
        if (url.pathname === '/__tus/folders') {
          res.writeHead(200, { 'Content-Type': 'application/json' });
          res.end(JSON.stringify(folders));
          return;
        }
        if (url.pathname === '/__tus/log') {
          res.writeHead(200, { 'Content-Type': 'application/json' });
          res.end(JSON.stringify(log));
          return;
        }
        if (url.pathname === '/__tus/uploads') {
          res.writeHead(200, { 'Content-Type': 'application/json' });
          res.end(JSON.stringify([...uploads.values()]));
          return;
        }
        if (!url.pathname.startsWith(base)) {
          go();
          return;
        }
        const id = url.pathname.slice(base.length);
        const tus = { 'Tus-Resumable': '1.0.0', 'Cache-Control': 'no-store' };
        switch (req.method) {
          case 'OPTIONS':
            res.writeHead(204, {
              ...tus,
              'Tus-Version': '1.0.0',
              'Tus-Extension': 'creation,termination',
              'Upload-Max-Chunk-Size': String(fakeChunkBytes)
            });
            res.end();
            return;
          case 'POST': {
            log.push({ method: 'POST', id: '', bytes: (await body(req)).length });
            const meta = metadata(req.headers['upload-metadata'] as string | undefined);
            const target = meta.target_path ?? '';
            const policy = meta.on_conflict || 'fail';
            if (target.includes('CON.')) {
              problem(res, 400, 'invalid_name', 'the name is reserved on Windows', 'reserved_name');
              return;
            }
            if (target.includes('nospace')) {
              problem(res, 507, 'insufficient_storage', 'not enough free space');
              return;
            }
            if (target.includes('taken') && policy === 'fail') {
              problem(res, 409, 'conflict', `an item already exists at ${target}`);
              return;
            }
            const upload: Upload = {
              id: `u${next++}`,
              length: Number(req.headers['upload-length']),
              offset: 0,
              target,
              policy
            };
            uploads.set(upload.id, upload);
            res.writeHead(201, { ...tus, Location: base + upload.id });
            res.end();
            return;
          }
          case 'HEAD': {
            const up = uploads.get(id);
            if (!up) {
              res.writeHead(404, tus);
              res.end();
              return;
            }
            res.writeHead(200, {
              ...tus,
              'Upload-Offset': String(up.offset),
              'Upload-Length': String(up.length)
            });
            res.end();
            return;
          }
          case 'PATCH': {
            const up = uploads.get(id);
            const data = await body(req);
            if (!up) {
              problem(res, 404, 'not_found', 'no such upload');
              return;
            }
            if (up.target.includes('flaky') && !up.flaked) {
              up.flaked = true;
              problem(res, 503, 'unavailable', 'busy, try again');
              return;
            }
            if (Number(req.headers['upload-offset']) !== up.offset) {
              res.writeHead(409, tus);
              res.end();
              return;
            }
            up.offset += data.length;
            const headers: Record<string, string> = { ...tus, 'Upload-Offset': String(up.offset) };
            if (up.offset === up.length) {
              up.itemPath =
                up.policy === 'rename' && up.target.includes('taken')
                  ? numbered(up.target)
                  : up.target;
              headers['Item-Path'] = up.itemPath;
            }
            res.writeHead(204, headers);
            res.end();
            return;
          }
          case 'DELETE':
            log.push({ method: 'DELETE', id, bytes: 0 });
            uploads.delete(id);
            res.writeHead(204, tus);
            res.end();
            return;
          default:
            go();
        }
      });
    }
  };
}
