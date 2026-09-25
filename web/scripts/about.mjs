// Writes src/lib/generated/components.json: the packages the web interface
// ships (the "dependencies" of package.json), with their versions and
// licenses, for the About section of Settings (S02.2-T01). It runs before
// dev, check, and build, so the list always matches what is installed.
import { mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const web = join(dirname(fileURLToPath(import.meta.url)), '..');
const read = (path) => JSON.parse(readFileSync(path, 'utf8'));
const pkg = read(join(web, 'package.json'));

const components = Object.keys(pkg.dependencies ?? {})
  .sort()
  .map((name) => {
    const installed = read(join(web, 'node_modules', name, 'package.json'));
    return { name, version: installed.version, license: installed.license ?? 'unknown' };
  });

const out = join(web, 'src', 'lib', 'generated');
mkdirSync(out, { recursive: true });
writeFileSync(join(out, 'components.json'), JSON.stringify(components, null, 2) + '\n');
