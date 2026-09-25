// Checks the licenses of the web interface's npm packages (S02.1-T05,
// NFR-029). Packages that ship in the built interface ("dependencies" and
// what they pull in) must carry a license from scripts/allowed-licenses.txt,
// the list that also holds for Go code linked into the binary. Development
// tools may also use the licenses in devOnlyLicenses below; they are never
// shipped. The policy and its reasons are in docs/licensing.md.
//
// Usage (from the repository root or from web/): node scripts/check-web-licenses.mjs
import { execSync } from 'node:child_process';
import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = join(dirname(fileURLToPath(import.meta.url)), '..');
const web = join(root, 'web');

// Permissive licenses accepted for tools that are never shipped
// (docs/licensing.md): BlueOak-1.0.0 for minimatch (linters), Python-2.0
// for argparse (js-yaml in the API type generator, openapi-typescript).
const devOnlyLicenses = ['BlueOak-1.0.0', 'Python-2.0'];

const allowed = new Set(
  readFileSync(join(root, 'scripts', 'allowed-licenses.txt'), 'utf8')
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line && !line.startsWith('#'))
);

// An SPDX name without the -only or -or-later suffix, as the list keeps it.
function base(name) {
  return name.trim().replace(/-(only|or-later)$/, '').replace(/\+$/, '');
}

// Reports whether an SPDX expression is acceptable: every part of an AND,
// and at least one part of an OR. Parentheses are flattened.
function acceptable(expression, names) {
  const text = expression.replace(/[()]/g, ' ').trim();
  if (/\sOR\s/i.test(text)) {
    return text.split(/\sOR\s/i).some((part) => acceptable(part, names));
  }
  return text.split(/\sAND\s/i).every((part) => names.has(base(part)));
}

// pnpm licenses list --json: { "<license>": [{ name, versions, ... }] }.
function licenses(prod) {
  // A fixed command line, so pnpm's .cmd shim on Windows runs as well.
  const command = prod ? 'pnpm licenses list --json --prod' : 'pnpm licenses list --json';
  return JSON.parse(execSync(command, { cwd: web, encoding: 'utf8' }));
}

let failures = 0;
function check(label, byLicense, names) {
  let count = 0;
  for (const [license, packages] of Object.entries(byLicense)) {
    for (const pkg of packages) {
      count++;
      if (!acceptable(license, names)) {
        failures++;
        console.error(`not allowed (${label}): ${pkg.name}@${pkg.versions.join(', ')}: ${license}`);
      }
    }
  }
  console.log(`${label}: ${count} packages checked`);
}

const shipped = licenses(true);
check('shipped', shipped, allowed);
check('all, including dev tools', licenses(false), new Set([...allowed, ...devOnlyLicenses]));
if (failures > 0) {
  console.error(`${failures} package(s) with a license outside the policy (docs/licensing.md)`);
  process.exit(1);
}
console.log('all licenses are allowed');
